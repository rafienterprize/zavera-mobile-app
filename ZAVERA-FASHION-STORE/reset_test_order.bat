@echo off
echo ========================================
echo Reset Test Order for Re-testing
echo ========================================
echo.
echo WARNING: This will delete all refunds for order ZVR-20260125-F08623F3
echo and reset order status to DELIVERED
echo.
set /p confirm="Are you sure? (Y/N): "
if /i not "%confirm%"=="Y" (
    echo Cancelled.
    pause
    exit /b
)

echo.
echo Deleting refund status history...
psql -U postgres -d zavera_db -c "DELETE FROM refund_status_history WHERE refund_id IN (SELECT r.id FROM refunds r JOIN orders o ON o.id = r.order_id WHERE o.order_code = 'ZVR-20260125-F08623F3');"

echo.
echo Deleting refund items...
psql -U postgres -d zavera_db -c "DELETE FROM refund_items WHERE refund_id IN (SELECT r.id FROM refunds r JOIN orders o ON o.id = r.order_id WHERE o.order_code = 'ZVR-20260125-F08623F3');"

echo.
echo Deleting refunds...
psql -U postgres -d zavera_db -c "DELETE FROM refunds WHERE order_id IN (SELECT id FROM orders WHERE order_code = 'ZVR-20260125-F08623F3');"

echo.
echo Resetting order status...
psql -U postgres -d zavera_db -c "UPDATE orders SET status = 'DELIVERED', refund_status = NULL, refund_amount = 0, refunded_at = NULL WHERE order_code = 'ZVR-20260125-F08623F3';"

echo.
echo Verifying reset...
psql -U postgres -d zavera_db -c "SELECT order_code, status, refund_status, refund_amount FROM orders WHERE order_code = 'ZVR-20260125-F08623F3';"

echo.
echo ========================================
echo Reset Complete!
echo ========================================
echo.
echo Order is now ready for re-testing.
echo Run: test_refund_with_payment.bat
echo.
pause
