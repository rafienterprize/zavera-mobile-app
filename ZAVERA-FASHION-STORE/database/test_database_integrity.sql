-- ============================================
-- ZAVERA DATABASE INTEGRITY TESTING
-- Run this to verify database consistency
-- ============================================

\echo '========================================='
\echo 'ZAVERA DATABASE INTEGRITY TEST'
\echo '========================================='
\echo ''

-- TEST 1: Product Data Consistency
\echo 'TEST 1: Product Data Consistency'
\echo '-----------------------------------------'

\echo 'Checking products with invalid categories...'
SELECT COUNT(*) as invalid_categories 
FROM products 
WHERE category NOT IN ('wanita', 'pria', 'anak', 'sports', 'luxury', 'beauty');
-- Expected: 0

\echo 'Checking products without subcategory...'
SELECT COUNT(*) as missing_subcategory 
FROM products 
WHERE subcategory IS NULL;
-- Expected: 0

\echo 'Checking products with negative stock...'
SELECT COUNT(*) as negative_stock 
FROM products 
WHERE stock < 0;
-- Expected: 0

\echo ''

-- TEST 2: Variant Data Consistency
\echo 'TEST 2: Variant Data Consistency'
\echo '-----------------------------------------'

\echo 'Checking variants with invalid product_id...'
SELECT COUNT(*) as orphan_variants 
FROM product_variants 
WHERE product_id NOT IN (SELECT id FROM products);
-- Expected: 0

\echo 'Checking variants with negative stock...'
SELECT COUNT(*) as negative_variant_stock 
FROM product_variants 
WHERE stock_quantity < 0;
-- Expected: 0

\echo 'Products with variants:'
SELECT category, COUNT(DISTINCT p.id) as products_with_variants
FROM products p
JOIN product_variants pv ON p.id = pv.product_id
GROUP BY category
ORDER BY category;

\echo ''

-- TEST 3: Order Data Consistency
\echo 'TEST 3: Order Data Consistency'
\echo '-----------------------------------------'

\echo 'Checking orders with invalid user_id...'
SELECT COUNT(*) as orphan_orders 
FROM orders 
WHERE user_id IS NOT NULL 
AND user_id NOT IN (SELECT id FROM users);
-- Expected: 0

\echo 'Checking order total consistency...'
SELECT COUNT(*) as inconsistent_totals
FROM (
    SELECT o.id, o.order_code, o.total_amount,
           (o.subtotal + o.shipping_cost + COALESCE(o.tax, 0) - COALESCE(o.discount, 0)) as calculated_total
    FROM orders o
) subquery
WHERE ABS(total_amount - calculated_total) > 0.01;
-- Expected: 0

\echo 'Order status distribution:'
SELECT status, COUNT(*) as count
FROM orders
GROUP BY status
ORDER BY status;

\echo ''

-- TEST 4: Cart Data Consistency
\echo 'TEST 4: Cart Data Consistency'
\echo '-----------------------------------------'

\echo 'Checking cart items with invalid product_id...'
SELECT COUNT(*) as invalid_cart_items 
FROM cart_items 
WHERE product_id NOT IN (SELECT id FROM products);
-- Expected: 0

\echo 'Checking cart items with invalid cart_id...'
SELECT COUNT(*) as orphan_cart_items 
FROM cart_items 
WHERE cart_id NOT IN (SELECT id FROM carts);
-- Expected: 0

\echo 'Active carts:'
SELECT COUNT(*) as active_carts FROM carts;

\echo ''

-- TEST 5: Payment Data Consistency
\echo 'TEST 5: Payment Data Consistency'
\echo '-----------------------------------------'

\echo 'Checking payments with invalid order_id...'
SELECT COUNT(*) as orphan_payments 
FROM payments 
WHERE order_id NOT IN (SELECT id FROM orders);
-- Expected: 0

\echo 'Checking payment amount consistency...'
SELECT COUNT(*) as amount_mismatch
FROM orders o
JOIN payments p ON o.id = p.order_id
WHERE ABS(o.total_amount - p.amount) > 0.01;
-- Expected: 0

\echo 'Payment status distribution:'
SELECT status, COUNT(*) as count
FROM payments
GROUP BY status
ORDER BY status;

\echo ''

-- TEST 6: Refund Data Consistency
\echo 'TEST 6: Refund Data Consistency'
\echo '-----------------------------------------'

\echo 'Checking refunds exceeding order amount...'
SELECT COUNT(*) as excessive_refunds
FROM refunds r
JOIN orders o ON r.order_id = o.id
WHERE r.refund_amount > o.total_amount;
-- Expected: 0

\echo 'Checking refund status validity...'
SELECT COUNT(*) as invalid_refund_status 
FROM refunds 
WHERE status NOT IN ('PENDING', 'PROCESSING', 'COMPLETED', 'FAILED');
-- Expected: 0

\echo 'Refund status distribution:'
SELECT status, COUNT(*) as count
FROM refunds
GROUP BY status
ORDER BY status;

\echo ''

-- TEST 7: Stock Consistency
\echo 'TEST 7: Stock Consistency'
\echo '-----------------------------------------'

\echo 'Products with stock vs variant stock mismatch:'
SELECT p.id, p.name, p.stock as product_stock, 
       COALESCE(SUM(pv.stock_quantity), 0) as variant_stock,
       p.stock - COALESCE(SUM(pv.stock_quantity), 0) as difference
FROM products p
LEFT JOIN product_variants pv ON p.id = pv.product_id
GROUP BY p.id, p.name, p.stock
HAVING ABS(p.stock - COALESCE(SUM(pv.stock_quantity), 0)) > 0
ORDER BY difference DESC
LIMIT 10;

\echo ''

-- TEST 8: Image Data
\echo 'TEST 8: Image Data'
\echo '-----------------------------------------'

\echo 'Products without images:'
SELECT COUNT(*) as products_without_images
FROM products p
WHERE NOT EXISTS (SELECT 1 FROM product_images pi WHERE pi.product_id = p.id);

\echo 'Products without primary image:'
SELECT COUNT(*) as products_without_primary
FROM products p
WHERE NOT EXISTS (
    SELECT 1 FROM product_images pi 
    WHERE pi.product_id = p.id AND pi.is_primary = true
);

\echo ''

-- TEST 9: Audit Trail
\echo 'TEST 9: Audit Trail'
\echo '-----------------------------------------'

\echo 'Recent admin actions:'
SELECT action, COUNT(*) as count
FROM admin_audit_logs
WHERE created_at > NOW() - INTERVAL '7 days'
GROUP BY action
ORDER BY count DESC
LIMIT 10;

\echo ''

-- TEST 10: Summary Statistics
\echo 'TEST 10: Summary Statistics'
\echo '========================================='

\echo 'Database Summary:'
SELECT 
    (SELECT COUNT(*) FROM products) as total_products,
    (SELECT COUNT(*) FROM product_variants) as total_variants,
    (SELECT COUNT(*) FROM users) as total_users,
    (SELECT COUNT(*) FROM orders) as total_orders,
    (SELECT COUNT(*) FROM payments) as total_payments,
    (SELECT COUNT(*) FROM refunds) as total_refunds,
    (SELECT COUNT(*) FROM carts) as total_carts,
    (SELECT COUNT(*) FROM cart_items) as total_cart_items;

\echo ''
\echo 'Products by Category:'
SELECT category, COUNT(*) as count, 
       SUM(stock) as total_stock,
       AVG(price) as avg_price
FROM products
GROUP BY category
ORDER BY category;

\echo ''
\echo '========================================='
\echo 'DATABASE INTEGRITY TEST COMPLETE'
\echo '========================================='
\echo ''
\echo 'Review results above:'
\echo '- All counts should be 0 for error checks'
\echo '- Distributions should look reasonable'
\echo '- Stock levels should be positive'
\echo ''
