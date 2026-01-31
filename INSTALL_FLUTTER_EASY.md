# 🚀 Cara MUDAH Install Flutter (Pakai VS Code Extension)

## ✅ Cara Paling Gampang - Pakai Extension

### 1️⃣ Install Flutter Extension di VS Code

1. **Buka VS Code**
2. **Tekan `Ctrl + Shift + X`** (buka Extensions)
3. **Search:** `Flutter`
4. **Install:** Extension "Flutter" by Dart Code
5. **Restart VS Code**

### 2️⃣ Download Flutter SDK via Extension

Setelah install extension:

1. **Tekan `Ctrl + Shift + P`** (Command Palette)
2. **Ketik:** `Flutter: New Project`
3. Extension akan **otomatis detect** kalau Flutter belum install
4. Klik **"Download SDK"** atau **"Locate SDK"**
5. Extension akan **download & setup Flutter otomatis!**

### 3️⃣ Atau Download Manual (Lebih Cepat)

**Link Direct Download:**
```
https://storage.googleapis.com/flutter_infra_release/releases/stable/windows/flutter_windows_3.19.0-stable.zip
```

**Langkah:**
1. Download file zip (~1.5 GB)
2. Extract ke: `C:\src\flutter`
3. Add to PATH: `C:\src\flutter\bin`
4. Restart VS Code

### 4️⃣ Tambah Flutter ke PATH

**Cara Cepat:**
1. Tekan `Win + R`
2. Ketik: `sysdm.cpl` → Enter
3. Tab "Advanced" → "Environment Variables"
4. Di "User variables" → Pilih "Path" → "Edit"
5. Klik "New" → Ketik: `C:\src\flutter\bin`
6. OK semua
7. **Restart VS Code & Terminal**

### 5️⃣ Verifikasi

Buka terminal baru di VS Code (`Ctrl + ~`):
```bash
flutter --version
flutter doctor
```

## 🎯 Install Android SDK (Untuk Run di HP)

### Opsi 1: Via VS Code (Mudah)

1. Run: `flutter doctor`
2. Akan muncul error "Android toolchain not found"
3. VS Code akan suggest install Android SDK
4. Klik "Install" → Ikuti wizard

### Opsi 2: Install Android Studio

1. Download: https://developer.android.com/studio
2. Install dengan default settings
3. Buka Android Studio
4. Welcome screen → "More Actions" → "SDK Manager"
5. Install:
   - Android SDK Platform
   - Android SDK Build-Tools
   - Android SDK Platform-Tools
6. Close Android Studio
7. Run: `flutter doctor --android-licenses` (ketik 'y' semua)

## 📱 Connect HP & Run

### 1. Connect HP
```
1. Colok HP pakai USB
2. Allow "USB Debugging" di HP
3. Centang "Always allow"
```

### 2. Cek Device di VS Code

Di VS Code, klik **status bar kanan bawah** → Harus muncul nama HP kamu

Atau di terminal:
```bash
flutter devices
```

### 3. Open Project

```bash
# Di VS Code, buka folder:
File → Open Folder → Pilih: zavera_mobile
```

### 4. Run App

**Cara 1 - Pakai VS Code (Paling Mudah):**
1. Tekan `F5` atau `Ctrl + F5`
2. Pilih device (HP kamu)
3. Wait... App akan install & run di HP!

**Cara 2 - Pakai Terminal:**
```bash
cd zavera_mobile
flutter pub get
flutter run
```

## ⚡ Super Quick Steps

Jika sudah punya Flutter:

```bash
# 1. Buka project di VS Code
code zavera_mobile

# 2. Tekan F5 (Run)
# 3. Pilih device (HP kamu)
# 4. Done! 🎉
```

## 🐛 Troubleshooting

### "Flutter SDK not found"
- Install Flutter extension di VS Code
- Atau download manual & add to PATH
- Restart VS Code

### "No devices found"
- Cek USB Debugging aktif di HP
- Coba kabel USB lain
- Run: `adb devices`

### "Android licenses not accepted"
```bash
flutter doctor --android-licenses
# Ketik 'y' untuk semua
```

## 📦 Yang Perlu Diinstall:

1. ✅ **Flutter SDK** (~1.5 GB) - Via extension atau manual
2. ✅ **Android SDK** (~3 GB) - Via Android Studio
3. ✅ **VS Code Extensions:**
   - Flutter
   - Dart

## ⏱️ Estimasi Waktu:

- Download Flutter: 10-15 menit
- Install Android Studio: 20-30 menit
- Setup: 5-10 menit
- **Total: ~45 menit**

## 🎯 Setelah Setup:

```bash
# Run app (30-60 detik):
cd zavera_mobile
flutter run

# Atau tekan F5 di VS Code
```

---

## 💡 Tips:

1. **Pakai WiFi cepat** untuk download
2. **Restart VS Code** setelah install Flutter
3. **Restart terminal** setelah add PATH
4. **Allow firewall** jika diminta
5. **First build lama** (3-5 menit), next builds cepat (30 detik)

## 🔗 Links:

- Flutter Download: https://flutter.dev/docs/get-started/install/windows
- Android Studio: https://developer.android.com/studio
- VS Code: https://code.visualstudio.com/

---

**Setelah selesai, tinggal:**
```bash
cd zavera_mobile
flutter run
```

**Dan app ZAVERA akan muncul di HP kamu!** 🎉
