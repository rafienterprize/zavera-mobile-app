# Stuck Payments - Admin Guide

## 📋 Apa itu "Stuck Payments"?

**Stuck Payments** adalah pembayaran yang sudah **lebih dari 1 jam** dalam status PENDING tapi belum ada update dari payment gateway (Midtrans).

### Contoh Kasus:
```
Order: ZVR-20260113-1CBE8BDA
Amount: Rp 914,000
Method: VA_BCA
Status: PENDING
Time: 110.8 hours (4+ hari!)
```

## 🔍 Kenapa Payment Bisa "Stuck"?

### 1. Customer Belum Bayar
- Customer create order tapi tidak jadi bayar
- VA number sudah expired
- Customer lupa atau berubah pikiran

### 2. Customer Sudah Bayar Tapi Sistem Belum Update
- Payment gateway webhook gagal
- Network issue saat callback
- Database tidak ter-update

### 3. Technical Issues
- Midtrans server down saat customer bayar
- Webhook URL tidak accessible
- Server restart saat payment processing

## 🎯 Cara Menangani Stuck Payments

### Step 1: Identifikasi di Dashboard

Dashboard menampilkan stuck payments dengan info:
- **Order Code**: ZVR-20260113-1CBE8BDA
- **Amount**: Rp 914,000
- **Method**: VA_BCA (Virtual Account BCA)
- **Hours Stuck**: 110.8h (sudah 4+ hari)

### Step 2: Klik "Check Order"

Button "Check Order" akan membawa Anda ke detail order:
```
/admin/orders/ZVR-20260113-1CBE8BDA
```

Di halaman ini Anda bisa lihat:
- Order details
- Payment information
- Customer contact
- Order timeline

### Step 3: Verifikasi Payment

**Option A: Check Midtrans Dashboard**
1. Login ke [Midtrans Dashboard](https://dashboard.midtrans.com)
2. Search order code: `ZVR-20260113-1CBE8BDA`
3. Check payment status:
   - **Pending**: Customer belum bayar
   - **Settlement**: Customer sudah bayar ✅
   - **Expire**: VA sudah expired
   - **Cancel**: Order dibatalkan

**Option B: Contact Customer**
1. Call/WhatsApp customer: `customer_phone`
2. Tanya: "Apakah sudah transfer?"
3. Jika sudah: Minta bukti transfer
4. Jika belum: Tanya apakah masih mau order

**Option C: Check Bank Statement**
1. Login ke internet banking
2. Check incoming transfers
3. Match amount dan tanggal
4. Verify customer name

### Step 4: Take Action

**Scenario 1: Customer Sudah Bayar**
```
Action: Update order status ke PAID
Steps:
1. Go to order detail page
2. Click "Update Status"
3. Select "PAID"
4. Add note: "Verified via bank statement"
5. Save
```

**Scenario 2: Customer Belum Bayar & Masih Mau Order**
```
Action: Create new order dengan payment baru
Steps:
1. Cancel old order
2. Create new order untuk customer
3. Generate new VA number
4. Send to customer
```

**Scenario 3: Customer Tidak Jadi Order**
```
Action: Cancel order
Steps:
1. Go to order detail page
2. Click "Cancel Order"
3. Reason: "Customer tidak jadi order"
4. Stock akan otomatis di-restore
```

**Scenario 4: Payment Expired**
```
Action: Cancel order
Steps:
1. Check VA expiry (biasanya 24 jam)
2. If expired: Cancel order
3. Reason: "Payment expired"
4. Stock akan otomatis di-restore
```

## 🚨 Critical Actions

### Payments Stuck > 24 Hours
**Priority: HIGH**
- Check Midtrans immediately
- Contact customer
- Resolve within 1 business day

### Payments Stuck > 72 Hours (3 days)
**Priority: CRITICAL**
- Assume customer tidak jadi
- Cancel order
- Restore stock
- Send cancellation email

### Payments Stuck > 7 Days
**Priority: AUTO-CANCEL**
- System should auto-cancel
- If not: Manual cancel immediately
- Free up stock for other customers

## 📊 Dashboard Features

### Payment Monitor Section

**Metrics Displayed:**
1. **Pending** (0): Orders awaiting payment (< 1 hour)
2. **Expiring Soon** (0): Orders < 1 hour before expiry
3. **Stuck** (1): Orders > 1 hour pending ⚠️
4. **Paid Today** (0): Successfully paid today

### Stuck Payments List

Shows up to 5 most critical stuck payments:
- Order code
- Payment method & bank
- Amount
- Hours stuck
- **"Check Order" button** → Direct link to order detail

### Action Buttons

**"Check Order"** button:
- Links to: `/admin/orders/{order_code}`
- Shows: Full order details
- Actions available:
  - Update status to PAID
  - Cancel order
  - Contact customer
  - View payment info

## 🔧 Manual Payment Verification

### For Virtual Account (VA)

**BCA:**
```
1. Login BCA internet banking
2. Menu: Informasi Rekening → Mutasi Rekening
3. Filter by date
4. Search amount: Rp 914,000
5. Check sender name matches customer
6. If found: Update order to PAID
```

**Mandiri:**
```
1. Login Mandiri internet banking
2. Menu: Rekening → Mutasi Rekening
3. Filter by date
4. Search amount
5. Verify and update
```

**BNI:**
```
1. Login BNI internet banking
2. Menu: Informasi → Mutasi Rekening
3. Filter by date
4. Search amount
5. Verify and update
```

### For QRIS

**Check Midtrans Dashboard:**
```
1. Login Midtrans
2. Transactions → Search order code
3. Check status
4. If settlement: Update to PAID
```

### For GoPay/E-Wallet

**Check Midtrans Dashboard:**
```
1. Login Midtrans
2. Transactions → Search order code
3. Check status
4. E-wallet usually auto-update
5. If stuck: Contact Midtrans support
```

## 🤖 Automated Solutions

### Payment Expiry Job

System automatically checks expired payments:
```go
// Runs every 1 minute
// Cancels orders with expired payments (> 24 hours)
// Restores stock automatically
```

### Payment Recovery Service

System tries to sync stuck payments:
```go
// Runs every 15 minutes
// Queries Midtrans for payment status
// Auto-updates if payment found
```

### Webhook Retry

If webhook fails:
```go
// Midtrans retries webhook
// Up to 5 times
// With exponential backoff
```

## 📈 Best Practices

### Daily Routine

**Morning (9 AM):**
1. Check dashboard for stuck payments
2. Verify payments > 24 hours
3. Contact customers if needed

**Afternoon (3 PM):**
1. Re-check stuck payments
2. Cancel expired orders
3. Update resolved payments

**Evening (6 PM):**
1. Final check before end of day
2. Resolve all critical issues
3. Prepare report for next day

### Weekly Review

**Every Monday:**
1. Review stuck payments from last week
2. Identify patterns (which bank/method)
3. Contact Midtrans if recurring issues
4. Update processes if needed

### Monthly Analysis

**End of Month:**
1. Count total stuck payments
2. Calculate resolution rate
3. Identify improvement areas
4. Update documentation

## 🎯 Success Metrics

### Target KPIs

**Stuck Payment Rate:**
```
Target: < 5% of total orders
Current: 1/13 = 7.7% (needs improvement)
```

**Resolution Time:**
```
Target: < 24 hours
Current: 110.8 hours (needs improvement)
```

**Auto-Resolution Rate:**
```
Target: > 80% (via webhook)
Current: Check logs
```

## 🔔 Alerts & Notifications

### Email Alerts (Future Enhancement)

**When to send:**
- Payment stuck > 24 hours
- Payment stuck > 72 hours
- Payment expired

**Who to notify:**
- Admin email
- Operations team
- Customer (if needed)

### SMS Alerts (Future Enhancement)

**Critical cases:**
- High-value orders (> Rp 1,000,000)
- VIP customers
- Urgent orders

## 📞 Customer Communication

### Template: Payment Reminder

```
Hi [Customer Name],

Kami melihat order Anda [ORDER_CODE] senilai [AMOUNT] 
masih dalam status pending.

Apakah Anda sudah melakukan pembayaran?

Jika sudah, mohon kirim bukti transfer ke WhatsApp kami.
Jika belum, silakan transfer ke:

Bank: [BANK]
VA Number: [VA_NUMBER]
Amount: [AMOUNT]

Terima kasih!
ZAVERA Team
```

### Template: Payment Confirmation

```
Hi [Customer Name],

Pembayaran Anda untuk order [ORDER_CODE] sudah kami terima!

Amount: [AMOUNT]
Status: PAID ✅

Order Anda akan segera diproses.

Terima kasih!
ZAVERA Team
```

### Template: Order Cancellation

```
Hi [Customer Name],

Order Anda [ORDER_CODE] telah dibatalkan karena 
pembayaran tidak diterima dalam 24 jam.

Jika Anda masih ingin order, silakan buat order baru.

Terima kasih!
ZAVERA Team
```

## ✅ Checklist: Handling Stuck Payment

- [ ] Identify stuck payment in dashboard
- [ ] Click "Check Order" button
- [ ] Review order details
- [ ] Check Midtrans dashboard
- [ ] Verify payment status
- [ ] Contact customer if needed
- [ ] Take appropriate action:
  - [ ] Update to PAID (if paid)
  - [ ] Cancel order (if not paid)
  - [ ] Create new order (if customer wants)
- [ ] Document resolution
- [ ] Update customer
- [ ] Monitor for similar issues

## 🎓 Training Resources

### For New Admins

1. **Read this guide** ✅
2. **Watch tutorial video** (if available)
3. **Shadow experienced admin** (1 day)
4. **Handle test case** (supervised)
5. **Handle real case** (supervised)
6. **Independent handling** (after approval)

### For Experienced Admins

1. **Review monthly** (refresh knowledge)
2. **Share best practices** (team meeting)
3. **Update documentation** (if process changes)
4. **Train new admins** (knowledge transfer)

## 📝 Summary

**Stuck Payments** = Pembayaran pending > 1 jam

**How to Handle:**
1. ✅ Check dashboard
2. ✅ Click "Check Order"
3. ✅ Verify payment
4. ✅ Take action (PAID/Cancel)
5. ✅ Update customer

**Goal:**
- Resolve within 24 hours
- Keep stuck rate < 5%
- Maintain customer satisfaction

**Remember:**
- Always verify before updating
- Document all actions
- Communicate with customer
- Learn from patterns

**Status**: ✅ READY TO USE
