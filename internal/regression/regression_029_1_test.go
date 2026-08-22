package service

import (
	"testing"
	"time"

	"cleanroom-monitor/internal/model"
)

func TestBug29_WarnAlertDoesNotBlockRelease(t *testing.T) {
	svc := newTestServices(t)
	cr := mustCleanroom(t, svc)
	room := mustRoom(t, svc, cr)
	point := mustPoint(t, svc, room)
	if _, err := svc.Rules.Create(testCtx(), &model.RuleInput{RoomID: room.ID, Code: "WARN-TEMP", ParamType: model.ParamTemp, Op: model.OpGt, Threshold: 30, DurationSec: 1, Level: model.AlertWarn, Enabled: true}); err != nil {
		t.Fatalf("create rule: %v", err)
	}
	if _, err := svc.Readings.Ingest(testCtx(), &model.ReadingBatch{Readings: []model.ReadingInput{{PointID: point.ID, SensorID: 1, ParamType: model.ParamTemp, Value: 40, MeasuredAt: time.Now().Format(time.RFC3339Nano)}}}); err != nil {
		t.Fatalf("ingest: %v", err)
	}
	batch, err := svc.Batches.Start(testCtx(), &model.BatchInput{RoomID: room.ID, Name: "放行批次", Product: "疫苗", Phase: "灌装"})
	if err != nil {
		t.Fatalf("start batch: %v", err)
	}
	if _, err := svc.Batches.Complete(testCtx(), batch.ID); err != nil {
		t.Fatalf("complete batch: %v", err)
	}
	updated, err := svc.Rooms.Get(testCtx(), room.ID)
	if err != nil {
		t.Fatalf("get room: %v", err)
	}
	if updated.Status != model.StateRelease {
		t.Fatalf("只有 warn 告警时应放行，得到状态 %s", updated.Status)
	}
}
