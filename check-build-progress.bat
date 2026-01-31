@echo off
echo ========================================
echo   CHECKING BUILD PROGRESS
echo ========================================
echo.
echo Build is running in background...
echo This can take 3-5 minutes for first build.
echo.
echo Checking Gradle process...
tasklist | findstr /i "java.exe gradle"
echo.
echo ========================================
echo The app will automatically:
echo   1. Build APK (3-5 min)
echo   2. Install on your phone
echo   3. Launch automatically
echo.
echo Just wait and watch your phone!
echo ========================================
pause
