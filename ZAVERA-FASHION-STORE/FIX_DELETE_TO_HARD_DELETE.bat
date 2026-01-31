@echo off
echo ========================================
echo FIX: CHANGE DELETE TO HARD DELETE
echo ========================================
echo.
echo CHANGES:
echo 1. Delete function now HARD DELETE (permanent)
echo 2. Products will be removed from database
echo 3. No more inactive products cluttering database
echo.
echo ========================================
echo STEP 1: CLEANUP EXISTING INACTIVE PRODUCTS
echo ========================================
echo.

set PGPASSWORD=Yan2692009

echo Showing inactive products...
psql -U postgres -d zavera_db -c "SELECT id, name, slug, is_active FROM products WHERE is_active = false ORDER BY id;"

echo.
echo Do you want to DELETE these inactive products? (Y/N)
set /p confirm=

if /i "%confirm%"=="Y" (
    echo.
    echo Deleting inactive products...
    
    psql -U postgres -d zavera_db -c "BEGIN; DELETE FROM product_images WHERE product_id IN (SELECT id FROM products WHERE is_active = false); DELETE FROM product_variants WHERE product_id IN (SELECT id FROM products WHERE is_active = false); DELETE FROM products WHERE is_active = false; COMMIT;"
    
    echo.
    echo ✅ Inactive products deleted!
    
    echo.
    echo Verifying...
    psql -U postgres -d zavera_db -c "SELECT COUNT(*) as remaining_inactive FROM products WHERE is_active = false;"
) else (
    echo.
    echo Skipping cleanup...
)

echo.
echo ========================================
echo STEP 2: RESTART BACKEND WITH NEW CODE
echo ========================================
echo.

taskkill /F /IM zavera.exe 2>nul
taskkill /F /IM zavera_brand_material.exe 2>nul
taskkill /F /IM zavera_brand_material_fix.exe 2>nul
timeout /t 2 /nobreak >nul

cd backend
start "Zavera Backend - Hard Delete" zavera_brand_material_fix.exe
cd ..

echo.
echo ========================================
echo DONE!
echo ========================================
echo.
echo Backend restarted with HARD DELETE function.
echo.
echo TEST:
echo 1. Go to admin products
echo 2. Delete a product
echo 3. Check database - product should be GONE
echo 4. Create product with same name - should work!
echo.
pause
