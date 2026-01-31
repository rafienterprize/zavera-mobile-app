# Install Android SDK Command Line Tools Only (Tanpa Android Studio)
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Installing Android SDK (Lightweight)" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$sdkUrl = "https://dl.google.com/android/repository/commandlinetools-win-11076708_latest.zip"
$downloadPath = "$env:USERPROFILE\Downloads\android-cmdline-tools.zip"
$androidHome = "C:\Android"
$cmdlineToolsPath = "$androidHome\cmdline-tools\latest"

Write-Host "[1/5] Downloading Android Command Line Tools (~150 MB)..." -ForegroundColor Yellow
Write-Host "This is much lighter than Android Studio!" -ForegroundColor Gray
Write-Host ""

try {
    # Download
    Invoke-WebRequest -Uri $sdkUrl -OutFile $downloadPath -UseBasicParsing
    Write-Host "[OK] Download complete!" -ForegroundColor Green
    Write-Host ""
    
    # Create directories
    Write-Host "[2/5] Creating Android SDK directory..." -ForegroundColor Yellow
    if (!(Test-Path $androidHome)) {
        New-Item -ItemType Directory -Path $androidHome -Force | Out-Null
    }
    if (!(Test-Path $cmdlineToolsPath)) {
        New-Item -ItemType Directory -Path $cmdlineToolsPath -Force | Out-Null
    }
    Write-Host "[OK] Directory ready!" -ForegroundColor Green
    Write-Host ""
    
    # Extract
    Write-Host "[3/5] Extracting tools..." -ForegroundColor Yellow
    Expand-Archive -Path $downloadPath -DestinationPath "$androidHome\temp" -Force
    
    # Move to correct location
    Get-ChildItem "$androidHome\temp\cmdline-tools" | Move-Item -Destination $cmdlineToolsPath -Force
    Remove-Item "$androidHome\temp" -Recurse -Force
    Write-Host "[OK] Extraction complete!" -ForegroundColor Green
    Write-Host ""
    
    # Set environment variables
    Write-Host "[4/5] Setting environment variables..." -ForegroundColor Yellow
    [System.Environment]::SetEnvironmentVariable("ANDROID_HOME", $androidHome, "User")
    
    $currentPath = [System.Environment]::GetEnvironmentVariable("Path", "User")
    $pathsToAdd = @(
        "$cmdlineToolsPath\bin",
        "$androidHome\platform-tools",
        "$androidHome\emulator"
    )
    
    foreach ($path in $pathsToAdd) {
        if ($currentPath -notlike "*$path*") {
            $currentPath = "$currentPath;$path"
        }
    }
    [System.Environment]::SetEnvironmentVariable("Path", $currentPath, "User")
    Write-Host "[OK] Environment variables set!" -ForegroundColor Green
    Write-Host ""
    
    # Install SDK components
    Write-Host "[5/5] Installing Android SDK components..." -ForegroundColor Yellow
    Write-Host "This will take 5-10 minutes..." -ForegroundColor Gray
    Write-Host ""
    
    $env:ANDROID_HOME = $androidHome
    $env:Path = "$env:Path;$cmdlineToolsPath\bin"
    
    & "$cmdlineToolsPath\bin\sdkmanager.bat" "platform-tools" "platforms;android-34" "build-tools;34.0.0"
    
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "  Android SDK Installed!" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "Next steps:" -ForegroundColor Cyan
    Write-Host "1. Close and reopen VS Code" -ForegroundColor White
    Write-Host "2. Run: flutter doctor --android-licenses" -ForegroundColor White
    Write-Host "   (Type 'y' for all)" -ForegroundColor Gray
    Write-Host "3. Connect your phone via USB" -ForegroundColor White
    Write-Host "4. Run: flutter devices" -ForegroundColor White
    Write-Host "5. Run: flutter run" -ForegroundColor White
    Write-Host ""
    Write-Host "Android SDK location: $androidHome" -ForegroundColor Gray
    Write-Host ""
    
    # Clean up
    Remove-Item $downloadPath -Force
    
} catch {
    Write-Host "[ERROR] Installation failed!" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
    Write-Host ""
    Write-Host "Alternative: Install Android Studio from:" -ForegroundColor Yellow
    Write-Host "https://developer.android.com/studio" -ForegroundColor Cyan
}

Write-Host "Press any key to exit..."
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
