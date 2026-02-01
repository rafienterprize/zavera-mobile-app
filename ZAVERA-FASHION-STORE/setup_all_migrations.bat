@echo off
echo ========================================
echo ZAVERA - Run ALL Migrations
echo ========================================
echo.

set PSQL_PATH=D:\postgresql\PostgreSQL\16\bin\psql.exe
set DB_NAME=zavera_db
set DB_USER=postgres

echo Running ALL migration files...
echo.

REM Core migrations
echo [1/20] migrate.sql
%PSQL_PATH% -U %DB_USER% -d %DB_NAME% -f database\migrate.sql 2>nul

echo [2/20] migrate_auth.sql
%PSQL_PATH% -U %DB_USER% -d %DB_NAME% -f database\migrate_auth.sql 2>nul

echo [3/20] migrate_categories.sql
%PSQL_PATH% -U %DB_USER% -d %DB_NAME% -f database\migrate_categories.sql 2>nul

echo [4/20] migrate_product_variants.sql
%PSQL_PATH% -U %DB_USER% -d %DB_NAME% -f database\migrate_product_variants.sql 2>nul

echo [5/20] migrate_product_dimensions.sql
%PSQL_PATH% -U %DB_USER% -d %DB_NAME% -f database\migrate_product_dimensions.sql 2>nul

echo [6/20] migrate_product_weight.sql
%PSQL_PATH% -U %DB_USER% -d %DB_NAME% -f database\migrate_product_weight.sql 2>nul

echo [7/20] migrate_brand_material.sql
%PSQL_PATH% -U %DB_USER% -d %DB_NAME% -f database\migrate_brand_material.sql 2>nul

echo [8/20] migrate_wishlist.sql
%PSQL_PATH% -U %DB_USER% -d %DB_NAME% -f database\migrate_wishlist.sql 2>nul

echo [9/20] migrate_shipping.sql
%PSQL_PATH% -U %DB_USER% -d %DB_NAME% -f database\migrate_shipping.sql 2>nul

echo [10/20] migrate_biteship.sql
%PSQL_PATH% -U %DB_USER% -d %DB_NAME% -f database\migrate_biteship.sql 2>nul

echo [11/20] migrate_core_payments.sql
%PSQL_PATH% -U %DB_USER% -d %DB_NAME% -f database\migrate_core_payments.sql 2>nul

echo [12/20] migrate_qris_gopay.sql
%PSQL_PATH% -U %DB_USER% -d %DB_NAME% -f database\migrate_qris_gopay.sql 2>nul

echo [13/20] migrate_payment_immutability.sql
%PSQL_PATH% -U %DB_USER% -d %DB_NAME% -f database\migrate_payment_immutability.sql 2>nul

echo [14/20] migrate_refund_enhancement.sql
%PSQL_PATH% -U %DB_USER% -d %DB_NAME% -f database\migrate_refund_enhancement.sql 2>nul

echo [15/20] migrate_subdistricts.sql
%PSQL_PATH% -U %DB_USER% -d %DB_NAME% -f database\migrate_subdistricts.sql 2>nul

echo [16/20] migrate_hardening.sql
%PSQL_PATH% -U %DB_USER% -d %DB_NAME% -f database\migrate_hardening.sql 2>nul

echo [17/20] migrate_shipping_hardening.sql
%PSQL_PATH% -U %DB_USER% -d %DB_NAME% -f database\migrate_shipping_hardening.sql 2>nul

echo [18/20] migrate_audit_fixes.sql
%PSQL_PATH% -U %DB_USER% -d %DB_NAME% -f database\migrate_audit_fixes.sql 2>nul

echo [19/20] migrate_commerce_upgrade.sql
%PSQL_PATH% -U %DB_USER% -d %DB_NAME% -f database\migrate_commerce_upgrade.sql 2>nul

echo [20/20] migrate_email_upgrade.sql
%PSQL_PATH% -U %DB_USER% -d %DB_NAME% -f database\migrate_email_upgrade.sql 2>nul

echo.
echo ========================================
echo All migrations completed!
echo ========================================
echo.
echo Now restart backend:
echo   cd backend
echo   go run main.go
echo.
pause
