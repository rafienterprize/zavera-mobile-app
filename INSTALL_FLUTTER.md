# 🚀 Cara Install Flutter di Windows & Run di HP

## 📥 STEP 1: Install Flutter

### Download Flutter SDK

1. **Download Flutter:**
   - Buka: https://docs.flutter.dev/get-started/install/windows
   - Klik "Download Flutter SDK" (sekitar 1.5 GB)
   - Atau direct link: https://storage.googleapis.com/flutter_infra_release/releases/stable/windows/flutter_windows_3.x.x-stable.zip

2. **Extract Flutter:**
   ```
   Extract file zip ke: C:\src\flutter
   (Jangan extract ke folder yang butuh admin permission seperti C:\Program Files)
   ```

3. **Tambahkan Flutter ke PATH:**
   
   **Cara 1 - Via GUI:**
   - Tekan `Win + R`, ketik `sysdm.cpl`, Enter
   - Tab "Advanced" → "Environment Variables"
   - Di "User variables", pilih "Path" → "Edit"
   - Klik "New" → Tambahkan: `C:\src\flutter\bin`
   - Klik OK semua

   **Cara 2 - Via PowerShell (Admin):**
   ```powershell
   [System.Environment]::SetEnvironmentVariable('Path', $env:Path + ';C:\src\flutter\bin', 'User')
   ```

4. **Restart Terminal/PowerShell** (PENTING!)

5. **Verifikasi Instalasi:**
   ```bash
   flutter --version
   flutter doctor
   ```

## 🔧 STEP 2: Install Android SDK (untuk run di HP)

Flutter butuh Android SDK untuk build ke Android device.

### Opsi A: Install Android Studio (Recommended)

1. **Download Android Studio:**
   - https://developer.android.com/studio
   - Install dengan default settings

2. **Install Android SDK:**
   - Buka Android Studio
   - Welcome screen → "More Actions" → "SDK Manager"
   - Install:
     - ✅ Android SDK Platform (API 33 atau terbaru)
     - ✅ Android SDK Build-Tools
     - ✅ Android SDK Platform-Tools
     - ✅ Android SDK Command-line Tools

3. **Accept Licenses:**
   ```bash
   flutter doctor --android-licenses
   # Ketik 'y' untuk accept semua
   ```

### Opsi B: Install Command Line Tools Only (Lebih Ringan)

1. **Download Command Line Tools:**
   - https://developer.android.com/studio#command-tools
   - Extract ke: `C:\Android\cmdline-tools\latest`

2. **Set Environment Variables:**
   ```
   ANDROID_HOME = C:\Android
   Path += C:\Android\cmdline-tools\latest\bin
   Path += C:\Android\platform-tools
   ```

3. **Install SDK Components:**
   ```bash
   sdkmanager "platform-tools" "platforms;android-33" "build-tools;33.0.0"
   flutter doctor --android-licenses
   ```

## 📱 STEP 3: Connect HP & Run App

### 1. Enable USB Debugging (Sudah Done ✅)

Kamu sudah aktifkan, bagus!

### 2. Connect HP ke Laptop

1. Colok HP ke laptop pakai kabel USB
2. Di HP, akan muncul notifikasi "USB Debugging"
3. Tap "Allow" atau "OK"
4. Centang "Always allow from this computer" (opsional)

### 3. Cek Device Terdeteksi

```bash
# Cek apakah HP terdeteksi
flutter devices

# Atau pakai adb
adb devices
```

Output yang diharapkan:
```
List of devices attached
ABC123XYZ    device
```

Jika muncul "unauthorized", cek HP kamu dan allow USB debugging.

### 4. Run Flutter App

```bash
# Masuk ke folder project
cd zavera_mobile

# Install dependencies (jika belum)
flutter pub get

# Run app di HP
flutter run
```

## 🎯 TROUBLESHOOTING

### Problem 1: "flutter: command not found"

**Solusi:**
- Pastikan Flutter sudah di PATH
- Restart terminal/PowerShell
- Coba buka terminal baru

### Problem 2: "No devices found"

**Solusi:**
- Cek kabel USB (coba kabel lain)
- Pastikan USB Debugging aktif di HP
- Install driver HP (biasanya auto-install)
- Coba mode "File Transfer" atau "PTP" di HP
- Restart adb: `adb kill-server` lalu `adb start-server`

### Problem 3: "Android licenses not accepted"

**Solusi:**
```bash
flutter doctor --android-licenses
# Ketik 'y' untuk semua
```

### Problem 4: "Gradle build failed"

**Solusi:**
```bash
cd zavera_mobile/android
./gradlew clean
cd ..
flutter clean
flutter pub get
flutter run
```

### Problem 5: HP tidak terdeteksi di Windows

**Solusi:**
1. Install USB Driver HP:
   - Google: "[Merk HP] USB Driver"
   - Atau pakai Universal ADB Driver
2. Coba port USB lain
3. Disable/Enable USB Debugging
4. Restart HP dan Laptop

## ⚡ QUICK START (Setelah Install)

```bash
# 1. Cek Flutter ready
flutter doctor

# 2. Connect HP (USB Debugging ON)

# 3. Cek device
flutter devices

# 4. Masuk folder project
cd zavera_mobile

# 5. Install dependencies
flutter pub get

# 6. Run app
flutter run
```

## 📊 Flutter Doctor Checklist

Setelah install, `flutter doctor` harus show:

```
[✓] Flutter (Channel stable, 3.x.x)
[✓] Android toolchain - develop for Android devices
[✓] Chrome - develop for the web
[✓] Android Studio (version 2023.x)
[✓] VS Code (version 1.x)
[✓] Connected device (1 available)
```

Minimal yang harus ✓:
- Flutter
- Android toolchain
- Connected device

## 🔥 ALTERNATIVE: Pakai Emulator (Jika HP Bermasalah)

### Setup Android Emulator:

1. **Buka Android Studio**
2. **Tools → Device Manager**
3. **Create Device:**
   - Phone: Pixel 4 atau 5
   - System Image: Android 11 atau 12 (download jika perlu)
   - Finish

4. **Start Emulator:**
   - Klik ▶️ di Device Manager

5. **Run Flutter:**
   ```bash
   flutter run
   ```

## 📞 Need Help?

Jika ada error, kirim output dari:
```bash
flutter doctor -v
flutter devices
adb devices
```

---

**Estimasi Waktu Install:**
- Download Flutter: 10-20 menit (tergantung internet)
- Install Android Studio: 15-30 menit
- Setup & Config: 10-15 menit
- **Total: ~1 jam**

**Setelah selesai, kamu bisa:**
```bash
cd zavera_mobile
flutter run
```

Dan app ZAVERA akan running di HP kamu! 🎉
