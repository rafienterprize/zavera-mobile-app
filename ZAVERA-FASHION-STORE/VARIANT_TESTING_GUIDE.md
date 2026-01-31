# Product Variant System - Testing Guide

## 🧪 Quick Testing Steps

### 1. Start Backend
```bash
cd backend
zavera_variants.exe
```
Backend should start on `http://localhost:8080`

### 2. Start Frontend
```bash
cd frontend
npm run dev
```
Frontend should start on `http://localhost:3000`

### 3. Test Admin Product Creation

#### Step 1: Login as Admin
1. Navigate to `http://localhost:3000/login`
2. Login with admin Google account (configured in `.env`)

#### Step 2: Create Product with Variants
1. Navigate to `http://localhost:3000/admin/products/add`
2. You should see a **single-page form** with 3 sections:
   - Left column: Basic Information + Product Variants
   - Right column: Product Images

#### Step 3: Fill Basic Information
- **Product Name**: "Classic Denim Jacket"
- **Description**: "Premium quality denim jacket with modern fit"
- **Category**: Select "Wanita" or "Pria"
- **Subcategory**: Select from dropdown (e.g., "Jacket")
- **Base Price**: 299000
- **Brand**: "Levi's"
- **Material**: "100% Cotton Denim"

#### Step 4: Upload Images
1. Click the upload area in the right column
2. Select multiple images (3-5 images recommended)
3. Wait for upload to complete
4. First image will be marked as "Primary"
5. You can remove images by hovering and clicking X

#### Step 5: Add Variants
1. Click "Add Variant" button
2. For each variant, set:
   - **Size**: Select from dropdown (XS, S, M, L, XL, XXL, XXXL)
   - **Color**: Select from dropdown (Black, White, Navy, Red, Blue, Green, Gray, Pink)
   - **Stock**: Enter quantity (e.g., 10)
   - **Price**: Enter price (can be different from base price)
   - **Weight**: Enter in grams (e.g., 400)
   - **Dimensions**: Length, Width, Height in cm (e.g., 70, 45, 3)
3. Click "Add Variant" again to add more variants
4. Example variants:
   - Size M, Black, Stock 15, Price 299000, 400g, 70×45×3cm
   - Size L, Black, Stock 10, Price 299000, 450g, 75×48×3cm
   - Size M, Navy, Stock 12, Price 299000, 400g, 70×45×3cm
   - Size L, Navy, Stock 8, Price 299000, 450g, 75×48×3cm

#### Step 6: Submit
1. Click "Create Product" button
2. Wait for success message
3. You should be redirected to `/admin/products`

### 4. Test Customer Product View

#### Step 1: View Product
1. Navigate to the product detail page
2. URL format: `http://localhost:3000/product/{id}`

#### Step 2: Verify Multi-Image Gallery
- ✅ Should see main image with thumbnails below
- ✅ Click thumbnails to change main image
- ✅ Use arrow buttons to navigate images
- ✅ See image counter (e.g., "1 / 5")

#### Step 3: Test Variant Selector
1. **Size Selection**:
   - Click on available sizes (should highlight)
   - Out-of-stock sizes should be disabled/crossed out
   - Selected size should have black background

2. **Color Selection**:
   - Click on color swatches
   - See actual color hex displayed
   - Out-of-stock colors should be disabled
   - Selected color should have black border

3. **Price Update**:
   - Price should update when variant is selected
   - If variants have different prices, should show range initially

4. **Stock Display**:
   - Should show "X in stock" when variant selected
   - Should show "Out of stock" if stock is 0
   - Low stock warning if stock < 10

#### Step 4: Add to Cart
1. Select size and color
2. Adjust quantity
3. Click "Tambah ke Keranjang"
4. Should see success modal
5. Verify cart shows correct variant (size, color)

### 5. Test API Endpoints

#### Get Product Variants
```bash
curl http://localhost:8080/api/products/1/variants
```

Expected response:
```json
{
  "variants": [
    {
      "id": 1,
      "product_id": 1,
      "sku": "PROD1-M-Black",
      "size": "M",
      "color": "Black",
      "color_hex": "#000000",
      "stock_quantity": 15,
      "available_stock": 15,
      "price": 299000,
      "weight_grams": 400,
      "is_active": true
    }
  ]
}
```

#### Get Available Options
```bash
curl http://localhost:8080/api/products/1/options
```

Expected response:
```json
{
  "size": ["M", "L"],
  "color": ["Black", "Navy"]
}
```

#### Find Specific Variant
```bash
curl -X POST http://localhost:8080/api/products/variants/find \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": 1,
    "size": "M",
    "color": "Black"
  }'
```

Expected response:
```json
{
  "variant": {
    "id": 1,
    "sku": "PROD1-M-Black",
    "size": "M",
    "color": "Black",
    "stock_quantity": 15,
    "price": 299000
  }
}
```

## ✅ Expected Behaviors

### Admin Form
- ✅ Single-page layout (no wizard steps)
- ✅ Category dropdown changes subcategory options
- ✅ Multi-image upload with preview
- ✅ Add/remove variants dynamically
- ✅ Each variant has full set of inputs
- ✅ Form validation prevents submission without required fields
- ✅ Loading states during upload and submission
- ✅ Dark theme consistent with admin dashboard

### Client Product Page
- ✅ Multi-image gallery with thumbnails
- ✅ Image navigation with arrows
- ✅ Image counter display
- ✅ Variant selector shows available options
- ✅ Disabled state for out-of-stock variants
- ✅ Price updates dynamically
- ✅ Stock availability shown
- ✅ Low stock warnings
- ✅ Add to cart with selected variant

### Stock Management
- ✅ Stock tracked per variant
- ✅ Cannot add more to cart than available stock
- ✅ Stock decreases after order
- ✅ Stock reservation during checkout
- ✅ Low stock alerts in admin

## 🐛 Common Issues & Solutions

### Issue: Images not uploading
**Solution**: Check Cloudinary credentials in `backend/.env`:
```
CLOUDINARY_CLOUD_NAME=your_cloud_name
CLOUDINARY_API_KEY=your_api_key
CLOUDINARY_API_SECRET=your_api_secret
```

### Issue: Variants not showing
**Solution**: 
1. Check backend logs for errors
2. Verify database migration ran successfully
3. Check browser console for API errors

### Issue: Admin form shows wizard instead of single page
**Solution**: 
1. Clear browser cache
2. Restart frontend dev server
3. Verify `page.tsx` is the correct file (not `page-complete.tsx`)

### Issue: Price not updating on variant selection
**Solution**:
1. Check browser console for errors
2. Verify variant has price set in database
3. Check API response includes price field

### Issue: Out-of-stock variants not disabled
**Solution**:
1. Verify `available_stock` field in API response
2. Check `is_active` flag is true
3. Verify stock_quantity > 0

## 📊 Test Data Examples

### Product 1: Classic Denim Jacket
- Category: Wanita
- Subcategory: Jacket
- Base Price: 299000
- Brand: Levi's
- Material: Cotton Denim
- Variants:
  - M, Black, 15 stock, 299000, 400g
  - L, Black, 10 stock, 299000, 450g
  - M, Navy, 12 stock, 299000, 400g
  - L, Navy, 8 stock, 299000, 450g

### Product 2: Premium T-Shirt
- Category: Pria
- Subcategory: T-Shirt
- Base Price: 149000
- Brand: Uniqlo
- Material: Cotton
- Variants:
  - S, White, 20 stock, 149000, 200g
  - M, White, 25 stock, 149000, 220g
  - L, White, 15 stock, 149000, 240g
  - S, Black, 18 stock, 149000, 200g
  - M, Black, 22 stock, 149000, 220g
  - L, Black, 12 stock, 149000, 240g

## 🎯 Success Criteria

The system is working correctly if:
- ✅ Admin can create products with multiple variants
- ✅ Admin can upload multiple images
- ✅ Each variant has independent stock and price
- ✅ Customers see multi-image gallery
- ✅ Customers can select size and color
- ✅ Price updates based on selected variant
- ✅ Out-of-stock variants are disabled
- ✅ Cart shows correct variant information
- ✅ Stock decreases after purchase
- ✅ Low stock warnings appear

## 🚀 Next Steps

After successful testing:
1. Add more products with variants
2. Test checkout flow with variants
3. Test stock reservation during checkout
4. Test admin stock management
5. Test low stock alerts
6. Test variant image management (if implemented)
7. Monitor performance with large number of variants
8. Test concurrent purchases (stock race conditions)

## 📞 Support

If you encounter issues:
1. Check backend logs: `backend/zavera_variants.exe`
2. Check frontend console: Browser DevTools
3. Check database: Verify migrations ran
4. Check API responses: Network tab in DevTools
5. Refer to `VARIANT_SYSTEM_COMPLETE.md` for architecture details
