# 🧪 Test Biteship Resi - Step by Step

## ⚠️ PENTING: Order Lama Tidak Punya Draft Order!

Order yang dibuat **SEBELUM** update code tidak punya draft order Biteship.
Untuk test resi dari Biteship, harus **CREATE ORDER BARU**.

---

## 📋 Test Checklist

### Step 1: Restart Backend dengan Code Baru
```bash
# Stop backend lama (Ctrl+C)

# Start backend baru
cd backend
.\zavera_COMPLETE.exe
```

**Cek Log:**
```
✅ Server running on :8080
✅ Database connected
```

---

### Step 2: Create Order Baru (Customer)

1. **Buka frontend:** http://localhost:3000
2. **Login sebagai customer** (bukan admin!)
3. **Add produk ke cart**
4. **Checkout:**
   - Pilih alamat
   - Pilih kurir: **JNE REG** atau **SiCepat REG**
   - Klik "Bayar Sekarang"

**Cek Backend Log:**
```
📦 Creating Biteship draft order for order 123
✅ Created Biteship draft order: draft_order_abc123 for order 123
```

**Jika TIDAK ada log ini:**
- Backend belum restart dengan code baru
- Atau ada error saat create draft order

---

### Step 3: Verifikasi Draft Order di Database

```sql
-- Cek order terbaru
SELECT order_code, status FROM orders ORDER BY created_at DESC LIMIT 1;

-- Cek draft order ID
SELECT 
  o.order_code,
  o.status,
  s.biteship_draft_order_id,
  s.provider_code
FROM orders o
LEFT JOIN shipments s ON o.id = s.order_id
WHERE o.order_code = 'ZVR-xxx';  -- Ganti dengan order code terbaru
```

**Expected Result:**
```
order_code: ZVR-20260129-xxx
status: PENDING
biteship_draft_order_id: draft_order_abc123  ← HARUS ADA!
provider_code: jne
```

**Jika biteship_draft_order_id NULL:**
- Draft order gagal dibuat
- Cek backend log untuk error
- Verifikasi TOKEN_BITESHIP valid

---

### Step 4: Bayar Order (Simulate Payment)

**Option A: Via Midtrans Sandbox**
1. Pilih payment method (VA/QRIS)
2. Bayar via Midtrans sandbox
3. Order status → PAID

**Option B: Manual Update (Testing)**
```sql
-- Mark as paid manually
UPDATE orders SET status = 'PAID', paid_at = NOW() WHERE order_code = 'ZVR-xxx';
UPDATE order_payments SET payment_status = 'PAID', paid_at = NOW() WHERE order_id = (SELECT id FROM orders WHERE order_code = 'ZVR-xxx');
```

---

### Step 5: Admin Pack Order

1. **Login sebagai admin:** http://localhost:3000/admin
2. **Buka Orders → Pilih order PAID**
3. **Klik "Proses Pesanan"**
4. Order status → PACKING

---

### Step 6: Admin Kirim Pesanan (AUTO-GENERATE RESI)

1. **Klik "Kirim Pesanan"**
2. **KOSONGKAN input resi** (jangan isi apa-apa!)
3. **Klik "Confirm"**

**Cek Backend Log:**
```
🚀 Auto-generating resi via Biteship for order ZVR-xxx
📦 Confirming Biteship draft order: draft_order_abc123
✅ Got resi from Biteship: JNE1234567890 (Tracking: track_ghi789)
✅ Order ZVR-xxx shipped with resi: JNE1234567890
```

**Modal Muncul:**
```
┌─────────────────────────────────────────┐
│ ✅ Resi Berhasil Di-Generate!           │
│                                         │
│ Nomor resi dari Biteship:               │
│ JNE1234567890  ← RESI REAL!            │
│                                         │
│ [OK]                                    │
└─────────────────────────────────────────┘
```

---

### Step 7: Verifikasi Resi di Database

```sql
-- Cek resi
SELECT 
  o.order_code,
  o.status,
  o.resi,
  s.tracking_number,
  s.biteship_tracking_id,
  s.biteship_waybill_id
FROM orders o
LEFT JOIN shipments s ON o.id = s.order_id
WHERE o.order_code = 'ZVR-xxx';
```

**Expected Result:**
```
order_code: ZVR-xxx
status: SHIPPED
resi: JNE1234567890  ← RESI REAL dari Biteship!
tracking_number: JNE1234567890
biteship_tracking_id: track_ghi789
biteship_waybill_id: JNE1234567890
```

**Verifikasi Format Resi:**
- ✅ REAL: `JNE1234567890` (format kurir)
- ❌ DUMMY: `JNE-123-1738123456` (format code)

---

### Step 8: Admin Lihat Resi

1. **Refresh order detail page**
2. **Scroll ke Shipment card**
3. **Lihat resi dengan button "Copy"**

**Expected:**
```
┌─────────────────────────────────────────┐
│ 🚚 Shipment                             │
│ Status: SHIPPED                         │
│                                         │
│ Nomor Resi              [Copy]          │
│ JNE1234567890                           │
│                                         │
│ Kurir: JNE REG                          │
│ Dikirim: 29 Jan 2025, 10:00            │
└─────────────────────────────────────────┘
```

---

### Step 9: Customer Track Pesanan

1. **Copy resi:** `JNE1234567890`
2. **Buka:** http://localhost:3000/track/JNE1234567890
3. **Lihat tracking page**

**Expected:**
```
┌─────────────────────────────────────────┐
│ Lacak Pesanan                           │
│                                         │
│ Nomor Resi: JNE1234567890               │
│ Kurir: JNE REG                          │
│ Status: SHIPPED                         │
│                                         │
│ Riwayat Pengiriman:                     │
│ ● Paket telah dibuat                    │
│   29 Jan 2025, 10:00                    │
└─────────────────────────────────────────┘
```

---

## 🐛 Troubleshooting

### Issue 1: Draft Order Tidak Dibuat
**Symptom:** `biteship_draft_order_id` NULL di database

**Check:**
```bash
# Cek backend log saat checkout
# Harus ada: "Creating Biteship draft order for order 123"
```

**Possible Causes:**
1. Backend belum restart dengan code baru
2. TOKEN_BITESHIP tidak valid
3. Biteship API error

**Solution:**
```bash
# 1. Restart backend
cd backend
.\zavera_COMPLETE.exe

# 2. Cek .env
cat .env | grep TOKEN_BITESHIP
# Harus ada: TOKEN_BITESHIP=biteship_test.eyJ...

# 3. Test Biteship API
curl -X GET "https://api.biteship.com/v1/couriers" \
  -H "Authorization: Bearer biteship_test.eyJ..."
```

---

### Issue 2: Resi Format Dummy (JNE-123-xxx)
**Symptom:** Resi format `JNE-123-1738123456`

**Cause:** Biteship confirm failed, fallback ke manual resi

**Check Backend Log:**
```
⚠️ Failed to confirm Biteship order: [error]
⚠️ No Biteship draft order found, generating manual resi
```

**Solution:**
- Verifikasi draft order ID ada di database
- Cek Biteship API status
- Create order baru untuk testing

---

### Issue 3: Modal Tidak Muncul
**Symptom:** Setelah klik "Confirm", modal tidak muncul

**Cause:** Frontend belum update

**Solution:**
```bash
# Restart frontend
cd frontend
npm run dev
```

---

## ✅ Success Criteria

### 1. Draft Order Created
```sql
SELECT biteship_draft_order_id FROM shipments WHERE order_id = 123;
-- Result: draft_order_abc123 (NOT NULL!)
```

### 2. Resi from Biteship
```sql
SELECT resi FROM orders WHERE order_code = 'ZVR-xxx';
-- Result: JNE1234567890 (NOT JNE-123-xxx!)
```

### 3. Admin Sees Resi
- ✅ Modal muncul dengan resi
- ✅ Shipment card menampilkan resi
- ✅ Button "Copy" berfungsi

### 4. Customer Can Track
- ✅ Tracking page accessible
- ✅ Resi info displayed
- ✅ Tracking history shown

---

## 📝 Quick Test Script

```bash
# 1. Restart backend
cd backend
.\zavera_COMPLETE.exe

# 2. Create order baru via frontend
# - Login customer
# - Checkout dengan JNE REG
# - Bayar

# 3. Cek draft order
psql -U postgres -d zavera_db -c "SELECT o.order_code, s.biteship_draft_order_id FROM orders o LEFT JOIN shipments s ON o.id = s.order_id ORDER BY o.created_at DESC LIMIT 1;"

# 4. Admin kirim pesanan
# - Login admin
# - Pack order
# - Kirim pesanan (kosongkan resi)

# 5. Cek resi
psql -U postgres -d zavera_db -c "SELECT order_code, resi FROM orders ORDER BY created_at DESC LIMIT 1;"

# 6. Verifikasi format resi
# ✅ REAL: JNE1234567890
# ❌ DUMMY: JNE-123-xxx
```

---

## 🎯 Next Steps

1. **Restart backend** dengan code baru
2. **Create order baru** (order lama tidak punya draft order)
3. **Test full flow** dari checkout sampai tracking
4. **Verifikasi resi** dari Biteship API (bukan dummy)

**PENTING:** Order lama (sebelum update) tidak akan punya draft order. Harus create order baru untuk test!
