.PHONY: build run migrate-up migrate-down test clean docker-up docker-down

# Build the application
build:
	go build -o banking-platform ./cmd/app

# Run the application
run:
	go run ./cmd/app

# Run database migrations (using goose)
migrate-up:
	goose -dir ./migration postgres "postgres://postgres:postgres@localhost:5433/banking?sslmode=disable" up

migrate-down:
	goose -dir ./migration postgres "postgres://postgres:postgres@localhost:5433/banking?sslmode=disable" down

migrate-status:
	goose -dir ./migration postgres "postgres://postgres:postgres@localhost:5433/banking?sslmode=disable" status

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -f banking-platform

# Docker commands
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f
