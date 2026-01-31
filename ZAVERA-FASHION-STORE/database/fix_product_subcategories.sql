-- ============================================
-- FIX PRODUCT SUBCATEGORIES
-- Memperbaiki subcategory produk yang salah atau NULL
-- ============================================

-- PRIA Category Fixes
-- Update Jeans products to Bottoms (Celana)
UPDATE products 
SET subcategory = 'Bottoms' 
WHERE category = 'pria' 
AND (subcategory = 'Jeans' OR name ILIKE '%jeans%' OR name ILIKE '%celana%');

-- Update Jacket products to Outerwear (Jaket)
UPDATE products 
SET subcategory = 'Outerwear' 
WHERE category = 'pria' 
AND (subcategory = 'Jacket' OR (name ILIKE '%jacket%' AND (subcategory IS NULL OR subcategory = 'Jacket')));

-- Update Bomber/Parasut jackets specifically
UPDATE products 
SET subcategory = 'Outerwear' 
WHERE category = 'pria' 
AND (name ILIKE '%bomber%' OR name ILIKE '%parasut%')
AND (subcategory IS NULL OR subcategory NOT IN ('Outerwear'));

-- VERIFICATION QUERIES
-- Check all products by category and subcategory
SELECT category, subcategory, COUNT(*) as count, STRING_AGG(name, ', ' ORDER BY name) as products
FROM products 
GROUP BY category, subcategory 
ORDER BY category, subcategory;

-- Check for products without subcategory
SELECT id, name, category, subcategory 
FROM products 
WHERE subcategory IS NULL;

-- Check PRIA products specifically
SELECT id, name, category, subcategory 
FROM products 
WHERE category = 'pria' 
ORDER BY subcategory, name;

-- ============================================
-- EXPECTED RESULTS AFTER FIX
-- ============================================
-- PRIA Category:
-- - Tops: Minimalist Cotton Tee, Premium Hoodie, Merino Wool Sweater
-- - Shirts: Slim Fit Shirt
-- - Bottoms: Tailored Trousers, Chino Pants, Mens Denim Jeans, Hip Hop Baggy Jeans, Hip Hop Baggy Jeans 22
-- - Outerwear: Classic Denim Jacket, Casual Blazer, Denim Jacket, Jacket Boomber, Jacket Parasut, Jacket Parasut 22
-- - Suits: Premium Wool Suit
-- - Footwear: Leather Oxford Shoes
