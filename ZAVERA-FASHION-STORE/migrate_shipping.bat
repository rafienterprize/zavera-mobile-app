@echo off
echo Running Shipping System Migration...
echo.

set PGPASSWORD=Yan2692009
psql -h localhost -U postgres -d zavera_db -f database/migrate_shipping.sql

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✅ Shipping migration completed successfully!
) else (
    echo.
    echo ❌ Migration failed!
)

pause
