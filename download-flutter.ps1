# Download Flutter SDK
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Downloading Flutter SDK..." -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$flutterUrl = "https://storage.googleapis.com/flutter_infra_release/releases/stable/windows/flutter_windows_3.19.6-stable.zip"
$downloadPath = "$env:USERPROFILE\Downloads\flutter_windows.zip"
$extractPath = "C:\src"

Write-Host "[1/4] Downloading Flutter SDK (~1.5 GB)..." -ForegroundColor Yellow
Write-Host "This may take 10-20 minutes depending on your internet speed..." -ForegroundColor Gray
Write-Host ""

try {
    # Download Flutter
    Invoke-WebRequest -Uri $flutterUrl -OutFile $downloadPath -UseBasicParsing
    Write-Host "[OK] Download complete!" -ForegroundColor Green
    Write-Host ""
    
    Write-Host "[2/4] Creating directory C:\src..." -ForegroundColor Yellow
    if (!(Test-Path $extractPath)) {
        New-Item -ItemType Directory -Path $extractPath -Force | Out-Null
    }
    Write-Host "[OK] Directory ready!" -ForegroundColor Green
    Write-Host ""
    
    Write-Host "[3/4] Extracting Flutter SDK..." -ForegroundColor Yellow
    Write-Host "This may take 5-10 minutes..." -ForegroundColor Gray
    Expand-Archive -Path $downloadPath -DestinationPath $extractPath -Force
    Write-Host "[OK] Extraction complete!" -ForegroundColor Green
    Write-Host ""
    
    Write-Host "[4/4] Adding Flutter to PATH..." -ForegroundColor Yellow
    $flutterBin = "C:\src\flutter\bin"
    $currentPath = [System.Environment]::GetEnvironmentVariable("Path", "User")
    
    if ($currentPath -notlike "*$flutterBin*") {
        [System.Environment]::SetEnvironmentVariable("Path", "$currentPath;$flutterBin", "User")
        Write-Host "[OK] Flutter added to PATH!" -ForegroundColor Green
    } else {
        Write-Host "[OK] Flutter already in PATH!" -ForegroundColor Green
    }
    Write-Host ""
    
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "  Flutter SDK Installed Successfully!" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "Next steps:" -ForegroundColor Cyan
    Write-Host "1. Close and reopen VS Code" -ForegroundColor White
    Write-Host "2. Open new terminal (Ctrl + ~)" -ForegroundColor White
    Write-Host "3. Run: flutter --version" -ForegroundColor White
    Write-Host "4. Run: flutter doctor" -ForegroundColor White
    Write-Host ""
    Write-Host "Flutter location: C:\src\flutter" -ForegroundColor Gray
    Write-Host ""
    
    # Clean up
    Remove-Item $downloadPath -Force
    
} catch {
    Write-Host "[ERROR] Failed to download/install Flutter!" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
    Write-Host ""
    Write-Host "Please download manually from:" -ForegroundColor Yellow
    Write-Host "https://docs.flutter.dev/get-started/install/windows" -ForegroundColor Cyan
}

Write-Host "Press any key to exit..."
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
