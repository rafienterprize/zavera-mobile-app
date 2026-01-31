@echo off
echo ========================================
echo   CHECKING PHONE CONNECTION
echo ========================================
echo.

echo Restarting ADB server...
adb kill-server
adb start-server
echo.

echo Checking connected devices...
echo.
flutter devices
echo.

echo ========================================
echo If you see your phone listed above, you're ready!
echo If not, check:
echo   1. USB cable is properly connected
echo   2. USB debugging is allowed on phone
echo   3. USB mode is set to "File Transfer"
echo ========================================
echo.
pause
