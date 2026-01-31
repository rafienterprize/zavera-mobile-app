# 🎯 SETUP GUIDE - Visual Walkthrough

## 📋 Current Status Check

Run this to verify everything is ready:
```bash
flutter doctor -v
```

You should see:
```
[√] Flutter (Channel stable, 3.38.9)
[√] Android toolchain - develop for Android devices
[√] Chrome - develop for the web
```

---

## 🔌 Step 1: Connect Your Phone

### What You Need:
- ✅ USB cable (data cable, not charge-only)
- ✅ Android phone with USB debugging enabled
- ✅ Laptop/PC

### Visual Guide:

```
   [PHONE] ----USB CABLE---- [LAPTOP]
      |                          |
      |                          |
   Notification:            Terminal:
   "Allow USB               flutter devices
    debugging?"
      |
   [TAP ALLOW]
```

### Commands to Run:

**Check if phone is detected:**
```bash
flutter devices
```

**Expected output:**
```
Found 4 connected devices:
  SM G991B (mobile) • 1234567890 • android-arm64 • Android 13 (API 33)
  Windows (desktop) • windows • windows-x64 • Microsoft Windows
  Chrome (web) • chrome • web-javascript • Google Chrome
  Edge (web) • edge • web-javascript • Microsoft Edge
```

Your phone should appear in the list!

---

## 🚀 Step 2: Run the App

### Option A: Use Batch File (Easiest)
**Double click:** `run-mobile-app.bat`

### Option B: Manual Command
```bash
cd zavera_mobile
flutter run
```

---

## ⏱️ What Happens During Build

### First Build (3-5 minutes):

```
[  +1 ms] Downloading Dart SDK...
[  +500 ms] Downloading Flutter dependencies...
[  +30 s] Resolving dependencies...
[  +60 s] Running Gradle tasks...
[  +90 s] Building APK...
[  +120 s] Installing APK on device...
[  +180 s] Launching app...
[  +200 s] ✓ Built build/app/outputs/flutter-apk/app-debug.apk
```

### Progress Indicators:
```
Building...                                    [    ]  0%
Downloading dependencies...                    [==  ] 25%
Running Gradle tasks...                        [====] 50%
Building APK...                                [======] 75%
Installing on device...                        [========] 100%
```

### When Complete:
```
✓ Built build/app/outputs/flutter-apk/app-debug.apk (18.5MB)
Syncing files to device SM G991B...
Flutter run key commands.
r Hot reload.
R Hot restart.
h List all available interactive commands.
d Detach (terminate "flutter run" but leave application running).
c Clear the screen
q Quit (terminate the application on the device).

Running with sound null safety

An Observatory debugger and profiler on SM G991B is available at: http://127.0.0.1:12345/
The Flutter DevTools debugger and profiler on SM G991B is available at: http://127.0.0.1:9100/
```

**App will automatically launch on your phone!** 📱

---

## 📱 What You'll See on Your Phone

### 1. Splash Screen (2 seconds)
```
┌─────────────────────┐
│                     │
│                     │
│      ZAVERA         │
│   Fashion Store     │
│                     │
│    [Loading...]     │
│                     │
└─────────────────────┘
```

### 2. Home Screen
```
┌─────────────────────┐
│  [ZAVERA]    [🛒][👤]│
├─────────────────────┤
│  [Hero Carousel]    │
│  ← → → →            │
├─────────────────────┤
│  Shop by Category   │
│  ┌────┐  ┌────┐    │
│  │Pria│  │Wanita│   │
│  └────┘  └────┘    │
│  ┌────┐  ┌────┐    │
│  │Anak│  │Sport│    │
│  └────┘  └────┘    │
├─────────────────────┤
│  Featured Products  │
│  ┌────┐  ┌────┐    │
│  │ 👕 │  │ 👗 │    │
│  │$50 │  │$75 │    │
│  └────┘  └────┘    │
├─────────────────────┤
│ [🏠] [📂] [🛒] [👤] │
└─────────────────────┘
```

### 3. Bottom Navigation
- 🏠 Home
- 📂 Categories
- 🛒 Cart
- 👤 Profile

---

## 🎮 Hot Reload Demo

While app is running, you can edit code and see changes instantly!

### Example:
1. App is running on phone
2. Edit `lib/screens/home_screen.dart`
3. Change text: "Featured Products" → "Produk Terbaru"
4. Press `r` in terminal
5. **Changes appear in 1-2 seconds!** ⚡

```
Terminal:
> r
Performing hot reload...
Reloaded 1 of 500 libraries in 1,234ms.
```

---

## 🔧 Troubleshooting Visual Guide

### Problem: Phone Not Detected

```
[PHONE]  ❌  [LAPTOP]
   ↓
Check:
1. USB cable → Try different cable
2. USB mode → Set to "File Transfer"
3. USB debugging → Check it's enabled
4. Restart ADB:
   adb kill-server
   adb start-server
```

### Problem: Build Failed

```
Error: Could not resolve dependencies
   ↓
Solution:
cd zavera_mobile
flutter clean
flutter pub get
flutter run
```

### Problem: App Crashes on Launch

```
App crashes immediately
   ↓
Check:
1. Android version (minimum: Android 5.0)
2. Storage space (need ~100MB free)
3. Check logs:
   flutter logs
```

---

## 🔌 Backend Connection (Optional)

### Network Diagram:

```
[PHONE]  ←WiFi→  [ROUTER]  ←WiFi→  [LAPTOP]
   |                                    |
   |                                    |
Mobile App                         Backend Server
(Flutter)                          (localhost:8080)
   |                                    |
   └────── HTTP Request ────────────────┘
           (via WiFi IP)
```

### Setup Steps:

**1. Get Laptop IP:**
```bash
ipconfig
```
Look for: `IPv4 Address: 192.168.1.100`

**2. Update API URL:**
Edit: `zavera_mobile/lib/services/api_service.dart`
```dart
static const String baseUrl = 'http://192.168.1.100:8080/api';
```

**3. Test Connection:**
Open browser on phone:
```
http://192.168.1.100:8080/api/products
```

Should see JSON data!

---

## 📊 Build Size Information

### First Build:
- **Time:** 3-5 minutes
- **APK Size:** ~18-20 MB
- **Download:** ~200 MB (dependencies)

### Subsequent Builds:
- **Time:** 30-60 seconds
- **APK Size:** ~18-20 MB
- **Download:** None (cached)

### Hot Reload:
- **Time:** 1-2 seconds
- **No rebuild:** Just updates changed code
- **Super fast!** ⚡

---

## ✅ Success Checklist

Before running, verify:

- [ ] Flutter SDK installed (`flutter --version` works)
- [ ] Android SDK installed (Android Studio)
- [ ] Android licenses accepted (`flutter doctor` shows ✓)
- [ ] Phone connected via USB
- [ ] USB debugging allowed on phone
- [ ] Phone appears in `flutter devices`
- [ ] Dependencies installed (`flutter pub get` done)

If all checked → **Ready to run!** 🚀

---

## 🎯 Quick Command Reference

| Command | Purpose |
|---------|---------|
| `flutter devices` | List connected devices |
| `flutter run` | Build and run app |
| `flutter clean` | Clean build cache |
| `flutter pub get` | Install dependencies |
| `flutter doctor` | Check setup |
| `adb devices` | Check ADB connection |
| `adb kill-server` | Restart ADB |

### While App Running:
| Key | Action |
|-----|--------|
| `r` | Hot reload (fast) |
| `R` | Hot restart (full restart) |
| `q` | Quit app |
| `h` | Show help |
| `c` | Clear screen |

---

## 🎉 Ready to Go!

Everything is set up. Now:

1. **Connect phone** → USB cable + Allow debugging
2. **Run app** → `run-mobile-app.bat` or `flutter run`
3. **Wait** → 3-5 minutes first time
4. **Enjoy!** → App launches on your phone

**Let's do this!** 🚀📱
