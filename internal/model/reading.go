package model

import "time"

// Reading 单条环境读数。
type Reading struct {
	ID         int64     `json:"id"`
	PointID    int64     `json:"point_id"`
	SensorID   int64     `json:"sensor_id"`
	ParamType  string    `json:"param_type"`
	Value      float64   `json:"value"`
	MeasuredAt time.Time `json:"measured_at"`
	Raw        string    `json:"raw"`
	CreatedAt  time.Time `json:"created_at"`
}

// ReadingBatch 批量上报的读数（网关/模拟器）。
type ReadingBatch struct {
	Readings []ReadingInput `json:"readings"`
}

// ReadingInput 单条读数入参。
type ReadingInput struct {
	PointID    int64   `json:"point_id"`
	SensorID   int64   `json:"sensor_id"`
	ParamType  string  `json:"param_type"`
	Value      float64 `json:"value"`
	MeasuredAt string  `json:"measured_at"` // RFC3339
	Raw        string  `json:"raw"`
}

// ReadingStats 按参数聚合统计结果。
type ReadingStats struct {
	ParamType  string  `json:"param_type"`
	Count      int     `json:"count"`
	Min        float64 `json:"min"`
	Max        float64 `json:"max"`
	Avg        float64 `json:"avg"`
	OutOfRange int     `json:"out_of_range"` // 超阈值条数
	Unit       string  `json:"unit"`
}

// RealtimeReading 看板实时读数（合并点位与房间信息）。
type RealtimeReading struct {
	PointID   int64     `json:"point_id"`
	RoomID    int64     `json:"room_id"`
	RoomCode  string    `json:"room_code"`
	ParamType string    `json:"param_type"`
	Value     float64   `json:"value"`
	MeasuredAt time.Time `json:"measured_at"`
	WithinRange bool    `json:"within_range"`
}