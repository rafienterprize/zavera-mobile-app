# Testing Dashboard Admin - Quick Guide

## ✅ Backend Status: RUNNING

Server berhasil berjalan di `http://localhost:8080`

## 🔐 Authentication Required

Semua endpoint dashboard memerlukan:
1. Login sebagai admin
2. Token JWT di header: `Authorization: Bearer <token>`
3. Email admin harus sesuai dengan `ADMIN_GOOGLE_EMAIL` di `.env`

## 📊 Dashboard Endpoints

### 1. Executive Metrics
```bash
GET http://localhost:8080/api/admin/dashboard/executive?period=today
```

**Query Parameters:**
- `period`: `today`, `week`, `month`, `year`

**Response:**
```json
{
  "gmv": 50000000,
  "revenue": 45000000,
  "pending_revenue": 5000000,
  "total_orders": 150,
  "paid_orders": 135,
  "avg_order_value": 333333,
  "conversion_rate": 90.0,
  "payment_methods": [...],
  "top_products": [...]
}
```

---

### 2. Payment Monitor (Real-time)
```bash
GET http://localhost:8080/api/admin/dashboard/payments
```

**Response:**
```json
{
  "pending_count": 25,
  "pending_amount": 5000000,
  "expiring_soon_count": 5,
  "expiring_soon_amount": 1000000,
  "stuck_payments": [...],
  "today_paid_count": 50,
  "today_paid_amount": 15000000,
  "method_performance": [...]
}
```

---

### 3. Inventory Alerts
```bash
GET http://localhost:8080/api/admin/dashboard/inventory
```

**Response:**
```json
{
  "out_of_stock": [...],
  "low_stock": [...],
  "fast_moving": [...]
}
```

---

### 4. Customer Insights
```bash
GET http://localhost:8080/api/admin/dashboard/customers
```

**Response:**
```json
{
  "total_customers": 500,
  "active_customers": 150,
  "new_customers": 50,
  "segments": [...],
  "top_customers": [...]
}
```

---

### 5. Conversion Funnel
```bash
GET http://localhost:8080/api/admin/dashboard/funnel?period=today
```

**Query Parameters:**
- `period`: `today`, `week`, `month`, `year`

**Response:**
```json
{
  "orders_created": 100,
  "orders_paid": 85,
  "orders_shipped": 80,
  "orders_delivered": 75,
  "orders_completed": 70,
  "payment_rate": 85.0,
  "fulfillment_rate": 80.0,
  "delivery_rate": 75.0,
  "completion_rate": 70.0,
  "drop_offs": [...]
}
```

---

### 6. Revenue Chart
```bash
GET http://localhost:8080/api/admin/dashboard/revenue-chart?period=7days
```

**Query Parameters:**
- `period`: `7days`, `30days`, `90days`, `year`

**Response:**
```json
{
  "data_points": [
    {"date": "2025-01-13", "orders": 10, "revenue": 5000000},
    {"date": "2025-01-14", "orders": 15, "revenue": 7500000}
  ]
}
```

---

## 🧪 Testing dengan Postman/Thunder Client

### Step 1: Login sebagai Admin
```bash
POST http://localhost:8080/api/auth/login
Content-Type: application/json

{
  "email": "admin@zavera.com",
  "password": "your_password"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {...}
}
```

### Step 2: Test Dashboard Endpoint
```bash
GET http://localhost:8080/api/admin/dashboard/executive?period=today
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

---

## 🌐 Testing Frontend

### Step 1: Start Frontend
```bash
cd frontend
npm run dev
```

### Step 2: Access Dashboard
```
http://localhost:3000/admin/dashboard
```

### Expected Features:
- ✅ Executive KPI cards (GMV, Revenue, Pending, AOV)
- ✅ Critical alerts banner (if issues exist)
- ✅ Payment monitor panel with real-time data
- ✅ Conversion funnel visualization
- ✅ Inventory alerts (out of stock, low stock, fast moving)
- ✅ Customer insights with RFM segmentation
- ✅ Top products & payment methods
- ✅ Period selector (Today/Week/Month/Year)
- ✅ Refresh button

---

## 🐛 Troubleshooting

### Error: "Unauthorized"
- Pastikan sudah login sebagai admin
- Cek token JWT masih valid
- Cek email admin di `.env` file: `ADMIN_GOOGLE_EMAIL`

### Error: "No data"
- Database mungkin kosih
- Buat beberapa test orders terlebih dahulu
- Gunakan endpoint `/api/checkout` untuk membuat order

### Error: "Connection refused"
- Pastikan backend running di port 8080
- Cek dengan: `curl http://localhost:8080/health`

### Frontend tidak load data
- Buka browser console (F12)
- Cek network tab untuk error API calls
- Pastikan `NEXT_PUBLIC_API_URL` di `.env.local` benar

---

## 📈 Sample Data untuk Testing

Jika database kosong, buat sample data:

### 1. Create Products
```bash
POST http://localhost:8080/api/admin/products
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Test Product",
  "price": 500000,
  "stock": 10,
  "category": "Wanita",
  "weight": 500
}
```

### 2. Create Test Order
```bash
POST http://localhost:8080/api/checkout
Content-Type: application/json

{
  "customer_name": "Test Customer",
  "customer_email": "test@example.com",
  "customer_phone": "08123456789",
  "items": [
    {"product_id": 1, "quantity": 2}
  ]
}
```

### 3. Mark Order as Paid (simulate payment)
```bash
PATCH http://localhost:8080/api/admin/orders/:code/status
Authorization: Bearer <token>
Content-Type: application/json

{
  "status": "PAID",
  "reason": "Test payment"
}
```

---

## ✅ Success Indicators

Dashboard berfungsi dengan baik jika:
- ✅ Semua cards menampilkan angka (bukan 0 semua)
- ✅ Payment monitor menunjukkan pending payments
- ✅ Conversion funnel menampilkan progress bars
- ✅ Inventory alerts menampilkan produk (jika ada)
- ✅ Customer insights menampilkan segmentasi
- ✅ Period selector mengubah data saat diklik
- ✅ Refresh button memuat ulang data
- ✅ Tidak ada error di browser console

---

## 🎯 Next Steps

Setelah testing berhasil:
1. ✅ Verifikasi semua metrics akurat
2. ✅ Test dengan data real (bukan sample)
3. ✅ Test responsive design (mobile/tablet)
4. ✅ Test performance dengan 1000+ orders
5. ✅ Implement remaining P0 features (Fraud Detection, Reconciliation UI)

---

## 📞 Support

Jika ada masalah:
1. Cek backend logs di terminal
2. Cek frontend console (F12)
3. Cek network tab untuk API errors
4. Pastikan database connection OK
5. Restart backend jika perlu

**Status**: ✅ READY FOR TESTING
