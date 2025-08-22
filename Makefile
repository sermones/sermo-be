.PHONY: help build run stop clean logs test

# Default target
help:
	@echo "Available commands:"
	@echo "  build     - Build Docker images"
	@echo "  run       - Start all services"
	@echo "  run-db    - Start only PostgreSQL database"
	@echo "  stop      - Stop all services"
	@echo "  stop-db   - Stop only PostgreSQL database"
	@echo "  clean     - Stop and remove containers, networks, volumes"
	@echo "  logs      - Show logs from all services"
	@echo "  logs-db   - Show logs from PostgreSQL only"
	@echo "  test      - Run tests"
	@echo "  dev       - Run in development mode (go run)"
	@echo "  swagger   - Generate Swagger documentation"

# Build Docker images
build:
	docker-compose build

# Start all services
run:
	docker-compose up -d

# Start only PostgreSQL database
run-db:
	docker-compose up -d postgres

# Stop all services
stop:
	docker-compose down

# Stop only PostgreSQL database
stop-db:
	docker-compose stop postgres

# Clean everything (containers, networks, volumes)
clean:
	docker-compose down -v --remove-orphans
	docker system prune -f

# Show logs from all services
logs:
	docker-compose logs -f

# Show logs from PostgreSQL only
logs-db:
	docker-compose logs -f postgres

# Run tests
test:
	go test ./...

# Development mode
dev:
	go run cmd/server/main.go

# Build and run
build-run: build run

# Generate Swagger documentation
swagger:
	/Users/chungjung-mac-m4/.asdf/installs/golang/1.25.0/bin/swag init -g cmd/server/main.go -o docs
	@echo "✅ Swagger documentation generated successfully!"
	@echo "📖 View docs at: docs/index.html"

# Restart services
restart: stop run
