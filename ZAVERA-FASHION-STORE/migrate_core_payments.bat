@echo off
echo Running Core Payments Migration...
echo.

set PGPASSWORD=postgres
psql -U postgres -d zavera -f database/migrate_core_payments.sql

echo.
echo Migration complete!
pause
