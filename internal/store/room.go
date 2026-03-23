package store

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"room-booking/internal/model"
)

func (s *Store) CreateRoom(ctx context.Context, name string, description *string, capacity *int) (*model.Room, error) {
	room := &model.Room{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Capacity:    capacity,
	}
	err := s.DB.QueryRowContext(ctx,
		`INSERT INTO rooms (id, name, description, capacity) VALUES ($1, $2, $3, $4) RETURNING created_at`,
		room.ID, room.Name, room.Description, room.Capacity,
	).Scan(&room.CreatedAt)
	if err != nil {
		return nil, err
	}
	return room, nil
}

func (s *Store) ListRooms(ctx context.Context) ([]model.Room, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT id, name, description, capacity, created_at FROM rooms ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rooms := make([]model.Room, 0)
	for rows.Next() {
		var r model.Room
		if err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.Capacity, &r.CreatedAt); err != nil {
			return nil, err
		}
		rooms = append(rooms, r)
	}
	return rooms, rows.Err()
}

func (s *Store) GetRoom(ctx context.Context, id uuid.UUID) (*model.Room, error) {
	var r model.Room
	err := s.DB.QueryRowContext(ctx,
		`SELECT id, name, description, capacity, created_at FROM rooms WHERE id = $1`, id,
	).Scan(&r.ID, &r.Name, &r.Description, &r.Capacity, &r.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}
