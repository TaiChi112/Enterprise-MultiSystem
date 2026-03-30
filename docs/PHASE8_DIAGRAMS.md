# Role & Objective
You are an Expert Enterprise Software Architect. 
Your objective is to reverse-engineer our existing Go Monorepo Microservices codebase and generate comprehensive Mermaid.js diagrams (Sequence, Class, and State diagrams) to help visualize the system architecture and data flows.

# 1. System Context
The project is a fully functional Microservices MVP (Gateway, IAM, POS, OMS, CRM, SCM, EDI, ERP, HRM, MDM, DSS, ECM, IDP). 
We need to document the architecture visually using Mermaid syntax inside a central Markdown file.

# 2. Output Requirements
You will create a new file named `docs/ARCHITECTURE.md`.
Inside this file, you must use valid ` ```mermaid ` code blocks for every diagram.
Do not invent features; only document what is logically present or intended in the current MVP ecosystem.

# 3. Execution Plan (Todo List)
Please execute this project step-by-step. Do not proceed to the next step until I confirm the result of the current step.

- [ ] STEP 1: Global Sequence Diagram (The Checkout Flow)
  - Task: Create `docs/ARCHITECTURE.md`. Add a Sequence Diagram illustrating a complex cross-service flow: A user authenticates via IAM, sends a purchase request via API Gateway to POS. POS verifies the customer with CRM, and then initiates an order in OMS. 
  - Expected Result: A Mermaid sequence diagram correctly showing actors, participants, and message flows.

- [ ] STEP 2: Domain Class Diagrams (Core Entities)
  - Task: Append to `docs/ARCHITECTURE.md`. Generate Mermaid Class Diagrams representing the core data structures for three major services: `pos-api`, `oms-api`, and `scm-api`.
  - Expected Result: Class diagrams showing attributes and relationships (e.g., Order has many OrderItems).

- [ ] STEP 3: State Diagram (Order Lifecycle)
  - Task: Append to `docs/ARCHITECTURE.md`. Generate a Mermaid State Diagram specifically for the Order Lifecycle managed by `oms-api`.
  - Expected Result: A state diagram showing transitions from initialization (e.g., Pending) to completion or cancellation.

- [ ] STEP 4: High-Level System Architecture Diagram
  - Task: Append to `docs/ARCHITECTURE.md`. Generate a Mermaid Graph (flowchart LR or TD) showing the API Gateway routing traffic to all the underlying microservices, demonstrating the overall topology.
  - Expected Result: A complete architectural map of the monorepo ecosystem.

# Instruction to Agent:
Acknowledge this master prompt by summarizing why combining Sequence, Class, and State diagrams provides a complete architectural picture of a Microservices system. Confirm you will use standard Mermaid syntax. Do not generate the diagrams yet. Ask for my permission to begin STEP 1.