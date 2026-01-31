@echo off
echo ========================================
echo Fix Cart Database Constraint
echo ========================================
echo.
echo This will:
echo 1. Remove old unique constraint
echo 2. Allow multiple variants of same product
echo 3. Clean up duplicate items
echo.
echo ========================================
echo.

psql -U postgres -d zavera -f database/fix_cart_constraint.sql

echo.
echo ========================================
echo Done! Now restart backend with:
echo   start-backend-COMPLETE.bat
echo ========================================
pause
