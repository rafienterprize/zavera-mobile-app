# 🚀 Panduan Setup ZAVERA Fashion Store

## Langkah-langkah Instalasi

### 1️⃣ Setup Database PostgreSQL

**Install PostgreSQL** (jika belum terinstall)

- Download dari: https://www.postgresql.org/download/windows/

**Buat database:**

```powershell
# Buka PowerShell sebagai Administrator
psql -U postgres

# Di dalam psql, jalankan:
CREATE DATABASE zavera_db;
\q
```

**Import data:**

```powershell
# Jalankan script SQL untuk membuat tabel dan data sample
psql -U postgres -d zavera_db -f "database/init.sql"
```

### 2️⃣ Setup Backend (Golang)

**Masuk ke folder backend:**

```powershell
cd backend
```

**Download dependencies:**

```powershell
go mod download
go mod tidy
```

**Buat file .env** (copy dari .env.example):

```powershell
Copy-Item .env.example .env
```

**Edit file .env** dengan text editor (notepad/VSCode):

```env
PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_NAME=zavera_db
DB_USER=postgres
DB_PASSWORD=password_postgres_anda

# Dapatkan dari https://dashboard.midtrans.com/
MIDTRANS_SERVER_KEY=your_server_key_here
MIDTRANS_ENV=sandbox
```

**Jalankan backend:**

```powershell
go run main.go
```

Backend akan berjalan di: **http://localhost:8080**

### 3️⃣ Setup Frontend (Next.js)

**Buka terminal baru, masuk ke folder frontend:**

```powershell
cd frontend
```

**Install dependencies:**

```powershell
npm install
```

**Buat file .env.local** (copy dari .env.example):

```powershell
Copy-Item .env.example .env.local
```

**Edit file .env.local:**

```env
NEXT_PUBLIC_API_URL=http://localhost:8080/api
NEXT_PUBLIC_MIDTRANS_CLIENT_KEY=your_client_key_here
```

**Jalankan frontend:**

```powershell
npm run dev
```

Frontend akan berjalan di: **http://localhost:3000**

### 4️⃣ Dapatkan Midtrans Keys

1. Daftar di [Midtrans](https://midtrans.com/)
2. Login ke [Dashboard Midtrans](https://dashboard.midtrans.com/)
3. Pilih mode **Sandbox** (untuk testing)
4. Buka **Settings → Access Keys**
5. Copy:
   - **Server Key** → untuk backend `.env`
   - **Client Key** → untuk frontend `.env.local`

## ✅ Cek Instalasi

1. **Backend**: Buka http://localhost:8080/api/products

   - Harus menampilkan data JSON produk

2. **Frontend**: Buka http://localhost:3000
   - Harus menampilkan website ZAVERA

## 🧪 Test Payment

Gunakan kartu test Midtrans:

- **Nomor Kartu**: 4811 1111 1111 1114
- **CVV**: 123
- **Exp Date**: Tanggal di masa depan (contoh: 12/25)
- **OTP/3DS**: 112233

## ❗ Troubleshooting

### Error: "database does not exist"

```powershell
psql -U postgres -c "CREATE DATABASE zavera_db"
```

### Error: "cannot find module"

```powershell
# Di folder frontend:
rm -r node_modules
npm install
```

### Error: "broken import" di backend

```powershell
# Di folder backend:
go mod tidy
go mod download
```

### Port sudah digunakan

- Backend: Ubah PORT di `.env`
- Frontend: Jalankan dengan `npm run dev -- -p 3001`

## 📁 Struktur Folder

```
ZAVERA FASHION STORE/
├── backend/          # API Golang
├── frontend/         # Website Next.js
├── database/         # SQL scripts
└── README.md         # Dokumentasi lengkap
```

## 🎯 Fitur Utama

- ✅ Katalog produk fashion
- ✅ Keranjang belanja
- ✅ Checkout & pembayaran real (Midtrans)
- ✅ Responsive design
- ✅ Stock management

## 📞 Bantuan

Jika ada error, cek:

1. PostgreSQL sudah running
2. File .env sudah terisi dengan benar
3. Dependencies sudah terinstall
4. Port tidak bentrok

---

**Selamat mencoba! 🎉**
