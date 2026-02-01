@echo off
echo ========================================
echo   Installing Android SDK
echo ========================================
echo.
echo This will install Android SDK (lightweight, ~500MB)
echo.
pause

powershell -Command "Start-Process powershell -ArgumentList '-ExecutionPolicy Bypass -File \"%~dp0install-android-sdk-only.ps1\"' -Verb RunAs"
