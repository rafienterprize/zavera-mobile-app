-- ============================================
-- COMMERCIAL HARDENING MIGRATION - PHASE 1
-- ZAVERA E-Commerce Critical Safety Layer
-- ============================================
-- This migration adds:
-- 1. Refund system tables
-- 2. Admin audit logging
-- 3. Payment sync/reconciliation tables
-- 4. New columns for orders, payments, shipments
-- ============================================

-- ============================================
-- REFUND STATUS ENUM
-- ============================================
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'refund_status') THEN
        CREATE TYPE refund_status AS ENUM (
            'PENDING',          -- Refund requested, awaiting processing
            'PROCESSING',       -- Refund being processed with payment gateway
            'PARTIAL',          -- Partial refund completed
            'COMPLETED',        -- Full refund completed
            'FAILED',           -- Refund failed
            'REJECTED',         -- Refund rejected by admin
            'CANCELLED'         -- Refund cancelled
        );
    END IF;
END $$;

-- ============================================
-- REFUND TYPE ENUM
-- ============================================
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'refund_type') THEN
        CREATE TYPE refund_type AS ENUM (
            'FULL',             -- Full order refund
            'PARTIAL',          -- Partial item refund
            'SHIPPING_ONLY',    -- Shipping cost refund only
            'ITEM_ONLY'         -- Specific items refund
        );
    END IF;
END $$;

-- ============================================
-- REFUND REASON ENUM
-- ============================================
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'refund_reason') THEN
        CREATE TYPE refund_reason AS ENUM (
            'CUSTOMER_REQUEST',     -- Customer requested cancellation
            'OUT_OF_STOCK',         -- Item out of stock after order
            'DAMAGED_ITEM',         -- Item damaged during shipping
            'WRONG_ITEM',           -- Wrong item shipped
            'LATE_DELIVERY',        -- Delivery too late
            'DUPLICATE_ORDER',      -- Duplicate order placed
            'FRAUD_SUSPECTED',      -- Suspected fraudulent order
            'ADMIN_DECISION',       -- Admin decided to refund
            'SHIPPING_FAILED',      -- Shipping failed/returned
            'OTHER'                 -- Other reason
        );
    END IF;
END $$;

-- ============================================
-- ADMIN ACTION TYPE ENUM
-- ============================================
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'admin_action_type') THEN
        CREATE TYPE admin_action_type AS ENUM (
            'FORCE_CANCEL',         -- Force cancel order
            'FORCE_REFUND',         -- Force refund
            'FORCE_RESHIP',         -- Force reship order
            'RECONCILE_PAYMENT',    -- Reconcile payment manually
            'UPDATE_STATUS',        -- Manual status update
            'RESTORE_STOCK',        -- Manual stock restoration
            'VOID_REFUND',          -- Void a refund
            'OVERRIDE_PAYMENT',     -- Override payment status
            'MANUAL_ADJUSTMENT'     -- Manual adjustment
        );
    END IF;
END $$;

-- ============================================
-- PAYMENT SYNC STATUS ENUM
-- ============================================
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'payment_sync_status') THEN
        CREATE TYPE payment_sync_status AS ENUM (
            'PENDING',          -- Sync pending
            'IN_PROGRESS',      -- Sync in progress
            'SYNCED',           -- Successfully synced
            'MISMATCH',         -- Status mismatch detected
            'FAILED',           -- Sync failed
            'RESOLVED'          -- Mismatch resolved
        );
    END IF;
END $$;

-- ============================================
-- REFUNDS TABLE
-- Master table for all refund requests
-- ============================================
CREATE TABLE IF NOT EXISTS refunds (
    id SERIAL PRIMARY KEY,
    
    -- Reference
    refund_code VARCHAR(50) UNIQUE NOT NULL,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE RESTRICT,
    payment_id INTEGER REFERENCES payments(id) ON DELETE RESTRICT,
    
    -- Refund details
    refund_type refund_type NOT NULL,
    reason refund_reason NOT NULL,
    reason_detail TEXT,
    
    -- Amounts
    original_amount DECIMAL(12, 2) NOT NULL,       -- Original payment amount
    refund_amount DECIMAL(12, 2) NOT NULL,         -- Amount to refund
    shipping_refund DECIMAL(12, 2) DEFAULT 0,      -- Shipping portion
    items_refund DECIMAL(12, 2) DEFAULT 0,         -- Items portion
    
    -- Status
    status refund_status DEFAULT 'PENDING',
    
    -- Gateway response
    gateway_refund_id VARCHAR(255),                -- Midtrans refund ID
    gateway_status VARCHAR(100),
    gateway_response JSONB,
    
    -- Idempotency
    idempotency_key VARCHAR(100) UNIQUE,
    
    -- Processing
    processed_by INTEGER REFERENCES users(id),
    processed_at TIMESTAMP,
    
    -- Audit
    requested_by INTEGER REFERENCES users(id),
    requested_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    
    -- Constraints
    CONSTRAINT chk_refund_amounts CHECK (
        refund_amount > 0 AND 
        refund_amount <= original_amount AND
        shipping_refund >= 0 AND
        items_refund >= 0
    )
);

CREATE INDEX IF NOT EXISTS idx_refunds_order ON refunds(order_id);
CREATE INDEX IF NOT EXISTS idx_refunds_payment ON refunds(payment_id);
CREATE INDEX IF NOT EXISTS idx_refunds_status ON refunds(status);
CREATE INDEX IF NOT EXISTS idx_refunds_code ON refunds(refund_code);
CREATE INDEX IF NOT EXISTS idx_refunds_idempotency ON refunds(idempotency_key);
CREATE INDEX IF NOT EXISTS idx_refunds_gateway ON refunds(gateway_refund_id);

-- ============================================
-- REFUND ITEMS TABLE
-- Individual items in a refund (for partial refunds)
-- ============================================
CREATE TABLE IF NOT EXISTS refund_items (
    id SERIAL PRIMARY KEY,
    refund_id INTEGER NOT NULL REFERENCES refunds(id) ON DELETE CASCADE,
    order_item_id INTEGER NOT NULL REFERENCES order_items(id) ON DELETE RESTRICT,
    
    -- Item details (snapshot)
    product_id INTEGER NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    quantity INTEGER NOT NULL,
    price_per_unit DECIMAL(12, 2) NOT NULL,
    refund_amount DECIMAL(12, 2) NOT NULL,
    
    -- Reason for this specific item
    item_reason TEXT,
    
    -- Stock restoration
    stock_restored BOOLEAN DEFAULT false,
    stock_restored_at TIMESTAMP,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT chk_refund_item_quantity CHECK (quantity > 0),
    CONSTRAINT chk_refund_item_amount CHECK (refund_amount >= 0)
);

CREATE INDEX IF NOT EXISTS idx_refund_items_refund ON refund_items(refund_id);
CREATE INDEX IF NOT EXISTS idx_refund_items_order_item ON refund_items(order_item_id);

-- ============================================
-- ADMIN AUDIT LOG TABLE
-- Immutable log of all admin actions
-- ============================================
CREATE TABLE IF NOT EXISTS admin_audit_log (
    id SERIAL PRIMARY KEY,
    
    -- Who
    admin_user_id INTEGER NOT NULL REFERENCES users(id),
    admin_email VARCHAR(255) NOT NULL,
    admin_ip VARCHAR(50),
    admin_user_agent TEXT,
    
    -- What
    action_type admin_action_type NOT NULL,
    action_detail TEXT NOT NULL,
    
    -- Target
    target_type VARCHAR(50) NOT NULL,              -- 'order', 'payment', 'refund', 'shipment'
    target_id INTEGER NOT NULL,
    target_code VARCHAR(100),                      -- order_code, refund_code, etc.
    
    -- Before/After state
    state_before JSONB,
    state_after JSONB,
    
    -- Result
    success BOOLEAN NOT NULL,
    error_message TEXT,
    
    -- Idempotency
    idempotency_key VARCHAR(100),
    
    -- Metadata
    metadata JSONB,
    
    -- Timestamp (immutable)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    
    -- This table is append-only, no updates allowed
    CONSTRAINT admin_audit_immutable CHECK (true)
);

CREATE INDEX IF NOT EXISTS idx_admin_audit_user ON admin_audit_log(admin_user_id);
CREATE INDEX IF NOT EXISTS idx_admin_audit_action ON admin_audit_log(action_type);
CREATE INDEX IF NOT EXISTS idx_admin_audit_target ON admin_audit_log(target_type, target_id);
CREATE INDEX IF NOT EXISTS idx_admin_audit_created ON admin_audit_log(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_admin_audit_idempotency ON admin_audit_log(idempotency_key);

-- ============================================
-- PAYMENT SYNC LOG TABLE
-- Tracks payment status sync with gateway
-- ============================================
CREATE TABLE IF NOT EXISTS payment_sync_log (
    id SERIAL PRIMARY KEY,
    
    -- Reference
    payment_id INTEGER NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    order_code VARCHAR(100) NOT NULL,
    
    -- Sync details
    sync_type VARCHAR(50) NOT NULL,                -- 'scheduled', 'manual', 'webhook', 'recovery'
    sync_status payment_sync_status DEFAULT 'PENDING',
    
    -- Local state
    local_payment_status VARCHAR(50),
    local_order_status VARCHAR(50),
    
    -- Gateway state
    gateway_status VARCHAR(50),
    gateway_transaction_id VARCHAR(255),
    gateway_response JSONB,
    
    -- Mismatch details
    has_mismatch BOOLEAN DEFAULT false,
    mismatch_type VARCHAR(100),
    mismatch_detail TEXT,
    
    -- Resolution
    resolved BOOLEAN DEFAULT false,
    resolved_by INTEGER REFERENCES users(id),
    resolved_at TIMESTAMP,
    resolution_action TEXT,
    
    -- Retry tracking
    retry_count INTEGER DEFAULT 0,
    last_retry_at TIMESTAMP,
    next_retry_at TIMESTAMP,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_payment_sync_payment ON payment_sync_log(payment_id);
CREATE INDEX IF NOT EXISTS idx_payment_sync_order ON payment_sync_log(order_id);
CREATE INDEX IF NOT EXISTS idx_payment_sync_status ON payment_sync_log(sync_status);
CREATE INDEX IF NOT EXISTS idx_payment_sync_mismatch ON payment_sync_log(has_mismatch) WHERE has_mismatch = true;
CREATE INDEX IF NOT EXISTS idx_payment_sync_retry ON payment_sync_log(next_retry_at) WHERE resolved = false;

-- ============================================
-- RECONCILIATION LOG TABLE
-- Daily reconciliation records
-- ============================================
CREATE TABLE IF NOT EXISTS reconciliation_log (
    id SERIAL PRIMARY KEY,
    
    -- Period
    reconciliation_date DATE NOT NULL,
    period_start TIMESTAMP NOT NULL,
    period_end TIMESTAMP NOT NULL,
    
    -- Summary
    total_orders INTEGER DEFAULT 0,
    total_payments INTEGER DEFAULT 0,
    total_amount DECIMAL(14, 2) DEFAULT 0,
    
    -- Status counts
    orders_pending INTEGER DEFAULT 0,
    orders_paid INTEGER DEFAULT 0,
    orders_cancelled INTEGER DEFAULT 0,
    orders_refunded INTEGER DEFAULT 0,
    
    payments_pending INTEGER DEFAULT 0,
    payments_success INTEGER DEFAULT 0,
    payments_failed INTEGER DEFAULT 0,
    
    -- Mismatches
    mismatches_found INTEGER DEFAULT 0,
    mismatches_resolved INTEGER DEFAULT 0,
    mismatch_details JSONB,
    
    -- Orphans
    orphan_orders INTEGER DEFAULT 0,              -- Orders without payments
    orphan_payments INTEGER DEFAULT 0,            -- Payments without orders
    orphan_details JSONB,
    
    -- Stuck payments
    stuck_payments INTEGER DEFAULT 0,
    stuck_payment_ids INTEGER[],
    
    -- Financial summary
    expected_revenue DECIMAL(14, 2) DEFAULT 0,
    actual_revenue DECIMAL(14, 2) DEFAULT 0,
    revenue_variance DECIMAL(14, 2) DEFAULT 0,
    
    total_refunds DECIMAL(14, 2) DEFAULT 0,
    
    -- Status
    status VARCHAR(50) DEFAULT 'PENDING',         -- PENDING, RUNNING, COMPLETED, FAILED
    
    -- Execution
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    run_by VARCHAR(100),                          -- 'cron', 'manual', admin email
    
    -- Errors
    error_count INTEGER DEFAULT 0,
    errors JSONB,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT unique_reconciliation_date UNIQUE (reconciliation_date)
);

CREATE INDEX IF NOT EXISTS idx_reconciliation_date ON reconciliation_log(reconciliation_date DESC);
CREATE INDEX IF NOT EXISTS idx_reconciliation_status ON reconciliation_log(status);
CREATE INDEX IF NOT EXISTS idx_reconciliation_mismatches ON reconciliation_log(mismatches_found) WHERE mismatches_found > 0;

-- ============================================
-- ADD NEW COLUMNS TO ORDERS TABLE
-- ============================================
ALTER TABLE orders ADD COLUMN IF NOT EXISTS refund_status VARCHAR(50);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS refund_amount DECIMAL(12, 2) DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS refunded_at TIMESTAMP;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS is_refundable BOOLEAN DEFAULT true;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS refund_deadline TIMESTAMP;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS last_synced_at TIMESTAMP;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS sync_status VARCHAR(50);

-- ============================================
-- ADD NEW COLUMNS TO PAYMENTS TABLE
-- ============================================
ALTER TABLE payments ADD COLUMN IF NOT EXISTS refund_status VARCHAR(50);
ALTER TABLE payments ADD COLUMN IF NOT EXISTS refunded_amount DECIMAL(12, 2) DEFAULT 0;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS refundable_amount DECIMAL(12, 2);
ALTER TABLE payments ADD COLUMN IF NOT EXISTS last_synced_at TIMESTAMP;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS sync_status VARCHAR(50);
ALTER TABLE payments ADD COLUMN IF NOT EXISTS gateway_status VARCHAR(100);
ALTER TABLE payments ADD COLUMN IF NOT EXISTS settlement_time TIMESTAMP;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS is_reconciled BOOLEAN DEFAULT false;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS reconciled_at TIMESTAMP;

-- ============================================
-- ADD NEW COLUMNS TO SHIPMENTS TABLE
-- ============================================
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS reship_count INTEGER DEFAULT 0;
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS original_shipment_id INTEGER REFERENCES shipments(id);
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS reship_reason TEXT;
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS is_replacement BOOLEAN DEFAULT false;

-- ============================================
-- ORDER STATUS HISTORY TABLE (if not exists)
-- ============================================
CREATE TABLE IF NOT EXISTS order_status_history (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    from_status VARCHAR(50),
    to_status VARCHAR(50) NOT NULL,
    changed_by VARCHAR(100),
    reason TEXT,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_order_status_history_order ON order_status_history(order_id);
CREATE INDEX IF NOT EXISTS idx_order_status_history_created ON order_status_history(created_at DESC);

-- ============================================
-- REFUND HISTORY TABLE
-- Track all refund status changes
-- ============================================
CREATE TABLE IF NOT EXISTS refund_status_history (
    id SERIAL PRIMARY KEY,
    refund_id INTEGER NOT NULL REFERENCES refunds(id) ON DELETE CASCADE,
    from_status refund_status,
    to_status refund_status NOT NULL,
    changed_by VARCHAR(100),
    reason TEXT,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_refund_status_history_refund ON refund_status_history(refund_id);

-- ============================================
-- UPDATE TRIGGERS
-- ============================================
DROP TRIGGER IF EXISTS update_refunds_updated_at ON refunds;
CREATE TRIGGER update_refunds_updated_at 
    BEFORE UPDATE ON refunds 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_payment_sync_log_updated_at ON payment_sync_log;
CREATE TRIGGER update_payment_sync_log_updated_at 
    BEFORE UPDATE ON payment_sync_log 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_reconciliation_log_updated_at ON reconciliation_log;
CREATE TRIGGER update_reconciliation_log_updated_at 
    BEFORE UPDATE ON reconciliation_log 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- PREVENT ADMIN AUDIT LOG UPDATES
-- ============================================
CREATE OR REPLACE FUNCTION prevent_audit_update()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'admin_audit_log is immutable - updates not allowed';
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS prevent_admin_audit_update ON admin_audit_log;
CREATE TRIGGER prevent_admin_audit_update
    BEFORE UPDATE ON admin_audit_log
    FOR EACH ROW
    EXECUTE FUNCTION prevent_audit_update();

-- ============================================
-- COMMENTS
-- ============================================
COMMENT ON TABLE refunds IS 'Master table for all refund requests and processing';
COMMENT ON TABLE refund_items IS 'Individual items included in partial refunds';
COMMENT ON TABLE admin_audit_log IS 'Immutable audit log of all admin actions - append only';
COMMENT ON TABLE payment_sync_log IS 'Payment status sync tracking with payment gateway';
COMMENT ON TABLE reconciliation_log IS 'Daily reconciliation records and summaries';
COMMENT ON COLUMN refunds.idempotency_key IS 'Unique key to prevent duplicate refund processing';
COMMENT ON COLUMN admin_audit_log.state_before IS 'JSON snapshot of entity state before action';
COMMENT ON COLUMN admin_audit_log.state_after IS 'JSON snapshot of entity state after action';
COMMENT ON COLUMN payment_sync_log.has_mismatch IS 'True if local and gateway status do not match';
COMMENT ON COLUMN reconciliation_log.stuck_payments IS 'Count of payments stuck in pending state';

-- ============================================
-- GRANT PERMISSIONS (adjust as needed)
-- ============================================
-- GRANT SELECT, INSERT ON admin_audit_log TO app_user;
-- GRANT SELECT, INSERT, UPDATE ON refunds TO app_user;
-- GRANT SELECT, INSERT, UPDATE ON payment_sync_log TO app_user;
-- GRANT SELECT, INSERT, UPDATE ON reconciliation_log TO app_user;

SELECT 'Commercial Hardening Migration Phase 1 completed successfully' AS status;
