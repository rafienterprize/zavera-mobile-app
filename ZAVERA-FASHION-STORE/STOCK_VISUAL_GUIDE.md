# Stock Display Visual Guide

## Admin Dashboard - Products List

### Product WITHOUT Variants (Simple Product)
```
┌─────────────────────────────────────────────────────────┐
│ Product: Basic T-Shirt                                  │
│ Stock: 25 ✓ (green/white - normal stock)              │
└─────────────────────────────────────────────────────────┘
```

### Product WITH Variants
```
┌─────────────────────────────────────────────────────────┐
│ Product: Premium T-Shirt                                │
│ Stock: 📦 Variants (gray - click Edit to see details)  │
└─────────────────────────────────────────────────────────┘
```

### Low Stock Product
```
┌─────────────────────────────────────────────────────────┐
│ Product: Limited Edition Shirt                          │
│ Stock: 8 ⚠️ (amber - low stock warning)               │
└─────────────────────────────────────────────────────────┘
```

## Admin Edit Page - Variants Tab

```
┌──────────────────────────────────────────────────────────────┐
│ Variants & Stock Management                                  │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│ ┌─────────┬────────┬────────┬────────┬────────┬──────────┐ │
│ │ Size    │ Color  │ Price  │ Stock  │ Status │ Actions  │ │
│ ├─────────┼────────┼────────┼────────┼────────┼──────────┤ │
│ │ M       │ Red    │ 150000 │   10   │ Active │ Edit Del │ │
│ │ M       │ Blue   │ 150000 │   15   │ Active │ Edit Del │ │
│ │ L       │ Red    │ 150000 │    8   │ Active │ Edit Del │ │
│ │ L       │ Blue   │ 150000 │   12   │ Active │ Edit Del │ │
│ └─────────┴────────┴────────┴────────┴────────┴──────────┘ │
│                                                               │
│ Total Stock: 45 items across 4 variants                      │
└──────────────────────────────────────────────────────────────┘
```

## Customer Product Page

### Scenario 1: Product with Variants - No Selection Yet
```
┌──────────────────────────────────────────────────────┐
│                                                       │
│              [Product Image]                          │
│                                                       │
│         ┌─────────────────────────┐                  │
│         │ Pilih ukuran dan warna  │ ← Overlay        │
│         └─────────────────────────┘                  │
│                                                       │
└──────────────────────────────────────────────────────┘

Premium T-Shirt
Rp 150.000

Ukuran: [ M ] [ L ] [ XL ]  ← Must select
Warna:  [ Red ] [ Blue ] [ Black ]  ← Must select

Jumlah: [ - ] 1 [ + ]
? item tersedia  ← Shows after selection

[Tambah ke Keranjang] ← Disabled until variant selected
```

### Scenario 2: Variant Selected - Stock Available
```
┌──────────────────────────────────────────────────────┐
│                                                       │
│              [Product Image]                          │
│                                                       │
│                                                       │
│                                                       │
└──────────────────────────────────────────────────────┘

Premium T-Shirt
Rp 150.000

Ukuran: [ M✓ ] [ L ] [ XL ]  ← M selected
Warna:  [ Red✓ ] [ Blue ] [ Black ]  ← Red selected

Jumlah: [ - ] 1 [ + ]
10 item tersedia  ← Shows available stock

[Tambah ke Keranjang] ← Enabled
```

### Scenario 3: Variant Selected - Low Stock
```
┌──────────────────────────────────────────────────────┐
│ ⚠️ Sisa 8                                            │
│              [Product Image]                          │
│                                                       │
│                                                       │
│                                                       │
└──────────────────────────────────────────────────────┘

Premium T-Shirt
Rp 150.000

Ukuran: [ M ] [ L✓ ] [ XL ]
Warna:  [ Red✓ ] [ Blue ] [ Black ]

Jumlah: [ - ] 1 [ + ]
8 item tersedia - Segera habis!  ← Low stock warning

[Tambah ke Keranjang]
```

### Scenario 4: Variant Selected - Out of Stock
```
┌──────────────────────────────────────────────────────┐
│                                                       │
│              [Product Image]                          │
│                                                       │
│         ┌─────────────────────────┐                  │
│         │      SOLD OUT           │ ← Overlay        │
│         └─────────────────────────┘                  │
│                                                       │
└──────────────────────────────────────────────────────┘

Premium T-Shirt
Rp 150.000

Ukuran: [ M ] [ L ] [ XL✓ ]
Warna:  [ Red✓ ] [ Blue ] [ Black ]

Jumlah: [ - ] 1 [ + ]
Stok habis  ← Red text

[Stok Habis] ← Disabled button
```

### Scenario 5: Simple Product - Out of Stock
```
┌──────────────────────────────────────────────────────┐
│                                                       │
│              [Product Image]                          │
│                                                       │
│         ┌─────────────────────────┐                  │
│         │      SOLD OUT           │ ← Overlay        │
│         └─────────────────────────┘                  │
│                                                       │
└──────────────────────────────────────────────────────┘

Basic T-Shirt
Rp 100.000

(No variant selector - simple product)

Jumlah: [ - ] 1 [ + ]
Stok habis

[Stok Habis] ← Disabled
```

## Key Visual Indicators

### Stock Status Colors
- **White/Green**: Normal stock (≥ 10 items)
- **Amber/Yellow**: Low stock (< 10 items) ⚠️
- **Red**: Out of stock (0 items) ❌
- **Gray**: Variant-based product 📦

### Overlays
1. **"Pilih ukuran dan warna"** (Light overlay)
   - Shown when variants exist but none selected
   - Semi-transparent background
   - Guides user to make selection

2. **"SOLD OUT"** (Dark overlay)
   - Shown when selected variant has no stock
   - Or when simple product has no stock
   - Darker background, more prominent

### Badges
- **"Sisa X"** - Low stock badge (amber background)
- **Category badge** - Top right of image
- **Image counter** - "1 / 3" for multiple images

## User Flow

### Creating Product with Variants
```
1. Admin → Products → Add Product
   ↓
2. Fill basic info (name, price, description)
   ↓
3. Upload images
   ↓
4. Click "Save" (product.stock = 0 at this point)
   ↓
5. Go to Edit → Variants & Stock tab
   ↓
6. Click "Bulk Generate Variants"
   ↓
7. Select sizes: [M, L, XL]
   Select colors: [Red, Blue, Black]
   Set stock per variant: 10
   ↓
8. Click "Generate" → Creates 9 variants (3×3)
   ↓
9. Total stock: 90 items (9 variants × 10 each)
   ↓
10. Admin list shows: "📦 Variants"
    Customer page shows: "Pilih ukuran dan warna"
```

### Customer Purchasing Flow
```
1. Browse products → Click product
   ↓
2. See "Pilih ukuran dan warna" overlay
   ↓
3. Select size: M
   ↓
4. Select color: Red
   ↓
5. Overlay disappears, shows "10 item tersedia"
   ↓
6. Adjust quantity (1-10)
   ↓
7. Click "Tambah ke Keranjang"
   ↓
8. Success! Stock reserved in cart
```

## Common Mistakes to Avoid

❌ **Wrong**: Setting stock on product, then adding variants
```
Product stock = 50
Add variants → Product stock still shows 50
But variants have 0 stock → Can't purchase!
```

✅ **Correct**: Add variants first, then set stock on variants
```
Product stock = 0 (automatic)
Add variants → Set stock on each variant
Variants have stock → Can purchase!
```

❌ **Wrong**: Expecting product.stock to show total
```
Admin sees: Stock = 0
Thinks: "No stock!"
Reality: 45 items in variants
```

✅ **Correct**: Check variants tab for total stock
```
Admin sees: "📦 Variants"
Clicks Edit → Variants tab
Sees: 45 total items across variants
```

## Summary

| Product Type | Admin Display | Customer Display (No Selection) | Customer Display (Selected) |
|--------------|---------------|--------------------------------|----------------------------|
| Simple (stock > 0) | Stock number | Product image | Product image |
| Simple (stock = 0) | 0 (red) | SOLD OUT overlay | SOLD OUT overlay |
| Variants (no selection) | 📦 Variants | "Pilih ukuran dan warna" | - |
| Variants (selected, stock > 0) | 📦 Variants | - | Shows stock count |
| Variants (selected, stock = 0) | 📦 Variants | - | SOLD OUT overlay |

This is the expected behavior and matches how major e-commerce platforms work!
