package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"room-booking/internal/model"
	"room-booking/internal/store"
)

type BookingService struct {
	store *store.Store
}

func NewBookingService(s *store.Store) *BookingService {
	return &BookingService{store: s}
}

func (s *BookingService) Create(ctx context.Context, userID, slotID uuid.UUID, createConferenceLink bool) (*model.Booking, error) {
	slot, err := s.store.GetSlotByID(ctx, slotID)
	if err != nil {
		return nil, err
	}
	if slot == nil {
		return nil, model.NewAppError(http.StatusNotFound, "SLOT_NOT_FOUND", "slot not found")
	}

	if slot.Start.Before(time.Now().UTC()) {
		return nil, model.NewAppError(http.StatusBadRequest, "INVALID_REQUEST", "cannot book a slot in the past")
	}

	var conferenceLink *string
	if createConferenceLink {
		link := fmt.Sprintf("https://conference.example.com/%s", uuid.New().String())
		conferenceLink = &link
	}

	booking, err := s.store.CreateBooking(ctx, slotID, userID, conferenceLink)
	if err != nil {
		if store.IsUniqueViolation(err) {
			return nil, model.NewAppError(http.StatusConflict, "SLOT_ALREADY_BOOKED", "slot is already booked")
		}
		return nil, err
	}
	return booking, nil
}

func (s *BookingService) ListAll(ctx context.Context, page, pageSize int) ([]model.Booking, *model.Pagination, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize
	bookings, total, err := s.store.ListBookings(ctx, pageSize, offset)
	if err != nil {
		return nil, nil, err
	}

	pagination := &model.Pagination{
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}
	return bookings, pagination, nil
}

func (s *BookingService) ListMy(ctx context.Context, userID uuid.UUID) ([]model.Booking, error) {
	return s.store.ListMyBookings(ctx, userID)
}

func (s *BookingService) Cancel(ctx context.Context, userID, bookingID uuid.UUID) (*model.Booking, error) {
	booking, err := s.store.GetBooking(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if booking == nil {
		return nil, model.NewAppError(http.StatusNotFound, "BOOKING_NOT_FOUND", "booking not found")
	}

	if booking.UserID != userID {
		return nil, model.NewAppError(http.StatusForbidden, "FORBIDDEN", "cannot cancel another user's booking")
	}

	if booking.Status == "cancelled" {
		return booking, nil
	}

	return s.store.CancelBooking(ctx, bookingID)
}
