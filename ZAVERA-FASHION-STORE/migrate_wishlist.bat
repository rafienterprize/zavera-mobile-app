@echo off
echo Running Wishlist Migration...
psql -U postgres -d zavera_db -f database/migrate_wishlist.sql
if %errorlevel% equ 0 (
    echo Migration completed successfully!
) else (
    echo Migration failed!
)
pause
