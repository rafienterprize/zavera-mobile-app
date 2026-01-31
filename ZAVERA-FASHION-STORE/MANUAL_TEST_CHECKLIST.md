# Manual Testing Checklist - ZAVERA

## 🚀 Quick Start

1. **Start Backend:** `cd backend && ./zavera_size_filter.exe`
2. **Start Frontend:** `cd frontend && npm run dev`
3. **Open Browser:** http://localhost:3000

---

## ✅ CUSTOMER FLOW (30 minutes)

### 1. Homepage & Navigation
- [ ] Homepage loads without errors
- [ ] Hero carousel works
- [ ] All category links work (WANITA, PRIA, ANAK, SPORTS, LUXURY, BEAUTY)
- [ ] Search bar visible
- [ ] Cart icon shows count

### 2. Browse Products (Category: PRIA)
- [ ] Click "PRIA" → 17 products shown
- [ ] Product images load
- [ ] Product prices displayed
- [ ] "SISA X" stock badge shows

### 3. Filter Products
- [ ] Click "Celana" → 5 products (Tailored Trousers, Chino Pants, Jeans)
- [ ] Click "Atasan" → 3 products (Cotton Tee, Hoodie, Sweater)
- [ ] Click "Semua" → All 17 products back
- [ ] Click size "L" → Only products with L variant
- [ ] Click size "M" → L unselects, M selects
- [ ] Click "Hapus Semua" → All filters cleared

### 4. Product Detail
- [ ] Click "Hip Hop Baggy Jeans 22"
- [ ] Product images load
- [ ] Variant selector shows: M, L, XL
- [ ] Select size "L" → Stock updates
- [ ] Select color "Black"
- [ ] Price displayed correctly
- [ ] Click "Add to Cart" → Success notification
- [ ] Cart icon count increases

### 5. Cart
- [ ] Go to `/cart`
- [ ] Item shows: Hip Hop Baggy Jeans 22, Size L, Black
- [ ] Price correct: Rp 330.000
- [ ] Click "+" → Quantity increases, subtotal updates
- [ ] Click "-" → Quantity decreases
- [ ] Click "Remove" → Item removed
- [ ] Add item back
- [ ] Click "Proceed to Checkout"

### 6. Checkout
- [ ] Fill form:
  - Name: Test Customer
  - Email: test@example.com
  - Phone: 081234567890
  - Province: DKI Jakarta
  - City: Jakarta Selatan
  - District: Kebayoran Baru
  - Subdistrict: Senayan
  - Address: Jl. Test No. 123
  - Postal Code: 12190
- [ ] Click "Calculate Shipping"
- [ ] Shipping options load (JNE, TIKI, etc.)
- [ ] Select "JNE REG"
- [ ] Total amount correct (Item + Shipping)
- [ ] Click "Continue to Payment"

### 7. Payment
- [ ] Order summary correct
- [ ] Select payment: BCA Virtual Account
- [ ] Click "Pay Now"
- [ ] Midtrans page opens
- [ ] Complete payment (Sandbox)
- [ ] Redirect to success page
- [ ] Order code displayed

### 8. Order Tracking
- [ ] Go to `/orders`
- [ ] Order appears in list
- [ ] Click order
- [ ] Order details correct
- [ ] Status shows "PAID"
- [ ] Payment info displayed

---

## 👨‍💼 ADMIN PANEL (30 minutes)

### 1. Admin Login
- [ ] Go to `/admin`
- [ ] Click "Login with Google"
- [ ] Login with: pemberani073@gmail.com
- [ ] Redirect to dashboard

### 2. Dashboard
- [ ] Statistics cards show:
  - Total Revenue
  - Total Orders
  - Pending Orders
  - Total Customers
- [ ] Recent orders list
- [ ] Charts/graphs load
- [ ] No console errors

### 3. Orders Management
- [ ] Go to `/admin/orders`
- [ ] All orders listed
- [ ] Filter by status: PAID
- [ ] Search by order code
- [ ] Click on recent order
- [ ] Order details complete
- [ ] Customer info shown
- [ ] Items listed correctly

### 4. Update Order Status
- [ ] Click "Mark as Shipped"
- [ ] Enter resi: JNE123456789
- [ ] Submit
- [ ] Status updates to "SHIPPED"
- [ ] Resi displayed
- [ ] Audit log recorded

### 5. Refund Process
- [ ] Find a DELIVERED order
- [ ] Click "Process Refund"
- [ ] Select "FULL" refund
- [ ] Reason: "Customer Request"
- [ ] Details: "Test refund"
- [ ] Submit
- [ ] Refund created (PENDING)
- [ ] Click "Process Refund" button
- [ ] If error 418: Click "Mark as Completed"
- [ ] Enter note: "Manual bank transfer completed"
- [ ] Refund status: COMPLETED
- [ ] Stock restored

### 6. Product Management
- [ ] Go to `/admin/products`
- [ ] All products listed
- [ ] Click "Add Product"
- [ ] Fill details:
  - Name: Test Product
  - Category: pria
  - Subcategory: Tops
  - Price: 100000
  - Stock: 10
- [ ] Upload image
- [ ] Save
- [ ] Product created
- [ ] Edit product
- [ ] Update stock
- [ ] Save changes

### 7. Variants Management
- [ ] Open product with variants
- [ ] View variants list
- [ ] Check stock levels
- [ ] Update variant stock
- [ ] Generate new variant
- [ ] Deactivate variant

### 8. Customer Management
- [ ] Go to `/admin/customers`
- [ ] Customers listed
- [ ] Search customer
- [ ] View customer details
- [ ] Check order history
- [ ] Stats displayed

---

## 🔌 API TESTING (10 minutes)

### Run Automated Test
```bash
# Run API test script
test_api_endpoints.bat
```

**Check Results:**
- [ ] All endpoints return 200 OK
- [ ] Products endpoint works
- [ ] Category filter works
- [ ] Variants endpoint works
- [ ] Cart endpoints accessible
- [ ] Shipping rates work

---

## 🗄️ DATABASE TESTING (10 minutes)

### Run Database Integrity Test
```bash
# Run database test script
test_database.bat
```

**Check Results:**
- [ ] No invalid categories (0)
- [ ] No missing subcategories (0)
- [ ] No negative stock (0)
- [ ] No orphan records (0)
- [ ] Order totals consistent
- [ ] Payment amounts match
- [ ] Refunds don't exceed orders
- [ ] All products have images

---

## 🐛 COMMON ISSUES TO CHECK

### Frontend Issues
- [ ] No console errors
- [ ] Images load properly
- [ ] Buttons clickable
- [ ] Forms validate
- [ ] Modals open/close
- [ ] Notifications show

### Backend Issues
- [ ] Server running on :8080
- [ ] No panic/crash
- [ ] Logs show requests
- [ ] Database connected
- [ ] External APIs work (Biteship, Midtrans)

### Database Issues
- [ ] PostgreSQL running
- [ ] Connections available
- [ ] No deadlocks
- [ ] Queries fast (< 1s)
- [ ] Indexes used

---

## 📊 PERFORMANCE CHECK

### Page Load Times (DevTools Network)
- [ ] Homepage: < 2s
- [ ] Category page: < 2s
- [ ] Product detail: < 1.5s
- [ ] Cart: < 1s
- [ ] Checkout: < 2s
- [ ] Admin dashboard: < 3s

### API Response Times
- [ ] GET /products: < 500ms
- [ ] GET /products/:id: < 200ms
- [ ] POST /cart/add: < 300ms
- [ ] POST /shipping/rates: < 2s
- [ ] POST /checkout: < 1s

---

## ✅ FINAL CHECKLIST

### Critical Features
- [ ] Browse products works
- [ ] Add to cart works
- [ ] Checkout works
- [ ] Payment works
- [ ] Order created
- [ ] Admin can view orders
- [ ] Admin can update status
- [ ] Refund works

### Nice to Have
- [ ] Filters work smoothly
- [ ] Images load fast
- [ ] No console errors
- [ ] Mobile responsive
- [ ] Email notifications
- [ ] SSE notifications

---

## 📝 TEST RESULTS

**Date:** _______________
**Tester:** _______________
**Environment:** Development / Production

### Summary
- Total Tests: _____ / _____
- Passed: _____
- Failed: _____
- Blocked: _____

### Critical Issues Found
1. ________________________________
2. ________________________________
3. ________________________________

### Minor Issues Found
1. ________________________________
2. ________________________________
3. ________________________________

### Recommendations
1. ________________________________
2. ________________________________
3. ________________________________

### Sign-off
- [ ] All critical features working
- [ ] No blocking issues
- [ ] Ready for production

**Signature:** _______________
**Date:** _______________

---

## 🎯 Quick Test (5 minutes)

If short on time, test these critical paths:

1. **Customer:** Browse → Add to Cart → Checkout → Payment
2. **Admin:** Login → View Orders → Update Status
3. **API:** Run `test_api_endpoints.bat`
4. **Database:** Run `test_database.bat`

**All green?** ✅ System is healthy!
**Any red?** ❌ Investigate and fix!
