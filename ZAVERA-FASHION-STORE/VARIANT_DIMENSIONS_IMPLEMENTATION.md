# ✅ Variant Dimensions Implementation - E-Commerce Standard

## 🎯 Tujuan
Implementasi sistem dimensi produk variant untuk kalkulasi ongkir yang akurat, mengikuti standard Tokopedia, Shopee, dan e-commerce lainnya.

## 📦 Logika Volumetric Weight (Standard E-Commerce)

### Cara Kerja:
Ketika customer beli multiple items dengan variant berbeda:

**Contoh:**
- 1x Baju Size L (500g, 30x20x5 cm)
- 1x Baju Size XL (600g, 32x22x5 cm)

**Kalkulasi Paket:**
1. **Berat Total** = 500g + 600g = **1100g** (dijumlahkan)
2. **Panjang Paket** = MAX(30, 32) = **32 cm** (ambil terbesar)
3. **Lebar Paket** = MAX(20, 22) = **22 cm** (ambil terbesar)
4. **Tinggi Paket** = 5 + 5 = **10 cm** (dijumlahkan karena ditumpuk)

**Berat Volumetrik** = (P x L x T) / 6000 = (32 x 22 x 10) / 6000 = **1.17 kg**

**Ongkir dihitung dari**: MAX(berat aktual, berat volumetrik) = **1.17 kg**

### Kenapa Tinggi Dijumlahkan?
Karena baju dilipat dan **ditumpuk** dalam box:
- Baju 1 dilipat = tinggi 5cm
- Baju 2 dilipat = tinggi 5cm  
- **Total tinggi box** = 10cm ✅

Panjang & lebar pakai yang **terbesar** karena box harus muat item terbesar.

## 🛠️ Implementasi yang Sudah Dikerjakan

### 1. ✅ Database Migration
**File**: `database/migrate_variant_dimensions.sql`

```sql
ALTER TABLE product_variants 
ADD COLUMN IF NOT EXISTS length_cm INT,
ADD COLUMN IF NOT EXISTS width_cm INT,
ADD COLUMN IF NOT EXISTS height_cm INT;
```

**Status**: ✅ Executed successfully

### 2. ✅ Backend Models Updated
**File**: `backend/models/product_variant.go`

Added fields:
```go
LengthCm  *int `json:"length_cm,omitempty"`
WidthCm   *int `json:"width_cm,omitempty"`
HeightCm  *int `json:"height_cm,omitempty"`
```

### 3. ✅ Backend DTOs Updated
**File**: `backend/dto/variant_dto.go`

Added to `CreateVariantRequest` and `UpdateVariantRequest`:
```go
Weight    *int `json:"weight"` // Alias for weight_grams
LengthCm  *int `json:"length_cm"`
Length    *int `json:"length"` // Alias
WidthCm   *int `json:"width_cm"`
Width     *int `json:"width"` // Alias
HeightCm  *int `json:"height_cm"`
Height    *int `json:"height"` // Alias
```

### 4. ✅ Frontend Types Updated
**File**: `frontend/src/types/variant.ts`

Added dimension fields to `ProductVariant` and `CreateVariantRequest`

### 5. ✅ Admin UI Updated

#### A. Edit Product Page
**File**: `frontend/src/app/admin/products/edit/[id]/page.tsx`
- ❌ **REMOVED** Weight/Length/Width fields (tidak ada fungsi di product level)
- ✅ Dimensi sekarang dikelola di variant level

#### B. Variant Manager
**File**: `frontend/src/components/admin/VariantManager.tsx`
- ✅ **ADDED** Shipping Dimensions section in edit variant form
- ✅ Fields: Weight (g), Length (cm), Width (cm), Height (cm)
- ✅ Tooltip explaining how dimensions are used for shipping

## 🚧 Yang Masih Perlu Dikerjakan

### 1. Backend Repository Updates
**File**: `backend/repository/variant_repository.go`

Perlu update:
- `Create()` method - add length_cm, width_cm, height_cm to INSERT
- `Update()` method - add length_cm, width_cm, height_cm to UPDATE
- `FindByID()` method - add length_cm, width_cm, height_cm to SELECT

### 2. Shipping Service Updates
**File**: `backend/service/shipping_service.go`

Method `calculateRatesForItems()` perlu update untuk:
```go
// Calculate dimensions for package
var totalWeight int
var maxLength, maxWidth, totalHeight int

for _, item := range items {
    // Get variant dimensions
    variant := getVariantByID(item.VariantID)
    
    // Weight: sum all
    totalWeight += variant.WeightGrams * item.Quantity
    
    // Length & Width: take maximum
    if variant.LengthCm > maxLength {
        maxLength = variant.LengthCm
    }
    if variant.WidthCm > maxWidth {
        maxWidth = variant.WidthCm
    }
    
    // Height: sum all (stacked)
    totalHeight += variant.HeightCm * item.Quantity
}

// Send to Biteship with dimensions
biteshipReq := GetRatesRequest{
    Weight: totalWeight,
    Length: maxLength,
    Width: maxWidth,
    Height: totalHeight,
    // ...
}
```

### 3. Client Product Page - Size Guide
**File**: `frontend/src/app/product/[id]/page.tsx`

Perlu tambahkan:
- Size Guide button/link
- Modal popup dengan size chart (seperti Zalora)
- Product dimensions display
- Fit guide (Slim, Regular, Oversized)

**Contoh dari Zalora:**
```
📏 Panduan Ukuran

Ukuran Produk:
- Lebar x Tinggi x Panjang box: 30 cm x 10 cm x 20 cm
- Berat: 0.7 kg

Panduan Ukuran Badan:
[Size Chart Table]
Size | Chest | Waist | Length
S    | 90cm  | 80cm  | 70cm
M    | 95cm  | 85cm  | 72cm
L    | 100cm | 90cm  | 74cm
```

## 📝 Testing Checklist

### Backend Testing
- [ ] Create variant with dimensions via API
- [ ] Update variant dimensions via API
- [ ] Get variant - verify dimensions returned
- [ ] Checkout with multiple variants - verify shipping calculation

### Frontend Testing
- [ ] Admin: Edit variant - see dimension fields
- [ ] Admin: Input dimensions - save successfully
- [ ] Admin: Edit product - no dimension fields (removed)
- [ ] Client: Product page - see size guide button
- [ ] Client: Click size guide - modal opens with info

### Shipping Calculation Testing
- [ ] Add 2 variants to cart (different sizes)
- [ ] Go to checkout
- [ ] Verify shipping cost calculated correctly
- [ ] Check backend logs for dimension calculation

## 🎨 UI/UX Improvements Needed

### 1. Size Guide Modal (Client)
Tambahkan modal seperti Zalora dengan:
- Product dimensions (P x L x T)
- Weight
- Size chart table
- Fit guide
- Care instructions

### 2. Dimension Input Helper (Admin)
Tambahkan helper text:
- "Ukuran produk setelah dilipat/dikemas"
- "Digunakan untuk kalkulasi ongkir"
- Preview volumetric weight

### 3. Bulk Edit Dimensions (Admin)
Untuk efficiency, tambahkan:
- Bulk update dimensions untuk multiple variants
- Copy dimensions from one variant to others
- Default dimensions per category

## 🚀 Next Steps

1. **PRIORITY 1**: Update backend repository untuk save/load dimensions
2. **PRIORITY 2**: Update shipping service untuk calculate volumetric weight
3. **PRIORITY 3**: Add size guide modal di product page
4. **PRIORITY 4**: Test end-to-end shipping calculation

## 📚 References

- Tokopedia Shipping Guide: https://seller.tokopedia.com/edu/cara-menghitung-berat-volumetrik/
- Shopee Volumetric Weight: https://seller.shopee.co.id/edu/article/11511
- Biteship API Docs: https://biteship.com/id/docs/api/rates

---
**Status**: 🟡 In Progress (60% complete)
**Last Updated**: January 28, 2026
