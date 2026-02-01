@echo off
echo ========================================
echo   ZAVERA - Daily Git Workflow Helper
echo ========================================
echo.

:menu
echo What do you want to do?
echo.
echo 1. Start new feature (create branch)
echo 2. Commit changes
echo 3. Push to GitHub
echo 4. Pull latest changes
echo 5. Create Pull Request (opens browser)
echo 6. Switch branch
echo 7. View status
echo 8. Exit
echo.
set /p choice="Enter choice (1-8): "

if "%choice%"=="1" goto :new_feature
if "%choice%"=="2" goto :commit
if "%choice%"=="3" goto :push
if "%choice%"=="4" goto :pull
if "%choice%"=="5" goto :pr
if "%choice%"=="6" goto :switch
if "%choice%"=="7" goto :status
if "%choice%"=="8" goto :end
goto :menu

:new_feature
echo.
echo Creating new feature branch...
echo.
echo Branch types:
echo - mobile/feature-name
echo - backend/feature-name
echo - frontend/feature-name
echo - fix/bug-name
echo.
set /p branch_name="Enter branch name: "
git checkout main
git pull origin main
git checkout -b %branch_name%
echo.
echo [OK] Branch '%branch_name%' created and checked out
echo.
pause
goto :menu

:commit
echo.
echo Committing changes...
echo.
git status
echo.
echo Commit types:
echo - feat: new feature
echo - fix: bug fix
echo - docs: documentation
echo - refactor: code restructure
echo.
set /p commit_msg="Enter commit message: "
git add .
git commit -m "%commit_msg%"
echo.
echo [OK] Changes committed
echo.
pause
goto :menu

:push
echo.
echo Pushing to GitHub...
echo.
for /f "tokens=*" %%i in ('git branch --show-current') do set current_branch=%%i
echo Current branch: %current_branch%
echo.
git push origin %current_branch%
echo.
echo [OK] Pushed to GitHub
echo.
echo Don't forget to create Pull Request!
echo.
pause
goto :menu

:pull
echo.
echo Pulling latest changes...
echo.
git checkout main
git pull origin main
echo.
echo [OK] Updated to latest
echo.
pause
goto :menu

:pr
echo.
echo Opening GitHub Pull Request page...
echo.
for /f "tokens=*" %%i in ('git remote get-url origin') do set repo_url=%%i
set repo_url=%repo_url:.git=%
start %repo_url%/pulls
echo.
pause
goto :menu

:switch
echo.
echo Available branches:
git branch -a
echo.
set /p target_branch="Enter branch name to switch: "
git checkout %target_branch%
echo.
echo [OK] Switched to '%target_branch%'
echo.
pause
goto :menu

:status
echo.
echo Git Status:
echo.
git status
echo.
echo Current branch:
git branch --show-current
echo.
echo Recent commits:
git log --oneline -5
echo.
pause
goto :menu

:end
echo.
echo Bye! Happy coding! 🚀
echo.
pause
exit
