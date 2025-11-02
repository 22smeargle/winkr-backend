.PHONY: help build run test test-integration clean docker-up docker-down migrate-up migrate-down lint fmt vet deps tidy

# Variables
APP_NAME := winkr-backend
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

# Help
help:
	@echo "Available commands:"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application"
	@echo "  test         - Run unit tests"
	@echo "  test-integration - Run integration tests"
	@echo "  test-cover   - Run tests with coverage"
	@echo "  test-all     - Run all tests (unit + integration)"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Download dependencies"
	@echo "  tidy         - Tidy dependencies"
	@echo "  fmt          - Format code"
	@echo "  vet          - Vet code"
	@echo "  lint         - Run linter"
	@echo "  docker-up    - Start development environment"
	@echo "  docker-down  - Stop development environment"
	@echo "  migrate-up   - Run database migrations"
	@echo "  migrate-down - Rollback database migrations"
	@echo "  migrate-create - Create new migration"
	@echo "  swagger      - Generate swagger documentation"
	@echo "  mock         - Generate mocks"

# Build
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p bin
	go build $(LDFLAGS) -o bin/$(APP_NAME) cmd/api/main.go

# Run
run:
	@echo "Running $(APP_NAME)..."
	go run cmd/api/main.go

# Unit Tests
test:
	@echo "Running unit tests..."
	go test -v ./...

# Integration Tests
test-integration:
	@echo "Running integration tests..."
	@cd tests/integration && go run runner.go

# All Tests
test-all: test test-integration
	@echo "All tests completed!"

# Test with coverage
test-cover:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html

# Dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Vet code
vet:
	@echo "Vetting code..."
	go vet ./...

# Lint
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Docker
docker-up:
	@echo "Starting development environment..."
	docker-compose -f docker/docker-compose.dev.yml up -d

docker-down:
	@echo "Stopping development environment..."
	docker-compose -f docker/docker-compose.dev.yml down

docker-logs:
	@echo "Showing docker logs..."
	docker-compose -f docker/docker-compose.dev.yml logs -f

# Database migrations
migrate-up:
	@echo "Running database migrations..."
	@if command -v migrate >/dev/null 2>&1; then \
		migrate -path migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)" up; \
	else \
		echo "migrate tool not installed. Install it with: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"; \
	fi

migrate-down:
	@echo "Rolling back database migrations..."
	@if command -v migrate >/dev/null 2>&1; then \
		migrate -path migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)" down; \
	else \
		echo "migrate tool not installed. Install it with: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"; \
	fi

migrate-force:
	@echo "Forcing database migration version..."
	@if command -v migrate >/dev/null 2>&1; then \
		migrate -path migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)" force $(VERSION); \
	else \
		echo "migrate tool not installed. Install it with: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"; \
	fi

migrate-create:
	@echo "Creating new migration..."
	@if command -v migrate >/dev/null 2>&1; then \
		read -p "Enter migration name: " name; \
		migrate create -ext sql -dir migrations -seq $$name; \
	else \
		echo "migrate tool not installed. Install it with: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"; \
	fi

# Swagger
swagger:
	@echo "Generating swagger documentation..."
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g cmd/api/main.go -o docs; \
	else \
		echo "swag tool not installed. Install it with: go install github.com/swaggo/swag/cmd/swag@latest"; \
	fi

# Mocks
mock:
	@echo "Generating mocks..."
	@if command -v mockgen >/dev/null 2>&1; then \
		mockgen -source=internal/domain/repositories/user_repository.go -destination=internal/domain/repositories/mocks/user_repository_mock.go -package=mocks; \
		mockgen -source=internal/domain/repositories/photo_repository.go -destination=internal/domain/repositories/mocks/photo_repository_mock.go -package=mocks; \
		mockgen -source=internal/domain/repositories/message_repository.go -destination=internal/domain/repositories/mocks/message_repository_mock.go -package=mocks; \
		mockgen -source=internal/domain/repositories/match_repository.go -destination=internal/domain/repositories/mocks/match_repository_mock.go -package=mocks; \
		mockgen -source=internal/domain/repositories/report_repository.go -destination=internal/domain/repositories/mocks/report_repository_mock.go -package=mocks; \
		mockgen -source=internal/domain/repositories/subscription_repository.go -destination=internal/domain/repositories/mocks/subscription_repository_mock.go -package=mocks; \
	else \
		echo "mockgen tool not installed. Install it with: go install github.com/golang/mock/mockgen@latest"; \
	fi

# Development setup
setup-dev:
	@echo "Setting up development environment..."
	@cp .env.example .env
	@echo "Please edit .env file with your configuration"
	@make deps
	@make docker-up
	@sleep 5
	@make migrate-up
	@echo "Development environment is ready!"

# Production build
build-prod:
	@echo "Building $(APP_NAME) for production..."
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=linux go build $(LDFLAGS) -a -installsuffix cgo -o bin/$(APP_NAME)-linux cmd/api/main.go
	CGO_ENABLED=0 GOOS=darwin go build $(LDFLAGS) -a -installsuffix cgo -o bin/$(APP_NAME)-darwin cmd/api/main.go
	CGO_ENABLED=0 GOOS=windows go build $(LDFLAGS) -a -installsuffix cgo -o bin/$(APP_NAME)-windows.exe cmd/api/main.go

# Install tools
install-tools:
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install github.com/golang/mock/mockgen@latest
	@go install github.com/air-verse/air@latest

# Development with hot reload
dev:
	@echo "Starting development server with hot reload..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "air tool not installed. Install it with: go install github.com/air-verse/air@latest"; \
	fi

# Security scan
security:
	@echo "Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec tool not installed. Install it with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# Benchmark
benchmark:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Profile
profile:
	@echo "Running profiling..."
	go test -cpuprofile=cpu.prof -memprofile=mem.prof -bench=. ./...
	@echo "Profile files generated: cpu.prof, mem.prof"

# Check all
check-all: fmt vet lint security test-all
	@echo "All checks completed!"