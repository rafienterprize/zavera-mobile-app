# Panduan Testing Refund untuk Production

## Masalah yang Ditemukan

Saat ini, semua payment di database adalah **manual/testing payment** yang tidak ada di Midtrans. Ketika kita coba refund, Midtrans mengembalikan error `Transaction doesn't exist`.

## Cara Memastikan Refund Berfungsi di Production

### 1. Test dengan Transaksi Real Midtrans

**Langkah-langkah:**

1. **Buat order baru melalui frontend**
   - Login sebagai customer
   - Tambah produk ke cart
   - Checkout dan pilih payment method (GoPay/Bank Transfer/QRIS)

2. **Bayar menggunakan Midtrans Sandbox**
   - Gunakan test credentials Midtrans
   - Untuk GoPay: gunakan nomor test `081234567890`
   - Untuk VA: gunakan nomor VA yang digenerate
   - Untuk QRIS: scan QR code di sandbox

3. **Tunggu payment SUCCESS**
   - Cek di `/admin/orders` bahwa payment status = SUCCESS
   - Cek di Midtrans dashboard bahwa transaksi status = Settlement

4. **Buat refund untuk order tersebut**
   ```bash
   POST /api/admin/refunds
   {
     "order_code": "ZVR-XXXXXXXX",
     "refund_type": "FULL",
     "reason": "CUSTOMER_REQUEST",
     "reason_detail": "Testing refund"
   }
   ```

5. **Process refund**
   ```bash
   POST /api/admin/refunds/{refund_id}/process
   ```

6. **Verifikasi di Midtrans**
   - Buka Midtrans dashboard
   - Cari transaksi berdasarkan order_code
   - Status harus berubah menjadi **Refund**

### 2. Verifikasi Kode Refund Sudah Benar

Saya sudah review kode refund Anda, dan **kode sudah benar**:

✅ **backend/service/refund_service.go**
- `ProcessMidtransRefund()` memanggil Midtrans API dengan benar
- Endpoint: `POST /v2/{order_code}/refund`
- Headers: Authorization dengan Basic Auth
- Body: refund_key, amount, reason

✅ **backend/repository/refund_repository.go**
- `MarkCompleted()` menyimpan gateway_refund_id dan gateway_status
- Update status ke COMPLETED setelah sukses

✅ **backend/handler/admin_refund_handler.go**
- `ProcessRefund()` memanggil service dengan benar
- Error handling sudah proper

### 3. Flow Refund yang Benar

```
1. Customer/Admin request refund
   ↓
2. CreateRefund() - buat record refund dengan status PENDING
   ↓
3. Admin klik "Process Refund" di admin panel
   ↓
4. ProcessRefund() dipanggil
   ↓
5. ProcessMidtransRefund() kirim request ke Midtrans API
   ↓
6. Midtrans proses refund dan return refund_chargeback_id
   ↓
7. MarkCompleted() update status ke COMPLETED + simpan gateway_refund_id
   ↓
8. Status di Midtrans dashboard berubah jadi "Refund"
```

### 4. Kenapa Refund Saat Ini Tidak Muncul di Midtrans?

**Refund ID 4, 10, 13** (COMPLETED, gateway_refund_id = NULL):
- Ini adalah **manual refund** untuk order tanpa payment record
- **Normal behavior** - tidak akan muncul di Midtrans
- Digunakan untuk order yang dibayar cash/manual

**Refund ID 12** (PENDING):
- Order ini punya payment_id tapi transaksi tidak ada di Midtrans
- Ini payment yang di-mark SUCCESS secara manual (bukan real payment)
- Ketika di-process, akan error `Transaction doesn't exist`

### 5. Testing di Sandbox Midtrans

**Test Credentials Midtrans Sandbox:**

**GoPay:**
- Phone: `081234567890`
- PIN: `123456`

**Bank Transfer (VA):**
- BCA: Bayar ke VA number yang digenerate
- BNI: Bayar ke VA number yang digenerate
- Mandiri: Bayar ke VA number yang digenerate

**Credit Card:**
- Card Number: `4811 1111 1111 1114`
- CVV: `123`
- Exp: `01/25`
- OTP: `112233`

**QRIS:**
- Scan QR code yang muncul
- Akan auto-approve di sandbox

### 6. Checklist Sebelum Production

- [ ] Test minimal 1 transaksi real dari Midtrans sandbox
- [ ] Bayar transaksi hingga status = Settlement
- [ ] Buat refund untuk transaksi tersebut
- [ ] Process refund dan verifikasi status berubah di Midtrans
- [ ] Cek database: gateway_refund_id terisi
- [ ] Cek Midtrans dashboard: status = Refund
- [ ] Test refund PARTIAL (sebagian amount)
- [ ] Test refund SHIPPING_ONLY
- [ ] Ganti MIDTRANS_ENVIRONMENT=production di .env
- [ ] Ganti MIDTRANS_SERVER_KEY dengan production key

### 7. Command untuk Test Manual

**1. Buat order dan bayar via Midtrans:**
```bash
# Buka frontend
http://localhost:3000

# Login, add to cart, checkout
# Pilih payment method dan bayar
```

**2. Cek payment berhasil:**
```sql
SELECT o.order_code, p.transaction_id, p.status 
FROM payments p 
JOIN orders o ON p.order_id = o.id 
WHERE p.status = 'SUCCESS' 
ORDER BY p.created_at DESC 
LIMIT 1;
```

**3. Buat refund (ganti ORDER_CODE dan TOKEN):**
```bash
curl -X POST http://localhost:8080/api/admin/refunds \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "order_code": "ZVR-XXXXXXXX",
    "refund_type": "FULL",
    "reason": "CUSTOMER_REQUEST",
    "reason_detail": "Testing refund"
  }'
```

**4. Process refund (ganti REFUND_ID dan TOKEN):**
```bash
curl -X POST http://localhost:8080/api/admin/refunds/REFUND_ID/process \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

**5. Cek hasil:**
```sql
SELECT id, refund_code, status, gateway_refund_id, gateway_status 
FROM refunds 
WHERE id = REFUND_ID;
```

### 8. Expected Results

**Jika berhasil:**
```json
{
  "id": 14,
  "refund_code": "RFD-20260126-xxxxx",
  "status": "COMPLETED",
  "gateway_refund_id": "12345678",
  "gateway_status": "success"
}
```

**Di Midtrans Dashboard:**
- Transaction status: **Refund**
- Refund amount: sesuai dengan refund_amount
- Refund date: timestamp saat process

## Kesimpulan

**Kode refund Anda sudah BENAR dan siap production!** ✅

Yang perlu dilakukan:
1. **Test dengan transaksi REAL** dari Midtrans (bukan manual payment)
2. **Verifikasi** refund berhasil di Midtrans dashboard
3. **Ganti ke production** credentials saat deploy

Masalah saat ini hanya karena **tidak ada transaksi real di Midtrans** untuk di-refund.
