@echo off
echo ========================================
echo ZAVERA API ENDPOINT TESTING
echo ========================================
echo.

REM Test 1: Health Check
echo [TEST 1] Health Check
curl -s http://localhost:8080/health
echo.
echo.

REM Test 2: Get All Products
echo [TEST 2] Get All Products
curl -s http://localhost:8080/products | jq "length"
echo.
echo.

REM Test 3: Get PRIA Products
echo [TEST 3] Get PRIA Category Products
curl -s "http://localhost:8080/products?category=pria" | jq "length"
echo.
echo.

REM Test 4: Get Product by ID
echo [TEST 4] Get Product by ID (ID: 46)
curl -s http://localhost:8080/products/46 | jq "{id, name, category, subcategory, available_sizes}"
echo.
echo.

REM Test 5: Get Product Variants
echo [TEST 5] Get Product Variants (Product ID: 46)
curl -s http://localhost:8080/products/46/variants | jq "length"
echo.
echo.

REM Test 6: Get Shipping Rates (requires valid data)
echo [TEST 6] Get Shipping Rates
curl -s -X POST http://localhost:8080/api/shipping/rates ^
-H "Content-Type: application/json" ^
  -d "{\"destination_area_id\":\"IDNP6IDNC148IDND1817IDZ12190\",\"items\":[{\"product_id\":46,\"quantity\":1,\"variant_id\":1}]}" | jq "length"
echo.
echo.

REM Test 7: Get Districts
echo [TEST 7] Get Districts (Province: DKI Jakarta)
curl -s "http://localhost:8080/api/shipping/districts?province=DKI%%20Jakarta" | jq "length"
echo.
echo.

echo ========================================
echo API ENDPOINT TESTS COMPLETE
echo ========================================
echo.
echo Summary:
echo - Health Check: Check if "status: ok"
echo - Products: Should return product count
echo - PRIA Products: Should return 17
echo - Product Detail: Should show product info
echo - Variants: Should show variant count
echo - Shipping: Should return courier options
echo - Districts: Should return district list
echo.
pause
