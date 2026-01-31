-- Migration: Add district_id column to user_addresses table
-- This is required for accurate shipping calculation using the new Komerce RajaOngkir API
-- The API requires district IDs (kecamatan) instead of city IDs for shipping cost calculation

-- Add district_id column to user_addresses
ALTER TABLE user_addresses 
ADD COLUMN IF NOT EXISTS district_id VARCHAR(20);

-- Add comment explaining the field
COMMENT ON COLUMN user_addresses.district_id IS 'Kecamatan ID from RajaOngkir API for shipping calculation';

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_user_addresses_district_id ON user_addresses(district_id);

-- Verify the migration
SELECT column_name, data_type, is_nullable 
FROM information_schema.columns 
WHERE table_name = 'user_addresses' 
AND column_name = 'district_id';
