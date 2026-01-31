-- ============================================
-- SHIPPING & FULFILLMENT HARDENING - PHASE 2
-- ZAVERA E-Commerce Operational Control Layer
-- ============================================
-- This migration adds:
-- 1. Enhanced shipment state machine
-- 2. Pickup & courier control
-- 3. Lost/stuck shipment detection
-- 4. Dispute system
-- 5. Reship engine support
-- ============================================

-- ============================================
-- DROP OLD SHIPMENT STATUS ENUM AND CREATE NEW
-- ============================================
DO $$ 
BEGIN
    -- Create new enum with all statuses
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'shipment_status_v2') THEN
        CREATE TYPE shipment_status_v2 AS ENUM (
            'PENDING',              -- Waiting for payment/processing
            'PROCESSING',           -- Payment received, preparing package
            'PICKUP_SCHEDULED',     -- Courier pickup scheduled
            'PICKUP_FAILED',        -- Courier failed to pickup
            'SHIPPED',              -- Handed to courier
            'IN_TRANSIT',           -- On the way
            'OUT_FOR_DELIVERY',     -- Out for delivery
            'DELIVERED',            -- Successfully delivered
            'DELIVERY_FAILED',      -- Delivery attempt failed
            'HELD_AT_WAREHOUSE',    -- Held at courier warehouse
            'RETURNED_TO_SENDER',   -- Returned to sender
            'LOST',                 -- Package lost
            'INVESTIGATION',        -- Under investigation
            'REPLACED',             -- Replaced with new shipment
            'CANCELLED'             -- Cancelled
        );
    END IF;
END $$;

-- ============================================
-- DISPUTE STATUS ENUM
-- ============================================
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'dispute_status') THEN
        CREATE TYPE dispute_status AS ENUM (
            'OPEN',                 -- Dispute opened
            'INVESTIGATING',        -- Under investigation
            'EVIDENCE_REQUIRED',    -- Waiting for evidence
            'PENDING_RESOLUTION',   -- Waiting for resolution decision
            'RESOLVED_REFUND',      -- Resolved with refund
            'RESOLVED_RESHIP',      -- Resolved with reship
            'RESOLVED_REJECTED',    -- Dispute rejected
            'CLOSED'                -- Closed
        );
    END IF;
END $$;

-- ============================================
-- DISPUTE TYPE ENUM
-- ============================================
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'dispute_type') THEN
        CREATE TYPE dispute_type AS ENUM (
            'LOST_PACKAGE',         -- Package lost in transit
            'DAMAGED_PACKAGE',      -- Package damaged
            'WRONG_ITEM',           -- Wrong item received
            'MISSING_ITEM',         -- Item missing from package
            'NOT_DELIVERED',        -- Marked delivered but not received
            'LATE_DELIVERY',        -- Significantly late delivery
            'FAKE_DELIVERY',        -- Fake delivery confirmation
            'OTHER'                 -- Other issues
        );
    END IF;
END $$;

-- ============================================
-- ADD NEW COLUMNS TO SHIPMENTS TABLE
-- ============================================

-- Pickup control
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS pickup_scheduled_at TIMESTAMP;
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS pickup_deadline TIMESTAMP;
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS pickup_attempts INTEGER DEFAULT 0;
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS last_pickup_attempt_at TIMESTAMP;
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS pickup_notes TEXT;

-- Tracking control
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS last_tracking_update TIMESTAMP;
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS days_without_update INTEGER DEFAULT 0;
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS tracking_stale BOOLEAN DEFAULT false;

-- Investigation
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS investigation_opened_at TIMESTAMP;
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS investigation_reason TEXT;
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS marked_lost_at TIMESTAMP;
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS lost_reason TEXT;

-- Delivery control
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS delivery_attempts INTEGER DEFAULT 0;
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS last_delivery_attempt_at TIMESTAMP;
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS delivery_notes TEXT;
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS recipient_name_confirmed VARCHAR(255);
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS delivery_photo_url TEXT;

-- Reship tracking (already added in phase 1, ensure exists)
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS reship_count INTEGER DEFAULT 0;
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS original_shipment_id INTEGER REFERENCES shipments(id);
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS is_replacement BOOLEAN DEFAULT false;
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS reship_reason TEXT;
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS replaced_by_shipment_id INTEGER REFERENCES shipments(id);

-- Cost tracking for reships
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS reship_cost DECIMAL(12, 2) DEFAULT 0;
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS reship_cost_bearer VARCHAR(50); -- 'company' or 'customer'

-- Status metadata
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS status_metadata JSONB;
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS requires_admin_action BOOLEAN DEFAULT false;
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS admin_action_reason TEXT;

-- ============================================
-- SHIPMENT STATUS HISTORY TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS shipment_status_history (
    id SERIAL PRIMARY KEY,
    shipment_id INTEGER NOT NULL REFERENCES shipments(id) ON DELETE CASCADE,
    from_status VARCHAR(50),
    to_status VARCHAR(50) NOT NULL,
    changed_by VARCHAR(100),          -- 'system', 'webhook', 'admin:email', 'cron'
    reason TEXT,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_shipment_status_history_shipment ON shipment_status_history(shipment_id);
CREATE INDEX IF NOT EXISTS idx_shipment_status_history_created ON shipment_status_history(created_at DESC);

-- ============================================
-- COURIER FAILURE LOG TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS courier_failure_log (
    id SERIAL PRIMARY KEY,
    shipment_id INTEGER NOT NULL REFERENCES shipments(id) ON DELETE CASCADE,
    
    -- Failure details
    failure_type VARCHAR(50) NOT NULL,    -- 'pickup_failed', 'delivery_failed', 'lost', 'damaged'
    failure_reason TEXT,
    failure_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Courier info
    courier_code VARCHAR(50),
    courier_name VARCHAR(100),
    courier_tracking VARCHAR(100),
    
    -- Location
    failure_location VARCHAR(255),
    
    -- Resolution
    resolved BOOLEAN DEFAULT false,
    resolved_at TIMESTAMP,
    resolved_by VARCHAR(100),
    resolution_action TEXT,
    
    -- Evidence
    evidence_urls TEXT[],
    notes TEXT,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_courier_failure_shipment ON courier_failure_log(shipment_id);
CREATE INDEX IF NOT EXISTS idx_courier_failure_type ON courier_failure_log(failure_type);
CREATE INDEX IF NOT EXISTS idx_courier_failure_unresolved ON courier_failure_log(resolved) WHERE resolved = false;

-- ============================================
-- DISPUTES TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS disputes (
    id SERIAL PRIMARY KEY,
    dispute_code VARCHAR(50) UNIQUE NOT NULL,
    
    -- References
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE RESTRICT,
    shipment_id INTEGER REFERENCES shipments(id) ON DELETE RESTRICT,
    refund_id INTEGER REFERENCES refunds(id) ON DELETE RESTRICT,
    
    -- Dispute details
    dispute_type dispute_type NOT NULL,
    status dispute_status DEFAULT 'OPEN',
    
    -- Description
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    customer_claim TEXT,
    
    -- Customer info
    customer_user_id INTEGER REFERENCES users(id),
    customer_email VARCHAR(255) NOT NULL,
    customer_phone VARCHAR(50),
    
    -- Evidence
    evidence_urls TEXT[],
    customer_evidence_urls TEXT[],
    courier_evidence_urls TEXT[],
    
    -- Investigation
    investigation_notes TEXT,
    investigation_started_at TIMESTAMP,
    investigation_completed_at TIMESTAMP,
    investigator_id INTEGER REFERENCES users(id),
    
    -- Resolution
    resolution dispute_status,
    resolution_notes TEXT,
    resolution_amount DECIMAL(12, 2),
    resolved_by INTEGER REFERENCES users(id),
    resolved_at TIMESTAMP,
    
    -- Linked actions
    reship_shipment_id INTEGER REFERENCES shipments(id),
    
    -- Deadlines
    response_deadline TIMESTAMP,
    resolution_deadline TIMESTAMP,
    
    -- Metadata
    metadata JSONB,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT chk_dispute_resolution_amount CHECK (resolution_amount >= 0 OR resolution_amount IS NULL)
);

CREATE INDEX IF NOT EXISTS idx_disputes_order ON disputes(order_id);
CREATE INDEX IF NOT EXISTS idx_disputes_shipment ON disputes(shipment_id);
CREATE INDEX IF NOT EXISTS idx_disputes_status ON disputes(status);
CREATE INDEX IF NOT EXISTS idx_disputes_code ON disputes(dispute_code);
CREATE INDEX IF NOT EXISTS idx_disputes_type ON disputes(dispute_type);
CREATE INDEX IF NOT EXISTS idx_disputes_open ON disputes(status) WHERE status IN ('OPEN', 'INVESTIGATING', 'EVIDENCE_REQUIRED', 'PENDING_RESOLUTION');

-- ============================================
-- DISPUTE MESSAGES TABLE
-- For communication thread on disputes
-- ============================================
CREATE TABLE IF NOT EXISTS dispute_messages (
    id SERIAL PRIMARY KEY,
    dispute_id INTEGER NOT NULL REFERENCES disputes(id) ON DELETE CASCADE,
    
    -- Sender
    sender_type VARCHAR(20) NOT NULL,     -- 'customer', 'admin', 'system'
    sender_id INTEGER REFERENCES users(id),
    sender_name VARCHAR(255),
    
    -- Message
    message TEXT NOT NULL,
    attachment_urls TEXT[],
    
    -- Visibility
    is_internal BOOLEAN DEFAULT false,    -- Internal admin notes
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_dispute_messages_dispute ON dispute_messages(dispute_id);

-- ============================================
-- SHIPMENT ALERTS TABLE
-- For tracking alerts and notifications
-- ============================================
CREATE TABLE IF NOT EXISTS shipment_alerts (
    id SERIAL PRIMARY KEY,
    shipment_id INTEGER NOT NULL REFERENCES shipments(id) ON DELETE CASCADE,
    
    -- Alert details
    alert_type VARCHAR(50) NOT NULL,      -- 'stuck', 'lost', 'pickup_failed', 'delivery_failed'
    alert_level VARCHAR(20) NOT NULL,     -- 'warning', 'critical', 'urgent'
    title VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Status
    acknowledged BOOLEAN DEFAULT false,
    acknowledged_by INTEGER REFERENCES users(id),
    acknowledged_at TIMESTAMP,
    
    resolved BOOLEAN DEFAULT false,
    resolved_by INTEGER REFERENCES users(id),
    resolved_at TIMESTAMP,
    resolution_notes TEXT,
    
    -- Auto-action
    auto_action_taken BOOLEAN DEFAULT false,
    auto_action_type VARCHAR(50),
    auto_action_at TIMESTAMP,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_shipment_alerts_shipment ON shipment_alerts(shipment_id);
CREATE INDEX IF NOT EXISTS idx_shipment_alerts_type ON shipment_alerts(alert_type);
CREATE INDEX IF NOT EXISTS idx_shipment_alerts_unresolved ON shipment_alerts(resolved) WHERE resolved = false;
CREATE INDEX IF NOT EXISTS idx_shipment_alerts_level ON shipment_alerts(alert_level);

-- ============================================
-- VALID SHIPMENT TRANSITIONS TABLE
-- Defines allowed status transitions
-- ============================================
CREATE TABLE IF NOT EXISTS shipment_status_transitions (
    id SERIAL PRIMARY KEY,
    from_status VARCHAR(50) NOT NULL,
    to_status VARCHAR(50) NOT NULL,
    requires_admin BOOLEAN DEFAULT false,
    auto_allowed BOOLEAN DEFAULT true,
    description TEXT,
    UNIQUE(from_status, to_status)
);

-- Insert valid transitions
INSERT INTO shipment_status_transitions (from_status, to_status, requires_admin, auto_allowed, description) VALUES
-- From PENDING
('PENDING', 'PROCESSING', false, true, 'Payment received, start processing'),
('PENDING', 'CANCELLED', false, true, 'Order cancelled before processing'),

-- From PROCESSING
('PROCESSING', 'PICKUP_SCHEDULED', false, true, 'Pickup scheduled with courier'),
('PROCESSING', 'SHIPPED', false, true, 'Direct ship without pickup scheduling'),
('PROCESSING', 'CANCELLED', true, false, 'Admin cancellation during processing'),

-- From PICKUP_SCHEDULED
('PICKUP_SCHEDULED', 'SHIPPED', false, true, 'Courier picked up package'),
('PICKUP_SCHEDULED', 'PICKUP_FAILED', false, true, 'Courier failed to pickup'),
('PICKUP_SCHEDULED', 'CANCELLED', true, false, 'Admin cancellation'),

-- From PICKUP_FAILED
('PICKUP_FAILED', 'PICKUP_SCHEDULED', false, true, 'Reschedule pickup'),
('PICKUP_FAILED', 'CANCELLED', true, false, 'Cancel after pickup failures'),

-- From SHIPPED
('SHIPPED', 'IN_TRANSIT', false, true, 'Package in transit'),
('SHIPPED', 'DELIVERED', false, true, 'Direct delivery'),
('SHIPPED', 'LOST', true, false, 'Package lost'),
('SHIPPED', 'INVESTIGATION', true, false, 'Open investigation'),

-- From IN_TRANSIT
('IN_TRANSIT', 'OUT_FOR_DELIVERY', false, true, 'Out for delivery'),
('IN_TRANSIT', 'DELIVERED', false, true, 'Delivered'),
('IN_TRANSIT', 'HELD_AT_WAREHOUSE', false, true, 'Held at warehouse'),
('IN_TRANSIT', 'RETURNED_TO_SENDER', false, true, 'Returned'),
('IN_TRANSIT', 'LOST', true, false, 'Package lost'),
('IN_TRANSIT', 'INVESTIGATION', true, false, 'Open investigation'),

-- From OUT_FOR_DELIVERY
('OUT_FOR_DELIVERY', 'DELIVERED', false, true, 'Successfully delivered'),
('OUT_FOR_DELIVERY', 'DELIVERY_FAILED', false, true, 'Delivery failed'),
('OUT_FOR_DELIVERY', 'HELD_AT_WAREHOUSE', false, true, 'Returned to warehouse'),

-- From DELIVERY_FAILED
('DELIVERY_FAILED', 'OUT_FOR_DELIVERY', false, true, 'Retry delivery'),
('DELIVERY_FAILED', 'HELD_AT_WAREHOUSE', false, true, 'Hold at warehouse'),
('DELIVERY_FAILED', 'RETURNED_TO_SENDER', false, true, 'Return to sender'),

-- From HELD_AT_WAREHOUSE
('HELD_AT_WAREHOUSE', 'OUT_FOR_DELIVERY', false, true, 'Retry delivery'),
('HELD_AT_WAREHOUSE', 'RETURNED_TO_SENDER', false, true, 'Return to sender'),
('HELD_AT_WAREHOUSE', 'DELIVERED', false, true, 'Customer pickup'),

-- From RETURNED_TO_SENDER
('RETURNED_TO_SENDER', 'REPLACED', true, false, 'Create replacement shipment'),
('RETURNED_TO_SENDER', 'CANCELLED', true, false, 'Cancel and refund'),

-- From INVESTIGATION
('INVESTIGATION', 'LOST', true, false, 'Confirmed lost'),
('INVESTIGATION', 'DELIVERED', true, false, 'Confirmed delivered'),
('INVESTIGATION', 'IN_TRANSIT', true, false, 'Found, back in transit'),
('INVESTIGATION', 'REPLACED', true, false, 'Replace package'),

-- From LOST
('LOST', 'REPLACED', true, false, 'Create replacement'),

-- Terminal states (no transitions out except admin override)
('DELIVERED', 'INVESTIGATION', true, false, 'Customer disputes delivery'),
('CANCELLED', 'PROCESSING', true, false, 'Reactivate cancelled shipment')

ON CONFLICT (from_status, to_status) DO NOTHING;

-- ============================================
-- UPDATE TRIGGERS
-- ============================================
DROP TRIGGER IF EXISTS update_disputes_updated_at ON disputes;
CREATE TRIGGER update_disputes_updated_at 
    BEFORE UPDATE ON disputes 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- FUNCTION: Validate shipment status transition
-- ============================================
CREATE OR REPLACE FUNCTION validate_shipment_transition(
    p_from_status VARCHAR(50),
    p_to_status VARCHAR(50),
    p_is_admin BOOLEAN DEFAULT false
) RETURNS BOOLEAN AS $$
DECLARE
    v_valid BOOLEAN;
    v_requires_admin BOOLEAN;
BEGIN
    -- Check if transition exists
    SELECT EXISTS(
        SELECT 1 FROM shipment_status_transitions 
        WHERE from_status = p_from_status AND to_status = p_to_status
    ) INTO v_valid;
    
    IF NOT v_valid THEN
        RETURN false;
    END IF;
    
    -- Check if admin required
    SELECT requires_admin INTO v_requires_admin
    FROM shipment_status_transitions 
    WHERE from_status = p_from_status AND to_status = p_to_status;
    
    IF v_requires_admin AND NOT p_is_admin THEN
        RETURN false;
    END IF;
    
    RETURN true;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- FUNCTION: Calculate days without tracking update
-- ============================================
CREATE OR REPLACE FUNCTION update_shipment_tracking_staleness() RETURNS void AS $$
BEGIN
    UPDATE shipments
    SET 
        days_without_update = EXTRACT(DAY FROM (NOW() - COALESCE(last_tracking_update, shipped_at, created_at))),
        tracking_stale = EXTRACT(DAY FROM (NOW() - COALESCE(last_tracking_update, shipped_at, created_at))) > 3
    WHERE status IN ('SHIPPED', 'IN_TRANSIT', 'OUT_FOR_DELIVERY');
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- COMMENTS
-- ============================================
COMMENT ON TABLE shipment_status_history IS 'Audit trail of all shipment status changes';
COMMENT ON TABLE courier_failure_log IS 'Log of courier failures (pickup, delivery, lost)';
COMMENT ON TABLE disputes IS 'Customer disputes for orders/shipments';
COMMENT ON TABLE dispute_messages IS 'Communication thread for disputes';
COMMENT ON TABLE shipment_alerts IS 'Alerts for shipment issues requiring attention';
COMMENT ON TABLE shipment_status_transitions IS 'Valid shipment status transitions';
COMMENT ON COLUMN shipments.days_without_update IS 'Days since last tracking update';
COMMENT ON COLUMN shipments.tracking_stale IS 'True if no tracking update for >3 days';
COMMENT ON COLUMN shipments.requires_admin_action IS 'True if shipment needs admin intervention';

SELECT 'Shipping Hardening Migration Phase 2 completed successfully' AS status;
