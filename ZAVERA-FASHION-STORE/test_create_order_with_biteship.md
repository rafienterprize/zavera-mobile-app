# 🧪 Test Create Order Baru dengan Biteship Draft Order

## Masalah yang Ditemukan:

Order 114 (ZVR-20260129-136AC89B):
- ✅ Draft order ID ada: `f1023272-8cd5-4e2e-b245-7768b4bf669f`
- ❌ Status: "placed" (bukan "draft")
- ❌ Waybill ID: NULL
- ❌ Courier info: NULL

**Root Cause:** Draft order dibuat tapi Biteship tidak assign courier dan generate waybill.

## Kemungkinan Penyebab:

### 1. Courier Code Tidak Valid
Draft order mungkin dibuat tanpa `courier_code` dan `courier_service_code` yang benar.

**Check:**
```json
{
  "courier_code": "jne",  ← Harus ada
  "courier_service_code": "reg"  ← Harus ada
}
```

### 2. Route Tidak Tersedia
JNE REG mungkin tidak tersedia untuk route Pedurungan → Tembalang (sama-sama Semarang, jarak dekat).

**Solusi:** Test dengan courier lain (SiCepat, J&T, dll)

### 3. Data Tidak Lengkap
Draft order mungkin missing required fields.

## ✅ Cara Test yang Benar:

### Step 1: Create Order Baru
1. Buka http://localhost:3000
2. Login sebagai customer (BUKAN admin)
3. Add produk ke cart
4. Checkout
5. **PENTING:** Pilih courier yang berbeda (SiCepat HALU atau J&T REG)
6. Klik "Bayar Sekarang"

### Step 2: Monitor Backend Log
```bash
# Watch backend log real-time
Get-Content backend_log.txt -Wait
```

**Expected Log:**
```
📦 Creating Biteship draft order for order 115
📤 Biteship API Request [POST /v1/draft_orders]: {
  "origin_postal_code": "50113",
  "destination_postal_code": "50279",
  "courier_code": "sicepat",  ← Check ini!
  "courier_service_code": "halu",  ← Check ini!
  "items": [...]
}
📡 Biteship API [POST /v1/draft_orders] Status: 200
✅ Created draft order with ID: draft_order_xyz789
```

### Step 3: Check Draft Order Status
```powershell
$token = "biteship_test.eyJ...";
$draftId = "draft_order_xyz789";
$headers = @{ "Authorization" = "Bearer $token" };
Invoke-RestMethod -Uri "https://api.biteship.com/v1/draft_orders/$draftId" -Headers $headers | ConvertTo-Json -Depth 5
```

**Expected:**
```json
{
  "status": "draft",  ← Harus "draft", bukan "placed"!
  "courier": {
    "company": "sicepat",
    "waybill_id": null  ← OK untuk draft
  }
}
```

### Step 4: Bayar Order
```sql
UPDATE orders SET status = 'PAID' WHERE id = 115;
UPDATE order_payments SET payment_status = 'PAID', paid_at = NOW() WHERE order_id = 115;
```

### Step 5: Admin Pack Order
Admin Dashboard → Orders → Pilih order baru → Klik "Proses Pesanan"

### Step 6: Generate Resi
Klik "Kirim Pesanan" → Klik "Generate dari Biteship"

**Expected Backend Log:**
```
🚀 Generating resi from Biteship for order ZVR-xxx
📦 Confirming Biteship draft order: draft_order_xyz789
📡 Biteship API [POST /v1/draft_orders/draft_order_xyz789/confirm] Status: 200
✅ Confirmed order - Waybill: SICEPAT123456, Tracking: track_abc
✅ Got resi from Biteship: SICEPAT123456
```

**Expected UI:**
```
Input Field: [SICEPAT123456]  ← Resi REAL dari Biteship!
```

## 🔍 Debug Order 114:

Untuk order 114 yang sudah "placed" tanpa waybill, ada 2 opsi:

### Opsi 1: Fallback ke Manual Resi (CURRENT)
System sudah handle ini dengan baik - fallback ke `ZVR-JNE-20260129-114-FDDA`

### Opsi 2: Contact Biteship Support
Jika perlu resi REAL dari Biteship untuk order ini, contact Biteship support dengan draft order ID: `f1023272-8cd5-4e2e-b245-7768b4bf669f`

## 📋 Recommendation:

**Untuk Testing:**
1. ✅ Create order BARU dengan courier SiCepat atau J&T
2. ✅ Monitor backend log untuk verify draft order created correctly
3. ✅ Test generate resi dengan order baru

**Untuk Production:**
- System sudah handle fallback dengan baik
- Jika Biteship gagal, otomatis generate manual resi
- Order tetap bisa dikirim

## 🎯 Next Steps:

1. Create order baru dengan SiCepat HALU
2. Verify draft order status = "draft" (bukan "placed")
3. Test generate resi
4. Verify dapat resi REAL dari Biteship

---

**Note:** Order 114 kemungkinan issue dengan Biteship API saat dibuat. Order baru seharusnya work dengan benar.
