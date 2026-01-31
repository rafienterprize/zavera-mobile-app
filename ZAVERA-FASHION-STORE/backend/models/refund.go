package models

import "time"

// RefundStatus represents the status of a refund
type RefundStatus string

const (
	RefundStatusPending    RefundStatus = "PENDING"
	RefundStatusProcessing RefundStatus = "PROCESSING"
	RefundStatusPartial    RefundStatus = "PARTIAL"
	RefundStatusCompleted  RefundStatus = "COMPLETED"
	RefundStatusFailed     RefundStatus = "FAILED"
	RefundStatusRejected   RefundStatus = "REJECTED"
	RefundStatusCancelled  RefundStatus = "CANCELLED"
)

// RefundType represents the type of refund
type RefundType string

const (
	RefundTypeFull         RefundType = "FULL"
	RefundTypePartial      RefundType = "PARTIAL"
	RefundTypeShippingOnly RefundType = "SHIPPING_ONLY"
	RefundTypeItemOnly     RefundType = "ITEM_ONLY"
)

// RefundReason represents the reason for refund
type RefundReason string

const (
	RefundReasonCustomerRequest RefundReason = "CUSTOMER_REQUEST"
	RefundReasonOutOfStock      RefundReason = "OUT_OF_STOCK"
	RefundReasonDamagedItem     RefundReason = "DAMAGED_ITEM"
	RefundReasonWrongItem       RefundReason = "WRONG_ITEM"
	RefundReasonLateDelivery    RefundReason = "LATE_DELIVERY"
	RefundReasonDuplicateOrder  RefundReason = "DUPLICATE_ORDER"
	RefundReasonFraudSuspected  RefundReason = "FRAUD_SUSPECTED"
	RefundReasonAdminDecision   RefundReason = "ADMIN_DECISION"
	RefundReasonShippingFailed  RefundReason = "SHIPPING_FAILED"
	RefundReasonOther           RefundReason = "OTHER"
)

// Refund represents a refund request
type Refund struct {
	ID               int            `json:"id" db:"id"`
	RefundCode       string         `json:"refund_code" db:"refund_code"`
	OrderID          int            `json:"order_id" db:"order_id"`
	PaymentID        *int           `json:"payment_id,omitempty" db:"payment_id"`
	RefundType       RefundType     `json:"refund_type" db:"refund_type"`
	Reason           RefundReason   `json:"reason" db:"reason"`
	ReasonDetail     string         `json:"reason_detail,omitempty" db:"reason_detail"`
	OriginalAmount   float64        `json:"original_amount" db:"original_amount"`
	RefundAmount     float64        `json:"refund_amount" db:"refund_amount"`
	ShippingRefund   float64        `json:"shipping_refund" db:"shipping_refund"`
	ItemsRefund      float64        `json:"items_refund" db:"items_refund"`
	Status           RefundStatus   `json:"status" db:"status"`
	GatewayRefundID  *string        `json:"gateway_refund_id,omitempty" db:"gateway_refund_id"`
	GatewayStatus    *string        `json:"gateway_status,omitempty" db:"gateway_status"`
	GatewayResponse  map[string]any `json:"gateway_response,omitempty" db:"gateway_response"`
	IdempotencyKey   *string        `json:"idempotency_key,omitempty" db:"idempotency_key"`
	ProcessedBy      *int           `json:"processed_by,omitempty" db:"processed_by"`
	ProcessedAt      *time.Time     `json:"processed_at,omitempty" db:"processed_at"`
	RequestedBy      *int           `json:"requested_by,omitempty" db:"requested_by"`
	RequestedAt      time.Time      `json:"requested_at" db:"requested_at"`
	CreatedAt        time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at" db:"updated_at"`
	CompletedAt      *time.Time     `json:"completed_at,omitempty" db:"completed_at"`
	Items            []RefundItem   `json:"items,omitempty" db:"-"`
}

// RefundItem represents an item in a partial refund
type RefundItem struct {
	ID              int        `json:"id" db:"id"`
	RefundID        int        `json:"refund_id" db:"refund_id"`
	OrderItemID     int        `json:"order_item_id" db:"order_item_id"`
	ProductID       int        `json:"product_id" db:"product_id"`
	ProductName     string     `json:"product_name" db:"product_name"`
	Quantity        int        `json:"quantity" db:"quantity"`
	PricePerUnit    float64    `json:"price_per_unit" db:"price_per_unit"`
	RefundAmount    float64    `json:"refund_amount" db:"refund_amount"`
	ItemReason      string     `json:"item_reason,omitempty" db:"item_reason"`
	StockRestored   bool       `json:"stock_restored" db:"stock_restored"`
	StockRestoredAt *time.Time `json:"stock_restored_at,omitempty" db:"stock_restored_at"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
}

// RefundStatusHistory represents a status change in a refund for audit trail
type RefundStatusHistory struct {
	ID        int          `json:"id" db:"id"`
	RefundID  int          `json:"refund_id" db:"refund_id"`
	OldStatus *string      `json:"old_status,omitempty" db:"old_status"`
	NewStatus RefundStatus `json:"new_status" db:"new_status"`
	Actor     string       `json:"actor" db:"actor"`
	Reason    string       `json:"reason,omitempty" db:"reason"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
}

// IsFinalStatus checks if refund is in a terminal state
func (s RefundStatus) IsFinalStatus() bool {
	switch s {
	case RefundStatusCompleted, RefundStatusFailed, RefundStatusRejected, RefundStatusCancelled:
		return true
	}
	return false
}

// CanProcess checks if refund can be processed
func (s RefundStatus) CanProcess() bool {
	return s == RefundStatusPending
}

// CanCancel checks if refund can be cancelled
func (s RefundStatus) CanCancel() bool {
	return s == RefundStatusPending || s == RefundStatusProcessing
}
