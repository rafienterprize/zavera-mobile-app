@echo off
echo ============================================
echo CLEAR USER DATA - ZAVERA DATABASE
echo ============================================
echo.
echo WARNING: Ini akan menghapus SEMUA data:
echo - Users
echo - Carts dan Cart Items
echo - Orders dan Order Items
echo - Payments
echo - Sessions dan Tokens
echo.
echo Data PRODUCTS dan PRODUCT_IMAGES TIDAK akan dihapus.
echo.
set /p confirm="Yakin ingin melanjutkan? (y/n): "
if /i not "%confirm%"=="y" (
    echo Dibatalkan.
    exit /b 0
)

echo.
echo Menjalankan clear_user_data.sql...
psql -U postgres -d zavera -f database/clear_user_data.sql

echo.
echo Selesai!
pause
