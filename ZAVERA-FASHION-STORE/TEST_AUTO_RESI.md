# 🧪 Test Guide: Auto-Generate Resi dari Biteship

## Quick Test Steps

### Prerequisites
1. Backend running: `.\zavera_COMPLETE.exe`
2. Frontend running: `npm run dev`
3. Biteship token configured in `.env`

---

## Test Scenario 1: Auto-Generate Resi (Happy Path)

### Step 1: Create Order
1. Buka frontend: http://localhost:3000
2. Login sebagai customer
3. Tambah produk ke cart
4. Checkout dengan alamat lengkap
5. Pilih kurir: **JNE REG** atau **SiCepat REG**
6. Klik "Bayar Sekarang"

**Expected Result:**
- ✅ Order created dengan status PENDING
- ✅ Draft order Biteship created (cek log backend)
- ✅ `shipments.biteship_draft_order_id` terisi

### Step 2: Payment
1. Pilih metode pembayaran (VA/QRIS/GoPay)
2. Lakukan pembayaran (test mode)
3. Order status berubah PAID

**Expected Result:**
- ✅ Order status: PAID
- ✅ Payment status: PAID
- ✅ Draft order masih tersimpan

### Step 3: Pack Order
1. Login sebagai admin: http://localhost:3000/admin
2. Buka Orders → Pilih order yang baru dibayar
3. Klik "Pack Order"

**Expected Result:**
- ✅ Order status: PACKING
- ✅ Button "Kirim Pesanan" muncul

### Step 4: Ship Order (AUTO-GENERATE RESI)
1. Klik "Kirim Pesanan"
2. **JANGAN input resi manual** (biarkan kosong atau langsung klik)
3. Tunggu beberapa detik

**Expected Result:**
- ✅ Order status: SHIPPED
- ✅ Nomor resi otomatis muncul (contoh: JNE1234567890)
- ✅ Resi terlihat di order detail
- ✅ Customer dapat email dengan resi

**Backend Log:**
```
🚀 Auto-generating resi via Biteship for order ORD-123
📦 Confirming Biteship draft order: draft_order_123abc
✅ Got resi from Biteship: JNE1234567890 (Tracking: track_789ghi)
✅ Order ORD-123 shipped with resi: JNE1234567890
```

---

## Test Scenario 2: Fallback Manual Resi

### Step 1: Create Order Without Draft
1. Buat order baru (sama seperti Test 1)
2. Hapus draft order ID dari database (simulate failure):
   ```sql
   UPDATE shipments 
   SET biteship_draft_order_id = NULL 
   WHERE order_id = (SELECT id FROM orders WHERE order_code = 'ORD-123');
   ```

### Step 2: Ship Order
1. Admin klik "Kirim Pesanan"
2. Biarkan kosong (tidak input resi)

**Expected Result:**
- ✅ Order status: SHIPPED
- ✅ Resi manual generated (contoh: JNE-123-1738123456)
- ✅ Order tetap bisa dikirim

**Backend Log:**
```
🚀 Auto-generating resi via Biteship for order ORD-123
⚠️ No Biteship draft order found, generating manual resi
✅ Order ORD-123 shipped with resi: JNE-123-1738123456
```

---

## Test Scenario 3: Manual Input (Backward Compatible)

### Step 1: Ship with Manual Resi
1. Admin klik "Kirim Pesanan"
2. **Input resi manual:** `TEST123456789`
3. Klik "Kirim"

**Expected Result:**
- ✅ Order status: SHIPPED
- ✅ Resi yang diinput muncul: TEST123456789
- ✅ Tidak call Biteship API

**Backend Log:**
```
✅ Order ORD-123 shipped with resi: TEST123456789
```

---

## Verification Checklist

### Database Check
```sql
-- Check shipment details
SELECT 
    s.id,
    o.order_code,
    s.tracking_number,
    s.biteship_draft_order_id,
    s.biteship_order_id,
    s.biteship_tracking_id,
    s.biteship_waybill_id,
    s.status
FROM shipments s
JOIN orders o ON s.order_id = o.id
WHERE o.order_code = 'ORD-123';
```

**Expected Result:**
```
tracking_number:          JNE1234567890
biteship_draft_order_id:  draft_order_123abc
biteship_order_id:        order_456def
biteship_tracking_id:     track_789ghi
biteship_waybill_id:      JNE1234567890
status:                   SHIPPED
```

### API Check (Optional)
```bash
# Check Biteship tracking
curl -X GET "https://api.biteship.com/v1/trackings/track_789ghi" \
  -H "Authorization: Bearer biteship_test.eyJ..."
```

**Expected Response:**
```json
{
  "success": true,
  "waybill_id": "JNE1234567890",
  "status": "confirmed",
  "history": [...]
}
```

---

## Common Issues & Solutions

### Issue 1: "No draft order ID found"
**Cause:** Draft order tidak dibuat saat checkout
**Check:**
```sql
SELECT biteship_draft_order_id FROM shipments WHERE order_id = 123;
```
**Solution:** System otomatis fallback ke manual resi

### Issue 2: "Failed to confirm Biteship order"
**Cause:** Draft order expired atau invalid
**Check Backend Log:**
```
⚠️ Failed to confirm Biteship order: draft order expired
```
**Solution:** System otomatis fallback ke manual resi

### Issue 3: Resi tidak muncul
**Cause:** Update shipment gagal
**Check:**
```sql
SELECT resi FROM orders WHERE order_code = 'ORD-123';
SELECT tracking_number FROM shipments WHERE order_id = 123;
```
**Solution:** Resi tetap tersimpan di `orders.resi`

---

## Performance Test

### Test Multiple Orders
1. Create 5 orders
2. Pay all orders
3. Pack all orders
4. Ship all orders (auto-generate resi)

**Expected Result:**
- ✅ All orders get unique resi numbers
- ✅ No duplicate resi
- ✅ All orders status SHIPPED
- ✅ Response time < 3 seconds per order

---

## Production Readiness Checklist

- [ ] Test auto-generate resi works
- [ ] Test fallback manual resi works
- [ ] Test manual input works
- [ ] Verify database updates correctly
- [ ] Check email notifications sent
- [ ] Verify tracking works
- [ ] Test with multiple couriers (JNE, SiCepat, J&T)
- [ ] Check error handling
- [ ] Verify logs are clear
- [ ] Test with production Biteship token

---

## Success Criteria

✅ **Auto-generate resi works:**
- Admin klik "Kirim Pesanan" → Resi otomatis muncul
- Tidak perlu input manual
- Resi dari Biteship API

✅ **Fallback works:**
- Jika Biteship gagal → Manual resi generated
- Order tetap bisa dikirim
- Tidak error

✅ **Backward compatible:**
- Admin masih bisa input resi manual
- System gunakan resi yang diinput
- Tidak break existing flow

---

## Next Steps After Testing

### If All Tests Pass ✅
1. Update production `.env` dengan Biteship production token
2. Deploy `zavera_COMPLETE.exe` ke production
3. Monitor logs untuk error
4. Inform team bahwa fitur sudah live

### If Tests Fail ❌
1. Check backend logs untuk error detail
2. Verify Biteship token valid
3. Check database schema correct
4. Review code changes
5. Re-test after fixes

---

**Happy Testing! 🧪**
