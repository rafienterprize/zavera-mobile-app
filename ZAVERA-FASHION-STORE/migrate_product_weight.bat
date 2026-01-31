@echo off
echo Running product weight migration (PostgreSQL)...
set PGPASSWORD=Yan2692009
psql -U postgres -d zavera_db -f database/migrate_product_weight.sql
echo Migration complete!
pause
