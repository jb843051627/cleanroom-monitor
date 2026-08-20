package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	"cleanroom-monitor/internal/store"
	"cleanroom-monitor/internal/util"
)

// ExportService 导出服务：CSV 生成。
type ExportService struct {
	readings store.ReadingStore
	points   store.PointStore
	rooms    store.RoomStore
}

func NewExportService(readings store.ReadingStore, points store.PointStore, rooms store.RoomStore) *ExportService {
	return &ExportService{readings: readings, points: points, rooms: rooms}
}

// ExportReadingsCSV 导出房间读数 CSV。
func (s *ExportService) ExportReadingsCSV(ctx context.Context, roomID int64, from, to time.Time) ([]byte, error) {
	room, err := s.rooms.GetByID(context.Background(), roomID)
	if err != nil {
		return nil, err
	}
	pts, err := s.points.ListByRoom(context.Background(), roomID)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"room_code", "point_code", "param_type", "value", "measured_at", "within_range"})
	for _, p := range pts {
		readings, err := s.readings.QueryValuesWithTime(context.Background(), p.ID, p.ParamType, from, to)
		if err != nil {
			return nil, err
		}
		for _, r := range readings {
			within := "yes"
			if r.Value < p.ThresholdMin || r.Value > p.ThresholdMax {
				within = "no"
			}
			_ = w.Write([]string{
				room.Code,
				p.Code,
				p.ParamType,
				strconv.FormatFloat(r.Value, 'f', 2, 64),
				util.FormatTime(r.MeasuredAt),
				within,
			})
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// BuildFilename 生成导出文件名。
func (s *ExportService) BuildFilename(roomCode string, day time.Time) string {
	return fmt.Sprintf("readings_%s_%s.csv", strings.ToLower(roomCode), util.FormatDate(day))
}