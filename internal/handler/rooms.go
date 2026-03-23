package handler

import (
	"encoding/json"
	"net/http"

	"room-booking/internal/model"
)

type createRoomRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Capacity    *int    `json:"capacity"`
}

func (h *Handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req createRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	room, err := h.rooms.Create(r.Context(), req.Name, req.Description, req.Capacity)
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{"room": room})
}

func (h *Handler) ListRooms(w http.ResponseWriter, r *http.Request) {
	rooms, err := h.rooms.List(r.Context())
	if err != nil {
		handleError(w, err)
		return
	}
	if rooms == nil {
		rooms = make([]model.Room, 0)
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"rooms": rooms})
}
