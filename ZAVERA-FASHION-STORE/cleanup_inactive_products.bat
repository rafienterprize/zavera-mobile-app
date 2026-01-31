@echo off
echo ========================================
echo CLEANUP INACTIVE PRODUCTS
echo ========================================
echo.
echo This will show all inactive products in database.
echo.
pause

set PGPASSWORD=Yan2692009
psql -U postgres -d zavera_db -f database/cleanup_inactive_products.sql

echo.
echo ========================================
echo REVIEW THE LIST ABOVE
echo ========================================
echo.
echo If you want to DELETE these products permanently:
echo 1. Open database/cleanup_inactive_products.sql
echo 2. Uncomment the DELETE section (remove /* and */)
echo 3. Run this script again
echo.
echo WARNING: Deletion is PERMANENT and cannot be undone!
echo.
pause
