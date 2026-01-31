# 📦 Cara Mendapatkan Resi Otomatis dari Biteship

## ✅ SUDAH SELESAI - SIAP DIGUNAKAN!

---

## 🎯 Untuk Admin: Di Mana Dapat Resi?

### Cara Lama (Manual) ❌
1. Login ke dashboard JNE/SiCepat/J&T
2. Create order manual
3. Copy nomor resi
4. Paste ke ZAVERA
5. Klik "Kirim Pesanan"

### Cara Baru (Otomatis) ✅
1. Klik **"Kirim Pesanan"**
2. **KOSONGKAN input resi** (jangan isi apa-apa)
3. Klik **"Confirm"**
4. ✨ **RESI OTOMATIS MUNCUL!**

---

## 📸 Screenshot Guide

### Step 1: Buka Order Detail
```
Admin Dashboard → Orders → Pilih order PACKING → Klik order
```

### Step 2: Klik "Kirim Pesanan"
```
[Button: Kirim Pesanan] ← Klik ini
```

### Step 3: Modal Muncul
```
┌─────────────────────────────────────────┐
│ Kirim Pesanan                           │
├─────────────────────────────────────────┤
│ ✨ Auto-Generate Resi dari Biteship     │
│                                         │
│ Biarkan kosong untuk auto-generate     │
│ nomor resi dari Biteship.               │
│                                         │
│ Nomor resi (opsional):                  │
│ ┌─────────────────────────────────────┐ │
│ │ [KOSONGKAN INI]                     │ │ ← JANGAN ISI!
│ └─────────────────────────────────────┘ │
│                                         │
│ 💡 Tip: Kosongkan untuk auto-generate  │
│                                         │
│ [Batal]  [Confirm] ← Klik ini          │
└─────────────────────────────────────────┘
```

### Step 4: Resi Otomatis Muncul!
```
✅ Resi otomatis: JNE1234567890

Order Detail:
┌─────────────────────────────────────────┐
│ Shipment                                │
│ Status: SHIPPED                         │
│                                         │
│ Nomor Resi:                             │
│ ┌─────────────────────────────────────┐ │
│ │ JNE1234567890                       │ │ ← RESI MUNCUL!
│ └─────────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

---

## 🔍 Di Mana Resi Muncul?

### 1. Toast Notification (Pop-up)
Setelah klik "Confirm", akan muncul notifikasi:
```
✅ Resi otomatis: JNE1234567890
```

### 2. Order Detail Page
Scroll ke bagian **Shipment Status Card**:
```
┌─────────────────────────────────────────┐
│ 🚚 Shipment                             │
│ Status: SHIPPED                         │
│                                         │
│ Nomor Resi:                             │
│ JNE1234567890                           │ ← DI SINI!
└─────────────────────────────────────────┘
```

### 3. Email ke Customer
Customer otomatis dapat email:
```
Subject: Pesanan Anda Telah Dikirim - ORD-123

Halo Customer,

Pesanan Anda telah dikirim!

Nomor Resi: JNE1234567890  ← DI SINI!
Kurir: JNE REG

Track pesanan Anda di: [Link Tracking]
```

### 4. Database
Resi tersimpan di 2 tempat:
```sql
-- Table: orders
SELECT resi FROM orders WHERE order_code = 'ORD-123';
-- Result: JNE1234567890

-- Table: shipments
SELECT tracking_number FROM shipments WHERE order_id = 123;
-- Result: JNE1234567890
```

---

## 🎬 Video Tutorial (Step-by-Step)

### Scenario: Order Sudah PAID, Siap Dikirim

**1. Pack Order (PAID → PACKING)**
```
Admin Dashboard
  → Orders
  → Pilih order dengan status PAID
  → Klik "Proses Pesanan"
  → Status berubah PACKING ✅
```

**2. Ship Order (PACKING → SHIPPED)**
```
Order Detail Page
  → Klik "Kirim Pesanan"
  → Modal muncul
  → KOSONGKAN input resi (jangan isi apa-apa!)
  → Klik "Confirm"
  → Loading... (2-3 detik)
  → ✅ Toast: "Resi otomatis: JNE1234567890"
  → Status berubah SHIPPED ✅
  → Resi muncul di Shipment card ✅
```

**3. Verifikasi Resi**
```
Order Detail Page
  → Scroll ke "Shipment" card
  → Lihat "Nomor Resi: JNE1234567890" ✅
  → Copy resi untuk tracking
```

---

## 🧪 Test Sekarang!

### Test Case 1: Auto-Generate (Recommended)
1. Buka order dengan status PACKING
2. Klik "Kirim Pesanan"
3. **KOSONGKAN input resi**
4. Klik "Confirm"
5. ✅ Resi otomatis: JNE1234567890

### Test Case 2: Manual Input (Backward Compatible)
1. Buka order dengan status PACKING
2. Klik "Kirim Pesanan"
3. **ISI resi manual:** TEST123456789
4. Klik "Confirm"
5. ✅ Resi manual: TEST123456789

---

## ❓ FAQ

### Q: Resi tidak muncul, kenapa?
**A:** Cek backend log untuk error. System akan fallback ke manual resi jika Biteship gagal.

### Q: Resi yang muncul format aneh (JNE-123-1738123456)?
**A:** Itu manual resi (fallback). Biteship mungkin gagal. Cek:
- Token Biteship valid?
- Draft order ada di database?
- Log backend ada error?

### Q: Bisa input resi manual?
**A:** Bisa! Isi input resi, lalu klik "Confirm". System akan gunakan resi yang Anda input.

### Q: Resi bisa di-track?
**A:** Ya! Resi dari Biteship bisa di-track via API. Resi manual tidak bisa auto-track.

### Q: Customer dapat notifikasi?
**A:** Ya! Customer otomatis dapat email dengan nomor resi setelah admin kirim pesanan.

---

## 🎉 Keuntungan Auto-Generate Resi

### Untuk Admin:
- ✅ **Hemat waktu** - Tidak perlu login ke dashboard kurir
- ✅ **Mengurangi error** - Tidak salah ketik resi
- ✅ **Lebih cepat** - 1 klik langsung dapat resi
- ✅ **Tidak ribet** - Tidak perlu copy-paste

### Untuk Customer:
- ✅ **Dapat resi lebih cepat** - Otomatis setelah admin kirim
- ✅ **Email notifikasi** - Langsung dapat email dengan resi
- ✅ **Tracking real-time** - Status update otomatis

### Untuk System:
- ✅ **Reliable** - Fallback ke manual jika Biteship gagal
- ✅ **Trackable** - Resi bisa di-track via API
- ✅ **Scalable** - Support semua kurir (JNE, SiCepat, J&T, dll)

---

## 🚀 Mulai Sekarang!

1. **Jalankan backend:** `.\zavera_COMPLETE.exe`
2. **Buka admin dashboard:** http://localhost:3000/admin
3. **Pilih order PACKING**
4. **Klik "Kirim Pesanan"**
5. **KOSONGKAN input resi**
6. **Klik "Confirm"**
7. **✅ RESI OTOMATIS MUNCUL!**

---

## 📞 Support

Jika ada masalah:
1. Cek backend log untuk error
2. Cek database `shipments` table
3. Verifikasi Biteship token valid
4. Lihat dokumentasi: `BITESHIP_AUTO_RESI_IMPLEMENTATION.md`

---

**Selamat! Anda tidak perlu lagi input resi manual! 🎉**
