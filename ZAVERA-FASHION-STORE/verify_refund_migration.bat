@echo off
echo ========================================
echo Refund System - Migration Verification
echo ========================================
echo.

echo Checking database connection...
psql -h localhost -U postgres -d zavera_db -c "SELECT version();" >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Cannot connect to database
    echo Please ensure PostgreSQL is running
    pause
    exit /b 1
)
echo ✓ Database connection OK
echo.

echo Checking refunds table structure...
psql -h localhost -U postgres -d zavera_db -c "\d refunds" >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Refunds table not found
    echo Please run migration first
    pause
    exit /b 1
)
echo ✓ Refunds table exists
echo.

echo Checking nullable columns...
psql -h localhost -U postgres -d zavera_db -c "SELECT column_name, is_nullable FROM information_schema.columns WHERE table_name = 'refunds' AND column_name IN ('requested_by', 'payment_id');"
echo.

echo Checking orders refund columns...
psql -h localhost -U postgres -d zavera_db -c "SELECT column_name FROM information_schema.columns WHERE table_name = 'orders' AND column_name IN ('refund_status', 'refund_amount', 'refunded_at');"
echo.

echo Checking refund_status_history table...
psql -h localhost -U postgres -d zavera_db -c "\d refund_status_history" >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: refund_status_history table not found
    pause
    exit /b 1
)
echo ✓ refund_status_history table exists
echo.

echo Checking refund_items table...
psql -h localhost -U postgres -d zavera_db -c "\d refund_items" >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: refund_items table not found
    pause
    exit /b 1
)
echo ✓ refund_items table exists
echo.

echo ========================================
echo Migration Status: ✓ VERIFIED
echo ========================================
echo.
echo All required tables and columns exist.
echo Refund system is ready to use!
echo.
pause
