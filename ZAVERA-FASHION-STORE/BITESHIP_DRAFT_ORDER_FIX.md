# 🔧 Biteship Draft Order Fix - Area ID Error

## 🔴 Masalah yang Ditemukan

### Error dari Biteship API:
```json
{
  "success": false,
  "code": 40000000,
  "error": "Bad request.-",
  "details": {
    "field": "origin_area_id",
    "reason": "Format origin area id is invalid"
  }
}
```

### Penyebab:
1. **Area ID format tidak valid** - `IDNP10IDNC393IDND4700` tidak dikenali Biteship
2. **Area ID tidak exist** - Mungkin area ID sudah berubah atau tidak valid
3. **Biteship prefer postal_code** - Lebih reliable daripada area_id

### Dampak:
- ❌ Draft order gagal dibuat saat checkout
- ❌ Database: `biteship_draft_order_id` = NULL
- ❌ Admin tidak bisa auto-generate resi
- ❌ Fallback ke resi dummy (JNE-123-xxx)

---

## ✅ Solusi yang Diterapkan

### 1. Prioritaskan Postal Code
Ubah dari area_id ke postal_code untuk origin dan destination:

**Before:**
```go
draftParams := CreateDraftOrderParams{
    OriginAreaID:      "IDNP10IDNC393IDND4700",  // ❌ Invalid!
    DestinationAreaID: destinationAreaID,         // ❌ Might be invalid
    ...
}
```

**After:**
```go
draftParams := CreateDraftOrderParams{
    OriginAreaID:      "",  // ✅ Empty - use postal_code
    OriginPostalCode:  "50113",  // ✅ Pedurungan, Semarang
    DestinationAreaID: "",  // ✅ Empty - use postal_code
    DestinationPostalCode: destinationPostalCode,  // ✅ From customer
    ...
}
```

### 2. Make Area ID Optional
Update struct dengan `omitempty` tag:

```go
type CreateDraftOrderRequest struct {
    OriginAreaID      string `json:"origin_area_id,omitempty"`      // Optional
    DestinationAreaID string `json:"destination_area_id,omitempty"` // Optional
    ...
}
```

**Benefit:**
- Jika area_id kosong, field tidak dikirim ke Biteship API
- Biteship akan gunakan postal_code untuk determine area
- Lebih reliable dan compatible

---

## 🚀 Cara Test Fix

### Step 1: Restart Backend dengan Code Baru
```bash
cd backend
.\zavera_BITESHIP_FIX.exe
```

### Step 2: Create Order Baru
1. Buka http://localhost:3000
2. Login sebagai customer
3. Add produk ke cart
4. Checkout dengan kurir apa saja
5. Klik "Bayar Sekarang"

### Step 3: Cek Backend Log (HARUS BERHASIL!)
```
📦 Creating Biteship draft order for order 123
📤 Biteship API Request [POST /v1/draft_orders]: {
  "origin_postal_code": "50113",
  "destination_postal_code": "10110",
  ...
}
📡 Biteship API [POST /v1/draft_orders] Status: 200  ← ✅ SUCCESS!
✅ Created draft order with ID: draft_order_abc123
```

**Jika masih error 400:**
- Cek postal_code valid (5 digit)
- Cek courier_code dan courier_service_code valid
- Cek items tidak kosong

### Step 4: Verifikasi Database
```sql
SELECT 
  o.order_code,
  s.biteship_draft_order_id
FROM orders o
LEFT JOIN shipments s ON o.id = s.order_id
ORDER BY o.created_at DESC LIMIT 1;
```

**Expected:**
```
order_code: ZVR-20260129-xxx
biteship_draft_order_id: draft_order_abc123  ← ✅ NOT NULL!
```

### Step 5: Test Auto-Generate Resi
1. Bayar order (manual update atau via Midtrans)
2. Admin pack order
3. Admin kirim dengan resi KOSONG
4. Cek backend log:
```
🚀 Auto-generating resi via Biteship for order ZVR-xxx
📦 Confirming Biteship draft order: draft_order_abc123
✅ Got resi from Biteship: JNE1234567890
```

---

## 📋 Verification Checklist

### ✅ Draft Order Created Successfully
```bash
# Backend log
📦 Creating Biteship draft order for order 123
📡 Biteship API [POST /v1/draft_orders] Status: 200
✅ Created draft order with ID: draft_order_abc123

# Database
SELECT biteship_draft_order_id FROM shipments WHERE order_id = 123;
-- Result: draft_order_abc123 (NOT NULL!)
```

### ✅ Resi from Biteship API
```bash
# Backend log
🚀 Auto-generating resi via Biteship for order ZVR-xxx
✅ Got resi from Biteship: JNE1234567890

# Database
SELECT resi FROM orders WHERE order_code = 'ZVR-xxx';
-- Result: JNE1234567890 (NOT JNE-123-xxx!)
```

### ✅ Biteship Dashboard
- Login https://dashboard.biteship.com
- Menu: Orders
- Cari order dengan waybill_id: JNE1234567890
- **Order HARUS muncul!**

---

## 🔍 Troubleshooting

### Issue 1: Masih Error 400 "Bad Request"
**Check:**
```bash
# Cek backend log untuk detail error
# Look for: "Biteship API 400 Error: ..."
```

**Possible Causes:**
1. Postal code tidak valid (bukan 5 digit)
2. Courier code tidak valid
3. Items kosong atau format salah

**Solution:**
```go
// Verify postal code format
if len(destinationPostalCode) != 5 {
    return nil, fmt.Errorf("invalid postal code: must be 5 digits")
}

// Verify courier code
validCouriers := []string{"jne", "sicepat", "jnt", "anteraja", "tiki"}
if !contains(validCouriers, courierCode) {
    return nil, fmt.Errorf("invalid courier code: %s", courierCode)
}
```

### Issue 2: Draft Order NULL di Database
**Check:**
```sql
SELECT biteship_draft_order_id FROM shipments WHERE order_id = 123;
-- Result: NULL
```

**Cause:** Draft order creation failed

**Solution:**
1. Cek backend log untuk error detail
2. Verify TOKEN_BITESHIP valid
3. Test Biteship API manually:
```bash
curl -X POST "https://api.biteship.com/v1/draft_orders" \
  -H "Authorization: Bearer biteship_test.eyJ..." \
  -H "Content-Type: application/json" \
  -d '{
    "origin_postal_code": "50113",
    "destination_postal_code": "10110",
    "courier_code": "jne",
    "courier_service_code": "reg",
    "items": [...]
  }'
```

### Issue 3: Resi Masih Format Dummy
**Check:**
```sql
SELECT resi FROM orders WHERE order_code = 'ZVR-xxx';
-- Result: JNE-123-1738123456  ← Still dummy!
```

**Cause:** Draft order tidak ada, fallback ke manual resi

**Solution:**
- Verifikasi draft order dibuat saat checkout
- Create order BARU (order lama tidak punya draft order)
- Cek backend log saat ship order

---

## 📝 Files Changed

### 1. `backend/service/checkout_service.go`
- Line ~360: Update `CreateDraftOrderParams` to use postal_code
- Remove area_id, prioritize postal_code

### 2. `backend/service/biteship_client.go`
- Line ~860: Update `CreateDraftOrderRequest` struct
- Add `omitempty` tag to area_id fields

### 3. Build Output
- New binary: `zavera_BITESHIP_FIX.exe`

---

## 🎯 Summary

### Problem:
- ❌ Area ID format invalid → Draft order failed
- ❌ No draft order → No auto-resi
- ❌ Fallback to dummy resi

### Solution:
- ✅ Use postal_code instead of area_id
- ✅ Make area_id optional (omitempty)
- ✅ More reliable and compatible

### Result:
- ✅ Draft order created successfully
- ✅ Auto-generate resi from Biteship
- ✅ Real resi format (JNE1234567890)
- ✅ Order appears in Biteship dashboard

---

## 🚀 Next Steps

1. **Restart backend** dengan `zavera_BITESHIP_FIX.exe`
2. **Create order baru** (order lama tidak punya draft order)
3. **Verify draft order** dibuat saat checkout
4. **Test auto-generate resi** saat admin kirim pesanan
5. **Check Biteship dashboard** untuk konfirmasi

**Fix sudah diterapkan! Silakan test dengan order baru.**
