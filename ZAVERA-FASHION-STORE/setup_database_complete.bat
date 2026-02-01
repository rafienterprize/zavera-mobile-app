@echo off
echo ========================================
echo ZAVERA Database Setup - Complete
echo ========================================
echo.

set PSQL_PATH=D:\postgresql\PostgreSQL\16\bin\psql.exe
set DB_NAME=zavera_db
set DB_USER=postgres

echo Step 1: Running main schema...
%PSQL_PATH% -U %DB_USER% -d %DB_NAME% -f database\schema.sql
if %errorlevel% neq 0 (
    echo Warning: Schema might already exist, continuing...
)

echo.
echo Step 2: Running migrations...
%PSQL_PATH% -U %DB_USER% -d %DB_NAME% -f database\migrate.sql
%PSQL_PATH% -U %DB_USER% -d %DB_NAME% -f database\migrate_categories.sql
%PSQL_PATH% -U %DB_USER% -d %DB_NAME% -f database\migrate_product_variants.sql
%PSQL_PATH% -U %DB_USER% -d %DB_NAME% -f database\migrate_brand_material.sql
%PSQL_PATH% -U %DB_USER% -d %DB_NAME% -f database\migrate_wishlist.sql
%PSQL_PATH% -U %DB_USER% -d %DB_NAME% -f database\migrate_shipping.sql
%PSQL_PATH% -U %DB_USER% -d %DB_NAME% -f database\migrate_biteship.sql

echo.
echo Step 3: Inserting sample data...
%PSQL_PATH% -U %DB_USER% -d %DB_NAME% -f database\init.sql

echo.
echo ========================================
echo Database setup complete!
echo ========================================
echo.
echo You can now:
echo 1. Start backend: cd backend ^&^& go run main.go
echo 2. Check products: http://localhost:8080/api/products
echo.
pause
