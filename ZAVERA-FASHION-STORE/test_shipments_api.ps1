# Test Shipments API
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "TEST SHIPMENTS API" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Step 1: Login
Write-Host "Step 1: Login as admin..." -ForegroundColor Green
$loginBody = @{
    email = "pemberani073@gmail.com"
    password = "admin123"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
    $token = $loginResponse.token
    Write-Host "✅ Login successful!" -ForegroundColor Green
    Write-Host ""
} catch {
    Write-Host "❌ Login failed: $_" -ForegroundColor Red
    exit 1
}

# Step 2: Get dashboard
Write-Host "Step 2: Get fulfillment dashboard..." -ForegroundColor Green
try {
    $dashboard = Invoke-RestMethod -Uri "http://localhost:8080/api/admin/fulfillment/dashboard" -Method GET -Headers @{Authorization="Bearer $token"}
    Write-Host "✅ Dashboard retrieved!" -ForegroundColor Green
    Write-Host "Status Counts:" -ForegroundColor Yellow
    $dashboard.status_counts.PSObject.Properties | ForEach-Object {
        Write-Host "  $($_.Name): $($_.Value)" -ForegroundColor White
    }
    Write-Host ""
} catch {
    Write-Host "❌ Failed to get dashboard: $_" -ForegroundColor Red
}

# Step 3: Get all shipments
Write-Host "Step 3: Get all shipments..." -ForegroundColor Green
try {
    $allShipments = Invoke-RestMethod -Uri "http://localhost:8080/api/admin/shipments" -Method GET -Headers @{Authorization="Bearer $token"}
    Write-Host "✅ All shipments retrieved!" -ForegroundColor Green
    Write-Host "Total: $($allShipments.total)" -ForegroundColor Yellow
    Write-Host "Shipments count: $($allShipments.shipments.Count)" -ForegroundColor Yellow
    Write-Host ""
} catch {
    Write-Host "❌ Failed to get all shipments: $_" -ForegroundColor Red
}

# Step 4: Get PENDING shipments
Write-Host "Step 4: Get PENDING shipments..." -ForegroundColor Green
try {
    $pendingShipments = Invoke-RestMethod -Uri "http://localhost:8080/api/admin/shipments?status=PENDING" -Method GET -Headers @{Authorization="Bearer $token"}
    Write-Host "✅ PENDING shipments retrieved!" -ForegroundColor Green
    Write-Host "Total: $($pendingShipments.total)" -ForegroundColor Yellow
    Write-Host "Shipments count: $($pendingShipments.shipments.Count)" -ForegroundColor Yellow
    
    if ($pendingShipments.shipments.Count -gt 0) {
        Write-Host ""
        Write-Host "First 3 PENDING shipments:" -ForegroundColor Cyan
        $pendingShipments.shipments | Select-Object -First 3 | ForEach-Object {
            Write-Host "  Order: $($_.order_code), Status: $($_.status), Provider: $($_.provider_name)" -ForegroundColor White
        }
    }
    Write-Host ""
} catch {
    Write-Host "❌ Failed to get PENDING shipments: $_" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "TEST COMPLETE" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
