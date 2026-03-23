package service

import (
	"context"
	"net/http"

	"room-booking/internal/model"
	"room-booking/internal/store"
)

type RoomService struct {
	store *store.Store
}

func NewRoomService(s *store.Store) *RoomService {
	return &RoomService{store: s}
}

func (s *RoomService) Create(ctx context.Context, name string, description *string, capacity *int) (*model.Room, error) {
	if name == "" {
		return nil, model.NewAppError(http.StatusBadRequest, "INVALID_REQUEST", "name is required")
	}
	return s.store.CreateRoom(ctx, name, description, capacity)
}

func (s *RoomService) List(ctx context.Context) ([]model.Room, error) {
	return s.store.ListRooms(ctx)
}
