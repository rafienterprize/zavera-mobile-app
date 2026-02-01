# Setup Guide untuk Kolaborasi

## Untuk Backend Developer (Temen yang handle database)

### 1. Setup Backend
```bash
cd ZAVERA-FASHION-STORE/backend
```

### 2. Update `.env` file
Edit `ZAVERA-FASHION-STORE/backend/.env`:
```env
DB_PASSWORD=your_actual_postgres_password
MIDTRANS_SERVER_KEY=your_midtrans_key
BITESHIP_API_KEY=your_biteship_key
```

### 3. Run Backend
```bash
go run main.go
```

Backend akan jalan di `http://localhost:8080`

### 4. Share IP Address ke Mobile Developer
Jalankan:
```bash
get-laptop-ip.bat
```

Kasih tau IP address ke temen (contoh: `192.168.1.100`)

---

## Untuk Mobile Developer (Yang handle Flutter APK)

### 1. Update API URL
Edit `zavera_mobile/lib/services/api_service.dart`:
```dart
static const String baseUrl = 'http://IP_TEMEN_KAMU:8080/api';
```

Ganti `IP_TEMEN_KAMU` dengan IP yang dikasih backend developer.

### 2. Run Flutter App
```bash
cd zavera_mobile
flutter run
```

Pilih device (HP Android atau emulator).

---

## Git Workflow

### Push Changes
```bash
git add .
git commit -m "Update: [deskripsi perubahan]"
git push origin main
```

### Pull Changes
```bash
git pull origin main
```

### JANGAN COMMIT FILE INI:
- `ZAVERA-FASHION-STORE/backend/.env` (sudah di .gitignore)
- File dengan password atau API keys

---

## Testing

### Test Backend API
Buka browser: `http://localhost:8080/api/products`

Harus muncul JSON produk-produk.

### Test Mobile App
1. Pastikan HP dan laptop di WiFi yang sama
2. Backend harus jalan
3. Run `flutter run` di folder `zavera_mobile`

---

## Troubleshooting

### Mobile app tidak bisa connect ke backend
- Cek IP address benar
- Cek backend jalan (`http://IP:8080/api/products`)
- Cek firewall Windows tidak block port 8080
- Pastikan HP dan laptop di network yang sama

### Backend error "relation does not exist"
Run migrations:
```bash
cd ZAVERA-FASHION-STORE
setup_all_migrations.bat
```

---

## Contact
- Backend: [Nama temen kamu]
- Mobile: [Nama kamu]
