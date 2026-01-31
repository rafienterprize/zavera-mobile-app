@echo off
echo ========================================
echo Starting Zavera Backend - COMPLETE FIX
echo ========================================
echo.
echo All Fixes Applied:
echo 1. Cart variant stock check fixed
echo 2. Cart metadata comparison fixed  
echo 3. Cart validation skip for variants
echo 4. Multiple variants supported
echo.
echo IMPORTANT: Run fix_cart_database.bat first!
echo.
echo ========================================
echo.

cd backend
zavera_COMPLETE_FIX.exe
