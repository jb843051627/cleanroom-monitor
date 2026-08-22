package service

import (
	"testing"
	"time"

	"cleanroom-monitor/internal/model"
)

func TestBug21_WrappedNotFoundCreatesAlert(t *testing.T) {
	svc := newTestServices(t)
	cr := mustCleanroom(t, svc)
	room := mustRoom(t, svc, cr)
	point := mustPoint(t, svc, room)
	if _, err := svc.Rules.Create(testCtx(), &model.RuleInput{RoomID: room.ID, Code: "R1", ParamType: model.ParamTemp, Op: model.OpGt, Threshold: 30, DurationSec: 1, Level: model.AlertWarn, Enabled: true}); err != nil {
		t.Fatalf("create rule: %v", err)
	}
	err := svc.Engine.Evaluate(testCtx(), []*model.Reading{{PointID: point.ID, ParamType: model.ParamTemp, Value: 40, MeasuredAt: time.Now()}})
	if err != nil {
		t.Fatalf("首次超标读数应创建告警，而非返回 not-found: %v", err)
	}
	count, err := svc.Alerts.List(testCtx(), model.AlertInput{RoomID: room.ID, Limit: 10})
	if err != nil || len(count) != 1 {
		t.Fatalf("应创建 1 条告警，得到 %d, err=%v", len(count), err)
	}
}
