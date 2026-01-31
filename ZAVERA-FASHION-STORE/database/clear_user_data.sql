-- ============================================
-- CLEAR USER DATA SCRIPT
-- Menghapus semua data user/transaksi
-- TIDAK menghapus: products, product_images, categories, 
--                  provinces, cities, subdistricts, email_templates
-- ============================================

-- PERINGATAN: Script ini akan menghapus SEMUA data user!
-- Pastikan Anda sudah backup database sebelum menjalankan script ini.

-- ============================================
-- 1. CLEAR DISPUTE & REFUND DATA
-- ============================================
TRUNCATE TABLE disputes CASCADE;
TRUNCATE TABLE refunds CASCADE;

-- ============================================
-- 2. CLEAR EMAIL LOGS (bukan templates)
-- ============================================
TRUNCATE TABLE email_logs CASCADE;

-- ============================================
-- 3. CLEAR SHIPPING & STOCK DATA
-- ============================================
TRUNCATE TABLE shipping_snapshots CASCADE;
TRUNCATE TABLE stock_movements CASCADE;

-- ============================================
-- 4. CLEAR ORDER DATA
-- ============================================
TRUNCATE TABLE order_status_history CASCADE;
TRUNCATE TABLE core_payment_sync_logs CASCADE;
TRUNCATE TABLE order_payments CASCADE;
TRUNCATE TABLE payments CASCADE;
TRUNCATE TABLE order_items CASCADE;
TRUNCATE TABLE orders CASCADE;

-- ============================================
-- 5. CLEAR CART DATA
-- ============================================
TRUNCATE TABLE cart_items CASCADE;
TRUNCATE TABLE carts CASCADE;

-- ============================================
-- 6. CLEAR USER AUTH DATA
-- ============================================
TRUNCATE TABLE email_verification_tokens CASCADE;
TRUNCATE TABLE password_reset_tokens CASCADE;
TRUNCATE TABLE user_sessions CASCADE;

-- ============================================
-- 7. CLEAR USERS (terakhir karena FK)
-- ============================================
TRUNCATE TABLE users CASCADE;

-- ============================================
-- 8. RESET SEQUENCES (auto-increment)
-- ============================================
ALTER SEQUENCE IF EXISTS users_id_seq RESTART WITH 1;
ALTER SEQUENCE IF EXISTS carts_id_seq RESTART WITH 1;
ALTER SEQUENCE IF EXISTS cart_items_id_seq RESTART WITH 1;
ALTER SEQUENCE IF EXISTS orders_id_seq RESTART WITH 1;
ALTER SEQUENCE IF EXISTS order_items_id_seq RESTART WITH 1;
ALTER SEQUENCE IF EXISTS payments_id_seq RESTART WITH 1;
ALTER SEQUENCE IF EXISTS order_payments_id_seq RESTART WITH 1;
ALTER SEQUENCE IF EXISTS core_payment_sync_logs_id_seq RESTART WITH 1;
ALTER SEQUENCE IF EXISTS disputes_id_seq RESTART WITH 1;
ALTER SEQUENCE IF EXISTS refunds_id_seq RESTART WITH 1;
ALTER SEQUENCE IF EXISTS email_logs_id_seq RESTART WITH 1;
ALTER SEQUENCE IF EXISTS shipping_snapshots_id_seq RESTART WITH 1;
ALTER SEQUENCE IF EXISTS stock_movements_id_seq RESTART WITH 1;
ALTER SEQUENCE IF EXISTS order_status_history_id_seq RESTART WITH 1;

-- ============================================
-- 9. VERIFIKASI
-- ============================================
SELECT '=== DATA USER (DIHAPUS) ===' AS info;
SELECT 'users' as table_name, COUNT(*) as count FROM users
UNION ALL SELECT 'carts', COUNT(*) FROM carts
UNION ALL SELECT 'cart_items', COUNT(*) FROM cart_items
UNION ALL SELECT 'orders', COUNT(*) FROM orders
UNION ALL SELECT 'order_items', COUNT(*) FROM order_items
UNION ALL SELECT 'payments', COUNT(*) FROM payments
UNION ALL SELECT 'order_payments', COUNT(*) FROM order_payments
UNION ALL SELECT 'core_payment_sync_logs', COUNT(*) FROM core_payment_sync_logs
UNION ALL SELECT 'disputes', COUNT(*) FROM disputes
UNION ALL SELECT 'refunds', COUNT(*) FROM refunds
UNION ALL SELECT 'email_logs', COUNT(*) FROM email_logs
UNION ALL SELECT 'shipping_snapshots', COUNT(*) FROM shipping_snapshots
UNION ALL SELECT 'stock_movements', COUNT(*) FROM stock_movements;

SELECT '=== DATA SISTEM (TIDAK DIHAPUS) ===' AS info;
SELECT 'products', COUNT(*) FROM products
UNION ALL SELECT 'product_images', COUNT(*) FROM product_images
UNION ALL SELECT 'categories', COUNT(*) FROM categories
UNION ALL SELECT 'provinces', COUNT(*) FROM provinces
UNION ALL SELECT 'cities', COUNT(*) FROM cities
UNION ALL SELECT 'subdistricts', COUNT(*) FROM subdistricts
UNION ALL SELECT 'email_templates', COUNT(*) FROM email_templates;

SELECT 'User data cleared successfully!' AS status;
