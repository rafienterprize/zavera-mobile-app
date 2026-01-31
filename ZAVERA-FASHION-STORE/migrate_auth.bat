@echo off
echo Running Auth Migration...
set PGPASSWORD=Yan2692009
psql -h localhost -U postgres -d zavera_db -f database/migrate_auth.sql
echo Migration complete!
pause
