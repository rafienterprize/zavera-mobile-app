-- Update courier logos for missing couriers
-- Add logo URLs for Ninja Express, RPX, SAP, IDexpress, Lion Parcel

-- Update existing couriers with logo URLs
UPDATE shipping_providers SET logo_url = '/images/couriers/ninja.png' WHERE code = 'ninja' OR code = 'ninjavan';
UPDATE shipping_providers SET logo_url = '/images/couriers/rpx.png' WHERE code = 'rpx';
UPDATE shipping_providers SET logo_url = '/images/couriers/sap.png' WHERE code = 'sap';
UPDATE shipping_providers SET logo_url = '/images/couriers/idexpress.png' WHERE code = 'idexpress' OR code = 'ide';
UPDATE shipping_providers SET logo_url = '/images/couriers/lion.png' WHERE code = 'lion' OR code = 'lionparcel';

-- Insert new couriers if they don't exist
INSERT INTO shipping_providers (code, name, logo_url, is_active, created_at, updated_at)
VALUES 
    ('ninja', 'Ninja Express', '/images/couriers/ninja.png', true, NOW(), NOW()),
    ('rpx', 'RPX', '/images/couriers/rpx.png', true, NOW(), NOW()),
    ('sap', 'SAP Express', '/images/couriers/sap.png', true, NOW(), NOW()),
    ('idexpress', 'ID Express', '/images/couriers/idexpress.png', true, NOW(), NOW()),
    ('lion', 'Lion Parcel', '/images/couriers/lion.png', true, NOW(), NOW())
ON CONFLICT (code) DO UPDATE SET
    logo_url = EXCLUDED.logo_url,
    updated_at = NOW();

SELECT 'Courier logos updated successfully' AS status;
