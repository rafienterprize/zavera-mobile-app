@echo off
echo ========================================
echo ZAVERA REFUND SYSTEM TEST
echo ========================================
echo.
echo Test Order: ZVR-20260127-B8B3ACCD
echo.
echo Step 1: Login as admin...
curl -X POST http://localhost:8080/api/auth/login ^
  -H "Content-Type: application/json" ^
  -d "{\"email\":\"pemberani073@gmail.com\",\"password\":\"admin123\"}" ^
  -o login_response.json
echo.
echo.

echo Step 2: Extract token...
for /f "tokens=2 delims=:," %%a in ('type login_response.json ^| findstr "token"') do set TOKEN=%%a
set TOKEN=%TOKEN:"=%
set TOKEN=%TOKEN: =%
echo Token: %TOKEN%
echo.

echo Step 3: Create FULL refund...
curl -X POST http://localhost:8080/api/admin/refunds ^
  -H "Content-Type: application/json" ^
  -H "Authorization: Bearer %TOKEN%" ^
  -d "{\"order_code\":\"ZVR-20260127-B8B3ACCD\",\"refund_type\":\"FULL\",\"reason\":\"CUSTOMER_REQUEST\",\"reason_detail\":\"Test refund system\",\"idempotency_key\":\"test-refund-%RANDOM%\"}"
echo.
echo.

echo Step 4: Check refunds for order...
curl -X GET "http://localhost:8080/api/admin/orders/ZVR-20260127-B8B3ACCD/refunds" ^
  -H "Authorization: Bearer %TOKEN%"
echo.
echo.

echo ========================================
echo TEST COMPLETE
echo ========================================
pause
