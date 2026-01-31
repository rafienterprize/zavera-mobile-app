@echo off
echo ========================================
echo Verifying Refund Result
echo ========================================
echo.

echo Checking refund record...
psql -U postgres -d zavera_db -c "SELECT r.id, r.refund_code, r.status, r.refund_amount, r.gateway_refund_id, r.created_at FROM refunds r JOIN orders o ON o.id = r.order_id WHERE o.order_code = 'ZVR-20260125-F08623F3' ORDER BY r.created_at DESC LIMIT 1;"

echo.
echo Checking order status...
psql -U postgres -d zavera_db -c "SELECT order_code, status, refund_status, refund_amount, total_amount FROM orders WHERE order_code = 'ZVR-20260125-F08623F3';"

echo.
echo Checking refund status history...
psql -U postgres -d zavera_db -c "SELECT h.from_status, h.to_status, h.changed_by, h.reason, h.created_at FROM refund_status_history h JOIN refunds r ON r.id = h.refund_id JOIN orders o ON o.id = r.order_id WHERE o.order_code = 'ZVR-20260125-F08623F3' ORDER BY h.created_at DESC;"

echo.
echo ========================================
echo Verification Complete!
echo ========================================
echo.
echo Expected Results:
echo - Refund status: COMPLETED
echo - Gateway refund ID: Number from Midtrans (NOT NULL or MANUAL_REFUND)
echo - Order status: REFUNDED
echo - Order refund_status: FULL
echo - Order refund_amount: 209000.00
echo.
echo Now check Midtrans Dashboard:
echo https://dashboard.sandbox.midtrans.com
echo.
echo Search for transaction: ZVR-20260125-F08623F3
echo Status should be: refund (changed from settlement)
echo.
pause
