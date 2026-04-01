# Scaling Guide: IAM API
**Primary Responsibility:** Authenticates users and issues JWT tokens for secure service access.

## How to Add a New Feature (File Routing)
Follow this flow to add a new entity or feature:
1. **Domain Models:** Define the struct -> [internal/domain/models.go](./internal/domain/models.go)
2. **Repository Layer:** Add identity persistence integration -> [internal/service/](./internal/service/)
3. **Service Layer:** Add business logic -> [internal/service/](./internal/service/)
4. **Handler Layer:** Parse HTTP request -> [internal/handler/](./internal/handler/)
5. **Router/Main:** Register endpoint -> [cmd/api/main.go](./cmd/api/main.go)

## Practice Roadmap (Future Features)
- [ ] Refresh token rotation flow
- [ ] Role-based permission matrix
- [ ] Account lockout and brute-force guard
