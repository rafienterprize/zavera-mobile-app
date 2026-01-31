@echo off
echo Running Payment Immutability Migration...
psql -h localhost -U postgres -d zavera_db -f database/migrate_payment_immutability.sql
echo Migration complete!
pause
