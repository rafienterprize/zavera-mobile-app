@echo off
echo ============================================
echo ZAVERA Commercial Hardening Migration
echo Phase 1: Critical Safety Layer
echo ============================================
echo.

set PGPASSWORD=postgres

echo Running hardening migration...
psql -U postgres -d zavera -f database/migrate_hardening.sql

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ============================================
    echo Migration completed successfully!
    echo ============================================
    echo.
    echo New tables created:
    echo   - refunds
    echo   - refund_items
    echo   - admin_audit_log
    echo   - payment_sync_log
    echo   - reconciliation_log
    echo   - refund_status_history
    echo.
    echo New columns added to:
    echo   - orders (refund_status, refund_amount, etc.)
    echo   - payments (refund_status, refunded_amount, etc.)
    echo   - shipments (reship_count, is_replacement, etc.)
    echo.
) else (
    echo.
    echo ============================================
    echo Migration FAILED!
    echo ============================================
    echo Please check the error messages above.
)

pause
