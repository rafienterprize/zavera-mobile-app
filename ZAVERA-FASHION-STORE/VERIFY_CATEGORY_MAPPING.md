# Verify Category Mapping

## Test: Admin Input → Database → Client Display

### Test Case 1: Wanita - Atasan

**Admin Input:**
1. Category: Wanita
2. Subcategory: Atasan

**Expected Database:**
```sql
category: wanita
subcategory: Tops
```

**Expected Client:**
- Filter shows: "Atasan"
- Product appears when filtering by "Atasan"

### Test Case 2: Pria - Kemeja

**Admin Input:**
1. Category: Pria
2. Subcategory: Kemeja

**Expected Database:**
```sql
category: pria
subcategory: Shirts
```

**Expected Client:**
- Filter shows: "Kemeja"
- Product appears when filtering by "Kemeja"

### Test Case 3: Pria - Celana

**Admin Input:**
1. Category: Pria
2. Subcategory: Celana

**Expected Database:**
```sql
category: pria
subcategory: Bottoms
```

**Expected Client:**
- Filter shows: "Celana"
- Product appears when filtering by "Celana"

## Complete Mapping Reference

### Wanita
| Admin Label | Database Value | Client Label |
|-------------|----------------|--------------|
| Dress | Dress | Dress |
| Atasan | Tops | Atasan |
| Bawahan | Bottoms | Bawahan |
| Outerwear | Outerwear | Outerwear |
| Aksesoris | Accessories | Aksesoris |

### Pria
| Admin Label | Database Value | Client Label |
|-------------|----------------|--------------|
| Atasan | Tops | Atasan |
| Kemeja | Shirts | Kemeja |
| Celana | Bottoms | Celana |
| Jaket | Outerwear | Jaket |
| Jas | Suits | Jas |
| Sepatu | Footwear | Sepatu |

### Anak
| Admin Label | Database Value | Client Label |
|-------------|----------------|--------------|
| Anak Laki-laki | Boys | Anak Laki-laki |
| Anak Perempuan | Girls | Anak Perempuan |
| Bayi | Baby | Bayi |
| Sepatu | Footwear | Sepatu |

### Sports
| Admin Label | Database Value | Client Label |
|-------------|----------------|--------------|
| Pakaian Olahraga | Activewear | Pakaian Olahraga |
| Sepatu | Footwear | Sepatu |
| Jaket | Outerwear | Jaket |
| Aksesoris | Accessories | Aksesoris |

### Luxury
| Admin Label | Database Value | Client Label |
|-------------|----------------|--------------|
| Aksesoris | Accessories | Aksesoris |
| Outerwear | Outerwear | Outerwear |

### Beauty
| Admin Label | Database Value | Client Label |
|-------------|----------------|--------------|
| Perawatan Kulit | Skincare | Perawatan Kulit |
| Makeup | Makeup | Makeup |
| Parfum | Fragrance | Parfum |

## How to Test

### Step 1: Create Test Product
```
Admin Panel → Add Product
- Name: Test Atasan Pria
- Category: Pria
- Subcategory: Atasan
- Price: 100000
- Add image & variant
- Create
```

### Step 2: Check Database
```sql
SELECT id, name, category, subcategory 
FROM products 
WHERE name = 'Test Atasan Pria';
```

**Expected:**
```
category: pria
subcategory: Tops  ← Database value (English)
```

### Step 3: Check Client
```
1. Go to: http://localhost:3000/pria
2. Filter by: Atasan
3. Product "Test Atasan Pria" should appear
```

## Verification SQL

```sql
-- Check all products with their categories
SELECT 
  id, 
  name, 
  category, 
  subcategory,
  CASE 
    WHEN category = 'pria' AND subcategory = 'Tops' THEN 'Atasan'
    WHEN category = 'pria' AND subcategory = 'Shirts' THEN 'Kemeja'
    WHEN category = 'pria' AND subcategory = 'Bottoms' THEN 'Celana'
    WHEN category = 'pria' AND subcategory = 'Outerwear' THEN 'Jaket'
    WHEN category = 'pria' AND subcategory = 'Suits' THEN 'Jas'
    WHEN category = 'pria' AND subcategory = 'Footwear' THEN 'Sepatu'
    ELSE subcategory
  END as client_display
FROM products 
WHERE category = 'pria'
ORDER BY id DESC;
```

## Common Issues

### Issue 1: Product tidak muncul di filter
**Cause:** Database value tidak match dengan mapping

**Check:**
```sql
SELECT subcategory FROM products WHERE name = 'Product Name';
```

**Fix:**
```sql
-- Example: Fix "Shirt" to "Shirts"
UPDATE products SET subcategory = 'Shirts' WHERE subcategory = 'Shirt';
```

### Issue 2: Filter menampilkan English label
**Cause:** Client mapping tidak ada untuk database value

**Check FilterPanel.tsx:**
```typescript
const SUBCATEGORY_MAPPING = {
  pria: {
    "Atasan": "Tops",  // ← Must have this mapping
    // ...
  }
}
```

### Issue 3: Admin dropdown masih English
**Cause:** Frontend belum di-refresh atau cache

**Fix:**
```bash
# Hard refresh
Ctrl+Shift+R

# Or restart frontend
cd frontend
npm run dev
```

## Success Criteria

✅ Admin dropdown shows Indonesian labels
✅ Database stores English values
✅ Client filter shows Indonesian labels
✅ Products appear in correct filter
✅ No mismatch between admin input and client display

## Test All Categories

Run this test for each category:

```bash
# 1. Create product in admin with specific subcategory
# 2. Check database value
psql -U postgres -d zavera_db -c "SELECT id, name, category, subcategory FROM products ORDER BY id DESC LIMIT 1;"

# 3. Check client display
# Open browser and filter by that subcategory

# 4. Verify product appears
```

## Summary

**Goal:** Admin input "Atasan" → Database "Tops" → Client display "Atasan"

**Mapping:** Admin Label → Database Value → Client Label (same as Admin)

**Files:**
- Admin: `frontend/src/app/admin/products/add/page.tsx`
- Client: `frontend/src/components/FilterPanel.tsx`
- Both must have EXACT same mapping!
