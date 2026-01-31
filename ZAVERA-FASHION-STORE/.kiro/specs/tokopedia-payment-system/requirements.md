# Requirements Document

**STATUS: FINAL — READY FOR DESIGN PHASE**

## Introduction

This document specifies the requirements for implementing a Tokopedia-style payment system using Midtrans Core API (NOT Snap UI). The system will feature a dedicated "Pembelian" page with "Menunggu Pembayaran" and "Daftar Transaksi" tabs, immutable payment method selection, and resumable payment flows. The payment method is locked once selected and persists until payment success or expiry.

**Critical Constraints:**
- NO Midtrans SNAP UI - Core API only
- Payment method is LOCKED once selected
- Payment must be resumable exactly like Tokopedia
- One active payment per order at any time

**Scope Boundaries:**
- IN SCOPE: Payment selection, VA generation, payment status tracking, webhook handling, Pembelian page
- OUT OF SCOPE: Post-payment fulfillment lifecycle (PACKING, SHIPPED, DELIVERED states are handled by existing system)

## Glossary

- **Midtrans Core API**: Direct API integration with Midtrans payment gateway without using Snap UI popup
- **Virtual Account (VA)**: Bank-specific virtual account number for payment transfer
- **Order**: A customer purchase record containing items, shipping, and payment information
- **Order Payment**: A separate record tracking payment method, VA details, and payment status for an order
- **Payment Method**: The selected bank VA type (bca_va, bri_va, mandiri_va)
- **Pembelian Page**: Customer purchase management page with payment and transaction history tabs
- **Menunggu Pembayaran**: "Waiting for Payment" tab showing pending orders with active payments
- **Daftar Transaksi**: "Transaction History" tab showing completed/cancelled/expired orders
- **Expiry Time**: The deadline for completing payment (24 hours from VA creation)
- **Idempotency**: Ensuring duplicate webhook calls don't corrupt data
- **Internal Order ID**: Database primary key for orders table (integer)
- **Order Code**: Human-readable order identifier (e.g., ZVR-20260113-XXXXXXXX)
- **Midtrans Order ID**: Unique identifier sent to Midtrans (format: ORDER_CODE-TIMESTAMP)
- **MENUNGGU_PEMBAYARAN**: Order status indicating awaiting payment
- **DIBAYAR**: Order status indicating payment completed
- **DIBATALKAN**: Order status indicating order cancelled
- **KADALUARSA**: Order status indicating payment expired

## State Transition Rules

### Order Status Transitions (Payment Scope Only)

| Current Status | Allowed Next Status | Trigger |
|----------------|---------------------|---------|
| MENUNGGU_PEMBAYARAN | DIBAYAR | Webhook: settlement |
| MENUNGGU_PEMBAYARAN | KADALUARSA | Webhook: expire OR on-access expiry check |
| MENUNGGU_PEMBAYARAN | DIBATALKAN | User cancellation (before payment selection only) |

**Note:** Transitions to PACKING, SHIPPED, DELIVERED, COMPLETED are out of scope for this document.

### Payment Status Transitions

| Current Status | Allowed Next Status | Trigger |
|----------------|---------------------|---------|
| PENDING | PAID | Webhook: settlement |
| PENDING | EXPIRED | Webhook: expire OR on-access expiry check |
| PENDING | CANCELLED | Webhook: cancel |
| PENDING | FAILED | Webhook: deny |

**Final States (No Further Transitions):** PAID, EXPIRED, CANCELLED, FAILED

### Valid State Combinations

| order.status | payment_status | Valid? | Description |
|--------------|----------------|--------|-------------|
| MENUNGGU_PEMBAYARAN | NULL (no record) | ✓ | Order created, payment not yet selected |
| MENUNGGU_PEMBAYARAN | PENDING | ✓ | Payment selected, awaiting transfer |
| DIBAYAR | PAID | ✓ | Payment completed successfully |
| KADALUARSA | EXPIRED | ✓ | Payment expired |
| DIBATALKAN | CANCELLED | ✓ | Order cancelled |
| DIBATALKAN | NULL | ✓ | Order cancelled before payment selection |
| MENUNGGU_PEMBAYARAN | PAID | ✗ | Invalid - must update order status |
| DIBAYAR | PENDING | ✗ | Invalid - inconsistent state |
| KADALUARSA | PENDING | ✗ | Invalid - must update payment status |

### Cancellation Authority Rules

1. **User Cancellation**: Allowed ONLY when order.status = MENUNGGU_PEMBAYARAN AND payment_status IS NULL (no payment record exists)
2. **User Cancellation After Payment Selection**: NOT allowed - user must wait for expiry or complete payment
3. **Webhook Cancellation**: Midtrans webhook with status "cancel" takes precedence over any user action
4. **Conflict Resolution**: If webhook arrives while user action is processing, webhook wins (row lock ensures atomicity)

## Requirements

### Requirement 1: Checkout Flow and Order Creation

**User Story:** As a customer, I want to complete checkout and be directed to payment selection, so that I can choose my preferred payment method.

#### Acceptance Criteria

1. WHEN a user completes checkout with valid cart and shipping THEN the System SHALL create an order with status MENUNGGU_PEMBAYARAN
2. WHEN creating an order THEN the System SHALL NOT create any payment record until user selects payment method
3. WHEN order is created successfully THEN the System SHALL redirect user to Payment Selection Page with order_id parameter
4. WHEN order creation fails THEN the System SHALL return error and NOT redirect to payment selection
5. WHEN creating order THEN the System SHALL reserve stock for all items in the order

### Requirement 2: Payment Method Selection

**User Story:** As a customer, I want to select a payment method from available VA options, so that I can pay using my preferred bank.

#### Acceptance Criteria

1. WHEN user lands on Payment Selection Page THEN the System SHALL display available VA options (BCA VA, BRI VA, Mandiri VA) with bank logos
2. WHEN displaying payment options THEN the System SHALL render radio button selection allowing only ONE option to be selected
3. WHEN user has not selected any payment method THEN the System SHALL disable the "Bayar Sekarang" button
4. WHEN user selects a payment method and clicks "Bayar Sekarang" THEN the System SHALL call backend to create payment
5. WHEN backend receives payment creation request THEN the System SHALL first check if active payment already exists for this order
6. IF active payment exists THEN the System SHALL return existing VA details instead of creating new payment
7. IF no active payment exists THEN the System SHALL call Midtrans Core API to generate VA number
8. WHEN Midtrans Core API returns VA details THEN the System SHALL persist payment_method, bank, va_number, transaction_id, expiry_time, and raw_response to order_payments table
9. WHEN payment record is created THEN the System SHALL redirect user to VA Detail Page

### Requirement 3: Payment Method Immutability

**User Story:** As a customer, I want my payment method to be locked after selection, so that I have a consistent payment experience like Tokopedia.

#### Acceptance Criteria

1. WHEN a payment record exists for an order with status PENDING THEN the System SHALL prevent creation of new payment records for that order
2. WHEN a user attempts to access Payment Selection Page for an order with existing PENDING payment THEN the System SHALL redirect to VA Detail Page with existing payment data
3. WHEN displaying VA Detail Page THEN the System SHALL NOT show any option to change payment method
4. WHEN payment_method field is set THEN the System SHALL treat it as immutable for the lifetime of that payment record
5. IF user wants different payment method THEN the System SHALL require waiting for expiry (changing payment method on existing order is NOT supported)

### Requirement 4: VA Detail Page Display

**User Story:** As a customer, I want to see detailed payment instructions, so that I can complete the VA transfer correctly.

#### Acceptance Criteria

1. WHEN displaying VA Detail Page THEN the System SHALL show bank logo prominently at the top
2. WHEN displaying VA Detail Page THEN the System SHALL show VA number in large, readable font with copy-to-clipboard button
3. WHEN user clicks copy button THEN the System SHALL copy VA number to clipboard and show success feedback
4. WHEN displaying VA Detail Page THEN the System SHALL show total payment amount formatted in Indonesian Rupiah
5. WHEN displaying VA Detail Page THEN the System SHALL show countdown timer displaying hours, minutes, and seconds until expiry
6. WHEN countdown reaches zero THEN the System SHALL update display to show "Waktu Habis" and disable payment actions
7. WHEN displaying VA Detail Page THEN the System SHALL show payment status badge (PENDING, DIBAYAR, KADALUARSA)
8. WHEN displaying VA Detail Page THEN the System SHALL show expandable payment instructions specific to selected bank (ATM, Mobile Banking, Internet Banking steps)
9. WHEN displaying VA Detail Page THEN the System SHALL show "Cek Status Bayar" button to manually refresh payment status

### Requirement 5: Manual Payment Status Check

**User Story:** As a customer, I want to manually check my payment status, so that I can see if my transfer was received.

#### Acceptance Criteria

1. WHEN user clicks "Cek Status Bayar" button THEN the System SHALL query current payment status from database
2. WHEN checking status THEN the System SHALL NOT call Midtrans API directly (rely on webhook for status updates)
3. WHEN status check returns PAID THEN the System SHALL update display to show success state
4. WHEN status check returns EXPIRED THEN the System SHALL update display to show expired state
5. WHEN status check returns PENDING THEN the System SHALL show message "Pembayaran belum diterima" and keep current display
6. WHEN user clicks "Cek Status Bayar" THEN the System SHALL implement rate limiting (max 1 request per 5 seconds)

### Requirement 6: Pembelian Page - Menunggu Pembayaran Tab

**User Story:** As a customer, I want to view my pending payments in the "Menunggu Pembayaran" tab, so that I can complete payments for my orders.

#### Acceptance Criteria

1. WHEN user visits Pembelian page THEN the System SHALL display tabbed interface with "Menunggu Pembayaran" as default active tab
2. WHEN "Menunggu Pembayaran" tab is active THEN the System SHALL query orders where order.status equals MENUNGGU_PEMBAYARAN
3. WHEN displaying pending orders THEN the System SHALL show orders that have payment record with payment_status PENDING
4. WHEN displaying pending orders THEN the System SHALL also show orders that have NO payment record yet (user abandoned before selecting method)
5. WHEN displaying each pending order card THEN the System SHALL show: order_code, order date, item summary, total amount
6. IF order has payment record THEN the System SHALL show: bank logo, VA number (masked: ****1234), expiry countdown
7. IF order has payment record THEN the System SHALL show "Lihat Detail" button leading to VA Detail Page
8. IF order has NO payment record THEN the System SHALL show "Pilih Pembayaran" button leading to Payment Selection Page
9. WHEN user clicks "Bayar Sekarang" or "Lihat Detail" THEN the System SHALL navigate to VA Detail Page with existing payment data
10. WHEN displaying countdown THEN the System SHALL update in real-time without page refresh

### Requirement 7: Pembelian Page - Daftar Transaksi Tab

**User Story:** As a customer, I want to view my transaction history in the "Daftar Transaksi" tab, so that I can see my completed and cancelled orders.

#### Acceptance Criteria

1. WHEN "Daftar Transaksi" tab is clicked THEN the System SHALL display orders where order.status NOT IN (MENUNGGU_PEMBAYARAN)
2. WHEN displaying completed orders THEN the System SHALL show: order_code, order date, item summary, total amount, payment method used, paid_at timestamp
3. WHEN displaying cancelled/expired orders THEN the System SHALL show: order_code, order date, item summary, total amount, status badge
4. WHEN displaying transaction history THEN the System SHALL present orders in read-only format without any payment action buttons
5. WHEN displaying transaction history THEN the System SHALL sort orders by created_at descending (newest first)
6. WHEN displaying transaction history THEN the System SHALL implement pagination with configurable page size

### Requirement 8: Payment Resume Logic

**User Story:** As a customer, I want to resume my pending payment, so that I can complete payment using the same VA number I was given.

#### Acceptance Criteria

1. WHEN user opens a pending order from any entry point THEN the System SHALL check if payment record exists
2. IF payment record exists with payment_status PENDING THEN the System SHALL check if expiry_time is greater than current server time
3. IF payment is valid (PENDING and not expired) THEN the System SHALL display existing VA details without calling Midtrans API
4. IF payment is valid THEN the System SHALL calculate and display remaining time as (expiry_time - current_time)
5. IF payment has expired (expiry_time <= current_time) but payment_status is still PENDING THEN the System SHALL trigger expiry handling
6. WHEN expiry is detected THEN the System SHALL update payment_status to EXPIRED
7. WHEN expiry is detected THEN the System SHALL update order.status to KADALUARSA
8. WHEN expiry is detected THEN the System SHALL restore reserved stock to inventory
9. WHEN displaying expired payment THEN the System SHALL show clear messaging: "Pembayaran telah kadaluarsa"
10. WHEN payment is expired THEN the System SHALL NOT allow any payment actions on that order

### Requirement 9: Midtrans Core API Integration

**User Story:** As a developer, I want to integrate with Midtrans Core API for VA generation, so that customers can pay via bank transfer.

#### Acceptance Criteria

1. WHEN initializing Midtrans client THEN the System SHALL load server_key from MIDTRANS_SERVER_KEY environment variable
2. WHEN initializing Midtrans client THEN the System SHALL determine environment (sandbox/production) from MIDTRANS_ENVIRONMENT variable
3. WHEN calling Midtrans Charge API THEN the System SHALL support payment_type values: bank_transfer with bank: bca, bri, mandiri
4. WHEN constructing Midtrans request THEN the System SHALL use unique order_id format: {ORDER_CODE}-{UNIX_TIMESTAMP}
5. WHEN constructing Midtrans request THEN the System SHALL include transaction_details with order_id and gross_amount
6. WHEN constructing Midtrans request THEN the System SHALL include customer_details with name, email, phone
7. WHEN Midtrans API returns success THEN the System SHALL extract va_number from response based on bank type
8. WHEN Midtrans API returns success THEN the System SHALL extract transaction_id and expiry_time from response
9. WHEN Midtrans API returns error THEN the System SHALL log error details and return user-friendly error message
10. WHEN calling Midtrans API THEN the System SHALL implement timeout of 30 seconds
11. WHEN Midtrans API call times out THEN the System SHALL NOT create payment record and return error to user

### Requirement 10: Webhook Handling

**User Story:** As a system administrator, I want the system to handle Midtrans webhooks correctly, so that payment status is updated reliably.

#### Acceptance Criteria

1. WHEN Midtrans sends webhook notification THEN the System SHALL parse JSON payload containing transaction_status, order_id, signature_key, gross_amount, status_code
2. WHEN processing webhook THEN the System SHALL validate signature using SHA512(order_id + status_code + gross_amount + server_key)
3. IF signature validation fails THEN the System SHALL log the attempt with IP address and reject with HTTP 200 (to prevent retries)
4. WHEN signature is valid THEN the System SHALL extract original order_code from Midtrans order_id (remove timestamp suffix)
5. WHEN processing webhook THEN the System SHALL acquire row lock on order record to prevent race conditions
6. WHEN transaction_status equals "settlement" THEN the System SHALL update payment_status to PAID
7. WHEN transaction_status equals "settlement" THEN the System SHALL update order.status to DIBAYAR
8. WHEN transaction_status equals "settlement" THEN the System SHALL set order.paid_at to current timestamp
9. WHEN transaction_status equals "expire" THEN the System SHALL update payment_status to EXPIRED
10. WHEN transaction_status equals "expire" THEN the System SHALL update order.status to KADALUARSA
11. WHEN transaction_status equals "expire" THEN the System SHALL restore reserved stock
12. WHEN transaction_status equals "cancel" or "deny" THEN the System SHALL update payment_status to CANCELLED/FAILED respectively
13. WHEN processing webhook THEN the System SHALL implement idempotency by checking current status before updating
14. IF payment is already in final status (PAID, EXPIRED, CANCELLED, FAILED) THEN the System SHALL skip processing and return success
15. WHEN webhook processing completes THEN the System SHALL return HTTP 200 with JSON body {"status": "ok"}
16. WHEN webhook processing fails THEN the System SHALL still return HTTP 200 to prevent Midtrans retries on non-recoverable errors

### Requirement 11: Data Persistence Requirements

**User Story:** As a developer, I want the data layer to support immutable payment methods and audit trail, so that payment integrity is maintained.

#### Acceptance Criteria

1. WHEN persisting order_payments THEN the System SHALL store: id, order_id, payment_method, bank, va_number, transaction_id, midtrans_order_id, expiry_time, payment_status, raw_response, created_at, updated_at, paid_at
2. WHEN creating order_payments record THEN the System SHALL enforce one active (PENDING) payment per order
3. WHEN storing payment_method THEN the System SHALL accept values: bca_va, bri_va, mandiri_va
4. WHEN storing payment_status THEN the System SHALL accept values: PENDING, PAID, EXPIRED, CANCELLED, FAILED
5. WHEN storing raw_response THEN the System SHALL persist complete Midtrans API response for audit
6. WHEN updating order_payments THEN the System SHALL track updated_at timestamp
7. WHEN payment succeeds THEN the System SHALL record paid_at timestamp

### Requirement 12: Expiry Handling

**User Story:** As a customer, I want expired payments to be handled gracefully, so that I understand why my payment failed and what options I have.

#### Acceptance Criteria

1. WHEN payment expiry_time is reached THEN the System SHALL mark payment_status as EXPIRED via scheduled job or on-access check
2. WHEN payment expires THEN the System SHALL update order.status to KADALUARSA
3. WHEN payment expires THEN the System SHALL restore all reserved stock for order items
4. WHEN stock is restored THEN the System SHALL create stock_movement record with type RELEASE
5. WHEN displaying expired order THEN the System SHALL show status badge "Kadaluarsa" with red styling
6. WHEN displaying expired order THEN the System SHALL show message: "Pembayaran telah melewati batas waktu"
7. WHEN order is expired THEN the System SHALL NOT allow any payment retry on the same order

### Requirement 13: Concurrency and Data Integrity

**User Story:** As a system administrator, I want the system to handle concurrent requests safely, so that payment data remains consistent.

#### Acceptance Criteria

1. WHEN creating payment record THEN the System SHALL use database transaction with row locking on order
2. WHEN multiple requests attempt to create payment for same order THEN the System SHALL ensure only one payment is created (first wins)
3. WHEN webhook and user action occur simultaneously THEN the System SHALL use row-level locks to ensure atomicity
4. WHEN updating payment status THEN the System SHALL verify current status before applying update (optimistic check)
5. WHEN restoring stock THEN the System SHALL use atomic database operations to prevent overselling

### Requirement 14: Security Requirements

**User Story:** As a security engineer, I want the payment system to be secure, so that customer payment data is protected.

#### Acceptance Criteria

1. WHEN storing Midtrans server_key THEN the System SHALL use environment variables, never hardcode
2. WHEN logging payment data THEN the System SHALL mask VA numbers in logs (show only last 4 digits)
3. WHEN exposing payment APIs THEN the System SHALL require authentication for all endpoints except webhook
4. WHEN receiving webhook THEN the System SHALL validate signature (IP whitelist is optional)
5. WHEN transmitting payment data THEN the System SHALL use HTTPS only

---

## Change Log

| Change | Reason |
|--------|--------|
| Added "Scope Boundaries" section | Explicitly mark post-payment fulfillment as out of scope to prevent scope creep |
| Added "State Transition Rules" section | Provide explicit valid/invalid state combinations to eliminate ambiguity |
| Added "Cancellation Authority Rules" | Clarify who can cancel and when, plus conflict resolution with webhooks |
| Split "Manual Payment Status Check" into separate Requirement 5 | Clarify behavior and limitations of "Cek Status Bayar" action |
| Added rate limiting to status check | Prevent abuse of manual status check feature |
| Clarified status check does NOT call Midtrans API | Prevent unsafe direct API calls, rely on webhook |
| Renamed Requirement 10 to "Data Persistence Requirements" | Avoid design-level language (schema) in requirements |
| Removed "Database Schema Requirements" title | Requirements should describe what, not how |
| Updated Requirement 3.5 | Removed "not implemented in this scope" - clarified user must wait for expiry |
| Updated Requirement 7.1 | Changed to NOT IN (MENUNGGU_PEMBAYARAN) for cleaner logic |
| Removed "cancellation/expiry reason" from Requirement 7.3 | Simplified - reason tracking is implementation detail |
| Removed IP whitelist as mandatory | Made optional since signature validation is sufficient |
| Removed rate limiting on copy action | Over-engineering for low-risk action |
