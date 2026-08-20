package model

import (
	"errors"
	"strings"
	"time"
)

// Alert 等级。
const (
	AlertWarn  = "warn"
	AlertAlarm = "alarm"
)

// Alert 状态。
const (
	AlertOpen       = "open"
	AlertAcknowledged = "acknowledged"
	AlertResolved   = "resolved"
)

// Alert 告警记录。
type Alert struct {
	ID          int64     `json:"id"`
	RuleID      int64     `json:"rule_id"`
	RoomID      int64     `json:"room_id"`
	PointID     int64     `json:"point_id"`
	Level       string    `json:"level"`
	Message     string    `json:"message"`
	Status      string    `json:"status"`
	OpenedAt    time.Time `json:"opened_at"`
	AckAt       *time.Time `json:"ack_at"`
	ResolvedAt  *time.Time `json:"resolved_at"`
}

// AlertInput 查询告警列表条件。
type AlertInput struct {
	RoomID   int64     `json:"room_id"`
	Level    string    `json:"level"`
	Status   string    `json:"status"`
	From     time.Time `json:"from"`
	To       time.Time `json:"to"`
	Limit    int       `json:"limit"`
}

// AlertLevels 合法告警等级。
var AlertLevels = map[string]bool{AlertWarn: true, AlertAlarm: true}

// AlertStatuses 合法告警状态。
var AlertStatuses = map[string]bool{AlertOpen: true, AlertAcknowledged: true, AlertResolved: true}

func (a *Alert) Validate() error {
	if !AlertLevels[a.Level] {
		return ErrInvalidLevel
	}
	if !AlertStatuses[a.Status] {
		return ErrInvalidAlertStatus
	}
	return nil
}

// IsOpen 是否未解决。
func (a *Alert) IsOpen() bool {
	return a.Status == AlertOpen || a.Status == AlertAcknowledged
}

var (
	ErrInvalidLevel       = errors.New("告警等级非法")
	ErrInvalidAlertStatus = errors.New("告警状态非法")
)

// Message 构造告警描述。
func (a *Alert) MessageText() string {
	parts := []string{"房间", roomName(a.RoomID)}
	if a.Level == AlertAlarm {
		parts = append(parts, "ALARM")
	} else {
		parts = append(parts, "WARN")
	}
	return strings.Join(parts, " ")
}

func roomName(_ int64) string { return "" }