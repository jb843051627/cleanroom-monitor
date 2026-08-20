package service

import (
	"context"
	"time"

	"cleanroom-monitor/internal/model"
	"cleanroom-monitor/internal/store"
)

// RoomService 房间管理。
type RoomService struct {
	rooms       store.RoomStore
	cleanrooms  store.CleanroomStore
}

func NewRoomService(rooms store.RoomStore, cleanrooms store.CleanroomStore) *RoomService {
	return &RoomService{rooms: rooms, cleanrooms: cleanrooms}
}

// Create 创建房间（校验洁净区存在）。
func (s *RoomService) Create(ctx context.Context, in *model.RoomInput) (*model.Room, error) {
	if _, err := s.cleanrooms.GetByID(ctx, in.CleanroomID); err != nil {
		return nil, err
	}
	r := &model.Room{
		CleanroomID:       in.CleanroomID,
		Name:              in.Name,
		Code:              in.Code,
		Kind:              in.Kind,
		AreaSqm:           in.AreaSqm,
		TargetPressure:    in.TargetPressure,
		PressureTolerance: in.PressureTolerance,
		TargetTemp:        in.TargetTemp,
		TargetHumidity:    in.TargetHumidity,
		Status:            model.NewRoomStatus,
		CreatedAt:         time.Now(),
	}
	if err := r.Validate(); err != nil {
		return nil, err
	}
	if err := s.rooms.Create(ctx, r); err != nil {
		return nil, err
	}
	return r, nil
}

// List 列出房间。
func (s *RoomService) List(ctx context.Context, page model.Page) ([]*model.Room, error) {
	page.Normalize()
	return s.rooms.List(ctx, page.Limit, page.Offset)
}

// ListByCleanroom 按洁净区列出房间。
func (s *RoomService) ListByCleanroom(ctx context.Context, cleanroomID int64) ([]*model.Room, error) {
	return s.rooms.ListByCleanroom(ctx, cleanroomID)
}

// Get 获取房间。
func (s *RoomService) Get(ctx context.Context, id int64) (*model.Room, error) {
	return s.rooms.GetByID(ctx, id)
}

// Update 更新房间。
func (s *RoomService) Update(ctx context.Context, id int64, in *model.RoomInput) (*model.Room, error) {
	r, err := s.rooms.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	r.Name = in.Name
	r.Code = in.Code
	r.Kind = in.Kind
	r.AreaSqm = in.AreaSqm
	r.TargetPressure = in.TargetPressure
	r.PressureTolerance = in.PressureTolerance
	r.TargetTemp = in.TargetTemp
	r.TargetHumidity = in.TargetHumidity
	if err := r.Validate(); err != nil {
		return nil, err
	}
	if err := s.rooms.Update(ctx, r); err != nil {
		return nil, err
	}
	return r, nil
}

// GetStatus 获取房间当前状态。
func (s *RoomService) GetStatus(ctx context.Context, id int64) (*model.CleanStatus, error) {
	r, err := s.rooms.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &model.CleanStatus{
		RoomID:  r.ID,
		State:   r.Status,
		ChangedAt: time.Now(),
	}, nil
}