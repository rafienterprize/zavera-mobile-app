# Cek Brand & Material Sekarang

## Yang Sudah Saya Tambahkan

Saya sudah tambahkan logging di frontend untuk debug kenapa Brand & Material tidak muncul.

## Cara Test

### 1. Refresh Browser
Buka halaman product Shirt Eiger:
```
http://localhost:3000/product/60
```

Tekan F5 atau Ctrl+R untuk refresh

### 2. Buka Console
Tekan F12 → Tab Console

### 3. Lihat Logs

**Yang harus muncul:**
```
🔍 Product data loaded: { ... }
🏷️ Brand: Eiger
🧵 Material: Cotton
🔍 Checking brand/material display:
  - product.brand: Eiger
  - product.material: Cotton
  - Should show? true
```

### 4. Lihat UI

**Yang harus muncul di halaman:**
Kotak abu-abu dengan:
```
┌─────────────────────────────┐
│ ℹ️ Detail Produk            │
├─────────────────────────────┤
│ Brand        Material       │
│ Eiger        Cotton         │
└─────────────────────────────┘
```

## Kalau Tidak Muncul

### Kemungkinan 1: Console log menunjukkan undefined
```
🏷️ Brand: undefined
🧵 Material: undefined
```

**Artinya:** Backend tidak kirim data

**Solusi:**
```bash
# Restart backend
RESTART_BACKEND_FIX2.bat
```

### Kemungkinan 2: Console log menunjukkan empty string
```
🏷️ Brand: ""
🧵 Material: ""
```

**Artinya:** Database punya empty string

**Solusi:**
```sql
UPDATE products 
SET brand = 'Eiger', material = 'Cotton' 
WHERE id = 60;
```

### Kemungkinan 3: Console log benar tapi UI tidak muncul
```
🏷️ Brand: Eiger
🧵 Material: Cotton
Should show? true
```
Tapi UI tidak ada kotak abu-abu

**Artinya:** CSS issue atau cache

**Solusi:**
1. Hard refresh: Ctrl+Shift+R
2. Clear cache
3. Restart frontend

## Cek Database

```sql
SELECT id, name, brand, material 
FROM products 
WHERE id = 60;
```

**Harus muncul:**
```
 id |    name     | brand | material 
----+-------------+-------+----------
 60 | Shirt Eiger | Eiger | Cotton
```

Kalau NULL atau kosong:
```sql
UPDATE products 
SET brand = 'Eiger', material = 'Cotton' 
WHERE id = 60;
```

## Yang Perlu Di-Share

Kalau masih tidak muncul, share:

1. **Screenshot console logs** (semua yang ada 🔍 🏷️ 🧵)
2. **Screenshot halaman product**
3. **Hasil query database:**
   ```sql
   SELECT id, name, brand, material FROM products WHERE id = 60;
   ```
4. **Backend console** (ada error tidak?)

## Test Cepat

```bash
# 1. Cek database
psql -U postgres -d zavera_db -c "SELECT id, name, brand, material FROM products WHERE id = 60;"

# 2. Kalau kosong, update
psql -U postgres -d zavera_db -c "UPDATE products SET brand = 'Eiger', material = 'Cotton' WHERE id = 60;"

# 3. Restart backend
RESTART_BACKEND_FIX2.bat

# 4. Refresh browser
# Buka: http://localhost:3000/product/60
# Tekan: Ctrl+Shift+R (hard refresh)

# 5. Cek console (F12)
```

## Hasil yang Diharapkan

### ✅ Berhasil
- Console log menunjukkan Brand: Eiger, Material: Cotton
- UI menampilkan kotak "Detail Produk"
- Brand dan Material terlihat jelas

### ❌ Gagal
- Console log undefined/empty
- UI tidak ada kotak "Detail Produk"
- Kotak ada tapi kosong

## Next

Setelah Brand & Material muncul:
1. ✅ Test product creation (buat product baru dengan brand/material)
2. ✅ Test edit variant (VariantManagerNew component)
3. ✅ Verify semua flow lengkap

Silakan test dan share hasilnya! 🚀
