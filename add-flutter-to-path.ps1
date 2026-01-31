# Add Flutter to PATH
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Adding Flutter to PATH..." -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Cari lokasi Flutter
Write-Host "Searching for Flutter SDK..." -ForegroundColor Yellow

$possiblePaths = @(
    "C:\flutter\bin",
    "C:\src\flutter\bin",
    "$env:USERPROFILE\flutter\bin",
    "$env:USERPROFILE\Downloads\flutter\bin",
    "D:\flutter\bin"
)

$flutterPath = $null

foreach ($path in $possiblePaths) {
    if (Test-Path "$path\flutter.bat") {
        $flutterPath = $path
        Write-Host "[FOUND] Flutter SDK at: $flutterPath" -ForegroundColor Green
        break
    }
}

if ($null -eq $flutterPath) {
    Write-Host "[ERROR] Flutter SDK not found!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please enter Flutter SDK path manually:" -ForegroundColor Yellow
    Write-Host "Example: C:\flutter\bin" -ForegroundColor Gray
    $flutterPath = Read-Host "Flutter bin path"
    
    if (!(Test-Path "$flutterPath\flutter.bat")) {
        Write-Host "[ERROR] Invalid path! flutter.bat not found." -ForegroundColor Red
        exit 1
    }
}

Write-Host ""
Write-Host "Adding to PATH..." -ForegroundColor Yellow

# Get current PATH
$currentPath = [System.Environment]::GetEnvironmentVariable("Path", "User")

# Check if already in PATH
if ($currentPath -like "*$flutterPath*") {
    Write-Host "[OK] Flutter already in PATH!" -ForegroundColor Green
} else {
    # Add to PATH
    $newPath = "$currentPath;$flutterPath"
    [System.Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    Write-Host "[OK] Flutter added to PATH!" -ForegroundColor Green
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "  Setup Complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host "1. Close and reopen VS Code (IMPORTANT!)" -ForegroundColor White
Write-Host "2. Open new terminal" -ForegroundColor White
Write-Host "3. Run: flutter --version" -ForegroundColor White
Write-Host "4. Run: flutter doctor" -ForegroundColor White
Write-Host ""
Write-Host "Flutter location: $flutterPath" -ForegroundColor Gray
Write-Host ""

Write-Host "Press any key to exit..."
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
