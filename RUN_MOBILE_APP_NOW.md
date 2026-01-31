# 📱 RUN MOBILE APP - LANGKAH TERAKHIR!

## ✅ Yang Sudah Selesai:
- ✅ Flutter SDK installed
- ✅ Android SDK installed  
- ✅ Android licenses accepted
- ✅ Mobile app code ready
- ✅ Dependencies installed

## 🎯 TINGGAL 2 LANGKAH!

### Langkah 1: Hubungkan HP ke Laptop
1. **Colok kabel USB** dari HP ke laptop
2. Di HP, akan muncul notifikasi **"Allow USB debugging"**
3. **Tap "Allow"** atau **"OK"**

### Langkah 2: Jalankan App
**Cara Termudah - Double click file ini:**
```
run-mobile-app.bat
```

**Atau manual di terminal:**
```bash
cd zavera_mobile
flutter run
```

**Build pertama: 3-5 menit ⏱️**
**Build berikutnya: 30-60 detik**

---

## 🔍 Cek HP Terdeteksi atau Tidak

**Double click:** `check-phone-connected.bat`

Atau manual:
```bash
flutter devices
```

Harus muncul nama HP kamu (contoh: "SM-G991B" atau "Redmi Note 10")

---

## ⚠️ Troubleshooting

### HP Tidak Terdeteksi?

**1. Cek Kabel USB**
- Coba kabel lain (beberapa kabel cuma bisa charge, tidak bisa data)

**2. Cek Mode USB di HP**
- Swipe down notifikasi HP
- Tap "USB options" atau "USB for..."
- Pilih **"File Transfer"** atau **"MTP"**

**3. Restart ADB**
```bash
adb kill-server
adb start-server
flutter devices
```

**4. Cabut-Colok Ulang**
- Cabut kabel USB
- Colok lagi
- Tap "Allow USB debugging" lagi

**5. Revoke & Try Again**
- Di HP: Settings → Developer Options → Revoke USB debugging authorizations
- Cabut-colok kabel lagi
- Allow lagi

---

## 🔌 Konfigurasi Backend (Opsional)

Kalau mau connect ke backend ZAVERA:

### 1. Cek IP Laptop
**Double click:** `get-laptop-ip.bat`

Atau manual:
```bash
ipconfig
```
Cari "IPv4 Address" di WiFi adapter (contoh: 192.168.1.100)

### 2. Update API URL
Edit file: `zavera_mobile/lib/services/api_service.dart`

Ganti baris ini:
```dart
static const String baseUrl = 'http://localhost:8080/api';
```

Jadi:
```dart
static const String baseUrl = 'http://192.168.1.100:8080/api';
```
*(Ganti dengan IP laptop kamu)*

### 3. Pastikan Backend Running
```bash
cd ZAVERA-FASHION-STORE
# Jalankan backend (sesuai cara biasa)
```

### 4. Test dari Browser HP
Buka browser di HP, akses:
```
http://192.168.1.100:8080/api/products
```
*(Ganti dengan IP laptop kamu)*

Kalau muncul JSON → ✅ Siap!

---

## 📱 Apa yang Terjadi Saat Run?

1. **Terminal menampilkan "Building..."** (sabar ya!)
2. **App otomatis install ke HP** (progress bar di terminal)
3. **App otomatis launch** di HP
4. **Kamu bisa lihat UI mobile app!**

---

## 🎮 Hot Reload (Saat App Running)

Setelah app running, kamu bisa edit code dan:
- Press **`r`** → Hot reload (cepat, 1-2 detik)
- Press **`R`** → Hot restart (restart app)
- Press **`q`** → Quit (stop app)

---

## 📂 File-File Helper

| File | Fungsi |
|------|--------|
| `run-mobile-app.bat` | Jalankan app di HP |
| `check-phone-connected.bat` | Cek HP terdeteksi atau tidak |
| `get-laptop-ip.bat` | Lihat IP laptop untuk config backend |
| `QUICK_START.md` | Panduan lengkap |
| `zavera_mobile/CONFIGURE_API.md` | Panduan config backend |

---

## 🚀 READY TO GO!

**Sekarang:**
1. Colok HP ke laptop
2. Allow USB debugging
3. Double click `run-mobile-app.bat`
4. Tunggu 3-5 menit
5. **DONE!** 🎉

---

## 💡 Tips

- **Pertama kali build lama** (3-5 menit) karena download dependencies
- **Build berikutnya cepat** (30-60 detik)
- **Hot reload sangat cepat** (1-2 detik) untuk edit UI
- **HP & Laptop harus di WiFi yang sama** (kalau mau connect backend)
- **Bisa run tanpa backend** (cuma lihat UI dulu)

---

**Kalau ada error, screenshot dan tanya!** 👍
