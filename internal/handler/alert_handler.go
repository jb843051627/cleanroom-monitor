package handler

import (
	"net/http"
	"strconv"

	"cleanroom-monitor/internal/model"
)

func (h *Handler) alertList(w http.ResponseWriter, r *http.Request) {
	in := model.AlertInput{Limit: 100}
	if v := r.URL.Query().Get("room_id"); v != "" {
		in.RoomID, _ = strconv.ParseInt(v, 10, 64)
	}
	in.Level = r.URL.Query().Get("level")
	in.Status = r.URL.Query().Get("status")
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			in.Limit = n
		}
	}
	list, err := h.services.Alerts.List(r.Context(), in)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": list})
}

func (h *Handler) alertAck(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	a, err := h.services.Alerts.Acknowledge(r.Context(), id)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (h *Handler) alertResolve(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	a, err := h.services.Alerts.Resolve(r.Context(), id)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, a)
}