@echo off
echo ============================================
echo BITESHIP MIGRATION - RajaOngkir to Biteship
echo ============================================
echo.

REM Load environment variables
for /f "tokens=1,2 delims==" %%a in (backend\.env) do (
    if not "%%a"=="" if not "%%b"=="" set %%a=%%b
)

echo Running Biteship migration...
psql -h %DB_HOST% -U %DB_USER% -d %DB_NAME% -f database/migrate_biteship.sql

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✅ Migration completed successfully!
    echo.
    echo Next steps:
    echo 1. Verify TOKEN_BITESHIP and BITESHIP_BASE_URL in backend/.env
    echo 2. Test Biteship API connection
    echo 3. Update application code to use Biteship
) else (
    echo.
    echo ❌ Migration failed with error code %ERRORLEVEL%
    echo Please check the error messages above
)

echo.
pause
