package service

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"room-booking/internal/model"
	"room-booking/internal/store"
)

type SlotService struct {
	store *store.Store
}

func NewSlotService(s *store.Store) *SlotService {
	return &SlotService{store: s}
}

func (s *SlotService) ListAvailable(ctx context.Context, roomID uuid.UUID, date time.Time) ([]model.Slot, error) {
	room, err := s.store.GetRoom(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, model.NewAppError(http.StatusNotFound, "ROOM_NOT_FOUND", "room not found")
	}

	schedule, err := s.store.GetScheduleByRoomID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if schedule == nil {
		return make([]model.Slot, 0), nil
	}

	goWeekday := int(date.Weekday()) // Sunday=0, Monday=1, ..., Saturday=6
	isoWeekday := goWeekday
	if goWeekday == 0 {
		isoWeekday = 7 // Sunday → 7 per ISO
	}

	found := false
	for _, d := range schedule.DaysOfWeek {
		if d == isoWeekday {
			found = true
			break
		}
	}
	if !found {
		return make([]model.Slot, 0), nil
	}

	if err := s.store.EnsureSlotsForDate(ctx, roomID, date, schedule); err != nil {
		return nil, err
	}

	return s.store.ListAvailableSlots(ctx, roomID, date)
}
