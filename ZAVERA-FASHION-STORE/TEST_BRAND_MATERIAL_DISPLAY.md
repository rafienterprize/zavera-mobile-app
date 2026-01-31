# Test Brand & Material Display

## Issue
Brand dan Material tidak muncul di halaman product detail customer, padahal sudah ada di database.

## What I've Added

### 1. Logging di Product Load
```javascript
console.log('🔍 Product data loaded:', productData);
console.log('🏷️ Brand:', productData.brand);
console.log('🧵 Material:', productData.material);
```

### 2. Logging di Render Check
```javascript
console.log('🔍 Checking brand/material display:');
console.log('  - product.brand:', product.brand);
console.log('  - product.material:', product.material);
console.log('  - Should show?', !!(product.brand || product.material));
```

## How to Test

### Step 1: Refresh Frontend
Jika frontend sudah running, refresh browser (Ctrl+R atau F5)

Jika belum running:
```bash
cd frontend
npm run dev
```

### Step 2: Open Product Page
```
http://localhost:3000/product/60
```
(Product ID 60 = Shirt Eiger yang punya Brand "Eiger" dan Material "Cotton")

### Step 3: Open Browser Console
Press F12 → Console tab

### Step 4: Check Logs

**Expected logs:**
```
🔍 Product data loaded: { id: 60, name: "Shirt Eiger", brand: "Eiger", material: "Cotton", ... }
🏷️ Brand: Eiger
🧵 Material: Cotton
🔍 Checking brand/material display:
  - product.brand: Eiger
  - product.material: Cotton
  - Should show? true
```

**Expected UI:**
Should see a gray box with:
```
Detail Produk
Brand          Material
Eiger          Cotton
```

## Possible Issues

### Issue 1: Brand/Material are undefined in logs
**Meaning:** Backend tidak mengirim data

**Check:**
1. Restart backend: `RESTART_BACKEND_FIX2.bat`
2. Check backend response:
   ```bash
   curl http://localhost:8080/api/products/60
   ```
3. Should see: `"brand":"Eiger","material":"Cotton"`

### Issue 2: Brand/Material are empty strings in logs
**Meaning:** Database memiliki empty string, bukan NULL

**Fix:**
```sql
UPDATE products 
SET brand = 'Eiger', material = 'Cotton' 
WHERE id = 60;
```

### Issue 3: "Should show?" is false
**Meaning:** Kondisi render tidak terpenuhi

**Check:**
- Are brand/material truthy values?
- Are they not empty strings?

### Issue 4: Logs show correct data but UI doesn't show
**Meaning:** CSS issue atau component tidak re-render

**Fix:**
1. Hard refresh: Ctrl+Shift+R
2. Clear cache
3. Check if element exists in DOM (Inspect Element)

## Database Verification

Check current data:
```sql
SELECT id, name, brand, material 
FROM products 
WHERE id = 60;
```

Expected result:
```
 id |    name     | brand | material 
----+-------------+-------+----------
 60 | Shirt Eiger | Eiger | Cotton
```

If brand/material are NULL or empty:
```sql
UPDATE products 
SET brand = 'Eiger', material = 'Cotton' 
WHERE id = 60;
```

## Test Other Products

Try products without brand/material:
```
http://localhost:3000/product/5
```
(Slim Fit Shirt - no brand/material)

**Expected:** No "Detail Produk" section should appear

## What to Share if Not Working

1. **Console logs:**
   - Screenshot of all console logs
   - Especially the 🔍 logs

2. **Network tab:**
   - F12 → Network tab
   - Find request to `/api/products/60`
   - Click it → Preview tab
   - Screenshot the response

3. **Database query result:**
   ```sql
   SELECT id, name, brand, material FROM products WHERE id = 60;
   ```

4. **Backend logs:**
   - Check backend console window
   - Any errors when loading product?

## Expected Behavior

### ✅ Success
- Console shows brand and material values
- "Detail Produk" section appears
- Brand and Material displayed correctly
- Gray box with 2-column grid layout

### ❌ Not Working
- Console shows undefined or empty
- No "Detail Produk" section
- Section appears but empty
- Data in console but not in UI

## Quick Fix Commands

### Restart Everything
```bash
# Backend
RESTART_BACKEND_FIX2.bat

# Frontend (in new terminal)
cd frontend
npm run dev
```

### Update Database
```sql
-- Update Shirt Eiger
UPDATE products 
SET brand = 'Eiger', material = 'Cotton' 
WHERE id = 60;

-- Check all products with brand/material
SELECT id, name, brand, material 
FROM products 
WHERE brand IS NOT NULL OR material IS NOT NULL;
```

### Clear Browser Cache
1. F12 → Application tab
2. Clear storage
3. Hard refresh (Ctrl+Shift+R)

## Next Steps

After brand/material display is working:
1. ✅ Verify on multiple products
2. ✅ Test products without brand/material (should not show section)
3. ✅ Test on mobile view
4. ✅ Move to product creation testing
