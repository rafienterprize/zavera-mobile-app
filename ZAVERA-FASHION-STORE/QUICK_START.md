# Quick Start - Product Variant System

## 🚀 Start the System

### 1. Start Backend (Terminal 1)
```bash
cd backend
zavera_variants.exe
```
✅ Backend running on `http://localhost:8080`

### 2. Start Frontend (Terminal 2)
```bash
cd frontend
npm run dev
```
✅ Frontend running on `http://localhost:3000`

## 🎯 Test the System

### Create Your First Product with Variants

1. **Login as Admin**
   - Go to: `http://localhost:3000/login`
   - Login with admin Google account

2. **Create Product**
   - Go to: `http://localhost:3000/admin/products/add`
   - You'll see a single-page form with 3 sections

3. **Fill Basic Info** (Left Column - Top)
   ```
   Product Name: Classic Denim Jacket
   Description: Premium quality denim jacket
   Category: Wanita
   Subcategory: Jacket
   Base Price: 299000
   Brand: Levi's
   Material: Cotton Denim
   ```

4. **Upload Images** (Right Column)
   - Click upload area
   - Select 3-5 images
   - Wait for upload
   - First image = Primary

5. **Add Variants** (Left Column - Bottom)
   - Click "Add Variant" button
   - Add these variants:
   
   **Variant 1:**
   ```
   Size: M
   Color: Black
   Stock: 15
   Price: 299000
   Weight: 400g
   Length: 70cm, Width: 45cm, Height: 3cm
   ```
   
   **Variant 2:**
   ```
   Size: L
   Color: Black
   Stock: 10
   Price: 299000
   Weight: 450g
   Length: 75cm, Width: 48cm, Height: 3cm
   ```
   
   **Variant 3:**
   ```
   Size: M
   Color: Navy
   Stock: 12
   Price: 299000
   Weight: 400g
   Length: 70cm, Width: 45cm, Height: 3cm
   ```

6. **Submit**
   - Click "Create Product"
   - Wait for success
   - Redirected to products list

### View Product as Customer

1. **Go to Product Page**
   - Navigate to: `http://localhost:3000/product/1`
   - (Replace 1 with your product ID)

2. **Test Image Gallery**
   - ✅ See main image
   - ✅ See thumbnails below
   - ✅ Click thumbnails to change image
   - ✅ Use arrow buttons to navigate
   - ✅ See image counter (1/5)

3. **Test Variant Selector**
   - ✅ Click size buttons (M, L)
   - ✅ Click color swatches (Black, Navy)
   - ✅ See price update
   - ✅ See stock display
   - ✅ Out-of-stock variants disabled

4. **Add to Cart**
   - Select size: M
   - Select color: Black
   - Set quantity: 1
   - Click "Tambah ke Keranjang"
   - ✅ See success modal
   - ✅ Go to cart
   - ✅ Verify variant shows (M - Black)

## ✅ Success Indicators

If everything works, you should see:
- ✅ Single-page admin form (no wizard)
- ✅ Category dropdown changes subcategories
- ✅ Multi-image upload works
- ✅ Can add/remove variants
- ✅ Each variant has all fields
- ✅ Product created successfully
- ✅ Customer sees image gallery
- ✅ Customer can select variants
- ✅ Price updates dynamically
- ✅ Stock shows correctly
- ✅ Cart shows variant details

## 🐛 Troubleshooting

### Backend won't start
```bash
# Check if port 8080 is in use
netstat -ano | findstr :8080

# Kill process if needed
taskkill /PID <process_id> /F

# Restart backend
cd backend
zavera_variants.exe
```

### Frontend won't start
```bash
# Check if port 3000 is in use
netstat -ano | findstr :3000

# Kill process if needed
taskkill /PID <process_id> /F

# Restart frontend
cd frontend
npm run dev
```

### Images not uploading
Check `backend/.env` has Cloudinary credentials:
```
CLOUDINARY_CLOUD_NAME=your_cloud_name
CLOUDINARY_API_KEY=your_api_key
CLOUDINARY_API_SECRET=your_api_secret
```

### Admin form shows wizard
1. Clear browser cache (Ctrl+Shift+Delete)
2. Hard refresh (Ctrl+F5)
3. Restart frontend dev server

### Variants not showing
1. Check backend logs for errors
2. Open browser DevTools → Network tab
3. Check API response for `/api/products/1/variants`
4. Verify database migration ran

## 📚 Documentation

For detailed information, see:
- `VARIANT_SYSTEM_COMPLETE.md` - Complete system documentation
- `VARIANT_TESTING_GUIDE.md` - Detailed testing procedures
- `IMPLEMENTATION_SUMMARY.md` - Implementation overview

## 🎉 You're Ready!

The system is fully functional and ready for production use. Start creating products with variants and test the complete flow from admin creation to customer purchase.

**Happy coding!** 🚀
