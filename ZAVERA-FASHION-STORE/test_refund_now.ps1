# ZAVERA REFUND SYSTEM TEST
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "ZAVERA REFUND SYSTEM TEST" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Test Order: ZVR-20260127-B8B3ACCD" -ForegroundColor Yellow
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
    Write-Host "Token: $($token.Substring(0, 20))..." -ForegroundColor Gray
    Write-Host ""
} catch {
    Write-Host "❌ Login failed: $_" -ForegroundColor Red
    exit 1
}

# Step 2: Create FULL refund
Write-Host "Step 2: Create FULL refund..." -ForegroundColor Green
$refundBody = @{
    order_code = "ZVR-20260127-B8B3ACCD"
    refund_type = "FULL"
    reason = "CUSTOMER_REQUEST"
    reason_detail = "Test refund system"
    idempotency_key = "test-refund-$(Get-Date -Format 'yyyyMMddHHmmss')"
} | ConvertTo-Json

try {
    $refundResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/admin/refunds" -Method POST -Body $refundBody -ContentType "application/json" -Headers @{Authorization="Bearer $token"}
    Write-Host "✅ Refund created!" -ForegroundColor Green
    Write-Host "Refund Code: $($refundResponse.refund_code)" -ForegroundColor Yellow
    Write-Host "Status: $($refundResponse.status)" -ForegroundColor Yellow
    Write-Host "Amount: $($refundResponse.refund_amount)" -ForegroundColor Yellow
    Write-Host ""
    
    $refundId = $refundResponse.id
} catch {
    Write-Host "❌ Refund creation failed: $_" -ForegroundColor Red
    Write-Host "Response: $($_.Exception.Response)" -ForegroundColor Red
    exit 1
}

# Step 3: Process refund
Write-Host "Step 3: Process refund..." -ForegroundColor Green
try {
    $processResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/admin/refunds/$refundId/process" -Method POST -ContentType "application/json" -Headers @{Authorization="Bearer $token"}
    Write-Host "✅ Refund processed!" -ForegroundColor Green
    Write-Host "Success: $($processResponse.success)" -ForegroundColor Yellow
    Write-Host "Message: $($processResponse.message)" -ForegroundColor Yellow
    Write-Host ""
} catch {
    $errorMsg = $_.Exception.Message
    Write-Host "⚠️ Refund processing error: $errorMsg" -ForegroundColor Yellow
    
    # Check if it's error 418 (manual processing required)
    if ($errorMsg -like "*MANUAL_PROCESSING_REQUIRED*" -or $errorMsg -like "*manual bank transfer*") {
        Write-Host "✅ This is expected! Error 418 detected - manual processing required" -ForegroundColor Green
        Write-Host "Refund should be in PENDING status with 'Mark as Completed' option" -ForegroundColor Cyan
        Write-Host ""
    } else {
        Write-Host "Response: $($_.Exception.Response)" -ForegroundColor Red
    }
}

# Step 4: Check refunds for order
Write-Host "Step 4: Check refunds for order..." -ForegroundColor Green
try {
    $refundsResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/admin/orders/ZVR-20260127-B8B3ACCD/refunds" -Method GET -Headers @{Authorization="Bearer $token"}
    Write-Host "✅ Refunds retrieved!" -ForegroundColor Green
    Write-Host "Total refunds: $($refundsResponse.Count)" -ForegroundColor Yellow
    
    foreach ($refund in $refundsResponse) {
        Write-Host ""
        Write-Host "  Refund Code: $($refund.refund_code)" -ForegroundColor Cyan
        Write-Host "  Status: $($refund.status)" -ForegroundColor Cyan
        Write-Host "  Amount: $($refund.refund_amount)" -ForegroundColor Cyan
        Write-Host "  Type: $($refund.refund_type)" -ForegroundColor Cyan
        if ($refund.gateway_refund_id) {
            Write-Host "  Gateway ID: $($refund.gateway_refund_id)" -ForegroundColor Cyan
        }
    }
    Write-Host ""
} catch {
    Write-Host "❌ Failed to retrieve refunds: $_" -ForegroundColor Red
}

# Step 5: Check order status
Write-Host "Step 5: Check order status..." -ForegroundColor Green
try {
    $orderResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/admin/orders/ZVR-20260127-B8B3ACCD" -Method GET -Headers @{Authorization="Bearer $token"}
    Write-Host "✅ Order retrieved!" -ForegroundColor Green
    Write-Host "Order Status: $($orderResponse.status)" -ForegroundColor Yellow
    Write-Host "Refund Status: $($orderResponse.refund_status)" -ForegroundColor Yellow
    Write-Host "Refund Amount: $($orderResponse.refund_amount)" -ForegroundColor Yellow
    Write-Host ""
} catch {
    Write-Host "❌ Failed to retrieve order: $_" -ForegroundColor Red
}

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "TEST COMPLETE" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "1. Open admin panel: http://localhost:3000/admin/orders/ZVR-20260127-B8B3ACCD" -ForegroundColor White
Write-Host "2. Check if refund shows in Refund History section" -ForegroundColor White
Write-Host "3. If status is PENDING, click 'Mark as Completed' button" -ForegroundColor White
Write-Host "4. Enter confirmation note and complete the refund" -ForegroundColor White
Write-Host ""
