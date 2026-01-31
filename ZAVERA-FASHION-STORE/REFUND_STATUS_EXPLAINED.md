# Penjelasan Status Refund di Midtrans

## Pertanyaan: "Kenapa status di Midtrans tidak berubah jadi Refunded?"

### Jawaban Singkat:
**Karena transaksi yang di-refund tidak ada di Midtrans (manual payment).** Untuk refund muncul di Midtrans, transaksi harus **benar-benar dibayar melalui Midtrans**, bukan di-mark SUCCESS secara manual.

---

## Penjelasan Detail

### 1. Jenis Refund di Sistem Anda

Ada **2 jenis refund** di sistem:

#### A. Manual Refund (payment_id = NULL)
```
Order → Tidak ada payment record → Refund langsung COMPLETED
```
- Untuk order yang dibayar cash/manual (bukan via Midtrans)
- Status langsung `COMPLETED` tanpa proses ke gateway
- `gateway_refund_id` = NULL
- **TIDAK akan muncul di Midtrans** ✅ (ini normal)

**Contoh di database Anda:**
- Refund ID 4, 10, 13 → Manual refund
- Status: COMPLETED
- gateway_refund_id: NULL

#### B. Gateway Refund (payment_id ada)
```
Order → Payment via Midtrans → Refund perlu diproses ke Midtrans
```
- Untuk order yang dibayar via Midtrans (GoPay/VA/QRIS/CC)
- Status awal `PENDING`
- Perlu di-**process** agar dikirim ke Midtrans
- Setelah sukses: `gateway_refund_id` terisi, status `COMPLETED`
- **AKAN muncul di Midtrans** ✅

**Contoh di database Anda:**
- Refund ID 12 → Gateway refund
- Status: PENDING (belum diproses)
- payment_id: 1
- Tapi transaksi tidak ada di Midtrans (manual payment)

---

### 2. Kenapa Refund ID 12 Gagal?

```bash
Error: "Transaction doesn't exist"
```

**Penyebab:**
- Payment ID 1 di-mark SUCCESS secara **manual** (bukan real payment dari Midtrans)
- Ketika kode refund coba kirim ke Midtrans, transaksi tidak ditemukan
- Midtrans return error 404

**Solusi:**
Test dengan **transaksi REAL** dari Midtrans:
1. Buat order baru via frontend
2. Bayar menggunakan Midtrans sandbox
3. Tunggu payment SUCCESS (real dari Midtrans)
4. Buat refund
5. Process refund
6. Cek Midtrans dashboard → status berubah jadi "Refund" ✅

---

### 3. Flow Refund yang Benar

```
┌─────────────────────────────────────────────────────────────┐
│ CUSTOMER BAYAR VIA MIDTRANS                                 │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│ Payment record dibuat dengan transaction_id dari Midtrans   │
│ Status: SUCCESS                                             │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│ Admin/Customer request refund                               │
│ CreateRefund() → Status: PENDING                            │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│ Admin klik "Process Refund"                                 │
│ ProcessRefund() → Kirim ke Midtrans API                     │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│ Midtrans proses refund                                      │
│ Return: refund_chargeback_id                                │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│ MarkCompleted() update database                             │
│ - Status: COMPLETED                                         │
│ - gateway_refund_id: "12345678"                             │
│ - gateway_status: "success"                                 │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│ ✅ Status di Midtrans dashboard berubah jadi "Refund"       │
└─────────────────────────────────────────────────────────────┘
```

---

### 4. Cara Test Refund Sebelum Production

**Step 1: Buat Transaksi Real**
```bash
# Buka frontend
http://localhost:3000

# Login → Add to cart → Checkout
# Pilih payment: GoPay/Bank Transfer/QRIS
```

**Step 2: Bayar di Midtrans Sandbox**
```
GoPay Test:
- Phone: 081234567890
- PIN: 123456

Bank Transfer:
- Gunakan VA number yang digenerate
- Auto-approve di sandbox

QRIS:
- Scan QR code
- Auto-approve di sandbox
```

**Step 3: Verifikasi Payment SUCCESS**
```sql
SELECT o.order_code, p.transaction_id, p.status 
FROM payments p 
JOIN orders o ON p.order_id = o.id 
WHERE p.status = 'SUCCESS' 
ORDER BY p.created_at DESC 
LIMIT 1;
```

**Step 4: Buat Refund**
```bash
# Via admin panel atau API
POST /api/admin/refunds
{
  "order_code": "ZVR-XXXXXXXX",
  "refund_type": "FULL",
  "reason": "CUSTOMER_REQUEST"
}
```

**Step 5: Process Refund**
```bash
# Klik "Process" di admin panel atau
POST /api/admin/refunds/{id}/process
```

**Step 6: Verifikasi di Midtrans**
```
1. Buka https://dashboard.sandbox.midtrans.com
2. Login dengan akun Midtrans Anda
3. Menu: Transactions
4. Cari order_code
5. Status harus: "Refund" ✅
```

---

### 5. Checklist Production Ready

- [ ] **Kode refund sudah benar** ✅ (sudah saya review)
- [ ] **Test dengan transaksi real** dari Midtrans sandbox
- [ ] **Verifikasi refund muncul** di Midtrans dashboard
- [ ] **Test semua jenis refund:**
  - [ ] FULL refund
  - [ ] PARTIAL refund
  - [ ] SHIPPING_ONLY refund
  - [ ] ITEM_ONLY refund
- [ ] **Ganti ke production:**
  - [ ] MIDTRANS_ENVIRONMENT=production
  - [ ] MIDTRANS_SERVER_KEY=production_key
  - [ ] Test 1 transaksi kecil di production

---

## Kesimpulan

### ✅ Kode Refund Anda SUDAH BENAR!

Yang perlu dilakukan:
1. **Test dengan transaksi REAL** (bukan manual payment)
2. **Verifikasi** di Midtrans dashboard
3. **Deploy** dengan production credentials

### ❌ Kenapa Sekarang Tidak Muncul di Midtrans?

Karena **tidak ada transaksi real** di database. Semua payment di-mark SUCCESS secara manual, bukan dari real Midtrans payment.

### 🎯 Next Steps

1. Jalankan `test_refund_complete.bat` untuk cek status
2. Baca `REFUND_TESTING_PRODUCTION_GUIDE.md` untuk panduan lengkap
3. Buat 1 order baru dan bayar via Midtrans sandbox
4. Test refund untuk order tersebut
5. Verifikasi muncul di Midtrans dashboard

**Setelah test berhasil, sistem refund siap production!** 🚀
