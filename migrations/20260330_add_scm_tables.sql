-- Patch: add SCM tables for STEP 1 (supplier and purchase_order).
-- Safe to run multiple times.

CREATE TABLE IF NOT EXISTS supplier (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    contact VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT supplier_name_not_empty CHECK (LENGTH(TRIM(name)) > 0)
);

CREATE INDEX IF NOT EXISTS idx_supplier_name ON supplier(name);
CREATE INDEX IF NOT EXISTS idx_supplier_created_at ON supplier(created_at DESC);

CREATE TABLE IF NOT EXISTS purchase_order (
    id SERIAL PRIMARY KEY,
    supplier_id INTEGER NOT NULL,
    product_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT purchase_order_quantity_positive CHECK (quantity > 0),
    CONSTRAINT purchase_order_status_valid CHECK (status IN ('draft', 'approved', 'transmitted', 'cancelled'))
);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_purchase_order_supplier'
    ) THEN
        ALTER TABLE purchase_order
            ADD CONSTRAINT fk_purchase_order_supplier
            FOREIGN KEY (supplier_id)
            REFERENCES supplier(id)
            ON DELETE RESTRICT;
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_purchase_order_product'
    ) THEN
        ALTER TABLE purchase_order
            ADD CONSTRAINT fk_purchase_order_product
            FOREIGN KEY (product_id)
            REFERENCES product(id)
            ON DELETE RESTRICT;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_purchase_order_supplier_id ON purchase_order(supplier_id);
CREATE INDEX IF NOT EXISTS idx_purchase_order_product_id ON purchase_order(product_id);
CREATE INDEX IF NOT EXISTS idx_purchase_order_status ON purchase_order(status);
CREATE INDEX IF NOT EXISTS idx_purchase_order_created_at ON purchase_order(created_at DESC);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_trigger
        WHERE tgname = 'trigger_supplier_update'
    ) THEN
        CREATE TRIGGER trigger_supplier_update
        BEFORE UPDATE ON supplier
        FOR EACH ROW
        EXECUTE FUNCTION update_timestamp();
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_trigger
        WHERE tgname = 'trigger_purchase_order_update'
    ) THEN
        CREATE TRIGGER trigger_purchase_order_update
        BEFORE UPDATE ON purchase_order
        FOR EACH ROW
        EXECUTE FUNCTION update_timestamp();
    END IF;
END $$;
