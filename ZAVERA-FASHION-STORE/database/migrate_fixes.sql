-- ============================================
-- ZAVERA DATABASE FIXES & CLEANUP
-- Production-ready fixes for identified issues
-- ============================================

-- ============================================
-- 1. ADD MISSING INDEXES
-- ============================================

-- Index for user orders (was missing in some migrations)
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);

-- Index for cart cleanup (abandoned carts)
CREATE INDEX IF NOT EXISTS idx_carts_updated_at ON carts(updated_at);

-- Index for order expiry job
CREATE INDEX IF NOT EXISTS idx_orders_pending_status ON orders(status, created_at) WHERE status = 'PENDING';

-- ============================================
-- 2. ADD PRODUCT_IMAGE COLUMN TO ORDER_ITEMS
-- (This column is used in code but may be missing)
-- ============================================
ALTER TABLE order_items ADD COLUMN IF NOT EXISTS product_image VARCHAR(500) DEFAULT '';

-- ============================================
-- 3. ENSURE ALL ENUM VALUES EXIST
-- ============================================

-- Add DELIVERED status if missing (for shipment tracking)
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_enum 
        WHERE enumlabel = 'DELIVERED' 
        AND enumtypid = (SELECT oid FROM pg_type WHERE typname = 'order_status')
    ) THEN
        ALTER TYPE order_status ADD VALUE 'DELIVERED' AFTER 'SHIPPED';
    END IF;
END $$;

-- ============================================
-- 4. CREATE CLEANUP FUNCTIONS
-- ============================================

-- Function to cleanup abandoned carts (older than 30 days)
CREATE OR REPLACE FUNCTION cleanup_abandoned_carts()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM carts 
    WHERE updated_at < NOW() - INTERVAL '30 days'
    AND id NOT IN (
        SELECT DISTINCT c.id FROM carts c
        INNER JOIN orders o ON o.metadata->>'cart_id' = c.id::text
    );
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Function to expire pending orders (older than 24 hours)
CREATE OR REPLACE FUNCTION expire_pending_orders()
RETURNS INTEGER AS $$
DECLARE
    expired_count INTEGER;
    order_record RECORD;
BEGIN
    expired_count := 0;
    
    FOR order_record IN 
        SELECT id FROM orders 
        WHERE status = 'PENDING' 
        AND created_at < NOW() - INTERVAL '24 hours'
        AND stock_reserved = true
    LOOP
        -- Restore stock for each order item
        UPDATE products p
        SET stock = p.stock + oi.quantity
        FROM order_items oi
        WHERE oi.order_id = order_record.id
        AND p.id = oi.product_id;
        
        -- Mark order as expired
        UPDATE orders 
        SET status = 'EXPIRED', 
            stock_reserved = false,
            updated_at = NOW()
        WHERE id = order_record.id;
        
        -- Record status change
        INSERT INTO order_status_history (order_id, from_status, to_status, changed_by, reason)
        VALUES (order_record.id, 'PENDING', 'EXPIRED', 'system', 'Payment timeout - 24 hours');
        
        expired_count := expired_count + 1;
    END LOOP;
    
    RETURN expired_count;
END;
$$ LANGUAGE plpgsql;

-- Function to cleanup expired verification tokens
CREATE OR REPLACE FUNCTION cleanup_expired_tokens()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM email_verification_tokens 
    WHERE expires_at < NOW() OR used_at IS NOT NULL;
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    DELETE FROM password_reset_tokens 
    WHERE expires_at < NOW() OR used_at IS NOT NULL;
    
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- 5. VERIFY AND FIX DATA INTEGRITY
-- ============================================

-- Ensure all orders have stock_reserved column set correctly
UPDATE orders 
SET stock_reserved = false 
WHERE status IN ('CANCELLED', 'FAILED', 'EXPIRED', 'COMPLETED', 'DELIVERED')
AND stock_reserved = true;

-- Ensure all orders have stock_reserved = true for pending/paid orders
UPDATE orders 
SET stock_reserved = true 
WHERE status IN ('PENDING', 'PAID', 'PROCESSING', 'SHIPPED')
AND stock_reserved = false;

-- ============================================
-- 6. ADD CONSTRAINTS FOR DATA INTEGRITY
-- ============================================

-- Ensure user_addresses has proper user_id for authenticated users
-- (Guest addresses can have NULL user_id)

-- ============================================
-- 7. CREATE VIEW FOR MONITORING
-- ============================================

CREATE OR REPLACE VIEW v_order_summary AS
SELECT 
    DATE(created_at) as order_date,
    status,
    COUNT(*) as order_count,
    SUM(total_amount) as total_revenue,
    AVG(total_amount) as avg_order_value
FROM orders
GROUP BY DATE(created_at), status
ORDER BY order_date DESC, status;

CREATE OR REPLACE VIEW v_pending_orders_alert AS
SELECT 
    id,
    order_code,
    customer_email,
    total_amount,
    created_at,
    EXTRACT(EPOCH FROM (NOW() - created_at))/3600 as hours_pending
FROM orders
WHERE status = 'PENDING'
AND created_at < NOW() - INTERVAL '1 hour'
ORDER BY created_at ASC;

-- ============================================
-- 8. COMMENTS FOR DOCUMENTATION
-- ============================================

COMMENT ON FUNCTION cleanup_abandoned_carts() IS 'Removes carts not updated in 30 days. Run daily via cron.';
COMMENT ON FUNCTION expire_pending_orders() IS 'Expires pending orders older than 24 hours and restores stock. Run hourly via cron.';
COMMENT ON FUNCTION cleanup_expired_tokens() IS 'Removes expired verification and reset tokens. Run daily via cron.';
COMMENT ON VIEW v_order_summary IS 'Daily order summary by status for monitoring dashboard.';
COMMENT ON VIEW v_pending_orders_alert IS 'Pending orders older than 1 hour for alerting.';

-- ============================================
-- 9. GRANT PERMISSIONS (adjust as needed)
-- ============================================

-- Example: GRANT EXECUTE ON FUNCTION cleanup_abandoned_carts() TO zavera_app;

