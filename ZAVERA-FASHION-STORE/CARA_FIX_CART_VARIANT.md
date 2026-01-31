# Cara Fix Cart Variant Bug

## Masalah
1. ❌ Item yang baru ditambahkan tidak muncul di cart
2. ❌ Checkout error: "insufficient stock for product"
3. ❌ Hanya item lama yang muncul di cart

## Penyebab
Backend mengecek `product.Stock` untuk produk dengan varian, padahal `product.Stock = 0` (stok ada di varian).

## Solusi

### Step 1: Stop Backend Lama
Tekan `Ctrl+C` di terminal backend yang sedang running

### Step 2: Jalankan Backend Baru
```bash
start-backend-FINAL.bat
```

ATAU manual:
```bash
cd backend
zavera_FINAL_FIX.exe
```

### Step 3: Test
1. **Clear cart** (klik "Clear All" di cart page)
2. **Add to cart** produk dengan varian:
   - Pilih ukuran (L)
   - Pilih warna (Black)
   - Klik "Tambah ke Keranjang"
3. **Cek cart page** - Item harus muncul ✅
4. **Proceed to checkout** - Harus berhasil ✅

## Verifikasi Fix Berhasil

### ✅ Success Indicators
1. **Add to cart**: Toast "ditambahkan ke keranjang" muncul
2. **Cart page**: Item baru muncul dengan ukuran yang benar
3. **Backend log**: `POST "/api/cart/items" - 200` (bukan 400)
4. **Checkout**: Tidak ada error "insufficient stock"

### ❌ Jika Masih Error
1. Pastikan backend yang running adalah `zavera_FINAL_FIX.exe`
2. Cek terminal backend untuk log error
3. Clear browser cache (Ctrl+Shift+Del)
4. Logout dan login ulang
5. Clear cart dan add ulang

## Technical Details

### Yang Diubah
**File**: `backend/service/cart_service.go`

**Logika Lama**:
```go
if product.Stock < req.Quantity {
    return error  // ❌ Selalu error untuk varian
}
```

**Logika Baru**:
```go
// Skip check jika product.Stock = 0 (produk dengan varian)
if product.Stock > 0 && product.Stock < req.Quantity {
    return error
}
```

### Kenapa Ini Fix Masalahnya?
- Produk dengan varian: `product.Stock = 0` (normal)
- Kondisi `product.Stock > 0` = false
- Skip stock check
- Add to cart berhasil ✅

## Troubleshooting

### Problem: Item masih tidak muncul di cart
**Solution**:
1. Buka browser console (F12)
2. Lihat request ke `/api/cart/items`
3. Cek response status:
   - 200: Berhasil ✅
   - 400: Masih error, backend belum update ❌

### Problem: Checkout masih error
**Solution**:
1. Cek error message di checkout
2. Jika "insufficient stock": Backend belum update
3. Restart backend dengan binary baru

### Problem: Cart count di header tidak update
**Solution**:
1. Refresh page (F5)
2. Atau logout dan login ulang

## Files

### Binary Baru
- ✅ `backend/zavera_FINAL_FIX.exe`
- ✅ `backend/zavera_cart_variant_fix.exe` (sama)

### Batch Script
- ✅ `start-backend-FINAL.bat`

### Dokumentasi
- 📖 `CART_VARIANT_STOCK_FIX.md` - Technical details
- 📖 `CARA_FIX_CART_VARIANT.md` - This file

## Summary

**Masalah**: Cart variant tidak bisa ditambahkan
**Penyebab**: Backend reject karena `product.Stock = 0`
**Solusi**: Skip stock check untuk produk dengan varian
**Binary**: `zavera_FINAL_FIX.exe`
**Status**: ✅ FIXED

---

**Jika masih ada masalah, kirim screenshot:**
1. Browser console (F12) saat add to cart
2. Backend terminal log
3. Cart page setelah add to cart
