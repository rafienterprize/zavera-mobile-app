# EXPIRED Order Logic Fix

## Problem
Order EXPIRED masih menampilkan action buttons (Cancel, Mark as Paid), padahal EXPIRED artinya pembayaran tidak dilakukan dan order sudah final.

## Correct Business Logic

### Order Status Flow:
```
PENDING → (customer pays) → PAID → PACKING → SHIPPED → DELIVERED → COMPLETED
   ↓
   (24h timeout, no payment)
   ↓
EXPIRED (FINAL - no action needed)
```

### Status Definitions:
- **PENDING**: Menunggu pembayaran customer → **Perlu action** (cancel/mark paid)
- **EXPIRED**: Pembayaran tidak dilakukan, sudah lewat waktu → **Sudah final, tidak perlu action**
- **CANCELLED**: Admin/customer cancel → **Sudah final**
- **FAILED**: Payment failed → **Sudah final**

## Changes Made

### 1. Frontend - Order Detail Page
**File:** `frontend/src/app/admin/orders/[code]/page.tsx`

**Before:**
```typescript
// SALAH - EXPIRED bisa di-cancel
const canCancel = ["PENDING", "PAID", "PACKING", "EXPIRED"].includes(order.status);
const isStuckPayment = order.status === "EXPIRED";
```

**After:**
```typescript
// BENAR - EXPIRED tidak bisa di-cancel
const canCancel = ["PENDING", "PAID", "PACKING"].includes(order.status);
const isStuckPayment = order.status === "PENDING" && order.payment?.status === "PENDING";
const canMarkAsPaid = order.status === "PENDING" && order.payment?.status === "PENDING";
```

**UI Changes:**
- **PENDING orders**: Show amber alert banner + action buttons (Cancel/Mark Paid)
- **EXPIRED orders**: Show gray info banner "Order Expired - No action needed"
- **No action buttons** for EXPIRED orders

### 2. Backend - Dashboard Stuck Payments
**File:** `backend/service/admin_dashboard_service.go`

**Before:**
```sql
-- SALAH - Menampilkan EXPIRED orders
WHERE o.status IN ('PENDING', 'EXPIRED')
```

**After:**
```sql
-- BENAR - Hanya PENDING orders
WHERE o.status = 'PENDING'
```

**Result:** Dashboard "Stuck Payments" hanya menampilkan order PENDING yang perlu action.

## UI Behavior

### PENDING Order (Need Action)
```
┌─────────────────────────────────────────────────┐
│ ⚠️ Payment Pending - Waiting for customer      │
│    [Check Midtrans] [WhatsApp]                 │
└─────────────────────────────────────────────────┘

┌─ Order Actions ─────────────────────────────────┐
│ ✓ Verification Steps:                           │
│   1. Check Midtrans dashboard                   │
│   2. Verify bank statement                      │
│   3. Contact customer                           │
│   4. Verify amount: Rp 914.000                  │
│                                                  │
│ [Cancel Order]                                  │
│ [Confirm Payment]                               │
│ ⚠️ Only if customer HAS PAID                    │
└─────────────────────────────────────────────────┘
```

### EXPIRED Order (No Action)
```
┌─────────────────────────────────────────────────┐
│ ⓧ Order Expired - Payment not completed        │
│    This order has expired and is automatically  │
│    closed. No action needed.                    │
└─────────────────────────────────────────────────┘

┌─ Customer ──────────────────────────────────────┐
│ Name: Sebastian Alexander                       │
│ Email: sebastian@gmail.com                      │
│ Phone: 628214162095                             │
└─────────────────────────────────────────────────┘

(No Order Actions card shown)
```

## Dashboard Behavior

### Stuck Payments Section
**Before:** Shows both PENDING and EXPIRED orders
**After:** Only shows PENDING orders that need manual verification

**Query Logic:**
```sql
WHERE op.payment_status = 'PENDING'
AND o.status = 'PENDING'              -- Only PENDING, not EXPIRED
AND op.created_at < NOW() - INTERVAL '1 hour'
```

## Files Modified

1. **`frontend/src/app/admin/orders/[code]/page.tsx`**
   - Updated `canCancel` to exclude EXPIRED
   - Updated `isStuckPayment` to only PENDING
   - Updated `canMarkAsPaid` to only PENDING
   - Added EXPIRED info banner
   - Changed PENDING alert banner color to amber

2. **`backend/service/admin_dashboard_service.go`**
   - Updated stuck payments query to only show PENDING orders
   - Line ~190: Changed `IN ('PENDING', 'EXPIRED')` to `= 'PENDING'`

## Testing Checklist

- [x] Backend compiled successfully
- [ ] Restart backend
- [ ] Test PENDING order → Should show action buttons
- [ ] Test EXPIRED order → Should NOT show action buttons
- [ ] Test EXPIRED order → Should show gray info banner
- [ ] Dashboard → Should only show PENDING in stuck payments
- [ ] Dashboard → EXPIRED orders should not appear

## Business Rules Summary

| Status | Can Cancel? | Can Mark Paid? | Show Actions? | Dashboard Stuck? |
|--------|-------------|----------------|---------------|------------------|
| PENDING | ✅ Yes | ✅ Yes | ✅ Yes | ✅ Yes (if >1h) |
| EXPIRED | ❌ No | ❌ No | ❌ No | ❌ No |
| PAID | ✅ Yes | ❌ No | ✅ Yes | ❌ No |
| CANCELLED | ❌ No | ❌ No | ❌ No | ❌ No |
| FAILED | ❌ No | ❌ No | ❌ No | ❌ No |

## Notes
- EXPIRED orders are automatically set by backend job after 24 hours
- EXPIRED is a final status - no manual intervention needed
- Admin should only focus on PENDING orders in dashboard
- EXPIRED orders remain in database for record keeping
