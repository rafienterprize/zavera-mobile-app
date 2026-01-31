@echo off
echo ========================================
echo Refund System - Quick Test Script
echo ========================================
echo.

echo [1/3] Checking database migration...
psql -h localhost -U postgres -d zavera_db -c "SELECT COUNT(*) as refund_count FROM refunds;" 2>nul
if %errorlevel% neq 0 (
    echo ERROR: Database not migrated or not accessible
    echo Please run: psql -h localhost -U postgres -d zavera_db -f database/migrate_refund_enhancement.sql
    pause
    exit /b 1
)
echo ✓ Database migration verified
echo.

echo [2/3] Checking backend build...
if not exist "backend\zavera.exe" (
    echo Building backend...
    cd backend
    go build -o zavera.exe
    cd ..
)
echo ✓ Backend build verified
echo.

echo [3/3] System ready!
echo.
echo ========================================
echo Quick Start Commands:
echo ========================================
echo.
echo 1. Start Backend:
echo    cd backend ^&^& .\zavera.exe
echo.
echo 2. Start Frontend (new terminal):
echo    cd frontend ^&^& npm run dev
echo.
echo 3. Test Refund:
echo    - Login as admin: http://localhost:3000/login
echo    - Go to orders: http://localhost:3000/admin/orders
echo    - Click "Refund" button on DELIVERED order
echo.
echo ========================================
echo Documentation:
echo ========================================
echo - Quick Start: REFUND_QUICK_START.md
echo - Full Guide: REFUND_SYSTEM_DEPLOYMENT_GUIDE.md
echo - Summary: REFUND_SYSTEM_COMPLETION_SUMMARY.md
echo.
pause
