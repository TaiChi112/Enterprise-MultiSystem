-- Patch: add missing foreign keys for OMS-related tables.
-- Safe to run multiple times: checks constraint existence first.

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_order_lifecycle_customer'
    ) THEN
        ALTER TABLE order_lifecycle
            ADD CONSTRAINT fk_order_lifecycle_customer
            FOREIGN KEY (customer_id)
            REFERENCES customer(id)
            ON DELETE RESTRICT;
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_order_item_oms_product'
    ) THEN
        ALTER TABLE order_item_oms
            ADD CONSTRAINT fk_order_item_oms_product
            FOREIGN KEY (product_id)
            REFERENCES product(id)
            ON DELETE RESTRICT;
    END IF;
END $$;
