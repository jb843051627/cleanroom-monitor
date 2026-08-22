package service

import (
	"context"
	"path/filepath"
	"testing"

	"cleanroom-monitor/internal/config"
	"cleanroom-monitor/internal/model"
	"cleanroom-monitor/internal/store"
)

func newTestServices(t *testing.T) *Services {
	t.Helper()
	db, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewServices(db, config.New())
}

func testCtx() context.Context { return context.Background() }

func mustCleanroom(t *testing.T, svc *Services) *model.Cleanroom {
	t.Helper()
	c, err := svc.Cleanrooms.Create(testCtx(), &model.CleanroomInput{Name: "药厂", Code: "CR1", Grade: "ISO7", AreaSqm: 100})
	if err != nil {
		t.Fatalf("create cleanroom: %v", err)
	}
	return c
}

func mustRoom(t *testing.T, svc *Services, cr *model.Cleanroom) *model.Room {
	t.Helper()
	r, err := svc.Rooms.Create(testCtx(), &model.RoomInput{
		CleanroomID: cr.ID, Name: "灌装间", Code: "RM1", Kind: "atelier", AreaSqm: 50,
		TargetPressure: 25, PressureTolerance: 5, TargetTemp: 22, TargetHumidity: 50,
	})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	return r
}

func mustPoint(t *testing.T, svc *Services, room *model.Room) *model.MonitoringPoint {
	t.Helper()
	p, err := svc.Points.Create(testCtx(), &model.PointInput{
		RoomID: room.ID, Code: "PT1", ParamType: model.ParamTemp,
		ThresholdMin: 18, ThresholdMax: 26, AlarmDurationSec: 5, Enabled: true,
	})
	if err != nil {
		t.Fatalf("create point: %v", err)
	}
	return p
}
