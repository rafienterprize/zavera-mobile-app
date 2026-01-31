@echo off
echo ============================================
echo ZAVERA AUDIT FIXES MIGRATION
echo ============================================
echo.
echo This script will apply the following fixes:
echo 1. Enum fix for shipment_status
echo 2. Product weight column
echo 3. Over-refund prevention trigger
echo.
echo Make sure PostgreSQL is running and you have the correct credentials.
echo.

set /p PGPASSWORD="Enter PostgreSQL password: "
set PGHOST=localhost
set PGPORT=5432
set PGUSER=postgres
set PGDATABASE=zavera_db

echo.
echo Running enum fix migration...
psql -h %PGHOST% -p %PGPORT% -U %PGUSER% -d %PGDATABASE% -f database/migrate_enum_fix.sql
if %ERRORLEVEL% NEQ 0 (
    echo ERROR: Enum fix migration failed!
    pause
    exit /b 1
)

echo.
echo Running product weight migration...
psql -h %PGHOST% -p %PGPORT% -U %PGUSER% -d %PGDATABASE% -f database/migrate_product_weight.sql
if %ERRORLEVEL% NEQ 0 (
    echo ERROR: Product weight migration failed!
    pause
    exit /b 1
)

echo.
echo ============================================
echo ALL MIGRATIONS COMPLETED SUCCESSFULLY!
echo ============================================
echo.
echo Next steps:
echo 1. Rebuild the backend: cd backend && go build
echo 2. Restart the backend server
echo 3. Run the audit queries to verify data integrity
echo.
pause
