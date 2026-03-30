# Role & Objective
You are an Expert Enterprise Cloud Architect & Go Developer.
Your objective is to transition my current Go Monolith (POS+WMS) into a Monorepo Microservices Architecture. We will introduce two new core components: an API Gateway and an IAM (Identity and Access Management) service.

# 1. System Context & Target Architecture
Currently, everything is tightly coupled in the `internal/` and `cmd/api/` directories. 
We need to refactor this into a Monorepo containing three distinct services:
1.  **pos-api**: The existing POS and WMS logic.
2.  **iam-api**: A new service responsible for user authentication and issuing JWTs.
3.  **api-gateway**: A new reverse proxy service that intercepts all client requests, validates the JWT with the `iam-api` (or internally if sharing the secret), and routes valid requests to the `pos-api`.

# 2. Tech Stack & Rules
- Backend: Go (Golang)
- Gateway: Use Go's standard `net/http/httputil` (ReverseProxy) or Fiber's proxy middleware.
- IAM: Generate standard JWTs (JSON Web Tokens).
- Shared Code: Move reusable components (e.g., config loading, observability, database connection logic) to a shared `pkg/` directory.

# 3. Execution Plan (Todo List) & Expected Results
Please execute this project step-by-step. Do not proceed to the next step until I confirm the result of the current step.

- [ ] STEP 1: Monorepo Restructuring
  - Task: Create a `services/pos-api/` directory and move the existing `cmd/api/` and `internal/` logic into it. Create a `pkg/` directory for shared code (like the Prometheus setup). Update `go.mod` if necessary.
  - Expected Result: A clean Monorepo directory structure. The existing POS API should still build and run successfully from its new location.

- [ ] STEP 2: Implement the IAM Service
  - Task: Create `services/iam-api/`. Build a simple endpoint (e.g., POST `/login`) that accepts a mock username/password and returns a signed JWT.
  - Expected Result: A running IAM service on a separate port. I can curl the login endpoint and receive a valid JWT token.

- [ ] STEP 3: Implement the API Gateway
  - Task: Create `services/api-gateway/`. Implement a reverse proxy that listens on a public port. Add a middleware that requires a valid JWT for specific routes (like `/api/orders`). If valid, proxy the request to the `pos-api` port.
  - Expected Result: A running Gateway. Requests without a token are rejected (HTTP 401). Requests with a valid token are routed to the POS API, and the correct response is returned to the client.

- [ ] STEP 4: Update Docker & Observability
  - Task: Update `docker-compose.yml` to spin up all three Go services (gateway, iam, pos) alongside the existing Prometheus and Grafana setup.
  - Expected Result: Running `docker-compose up` launches the entire microservices ecosystem.

# Instruction to Agent:
Acknowledge this master prompt by summarizing the transition from Monolith to Monorepo Microservices. Describe the request flow from Client -> Gateway -> IAM/POS. Do not write any code yet. Ask for my permission to begin STEP 1.