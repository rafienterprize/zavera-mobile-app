@echo off
echo ========================================
echo Creating Sample Data for Dashboard
echo ========================================
echo.

echo Creating sample shipments...
psql -U postgres -d zavera_db -f database\create_sample_shipments.sql
echo.

echo Creating sample audit logs...
psql -U postgres -d zavera_db -f database\create_sample_audit_logs.sql
echo.

echo ========================================
echo Sample data created successfully!
echo ========================================
echo.
echo Now refresh your browser to see:
echo - Shipments page with data
echo - Audit logs with entries
echo - Courier performance analytics
echo.
pause
