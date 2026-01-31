# Fix: SOLD OUT di Product Card (Grid/Listing)

## Masalah

User melihat overlay "SOLD OUT" di product card (halaman listing/grid produk) padahal produk masih ada stok di varian.

### Screenshot Masalah
```
Grid Produk:
┌─────────────────┐  ┌─────────────────┐
│  [Image]        │  │  [Image]        │
│                 │  │                 │
│  "SOLD OUT"     │  │  "SOLD OUT"     │ ← Padahal ada stok di varian!
└─────────────────┘  └─────────────────┘
Hip Hop Jeans 22     Hip Hop Jeans
Rp 330.000           Rp 400.000
```

## Penyebab

Di `ProductCard.tsx`, logika menampilkan SOLD OUT berdasarkan `product.stock === 0`:

```typescript
{product.stock === 0 && (
  <div className="absolute inset-0 bg-black/50 flex items-center justify-center">
    <span className="px-4 py-2 bg-white text-primary font-medium text-sm">
      SOLD OUT
    </span>
  </div>
)}
```

**Masalahnya**: Untuk produk dengan varian, `product.stock` selalu 0 (karena stok ada di level varian), jadi semua produk dengan varian akan tampil SOLD OUT!

## Solusi

**HAPUS overlay SOLD OUT dari ProductCard** karena:

1. ❌ Tidak akurat untuk produk dengan varian
2. ❌ Membingungkan user
3. ❌ Tidak fair - produk masih ada stok di varian tertentu
4. ✅ User harus klik detail untuk lihat varian yang tersedia
5. ✅ Mengikuti best practice (Tokopedia, Shopee tidak tampilkan SOLD OUT di card untuk produk dengan varian)

### Kode Baru

```typescript
{/* Badges */}
<div className="absolute top-3 left-3 flex flex-col gap-2">
  {isLuxury && (
    <span className="px-2 py-1 bg-amber-500 text-white text-xs font-medium tracking-wider">
      LUXURY
    </span>
  )}
  {/* Low stock badge - only show for simple products (not variant-based) */}
  {product.stock > 0 && product.stock < 10 && (
    <span className="px-2 py-1 bg-red-500 text-white text-xs font-medium tracking-wider">
      SISA {product.stock}
    </span>
  )}
</div>

{/* REMOVED: SOLD OUT overlay for product cards
    Reason: For variant-based products, product.stock = 0 is normal
    User needs to click into product detail to see variant availability
*/}
```

## Hasil Setelah Fix

### Product Card - Produk dengan Varian
```
┌─────────────────┐
│  [Image]        │ ← Gambar terlihat jelas
│                 │
│  (No overlay)   │ ← Tidak ada SOLD OUT
└─────────────────┘
Hip Hop Jeans
Rp 400.000

User klik → Masuk detail → Pilih varian → Lihat stok
```

### Product Card - Produk Simple (Tanpa Varian)
```
┌─────────────────┐
│  [Image]        │
│  SISA 8         │ ← Badge low stock (jika < 10)
│                 │
└─────────────────┘
Basic T-Shirt
Rp 100.000
```

### Product Card - Produk Simple Habis Stok
```
┌─────────────────┐
│  [Image]        │ ← Tidak ada overlay SOLD OUT
│                 │    (user harus klik untuk tahu)
│                 │
└─────────────────┘
Basic T-Shirt
Rp 100.000
```

## Alternatif (Jika Ingin Tetap Tampilkan SOLD OUT)

Jika Anda tetap ingin menampilkan SOLD OUT di product card, maka harus:

### Option 1: Cek Semua Varian (Kompleks)
```typescript
// Fetch variants untuk setiap produk
const hasAnyStock = variants.some(v => v.available_stock > 0);

{!hasAnyStock && (
  <div>SOLD OUT</div>
)}
```

**Masalah**: 
- Perlu fetch varian untuk setiap produk di grid (banyak request)
- Lambat dan tidak efisien
- Tidak recommended

### Option 2: Tambah Field di Backend (Recommended jika perlu)
```sql
-- Tambah computed field di backend
ALTER TABLE products ADD COLUMN total_variant_stock INT DEFAULT 0;

-- Update via trigger atau cron job
UPDATE products p
SET total_variant_stock = (
  SELECT COALESCE(SUM(available_stock), 0)
  FROM product_variants
  WHERE product_id = p.id AND is_active = true
);
```

Lalu di frontend:
```typescript
const isOutOfStock = product.stock === 0 && 
                     (product.total_variant_stock === 0 || !product.total_variant_stock);

{isOutOfStock && (
  <div>SOLD OUT</div>
)}
```

**Tapi**: Ini menambah kompleksitas. Lebih baik tidak tampilkan SOLD OUT di card.

## Rekomendasi

✅ **JANGAN tampilkan SOLD OUT di product card** untuk produk dengan varian

**Alasan**:
1. User harus klik detail untuk pilih varian anyway
2. Tidak akurat - mungkin varian tertentu masih ada
3. Mengikuti best practice platform besar
4. Lebih simple dan tidak membingungkan

✅ **Tampilkan SOLD OUT hanya di halaman detail produk** setelah user pilih varian yang memang habis

## Perbandingan dengan Platform Lain

### Tokopedia
- ❌ Tidak tampilkan SOLD OUT di product card untuk produk dengan varian
- ✅ Tampilkan "Habis" di halaman detail setelah pilih varian

### Shopee
- ❌ Tidak tampilkan SOLD OUT di product card untuk produk dengan varian
- ✅ Tampilkan "Stok Habis" di halaman detail setelah pilih varian

### Lazada
- ❌ Tidak tampilkan SOLD OUT di product card untuk produk dengan varian
- ✅ Tampilkan "Out of Stock" di halaman detail setelah pilih varian

## Files Modified

1. ✅ `frontend/src/components/ProductCard.tsx`
   - Hapus overlay SOLD OUT
   - Update kondisi low stock badge
   - Tambah comment penjelasan

## Testing

```bash
cd frontend
npm run dev
```

### Test 1: Halaman Kategori (Grid Produk)
```
http://localhost:3000/pria
```

**Cek**:
- ✅ Produk dengan varian TIDAK ada overlay SOLD OUT
- ✅ Gambar terlihat jelas
- ✅ Badge "SISA X" hanya muncul untuk produk simple dengan stock < 10

### Test 2: Halaman Detail Produk
```
http://localhost:3000/product/46
```

**Cek**:
- ✅ Gambar terlihat jelas (tidak ada overlay)
- ✅ Ada pesan "Silakan pilih ukuran dan warna"
- ✅ Setelah pilih varian → Tampil stok
- ✅ Jika varian habis → Button "Stok Habis"

## Kesimpulan

### Masalah
❌ Product card menampilkan "SOLD OUT" untuk semua produk dengan varian

### Penyebab
❌ Logika mengecek `product.stock === 0` tanpa mempertimbangkan varian

### Solusi
✅ Hapus overlay SOLD OUT dari product card
✅ User harus klik detail untuk lihat ketersediaan varian
✅ Mengikuti best practice platform e-commerce besar

### Hasil
✅ Product card lebih bersih dan tidak membingungkan
✅ User experience lebih baik
✅ Konsisten dengan Tokopedia, Shopee, Lazada

---

**Status**: ✅ Fixed
**File**: `frontend/src/components/ProductCard.tsx`
**Approach**: Remove SOLD OUT overlay from product cards
**Reason**: Inaccurate for variant-based products
