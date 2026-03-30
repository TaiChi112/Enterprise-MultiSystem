.PHONY: help build run clean db-up db-down db-init test lint smoke smoke-up smoke-up-build smoke-down

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
	@echo "  make smoke        - One-shot smoke test via gateway (health/login/protected route)"
	@echo "  make smoke-up     - Bring stack up then run smoke test"
	@echo "  make smoke-up-build - Build+start stack then run smoke test"
	@echo "  make smoke-down   - Stop and remove stack after smoke testing"
	@echo "  make tidy         - Tidy go.mod"
	@echo ""

build:
	go build -o bin/api ./services/pos-api/cmd/api/main.go

run: build
	./bin/api

dev:
	go run ./services/pos-api/cmd/api/main.go

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

smoke:
	@set -eu; \
	echo "[smoke] checking gateway health"; \
	HEALTH_CODE=$$(curl -sS -o /tmp/smoke_health.json -w "%{http_code}" http://localhost:8080/api/health); \
	if [ "$$HEALTH_CODE" != "200" ]; then \
		echo "[smoke] gateway health failed with status $$HEALTH_CODE"; \
		cat /tmp/smoke_health.json; \
		exit 1; \
	fi; \
	echo "[smoke] login via gateway"; \
	LOGIN_JSON=$$(curl -sS -X POST http://localhost:8080/login -H "Content-Type: application/json" -d '{"username":"admin","password":"admin123"}'); \
	TOKEN=$$(echo "$$LOGIN_JSON" | sed -n 's/.*"access_token":"\([^"]*\)".*/\1/p'); \
	if [ -z "$$TOKEN" ]; then \
		echo "[smoke] login did not return access_token"; \
		echo "$$LOGIN_JSON"; \
		exit 1; \
	fi; \
	echo "[smoke] protected route without token should be 401"; \
	NO_TOKEN_CODE=$$(curl -sS -o /tmp/smoke_no_token.json -w "%{http_code}" http://localhost:8080/api/orders/1); \
	if [ "$$NO_TOKEN_CODE" != "401" ]; then \
		echo "[smoke] expected 401 without token, got $$NO_TOKEN_CODE"; \
		cat /tmp/smoke_no_token.json; \
		exit 1; \
	fi; \
	echo "[smoke] protected route with token should pass auth (not 401/502)"; \
	WITH_TOKEN_CODE=$$(curl -sS -o /tmp/smoke_with_token.json -w "%{http_code}" http://localhost:8080/api/orders/1 -H "Authorization: Bearer $$TOKEN"); \
	if [ "$$WITH_TOKEN_CODE" = "401" ] || [ "$$WITH_TOKEN_CODE" = "502" ]; then \
		echo "[smoke] expected proxied response, got $$WITH_TOKEN_CODE"; \
		cat /tmp/smoke_with_token.json; \
		exit 1; \
	fi; \
	echo "[smoke] PASSED (health=200, login token issued, unauthorized blocked, authorized proxied status=$$WITH_TOKEN_CODE)"

smoke-up:
	@set -eu; \
	echo "[smoke-up] starting docker compose services"; \
	docker compose up -d; \
	echo "[smoke-up] running smoke checks"; \
	$(MAKE) smoke

smoke-up-build:
	@set -eu; \
	echo "[smoke-up-build] building and starting docker compose services"; \
	docker compose up -d --build; \
	echo "[smoke-up-build] running smoke checks"; \
	$(MAKE) smoke

smoke-down:
	@set -eu; \
	echo "[smoke-down] stopping docker compose services"; \
	docker compose down
