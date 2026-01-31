@echo off
echo Running QRIS and GoPay payment method migration...
echo.

set PGPASSWORD=Bismillah

psql -h localhost -U postgres -d zavera -f database/migrate_qris_gopay.sql

echo.
echo Migration completed!
pause
