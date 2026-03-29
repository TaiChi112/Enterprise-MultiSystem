-- ============================================================================
-- POS & WMS MVP - Database Schema
-- Database: PostgreSQL
-- Purpose: Product Catalog, Inventory Management, Sales Transactions
-- ============================================================================

-- ============================================================================
-- 1. PRODUCT TABLE - Master data for products
-- ============================================================================
CREATE TABLE product (
    id SERIAL PRIMARY KEY,
    sku VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(12, 2) NOT NULL,
    cost DECIMAL(12, 2),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT product_price_positive CHECK (price > 0),
    CONSTRAINT product_cost_non_negative CHECK (cost >= 0)
);

CREATE INDEX idx_product_sku ON product(sku);
CREATE INDEX idx_product_is_active ON product(is_active);
CREATE INDEX idx_product_created_at ON product(created_at DESC);

-- ============================================================================
-- 2. BRANCH TABLE - Physical branch/location data
-- ============================================================================
CREATE TABLE branch (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    address TEXT,
    phone VARCHAR(20),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_branch_is_active ON branch(is_active);
CREATE INDEX idx_branch_created_at ON branch(created_at DESC);

-- ============================================================================
-- 3. INVENTORY TABLE - Stock levels per product per branch
-- ============================================================================
CREATE TABLE inventory (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL,
    branch_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 0,
    minimum_qty INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_inventory_product FOREIGN KEY (product_id) 
        REFERENCES product(id) ON DELETE CASCADE,
    CONSTRAINT fk_inventory_branch FOREIGN KEY (branch_id) 
        REFERENCES branch(id) ON DELETE CASCADE,
    CONSTRAINT inventory_quantity_non_negative CHECK (quantity >= 0),
    CONSTRAINT inventory_minimum_qty_non_negative CHECK (minimum_qty >= 0),
    CONSTRAINT inventory_unique_product_branch UNIQUE (product_id, branch_id)
);

CREATE INDEX idx_inventory_product_id ON inventory(product_id);
CREATE INDEX idx_inventory_branch_id ON inventory(branch_id);
CREATE INDEX idx_inventory_quantity_low ON inventory(quantity) WHERE quantity <= minimum_qty;

-- ============================================================================
-- 4. ORDER TABLE - Sales transaction header
-- ============================================================================
CREATE TABLE "order" (
    id SERIAL PRIMARY KEY,
    branch_id INTEGER NOT NULL,
    customer_name VARCHAR(255),
    total_amount DECIMAL(12, 2) NOT NULL DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'completed',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_order_branch FOREIGN KEY (branch_id) 
        REFERENCES branch(id) ON DELETE RESTRICT,
    CONSTRAINT order_total_amount_non_negative CHECK (total_amount >= 0),
    CONSTRAINT order_status_valid CHECK (status IN ('pending', 'completed', 'cancelled', 'refunded'))
);

CREATE INDEX idx_order_branch_id ON "order"(branch_id);
CREATE INDEX idx_order_created_at ON "order"(created_at DESC);
CREATE INDEX idx_order_branch_created ON "order"(branch_id, created_at DESC);
CREATE INDEX idx_order_status ON "order"(status);

-- ============================================================================
-- 5. ORDER_ITEM TABLE - Sales transaction details (line items)
-- ============================================================================
CREATE TABLE order_item (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL,
    product_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL,
    unit_price DECIMAL(12, 2) NOT NULL,
    discount DECIMAL(12, 2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_order_item_order FOREIGN KEY (order_id) 
        REFERENCES "order"(id) ON DELETE CASCADE,
    CONSTRAINT fk_order_item_product FOREIGN KEY (product_id) 
        REFERENCES product(id) ON DELETE RESTRICT,
    CONSTRAINT order_item_quantity_positive CHECK (quantity > 0),
    CONSTRAINT order_item_unit_price_positive CHECK (unit_price > 0),
    CONSTRAINT order_item_discount_non_negative CHECK (discount >= 0),
    CONSTRAINT order_item_unique_product_per_order UNIQUE (order_id, product_id)
);

CREATE INDEX idx_order_item_order_id ON order_item(order_id);
CREATE INDEX idx_order_item_product_id ON order_item(product_id);

-- ============================================================================
-- AUDIT TRIGGER - Auto-update updated_at timestamp
-- ============================================================================
CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_product_update BEFORE UPDATE ON product
    FOR EACH ROW EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER trigger_branch_update BEFORE UPDATE ON branch
    FOR EACH ROW EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER trigger_inventory_update BEFORE UPDATE ON inventory
    FOR EACH ROW EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER trigger_order_update BEFORE UPDATE ON "order"
    FOR EACH ROW EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER trigger_order_item_update BEFORE UPDATE ON order_item
    FOR EACH ROW EXECUTE FUNCTION update_timestamp();

-- ============================================================================
-- SAMPLE DATA (Optional - for testing)
-- ============================================================================
-- Uncomment below to insert sample data

/*
INSERT INTO branch (name, address, phone) VALUES
    ('Bangkok Store', '123 Silom Rd, Bangkok', '02-123-4567'),
    ('Chiang Mai Store', '456 Nimman Rd, Chiang Mai', '053-234-5678'),
    ('Phuket Store', '789 Patong Beach, Phuket', '076-345-6789');

INSERT INTO product (sku, name, description, price, cost, is_active) VALUES
    ('SKU-001', 'iPhone 15 Pro', 'Latest Apple smartphone', 35999.00, 28000.00, TRUE),
    ('SKU-002', 'Samsung Galaxy S24', 'Premium Android Phone', 32999.00, 25000.00, TRUE),
    ('SKU-003', 'USB-C Cable', '2m white cable', 299.00, 80.00, TRUE);

INSERT INTO inventory (product_id, branch_id, quantity, minimum_qty) VALUES
    (1, 1, 15, 5),
    (1, 2, 8, 5),
    (2, 1, 10, 5),
    (3, 1, 50, 20),
    (3, 2, 30, 20);
*/

-- ============================================================================
-- SCHEMA SUMMARY
-- ============================================================================
-- Tables: 5 (product, branch, inventory, order, order_item)
-- Primary Keys: All tables
-- Foreign Keys: inventory(product_id, branch_id), order(branch_id), order_item(order_id, product_id)
-- Unique Constraints: product(sku), inventory(product_id, branch_id), order_item(order_id, product_id)
-- Check Constraints: price/amount positive, quantities non-negative, status enum
-- Indexes: 17 total (optimized for common queries)
-- Audit Trail: Auto-update timestamps via triggers
-- ============================================================================
