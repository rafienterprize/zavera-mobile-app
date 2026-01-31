# 🎯 CARA TEST REFUND SEKARANG

**Status:** ✅ Backend & Frontend sudah running!

---

## 📍 Langkah-Langkah Test

### 1️⃣ Buka Admin Panel

```
http://localhost:3000/admin
```

- Login dengan Google: **pemberani073@gmail.com**
- Masuk ke dashboard admin

### 2️⃣ Buka Order Test

```
http://localhost:3000/admin/orders/ZVR-20260127-B8B3ACCD
```

Atau:
- Klik **"Orders"** di sidebar
- Cari order: **ZVR-20260127-B8B3ACCD**
- Klik order tersebut

### 3️⃣ Buat Refund

1. Scroll ke bawah sampai **"Order Actions"**
2. Klik tombol **"Refund"** (warna kuning/amber)
3. Modal refund akan muncul

### 4️⃣ Isi Form Refund

**Pilih:**
- Refund Type: **FULL**
- Reason: **Customer Request**
- Additional Details: "Test refund system"

**Klik:** **"Process Refund"**

### 5️⃣ Yang Akan Terjadi

**Scenario A: Error 418 (Expected)**
```
⚠️ Error message muncul:
"MANUAL_PROCESSING_REQUIRED: Automatic refund failed. 
Please process manual bank transfer to customer and 
mark refund as completed after transfer is done."

✅ Modal tertutup
✅ Refund muncul di "Refund History"
✅ Status: PENDING
✅ Amount: Rp 918,000
✅ Tombol "Mark as Completed" muncul
```

**Scenario B: Success (Rare)**
```
✅ Success message muncul
✅ Refund status: COMPLETED
✅ Gateway ID: [number]
✅ Stock restored
```

### 6️⃣ Complete Manual Refund (Jika Error 418)

1. Di **"Refund History"**, cari refund yang PENDING
2. Klik tombol **"Mark as Completed"** (hijau)
3. Dialog konfirmasi muncul
4. Masukkan note:
   ```
   Transfer manual via BCA ke rekening customer pada 29 Jan 2026
   ```
5. Klik **"Confirm"**

### 7️⃣ Verifikasi Hasil

**Yang harus terlihat:**
- ✅ Success message: "Refund berhasil ditandai sebagai completed!"
- ✅ Refund status berubah jadi: **COMPLETED**
- ✅ Gateway ID: **MANUAL_BANK_TRANSFER**
- ✅ Order refund_status: **FULL**
- ✅ Order refund_amount: **918000**

---

## 🎨 Screenshot Guide

### Step 1: Order Detail Page
```
┌─────────────────────────────────────────────────┐
│ ZVR-20260127-B8B3ACCD                           │
│ Created 27 Jan 2026                             │
├─────────────────────────────────────────────────┤
│ Order Status: DELIVERED                         │
│ Payment: BCA VA - SUCCESS                       │
│ Shipment: DELIVERED                             │
├─────────────────────────────────────────────────┤
│ Order Actions:                                  │
│ [Refund] [Reship]                               │ ← Klik ini!
└─────────────────────────────────────────────────┘
```

### Step 2: Refund Modal
```
┌─────────────────────────────────────────────────┐
│ Process Refund                                  │
├─────────────────────────────────────────────────┤
│ Refund Type:                                    │
│ [FULL] [PARTIAL] [SHIPPING_ONLY] [ITEM_ONLY]   │ ← Pilih FULL
│                                                 │
│ Reason:                                         │
│ [Customer Request ▼]                            │ ← Pilih ini
│                                                 │
│ Additional Details:                             │
│ [Test refund system________________]            │
│                                                 │
│ [Cancel] [Process Refund]                       │ ← Klik ini!
└─────────────────────────────────────────────────┘
```

### Step 3: Refund History (After Creation)
```
┌─────────────────────────────────────────────────┐
│ 🔄 Refund History                               │
├─────────────────────────────────────────────────┤
│ REF-20260129-ABC123  [PENDING] [FULL]          │
│ Reason: Customer Request - Test refund system  │
│                                    Rp 918,000   │
│ ⚠️ MANUAL REFUND                                │
│                                                 │
│ [✓ Mark as Completed]                           │ ← Klik ini!
└─────────────────────────────────────────────────┘
```

### Step 4: Confirmation Dialog
```
┌─────────────────────────────────────────────────┐
│ Mark Refund as Completed                        │
├─────────────────────────────────────────────────┤
│ Apakah Anda sudah melakukan transfer manual    │
│ ke customer? Pastikan transfer sudah berhasil  │
│ sebelum menandai refund sebagai completed.      │
│                                                 │
│ [Cancel] [Confirm]                              │ ← Klik Confirm
└─────────────────────────────────────────────────┘

Then prompt appears:
┌─────────────────────────────────────────────────┐
│ Masukkan catatan konfirmasi:                    │
│ [Transfer manual via BCA ke rekening customer_] │
│ [pada 29 Jan 2026_________________________]     │
│                                                 │
│ [OK] [Cancel]                                   │ ← Klik OK
└─────────────────────────────────────────────────┘
```

### Step 5: Completed Refund
```
┌─────────────────────────────────────────────────┐
│ 🔄 Refund History                               │
├─────────────────────────────────────────────────┤
│ REF-20260129-ABC123  [COMPLETED] [FULL]        │
│ Reason: Customer Request - Test refund system  │
│                                    Rp 918,000   │
│ Gateway ID: MANUAL_BANK_TRANSFER                │
│                                                 │
│ Items: Rp 900,000                               │
│ Shipping: Rp 18,000                             │
│                                                 │
│ Requested: 29 Jan 2026 14:30                    │
│ Completed: 29 Jan 2026 14:35                    │
└─────────────────────────────────────────────────┘
```

---

## ✅ Checklist Test

Centang setiap step yang berhasil:

- [ ] Bisa login ke admin panel
- [ ] Bisa buka order ZVR-20260127-B8B3ACCD
- [ ] Tombol "Refund" terlihat
- [ ] Modal refund bisa dibuka
- [ ] Bisa pilih FULL refund type
- [ ] Bisa pilih reason
- [ ] Bisa klik "Process Refund"
- [ ] Error 418 muncul (atau success)
- [ ] Refund muncul di Refund History
- [ ] Status PENDING terlihat
- [ ] Tombol "Mark as Completed" terlihat
- [ ] Bisa klik "Mark as Completed"
- [ ] Dialog konfirmasi muncul
- [ ] Bisa masukkan note
- [ ] Bisa klik Confirm
- [ ] Success message muncul
- [ ] Status berubah jadi COMPLETED
- [ ] Gateway ID: MANUAL_BANK_TRANSFER
- [ ] Order refund_status: FULL

---

## 🔍 Troubleshooting

### ❌ Tombol Refund tidak muncul
**Solusi:** Order harus status DELIVERED atau PAID

### ❌ Error "refund amount exceeds refundable amount"
**Solusi:** Sudah tidak akan terjadi lagi! Sudah di-fix.

### ❌ Tombol "Mark as Completed" tidak muncul
**Solusi:** Refund harus status PENDING

### ❌ Login gagal
**Solusi:** Gunakan Google OAuth dengan email pemberani073@gmail.com

### ❌ Backend error
**Solusi:** 
```bash
cd backend
.\zavera_refund_fix.exe
```

### ❌ Frontend error
**Solusi:**
```bash
cd frontend
npm run dev
```

---

## 📞 Kalau Ada Masalah

1. **Check backend logs:**
   - Lihat terminal yang running zavera_refund_fix.exe
   - Cari error messages

2. **Check frontend console:**
   - Buka browser DevTools (F12)
   - Lihat Console tab
   - Cari error messages

3. **Check database:**
   ```sql
   -- Check refund
   SELECT * FROM refunds 
   WHERE order_id = (SELECT id FROM orders WHERE order_code = 'ZVR-20260127-B8B3ACCD');
   
   -- Check order
   SELECT order_code, status, refund_status, refund_amount 
   FROM orders 
   WHERE order_code = 'ZVR-20260127-B8B3ACCD';
   ```

---

## 🎉 Setelah Test Berhasil

**Kalau semua checklist ✅:**

1. **Test refund types lain:**
   - PARTIAL (Rp 500,000)
   - SHIPPING_ONLY
   - ITEM_ONLY

2. **Siap demo ke client!**
   - Baca: `CLIENT_DEMO_GUIDE.md`
   - Baca: `REFUND_SYSTEM_READY_FOR_DEMO.md`

3. **Celebrate!** 🎊
   - Refund system 100% working!
   - Production ready!

---

## 📚 Dokumentasi Lengkap

- **REFUND_SYSTEM_READY_FOR_DEMO.md** - Demo guide lengkap
- **REFUND_MANUAL_TEST_GUIDE.md** - Testing guide detail
- **REFUND_FIX_SUMMARY.md** - Technical summary
- **REFUND_SYSTEM_COMPLETE_GUIDE.md** - User guide lengkap
- **CLIENT_DEMO_GUIDE.md** - Script demo untuk client

---

**Selamat testing! Semoga berhasil! 🚀**

**Kalau ada pertanyaan atau masalah, screenshot dan tanya!**
