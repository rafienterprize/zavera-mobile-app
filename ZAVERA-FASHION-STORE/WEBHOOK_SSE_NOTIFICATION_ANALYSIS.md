# Analisis: Webhook SSE Notification Issue

## 🔴 MASALAH KRITIS

### Situasi:
1. Client bayar di Midtrans simulator → Status jadi PAID ✅
2. **TAPI** notifikasi SSE ke admin **HANYA** muncul jika client klik "Cek Status Pembayaran" ❌
3. Jika client langsung keluar tanpa klik → **Admin tidak dapat notifikasi real-time** ❌

### Dampak:
- **Client**: Tidak ada masalah (status sudah PAID di database)
- **Admin**: **MASALAH BESAR** - Tidak tahu ada pembayaran baru masuk secara real-time!

---

## 🔍 ROOT CAUSE ANALYSIS

### Alur Pembayaran Saat Ini:

```
┌─────────────────────────────────────────────────────────────┐
│ SCENARIO 1: Client Klik "Cek Status Pembayaran"            │
└─────────────────────────────────────────────────────────────┘

1. Client bayar di simulator Midtrans
2. Midtrans webhook → Backend (update DB ke PAID)
3. Client klik "Cek Status Pembayaran"
4. Backend CheckPaymentStatus → Kirim SSE notification ✅
5. Admin dapat notifikasi ✅

┌─────────────────────────────────────────────────────────────┐
│ SCENARIO 2: Client Langsung Keluar (MASALAH!)              │
└─────────────────────────────────────────────────────────────┘

1. Client bayar di simulator Midtrans
2. Midtrans webhook → Backend (update DB ke PAID)
3. Client langsung keluar (tidak klik cek status)
4. ❌ TIDAK ADA SSE notification ke admin
5. Admin tidak tahu ada pembayaran baru ❌
```

### Penyebab:
**Webhook handler SUDAH mengirim SSE notification**, tapi kemungkinan:
1. **Midtrans simulator tidak memanggil webhook** secara otomatis
2. **Webhook URL tidak terkonfigurasi** dengan benar di Midtrans
3. **Webhook dipanggil tapi gagal** (network issue, signature validation, dll)

---

## ✅ SOLUSI YANG SUDAH DIIMPLEMENTASIKAN

### 1. Webhook Handler Sudah Benar
File: `backend/service/core_payment_service.go`

```go
func (s *corePaymentService) handleWebhookSettlement(...) error {
    // Update payment & order to PAID
    // ...
    
    // ✅ SUDAH ADA: Send notification to admin dashboard
    log.Printf("📢 Sending payment notification for order %s", order.OrderCode)
    NotifyPaymentReceived(order.OrderCode, string(payment.PaymentMethod), order.TotalAmount)
    log.Printf("📢 Payment notification sent")
    
    return nil
}
```

### 2. Email Service Ditambahkan ke Webhook
**BARU DITAMBAHKAN**: Email dikirim setelah webhook berhasil

```go
func (s *corePaymentService) ProcessCoreWebhook(...) error {
    // ... process webhook ...
    
    // Commit transaction
    tx.Commit()
    
    // ✅ BARU: Send email for successful payment (async, after commit)
    if notification.TransactionStatus == "settlement" || notification.TransactionStatus == "capture" {
        if s.emailService != nil {
            go func() {
                freshOrder, err := s.orderRepo.FindByOrderCode(orderCode)
                if err == nil {
                    paymentMethodStr := string(payment.PaymentMethod)
                    s.emailService.SendPaymentSuccess(freshOrder, paymentMethodStr)
                }
            }()
        }
    }
    
    return nil
}
```

---

## 🧪 CARA TESTING

### Test 1: Verifikasi Webhook Dipanggil
1. Bayar di Midtrans simulator
2. Cek backend logs untuk:
   ```
   🔔 Core Webhook received from IP: ...
   📦 Webhook payload: order_id=..., status=settlement
   💰 Processing settlement for order ...
   📢 Sending payment notification for order ...
   ✅ Webhook processed successfully
   ```

### Test 2: Verifikasi SSE Notification
1. Buka admin dashboard (pastikan SSE connection aktif)
2. Bayar di Midtrans simulator
3. **Jangan klik "Cek Status"** di client
4. Cek apakah notifikasi muncul di admin dashboard

### Test 3: Verifikasi Email Dikirim
1. Bayar di Midtrans simulator
2. Cek backend logs untuk:
   ```
   ✅ Payment success email sent for order ...
   ```
3. Cek inbox customer untuk email konfirmasi

---

## 🔧 KONFIGURASI WEBHOOK MIDTRANS

### Untuk Testing (Simulator):
Midtrans simulator **TIDAK** memanggil webhook secara otomatis. Anda harus:

1. **Manual Webhook Test**:
   - Buka Midtrans Dashboard
   - Settings → Configuration → Notification URL
   - Set: `https://your-domain.com/api/webhook/midtrans/core`
   - Test webhook manually

2. **Ngrok untuk Local Testing**:
   ```bash
   # Install ngrok
   ngrok http 8080
   
   # Copy HTTPS URL (e.g., https://abc123.ngrok.io)
   # Set di Midtrans: https://abc123.ngrok.io/api/webhook/midtrans/core
   ```

### Untuk Production:
1. Set Notification URL di Midtrans Dashboard:
   ```
   https://api.zavera.com/api/webhook/midtrans/core
   ```

2. Pastikan endpoint accessible dari internet

3. Verifikasi signature validation aktif

---

## 📊 MONITORING & DEBUGGING

### Check Webhook Logs:
```bash
# Backend logs
tail -f backend.log | grep "Webhook"

# Expected output:
🔔 Core Webhook received from IP: 103.127.132.64
📦 Webhook payload: order_id=ZVR-20260122-ABC123-1737558000, status=settlement
💰 Processing settlement for order ZVR-20260122-ABC123
📢 Sending payment notification for order ZVR-20260122-ABC123
✅ Webhook processed successfully for order: ZVR-20260122-ABC123
```

### Check SSE Connection:
```bash
# Admin dashboard console
# Should see:
✅ SSE Connected
📢 New notification: Payment received for order ZVR-20260122-ABC123
```

### Check Database:
```sql
-- Check payment status
SELECT order_code, status, paid_at 
FROM orders 
WHERE order_code = 'ZVR-20260122-ABC123';

-- Check webhook sync logs
SELECT * FROM core_payment_sync_logs 
WHERE order_code = 'ZVR-20260122-ABC123' 
ORDER BY created_at DESC;
```

---

## 🎯 KESIMPULAN

### Masalah Sebenarnya:
**Bukan bug di code**, tapi **Midtrans simulator tidak memanggil webhook secara otomatis**.

### Solusi:
1. ✅ **Code sudah benar** - Webhook handler sudah mengirim SSE notification
2. ✅ **Email service ditambahkan** - Customer dapat konfirmasi email
3. ⚠️ **Perlu konfigurasi webhook** - Set notification URL di Midtrans
4. ⚠️ **Untuk testing lokal** - Gunakan ngrok untuk expose localhost

### Production Checklist:
- [ ] Set webhook URL di Midtrans Dashboard
- [ ] Verifikasi webhook accessible dari internet
- [ ] Test webhook dengan real payment
- [ ] Monitor webhook logs untuk errors
- [ ] Setup alerting untuk webhook failures

---

## 🚀 NEXT STEPS

1. **Setup Ngrok** untuk local testing:
   ```bash
   ngrok http 8080
   ```

2. **Configure Midtrans**:
   - Login ke Midtrans Dashboard
   - Settings → Configuration
   - Set Notification URL: `https://your-ngrok-url.ngrok.io/api/webhook/midtrans/core`

3. **Test End-to-End**:
   - Bayar di simulator
   - Jangan klik "Cek Status"
   - Verifikasi admin dapat notifikasi
   - Verifikasi customer dapat email

4. **Production Deployment**:
   - Deploy backend dengan public URL
   - Update webhook URL di Midtrans
   - Monitor webhook logs
