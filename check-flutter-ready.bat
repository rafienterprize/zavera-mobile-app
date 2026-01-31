@echo off
echo ========================================
echo   ZAVERA Mobile - Flutter Check
echo ========================================
echo.

echo [1/4] Checking Flutter installation...
flutter --version
if %errorlevel% neq 0 (
    echo [ERROR] Flutter not found! Please install Flutter first.
    echo See INSTALL_FLUTTER.md for instructions.
    pause
    exit /b 1
)
echo [OK] Flutter installed!
echo.

echo [2/4] Running Flutter Doctor...
flutter doctor
echo.

echo [3/4] Checking connected devices...
flutter devices
echo.

echo [4/4] Checking ADB devices...
adb devices
echo.

echo ========================================
echo   Check Complete!
echo ========================================
echo.
echo Next steps:
echo 1. Make sure your phone is connected via USB
echo 2. Enable USB Debugging on your phone
echo 3. Run: cd zavera_mobile
echo 4. Run: flutter pub get
echo 5. Run: flutter run
echo.
pause
