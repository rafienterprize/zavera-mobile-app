@echo off
echo ========================================
echo Stock Display Test Script
echo ========================================
echo.

echo This script will help you verify the stock display fixes
echo.

echo TEST 1: Check if product has variants
echo ----------------------------------------
curl -s http://localhost:8080/api/products/46/variants
echo.
echo.

echo TEST 2: Get product details
echo ----------------------------------------
curl -s http://localhost:8080/api/products/46
echo.
echo.

echo TEST 3: Check variant stock summary
echo ----------------------------------------
curl -s -H "Authorization: Bearer YOUR_TOKEN_HERE" http://localhost:8080/api/admin/variants/stock-summary/46
echo.
echo.

echo ========================================
echo Test Complete!
echo ========================================
echo.
echo WHAT TO CHECK:
echo.
echo 1. TEST 1 should return array of variants with stock_quantity
echo    - If empty array: Product has no variants (uses product.stock)
echo    - If has items: Product uses variant stock (product.stock = 0 is normal)
echo.
echo 2. TEST 2 should show product.stock
echo    - If 0 and has variants: This is CORRECT
echo    - If 0 and no variants: Product is out of stock
echo.
echo 3. TEST 3 shows total stock across all variants
echo    - Requires admin authentication token
echo.
echo ADMIN DASHBOARD:
echo - Products with variants show "Variants" icon
echo - Products without variants show stock number
echo.
echo CUSTOMER PAGE:
echo - Before variant selection: "Pilih ukuran dan warna"
echo - After variant selection: Shows actual stock
echo - Out of stock variant: "SOLD OUT"
echo.

pause
