@echo off
echo ========================================
echo  Add Brand and Material Columns
echo ========================================
echo.

set PGPASSWORD=Yan2692009
psql -U postgres -d zavera_db -f database/migrate_brand_material.sql

echo.
echo ========================================
echo  Migration Complete!
echo ========================================
echo.
echo Brand and Material columns have been added to products table.
echo.
pause
