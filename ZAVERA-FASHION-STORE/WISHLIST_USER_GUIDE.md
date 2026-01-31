# 🛍️ Panduan Penggunaan Fitur Wishlist - ZAVERA Fashion Store

## ✅ Status Implementasi
**FITUR WISHLIST SUDAH SIAP DIGUNAKAN!**

Semua komponen backend dan frontend sudah terintegrasi dengan sempurna:
- ✅ Backend API (Go/Gin) - Compiled successfully
- ✅ Database migration - Executed
- ✅ Frontend Context (React) - No TypeScript errors
- ✅ UI Components - Fully integrated
- ✅ Routes registered - All endpoints active

---

## 📋 Cara Menggunakan Wishlist

### 1️⃣ Menambahkan Produk ke Wishlist

**Dari Halaman Produk:**
1. Login terlebih dahulu (wajib)
2. Browse produk di halaman utama atau kategori
3. Hover mouse ke product card
4. Klik icon **❤️ (heart)** yang muncul di pojok kanan atas
5. Toast notification akan muncul: "Added to wishlist"
6. Icon heart akan berubah menjadi merah (filled)

**Indikator Produk Sudah di Wishlist:**
- Icon heart berwarna merah dengan background merah
- Badge counter di header bertambah

### 2️⃣ Melihat Wishlist

**Akses Halaman Wishlist:**
1. Klik icon **❤️** di header (sebelah cart icon)
2. Atau akses langsung: `http://localhost:3000/wishlist`

**Tampilan Wishlist:**
- Grid layout responsive (1-4 kolom tergantung ukuran layar)
- Setiap item menampilkan:
  - Gambar produk
  - Nama produk
  - Harga
  - Status ketersediaan (Available/Out of Stock)
  - Tombol "MOVE TO CART"
  - Tombol "Remove" (X)

### 3️⃣ Memindahkan ke Cart

1. Buka halaman wishlist
2. Klik tombol **"MOVE TO CART"** pada produk yang diinginkan
3. Produk akan:
   - Ditambahkan ke cart dengan quantity 1
   - Dihapus dari wishlist
   - Toast notification: "Moved to cart"

**Catatan:**
- Tombol disabled jika produk out of stock
- Otomatis redirect ke login jika belum login

### 4️⃣ Menghapus dari Wishlist

**Cara 1 - Dari Halaman Wishlist:**
1. Klik tombol **X** (remove) pada produk
2. Toast notification: "Removed from wishlist"

**Cara 2 - Dari Product Card:**
1. Hover ke product card
2. Klik icon heart merah (filled)
3. Icon akan kembali menjadi outline (tidak filled)

---

## 🎨 Fitur UI/UX

### Real-time Counter
- Badge merah di header menampilkan jumlah item di wishlist
- Update otomatis setiap ada perubahan
- Animasi smooth saat counter berubah

### Empty State
Jika wishlist kosong, akan tampil:
- Icon heart besar
- Pesan: "Your wishlist is empty"
- Tombol "EXPLORE PRODUCTS" untuk browse

### Loading State
- Skeleton loading saat data sedang dimuat
- Spinner animation

### Responsive Design
- Mobile: 1 kolom
- Tablet: 2-3 kolom
- Desktop: 4 kolom

---

## 🔐 Autentikasi

**Wishlist WAJIB Login:**
- Semua endpoint wishlist memerlukan authentication token
- Jika belum login, akan redirect ke halaman login
- Setelah login, otomatis kembali ke halaman sebelumnya

**Flow Login:**
```
User klik heart → Belum login → Redirect ke /login?redirect=/products
→ Login berhasil → Kembali ke /products → Klik heart lagi → Berhasil
```

---

## 🔌 API Endpoints

### 1. Get Wishlist
```
GET /api/wishlist
Authorization: Bearer <token>

Response:
{
  "items": [
    {
      "id": 1,
      "product_id": 123,
      "product_name": "Kemeja Formal Pria",
      "product_image": "https://...",
      "product_price": 299000,
      "product_stock": 10,
      "is_available": true,
      "added_at": "2024-01-27T10:00:00Z"
    }
  ],
  "count": 1
}
```

### 2. Add to Wishlist
```
POST /api/wishlist
Authorization: Bearer <token>
Content-Type: application/json

Body:
{
  "product_id": 123
}

Response: Same as GET /api/wishlist
```

### 3. Remove from Wishlist
```
DELETE /api/wishlist/:productId
Authorization: Bearer <token>

Response: Same as GET /api/wishlist
```

### 4. Move to Cart
```
POST /api/wishlist/:productId/move-to-cart
Authorization: Bearer <token>

Response: Cart data (same as GET /api/cart)
```

---

## 🧪 Testing

### Manual Testing Steps

**1. Test Add to Wishlist:**
```bash
# Login first to get token
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password"}'

# Copy the token from response

# Add product to wishlist
curl -X POST http://localhost:8080/api/wishlist \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"product_id": 1}'
```

**2. Test Get Wishlist:**
```bash
curl -X GET http://localhost:8080/api/wishlist \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

**3. Test Remove from Wishlist:**
```bash
curl -X DELETE http://localhost:8080/api/wishlist/1 \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

**4. Test Move to Cart:**
```bash
curl -X POST http://localhost:8080/api/wishlist/1/move-to-cart \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### Automated Test Script
Jalankan file `test_wishlist.bat` untuk test cepat:
```bash
test_wishlist.bat
```

---

## 🗄️ Database Schema

```sql
CREATE TABLE wishlists (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id),
  product_id INTEGER NOT NULL REFERENCES products(id),
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  UNIQUE(user_id, product_id)
);

-- Trigger untuk auto-update wishlist_count di products table
CREATE TRIGGER update_wishlist_count_on_insert ...
CREATE TRIGGER update_wishlist_count_on_delete ...
```

---

## 🚀 Cara Menjalankan

### Backend
```bash
cd backend
.\zavera_refund_friendly_errors.exe
# atau
.\zavera_wishlist_test.exe
```

Backend akan berjalan di: `http://localhost:8080`

### Frontend
```bash
cd frontend
npm run dev
```

Frontend akan berjalan di: `http://localhost:3000`

---

## 🐛 Troubleshooting

### Problem: "Please login to add items to wishlist"
**Solution:** User belum login. Klik tombol login di header.

### Problem: "Product not found"
**Solution:** Product ID tidak valid atau produk sudah dihapus.

### Problem: "Product is not available"
**Solution:** Produk sudah tidak aktif (is_active = false).

### Problem: Wishlist count tidak update
**Solution:** 
1. Refresh halaman
2. Cek console browser untuk error
3. Pastikan backend running
4. Cek network tab untuk API response

### Problem: "Failed to add to wishlist"
**Solution:**
1. Cek apakah token masih valid
2. Cek apakah backend running
3. Cek database connection
4. Lihat backend logs untuk detail error

---

## 📊 Fitur Tambahan

### Duplicate Prevention
- Database constraint: UNIQUE(user_id, product_id)
- User tidak bisa menambahkan produk yang sama 2x
- Backend akan return error jika duplicate

### Stock Validation
- Produk out of stock tetap bisa di wishlist
- Tapi tombol "Move to Cart" akan disabled
- Badge "OUT OF STOCK" akan muncul

### Product Deletion Handling
- Jika produk dihapus, item wishlist akan di-skip saat load
- Tidak akan error, hanya tidak ditampilkan

---

## 🎯 Best Practices

### Untuk User:
1. Login sebelum browse untuk pengalaman lebih baik
2. Gunakan wishlist untuk save produk favorit
3. Cek wishlist secara berkala untuk promo/diskon
4. Move to cart saat siap checkout

### Untuk Developer:
1. Selalu handle authentication error
2. Implement optimistic UI updates
3. Show loading states
4. Provide clear error messages
5. Test dengan berbagai skenario (login/logout, network error, dll)

---

## 📝 Changelog

### Version 1.0.0 (Current)
- ✅ Complete wishlist CRUD operations
- ✅ Real-time counter in header
- ✅ Move to cart functionality
- ✅ Responsive grid layout
- ✅ Empty state & loading states
- ✅ Toast notifications
- ✅ Authentication integration
- ✅ Stock validation
- ✅ Dark theme with orange accents

---

## 🔮 Future Enhancements (Optional)

- [ ] Share wishlist via link
- [ ] Wishlist collections/folders
- [ ] Price drop notifications
- [ ] Back in stock alerts
- [ ] Wishlist analytics (most wishlisted products)
- [ ] Export wishlist to PDF
- [ ] Social sharing (WhatsApp, Instagram)

---

## 📞 Support

Jika ada pertanyaan atau issue:
1. Cek dokumentasi ini terlebih dahulu
2. Lihat backend logs untuk error details
3. Cek browser console untuk frontend errors
4. Review `WISHLIST_IMPLEMENTATION.md` untuk technical details

---

**🎉 SELAMAT! Fitur Wishlist Sudah Siap Digunakan!**

Silakan test dan nikmati fitur wishlist yang sudah terintegrasi penuh dengan sistem ZAVERA Fashion Store.
