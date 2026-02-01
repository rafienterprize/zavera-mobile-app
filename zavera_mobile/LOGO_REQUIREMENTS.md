# 🎨 Logo Requirements untuk ZAVERA App

## ✅ Nama App Sudah Diubah
App name sekarang: **ZAVERA** (bukan "zavera_mobile")

---

## 📐 Ukuran Logo yang Dibutuhkan

### Android App Icon (Launcher Icon)

Kamu perlu buat logo dalam **5 ukuran berbeda**:

| Density | Ukuran (px) | Folder | Keterangan |
|---------|-------------|--------|------------|
| **mdpi** | 48 x 48 | `android/app/src/main/res/mipmap-mdpi/` | Low density |
| **hdpi** | 72 x 72 | `android/app/src/main/res/mipmap-hdpi/` | Medium density |
| **xhdpi** | 96 x 96 | `android/app/src/main/res/mipmap-xhdpi/` | High density |
| **xxhdpi** | 144 x 144 | `android/app/src/main/res/mipmap-xxhdpi/` | Extra high density |
| **xxxhdpi** | 192 x 192 | `android/app/src/main/res/mipmap-xxxhdpi/` | Extra extra high density |

**Nama file:** `ic_launcher.png` (untuk semua ukuran)

---

## 🎯 Rekomendasi Ukuran untuk Design

### 1. Master Logo (untuk design)
**Ukuran:** **1024 x 1024 px**
- Format: PNG dengan background transparan
- Ini ukuran master, nanti di-resize ke ukuran lain

### 2. Adaptive Icon (Android 8.0+)
**Ukuran:** **432 x 432 px** (dengan safe zone 108px dari tepi)
- Foreground layer: Logo utama (transparan)
- Background layer: Warna solid atau pattern

---

## 🛠️ Cara Mudah: Pakai Tool Auto-Generate

### Option 1: flutter_launcher_icons (RECOMMENDED)

1. **Install package:**
```yaml
# pubspec.yaml
dev_dependencies:
  flutter_launcher_icons: ^0.13.1

flutter_launcher_icons:
  android: true
  ios: false
  image_path: "assets/images/logo.png"  # Logo 1024x1024
  adaptive_icon_background: "#000000"   # Warna background
  adaptive_icon_foreground: "assets/images/logo_foreground.png"
```

2. **Siapkan logo:**
- `assets/images/logo.png` → 1024x1024 px (logo lengkap)
- `assets/images/logo_foreground.png` → 432x432 px (logo untuk adaptive icon)

3. **Generate:**
```bash
flutter pub get
flutter pub run flutter_launcher_icons
```

**DONE!** Semua ukuran auto-generated.

---

### Option 2: Manual (Pakai Photoshop/Figma)

1. **Buat logo 1024x1024 px**
2. **Export ke 5 ukuran:**
   - 48x48 → save ke `mipmap-mdpi/ic_launcher.png`
   - 72x72 → save ke `mipmap-hdpi/ic_launcher.png`
   - 96x96 → save ke `mipmap-xhdpi/ic_launcher.png`
   - 144x144 → save ke `mipmap-xxhdpi/ic_launcher.png`
   - 192x192 → save ke `mipmap-xxxhdpi/ic_launcher.png`

---

## 🎨 Design Guidelines

### Logo Style
- **Simple & Clean** - Harus jelas di ukuran kecil (48x48)
- **High Contrast** - Mudah dibedakan dari background
- **No Text** - Atau text minimal (logo mark aja)
- **Square Format** - 1:1 ratio

### Colors
- **Background:** Hitam (#000000) atau putih (#FFFFFF)
- **Logo:** Kontras dengan background
- **Avoid:** Gradients kompleks (susah di ukuran kecil)

### Safe Zone
- **Padding:** Minimal 10% dari tepi
- Contoh: Logo 1024x1024 → content area 820x820 (padding 102px)

---

## 📁 Struktur Folder

```
android/app/src/main/res/
├── mipmap-mdpi/
│   └── ic_launcher.png (48x48)
├── mipmap-hdpi/
│   └── ic_launcher.png (72x72)
├── mipmap-xhdpi/
│   └── ic_launcher.png (96x96)
├── mipmap-xxhdpi/
│   └── ic_launcher.png (144x144)
└── mipmap-xxxhdpi/
    └── ic_launcher.png (192x192)
```

---

## 🚀 Quick Start (Pakai flutter_launcher_icons)

### Step 1: Update pubspec.yaml
```yaml
dev_dependencies:
  flutter_launcher_icons: ^0.13.1

flutter_launcher_icons:
  android: true
  ios: false
  image_path: "assets/images/zavera_logo.png"
  adaptive_icon_background: "#000000"
  adaptive_icon_foreground: "assets/images/zavera_logo_foreground.png"
```

### Step 2: Buat Logo
1. **zavera_logo.png** → 1024x1024 px (logo lengkap dengan background)
2. **zavera_logo_foreground.png** → 432x432 px (logo aja, transparan)

Save ke folder: `zavera_mobile/assets/images/`

### Step 3: Generate
```bash
cd zavera_mobile
flutter pub get
flutter pub run flutter_launcher_icons
```

### Step 4: Test
```bash
flutter run
```

Cek icon di home screen HP!

---

## 📱 Preview di HP

Setelah install, logo akan muncul di:
- Home screen
- App drawer
- Recent apps
- Notification (kalau ada)

---

## ✅ Checklist

- [ ] Buat logo master 1024x1024 px
- [ ] Buat logo foreground 432x432 px (untuk adaptive icon)
- [ ] Save ke `assets/images/`
- [ ] Update `pubspec.yaml` dengan config `flutter_launcher_icons`
- [ ] Run `flutter pub run flutter_launcher_icons`
- [ ] Test di HP
- [ ] Verify icon muncul di home screen

---

## 💡 Tips

1. **Test di HP real** - Icon bisa terlihat beda di emulator vs HP
2. **Check di dark mode** - Pastikan logo jelas di dark & light theme
3. **Simple is better** - Logo kompleks susah dilihat di ukuran kecil
4. **Use PNG** - Dengan transparency untuk adaptive icon

---

## 🎊 Contoh Logo ZAVERA

Untuk brand fashion seperti ZAVERA, rekomendasi:

**Style 1: Minimalist Text**
- Text "ZAVERA" dengan font elegant (Playfair Display)
- Background hitam
- Text putih/gold

**Style 2: Monogram**
- Letter "Z" stylized
- Simple & iconic
- Easy to recognize

**Style 3: Icon + Text**
- Small icon (fashion-related)
- Text "ZAVERA" di bawah
- Balanced composition

---

**Need help?** Kalau udah buat logo, kasih tau aku nanti aku bantu setup!
