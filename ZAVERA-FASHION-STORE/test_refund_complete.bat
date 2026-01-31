@echo off
echo ========================================
echo REFUND SYSTEM COMPLETE TEST
echo ========================================
echo.

REM Set database credentials
set PGPASSWORD=Yan2692009

echo [1/5] Checking for real Midtrans transactions...
psql -U postgres -d zavera_db -c "SELECT o.order_code, p.id as payment_id, p.transaction_id, p.amount, p.status FROM payments p JOIN orders o ON p.order_id = o.id WHERE p.status = 'SUCCESS' ORDER BY p.created_at DESC LIMIT 5;"

echo.
echo [2/5] Checking existing refunds...
psql -U postgres -d zavera_db -c "SELECT r.id, r.refund_code, r.status, r.gateway_refund_id, r.gateway_status, o.order_code FROM refunds r JOIN orders o ON r.order_id = o.id ORDER BY r.created_at DESC LIMIT 5;"

echo.
echo [3/5] Checking refund with payment_id (should be processed to Midtrans)...
psql -U postgres -d zavera_db -c "SELECT r.id, r.refund_code, r.status, r.payment_id, r.gateway_refund_id, o.order_code FROM refunds r JOIN orders o ON r.order_id = o.id WHERE r.payment_id IS NOT NULL;"

echo.
echo [4/5] Checking manual refunds (no payment_id, won't appear in Midtrans)...
psql -U postgres -d zavera_db -c "SELECT r.id, r.refund_code, r.status, r.payment_id, r.gateway_refund_id, o.order_code FROM refunds r JOIN orders o ON r.order_id = o.id WHERE r.payment_id IS NULL;"

echo.
echo ========================================
echo SUMMARY:
echo ========================================
echo - Refunds WITH payment_id: Need to be processed to Midtrans
echo - Refunds WITHOUT payment_id: Manual refunds (won't appear in Midtrans)
echo.
echo TO TEST REFUND TO MIDTRANS:
echo 1. Create a NEW order via frontend
echo 2. Pay using Midtrans sandbox (GoPay/VA/QRIS)
echo 3. Wait for payment SUCCESS
echo 4. Create refund for that order
echo 5. Process refund via admin panel
echo 6. Check Midtrans dashboard - status should change to "Refund"
echo.
echo See REFUND_TESTING_PRODUCTION_GUIDE.md for detailed instructions
echo ========================================

pause
