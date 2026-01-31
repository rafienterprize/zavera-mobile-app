@echo off
echo ========================================
echo Migrating Subdistricts (Kecamatan) Data
echo ========================================

set PGPASSWORD=postgres
psql -h localhost -U postgres -d zavera -f database/migrate_subdistricts.sql

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ========================================
    echo Migration completed successfully!
    echo Subdistricts data has been added.
    echo ========================================
) else (
    echo.
    echo ========================================
    echo Migration failed! Check the error above.
    echo ========================================
)

pause
