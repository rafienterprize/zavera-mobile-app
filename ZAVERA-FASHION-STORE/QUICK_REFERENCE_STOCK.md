# Stock System - Quick Reference Card

## 🎯 Quick Answer: Why Stock Shows 0?

**If product has variants → Stock = 0 is NORMAL**

Stock is stored in variants, not the product itself. This is how all major e-commerce platforms work (Tokopedia, Shopee, Lazada).

---

## 📊 Admin Dashboard

### What You See
```
Product Name          Stock
─────────────────────────────
Basic T-Shirt         25      ← Simple product
Premium T-Shirt       📦 Variants  ← Variant product
Limited Edition       8 ⚠️    ← Low stock
```

### What It Means
- **Number**: Product has direct stock (no variants)
- **📦 Variants**: Stock is in variants (click Edit to see)
- **⚠️ Warning**: Stock < 10 items
- **Red 0**: Simple product out of stock

---

## 🛍️ Customer Product Page

### Three Scenarios

#### 1️⃣ Product with Variants - No Selection
```
┌─────────────────────────┐
│   [Product Image]       │
│                         │
│ "Pilih ukuran dan warna"│ ← Guide user
└─────────────────────────┘

Button: DISABLED
```

#### 2️⃣ Variant Selected - Has Stock
```
┌─────────────────────────┐
│   [Product Image]       │
│                         │
│   (No overlay)          │
└─────────────────────────┘

Stock: "10 item tersedia"
Button: ENABLED
```

#### 3️⃣ Variant Selected - No Stock
```
┌─────────────────────────┐
│   [Product Image]       │
│                         │
│    "SOLD OUT"           │ ← Out of stock
└─────────────────────────┘

Stock: "Stok habis"
Button: DISABLED
```

---

## 🔧 How to Check Variant Stock

### Method 1: Admin UI
1. Admin → Products
2. Click "Edit" on product
3. Go to "Variants & Stock" tab
4. See all variants with stock

### Method 2: API
```bash
curl http://localhost:8080/api/products/46/variants
```

---

## ✅ Creating Product with Variants

### Correct Flow
```
1. Create product (stock will be 0)
2. Click Edit
3. Go to "Variants & Stock" tab
4. Click "Bulk Generate Variants"
5. Select sizes and colors
6. Set stock per variant
7. Click Generate
```

### Result
```
Product: Premium T-Shirt (stock = 0)
├── M-Red: 10 items
├── M-Blue: 15 items
├── L-Red: 8 items
└── L-Blue: 12 items

Total: 45 items available ✓
Admin shows: "📦 Variants"
Customer sees: "Pilih ukuran dan warna"
```

---

## ❌ Common Mistakes

### Mistake 1: Setting Product Stock First
```
❌ Create product → Set stock = 50 → Add variants
   Result: Product stock = 50, but variants have 0
   Problem: Can't purchase!

✅ Create product → Add variants → Set variant stock
   Result: Product stock = 0, variants have stock
   Success: Can purchase!
```

### Mistake 2: Thinking Stock = 0 is Error
```
❌ See stock = 0 → Think "No stock!"
   Reality: 45 items in variants

✅ See "📦 Variants" → Click Edit → Check variants
   Reality: See actual stock per variant
```

---

## 🐛 Troubleshooting

### Problem: SOLD OUT showing incorrectly
**Check**:
1. Open browser console (F12)
2. Look for: "Fetched variants:", "Variants count:"
3. Verify variants array has items
4. Check variant is_active = true

### Problem: Can't add to cart
**Check**:
1. Variant is selected (for variant products)
2. available_stock > 0
3. User is logged in
4. No errors in console

### Problem: Admin shows 0 instead of "Variants"
**Check**:
1. Clear browser cache (Ctrl+Shift+R)
2. Verify product has variants
3. Run: `curl http://localhost:8080/api/products/46/variants`

---

## 📁 Files Changed

### Frontend
- ✅ `frontend/src/app/admin/products/page.tsx` - Admin stock display
- ✅ `frontend/src/app/product/[id]/page.tsx` - Customer overlay logic

### Backend
- ✅ `backend/zavera_stock_fix.exe` - Latest compiled binary

### Documentation
- 📖 `STOCK_SYSTEM_EXPLAINED.md` - Full technical details
- 📖 `STOCK_VISUAL_GUIDE.md` - Visual examples
- 📖 `STOCK_FIX_SUMMARY.md` - Complete summary
- 📖 `QUICK_REFERENCE_STOCK.md` - This file

---

## 🚀 Testing

### Quick Test
```bash
# 1. Start backend
start-backend.bat

# 2. Check variant data
curl http://localhost:8080/api/products/46/variants

# 3. Open admin
http://localhost:3000/admin/products

# 4. Open product page
http://localhost:3000/product/46
```

### Expected Results
- ✅ Admin shows "📦 Variants" for variant products
- ✅ Customer sees "Pilih ukuran dan warna" before selection
- ✅ Stock appears after variant selection
- ✅ SOLD OUT only when actually out of stock

---

## 💡 Key Takeaway

**Stock = 0 for variant products is CORRECT and EXPECTED**

This is not a bug, it's how variant-based inventory works:
- Product = Container
- Variants = Actual SKUs with stock
- Total stock = Sum of all variant stocks

Just like Tokopedia, Shopee, and every major e-commerce platform! 🎉

---

## 📞 Need Help?

1. Check browser console for errors
2. Review `STOCK_SYSTEM_EXPLAINED.md` for details
3. Run `test_stock_display.bat` for diagnostics
4. Verify backend is running on port 8080
5. Check database for variant data

---

**Last Updated**: January 27, 2026
**Status**: ✅ Fixed and Working
**Compiled Binary**: `backend/zavera_stock_fix.exe`
