@echo off
echo ====================================
echo ZAVERA - Category Migration
echo ====================================
echo.

set PGPASSWORD=Yan2692009

echo Running category migration...
psql -U postgres -h localhost -d zavera_db -f "database\migrate_categories.sql"

echo.
echo ====================================
echo Migration completed!
echo ====================================
echo.
echo New categories added:
echo - Wanita (Women's Fashion)
echo - Pria (Men's Fashion)
echo - Anak (Kids)
echo - Sports
echo - Luxury
echo - Beauty
echo.
echo Please restart the backend server to apply changes.
echo.
pause
