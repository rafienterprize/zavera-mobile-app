-- Cleanup Inactive Products
-- This script will permanently delete all inactive products and their related data

-- Show what will be deleted
SELECT 
    p.id,
    p.name,
    p.slug,
    p.is_active,
    COUNT(DISTINCT pi.id) as image_count,
    COUNT(DISTINCT pv.id) as variant_count
FROM products p
LEFT JOIN product_images pi ON p.id = pi.product_id
LEFT JOIN product_variants pv ON p.id = pv.product_id
WHERE p.is_active = false
GROUP BY p.id, p.name, p.slug, p.is_active
ORDER BY p.id;

-- Uncomment below to actually delete
-- WARNING: This is permanent and cannot be undone!

/*
BEGIN;

-- Delete product images for inactive products
DELETE FROM product_images 
WHERE product_id IN (SELECT id FROM products WHERE is_active = false);

-- Delete product variants for inactive products
DELETE FROM product_variants 
WHERE product_id IN (SELECT id FROM products WHERE is_active = false);

-- Delete inactive products
DELETE FROM products WHERE is_active = false;

COMMIT;

-- Verify deletion
SELECT COUNT(*) as remaining_inactive_products 
FROM products 
WHERE is_active = false;
*/
