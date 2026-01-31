-- ============================================
-- FIX PRODUCT IMAGES
-- Ensure all products have at least one image
-- ============================================

-- First, let's see which products don't have images
SELECT p.id, p.name, p.slug, p.category
FROM products p
LEFT JOIN product_images pi ON pi.product_id = p.id
WHERE pi.id IS NULL;

-- Add images for products that don't have any
INSERT INTO product_images (product_id, image_url, is_primary, display_order)
SELECT p.id, 
  CASE p.category
    WHEN 'wanita' THEN 'https://images.unsplash.com/photo-1595777457583-95e059d581b8?w=800&q=80'
    WHEN 'pria' THEN 'https://images.unsplash.com/photo-1617137968427-85924c800a22?w=800&q=80'
    WHEN 'anak' THEN 'https://images.unsplash.com/photo-1519238263530-99bdd11df2ea?w=800&q=80'
    WHEN 'sports' THEN 'https://images.unsplash.com/photo-1571019613454-1cb2f99b2d8b?w=800&q=80'
    WHEN 'luxury' THEN 'https://images.unsplash.com/photo-1584917865442-de89df76afd3?w=800&q=80'
    WHEN 'beauty' THEN 'https://images.unsplash.com/photo-1596462502278-27bfdc403348?w=800&q=80'
    ELSE 'https://images.unsplash.com/photo-1617137968427-85924c800a22?w=800&q=80'
  END,
  true, 1
FROM products p
WHERE NOT EXISTS (SELECT 1 FROM product_images pi WHERE pi.product_id = p.id);

-- Verify all products now have images
SELECT p.id, p.name, p.slug, pi.image_url, pi.is_primary
FROM products p
LEFT JOIN product_images pi ON pi.product_id = p.id
ORDER BY p.id;
