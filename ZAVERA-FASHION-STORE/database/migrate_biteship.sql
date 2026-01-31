-- ============================================
-- BITESHIP MIGRATION SCRIPT
-- Migrates from RajaOngkir/Kommerce to Biteship
-- ============================================

-- ============================================
-- 1. CREATE BITESHIP_LOCATIONS TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS biteship_locations (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    location_id VARCHAR(100) NOT NULL,
    area_id VARCHAR(100) NOT NULL,
    area_name VARCHAR(500) NOT NULL,
    contact_name VARCHAR(255) NOT NULL,
    contact_phone VARCHAR(50) NOT NULL,
    address TEXT NOT NULL,
    postal_code VARCHAR(10),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for biteship_locations
CREATE INDEX IF NOT EXISTS idx_biteship_locations_user_id ON biteship_locations(user_id);
CREATE INDEX IF NOT EXISTS idx_biteship_locations_location_id ON biteship_locations(location_id);
CREATE INDEX IF NOT EXISTS idx_biteship_locations_area_id ON biteship_locations(area_id);

-- ============================================
-- 2. ADD BITESHIP COLUMNS TO SHIPMENTS TABLE
-- ============================================
ALTER TABLE shipments 
ADD COLUMN IF NOT EXISTS biteship_draft_order_id VARCHAR(100),
ADD COLUMN IF NOT EXISTS biteship_order_id VARCHAR(100),
ADD COLUMN IF NOT EXISTS biteship_tracking_id VARCHAR(100),
ADD COLUMN IF NOT EXISTS biteship_waybill_id VARCHAR(100);

-- Indexes for new shipments columns
CREATE INDEX IF NOT EXISTS idx_shipments_biteship_draft_order_id ON shipments(biteship_draft_order_id);
CREATE INDEX IF NOT EXISTS idx_shipments_biteship_tracking_id ON shipments(biteship_tracking_id);
CREATE INDEX IF NOT EXISTS idx_shipments_biteship_waybill_id ON shipments(biteship_waybill_id);

-- ============================================
-- 3. RENAME RAJAONGKIR_RAW_JSON TO BITESHIP_RAW_JSON
-- ============================================
-- Check if column exists before renaming
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'shipping_snapshots' 
        AND column_name = 'rajaongkir_raw_json'
    ) THEN
        ALTER TABLE shipping_snapshots 
        RENAME COLUMN rajaongkir_raw_json TO biteship_raw_json;
    END IF;
END $$;

-- Add biteship_raw_json if it doesn't exist (for fresh installs)
ALTER TABLE shipping_snapshots 
ADD COLUMN IF NOT EXISTS biteship_raw_json JSONB DEFAULT '{}';

-- ============================================
-- 4. ADD AREA_ID COLUMNS TO SHIPPING_SNAPSHOTS
-- ============================================
ALTER TABLE shipping_snapshots 
ADD COLUMN IF NOT EXISTS origin_area_id VARCHAR(100),
ADD COLUMN IF NOT EXISTS origin_area_name VARCHAR(500),
ADD COLUMN IF NOT EXISTS destination_area_id VARCHAR(100),
ADD COLUMN IF NOT EXISTS destination_area_name VARCHAR(500);

-- ============================================
-- 5. ADD AREA_ID COLUMN TO USER_ADDRESSES
-- ============================================
ALTER TABLE user_addresses 
ADD COLUMN IF NOT EXISTS area_id VARCHAR(100);

-- Index for area_id lookup
CREATE INDEX IF NOT EXISTS idx_user_addresses_area_id ON user_addresses(area_id);

-- ============================================
-- 6. VERIFICATION QUERIES
-- ============================================
-- Run these to verify migration success:
-- SELECT column_name FROM information_schema.columns WHERE table_name = 'biteship_locations';
-- SELECT column_name FROM information_schema.columns WHERE table_name = 'shipments' AND column_name LIKE 'biteship%';
-- SELECT column_name FROM information_schema.columns WHERE table_name = 'shipping_snapshots' AND column_name LIKE '%area%';
-- SELECT column_name FROM information_schema.columns WHERE table_name = 'user_addresses' AND column_name = 'area_id';
