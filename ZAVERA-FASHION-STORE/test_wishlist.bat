@echo off
echo ========================================
echo ZAVERA WISHLIST FEATURE TEST
echo ========================================
echo.

echo Testing Wishlist Endpoints...
echo.

echo 1. Testing GET /api/wishlist (requires authentication)
echo    Expected: 401 Unauthorized (no token) or wishlist data (with valid token)
curl -X GET http://localhost:8080/api/wishlist
echo.
echo.

echo 2. Testing POST /api/wishlist (requires authentication)
echo    Expected: 401 Unauthorized (no token)
curl -X POST http://localhost:8080/api/wishlist -H "Content-Type: application/json" -d "{\"product_id\": 1}"
echo.
echo.

echo 3. Testing Health Check
curl -X GET http://localhost:8080/health
echo.
echo.

echo ========================================
echo TEST COMPLETE
echo ========================================
echo.
echo NOTES:
echo - Backend must be running on port 8080
echo - Wishlist endpoints require authentication
echo - To test with authentication, login first and use the token
echo.
pause
