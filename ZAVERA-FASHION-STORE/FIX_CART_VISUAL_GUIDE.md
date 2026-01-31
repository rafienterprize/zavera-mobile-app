# 🎯 FIX CART - PANDUAN VISUAL

## 🔴 MASALAH SEKARANG

```
User mencoba add 2 variants:
┌─────────────────────────────┐
│ Hip Hop Jeans - XL × 1      │ ✅ Berhasil ditambahkan
└─────────────────────────────┘

┌─────────────────────────────┐
│ Hip Hop Jeans - L × 2       │ ❌ ERROR: duplicate key!
└─────────────────────────────┘

Cart hanya punya:
┌─────────────────────────────┐
│ Hip Hop Jeans - XL × 1      │ (L hilang!)
└─────────────────────────────┘
```

**Kenapa?** Database punya constraint yang bilang:
> "Tidak boleh ada 2 items dengan cart_id dan product_id yang sama!"

Tapi XL dan L itu product_id-nya SAMA (47), cuma metadata-nya beda!

---

## ✅ SOLUSI

### Step 1: Hapus Constraint Database

```bash
fix_cart_database.bat
```

**Yang terjadi:**
```sql
BEFORE:
┌──────────┬────────────┬──────────┬──────────┐
│ cart_id  │ product_id │ metadata │ quantity │
├──────────┼────────────┼──────────┼──────────┤
│    1     │     47     │ {"XL"}   │    1     │ ✅
│    1     │     47     │ {"L"}    │    2     │ ❌ BLOCKED by constraint!
└──────────┴────────────┴──────────┴──────────┘

AFTER:
┌──────────┬────────────┬──────────┬──────────┐
│ cart_id  │ product_id │ metadata │ quantity │
├──────────┼────────────┼──────────┼──────────┤
│    1     │     47     │ {"XL"}   │    1     │ ✅
│    1     │     47     │ {"L"}    │    2     │ ✅ ALLOWED!
└──────────┴────────────┴──────────┴──────────┘
```

---

### Step 2: Restart Backend

```bash
start-backend-COMPLETE.bat
```

**Yang terjadi:**
```
OLD Backend:
┌─────────────────────────────────────┐
│ Check: product.Stock < quantity     │ ❌ Error untuk variants!
│ Compare: cart_id + product_id       │ ❌ Tidak cek metadata!
└─────────────────────────────────────┘

NEW Backend (zavera_COMPLETE_FIX.exe):
┌─────────────────────────────────────┐
│ Check: IF stock > 0 THEN validate   │ ✅ Skip untuk variants!
│ Compare: cart_id + product_id       │
│          + metadata                 │ ✅ Cek metadata!
└─────────────────────────────────────┘
```

---

### Step 3: Test

```
1. Clear cart lama
   ┌─────────────────────────────┐
   │ Cart: Empty                 │
   └─────────────────────────────┘

2. Add XL
   ┌─────────────────────────────┐
   │ Hip Hop Jeans - XL × 1      │ ✅
   └─────────────────────────────┘

3. Add L
   ┌─────────────────────────────┐
   │ Hip Hop Jeans - XL × 1      │ ✅
   │ Hip Hop Jeans - L × 2       │ ✅ Berhasil!
   └─────────────────────────────┘

4. Checkout
   ┌─────────────────────────────┐
   │ ✅ No error "undefined"     │
   │ ✅ Semua items muncul       │
   │ ✅ Bisa pilih courier       │
   │ ✅ Bisa bayar               │
   └─────────────────────────────┘
```

---

## 🔍 CARA CEK BERHASIL

### Test 1: Add Multiple Variants

```
Action:
1. Buka: http://localhost:3000/product/47
2. Pilih XL, Add to Cart
3. Pilih L, Add to Cart

Expected Backend Log:
✅ POST "/api/cart/items" - 200 (XL)
✅ POST "/api/cart/items" - 200 (L)

❌ JANGAN ADA:
   "duplicate key violates unique constraint"
   "insufficient stock"
```

### Test 2: Cart Display

```
Action:
1. Buka: http://localhost:3000/cart

Expected:
┌─────────────────────────────────────────┐
│ Hip Hop Baggy Jeans                     │
│ Size: XL, Color: Black                  │
│ Quantity: 1                             │
│ Price: Rp 250,000                       │
├─────────────────────────────────────────┤
│ Hip Hop Baggy Jeans                     │
│ Size: L, Color: Black                   │
│ Quantity: 2                             │
│ Price: Rp 500,000                       │
└─────────────────────────────────────────┘
Total: Rp 750,000

❌ JANGAN:
   - Hanya 1 item (L hilang)
   - Error "undefined items available"
```

### Test 3: Checkout

```
Action:
1. Cart → Proceed to Checkout
2. Isi alamat
3. Pilih courier
4. Pilih payment

Expected:
✅ Semua items muncul di checkout
✅ Bisa pilih courier (tidak redirect ke cart)
✅ Bisa pilih payment
✅ Bisa klik "Bayar Sekarang"

❌ JANGAN:
   - Redirect balik ke cart
   - Error "undefined items"
   - Cart kosong di checkout
```

---

## 🚨 TROUBLESHOOTING

### Problem: Masih error "duplicate key"

```
Diagnosis:
┌─────────────────────────────────────┐
│ Constraint masih ada di database   │
└─────────────────────────────────────┘

Check:
psql -U postgres -d zavera -c "SELECT conname FROM pg_constraint WHERE conrelid = 'cart_items'::regclass;"

Output:
┌──────────────────────────────────────┐
│ cart_items_cart_id_product_id_key    │ ❌ MASIH ADA!
└──────────────────────────────────────┘

Fix:
psql -U postgres -d zavera -c "ALTER TABLE cart_items DROP CONSTRAINT cart_items_cart_id_product_id_key;"

Output:
┌──────────────────────────────────────┐
│ ALTER TABLE                          │ ✅ BERHASIL!
└──────────────────────────────────────┘
```

### Problem: Masih error "undefined items"

```
Diagnosis:
┌─────────────────────────────────────┐
│ Backend masih pakai binary lama     │
└─────────────────────────────────────┘

Check:
tasklist | findstr zavera

Output:
┌──────────────────────────────────────┐
│ zavera.exe                           │ ❌ BINARY LAMA!
└──────────────────────────────────────┘

Fix:
1. Stop: Ctrl+C atau taskkill /F /IM zavera.exe
2. Start: start-backend-COMPLETE.bat

Output:
┌──────────────────────────────────────┐
│ zavera_COMPLETE_FIX.exe              │ ✅ BINARY BARU!
└──────────────────────────────────────┘
```

### Problem: Cart kosong di checkout

```
Diagnosis:
┌─────────────────────────────────────┐
│ Cart lama punya data rusak          │
└─────────────────────────────────────┘

Fix:
1. Logout
2. Clear browser cache (Ctrl+Shift+Delete)
3. Login
4. Clear cart (klik "Clear All")
5. Add items baru
6. Checkout

Result:
┌─────────────────────────────────────┐
│ ✅ Cart bersih, checkout berhasil   │
└─────────────────────────────────────┘
```

---

## 📊 FLOW DIAGRAM

### BEFORE (Broken):

```
User Add XL
    ↓
Backend: Check stock (product.Stock = 0)
    ↓
❌ Error: "insufficient stock"
    ↓
Cart: Empty

User Add L
    ↓
Backend: Check duplicate (cart_id + product_id)
    ↓
❌ Error: "duplicate key"
    ↓
Cart: Only XL (L hilang)

User Checkout
    ↓
Backend: Validate cart (product.Stock = 0)
    ↓
❌ Error: "undefined items available"
    ↓
Checkout: Empty
```

### AFTER (Fixed):

```
User Add XL
    ↓
Backend: Check stock (IF stock > 0)
    ↓
✅ Skip check (stock = 0 for variants)
    ↓
Backend: Check duplicate (cart_id + product_id + metadata)
    ↓
✅ Not found, insert new
    ↓
Cart: XL × 1

User Add L
    ↓
Backend: Check stock (IF stock > 0)
    ↓
✅ Skip check (stock = 0 for variants)
    ↓
Backend: Check duplicate (cart_id + product_id + metadata)
    ↓
✅ Not found (metadata different), insert new
    ↓
Cart: XL × 1, L × 2

User Checkout
    ↓
Backend: Validate cart (IF stock > 0)
    ↓
✅ Skip validation (stock = 0 for variants)
    ↓
Checkout: XL × 1, L × 2 ✅
```

---

## 🎯 CHECKLIST LENGKAP

### Pre-Fix:
- [ ] Backend running (cek dengan: `tasklist | findstr zavera`)
- [ ] Database running (cek dengan: `psql -U postgres -d zavera -c "SELECT 1;"`)
- [ ] Browser buka di: http://localhost:3000

### Fix:
- [ ] **Step 1**: Jalankan `fix_cart_database.bat`
  - [ ] Output: "ALTER TABLE" ✅
  - [ ] Tidak ada error ✅
  
- [ ] **Step 2**: Stop backend lama (Ctrl+C)
  - [ ] Jalankan `start-backend-COMPLETE.bat`
  - [ ] Output: "Starting Zavera Backend - COMPLETE FIX" ✅
  
- [ ] **Step 3**: Clear cart
  - [ ] Buka: http://localhost:3000/cart
  - [ ] Klik "Clear All" ✅

### Test:
- [ ] Add XL → Cart punya XL ✅
- [ ] Add L → Cart punya XL + L ✅
- [ ] Tidak ada error "duplicate key" ✅
- [ ] Checkout → Semua items muncul ✅
- [ ] Tidak ada error "undefined items" ✅
- [ ] Bisa pilih courier ✅
- [ ] Bisa bayar ✅

### Done! 🎉

---

**INGAT**: Langkah 1 (fix database) WAJIB dijalankan dulu! Ini yang paling penting!
