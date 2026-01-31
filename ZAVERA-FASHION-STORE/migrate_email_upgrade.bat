@echo off
echo ============================================
echo ZAVERA Email System Upgrade Migration
echo ============================================
echo.

set PGPASSWORD=Yan2692009
psql -h localhost -U postgres -d zavera_db -f database/migrate_email_upgrade.sql

echo.
echo Migration completed!
pause
