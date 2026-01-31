# Cara Verify Kategori Sekarang

## Yang Sudah Dipastikan

Mapping kategori di admin **SUDAH PERSIS SAMA** dengan client!

**Contoh Flow:**
```
Admin Input: "Atasan" 
    ↓
Database: "Tops" (English value)
    ↓
Client Display: "Atasan" (Indonesian label)
```

## Test Sekarang

### Test 1: Pria - Atasan

1. **Buka Admin:**
   ```
   http://localhost:3000/admin/products/add
   ```

2. **Isi Form:**
   - Name: `Test Atasan Pria`
   - Category: `Pria`
   - Subcategory: `Atasan` ← Pilih ini
   - Price: `100000`
   - Brand: `Test`
   - Material: `Cotton`
   - Upload 1 image
   - Add 1 variant (M, Black, 10)

3. **Create Product**

4. **Cek Database:**
   ```bash
   TEST_CATEGORY_MAPPING_NOW.bat
   ```
   
   Atau manual:
   ```sql
   SELECT id, name, category, subcategory 
   FROM products 
   WHERE name = 'Test Atasan Pria';
   ```
   
   **Harus muncul:**
   ```
   category: pria
   subcategory: Tops  ← Database value (English)
   ```

5. **Cek Client:**
   ```
   http://localhost:3000/pria
   ```
   - Click filter "Atasan"
   - Product "Test Atasan Pria" **HARUS MUNCUL**

### Test 2: Pria - Kemeja

1. **Admin Input:**
   - Category: Pria
   - Subcategory: Kemeja

2. **Database Check:**
   ```sql
   SELECT subcategory FROM products WHERE name = 'Test Kemeja';
   ```
   Expected: `Shirts`

3. **Client Check:**
   - Filter by "Kemeja"
   - Product harus muncul

### Test 3: Wanita - Bawahan

1. **Admin Input:**
   - Category: Wanita
   - Subcategory: Bawahan

2. **Database Check:**
   Expected: `Bottoms`

3. **Client Check:**
   - Filter by "Bawahan"
   - Product harus muncul

## Mapping Lengkap

### Pria (Yang Paling Sering Dipakai)
```
Admin → Database → Client
Atasan → Tops → Atasan
Kemeja → Shirts → Kemeja
Celana → Bottoms → Celana
Jaket → Outerwear → Jaket
Jas → Suits → Jas
Sepatu → Footwear → Sepatu
```

### Wanita
```
Admin → Database → Client
Dress → Dress → Dress
Atasan → Tops → Atasan
Bawahan → Bottoms → Bawahan
Outerwear → Outerwear → Outerwear
Aksesoris → Accessories → Aksesoris
```

### Anak
```
Admin → Database → Client
Anak Laki-laki → Boys → Anak Laki-laki
Anak Perempuan → Girls → Anak Perempuan
Bayi → Baby → Bayi
Sepatu → Footwear → Sepatu
```

## Kalau Product Tidak Muncul di Filter

### Kemungkinan 1: Database value salah

**Cek:**
```sql
SELECT id, name, subcategory FROM products WHERE name LIKE '%Test%';
```

**Kalau salah, fix:**
```sql
-- Example: Fix "Shirt" to "Shirts"
UPDATE products SET subcategory = 'Shirts' WHERE id = 123;
```

### Kemungkinan 2: Frontend belum refresh

**Fix:**
```bash
# Hard refresh browser
Ctrl+Shift+R

# Atau restart frontend
cd frontend
npm run dev
```

### Kemungkinan 3: Mapping tidak match

**Cek admin mapping:**
```typescript
// frontend/src/app/admin/products/add/page.tsx
const CATEGORIES = {
  pria: {
    subcategories: [
      { label: 'Atasan', value: 'Tops' },  // ← Harus ada
      // ...
    ]
  }
}
```

**Cek client mapping:**
```typescript
// frontend/src/components/FilterPanel.tsx
const SUBCATEGORY_MAPPING = {
  pria: {
    "Atasan": "Tops",  // ← Harus sama!
    // ...
  }
}
```

## Verification Script

```bash
# Run test script
TEST_CATEGORY_MAPPING_NOW.bat

# Akan menampilkan:
# 1. Test products di database
# 2. Semua distinct subcategories
# 3. Mapping verification untuk Pria
```

## Expected Output

```
Test Atasan Pria | pria | Tops | Atasan
Test Kemeja Pria | pria | Shirts | Kemeja
Test Celana Pria | pria | Bottoms | Celana
```

## Success Criteria

✅ Admin dropdown menampilkan label Indonesia
✅ Database menyimpan value English
✅ Client filter menampilkan label Indonesia (sama dengan admin)
✅ Product muncul di filter yang benar
✅ Tidak ada mismatch antara admin input dan client display

## Quick Test

```bash
# 1. Create product
# Admin: Pria → Atasan

# 2. Check database
psql -U postgres -d zavera_db -c "SELECT name, subcategory FROM products ORDER BY id DESC LIMIT 1;"
# Expected: subcategory = 'Tops'

# 3. Check client
# Browser: http://localhost:3000/pria
# Filter: Atasan
# Result: Product muncul ✅
```

## Summary

**Mapping sudah benar!** Admin dan Client menggunakan label Indonesia yang sama, dengan database menyimpan value English.

**Test sekarang:**
1. Create product di admin dengan subcategory "Atasan"
2. Verify database punya "Tops"
3. Verify client filter "Atasan" menampilkan product

Kalau semua test pass, mapping sudah 100% benar! 🚀
