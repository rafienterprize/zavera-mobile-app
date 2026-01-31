package models

import "time"

// DisputeStatus represents the status of a dispute
type DisputeStatus string

const (
	DisputeStatusOpen              DisputeStatus = "OPEN"
	DisputeStatusInvestigating     DisputeStatus = "INVESTIGATING"
	DisputeStatusEvidenceRequired  DisputeStatus = "EVIDENCE_REQUIRED"
	DisputeStatusPendingResolution DisputeStatus = "PENDING_RESOLUTION"
	DisputeStatusResolvedRefund    DisputeStatus = "RESOLVED_REFUND"
	DisputeStatusResolvedReship    DisputeStatus = "RESOLVED_RESHIP"
	DisputeStatusResolvedRejected  DisputeStatus = "RESOLVED_REJECTED"
	DisputeStatusClosed            DisputeStatus = "CLOSED"
)

// DisputeType represents the type of dispute
type DisputeType string

const (
	DisputeTypeLostPackage    DisputeType = "LOST_PACKAGE"
	DisputeTypeDamagedPackage DisputeType = "DAMAGED_PACKAGE"
	DisputeTypeWrongItem      DisputeType = "WRONG_ITEM"
	DisputeTypeMissingItem    DisputeType = "MISSING_ITEM"
	DisputeTypeNotDelivered   DisputeType = "NOT_DELIVERED"
	DisputeTypeLateDelivery   DisputeType = "LATE_DELIVERY"
	DisputeTypeFakeDelivery   DisputeType = "FAKE_DELIVERY"
	DisputeTypeOther          DisputeType = "OTHER"
)

// IsFinalStatus checks if dispute is in terminal state
func (s DisputeStatus) IsFinalStatus() bool {
	switch s {
	case DisputeStatusResolvedRefund, DisputeStatusResolvedReship,
		DisputeStatusResolvedRejected, DisputeStatusClosed:
		return true
	}
	return false
}

// IsResolved checks if dispute is resolved
func (s DisputeStatus) IsResolved() bool {
	switch s {
	case DisputeStatusResolvedRefund, DisputeStatusResolvedReship, DisputeStatusResolvedRejected:
		return true
	}
	return false
}

// Dispute represents a customer dispute
type Dispute struct {
	ID                       int            `json:"id" db:"id"`
	DisputeCode              string         `json:"dispute_code" db:"dispute_code"`
	OrderID                  int            `json:"order_id" db:"order_id"`
	ShipmentID               *int           `json:"shipment_id,omitempty" db:"shipment_id"`
	RefundID                 *int           `json:"refund_id,omitempty" db:"refund_id"`
	DisputeType              DisputeType    `json:"dispute_type" db:"dispute_type"`
	Status                   DisputeStatus  `json:"status" db:"status"`
	Title                    string         `json:"title" db:"title"`
	Description              string         `json:"description" db:"description"`
	CustomerClaim            string         `json:"customer_claim,omitempty" db:"customer_claim"`
	CustomerUserID           *int           `json:"customer_user_id,omitempty" db:"customer_user_id"`
	CustomerEmail            string         `json:"customer_email" db:"customer_email"`
	CustomerPhone            string         `json:"customer_phone,omitempty" db:"customer_phone"`
	EvidenceURLs             []string       `json:"evidence_urls,omitempty" db:"evidence_urls"`
	CustomerEvidenceURLs     []string       `json:"customer_evidence_urls,omitempty" db:"customer_evidence_urls"`
	CourierEvidenceURLs      []string       `json:"courier_evidence_urls,omitempty" db:"courier_evidence_urls"`
	InvestigationNotes       string         `json:"investigation_notes,omitempty" db:"investigation_notes"`
	InvestigationStartedAt   *time.Time     `json:"investigation_started_at,omitempty" db:"investigation_started_at"`
	InvestigationCompletedAt *time.Time     `json:"investigation_completed_at,omitempty" db:"investigation_completed_at"`
	InvestigatorID           *int           `json:"investigator_id,omitempty" db:"investigator_id"`
	Resolution               DisputeStatus  `json:"resolution,omitempty" db:"resolution"`
	ResolutionNotes          string         `json:"resolution_notes,omitempty" db:"resolution_notes"`
	ResolutionAmount         *float64       `json:"resolution_amount,omitempty" db:"resolution_amount"`
	ResolvedBy               *int           `json:"resolved_by,omitempty" db:"resolved_by"`
	ResolvedAt               *time.Time     `json:"resolved_at,omitempty" db:"resolved_at"`
	ReshipShipmentID         *int           `json:"reship_shipment_id,omitempty" db:"reship_shipment_id"`
	ResponseDeadline         *time.Time     `json:"response_deadline,omitempty" db:"response_deadline"`
	ResolutionDeadline       *time.Time     `json:"resolution_deadline,omitempty" db:"resolution_deadline"`
	Metadata                 map[string]any `json:"metadata,omitempty" db:"metadata"`
	CreatedAt                time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt                time.Time      `json:"updated_at" db:"updated_at"`
	Messages                 []DisputeMessage `json:"messages,omitempty" db:"-"`
}

// DisputeMessage represents a message in a dispute thread
type DisputeMessage struct {
	ID             int       `json:"id" db:"id"`
	DisputeID      int       `json:"dispute_id" db:"dispute_id"`
	SenderType     string    `json:"sender_type" db:"sender_type"` // customer, admin, system
	SenderID       *int      `json:"sender_id,omitempty" db:"sender_id"`
	SenderName     string    `json:"sender_name,omitempty" db:"sender_name"`
	Message        string    `json:"message" db:"message"`
	AttachmentURLs []string  `json:"attachment_urls,omitempty" db:"attachment_urls"`
	IsInternal     bool      `json:"is_internal" db:"is_internal"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// CourierFailureLog represents a courier failure record
type CourierFailureLog struct {
	ID               int        `json:"id" db:"id"`
	ShipmentID       int        `json:"shipment_id" db:"shipment_id"`
	FailureType      string     `json:"failure_type" db:"failure_type"`
	FailureReason    string     `json:"failure_reason,omitempty" db:"failure_reason"`
	FailureTime      time.Time  `json:"failure_time" db:"failure_time"`
	CourierCode      string     `json:"courier_code,omitempty" db:"courier_code"`
	CourierName      string     `json:"courier_name,omitempty" db:"courier_name"`
	CourierTracking  string     `json:"courier_tracking,omitempty" db:"courier_tracking"`
	FailureLocation  string     `json:"failure_location,omitempty" db:"failure_location"`
	Resolved         bool       `json:"resolved" db:"resolved"`
	ResolvedAt       *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
	ResolvedBy       string     `json:"resolved_by,omitempty" db:"resolved_by"`
	ResolutionAction string     `json:"resolution_action,omitempty" db:"resolution_action"`
	EvidenceURLs     []string   `json:"evidence_urls,omitempty" db:"evidence_urls"`
	Notes            string     `json:"notes,omitempty" db:"notes"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
}

// ShipmentStatusHistory represents a shipment status change record
type ShipmentStatusHistory struct {
	ID         int            `json:"id" db:"id"`
	ShipmentID int            `json:"shipment_id" db:"shipment_id"`
	FromStatus string         `json:"from_status,omitempty" db:"from_status"`
	ToStatus   string         `json:"to_status" db:"to_status"`
	ChangedBy  string         `json:"changed_by,omitempty" db:"changed_by"`
	Reason     string         `json:"reason,omitempty" db:"reason"`
	Metadata   map[string]any `json:"metadata,omitempty" db:"metadata"`
	CreatedAt  time.Time      `json:"created_at" db:"created_at"`
}

// ShipmentAlert represents an alert for a shipment issue
type ShipmentAlert struct {
	ID              int        `json:"id" db:"id"`
	ShipmentID      int        `json:"shipment_id" db:"shipment_id"`
	AlertType       string     `json:"alert_type" db:"alert_type"`
	AlertLevel      string     `json:"alert_level" db:"alert_level"`
	Title           string     `json:"title" db:"title"`
	Description     string     `json:"description,omitempty" db:"description"`
	Acknowledged    bool       `json:"acknowledged" db:"acknowledged"`
	AcknowledgedBy  *int       `json:"acknowledged_by,omitempty" db:"acknowledged_by"`
	AcknowledgedAt  *time.Time `json:"acknowledged_at,omitempty" db:"acknowledged_at"`
	Resolved        bool       `json:"resolved" db:"resolved"`
	ResolvedBy      *int       `json:"resolved_by,omitempty" db:"resolved_by"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
	ResolutionNotes string     `json:"resolution_notes,omitempty" db:"resolution_notes"`
	AutoActionTaken bool       `json:"auto_action_taken" db:"auto_action_taken"`
	AutoActionType  string     `json:"auto_action_type,omitempty" db:"auto_action_type"`
	AutoActionAt    *time.Time `json:"auto_action_at,omitempty" db:"auto_action_at"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
}
