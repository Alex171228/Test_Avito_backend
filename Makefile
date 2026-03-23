.PHONY: up down build seed test lint

up:
	docker-compose up --build -d

down:
	docker-compose down -v

build:
	go build -o bin/server ./cmd/server

seed:
	@echo "Seeding database with test data..."
	docker-compose exec -T postgres psql -U postgres -d room_booking -c " \
		INSERT INTO rooms (id, name, description, capacity) VALUES \
			('a0000000-0000-0000-0000-000000000001', 'Meeting Room A', 'Large meeting room', 10), \
			('a0000000-0000-0000-0000-000000000002', 'Meeting Room B', 'Small meeting room', 4), \
			('a0000000-0000-0000-0000-000000000003', 'Conference Hall', 'Main conference hall', 50) \
		ON CONFLICT DO NOTHING; \
		INSERT INTO schedules (id, room_id, days_of_week, start_time, end_time) VALUES \
			('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001', '{1,2,3,4,5}', '09:00', '18:00'), \
			('b0000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000002', '{1,2,3,4,5}', '10:00', '17:00'), \
			('b0000000-0000-0000-0000-000000000003', 'a0000000-0000-0000-0000-000000000003', '{1,2,3,4,5,6}', '08:00', '20:00') \
		ON CONFLICT DO NOTHING;"

test:
	go test ./... -v -cover

lint:
	golangci-lint run
