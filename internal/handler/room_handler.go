package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"cleanroom-monitor/internal/model"
)

func (h *Handler) roomCreate(w http.ResponseWriter, r *http.Request) {
	var in model.RoomInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeErr(w, http.StatusBadRequest, model.ErrInvalidInput)
		return
	}
	room, err := h.services.Rooms.Create(r.Context(), &in)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusCreated, room)
}

func (h *Handler) roomList(w http.ResponseWriter, r *http.Request) {
	page := model.Page{Limit: 100, Offset: 0}
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			page.Limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			page.Offset = n
		}
	}
	if v := r.URL.Query().Get("cleanroom_id"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			list, err := h.services.Rooms.ListByCleanroom(r.Context(), n)
			if err != nil {
				writeErr(w, statusFromError(err), err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"items": list})
			return
		}
	}
	list, err := h.services.Rooms.List(r.Context(), page)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": list})
}

func (h *Handler) roomGet(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	room, err := h.services.Rooms.Get(r.Context(), id)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, room)
}

func (h *Handler) roomStatus(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	st, err := h.services.Rooms.GetStatus(r.Context(), id)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, st)
}