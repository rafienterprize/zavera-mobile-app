@echo off
echo ========================================
echo TESTING PRODUCT CREATION ERROR HANDLING
echo ========================================
echo.
echo This will:
echo 1. Restart backend with new error handling
echo 2. You need to refresh browser (Ctrl+Shift+R)
echo 3. Try creating product with duplicate name
echo.
pause

echo.
echo ========================================
echo STEP 1: STOPPING OLD BACKEND...
echo ========================================
taskkill /F /IM zavera.exe 2>nul
taskkill /F /IM zavera_brand_material.exe 2>nul
taskkill /F /IM zavera_brand_material_fix.exe 2>nul
timeout /t 2 /nobreak >nul

echo.
echo ========================================
echo STEP 2: STARTING NEW BACKEND...
echo ========================================
cd backend
start "Zavera Backend - Error Handling Fix" zavera_brand_material_fix.exe
cd ..

echo.
echo ========================================
echo BACKEND RESTARTED!
echo ========================================
echo.
echo NEXT STEPS:
echo 1. Go to browser: http://localhost:3000/admin/products/add
echo 2. Press Ctrl+Shift+R to hard refresh (clear cache)
echo 3. Try creating product with name "Shirt Elper V2 22"
echo 4. You should see clear error message popup
echo 5. Try with different name - should work!
echo.
echo ========================================
pause
