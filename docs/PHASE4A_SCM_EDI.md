# Role & Objective
You are an Expert Enterprise Cloud Architect & Go Developer.
Your objective is to expand our existing Go Monorepo Microservices architecture by introducing two new components focused on supply chain and external communication: SCM (Supply Chain Management) and an EDI (Electronic Data Interchange) Gateway.

# 1. System Context & Target Architecture
Currently, we have `api-gateway`, `iam-api`, `pos-api` (acting as POS+WMS), `crm-api`, and `oms-api`.
To automate inventory replenishment, we need:
1.  **scm-api**: Manages Supplier profiles and generates Purchase Orders (POs) when internal inventory (in pos-api) falls below a threshold.
2.  **edi-worker/api**: Acts as a B2B integration layer. It listens for approved POs from the `scm-api`, transforms them into a standardized format (e.g., a specific JSON payload simulating an EDI document), and simulates transmitting them to an external vendor.

# 2. Tech Stack & Rules
- Backend: Go (Golang)
- Architecture: Continue using the Monorepo structure inside `services/`.
- Communication: Service-to-service HTTP calls for MVP (e.g., SCM polling POS for low stock, SCM sending payload to EDI).
- Standards: Keep domain logic isolated. Write clean Go code.

# 3. Execution Plan (Todo List) & Expected Results
Please execute this project step-by-step. Do not proceed to the next step until I confirm the result of the current step.

- [ ] STEP 1: Implement the SCM Service
  - Task: Create `services/scm-api/`. Define Domain Models for `Supplier` (ID, Name, Contact) and `PurchaseOrder` (ID, SupplierID, ProductID, Quantity, Status). Create standard CRUD endpoints for Suppliers.
  - Expected Result: A running SCM service. I can create a supplier via API.

- [ ] STEP 2: Implement the EDI Simulator Service
  - Task: Create `services/edi-api/`. This service exposes an endpoint to accept an internal PO payload, transforms it into an "External EDI Format" (just a struct mapping), and logs "TRANSMITTED TO VENDOR successfully" to simulate an external HTTP call.
  - Expected Result: A running EDI service that acts as a dummy external gateway.

- [ ] STEP 3: API Gateway Integration
  - Task: Update `services/api-gateway/` to route `/scm/*` to the `scm-api`. Ensure these routes are protected by the IAM JWT middleware.
  - Expected Result: I can access the SCM endpoints securely via the API Gateway port.

- [ ] STEP 4: Inventory Threshold Check Flow (POS -> SCM -> EDI)
  - Task: Implement a workflow where SCM can create a Purchase Order. Add an endpoint in SCM like `POST /scm/replenish` which accepts a ProductID. SCM will create a PO for a default supplier, and then send it to the `edi-api` for transmission. Update `docker-compose.yml` to include the 2 new services.
  - Expected Result: A complete flow showing internal system triggering an external B2B communication simulation.

# Instruction to Agent:
Acknowledge this master prompt by summarizing the relationship between WMS (inside pos-api), SCM, and EDI. Explain the data flow in STEP 4. Do not write any code yet. Ask for my permission to begin STEP 1.