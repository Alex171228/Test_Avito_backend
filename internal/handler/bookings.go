package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"room-booking/internal/middleware"
	"room-booking/internal/model"
)

type createBookingRequest struct {
	SlotID               string `json:"slotId"`
	CreateConferenceLink bool   `json:"createConferenceLink"`
}

func (h *Handler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	var req createBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	slotID, err := uuid.Parse(req.SlotID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid slotId")
		return
	}

	booking, err := h.bookings.Create(r.Context(), claims.UserID, slotID, req.CreateConferenceLink)
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{"booking": booking})
}

func (h *Handler) ListBookings(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	bookings, pagination, err := h.bookings.ListAll(r.Context(), page, pageSize)
	if err != nil {
		handleError(w, err)
		return
	}
	if bookings == nil {
		bookings = make([]model.Booking, 0)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"bookings":   bookings,
		"pagination": pagination,
	})
}

func (h *Handler) ListMyBookings(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	bookings, err := h.bookings.ListMy(r.Context(), claims.UserID)
	if err != nil {
		handleError(w, err)
		return
	}
	if bookings == nil {
		bookings = make([]model.Booking, 0)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"bookings": bookings})
}

func (h *Handler) CancelBooking(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	bookingIDStr := chi.URLParam(r, "bookingId")
	bookingID, err := uuid.Parse(bookingIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid bookingId")
		return
	}

	booking, err := h.bookings.Cancel(r.Context(), claims.UserID, bookingID)
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"booking": booking})
}
