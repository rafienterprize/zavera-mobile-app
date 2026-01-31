@echo off
echo ========================================
echo ZAVERA - Shipping Hardening Migration
echo Phase 2: Fulfillment Control Layer
echo ========================================
echo.

REM Load environment variables from backend/.env
for /f "tokens=1,2 delims==" %%a in (backend\.env) do (
    if "%%a"=="DB_HOST" set DB_HOST=%%b
    if "%%a"=="DB_PORT" set DB_PORT=%%b
    if "%%a"=="DB_USER" set DB_USER=%%b
    if "%%a"=="DB_PASSWORD" set DB_PASSWORD=%%b
    if "%%a"=="DB_NAME" set DB_NAME=%%b
)

REM Set defaults if not found
if "%DB_HOST%"=="" set DB_HOST=localhost
if "%DB_PORT%"=="" set DB_PORT=5432
if "%DB_USER%"=="" set DB_USER=postgres
if "%DB_NAME%"=="" set DB_NAME=zavera

echo Database: %DB_NAME%@%DB_HOST%:%DB_PORT%
echo.

REM Set PGPASSWORD for psql
set PGPASSWORD=%DB_PASSWORD%

echo Running shipping hardening migration...
psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% -f database/migrate_shipping_hardening.sql

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ========================================
    echo Migration completed successfully!
    echo ========================================
    echo.
    echo New tables created:
    echo   - disputes
    echo   - dispute_messages
    echo   - courier_failure_log
    echo   - shipment_status_history
    echo   - shipment_alerts
    echo.
    echo Shipment table updated with:
    echo   - New status values (15 statuses)
    echo   - Pickup control columns
    echo   - Tracking control columns
    echo   - Investigation columns
    echo   - Reship tracking columns
    echo.
    echo New admin endpoints available:
    echo   POST /api/admin/shipments/:id/investigate
    echo   POST /api/admin/shipments/:id/mark-lost
    echo   POST /api/admin/shipments/:id/reship
    echo   POST /api/admin/shipments/:id/override-status
    echo   POST /api/admin/disputes
    echo   GET  /api/admin/disputes/open
    echo   POST /api/admin/disputes/:id/resolve
    echo   GET  /api/admin/fulfillment/dashboard
    echo.
) else (
    echo.
    echo ========================================
    echo Migration FAILED!
    echo ========================================
    echo Please check the error messages above.
)

pause
