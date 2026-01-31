# Design Document: Refund System Enhancement

## Overview

The Refund System Enhancement redesigns the refund processing flow to ensure data integrity, reliable payment gateway integration, and excellent user experience. The system handles full refunds, partial refunds, shipping-only refunds, and item-specific refunds while maintaining proper audit trails and preventing duplicate processing.

### Key Design Goals

1. **Data Integrity**: Eliminate foreign key constraint violations through proper validation and nullable fields
2. **Payment Reliability**: Ensure money actually returns to customers via Midtrans gateway integration
3. **User Experience**: Provide clear refund status and timeline information to customers
4. **Admin Safety**: Implement validation and confirmation workflows to prevent errors
5. **Error Recovery**: Handle failures gracefully with retry mechanisms
6. **Audit Compliance**: Maintain complete audit trails for all refund operations

## Architecture

### System Components

```
┌─────────────────┐         ┌──────────────────┐         ┌─────────────────┐
│  Admin Panel    │────────▶│  Refund Service  │────────▶│  Midtrans API   │
│  (Frontend)     │         │  (Backend)       │         │  (Gateway)      │
└─────────────────┘         └──────────────────┘         └─────────────────┘
                                     │
                                     ▼
                            ┌──────────────────┐
                            │   PostgreSQL     │
                            │   Database       │
                            └──────────────────┘
                                     │
                                     ▼
                            ┌──────────────────┐
                            │  Customer Portal │
                            │  (Frontend)      │
                            └──────────────────┘
```

### Refund Processing Flow

```
1. Admin initiates refund
   ↓
2. Validate refund request (order status, amount, etc.)
   ↓
3. Create refund record in database (with transaction)
   ↓
4. Check if manual refund (no payment record)
   ├─ Yes → Mark as COMPLETED, skip gateway
   └─ No → Continue to gateway processing
   ↓
5. Call Midtrans refund API
   ├─ Success → Store gateway refund ID, mark COMPLETED
   └─ Failure → Mark FAILED, allow retry
   ↓
6. Update order status (REFUNDED if full refund)
   ↓
7. Restore product stock
   ↓
8. Send notification to customer
```

## Components and Interfaces

### 1. Refund Service (Backend)

**File**: `backend/service/refund_service.go`

**Key Methods**:

```go
// Core refund operations
CreateRefund(req *dto.RefundRequest, requestedBy *int) (*models.Refund, error)
ProcessRefund(refundID int, processedBy int) error
GetRefund(refundID int) (*models.Refund, error)
GetRefundsByOrder(orderID int) ([]*models.Refund, error)

// Specific refund types
FullRefund(orderCode string, reason, detail string, requestedBy *int, idempotencyKey string) (*models.Refund, error)
PartialRefund(orderCode string, amount float64, reason, detail string, requestedBy *int, idempotencyKey string) (*models.Refund, error)
ShippingOnlyRefund(orderCode string, reason, detail string, requestedBy *int, idempotencyKey string) (*models.Refund, error)
ItemRefund(orderCode string, items []dto.RefundItemRequest, reason, detail string, requestedBy *int, idempotencyKey string) (*models.Refund, error)

// Gateway operations
ProcessMidtransRefund(refund *models.Refund) (*dto.MidtransRefundResponse, error)
CheckMidtransRefundStatus(refundCode string) (*dto.MidtransRefundResponse, error)

// Helper methods
validateRefundRequest(req *dto.RefundRequest) error
calculateRefundAmount(order *models.Order, payment *models.Payment, req *dto.RefundRequest) (total, shipping, items float64, error)
updateOrderRefundStatus(orderID int) error
restoreRefundedStock(refund *models.Refund) error
```

**Key Improvements**:

1. **Nullable Foreign Keys**: Change `requested_by` and `payment_id` to nullable (`*int`) to handle manual refunds
2. **Transaction Safety**: Wrap all database operations in transactions with row-level locking
3. **Idempotency**: Check idempotency keys before creating refunds
4. **Manual Refund Path**: Separate code path for orders without payment records
5. **Error Context**: Return descriptive errors with context for debugging

### 2. Refund Repository (Backend)

**File**: `backend/repository/refund_repository.go`

**Key Methods**:

```go
// CRUD operations
Create(refund *models.Refund) error
CreateWithTx(tx *sql.Tx, refund *models.Refund) error
FindByID(id int) (*models.Refund, error)
FindByCode(code string) (*models.Refund, error)
FindByOrderID(orderID int) ([]*models.Refund, error)
FindByIdempotencyKey(key string) (*models.Refund, error)

// Status management
UpdateStatus(refundID int, status models.RefundStatus, processedBy *int) error
MarkCompleted(refundID int, gatewayRefundID string, gatewayResponse map[string]any) error
MarkFailed(refundID int, errorMessage string, processedBy *int) error

// Refund items
CreateRefundItem(item *models.RefundItem) error
CreateRefundItemWithTx(tx *sql.Tx, item *models.RefundItem) error
FindItemsByRefundID(refundID int) ([]models.RefundItem, error)
MarkItemStockRestored(itemID int) error

// Status history
RecordStatusChange(refundID int, oldStatus, newStatus models.RefundStatus, actor, reason string) error
GetStatusHistory(refundID int) ([]models.RefundStatusHistory, error)

// Helper methods
GetDB() *sql.DB
```

### 3. Admin Refund API (Backend)

**File**: `backend/handlers/admin_refund_handler.go`

**Endpoints**:

```
POST   /admin/refunds                    - Create a new refund
POST   /admin/refunds/:id/process        - Process a pending refund
POST   /admin/refunds/:id/retry          - Retry a failed refund
GET    /admin/refunds/:id                - Get refund details
GET    /admin/orders/:code/refunds       - List refunds for an order
GET    /admin/refunds                    - List all refunds (with pagination)
```

**Request/Response DTOs**:

```go
// RefundRequest - Create refund
type RefundRequest struct {
    OrderCode      string               `json:"order_code" binding:"required"`
    RefundType     string               `json:"refund_type" binding:"required,oneof=FULL PARTIAL SHIPPING_ONLY ITEM_ONLY"`
    Reason         string               `json:"reason" binding:"required"`
    ReasonDetail   string               `json:"reason_detail"`
    Amount         *float64             `json:"amount,omitempty"`
    Items          []RefundItemRequest  `json:"items,omitempty"`
    IdempotencyKey string               `json:"idempotency_key"`
}

// RefundItemRequest - Item in partial refund
type RefundItemRequest struct {
    OrderItemID int    `json:"order_item_id" binding:"required"`
    Quantity    int    `json:"quantity" binding:"required,min=1"`
    Reason      string `json:"reason"`
}

// RefundResponse - Refund details
type RefundResponse struct {
    ID              int                    `json:"id"`
    RefundCode      string                 `json:"refund_code"`
    OrderCode       string                 `json:"order_code"`
    RefundType      string                 `json:"refund_type"`
    Reason          string                 `json:"reason"`
    ReasonDetail    string                 `json:"reason_detail"`
    OriginalAmount  float64                `json:"original_amount"`
    RefundAmount    float64                `json:"refund_amount"`
    ShippingRefund  float64                `json:"shipping_refund"`
    ItemsRefund     float64                `json:"items_refund"`
    Status          string                 `json:"status"`
    GatewayRefundID string                 `json:"gateway_refund_id,omitempty"`
    ProcessedBy     *int                   `json:"processed_by,omitempty"`
    ProcessedAt     *time.Time             `json:"processed_at,omitempty"`
    RequestedBy     *int                   `json:"requested_by,omitempty"`
    RequestedAt     time.Time              `json:"requested_at"`
    CompletedAt     *time.Time             `json:"completed_at,omitempty"`
    Items           []RefundItemResponse   `json:"items,omitempty"`
    StatusHistory   []StatusHistoryResponse `json:"status_history,omitempty"`
}
```

### 4. Customer Refund API (Backend)

**File**: `backend/handlers/customer_refund_handler.go`

**Endpoints**:

```
GET    /customer/orders/:code/refunds    - Get refunds for customer's order
GET    /customer/refunds/:code           - Get refund details by refund code
```

### 5. Admin Panel UI (Frontend)

**File**: `frontend/src/app/admin/orders/[code]/page.tsx`

**Refund Modal Component**:

```typescript
interface RefundModalProps {
  order: OrderDetail;
  onClose: () => void;
  onSuccess: () => void;
}

const RefundModal: React.FC<RefundModalProps> = ({ order, onClose, onSuccess }) => {
  const [refundType, setRefundType] = useState<'FULL' | 'PARTIAL' | 'SHIPPING_ONLY' | 'ITEM_ONLY'>('FULL');
  const [reason, setReason] = useState('');
  const [reasonDetail, setReasonDetail] = useState('');
  const [amount, setAmount] = useState<number | null>(null);
  const [selectedItems, setSelectedItems] = useState<RefundItemSelection[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  // Validation logic
  // Submit handler
  // UI rendering
};
```

**Key Features**:
- Refund type selection (Full, Partial, Shipping Only, Item Only)
- Reason selection with predefined options
- Amount input for partial refunds
- Item selection for item-only refunds
- Validation before submission
- Loading states and error handling
- Success confirmation with gateway refund ID

### 6. Customer Portal UI (Frontend)

**File**: `frontend/src/app/account/pembelian/page.tsx`

**Refund Status Display**:

```typescript
// Enhanced TransactionCard component
const TransactionCard = ({ order }: { order: TransactionHistoryItem }) => {
  // ... existing code ...

  // Add refund information section
  {order.refund_status && (
    <div className="mt-4 p-4 bg-orange-50 rounded-lg border border-orange-100">
      <div className="flex items-center gap-2 mb-2">
        <RefreshCcw className="text-orange-600" size={16} />
        <span className="text-sm font-semibold text-orange-800">
          Refund {order.refund_status}
        </span>
      </div>
      <div className="space-y-1 text-sm">
        <div className="flex justify-between">
          <span className="text-gray-600">Refund Amount:</span>
          <span className="font-semibold text-gray-900">
            {formatCurrency(order.refund_amount)}
          </span>
        </div>
        <div className="flex justify-between">
          <span className="text-gray-600">Status:</span>
          <span className={`font-medium ${getRefundStatusColor(order.refund_status)}`}>
            {getRefundStatusLabel(order.refund_status)}
          </span>
        </div>
        {order.refund_completed_at && (
          <div className="flex justify-between">
            <span className="text-gray-600">Completed:</span>
            <span className="text-gray-900">
              {formatDate(order.refund_completed_at)}
            </span>
          </div>
        )}
        <p className="text-xs text-gray-500 mt-2">
          {getRefundTimeline(order.payment_method, order.refund_status)}
        </p>
      </div>
      <Link
        href={`/orders/${order.order_code}#refund`}
        className="text-sm font-medium text-orange-700 hover:underline mt-2 inline-block"
      >
        View Refund Details →
      </Link>
    </div>
  )}
};
```

**File**: `frontend/src/app/orders/[code]/page.tsx`

**Refund Details Section**:

```typescript
// New component for order detail page
const RefundDetailsSection = ({ orderCode }: { orderCode: string }) => {
  const [refunds, setRefunds] = useState<RefundDetail[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadRefunds();
  }, [orderCode]);

  const loadRefunds = async () => {
    try {
      const response = await api.get(`/customer/orders/${orderCode}/refunds`);
      setRefunds(response.data);
    } catch (error) {
      console.error('Failed to load refunds:', error);
    } finally {
      setLoading(false);
    }
  };

  // Render refund timeline, status, and details
};
```

## Data Models

### Database Schema Changes

**1. Refunds Table** - Make foreign keys nullable:

```sql
ALTER TABLE refunds 
  ALTER COLUMN requested_by DROP NOT NULL,
  ALTER COLUMN payment_id DROP NOT NULL;

-- Add index for idempotency key
CREATE INDEX idx_refunds_idempotency_key ON refunds(idempotency_key) WHERE idempotency_key IS NOT NULL;

-- Add index for order_id lookups
CREATE INDEX idx_refunds_order_id ON refunds(order_id);
```

**2. Orders Table** - Add refund tracking fields:

```sql
ALTER TABLE orders
  ADD COLUMN IF NOT EXISTS refund_status VARCHAR(20),
  ADD COLUMN IF NOT EXISTS refund_amount DECIMAL(10,2) DEFAULT 0,
  ADD COLUMN IF NOT EXISTS refunded_at TIMESTAMP;

-- Add index for refund status queries
CREATE INDEX idx_orders_refund_status ON orders(refund_status) WHERE refund_status IS NOT NULL;
```

**3. Refund Status History Table** - Track all status changes:

```sql
CREATE TABLE IF NOT EXISTS refund_status_history (
  id SERIAL PRIMARY KEY,
  refund_id INTEGER NOT NULL REFERENCES refunds(id) ON DELETE CASCADE,
  old_status VARCHAR(20),
  new_status VARCHAR(20) NOT NULL,
  actor VARCHAR(100) NOT NULL,
  reason TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_refund_status_history_refund_id (refund_id)
);
```

### Go Models

**Refund Model** - Updated with nullable fields:

```go
type Refund struct {
    ID               int            `json:"id" db:"id"`
    RefundCode       string         `json:"refund_code" db:"refund_code"`
    OrderID          int            `json:"order_id" db:"order_id"`
    PaymentID        *int           `json:"payment_id,omitempty" db:"payment_id"` // NULLABLE
    RefundType       RefundType     `json:"refund_type" db:"refund_type"`
    Reason           RefundReason   `json:"reason" db:"reason"`
    ReasonDetail     string         `json:"reason_detail,omitempty" db:"reason_detail"`
    OriginalAmount   float64        `json:"original_amount" db:"original_amount"`
    RefundAmount     float64        `json:"refund_amount" db:"refund_amount"`
    ShippingRefund   float64        `json:"shipping_refund" db:"shipping_refund"`
    ItemsRefund      float64        `json:"items_refund" db:"items_refund"`
    Status           RefundStatus   `json:"status" db:"status"`
    GatewayRefundID  string         `json:"gateway_refund_id,omitempty" db:"gateway_refund_id"`
    GatewayStatus    string         `json:"gateway_status,omitempty" db:"gateway_status"`
    GatewayResponse  map[string]any `json:"gateway_response,omitempty" db:"gateway_response"`
    IdempotencyKey   string         `json:"idempotency_key,omitempty" db:"idempotency_key"`
    ProcessedBy      *int           `json:"processed_by,omitempty" db:"processed_by"` // NULLABLE
    ProcessedAt      *time.Time     `json:"processed_at,omitempty" db:"processed_at"`
    RequestedBy      *int           `json:"requested_by,omitempty" db:"requested_by"` // NULLABLE
    RequestedAt      time.Time      `json:"requested_at" db:"requested_at"`
    CreatedAt        time.Time      `json:"created_at" db:"created_at"`
    UpdatedAt        time.Time      `json:"updated_at" db:"updated_at"`
    CompletedAt      *time.Time     `json:"completed_at,omitempty" db:"completed_at"`
    Items            []RefundItem   `json:"items,omitempty" db:"-"`
}
```

**Order Model** - Add refund fields:

```go
type Order struct {
    // ... existing fields ...
    RefundStatus  *string    `json:"refund_status,omitempty" db:"refund_status"`
    RefundAmount  float64    `json:"refund_amount" db:"refund_amount"`
    RefundedAt    *time.Time `json:"refunded_at,omitempty" db:"refunded_at"`
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*


### Property Reflection

After analyzing all acceptance criteria, I identified several redundant properties that can be consolidated:

**Redundancies Identified:**
1. Properties 9.2, 9.3, 9.4, 9.5 are redundant with 9.1, 1.6, 8.6, and 1.6 respectively
2. Property 10.6 is redundant with 2.8 (gateway response storage)
3. Properties 13.1, 13.2, 13.3 are redundant with 1.2 and 2.7 (manual refund handling)
4. Property 15.4 is redundant with 8.6 (refund amount validation)

**Consolidation Strategy:**
- Combine idempotency checking (9.1, 9.2) into a single comprehensive property
- Combine concurrency control (1.6, 9.3, 9.5) into a single property
- Remove duplicate gateway response storage property
- Remove duplicate manual refund properties
- Remove duplicate validation properties

This reduces the total number of properties while maintaining complete coverage of all requirements.

### Correctness Properties

**Property 1: Foreign Key Validation**
*For any* refund creation request with a `requested_by` user ID, if the user ID does not exist in the users table, then the refund creation SHALL fail with a descriptive error message identifying the missing user reference.
**Validates: Requirements 1.1, 1.3**

**Property 2: Nullable Payment ID Handling**
*For any* order (with or without a payment record), creating a refund SHALL succeed regardless of whether `payment_id` is NULL or references a valid payment.
**Validates: Requirements 1.2**

**Property 3: Transaction Atomicity**
*For any* refund creation attempt, either all database changes (refund record, refund items, status history) SHALL be committed together, or all changes SHALL be rolled back if any operation fails.
**Validates: Requirements 1.4, 1.5, 7.1**

**Property 4: Concurrent Refund Prevention**
*For any* order, when multiple refund creation requests are processed concurrently, the system SHALL use row-level locking to ensure only one refund is created at a time, preventing race conditions and duplicate refunds.
**Validates: Requirements 1.6, 9.3, 9.5**

**Property 5: Order Existence Validation**
*For any* refund creation request with an order code, if the order does not exist, then the refund creation SHALL fail with an error message indicating the order was not found.
**Validates: Requirements 1.7**

**Property 6: Gateway Refund Success Handling**
*For any* refund processed through Midtrans that returns a successful response (status code 200 or 201), the system SHALL store the gateway refund ID in the `gateway_refund_id` field and mark the refund status as COMPLETED.
**Validates: Requirements 2.3**

**Property 7: Gateway Refund Failure Handling**
*For any* refund processed through Midtrans that returns an error response, the system SHALL mark the refund status as FAILED and store the error message in the database.
**Validates: Requirements 2.4, 7.2, 7.3**

**Property 8: Payment Method Consistency**
*For any* refund regardless of payment method (VA, E_Wallet, QRIS, Credit Card), the system SHALL use the same Midtrans refund API endpoint format: `/v2/{order_code}/refund`.
**Validates: Requirements 2.6**

**Property 9: Manual Refund Processing**
*For any* refund where the order has no payment record (`payment_id` is NULL), the system SHALL skip Midtrans gateway processing, set status to COMPLETED immediately, and set `gateway_refund_id` to "MANUAL_REFUND".
**Validates: Requirements 2.7, 13.2, 13.3, 13.4**

**Property 10: Gateway Response Persistence**
*For any* refund that calls the Midtrans API (success or failure), the complete gateway response SHALL be stored in the `gateway_response` field for audit purposes.
**Validates: Requirements 2.8, 10.6**

**Property 11: Refund Status Display**
*For any* refunded order displayed in the customer portal, the rendered output SHALL contain the refund status, refund amount, and processing date.
**Validates: Requirements 3.1, 3.2, 3.3**

**Property 12: Refund Timeline Display**
*For any* refunded order displayed in the customer portal, the rendered output SHALL contain an estimated refund completion timeline based on the payment method.
**Validates: Requirements 3.4, 12.5**

**Property 13: Refund Button Visibility**
*For any* order with status DELIVERED or COMPLETED displayed in the admin panel, the rendered output SHALL contain a "Refund" action button.
**Validates: Requirements 4.1**

**Property 14: Refund Amount Validation**
*For any* refund request, if the refund amount exceeds the refundable balance (order total minus sum of completed refunds), then the refund creation SHALL fail with an error message indicating the amount exceeds the refundable balance.
**Validates: Requirements 4.4, 8.6, 9.4, 15.4**

**Property 15: Refund History Display**
*For any* order with one or more refunds, the admin panel SHALL display all refund attempts with their timestamps, statuses, and amounts in chronological order.
**Validates: Requirements 4.8, 10.7**

**Property 16: Full Refund Order Status Update**
*For any* completed full refund, the order status SHALL be updated to REFUNDED and the order's `refund_status` field SHALL be set to "FULL".
**Validates: Requirements 5.1, 5.4**

**Property 17: Partial Refund Order Status Preservation**
*For any* completed partial refund, the order status SHALL remain as DELIVERED or COMPLETED (unchanged) and the order's `refund_status` field SHALL be set to "PARTIAL".
**Validates: Requirements 5.2, 5.4**

**Property 18: Order Status Change Audit**
*For any* order status change triggered by a refund, a corresponding audit record SHALL be created in the order status history table with the old status, new status, actor, and reason.
**Validates: Requirements 5.3**

**Property 19: Refund Amount Aggregation**
*For any* order with multiple completed refunds, the order's `refund_amount` field SHALL equal the sum of all completed refund amounts.
**Validates: Requirements 5.5, 5.6**

**Property 20: First Refund Timestamp**
*For any* order receiving its first completed refund, the order's `refunded_at` timestamp SHALL be set to the refund completion time.
**Validates: Requirements 5.7**

**Property 21: Full Refund Stock Restoration**
*For any* completed full refund, the product stock for all order items SHALL be incremented by their ordered quantities.
**Validates: Requirements 6.1, 6.3**

**Property 22: Partial Refund Stock Restoration**
*For any* completed item-specific refund, the product stock SHALL be incremented only for the refunded items by their refunded quantities.
**Validates: Requirements 6.2, 6.3**

**Property 23: Stock Restoration Flag**
*For any* refund item after stock restoration completes, the `stock_restored` flag SHALL be set to true.
**Validates: Requirements 6.4**

**Property 24: Stock Restoration Error Isolation**
*For any* refund where stock restoration fails, the refund SHALL still be marked as COMPLETED and the error SHALL be logged without failing the entire refund operation.
**Validates: Requirements 6.5**

**Property 25: Shipping Refund Stock Preservation**
*For any* completed shipping-only refund, product stock SHALL remain unchanged (no stock restoration).
**Validates: Requirements 6.6**

**Property 26: Stock Restoration Idempotency**
*For any* refund item, calling stock restoration multiple times SHALL only increment product stock once (idempotent operation).
**Validates: Requirements 6.7**

**Property 27: Retry Idempotency**
*For any* failed refund that is retried, the system SHALL use the same `idempotency_key` for all retry attempts to prevent duplicate gateway charges.
**Validates: Requirements 7.5, 9.5**

**Property 28: Error Logging**
*For any* error that occurs during refund processing, the system SHALL log the error with sufficient context including refund code, order code, error message, and stack trace.
**Validates: Requirements 7.7**

**Property 29: Full Refund Amount Calculation**
*For any* FULL refund type, the `refund_amount` SHALL equal the order's `total_amount`, `shipping_refund` SHALL equal `shipping_cost`, and `items_refund` SHALL equal `subtotal`.
**Validates: Requirements 8.1, 8.7**

**Property 30: Shipping Only Refund Amount Calculation**
*For any* SHIPPING_ONLY refund type, the `refund_amount` SHALL equal the order's `shipping_cost`, `shipping_refund` SHALL equal `shipping_cost`, and `items_refund` SHALL be zero.
**Validates: Requirements 8.2, 8.7**

**Property 31: Partial Refund Amount Calculation**
*For any* PARTIAL refund type with a specified amount, the `refund_amount` SHALL equal the specified amount (if it does not exceed the refundable balance).
**Validates: Requirements 8.3**

**Property 32: Item Refund Amount Calculation**
*For any* ITEM_ONLY refund type, the `refund_amount` SHALL equal the sum of (quantity × price_per_unit) for all selected refund items.
**Validates: Requirements 8.4**

**Property 33: Refundable Balance Calculation**
*For any* order, the refundable balance SHALL equal the order's `total_amount` minus the sum of all completed refund amounts for that order.
**Validates: Requirements 8.5**

**Property 34: Idempotency Key Checking**
*For any* refund creation request with an idempotency key that already exists, the system SHALL return the existing refund record without creating a new refund.
**Validates: Requirements 9.1, 9.2**

**Property 35: Idempotency Key Storage**
*For any* refund created with an idempotency key, the key SHALL be stored in the `idempotency_key` field for audit purposes.
**Validates: Requirements 9.6**

**Property 36: Refund Creation Audit**
*For any* refund created, the system SHALL record the `requested_by` user ID and `requested_at` timestamp.
**Validates: Requirements 10.1**

**Property 37: Status Change Audit**
*For any* refund status change, the system SHALL create a record in the `refund_status_history` table containing the old status, new status, actor, reason, and timestamp.
**Validates: Requirements 10.2, 10.3**

**Property 38: Refund Processing Audit**
*For any* refund that is processed by an administrator, the system SHALL record the `processed_by` user ID and `processed_at` timestamp.
**Validates: Requirements 10.4**

**Property 39: Refund Completion Audit**
*For any* refund that reaches COMPLETED status, the system SHALL record the `completed_at` timestamp.
**Validates: Requirements 10.5**

**Property 40: Notification Content Completeness**
*For any* refund notification email sent to a customer, the email SHALL include the order code, refund amount, refund status, and a link to view refund details.
**Validates: Requirements 11.5, 11.6**

**Property 41: Manual Refund Amount Calculation**
*For any* manual refund (order without payment record), the refund amount calculation SHALL use the order's total amounts instead of payment amounts.
**Validates: Requirements 13.5**

**Property 42: Manual Refund Display Indicator**
*For any* manual refund displayed in the admin panel, the UI SHALL clearly indicate it was processed manually (e.g., showing "MANUAL_REFUND" as the gateway ID).
**Validates: Requirements 13.6**

**Property 43: Manual Refund Side Effects**
*For any* completed manual refund, the system SHALL still update the order status and restore product stock as it would for a regular refund.
**Validates: Requirements 13.7**

**Property 44: API Authentication Validation**
*For any* refund API endpoint call without valid authentication credentials, the system SHALL return a 401 Unauthorized or 403 Forbidden response.
**Validates: Requirements 14.6**

**Property 45: API Response Completeness**
*For any* refund retrieval API call, the response SHALL include the refund details, all refund items, and the complete status history.
**Validates: Requirements 14.7**

**Property 46: Order Status Validation**
*For any* refund creation request, if the order status is not DELIVERED or COMPLETED, then the refund creation SHALL fail with an error message indicating the order is not eligible for refund.
**Validates: Requirements 15.1**

**Property 47: Payment Status Validation**
*For any* refund creation request for an order with a payment record, if the payment status is not SUCCESS, then the refund creation SHALL fail with an error message indicating the payment is not settled.
**Validates: Requirements 15.2**

**Property 48: Refund Amount Positivity Validation**
*For any* refund creation request, if the refund amount is zero or negative, then the refund creation SHALL fail with an error message indicating the amount must be positive.
**Validates: Requirements 15.3**

**Property 49: Item Existence Validation**
*For any* item refund creation request, if any specified order item ID does not exist in the order, then the refund creation SHALL fail with an error message identifying the non-existent item.
**Validates: Requirements 15.5**

**Property 50: Item Quantity Validation**
*For any* item refund creation request, if any refund quantity exceeds the ordered quantity for that item, then the refund creation SHALL fail with an error message indicating the quantity exceeds the ordered amount.
**Validates: Requirements 15.6**

**Property 51: Validation Error Handling**
*For any* refund creation request that fails validation, no refund record SHALL be created in the database and a descriptive error message SHALL be returned to the caller.
**Validates: Requirements 15.7**

## Error Handling

### Error Categories

1. **Validation Errors** (400 Bad Request)
   - Invalid refund type
   - Refund amount exceeds refundable balance
   - Order not eligible for refund
   - Invalid item IDs or quantities
   - Missing required fields

2. **Authorization Errors** (401/403)
   - Missing authentication token
   - Invalid authentication token
   - Insufficient permissions

3. **Not Found Errors** (404)
   - Order not found
   - Refund not found
   - User not found

4. **Conflict Errors** (409)
   - Idempotency key conflict (returns existing refund)
   - Concurrent refund attempt (retry with backoff)

5. **Gateway Errors** (502/503)
   - Midtrans API timeout
   - Midtrans API error response
   - Network connectivity issues

6. **Database Errors** (500)
   - Foreign key constraint violation
   - Transaction deadlock
   - Connection pool exhaustion

### Error Response Format

```json
{
  "error": "REFUND_AMOUNT_EXCEEDS_BALANCE",
  "message": "Refund amount Rp 1,000,000 exceeds refundable balance Rp 850,000",
  "details": {
    "order_code": "ORD-2024-001",
    "requested_amount": 1000000,
    "refundable_balance": 850000,
    "total_refunded": 150000
  },
  "timestamp": "2024-01-13T10:30:00Z"
}
```

### Retry Strategy

**Gateway Failures**:
- Automatic retry: No (requires manual admin retry)
- Retry mechanism: Admin clicks "Retry" button in UI
- Idempotency: Same idempotency key used for all retries
- Max retries: Unlimited (admin decides when to stop)

**Database Failures**:
- Automatic retry: Yes (3 attempts with exponential backoff)
- Backoff: 100ms, 200ms, 400ms
- Transaction rollback: Automatic on all failures

## Testing Strategy

### Dual Testing Approach

The refund system requires both unit tests and property-based tests for comprehensive coverage:

**Unit Tests**: Verify specific examples, edge cases, and error conditions
- Specific refund scenarios (full refund of Rp 100,000 order)
- Edge cases (refund with zero shipping cost, single-item order)
- Error conditions (invalid order code, missing payment)
- Integration points (Midtrans API mocking, database transactions)

**Property Tests**: Verify universal properties across all inputs
- Refund amount calculations for all refund types
- Stock restoration for all item combinations
- Idempotency for all concurrent scenarios
- Transaction atomicity for all failure points

### Property-Based Testing Configuration

**Library**: Use `gopter` for Go property-based testing

**Configuration**:
- Minimum 100 iterations per property test
- Each property test references its design document property
- Tag format: `// Feature: refund-system-enhancement, Property {number}: {property_text}`

**Example Property Test**:

```go
// Feature: refund-system-enhancement, Property 29: Full Refund Amount Calculation
func TestProperty_FullRefundAmountCalculation(t *testing.T) {
    parameters := gopter.DefaultTestParameters()
    parameters.MinSuccessfulTests = 100
    
    properties := gopter.NewProperties(parameters)
    
    properties.Property("Full refund amount equals order total", prop.ForAll(
        func(order *models.Order) bool {
            refund, err := refundService.FullRefund(
                order.OrderCode,
                models.RefundReasonCustomerRequest,
                "Test refund",
                nil,
                generateIdempotencyKey(),
            )
            
            if err != nil {
                return false
            }
            
            return refund.RefundAmount == order.TotalAmount &&
                   refund.ShippingRefund == order.ShippingCost &&
                   refund.ItemsRefund == order.Subtotal
        },
        genValidOrder(),
    ))
    
    properties.TestingRun(t)
}
```

### Test Coverage Goals

- Unit test coverage: >80% of refund service code
- Property test coverage: All 51 correctness properties
- Integration test coverage: All API endpoints
- E2E test coverage: Critical user flows (admin refund, customer view)

### Testing Environments

1. **Local Development**: SQLite in-memory database, mocked Midtrans
2. **CI/CD Pipeline**: PostgreSQL test database, mocked Midtrans
3. **Staging**: PostgreSQL staging database, Midtrans sandbox
4. **Production**: PostgreSQL production database, Midtrans production

## Implementation Notes

### Database Migration

```sql
-- Migration: Add refund tracking to orders table
ALTER TABLE orders
  ADD COLUMN IF NOT EXISTS refund_status VARCHAR(20),
  ADD COLUMN IF NOT EXISTS refund_amount DECIMAL(10,2) DEFAULT 0,
  ADD COLUMN IF NOT EXISTS refunded_at TIMESTAMP;

-- Migration: Make refunds foreign keys nullable
ALTER TABLE refunds 
  ALTER COLUMN requested_by DROP NOT NULL,
  ALTER COLUMN payment_id DROP NOT NULL;

-- Migration: Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_refunds_idempotency_key 
  ON refunds(idempotency_key) WHERE idempotency_key IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_refunds_order_id ON refunds(order_id);
CREATE INDEX IF NOT EXISTS idx_orders_refund_status 
  ON orders(refund_status) WHERE refund_status IS NOT NULL;

-- Migration: Create refund status history table
CREATE TABLE IF NOT EXISTS refund_status_history (
  id SERIAL PRIMARY KEY,
  refund_id INTEGER NOT NULL REFERENCES refunds(id) ON DELETE CASCADE,
  old_status VARCHAR(20),
  new_status VARCHAR(20) NOT NULL,
  actor VARCHAR(100) NOT NULL,
  reason TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_refund_status_history_refund_id 
  ON refund_status_history(refund_id);
```

### Midtrans API Integration

**Endpoint**: `POST https://api.sandbox.midtrans.com/v2/{order_code}/refund`

**Authentication**: Basic Auth with server key

**Request Body**:
```json
{
  "refund_key": "REFUND-2024-001",
  "amount": 100000,
  "reason": "CUSTOMER_REQUEST"
}
```

**Success Response** (200/201):
```json
{
  "status_code": "200",
  "status_message": "Success, refund is processed",
  "refund_chargeback_id": 12345,
  "refund_amount": "100000.00",
  "refund_key": "REFUND-2024-001"
}
```

**Error Response** (400/500):
```json
{
  "status_code": "412",
  "status_message": "Merchant cannot modify the status of the transaction",
  "id": "ORD-2024-001"
}
```

### Performance Considerations

1. **Database Locking**: Use `SELECT ... FOR UPDATE` to prevent race conditions
2. **Connection Pooling**: Ensure sufficient database connections for concurrent refunds
3. **API Timeouts**: Set 30-second timeout for Midtrans API calls
4. **Batch Operations**: Process stock restoration in batches for large orders
5. **Caching**: Cache refund status for frequently accessed orders

### Security Considerations

1. **Authentication**: All admin endpoints require valid JWT token
2. **Authorization**: Verify admin role before allowing refund operations
3. **Audit Logging**: Log all refund operations with user ID and timestamp
4. **Input Validation**: Sanitize all user inputs to prevent SQL injection
5. **Rate Limiting**: Limit refund creation to 10 requests per minute per admin
6. **Idempotency**: Prevent duplicate refunds using idempotency keys

### Monitoring and Alerting

1. **Metrics to Track**:
   - Refund creation rate
   - Refund success/failure rate
   - Gateway API response times
   - Database transaction durations
   - Stock restoration failures

2. **Alerts to Configure**:
   - Gateway failure rate > 5%
   - Refund processing time > 10 seconds
   - Database transaction deadlocks
   - Stock restoration failures
   - Concurrent refund conflicts

3. **Logging Requirements**:
   - Log all refund operations (INFO level)
   - Log all gateway API calls (DEBUG level)
   - Log all errors with stack traces (ERROR level)
   - Log all audit trail events (INFO level)
