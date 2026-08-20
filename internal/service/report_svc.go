package service

import (
	"context"
	"time"

	"cleanroom-monitor/internal/model"
	"cleanroom-monitor/internal/store"
	"cleanroom-monitor/internal/util"
)

// ReportService 报表服务：趋势、汇总、达标率。
type ReportService struct {
	readings    store.ReadingStore
	points      store.PointStore
	rooms       store.RoomStore
	cleanrooms  store.CleanroomStore
	alerts      store.AlertStore
	status      store.StatusStore
}

func NewReportService(readings store.ReadingStore, points store.PointStore, rooms store.RoomStore,
	cleanrooms store.CleanroomStore, alerts store.AlertStore, status store.StatusStore) *ReportService {
	return &ReportService{readings: readings, points: points, rooms: rooms, cleanrooms: cleanrooms, alerts: alerts, status: status}
}

// TrendPoint 趋势点。
type TrendPoint struct {
	Time  time.Time `json:"time"`
	Value float64   `json:"value"`
	Min   float64   `json:"min"`
	Max   float64   `json:"max"`
}

// TrendResult 趋势报表结果。
type TrendResult struct {
	RoomCode  string       `json:"room_code"`
	PointCode string       `json:"point_code"`
	ParamType string       `json:"param_type"`
	Unit      string       `json:"unit"`
	Points    []TrendPoint `json:"points"`
}

// Trend 房间某参数趋势。
func (s *ReportService) Trend(ctx context.Context, roomID int64, paramType string, from, to time.Time, limit int) (*TrendResult, error) {
	room, _ := s.rooms.GetByID(ctx, roomID)
	if !model.ParamTypes[paramType] {
		return nil, model.ErrInvalidParamType
	}
	pts, err := s.points.ListByRoom(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	res := &TrendResult{
		RoomCode:  room.Code,
		ParamType: paramType,
		Unit:      model.ParamUnits[paramType],
	}
	for _, p := range pts {
		if p.ParamType != paramType {
			continue
		}
		readings, err := s.readings.QueryValuesWithTime(ctx, p.ID, paramType, from, to)
		if err != nil {
			return nil, err
		}
		if len(readings) == 0 {
			continue
		}
		if len(res.PointCode) == 0 {
			res.PointCode = p.Code
		}
		for _, r := range readings {
			res.Points = append(res.Points, TrendPoint{
				Time:  r.MeasuredAt,
				Value: r.Value,
				Min:   p.ThresholdMin,
				Max:   p.ThresholdMax,
			})
		}
	}
	if len(res.Points) > limit {
		res.Points = res.Points[len(res.Points)-limit:]
	}
	return res, nil
}

// RoomCompliance 房间达标率。
type RoomCompliance struct {
	RoomID     int64   `json:"room_id"`
	RoomCode   string  `json:"room_code"`
	ParamType  string  `json:"param_type"`
	Readings   int     `json:"readings"`
	OutOfRange int     `json:"out_of_range"`
	Rate       float64 `json:"rate"` // 0-100
}

// SummaryResult 汇总报表。
type SummaryResult struct {
	GeneratedAt time.Time         `json:"generated_at"`
	RoomCount   int               `json:"room_count"`
	Cleanrooms  int               `json:"cleanrooms"`
	Compliance  []RoomCompliance  `json:"compliance"`
	OpenAlerts  int               `json:"open_alerts"`
	DueSensors  int               `json:"due_sensors"`
}

// Summary 汇总各房间各参数达标率。
func (s *ReportService) Summary(ctx context.Context, from, to time.Time) (*SummaryResult, error) {
	res := &SummaryResult{GeneratedAt: time.Now()}
	rooms, err := s.rooms.List(ctx, 1000, 0)
	if err != nil {
		return nil, err
	}
	res.RoomCount = len(rooms)
	cc, err := s.cleanrooms.Count(ctx)
	if err != nil {
		return nil, err
	}
	res.Cleanrooms = cc
	openAlerts, err := s.alerts.List(ctx, model.AlertInput{Status: model.AlertOpen})
	if err != nil {
		return nil, err
	}
	res.OpenAlerts = len(openAlerts)
	for _, r := range rooms {
		pts, err := s.points.ListByRoom(ctx, r.ID)
		if err != nil {
			return nil, err
		}
		seen := map[string]bool{}
		for _, p := range pts {
			if seen[p.ParamType] {
				continue
			}
			seen[p.ParamType] = true
			vals, err := s.readings.QueryValues(ctx, p.ID, p.ParamType, from, to)
			if err != nil {
				return nil, err
			}
			oor := 0
			for _, v := range vals {
				if v < p.ThresholdMin || v > p.ThresholdMax {
					oor++
				}
			}
			rate := 0.0
			if len(vals) > 0 {
				rate = float64(len(vals)-oor) / float64(len(vals)) * 100
			}
			res.Compliance = append(res.Compliance, RoomCompliance{
				RoomID:     r.ID,
				RoomCode:   r.Code,
				ParamType:  p.ParamType,
				Readings:   len(vals),
				OutOfRange: oor,
				Rate:       util.Round(rate, 1),
			})
		}
	}
	return res, nil
}

// StateDistribution 房间状态分布（看板用）。
type StateDistribution struct {
	State string `json:"state"`
	Count int    `json:"count"`
}

// StateDistribution 统计各状态房间数。
func (s *ReportService) StateDistribution(ctx context.Context) ([]StateDistribution, error) {
	var out []StateDistribution
	for state := range model.States {
		n, err := s.status.CountByState(ctx, state)
		if err != nil {
			return nil, err
		}
		out = append(out, StateDistribution{State: state, Count: n})
	}
	return out, nil
}