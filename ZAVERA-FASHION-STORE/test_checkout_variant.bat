@echo off
echo ========================================
echo Testing Variant Product Checkout
echo ========================================
echo.

echo Step 1: Check cart items
psql -U postgres -d zavera_db -c "SELECT ci.id, ci.product_id, ci.variant_id, ci.quantity, p.name, pv.variant_name, pv.size, pv.color, pv.stock_quantity FROM cart_items ci JOIN products p ON p.id = ci.product_id LEFT JOIN product_variants pv ON pv.id = ci.variant_id WHERE ci.cart_id = (SELECT id FROM carts WHERE user_id = 1 LIMIT 1)"
echo.

echo Step 2: Check variant stock
psql -U postgres -d zavera_db -c "SELECT id, product_id, variant_name, size, color, stock_quantity FROM product_variants WHERE id IN (5, 6)"
echo.

echo Step 3: Verify foreign key constraint
psql -U postgres -d zavera_db -c "SELECT tc.constraint_name, tc.table_name, kcu.column_name, ccu.table_name AS foreign_table_name FROM information_schema.table_constraints AS tc JOIN information_schema.key_column_usage AS kcu ON tc.constraint_name = kcu.constraint_name JOIN information_schema.constraint_column_usage AS ccu ON ccu.constraint_name = tc.constraint_name WHERE tc.constraint_type = 'FOREIGN KEY' AND tc.table_name='order_items' AND kcu.column_name='variant_id'"
echo.

echo ========================================
echo Now test checkout in the browser:
echo 1. Go to http://localhost:3000/checkout
echo 2. Fill in shipping address
echo 3. Select shipping method
echo 4. Select payment method
echo 5. Click "Bayar Sekarang"
echo ========================================
pause
