# ZAVERA Refund System - Complete Guide

**Status:** ✅ **FULLY FUNCTIONAL**  
**Date:** January 29, 2026

---

## 🎯 Overview

Refund system ZAVERA sekarang **fully functional** dengan support untuk:
- ✅ Automatic refund via Midtrans
- ✅ Manual refund processing (untuk error 418)
- ✅ Multiple refund types (FULL, PARTIAL, SHIPPING_ONLY, ITEM_ONLY)
- ✅ Stock restoration
- ✅ Audit trail
- ✅ Retry failed refunds
- ✅ Mark as completed manually

---

## 🔄 Refund Flow

### Scenario 1: Automatic Refund (Success)

```
1. Admin clicks "Refund" button
   ↓
2. Fill refund form (type, reason, details)
   ↓
3. Click "Process Refund"
   ↓
4. System creates refund (status: PENDING)
   ↓
5. System calls Midtrans API
   ↓
6. Midtrans processes refund ✅
   ↓
7. Status changes to COMPLETED
   ↓
8. Stock restored automatically
   ↓
9. Customer receives refund
```

### Scenario 2: Manual Processing Required (Error 418)

```
1. Admin clicks "Refund" button
   ↓
2. Fill refund form (type, reason, details)
   ↓
3. Click "Process Refund"
   ↓
4. System creates refund (status: PENDING)
   ↓
5. System calls Midtrans API
   ↓
6. Midtrans returns Error 418 ⚠️
   (Payment provider requires settlement time)
   ↓
7. Status stays PENDING
   ↓
8. Error message shown:
   "MANUAL_PROCESSING_REQUIRED: Automatic refund failed.
    Please process manual bank transfer to customer and
    mark refund as completed after transfer is done."
   ↓
9. Admin sees "Mark as Completed" button
   ↓
10. Admin processes manual bank transfer to customer
   ↓
11. Admin clicks "Mark as Completed"
   ↓
12. Confirmation dialog appears
   ↓
13. Admin enters note (e.g., "Transfer manual via BCA...")
   ↓
14. Status changes to COMPLETED
   ↓
15. Stock restored automatically
   ↓
16. Refund complete ✅
```

### Scenario 3: Failed Refund (Retry)

```
1. Refund fails (status: FAILED)
   ↓
2. Admin sees "Retry Refund" button
   ↓
3. Admin clicks "Retry Refund"
   ↓
4. Confirmation dialog appears
   ↓
5. System retries with Midtrans
   ↓
6. If success → COMPLETED
   If error 418 → PENDING (manual processing)
   If other error → FAILED (can retry again)
```

---

## 📋 How to Use

### Step 1: Access Order Detail

1. Go to **Admin Panel** → **Orders**
2. Click on an order with status **DELIVERED** or **PAID**
3. Order detail page opens

### Step 2: Initiate Refund

1. Scroll to **Order Actions** section
2. Click **"Refund"** button
3. Refund modal opens

### Step 3: Fill Refund Form

#### Refund Type Options:

**1. FULL** - Refund semua (items + shipping)
- Automatically calculates total amount
- Refunds entire order

**2. PARTIAL** - Refund sebagian amount
- Enter custom amount
- Must be ≤ refundable balance

**3. SHIPPING_ONLY** - Refund ongkir saja
- Only refunds shipping cost
- Items not refunded

**4. ITEM_ONLY** - Refund items tertentu
- Select which items to refund
- Adjust quantities
- Shipping not refunded

#### Reason Options:
- Customer Request
- Damaged Product
- Wrong Item
- Quality Issue
- Late Delivery
- Other

#### Additional Details:
- Optional text field
- Add more context about refund

### Step 4: Process Refund

1. Click **"Process Refund"** button
2. System creates refund
3. System attempts automatic refund

**If Successful:**
- ✅ Success message shown
- ✅ Refund status: COMPLETED
- ✅ Stock restored
- ✅ Done!

**If Error 418 (Manual Processing Required):**
- ⚠️ Error message shown with instructions
- ⚠️ Refund status: PENDING
- ⚠️ "Mark as Completed" button appears
- ⚠️ Continue to Step 5

### Step 5: Manual Processing (If Required)

1. **Process Manual Bank Transfer:**
   - Get customer bank details
   - Transfer refund amount manually
   - Keep proof of transfer

2. **Mark as Completed:**
   - Click **"Mark as Completed"** button
   - Confirmation dialog appears
   - Enter note with transfer details:
     ```
     Example: "Transfer manual via BCA ke rekening customer 
     1234567890 a.n. John Doe pada 29 Jan 2026 pukul 14:30"
     ```
   - Click **"Confirm"**

3. **Verification:**
   - ✅ Status changes to COMPLETED
   - ✅ Stock restored
   - ✅ Gateway ID: "MANUAL_BANK_TRANSFER"
   - ✅ Note saved in audit trail

---

## 🎨 UI Elements

### Refund History Section

Shows all refunds for an order:

```
┌─────────────────────────────────────────────────┐
│ 🔄 Refund History                               │
├─────────────────────────────────────────────────┤
│ REF-20260129-ABC123  [COMPLETED] [FULL]        │
│ Reason: Customer Request - Changed mind        │
│                                    Rp 918,000   │
│ Gateway ID: 12345678                            │
│                                                 │
│ Items: Rp 900,000                               │
│ Shipping: Rp 18,000                             │
│                                                 │
│ Requested: 29 Jan 2026 14:30                    │
│ Completed: 29 Jan 2026 14:35                    │
└─────────────────────────────────────────────────┘
```

### Pending Refund with Manual Processing

```
┌─────────────────────────────────────────────────┐
│ REF-20260129-XYZ789  [PENDING] [FULL]          │
│ Reason: Customer Request                        │
│                                    Rp 918,000   │
│ ⚠️ MANUAL REFUND                                │
│                                                 │
│ [✓ Mark as Completed]                           │
└─────────────────────────────────────────────────┘
```

### Failed Refund with Retry

```
┌─────────────────────────────────────────────────┐
│ REF-20260129-DEF456  [FAILED] [FULL]           │
│ Reason: Customer Request                        │
│                                    Rp 918,000   │
│ Gateway ID: Error - Connection timeout          │
│                                                 │
│ [🔄 Retry Refund]                               │
└─────────────────────────────────────────────────┘
```

---

## 🔧 Technical Details

### Backend Endpoints

**Create Refund:**
```
POST /api/admin/refunds
Authorization: Bearer {token}

Body:
{
  "order_code": "ZVR-20260129-ABC123",
  "refund_type": "FULL",
  "reason": "CUSTOMER_REQUEST",
  "reason_detail": "Changed mind",
  "idempotency_key": "unique-key"
}

Response:
{
  "id": 1,
  "refund_code": "REF-20260129-ABC123",
  "status": "PENDING",
  "refund_amount": 918000,
  ...
}
```

**Process Refund:**
```
POST /api/admin/refunds/:id/process
Authorization: Bearer {token}

Response (Success):
{
  "success": true,
  "message": "Refund processed successfully",
  "refund_code": "REF-20260129-ABC123",
  "gateway_refund_id": "12345678"
}

Response (Error 418):
{
  "error": "GATEWAY_ERROR",
  "message": "Payment gateway error",
  "details": {
    "error": "MANUAL_PROCESSING_REQUIRED: Automatic refund failed...",
    "refund_id": 1
  }
}
```

**Mark as Completed:**
```
POST /api/admin/refunds/:id/mark-completed
Authorization: Bearer {token}

Body:
{
  "note": "Transfer manual via BCA ke rekening customer..."
}

Response:
{
  "success": true,
  "message": "Refund marked as completed successfully",
  "refund_code": "REF-20260129-ABC123",
  "gateway_refund_id": "MANUAL_BANK_TRANSFER"
}
```

**Retry Refund:**
```
POST /api/admin/refunds/:id/retry
Authorization: Bearer {token}

Response:
{
  "success": true,
  "message": "Refund retry completed successfully",
  "refund_code": "REF-20260129-ABC123",
  "gateway_refund_id": "12345678"
}
```

### Frontend Implementation

**Error Handling:**
```typescript
try {
  const processResponse = await api.post(
    `/admin/refunds/${response.data.id}/process`,
    {},
    { headers: { Authorization: `Bearer ${token}` } }
  );
} catch (processError: any) {
  const errorMsg = processError.response?.data?.message || '';
  
  // Check for manual processing required
  if (errorMsg.includes('MANUAL_PROCESSING_REQUIRED')) {
    setRefundError('MANUAL_PROCESSING_REQUIRED: ...');
    // Show refund in PENDING state with Mark as Completed button
    return;
  }
  
  throw processError;
}
```

**Mark as Completed:**
```typescript
const handleMarkRefundCompleted = async (refundId: number) => {
  // Show confirmation
  setConfirmConfig({
    title: 'Mark Refund as Completed',
    message: 'Apakah Anda sudah melakukan transfer manual...',
    onConfirm: async () => {
      // Prompt for note
      const note = prompt('Masukkan catatan konfirmasi...');
      
      // Call API
      await api.post(`/admin/refunds/${refundId}/mark-completed`, {
        note: note.trim()
      }, {
        headers: { Authorization: `Bearer ${token}` }
      });
      
      // Reload data
      loadRefunds();
      loadOrder();
    }
  });
};
```

---

## 🧪 Testing Guide

### Test Case 1: Successful Automatic Refund

**Prerequisites:**
- Order with status DELIVERED
- Payment settled > 24 hours ago

**Steps:**
1. Go to order detail
2. Click "Refund"
3. Select "FULL"
4. Select reason "Customer Request"
5. Click "Process Refund"

**Expected Result:**
- ✅ Success message shown
- ✅ Refund status: COMPLETED
- ✅ Gateway ID populated
- ✅ Stock restored

### Test Case 2: Manual Processing (Error 418)

**Prerequisites:**
- Order with status DELIVERED
- Payment settled < 24 hours ago (or test environment)

**Steps:**
1. Go to order detail
2. Click "Refund"
3. Select "FULL"
4. Select reason "Customer Request"
5. Click "Process Refund"

**Expected Result:**
- ⚠️ Error message shown with "MANUAL_PROCESSING_REQUIRED"
- ⚠️ Refund status: PENDING
- ⚠️ "Mark as Completed" button visible

**Continue:**
6. Click "Mark as Completed"
7. Confirm dialog appears
8. Enter note: "Test manual transfer"
9. Click confirm

**Expected Result:**
- ✅ Success message shown
- ✅ Refund status: COMPLETED
- ✅ Gateway ID: "MANUAL_BANK_TRANSFER"
- ✅ Stock restored

### Test Case 3: Partial Refund

**Steps:**
1. Go to order detail
2. Click "Refund"
3. Select "PARTIAL"
4. Enter amount: 500000
5. Select reason
6. Click "Process Refund"

**Expected Result:**
- ✅ Refund created with amount 500000
- ✅ Refundable balance updated

### Test Case 4: Item-Only Refund

**Steps:**
1. Go to order detail
2. Click "Refund"
3. Select "ITEM_ONLY"
4. Select items and quantities
5. Select reason
6. Click "Process Refund"

**Expected Result:**
- ✅ Only selected items refunded
- ✅ Shipping not refunded
- ✅ Stock restored for refunded items only

### Test Case 5: Retry Failed Refund

**Prerequisites:**
- Refund with status FAILED

**Steps:**
1. Go to order detail
2. Find failed refund in history
3. Click "Retry Refund"
4. Confirm dialog

**Expected Result:**
- ✅ Refund retried
- ✅ Status updated based on result

---

## 📊 Refund Status Flow

```
PENDING
  ↓
  ├─→ [Process] → PROCESSING
  │                  ↓
  │                  ├─→ [Success] → COMPLETED ✅
  │                  ├─→ [Error 418] → PENDING ⚠️
  │                  └─→ [Other Error] → FAILED ❌
  │
  └─→ [Mark Completed] → COMPLETED ✅

FAILED
  └─→ [Retry] → PROCESSING
                   ↓
                   (same as above)
```

---

## ⚠️ Important Notes

### For Admin:

1. **Always verify payment before marking as completed**
   - Check bank statement
   - Verify transfer proof
   - Confirm with customer

2. **Enter detailed notes**
   - Include transfer date and time
   - Include bank name
   - Include account details
   - Include reference number

3. **Stock restoration is automatic**
   - Don't manually adjust stock
   - System handles it automatically

4. **Refundable balance**
   - Cannot exceed order total
   - Previous refunds deducted
   - Shipping included in balance

### For Developers:

1. **Error 418 handling**
   - Keep status as PENDING
   - Don't mark as FAILED
   - Show manual processing option

2. **Idempotency**
   - Use unique idempotency keys
   - Prevents duplicate refunds
   - Format: `{order_code}-{timestamp}`

3. **Stock restoration**
   - Happens on COMPLETED status
   - Restores to product_variants
   - Updates product.stock
   - Logged in audit trail

4. **Audit trail**
   - All actions logged
   - Status changes tracked
   - User actions recorded
   - Notes preserved

---

## 🎉 Success Indicators

### Refund is Successful When:

1. ✅ Status is COMPLETED
2. ✅ Gateway ID populated (or "MANUAL_BANK_TRANSFER")
3. ✅ Stock restored (stock_restored = true)
4. ✅ Order refund_status updated
5. ✅ Order refund_amount updated
6. ✅ Audit trail recorded
7. ✅ Customer receives money

---

## 🔍 Troubleshooting

### Issue: Refund button not showing

**Cause:** Order status not eligible  
**Solution:** Order must be DELIVERED or PAID

### Issue: Error 418 every time

**Cause:** Payment too recent  
**Solution:** Use manual processing flow

### Issue: Mark as Completed not working

**Cause:** Missing note or wrong status  
**Solution:** 
- Ensure note is provided
- Refund must be PENDING status

### Issue: Stock not restored

**Cause:** Refund not COMPLETED  
**Solution:** Complete refund first, stock restores automatically

### Issue: Cannot retry failed refund

**Cause:** Refund not in FAILED status  
**Solution:** Only FAILED refunds can be retried

---

## 📚 Related Documentation

- **REFUND_ERROR_418_SOLUTION.md** - Error 418 handling
- **REFUND_SYSTEM_README.md** - System overview
- **REFUND_TESTING_GUIDE.md** - Testing procedures
- **API_DOCS.md** - API reference

---

## ✅ Checklist for Demo

Before demo, verify:

- [ ] Backend running (zavera_size_filter.exe)
- [ ] Frontend running (npm run dev)
- [ ] Database connected
- [ ] Test order with DELIVERED status exists
- [ ] Admin can login
- [ ] Refund button visible
- [ ] Can create refund
- [ ] Can see refund history
- [ ] Mark as Completed button works
- [ ] Success messages show correctly

---

## 🎊 Conclusion

**Refund system is now FULLY FUNCTIONAL!**

✅ Automatic refund via Midtrans  
✅ Manual processing for error 418  
✅ Multiple refund types supported  
✅ Stock restoration automatic  
✅ Complete audit trail  
✅ Retry mechanism for failures  
✅ User-friendly UI  
✅ Clear error messages  
✅ Production-ready  

**System ready for demo and production use!** 🚀

---

**Last Updated:** January 29, 2026  
**Status:** ✅ COMPLETE  
**Version:** 1.0.0
