package service

import (
	"context"
	"time"

	"cleanroom-monitor/internal/model"
	"cleanroom-monitor/internal/store"
)

// RuleService 告警规则管理。
type RuleService struct {
	rules store.RuleStore
	rooms store.RoomStore
}

func NewRuleService(rules store.RuleStore, rooms store.RoomStore) *RuleService {
	return &RuleService{rules: rules, rooms: rooms}
}

// Create 创建规则（校验房间存在）。
func (s *RuleService) Create(ctx context.Context, in *model.RuleInput) (*model.AlertRule, error) {
	if _, err := s.rooms.GetByID(ctx, in.RoomID); err != nil {
		return nil, err
	}
	r := &model.AlertRule{
		RoomID:      in.RoomID,
		Code:        in.Code,
		ParamType:   in.ParamType,
		Op:          in.Op,
		Threshold:   in.Threshold,
		DurationSec: in.DurationSec,
		Level:       in.Level,
		Enabled:     in.Enabled,
		CreatedAt:   time.Now(),
	}
	if err := r.Validate(); err != nil {
		return nil, err
	}
	if err := s.rules.Create(ctx, r); err != nil {
		return nil, err
	}
	return r, nil
}

// List 列出规则（可按房间过滤）。
func (s *RuleService) List(ctx context.Context, roomID int64) ([]*model.AlertRule, error) {
	if roomID > 0 {
		return s.rules.ListByRoom(ctx, roomID)
	}
	return s.rules.ListAll(ctx)
}

// Toggle 启停规则。
func (s *RuleService) Toggle(ctx context.Context, id int64, enabled bool) (*model.AlertRule, error) {
	if err := s.rules.SetEnabled(ctx, id, enabled); err != nil {
		return nil, err
	}
	return s.rules.GetByID(ctx, id)
}