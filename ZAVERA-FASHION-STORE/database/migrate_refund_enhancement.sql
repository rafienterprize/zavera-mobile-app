-- ============================================
-- REFUND SYSTEM ENHANCEMENT MIGRATION
-- ZAVERA E-Commerce Refund System Improvements
-- ============================================
-- This migration enhances the refund system to:
-- 1. Support manual refunds (orders without payment records)
-- 2. Improve database integrity and constraint handling
-- 3. Add performance indexes
-- 4. Ensure proper audit trail support
-- ============================================

-- ============================================
-- MAKE FOREIGN KEYS NULLABLE IN REFUNDS TABLE
-- ============================================
-- Allow refunds for orders without payment records (manual orders)
-- Allow refunds without a specific requesting user (system-initiated)

ALTER TABLE refunds 
  ALTER COLUMN requested_by DROP NOT NULL,
  ALTER COLUMN payment_id DROP NOT NULL;

COMMENT ON COLUMN refunds.requested_by IS 'User who requested the refund - NULL for system-initiated refunds';
COMMENT ON COLUMN refunds.payment_id IS 'Payment record reference - NULL for manual refunds (orders marked paid manually)';

-- ============================================
-- ADD PERFORMANCE INDEXES
-- ============================================
-- These indexes improve query performance for common refund operations

-- Index for idempotency key lookups (only non-null values)
CREATE INDEX IF NOT EXISTS idx_refunds_idempotency_key 
  ON refunds(idempotency_key) 
  WHERE idempotency_key IS NOT NULL;

-- Index for order_id lookups (frequently used for refund history)
CREATE INDEX IF NOT EXISTS idx_refunds_order_id 
  ON refunds(order_id);

-- Index for refund status queries (used in admin dashboard)
CREATE INDEX IF NOT EXISTS idx_refunds_status 
  ON refunds(status);

-- Index for orders with refund status (used in customer portal)
CREATE INDEX IF NOT EXISTS idx_orders_refund_status 
  ON orders(refund_status) 
  WHERE refund_status IS NOT NULL;

-- Index for refund status history lookups
CREATE INDEX IF NOT EXISTS idx_refund_status_history_refund_id 
  ON refund_status_history(refund_id);

-- Index for refund items by refund_id
CREATE INDEX IF NOT EXISTS idx_refund_items_refund_id 
  ON refund_items(refund_id);

-- ============================================
-- VERIFY REQUIRED COLUMNS EXIST
-- ============================================
-- These columns should already exist from migrate_hardening.sql
-- This section verifies they exist and adds them if missing

-- Verify orders table has refund tracking columns
ALTER TABLE orders ADD COLUMN IF NOT EXISTS refund_status VARCHAR(50);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS refund_amount DECIMAL(12, 2) DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS refunded_at TIMESTAMP;

-- ============================================
-- VERIFY REFUND STATUS HISTORY TABLE EXISTS
-- ============================================
-- This table should already exist from migrate_hardening.sql
-- This section verifies it exists and creates it if missing

CREATE TABLE IF NOT EXISTS refund_status_history (
    id SERIAL PRIMARY KEY,
    refund_id INTEGER NOT NULL REFERENCES refunds(id) ON DELETE CASCADE,
    from_status VARCHAR(20),
    to_status VARCHAR(20) NOT NULL,
    changed_by VARCHAR(100) NOT NULL,
    reason TEXT,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Ensure index exists
CREATE INDEX IF NOT EXISTS idx_refund_status_history_refund_id 
  ON refund_status_history(refund_id);

COMMENT ON TABLE refund_status_history IS 'Audit trail of all refund status changes';
COMMENT ON COLUMN refund_status_history.changed_by IS 'Actor who changed the status (user email or "SYSTEM")';
COMMENT ON COLUMN refund_status_history.from_status IS 'Previous status - NULL for initial status';
COMMENT ON COLUMN refund_status_history.to_status IS 'New status after change';

-- ============================================
-- ADD CONSTRAINTS FOR DATA INTEGRITY
-- ============================================

-- Ensure refund_amount on orders is non-negative
-- Note: This constraint may already exist, so we ignore errors
DO $$
BEGIN
    ALTER TABLE orders 
    ADD CONSTRAINT chk_orders_refund_amount_non_negative 
    CHECK (refund_amount >= 0);
EXCEPTION
    WHEN duplicate_object THEN
        NULL; -- Constraint already exists, ignore
END $$;

-- ============================================
-- UPDATE COMMENTS FOR CLARITY
-- ============================================

COMMENT ON COLUMN orders.refund_status IS 'Refund status: FULL (fully refunded) or PARTIAL (partially refunded)';
COMMENT ON COLUMN orders.refund_amount IS 'Total amount refunded for this order (sum of all completed refunds)';
COMMENT ON COLUMN orders.refunded_at IS 'Timestamp when the first refund was completed';

COMMENT ON COLUMN refunds.gateway_refund_id IS 'Refund ID from payment gateway (Midtrans) - "MANUAL_REFUND" for manual refunds';
COMMENT ON COLUMN refunds.idempotency_key IS 'Unique key to prevent duplicate refund processing - used for retry safety';

-- ============================================
-- MIGRATION COMPLETE
-- ============================================

SELECT 'Refund System Enhancement Migration completed successfully!' AS status;
SELECT 'Key changes:' AS summary;
SELECT '  1. Made refunds.requested_by nullable (supports system-initiated refunds)' AS change_1;
SELECT '  2. Made refunds.payment_id nullable (supports manual refunds)' AS change_2;
SELECT '  3. Added performance indexes for refund queries' AS change_3;
SELECT '  4. Verified refund tracking columns exist on orders table' AS change_4;
SELECT '  5. Verified refund_status_history table exists for audit trail' AS change_5;
