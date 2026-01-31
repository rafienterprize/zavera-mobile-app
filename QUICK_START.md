# 🚀 QUICK START - Run Mobile App on Your Phone

## Current Status
✅ Flutter SDK installed and working
✅ Android SDK installed with all licenses accepted
✅ Mobile app code ready
⏳ **NEXT: Connect your phone and run!**

---

## 🎯 FASTEST WAY - 3 Steps Only!

### Step 1: Connect Your Phone
1. **Plug USB cable** from phone to computer
2. On your phone, tap **"Allow USB debugging"** when notification appears
3. Run: `check-phone-connected.bat` to verify

### Step 2: Configure API (Optional - for backend connection)
1. Run: `get-laptop-ip.bat` to get your laptop IP
2. Edit: `zavera_mobile/lib/services/api_service.dart`
3. Change: `http://localhost:8080/api` → `http://YOUR_IP:8080/api`

**Skip this if you just want to see the UI first!**

### Step 3: Run the App
**Double-click:** `run-mobile-app.bat`

OR run manually:
```bash
cd zavera_mobile
flutter run
```

**First build: 3-5 minutes ⏱️**
**Next builds: 30-60 seconds**

---

## 📱 What You'll See

1. Terminal shows "Building..." (be patient!)
2. App installs on your phone automatically
3. App launches automatically
4. You can browse products, add to cart, etc.

---

## 🔧 Detailed Step-by-Step

### 1. Connect Your Phone via USB
1. Take your USB cable
2. Connect your phone to your computer
3. On your phone, you should see a notification asking to "Allow USB debugging"
4. **Tap "Allow" or "OK"** on your phone

### 2. Verify Phone is Connected
Run this command:
```bash
flutter devices
```

You should see your phone listed (e.g., "SM-G991B" or similar device name)

### 3. Run the App
Once your phone appears, run:
```bash
cd zavera_mobile
flutter run
```

**First build takes 3-5 minutes** ⏱️
Subsequent builds: 30-60 seconds

## Troubleshooting

### Phone Not Detected?
1. **Check USB cable** - Try a different cable (some cables are charge-only)
2. **Check USB mode** - On your phone, swipe down notifications and tap USB options, select "File Transfer" or "MTP"
3. **Restart ADB**:
   ```bash
   adb kill-server
   adb start-server
   flutter devices
   ```

### Still Not Working?
1. Unplug and replug the USB cable
2. On your phone: Settings → Developer Options → Revoke USB debugging authorizations → Try again
3. Try a different USB port on your computer

## What Happens When You Run?
1. Flutter compiles the app (first time: 3-5 min)
2. App installs on your phone automatically
3. App launches automatically
4. You can see logs in the terminal
5. Hot reload works - press `r` to reload, `R` to restart

## Quick Commands
- `flutter devices` - List connected devices
- `flutter run` - Build and run app
- `r` - Hot reload (while app is running)
- `R` - Hot restart (while app is running)
- `q` - Quit and stop app

## API Configuration
Before running, you may need to configure the backend API URL in:
`zavera_mobile/lib/services/api_service.dart`

See `CONFIGURE_API.md` for details.
