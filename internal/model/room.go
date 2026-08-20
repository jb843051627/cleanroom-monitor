package model

import (
	"errors"
	"time"
)

// Room 房间：洁净区下的物理房间，绑定环境目标参数。
type Room struct {
	ID                int64     `json:"id"`
	CleanroomID       int64     `json:"cleanroom_id"`
	Name              string    `json:"name"`
	Code              string    `json:"code"`
	Kind              string    `json:"kind"` // classify / atelier / locker
	AreaSqm           float64   `json:"area_sqm"`
	TargetPressure    float64   `json:"target_pressure"` // 目标压差 Pa
	PressureTolerance float64   `json:"pressure_tolerance"`
	TargetTemp        float64   `json:"target_temp"` // 目标温度 ℃
	TargetHumidity    float64   `json:"target_humidity"`
	Status            string    `json:"status"` // 当前洁净状态
	CreatedAt         time.Time `json:"created_at"`
}

// RoomInput 创建/更新房间入参。
type RoomInput struct {
	CleanroomID       int64   `json:"cleanroom_id"`
	Name              string  `json:"name"`
	Code              string  `json:"code"`
	Kind              string  `json:"kind"`
	AreaSqm           float64 `json:"area_sqm"`
	TargetPressure    float64 `json:"target_pressure"`
	PressureTolerance float64 `json:"pressure_tolerance"`
	TargetTemp        float64 `json:"target_temp"`
	TargetHumidity    float64 `json:"target_humidity"`
}

func (r *Room) Validate() error {
	if r.CleanroomID <= 0 {
		return ErrCleanroomRequired
	}
	if r.Name == "" {
		return ErrNameRequired
	}
	if r.Code == "" {
		return ErrCodeRequired
	}
	switch r.Kind {
	case "classify", "atelier", "locker":
	default:
		return ErrInvalidKind
	}
	if r.TargetPressure <= 0 || r.PressureTolerance <= 0 {
		return ErrInvalidPressureTarget
	}
	return nil
}

// RoomKindNames 房间类型中文名（用于报表展示）。
func RoomKindNames() map[string]string {
	return map[string]string{
		"classify": "分类间",
		"atelier":  "操作间",
		"locker":   "更衣室",
	}
}

// NewRoomStatus 默认房间初始状态。
const NewRoomStatus = "at_rest"

// sentinel errors
var (
	ErrCleanroomRequired = errors.New("cleanroom 必填")
	ErrInvalidKind       = errors.New("房间类型非法")
	ErrInvalidPressureTarget = errors.New("压差目标与容差必须大于 0")
)