-- Fix typo in column name: biteship_draft_order_idd -> biteship_draft_order_id
-- Run this if you see error: column "biteship_draft_order_idd" does not exist

-- Check if the typo column exists and rename it
DO $$
BEGIN
    -- If typo column exists, rename it
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'shipments' AND column_name = 'biteship_draft_order_idd'
    ) THEN
        ALTER TABLE shipments RENAME COLUMN biteship_draft_order_idd TO biteship_draft_order_id;
        RAISE NOTICE 'Renamed biteship_draft_order_idd to biteship_draft_order_id';
    END IF;
    
    -- If correct column doesn't exist, add it
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'shipments' AND column_name = 'biteship_draft_order_id'
    ) THEN
        ALTER TABLE shipments ADD COLUMN biteship_draft_order_id VARCHAR(100);
        RAISE NOTICE 'Added biteship_draft_order_id column';
    END IF;
END $$;

-- Ensure all Biteship columns exist
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS biteship_draft_order_id VARCHAR(100);
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS biteship_order_id VARCHAR(100);
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS biteship_tracking_id VARCHAR(100);
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS biteship_waybill_id VARCHAR(100);

-- Create indexes if not exist
CREATE INDEX IF NOT EXISTS idx_shipments_biteship_draft_order_id ON shipments(biteship_draft_order_id);
CREATE INDEX IF NOT EXISTS idx_shipments_biteship_tracking_id ON shipments(biteship_tracking_id);
CREATE INDEX IF NOT EXISTS idx_shipments_biteship_waybill_id ON shipments(biteship_waybill_id);
