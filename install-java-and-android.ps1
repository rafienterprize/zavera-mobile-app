# Install Java JDK and Android SDK
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Installing Java JDK + Android SDK" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check if Java already installed
$javaVersion = java -version 2>&1
if ($javaVersion -match "version") {
    Write-Host "[OK] Java already installed!" -ForegroundColor Green
    Write-Host $javaVersion[0] -ForegroundColor Gray
    Write-Host ""
} else {
    Write-Host "[1/3] Downloading Java JDK 17..." -ForegroundColor Yellow
    Write-Host "This will take 5-10 minutes (~180 MB)" -ForegroundColor Gray
    Write-Host ""
    
    $jdkUrl = "https://download.oracle.com/java/17/latest/jdk-17_windows-x64_bin.exe"
    $jdkInstaller = "$env:TEMP\jdk-17-installer.exe"
    
    try {
        Invoke-WebRequest -Uri $jdkUrl -OutFile $jdkInstaller -UseBasicParsing
        Write-Host "[OK] Download complete!" -ForegroundColor Green
        Write-Host ""
        
        Write-Host "[2/3] Installing Java JDK..." -ForegroundColor Yellow
        Write-Host "Please follow the installer prompts..." -ForegroundColor Gray
        Start-Process -FilePath $jdkInstaller -Wait
        
        # Set JAVA_HOME
        $javaHome = "C:\Program Files\Java\jdk-17"
        if (Test-Path $javaHome) {
            [System.Environment]::SetEnvironmentVariable("JAVA_HOME", $javaHome, "User")
            $env:JAVA_HOME = $javaHome
            
            $currentPath = [System.Environment]::GetEnvironmentVariable("Path", "User")
            if ($currentPath -notlike "*$javaHome\bin*") {
                [System.Environment]::SetEnvironmentVariable("Path", "$currentPath;$javaHome\bin", "User")
            }
            Write-Host "[OK] Java installed!" -ForegroundColor Green
        }
        
        Remove-Item $jdkInstaller -Force
    } catch {
        Write-Host "[ERROR] Failed to install Java!" -ForegroundColor Red
        Write-Host "Please install manually from: https://www.oracle.com/java/technologies/downloads/" -ForegroundColor Yellow
        exit 1
    }
}

Write-Host ""
Write-Host "[3/3] Installing Android SDK components..." -ForegroundColor Yellow
Write-Host "This will take 5-10 minutes..." -ForegroundColor Gray
Write-Host ""

$env:ANDROID_HOME = "C:\Android"
$sdkManager = "C:\Android\cmdline-tools\latest\bin\sdkmanager.bat"

if (Test-Path $sdkManager) {
    & $sdkManager "platform-tools" "platforms;android-34" "build-tools;34.0.0" --sdk_root="C:\Android"
    
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "  Installation Complete!" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "Next steps:" -ForegroundColor Cyan
    Write-Host "1. Close and reopen VS Code" -ForegroundColor White
    Write-Host "2. Run: flutter doctor --android-licenses" -ForegroundColor White
    Write-Host "3. Run: flutter build apk --debug" -ForegroundColor White
    Write-Host ""
} else {
    Write-Host "[ERROR] Android SDK not found!" -ForegroundColor Red
    Write-Host "Please run install-android-sdk-only.ps1 first" -ForegroundColor Yellow
}

Write-Host "Press any key to exit..."
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
