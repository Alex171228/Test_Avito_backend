package service

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"room-booking/internal/model"
	"room-booking/internal/store"
)

type ScheduleService struct {
	store *store.Store
}

func NewScheduleService(s *store.Store) *ScheduleService {
	return &ScheduleService{store: s}
}

var timePattern = regexp.MustCompile(`^([01]?[0-9]|2[0-3]):[0-5][0-9]$`)

func (s *ScheduleService) Create(ctx context.Context, roomID uuid.UUID, daysOfWeek []int, startTime, endTime string) (*model.Schedule, error) {
	room, err := s.store.GetRoom(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, model.NewAppError(http.StatusNotFound, "ROOM_NOT_FOUND", "room not found")
	}

	existing, err := s.store.GetScheduleByRoomID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, model.NewAppError(http.StatusConflict, "SCHEDULE_EXISTS",
			"schedule for this room already exists and cannot be changed")
	}

	if len(daysOfWeek) == 0 {
		return nil, model.NewAppError(http.StatusBadRequest, "INVALID_REQUEST", "daysOfWeek must not be empty")
	}
	for _, d := range daysOfWeek {
		if d < 1 || d > 7 {
			return nil, model.NewAppError(http.StatusBadRequest, "INVALID_REQUEST",
				fmt.Sprintf("daysOfWeek values must be between 1 and 7, got %d", d))
		}
	}

	if !timePattern.MatchString(startTime) {
		return nil, model.NewAppError(http.StatusBadRequest, "INVALID_REQUEST", "invalid startTime format, expected HH:MM")
	}
	if !timePattern.MatchString(endTime) {
		return nil, model.NewAppError(http.StatusBadRequest, "INVALID_REQUEST", "invalid endTime format, expected HH:MM")
	}

	startMinutes := ToMinutes(startTime)
	endMinutes := ToMinutes(endTime)
	if startMinutes >= endMinutes {
		return nil, model.NewAppError(http.StatusBadRequest, "INVALID_REQUEST", "startTime must be before endTime")
	}
	if endMinutes-startMinutes < 30 {
		return nil, model.NewAppError(http.StatusBadRequest, "INVALID_REQUEST", "time range must be at least 30 minutes")
	}

	sch, err := s.store.CreateSchedule(ctx, roomID, daysOfWeek, startTime, endTime)
	if err != nil {
		if store.IsUniqueViolation(err) {
			return nil, model.NewAppError(http.StatusConflict, "SCHEDULE_EXISTS",
				"schedule for this room already exists and cannot be changed")
		}
		return nil, err
	}
	return sch, nil
}

func ToMinutes(t string) int {
	parts := strings.Split(t, ":")
	h, _ := strconv.Atoi(parts[0])
	m, _ := strconv.Atoi(parts[1])
	return h*60 + m
}
