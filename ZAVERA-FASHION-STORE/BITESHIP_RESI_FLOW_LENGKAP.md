# 🚀 Biteship Auto-Resi Flow - LENGKAP & REAL

## ✅ STATUS: COMPLETE - RESI DARI BITESHIP API (BUKAN DUMMY!)

---

## 🎯 Cara Kerja (End-to-End)

### 1️⃣ CUSTOMER CHECKOUT
**Apa yang Terjadi:**
```
Customer → Pilih Kurir (JNE REG) → Klik "Bayar Sekarang"
```

**Backend Process:**
```go
// checkout_service.go - CheckoutWithShipping()

1. Create Order (status: PENDING)
2. Create Shipment (status: PENDING)
3. 🔥 CREATE BITESHIP DRAFT ORDER 🔥
   ↓
   POST https://api.biteship.com/v1/draft_orders
   {
     "origin_area_id": "IDNP10IDNC393IDND4700",
     "destination_area_id": "IDNP6IDNC76IDND760",
     "courier_code": "jne",
     "courier_service_code": "reg",
     "items": [...]
   }
   ↓
   Response: {
     "success": true,
     "id": "draft_order_abc123",  ← SIMPAN INI!
     "status": "draft"
   }

4. Save draft_order_id ke database:
   UPDATE shipments 
   SET biteship_draft_order_id = 'draft_order_abc123'
   WHERE order_id = 123;
```

**Database State:**
```sql
orders:
  id: 123
  order_code: ORD-123
  status: PENDING
  
shipments:
  id: 456
  order_id: 123
  biteship_draft_order_id: 'draft_order_abc123'  ← PENTING!
  status: PENDING
```

---

### 2️⃣ CUSTOMER BAYAR
**Apa yang Terjadi:**
```
Customer → Bayar via VA/QRIS → Payment Success
```

**Backend Process:**
```
Midtrans Webhook → Update Order Status
  PENDING → PAID
  
Draft order masih tersimpan, siap dikonfirmasi!
```

---

### 3️⃣ ADMIN PACK ORDER
**Apa yang Terjadi:**
```
Admin → Klik "Proses Pesanan"
```

**Backend Process:**
```
Order Status: PAID → PACKING
Draft order masih tersimpan!
```

---

### 4️⃣ ADMIN KIRIM PESANAN (AUTO-GENERATE RESI)
**Apa yang Terjadi:**
```
Admin → Klik "Kirim Pesanan" → KOSONGKAN input resi → Klik "Confirm"
```

**Backend Process:**
```go
// admin_order_service.go - ShipOrder()

1. Check if resi input is empty
   if resi == "" {
     
2. Check if draft order exists
   SELECT biteship_draft_order_id FROM shipments WHERE order_id = 123;
   Result: 'draft_order_abc123' ✅
   
3. 🔥 CONFIRM DRAFT ORDER TO BITESHIP 🔥
   ↓
   POST https://api.biteship.com/v1/draft_orders/draft_order_abc123/confirm
   ↓
   Response: {
     "success": true,
     "id": "order_def456",
     "waybill_id": "JNE1234567890",  ← INI RESI REAL DARI BITESHIP!
     "tracking_id": "track_ghi789",
     "status": "confirmed"
   }
   
4. Save resi to database:
   UPDATE orders SET resi = 'JNE1234567890' WHERE id = 123;
   UPDATE shipments SET 
     tracking_number = 'JNE1234567890',
     biteship_tracking_id = 'track_ghi789',
     biteship_waybill_id = 'JNE1234567890',
     status = 'SHIPPED'
   WHERE id = 456;
   
5. Return resi to frontend:
   return "JNE1234567890", nil
}
```

**Frontend Process:**
```typescript
// admin/orders/[code]/page.tsx - handleShipOrder()

const response = await api.post('/admin/orders/ORD-123/ship', { resi: "" });

// Response dari backend:
{
  "message": "Order shipped successfully",
  "status": "SHIPPED",
  "resi": "JNE1234567890"  ← RESI REAL DARI BITESHIP!
}

// Tampilkan modal dengan resi:
setConfirmConfig({
  title: '✅ Resi Berhasil Di-Generate!',
  message: `Nomor resi dari Biteship:\n\nJNE1234567890\n\n...`,
  ...
});
setShowConfirm(true);
```

**Admin Lihat Resi:**
```
┌─────────────────────────────────────────┐
│ ✅ Resi Berhasil Di-Generate!           │
├─────────────────────────────────────────┤
│ Nomor resi dari Biteship:               │
│                                         │
│ JNE1234567890  ← RESI REAL!            │
│                                         │
│ Pesanan telah dikirim dan customer     │
│ akan menerima email notifikasi.        │
│                                         │
│ [OK]                                    │
└─────────────────────────────────────────┘
```

---

## 🔍 Verifikasi Resi REAL dari Biteship

### 1. Cek Backend Log
```
📦 Creating Biteship draft order for order 123
✅ Created Biteship draft order: draft_order_abc123 for order 123

🚀 Auto-generating resi via Biteship for order ORD-123
📦 Confirming Biteship draft order: draft_order_abc123
✅ Got resi from Biteship: JNE1234567890 (Tracking: track_ghi789)
✅ Order ORD-123 shipped with resi: JNE1234567890
```

### 2. Cek Database
```sql
-- Check draft order ID (created at checkout)
SELECT biteship_draft_order_id FROM shipments WHERE order_id = 123;
-- Result: draft_order_abc123 ✅

-- Check resi (generated when shipped)
SELECT resi FROM orders WHERE id = 123;
-- Result: JNE1234567890 ✅

-- Check Biteship tracking info
SELECT 
  tracking_number,
  biteship_tracking_id,
  biteship_waybill_id
FROM shipments WHERE order_id = 123;
-- Result:
-- tracking_number: JNE1234567890
-- biteship_tracking_id: track_ghi789
-- biteship_waybill_id: JNE1234567890
```

### 3. Cek Biteship Dashboard
Login ke https://dashboard.biteship.com
- Lihat "Orders" → Cari order dengan ID `order_def456`
- Verifikasi waybill_id: `JNE1234567890`
- Status: `confirmed`

### 4. Track Resi via Biteship API
```bash
curl -X GET "https://api.biteship.com/v1/trackings/track_ghi789" \
  -H "Authorization: Bearer biteship_test.eyJ..."
```

Response:
```json
{
  "success": true,
  "waybill_id": "JNE1234567890",
  "courier_code": "jne",
  "status": "confirmed",
  "history": [
    {
      "note": "Paket telah dibuat",
      "status": "confirmed",
      "updated_at": "2025-01-29T10:00:00Z"
    }
  ]
}
```

---

## 🚨 Perbedaan: REAL vs DUMMY

### ❌ DUMMY Resi (Fallback Manual)
```
Format: JNE-123-1738123456
Source: Generated by code (resi_service.go)
Trackable: NO
Biteship Dashboard: NOT FOUND
```

### ✅ REAL Resi (Biteship API)
```
Format: JNE1234567890 (format kurir asli)
Source: Biteship API (waybill_id)
Trackable: YES via Biteship API
Biteship Dashboard: FOUND with order details
```

---

## 📋 Test Checklist

### ✅ Test 1: Verifikasi Draft Order Created
1. Customer checkout dengan JNE REG
2. Cek backend log:
   ```
   📦 Creating Biteship draft order for order 123
   ✅ Created Biteship draft order: draft_order_abc123
   ```
3. Cek database:
   ```sql
   SELECT biteship_draft_order_id FROM shipments WHERE order_id = 123;
   -- Harus ada value: draft_order_abc123
   ```

### ✅ Test 2: Verifikasi Resi dari Biteship
1. Admin klik "Kirim Pesanan"
2. KOSONGKAN input resi
3. Klik "Confirm"
4. Cek backend log:
   ```
   🚀 Auto-generating resi via Biteship for order ORD-123
   📦 Confirming Biteship draft order: draft_order_abc123
   ✅ Got resi from Biteship: JNE1234567890
   ```
5. Modal muncul dengan resi: `JNE1234567890`
6. Cek database:
   ```sql
   SELECT resi FROM orders WHERE order_code = 'ORD-123';
   -- Result: JNE1234567890 (bukan JNE-123-xxx)
   ```

### ✅ Test 3: Verifikasi Resi Trackable
1. Copy resi: `JNE1234567890`
2. Track via Biteship API:
   ```bash
   curl -X GET "https://api.biteship.com/v1/trackings/track_ghi789" \
     -H "Authorization: Bearer biteship_test.eyJ..."
   ```
3. Response harus success dengan tracking history

---

## 🐛 Troubleshooting

### Issue: "No draft order ID found"
**Cause:** Draft order tidak dibuat saat checkout
**Check:**
```sql
SELECT biteship_draft_order_id FROM shipments WHERE order_id = 123;
-- Result: NULL atau empty
```
**Solution:** 
- Cek backend log saat checkout
- Verifikasi Biteship token valid
- Cek error di log: "Failed to create Biteship draft order"

### Issue: Resi format dummy (JNE-123-xxx)
**Cause:** Biteship confirm failed, fallback ke manual resi
**Check Backend Log:**
```
⚠️ Failed to confirm Biteship order: [error detail]
⚠️ No Biteship draft order found, generating manual resi
```
**Solution:**
- Verifikasi draft order ID ada di database
- Cek Biteship API status
- Verifikasi token valid

### Issue: "Failed to confirm Biteship order"
**Cause:** Draft order expired atau invalid
**Check:**
- Draft order dibuat > 24 jam yang lalu? (expired)
- Draft order sudah dikonfirmasi sebelumnya? (duplicate)
**Solution:**
- Create order baru untuk testing
- Jangan re-use order lama

---

## 🎉 Success Criteria

✅ **Draft Order Created at Checkout:**
- Backend log: "Created Biteship draft order: draft_order_abc123"
- Database: `biteship_draft_order_id` terisi

✅ **Resi from Biteship API:**
- Backend log: "Got resi from Biteship: JNE1234567890"
- Format: Sesuai format kurir (bukan JNE-123-xxx)
- Database: `resi` dan `tracking_number` terisi

✅ **Admin Sees Resi:**
- Modal muncul dengan resi dari Biteship
- Resi muncul di Shipment card
- Customer dapat email dengan resi

✅ **Resi Trackable:**
- Bisa di-track via Biteship API
- Muncul di Biteship dashboard
- Status update real-time

---

## 🚀 Production Deployment

### 1. Update Environment
```env
# Production Biteship token
TOKEN_BITESHIP=biteship_live.eyJ...

# Production Biteship URL (same as test)
BITESHIP_BASE_URL=https://api.biteship.com
```

### 2. Deploy Backend
```bash
# Build production
go build -o zavera_production.exe .

# Run
.\zavera_production.exe
```

### 3. Verify Production
- Test checkout → Verify draft order created
- Test ship order → Verify resi from Biteship
- Check Biteship dashboard → Verify order appears
- Track resi → Verify tracking works

---

## 📚 API Documentation

### Create Draft Order
```
POST https://api.biteship.com/v1/draft_orders
Authorization: Bearer {TOKEN_BITESHIP}

Request:
{
  "origin_area_id": "IDNP10IDNC393IDND4700",
  "destination_area_id": "IDNP6IDNC76IDND760",
  "courier_code": "jne",
  "courier_service_code": "reg",
  "items": [...]
}

Response:
{
  "success": true,
  "id": "draft_order_abc123",
  "status": "draft"
}
```

### Confirm Draft Order
```
POST https://api.biteship.com/v1/draft_orders/{draft_order_id}/confirm
Authorization: Bearer {TOKEN_BITESHIP}

Response:
{
  "success": true,
  "id": "order_def456",
  "waybill_id": "JNE1234567890",  ← RESI REAL!
  "tracking_id": "track_ghi789",
  "status": "confirmed"
}
```

### Track Shipment
```
GET https://api.biteship.com/v1/trackings/{tracking_id}
Authorization: Bearer {TOKEN_BITESHIP}

Response:
{
  "success": true,
  "waybill_id": "JNE1234567890",
  "status": "confirmed",
  "history": [...]
}
```

---

**RESI SEKARANG 100% DARI BITESHIP API - BUKAN DUMMY!** ✅
