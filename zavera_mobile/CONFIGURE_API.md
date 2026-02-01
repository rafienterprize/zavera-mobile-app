# 🔌 Konfigurasi API Backend

Sebelum run mobile app, pastikan API URL sudah benar!

## 📍 Cek IP Laptop Kamu

### Windows:
```bash
ipconfig
```
Cari "IPv4 Address" di WiFi adapter, contoh: `192.168.1.100`

### Atau pakai PowerShell:
```powershell
(Get-NetIPAddress -AddressFamily IPv4 | Where-Object {$_.InterfaceAlias -like "*Wi-Fi*"}).IPAddress
```

## ⚙️ Update API URL

Edit file: `lib/services/api_service.dart`

```dart
class ApiService {
  // GANTI IP INI dengan IP laptop kamu!
  static const String baseUrl = 'http://192.168.1.100:8080/api';
  
  // Contoh lain:
  // static const String baseUrl = 'http://192.168.0.105:8080/api';
  // static const String baseUrl = 'http://10.0.0.50:8080/api';
```

## 🎯 Pilih URL Sesuai Environment:

| Environment | Base URL | Kapan Dipakai |
|------------|----------|---------------|
| **Real Device (WiFi)** | `http://192.168.x.x:8080/api` | HP & Laptop di WiFi yang sama |
| **Android Emulator** | `http://10.0.2.2:8080/api` | Pakai emulator Android Studio |
| **iOS Simulator** | `http://localhost:8080/api` | Pakai simulator iOS (Mac only) |

## ✅ Checklist Sebelum Run:

- [ ] Backend ZAVERA sudah running di `http://localhost:8080`
- [ ] HP dan Laptop di WiFi yang sama
- [ ] IP laptop sudah dicek (ipconfig)
- [ ] File `api_service.dart` sudah diupdate dengan IP yang benar
- [ ] Firewall tidak block port 8080

## 🧪 Test API Connection

Setelah update IP, test dari HP:

1. Buka browser di HP
2. Akses: `http://192.168.x.x:8080/api/products` (ganti dengan IP kamu)
3. Harus muncul JSON data products

Jika tidak bisa akses, cek:
- Backend running?
- Firewall block port 8080?
- HP & Laptop di WiFi yang sama?

## 🔥 Allow Firewall (Jika Perlu)

Windows Firewall mungkin block port 8080:

```powershell
# Run as Administrator
New-NetFirewallRule -DisplayName "ZAVERA Backend" -Direction Inbound -LocalPort 8080 -Protocol TCP -Action Allow
```

Atau manual:
1. Windows Security → Firewall & network protection
2. Advanced settings → Inbound Rules → New Rule
3. Port → TCP → 8080 → Allow

## 📝 Contoh Lengkap:

**Laptop IP:** `192.168.1.100`

**File:** `lib/services/api_service.dart`
```dart
class ApiService {
  static const String baseUrl = 'http://192.168.1.100:8080/api';
  // ...
}
```

**Test di browser HP:**
```
http://192.168.1.100:8080/api/products
```

Jika muncul JSON → ✅ Siap run app!

---

**Setelah konfigurasi, run:**
```bash
cd zavera_mobile
flutter run
```
