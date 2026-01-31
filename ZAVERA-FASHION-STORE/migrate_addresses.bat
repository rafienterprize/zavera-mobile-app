@echo off
echo Adding area_id and area_name columns to user_addresses table...

set PGPASSWORD=Yan2692009
psql -U postgres -d zavera_db -c "ALTER TABLE user_addresses ADD COLUMN IF NOT EXISTS area_id VARCHAR(100);"
psql -U postgres -d zavera_db -c "ALTER TABLE user_addresses ADD COLUMN IF NOT EXISTS area_name VARCHAR(500);"

echo Migration complete!
pause
