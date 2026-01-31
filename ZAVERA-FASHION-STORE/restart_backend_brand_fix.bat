@echo off
echo ========================================
echo STOPPING OLD BACKEND...
echo ========================================
taskkill /F /IM zavera.exe 2>nul
taskkill /F /IM zavera_brand_material.exe 2>nul
taskkill /F /IM zavera_brand_material_fix.exe 2>nul
timeout /t 2 /nobreak >nul

echo.
echo ========================================
echo STARTING NEW BACKEND WITH BRAND/MATERIAL FIX...
echo ========================================
cd backend
start "Zavera Backend - Brand Material Fix" zavera_brand_material_fix.exe

echo.
echo ========================================
echo Backend started! Check the new window for logs.
echo ========================================
pause
