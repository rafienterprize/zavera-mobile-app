# Stock System Explanation

## Overview
The Zavera e-commerce platform uses a **variant-based stock system** similar to major platforms like Tokopedia, Shopee, and Lazada. This document explains how stock works for products with and without variants.

## Two Types of Products

### 1. Simple Products (No Variants)
- Stock is stored at the **product level**
- `product.stock` contains the actual inventory count
- Example: A basic t-shirt with no size/color options

### 2. Variant Products (With Variants)
- Stock is stored at the **variant level**
- `product.stock = 0` (this is normal and expected)
- Each variant has its own `stock_quantity` and `available_stock`
- Example: A t-shirt with multiple sizes (S, M, L, XL) and colors (Red, Blue, Black)

## Why Product Stock Shows 0

When you create a product with variants:
1. The product itself has `stock = 0`
2. Each variant (e.g., "Size M - Red") has its own stock count
3. This is **by design** - the product is just a container for variants

**Example:**
```
Product: "Premium T-Shirt" (stock = 0)
├── Variant: Size M - Red (stock = 10)
├── Variant: Size M - Blue (stock = 15)
├── Variant: Size L - Red (stock = 8)
└── Variant: Size L - Blue (stock = 12)

Total Available Stock: 45 items
```

## Admin Dashboard Display

### Current Behavior
- Products with variants show "Variants" label instead of stock number
- Products without variants show actual stock count
- Low stock warning (< 10) shows amber color
- Out of stock (0) shows red color for simple products

### Stock Column Logic
```typescript
if (product.stock === 0) {
  // This is a variant-based product
  display: "Variants" icon
} else {
  // This is a simple product
  display: actual stock number
}
```

## Customer Product Page

### SOLD OUT Overlay Logic

The product detail page shows different overlays based on the situation:

1. **Simple Product - Out of Stock**
   - Condition: `variants.length === 0 && stock === 0`
   - Display: "SOLD OUT" overlay

2. **Variant Product - No Selection**
   - Condition: `variants.length > 0 && !selectedVariant`
   - Display: "Pilih ukuran dan warna" (Select size and color)

3. **Variant Product - Selected Variant Out of Stock**
   - Condition: `variants.length > 0 && selectedVariant && variant.stock === 0`
   - Display: "SOLD OUT" overlay

4. **Stock Available**
   - No overlay, product can be added to cart

### Stock Display
- Shows available stock count below quantity selector
- Low stock warning (< 10): "X item tersedia - Segera habis!"
- Out of stock: "Stok habis" in red

## How to Check Variant Stock

### Via Admin Panel
1. Go to Admin → Products
2. Click "Edit" on a product
3. Go to "Variants & Stock" tab
4. See all variants with their individual stock counts

### Via API
```bash
GET /api/products/{productId}/variants
```

Response:
```json
[
  {
    "id": 1,
    "product_id": 46,
    "size": "M",
    "color": "Red",
    "stock_quantity": 10,
    "available_stock": 10,
    "is_active": true
  },
  ...
]
```

## Stock Management

### Update Variant Stock
1. **Via Admin UI**: Edit product → Variants & Stock tab → Edit variant
2. **Via API**: `PUT /api/admin/variants/stock/{variantId}`

### Bulk Generate Variants
When you bulk generate variants, you can set:
- `stock_per_variant`: Stock count for each generated variant
- Example: Generate 12 variants (3 sizes × 4 colors) with 10 stock each = 120 total items

## Common Scenarios

### Scenario 1: Product Shows SOLD OUT but Has Stock
**Cause**: Product has variants but customer hasn't selected size/color yet

**Solution**: Customer needs to select a variant first. The overlay now shows "Pilih ukuran dan warna" instead of "SOLD OUT"

### Scenario 2: Admin Dashboard Shows Stock = 0
**Cause**: Product uses variants for stock management

**Solution**: This is normal. Click "Edit" to see individual variant stocks. The dashboard now shows "Variants" label for these products.

### Scenario 3: Created Product with Stock but Shows 0
**Cause**: You set stock during product creation, but then added variants

**Solution**: When variants are added, stock moves to variant level. Set stock on each variant instead.

## Best Practices

1. **Decide Early**: Choose whether product will have variants before setting stock
2. **Consistent Approach**: Either use product-level stock OR variant-level stock, not both
3. **Bulk Generate**: Use bulk generate feature to create all size/color combinations at once
4. **Stock Monitoring**: Check "Variants & Stock" tab regularly for low stock alerts
5. **Customer Experience**: Ensure all active variants have stock > 0

## Technical Details

### Database Schema
```sql
-- Product table
products (
  id, name, price, stock, -- stock = 0 for variant products
  ...
)

-- Variant table
product_variants (
  id, product_id, size, color,
  stock_quantity,      -- Physical inventory
  reserved_stock,      -- Temporarily reserved (in carts)
  available_stock,     -- stock_quantity - reserved_stock
  ...
)
```

### Stock Calculation
```
available_stock = stock_quantity - reserved_stock
```

When customer adds to cart:
- `reserved_stock` increases
- `available_stock` decreases
- `stock_quantity` stays same until order is completed

## Troubleshooting

### Issue: Variant stock not updating
**Check**: 
1. Variant ID is correct
2. Using correct API endpoint (`/api/admin/variants/stock/{id}`)
3. Auth token is valid

### Issue: SOLD OUT showing incorrectly
**Check**:
1. Browser console for variant fetch errors
2. Variant `is_active = true`
3. Variant `available_stock > 0`

### Issue: Can't add to cart
**Check**:
1. Variant is selected (for variant products)
2. Quantity <= available_stock
3. User is logged in

## Files Modified

### Frontend
- `frontend/src/app/product/[id]/page.tsx` - SOLD OUT logic and variant selection
- `frontend/src/app/admin/products/page.tsx` - Admin stock display
- `frontend/src/lib/variantApi.ts` - API response handling

### Backend
- `backend/handler/variant_handler.go` - Variant endpoints
- `backend/service/variant_service.go` - Stock management logic
- `backend/models/product_variant.go` - Variant model

## Compiled Binary
- `backend/zavera_stock_fix.exe` - Latest build with stock fixes
