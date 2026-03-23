package store

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"room-booking/internal/model"
)

func (s *Store) CreateBooking(ctx context.Context, slotID, userID uuid.UUID, conferenceLink *string) (*model.Booking, error) {
	booking := &model.Booking{
		ID:             uuid.New(),
		SlotID:         slotID,
		UserID:         userID,
		Status:         "active",
		ConferenceLink: conferenceLink,
	}
	err := s.DB.QueryRowContext(ctx,
		`INSERT INTO bookings (id, slot_id, user_id, status, conference_link)
		 VALUES ($1, $2, $3, $4, $5) RETURNING created_at`,
		booking.ID, booking.SlotID, booking.UserID, booking.Status, booking.ConferenceLink,
	).Scan(&booking.CreatedAt)
	if err != nil {
		return nil, err
	}
	return booking, nil
}

func (s *Store) ListBookings(ctx context.Context, limit, offset int) ([]model.Booking, int, error) {
	var total int
	err := s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM bookings`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.DB.QueryContext(ctx,
		`SELECT id, slot_id, user_id, status, conference_link, created_at
		 FROM bookings ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	bookings := make([]model.Booking, 0)
	for rows.Next() {
		var b model.Booking
		if err := rows.Scan(&b.ID, &b.SlotID, &b.UserID, &b.Status, &b.ConferenceLink, &b.CreatedAt); err != nil {
			return nil, 0, err
		}
		bookings = append(bookings, b)
	}
	return bookings, total, rows.Err()
}

func (s *Store) ListMyBookings(ctx context.Context, userID uuid.UUID) ([]model.Booking, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT b.id, b.slot_id, b.user_id, b.status, b.conference_link, b.created_at
		FROM bookings b
		JOIN slots s ON s.id = b.slot_id
		WHERE b.user_id = $1 AND s.start_time >= NOW()
		ORDER BY s.start_time
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bookings := make([]model.Booking, 0)
	for rows.Next() {
		var b model.Booking
		if err := rows.Scan(&b.ID, &b.SlotID, &b.UserID, &b.Status, &b.ConferenceLink, &b.CreatedAt); err != nil {
			return nil, err
		}
		bookings = append(bookings, b)
	}
	return bookings, rows.Err()
}

func (s *Store) GetBooking(ctx context.Context, id uuid.UUID) (*model.Booking, error) {
	var b model.Booking
	err := s.DB.QueryRowContext(ctx,
		`SELECT id, slot_id, user_id, status, conference_link, created_at FROM bookings WHERE id = $1`, id,
	).Scan(&b.ID, &b.SlotID, &b.UserID, &b.Status, &b.ConferenceLink, &b.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (s *Store) CancelBooking(ctx context.Context, id uuid.UUID) (*model.Booking, error) {
	var b model.Booking
	err := s.DB.QueryRowContext(ctx,
		`UPDATE bookings SET status = 'cancelled' WHERE id = $1
		 RETURNING id, slot_id, user_id, status, conference_link, created_at`, id,
	).Scan(&b.ID, &b.SlotID, &b.UserID, &b.Status, &b.ConferenceLink, &b.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &b, nil
}
