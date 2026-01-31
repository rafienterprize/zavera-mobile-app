@echo off
echo ========================================
echo FIX CART - ALL IN ONE
echo ========================================
echo.
echo Ini akan:
echo 1. Fix database constraint
echo 2. Clear cart lama
echo 3. Restart backend dengan fix baru
echo.
echo ========================================
echo.
pause

echo.
echo [1/3] Fixing database constraint...
echo ========================================
psql -U postgres -d zavera -c "ALTER TABLE cart_items DROP CONSTRAINT IF EXISTS cart_items_cart_id_product_id_key;"

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo ❌ ERROR: Database fix gagal!
    echo.
    echo Pastikan:
    echo - PostgreSQL running
    echo - Database 'zavera' ada
    echo - User 'postgres' bisa akses
    echo.
    pause
    exit /b 1
)

echo ✅ Database constraint dihapus!
echo.

echo [2/3] Clearing old cart data...
echo ========================================
psql -U postgres -d zavera -c "DELETE FROM cart_items WHERE created_at < NOW() - INTERVAL '1 hour';"
echo ✅ Old cart data cleared!
echo.

echo [3/3] Backend akan restart...
echo ========================================
echo.
echo INSTRUKSI:
echo 1. Stop backend lama (Ctrl+C di terminal backend)
echo 2. Jalankan: start-backend-COMPLETE.bat
echo 3. Clear cart di browser: http://localhost:3000/cart
echo 4. Test add XL dan L
echo.
echo ========================================
echo ✅ Database fix SELESAI!
echo ========================================
echo.
echo Next steps:
echo 1. Stop backend lama (Ctrl+C)
echo 2. Run: start-backend-COMPLETE.bat
echo 3. Clear cart di browser
echo 4. Test add multiple variants
echo.
pause
