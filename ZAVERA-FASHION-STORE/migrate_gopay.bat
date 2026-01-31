@echo off
echo Running GoPay migration...
cd database
psql -U postgres -d zavera -f migrate_gopay.sql
echo Migration complete!
pause
