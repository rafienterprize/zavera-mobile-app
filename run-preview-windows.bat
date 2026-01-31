@echo off
echo ========================================
echo   ZAVERA Mobile - Windows Preview
echo ========================================
echo.

echo Checking Developer Mode...
echo.

cd zavera_mobile

echo Starting Flutter app on Windows...
echo This will open a window showing mobile app preview.
echo.
echo Press 'r' for hot reload
echo Press 'q' to quit
echo.

flutter run -d windows

pause
