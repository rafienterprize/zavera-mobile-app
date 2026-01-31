@echo off
echo Testing Bulk Generate Variants API
echo.

REM Get admin token first (replace with your actual admin email)
set ADMIN_EMAIL=your-admin@gmail.com

echo Step 1: Login as admin...
curl -X POST http://localhost:8080/api/auth/login ^
  -H "Content-Type: application/json" ^
  -d "{\"email\":\"%ADMIN_EMAIL%\",\"password\":\"your-password\"}" ^
  > admin_token.json

echo.
echo Step 2: Bulk generate variants for product ID 1...
echo Generating: Sizes [S, M, L, XL] x Colors [Black, White, Navy]
echo.

curl -X POST http://localhost:8080/api/admin/variants/bulk-generate ^
  -H "Content-Type: application/json" ^
  -H "Authorization: Bearer YOUR_TOKEN_HERE" ^
  -d "{\"product_id\":1,\"sizes\":[\"S\",\"M\",\"L\",\"XL\"],\"colors\":[\"Black\",\"White\",\"Navy\"],\"base_price\":400000,\"stock_per_variant\":10}"

echo.
echo.
echo Step 3: Get variants for product 1...
curl http://localhost:8080/api/products/1/variants

echo.
echo.
echo Step 4: Get product with variants...
curl http://localhost:8080/api/products/1/with-variants

echo.
pause
