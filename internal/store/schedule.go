package store

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"room-booking/internal/model"
)

func (s *Store) CreateSchedule(ctx context.Context, roomID uuid.UUID, daysOfWeek []int, startTime, endTime string) (*model.Schedule, error) {
	sch := &model.Schedule{
		ID:         uuid.New(),
		RoomID:     roomID,
		DaysOfWeek: daysOfWeek,
		StartTime:  startTime,
		EndTime:    endTime,
	}
	_, err := s.DB.ExecContext(ctx,
		`INSERT INTO schedules (id, room_id, days_of_week, start_time, end_time) VALUES ($1, $2, $3, $4, $5)`,
		sch.ID, sch.RoomID, pq.Array(sch.DaysOfWeek), sch.StartTime, sch.EndTime,
	)
	if err != nil {
		return nil, err
	}
	return sch, nil
}

func (s *Store) GetScheduleByRoomID(ctx context.Context, roomID uuid.UUID) (*model.Schedule, error) {
	var sch model.Schedule
	var days pq.Int64Array
	err := s.DB.QueryRowContext(ctx,
		`SELECT id, room_id, days_of_week, start_time, end_time FROM schedules WHERE room_id = $1`, roomID,
	).Scan(&sch.ID, &sch.RoomID, &days, &sch.StartTime, &sch.EndTime)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	sch.DaysOfWeek = make([]int, len(days))
	for i, d := range days {
		sch.DaysOfWeek[i] = int(d)
	}
	return &sch, nil
}
