@echo off
echo ============================================
echo MIGRATE PRODUCT DIMENSIONS
echo ============================================
echo.
echo This will add length, width, height columns to products table
echo Required for accurate Biteship shipping cost calculation
echo.
pause

echo.
echo Running migration...
echo.

psql -U postgres -d zavera_db -f database\migrate_product_dimensions.sql

echo.
echo ============================================
echo Migration completed!
echo ============================================
echo.
echo Next steps:
echo 1. Go to Admin Dashboard
echo 2. Edit your products
echo 3. Set correct dimensions (length, width, height in cm)
echo 4. Set correct weight (in grams)
echo 5. Save
echo.
echo Then test checkout to verify shipping costs
echo.
pause
