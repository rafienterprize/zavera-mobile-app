# 🎨 Product Variants System

Complete variant management system for fashion e-commerce.

## 🚀 Quick Start

```bash
# 1. Migrate database
migrate_product_variants.bat

# 2. Run backend
cd backend
.\zavera_variants.exe

# 3. Test
# Admin: http://localhost:3000/admin/products/edit/1
# Customer: http://localhost:3000/product/1
```

## 📖 Documentation

- **[Quick Start Guide](VARIANT_QUICK_START.md)** - 5-minute setup
- **[Complete Guide](VARIANT_SYSTEM_GUIDE.md)** - Full documentation
- **[Implementation Summary](VARIANT_IMPLEMENTATION_COMPLETE.md)** - What's included

## ✨ Features

### Admin
- ✅ Bulk generate variants (size × color matrix)
- ✅ Manual variant creation
- ✅ Stock management per variant
- ✅ Low stock alerts
- ✅ Multi-image per variant
- ✅ Price override per variant

### Customer
- ✅ Dynamic size selector
- ✅ Color swatches with hex codes
- ✅ Real-time price updates
- ✅ Stock availability display
- ✅ Disabled out-of-stock options

### System
- ✅ Stock reservation (15-min timeout)
- ✅ Transaction-safe operations
- ✅ Auto-generated SKUs
- ✅ Concurrent purchase protection

## 🎯 Usage

### Admin: Bulk Generate

1. Edit product → Tab "Variants & Stock"
2. Click "Bulk Generate"
3. Enter: `S, M, L, XL` (sizes)
4. Enter: `Black, White, Navy` (colors)
5. Set stock: `10` per variant
6. Generate → 12 variants created!

### Customer: Select Variant

1. Open product detail
2. Click size button
3. Click color swatch
4. Price & stock auto-update
5. Add to cart

## 📊 API Endpoints

```bash
# Public
GET  /api/products/:id/variants
GET  /api/products/:id/with-variants
POST /api/variants/check-availability

# Admin
POST /api/admin/variants/bulk-generate
PUT  /api/admin/variants/stock/:id
GET  /api/admin/variants/low-stock
```

## 🗄️ Database

**Tables:**
- `product_variants` - Variant data
- `variant_images` - Images per variant
- `stock_reservations` - Temporary holds
- `variant_attributes` - Configurable attributes

**Functions:**
- `reserve_stock()` - Reserve with timeout
- `complete_reservation()` - Convert to order
- `get_available_stock()` - Total - reserved

## 🧪 Test

```bash
# Test bulk generate
test_bulk_generate_variants.bat

# Test API
curl http://localhost:8080/api/products/1/variants
```

## 📁 Files

### Backend
- `backend/models/product_variant.go`
- `backend/repository/variant_repository.go`
- `backend/service/variant_service.go`
- `backend/handler/variant_handler.go`
- `backend/dto/variant_dto.go`

### Frontend
- `frontend/src/types/variant.ts`
- `frontend/src/lib/variantApi.ts`
- `frontend/src/components/VariantSelector.tsx`
- `frontend/src/components/admin/VariantManager.tsx`
- `frontend/src/app/admin/variants/page.tsx`

### Database
- `database/migrate_product_variants.sql`

## 🎓 Examples

### Bulk Generate Request
```json
{
  "product_id": 1,
  "sizes": ["S", "M", "L", "XL"],
  "colors": ["Black", "White", "Navy"],
  "base_price": 400000,
  "stock_per_variant": 10
}
```

### Create Variant Request
```json
{
  "product_id": 1,
  "size": "M",
  "color": "Black",
  "color_hex": "#000000",
  "stock_quantity": 50,
  "price": 450000,
  "is_active": true
}
```

## 🔧 Configuration

No additional configuration needed. Uses existing database connection.

## 📈 Performance

- Supports 1000+ variants per product
- Transaction-safe stock operations
- Indexed queries for fast lookups
- Auto-cleanup expired reservations

## 🐛 Troubleshooting

**Variants not showing?**
- Check `is_active = true`
- Check API: `/api/products/:id/variants`

**Stock issues?**
- Use `get_available_stock()` function
- Check reservations table

**Duplicate SKU?**
- Leave SKU empty for auto-generation

## ✅ Status

**Implementation**: 100% Complete
**Testing**: Ready
**Production**: Ready

---

**Need help?** Check the [Complete Guide](VARIANT_SYSTEM_GUIDE.md)
