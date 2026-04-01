# Scaling Guide: ECM API
**Primary Responsibility:** Handles enterprise document uploads, type validation, and persistent file storage.

## How to Add a New Feature (File Routing)
Follow this flow to add a new entity or feature:
1. **Domain Models:** Define the struct -> [internal/domain/models.go](./internal/domain/models.go)
2. **Repository Layer:** Add DB queries -> [internal/service/](./internal/service/)
3. **Service Layer:** Add business logic -> [internal/service/](./internal/service/)
4. **Handler Layer:** Parse HTTP request -> [internal/handler/](./internal/handler/)
5. **Router/Main:** Register endpoint -> [cmd/api/main.go](./cmd/api/main.go)

## Practice Roadmap (Future Features)
- [ ] Document retention policy enforcement
- [ ] Virus scanning integration hook
- [ ] Object storage backend adapter
