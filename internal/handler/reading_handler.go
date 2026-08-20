package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"cleanroom-monitor/internal/model"
	"cleanroom-monitor/internal/store"
	"cleanroom-monitor/internal/util"
)

// gatewayIngest 转发到采集网关（由 router 挂载 gateway 处理）。
func (h *Handler) gatewayIngest(w http.ResponseWriter, r *http.Request) {
	var batch model.ReadingBatch
	if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
		writeErr(w, http.StatusBadRequest, model.ErrInvalidInput)
		return
	}
	n, err := h.services.Readings.Ingest(r.Context(), &batch)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"accepted": n})
}

func (h *Handler) readingsQuery(w http.ResponseWriter, r *http.Request) {
	q := store.ReadingQuery{Limit: 200}
	if v := r.URL.Query().Get("point_id"); v != "" {
		q.PointID, _ = strconv.ParseInt(v, 10, 64)
	}
	if v := r.URL.Query().Get("room_id"); v != "" {
		q.RoomID, _ = strconv.ParseInt(v, 10, 64)
	}
	q.ParamType = r.URL.Query().Get("param_type")
	if v := r.URL.Query().Get("from"); v != "" {
		q.From, _ = util.ParseTime(v)
	}
	if v := r.URL.Query().Get("to"); v != "" {
		q.To, _ = util.ParseTime(v)
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			q.Limit = n
		}
	}
	list, err := h.services.Readings.Query(r.Context(), q)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": list})
}

func (h *Handler) readingsStats(w http.ResponseWriter, r *http.Request) {
	paramType := r.URL.Query().Get("param_type")
	if !model.ParamTypes[paramType] {
		writeErr(w, http.StatusBadRequest, model.ErrInvalidParamType)
		return
	}
	from, to := parseRange(r)
	st, err := h.services.Readings.Stats(r.Context(), paramType, from, to)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, st)
}

func (h *Handler) readingsRealtime(w http.ResponseWriter, r *http.Request) {
	var roomID int64
	if v := r.URL.Query().Get("room_id"); v != "" {
		roomID, _ = strconv.ParseInt(v, 10, 64)
	}
	snap, err := h.services.Readings.RealtimeSnapshot(r.Context(), roomID)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, snap)
}

// parseRange 解析 from/to 时间范围（默认最近 24h）。
func parseRange(r *http.Request) (time.Time, time.Time) {
	now := time.Now()
	from := now.Add(-24 * time.Hour)
	to := now
	if v := r.URL.Query().Get("from"); v != "" {
		if t, err := util.ParseTime(v); err == nil && !t.IsZero() {
			from = t
		}
	}
	if v := r.URL.Query().Get("to"); v != "" {
		if t, err := util.ParseTime(v); err == nil && !t.IsZero() {
			to = t
		}
	}
	return from, to
}