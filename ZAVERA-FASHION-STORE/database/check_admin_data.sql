-- ============================================
-- ZAVERA ADMIN DASHBOARD DATA CHECK
-- Run this to verify data exists in database
-- ============================================

-- Check Orders
SELECT 'ORDERS SUMMARY' as check_type;
SELECT 
    COUNT(*) as total_orders,
    COALESCE(SUM(total_amount), 0) as total_revenue,
    COUNT(*) FILTER (WHERE status = 'PENDING') as pending,
    COUNT(*) FILTER (WHERE status = 'PAID') as paid,
    COUNT(*) FILTER (WHERE status = 'PROCESSING') as processing,
    COUNT(*) FILTER (WHERE status = 'SHIPPED') as shipped,
    COUNT(*) FILTER (WHERE status = 'DELIVERED') as delivered,
    COUNT(*) FILTER (WHERE status = 'CANCELLED') as cancelled
FROM orders;

-- Check Products
SELECT 'PRODUCTS SUMMARY' as check_type;
SELECT 
    COUNT(*) as total_products,
    COUNT(*) FILTER (WHERE is_active = true) as active_products,
    COUNT(*) FILTER (WHERE stock < 10) as low_stock,
    COUNT(*) FILTER (WHERE stock = 0) as out_of_stock
FROM products;

-- Check Shipments
SELECT 'SHIPMENTS SUMMARY' as check_type;
SELECT 
    COUNT(*) as total_shipments,
    COUNT(*) FILTER (WHERE status = 'PENDING') as pending,
    COUNT(*) FILTER (WHERE status = 'SHIPPED') as shipped,
    COUNT(*) FILTER (WHERE status = 'IN_TRANSIT') as in_transit,
    COUNT(*) FILTER (WHERE status = 'DELIVERED') as delivered
FROM shipments;

-- Check Payments
SELECT 'PAYMENTS SUMMARY' as check_type;
SELECT 
    COUNT(*) as total_payments,
    COUNT(*) FILTER (WHERE status = 'PENDING') as pending,
    COUNT(*) FILTER (WHERE status = 'SUCCESS') as success,
    COUNT(*) FILTER (WHERE status = 'FAILED') as failed
FROM payments;

-- Recent Orders
SELECT 'RECENT ORDERS (Last 5)' as check_type;
SELECT order_code, customer_name, total_amount, status, created_at
FROM orders
ORDER BY created_at DESC
LIMIT 5;
