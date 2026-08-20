package model

import (
	"errors"
	"time"
)

// 房间洁净状态机状态。
const (
	StateAtRest    = "at_rest"   // 静态（无批次）
	StateNormal    = "normal"    // 正常生产
	StateAlert     = "alert"     // 有 warn 告警
	StateAlarm     = "alarm"     // 有 alarm 告警
	StateRestricted = "restricted" // 受限（读数停滞或长时间超标）
	StateRelease   = "release"   // 批次放行
)

// CleanStatus 房间洁净状态流转记录。
type CleanStatus struct {
	ID        int64     `json:"id"`
	RoomID    int64     `json:"room_id"`
	BatchID   int64     `json:"batch_id"`
	State     string    `json:"state"`
	Reason    string    `json:"reason"`
	ChangedAt time.Time `json:"changed_at"`
}

// StateTransition 状态转移。
type StateTransition struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Reason string `json:"reason"`
}

// ValidTransitions 允许的状态转移映射（状态机核心）。
var ValidTransitions = map[string]map[string]bool{
	StateAtRest: {
		StateNormal: true,
		StateAlert:  true,
		StateAlarm:  true,
	},
	StateNormal: {
		StateAlert:   true,
		StateAlarm:   true,
		StateRestricted: true,
		StateRelease: true,
		StateAtRest:  true,
	},
	StateAlert: {
		StateNormal: true,
		StateAlarm:  true,
		StateRestricted: true,
		StateRelease: true,
	},
	StateAlarm: {
		StateRestricted: true,
		StateNormal:     true,
		StateAlert:      true,
	},
	StateRestricted: {
		StateNormal: true,
	},
	StateRelease: {
		StateNormal: true,
		StateAtRest:  true,
	},
}

// CanTransition 判断状态转移是否合法。
func CanTransition(from, to string) bool {
	allowed, ok := ValidTransitions[from]
	if !ok {
		return false
	}
	return allowed[to]
}

// States 所有合法状态。
var States = map[string]bool{
	StateAtRest: true, StateNormal: true, StateAlert: true,
	StateAlarm: true, StateRestricted: true, StateRelease: true,
}

func (s *CleanStatus) Validate() error {
	if s.RoomID <= 0 {
		return ErrRoomRequired
	}
	if !States[s.State] {
		return ErrInvalidState
	}
	return nil
}

var (
	ErrInvalidState = errors.New("洁净状态非法")
)

// Reason 常量（报表展示用）。
const (
	ReasonBatchStarted     = "批次开始"
	ReasonBatchCompleted   = "批次完成"
	ReasonBatchAborted     = "批次中止"
	ReasonWarnAlert        = "出现 warn 告警"
	ReasonAlarmAlert       = "出现 alarm 告警"
	ReasonNoReadings       = "长时间无读数"
	ReasonExceedThreshold  = "持续超阈值"
	ReasonAllClear         = "告警全部解除"
	ReasonBatchRelease     = "批次放行"
)