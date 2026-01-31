# Refund DTOs Summary

This document provides an overview of all refund-related Data Transfer Objects (DTOs) created for the refund system enhancement.

## Location
All refund DTOs are defined in `backend/dto/hardening_dto.go`

## Request DTOs

### 1. RefundRequest
**Purpose**: Create a new refund request  
**Validation Tags**: 
- `order_code`: required
- `refund_type`: required, must be one of: FULL, PARTIAL, SHIPPING_ONLY, ITEM_ONLY
- `reason`: required
- `amount`: optional (required for PARTIAL refunds)
- `items`: optional (required for ITEM_ONLY refunds)

**Fields**:
```go
OrderCode      string              // Order code to refund
RefundType     string              // Type of refund
Reason         string              // Reason for refund
ReasonDetail   string              // Additional details
Amount         *float64            // Amount for partial refunds
Items          []RefundItemRequest // Items for item-specific refunds
IdempotencyKey string              // Prevent duplicate processing
```

### 2. RefundItemRequest
**Purpose**: Specify items in a partial refund  
**Validation Tags**:
- `order_item_id`: required
- `quantity`: required, minimum 1

**Fields**:
```go
OrderItemID int    // ID of the order item
Quantity    int    // Quantity to refund
Reason      string // Reason for this item
```

### 3. ProcessRefundRequest
**Purpose**: Process a pending refund  
**Fields**:
```go
ProcessedBy int // User ID processing the refund
```

### 4. RetryRefundRequest
**Purpose**: Retry a failed refund  
**Fields**:
```go
ProcessedBy int // User ID retrying the refund
```

## Response DTOs

### 5. RefundResponse
**Purpose**: Complete refund details in API response  
**Fields**:
```go
ID              int                     // Refund ID
RefundCode      string                  // Unique refund code
OrderCode       string                  // Associated order code
OrderID         int                     // Associated order ID
PaymentID       *int                    // Payment ID (nullable for manual refunds)
RefundType      string                  // Type of refund
Reason          string                  // Refund reason
ReasonDetail    string                  // Additional details
OriginalAmount  float64                 // Original order amount
RefundAmount    float64                 // Total refund amount
ShippingRefund  float64                 // Shipping portion
ItemsRefund     float64                 // Items portion
Status          string                  // Current status
GatewayRefundID string                  // Gateway refund ID
GatewayStatus   string                  // Gateway status
IdempotencyKey  string                  // Idempotency key
ProcessedBy     *int                    // User who processed
ProcessedAt     *time.Time              // Processing timestamp
RequestedBy     *int                    // User who requested
RequestedAt     time.Time               // Request timestamp
CompletedAt     *time.Time              // Completion timestamp
CreatedAt       time.Time               // Creation timestamp
UpdatedAt       time.Time               // Last update timestamp
Items           []RefundItemResponse    // Refund items
StatusHistory   []StatusHistoryResponse // Status change history
```

### 6. RefundItemResponse
**Purpose**: Refund item details in API response  
**Fields**:
```go
ID              int        // Item ID
RefundID        int        // Associated refund ID
OrderItemID     int        // Original order item ID
ProductID       int        // Product ID
ProductName     string     // Product name
Quantity        int        // Refunded quantity
PricePerUnit    float64    // Price per unit
RefundAmount    float64    // Total refund for this item
ItemReason      string     // Reason for this item
StockRestored   bool       // Whether stock was restored
StockRestoredAt *time.Time // Stock restoration timestamp
CreatedAt       time.Time  // Creation timestamp
```

### 7. StatusHistoryResponse
**Purpose**: Refund status change history  
**Fields**:
```go
ID        int       // History entry ID
RefundID  int       // Associated refund ID
OldStatus *string   // Previous status
NewStatus string    // New status
Actor     string    // Who made the change
Reason    string    // Reason for change
CreatedAt time.Time // Change timestamp
```

### 8. RefundListResponse
**Purpose**: Paginated list of refunds  
**Fields**:
```go
Refunds    []RefundResponse // List of refunds
TotalCount int              // Total number of refunds
Page       int              // Current page
PageSize   int              // Items per page
```

### 9. RefundSuccessResponse
**Purpose**: Successful refund operation response  
**Fields**:
```go
Success         bool   // Operation success flag
Message         string // Success message
RefundCode      string // Refund code
GatewayRefundID string // Gateway refund ID
```

### 10. RefundErrorResponse
**Purpose**: Refund error response  
**Fields**:
```go
Error   string                 // Error code
Message string                 // Error message
Details map[string]interface{} // Additional error details
```

## Customer-Facing DTOs

### 11. CustomerRefundResponse
**Purpose**: Customer-friendly refund information  
**Fields**:
```go
RefundCode     string               // Refund code
OrderCode      string               // Order code
RefundType     string               // Type of refund
RefundAmount   float64              // Total refund amount
ShippingRefund float64              // Shipping portion
ItemsRefund    float64              // Items portion
Status         string               // Current status
StatusLabel    string               // Human-readable status
Timeline       string               // Estimated timeline message
RequestedAt    time.Time            // Request timestamp
ProcessedAt    *time.Time           // Processing timestamp
CompletedAt    *time.Time           // Completion timestamp
Items          []RefundItemResponse // Refund items
```

### 12. CustomerRefundListResponse
**Purpose**: List of refunds for a customer  
**Fields**:
```go
Refunds []CustomerRefundResponse // List of refunds
Count   int                      // Total count
```

## Midtrans Gateway DTOs

### 13. MidtransRefundRequest
**Purpose**: Request to Midtrans refund API  
**Fields**:
```go
RefundKey string  // Unique refund key
Amount    float64 // Refund amount
Reason    string  // Refund reason
```

### 14. MidtransRefundResponse
**Purpose**: Response from Midtrans refund API  
**Fields**:
```go
StatusCode           string // Response status code
StatusMessage        string // Response message
RefundChargebackID   int    // Midtrans refund ID
RefundAmount         string // Refund amount
RefundKey            string // Refund key
TransactionID        string // Transaction ID
GrossAmount          string // Gross amount
Currency             string // Currency code
OrderID              string // Order ID
PaymentType          string // Payment type
TransactionTime      string // Transaction time
TransactionStatus    string // Transaction status
FraudStatus          string // Fraud status
RefundChargebackTime string // Refund time
Bank                 string // Bank name
```

## Validation Rules

### RefundRequest Validation
- `order_code`: Must be provided
- `refund_type`: Must be one of: FULL, PARTIAL, SHIPPING_ONLY, ITEM_ONLY
- `reason`: Must be provided
- `amount`: Required when refund_type is PARTIAL
- `items`: Required when refund_type is ITEM_ONLY

### RefundItemRequest Validation
- `order_item_id`: Must be provided
- `quantity`: Must be at least 1

## Usage Examples

### Creating a Full Refund
```go
req := dto.RefundRequest{
    OrderCode:      "ORD-2024-001",
    RefundType:     "FULL",
    Reason:         "CUSTOMER_REQUEST",
    ReasonDetail:   "Customer changed mind",
    IdempotencyKey: "unique-key-123",
}
```

### Creating a Partial Refund
```go
amount := 50000.0
req := dto.RefundRequest{
    OrderCode:      "ORD-2024-001",
    RefundType:     "PARTIAL",
    Reason:         "DAMAGED_ITEM",
    ReasonDetail:   "One item was damaged",
    Amount:         &amount,
    IdempotencyKey: "unique-key-456",
}
```

### Creating an Item-Specific Refund
```go
req := dto.RefundRequest{
    OrderCode:  "ORD-2024-001",
    RefundType: "ITEM_ONLY",
    Reason:     "WRONG_ITEM",
    Items: []dto.RefundItemRequest{
        {
            OrderItemID: 1,
            Quantity:    2,
            Reason:      "Wrong color received",
        },
    },
    IdempotencyKey: "unique-key-789",
}
```

### Midtrans Refund Request
```go
req := dto.MidtransRefundRequest{
    RefundKey: "REF-2024-001",
    Amount:    100000.0,
    Reason:    "CUSTOMER_REQUEST",
}
```

## Requirements Coverage

These DTOs satisfy the following requirements from the design document:

- **Requirement 8.1**: Full refund DTO structure
- **Requirement 8.2**: Shipping-only refund support
- **Requirement 8.3**: Partial refund with amount
- **Requirement 8.4**: Item-specific refund with items array
- **Requirement 14.1**: API request/response DTOs for all endpoints

## Testing

All DTOs have been tested for:
- JSON marshaling/unmarshaling
- Field validation
- Nullable field handling
- Nested structure support

Test file: `backend/dto/refund_dto_test.go`

Run tests with:
```bash
go test -v ./dto -run TestRefund
go test -v ./dto -run TestMidtrans
go test -v ./dto -run TestCustomer
```

## Notes

1. All DTOs use proper JSON tags for API serialization
2. Validation tags are included for request DTOs
3. Nullable fields use pointers (*int, *time.Time) to distinguish between zero values and null
4. Customer-facing DTOs include human-readable labels and timeline messages
5. Midtrans DTOs match the actual Midtrans API specification
6. All timestamps use time.Time for proper timezone handling
