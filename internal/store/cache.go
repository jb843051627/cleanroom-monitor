package store

import (
	"sync"
	"time"

	"cleanroom-monitor/internal/model"
)

// Cache 实时看板缓存：房间级最新读数快照与状态摘要。
// 读写并发安全（RWMutex 保护），由 service 层刷新，handler 层只读。
type Cache struct {
	mu        sync.RWMutex
	snapshot  map[int64]*RoomSnapshot // key = room_id
	updatedAt time.Time
}

// RoomSnapshot 房间实时快照。
type RoomSnapshot struct {
	RoomID   int64             `json:"room_id"`
	RoomCode string            `json:"room_code"`
	Status   string            `json:"status"`
	Realtime []model.RealtimeReading `json:"realtime"`
	OpenAlerts int             `json:"open_alerts"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// NewCache 创建缓存。
func NewCache() *Cache {
	return &Cache{snapshot: make(map[int64]*RoomSnapshot)}
}

// Get 读取房间快照。
func (c *Cache) Get(roomID int64) (*RoomSnapshot, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	s, ok := c.snapshot[roomID]
	return s, ok
}

// GetAll 读取全部快照。
func (c *Cache) GetAll() []*RoomSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]*RoomSnapshot, 0, len(c.snapshot))
	for _, s := range c.snapshot {
		out = append(out, s)
	}
	return out
}

// Set 写入房间快照。
func (c *Cache) Set(roomID int64, s *RoomSnapshot) {
	c.mu.Lock()
	defer c.mu.Unlock()
	s.UpdatedAt = time.Now()
	c.snapshot[roomID] = s
	c.updatedAt = time.Now()
}

// Remove 移除房间快照。
func (c *Cache) Remove(roomID int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.snapshot, roomID)
}

// UpdatedAt 缓存最近刷新时间。
func (c *Cache) UpdatedAt() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.updatedAt
}

// RoomID 由 RoomSnapshot 构造（供排序去重）。
func (s *RoomSnapshot) RoomIDKey() int64 { return s.RoomID }

// SnapshotRoomIDs 返回全部房间 ID。
func (c *Cache) SnapshotRoomIDs() []int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]int64, 0, len(c.snapshot))
	for id := range c.snapshot {
		out = append(out, id)
	}
	return out
}

// Clear 清空缓存。
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.snapshot = make(map[int64]*RoomSnapshot)
}