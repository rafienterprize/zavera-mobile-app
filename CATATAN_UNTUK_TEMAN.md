# Catatan untuk Melanjutkan Development

## Yang Sudah Dikerjakan ✅

### 1. Mobile App Setup
- ✅ File `.env` sudah dibuat di `zavera_mobile/.env`
- ✅ AndroidManifest.xml sudah ditambahkan permission INTERNET dan `usesCleartextTraffic=true`
- ✅ Home screen sudah dibersihkan (tidak ada section produk lagi)
- ✅ Category detail screen sudah bisa fetch produk dari backend berdasarkan kategori
- ✅ API service sudah lengkap dengan semua endpoint backend

### 2. Struktur Home Screen Sekarang
1. Search bar + Wishlist + Cart icons
2. Category navigation (Wanita, Pria, Sports, Anak, Luxury, Beauty)
3. Banner carousel (3 banners)
4. Jelajahi Kategori section
5. New Arrivals section (2 category cards: Wanita & Pria)
6. Trending Now / Luxury Collection section

### 3. File Penting
- `zavera_mobile/.env` - Konfigurasi API URL
- `zavera_mobile/lib/services/api_service.dart` - Semua API calls
- `zavera_mobile/lib/screens/home_screen.dart` - Home screen (clean, no products)
- `zavera_mobile/lib/screens/category_detail_screen.dart` - Category screen (with products)
- `zavera_mobile/android/app/src/main/AndroidManifest.xml` - Android permissions

---

## Yang Perlu Dilakukan Sekarang ⚠️

### MASALAH: Mobile App Tidak Bisa Connect ke Backend

**Error yang muncul:**
```
SocketException: Connection refused (OS Error: Connection refused, errno = 111)
```

**Penyebabnya:**
Mobile app mencoba connect ke `http://172.20.9.184:8080/api` tapi backend tidak bisa diakses.

---

## SOLUSI: Langkah-langkah untuk Teman

### Step 1: Pastikan Backend Running
```bash
cd ZAVERA-FASHION-STORE/backend
go run main.go
```
Tunggu sampai muncul: `Server running on :8080`

### Step 2: Cek IP Laptop Kamu
**Windows:**
```bash
ipconfig
```
Cari "IPv4 Address" di bagian WiFi adapter (contoh: `192.168.1.50`)

**Mac/Linux:**
```bash
ifconfig
```

### Step 3: Update File .env Mobile
Edit file `zavera_mobile/.env`:
```env
API_BASE_URL=http://IP_LAPTOP_KAMU:8080/api
```
Contoh: `API_BASE_URL=http://192.168.1.50:8080/api`

### Step 4: Pastikan 1 WiFi
HP dan laptop harus konek ke **WiFi yang sama**!

### Step 5: Test di Browser HP Dulu
Buka browser di HP, ketik:
```
http://IP_LAPTOP_KAMU:8080/api/products
```

**Kalau muncul JSON data produk** → Backend bisa diakses ✅
**Kalau error/timeout** → Cek firewall atau WiFi ❌

### Step 6: Jalankan Flutter App
```bash
cd zavera_mobile
flutter run -d DEVICE_ID
```

Atau kalau sudah running, hot restart dengan tekan `R` di terminal.

---

## Troubleshooting

### Kalau Masih Connection Refused:

**1. Cek Firewall Windows**
- Buka Windows Defender Firewall
- Allow port 8080 untuk Go application
- Atau matikan firewall sementara untuk testing

**2. Cek Backend Logs**
Pastikan backend tidak ada error saat startup

**3. Cek WiFi**
Pastikan HP dan laptop benar-benar 1 WiFi (bukan hotspot HP)

**4. Test dengan Postman/Browser Desktop**
Test dulu dari laptop: `http://localhost:8080/api/products`
Kalau berhasil, berarti backend OK, masalahnya di network.

---

## Alternatif: Deploy Backend ke Cloud (Recommended)

Kalau IP terus berubah-ubah, lebih baik deploy backend ke cloud:

### Railway.app (Gratis $5/bulan)
1. Push code ke GitHub
2. Connect Railway ke GitHub repo
3. Deploy backend
4. Dapat URL tetap: `https://zavera-backend.up.railway.app`
5. Update `.env` mobile: `API_BASE_URL=https://zavera-backend.up.railway.app/api`

**Keuntungan:**
- Tidak perlu IP lagi
- Bisa diakses dari mana saja
- Tidak perlu 1 WiFi
- Backend selalu online

---

## File yang Sudah Diubah (Commit Terakhir)

```
zavera_mobile/lib/screens/home_screen.dart - Dibersihkan, no products section
zavera_mobile/lib/screens/category_detail_screen.dart - Added API fetch
zavera_mobile/android/app/src/main/AndroidManifest.xml - Added permissions
zavera_mobile/.env - Updated API URL
```

---

## Kontak

Kalau ada masalah, hubungi Raffi atau tanya di grup.

Good luck! 🚀
