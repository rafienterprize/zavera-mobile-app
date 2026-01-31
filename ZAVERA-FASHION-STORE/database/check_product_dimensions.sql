-- Check product dimensions for shipping calculation
-- Run this to verify dimensions are set correctly

SELECT 
    id,
    name,
    weight,
    length,
    width,
    height,
    stock,
    price
FROM products
WHERE is_active = true
ORDER BY id DESC
LIMIT 10;
