@echo off
echo ============================================
echo TEST SHIPPING API WITH DIMENSIONS
echo ============================================
echo.
echo Testing shipping rates for Denim Jacket (2 items)
echo Origin: 50113 (Pedurungan, Semarang)
echo Destination: 50122 (Semarang Timur)
echo.
echo Product dimensions from database:
echo - Weight: 700g per item
echo - Length: 30cm
echo - Width: 25cm  
echo - Height: 10cm
echo - Quantity: 2
echo.
echo Expected volumetric weight: (30x25x10)/6000 x 2 = 2.5kg
echo Expected actual weight: 700g x 2 = 1.4kg
echo Biteship will use: 2.5kg (volumetric is higher)
echo.
pause

echo.
echo Calling API...
echo.

curl -X GET "http://localhost:8080/api/checkout/shipping-options?destination_postal_code=50122&courier=jne,sicepat,anteraja,tiki" -H "X-Session-ID: cd5b4b7c-3ecb-49cc-8aa1-1e17215c4d6b"

echo.
echo.
echo ============================================
echo CHECK BACKEND TERMINAL LOGS
echo ============================================
echo.
echo Look for these lines in backend terminal:
echo.
echo   Item 1: Denim Jacket - Weight: 700g, Dimensions: 30x25x10 cm, Qty: 2
echo.
echo If you see "Dimensions: 0x0x0", dimensions not loaded!
echo If you see "Dimensions: 30x25x10", dimensions are correct!
echo.
pause
