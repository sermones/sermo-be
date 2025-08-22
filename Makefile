.PHONY: help build run stop clean logs test

# Default target
help:
	@echo "Available commands:"
	@echo "  build     - Build Docker images"
	@echo "  run       - Start all services (clean rebuild with --no-cache)"
	@echo "  run-db    - Start only PostgreSQL database"
	@echo "  stop      - Stop all services"
	@echo "  stop-db   - Stop only PostgreSQL database"
	@echo "  clean     - Stop and remove containers, networks, volumes"
	@echo "  clean-all - Stop and remove containers, networks, and images (preserve DB volumes)"
	@echo "  logs      - Show logs from all services"
	@echo "  logs-db   - Show logs from PostgreSQL only"
	@echo "  test      - Run tests"
	@echo "  dev       - Run in development mode (go run)"
	@echo "  swagger   - Generate Swagger documentation"

# Build Docker images
build: copy-docs
	@echo "🔑 SSH 에이전트를 확인하고 설정합니다..."
	@if [ -z "$$SSH_AUTH_SOCK" ]; then \
		echo "SSH 에이전트를 시작합니다..."; \
		eval $$(ssh-agent -s); \
		ssh-add ~/.ssh/id_rsa 2>/dev/null || echo "⚠️  SSH 키를 추가할 수 없습니다. ssh-add를 수동으로 실행하세요."; \
	fi
	@echo "✅ SSH 에이전트가 설정되었습니다."
	docker-compose build

# Start all services (clean rebuild)
run: clean-all copy-docs
	@echo "🔑 SSH 에이전트를 확인하고 설정합니다..."
	@if [ -z "$$SSH_AUTH_SOCK" ]; then \
		echo "SSH 에이전트를 시작합니다..."; \
		eval $$(ssh-agent -s); \
		ssh-add ~/.ssh/id_rsa 2>/dev/null || echo "⚠️  SSH 키를 추가할 수 없습니다. ssh-add를 수동으로 실행하세요."; \
	fi
	@echo "✅ SSH 에이전트가 설정되었습니다."
	@echo "🏗️  Docker 이미지를 새로 빌드합니다..."
	docker-compose build --no-cache
	@echo "🚀 서비스를 시작합니다..."
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

# Clean everything including images (preserve DB volumes)
clean-all:
	@echo "🧹 모든 컨테이너, 네트워크를 정리합니다..."
	docker-compose down --remove-orphans
	@echo "🗑️  Docker 이미지를 삭제합니다..."
	docker-compose down --rmi all
	@echo "🧽 Docker 시스템을 정리합니다..."
	docker system prune -af
	@echo "💾 데이터베이스 볼륨은 보존됩니다."
	@echo "✅ 정리가 완료되었습니다."

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

# Database migrations
migrate:
	@echo "🗄️  데이터베이스 마이그레이션을 실행합니다..."
	@go run cmd/server/main.go --migrate-only

# Setup SSH key for Docker build
setup-ssh:
	@echo "🔑 SSH 에이전트를 설정합니다..."
	@if [ -z "$$SSH_AUTH_SOCK" ]; then \
		echo "SSH 에이전트를 시작합니다..."; \
		eval $$(ssh-agent -s); \
	fi
	@ssh-add ~/.ssh/id_rsa 2>/dev/null || echo "⚠️  SSH 키를 추가할 수 없습니다. ssh-add를 수동으로 실행하세요."
	@echo "✅ SSH 에이전트가 설정되었습니다."
	@echo "이제 'make build' 또는 'make run'을 실행할 수 있습니다."

# Build and run
build-run: copy-docs build run

# Generate Swagger documentation
swagger:
	@echo "📚 Swagger 문서를 생성합니다..."
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g cmd/server/main.go -o docs; \
		echo "✅ Swagger documentation generated successfully!"; \
		echo "📖 View docs at: docs/index.html"; \
	elif [ -f "$$(go env GOBIN)/swag" ]; then \
		$$(go env GOBIN)/swag init -g cmd/server/main.go -o docs; \
		echo "✅ Swagger documentation generated successfully!"; \
		echo "📖 View docs at: docs/index.html"; \
	else \
		echo "⚠️  swag 명령어를 찾을 수 없습니다."; \
		echo "다음 명령어로 설치하세요:"; \
		echo "go install github.com/swaggo/swag/cmd/swag@latest"; \
		exit 1; \
	fi

# Restart services
restart: stop run

# Copy docs from host to container
copy-docs:
	@echo "📄 docs 폴더를 확인합니다..."
	@if [ ! -d "docs" ]; then \
		echo "⚠️  docs 폴더가 없습니다. swagger 명령어를 먼저 실행하세요:"; \
		echo "make swagger"; \
		exit 1; \
	fi
	@echo "✅ docs 폴더가 준비되었습니다."
