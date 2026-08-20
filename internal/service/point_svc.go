package service

import (
	"context"
	"time"

	"cleanroom-monitor/internal/model"
	"cleanroom-monitor/internal/store"
)

// PointService 监测点管理。
type PointService struct {
	points store.PointStore
	rooms  store.RoomStore
}

func NewPointService(points store.PointStore, rooms store.RoomStore) *PointService {
	return &PointService{points: points, rooms: rooms}
}

// Create 创建监测点（校验房间存在）。
func (s *PointService) Create(ctx context.Context, in *model.PointInput) (*model.MonitoringPoint, error) {
	if _, err := s.rooms.GetByID(ctx, in.RoomID); err != nil {
		return nil, err
	}
	p := &model.MonitoringPoint{
		RoomID:           in.RoomID,
		Code:             in.Code,
		ParamType:        in.ParamType,
		ThresholdMin:     in.ThresholdMin,
		ThresholdMax:     in.ThresholdMax,
		AlarmDurationSec: in.AlarmDurationSec,
		Enabled:          in.Enabled,
		CreatedAt:        time.Now(),
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	if err := s.points.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// List 列出监测点。
func (s *PointService) List(ctx context.Context, roomID int64) ([]*model.MonitoringPoint, error) {
	if roomID > 0 {
		return s.points.ListByRoom(ctx, roomID)
	}
	return s.points.ListAll(ctx)
}

// Get 获取监测点。
func (s *PointService) Get(ctx context.Context, id int64) (*model.MonitoringPoint, error) {
	return s.points.GetByID(ctx, id)
}

// Toggle 启停监测点。
func (s *PointService) Toggle(ctx context.Context, id int64, enabled bool) (*model.MonitoringPoint, error) {
	if err := s.points.SetEnabled(ctx, id, enabled); err != nil {
		return nil, err
	}
	return s.points.GetByID(ctx, id)
}