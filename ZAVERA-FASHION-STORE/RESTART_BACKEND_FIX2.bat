@echo off
echo ========================================
echo RESTARTING BACKEND (FIX 2)
echo ========================================
echo.

echo Stopping any running backend processes...
taskkill /F /IM zavera*.exe 2>nul
timeout /t 2 /nobreak >nul

echo.
echo Starting backend with brand/material fix 2...
cd backend
start "Zavera Backend Fix 2" zavera_brand_material_fix2.exe

echo.
echo ========================================
echo Backend started!
echo Check the new window for logs
echo ========================================
pause
