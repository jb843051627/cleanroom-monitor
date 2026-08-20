package service

import (
	"context"
	"fmt"
	"time"

	"cleanroom-monitor/internal/model"
	"cleanroom-monitor/internal/store"
)

// BatchService 批次管理：驱动房间洁净状态机。
type BatchService struct {
	batches store.BatchStore
	rooms   store.RoomStore
	status  store.StatusStore
	alerts  store.AlertStore
}

func NewBatchService(batches store.BatchStore, rooms store.RoomStore, status store.StatusStore, alerts store.AlertStore) *BatchService {
	return &BatchService{batches: batches, rooms: rooms, status: status, alerts: alerts}
}

// Start 启动批次：校验房间存在、无进行中批次，状态机 at_rest → normal。
func (s *BatchService) Start(ctx context.Context, in *model.BatchInput) (*model.CleanBatch, error) {
	room, err := s.rooms.GetByID(ctx, in.RoomID)
	if err != nil {
		return nil, err
	}
	if _, err := s.batches.GetActiveByRoom(ctx, in.RoomID); err == nil {
		return nil, model.ErrBatchActive
	}
	b := &model.CleanBatch{
		RoomID:    in.RoomID,
		Name:      in.Name,
		Product:   in.Product,
		Phase:     in.Phase,
		StartAt:   time.Now(),
		Status:    model.BatchInProgress,
		CreatedAt: time.Now(),
	}
	if err := b.Validate(); err != nil {
		return nil, err
	}
	if err := s.batches.Create(ctx, b); err != nil {
		return nil, err
	}
	_ = s.transition(ctx, room, model.StateNormal, model.ReasonBatchStarted, b.ID)
	return b, nil
}

// Complete 完成批次：状态机 → release（若房间无未决 alarm）否则 restricted。
func (s *BatchService) Complete(ctx context.Context, id int64) (*model.CleanBatch, error) {
	b, err := s.batches.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("批次不存在: %v", err)
	}
	if b.Status != model.BatchInProgress && b.Status != model.BatchPlanning {
		return nil, model.ErrConflict
	}
	room, err := s.rooms.GetByID(ctx, b.RoomID)
	if err != nil {
		return nil, fmt.Errorf("房间不存在: %v", err)
	}
	endAt := time.Now()
	if err := s.batches.UpdateStatus(ctx, id, model.BatchCompleted, &endAt); err != nil {
		return nil, err
	}
	b.Status = model.BatchCompleted
	b.EndAt = &endAt
	openAlarm, err := s.alerts.CountOpenByLevel(ctx, b.RoomID, model.AlertAlarm)
	if err != nil {
		return nil, err
	}
	state := model.StateRelease
	reason := model.ReasonBatchRelease
	if openAlarm > 0 {
		state = model.StateRestricted
		reason = model.ReasonAlarmAlert
	}
	_ = s.transition(ctx, room, state, reason, b.ID)
	return b, nil
}

// Abort 中止批次：状态机 → at_rest。
func (s *BatchService) Abort(ctx context.Context, id int64) (*model.CleanBatch, error) {
	b, err := s.batches.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if b.Status != model.BatchInProgress && b.Status != model.BatchPlanning {
		return nil, model.ErrConflict
	}
	room, err := s.rooms.GetByID(ctx, b.RoomID)
	if err != nil {
		return nil, err
	}
	endAt := time.Now()
	if err := s.batches.UpdateStatus(ctx, id, model.BatchAborted, &endAt); err != nil {
		return nil, err
	}
	b.Status = model.BatchAborted
	b.EndAt = &endAt
	_ = s.transition(ctx, room, model.StateAtRest, model.ReasonBatchAborted, b.ID)
	return b, nil
}

// List 列出批次。
func (s *BatchService) List(ctx context.Context, page model.Page) ([]*model.CleanBatch, error) {
	page.Normalize()
	return s.batches.List(ctx, page.Limit, page.Offset)
}

// ListByRoom 按房间列出批次。
func (s *BatchService) ListByRoom(ctx context.Context, roomID int64) ([]*model.CleanBatch, error) {
	return s.batches.ListByRoom(ctx, roomID)
}

func (s *BatchService) transition(ctx context.Context, room *model.Room, to, reason string, batchID int64) error {
	if !model.CanTransition(room.Status, to) {
		return model.ErrStateConflict
	}
	if err := s.rooms.UpdateStatus(ctx, room.ID, to); err != nil {
		return err
	}
	return s.status.Append(ctx, &model.CleanStatus{
		RoomID:    room.ID,
		BatchID:   batchID,
		State:     to,
		Reason:    reason,
		ChangedAt: time.Now(),
	})
}