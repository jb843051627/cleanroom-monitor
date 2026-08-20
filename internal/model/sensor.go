package model

import (
	"errors"
	"time"
)

// Sensor 传感器状态。
const (
	SensorActive  = "active"
	SensorFault   = "fault"
	SensorOffline = "offline"
)

// Sensor 物理传感器，绑定监测点。
type Sensor struct {
	ID               int64     `json:"id"`
	PointID          int64     `json:"point_id"`
	Serial           string    `json:"serial"`
	Vendor           string    `json:"vendor"`
	LastSeenAt       time.Time `json:"last_seen_at"`
	Battery          float64   `json:"battery"` // 剩余电量 %
	Status           string    `json:"status"`
	CalibrationDueAt time.Time `json:"calibration_due_at"`
	CreatedAt        time.Time `json:"created_at"`
}

// SensorInput 注册传感器入参。
type SensorInput struct {
	PointID  int64  `json:"point_id"`
	Serial   string `json:"serial"`
	Vendor   string `json:"vendor"`
	Battery  float64 `json:"battery"`
	CalibrationDueAt time.Time `json:"calibration_due_at"`
}

func (s *Sensor) Validate() error {
	if s.PointID <= 0 {
		return ErrPointRequired
	}
	if s.Serial == "" {
		return ErrSerialRequired
	}
	if s.Vendor == "" {
		return ErrVendorRequired
	}
	if s.Battery < 0 || s.Battery > 100 {
		return ErrInvalidBattery
	}
	return nil
}

// IsDueForCalibration 校准是否已到期。
func (s *Sensor) IsDueForCalibration(now time.Time) bool {
	return !s.CalibrationDueAt.IsZero() && !now.Before(s.CalibrationDueAt)
}

// MarkStatusBySeen 根据最近上报时间刷新传感器在线状态。
func (s *Sensor) MarkStatusBySeen(now time.Time, offlineAfter time.Duration) {
	if s.LastSeenAt.IsZero() || now.Sub(s.LastSeenAt) > offlineAfter {
		s.Status = SensorOffline
		return
	}
	if s.Status == SensorOffline {
		s.Status = SensorActive
	}
}

var (
	ErrPointRequired    = errors.New("point 必填")
	ErrSerialRequired   = errors.New("序列号必填")
	ErrVendorRequired   = errors.New("厂商必填")
	ErrInvalidBattery   = errors.New("电量必须在 0-100")
)