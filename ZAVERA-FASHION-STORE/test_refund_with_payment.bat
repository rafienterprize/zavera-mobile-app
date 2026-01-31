@echo off
echo ========================================
echo Testing Refund with Payment Record
echo ========================================
echo.

echo Step 1: Checking order with payment record...
psql -U postgres -d zavera_db -c "SELECT o.order_code, o.status, o.total_amount, p.id as payment_id, p.status as payment_status, p.payment_method FROM orders o LEFT JOIN payments p ON p.order_id = o.id WHERE o.order_code = 'ZVR-20260125-F08623F3';"

echo.
echo Step 2: Ensuring order is DELIVERED...
psql -U postgres -d zavera_db -c "UPDATE orders SET status = 'DELIVERED' WHERE order_code = 'ZVR-20260125-F08623F3';"

echo.
echo Step 3: Checking current refund status...
psql -U postgres -d zavera_db -c "SELECT r.id, r.refund_code, r.status, r.refund_amount, r.gateway_refund_id FROM refunds r JOIN orders o ON o.id = r.order_id WHERE o.order_code = 'ZVR-20260125-F08623F3';"

echo.
echo ========================================
echo Ready for Testing!
echo ========================================
echo.
echo Next steps:
echo 1. Open: http://localhost:3000/admin/orders/ZVR-20260125-F08623F3
echo 2. Click "Refund" button
echo 3. Select refund type (FULL recommended)
echo 4. Choose reason and click "Process Refund"
echo 5. Check Midtrans dashboard: https://dashboard.sandbox.midtrans.com
echo.
echo After refund, run: verify_refund_result.bat
echo.
pause
