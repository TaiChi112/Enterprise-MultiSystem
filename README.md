# POS & WMS MVP - Point of Sale & Warehouse Management System

An enterprise-grade MVP combining **Point of Sale (POS)** and **Warehouse Management System (WMS)** built with Go, PostgreSQL, and Fiber.

## 📋 Project Overview

This MVP manages:
- **Product Catalog**: Master data for products with SKU, pricing, and cost tracking
- **Multi-Branch Inventory**: Track stock levels across multiple branch locations
- **Sales Transactions**: Process orders with automatic inventory deduction (ACID properties)
- **Order History**: Complete transaction audit trail with timestamps

## 🏗️ Architecture

```
┌──────────────────────────────────────────────────────────┐
│                    HTTP Clients                          │
└──────────────────────┬───────────────────────────────────┘
                       │ REST API
┌──────────────────────▼───────────────────────────────────┐
│              Handler Layer (Fiber Routes)               │
├──────────────────────────────────────────────────────────┤
│ /api/products, /api/branches, /api/inventory, /api/sales│
└──────────────────────┬───────────────────────────────────┘
                       │ Business Logic
┌──────────────────────▼───────────────────────────────────┐
│               Service Layer (Business Logic)            │
├──────────────────────────────────────────────────────────┤
│ - Product Management                                     │
│ - Inventory Tracking                                     │
│ - Sale Processing (ACID Transaction)                    │
└──────────────────────┬───────────────────────────────────┘
                       │ Data Access
┌──────────────────────▼───────────────────────────────────┐
│          Repository Layer (Data Access Layer)           │
├──────────────────────────────────────────────────────────┤
│ - Product Repository                                     │
│ - Branch Repository                                      │
│ - Inventory Repository                                   │
│ - Order Repository                                       │
└──────────────────────┬───────────────────────────────────┘
                       │ SQL Queries
┌──────────────────────▼───────────────────────────────────┐
│                   PostgreSQL Database                    │
├──────────────────────────────────────────────────────────┤
│ Tables: product, branch, inventory, order, order_item   │
└──────────────────────────────────────────────────────────┘
```

## 🔧 Tech Stack

| Component | Technology | Version |
|-----------|-----------|---------|
| **Language** | Go (Golang) | 1.21+ |
| **Web Framework** | Fiber | v2 |
| **Database** | PostgreSQL | 14+ |
| **Driver** | lib/pq | Latest |

## 📁 Project Structure

```
POS_Basis_WMS/
├── cmd/
│   └── api/
│       └── main.go                 # Entry point
├── config/
│   └── config.go                   # Configuration management
├── internal/
│   ├── domain/
│   │   └── models.go               # Database models & DTOs
│   ├── handler/
│   │   └── handler.go              # HTTP handlers & routes
│   ├── repository/
│   │   ├── database.go             # Database connection
│   │   ├── repository.go           # Product/Inventory CRUD
│   │   └── order_repository.go     # Order/OrderItem CRUD
│   └── service/
│       └── service.go              # Business logic & transactions
├── schema.sql                       # Database schema
├── docker-compose.yml              # PostgreSQL setup
├── .env.example                    # Environment variables template
├── go.mod                          # Go module definition
└── Makefile                        # Build commands

```

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- PostgreSQL 14+ (or use Docker)

### 1. Setup Database

```bash
# Start PostgreSQL with Docker
docker compose up -d

# Verify connection
docker compose exec postgres psql -U postgres -d pos_wms -c "SELECT COUNT(*) FROM product;"
```

### 2. Configure Environment

```bash
cp .env.example .env
# Edit .env if needed (defaults are fine for Docker setup)
```

### 3. Build & Run

```bash
# Build
go build -o bin/api ./cmd/api/main.go

# Run
./bin/api

# Or run directly
go run ./cmd/api/main.go
```

The server will start on `http://localhost:3000`

### 4. Test API

```bash
# Health check
curl http://localhost:3000/api/health
```

---

## 📚 API Endpoints

### Health Check
```bash
GET /api/health
```

### Products

#### Create Product
```bash
curl -X POST http://localhost:3000/api/products \
  -H "Content-Type: application/json" \
  -d '{
    "sku": "SKU-001",
    "name": "iPhone 15 Pro",
    "description": "Latest Apple smartphone",
    "price": 35999.00,
    "cost": 28000.00
  }'
```

#### Get Product by ID
```bash
curl http://localhost:3000/api/products/1
```

#### Get All Products
```bash
curl "http://localhost:3000/api/products?limit=10&offset=0"
```

### Branches

#### Create Branch
```bash
curl -X POST http://localhost:3000/api/branches \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Bangkok Store",
    "address": "123 Silom Rd, Bangkok",
    "phone": "02-123-4567"
  }'
```

#### Get All Branches
```bash
curl http://localhost:3000/api/branches
```

### Inventory

#### Get Inventory for Product at Branch
```bash
curl "http://localhost:3000/api/inventory/product/1/branch/1"
```

#### Get All Inventory for Branch
```bash
curl "http://localhost:3000/api/inventory/branch/1?limit=50&offset=0"
```

#### Get Low Stock Items
```bash
curl http://localhost:3000/api/inventory/low-stock/1
```

### Sales (Process Order with Inventory Deduction)

#### Process Sale ⭐ **MAIN FEATURE**
This endpoint demonstrates ACID transaction properties:
1. Creates an Order
2. Creates OrderItems
3. Deducts Inventory
4. Atomically commits or rolls back all changes

```bash
curl -X POST http://localhost:3000/api/sales \
  -H "Content-Type: application/json" \
  -d '{
    "branch_id": 1,
    "customer_name": "John Doe",
    "items": [
      {
        "product_id": 1,
        "quantity": 2,
        "discount": 0
      },
      {
        "product_id": 3,
        "quantity": 5,
        "discount": 50
      }
    ]
  }'
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "branch_id": 1,
    "customer_name": "John Doe",
    "total_amount": 72448.50,
    "status": "completed",
    "order_items": [
      {
        "id": 1,
        "product_id": 1,
        "quantity": 2,
        "unit_price": 35999.00,
        "discount": 0,
        "created_at": "2024-01-15T10:30:00Z",
        "updated_at": "2024-01-15T10:30:00Z"
      },
      {
        "id": 2,
        "product_id": 3,
        "quantity": 5,
        "unit_price": 299.00,
        "discount": 50,
        "created_at": "2024-01-15T10:30:00Z",
        "updated_at": "2024-01-15T10:30:00Z"
      }
    ],
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  },
  "message": "Sale processed successfully"
}
```

#### Get Order by ID
```bash
curl http://localhost:3000/api/orders/1
```

#### Get Orders by Branch
```bash
curl "http://localhost:3000/api/orders/branch/1?limit=50&offset=0"
```

---

## 🗄️ Database Schema

### Product Table
```sql
CREATE TABLE product (
    id SERIAL PRIMARY KEY,
    sku VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(12, 2) NOT NULL,
    cost DECIMAL(12, 2),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

### Branch Table
```sql
CREATE TABLE branch (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    address TEXT,
    phone VARCHAR(20),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

### Inventory Table
```sql
CREATE TABLE inventory (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES product(id),
    branch_id INTEGER NOT NULL REFERENCES branch(id),
    quantity INTEGER NOT NULL DEFAULT 0,
    minimum_qty INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(product_id, branch_id)
);
```

### Order Table
```sql
CREATE TABLE "order" (
    id SERIAL PRIMARY KEY,
    branch_id INTEGER NOT NULL REFERENCES branch(id),
    customer_name VARCHAR(255),
    total_amount DECIMAL(12, 2) NOT NULL DEFAULT 0,
    status VARCHAR(50) DEFAULT 'completed',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

### OrderItem Table
```sql
CREATE TABLE order_item (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES "order"(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES product(id),
    quantity INTEGER NOT NULL,
    unit_price DECIMAL(12, 2) NOT NULL,
    discount DECIMAL(12, 2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(order_id, product_id)
);
```

---

## 🔐 ACID Transaction Example

The `ProcessSale` method in `service.go` demonstrates ACID properties:

```go
func (s *Service) ProcessSale(ctx context.Context, req *domain.ProcessSaleRequest) (*domain.Order, error) {
    // 1. Begin transaction with Serializable isolation
    tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
        Isolation: sql.LevelSerializable,
    })
    
    // 2. Validate: Check branch, products, inventory
    // 3. Create: Insert Order
    // 4. Create: Insert OrderItems
    // 5. Deduct: Update inventory quantities
    
    // 6. Commit or Rollback atomically
    if err := tx.Commit(); err != nil {
        return nil, err
    }
    
    return order, nil
}
```

**Why ACID matters for POS:**
- **Atomicity**: All or nothing - no partial orders
- **Consistency**: Inventory always matches orders
- **Isolation**: Concurrent transactions don't interfere
- **Durability**: Committed data won't be lost

---

## 📊 Sample Test Data

```bash
# Create branches
curl -X POST http://localhost:3000/api/branches \
  -H "Content-Type: application/json" \
  -d '{"name": "Bangkok Store", "address": "123 Silom Rd", "phone": "02-123-4567"}'

curl -X POST http://localhost:3000/api/branches \
  -H "Content-Type: application/json" \
  -d '{"name": "Chiang Mai Store", "address": "456 Nimman Rd", "phone": "053-234-5678"}'

# Create products
curl -X POST http://localhost:3000/api/products \
  -H "Content-Type: application/json" \
  -d '{
    "sku": "SKU-001",
    "name": "iPhone 15 Pro",
    "description": "Latest Apple smartphone",
    "price": 35999,
    "cost": 28000
  }'

curl -X POST http://localhost:3000/api/products \
  -H "Content-Type: application/json" \
  -d '{
    "sku": "SKU-002",
    "name": "USB-C Cable",
    "price": 299,
    "cost": 80
  }'

# Process a sale
curl -X POST http://localhost:3000/api/sales \
  -H "Content-Type: application/json" \
  -d '{
    "branch_id": 1,
    "customer_name": "John Doe",
    "items": [
      {"product_id": 1, "quantity": 1, "discount": 0},
      {"product_id": 2, "quantity": 2, "discount": 0}
    ]
  }'
```

---

## 🛠️ Development

### Useful Make Commands

```bash
# Build and run
make run

# Development mode (auto-reload with hot reload tool)
make dev

# Database operations
make db-up      # Start PostgreSQL
make db-down    # Stop PostgreSQL
make db-init    # Initialize schema

# Testing
make test
make lint

# Dependencies
make tidy
make deps
```

---

## 🔄 Future Enhancements (Post-MVP)

1. **User Authentication & Authorization**
   - Staff accounts with role-based access
   - API key authentication

2. **Advanced Inventory Features**
   - Stock reservations (pending orders)
   - Inventory adjustments & corrections
   - Multi-warehouse transfers

3. **Reporting & Analytics**
   - Sales reports by branch/date/product
   - Inventory analytics
   - Low stock alerts

4. **AI Integration**
   - Demand forecasting
   - Dynamic pricing
   - Customer analytics

5. **Payment Processing**
   - Multiple payment methods
   - Refund handling
   - Payment gateway integration

---

## 📝 Error Handling

All endpoints return consistent error responses:

```json
{
  "success": false,
  "error": "Product not found",
  "code": "PRODUCT_NOT_FOUND"
}
```

---

## 🐛 Debugging

Enable debug logging by setting `DEBUG=1`:

```bash
DEBUG=1 go run ./cmd/api/main.go
```

Check PostgreSQL logs:

```bash
docker compose logs -f postgres
```

---

## 📄 License

MIT License - Feel free to use this for your learning and projects.

---

## ✨ Key Features Implemented

✅ Clean Architecture (Handler → Service → Repository)  
✅ CRUD operations for all entities  
✅ ACID transaction for sale processing  
✅ Inventory management with low-stock detection  
✅ Multi-branch support  
✅ PostgreSQL with proper constraints and indexes  
✅ Fiber REST API with error handling  
✅ Environment-based configuration  
✅ Docker support  
✅ Audit timestamps on all records  

---

## 📞 Support

For issues or questions, please refer to the codebase documentation or create an issue in your repository.

**Happy coding! 🚀**
