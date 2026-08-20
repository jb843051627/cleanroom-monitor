package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"cleanroom-monitor/internal/model"
)

func (h *Handler) batchStart(w http.ResponseWriter, r *http.Request) {
	var in model.BatchInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeErr(w, http.StatusBadRequest, model.ErrInvalidInput)
		return
	}
	b, err := h.services.Batches.Start(r.Context(), &in)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusCreated, b)
}

func (h *Handler) batchComplete(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	b, err := h.services.Batches.Complete(r.Context(), id)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (h *Handler) batchAbort(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	b, err := h.services.Batches.Abort(r.Context(), id)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (h *Handler) batchList(w http.ResponseWriter, r *http.Request) {
	page := model.Page{Limit: 100, Offset: 0}
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			page.Limit = n
		}
	}
	if v := r.URL.Query().Get("room_id"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			list, err := h.services.Batches.ListByRoom(r.Context(), n)
			if err != nil {
				writeErr(w, statusFromError(err), err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"items": list})
			return
		}
	}
	list, err := h.services.Batches.List(r.Context(), page)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": list})
}