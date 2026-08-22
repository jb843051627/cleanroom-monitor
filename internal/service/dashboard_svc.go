package service

import (
	"context"
	"errors"
	"time"

	"cleanroom-monitor/internal/model"
	"cleanroom-monitor/internal/store"
)

// DashboardSnapshot 看板快照。
type DashboardSnapshot struct {
	Rooms      []*store.RoomSnapshot `json:"rooms"`
	OpenAlerts int                   `json:"open_alerts"`
	UpdatedAt  time.Time             `json:"updated_at"`
}

// DashboardService 看板服务：缓存快照刷新与读取。
type DashboardService struct {
	rooms    store.RoomStore
	points   store.PointStore
	cache    *store.Cache
	alerts   store.AlertStore
	readings store.ReadingStore
}

func NewDashboardService(rooms store.RoomStore, points store.PointStore, cache *store.Cache,
	alerts store.AlertStore, readings store.ReadingStore) *DashboardService {
	return &DashboardService{rooms: rooms, points: points, cache: cache, alerts: alerts, readings: readings}
}

// Snapshot 读取看板快照（缓存优先）。
func (s *DashboardService) Snapshot(ctx context.Context) (*DashboardSnapshot, error) {
	snaps := s.cache.GetAll()
	openAlerts := 0
	for _, snap := range snaps {
		openAlerts += snap.OpenAlerts
	}
	return &DashboardSnapshot{
		Rooms:      snaps,
		OpenAlerts: openAlerts,
		UpdatedAt:  time.Now(),
	}, nil
}

// Refresh 刷新全部房间快照（后台定期调用）。
// 单个房间失败不中断其余房间的刷新，但错误会被聚合返回，避免被静默吞掉。
func (s *DashboardService) Refresh(ctx context.Context) error {
	rooms, err := s.rooms.List(ctx, 1000, 0)
	if err != nil {
		return err
	}
	var errs []error
	for _, r := range rooms {
		snap, err := s.buildRoomSnapshot(ctx, r)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		s.cache.Set(r.ID, snap)
	}
	return errors.Join(errs...)
}

func (s *DashboardService) buildRoomSnapshot(ctx context.Context, room *model.Room) (*store.RoomSnapshot, error) {
	snap := &store.RoomSnapshot{
		RoomID:   room.ID,
		RoomCode: room.Code,
		Status:   room.Status,
	}
	openAlerts, err := s.alerts.CountOpenByRoom(ctx, room.ID)
	if err != nil {
		return nil, err
	}
	snap.OpenAlerts = openAlerts
	points, err := s.points.ListByRoom(ctx, room.ID)
	if err != nil {
		return nil, err
	}
	var realtime []model.RealtimeReading
	for _, p := range points {
		latest, err := s.readings.LatestByPoint(ctx, p.ID)
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				continue
			}
			return nil, err
		}
		realtime = append(realtime, model.RealtimeReading{
			PointID:    p.ID,
			RoomID:     p.RoomID,
			ParamType:  p.ParamType,
			Value:      latest.Value,
			MeasuredAt: latest.MeasuredAt,
			WithinRange: latest.Value >= p.ThresholdMin && latest.Value <= p.ThresholdMax,
		})
	}
	snap.Realtime = realtime
	return snap, nil
}