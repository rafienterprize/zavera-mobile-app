@echo off
echo ========================================
echo TESTING SLUG FIX - SOFT DELETE ISSUE
echo ========================================
echo.
echo PROBLEM FIXED:
echo - Deleted products no longer block slug reuse
echo - Can create product with same name after delete
echo.
echo ========================================
echo RESTARTING BACKEND...
echo ========================================
taskkill /F /IM zavera.exe 2>nul
taskkill /F /IM zavera_brand_material.exe 2>nul
taskkill /F /IM zavera_brand_material_fix.exe 2>nul
timeout /t 2 /nobreak >nul

cd backend
start "Zavera Backend - Slug Fix" zavera_brand_material_fix.exe
cd ..

echo.
echo ========================================
echo BACKEND RESTARTED!
echo ========================================
echo.
echo TEST STEPS:
echo.
echo 1. Go to: http://localhost:3000/admin/products/add
echo 2. Create product "Shirt Eiger"
echo 3. Expected: SUCCESS (no more "slug already exists" error)
echo.
echo 4. Go to products list
echo 5. Delete "Shirt Eiger"
echo 6. Create "Shirt Eiger" again
echo 7. Expected: SUCCESS (can reuse name after delete)
echo.
echo ========================================
pause
