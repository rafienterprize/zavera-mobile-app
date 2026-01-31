-- ============================================
-- PHASE 2 MIGRATION: Order & Inventory Robust
-- Run this after initial schema.sql
-- ============================================

-- Add EXPIRED status to order_status enum if not exists
DO $$ 
BEGIN
    -- Check if EXPIRED already exists in the enum
    IF NOT EXISTS (
        SELECT 1 FROM pg_enum 
        WHERE enumlabel = 'EXPIRED' 
        AND enumtypid = (SELECT oid FROM pg_type WHERE typname = 'order_status')
    ) THEN
        ALTER TYPE order_status ADD VALUE 'EXPIRED' AFTER 'FAILED';
    END IF;
END $$;

-- Add COMPLETED status to order_status enum if not exists
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_enum 
        WHERE enumlabel = 'COMPLETED' 
        AND enumtypid = (SELECT oid FROM pg_type WHERE typname = 'order_status')
    ) THEN
        ALTER TYPE order_status ADD VALUE 'COMPLETED' AFTER 'DELIVERED';
    END IF;
END $$;

-- Add new timestamp columns to orders table
ALTER TABLE orders 
ADD COLUMN IF NOT EXISTS shipped_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS completed_at TIMESTAMP;

-- Add stock_reserved column to track if stock was reserved for this order
ALTER TABLE orders
ADD COLUMN IF NOT EXISTS stock_reserved BOOLEAN DEFAULT true;

-- Add index for stock_reserved orders (for cleanup jobs)
CREATE INDEX IF NOT EXISTS idx_orders_stock_reserved ON orders(stock_reserved) WHERE stock_reserved = true;

-- Add index for pending orders older than X (for expiry job)
CREATE INDEX IF NOT EXISTS idx_orders_pending_created ON orders(created_at) WHERE status = 'PENDING';

-- Add processed_at to track when webhook was processed (for idempotency)
ALTER TABLE payments
ADD COLUMN IF NOT EXISTS processed_at TIMESTAMP;

-- Create order_status_history table for audit trail
CREATE TABLE IF NOT EXISTS order_status_history (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    from_status VARCHAR(50),
    to_status VARCHAR(50) NOT NULL,
    changed_by VARCHAR(100), -- 'system', 'webhook', 'admin', etc.
    reason TEXT,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_order_status_history_order ON order_status_history(order_id);
CREATE INDEX IF NOT EXISTS idx_order_status_history_created ON order_status_history(created_at DESC);

-- ============================================
-- COMMENTS FOR DOCUMENTATION
-- ============================================
COMMENT ON COLUMN orders.stock_reserved IS 'Indicates if stock is currently reserved for this order. Set to false after stock is restored.';
COMMENT ON COLUMN orders.shipped_at IS 'Timestamp when order was shipped';
COMMENT ON COLUMN orders.completed_at IS 'Timestamp when order was marked as completed/delivered';
COMMENT ON COLUMN payments.processed_at IS 'Timestamp when webhook notification was processed (for idempotency)';
COMMENT ON TABLE order_status_history IS 'Audit trail for order status changes';
