# Role & Objective
You are an Expert Enterprise Cloud Architect, Data Engineer, and Go Developer.
Your objective is to implement the final overarching layer of our Go Monorepo Microservices ecosystem: "The Intelligence Layer". This phase introduces MDM (Master Data Management) and DSS (Decision Support System) to prepare the system for advanced AI and Data Science integration.

# 1. System Context & Target Architecture
We have a fully functional operational and managerial suite (Gateway, IAM, POS, CRM, OMS, SCM, EDI, ERP, HRM). 
Now, we need:
1.  **mdm-api**: A service to enforce data consistency. It will act as the "Single Source of Truth" for global entities. For this MVP, it will provide an endpoint to standardize and validate Customer and Supplier data formats before they are saved in their respective services.
2.  **dss-api**: A Decision Support System service. It will pull aggregated historical data from the `erp-api` and apply basic analytical logic (e.g., trend calculation) to serve as a backend for a future AI Agent Dashboard.

# 2. Tech Stack & Rules
- Backend: Go (Golang)
- Architecture: Monorepo structure inside `services/`.
- Communication: Service-to-service HTTP calls.
- Concurrency: Use Go's Goroutines to fetch and process analytical data in the DSS service efficiently.

# 3. Execution Plan (Todo List) & Expected Results
Please execute this project step-by-step. Do not proceed to the next step until I confirm the result of the current step.

- [ ] STEP 1: Implement the MDM Service
  - Task: Create `services/mdm-api/`. Create an endpoint `POST /mdm/validate/entity` that accepts raw JSON data, standardizes formats (e.g., phone numbers, email capitalization), and returns the clean entity. 
  - Expected Result: A running MDM service capable of standardizing data payloads.

- [ ] STEP 2: Implement the DSS Service
  - Task: Create `services/dss-api/`. Define an endpoint `GET /dss/insights/sales-trend`. This endpoint must fetch data from the existing `erp-api` and calculate a simple metric (e.g., simulated month-over-month growth).
  - Expected Result: A running DSS service returning structured JSON insights derived from ERP data.

- [ ] STEP 3: API Gateway Integration
  - Task: Update `services/api-gateway/` to route `/mdm/*` to `mdm-api` and `/dss/*` to `dss-api`. Protect these routes with IAM JWT. Update the root `docker-compose.yml` to include these 2 new services.
  - Expected Result: I can securely access the new intelligence endpoints via the Gateway.

# Instruction to Agent:
Acknowledge this master prompt by summarizing the difference between ERP (which aggregates current financial data) and DSS (which analyzes data for future insights). Do not write any code yet. Ask for my permission to begin STEP 1.