@echo off
echo Testing New Dashboard Endpoints...
echo.

echo 1. Testing System Health Endpoint...
curl -X GET "http://localhost:8080/api/admin/system/health" -H "Authorization: Bearer YOUR_TOKEN_HERE"
echo.
echo.

echo 2. Testing Courier Performance Endpoint...
curl -X GET "http://localhost:8080/api/admin/analytics/courier-performance" -H "Authorization: Bearer YOUR_TOKEN_HERE"
echo.
echo.

echo 3. Testing Shipments List Endpoint...
curl -X GET "http://localhost:8080/api/admin/shipments?page=1" -H "Authorization: Bearer YOUR_TOKEN_HERE"
echo.
echo.

echo 4. Testing Executive Dashboard with Previous Period...
curl -X GET "http://localhost:8080/api/admin/dashboard/executive?period=month" -H "Authorization: Bearer YOUR_TOKEN_HERE"
echo.
echo.

echo All tests complete!
pause
