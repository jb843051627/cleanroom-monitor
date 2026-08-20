package model

import "time"

// Cleanroom 洁净区：一个洁净区可包含多个房间。
type Cleanroom struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Grade     string    `json:"grade"` // ISO5 / ISO7 / ISO8
	AreaSqm   float64   `json:"area_sqm"`
	Status    string    `json:"status"` // 见 CleanStatus.State
	CreatedAt time.Time `json:"created_at"`
}

// CleanroomInput 创建/更新洁净区入参。
type CleanroomInput struct {
	Name    string  `json:"name"`
	Code    string  `json:"code"`
	Grade   string  `json:"grade"`
	AreaSqm float64 `json:"area_sqm"`
}

func (c *Cleanroom) Validate() error {
	if c.Name == "" {
		return ErrNameRequired
	}
	if c.Code == "" {
		return ErrCodeRequired
	}
	switch c.Grade {
	case "ISO5", "ISO7", "ISO8":
	default:
		return ErrInvalidGrade
	}
	if c.AreaSqm <= 0 {
		return ErrInvalidArea
	}
	return nil
}