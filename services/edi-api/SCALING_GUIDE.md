# Scaling Guide: EDI API
**Primary Responsibility:** Transforms internal purchase order payloads into external EDI-like messages and simulates partner transmission.

## How to Add a New Feature (File Routing)
Follow this flow to add a new entity or feature:
1. **Domain Models:** Define the struct -> [internal/domain/models.go](./internal/domain/models.go)
2. **Repository Layer:** Add DB queries -> [internal/service/](./internal/service/)
3. **Service Layer:** Add business logic -> [internal/service/](./internal/service/)
4. **Handler Layer:** Parse HTTP request -> [internal/handler/](./internal/handler/)
5. **Router/Main:** Register endpoint -> [cmd/api/main.go](./cmd/api/main.go)

## Practice Roadmap (Future Features)
- [ ] EDIFACT message mapper
- [ ] Partner endpoint retry queue
- [ ] Outbound message signature verification
