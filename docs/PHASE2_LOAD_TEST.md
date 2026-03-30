# Role & Objective
You are an Expert Enterprise Performance Engineer & Go Developer.
Your objective is to extend my existing Go-based POS + WMS MVP by implementing a Load Testing Bot (Chaos Engineering) and an Observability Dashboard.

# 1. System Context
We have a working POS + WMS system. Now, we need to stress-test it to see how Go handles high concurrency. We will build a bot to spam requests, instrument our target API to expose metrics, and visualize the system's behavior under stress using a dashboard.

# 2. Tech Stack & Tools
- Target API Metrics: Go `prometheus/client_golang` (to expose /metrics endpoint)
- Load Testing Bot: Go (using Goroutines, WaitGroup, and HTTP client)
- Dashboard & Scraping: Prometheus & Grafana (run via Docker Compose)
- Defense Mechanism: Go standard Rate Limiting or Fiber Rate Limiter middleware

# 3. Execution Plan (Todo List) & Expected Results
Please execute this project step-by-step. Do not proceed to the next step until I confirm the result of the current step.

- [ ] STEP 1: Instrument the Target API
  - Task: Add Prometheus metrics middleware to the existing Go POS API. Expose a `/metrics` endpoint. Track at least `http_requests_total` and `http_request_duration_seconds`.
  - Expected Result: Go code updated. I can access `http://localhost:<port>/metrics` and see raw Prometheus text data.

- [ ] STEP 2: Setup Observability Infrastructure
  - Task: Create a `docker-compose.yml` file to spin up Prometheus and Grafana. Provide a basic `prometheus.yml` configuration to scrape the Go API.
  - Expected Result: Docker Compose file and config file created. Instructions on how to start them and access the Grafana UI.

- [ ] STEP 3: Create the Go Load Bot
  - Task: Create a new standalone Go script (e.g., `cmd/loadbot/main.go`). It must use Goroutines to send thousands of concurrent POST requests to the POS "Create Order" endpoint.
  - Expected Result: A Go program that accepts parameters (e.g., number of workers, total requests) and executes concurrent HTTP requests efficiently without crashing itself.

- [ ] STEP 4: Implement System Defense (Rate Limiting)
  - Task: Add a Rate Limiting middleware to the POS API to protect it from the Load Bot. Return HTTP 429 (Too Many Requests) when limits are exceeded.
  - Expected Result: Go API code updated. When running the Load Bot again, we should see HTTP 429 responses, and the database should not be overwhelmed.

## STEP 4 Runbook Checklist (Team Operations)

Use this checklist in order for repeatable validation.

### A. Pre-Run Checks
- [ ] API source is up to date and builds: `go test ./...`
- [ ] PostgreSQL is reachable and healthy
- [ ] Test data exists for one branch and at least one product with enough inventory
- [ ] Prometheus stack is running (optional but recommended): `docker compose up -d`

### B. Start API with Rate Limit Policy
- [ ] Set limit policy (example): `RATE_LIMIT_MAX=5` and `RATE_LIMIT_WINDOW=1s`
- [ ] Start API: `PORT=3003 RATE_LIMIT_MAX=5 RATE_LIMIT_WINDOW=1s go run ./services/pos-api/cmd/api/main.go`
- [ ] Verify health endpoint: `curl http://localhost:3003/api/health`

### C. Execute Load Test Profiles
- [ ] Smoke run:
  - `go run ./cmd/loadbot/main.go -url http://localhost:3003/api/sales -profile smoke -branch-id <id> -product-ids <id>`
- [ ] Stress run:
  - `go run ./cmd/loadbot/main.go -url http://localhost:3003/api/sales -profile stress -branch-id <id> -product-ids <id>`
- [ ] Spike run:
  - `go run ./cmd/loadbot/main.go -url http://localhost:3003/api/sales -profile spike -branch-id <id> -product-ids <id>`

### D. Validate Protection Behavior
- [ ] Loadbot summary contains HTTP `429`
- [ ] API still responds on `/api/health` during and after load
- [ ] Prometheus shows increase in:
  - `http_rate_limit_429_total`
  - `http_requests_total{status="429"}`
- [ ] Optional quick snapshot endpoint returns increasing count:
  - `curl http://localhost:3003/metrics/rate-limit`

### E. Database Stability Checks
- [ ] Active DB sessions stay in expected range during test
- [ ] No runaway transaction growth or lock contention symptoms
- [ ] API process remains stable (no crash/restart loop)

### F. Exit Criteria (PASS)
- [ ] Rate limiter returns `429` under high request pressure
- [ ] API availability remains healthy
- [ ] DB stays responsive and not overwhelmed
- [ ] Metrics are visible in Prometheus/Grafana for post-analysis

### G. Rollback / Recovery
- [ ] Stop loadbot process
- [ ] Reduce limiter aggressiveness or revert env values if legitimate traffic is blocked
- [ ] Restart API with safe defaults:
  - `RATE_LIMIT_MAX=50`
  - `RATE_LIMIT_WINDOW=1s`
- [ ] Re-check health and baseline metrics before ending incident/testing window

# Instruction to Agent:
Acknowledge this master prompt by summarizing the architecture of this testing phase. Do not write any code yet. Ask for my permission to begin STEP 1.