-- Setup test cart with 3 products for QA testing
TRUNCATE TABLE carts, cart_items RESTART IDENTITY CASCADE;

-- Create a single cart
INSERT INTO carts (session_id, created_at, updated_at) 
VALUES ('test-qa-session', NOW(), NOW());

-- Add 3 items to cart (Product 1: 350g, Product 2: 600g, Product 4: 700g)
INSERT INTO cart_items (cart_id, product_id, quantity, price_snapshot, metadata, created_at, updated_at)
SELECT 1, 1, 1, price, '{"size":"M"}'::jsonb, NOW(), NOW() FROM products WHERE id = 1;

INSERT INTO cart_items (cart_id, product_id, quantity, price_snapshot, metadata, created_at, updated_at)
SELECT 1, 2, 1, price, '{"size":"L"}'::jsonb, NOW(), NOW() FROM products WHERE id = 2;

INSERT INTO cart_items (cart_id, product_id, quantity, price_snapshot, metadata, created_at, updated_at)
SELECT 1, 4, 1, price, '{"size":"XL"}'::jsonb, NOW(), NOW() FROM products WHERE id = 4;

-- Verify cart contents
SELECT 
    c.session_id,
    ci.product_id,
    p.name,
    p.weight as weight_grams,
    ci.price_snapshot
FROM carts c 
JOIN cart_items ci ON c.id = ci.cart_id 
JOIN products p ON ci.product_id = p.id;

-- Show total weight
SELECT 
    SUM(p.weight) as total_weight_grams,
    SUM(ci.price_snapshot) as subtotal
FROM carts c 
JOIN cart_items ci ON c.id = ci.cart_id 
JOIN products p ON ci.product_id = p.id;
