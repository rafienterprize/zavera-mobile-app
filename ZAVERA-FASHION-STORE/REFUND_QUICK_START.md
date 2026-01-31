# Refund System - Quick Start Guide

## 🚀 Quick Setup (5 Minutes)

### Step 1: Database Migration
```bash
psql -h localhost -U postgres -d zavera_db -f database/migrate_refund_enhancement.sql
```

### Step 2: Start Backend
```bash
cd backend
.\zavera.exe
```

### Step 3: Start Frontend (New Terminal)
```bash
cd frontend
npm run dev
```

## 🎯 Quick Test Flow

### Test 1: Admin Create Refund (2 minutes)

1. **Login Admin**
   - Go to: http://localhost:3000/login
   - Login dengan admin email

2. **Find Order**
   - Go to: http://localhost:3000/admin/orders
   - Pilih order dengan status DELIVERED atau COMPLETED

3. **Create Refund**
   - Klik tombol "Refund" (orange button)
   - Pilih refund type: **FULL**
   - Pilih reason: **Customer Request**
   - Isi detail: "Testing refund system"
   - Klik "Process Refund"

4. **Verify Success**
   - ✅ Modal shows success message
   - ✅ Refund appears in history section
   - ✅ Status shows "Selesai" (green badge)
   - ✅ Gateway refund ID displayed

### Test 2: Customer View Refund (1 minute)

1. **Login Customer**
   - Logout admin
   - Login dengan customer yang punya order di-refund

2. **View Purchase History**
   - Go to: http://localhost:3000/account/pembelian?tab=history
   - ✅ Order shows "Dikembalikan" badge (orange)
   - ✅ Refund notice displayed

3. **View Refund Details**
   - Klik "Lihat Detail Transaksi"
   - ✅ "Informasi Pengembalian Dana" section visible
   - ✅ Shows refund amount, status, timeline
   - ✅ Shows reason & completion date

## 🧪 Test Scenarios

### Scenario A: Full Refund (Paid Order)
```
Order Status: DELIVERED
Payment: SUCCESS (BCA VA)
Refund Type: FULL
Expected: ✅ Refund COMPLETED, Order status → REFUNDED, Stock restored
```

### Scenario B: Manual Refund (No Payment)
```
Order Status: COMPLETED
Payment: NULL (manual order)
Refund Type: FULL
Expected: ✅ Refund COMPLETED immediately, Gateway ID = "MANUAL_REFUND"
```

### Scenario C: Partial Refund (Item Only)
```
Order Status: DELIVERED
Payment: SUCCESS
Refund Type: ITEM_ONLY
Items: Select 1-2 items
Expected: ✅ Refund COMPLETED, Only selected items stock restored
```

### Scenario D: Shipping Only Refund
```
Order Status: DELIVERED
Payment: SUCCESS
Refund Type: SHIPPING_ONLY
Expected: ✅ Refund COMPLETED, No stock restoration
```

## 🔍 Verification Checklist

### Backend Verification
```bash
# Check backend is running
curl http://localhost:8080/health

# Should return: {"status":"ok"}
```

### Database Verification
```sql
-- Check refunds table
SELECT * FROM refunds ORDER BY created_at DESC LIMIT 5;

-- Check refund history
SELECT * FROM refund_status_history ORDER BY changed_at DESC LIMIT 10;

-- Check orders with refunds
SELECT order_code, status, refund_status, refund_amount 
FROM orders 
WHERE refund_status IS NOT NULL;
```

### API Verification
```bash
# Test admin refund endpoint (replace {token} and {order_code})
curl -X POST http://localhost:8080/api/admin/refunds \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{
    "order_code": "{order_code}",
    "refund_type": "FULL",
    "reason": "Customer Request",
    "reason_detail": "Testing",
    "idempotency_key": "test-123"
  }'
```

## 🐛 Troubleshooting

### Issue: Migration Failed
```bash
# Check if already applied
psql -h localhost -U postgres -d zavera_db -c "\d refunds"

# If columns exist, migration is done
```

### Issue: Backend Won't Start
```bash
# Check port 8080 is free
netstat -ano | findstr :8080

# Kill process if needed
taskkill /PID {pid} /F

# Rebuild
cd backend
go clean
go build
.\zavera.exe
```

### Issue: Refund Button Not Showing
- ✅ Check order status is DELIVERED or COMPLETED
- ✅ Check you're logged in as admin
- ✅ Refresh page (Ctrl+F5)

### Issue: Refund Creation Failed
- ✅ Check order has payment record (or use manual refund)
- ✅ Check refund amount <= order total
- ✅ Check Midtrans credentials in backend/.env
- ✅ Check backend console for error logs

## 📊 Expected Results

### After Full Refund:
- ✅ Refund record created with status COMPLETED
- ✅ Order status changed to REFUNDED
- ✅ Order refund_status = "FULL"
- ✅ Order refund_amount = order total
- ✅ Product stock restored
- ✅ Refund history recorded
- ✅ Customer can see refund in UI

### After Partial Refund:
- ✅ Refund record created with status COMPLETED
- ✅ Order status unchanged (still DELIVERED)
- ✅ Order refund_status = "PARTIAL"
- ✅ Order refund_amount = refund amount
- ✅ Selected items stock restored
- ✅ Customer can see refund in UI

## 🎓 Key Features to Test

1. **Refund Types**
   - [ ] FULL refund
   - [ ] PARTIAL refund
   - [ ] SHIPPING_ONLY refund
   - [ ] ITEM_ONLY refund

2. **Payment Methods**
   - [ ] BCA VA
   - [ ] BNI VA
   - [ ] BRI VA
   - [ ] Mandiri VA
   - [ ] Manual (no payment)

3. **Admin Features**
   - [ ] Create refund
   - [ ] View refund history
   - [ ] Retry failed refund
   - [ ] View refund details

4. **Customer Features**
   - [ ] View refund badge in purchase history
   - [ ] View refund details in order page
   - [ ] See timeline estimates
   - [ ] See status messages

## 📝 Notes

- **Sandbox Mode**: Midtrans sandbox auto-approves refunds
- **Manual Refunds**: Orders without payment skip gateway
- **Idempotency**: Same idempotency_key returns same refund
- **Stock Restoration**: Automatic based on refund type
- **Audit Trail**: All changes recorded in refund_status_history

## ✅ Success Criteria

Your refund system is working correctly if:
- ✅ Admin can create refunds via UI
- ✅ Refunds process to Midtrans successfully
- ✅ Order status updates correctly
- ✅ Stock restores correctly
- ✅ Customer can view refund information
- ✅ Failed refunds can be retried
- ✅ Manual refunds work without payment

## 🎉 You're Ready!

Refund system is now fully functional and ready for testing in sandbox environment. Enjoy! 🚀

---

**Need Help?** Check `REFUND_SYSTEM_DEPLOYMENT_GUIDE.md` for detailed documentation.
