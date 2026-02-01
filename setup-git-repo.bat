@echo off
echo ========================================
echo   ZAVERA Project - Git Setup
echo ========================================
echo.

REM Check if git is installed
git --version >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Git not installed!
    echo Please install Git from: https://git-scm.com/download/win
    pause
    exit /b 1
)

echo [OK] Git is installed
echo.

REM Check if already initialized
if exist .git (
    echo [INFO] Git repository already initialized
    echo.
    goto :configure
)

echo [1/5] Initializing Git repository...
git init
git branch -M main
echo [OK] Repository initialized
echo.

:configure
echo [2/5] Configuring Git...
set /p USERNAME="Enter your name: "
set /p EMAIL="Enter your email: "

git config user.name "%USERNAME%"
git config user.email "%EMAIL%"
echo [OK] Git configured
echo.

echo [3/5] Adding files...
git add .
echo [OK] Files added
echo.

echo [4/5] Creating initial commit...
git commit -m "Initial commit: ZAVERA Fashion Store project"
echo [OK] Commit created
echo.

echo [5/5] Setup remote (optional)...
echo.
echo To connect to GitHub:
echo 1. Create a new repository on GitHub
echo 2. Copy the repository URL
echo 3. Run: git remote add origin YOUR_REPO_URL
echo 4. Run: git push -u origin main
echo.

echo ========================================
echo   Git Setup Complete!
echo ========================================
echo.
echo Next steps:
echo 1. Create GitHub repository
echo 2. git remote add origin https://github.com/USERNAME/zavera-project.git
echo 3. git push -u origin main
echo.
echo Read GIT_WORKFLOW_GUIDE.md for more info!
echo.
pause
