@echo off
echo ========================================
echo ZAVERA - Restart Backend + Frontend
echo ========================================
echo.

echo [1/2] Starting Backend...
start "Zavera Backend" cmd /k "cd backend && zavera_variants_fixed2.exe"
timeout /t 3 /nobreak >nul

echo [2/2] Starting Frontend...
start "Zavera Frontend" cmd /k "cd frontend && npm run dev"

echo.
echo ========================================
echo Both services are starting...
echo Backend: http://localhost:8080
echo Frontend: http://localhost:3000
echo ========================================
echo.
echo Press any key to exit this window...
pause >nul
