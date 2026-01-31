# Implementasi Filter Ukuran (Size Filter)

## 🎯 Tujuan

Membuat filter ukuran berfungsi dengan benar - ketika user klik "L", hanya produk yang punya variant size "L" yang muncul.

## 🔴 Masalah Sebelumnya

- Filter ukuran tidak berfungsi
- Semua produk tetap muncul meskipun sudah pilih size
- ProductResponse tidak include available sizes dari variants

## ✅ Solusi yang Diimplementasikan

### 1. Backend Changes

#### A. Update ProductResponse DTO
**File:** `backend/dto/dto.go`

```go
type ProductResponse struct {
    // ... existing fields
    AvailableSizes []string `json:"available_sizes,omitempty"` // NEW
}
```

#### B. Update Product Service
**File:** `backend/service/product_service.go`

**Changes:**
1. Add `variantRepo` to service struct
2. Update `NewProductService` to accept variant repository
3. Update `toProductResponse` to fetch available sizes from variants

```go
// Get available sizes from active variants
variants, err := s.variantRepo.GetByProductID(p.ID)
if err == nil && len(variants) > 0 {
    sizeMap := make(map[string]bool)
    for _, v := range variants {
        if v.IsActive && v.Size != nil && *v.Size != "" && v.StockQuantity > 0 {
            sizeMap[*v.Size] = true
        }
    }
    
    // Convert map to sorted slice
    var sizes []string
    sizeOrder := []string{"XS", "S", "M", "L", "XL", "XXL"}
    for _, size := range sizeOrder {
        if sizeMap[size] {
            sizes = append(sizes, size)
        }
    }
    response.AvailableSizes = sizes
}
```

**Logic:**
- Fetch all variants for product
- Filter only active variants with stock > 0
- Extract unique sizes
- Sort by standard size order (XS, S, M, L, XL, XXL)

#### C. Update Routes
**File:** `backend/routes/routes.go`

```go
productService := service.NewProductService(productRepo, variantRepo)
```

### 2. Frontend Changes

#### A. Update Product Type
**File:** `frontend/src/types/index.ts`

```typescript
export interface Product {
    // ... existing fields
    available_sizes?: string[]; // NEW
}
```

#### B. Update Filter Logic
**File:** `frontend/src/components/CategoryPage.tsx`

```typescript
// Apply size filter - only show products that have the selected sizes
if (filters.sizes.length > 0) {
  result = result.filter((p) => {
    // If product has available_sizes, check if any selected size is available
    if (p.available_sizes && p.available_sizes.length > 0) {
      return filters.sizes.some((selectedSize) =>
        p.available_sizes!.includes(selectedSize)
      );
    }
    // If product doesn't have available_sizes, don't show it when size filter is active
    return false;
  });
}
```

**Logic:**
- If user selects sizes (e.g., "L", "XL")
- Filter products to only show those with available_sizes containing at least one selected size
- Products without variants (no available_sizes) are hidden when size filter is active

## 📊 Data Flow

```
1. User clicks size "L"
   ↓
2. Frontend filters products where available_sizes includes "L"
   ↓
3. Only products with L variants (stock > 0) are shown
```

## 🧪 Testing

### Backend Test

```bash
# Start backend
cd backend
./zavera_size_filter.exe

# Test API
curl http://localhost:8080/products?category=pria
```

**Expected Response:**
```json
[
  {
    "id": 46,
    "name": "Hip Hop Baggy Jeans",
    "available_sizes": ["M", "L", "XL"]  // ✅ NEW FIELD
  },
  {
    "id": 47,
    "name": "Hip Hop Baggy Jeans 22",
    "available_sizes": ["M", "L", "XL"]  // ✅ NEW FIELD
  },
  {
    "id": 1,
    "name": "Minimalist Cotton Tee",
    "available_sizes": []  // No variants yet
  }
]
```

### Frontend Test

1. **Buka halaman PRIA:** `http://localhost:3000/pria`

2. **Test Filter L:**
   - Klik size "L"
   - Produk yang muncul:
     - ✅ Hip Hop Baggy Jeans (punya variant L)
     - ✅ Hip Hop Baggy Jeans 22 (punya variant L)
     - ✅ Jacket Parasut 22 (punya variant L)
   - Produk yang TIDAK muncul:
     - ❌ Minimalist Cotton Tee (tidak punya variant)
     - ❌ Jacket Parasut (hanya punya XL)

3. **Test Filter M:**
   - Klik size "M"
   - Produk yang muncul:
     - ✅ Hip Hop Baggy Jeans (punya variant M)
     - ✅ Hip Hop Baggy Jeans 22 (punya variant M)
     - ✅ Jacket Parasut 22 (punya variant M)

4. **Test Multiple Sizes:**
   - Klik "M" dan "L"
   - Produk yang muncul: semua yang punya M ATAU L

5. **Clear Filter:**
   - Klik "Hapus Semua"
   - Semua produk muncul kembali

## 📋 Current Product Variants Status

### Products WITH Variants (Size Filter Works):
| Product | Available Sizes |
|---------|----------------|
| Hip Hop Baggy Jeans | M, L, XL |
| Hip Hop Baggy Jeans 22 | M, L, XL |
| Jacket Parasut | XL |
| Jacket Parasut 22 | M, L, XL |

### Products WITHOUT Variants (Hidden When Size Filter Active):
- Minimalist Cotton Tee
- Classic Denim Jacket
- Tailored Trousers
- Premium Hoodie
- Slim Fit Shirt
- Casual Blazer
- Premium Wool Suit
- Leather Oxford Shoes
- Merino Wool Sweater
- Chino Pants
- Denim Jacket
- Mens Denim Jeans
- Jacket Boomber

## 🔧 Next Steps (Optional)

### Generate Variants for All Products

Untuk membuat semua produk bisa difilter by size, perlu generate variants:

```sql
-- Example: Generate variants for Minimalist Cotton Tee
INSERT INTO product_variants (
    product_id, sku, variant_name, size, color,
    price, stock_quantity, is_active, is_default
) VALUES
(1, 'MCT-S-BLACK', 'S - Black', 'S', 'Black', 299000, 10, true, false),
(1, 'MCT-M-BLACK', 'M - Black', 'M', 'Black', 299000, 15, true, true),
(1, 'MCT-L-BLACK', 'L - Black', 'L', 'Black', 299000, 12, true, false),
(1, 'MCT-XL-BLACK', 'XL - Black', 'XL', 'Black', 299000, 8, true, false);
```

Atau gunakan admin panel untuk generate variants.

## 📁 Files Modified

### Backend:
1. ✅ `backend/dto/dto.go` - Add AvailableSizes field
2. ✅ `backend/service/product_service.go` - Fetch available sizes from variants
3. ✅ `backend/routes/routes.go` - Pass variant repo to product service

### Frontend:
1. ✅ `frontend/src/types/index.ts` - Add available_sizes to Product type
2. ✅ `frontend/src/components/CategoryPage.tsx` - Implement size filtering logic

## ✅ Summary

**Before:**
- Filter size tidak berfungsi ❌
- Semua produk tetap muncul ❌
- ProductResponse tidak include sizes ❌

**After:**
- Filter size berfungsi dengan benar ✅
- Hanya produk dengan size yang dipilih yang muncul ✅
- ProductResponse include available_sizes dari variants ✅
- Products tanpa variants hidden saat size filter active ✅

**Build Status:**
- Backend: ✅ Success (`zavera_size_filter.exe`)
- Frontend: ✅ Success

**Ready to Test!** 🚀

---

**Note:** Produk yang belum punya variants tidak akan muncul saat size filter aktif. Ini by design untuk memastikan user hanya melihat produk yang benar-benar tersedia dalam size yang mereka pilih.
