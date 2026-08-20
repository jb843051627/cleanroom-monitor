package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"cleanroom-monitor/internal/model"
)

func (h *Handler) pointCreate(w http.ResponseWriter, r *http.Request) {
	var in model.PointInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeErr(w, http.StatusBadRequest, model.ErrInvalidInput)
		return
	}
	p, err := h.services.Points.Create(r.Context(), &in)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (h *Handler) pointList(w http.ResponseWriter, r *http.Request) {
	var roomID int64
	if v := r.URL.Query().Get("room_id"); v != "" {
		roomID, _ = strconv.ParseInt(v, 10, 64)
	}
	list, err := h.services.Points.List(r.Context(), roomID)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": list})
}

func (h *Handler) pointToggle(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	var in struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeErr(w, http.StatusBadRequest, model.ErrInvalidInput)
		return
	}
	p, err := h.services.Points.Toggle(r.Context(), id, in.Enabled)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}