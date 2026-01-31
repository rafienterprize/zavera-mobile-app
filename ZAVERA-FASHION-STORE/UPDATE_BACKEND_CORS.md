# 🔧 Update Backend CORS untuk Cloudflare Pages

## ⚠️ PENTING: Lakukan ini SETELAH deploy frontend!

Setelah frontend Anda live di Cloudflare Pages, backend perlu diupdate agar bisa menerima request dari domain Cloudflare.

---

## 📝 Langkah-Langkah

### Step 1: Buka File Backend

Buka file: `backend/main.go`

### Step 2: Cari CORS Configuration

Cari bagian code yang seperti ini (sekitar line 50-60):

```go
config := cors.Config{
    AllowOrigins: []string{
        "http://localhost:3000",
    },
    AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
    AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
    AllowCredentials: true,
}
```

### Step 3: Tambahkan Domain Cloudflare

Update menjadi seperti ini:

```go
config := cors.Config{
    AllowOrigins: []string{
        "http://localhost:3000",                      // ← Development
        "https://zavera-fashion-store.pages.dev",     // ← Production (Cloudflare)
    },
    AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
    AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
    AllowCredentials: true,
}
```

**⚠️ GANTI** `zavera-fashion-store.pages.dev` dengan domain Cloudflare Anda yang sebenarnya!

### Step 4: Rebuild Backend

```bash
go build -o zavera.exe
```

### Step 5: Restart Backend

Stop backend yang sedang running (Ctrl+C), lalu start lagi:

```bash
.\zavera.exe
```

---

## ✅ Verifikasi CORS Berhasil

### Test 1: Buka Frontend di Browser

Buka: `https://zavera-fashion-store.pages.dev`

### Test 2: Buka Browser Console

Tekan `F12` → Tab "Console"

### Test 3: Test API Call

Coba login atau load products. Jika berhasil, tidak ada error CORS di console.

### ❌ Jika Ada Error CORS

Error akan terlihat seperti ini di console:

```
Access to fetch at 'http://your-backend.com/api/products' from origin 
'https://zavera-fashion-store.pages.dev' has been blocked by CORS policy: 
No 'Access-Control-Allow-Origin' header is present on the requested resource.
```

**Solusi:**
1. Pastikan domain Cloudflare sudah ditambahkan ke `AllowOrigins`
2. Pastikan backend sudah di-rebuild
3. Pastikan backend sudah di-restart
4. Clear browser cache (Ctrl+Shift+Delete)
5. Refresh page (Ctrl+F5)

---

## 🌐 Jika Pakai Custom Domain

Jika nanti Anda add custom domain (misalnya `www.zavera.com`), tambahkan juga:

```go
config := cors.Config{
    AllowOrigins: []string{
        "http://localhost:3000",                      // Development
        "https://zavera-fashion-store.pages.dev",     // Cloudflare default
        "https://www.zavera.com",                     // Custom domain
    },
    AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
    AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
    AllowCredentials: true,
}
```

Jangan lupa rebuild dan restart backend setiap kali update CORS!

---

## 🔒 Security Notes

### ✅ DO:
- Hanya tambahkan domain yang Anda kontrol
- Gunakan HTTPS untuk production
- Restart backend setelah update CORS

### ❌ DON'T:
- Jangan gunakan wildcard `*` di production
- Jangan tambahkan domain yang tidak Anda kenal
- Jangan lupa restart backend setelah update

---

## 🐛 Troubleshooting

### Issue: CORS error masih muncul setelah update

**Checklist:**
- [ ] Domain Cloudflare sudah benar di `AllowOrigins`?
- [ ] Backend sudah di-rebuild? (`go build -o zavera.exe`)
- [ ] Backend sudah di-restart? (Stop lalu start lagi)
- [ ] Browser cache sudah di-clear?
- [ ] Tidak ada typo di domain?

### Issue: Backend tidak bisa di-build

**Error:** `package cors not found`

**Solusi:**
```bash
go get github.com/gin-contrib/cors
go mod tidy
go build -o zavera.exe
```

### Issue: Backend crash setelah restart

**Solusi:**
1. Check error message di terminal
2. Pastikan tidak ada syntax error di `main.go`
3. Pastikan port 8080 tidak digunakan aplikasi lain
4. Coba build ulang: `go build -o zavera.exe`

---

## 📋 Quick Reference

**File to edit:** `backend/main.go`

**What to add:**
```go
"https://zavera-fashion-store.pages.dev",  // Add this line
```

**Commands:**
```bash
# Rebuild
go build -o zavera.exe

# Restart
.\zavera.exe
```

**Test:**
- Open frontend in browser
- Check console for CORS errors
- Try login or load products

---

## ✅ Done!

Setelah CORS diupdate, frontend Cloudflare Anda bisa berkomunikasi dengan backend!

Test semua fitur untuk memastikan semuanya bekerja:
- ✅ Homepage load
- ✅ Products tampil
- ✅ Login works
- ✅ Admin panel accessible
- ✅ Biteship resi generation works
- ✅ Checkout flow works
- ✅ Payment works

---

**Happy deploying! 🚀**

