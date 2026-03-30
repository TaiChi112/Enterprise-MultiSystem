# Architecture Diagrams

## STEP 1: Global Sequence Diagram (Checkout Flow)

```mermaid
sequenceDiagram
    autonumber
    actor User
    participant IAM as IAM API
    participant GW as API Gateway
    participant POS as POS API
    participant CRM as CRM API
    participant OMS as OMS API

    User->>IAM: POST /login (credentials)
    IAM-->>User: 200 OK + JWT

    User->>GW: POST /api/sales (Bearer JWT)
    GW->>GW: Validate JWT
    GW->>POS: Forward purchase request

    POS->>CRM: Verify customer profile/eligibility
    CRM-->>POS: Customer verified

    POS->>OMS: Create order
    OMS-->>POS: Order created (orderId, status)

    POS-->>GW: Sale accepted + order summary
    GW-->>User: 200 OK + checkout result
```

## STEP 2: Domain Class Diagrams (Core Entities)

### pos-api

```mermaid
classDiagram
    class Product {
        +int ID
        +string SKU
        +string Name
        +string Description
        +float64 Price
        +float64 Cost
        +bool IsActive
        +time.Time CreatedAt
        +time.Time UpdatedAt
    }

    class Branch {
        +int ID
        +string Name
        +string Address
        +string Phone
        +bool IsActive
        +time.Time CreatedAt
        +time.Time UpdatedAt
    }

    class Inventory {
        +int ID
        +int ProductID
        +int BranchID
        +int Quantity
        +int MinimumQty
        +time.Time CreatedAt
        +time.Time UpdatedAt
    }

    class Order {
        +int ID
        +int BranchID
        +string CustomerName
        +float64 TotalAmount
        +string Status
        +time.Time CreatedAt
        +time.Time UpdatedAt
    }

    class OrderItem {
        +int ID
        +int OrderID
        +int ProductID
        +int Quantity
        +float64 UnitPrice
        +float64 Discount
        +time.Time CreatedAt
        +time.Time UpdatedAt
    }

    Product "1" <-- "0..*" Inventory : tracked as
    Branch "1" <-- "0..*" Inventory : stocked at
    Branch "1" <-- "0..*" Order : placed at
    Order "1" *-- "1..*" OrderItem : contains
    Product "1" <-- "0..*" OrderItem : purchased product
```

### oms-api

```mermaid
classDiagram
    class OrderLifecycle {
        +int ID
        +string OrderNumber
        +int CustomerID
        +string Status
        +float64 TotalAmount
        +string Description
        +bool IsActive
        +time.Time CreatedAt
        +time.Time UpdatedAt
    }

    class OrderItem {
        +int ID
        +int OrderID
        +int ProductID
        +string ProductName
        +int Quantity
        +float64 UnitPrice
        +float64 LineTotal
        +time.Time CreatedAt
        +time.Time UpdatedAt
    }

    OrderLifecycle "1" *-- "0..*" OrderItem : has items
```

### scm-api

```mermaid
classDiagram
    class Supplier {
        +int ID
        +string Name
        +string Contact
        +time.Time CreatedAt
        +time.Time UpdatedAt
    }

    class PurchaseOrder {
        +int ID
        +int SupplierID
        +int ProductID
        +int Quantity
        +string Status
        +float64 UnitCost
        +float64 TotalCost
        +time.Time CreatedAt
        +time.Time UpdatedAt
    }

    Supplier "1" <-- "0..*" PurchaseOrder : receives orders
```

## STEP 3: State Diagram (Order Lifecycle)

```mermaid
stateDiagram-v2
    [*] --> Pending : InitializeOrder

    Pending --> Paid : UpdateOrderStatus(paid)
    Pending --> Cancelled : UpdateOrderStatus(cancelled)

    Paid --> Shipped : UpdateOrderStatus(shipped)
    Paid --> Cancelled : UpdateOrderStatus(cancelled)

    Shipped --> Completed : UpdateOrderStatus(completed)
    Shipped --> Cancelled : UpdateOrderStatus(cancelled)

    Completed --> [*]
    Cancelled --> [*]
```

## STEP 4: High-Level System Architecture Diagram

```mermaid
flowchart LR
    U[Client / User]
    GW[api-gateway]

    IAM[iam-api]
    POS[pos-api]
    CRM[crm-api]
    OMS[oms-api]
    SCM[scm-api]
    EDI[edi-api]
    HRM[hrm-api]
    ERP[erp-api]
    MDM[mdm-api]
    DSS[dss-api]
    ECM[ecm-api]
    IDP[idp-api]
    PG[(PostgreSQL)]

    U -->|HTTP| GW

    GW -->|/login| IAM
    GW -->|default /api/*| POS
    GW -->|/api/customers| CRM
    GW -->|/api/orders| OMS
    GW -->|/scm| SCM
    GW -->|/hrm| HRM
    GW -->|/erp| ERP
    GW -->|/mdm| MDM
    GW -->|/dss| DSS
    GW -->|/ecm| ECM
    GW -->|/idp| IDP

    POS --> PG
    CRM --> PG
    OMS --> PG
    SCM --> PG
    HRM --> PG

    SCM -->|EDI_API_URL /edi/transmit| EDI
    ERP -->|OMS_API_URL| OMS
    ERP -->|SCM_API_URL| SCM
    ERP -->|HRM_API_URL| HRM
    DSS -->|ERP_API_URL| ERP
```
