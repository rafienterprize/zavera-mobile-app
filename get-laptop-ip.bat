@echo off
echo ========================================
echo   GET LAPTOP IP ADDRESS
echo ========================================
echo.
echo Your WiFi IP Address:
echo.
powershell -Command "(Get-NetIPAddress -AddressFamily IPv4 | Where-Object {$_.InterfaceAlias -like '*Wi-Fi*' -or $_.InterfaceAlias -like '*Wireless*'}).IPAddress"
echo.
echo ========================================
echo Use this IP in api_service.dart:
echo   http://YOUR_IP:8080/api
echo.
echo Example:
echo   http://192.168.1.100:8080/api
echo ========================================
echo.
pause
