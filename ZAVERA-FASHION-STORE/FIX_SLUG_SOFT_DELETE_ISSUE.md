# Fix: Slug Uniqueness with Soft Delete

## Masalah

Ketika admin delete product, product di-soft delete (is_active = false) tapi slug masih ada di database. Ketika create product baru dengan nama yang sama, terjadi error "slug already exists" meskipun product sudah "dihapus".

### Scenario:
1. Admin create product "Shirt Eiger" → slug: "shirt-eiger"
2. Admin delete product → is_active = false (tapi slug masih "shirt-eiger")
3. Admin create product "Shirt Eiger" lagi → ERROR: slug already exists ❌

## Root Cause

Di `backend/service/admin_product_service.go`, slug uniqueness check tidak mempertimbangkan `is_active`:

```go
// OLD CODE - WRONG ❌
SELECT EXISTS(SELECT 1 FROM products WHERE slug = $1)
```

Query ini check semua products, termasuk yang sudah di-delete (is_active = false).

## Solution

Update query untuk hanya check active products:

```go
// NEW CODE - CORRECT ✅
SELECT EXISTS(SELECT 1 FROM products WHERE slug = $1 AND is_active = true)
```

Sekarang slug hanya di-check untuk products yang active. Jika product sudah di-delete, slug-nya bisa dipakai lagi.

## Changes Made

### File: `backend/service/admin_product_service.go`

**Before:**
```go
// Check slug uniqueness
var exists bool
err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM products WHERE slug = $1)", slug).Scan(&exists)
if err != nil {
    return nil, err
}
if exists {
    return nil, ErrDuplicateSlug
}
```

**After:**
```go
// Check slug uniqueness (only among active products)
var exists bool
err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM products WHERE slug = $1 AND is_active = true)", slug).Scan(&exists)
if err != nil {
    return nil, err
}
if exists {
    return nil, ErrDuplicateSlug
}
```

## Testing

### Test Case 1: Create → Delete → Create Same Name
1. Create product "Shirt Eiger"
2. Delete product (soft delete)
3. Create product "Shirt Eiger" lagi
4. **Expected:** ✅ Berhasil dibuat (tidak error)

### Test Case 2: Create → Create Same Name (Without Delete)
1. Create product "Shirt Eiger"
2. Create product "Shirt Eiger" lagi (tanpa delete)
3. **Expected:** ❌ Error "Produk Sudah Ada"

### Test Case 3: Multiple Soft Deletes
1. Create product "Shirt Eiger" → Delete
2. Create product "Shirt Eiger" → Delete
3. Create product "Shirt Eiger" → Delete
4. Create product "Shirt Eiger" lagi
5. **Expected:** ✅ Berhasil dibuat

## Restart Backend

Jalankan:
```
RESTART_BACKEND_NOW.bat
```

Atau manual:
```bash
cd backend
taskkill /F /IM zavera_brand_material_fix.exe
zavera_brand_material_fix.exe
```

## Database State

### Before Fix:
```sql
-- Products table
id | name         | slug         | is_active
1  | Shirt Eiger  | shirt-eiger  | false     (deleted)
-- Try create "Shirt Eiger" → ERROR ❌

-- Slug check query:
SELECT EXISTS(SELECT 1 FROM products WHERE slug = 'shirt-eiger')
-- Returns: true (found deleted product)
```

### After Fix:
```sql
-- Products table
id | name         | slug         | is_active
1  | Shirt Eiger  | shirt-eiger  | false     (deleted)
-- Try create "Shirt Eiger" → SUCCESS ✅

-- Slug check query:
SELECT EXISTS(SELECT 1 FROM products WHERE slug = 'shirt-eiger' AND is_active = true)
-- Returns: false (deleted product not counted)
```

## Alternative Solutions Considered

### Option 1: Hard Delete (Rejected)
Delete product permanently dari database.

**Pros:**
- No slug collision
- Cleaner database

**Cons:**
- ❌ Lose order history
- ❌ Lose sales data
- ❌ Cannot restore deleted products
- ❌ Break referential integrity

### Option 2: Append Timestamp to Slug (Rejected)
Generate slug dengan timestamp: "shirt-eiger-1738234567"

**Pros:**
- Always unique
- No collision

**Cons:**
- ❌ Ugly URLs
- ❌ Bad for SEO
- ❌ Confusing for users

### Option 3: Check Only Active Products (CHOSEN ✅)
Only check slug uniqueness among active products.

**Pros:**
- ✅ Clean URLs
- ✅ Can reuse names after delete
- ✅ Maintain order history
- ✅ Simple implementation

**Cons:**
- Multiple deleted products can have same slug (acceptable)

## Edge Cases Handled

### Case 1: Multiple Deleted Products with Same Slug
```sql
id | name         | slug         | is_active
1  | Shirt Eiger  | shirt-eiger  | false
2  | Shirt Eiger  | shirt-eiger  | false
3  | Shirt Eiger  | shirt-eiger  | false
```
**Result:** ✅ Can create new "Shirt Eiger" (slug check only looks at active)

### Case 2: Active + Deleted with Same Slug
```sql
id | name         | slug         | is_active
1  | Shirt Eiger  | shirt-eiger  | false
2  | Shirt Eiger  | shirt-eiger  | true
```
**Result:** ❌ Cannot create new "Shirt Eiger" (active product exists)

### Case 3: Restore Deleted Product
If we implement "restore" feature later:
```sql
-- Before restore
id | name         | slug         | is_active
1  | Shirt Eiger  | shirt-eiger  | false

-- After restore
id | name         | slug         | is_active
1  | Shirt Eiger  | shirt-eiger  | true
```
**Result:** ✅ Works fine (slug becomes active again)

## Impact on Other Features

### Orders
- ✅ No impact - orders reference product_id, not slug
- ✅ Order history preserved even after delete

### URLs
- ✅ No impact - deleted products not accessible anyway
- ✅ New product with same name gets same clean URL

### SEO
- ✅ Positive - can reuse good URLs
- ✅ No ugly timestamps in URLs

### Admin Panel
- ✅ Better UX - can recreate products after accidental delete
- ✅ No confusion with "slug already exists" errors

## Future Considerations

### If We Need True Uniqueness:
Add unique constraint on (slug, is_active):
```sql
CREATE UNIQUE INDEX idx_products_slug_active 
ON products(slug) 
WHERE is_active = true;
```

This ensures database-level enforcement of our business rule.

## Rollback Plan

If this causes issues, revert to old behavior:

```go
// Revert to checking all products
err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM products WHERE slug = $1)", slug).Scan(&exists)
```

Then rebuild and restart backend.

## Monitoring

After deployment, monitor:
- Product creation success rate
- Slug collision errors
- User complaints about "already exists" errors

Expected improvements:
- ✅ Fewer "slug already exists" errors
- ✅ Better admin UX
- ✅ More successful product creations

## Notes

- Soft delete is standard practice for e-commerce
- Preserves data integrity and history
- Allows for potential "restore" feature
- Better for analytics and reporting
