# Scaling Guide: API Gateway
**Primary Responsibility:** Enforces authentication policy and routes client requests to upstream microservices.

## How to Add a New Feature (File Routing)
Follow this flow to add a new entity or feature:
1. **Domain Models:** Define gateway structs -> [cmd/api/main.go](./cmd/api/main.go)
2. **Repository Layer:** Add upstream integration logic -> [cmd/api/main.go](./cmd/api/main.go)
3. **Service Layer:** Add routing and auth policy logic -> [cmd/api/main.go](./cmd/api/main.go)
4. **Handler Layer:** Parse HTTP request -> [cmd/api/main.go](./cmd/api/main.go)
5. **Router/Main:** Register endpoint -> [cmd/api/main.go](./cmd/api/main.go)

## Practice Roadmap (Future Features)
- [ ] Per-route rate limiting policy
- [ ] Circuit breaker for upstream failures
- [ ] Distributed request tracing headers
