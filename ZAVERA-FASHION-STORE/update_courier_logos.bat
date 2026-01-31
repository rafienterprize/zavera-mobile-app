@echo off
echo ============================================
echo UPDATE COURIER LOGOS
echo ============================================
echo.
echo This script will update courier logos in database
echo.
pause

echo.
echo Updating database...
echo.

psql -U postgres -d zavera_db -f database\update_courier_logos.sql

echo.
echo ============================================
echo Courier logos updated!
echo ============================================
echo.
echo New logos added:
echo - Ninja Express: /images/couriers/ninja.png
echo - RPX: /images/couriers/rpx.png
echo - SAP Express: /images/couriers/sap.png
echo - ID Express: /images/couriers/idexpress.png
echo - Lion Parcel: /images/couriers/lion.png
echo.
echo Please refresh your browser to see the new logos!
echo.
pause
