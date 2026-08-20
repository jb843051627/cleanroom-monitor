package service

import (
	"context"
	"time"

	"cleanroom-monitor/internal/model"
	"cleanroom-monitor/internal/store"
)

// AlertService 告警管理。
type AlertService struct {
	alerts store.AlertStore
}

func NewAlertService(alerts store.AlertStore) *AlertService {
	return &AlertService{alerts: alerts}
}

// List 查询告警列表。
func (s *AlertService) List(ctx context.Context, in model.AlertInput) ([]*model.Alert, error) {
	return s.alerts.List(ctx, in)
}

// Acknowledge 确认告警。
func (s *AlertService) Acknowledge(ctx context.Context, id int64) (*model.Alert, error) {
	a, err := s.alerts.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if a.Status == model.AlertResolved {
		return nil, model.ErrConflict
	}
	if err := s.alerts.SetAck(ctx, id, time.Now()); err != nil {
		return nil, err
	}
	a.Status = model.AlertAcknowledged
	now := time.Now()
	a.AckAt = &now
	return a, nil
}

// Resolve 解决告警。
func (s *AlertService) Resolve(ctx context.Context, id int64) (*model.Alert, error) {
	a, err := s.alerts.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if a.Status == model.AlertResolved {
		return nil, model.ErrConflict
	}
	if err := s.alerts.SetResolved(ctx, id, time.Now()); err != nil {
		return nil, err
	}
	a.Status = model.AlertResolved
	now := time.Now()
	a.ResolvedAt = &now
	return a, nil
}