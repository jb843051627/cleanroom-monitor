package model

import (
	"errors"
	"time"
)

// 批次状态。
const (
	BatchPlanning  = "planning"
	BatchInProgress = "in_progress"
	BatchCompleted = "completed"
	BatchAborted   = "aborted"
)

// CleanBatch 生产批次：关联房间，驱动洁净状态机。
type CleanBatch struct {
	ID        int64     `json:"id"`
	RoomID    int64     `json:"room_id"`
	Name      string    `json:"name"`
	Product   string    `json:"product"`
	Phase     string    `json:"phase"`
	StartAt   time.Time `json:"start_at"`
	EndAt     *time.Time `json:"end_at"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// BatchInput 启动批次入参。
type BatchInput struct {
	RoomID  int64  `json:"room_id"`
	Name    string `json:"name"`
	Product string `json:"product"`
	Phase   string `json:"phase"`
}

// BatchStatuses 合法批次状态。
var BatchStatuses = map[string]bool{
	BatchPlanning: true, BatchInProgress: true, BatchCompleted: true, BatchAborted: true,
}

func (b *CleanBatch) Validate() error {
	if b.RoomID <= 0 {
		return ErrRoomRequired
	}
	if b.Name == "" {
		return ErrNameRequired
	}
	if b.Product == "" {
		return ErrProductRequired
	}
	if !BatchStatuses[b.Status] {
		return ErrInvalidBatchStatus
	}
	return nil
}

var (
	ErrProductRequired   = errors.New("产品必填")
	ErrInvalidBatchStatus = errors.New("批次状态非法")
)