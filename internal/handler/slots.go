package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"room-booking/internal/model"
)

func (h *Handler) ListSlots(w http.ResponseWriter, r *http.Request) {
	roomIDStr := chi.URLParam(r, "roomId")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid roomId")
		return
	}

	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "date parameter is required")
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid date format, expected YYYY-MM-DD")
		return
	}

	slots, err := h.slots.ListAvailable(r.Context(), roomID, date)
	if err != nil {
		handleError(w, err)
		return
	}
	if slots == nil {
		slots = make([]model.Slot, 0)
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"slots": slots})
}
