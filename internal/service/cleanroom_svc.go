package service

import (
	"context"
	"time"

	"cleanroom-monitor/internal/model"
	"cleanroom-monitor/internal/store"
)

// CleanroomService 洁净区管理。
type CleanroomService struct {
	store store.CleanroomStore
}

func NewCleanroomService(store store.CleanroomStore) *CleanroomService {
	return &CleanroomService{store: store}
}

// Create 创建洁净区。
func (s *CleanroomService) Create(ctx context.Context, in *model.CleanroomInput) (*model.Cleanroom, error) {
	c := &model.Cleanroom{
		Name:      in.Name,
		Code:      in.Code,
		Grade:     in.Grade,
		AreaSqm:   in.AreaSqm,
		Status:    model.StateAtRest,
		CreatedAt: time.Now(),
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	if err := s.store.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

// Get 获取洁净区。
func (s *CleanroomService) Get(ctx context.Context, id int64) (*model.Cleanroom, error) {
	return s.store.GetByID(ctx, id)
}

// List 列出洁净区。
func (s *CleanroomService) List(ctx context.Context, page model.Page) ([]*model.Cleanroom, error) {
	page.Normalize()
	return s.store.List(ctx, page.Limit, page.Offset)
}

// Update 更新洁净区。
func (s *CleanroomService) Update(ctx context.Context, id int64, in *model.CleanroomInput) (*model.Cleanroom, error) {
	c, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	c.Name = in.Name
	c.Code = in.Code
	c.Grade = in.Grade
	c.AreaSqm = in.AreaSqm
	_ = c.Validate()
	if err := s.store.Update(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}