@echo off
echo ========================================
echo STOPPING ALL BACKEND PROCESSES...
echo ========================================
taskkill /F /IM zavera.exe 2>nul
taskkill /F /IM zavera_brand_material.exe 2>nul
taskkill /F /IM zavera_brand_material_fix.exe 2>nul
timeout /t 2 /nobreak >nul

echo.
echo ========================================
echo STARTING NEW BACKEND...
echo ========================================
cd backend
start "Zavera Backend - Brand Material Fix" zavera_brand_material_fix.exe

echo.
echo ========================================
echo Backend restarted! Check the new window.
echo Try creating product again in admin panel.
echo ========================================
pause
