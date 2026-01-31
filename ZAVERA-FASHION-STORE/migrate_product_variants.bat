@echo off
psql -U postgres -d zavera_db -f database/migrate_product_variants.sql
echo Product variants migration completed!
pause
