# Product Variants System - Complete Guide

## Overview

Sistem variant lengkap untuk e-commerce fashion yang mendukung:
- Multiple variants per product (size, color, material, dll)
- Stock tracking per variant
- Price override per variant
- Multi-image per variant
- Stock reservation system
- Low stock alerts

## Database Schema

### Tables Created

1. **product_variants** - Main variant data
   - SKU, size, color, material, pattern, fit, sleeve
   - Price override (optional)
   - Stock quantity per variant
   - Reserved stock tracking
   - Active/inactive status

2. **variant_images** - Images per variant
   - Multiple images per variant
   - Primary image designation
   - Position ordering

3. **stock_reservations** - Temporary stock holds
   - 15-minute timeout
   - Prevents overselling
   - Auto-cleanup of expired reservations

4. **variant_attributes** - Configurable attributes
   - Pre-populated: size, color, material, pattern, fit, sleeve

## API Endpoints

### Public Endpoints (Customer)

```bash
# Get product variants
GET /api/products/:id/variants

# Get product with variants and price range
GET /api/products/:id/with-variants

# Get available options (sizes, colors)
GET /api/products/:id/options

# Find specific variant
POST /api/products/variants/find
Body: { "product_id": 1, "size": "M", "color": "Black" }

# Get variant by ID
GET /api/variants/:id

# Get variant by SKU
GET /api/variants/sku/:sku

# Check stock availability
POST /api/variants/check-availability
Body: { "variant_id": 1, "quantity": 2 }

# Get variant attributes
GET /api/variants/attributes
```

### Admin Endpoints

```bash
# Create single variant
POST /api/admin/variants
Body: {
  "product_id": 1,
  "size": "M",
  "color": "Black",
  "color_hex": "#000000",
  "stock_quantity": 50,
  "price": 450000,
  "is_active": true
}

# Bulk generate variants (size × color matrix)
POST /api/admin/variants/bulk-generate
Body: {
  "product_id": 1,
  "sizes": ["S", "M", "L", "XL"],
  "colors": ["Black", "White", "Navy"],
  "base_price": 400000,
  "stock_per_variant": 10
}

# Update variant
PUT /api/admin/variants/:id
Body: { "stock_quantity": 25, "is_active": true }

# Delete variant
DELETE /api/admin/variants/:id

# Update stock
PUT /api/admin/variants/stock/:variantId
Body: { "quantity": 30 }

# Adjust stock (add/subtract)
POST /api/admin/variants/stock/:variantId/adjust
Body: { "delta": -5 }

# Get low stock variants
GET /api/admin/variants/low-stock

# Get stock summary for product
GET /api/admin/variants/stock-summary/:productId

# Add variant image
POST /api/admin/variants/images
Body: {
  "variant_id": 1,
  "image_url": "https://...",
  "is_primary": true
}

# Delete variant image
DELETE /api/admin/variants/images/:imageId

# Set primary image
POST /api/admin/variants/images/:variantId/primary
Body: { "image_id": 5 }

# Reorder images
POST /api/admin/variants/images/:variantId/reorder
Body: { "image_ids": [5, 3, 7, 2] }
```

## Admin UI Usage

### 1. Access Variant Manager

```
/admin/products/edit/:id → Tab "Variants & Stock"
```

### 2. Bulk Generate Variants

1. Click "Bulk Generate" button
2. Enter sizes (comma-separated): `S, M, L, XL`
3. Enter colors (comma-separated): `Black, White, Navy, Red`
4. Set stock per variant: `10`
5. Set base price (optional): `400000`
6. Click "Generate" → Creates 16 variants (4 sizes × 4 colors)

### 3. Manual Variant Creation

1. Click "Add Variant"
2. Fill form:
   - SKU (auto-generated if empty)
   - Size, Color, Color Hex
   - Price (overrides product price)
   - Stock quantity
   - Low stock threshold
3. Click "Create Variant"

### 4. Stock Management

- **Quick Update**: Edit stock directly in table
- **Bulk Adjust**: Use adjust endpoint for inventory changes
- **Low Stock Alerts**: View at `/admin/variants`

### 5. Image Management

- Upload images per variant
- Set primary image
- Reorder images by drag-drop (coming soon)

## Customer UI Usage

### Product Detail Page

Variant selector automatically appears when product has variants:

1. **Size Selector** - Buttons for available sizes
2. **Color Selector** - Color swatches with hex codes
3. **Dynamic Price** - Updates when variant selected
4. **Stock Display** - Shows available stock for selected variant
5. **Disabled States** - Out-of-stock variants are disabled/crossed

### Features

- Price range display if variants have different prices
- Real-time stock availability
- Visual feedback for selection
- Variant details (SKU, material, pattern)

## Stock Reservation System

### How It Works

1. **Add to Cart** → Reserves stock for 15 minutes
2. **Checkout** → Converts reservation to order
3. **Timeout** → Auto-releases after 15 minutes
4. **Cancel** → Manually release reservation

### Database Functions

```sql
-- Reserve stock
SELECT reserve_stock(variant_id, customer_id, session_id, quantity, 15);

-- Complete reservation (on order)
SELECT complete_reservation(reservation_id, order_id);

-- Cancel reservation
SELECT cancel_reservation(reservation_id);

-- Get available stock (total - reserved)
SELECT get_available_stock(variant_id);

-- Clean expired reservations
SELECT clean_expired_reservations();
```

## Integration with Cart & Checkout

### Cart Items

```typescript
{
  product_id: 1,
  variant_id: 5,  // NEW
  quantity: 2,
  variant_sku: "JACKET-M-BLACK",  // NEW
  variant_attributes: {  // NEW
    size: "M",
    color: "Black"
  }
}
```

### Order Items

```typescript
{
  product_id: 1,
  variant_id: 5,  // NEW
  variant_sku: "JACKET-M-BLACK",  // NEW
  variant_name: "M - Black",  // NEW
  price: 450000,  // Variant price or product price
  quantity: 2
}
```

## Best Practices

### 1. SKU Generation

- Auto-generated: `PRODUCT-SIZE-COLOR`
- Example: `CLASSIC-DENIM-JACKET-M-BLACK`
- Sanitized (uppercase, no special chars)
- Unique constraint enforced

### 2. Stock Management

- Always use `available_stock` (not `stock_quantity`)
- Available = Total - Reserved
- Set low stock threshold per variant (default: 5)
- Monitor low stock alerts daily

### 3. Pricing Strategy

- Set base price on product
- Override per variant if needed
- Display price range on listing
- Show specific price on detail

### 4. Image Strategy

- Upload product images first
- Add variant-specific images for colors
- Set primary image per variant
- Fallback to product images if no variant images

### 5. Variant Combinations

- Unique constraint: (product_id, size, color)
- Can't have duplicate size+color combo
- Use NULL for optional attributes
- Position field for custom ordering

## Testing Checklist

### Backend Tests

- [ ] Create variant with auto-generated SKU
- [ ] Bulk generate 12 variants (3 sizes × 4 colors)
- [ ] Update variant stock
- [ ] Reserve stock (check available_stock decreases)
- [ ] Complete reservation (check stock_quantity decreases)
- [ ] Cancel reservation (check available_stock increases)
- [ ] Get low stock variants
- [ ] Delete variant (should fail if has orders)

### Frontend Tests

- [ ] Admin: Bulk generate variants
- [ ] Admin: Edit variant stock
- [ ] Admin: View low stock alerts
- [ ] Customer: See variant selector
- [ ] Customer: Select size and color
- [ ] Customer: See price update
- [ ] Customer: See stock availability
- [ ] Customer: Disabled out-of-stock variants
- [ ] Cart: Shows variant details
- [ ] Checkout: Reserves stock

## Troubleshooting

### Issue: Variants not showing on product detail

**Solution:**
1. Check if product has variants: `GET /api/products/:id/variants`
2. Check if variants are active: `is_active = true`
3. Check browser console for API errors

### Issue: Stock overselling

**Solution:**
1. Use `get_available_stock()` function
2. Enable stock reservation system
3. Run `clean_expired_reservations()` periodically

### Issue: Duplicate SKU error

**Solution:**
1. Let system auto-generate SKU
2. Or ensure manual SKU is unique
3. Check existing SKUs: `SELECT sku FROM product_variants`

### Issue: Price not updating

**Solution:**
1. Check if variant has price override
2. Verify `selectedVariant` state in React
3. Check `getCurrentPrice()` function logic

## Migration Guide

### From Simple Product to Variants

1. **Backup database**
2. **Run migration**: `migrate_product_variants.bat`
3. **Create default variant** for existing products:
   ```sql
   INSERT INTO product_variants (product_id, sku, variant_name, stock_quantity, is_active, is_default)
   SELECT id, CONCAT('PROD-', id), 'Default', stock, true, true
   FROM products;
   ```
4. **Update cart items** to use variant_id
5. **Test checkout flow**

## Performance Optimization

### Indexes Created

```sql
CREATE INDEX idx_variants_product ON product_variants(product_id);
CREATE INDEX idx_variants_sku ON product_variants(sku);
CREATE INDEX idx_variants_active ON product_variants(is_active);
CREATE INDEX idx_variants_stock ON product_variants(stock_quantity);
CREATE INDEX idx_variant_images_variant ON variant_images(variant_id);
```

### Query Optimization

- Use `get_available_stock()` function (cached calculation)
- Fetch variants with product in single query
- Lazy load variant images
- Cache variant attributes

## Future Enhancements

- [ ] Variant import/export (CSV)
- [ ] Variant templates
- [ ] Bulk stock adjustment
- [ ] Variant analytics
- [ ] Size chart per category
- [ ] Color palette management
- [ ] Variant bundles/combos
- [ ] Pre-order variants
- [ ] Variant reviews

## Support

For issues or questions:
1. Check API logs: Backend console
2. Check browser console: Network tab
3. Verify database: `SELECT * FROM product_variants WHERE product_id = X`
4. Test API directly: Use Postman or curl

---

**System Status**: ✅ Production Ready
**Last Updated**: January 2026
**Version**: 1.0.0
