# Stuck Payment Action Guide - IMPROVED ✅

## 🎯 Problem & Solution

**User Feedback:** "Mark as PAID itu tidak logis kalau customer belum bayar. Button ini harus jelas bahwa HANYA untuk customer yang SUDAH BAYAR."

**Before:** Button "Mark as PAID" kurang jelas, bisa disalahgunakan untuk order yang customer belum bayar.

**After:** UI diperbaiki dengan:
- ✅ Button label lebih jelas: **"Confirm Payment Received"**
- ✅ Warning merah tebal: **"ONLY if customer HAS PAID"**
- ✅ Dual buttons: **"Cancel Order (Recommended)"** vs **"Confirm Payment Received"**
- ✅ Verification checklist wajib diisi
- ✅ Quick links ke Midtrans & WhatsApp customer

## ✅ New Feature: Confirm Payment Received

### When It Appears

Button **"Confirm Payment Received"** muncul ketika:
1. Order status = **EXPIRED** atau **PENDING**
2. Payment status = **PENDING**
3. Payment sudah > 1 jam (stuck payment)

### What It Does

Mengupdate order status dari **EXPIRED/PENDING** → **PAID** setelah admin **VERIFIKASI** bahwa customer **SUDAH BAYAR**.

### ⚠️ CRITICAL: When to Use

**ONLY use this button when:**
- ✅ Customer HAS ACTUALLY PAID (verified via bank/Midtrans)
- ✅ Payment amount matches order total
- ✅ You have proof of payment (screenshot/bank statement)

**DO NOT use if:**
- ❌ Customer has NOT paid
- ❌ Payment not verified
- ❌ No proof of payment
- ❌ Amount doesn't match

**If customer has NOT paid → Use "Cancel Order" instead!**

## 📋 Step-by-Step: Menangani Stuck Payment

### Step 1: Dari Dashboard

1. Lihat **"Stuck Payments Detected"** section
2. Klik button **"Check Order"** pada payment yang stuck
3. Akan dibawa ke halaman order detail

### Step 2: Di Halaman Order Detail

Anda akan melihat **STUCK PAYMENT ALERT** yang prominent:

**Alert Banner (Red with border):**
```
⚠️ STUCK PAYMENT - Action Required
Order expired but payment still pending. 
You must verify if customer has actually paid before taking action.

📋 Verification Steps:
1. Check Midtrans dashboard for payment status
2. Verify bank statement shows incoming transfer
3. Contact customer to confirm payment
4. Verify amount matches order total: Rp 914,000

[🔍 Check Midtrans] [💬 WhatsApp Customer]

Decision Buttons:
┌─────────────────────────────────┬─────────────────────────────────┐
│ [Cancel Order (Recommended)]    │ [Confirm Payment Received]      │
│ Customer tidak bayar /          │ ⚠️ ONLY if customer HAS PAID    │
│ tidak jadi order                │                                 │
└─────────────────────────────────┴─────────────────────────────────┘
```

**Key Improvements:**
- ✅ **"Cancel Order"** is now the PRIMARY/RECOMMENDED action (red, prominent)
- ✅ **"Confirm Payment Received"** has clear warning: "ONLY if customer HAS PAID"
- ✅ Quick action buttons to verify payment (Midtrans, WhatsApp)
- ✅ Clear verification steps before taking action

### Step 3: Verifikasi Payment

**SEBELUM klik "Mark as PAID", WAJIB verifikasi:**

#### Option A: Check Midtrans Dashboard
```
1. Login https://dashboard.midtrans.com
2. Search: ZVR-20260113-1CBE8BDA
3. Check status:
   - Settlement ✅ → Customer sudah bayar
   - Pending ❌ → Customer belum bayar
   - Expire ❌ → VA sudah expired
```

#### Option B: Check Bank Statement
```
1. Login internet banking (BCA/Mandiri/BNI)
2. Menu: Mutasi Rekening
3. Filter tanggal order
4. Cari transfer Rp 914,000
5. Verify nama pengirim = customer name
```

#### Option C: Contact Customer
```
WhatsApp/Call: 6282141620950
Tanya: "Apakah sudah transfer untuk order ZVR-20260113-1CBE8BDA?"
Jika sudah: Minta bukti transfer (screenshot)
```

### Step 4: Confirm Payment Received

**Jika customer SUDAH BAYAR (verified):**

1. Klik button **"Confirm Payment Received"** (amber button)
2. Modal akan muncul dengan **CRITICAL WARNING**:
   ```
   ⚠️ Confirm Payment Received
   
   🚨 CRITICAL WARNING
   This action should ONLY be used when:
   • Customer HAS ACTUALLY PAID (verified via bank/Midtrans)
   • Payment amount matches order total: Rp 914,000
   • You have proof of payment (screenshot/bank statement)
   
   ⚠️ If customer has NOT paid, use "Cancel Order" instead!
   
   ✅ Verification Checklist (Complete ALL):
   ☐ Checked Midtrans dashboard - payment status is "Settlement"
   ☐ Verified bank statement shows incoming transfer
   ☐ Confirmed amount matches: Rp 914,000
   ☐ Contacted customer and received payment proof
   ```

3. Isi **Verification Details** (REQUIRED):
   ```
   Example:
   "Verified via BCA internet banking on 2026-01-13 at 14:30 WIB.
   Transfer received: Rp 914,000.
   Sender name: Sebastian Alexander (matches customer name).
   Screenshot saved in Google Drive folder: Payments/2026-01.
   Midtrans dashboard shows Settlement status."
   ```

4. Klik **"✅ Yes, Customer Has Paid"** (amber button)

5. Order status akan berubah:
   - EXPIRED → **PAID** ✅
   - Payment status → **PAID** ✅
   - Stock tetap reserved
   - Order siap diproses

**Important:** Verification details akan tersimpan di audit log untuk accountability.

### Step 5: Process Order

Setelah status PAID, lanjutkan normal flow:

1. **Proses Pesanan** → Status: PACKING
2. **Kirim Pesanan** (input resi) → Status: SHIPPED
3. **Tandai Selesai** → Status: DELIVERED

## 🚫 Jika Customer BELUM BAYAR

**Jika customer TIDAK JADI bayar:**

1. Klik button **"Cancel"** (red button)
2. Isi reason: "Customer tidak jadi order, payment expired"
3. Klik **"Confirm"**
4. Order akan dibatalkan
5. Stock otomatis di-restore

## 🎨 UI Features (IMPROVED)

### Alert Banner (Prominent Red)

Muncul di atas halaman jika stuck payment:
```
🔴 ⚠️ STUCK PAYMENT - Action Required
Order expired but payment still pending.
You must verify if customer has actually paid before taking action.

📋 Verification Steps:
1. Check Midtrans dashboard
2. Verify bank statement
3. Contact customer
4. Verify amount matches: Rp 914,000

[🔍 Check Midtrans] [💬 WhatsApp Customer]

┌─────────────────────────────────┬─────────────────────────────────┐
│ [Cancel Order (Recommended)]    │ [Confirm Payment Received]      │
│ RED BUTTON - PRIMARY ACTION     │ AMBER BUTTON - SECONDARY        │
│ Customer tidak bayar            │ ⚠️ ONLY if customer HAS PAID    │
└─────────────────────────────────┴─────────────────────────────────┘
```

### Dual Action Buttons

**Cancel Order (Recommended):**
- Color: Red (bg-red-500)
- Position: LEFT (primary position)
- Label: "Cancel Order (Recommended)"
- Subtitle: "Customer tidak bayar / tidak jadi order"
- Use when: Customer has NOT paid

**Confirm Payment Received:**
- Color: Amber (bg-amber-500)
- Position: RIGHT (secondary position)
- Label: "Confirm Payment Received"
- Subtitle: "⚠️ ONLY if customer HAS PAID"
- Use when: Customer HAS ACTUALLY PAID (verified)

### Verification Modal (IMPROVED)

```
Title: ⚠️ Confirm Payment Received

🚨 CRITICAL WARNING (Red box with border)
This action should ONLY be used when:
• Customer HAS ACTUALLY PAID (verified via bank/Midtrans)
• Payment amount matches order total: Rp 914,000
• You have proof of payment (screenshot/bank statement)

⚠️ If customer has NOT paid, use "Cancel Order" instead!

✅ Verification Checklist (Complete ALL):
☐ Checked Midtrans dashboard - payment status is "Settlement"
☐ Verified bank statement shows incoming transfer
☐ Confirmed amount matches: Rp 914,000
☐ Contacted customer and received payment proof

Verification Details * (REQUIRED)
[Textarea with amber border - must be filled]
Include: verification method, date/time, amount, sender name, proof location

[Batal] [✅ Yes, Customer Has Paid]
```

### Success Flow

```
1. Click "Mark as PAID"
2. Fill verification details
3. Click "Confirm Payment"
4. ✅ Order updated to PAID
5. Page refreshes
6. Alert banner disappears
7. "Proses Pesanan" button appears
```

## 📊 Complete Workflow

### Scenario 1: Customer Sudah Bayar (Happy Path)

```
Dashboard → Stuck Payment Alert
    ↓
Click "Check Order"
    ↓
Order Detail Page (EXPIRED, Payment PENDING)
    ↓
Verify Payment (Midtrans/Bank)
    ↓
Click "Mark as PAID"
    ↓
Fill Verification Details
    ↓
Confirm Payment
    ↓
✅ Order Status: PAID
    ↓
Click "Proses Pesanan"
    ↓
Click "Kirim Pesanan" (input resi)
    ↓
✅ Order Status: SHIPPED
    ↓
Customer receives order
    ↓
Click "Tandai Selesai"
    ↓
✅ Order Status: DELIVERED
```

### Scenario 2: Customer Belum Bayar

```
Dashboard → Stuck Payment Alert
    ↓
Click "Check Order"
    ↓
Order Detail Page (EXPIRED, Payment PENDING)
    ↓
Verify Payment (Midtrans/Bank)
    ↓
❌ No payment found
    ↓
Contact Customer (optional)
    ↓
Customer tidak jadi order
    ↓
Click "Cancel"
    ↓
Fill Reason
    ↓
Confirm Cancel
    ↓
✅ Order Status: CANCELLED
✅ Stock Restored
```

### Scenario 3: Customer Mau Order Lagi

```
Dashboard → Stuck Payment Alert
    ↓
Click "Check Order"
    ↓
Order Detail Page (EXPIRED, Payment PENDING)
    ↓
Contact Customer
    ↓
Customer: "Saya masih mau order"
    ↓
Click "Cancel" (old order)
    ↓
Create New Order (manual/via system)
    ↓
Send new VA number to customer
    ↓
Customer pays new order
    ↓
✅ New order processed normally
```

## 🔒 Security & Validation

### Backend Validation

```go
// Only allow PENDING/EXPIRED orders to be marked as PAID
if order.Status != "PENDING" && order.Status != "EXPIRED" {
    return error("Cannot mark this order as paid")
}

// Require admin authentication
if !isAdmin(user) {
    return error("Unauthorized")
}

// Require reason/verification details
if reason == "" {
    return error("Verification details required")
}

// Record audit log
auditLog.Record(adminEmail, "MARK_AS_PAID", orderCode, reason)
```

### Frontend Validation

```typescript
// Button only shows for stuck payments
const canMarkAsPaid = 
  (order.status === "PENDING" || order.status === "EXPIRED") && 
  order.payment?.status === "PENDING";

// Require verification details
disabled={!actionReason.trim() || actionLoading !== null}

// Show verification checklist
⚠️ Verification Checklist (in modal)
```

## 📝 Audit Trail

Setiap action "Mark as PAID" akan tercatat di:

### Admin Audit Log

```
Admin: admin@zavera.com
Action: MARK_AS_PAID
Order: ZVR-20260113-1CBE8BDA
Reason: "Verified via BCA statement, transfer received..."
Timestamp: 2026-01-19 15:30:00
```

### Order Status History

```
EXPIRED → PAID
Changed by: admin@zavera.com
Reason: "Verified via BCA statement..."
Timestamp: 2026-01-19 15:30:00
```

## 🎯 Best Practices

### DO ✅

1. **Always verify** payment before marking as PAID
2. **Check Midtrans** dashboard first
3. **Verify bank statement** if needed
4. **Contact customer** if unsure
5. **Document verification** in reason field
6. **Include details**: date, time, amount, source
7. **Double-check amount** matches order total

### DON'T ❌

1. ❌ Mark as PAID without verification
2. ❌ Trust customer word only (need proof)
3. ❌ Skip checking Midtrans
4. ❌ Leave reason field empty
5. ❌ Mark wrong order as paid
6. ❌ Forget to process order after marking paid

## 📞 Customer Communication

### After Marking as PAID

Send WhatsApp/Email:
```
Hi Sebastian,

Pembayaran Anda untuk order ZVR-20260113-1CBE8BDA 
sudah kami terima dan verifikasi! ✅

Amount: Rp 914,000
Status: PAID

Pesanan Anda akan segera kami proses dan kirim.
Anda akan menerima nomor resi via email/WhatsApp.

Terima kasih!
ZAVERA Team
```

### If Cancelling

Send WhatsApp/Email:
```
Hi Sebastian,

Order Anda ZVR-20260113-1CBE8BDA telah dibatalkan 
karena pembayaran tidak diterima dalam waktu yang ditentukan.

Jika Anda masih ingin order, silakan buat order baru 
di website kami.

Terima kasih!
ZAVERA Team
```

## ✅ Success Metrics

### Before Fix
- ❌ Admin bingung apa yang harus dilakukan
- ❌ Stuck payments tidak ter-resolve
- ❌ Customer complain order tidak diproses
- ❌ Stock terkunci di order expired

### After Fix
- ✅ Clear action button "Mark as PAID"
- ✅ Verification checklist membantu admin
- ✅ Stuck payments ter-resolve cepat
- ✅ Customer happy order diproses
- ✅ Stock management lebih baik

## 🎓 Training Checklist

For new admins:

- [ ] Understand what is "stuck payment"
- [ ] Know how to access Midtrans dashboard
- [ ] Know how to check bank statement
- [ ] Practice verifying payment
- [ ] Practice marking order as PAID
- [ ] Practice cancelling expired order
- [ ] Know how to contact customer
- [ ] Understand audit trail importance

## 📊 Summary

**Problem**: Button "Check Order" tidak ada action untuk resolve stuck payment

**Solution**: Added **"Mark as PAID"** button dengan:
- ✅ Verification checklist
- ✅ Reason/details required
- ✅ Audit trail
- ✅ Clear UI/UX
- ✅ Security validation

**Result**: Admin sekarang bisa:
1. Identify stuck payment di dashboard
2. Click "Check Order"
3. Verify payment (Midtrans/Bank)
4. Click "Mark as PAID"
5. Fill verification details
6. Confirm → Order updated to PAID
7. Process order normally

**Status**: ✅ COMPLETE & READY TO USE

Sekarang button "Check Order" **TIDAK PERCUMA** lagi! Ada action yang jelas untuk menyelesaikan stuck payments! 🎉
