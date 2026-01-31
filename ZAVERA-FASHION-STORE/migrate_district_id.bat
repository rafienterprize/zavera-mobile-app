@echo off
echo Running district_id migration...
psql -U postgres -d zavera_db -f database/migrate_district_id.sql
echo Migration complete!
pause
