# Debug: Webhook SSE Notification Not Working

## Status
- ✅ Ngrok running
- ✅ Webhook URL configured in Midtrans
- ✅ Webhook called (3 POST requests in ngrok logs)
- ❌ SSE notification NOT sent to admin

## Debugging Steps

### 1. Check Backend Logs
Restart backend dan cek logs saat bayar di simulator:

```bash
# Windows
cd backend
.\zavera.exe

# Look for these logs:
🔔 Core Webhook received from IP: ...
🔔 Webhook Headers: ...
📦 Webhook payload: order_id=..., status=..., tx_id=..., fraud=...
📦 Full notification: ...
🔐 Signature verification:
   Order ID: ...
   Status Code: ...
   Gross Amount: ...
   Calculated: ...
   Received: ...
   Match: true/false
```

### 2. Kemungkinan Masalah

#### A. Signature Validation Gagal
Jika log menunjukkan `Match: false`:

**Penyebab**: Server key tidak match dengan Midtrans

**Solusi**:
1. Cek `.env` file:
   ```
   MIDTRANS_SERVER_KEY=SB-Mid-server-xxxxx
   ```

2. Pastikan menggunakan **Sandbox Server Key** (bukan Production)

3. Copy dari Midtrans Dashboard:
   - Settings → Access Keys
   - Copy **Server Key** (bukan Client Key!)

#### B. Order ID Format Salah
Jika log menunjukkan error "order not found":

**Penyebab**: Order ID format tidak sesuai

**Cek**:
- Order ID di webhook: `ZVR-20260122-ABC123-1737558000`
- Order code di database: `ZVR-20260122-ABC123`

**Solusi**: Sudah ada `extractOrderCode()` function

#### C. Transaction Status Tidak Dikenali
Jika log menunjukkan "Unhandled transaction status":

**Cek**: Status dari Midtrans simulator
- Expected: `settlement` atau `capture`
- Actual: ???

#### D. SSE Broker Tidak Aktif
Jika webhook berhasil tapi notifikasi tidak muncul:

**Cek**:
```bash
# Backend startup logs
🚀 SSE Broker starting...
✅ SSE Broker started successfully
```

**Cek Admin Dashboard Console**:
```javascript
// Should see:
✅ SSE Connected
```

### 3. Manual Test Webhook

Jika webhook tidak dipanggil otomatis, test manual:

```bash
# Test webhook endpoint
curl -X POST http://localhost:8080/api/webhook/midtrans/core \
  -H "Content-Type: application/json" \
  -d '{
    "transaction_time": "2026-01-22 21:00:00",
    "transaction_status": "settlement",
    "transaction_id": "test-123",
    "status_message": "midtrans payment notification",
    "status_code": "200",
    "signature_key": "test-signature",
    "payment_type": "bank_transfer",
    "order_id": "ZVR-20260122-ABC123-1737558000",
    "merchant_id": "G123456789",
    "gross_amount": "717000.00",
    "fraud_status": "accept",
    "currency": "IDR"
  }'
```

**Expected Response**:
```json
{
  "status": "ok"
}
```

**Expected Backend Logs**:
```
🔔 Core Webhook received from IP: ::1
📦 Webhook payload: order_id=ZVR-20260122-ABC123-1737558000, status=settlement
🔐 Signature verification: ...
💰 Processing settlement for order ZVR-20260122-ABC123
📢 Sending payment notification for order ZVR-20260122-ABC123
✅ Webhook processed successfully
```

### 4. Bypass Signature Validation (TESTING ONLY!)

Jika signature terus gagal, temporary bypass untuk testing:

**File**: `backend/service/core_payment_service.go`

```go
func (s *corePaymentService) ProcessCoreWebhook(notification CoreWebhookNotification) error {
    log.Printf("🔔 ProcessCoreWebhook: order_id=%s, status=%s", notification.OrderID, notification.TransactionStatus)

    // TEMPORARY: Skip signature validation for testing
    // TODO: Remove this in production!
    /*
    if !s.verifySignature(notification) {
        log.Printf("❌ Invalid signature for order: %s", notification.OrderID)
        return ErrInvalidSignature
    }
    */
    log.Printf("⚠️ SIGNATURE VALIDATION BYPASSED - TESTING ONLY!")
    
    // ... rest of code
}
```

**PENTING**: Jangan lupa enable kembali signature validation setelah testing!

### 5. Check Midtrans Configuration

#### Sandbox vs Production
Pastikan menggunakan environment yang benar:

```env
# .env file
MIDTRANS_SERVER_KEY=SB-Mid-server-xxxxx  # SB = Sandbox
MIDTRANS_CLIENT_KEY=SB-Mid-client-xxxxx
MIDTRANS_IS_PRODUCTION=false
```

#### Notification URL
Di Midtrans Dashboard:
- Settings → Configuration → Notification URL
- Set: `https://your-ngrok-url.ngrok.io/api/webhook/midtrans/core`
- **PENTING**: Harus HTTPS (ngrok provides this)

#### Payment Notification
Di Midtrans Dashboard:
- Settings → Configuration
- Enable: "Payment Notification"
- Enable: "Recurring Notification"

### 6. Network Issues

#### Check Ngrok
```bash
# Ngrok web interface
http://127.0.0.1:4040

# Should show:
- Status: online
- Forwarding: https://xxx.ngrok.io -> http://localhost:8080
- Requests: Should see POST /api/webhook/midtrans/core
```

#### Check Firewall
Pastikan port 8080 tidak diblock:
```bash
# Windows
netstat -ano | findstr :8080

# Should show:
TCP    0.0.0.0:8080    0.0.0.0:0    LISTENING    <PID>
```

### 7. Database Check

Setelah bayar di simulator, cek database:

```sql
-- Check order status
SELECT id, order_code, status, paid_at, created_at 
FROM orders 
WHERE order_code LIKE 'ZVR-20260122%' 
ORDER BY created_at DESC 
LIMIT 5;

-- Check payment status
SELECT op.id, op.order_id, o.order_code, op.payment_status, op.paid_at
FROM order_payments op
JOIN orders o ON op.order_id = o.id
WHERE o.order_code LIKE 'ZVR-20260122%'
ORDER BY op.created_at DESC
LIMIT 5;

-- Check webhook sync logs
SELECT * FROM core_payment_sync_logs 
WHERE order_code LIKE 'ZVR-20260122%'
ORDER BY created_at DESC 
LIMIT 5;
```

**Expected**:
- Order status: `PAID`
- Payment status: `PAID`
- Sync log: `sync_type = 'webhook'`, `sync_status = 'SYNCED'`

### 8. SSE Connection Check

#### Admin Dashboard Console
```javascript
// Open browser console (F12)
// Should see:
EventSource {url: 'http://localhost:8080/api/admin/events', ...}
readyState: 1  // 1 = OPEN, 0 = CONNECTING, 2 = CLOSED
```

#### Backend SSE Logs
```bash
# When admin opens dashboard
[GIN] GET /api/admin/events
📡 New SSE client connected: <client-id>
📊 Active SSE clients: 1
```

#### Test SSE Manually
```bash
# Terminal 1: Connect to SSE
curl -N -H "Authorization: Bearer <admin-token>" \
  http://localhost:8080/api/admin/events

# Terminal 2: Trigger notification
# (bayar di simulator atau manual webhook test)

# Terminal 1 should show:
data: {"type":"payment_received","message":"...","severity":"info"}
```

## Quick Checklist

- [ ] Backend running dengan logs visible
- [ ] Ngrok running dan forwarding ke localhost:8080
- [ ] Webhook URL di Midtrans = ngrok HTTPS URL
- [ ] Server key di .env = Sandbox server key dari Midtrans
- [ ] Admin dashboard open dengan SSE connected
- [ ] Bayar di simulator
- [ ] Check backend logs untuk webhook received
- [ ] Check signature validation result
- [ ] Check SSE notification sent
- [ ] Check admin dashboard untuk notification

## Expected Flow

```
1. Client bayar di simulator
   ↓
2. Midtrans → POST webhook ke ngrok URL
   ↓
3. Ngrok → Forward ke localhost:8080/api/webhook/midtrans/core
   ↓
4. Backend: Receive webhook
   ↓
5. Backend: Verify signature ✅
   ↓
6. Backend: Update order & payment to PAID
   ↓
7. Backend: Call NotifyPaymentReceived()
   ↓
8. Backend: SSE Broker → Broadcast to all admin clients
   ↓
9. Admin Dashboard: Receive SSE event
   ↓
10. Admin Dashboard: Show toast notification ✅
```

## Next Steps

1. **Restart backend** dengan logging enabled
2. **Open admin dashboard** (pastikan SSE connected)
3. **Bayar di simulator**
4. **Copy paste backend logs** ke sini untuk analisis
5. Kita akan debug berdasarkan logs

## Common Issues & Solutions

| Issue | Symptom | Solution |
|-------|---------|----------|
| Signature mismatch | `Match: false` | Check server key di .env |
| Order not found | `order not found` | Check order code format |
| Webhook not called | No logs | Check ngrok & Midtrans config |
| SSE not connected | No notification | Check admin auth & SSE endpoint |
| Status not handled | `Unhandled status` | Check transaction_status value |

