-- ============================================
-- ZAVERA ENUM FIX MIGRATION
-- Fixes shipment_status enum to include all required values
-- Run this BEFORE applying other fixes
-- ============================================

-- Step 1: Add missing values to existing shipment_status enum
DO $$ 
BEGIN
    -- Add PICKUP_SCHEDULED if not exists
    IF NOT EXISTS (
        SELECT 1 FROM pg_enum 
        WHERE enumlabel = 'PICKUP_SCHEDULED' 
        AND enumtypid = (SELECT oid FROM pg_type WHERE typname = 'shipment_status')
    ) THEN
        ALTER TYPE shipment_status ADD VALUE 'PICKUP_SCHEDULED' AFTER 'PROCESSING';
    END IF;
    
    -- Add PICKUP_FAILED if not exists
    IF NOT EXISTS (
        SELECT 1 FROM pg_enum 
        WHERE enumlabel = 'PICKUP_FAILED' 
        AND enumtypid = (SELECT oid FROM pg_type WHERE typname = 'shipment_status')
    ) THEN
        ALTER TYPE shipment_status ADD VALUE 'PICKUP_FAILED' AFTER 'PICKUP_SCHEDULED';
    END IF;
    
    -- Add DELIVERY_FAILED if not exists
    IF NOT EXISTS (
        SELECT 1 FROM pg_enum 
        WHERE enumlabel = 'DELIVERY_FAILED' 
        AND enumtypid = (SELECT oid FROM pg_type WHERE typname = 'shipment_status')
    ) THEN
        ALTER TYPE shipment_status ADD VALUE 'DELIVERY_FAILED' AFTER 'DELIVERED';
    END IF;
    
    -- Add HELD_AT_WAREHOUSE if not exists
    IF NOT EXISTS (
        SELECT 1 FROM pg_enum 
        WHERE enumlabel = 'HELD_AT_WAREHOUSE' 
        AND enumtypid = (SELECT oid FROM pg_type WHERE typname = 'shipment_status')
    ) THEN
        ALTER TYPE shipment_status ADD VALUE 'HELD_AT_WAREHOUSE' AFTER 'DELIVERY_FAILED';
    END IF;
    
    -- Add RETURNED_TO_SENDER if not exists
    IF NOT EXISTS (
        SELECT 1 FROM pg_enum 
        WHERE enumlabel = 'RETURNED_TO_SENDER' 
        AND enumtypid = (SELECT oid FROM pg_type WHERE typname = 'shipment_status')
    ) THEN
        ALTER TYPE shipment_status ADD VALUE 'RETURNED_TO_SENDER' AFTER 'HELD_AT_WAREHOUSE';
    END IF;
    
    -- Add LOST if not exists
    IF NOT EXISTS (
        SELECT 1 FROM pg_enum 
        WHERE enumlabel = 'LOST' 
        AND enumtypid = (SELECT oid FROM pg_type WHERE typname = 'shipment_status')
    ) THEN
        ALTER TYPE shipment_status ADD VALUE 'LOST' AFTER 'RETURNED_TO_SENDER';
    END IF;
    
    -- Add INVESTIGATION if not exists
    IF NOT EXISTS (
        SELECT 1 FROM pg_enum 
        WHERE enumlabel = 'INVESTIGATION' 
        AND enumtypid = (SELECT oid FROM pg_type WHERE typname = 'shipment_status')
    ) THEN
        ALTER TYPE shipment_status ADD VALUE 'INVESTIGATION' AFTER 'LOST';
    END IF;
    
    -- Add REPLACED if not exists
    IF NOT EXISTS (
        SELECT 1 FROM pg_enum 
        WHERE enumlabel = 'REPLACED' 
        AND enumtypid = (SELECT oid FROM pg_type WHERE typname = 'shipment_status')
    ) THEN
        ALTER TYPE shipment_status ADD VALUE 'REPLACED' AFTER 'INVESTIGATION';
    END IF;
    
    -- Add CANCELLED if not exists
    IF NOT EXISTS (
        SELECT 1 FROM pg_enum 
        WHERE enumlabel = 'CANCELLED' 
        AND enumtypid = (SELECT oid FROM pg_type WHERE typname = 'shipment_status')
    ) THEN
        ALTER TYPE shipment_status ADD VALUE 'CANCELLED' AFTER 'REPLACED';
    END IF;
END $$;

-- Step 2: Drop the unused shipment_status_v2 type if it exists
DROP TYPE IF EXISTS shipment_status_v2;

-- Step 3: Add compound index for refund queries
CREATE INDEX IF NOT EXISTS idx_refunds_order_status ON refunds(order_id, status);

-- Step 4: Add constraint to prevent over-refund (advisory - app still needs to check)
-- Note: This is a soft constraint via trigger since PostgreSQL doesn't support 
-- cross-row CHECK constraints

CREATE OR REPLACE FUNCTION check_refund_amount()
RETURNS TRIGGER AS $$
DECLARE
    v_paid_amount DECIMAL(12,2);
    v_total_refunded DECIMAL(12,2);
BEGIN
    -- Get paid amount
    SELECT COALESCE(p.amount, 0) INTO v_paid_amount
    FROM payments p
    WHERE p.order_id = NEW.order_id AND p.status = 'SUCCESS'
    LIMIT 1;
    
    -- Get total refunded (excluding current if update)
    SELECT COALESCE(SUM(refund_amount), 0) INTO v_total_refunded
    FROM refunds
    WHERE order_id = NEW.order_id 
    AND status IN ('COMPLETED', 'PROCESSING', 'PENDING')
    AND id != COALESCE(NEW.id, 0);
    
    -- Check if over-refund
    IF (v_total_refunded + NEW.refund_amount) > v_paid_amount THEN
        RAISE EXCEPTION 'Refund amount would exceed paid amount. Paid: %, Already refunded: %, Requested: %',
            v_paid_amount, v_total_refunded, NEW.refund_amount;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_check_refund_amount ON refunds;
CREATE TRIGGER trg_check_refund_amount
    BEFORE INSERT OR UPDATE ON refunds
    FOR EACH ROW
    EXECUTE FUNCTION check_refund_amount();

SELECT 'Enum fix migration completed successfully' AS status;
