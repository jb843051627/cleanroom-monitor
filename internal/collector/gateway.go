package collector

import (
	"encoding/json"
	"log"
	"net/http"

	"cleanroom-monitor/internal/model"
)

// Gateway 上报网关：HTTP 入口，接收批量读数并投递到采集器。
type Gateway struct {
	collect *Collector
}

// NewGateway 创建网关。
func NewGateway(collect *Collector) *Gateway {
	return &Gateway{collect: collect}
}

// HandleIngest POST /api/v1/readings。
func (g *Gateway) HandleIngest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var batch model.ReadingBatch
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&batch); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if len(batch.Readings) == 0 {
		http.Error(w, "empty readings", http.StatusBadRequest)
		return
	}
	g.collect.Enqueue(&batch)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"accepted": len(batch.Readings),
		"queued":   true,
	})
}

// HandleHealth GET /healthz。
func (g *Gateway) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": "ok",
		"queue":  g.collect.QueueLen(),
	})
}

// Drain 排空队列（优雅停机用）。
func (g *Gateway) Drain() {
	log.Printf("[gateway] 排空完成")
}