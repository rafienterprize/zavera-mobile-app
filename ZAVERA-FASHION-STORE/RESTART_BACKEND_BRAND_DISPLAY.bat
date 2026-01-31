@echo off
echo ========================================
echo RESTARTING BACKEND (Brand Display Fix)
echo ========================================
echo.

echo Stopping any running backend processes...
taskkill /F /IM zavera*.exe 2>nul
timeout /t 2 /nobreak >nul

echo.
echo Starting backend with brand/material display fix...
cd backend
start "Zavera Backend - Brand Display" zavera_brand_material_display.exe

echo.
echo ========================================
echo Backend started!
echo Check the new window for logs
echo ========================================
echo.
echo Now test: http://localhost:3000/product/60
echo Brand and Material should appear!
echo.
pause
