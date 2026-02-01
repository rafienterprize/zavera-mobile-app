# 📁 Logo Files - Taruh Logo Disini!

## 📂 Folder ini untuk logo ZAVERA

### File yang perlu kamu taruh:

1. **zavera_logo.png**
   - Ukuran: 1024 x 1024 px
   - Format: PNG
   - Background: Hitam atau putih (sesuai design)
   - Ini logo lengkap dengan background

2. **zavera_logo_foreground.png**
   - Ukuran: 432 x 432 px
   - Format: PNG dengan transparency
   - Background: Transparan
   - Ini logo aja tanpa background (untuk adaptive icon)

---

## 🎯 Setelah taruh logo:

1. Update `pubspec.yaml` (sudah aku setup)
2. Run command:
   ```bash
   flutter pub get
   flutter pub run flutter_launcher_icons
   ```
3. Logo akan auto-generate ke semua ukuran!

---

## 📐 Ukuran yang dibutuhkan:

- Master: 1024x1024 px (zavera_logo.png)
- Foreground: 432x432 px (zavera_logo_foreground.png)

Cek `LOGO_REQUIREMENTS.md` untuk detail lengkap!
