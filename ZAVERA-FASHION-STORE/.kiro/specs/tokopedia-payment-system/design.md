# Design Document: Tokopedia-Style Payment System

## Overview

This design document describes the architecture and implementation approach for a Tokopedia-style payment system using Midtrans Core API. The system enables customers to select VA payment methods, receive immutable payment details, and resume pending payments through a dedicated "Pembelian" page.

**Design Principles:**
- Core API only (NO Midtrans Snap UI)
- Payment method immutability after selection
- Resumable payments with same VA number
- High concurrency support with row-level locking
- Idempotent webhook processing

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              FRONTEND (Next.js)                              │
├─────────────────────────────────────────────────────────────────────────────┤
│  Payment Selection Page  │  VA Detail Page  │  Pembelian Page (Tabs)        │
│  - Bank radio buttons    │  - VA display    │  - Menunggu Pembayaran        │
│  - Bayar Sekarang btn    │  - Countdown     │  - Daftar Transaksi           │
│                          │  - Copy VA       │                               │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              BACKEND (Go/Gin)                                │
├─────────────────────────────────────────────────────────────────────────────┤
│  CorePaymentHandler      │  CorePaymentService    │  OrderPaymentRepository  │
│  - POST /payments/create │  - CreateVAPayment()   │  - Create()              │
│  - GET /payments/:id     │  - GetPaymentDetails() │  - FindByOrderID()       │
│  - POST /payments/check  │  - CheckExpiry()       │  - UpdateStatus()        │
│  - POST /webhook/core    │  - ProcessWebhook()    │  - FindPendingByOrder()  │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           EXTERNAL SERVICES                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│  Midtrans Core API                    │  PostgreSQL Database                 │
│  - POST /v2/charge (VA generation)    │  - orders table                      │
│  - Webhook notifications              │  - order_payments table              │
│                                       │  - stock_movements table             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Components and Interfaces

### Domain Entities

#### Order Entity
Represents a customer purchase. Extended with payment-specific status values.

**Attributes:**
- id: int (PK)
- order_code: string (unique, human-readable)
- user_id: int (FK, nullable for guest)
- customer_name, customer_email, customer_phone: string
- subtotal, shipping_cost, tax, discount, total_amount: decimal
- status: OrderStatus enum
- stock_reserved: boolean
- created_at, updated_at, paid_at, cancelled_at: timestamp

**Order Status Values (Payment Scope):**
- MENUNGGU_PEMBAYARAN: Awaiting payment selection or transfer
- DIBAYAR: Payment completed successfully
- DIBATALKAN: Order cancelled
- KADALUARSA: Payment expired

#### OrderPayment Entity
Tracks payment method selection and VA details. Immutable after creation.

**Attributes:**
- id: int (PK)
- order_id: int (FK to orders)
- payment_method: PaymentMethod enum (bca_va, bri_va, mandiri_va)
- bank: string (bca, bri, mandiri)
- va_number: string
- transaction_id: string (Midtrans transaction ID)
- midtrans_order_id: string (ORDER_CODE-TIMESTAMP format)
- expiry_time: timestamp
- payment_status: PaymentStatus enum
- raw_response: JSONB (complete Midtrans response)
- created_at, updated_at, paid_at: timestamp

**Payment Status Values:**
- PENDING: VA generated, awaiting transfer
- PAID: Payment received (settlement)
- EXPIRED: Payment window expired
- CANCELLED: Payment cancelled by Midtrans
- FAILED: Payment denied by Midtrans

### State Machines

#### Order State Machine (Payment Scope)

```
                    ┌─────────────────────────┐
                    │  MENUNGGU_PEMBAYARAN    │
                    │  (Initial State)        │
                    └───────────┬─────────────┘
                                │
            ┌───────────────────┼───────────────────┐
            │                   │                   │
            ▼                   ▼                   ▼
    ┌───────────────┐   ┌───────────────┐   ┌───────────────┐
    │   DIBAYAR     │   │  KADALUARSA   │   │  DIBATALKAN   │
    │   (Final)     │   │   (Final)     │   │   (Final)     │
    └───────────────┘   └───────────────┘   └───────────────┘
    
Triggers:
- MENUNGGU_PEMBAYARAN → DIBAYAR: Webhook settlement
- MENUNGGU_PEMBAYARAN → KADALUARSA: Webhook expire OR on-access expiry
- MENUNGGU_PEMBAYARAN → DIBATALKAN: User cancel (before payment selection only)
```

#### Payment State Machine

```
                    ┌─────────────────────────┐
                    │       PENDING           │
                    │   (Initial State)       │
                    └───────────┬─────────────┘
                                │
        ┌───────────────────────┼───────────────────────┐
        │                       │                       │
        ▼                       ▼                       ▼
┌───────────────┐       ┌───────────────┐       ┌───────────────┐
│     PAID      │       │   EXPIRED     │       │  CANCELLED    │
│   (Final)     │       │   (Final)     │       │   (Final)     │
└───────────────┘       └───────────────┘       └───────────────┘
                                                        │
                                                        ▼
                                                ┌───────────────┐
                                                │    FAILED     │
                                                │   (Final)     │
                                                └───────────────┘

Triggers:
- PENDING → PAID: Webhook transaction_status = "settlement"
- PENDING → EXPIRED: Webhook transaction_status = "expire" OR on-access check
- PENDING → CANCELLED: Webhook transaction_status = "cancel"
- PENDING → FAILED: Webhook transaction_status = "deny"
```

### Service Interfaces

#### CorePaymentService Interface

```go
type CorePaymentService interface {
    // Create VA payment via Midtrans Core API
    // Returns existing payment if one already exists (idempotent)
    CreateVAPayment(orderID int, paymentMethod string) (*OrderPayment, error)
    
    // Get payment details for an order
    // Triggers expiry check if payment is PENDING and expired
    GetPaymentByOrderID(orderID int) (*OrderPayment, error)
    
    // Check payment status (database only, no Midtrans API call)
    CheckPaymentStatus(paymentID int) (*PaymentStatusResponse, error)
    
    // Process Midtrans webhook notification
    ProcessCoreWebhook(notification CoreWebhookNotification) error
    
    // Get pending orders for user (Menunggu Pembayaran tab)
    GetPendingOrders(userID int, page, pageSize int) (*PendingOrdersResponse, error)
    
    // Get transaction history for user (Daftar Transaksi tab)
    GetTransactionHistory(userID int, page, pageSize int) (*TransactionHistoryResponse, error)
}
```

#### MidtransCoreClient Interface

```go
type MidtransCoreClient interface {
    // Generate VA via Midtrans Core API /v2/charge
    ChargeVA(request ChargeVARequest) (*ChargeVAResponse, error)
}
```

## Data Models

### OrderPayment Model

```go
type OrderPayment struct {
    ID              int            `json:"id" db:"id"`
    OrderID         int            `json:"order_id" db:"order_id"`
    PaymentMethod   string         `json:"payment_method" db:"payment_method"`
    Bank            string         `json:"bank" db:"bank"`
    VANumber        string         `json:"va_number" db:"va_number"`
    TransactionID   string         `json:"transaction_id" db:"transaction_id"`
    MidtransOrderID string         `json:"midtrans_order_id" db:"midtrans_order_id"`
    ExpiryTime      time.Time      `json:"expiry_time" db:"expiry_time"`
    PaymentStatus   PaymentStatus  `json:"payment_status" db:"payment_status"`
    RawResponse     map[string]any `json:"raw_response" db:"raw_response"`
    CreatedAt       time.Time      `json:"created_at" db:"created_at"`
    UpdatedAt       time.Time      `json:"updated_at" db:"updated_at"`
    PaidAt          *time.Time     `json:"paid_at,omitempty" db:"paid_at"`
}

type PaymentMethod string
const (
    PaymentMethodBCAVA     PaymentMethod = "bca_va"
    PaymentMethodBRIVA     PaymentMethod = "bri_va"
    PaymentMethodMandiriVA PaymentMethod = "mandiri_va"
)

type PaymentStatus string
const (
    PaymentStatusPending   PaymentStatus = "PENDING"
    PaymentStatusPaid      PaymentStatus = "PAID"
    PaymentStatusExpired   PaymentStatus = "EXPIRED"
    PaymentStatusCancelled PaymentStatus = "CANCELLED"
    PaymentStatusFailed    PaymentStatus = "FAILED"
)
```

### API Request/Response Models

```go
// Create VA Payment
type CreateVAPaymentRequest struct {
    OrderID       int    `json:"order_id" binding:"required"`
    PaymentMethod string `json:"payment_method" binding:"required,oneof=bca_va bri_va mandiri_va"`
}

type CreateVAPaymentResponse struct {
    PaymentID     int       `json:"payment_id"`
    OrderID       int       `json:"order_id"`
    OrderCode     string    `json:"order_code"`
    PaymentMethod string    `json:"payment_method"`
    Bank          string    `json:"bank"`
    VANumber      string    `json:"va_number"`
    Amount        float64   `json:"amount"`
    ExpiryTime    time.Time `json:"expiry_time"`
    Status        string    `json:"status"`
}

// Payment Details
type PaymentDetailsResponse struct {
    PaymentID       int       `json:"payment_id"`
    OrderID         int       `json:"order_id"`
    OrderCode       string    `json:"order_code"`
    PaymentMethod   string    `json:"payment_method"`
    Bank            string    `json:"bank"`
    BankLogo        string    `json:"bank_logo"`
    VANumber        string    `json:"va_number"`
    Amount          float64   `json:"amount"`
    ExpiryTime      time.Time `json:"expiry_time"`
    RemainingSeconds int      `json:"remaining_seconds"`
    Status          string    `json:"status"`
    Instructions    []PaymentInstruction `json:"instructions"`
}

type PaymentInstruction struct {
    Channel string   `json:"channel"` // ATM, Mobile Banking, Internet Banking
    Steps   []string `json:"steps"`
}

// Pending Orders (Menunggu Pembayaran)
type PendingOrdersResponse struct {
    Orders     []PendingOrderItem `json:"orders"`
    TotalCount int                `json:"total_count"`
    Page       int                `json:"page"`
    PageSize   int                `json:"page_size"`
}

type PendingOrderItem struct {
    OrderID          int        `json:"order_id"`
    OrderCode        string     `json:"order_code"`
    TotalAmount      float64    `json:"total_amount"`
    ItemCount        int        `json:"item_count"`
    ItemSummary      string     `json:"item_summary"`
    CreatedAt        time.Time  `json:"created_at"`
    HasPayment       bool       `json:"has_payment"`
    PaymentMethod    *string    `json:"payment_method,omitempty"`
    Bank             *string    `json:"bank,omitempty"`
    BankLogo         *string    `json:"bank_logo,omitempty"`
    VANumberMasked   *string    `json:"va_number_masked,omitempty"` // ****1234
    ExpiryTime       *time.Time `json:"expiry_time,omitempty"`
    RemainingSeconds *int       `json:"remaining_seconds,omitempty"`
}

// Transaction History (Daftar Transaksi)
type TransactionHistoryResponse struct {
    Orders     []TransactionHistoryItem `json:"orders"`
    TotalCount int                      `json:"total_count"`
    Page       int                      `json:"page"`
    PageSize   int                      `json:"page_size"`
}

type TransactionHistoryItem struct {
    OrderID       int       `json:"order_id"`
    OrderCode     string    `json:"order_code"`
    TotalAmount   float64   `json:"total_amount"`
    ItemCount     int       `json:"item_count"`
    ItemSummary   string    `json:"item_summary"`
    Status        string    `json:"status"`
    PaymentMethod *string   `json:"payment_method,omitempty"`
    PaidAt        *time.Time `json:"paid_at,omitempty"`
    CreatedAt     time.Time `json:"created_at"`
}

// Midtrans Core Webhook
type CoreWebhookNotification struct {
    TransactionTime   string `json:"transaction_time"`
    TransactionStatus string `json:"transaction_status"`
    TransactionID     string `json:"transaction_id"`
    StatusMessage     string `json:"status_message"`
    StatusCode        string `json:"status_code"`
    SignatureKey      string `json:"signature_key"`
    PaymentType       string `json:"payment_type"`
    OrderID           string `json:"order_id"` // Midtrans order ID (ORDER_CODE-TIMESTAMP)
    MerchantID        string `json:"merchant_id"`
    GrossAmount       string `json:"gross_amount"`
    FraudStatus       string `json:"fraud_status"`
    Currency          string `json:"currency"`
    // VA-specific fields
    VANumbers         []VANumber `json:"va_numbers,omitempty"`
    PermataVANumber   string     `json:"permata_va_number,omitempty"`
}

type VANumber struct {
    Bank     string `json:"bank"`
    VANumber string `json:"va_number"`
}
```



## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

Based on the acceptance criteria analysis, the following correctness properties must be validated through property-based testing:

### Property 1: Order Creation Decouples from Payment
*For any* valid checkout request, creating an order SHALL result in order.status = MENUNGGU_PEMBAYARAN AND no payment record exists for that order.
**Validates: Requirements 1.1, 1.2**

### Property 2: Stock Reservation on Order Creation
*For any* order with N items of quantity Q each, after order creation, the stock for each product SHALL be decremented by exactly Q.
**Validates: Requirements 1.5**

### Property 3: Payment Creation Idempotency
*For any* order with an existing PENDING payment, calling CreateVAPayment SHALL return the existing payment details without creating a new payment record.
**Validates: Requirements 2.5, 2.6, 3.1**

### Property 4: Payment Data Persistence Completeness
*For any* successful payment creation, the order_payments record SHALL contain non-null values for: payment_method, bank, va_number, transaction_id, midtrans_order_id, expiry_time, payment_status, raw_response.
**Validates: Requirements 2.8**

### Property 5: Payment Method Immutability
*For any* order_payment record, the payment_method field SHALL never change after initial creation, regardless of subsequent operations.
**Validates: Requirements 3.4**

### Property 6: Status Check Returns Current State
*For any* payment with status S in the database, calling CheckPaymentStatus SHALL return status S.
**Validates: Requirements 5.1**

### Property 7: Pending Orders Query Correctness
*For any* user, GetPendingOrders SHALL return exactly the orders where order.status = MENUNGGU_PEMBAYARAN, including both orders with and without payment records.
**Validates: Requirements 6.2, 6.3, 6.4**

### Property 8: Transaction History Query Correctness
*For any* user, GetTransactionHistory SHALL return exactly the orders where order.status NOT IN (MENUNGGU_PEMBAYARAN), sorted by created_at descending.
**Validates: Requirements 7.1, 7.5**

### Property 9: Remaining Time Calculation
*For any* valid payment with expiry_time E and current_time T where E > T, the remaining_seconds SHALL equal (E - T) in seconds.
**Validates: Requirements 8.4**

### Property 10: On-Access Expiry Detection
*For any* payment where payment_status = PENDING AND expiry_time <= current_time, accessing the payment SHALL trigger expiry handling resulting in payment_status = EXPIRED AND order.status = KADALUARSA.
**Validates: Requirements 8.5, 8.6, 8.7**

### Property 11: Stock Restoration on Expiry
*For any* order that transitions to KADALUARSA, the stock for each order item SHALL be restored (incremented) by the item quantity.
**Validates: Requirements 8.8, 12.3**

### Property 12: Expired Order Action Blocking
*For any* order with status KADALUARSA, attempting to create a payment SHALL fail with an appropriate error.
**Validates: Requirements 8.10**

### Property 13: Webhook Signature Validation
*For any* webhook notification, the signature SHALL be validated as SHA512(order_id + status_code + gross_amount + server_key). Invalid signatures SHALL be rejected.
**Validates: Requirements 10.2, 10.3**

### Property 14: Midtrans Order ID Parsing
*For any* Midtrans order_id in format "{ORDER_CODE}-{TIMESTAMP}", extracting the order_code SHALL return the original ORDER_CODE.
**Validates: Requirements 10.4**

### Property 15: Webhook Settlement Transition
*For any* webhook with transaction_status = "settlement" for a PENDING payment, processing SHALL result in payment_status = PAID AND order.status = DIBAYAR AND paid_at IS NOT NULL.
**Validates: Requirements 10.6, 10.7, 10.8**

### Property 16: Webhook Expire Transition
*For any* webhook with transaction_status = "expire" for a PENDING payment, processing SHALL result in payment_status = EXPIRED AND order.status = KADALUARSA AND stock restored.
**Validates: Requirements 10.9, 10.10, 10.11**

### Property 17: Webhook Idempotency
*For any* payment in a final status (PAID, EXPIRED, CANCELLED, FAILED), processing a duplicate webhook SHALL NOT change the payment or order state.
**Validates: Requirements 10.13, 10.14**

### Property 18: Payment Data Round-Trip
*For any* OrderPayment object, serializing to database and deserializing back SHALL produce an equivalent object.
**Validates: Requirements 11.1-11.7**

## Error Handling

### Payment Creation Errors

| Error Code | Condition | HTTP Status | User Message |
|------------|-----------|-------------|--------------|
| ORDER_NOT_FOUND | Order ID doesn't exist | 404 | Pesanan tidak ditemukan |
| ORDER_NOT_PENDING | Order status != MENUNGGU_PEMBAYARAN | 400 | Pesanan tidak dalam status menunggu pembayaran |
| PAYMENT_EXISTS | Active PENDING payment exists | 200 | (Return existing payment) |
| INVALID_PAYMENT_METHOD | Payment method not in allowed list | 400 | Metode pembayaran tidak valid |
| MIDTRANS_ERROR | Midtrans API returns error | 502 | Gagal membuat pembayaran, silakan coba lagi |
| MIDTRANS_TIMEOUT | Midtrans API timeout | 504 | Layanan pembayaran sedang sibuk |

### Payment Retrieval Errors

| Error Code | Condition | HTTP Status | User Message |
|------------|-----------|-------------|--------------|
| PAYMENT_NOT_FOUND | No payment for order | 404 | Pembayaran tidak ditemukan |
| PAYMENT_EXPIRED | Payment has expired | 410 | Pembayaran telah kadaluarsa |
| UNAUTHORIZED | User doesn't own order | 403 | Anda tidak memiliki akses |

### Webhook Processing Errors

| Error Code | Condition | Action |
|------------|-----------|--------|
| INVALID_SIGNATURE | Signature mismatch | Log + Return 200 |
| ORDER_NOT_FOUND | Order code not found | Log + Return 200 |
| PAYMENT_NOT_FOUND | No payment for order | Log + Return 200 |
| ALREADY_PROCESSED | Payment in final state | Skip + Return 200 |

## Testing Strategy

### Dual Testing Approach

This system requires both unit tests and property-based tests:

**Unit Tests** verify specific examples and edge cases:
- Specific VA number formats for each bank
- Specific error responses from Midtrans
- Specific countdown timer edge cases (0 seconds, negative)

**Property-Based Tests** verify universal properties:
- All 18 correctness properties defined above
- Use fast-check library for TypeScript frontend tests
- Use testing/quick or gopter for Go backend tests

### Property-Based Testing Configuration

- **Library**: gopter (Go), fast-check (TypeScript)
- **Iterations**: Minimum 100 per property
- **Shrinking**: Enabled for failure case minimization

### Test Annotation Format

Each property-based test MUST include:
```go
// **Feature: tokopedia-payment-system, Property 3: Payment Creation Idempotency**
// **Validates: Requirements 2.5, 2.6, 3.1**
func TestPaymentCreationIdempotency(t *testing.T) {
    // ...
}
```

### Generator Requirements

Property tests require smart generators for:
- Valid order data (with items, shipping, customer info)
- Valid payment methods (bca_va, bri_va, mandiri_va)
- Valid/invalid webhook signatures
- Expiry times (past, future, edge cases)
- Midtrans order IDs (ORDER_CODE-TIMESTAMP format)

## API Endpoints

### Payment APIs (Authenticated)

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | /api/payments/core/create | Create VA payment | Required |
| GET | /api/payments/core/:order_id | Get payment details | Required |
| POST | /api/payments/core/check | Check payment status | Required |
| GET | /api/pembelian/pending | Get pending orders | Required |
| GET | /api/pembelian/history | Get transaction history | Required |

### Webhook API (Public)

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | /api/webhook/midtrans/core | Midtrans Core API webhook | Signature |

### API Request/Response Details

#### POST /api/payments/core/create

**Request:**
```json
{
  "order_id": 12345,
  "payment_method": "bca_va"
}
```

**Success Response (201 Created):**
```json
{
  "payment_id": 1,
  "order_id": 12345,
  "order_code": "ZVR-20260113-ABC12345",
  "payment_method": "bca_va",
  "bank": "bca",
  "bank_logo": "/images/banks/bca.svg",
  "va_number": "12345678901234567",
  "amount": 758000,
  "expiry_time": "2026-01-14T10:30:00Z",
  "remaining_seconds": 86400,
  "status": "PENDING",
  "instructions": [
    {
      "channel": "ATM BCA",
      "steps": [
        "Masukkan kartu ATM dan PIN",
        "Pilih menu Transaksi Lainnya",
        "Pilih Transfer ke BCA Virtual Account",
        "Masukkan nomor VA: 12345678901234567",
        "Konfirmasi pembayaran"
      ]
    }
  ]
}
```

**Idempotent Response (200 OK):** Same as above if payment already exists

#### GET /api/payments/core/:order_id

**Success Response (200 OK):**
```json
{
  "payment_id": 1,
  "order_id": 12345,
  "order_code": "ZVR-20260113-ABC12345",
  "payment_method": "bca_va",
  "bank": "bca",
  "bank_logo": "/images/banks/bca.svg",
  "va_number": "12345678901234567",
  "amount": 758000,
  "expiry_time": "2026-01-14T10:30:00Z",
  "remaining_seconds": 43200,
  "status": "PENDING",
  "instructions": [...]
}
```

#### POST /api/payments/core/check

**Request:**
```json
{
  "payment_id": 1
}
```

**Response (200 OK):**
```json
{
  "payment_id": 1,
  "status": "PENDING",
  "message": "Pembayaran belum diterima"
}
```

#### GET /api/pembelian/pending

**Query Parameters:** `page=1&page_size=10`

**Response (200 OK):**
```json
{
  "orders": [
    {
      "order_id": 12345,
      "order_code": "ZVR-20260113-ABC12345",
      "total_amount": 758000,
      "item_count": 2,
      "item_summary": "Minimalist Cotton Tee + 1 lainnya",
      "created_at": "2026-01-13T10:30:00Z",
      "has_payment": true,
      "payment_method": "bca_va",
      "bank": "bca",
      "bank_logo": "/images/banks/bca.svg",
      "va_number_masked": "****4567",
      "expiry_time": "2026-01-14T10:30:00Z",
      "remaining_seconds": 43200
    },
    {
      "order_id": 12344,
      "order_code": "ZVR-20260113-XYZ98765",
      "total_amount": 299000,
      "item_count": 1,
      "item_summary": "Classic Denim Jacket",
      "created_at": "2026-01-13T09:00:00Z",
      "has_payment": false
    }
  ],
  "total_count": 2,
  "page": 1,
  "page_size": 10
}
```

#### POST /api/webhook/midtrans/core

**Request (from Midtrans):**
```json
{
  "transaction_time": "2026-01-13 11:30:00",
  "transaction_status": "settlement",
  "transaction_id": "abc123-def456",
  "status_code": "200",
  "signature_key": "sha512hash...",
  "payment_type": "bank_transfer",
  "order_id": "ZVR-20260113-ABC12345-1736765400",
  "gross_amount": "758000.00",
  "va_numbers": [
    {"bank": "bca", "va_number": "12345678901234567"}
  ]
}
```

**Response (200 OK):**
```json
{
  "status": "ok"
}
```

## Frontend Interaction Contract

### What Frontend CAN Do

1. **Call POST /api/payments/core/create** with order_id and payment_method
2. **Call GET /api/payments/core/:order_id** to get payment details
3. **Call POST /api/payments/core/check** to refresh status (rate limited)
4. **Call GET /api/pembelian/pending** to list pending orders
5. **Call GET /api/pembelian/history** to list transaction history
6. **Display countdown timer** using expiry_time from API
7. **Copy VA number** to clipboard
8. **Navigate** between Payment Selection, VA Detail, and Pembelian pages

### What Frontend MUST NOT Do

1. **Call Midtrans API directly** - All Midtrans calls go through backend
2. **Store server_key** - Never expose Midtrans credentials to frontend
3. **Allow payment method change** after payment is created
4. **Create multiple payments** for same order (backend enforces, but UI should prevent)
5. **Show payment actions** for expired/completed orders
6. **Bypass rate limiting** on status check

### State-Driven UI Rules

| Order Status | Payment Status | UI State |
|--------------|----------------|----------|
| MENUNGGU_PEMBAYARAN | NULL | Show "Pilih Pembayaran" button |
| MENUNGGU_PEMBAYARAN | PENDING | Show VA Detail with countdown |
| MENUNGGU_PEMBAYARAN | PENDING (expired) | Show "Kadaluarsa" message |
| DIBAYAR | PAID | Show success state, no actions |
| KADALUARSA | EXPIRED | Show expired state, no actions |
| DIBATALKAN | * | Show cancelled state, no actions |

### Countdown Timer Implementation

```typescript
// Frontend countdown logic
function calculateRemaining(expiryTime: string): number {
  const expiry = new Date(expiryTime).getTime();
  const now = Date.now();
  return Math.max(0, Math.floor((expiry - now) / 1000));
}

// Update every second
useEffect(() => {
  const interval = setInterval(() => {
    const remaining = calculateRemaining(payment.expiry_time);
    setRemainingSeconds(remaining);
    if (remaining <= 0) {
      // Trigger status check to confirm expiry
      checkPaymentStatus();
    }
  }, 1000);
  return () => clearInterval(interval);
}, [payment.expiry_time]);
```

## Sequence Diagrams

### Payment Creation Flow

```
User          Frontend         Backend          Midtrans
 |               |                |                |
 |--Select BCA-->|                |                |
 |               |--POST /create->|                |
 |               |                |--Check existing|
 |               |                |  payment       |
 |               |                |                |
 |               |                |--POST /charge->|
 |               |                |<--VA details---|
 |               |                |                |
 |               |                |--Save payment--|
 |               |<--VA response--|                |
 |<--Show VA-----|                |                |
```

### Payment Resume Flow

```
User          Frontend         Backend          Database
 |               |                |                |
 |--Open order-->|                |                |
 |               |--GET /payment->|                |
 |               |                |--Find payment->|
 |               |                |<--Payment data-|
 |               |                |                |
 |               |                |--Check expiry--|
 |               |                |  (if expired,  |
 |               |                |   update status|
 |               |                |   restore stock)|
 |               |<--Payment data-|                |
 |<--Show VA-----|                |                |
```

### Webhook Processing Flow

```
Midtrans       Backend          Database
   |              |                |
   |--POST /webhook->              |
   |              |--Verify sig----|
   |              |                |
   |              |--Extract order-|
   |              |  code          |
   |              |                |
   |              |--Lock order--->|
   |              |                |
   |              |--Check status->|
   |              |<--Current------|
   |              |                |
   |              |--Update------->|
   |              |  payment       |
   |              |                |
   |              |--Update------->|
   |              |  order         |
   |              |                |
   |              |--Commit------->|
   |<--200 OK-----|                |
```
