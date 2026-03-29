.PHONY: help build run clean db-up db-down db-init test lint

help:
	@echo "POS & WMS MVP - Available Commands:"
	@echo ""
	@echo "  make build        - Build the Go application"
	@echo "  make run          - Run the application"
	@echo "  make dev          - Run in development mode (with hot reload)"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make db-up        - Start PostgreSQL in Docker"
	@echo "  make db-down      - Stop PostgreSQL container"
	@echo "  make db-init      - Initialize database schema"
	@echo "  make test         - Run tests"
	@echo "  make lint         - Run linter"
	@echo "  make tidy         - Tidy go.mod"
	@echo ""

build:
	go build -o bin/api ./cmd/api/main.go

run: build
	./bin/api

dev:
	go run ./cmd/api/main.go

clean:
	rm -rf bin/

db-up:
	docker compose up -d

db-down:
	docker compose down

db-init:
	docker compose exec postgres psql -U postgres -d pos_wms -f /docker-entrypoint-initdb.d/01-schema.sql

db-logs:
	docker compose logs -f postgres

test:
	go test ./...

lint:
	golangci-lint run ./...

tidy:
	go mod tidy

deps:
	go get -u ./...
