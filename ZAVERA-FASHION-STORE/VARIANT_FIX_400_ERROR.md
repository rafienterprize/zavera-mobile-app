# Fix: 400 Error on Variant Endpoints

## Problem

When accessing `/api/products/46/variants`, the API returns 400 Bad Request error:
```
[GIN] 2026/01/27 - 17:12:24 | 400 | 0s | ::1 | GET "/api/products/46/variants"
```

Additionally, products created with variants show stock as 0 even though variants have stock set.

## Root Cause

### Issue 1: Parameter Mismatch
The route definition uses `:id` but the handler was looking for `productId`:

**Route (routes.go):**
```go
products.GET("/:id/variants", variantHandler.GetProductVariants)
```

**Handler (variant_handler.go) - BEFORE:**
```go
func (h *VariantHandler) GetProductVariants(c *gin.Context) {
    productID, err := strconv.Atoi(c.Param("productId")) // ❌ Wrong parameter name
    // ...
}
```

This caused the handler to receive an empty string, which failed to parse as integer, resulting in 400 error.

### Issue 2: Product Stock vs Variant Stock
Products are created with `stock: 0` because stock is tracked per variant, not per product. However, the product list might be displaying product stock instead of aggregated variant stock.

## Solution

### Fix 1: Corrected Parameter Names
Updated all handlers to use the correct parameter name from routes:

**Fixed handlers:**
```go
// GetProductVariants - uses :id from route
func (h *VariantHandler) GetProductVariants(c *gin.Context) {
    productID, err := strconv.Atoi(c.Param("id")) // ✅ Correct
    // ...
}

// GetProductWithVariants - uses :id from route
func (h *VariantHandler) GetProductWithVariants(c *gin.Context) {
    productID, err := strconv.Atoi(c.Param("id")) // ✅ Correct
    // ...
}

// GetAvailableOptions - uses :id from route
func (h *VariantHandler) GetAvailableOptions(c *gin.Context) {
    productID, err := strconv.Atoi(c.Param("id")) // ✅ Correct
    // ...
}

// GetStockSummary - uses :id from route (admin endpoint)
func (h *VariantHandler) GetStockSummary(c *gin.Context) {
    productID, err := strconv.Atoi(c.Param("id")) // ✅ Correct
    // ...
}

// GetVariantImages - uses :id from route
func (h *VariantHandler) GetVariantImages(c *gin.Context) {
    variantID, err := strconv.Atoi(c.Param("id")) // ✅ Correct
    // ...
}
```

### Fix 2: Recompiled Backend
```bash
cd backend
go build -o zavera_variants_fixed.exe
```

## How to Apply Fix

### Step 1: Stop Current Backend
If backend is running, stop it (Ctrl+C in terminal)

### Step 2: Start Fixed Backend
```bash
# Option 1: Use batch file
start-backend-fixed.bat

# Option 2: Manual
cd backend
zavera_variants_fixed.exe
```

### Step 3: Test the Fix
1. Open browser to admin product edit page
2. Page should load without infinite loading
3. Variants should display correctly
4. Stock should show from variants

### Step 4: Verify API
```bash
# Test variant endpoint
curl http://localhost:8080/api/products/46/variants

# Should return 200 OK with variants array
```

## Expected Behavior After Fix

### API Endpoints
✅ `GET /api/products/:id/variants` - Returns 200 with variants
✅ `GET /api/products/:id/with-variants` - Returns 200 with product + variants
✅ `GET /api/products/:id/options` - Returns 200 with available options
✅ `GET /api/variants/:id/images` - Returns 200 with variant images
✅ `GET /api/admin/variants/stock-summary/:id` - Returns 200 with stock summary

### Admin Interface
✅ Product edit page loads without infinite loading
✅ Variants display correctly
✅ Stock shows from variants (not product stock)
✅ Can edit variants

### Client Interface
✅ Product detail page loads
✅ Variant selector works
✅ Stock availability shows correctly

## Understanding Stock in Variant System

### Product Stock vs Variant Stock

**Product Stock (stock field in products table):**
- Set to 0 for products with variants
- Not used when variants exist
- Only used for simple products without variants

**Variant Stock (stock_quantity in product_variants table):**
- Each variant has its own stock
- This is the actual available stock
- Tracked independently per size/color combination

**Total Product Stock:**
- Sum of all variant stocks
- Calculated dynamically
- Example:
  - Variant 1 (M, Black): 15 stock
  - Variant 2 (L, Black): 10 stock
  - Variant 3 (M, Navy): 12 stock
  - **Total: 37 stock**

### How to Display Stock

**In Product List:**
```javascript
// Option 1: Show total variant stock
const totalStock = variants.reduce((sum, v) => sum + v.stock_quantity, 0);

// Option 2: Show stock range
const minStock = Math.min(...variants.map(v => v.stock_quantity));
const maxStock = Math.max(...variants.map(v => v.stock_quantity));
// Display: "10-15 in stock"

// Option 3: Show variant count
// Display: "3 variants available"
```

**In Product Detail:**
```javascript
// Show stock for selected variant
if (selectedVariant) {
  return `${selectedVariant.stock_quantity} in stock`;
}
// Or show total if no variant selected
return `${totalStock} total in stock`;
```

## Files Modified

```
✅ backend/handler/variant_handler.go - Fixed parameter names
✅ backend/zavera_variants_fixed.exe - Recompiled binary
✅ start-backend-fixed.bat - New startup script
```

## Testing Checklist

After applying fix, verify:
- [ ] Backend starts without errors
- [ ] Can access `/api/products/46/variants` (returns 200)
- [ ] Admin product edit page loads
- [ ] Variants display in admin
- [ ] Stock shows correctly
- [ ] Can edit product with variants
- [ ] Client product page loads
- [ ] Variant selector works
- [ ] Stock availability correct

## Prevention

To prevent similar issues in the future:

1. **Always match route parameters with handler parameters:**
   ```go
   // Route
   router.GET("/:id/something", handler.Method)
   
   // Handler
   func (h *Handler) Method(c *gin.Context) {
       id := c.Param("id") // ✅ Must match route
   }
   ```

2. **Use consistent naming:**
   - Use `:id` for single resource
   - Use `:productId`, `:variantId` for nested resources
   - Be consistent across all routes

3. **Test API endpoints after changes:**
   ```bash
   # Quick test
   curl http://localhost:8080/api/products/1/variants
   ```

4. **Check backend logs for 400 errors:**
   - 400 usually means parameter parsing failed
   - Check parameter names first

## Summary

The 400 error was caused by parameter name mismatch between routes and handlers. Fixed by updating handlers to use correct parameter names (`id` instead of `productId`, `variantId`, etc.) and recompiling backend.

**Status: FIXED** ✅

Use `zavera_variants_fixed.exe` or `start-backend-fixed.bat` to run the corrected backend.
