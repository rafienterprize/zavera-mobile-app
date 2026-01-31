@echo off
echo ============================================
echo ZAVERA Commerce Platform Upgrade Migration
echo ============================================
echo.

set PGPASSWORD=postgres
psql -h localhost -U postgres -d zavera -f database/migrate_commerce_upgrade.sql

echo.
echo Migration completed!
pause
