package store

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"room-booking/internal/model"
)

func (s *Store) EnsureSlotsForDate(ctx context.Context, roomID uuid.UUID, date time.Time, schedule *model.Schedule) error {
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.Add(24 * time.Hour)

	var count int
	err := s.DB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM slots WHERE room_id = $1 AND start_time >= $2 AND start_time < $3`,
		roomID, dayStart, dayEnd,
	).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	startH, startM := ParseTimeStr(schedule.StartTime)
	endH, endM := ParseTimeStr(schedule.EndTime)
	start := time.Date(date.Year(), date.Month(), date.Day(), startH, startM, 0, 0, time.UTC)
	end := time.Date(date.Year(), date.Month(), date.Day(), endH, endM, 0, 0, time.UTC)

	var values []string
	var args []interface{}
	idx := 1
	for t := start; t.Add(30 * time.Minute).Before(end) || t.Add(30*time.Minute).Equal(end); t = t.Add(30 * time.Minute) {
		values = append(values, fmt.Sprintf("($%d, $%d, $%d, $%d)", idx, idx+1, idx+2, idx+3))
		args = append(args, uuid.New(), roomID, t, t.Add(30*time.Minute))
		idx += 4
	}

	if len(values) == 0 {
		return nil
	}

	query := fmt.Sprintf(
		"INSERT INTO slots (id, room_id, start_time, end_time) VALUES %s ON CONFLICT DO NOTHING",
		strings.Join(values, ", "),
	)
	_, err = s.DB.ExecContext(ctx, query, args...)
	return err
}

func (s *Store) ListAvailableSlots(ctx context.Context, roomID uuid.UUID, date time.Time) ([]model.Slot, error) {
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.Add(24 * time.Hour)

	rows, err := s.DB.QueryContext(ctx, `
		SELECT s.id, s.room_id, s.start_time, s.end_time
		FROM slots s
		WHERE s.room_id = $1 AND s.start_time >= $2 AND s.start_time < $3
		  AND NOT EXISTS (
			SELECT 1 FROM bookings b WHERE b.slot_id = s.id AND b.status = 'active'
		  )
		ORDER BY s.start_time
	`, roomID, dayStart, dayEnd)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	slots := make([]model.Slot, 0)
	for rows.Next() {
		var sl model.Slot
		if err := rows.Scan(&sl.ID, &sl.RoomID, &sl.Start, &sl.End); err != nil {
			return nil, err
		}
		slots = append(slots, sl)
	}
	return slots, rows.Err()
}

func (s *Store) GetSlotByID(ctx context.Context, id uuid.UUID) (*model.Slot, error) {
	var sl model.Slot
	err := s.DB.QueryRowContext(ctx,
		`SELECT id, room_id, start_time, end_time FROM slots WHERE id = $1`, id,
	).Scan(&sl.ID, &sl.RoomID, &sl.Start, &sl.End)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sl, nil
}

func ParseTimeStr(s string) (int, int) {
	parts := strings.Split(s, ":")
	h, _ := strconv.Atoi(parts[0])
	m, _ := strconv.Atoi(parts[1])
	return h, m
}
