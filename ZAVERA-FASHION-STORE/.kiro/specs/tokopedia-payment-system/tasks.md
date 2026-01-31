# Implementation Plan

## Phase 1: Backend Foundation

- [x] 1. Database Migration and Models


  - [x] 1.1 Create database migration for order_payments table


    - Add columns: id, order_id, payment_method, bank, va_number, transaction_id, midtrans_order_id, expiry_time, payment_status, raw_response (JSONB), created_at, updated_at, paid_at
    - Add unique constraint on (order_id) where payment_status = 'PENDING'
    - Add foreign key to orders table
    - Add indexes on order_id, payment_status, expiry_time
    - _Requirements: 11.1, 11.2_


  - [x] 1.2 Create OrderPayment model in Go

    - Define OrderPayment struct with all fields
    - Define PaymentMethod enum (bca_va, bri_va, mandiri_va)
    - Define PaymentStatus enum (PENDING, PAID, EXPIRED, CANCELLED, FAILED)
    - Add JSON and DB tags
    - _Requirements: 11.3, 11.4_

  - [ ]* 1.3 Write property test for OrderPayment data round-trip
    - **Property 18: Payment Data Round-Trip**
    - **Validates: Requirements 11.1-11.7**

- [x] 2. Order Payment Repository



  - [x] 2.1 Create OrderPaymentRepository interface

    - Define Create, FindByOrderID, FindPendingByOrderID, UpdateStatus, UpdateToPaid, UpdateToExpired methods
    - _Requirements: 2.8, 3.1_


  - [x] 2.2 Implement OrderPaymentRepository


    - Implement Create with row locking on order
    - Implement FindByOrderID with expiry check
    - Implement FindPendingByOrderID
    - Implement UpdateStatus with optimistic locking
    - Implement UpdateToPaid with paid_at timestamp
    - Implement UpdateToExpired with stock restoration trigger
    - _Requirements: 13.1, 13.2, 13.4_

  - [ ]* 2.3 Write property test for payment creation idempotency
    - **Property 3: Payment Creation Idempotency**
    - **Validates: Requirements 2.5, 2.6, 3.1**

  - [ ]* 2.4 Write property test for payment method immutability
    - **Property 5: Payment Method Immutability**
    - **Validates: Requirements 3.4**

- [x] 3. Checkpoint - Ensure all tests pass


  - Ensure all tests pass, ask the user if questions arise.

## Phase 2: Midtrans Core API Integration

- [x] 4. Midtrans Core API Client



  - [x] 4.1 Create MidtransCoreClient interface and implementation

    - Load server_key from MIDTRANS_SERVER_KEY environment variable
    - Load environment from MIDTRANS_ENVIRONMENT variable
    - Implement ChargeVA method for bank_transfer payment type
    - Support BCA, BRI, Mandiri VA generation
    - Implement 30-second timeout
    - _Requirements: 9.1, 9.2, 9.3, 9.10_


  - [x] 4.2 Implement VA response parsing


    - Extract va_number based on bank type (va_numbers array for BCA/BRI, permata_va_number for Mandiri)
    - Extract transaction_id and expiry_time
    - Store complete raw_response for audit
    - _Requirements: 9.7, 9.8, 11.5_


  - [x] 4.3 Implement error handling for Midtrans API


    - Handle network errors with user-friendly messages
    - Handle timeout errors
    - Log error details for debugging
    - _Requirements: 9.9, 9.11_

- [x] 5. Core Payment Service



  - [x] 5.1 Create CorePaymentService interface

    - Define CreateVAPayment, GetPaymentByOrderID, CheckPaymentStatus, ProcessCoreWebhook methods
    - _Requirements: 2.4, 5.1, 8.1_


  - [x] 5.2 Implement CreateVAPayment method


    - Check if order exists and status is MENUNGGU_PEMBAYARAN
    - Check if active PENDING payment already exists (return existing if found)
    - Generate unique midtrans_order_id: {ORDER_CODE}-{UNIX_TIMESTAMP}
    - Call Midtrans Core API to generate VA
    - Persist payment record with all fields
    - Return payment details
    - _Requirements: 2.5, 2.6, 2.7, 2.8, 9.4, 9.5, 9.6_

  - [ ]* 5.3 Write property test for payment data persistence completeness
    - **Property 4: Payment Data Persistence Completeness**
    - **Validates: Requirements 2.8**


  - [x] 5.4 Implement GetPaymentByOrderID method


    - Find payment by order ID
    - Check expiry on access (if PENDING and expiry_time <= now, trigger expiry handling)
    - Calculate remaining_seconds
    - Return payment details with bank instructions
    - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5_

  - [ ]* 5.5 Write property test for remaining time calculation
    - **Property 9: Remaining Time Calculation**
    - **Validates: Requirements 8.4**

  - [ ]* 5.6 Write property test for on-access expiry detection
    - **Property 10: On-Access Expiry Detection**

    - **Validates: Requirements 8.5, 8.6, 8.7**

  - [x] 5.7 Implement CheckPaymentStatus method


    - Query current payment status from database (NO Midtrans API call)
    - Return status with appropriate message
    - _Requirements: 5.1, 5.2_

  - [ ]* 5.8 Write property test for status check returns current state
    - **Property 6: Status Check Returns Current State**
    - **Validates: Requirements 5.1**

- [x] 6. Checkpoint - Ensure all tests pass

  - Ensure all tests pass, ask the user if questions arise.

## Phase 3: Expiry and Stock Handling

- [x] 7. Expiry Handling Service



  - [x] 7.1 Implement expiry handling logic


    - Update payment_status to EXPIRED
    - Update order.status to KADALUARSA
    - Trigger stock restoration
    - Create stock_movement record with type RELEASE
    - _Requirements: 8.6, 8.7, 8.8, 12.1, 12.2, 12.3, 12.4_

  - [ ]* 7.2 Write property test for stock restoration on expiry
    - **Property 11: Stock Restoration on Expiry**
    - **Validates: Requirements 8.8, 12.3**

  - [ ]* 7.3 Write property test for expired order action blocking
    - **Property 12: Expired Order Action Blocking**
    - **Validates: Requirements 8.10**

- [x] 8. Checkpoint - Ensure all tests pass

  - Ensure all tests pass, ask the user if questions arise.

## Phase 4: Webhook Processing

- [x] 9. Webhook Handler and Processing

  - [x] 9.1 Create webhook signature validation


    - Implement SHA512(order_id + status_code + gross_amount + server_key)
    - Log invalid signature attempts with IP address
    - Return HTTP 200 even on invalid signature (prevent retries)
    - _Requirements: 10.2, 10.3_

  - [ ]* 9.2 Write property test for webhook signature validation
    - **Property 13: Webhook Signature Validation**
    - **Validates: Requirements 10.2, 10.3**


  - [x] 9.3 Implement Midtrans order ID parsing


    - Extract original order_code from {ORDER_CODE}-{TIMESTAMP} format
    - Handle edge cases (malformed IDs)
    - _Requirements: 10.4_

  - [ ]* 9.4 Write property test for Midtrans order ID parsing
    - **Property 14: Midtrans Order ID Parsing**
    - **Validates: Requirements 10.4**

  - [x] 9.5 Implement ProcessCoreWebhook method

    - Parse webhook JSON payload
    - Validate signature
    - Extract order_code from midtrans_order_id
    - Acquire row lock on order record
    - Check current payment status (idempotency)
    - Process based on transaction_status:
      - settlement → PAID, order DIBAYAR, set paid_at
      - expire → EXPIRED, order KADALUARSA, restore stock
      - cancel → CANCELLED
      - deny → FAILED
    - Return HTTP 200 always
    - _Requirements: 10.1, 10.5, 10.6-10.16_

  - [ ]* 9.6 Write property test for webhook settlement transition
    - **Property 15: Webhook Settlement Transition**
    - **Validates: Requirements 10.6, 10.7, 10.8**

  - [ ]* 9.7 Write property test for webhook expire transition
    - **Property 16: Webhook Expire Transition**
    - **Validates: Requirements 10.9, 10.10, 10.11**

  - [ ]* 9.8 Write property test for webhook idempotency
    - **Property 17: Webhook Idempotency**
    - **Validates: Requirements 10.13, 10.14**

- [x] 10. Checkpoint - Ensure all tests pass

  - Ensure all tests pass, ask the user if questions arise.

## Phase 5: API Handlers

- [x] 11. Core Payment Handler


  - [x] 11.1 Create CorePaymentHandler struct

    - Inject CorePaymentService dependency
    - _Requirements: 2.4_


  - [x] 11.2 Implement POST /api/payments/core/create endpoint


    - Parse CreateVAPaymentRequest (order_id, payment_method)
    - Validate payment_method is one of: bca_va, bri_va, mandiri_va
    - Require authentication
    - Call CorePaymentService.CreateVAPayment
    - Return CreateVAPaymentResponse with VA details and instructions
    - Handle errors with appropriate HTTP status codes
    - _Requirements: 2.4, 2.9, 14.3_


  - [x] 11.3 Implement GET /api/payments/core/:order_id endpoint


    - Require authentication
    - Verify user owns the order
    - Call CorePaymentService.GetPaymentByOrderID
    - Return PaymentDetailsResponse
    - Handle expired payment (410 Gone)

    - _Requirements: 8.1, 8.3, 14.3_

  - [x] 11.4 Implement POST /api/payments/core/check endpoint


    - Parse CheckPaymentStatusRequest (payment_id)
    - Require authentication
    - Implement rate limiting (max 1 request per 5 seconds per user)
    - Call CorePaymentService.CheckPaymentStatus

    - Return PaymentStatusResponse
    - _Requirements: 5.1, 5.6_

  - [x] 11.5 Implement POST /api/webhook/midtrans/core endpoint


    - Parse CoreWebhookNotification
    - Call CorePaymentService.ProcessCoreWebhook
    - Always return HTTP 200
    - _Requirements: 10.15, 10.16_

- [x] 12. Pembelian Handler





  - [x] 12.1 Implement GET /api/pembelian/pending endpoint


    - Require authentication
    - Query orders where status = MENUNGGU_PEMBAYARAN for user
    - Include orders with and without payment records
    - Calculate remaining_seconds for each payment
    - Mask VA numbers (show only last 4 digits)
    - Return PendingOrdersResponse with pagination
    - _Requirements: 6.2, 6.3, 6.4, 6.5, 6.6, 14.2_

  - [ ]* 12.2 Write property test for pending orders query correctness
    - **Property 7: Pending Orders Query Correctness**
    - **Validates: Requirements 6.2, 6.3, 6.4**


  - [x] 12.3 Implement GET /api/pembelian/history endpoint


    - Require authentication
    - Query orders where status NOT IN (MENUNGGU_PEMBAYARAN) for user
    - Sort by created_at descending
    - Return TransactionHistoryResponse with pagination
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.6_

  - [ ]* 12.4 Write property test for transaction history query correctness
    - **Property 8: Transaction History Query Correctness**
    - **Validates: Requirements 7.1, 7.5**

- [x] 13. Register Routes

  - [x] 13.1 Add Core Payment routes to router


    - POST /api/payments/core/create (authenticated)
    - GET /api/payments/core/:order_id (authenticated)
    - POST /api/payments/core/check (authenticated)
    - POST /api/webhook/midtrans/core (public)
    - GET /api/pembelian/pending (authenticated)
    - GET /api/pembelian/history (authenticated)
    - _Requirements: 14.3_

- [x] 14. Checkpoint - Ensure all tests pass

  - Ensure all tests pass, ask the user if questions arise.

## Phase 6: Frontend - Payment Selection Page

- [x] 15. Payment Selection Page

  - [x] 15.1 Create PaymentSelectionPage component


    - Route: /checkout/payment?order_id={id}
    - Fetch order details on mount
    - Check if payment already exists (redirect to VA detail if so)
    - Display order summary (items, total)
    - _Requirements: 2.1, 3.2_

  - [x] 15.2 Create BankSelector component


    - Display BCA, BRI, Mandiri options with bank logos
    - Radio button selection (only one selectable)
    - Highlight selected option
    - Tokopedia-style card design
    - _Requirements: 2.1, 2.2_

  - [x] 15.3 Implement payment method selection state


    - Track selected payment method
    - Disable "Bayar Sekarang" button when no selection
    - Enable button when method selected
    - _Requirements: 2.3_

  - [x] 15.4 Implement payment creation flow


    - On "Bayar Sekarang" click, call POST /api/payments/core/create
    - Show loading state during API call
    - On success, redirect to VA Detail page
    - On error, show error toast
    - _Requirements: 2.4, 2.9_

- [x] 16. Checkpoint - Ensure all tests pass

  - Ensure all tests pass, ask the user if questions arise.

## Phase 7: Frontend - VA Detail Page

- [x] 17. VA Detail Page

  - [x] 17.1 Create VADetailPage component


    - Route: /checkout/payment/detail?order_id={id}
    - Fetch payment details on mount
    - Handle expired payment state
    - _Requirements: 4.1, 8.1_

  - [x] 17.2 Create VACard component


    - Display bank logo prominently
    - Display VA number in large font
    - Copy-to-clipboard button with success feedback
    - Display total amount in IDR format
    - _Requirements: 4.1, 4.2, 4.3, 4.4_

  - [x] 17.3 Create CountdownTimer component


    - Display hours:minutes:seconds format
    - Update every second
    - Show "Waktu Habis" when countdown reaches zero
    - Trigger status check when expired
    - _Requirements: 4.5, 4.6_

  - [x] 17.4 Create StatusBadge component


    - Display PENDING (yellow), DIBAYAR (green), KADALUARSA (red)
    - Tokopedia-style badge design
    - _Requirements: 4.7_

  - [x] 17.5 Create PaymentInstructions component


    - Expandable accordion for each channel (ATM, Mobile Banking, Internet Banking)
    - Bank-specific instructions
    - _Requirements: 4.8_

  - [x] 17.6 Implement "Cek Status Bayar" button


    - Call POST /api/payments/core/check
    - Rate limit: disable button for 5 seconds after click
    - Update status badge on response
    - Show appropriate message
    - _Requirements: 4.9, 5.3, 5.4, 5.5, 5.6_

- [x] 18. Checkpoint - Ensure all tests pass

  - Ensure all tests pass, ask the user if questions arise.

## Phase 8: Frontend - Pembelian Page

- [x] 19. Pembelian Page with Tabs

  - [x] 19.1 Create PembelianPage component


    - Route: /account/pembelian
    - Tabbed interface with "Menunggu Pembayaran" and "Daftar Transaksi"
    - Default to "Menunggu Pembayaran" tab
    - _Requirements: 6.1_

  - [x] 19.2 Create PendingOrdersList component


    - Fetch from GET /api/pembelian/pending
    - Display order cards with:
      - Order code, date, item summary, total
      - Bank logo and masked VA number (if payment exists)
      - Countdown timer (if payment exists)
      - "Lihat Detail" button (if payment exists)
      - "Pilih Pembayaran" button (if no payment)
    - Real-time countdown updates
    - _Requirements: 6.2, 6.3, 6.4, 6.5, 6.6, 6.7, 6.8, 6.9, 6.10_

  - [x] 19.3 Create TransactionHistoryList component


    - Fetch from GET /api/pembelian/history
    - Display order cards with:
      - Order code, date, item summary, total
      - Status badge
      - Payment method and paid_at (if paid)
    - Read-only (no action buttons)
    - Pagination controls
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.6_

  - [x] 19.4 Create PendingOrderCard component


    - Tokopedia-style card design
    - Conditional rendering based on has_payment
    - Click handlers for navigation
    - _Requirements: 6.5, 6.6, 6.7, 6.8_

  - [x] 19.5 Create TransactionHistoryCard component


    - Tokopedia-style card design
    - Status-based styling
    - Read-only display
    - _Requirements: 7.2, 7.3, 7.4_

- [x] 20. Checkpoint - Ensure all tests pass

  - Ensure all tests pass, ask the user if questions arise.

## Phase 9: Frontend - Integration and Polish

- [x] 21. API Integration

  - [x] 21.1 Create corePaymentApi module


    - createVAPayment(orderId, paymentMethod)
    - getPaymentDetails(orderId)
    - checkPaymentStatus(paymentId)
    - getPendingOrders(page, pageSize)
    - getTransactionHistory(page, pageSize)
    - _Requirements: 2.4, 5.1, 6.2, 7.1_

  - [x] 21.2 Add error handling and loading states


    - Loading spinners during API calls
    - Error toasts for failures
    - Retry logic for transient errors
    - _Requirements: 9.9_

- [x] 22. Navigation and Routing

  - [x] 22.1 Update checkout flow to redirect to Payment Selection


    - After successful checkout, redirect to /checkout/payment?order_id={id}
    - _Requirements: 1.3_

  - [x] 22.2 Add Pembelian link to user menu/header


    - Link to /account/pembelian
    - Show pending count badge (optional)
    - _Requirements: 6.1_

  - [x] 22.3 Implement redirect logic for existing payments


    - If user visits Payment Selection with existing payment, redirect to VA Detail
    - _Requirements: 3.2_

- [x] 23. UI Polish

  - [x] 23.1 Add bank logo assets


    - /public/images/banks/bca.svg
    - /public/images/banks/bri.svg
    - /public/images/banks/mandiri.svg
    - _Requirements: 4.1_

  - [x] 23.2 Implement Tokopedia-style design system


    - Clean typography (Inter/system fonts)
    - Consistent spacing and padding
    - Card shadows and borders
    - Color palette (green for success, yellow for pending, red for expired)
    - Mobile-responsive layouts
    - _Requirements: 4.1-4.9_

  - [x] 23.3 Add copy-to-clipboard feedback


    - Toast notification on successful copy
    - Visual feedback on button
    - _Requirements: 4.3_

- [x] 24. Final Checkpoint - Ensure all tests pass

  - Ensure all tests pass, ask the user if questions arise.

## Phase 10: Property-Based Tests (Backend)

- [ ]* 25. Backend Property Tests Setup
  - [ ]* 25.1 Set up gopter for Go property-based testing
    - Add gopter dependency
    - Create test generators for OrderPayment, Order, PaymentMethod
    - Configure 100 iterations per property
    - _Requirements: Testing Strategy_

  - [ ]* 25.2 Write property test for order creation decouples from payment
    - **Property 1: Order Creation Decouples from Payment**
    - **Validates: Requirements 1.1, 1.2**

  - [ ]* 25.3 Write property test for stock reservation on order creation
    - **Property 2: Stock Reservation on Order Creation**
    - **Validates: Requirements 1.5**

- [x] 26. Final Checkpoint - Ensure all tests pass

  - ✅ Backend: `go build` successful, `go test ./...` all tests pass
  - ✅ Frontend: `npm run build` successful
  - All implementation tasks completed. Optional property-based tests (marked with `*`) are not implemented as they are optional.
