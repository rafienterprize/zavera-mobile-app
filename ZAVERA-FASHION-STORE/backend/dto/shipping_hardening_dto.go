package dto

import "time"

// ============================================
// SHIPMENT CONTROL DTOs
// ============================================

// SchedulePickupRequest represents a pickup scheduling request
type SchedulePickupRequest struct {
	PickupDate     string `json:"pickup_date" binding:"required"` // YYYY-MM-DD
	PickupTimeSlot string `json:"pickup_time_slot,omitempty"`     // e.g., "09:00-12:00"
	Notes          string `json:"notes,omitempty"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

// MarkShippedRequest represents a mark shipped request
type MarkShippedRequest struct {
	TrackingNumber string `json:"tracking_number" binding:"required"`
	Notes          string `json:"notes,omitempty"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

// UpdateShipmentStatusRequest represents a status update request
type UpdateShipmentStatusRequest struct {
	Status         string `json:"status" binding:"required"`
	Reason         string `json:"reason" binding:"required"`
	Notes          string `json:"notes,omitempty"`
	IsAdmin        bool   `json:"is_admin,omitempty"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

// InvestigateShipmentRequest represents an investigation request
type InvestigateShipmentRequest struct {
	Reason         string `json:"reason" binding:"required"`
	Notes          string `json:"notes,omitempty"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

// MarkLostRequest represents a mark lost request
type MarkLostRequest struct {
	Reason           string `json:"reason" binding:"required"`
	CreateDispute    bool   `json:"create_dispute"`
	AutoReship       bool   `json:"auto_reship"`
	IdempotencyKey   string `json:"idempotency_key,omitempty"`
}

// ReshipRequest represents a reship request
type ReshipRequest struct {
	Reason           string  `json:"reason" binding:"required"`
	CostBearer       string  `json:"cost_bearer" binding:"required,oneof=company customer"`
	NewTrackingNo    string  `json:"new_tracking_no,omitempty"`
	UseNewAddress    bool    `json:"use_new_address"`
	NewAddressID     *int    `json:"new_address_id,omitempty"`
	IdempotencyKey   string  `json:"idempotency_key,omitempty"`
}

// OverrideStatusRequest represents an admin status override
type OverrideStatusRequest struct {
	NewStatus      string `json:"new_status" binding:"required"`
	Reason         string `json:"reason" binding:"required"`
	BypassValidation bool `json:"bypass_validation"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

// ============================================
// DISPUTE DTOs
// ============================================

// CreateDisputeRequest represents a dispute creation request
type CreateDisputeRequest struct {
	OrderCode       string   `json:"order_code" binding:"required"`
	ShipmentID      *int     `json:"shipment_id,omitempty"`
	DisputeType     string   `json:"dispute_type" binding:"required"`
	Title           string   `json:"title" binding:"required"`
	Description     string   `json:"description" binding:"required"`
	CustomerClaim   string   `json:"customer_claim,omitempty"`
	EvidenceURLs    []string `json:"evidence_urls,omitempty"`
	IdempotencyKey  string   `json:"idempotency_key,omitempty"`
}

// UpdateDisputeRequest represents a dispute update request
type UpdateDisputeRequest struct {
	Status              string   `json:"status,omitempty"`
	InvestigationNotes  string   `json:"investigation_notes,omitempty"`
	EvidenceURLs        []string `json:"evidence_urls,omitempty"`
	IdempotencyKey      string   `json:"idempotency_key,omitempty"`
}

// ResolveDisputeRequest represents a dispute resolution request
type ResolveDisputeRequest struct {
	Resolution       string   `json:"resolution" binding:"required,oneof=RESOLVED_REFUND RESOLVED_RESHIP RESOLVED_REJECTED"`
	ResolutionNotes  string   `json:"resolution_notes" binding:"required"`
	ResolutionAmount *float64 `json:"resolution_amount,omitempty"`
	CreateRefund     bool     `json:"create_refund"`
	CreateReship     bool     `json:"create_reship"`
	IdempotencyKey   string   `json:"idempotency_key,omitempty"`
}

// AddDisputeMessageRequest represents adding a message to dispute
type AddDisputeMessageRequest struct {
	Message        string   `json:"message" binding:"required"`
	AttachmentURLs []string `json:"attachment_urls,omitempty"`
	IsInternal     bool     `json:"is_internal"`
}

// DisputeResponse represents a dispute in API response
type DisputeResponse struct {
	ID                     int                      `json:"id"`
	DisputeCode            string                   `json:"dispute_code"`
	OrderID                int                      `json:"order_id"`
	OrderCode              string                   `json:"order_code,omitempty"`
	ShipmentID             *int                     `json:"shipment_id,omitempty"`
	DisputeType            string                   `json:"dispute_type"`
	Status                 string                   `json:"status"`
	Title                  string                   `json:"title"`
	Description            string                   `json:"description"`
	CustomerClaim          string                   `json:"customer_claim,omitempty"`
	CustomerEmail          string                   `json:"customer_email"`
	EvidenceURLs           []string                 `json:"evidence_urls,omitempty"`
	InvestigationNotes     string                   `json:"investigation_notes,omitempty"`
	Resolution             string                   `json:"resolution,omitempty"`
	ResolutionNotes        string                   `json:"resolution_notes,omitempty"`
	ResolutionAmount       *float64                 `json:"resolution_amount,omitempty"`
	ReshipShipmentID       *int                     `json:"reship_shipment_id,omitempty"`
	ResponseDeadline       *string                  `json:"response_deadline,omitempty"`
	ResolutionDeadline     *string                  `json:"resolution_deadline,omitempty"`
	CreatedAt              string                   `json:"created_at"`
	ResolvedAt             *string                  `json:"resolved_at,omitempty"`
	Messages               []DisputeMessageResponse `json:"messages,omitempty"`
}

// DisputeMessageResponse represents a dispute message in response
type DisputeMessageResponse struct {
	ID             int      `json:"id"`
	SenderType     string   `json:"sender_type"`
	SenderName     string   `json:"sender_name,omitempty"`
	Message        string   `json:"message"`
	AttachmentURLs []string `json:"attachment_urls,omitempty"`
	IsInternal     bool     `json:"is_internal"`
	CreatedAt      string   `json:"created_at"`
}

// ============================================
// SHIPMENT MONITORING DTOs
// ============================================

// StuckShipmentResponse represents a stuck shipment
type StuckShipmentResponse struct {
	ShipmentID        int       `json:"shipment_id"`
	OrderID           int       `json:"order_id"`
	OrderCode         string    `json:"order_code"`
	TrackingNumber    string    `json:"tracking_number"`
	Status            string    `json:"status"`
	ProviderCode      string    `json:"provider_code"`
	DaysWithoutUpdate int       `json:"days_without_update"`
	LastTrackingUpdate *time.Time `json:"last_tracking_update,omitempty"`
	ShippedAt         *time.Time `json:"shipped_at,omitempty"`
	AlertLevel        string    `json:"alert_level"`
}

// PickupFailureResponse represents a pickup failure
type PickupFailureResponse struct {
	ShipmentID      int       `json:"shipment_id"`
	OrderID         int       `json:"order_id"`
	OrderCode       string    `json:"order_code"`
	PickupAttempts  int       `json:"pickup_attempts"`
	LastAttemptAt   *time.Time `json:"last_attempt_at,omitempty"`
	ScheduledAt     *time.Time `json:"scheduled_at,omitempty"`
	Deadline        *time.Time `json:"deadline,omitempty"`
	RequiresAdmin   bool      `json:"requires_admin"`
}

// ShipmentAlertResponse represents a shipment alert
type ShipmentAlertResponse struct {
	ID              int       `json:"id"`
	ShipmentID      int       `json:"shipment_id"`
	OrderCode       string    `json:"order_code"`
	AlertType       string    `json:"alert_type"`
	AlertLevel      string    `json:"alert_level"`
	Title           string    `json:"title"`
	Description     string    `json:"description,omitempty"`
	Acknowledged    bool      `json:"acknowledged"`
	Resolved        bool      `json:"resolved"`
	CreatedAt       string    `json:"created_at"`
}

// ShipmentStatusHistoryResponse represents status history
type ShipmentStatusHistoryResponse struct {
	FromStatus string `json:"from_status,omitempty"`
	ToStatus   string `json:"to_status"`
	ChangedBy  string `json:"changed_by,omitempty"`
	Reason     string `json:"reason,omitempty"`
	CreatedAt  string `json:"created_at"`
}

// EnhancedShipmentResponse represents detailed shipment info
type EnhancedShipmentResponse struct {
	ID                  int                             `json:"id"`
	OrderID             int                             `json:"order_id"`
	OrderCode           string                          `json:"order_code"`
	ProviderCode        string                          `json:"provider_code"`
	ProviderName        string                          `json:"provider_name"`
	ServiceName         string                          `json:"service_name"`
	TrackingNumber      string                          `json:"tracking_number"`
	Status              string                          `json:"status"`
	Cost                float64                         `json:"cost"`
	Weight              int                             `json:"weight"`
	Origin              string                          `json:"origin"`
	Destination         string                          `json:"destination"`
	
	// Timestamps
	CreatedAt           string                          `json:"created_at"`
	ShippedAt           *string                         `json:"shipped_at,omitempty"`
	DeliveredAt         *string                         `json:"delivered_at,omitempty"`
	
	// Pickup info
	PickupScheduledAt   *string                         `json:"pickup_scheduled_at,omitempty"`
	PickupAttempts      int                             `json:"pickup_attempts"`
	
	// Tracking info
	LastTrackingUpdate  *string                         `json:"last_tracking_update,omitempty"`
	DaysWithoutUpdate   int                             `json:"days_without_update"`
	TrackingStale       bool                            `json:"tracking_stale"`
	
	// Problem indicators
	IsProblematic       bool                            `json:"is_problematic"`
	RequiresAdminAction bool                            `json:"requires_admin_action"`
	AdminActionReason   string                          `json:"admin_action_reason,omitempty"`
	
	// Reship info
	IsReplacement       bool                            `json:"is_replacement"`
	OriginalShipmentID  *int                            `json:"original_shipment_id,omitempty"`
	ReplacedByID        *int                            `json:"replaced_by_id,omitempty"`
	
	// History
	StatusHistory       []ShipmentStatusHistoryResponse `json:"status_history,omitempty"`
	TrackingHistory     []TrackingEventResponse         `json:"tracking_history,omitempty"`
	Alerts              []ShipmentAlertResponse         `json:"alerts,omitempty"`
}

// CourierFailureResponse represents a courier failure
type CourierFailureResponse struct {
	ID               int       `json:"id"`
	ShipmentID       int       `json:"shipment_id"`
	FailureType      string    `json:"failure_type"`
	FailureReason    string    `json:"failure_reason,omitempty"`
	FailureTime      string    `json:"failure_time"`
	CourierCode      string    `json:"courier_code,omitempty"`
	FailureLocation  string    `json:"failure_location,omitempty"`
	Resolved         bool      `json:"resolved"`
	ResolutionAction string    `json:"resolution_action,omitempty"`
}

// ============================================
// FULFILLMENT DASHBOARD DTOs
// ============================================

// FulfillmentDashboardResponse represents fulfillment overview
type FulfillmentDashboardResponse struct {
	// Counts by status
	StatusCounts        map[string]int `json:"status_counts"`
	
	// Problem counts
	StuckShipments      int            `json:"stuck_shipments"`
	PickupFailures      int            `json:"pickup_failures"`
	DeliveryFailures    int            `json:"delivery_failures"`
	LostPackages        int            `json:"lost_packages"`
	UnderInvestigation  int            `json:"under_investigation"`
	
	// Alerts
	UnresolvedAlerts    int            `json:"unresolved_alerts"`
	CriticalAlerts      int            `json:"critical_alerts"`
	
	// Disputes
	OpenDisputes        int            `json:"open_disputes"`
	PendingResolution   int            `json:"pending_resolution"`
	
	// Performance
	AvgDeliveryDays     float64        `json:"avg_delivery_days"`
	OnTimeDeliveryRate  float64        `json:"on_time_delivery_rate"`
	
	// Recent issues
	RecentAlerts        []ShipmentAlertResponse `json:"recent_alerts,omitempty"`
}


// ShipmentListItem represents a shipment in list view
type ShipmentListItem struct {
	ID                int    `json:"id"`
	OrderCode         string `json:"order_code"`
	TrackingNumber    string `json:"tracking_number"`
	Status            string `json:"status"`
	ProviderCode      string `json:"provider_code"`
	ProviderName      string `json:"provider_name"`
	DaysWithoutUpdate int    `json:"days_without_update"`
	CreatedAt         string `json:"created_at"`
}
