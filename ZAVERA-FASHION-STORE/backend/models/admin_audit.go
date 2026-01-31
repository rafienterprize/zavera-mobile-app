package models

import "time"

// AdminActionType represents types of admin actions
type AdminActionType string

const (
	AdminActionForceCancel      AdminActionType = "FORCE_CANCEL"
	AdminActionForceRefund      AdminActionType = "FORCE_REFUND"
	AdminActionForceReship      AdminActionType = "FORCE_RESHIP"
	AdminActionReconcilePayment AdminActionType = "RECONCILE_PAYMENT"
	AdminActionUpdateStatus     AdminActionType = "UPDATE_STATUS"
	AdminActionRestoreStock     AdminActionType = "RESTORE_STOCK"
	AdminActionVoidRefund       AdminActionType = "VOID_REFUND"
	AdminActionOverridePayment  AdminActionType = "OVERRIDE_PAYMENT"
	AdminActionManualAdjustment AdminActionType = "MANUAL_ADJUSTMENT"
)

// AdminAuditLog represents an immutable admin action log entry
type AdminAuditLog struct {
	ID             int            `json:"id" db:"id"`
	AdminUserID    *int           `json:"admin_user_id,omitempty" db:"admin_user_id"`
	AdminEmail     string         `json:"admin_email" db:"admin_email"`
	AdminIP        string         `json:"admin_ip,omitempty" db:"admin_ip"`
	AdminUserAgent string         `json:"admin_user_agent,omitempty" db:"admin_user_agent"`
	ActionType     AdminActionType `json:"action_type" db:"action_type"`
	ActionDetail   string         `json:"action_detail" db:"action_detail"`
	TargetType     string         `json:"target_type" db:"target_type"`
	TargetID       int            `json:"target_id" db:"target_id"`
	TargetCode     string         `json:"target_code,omitempty" db:"target_code"`
	StateBefore    map[string]any `json:"state_before,omitempty" db:"state_before"`
	StateAfter     map[string]any `json:"state_after,omitempty" db:"state_after"`
	Success        bool           `json:"success" db:"success"`
	ErrorMessage   string         `json:"error_message,omitempty" db:"error_message"`
	IdempotencyKey string         `json:"idempotency_key,omitempty" db:"idempotency_key"`
	Metadata       map[string]any `json:"metadata,omitempty" db:"metadata"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
}

// PaymentSyncStatus represents payment sync status
type PaymentSyncStatus string

const (
	PaymentSyncStatusPending    PaymentSyncStatus = "PENDING"
	PaymentSyncStatusInProgress PaymentSyncStatus = "IN_PROGRESS"
	PaymentSyncStatusSynced     PaymentSyncStatus = "SYNCED"
	PaymentSyncStatusMismatch   PaymentSyncStatus = "MISMATCH"
	PaymentSyncStatusFailed     PaymentSyncStatus = "FAILED"
	PaymentSyncStatusResolved   PaymentSyncStatus = "RESOLVED"
)

// PaymentSyncLog represents a payment sync record
type PaymentSyncLog struct {
	ID                   int               `json:"id" db:"id"`
	PaymentID            int               `json:"payment_id" db:"payment_id"`
	OrderID              int               `json:"order_id" db:"order_id"`
	OrderCode            string            `json:"order_code" db:"order_code"`
	SyncType             string            `json:"sync_type" db:"sync_type"`
	SyncStatus           PaymentSyncStatus `json:"sync_status" db:"sync_status"`
	LocalPaymentStatus   string            `json:"local_payment_status,omitempty" db:"local_payment_status"`
	LocalOrderStatus     string            `json:"local_order_status,omitempty" db:"local_order_status"`
	GatewayStatus        string            `json:"gateway_status,omitempty" db:"gateway_status"`
	GatewayTransactionID string            `json:"gateway_transaction_id,omitempty" db:"gateway_transaction_id"`
	GatewayResponse      map[string]any    `json:"gateway_response,omitempty" db:"gateway_response"`
	HasMismatch          bool              `json:"has_mismatch" db:"has_mismatch"`
	MismatchType         string            `json:"mismatch_type,omitempty" db:"mismatch_type"`
	MismatchDetail       string            `json:"mismatch_detail,omitempty" db:"mismatch_detail"`
	Resolved             bool              `json:"resolved" db:"resolved"`
	ResolvedBy           *int              `json:"resolved_by,omitempty" db:"resolved_by"`
	ResolvedAt           *time.Time        `json:"resolved_at,omitempty" db:"resolved_at"`
	ResolutionAction     string            `json:"resolution_action,omitempty" db:"resolution_action"`
	RetryCount           int               `json:"retry_count" db:"retry_count"`
	LastRetryAt          *time.Time        `json:"last_retry_at,omitempty" db:"last_retry_at"`
	NextRetryAt          *time.Time        `json:"next_retry_at,omitempty" db:"next_retry_at"`
	CreatedAt            time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time         `json:"updated_at" db:"updated_at"`
}

// ReconciliationLog represents a daily reconciliation record
type ReconciliationLog struct {
	ID                  int            `json:"id" db:"id"`
	ReconciliationDate  time.Time      `json:"reconciliation_date" db:"reconciliation_date"`
	PeriodStart         time.Time      `json:"period_start" db:"period_start"`
	PeriodEnd           time.Time      `json:"period_end" db:"period_end"`
	TotalOrders         int            `json:"total_orders" db:"total_orders"`
	TotalPayments       int            `json:"total_payments" db:"total_payments"`
	TotalAmount         float64        `json:"total_amount" db:"total_amount"`
	OrdersPending       int            `json:"orders_pending" db:"orders_pending"`
	OrdersPaid          int            `json:"orders_paid" db:"orders_paid"`
	OrdersCancelled     int            `json:"orders_cancelled" db:"orders_cancelled"`
	OrdersRefunded      int            `json:"orders_refunded" db:"orders_refunded"`
	PaymentsPending     int            `json:"payments_pending" db:"payments_pending"`
	PaymentsSuccess     int            `json:"payments_success" db:"payments_success"`
	PaymentsFailed      int            `json:"payments_failed" db:"payments_failed"`
	MismatchesFound     int            `json:"mismatches_found" db:"mismatches_found"`
	MismatchesResolved  int            `json:"mismatches_resolved" db:"mismatches_resolved"`
	MismatchDetails     map[string]any `json:"mismatch_details,omitempty" db:"mismatch_details"`
	OrphanOrders        int            `json:"orphan_orders" db:"orphan_orders"`
	OrphanPayments      int            `json:"orphan_payments" db:"orphan_payments"`
	OrphanDetails       map[string]any `json:"orphan_details,omitempty" db:"orphan_details"`
	StuckPayments       int            `json:"stuck_payments" db:"stuck_payments"`
	StuckPaymentIDs     []int          `json:"stuck_payment_ids,omitempty" db:"stuck_payment_ids"`
	ExpectedRevenue     float64        `json:"expected_revenue" db:"expected_revenue"`
	ActualRevenue       float64        `json:"actual_revenue" db:"actual_revenue"`
	RevenueVariance     float64        `json:"revenue_variance" db:"revenue_variance"`
	TotalRefunds        float64        `json:"total_refunds" db:"total_refunds"`
	Status              string         `json:"status" db:"status"`
	StartedAt           *time.Time     `json:"started_at,omitempty" db:"started_at"`
	CompletedAt         *time.Time     `json:"completed_at,omitempty" db:"completed_at"`
	RunBy               string         `json:"run_by,omitempty" db:"run_by"`
	ErrorCount          int            `json:"error_count" db:"error_count"`
	Errors              map[string]any `json:"errors,omitempty" db:"errors"`
	CreatedAt           time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at" db:"updated_at"`
}
