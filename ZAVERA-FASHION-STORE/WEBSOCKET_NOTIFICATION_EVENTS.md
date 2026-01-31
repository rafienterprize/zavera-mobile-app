# WebSocket Notification Events - Implementation Status

## Overview
Real-time notifications untuk admin dashboard via WebSocket. Admin akan mendapat notifikasi instant untuk semua event penting.

## Implemented Events ✅

### 1. Order Created ✅
**Trigger:** `backend/service/checkout_service.go` line 313
```go
NotifyOrderCreated(order.OrderCode, customerName, order.TotalAmount)
```
**When:** Saat customer selesai checkout dan order dibuat
**Notification:** "🛍️ New Order - {customer} placed an order #{orderCode} (Rp {amount})"

### 2. Payment Received ✅
**Trigger:** `backend/service/core_payment_service.go` line 810
```go
NotifyPaymentReceived(order.OrderCode, string(payment.PaymentMethod), order.TotalAmount)
```
**When:** Saat Midtrans webhook confirm payment settlement
**Notification:** "💰 Payment Received - Order #{orderCode} paid via {method} (Rp {amount})"

### 3. Payment Expired ✅
**Trigger:** `backend/service/core_payment_service.go` line 843
```go
NotifyPaymentExpired(order.OrderCode)
```
**When:** Saat payment expired (24 jam tidak dibayar)
**Notification:** "⏰ Payment Expired - Order #{orderCode} payment has expired"

## Events to Implement ⏳

### 4. Order Cancelled
**File:** `backend/service/admin_order_service.go`
**Function:** `CancelOrderAdmin()`
**Notification:** "❌ Order Cancelled - Order #{orderCode} cancelled by admin"

### 5. Order Shipped
**File:** `backend/service/admin_order_service.go`
**Function:** `ShipOrder()`
**Notification:** "📦 Order Shipped - Order #{orderCode} shipped via {courier}"

### 6. Order Delivered
**File:** `backend/service/admin_order_service.go`
**Function:** `DeliverOrder()`
**Notification:** "✅ Order Delivered - Order #{orderCode} delivered successfully"

### 7. Refund Requested
**File:** `backend/service/refund_service.go`
**Function:** `CreateRefund()`
**Notification:** "💸 Refund Request - Refund requested for order #{orderCode} (Rp {amount})"

### 8. Dispute Created
**File:** `backend/service/dispute_service.go`
**Function:** `CreateDispute()`
**Notification:** "⚠️ New Dispute - Dispute #{disputeCode} created for order #{orderCode}"

### 9. Stock Low Alert
**File:** `backend/service/admin_product_service.go`
**Function:** `UpdateStock()` or `CreateProduct()`
**Notification:** "⚠️ Low Stock Alert - {productName} has only {stock} items left"
**Trigger:** When stock < 5

### 10. New User Registration
**File:** `backend/service/auth_service.go`
**Function:** `Register()`
**Notification:** "👤 New User - {name} ({email}) registered"

### 11. Shipment Status Update
**File:** `backend/service/shipment_monitor_service.go`
**Function:** Tracking job updates
**Notification:** "📦 Shipment Update - Order #{orderCode} - {status} ({courier})"

### 12. Payment Recovery
**File:** `backend/service/payment_recovery_service.go`
**Function:** Recovery job finds stuck payment
**Notification:** "🔄 Payment Recovered - Order #{orderCode} payment status updated"

## Testing Checklist

- [x] Order Created - Test by creating new order
- [x] Payment Received - Test by completing payment
- [x] Payment Expired - Test by waiting 24 hours or manual expire
- [ ] Order Cancelled - Test by cancelling order from admin
- [ ] Order Shipped - Test by marking order as shipped
- [ ] Order Delivered - Test by marking order as delivered
- [ ] Refund Requested - Test by creating refund
- [ ] Dispute Created - Test by creating dispute
- [ ] Stock Low - Test by reducing stock below 5
- [ ] New User - Test by registering new user
- [ ] Shipment Update - Test by tracking job
- [ ] Payment Recovery - Test by recovery job

## Frontend Display

Notifications appear in:
1. **Bell Icon Badge** - Shows unread count
2. **Dropdown Panel** - List of all notifications
3. **Browser Notification** - Native OS notification
4. **Sound Alert** - Beep sound
5. **Auto-refresh Dashboard** - Dashboard data refreshes on new order/payment

## Current Status

**Implemented:** 3/12 events (25%)
**Next Priority:**
1. Order Cancelled
2. Order Shipped  
3. Stock Low Alert
4. New User Registration

Mau saya lanjutkan implementasi event-event lainnya?
