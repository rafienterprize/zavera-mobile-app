-- ============================================
-- ZAVERA E-COMMERCE - CATEGORY MIGRATION
-- Add category support for fashion collections
-- ============================================

-- Add category column to products table
ALTER TABLE products ADD COLUMN IF NOT EXISTS category VARCHAR(50) DEFAULT 'wanita';
ALTER TABLE products ADD COLUMN IF NOT EXISTS subcategory VARCHAR(100);

-- Create index for category filtering
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);

-- Update existing products with categories
UPDATE products SET category = 'pria', subcategory = 'Tops' WHERE slug = 'minimalist-cotton-tee';
UPDATE products SET category = 'pria', subcategory = 'Outerwear' WHERE slug = 'classic-denim-jacket';
UPDATE products SET category = 'pria', subcategory = 'Bottoms' WHERE slug = 'tailored-trousers';
UPDATE products SET category = 'pria', subcategory = 'Tops' WHERE slug = 'premium-hoodie';
UPDATE products SET category = 'pria', subcategory = 'Shirts' WHERE slug = 'slim-fit-shirt';
UPDATE products SET category = 'pria', subcategory = 'Outerwear' WHERE slug = 'casual-blazer';
UPDATE products SET category = 'wanita', subcategory = 'Tops' WHERE slug = 'knit-sweater';
UPDATE products SET category = 'wanita', subcategory = 'Bottoms' WHERE slug = 'relaxed-fit-pants';

-- ============================================
-- INSERT NEW PRODUCTS FOR ALL CATEGORIES
-- ============================================

-- WANITA (Women's) Products
INSERT INTO products (name, slug, description, price, stock, is_active, category, subcategory) VALUES
('Elegant Silk Dress', 'elegant-silk-dress', 'Luxurious silk dress with flowing silhouette. Perfect for special occasions.', 1899000, 25, true, 'wanita', 'Dress'),
('Floral Maxi Skirt', 'floral-maxi-skirt', 'Beautiful floral print maxi skirt with elegant draping.', 799000, 40, true, 'wanita', 'Bottoms'),
('Cashmere Cardigan', 'cashmere-cardigan', 'Soft cashmere cardigan in neutral tones. Ultimate comfort and style.', 1299000, 30, true, 'wanita', 'Outerwear'),
('Satin Blouse', 'satin-blouse', 'Elegant satin blouse with subtle sheen. Perfect for office or evening.', 649000, 45, true, 'wanita', 'Tops'),
('High-Waist Palazzo Pants', 'high-waist-palazzo-pants', 'Flowing palazzo pants with high waist design. Effortlessly chic.', 899000, 35, true, 'wanita', 'Bottoms'),
('Lace Evening Gown', 'lace-evening-gown', 'Stunning lace evening gown for formal occasions.', 2499000, 15, true, 'wanita', 'Dress')
ON CONFLICT (slug) DO NOTHING;

-- PRIA (Men's) Additional Products
INSERT INTO products (name, slug, description, price, stock, is_active, category, subcategory) VALUES
('Premium Wool Suit', 'premium-wool-suit', 'Tailored wool suit with modern fit. Business elegance redefined.', 3499000, 20, true, 'pria', 'Suits'),
('Leather Oxford Shoes', 'leather-oxford-shoes', 'Classic leather oxford shoes. Handcrafted quality.', 1899000, 25, true, 'pria', 'Footwear'),
('Merino Wool Sweater', 'merino-wool-sweater', 'Fine merino wool sweater. Soft, warm, and sophisticated.', 899000, 40, true, 'pria', 'Tops'),
('Chino Pants', 'chino-pants', 'Classic chino pants with perfect fit. Versatile wardrobe essential.', 599000, 50, true, 'pria', 'Bottoms')
ON CONFLICT (slug) DO NOTHING;

-- ANAK (Kids) Products
INSERT INTO products (name, slug, description, price, stock, is_active, category, subcategory) VALUES
('Kids Denim Jacket', 'kids-denim-jacket', 'Stylish denim jacket for kids. Durable and trendy.', 449000, 40, true, 'anak', 'Boys'),
('Girls Floral Dress', 'girls-floral-dress', 'Adorable floral dress for girls. Perfect for any occasion.', 399000, 45, true, 'anak', 'Girls'),
('Kids Sneakers', 'kids-sneakers', 'Comfortable sneakers for active kids. Fun colors and designs.', 349000, 60, true, 'anak', 'Footwear'),
('Baby Romper Set', 'baby-romper-set', 'Soft cotton romper set for babies. Gentle on delicate skin.', 299000, 50, true, 'anak', 'Baby'),
('Boys Polo Shirt', 'boys-polo-shirt', 'Classic polo shirt for boys. Smart casual style.', 249000, 55, true, 'anak', 'Boys'),
('Girls Tutu Skirt', 'girls-tutu-skirt', 'Playful tutu skirt for girls. Perfect for parties.', 279000, 40, true, 'anak', 'Girls')
ON CONFLICT (slug) DO NOTHING;

-- SPORTS Products
INSERT INTO products (name, slug, description, price, stock, is_active, category, subcategory) VALUES
('Performance Running Shoes', 'performance-running-shoes', 'High-performance running shoes with advanced cushioning.', 1499000, 35, true, 'sports', 'Footwear'),
('Yoga Leggings', 'yoga-leggings', 'Flexible yoga leggings with moisture-wicking fabric.', 549000, 50, true, 'sports', 'Activewear'),
('Training Tank Top', 'training-tank-top', 'Breathable training tank top for intense workouts.', 299000, 60, true, 'sports', 'Activewear'),
('Sports Jacket', 'sports-jacket', 'Lightweight sports jacket with wind resistance.', 899000, 30, true, 'sports', 'Outerwear'),
('Gym Shorts', 'gym-shorts', 'Comfortable gym shorts with quick-dry technology.', 349000, 55, true, 'sports', 'Activewear'),
('Sports Bra', 'sports-bra', 'Supportive sports bra for high-impact activities.', 399000, 45, true, 'sports', 'Activewear')
ON CONFLICT (slug) DO NOTHING;

-- LUXURY Products
INSERT INTO products (name, slug, description, price, stock, is_active, category, subcategory) VALUES
('Designer Leather Handbag', 'designer-leather-handbag', 'Exquisite designer handbag crafted from premium Italian leather.', 8999000, 10, true, 'luxury', 'Accessories'),
('Silk Evening Clutch', 'silk-evening-clutch', 'Elegant silk clutch with gold hardware. Limited edition.', 3499000, 15, true, 'luxury', 'Accessories'),
('Cashmere Coat', 'cashmere-coat', 'Luxurious full-length cashmere coat. Timeless elegance.', 12999000, 8, true, 'luxury', 'Outerwear'),
('Diamond Watch', 'diamond-watch', 'Sophisticated timepiece with diamond accents. Swiss movement.', 25999000, 5, true, 'luxury', 'Accessories'),
('Designer Sunglasses', 'designer-sunglasses', 'Premium designer sunglasses with UV protection.', 4999000, 20, true, 'luxury', 'Accessories'),
('Luxury Silk Scarf', 'luxury-silk-scarf', 'Hand-printed silk scarf. Artistic design meets luxury.', 2499000, 25, true, 'luxury', 'Accessories')
ON CONFLICT (slug) DO NOTHING;

-- BEAUTY Products
INSERT INTO products (name, slug, description, price, stock, is_active, category, subcategory) VALUES
('Premium Face Serum', 'premium-face-serum', 'Advanced anti-aging serum with hyaluronic acid and vitamin C.', 899000, 40, true, 'beauty', 'Skincare'),
('Luxury Lipstick Set', 'luxury-lipstick-set', 'Collection of 6 premium lipsticks in trending shades.', 1299000, 35, true, 'beauty', 'Makeup'),
('Rose Gold Perfume', 'rose-gold-perfume', 'Elegant fragrance with notes of rose, jasmine, and sandalwood.', 1899000, 30, true, 'beauty', 'Fragrance'),
('Hydrating Face Cream', 'hydrating-face-cream', 'Deep hydrating cream with natural ingredients.', 649000, 50, true, 'beauty', 'Skincare'),
('Eyeshadow Palette', 'eyeshadow-palette', 'Professional eyeshadow palette with 18 versatile shades.', 799000, 45, true, 'beauty', 'Makeup'),
('Luxury Body Lotion', 'luxury-body-lotion', 'Nourishing body lotion with shea butter and vitamin E.', 449000, 55, true, 'beauty', 'Skincare')
ON CONFLICT (slug) DO NOTHING;

-- Add images for new products
INSERT INTO product_images (product_id, image_url, is_primary, display_order)
SELECT p.id, 
  CASE p.category
    WHEN 'wanita' THEN 'https://images.unsplash.com/photo-1595777457583-95e059d581b8?w=800&q=80'
    WHEN 'pria' THEN 'https://images.unsplash.com/photo-1617137968427-85924c800a22?w=800&q=80'
    WHEN 'anak' THEN 'https://images.unsplash.com/photo-1519238263530-99bdd11df2ea?w=800&q=80'
    WHEN 'sports' THEN 'https://images.unsplash.com/photo-1571019613454-1cb2f99b2d8b?w=800&q=80'
    WHEN 'luxury' THEN 'https://images.unsplash.com/photo-1584917865442-de89df76afd3?w=800&q=80'
    WHEN 'beauty' THEN 'https://images.unsplash.com/photo-1596462502278-27bfdc403348?w=800&q=80'
  END,
  true, 1
FROM products p
WHERE NOT EXISTS (SELECT 1 FROM product_images pi WHERE pi.product_id = p.id);
