package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type Store struct {
	DB *sql.DB
}

func New(databaseURL string) (*Store, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return &Store{DB: db}, nil
}

func (s *Store) Migrate(ctx context.Context) error {
	_, err := s.DB.ExecContext(ctx, migrateSQL)
	return err
}

func (s *Store) Close() error {
	return s.DB.Close()
}

func IsUniqueViolation(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505"
	}
	return false
}

const migrateSQL = `
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS users (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	email VARCHAR(255) UNIQUE NOT NULL,
	password_hash VARCHAR(255) NOT NULL DEFAULT '',
	role VARCHAR(10) NOT NULL CHECK (role IN ('admin', 'user')),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS rooms (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	name VARCHAR(255) NOT NULL,
	description TEXT,
	capacity INTEGER,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS schedules (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	room_id UUID NOT NULL UNIQUE REFERENCES rooms(id) ON DELETE CASCADE,
	days_of_week INTEGER[] NOT NULL,
	start_time VARCHAR(5) NOT NULL,
	end_time VARCHAR(5) NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS slots (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
	start_time TIMESTAMPTZ NOT NULL,
	end_time TIMESTAMPTZ NOT NULL,
	UNIQUE (room_id, start_time, end_time)
);

CREATE INDEX IF NOT EXISTS idx_slots_room_start ON slots (room_id, start_time);

CREATE TABLE IF NOT EXISTS bookings (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	slot_id UUID NOT NULL REFERENCES slots(id) ON DELETE CASCADE,
	user_id UUID NOT NULL,
	status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'cancelled')),
	conference_link TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_bookings_slot_status ON bookings (slot_id, status);
CREATE INDEX IF NOT EXISTS idx_bookings_user_id ON bookings (user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_bookings_slot_active ON bookings (slot_id) WHERE status = 'active';

INSERT INTO users (id, email, role) VALUES
	('00000000-0000-0000-0000-000000000001', 'admin@test.com', 'admin'),
	('00000000-0000-0000-0000-000000000002', 'user@test.com', 'user')
ON CONFLICT (id) DO NOTHING;
`
