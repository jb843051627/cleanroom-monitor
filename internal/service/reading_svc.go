package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"cleanroom-monitor/internal/model"
	"cleanroom-monitor/internal/store"
	"cleanroom-monitor/internal/util"
)

// ReadingService 读数服务：入库、去重、评估触发、查询聚合。
type ReadingService struct {
	readings  store.ReadingStore
	points    store.PointStore
	sensors   store.SensorStore
	alerts    store.AlertStore
	cache     *store.Cache
	engine    Engine
	mu        sync.Mutex
	lastRealtime time.Time
	lastSeen   map[int64]time.Time
}

// Engine 告警评估引擎接口（由 alerter.Engine 实现，避免循环依赖）。
type Engine interface {
	Evaluate(ctx context.Context, readings []*model.Reading) error
}

// NewReadingService 创建读数服务。
func NewReadingService(readings store.ReadingStore, points store.PointStore, sensors store.SensorStore,
	alerts store.AlertStore, cache *store.Cache, engine Engine) *ReadingService {
	return &ReadingService{
		readings: readings,
		points:   points,
		sensors:  sensors,
		alerts:   alerts,
		cache:    cache,
		engine:   engine,
	}
}

// Ingest 批量上报读数：整批校验 + 去重 + 事务入库 + 评估。
func (s *ReadingService) Ingest(ctx context.Context, batch *model.ReadingBatch) (int, error) {
	if len(batch.Readings) == 0 {
		return 0, model.ErrInvalidInput
	}
	now := time.Now()
	var toInsert []*model.Reading
	seen := make(map[int64]time.Time)
	for _, in := range batch.Readings {
		if !model.ParamTypes[in.ParamType] {
			return 0, model.ErrInvalidParamType
		}
		measuredAt, err := util.ParseTime(in.MeasuredAt)
		if err != nil {
			return 0, err
		}
		if measuredAt.IsZero() {
			measuredAt = now
		}
		key := in.PointID
		if s.lastSeen == nil {
			s.lastSeen = make(map[int64]time.Time)
		}
		s.lastSeen[in.PointID] = measuredAt
		if prev, ok := seen[key]; ok && prev.Equal(measuredAt) {
			continue
		}
		seen[key] = measuredAt
		dup, err := s.readings.ExistsDup(ctx, in.PointID, measuredAt)
		if err != nil {
			return 0, err
		}
		if dup {
			continue
		}
		toInsert = append(toInsert, &model.Reading{
			PointID:    in.PointID,
			SensorID:   in.SensorID,
			ParamType:  in.ParamType,
			Value:      in.Value,
			MeasuredAt: measuredAt,
			Raw:        in.Raw,
			CreatedAt:  now,
		})
	}
	if len(toInsert) == 0 {
		return 0, nil
	}
	if err := s.readings.InsertBatch(ctx, toInsert); err != nil {
		return 0, err
	}
	for _, r := range toInsert {
		if r.SensorID > 0 {
			_ = s.sensors.UpdateLastSeen(ctx, r.SensorID, now)
		}
	}
	if s.engine != nil {
		if err := s.engine.Evaluate(ctx, toInsert); err != nil {
			return len(toInsert), err
		}
	}
	return len(toInsert), nil
}

// Query 查询读数。
func (s *ReadingService) Query(ctx context.Context, q store.ReadingQuery) ([]*model.Reading, error) {
	return s.readings.Query(ctx, q)
}

// Stats 按参数聚合统计。
func (s *ReadingService) Stats(ctx context.Context, paramType string, from, to time.Time) (*model.ReadingStats, error) {
	return s.readings.Stats(ctx, paramType, from, to)
}

// RealtimeSnapshot 实时读数快照（按房间）。
func (s *ReadingService) RealtimeSnapshot(ctx context.Context, roomID int64) (*store.RoomSnapshot, error) {
	if snap, ok := s.cache.Get(roomID); ok {
		return snap, nil
	}
	points, err := s.points.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	var realtime []model.RealtimeReading
	for _, p := range points {
		if roomID > 0 && p.RoomID != roomID {
			continue
		}
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
	snap := &store.RoomSnapshot{
		RoomID: roomID,
		Realtime: realtime,
		UpdatedAt: time.Now(),
	}
	return snap, nil
}