@echo off
echo ====================================
echo Setup Database ZAVERA
echo ====================================
echo.

set PGPASSWORD=Yan2692009

echo [1/2] Membuat database zavera_db...
psql -U postgres -h localhost -c "DROP DATABASE IF EXISTS zavera_db;"
psql -U postgres -h localhost -c "CREATE DATABASE zavera_db;"

echo.
echo [2/2] Import data produk...
psql -U postgres -h localhost -d zavera_db -f "database\init.sql"

echo.
echo ====================================
echo Setup database selesai!
echo ====================================
pause
