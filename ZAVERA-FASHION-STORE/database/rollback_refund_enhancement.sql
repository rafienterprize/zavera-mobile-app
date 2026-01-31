-- ============================================
-- ROLLBACK: REFUND SYSTEM ENHANCEMENT MIGRATION
-- ZAVERA E-Commerce Refund System
-- ============================================
-- This script rolls back the refund enhancement migration
-- WARNING: Use with caution in production environments
-- ============================================

-- ============================================
-- RESTORE NOT NULL CONSTRAINTS
-- ============================================
-- Note: This will fail if there are existing NULL values
-- You must clean up NULL values before running this rollback

-- Check for NULL values before rollback
DO $$
DECLARE
    null_payment_count INTEGER;
    null_requested_by_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO null_payment_count
    FROM refunds
    WHERE payment_id IS NULL;

    SELECT COUNT(*) INTO null_requested_by_count
    FROM refunds
    WHERE requested_by IS NULL;

    IF null_payment_count > 0 THEN
        RAISE WARNING 'Found % refunds with NULL payment_id. These must be handled before rollback.', null_payment_count;
    END IF;

    IF null_requested_by_count > 0 THEN
        RAISE WARNING 'Found % refunds with NULL requested_by. These must be handled before rollback.', null_requested_by_count;
    END IF;

    IF null_payment_count > 0 OR null_requested_by_count > 0 THEN
        RAISE EXCEPTION 'Cannot rollback: NULL values exist in refunds table. Clean up data first.';
    END IF;
END $$;

-- Restore NOT NULL constraints
ALTER TABLE refunds 
  ALTER COLUMN requested_by SET NOT NULL,
  ALTER COLUMN payment_id SET NOT NULL;

-- ============================================
-- REMOVE PERFORMANCE INDEXES
-- ============================================
-- Note: Only remove indexes added by the enhancement migration
-- Keep indexes that existed before

DROP INDEX IF EXISTS idx_refunds_idempotency_key;
DROP INDEX IF EXISTS idx_refunds_order_id;
DROP INDEX IF EXISTS idx_orders_refund_status;
DROP INDEX IF EXISTS idx_refund_status_history_refund_id;
DROP INDEX IF EXISTS idx_refund_items_refund_id;

-- ============================================
-- REMOVE CONSTRAINT
-- ============================================

ALTER TABLE orders 
DROP CONSTRAINT IF EXISTS chk_orders_refund_amount_non_negative;

-- ============================================
-- REMOVE COMMENTS
-- ============================================

COMMENT ON COLUMN refunds.requested_by IS NULL;
COMMENT ON COLUMN refunds.payment_id IS NULL;
COMMENT ON COLUMN orders.refund_status IS NULL;
COMMENT ON COLUMN orders.refund_amount IS NULL;
COMMENT ON COLUMN orders.refunded_at IS NULL;
COMMENT ON COLUMN refunds.gateway_refund_id IS NULL;
COMMENT ON COLUMN refunds.idempotency_key IS NULL;

-- ============================================
-- ROLLBACK COMPLETE
-- ============================================

SELECT 'Refund System Enhancement Migration ROLLED BACK successfully!' AS status;
SELECT 'WARNING: The following changes were reverted:' AS warning;
SELECT '  1. Restored NOT NULL constraint on refunds.requested_by' AS change_1;
SELECT '  2. Restored NOT NULL constraint on refunds.payment_id' AS change_2;
SELECT '  3. Removed performance indexes' AS change_3;
SELECT '  4. Removed data integrity constraint' AS change_4;
SELECT '  5. Removed documentation comments' AS change_5;
SELECT 'NOTE: Orders table columns (refund_status, refund_amount, refunded_at) were NOT removed' AS note_1;
SELECT 'NOTE: refund_status_history table was NOT removed' AS note_2;
