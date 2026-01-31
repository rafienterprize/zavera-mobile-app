@echo off
echo ========================================
echo   ZAVERA MOBILE APP - RUN ON PHONE
echo ========================================
echo.

cd zavera_mobile

echo Checking for connected devices...
flutter devices
echo.

echo ========================================
echo Starting app build and installation...
echo This will take 3-5 minutes on first run
echo ========================================
echo.

flutter run

pause
