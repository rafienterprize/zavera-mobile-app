# Comprehensive System Testing - ZAVERA Fashion Store

## 🎯 Tujuan Testing

Menguji seluruh sistem end-to-end untuk memastikan semua fitur berfungsi dengan baik:
- Customer Flow (Browse → Cart → Checkout → Payment)
- Admin Panel (Orders, Products, Customers, Refunds)
- API Endpoints
- Database Integrity

---

## 📋 Pre-Test Checklist

### 1. Start Services

```bash
# Terminal 1 - Backend
cd backend
./zavera_size_filter.exe

# Terminal 2 - Frontend
cd frontend
npm run dev

# Terminal 3 - Database Check
psql -U postgres -d zavera_db
```

### 2. Verify Services Running

- ✅ Backend: http://localhost:8080
- ✅ Frontend: http://localhost:3000
- ✅ Database: PostgreSQL running

---

## 🛍️ PART 1: CUSTOMER FLOW TESTING

### Test 1.1: Browse Products

**URL:** `http://localhost:3000`

**Steps:**
1. ✅ Homepage loads with hero carousel
2. ✅ Click "PRIA" category
3. ✅ Verify 17 products displayed
4. ✅ Test filter "Celana" → 5 products shown
5. ✅ Test filter size "L" → Only products with L variant shown
6. ✅ Clear filters → All products shown again

**Expected Results:**
- All products load with images
- Filters work correctly
- Product count updates
- No console errors

---

### Test 1.2: Product Detail Page

**Steps:**
1. ✅ Click on "Hip Hop Baggy Jeans 22"
2. ✅ Verify product details displayed
3. ✅ Check variant selector shows: M, L, XL
4. ✅ Select size "L"
5. ✅ Select color "Black"
6. ✅ Check stock display
7. ✅ Click "Add to Cart"

**Expected Results:**
- Product images load
- Variant selector works
- Stock updates when variant selected
- Add to cart success notification
- Cart icon shows item count

---

### Test 1.3: Cart Functionality

**URL:** `http://localhost:3000/cart`

**Steps:**
1. ✅ Verify cart shows added item
2. ✅ Check item details: name, size, color, price
3. ✅ Test quantity increase (+)
4. ✅ Test quantity decrease (-)
5. ✅ Verify subtotal updates
6. ✅ Test "Remove" item
7. ✅ Add item back
8. ✅ Click "Proceed to Checkout"

**Expected Results:**
- Cart items display correctly
- Quantity updates work
- Price calculations correct
- Remove item works
- Checkout button enabled

---

### Test 1.4: Checkout Process

**URL:** `http://localhost:3000/checkout`

**Steps:**
1. ✅ Fill customer information:
   - Name: "Test Customer"
   - Email: "test@example.com"
   - Phone: "081234567890"
2. ✅ Fill shipping address:
   - Province: "DKI Jakarta"
   - City: "Jakarta Selatan"
   - District: "Kebayoran Baru"
   - Subdistrict: "Senayan"
   - Address: "Jl. Test No. 123"
   - Postal Code: "12190"
3. ✅ Click "Calculate Shipping"
4. ✅ Select courier (e.g., "JNE REG")
5. ✅ Verify total amount
6. ✅ Click "Continue to Payment"

**Expected Results:**
- Form validation works
- Shipping rates load from Biteship
- Multiple courier options shown
- Total calculation correct
- Proceed to payment

---

### Test 1.5: Payment Process

**URL:** `http://localhost:3000/checkout/payment`

**Steps:**
1. ✅ Verify order summary displayed
2. ✅ Select payment method:
   - Option A: Bank Transfer (BCA VA)
   - Option B: QRIS
   - Option C: GoPay
3. ✅ Click "Pay Now"
4. ✅ Verify Midtrans payment page opens
5. ✅ Complete payment (Sandbox)
6. ✅ Verify redirect to success page

**Expected Results:**
- Payment methods load
- Midtrans integration works
- Payment successful
- Order created in database
- Email notification sent

---

### Test 1.6: Order Tracking

**URL:** `http://localhost:3000/orders`

**Steps:**
1. ✅ Login/verify email
2. ✅ View order list
3. ✅ Click on recent order
4. ✅ Verify order details
5. ✅ Check order status
6. ✅ View payment info
7. ✅ Check shipping info (if shipped)

**Expected Results:**
- Orders displayed
- Order details correct
- Status accurate
- Timeline shows progress

---

## 👨‍💼 PART 2: ADMIN PANEL TESTING

### Test 2.1: Admin Login

**URL:** `http://localhost:3000/admin`

**Steps:**
1. ✅ Click "Login with Google"
2. ✅ Login with admin account: `pemberani073@gmail.com`
3. ✅ Verify redirect to dashboard

**Expected Results:**
- Google OAuth works
- Only admin email can access
- Dashboard loads

---

### Test 2.2: Admin Dashboard

**URL:** `http://localhost:3000/admin/dashboard`

**Steps:**
1. ✅ Verify statistics cards:
   - Total Revenue
   - Total Orders
   - Pending Orders
   - Total Customers
2. ✅ Check recent orders list
3. ✅ Verify charts/graphs load
4. ✅ Test date range filter

**Expected Results:**
- All stats display correctly
- Numbers match database
- Charts render
- Real-time updates via SSE

---

### Test 2.3: Order Management

**URL:** `http://localhost:3000/admin/orders`

**Steps:**
1. ✅ View all orders
2. ✅ Test filters:
   - Status: PENDING, PAID, SHIPPED, DELIVERED
   - Date range
   - Search by order code
3. ✅ Click on an order
4. ✅ View order details
5. ✅ Test actions:
   - Mark as Shipped (add resi)
   - Mark as Delivered
   - Cancel Order
6. ✅ Verify audit log

**Expected Results:**
- Orders list loads
- Filters work
- Order details complete
- Status updates work
- Audit trail recorded

---

### Test 2.4: Refund Management

**URL:** `http://localhost:3000/admin/orders/[code]`

**Steps:**
1. ✅ Open a DELIVERED order
2. ✅ Click "Process Refund"
3. ✅ Select refund type:
   - FULL
   - PARTIAL
   - SHIPPING_ONLY
   - ITEM_ONLY
4. ✅ Fill refund details
5. ✅ Submit refund request
6. ✅ Process refund (if auto-refund works)
7. ✅ OR Mark as completed manually (if error 418)

**Expected Results:**
- Refund form works
- Validation correct
- Midtrans refund API called
- Stock restored
- Order status updated

---

### Test 2.5: Product Management

**URL:** `http://localhost:3000/admin/products`

**Steps:**
1. ✅ View all products
2. ✅ Click "Add Product"
3. ✅ Fill product details:
   - Name, Description, Price
   - Category, Subcategory
   - Weight, Dimensions
   - Upload images
4. ✅ Save product
5. ✅ Edit existing product
6. ✅ Generate variants
7. ✅ Update stock
8. ✅ Deactivate product

**Expected Results:**
- Product CRUD works
- Image upload to Cloudinary
- Variants generated
- Stock management works

---

### Test 2.6: Customer Management

**URL:** `http://localhost:3000/admin/customers`

**Steps:**
1. ✅ View all customers
2. ✅ Search customer
3. ✅ View customer details
4. ✅ Check order history
5. ✅ View customer stats

**Expected Results:**
- Customer list loads
- Search works
- Details accurate
- Order history complete

---

### Test 2.7: Shipment Tracking

**URL:** `http://localhost:3000/admin/shipments`

**Steps:**
1. ✅ View all shipments
2. ✅ Check tracking status
3. ✅ Test manual tracking update
4. ✅ Verify auto-tracking job

**Expected Results:**
- Shipments listed
- Tracking info from courier
- Status updates work
- Background job running

---

## 🔌 PART 3: API ENDPOINT TESTING

### Test 3.1: Product Endpoints

```bash
# Get all products
curl http://localhost:8080/products

# Get products by category
curl http://localhost:8080/products?category=pria

# Get product by ID
curl http://localhost:8080/products/1

# Get product variants
curl http://localhost:8080/products/46/variants
```

**Expected:** 200 OK, JSON response with products

---

### Test 3.2: Cart Endpoints

```bash
# Add to cart
curl -X POST http://localhost:8080/cart/add \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": 46,
    "variant_id": 1,
    "quantity": 1
  }'

# Get cart
curl http://localhost:8080/cart

# Update cart item
curl -X PUT http://localhost:8080/cart/items/1 \
  -H "Content-Type: application/json" \
  -d '{"quantity": 2}'

# Remove from cart
curl -X DELETE http://localhost:8080/cart/items/1
```

**Expected:** Cart operations successful

---

### Test 3.3: Checkout Endpoints

```bash
# Get shipping rates
curl -X POST http://localhost:8080/shipping/rates \
  -H "Content-Type: application/json" \
  -d '{
    "destination_area_id": "IDNP6IDNC148IDND1817IDZ12190",
    "items": [{"product_id": 46, "quantity": 1}]
  }'

# Create order
curl -X POST http://localhost:8080/checkout \
  -H "Content-Type: application/json" \
  -d '{
    "customer_name": "Test",
    "email": "test@example.com",
    "phone": "081234567890",
    "items": [{"product_id": 46, "quantity": 1}]
  }'
```

**Expected:** Shipping rates returned, order created

---

### Test 3.4: Payment Endpoints

```bash
# Create payment (Core API)
curl -X POST http://localhost:8080/payments/core/create \
  -H "Content-Type: application/json" \
  -d '{
    "order_code": "ORD-xxx",
    "payment_method": "bca_va"
  }'

# Check payment status
curl http://localhost:8080/payments/status/ORD-xxx

# Webhook callback (simulate)
curl -X POST http://localhost:8080/payments/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "order_id": "ORD-xxx",
    "transaction_status": "settlement"
  }'
```

**Expected:** Payment created, status checked, webhook processed

---

### Test 3.5: Admin Endpoints

```bash
# Get dashboard stats
curl http://localhost:8080/admin/dashboard \
  -H "Authorization: Bearer ADMIN_TOKEN"

# Get all orders
curl http://localhost:8080/admin/orders \
  -H "Authorization: Bearer ADMIN_TOKEN"

# Update order status
curl -X PUT http://localhost:8080/admin/orders/1/status \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status": "SHIPPED", "resi": "JNE123456"}'

# Create refund
curl -X POST http://localhost:8080/admin/refunds \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "order_code": "ORD-xxx",
    "refund_type": "FULL",
    "reason": "CUSTOMER_REQUEST"
  }'
```

**Expected:** Admin operations successful

---

## 🗄️ PART 4: DATABASE INTEGRITY TESTING

### Test 4.1: Data Consistency

```sql
-- Check products have valid categories
SELECT COUNT(*) FROM products WHERE category NOT IN ('wanita', 'pria', 'anak', 'sports', 'luxury', 'beauty');
-- Expected: 0

-- Check all products have subcategory
SELECT COUNT(*) FROM products WHERE subcategory IS NULL;
-- Expected: 0

-- Check variants have valid product_id
SELECT COUNT(*) FROM product_variants WHERE product_id NOT IN (SELECT id FROM products);
-- Expected: 0

-- Check orders have valid user_id
SELECT COUNT(*) FROM orders WHERE user_id NOT IN (SELECT id FROM users);
-- Expected: 0

-- Check cart items have valid product_id
SELECT COUNT(*) FROM cart_items WHERE product_id NOT IN (SELECT id FROM products);
-- Expected: 0
```

---

### Test 4.2: Stock Consistency

```sql
-- Check product stock matches variant stock
SELECT p.id, p.name, p.stock as product_stock, 
       COALESCE(SUM(pv.stock_quantity), 0) as variant_stock
FROM products p
LEFT JOIN product_variants pv ON p.id = pv.product_id
GROUP BY p.id, p.name, p.stock
HAVING p.stock != COALESCE(SUM(pv.stock_quantity), 0);
-- Expected: 0 rows (or acceptable differences)

-- Check no negative stock
SELECT * FROM products WHERE stock < 0;
-- Expected: 0

SELECT * FROM product_variants WHERE stock_quantity < 0;
-- Expected: 0
```

---

### Test 4.3: Order Integrity

```sql
-- Check order totals match items
SELECT o.id, o.order_code, o.total_amount,
       SUM(oi.quantity * oi.price_per_unit) + o.shipping_cost as calculated_total
FROM orders o
JOIN order_items oi ON o.id = oi.order_id
GROUP BY o.id, o.order_code, o.total_amount, o.shipping_cost
HAVING o.total_amount != SUM(oi.quantity * oi.price_per_unit) + o.shipping_cost;
-- Expected: 0 rows

-- Check payment amounts match order amounts
SELECT o.id, o.order_code, o.total_amount, p.amount as payment_amount
FROM orders o
JOIN payments p ON o.id = p.order_id
WHERE o.total_amount != p.amount;
-- Expected: 0 rows
```

---

### Test 4.4: Refund Integrity

```sql
-- Check refund amounts don't exceed order amounts
SELECT r.id, r.refund_code, r.order_id, r.refund_amount, o.total_amount
FROM refunds r
JOIN orders o ON r.order_id = o.id
WHERE r.refund_amount > o.total_amount;
-- Expected: 0 rows

-- Check refund status consistency
SELECT * FROM refunds 
WHERE status NOT IN ('PENDING', 'PROCESSING', 'COMPLETED', 'FAILED');
-- Expected: 0 rows
```

---

## 📊 PART 5: PERFORMANCE TESTING

### Test 5.1: Page Load Times

**Measure:**
- Homepage: < 2s
- Category page: < 2s
- Product detail: < 1.5s
- Cart: < 1s
- Checkout: < 2s
- Admin dashboard: < 3s

**Tool:** Browser DevTools Network tab

---

### Test 5.2: API Response Times

```bash
# Test product list endpoint
time curl http://localhost:8080/products?category=pria
# Expected: < 500ms

# Test shipping rates
time curl -X POST http://localhost:8080/shipping/rates -d '{...}'
# Expected: < 2s (external API)

# Test cart operations
time curl http://localhost:8080/cart
# Expected: < 200ms
```

---

### Test 5.3: Database Query Performance

```sql
-- Check slow queries
SELECT query, mean_exec_time, calls
FROM pg_stat_statements
WHERE mean_exec_time > 1000
ORDER BY mean_exec_time DESC
LIMIT 10;

-- Check missing indexes
SELECT schemaname, tablename, attname, n_distinct, correlation
FROM pg_stats
WHERE schemaname = 'public'
AND correlation < 0.1
ORDER BY n_distinct DESC;
```

---

## 🔒 PART 6: SECURITY TESTING

### Test 6.1: Authentication

- ✅ Unauthenticated users cannot access admin
- ✅ Non-admin Google accounts rejected
- ✅ JWT tokens expire correctly
- ✅ Session management works

### Test 6.2: Authorization

- ✅ Users can only access their own orders
- ✅ Admin can access all orders
- ✅ API endpoints require proper auth

### Test 6.3: Input Validation

- ✅ SQL injection prevention
- ✅ XSS prevention
- ✅ CSRF protection
- ✅ File upload validation

---

## 📝 TEST RESULTS TEMPLATE

```markdown
## Test Execution: [Date]

### Customer Flow
- [ ] Browse Products: PASS/FAIL
- [ ] Product Detail: PASS/FAIL
- [ ] Cart: PASS/FAIL
- [ ] Checkout: PASS/FAIL
- [ ] Payment: PASS/FAIL
- [ ] Order Tracking: PASS/FAIL

### Admin Panel
- [ ] Login: PASS/FAIL
- [ ] Dashboard: PASS/FAIL
- [ ] Orders: PASS/FAIL
- [ ] Refunds: PASS/FAIL
- [ ] Products: PASS/FAIL
- [ ] Customers: PASS/FAIL

### API Endpoints
- [ ] Products: PASS/FAIL
- [ ] Cart: PASS/FAIL
- [ ] Checkout: PASS/FAIL
- [ ] Payment: PASS/FAIL
- [ ] Admin: PASS/FAIL

### Database
- [ ] Data Consistency: PASS/FAIL
- [ ] Stock Consistency: PASS/FAIL
- [ ] Order Integrity: PASS/FAIL
- [ ] Refund Integrity: PASS/FAIL

### Performance
- [ ] Page Load Times: PASS/FAIL
- [ ] API Response Times: PASS/FAIL
- [ ] Database Performance: PASS/FAIL

### Security
- [ ] Authentication: PASS/FAIL
- [ ] Authorization: PASS/FAIL
- [ ] Input Validation: PASS/FAIL

### Issues Found
1. [Issue description]
2. [Issue description]

### Recommendations
1. [Recommendation]
2. [Recommendation]
```

---

## 🚀 Quick Test Script

Saya akan buat script otomatis untuk test API endpoints:

```bash
# Save as test_api.sh
#!/bin/bash

echo "=== ZAVERA API Testing ==="
echo ""

# Test 1: Products
echo "Test 1: Get Products"
curl -s http://localhost:8080/products | jq '.[] | {id, name, category}' | head -20
echo ""

# Test 2: Product by Category
echo "Test 2: Get PRIA Products"
curl -s "http://localhost:8080/products?category=pria" | jq 'length'
echo ""

# Test 3: Health Check
echo "Test 3: Health Check"
curl -s http://localhost:8080/health | jq '.'
echo ""

echo "=== Tests Complete ==="
```

---

## 📞 Support

Jika menemukan bug atau issue:
1. Catat langkah reproduksi
2. Screenshot error
3. Check console logs
4. Check database state

**Ready to start testing!** 🧪
