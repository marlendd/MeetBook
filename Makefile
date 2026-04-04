.PHONY: up down seed test test-e2e lint swagger

swagger:
	swag init -g cmd/api/main.go -o docs

up:
	docker-compose up --build -d

down:
	docker-compose down

seed:
	docker-compose exec -T db psql -U postgres -d booking -c "\
		INSERT INTO rooms (id, name, description, capacity) VALUES \
		  ('aaaaaaaa-0000-0000-0000-000000000001', 'Малая', 'Переговорка на 4 человека', 4), \
		  ('aaaaaaaa-0000-0000-0000-000000000002', 'Большая', 'Переговорка на 12 человек', 12) \
		ON CONFLICT (id) DO NOTHING; \
		INSERT INTO schedules (id, room_id, days_of_week, start_time, end_time) VALUES \
		  ('bbbbbbbb-0000-0000-0000-000000000001', 'aaaaaaaa-0000-0000-0000-000000000001', '{1,2,3,4,5}', '09:00', '18:00'), \
		  ('bbbbbbbb-0000-0000-0000-000000000002', 'aaaaaaaa-0000-0000-0000-000000000002', '{1,2,3,4,5,6,7}', '08:00', '20:00') \
		ON CONFLICT (room_id) DO NOTHING;"

test:
	go test ./tests/unit/... -coverprofile=coverage.out -coverpkg=./internal/...
	go tool cover -func=coverage.out | tail -1

test-e2e:
	go test ./tests/e2e/... -v -timeout 120s

lint:
	golangci-lint run ./...
