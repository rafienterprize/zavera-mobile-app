# Product Variants - Quick Start Guide

## 🚀 Setup (5 menit)

### 1. Migrasi Database
```bash
migrate_product_variants.bat
```

### 2. Build & Run Backend
```bash
cd backend
go build -o zavera_variants.exe
.\zavera_variants.exe
```

### 3. Run Frontend
```bash
cd frontend
npm run dev
```

## 📦 Cara Pakai - Admin

### Bulk Generate Variants (Tercepat)

1. Login sebagai admin
2. Buka `/admin/products`
3. Klik product → Edit
4. Tab "Variants & Stock"
5. Klik "Bulk Generate"
6. Isi:
   - Sizes: `S, M, L, XL`
   - Colors: `Black, White, Navy`
   - Stock: `10`
7. Klik "Generate" → Selesai! 12 variants dibuat

### Manual Add Variant

1. Tab "Variants & Stock"
2. Klik "Add Variant"
3. Isi form (SKU auto-generated)
4. Klik "Create"

### Update Stock

Langsung edit di table, tekan Enter.

### Low Stock Alerts

Buka `/admin/variants` untuk lihat produk yang stoknya menipis.

## 🛍️ Cara Pakai - Customer

### Product Detail

1. Buka product yang punya variants
2. Pilih size (button)
3. Pilih color (color swatch)
4. Harga otomatis update
5. Stock availability muncul
6. Add to cart

### Features

- ✅ Size selector dengan disabled state untuk out-of-stock
- ✅ Color swatches dengan hex code
- ✅ Dynamic price update
- ✅ Real-time stock display
- ✅ Price range di listing (jika variant beda harga)

## 🧪 Test Cepat

### Test API (Postman/curl)

```bash
# 1. Get variants
curl http://localhost:8080/api/products/1/variants

# 2. Get product with variants
curl http://localhost:8080/api/products/1/with-variants

# 3. Check availability
curl -X POST http://localhost:8080/api/variants/check-availability \
  -H "Content-Type: application/json" \
  -d '{"variant_id":1,"quantity":2}'
```

### Test UI

1. **Admin**: Bulk generate 12 variants → Cek table
2. **Customer**: Buka product detail → Pilih size & color
3. **Cart**: Add to cart → Cek variant details muncul

## 📊 Monitoring

### Low Stock

```
/admin/variants → Tab "Low Stock Alerts"
```

### Stock Summary per Product

```
/admin/products/edit/:id → Tab "Variants & Stock"
```

Lihat:
- Total stock
- Reserved stock
- Available stock

## ⚡ Tips

1. **Bulk Generate** lebih cepat dari manual
2. **SKU auto-generated** - biarkan kosong
3. **Price override** - isi jika beda dari base price
4. **Low stock threshold** - default 5, bisa diubah per variant
5. **Color hex** - gunakan untuk color swatches

## 🐛 Troubleshooting

**Variants tidak muncul?**
- Cek `is_active = true`
- Cek API: `/api/products/:id/variants`

**Stock overselling?**
- Sistem sudah pakai reservation
- Available stock = Total - Reserved

**Duplicate SKU error?**
- Biarkan SKU kosong untuk auto-generate
- Atau pastikan SKU unique

## 📝 Contoh Data

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

Hasil: 12 variants (4 sizes × 3 colors)

### Manual Variant

```json
{
  "product_id": 1,
  "size": "M",
  "color": "Black",
  "color_hex": "#000000",
  "stock_quantity": 50,
  "price": 450000,
  "low_stock_threshold": 5,
  "is_active": true
}
```

## ✅ Checklist Implementasi

- [x] Database migration
- [x] Backend API (30+ endpoints)
- [x] Admin UI (bulk generate, manage)
- [x] Customer UI (variant selector)
- [x] Stock reservation system
- [x] Low stock alerts
- [x] Multi-image per variant
- [x] Price override per variant

## 🎯 Next Steps

1. Migrate existing products ke variants
2. Upload variant images (untuk colors)
3. Set low stock thresholds
4. Monitor low stock alerts
5. Test checkout flow

---

**Status**: ✅ Ready to Use
**Dokumentasi Lengkap**: `VARIANT_SYSTEM_GUIDE.md`
