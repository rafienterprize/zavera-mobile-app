# Quick Push to Branch Raffi
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  PUSH TO BRANCH RAFFI" -ForegroundColor Yellow
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Cek status
Write-Host "[1/4] Checking status..." -ForegroundColor Green
git status --short

Write-Host ""
$continue = Read-Host "Ada perubahan yang mau di-push? (y/n)"

if ($continue -ne "y") {
    Write-Host "Push dibatalkan." -ForegroundColor Yellow
    exit
}

# Add all changes
Write-Host ""
Write-Host "[2/4] Adding changes..." -ForegroundColor Green
git add .

# Commit
Write-Host ""
$message = Read-Host "Commit message (contoh: feat: add login screen)"

if ([string]::IsNullOrWhiteSpace($message)) {
    $message = "feat: update code"
    Write-Host "Using default message: $message" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "[3/4] Committing..." -ForegroundColor Green
git commit -m "$message"

# Push
Write-Host ""
Write-Host "[4/4] Pushing to branch raffi..." -ForegroundColor Green
git push origin raffi

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  ✅ PUSH BERHASIL!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Perubahan sudah di-push ke branch raffi di GitHub!" -ForegroundColor Green
Write-Host ""

Read-Host "Press Enter to exit"
