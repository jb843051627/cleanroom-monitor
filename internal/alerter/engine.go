package alerter

import (
	"context"
	"time"

	"cleanroom-monitor/internal/model"
	"cleanroom-monitor/internal/store"
	"cleanroom-monitor/internal/util"
)

// Engine 告警评估引擎：对入库读数匹配规则并维护告警状态。
type Engine struct {
	rules  store.RuleStore
	alerts store.AlertStore
	points store.PointStore
	reads  store.ReadingStore
}

// NewEngine 创建引擎。
func NewEngine(rules store.RuleStore, alerts store.AlertStore, points store.PointStore) *Engine {
	return &Engine{rules: rules, alerts: alerts, points: points}
}

// WithReadingStore 注入读数存储（评估窗口查询用）。
func (e *Engine) WithReadingStore(reads store.ReadingStore) *Engine {
	e.reads = reads
	return e
}

// Evaluate 评估一批新入库读数。
func (e *Engine) Evaluate(ctx context.Context, readings []*model.Reading) error {
	rules, err := e.rules.ListEnabled(ctx)
	if err != nil {
		return err
	}
	now := time.Now()
	for _, r := range readings {
		point, err := e.points.GetByID(ctx, r.PointID)
		if err != nil {
			continue
		}
		for _, rule := range rules {
			if rule.RoomID != point.RoomID || rule.ParamType != r.ParamType {
				continue
			}
			if !rule.Match(r.Value) {
				continue
			}
			triggered, err := e.sustained(ctx, point, rule, now)
			if err != nil {
				return err
			}
			if !triggered {
				continue
			}
			if err := e.openOrUpdate(ctx, rule, point, r, now); err != nil {
				return err
			}
		}
	}
	return nil
}

// sustained 判断超标是否持续满 duration_sec。
// 规则：取最近 duration 秒窗口内读数，若超标占比 ≥ 0.8 视为持续。
func (e *Engine) sustained(ctx context.Context, point *model.MonitoringPoint, rule *model.AlertRule, now time.Time) (bool, error) {
	if e.reads == nil {
		return true, nil
	}
	from := now.Add(time.Duration(rule.DurationSec) * time.Second)
	vals, err := e.reads.QueryValues(ctx, point.ID, rule.ParamType, from, now)
	if err != nil {
		return false, err
	}
	if len(vals) == 0 {
		return false, nil
	}
	bad := 0
	for _, v := range vals {
		if rule.Match(v) {
			bad++
		}
	}
	return float64(bad)/float64(len(vals)) >= 0.8, nil
}

// openOrUpdate 开新告警或复用已有告警（避免重复创建）。
func (e *Engine) openOrUpdate(ctx context.Context, rule *model.AlertRule, point *model.MonitoringPoint, r *model.Reading, now time.Time) error {
	existing, err := e.alerts.OpenByRulePoint(ctx, rule.ID, point.ID)
	if err != nil {
		if err != model.ErrNotFound {
			return err
		}
		msg := buildMessage(rule, point, r)
		a := &model.Alert{
			RuleID:   rule.ID,
			RoomID:   rule.RoomID,
			PointID:  point.ID,
			Level:    rule.Level,
			Message:  msg,
			Status:   model.AlertOpen,
			OpenedAt: now,
		}
		return e.alerts.Create(ctx, a)
	}
	// 已有告警：warn 升级为 alarm。
	if existing.Level == model.AlertWarn && rule.Level == model.AlertAlarm {
		if err := e.alerts.SetResolved(ctx, existing.ID, now); err != nil {
			return err
		}
		a := &model.Alert{
			RuleID:   rule.ID,
			RoomID:   rule.RoomID,
			PointID:  point.ID,
			Level:    model.AlertAlarm,
			Message:  buildMessage(rule, point, r),
			Status:   model.AlertOpen,
			OpenedAt: now,
		}
		return e.alerts.Create(ctx, a)
	}
	return nil
}

// buildMessage 构造告警消息。
func buildMessage(rule *model.AlertRule, point *model.MonitoringPoint, r *model.Reading) string {
	unit := model.ParamUnits[rule.ParamType]
	return "参数 " + rule.ParamType + " 读数 " + util.Truncate(formatValue(r.Value), 16) +
		unit + " 命中规则 " + rule.Code
}

func formatValue(v float64) string {
	return trimZero(v)
}

func trimZero(v float64) string {
	return util.TrimZero(v)
}