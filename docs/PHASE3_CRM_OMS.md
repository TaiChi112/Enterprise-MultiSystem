# Role & Objective
You are an Expert Enterprise Cloud Architect & Go Developer.
Your objective is to expand our existing Go Monorepo Microservices architecture by introducing two new services: CRM (Customer Relationship Management) and OMS (Order Management System).

# 1. System Context & Target Architecture
Currently, we have `api-gateway`, `iam-api`, and `pos-api` (which handles basic transactions and inventory). 
To scale into an Enterprise Commerce Platform, we need:
1.  **crm-api**: Manages customer profiles, loyalty points, and membership levels.
2.  **oms-api**: Acts as the central orchestrator for order lifecycles (e.g., Pending -> Paid -> Shipped -> Completed). It will eventually replace the simple order creation in POS for cross-channel orders.
3.  **Integration**: The `api-gateway` must route traffic to these new services. The services will communicate via REST (HTTP) for this MVP phase.

# 2. Tech Stack & Rules
- Backend: Go (Golang)
- Architecture: Continue using the Monorepo structure (`services/crm-api` and `services/oms-api`). Follow Clean Architecture (Handler -> Service -> Repository).
- Database: PostgreSQL. (Assume logical separation: each service handles its own tables/domain).
- Communication: Internal service-to-service HTTP calls or through the API Gateway.

# 3. Execution Plan (Todo List) & Expected Results
Please execute this project step-by-step. Do not proceed to the next step until I confirm the result of the current step.

- [ ] STEP 1: Implement the CRM Service
  - Task: Create `services/crm-api/`. Define Domain Models for `Customer` (ID, Name, Email, Phone, LoyaltyPoints). Implement standard CRUD API endpoints.
  - Expected Result: A running CRM service. I can create a customer and retrieve their details.

- [ ] STEP 2: Implement the OMS Service
  - Task: Create `services/oms-api/`. Define Domain Models for `OrderLifecycle` tracking the status of an order. Create an endpoint to "Initialize Order" which sets status to "Pending".
  - Expected Result: A running OMS service capable of tracking order states independently of the POS physical transaction.

- [ ] STEP 3: API Gateway & IAM Integration
  - Task: Update `services/api-gateway/` to route `/customers/*` to the `crm-api` and `/orders/*` to the `oms-api`. Ensure these routes are protected by the existing IAM JWT middleware.
  - Expected Result: I can access the new CRM and OMS endpoints securely via the API Gateway port.

- [ ] STEP 4: Service-to-Service Communication (POS -> CRM)
  - Task: Update `pos-api` to accept a `CustomerID` during a transaction. The `pos-api` should make an internal HTTP call to `crm-api` to validate the customer and optionally award loyalty points after a successful sale.
  - Expected Result: A complete transaction flow where POS talks to CRM, proving microservices integration. Update `docker-compose.yml` to include all 5 services.

# Instruction to Agent:
Acknowledge this master prompt by summarizing the purpose of CRM and OMS in this architecture. Explain how the POS service will communicate with the CRM service in STEP 4. Do not write any code yet. Ask for my permission to begin STEP 1.