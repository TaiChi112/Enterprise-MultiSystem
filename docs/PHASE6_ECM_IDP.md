# Role & Objective
You are an Expert Enterprise Cloud Architect & Go Developer.
Your objective is to implement Phase 6 of our Go Monorepo Microservices ecosystem: The Document & AI Pipeline. We will introduce ECM (Enterprise Content Management) for file handling and IDP (Intelligent Document Processing) to simulate AI-driven text extraction.

# 1. System Context & Target Architecture
We have a robust structured-data ecosystem. Now we must handle unstructured data (PDFs, images).
1.  **ecm-api**: A service dedicated to accepting file uploads, validating file types (e.g., image/jpeg, application/pdf), and storing them (local disk for this MVP).
2.  **idp-api** (or embedded in ECM for MVP): A service that takes the uploaded file's metadata, simulates passing it to an AI/OCR agent, and returns structured JSON (e.g., extracting "Total Amount" from a simulated receipt).

# 2. Tech Stack & Rules
- Backend: Go (Golang)
- Architecture: Monorepo structure inside `services/`.
- File Handling: Use standard Go `net/http` or Fiber methods for `multipart/form-data`.
- Pipeline: The ECM will receive the file, save it, and immediately call the IDP logic to extract data, returning both the File ID and the Extracted Data to the user.

# 3. Execution Plan (Todo List) & Expected Results
Please execute this project step-by-step. Do not proceed to the next step until I confirm the result of the current step.

- [ ] STEP 1: Implement the ECM Service & File Upload
  - Task: Create `services/ecm-api/`. Create an endpoint `POST /ecm/upload` that accepts a multipart file. Validate that it is an image or PDF. Save the file to a local `./uploads` directory.
  - Expected Result: A running ECM service. I can upload a file using Postman or cURL and see it saved on disk.

- [ ] STEP 2: Implement the IDP Simulation Engine
  - Task: Create `services/idp-api/` (or add to ECM). Create a function/endpoint that accepts a File ID or File Path. It should simulate an AI processing delay (e.g., `time.Sleep(2 * time.Second)`) and return a hardcoded structured JSON representing extracted data (e.g., `{"document_type": "invoice", "amount": 1500.00}`).
  - Expected Result: Simulated AI extraction logic working seamlessly in Go.

- [ ] STEP 3: API Gateway Integration
  - Task: Update `services/api-gateway/` to route `/ecm/*` to the `ecm-api` and `/idp/*` to the `idp-api`. Protect with IAM JWT. Update `docker-compose.yml` to include the new services.
  - Expected Result: I can securely upload files and trigger data extraction via the API Gateway.

# Instruction to Agent:
Acknowledge this master prompt by explaining the complexities of handling `multipart/form-data` in Go and why it is different from parsing standard JSON requests. Do not write any code yet. Ask for my permission to begin STEP 1.