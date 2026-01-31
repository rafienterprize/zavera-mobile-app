# Cloudinary Image Upload Implementation Plan

## 📋 Overview
Implementasi fitur upload image produk ke Cloudinary dengan dimensi produk untuk akurasi harga Biteship.

## ✅ Yang Sudah Dikerjakan

### 1. Database Migration
- ✅ File: `database/migrate_product_dimensions.sql`
- ✅ Added columns: `length`, `width`, `height` (in cm)
- ✅ Migration executed successfully

### 2. Backend - Models
- ✅ Updated `Product` model dengan dimensi (length, width, height)
- ✅ Updated DTOs: `CreateProductRequest`, `UpdateProductRequest`, `AdminProductResponse`

### 3. Backend - Cloudinary Service
- ✅ File: `backend/service/cloudinary_service.go`
- ✅ Functions:
  - `UploadImage()` - Upload ke Cloudinary
  - `DeleteImage()` - Delete dari Cloudinary
  - `ValidateImageFile()` - Validasi file (max 5MB, JPG/PNG/WEBP)
  - `ExtractPublicIDFromURL()` - Extract public ID dari URL
- ✅ Cloudinary SDK installed: `github.com/cloudinary/cloudinary-go/v2`

## 🔄 Yang Perlu Dikerjakan

### 4. Backend - Upload Handler
**File**: `backend/handler/admin_product_handler.go`

Perlu ditambahkan:
```go
// UploadProductImage handles multipart file upload
// POST /api/admin/products/upload-image
func (h *AdminProductHandler) UploadProductImage(c *gin.Context) {
    // 1. Get file from form
    // 2. Validate file
    // 3. Upload to Cloudinary
    // 4. Return URL
}
```

### 5. Backend - Update Product Service
**File**: `backend/service/admin_product_service.go`

Perlu update:
- `CreateProduct()` - Include length, width, height
- `UpdateProduct()` - Include length, width, height
- Query SQL untuk include dimensi

### 6. Backend - Update Repository
**File**: `backend/repository/product_repository.go`

Perlu update query:
```sql
INSERT INTO products (name, slug, description, price, stock, weight, length, width, height, ...)
SELECT id, name, slug, description, price, stock, weight, length, width, height, ...
```

### 7. Backend - Routes
**File**: `backend/routes/routes.go`

Perlu ditambahkan:
```go
admin.POST("/products/upload-image", adminProductHandler.UploadProductImage)
```

### 8. Frontend - File Upload Component
**File**: `frontend/src/components/ImageUpload.tsx` (NEW)

Features:
- Drag & drop area
- File preview
- Progress indicator
- Multiple image support
- Delete uploaded image

### 9. Frontend - Update Add Product Modal
**File**: `frontend/src/app/admin/products/page.tsx`

Perlu ditambahkan:
- Image upload component (replace URL input)
- Dimensi inputs (length, width, height)
- Preview uploaded images

### 10. Frontend - API Client
**File**: `frontend/src/lib/adminApi.ts`

Perlu ditambahkan:
```typescript
export async function uploadProductImage(file: File): Promise<string> {
  const formData = new FormData();
  formData.append('image', file);
  // Upload and return URL
}
```

## 📐 Biteship Dimension Requirements

Untuk akurasi harga shipping, produk harus punya:

| Field | Unit | Default | Description |
|-------|------|---------|-------------|
| Weight | gram | 500 | Berat produk |
| Length | cm | 30 | Panjang kemasan |
| Width | cm | 20 | Lebar kemasan |
| Height | cm | 5 | Tinggi kemasan |

**Contoh Produk:**
- T-Shirt: 200g, 30x20x3 cm
- Jacket: 600g, 40x30x8 cm
- Shoes: 800g, 35x25x12 cm

## 🎨 UI/UX Design

### Add Product Modal (Updated)
```
┌─────────────────────────────────────────┐
│ Add New Product                         │
├─────────────────────────────────────────┤
│ Name: [________________]                │
│ Description: [__________]               │
│                                         │
│ Price: [_____] Stock: [___]            │
│                                         │
│ ┌─ Product Dimensions (for shipping) ─┐│
│ │ Weight: [500] grams                 ││
│ │ Length: [30] cm                     ││
│ │ Width:  [20] cm                     ││
│ │ Height: [5] cm                      ││
│ └─────────────────────────────────────┘│
│                                         │
│ Category: [Wanita ▼]                   │
│                                         │
│ ┌─ Product Images ───────────────────┐ │
│ │ ┌─────────┐ ┌─────────┐           │ │
│ │ │  Drag   │ │ Preview │           │ │
│ │ │  Drop   │ │  Image  │           │ │
│ │ │  Here   │ │         │           │ │
│ │ └─────────┘ └─────────┘           │ │
│ │ or click to browse                 │ │
│ │ Max 5MB • JPG, PNG, WEBP          │ │
│ └────────────────────────────────────┘ │
│                                         │
│ [Cancel] [Create Product]              │
└─────────────────────────────────────────┘
```

### Image Upload Component
```
┌─────────────────────────────────────────┐
│ ┌─────────────────────────────────────┐ │
│ │         📤 Upload Image             │ │
│ │                                     │ │
│ │   Drag and drop image here          │ │
│ │   or click to browse                │ │
│ │                                     │ │
│ │   PNG, JPG, WEBP (Max 5MB)         │ │
│ └─────────────────────────────────────┘ │
│                                         │
│ Uploaded Images:                        │
│ ┌──────┐ ┌──────┐ ┌──────┐            │
│ │ IMG1 │ │ IMG2 │ │ IMG3 │            │
│ │  ✓   │ │  ✓   │ │  ✓   │            │
│ │  🗑️  │ │  🗑️  │ │  🗑️  │            │
│ └──────┘ └──────┘ └──────┘            │
└─────────────────────────────────────────┘
```

## 🔐 Security & Validation

### Backend Validation
- ✅ File size: Max 5MB
- ✅ File type: JPG, PNG, WEBP only
- ✅ MIME type validation
- ✅ Dimensions: > 0

### Cloudinary Settings
- Folder: `zavera/products/`
- Auto optimization: `q_auto,f_auto`
- Quality: `auto:good`
- Secure URL: HTTPS only

## 📝 Environment Variables

Already configured in `.env`:
```env
CLOUDINARY_CLOUD_NAME=dmofyz5tv
CLOUDINARY_API_KEY=836739665788915
CLOUDINARY_API_SECRET=II6aj86bAZjBl3VRwmUwtH04yck
```

## 🚀 Implementation Steps

### Phase 1: Backend (Priority)
1. ✅ Database migration
2. ✅ Update models & DTOs
3. ✅ Create Cloudinary service
4. ⏳ Add upload handler
5. ⏳ Update product service
6. ⏳ Update repository queries
7. ⏳ Add routes

### Phase 2: Frontend
1. ⏳ Create ImageUpload component
2. ⏳ Update Add Product modal
3. ⏳ Add dimension inputs
4. ⏳ Update API client
5. ⏳ Test upload flow

### Phase 3: Testing
1. ⏳ Test image upload
2. ⏳ Test product creation with image
3. ⏳ Test Biteship with dimensions
4. ⏳ Test image deletion

## 📊 Database Schema Changes

```sql
-- products table (UPDATED)
ALTER TABLE products 
ADD COLUMN length INTEGER DEFAULT 10,
ADD COLUMN width INTEGER DEFAULT 10,
ADD COLUMN height INTEGER DEFAULT 5;

-- product_images table (NO CHANGE)
-- Already supports multiple images per product
```

## 🎯 Success Criteria

- [x] Database has dimension columns
- [x] Cloudinary service created
- [ ] Admin can upload image via drag & drop
- [ ] Image stored in Cloudinary
- [ ] Product created with image URL
- [ ] Dimensions saved correctly
- [ ] Biteship uses dimensions for accurate pricing
- [ ] Old products work with default dimensions

## 📚 References

- Cloudinary Go SDK: https://github.com/cloudinary/cloudinary-go
- Biteship API: https://biteship.com/docs
- Multipart Upload: https://gin-gonic.com/docs/examples/upload-file/

---

## ❓ Questions for Review

1. **Apakah UI design sudah sesuai dengan yang Anda inginkan?**
2. **Apakah default dimensions (30x20x5 cm, 500g) sudah sesuai?**
3. **Apakah perlu support multiple images per product?** (database sudah support)
4. **Apakah perlu crop/resize image otomatis?** (Cloudinary bisa)
5. **Apakah perlu image compression?** (sudah auto dengan q_auto)

Silakan review plan ini terlebih dahulu sebelum saya lanjutkan implementasi lengkap!
