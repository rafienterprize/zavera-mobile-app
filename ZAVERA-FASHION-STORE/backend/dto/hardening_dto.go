package dto

import "time"

// ============================================
// REFUND DTOs
// ============================================

// RefundRequest represents a refund request
type RefundRequest struct {
	OrderCode      string              `json:"order_code" binding:"required"`
	RefundType     string              `json:"refund_type" binding:"required,oneof=FULL PARTIAL SHIPPING_ONLY ITEM_ONLY"`
	Reason         string              `json:"reason" binding:"required"`
	ReasonDetail   string              `json:"reason_detail,omitempty"`
	Amount         *float64            `json:"amount,omitempty"`         // For partial refunds
	Items          []RefundItemRequest `json:"items,omitempty"`          // For item-specific refunds
	IdempotencyKey string              `json:"idempotency_key,omitempty"`
}

// RefundItemRequest represents an item in a refund request
type RefundItemRequest struct {
	OrderItemID int    `json:"order_item_id" binding:"required"`
	Quantity    int    `json:"quantity" binding:"required,min=1"`
	Reason      string `json:"reason,omitempty"`
}

// RefundResponse represents a refund response with complete details
type RefundResponse struct {
	ID              int                     `json:"id"`
	RefundCode      string                  `json:"refund_code"`
	OrderCode       string                  `json:"order_code"`
	OrderID         int                     `json:"order_id"`
	PaymentID       *int                    `json:"payment_id,omitempty"`
	RefundType      string                  `json:"refund_type"`
	Reason          string                  `json:"reason"`
	ReasonDetail    string                  `json:"reason_detail,omitempty"`
	OriginalAmount  float64                 `json:"original_amount"`
	RefundAmount    float64                 `json:"refund_amount"`
	ShippingRefund  float64                 `json:"shipping_refund"`
	ItemsRefund     float64                 `json:"items_refund"`
	Status          string                  `json:"status"`
	GatewayRefundID *string                 `json:"gateway_refund_id,omitempty"`
	GatewayStatus   *string                 `json:"gateway_status,omitempty"`
	IdempotencyKey  *string                 `json:"idempotency_key,omitempty"`
	ProcessedBy     *int                    `json:"processed_by,omitempty"`
	ProcessedAt     *time.Time              `json:"processed_at,omitempty"`
	RequestedBy     *int                    `json:"requested_by,omitempty"`
	RequestedAt     time.Time               `json:"requested_at"`
	CompletedAt     *time.Time              `json:"completed_at,omitempty"`
	CreatedAt       time.Time               `json:"created_at"`
	UpdatedAt       time.Time               `json:"updated_at"`
	Items           []RefundItemResponse    `json:"items,omitempty"`
	StatusHistory   []StatusHistoryResponse `json:"status_history,omitempty"`
}

// RefundItemResponse represents a refund item in response
type RefundItemResponse struct {
	ID              int        `json:"id"`
	RefundID        int        `json:"refund_id"`
	OrderItemID     int        `json:"order_item_id"`
	ProductID       int        `json:"product_id"`
	ProductName     string     `json:"product_name"`
	Quantity        int        `json:"quantity"`
	PricePerUnit    float64    `json:"price_per_unit"`
	RefundAmount    float64    `json:"refund_amount"`
	ItemReason      string     `json:"item_reason,omitempty"`
	StockRestored   bool       `json:"stock_restored"`
	StockRestoredAt *time.Time `json:"stock_restored_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// StatusHistoryResponse represents a refund status change in API response
type StatusHistoryResponse struct {
	ID        int       `json:"id"`
	RefundID  int       `json:"refund_id"`
	OldStatus *string   `json:"old_status,omitempty"`
	NewStatus string    `json:"new_status"`
	Actor     string    `json:"actor"`
	Reason    string    `json:"reason,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// RefundListResponse represents a paginated list of refunds
type RefundListResponse struct {
	Refunds    []RefundResponse `json:"refunds"`
	TotalCount int              `json:"total_count"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
}

// ProcessRefundRequest represents the request to process a pending refund
type ProcessRefundRequest struct {
	ProcessedBy int `json:"processed_by" binding:"required"`
}

// RetryRefundRequest represents the request to retry a failed refund
type RetryRefundRequest struct {
	ProcessedBy int `json:"processed_by" binding:"required"`
}

// RefundSuccessResponse represents a successful refund operation response
type RefundSuccessResponse struct {
	Success         bool    `json:"success"`
	Message         string  `json:"message"`
	RefundCode      string  `json:"refund_code"`
	GatewayRefundID *string `json:"gateway_refund_id,omitempty"`
}

// RefundErrorResponse represents a refund error response
type RefundErrorResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// CustomerRefundResponse represents refund information for customers
type CustomerRefundResponse struct {
	RefundCode     string               `json:"refund_code"`
	OrderCode      string               `json:"order_code"`
	RefundType     string               `json:"refund_type"`
	RefundAmount   float64              `json:"refund_amount"`
	ShippingRefund float64              `json:"shipping_refund"`
	ItemsRefund    float64              `json:"items_refund"`
	Status         string               `json:"status"`
	StatusLabel    string               `json:"status_label"`
	Timeline       string               `json:"timeline"`
	RequestedAt    time.Time            `json:"requested_at"`
	ProcessedAt    *time.Time           `json:"processed_at,omitempty"`
	CompletedAt    *time.Time           `json:"completed_at,omitempty"`
	Items          []RefundItemResponse `json:"items,omitempty"`
}

// CustomerRefundListResponse represents a list of refunds for a customer
type CustomerRefundListResponse struct {
	Refunds []CustomerRefundResponse `json:"refunds"`
	Count   int                      `json:"count"`
}

// ============================================
// ADMIN FORCE ACTION DTOs
// ============================================

// ForceCancelRequest represents a force cancel request
type ForceCancelRequest struct {
	Reason         string `json:"reason" binding:"required"`
	RestoreStock   bool   `json:"restore_stock"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

// ForceRefundRequest represents a force refund request
type ForceRefundRequest struct {
	RefundType     string              `json:"refund_type" binding:"required,oneof=FULL PARTIAL SHIPPING_ONLY"`
	Reason         string              `json:"reason" binding:"required"`
	Amount         *float64            `json:"amount,omitempty"`
	Items          []RefundItemRequest `json:"items,omitempty"`
	SkipGateway    bool                `json:"skip_gateway"`    // For manual reconciliation
	IdempotencyKey string              `json:"idempotency_key,omitempty"`
}

// ForceReshipRequest represents a force reship request
type ForceReshipRequest struct {
	Reason         string `json:"reason" binding:"required"`
	NewTrackingNo  string `json:"new_tracking_no,omitempty"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

// ReconcilePaymentRequest represents a payment reconciliation request
type ReconcilePaymentRequest struct {
	Action         string `json:"action" binding:"required,oneof=MARK_PAID MARK_FAILED MARK_EXPIRED SYNC_GATEWAY"`
	Reason         string `json:"reason" binding:"required"`
	TransactionID  string `json:"transaction_id,omitempty"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

// AdminActionResponse represents a generic admin action response
type AdminActionResponse struct {
	Success     bool           `json:"success"`
	Message     string         `json:"message"`
	AuditLogID  int            `json:"audit_log_id"`
	StateBefore map[string]any `json:"state_before,omitempty"`
	StateAfter  map[string]any `json:"state_after,omitempty"`
}

// ============================================
// PAYMENT SYNC DTOs
// ============================================

// PaymentSyncRequest represents a payment sync request
type PaymentSyncRequest struct {
	PaymentID int    `json:"payment_id" binding:"required"`
	SyncType  string `json:"sync_type" binding:"required,oneof=manual recovery"`
}

// PaymentSyncResponse represents a payment sync response
type PaymentSyncResponse struct {
	PaymentID          int            `json:"payment_id"`
	OrderCode          string         `json:"order_code"`
	LocalStatus        string         `json:"local_status"`
	GatewayStatus      string         `json:"gateway_status"`
	HasMismatch        bool           `json:"has_mismatch"`
	MismatchType       string         `json:"mismatch_type,omitempty"`
	Resolved           bool           `json:"resolved"`
	ResolutionAction   string         `json:"resolution_action,omitempty"`
	GatewayResponse    map[string]any `json:"gateway_response,omitempty"`
}

// StuckPaymentResponse represents a stuck payment
type StuckPaymentResponse struct {
	PaymentID     int       `json:"payment_id"`
	OrderID       int       `json:"order_id"`
	OrderCode     string    `json:"order_code"`
	Amount        float64   `json:"amount"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	HoursStuck    float64   `json:"hours_stuck"`
	RetryCount    int       `json:"retry_count"`
	LastSyncedAt  *time.Time `json:"last_synced_at,omitempty"`
}

// ============================================
// RECONCILIATION DTOs
// ============================================

// ReconciliationRequest represents a reconciliation request
type ReconciliationRequest struct {
	Date string `json:"date" binding:"required"` // YYYY-MM-DD format
}

// ReconciliationSummary represents reconciliation summary
type ReconciliationSummary struct {
	Date               string                 `json:"date"`
	TotalOrders        int                    `json:"total_orders"`
	TotalPayments      int                    `json:"total_payments"`
	TotalAmount        float64                `json:"total_amount"`
	OrdersByStatus     map[string]int         `json:"orders_by_status"`
	PaymentsByStatus   map[string]int         `json:"payments_by_status"`
	MismatchesFound    int                    `json:"mismatches_found"`
	MismatchesResolved int                    `json:"mismatches_resolved"`
	OrphanOrders       int                    `json:"orphan_orders"`
	OrphanPayments     int                    `json:"orphan_payments"`
	StuckPayments      int                    `json:"stuck_payments"`
	ExpectedRevenue    float64                `json:"expected_revenue"`
	ActualRevenue      float64                `json:"actual_revenue"`
	RevenueVariance    float64                `json:"revenue_variance"`
	TotalRefunds       float64                `json:"total_refunds"`
	Status             string                 `json:"status"`
	Mismatches         []MismatchDetail       `json:"mismatches,omitempty"`
	Orphans            []OrphanDetail         `json:"orphans,omitempty"`
}

// MismatchDetail represents a mismatch detail
type MismatchDetail struct {
	OrderCode     string `json:"order_code"`
	PaymentID     int    `json:"payment_id"`
	LocalStatus   string `json:"local_status"`
	GatewayStatus string `json:"gateway_status"`
	MismatchType  string `json:"mismatch_type"`
	Amount        float64 `json:"amount"`
}

// OrphanDetail represents an orphan record detail
type OrphanDetail struct {
	Type      string  `json:"type"` // "order" or "payment"
	ID        int     `json:"id"`
	Code      string  `json:"code,omitempty"`
	Amount    float64 `json:"amount"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"created_at"`
}

// ============================================
// MIDTRANS REFUND DTOs
// ============================================

// MidtransRefundRequest represents Midtrans refund API request
type MidtransRefundRequest struct {
	RefundKey string  `json:"refund_key"`
	Amount    float64 `json:"amount"`
	Reason    string  `json:"reason"`
}

// MidtransRefundResponse represents Midtrans refund API response
type MidtransRefundResponse struct {
	StatusCode           string `json:"status_code"`
	StatusMessage        string `json:"status_message"`
	RefundChargebackID   int    `json:"refund_chargeback_id,omitempty"`
	RefundAmount         string `json:"refund_amount,omitempty"`
	RefundKey            string `json:"refund_key,omitempty"`
	TransactionID        string `json:"transaction_id,omitempty"`
	GrossAmount          string `json:"gross_amount,omitempty"`
	Currency             string `json:"currency,omitempty"`
	OrderID              string `json:"order_id,omitempty"`
	PaymentType          string `json:"payment_type,omitempty"`
	TransactionTime      string `json:"transaction_time,omitempty"`
	TransactionStatus    string `json:"transaction_status,omitempty"`
	FraudStatus          string `json:"fraud_status,omitempty"`
	RefundChargebackTime string `json:"refund_chargeback_time,omitempty"`
	Bank                 string `json:"bank,omitempty"`
}

// MidtransStatusResponse represents Midtrans status check response
type MidtransStatusResponse struct {
	StatusCode        string `json:"status_code"`
	StatusMessage     string `json:"status_message"`
	TransactionID     string `json:"transaction_id"`
	OrderID           string `json:"order_id"`
	GrossAmount       string `json:"gross_amount"`
	PaymentType       string `json:"payment_type"`
	TransactionTime   string `json:"transaction_time"`
	TransactionStatus string `json:"transaction_status"`
	FraudStatus       string `json:"fraud_status"`
	SettlementTime    string `json:"settlement_time,omitempty"`
}
