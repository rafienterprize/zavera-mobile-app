# 🧪 Cara Test Resi dari Biteship API (REAL)

## ⚠️ PENTING!

Order yang ada sekarang **TIDAK AKAN** dapat resi dari Biteship karena:
1. Order dibuat sebelum code update
2. Tidak ada `biteship_draft_order_id` di database
3. System fallback ke manual resi (dummy)

**Harus create ORDER BARU untuk test!**

---

## 📋 Step-by-Step Test

### Step 1: Verifikasi Backend Running dengan Code Baru

```bash
# Stop backend lama (Ctrl+C di terminal backend)

# Start backend baru
cd backend
.\zavera_COMPLETE.exe
```

**Cek log startup:**
```
Server running on :8080
Database connected
```

---

### Step 2: Create Order Baru (WAJIB!)

#### A. Via Frontend (Recommended)

1. **Buka:** http://localhost:3000
2. **Logout** jika sedang login sebagai admin
3. **Login sebagai customer** atau register baru
4. **Add produk ke cart:**
   - Pilih produk apa saja
   - Klik "Add to Cart"
5. **Checkout:**
   - Klik icon cart → "Checkout"
   - Isi alamat lengkap
   - **PENTING:** Pilih kurir **JNE REG** atau **SiCepat REG**
   - Klik "Bayar Sekarang"

**Cek Backend Log (PENTING!):**
```
📦 Creating Biteship draft order for order 123
POST https://api.biteship.com/v1/draft_orders
✅ Created Biteship draft order: draft_order_abc123 for order 123
```

**Jika TIDAK ada log ini:**
- Backend belum restart dengan code baru
- Atau TOKEN_BITESHIP tidak valid
- Atau Biteship API error

#### B. Verifikasi Draft Order di Database

```sql
-- Cek order terbaru
SELECT order_code, status FROM orders ORDER BY created_at DESC LIMIT 1;

-- Cek draft order (HARUS ADA!)
SELECT 
  o.order_code,
  o.status,
  s.biteship_draft_order_id,
  s.provider_code
FROM orders o
LEFT JOIN shipments s ON o.id = s.order_id
ORDER BY o.created_at DESC LIMIT 1;
```

**Expected Result:**
```
order_code: ZVR-20260129-NEWORDER
status: PENDING
biteship_draft_order_id: draft_order_abc123  ← HARUS ADA!
provider_code: jne
```

**Jika biteship_draft_order_id NULL:**
- ❌ Draft order GAGAL dibuat
- Backend belum restart
- TOKEN_BITESHIP tidak valid
- **STOP! Fix ini dulu sebelum lanjut**

---

### Step 3: Bayar Order

#### Option A: Via Midtrans Sandbox (Recommended)
1. Pilih payment method (VA/QRIS/GoPay)
2. Follow payment flow
3. Bayar via Midtrans sandbox

#### Option B: Manual Update (Testing Only)
```sql
-- Get order ID
SELECT id, order_code FROM orders ORDER BY created_at DESC LIMIT 1;

-- Mark as paid
UPDATE orders 
SET status = 'PAID', paid_at = NOW() 
WHERE order_code = 'ZVR-xxx';  -- Ganti dengan order code terbaru

-- Update payment
UPDATE order_payments 
SET payment_status = 'PAID', paid_at = NOW() 
WHERE order_id = (SELECT id FROM orders WHERE order_code = 'ZVR-xxx');
```

---

### Step 4: Admin Pack Order

1. **Login admin:** http://localhost:3000/admin
2. **Buka Orders**
3. **Pilih order PAID** (yang baru dibuat)
4. **Klik "Proses Pesanan"**
5. Order status → PACKING

---

### Step 5: Admin Kirim Pesanan (GENERATE RESI DARI BITESHIP)

1. **Klik "Kirim Pesanan"**
2. **KOSONGKAN input resi** (jangan isi apa-apa!)
3. **Klik "Confirm"**

**Cek Backend Log (CRITICAL!):**
```
🚀 Auto-generating resi via Biteship for order ZVR-xxx
📦 Confirming Biteship draft order: draft_order_abc123
POST https://api.biteship.com/v1/draft_orders/draft_order_abc123/confirm
✅ Got resi from Biteship: JNE1234567890 (Tracking: track_ghi789)
✅ Order ZVR-xxx shipped with resi: JNE1234567890
```

**Modal Harus Muncul:**
```
┌─────────────────────────────────────────┐
│ ✅ Resi Berhasil Di-Generate!           │
│                                         │
│ Nomor resi dari Biteship:               │
│ JNE1234567890  ← INI RESI REAL!        │
│                                         │
│ Pesanan telah dikirim dan customer     │
│ akan menerima email notifikasi.        │
│                                         │
│ [OK]                                    │
└─────────────────────────────────────────┘
```

---

### Step 6: Verifikasi Resi dari Biteship

#### A. Cek Database
```sql
SELECT 
  o.order_code,
  o.status,
  o.resi,
  s.biteship_draft_order_id,
  s.biteship_order_id,
  s.biteship_tracking_id,
  s.biteship_waybill_id
FROM orders o
LEFT JOIN shipments s ON o.id = s.order_id
WHERE o.order_code = 'ZVR-xxx';  -- Order terbaru
```

**Expected Result (REAL dari Biteship):**
```
order_code: ZVR-xxx
status: SHIPPED
resi: JNE1234567890  ← Format kurir asli
biteship_draft_order_id: draft_order_abc123  ← Ada!
biteship_order_id: order_def456  ← Ada!
biteship_tracking_id: track_ghi789  ← Ada!
biteship_waybill_id: JNE1234567890  ← Sama dengan resi!
```

**Jika Dummy/Fallback:**
```
resi: JNE-123-1738123456  ← Format code
biteship_draft_order_id: NULL  ← Tidak ada!
biteship_order_id: NULL
biteship_tracking_id: NULL
biteship_waybill_id: NULL
```

#### B. Cek Biteship Dashboard
1. Login: https://dashboard.biteship.com
2. Menu: **Orders**
3. Cari order dengan:
   - Order ID: `order_def456`
   - Waybill ID: `JNE1234567890`

**Harus muncul di dashboard!**

#### C. Cek Format Resi
- ✅ **REAL**: `JNE1234567890` (10-15 digit, format kurir)
- ❌ **DUMMY**: `JNE-123-1738123456` (ada dash, format code)

---

### Step 7: Admin Lihat Resi

1. **Refresh order detail page**
2. **Scroll ke Shipment card**

**Expected:**
```
┌─────────────────────────────────────────┐
│ 🚚 Shipment                             │
│ Status: SHIPPED                         │
│                                         │
│ Nomor Resi              [Copy]          │
│ JNE1234567890  ← RESI REAL!            │
│                                         │
│ Kurir: JNE REG                          │
│ Dikirim: 29 Jan 2025, 10:00            │
└─────────────────────────────────────────┘
```

**Klik "Copy":**
- ✅ Toast: "Resi berhasil di-copy!"
- Resi ter-copy ke clipboard

---

## 🔍 Troubleshooting

### Issue 1: Draft Order Tidak Dibuat

**Symptom:**
- Backend log tidak ada "Creating Biteship draft order"
- Database: `biteship_draft_order_id` NULL

**Check:**
```bash
# 1. Cek backend running
ps aux | grep zavera_COMPLETE

# 2. Cek .env
cat backend/.env | grep TOKEN_BITESHIP

# 3. Test Biteship API
curl -X GET "https://api.biteship.com/v1/couriers" \
  -H "Authorization: Bearer biteship_test.eyJ..."
```

**Solution:**
```bash
# Restart backend
cd backend
.\zavera_COMPLETE.exe

# Create order baru
# Cek log: "Creating Biteship draft order"
```

---

### Issue 2: Resi Format Dummy

**Symptom:**
- Resi: `JNE-123-1738123456`
- Backend log: "No Biteship draft order found"

**Cause:**
- Order tidak punya draft order
- Biteship confirm failed

**Solution:**
- Create order BARU
- Verifikasi draft order dibuat saat checkout

---

### Issue 3: Biteship API Error

**Symptom:**
- Backend log: "Failed to create Biteship draft order"
- Error: 401 Unauthorized

**Check:**
```bash
# Cek TOKEN_BITESHIP
cat backend/.env | grep TOKEN_BITESHIP

# Test token
curl -X GET "https://api.biteship.com/v1/couriers" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Solution:**
- Verifikasi token valid
- Login Biteship dashboard
- Generate token baru jika perlu

---

## ✅ Success Checklist

### 1. Draft Order Created
```sql
SELECT biteship_draft_order_id FROM shipments 
WHERE order_id = (SELECT id FROM orders ORDER BY created_at DESC LIMIT 1);
-- Result: draft_order_abc123 (NOT NULL!)
```

### 2. Resi from Biteship API
```sql
SELECT resi, biteship_waybill_id FROM orders 
ORDER BY created_at DESC LIMIT 1;
-- Result: JNE1234567890 (NOT JNE-123-xxx!)
```

### 3. Biteship Dashboard
- Login dashboard.biteship.com
- Orders → Cari waybill_id
- **Harus muncul!**

### 4. Admin Sees Resi
- Modal muncul dengan resi
- Shipment card menampilkan resi
- Button "Copy" berfungsi

---

## 📝 Quick Verification Script

```bash
# 1. Cek order terbaru
psql -U postgres -d zavera_db -c "SELECT order_code, status, resi FROM orders ORDER BY created_at DESC LIMIT 1;"

# 2. Cek draft order
psql -U postgres -d zavera_db -c "SELECT o.order_code, s.biteship_draft_order_id, s.biteship_waybill_id FROM orders o LEFT JOIN shipments s ON o.id = s.order_id ORDER BY o.created_at DESC LIMIT 1;"

# 3. Verifikasi format resi
# ✅ REAL: JNE1234567890 (no dash)
# ❌ DUMMY: JNE-123-xxx (with dash)
```

---

## 🎯 Summary

**Untuk dapat resi REAL dari Biteship:**

1. ✅ Backend harus running dengan code baru
2. ✅ Create ORDER BARU (order lama tidak punya draft order)
3. ✅ Verifikasi draft order dibuat saat checkout
4. ✅ Admin kirim dengan input resi KOSONG
5. ✅ Verifikasi resi format kurir (bukan JNE-123-xxx)
6. ✅ Cek Biteship dashboard (harus muncul)

**Order lama TIDAK AKAN dapat resi dari Biteship!**

**Harus test dengan order baru!**
