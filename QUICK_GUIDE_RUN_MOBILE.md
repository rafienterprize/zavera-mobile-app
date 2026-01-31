# 🚀 Quick Guide: Run ZAVERA Mobile di HP

## ✅ Yang Sudah Kamu Lakukan:
- USB Debugging di HP sudah aktif ✅
- HP sudah siap untuk development ✅

## 📋 Langkah Selanjutnya:

### 1️⃣ Install Flutter (Sekali Aja)

**Download & Install:**
```
1. Download Flutter SDK dari: https://flutter.dev/docs/get-started/install/windows
2. Extract ke: C:\src\flutter
3. Tambahkan ke PATH: C:\src\flutter\bin
4. Restart PowerShell/Terminal
```

**Atau lihat panduan lengkap di:** `INSTALL_FLUTTER.md`

### 2️⃣ Install Android SDK (Sekali Aja)

**Cara Mudah - Install Android Studio:**
```
1. Download: https://developer.android.com/studio
2. Install dengan default settings
3. Buka Android Studio → SDK Manager
4. Install Android SDK Platform & Build Tools
5. Run: flutter doctor --android-licenses (ketik 'y' semua)
```

### 3️⃣ Connect HP ke Laptop

```
1. Colok HP pakai kabel USB
2. Di HP, allow "USB Debugging" 
3. Centang "Always allow from this computer"
```

### 4️⃣ Cek HP Terdeteksi

**Double-click:** `check-flutter-ready.bat`

Atau manual:
```bash
flutter devices
```

Harus muncul HP kamu di list!

### 5️⃣ Run App di HP

**Cara 1 - Pakai Script (Mudah):**
```
Double-click: run-mobile-app.bat
```

**Cara 2 - Manual:**
```bash
cd zavera_mobile
flutter pub get
flutter run
```

### 6️⃣ Wait & Enjoy! 🎉

```
- First build: 3-5 menit (download dependencies)
- Next builds: 30-60 detik
- App akan otomatis install & run di HP kamu
```

## 🎮 Controls Saat App Running:

```
r  - Hot reload (reload code tanpa restart)
R  - Hot restart (restart app)
q  - Quit/Stop app
```

## ⚡ Estimasi Waktu:

| Step | Waktu |
|------|-------|
| Install Flutter | 15-20 menit |
| Install Android Studio | 20-30 menit |
| Setup & Config | 10 menit |
| First Run App | 5 menit |
| **TOTAL** | **~1 jam** |

Setelah setup pertama, next time cuma:
```bash
flutter run  # 30-60 detik
```

## 🐛 Troubleshooting Cepat:

### HP Tidak Terdeteksi?
```bash
# Cek status
adb devices

# Jika "unauthorized", allow di HP
# Jika tidak muncul, coba:
adb kill-server
adb start-server
adb devices
```

### Error "Android licenses not accepted"?
```bash
flutter doctor --android-licenses
# Ketik 'y' untuk semua
```

### Error Build?
```bash
cd zavera_mobile
flutter clean
flutter pub get
flutter run
```

## 📞 Butuh Bantuan?

Kirim screenshot dari:
```bash
flutter doctor -v
flutter devices
```

---

## 🎯 TL;DR (Super Quick)

Jika Flutter sudah install:
```bash
# 1. Connect HP (USB Debugging ON)
# 2. Run:
cd zavera_mobile
flutter run
```

Jika Flutter belum install:
```
1. Download Flutter: https://flutter.dev
2. Extract ke C:\src\flutter
3. Add to PATH: C:\src\flutter\bin
4. Restart terminal
5. flutter doctor
6. flutter run
```

**That's it!** 🚀
