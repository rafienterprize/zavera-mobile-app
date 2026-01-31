# 🚀 Install Android Studio - Quick Guide

## Download & Install (20-30 menit)

### 1. Download Android Studio
**Link:** https://developer.android.com/studio

**Size:** ~1 GB download, ~3 GB setelah install

### 2. Install
1. Run installer
2. Pilih "Standard" installation
3. Accept semua default settings
4. Wait... (10-15 menit)

### 3. First Launch Setup
1. Buka Android Studio
2. Welcome screen akan muncul
3. Klik "More Actions" → "SDK Manager"
4. Pastikan terinstall:
   - ✅ Android SDK Platform (API 34 atau terbaru)
   - ✅ Android SDK Build-Tools
   - ✅ Android SDK Platform-Tools
5. Klik "Apply" jika ada yang belum terinstall

### 4. Accept Licenses
Buka terminal/PowerShell:
```bash
flutter doctor --android-licenses
```
Ketik **'y'** untuk semua (tekan Enter berkali-kali)

### 5. Verify
```bash
flutter doctor
```

Harus muncul:
```
[✓] Flutter
[✓] Android toolchain
[✓] Chrome
```

## 📱 Run di HP

### 1. Connect HP
- Colok USB
- Allow "USB Debugging" di HP
- Centang "Always allow"

### 2. Check Device
```bash
flutter devices
```

Harus muncul nama HP kamu!

### 3. Run App
```bash
cd zavera_mobile
flutter run
```

**First build:** 3-5 menit
**Next builds:** 30-60 detik

App akan otomatis install & run di HP! 🎉

## ⚡ Super Quick Steps

```bash
# 1. Install Android Studio (download dari link di atas)
# 2. Accept licenses:
flutter doctor --android-licenses

# 3. Connect HP & run:
cd zavera_mobile
flutter run
```

## 🎯 Estimasi Total Waktu

- Download: 10-15 menit
- Install: 10-15 menit
- Setup: 5 menit
- First run: 5 menit
- **Total: ~40 menit**

Setelah itu, next time cuma:
```bash
flutter run  # 30-60 detik
```

---

**Download sekarang:** https://developer.android.com/studio
