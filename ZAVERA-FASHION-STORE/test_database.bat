@echo off
echo ========================================
echo ZAVERA DATABASE INTEGRITY TESTING
echo ========================================
echo.
echo Running database integrity checks...
echo.

set PGPASSWORD=Yan2692009
psql -U postgres -d zavera_db -f database\test_database_integrity.sql

echo.
echo ========================================
echo Press any key to exit...
pause > nul
