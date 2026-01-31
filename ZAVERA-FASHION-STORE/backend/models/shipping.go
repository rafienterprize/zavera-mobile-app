package models

import "time"

// ShipmentStatus represents the status of a shipment (v2 - hardened)
type ShipmentStatus string

const (
	ShipmentStatusPending          ShipmentStatus = "PENDING"
	ShipmentStatusProcessing       ShipmentStatus = "PROCESSING"
	ShipmentStatusPickupScheduled  ShipmentStatus = "PICKUP_SCHEDULED"
	ShipmentStatusPickupFailed     ShipmentStatus = "PICKUP_FAILED"
	ShipmentStatusShipped          ShipmentStatus = "SHIPPED"
	ShipmentStatusInTransit        ShipmentStatus = "IN_TRANSIT"
	ShipmentStatusOutForDelivery   ShipmentStatus = "OUT_FOR_DELIVERY"
	ShipmentStatusDelivered        ShipmentStatus = "DELIVERED"
	ShipmentStatusDeliveryFailed   ShipmentStatus = "DELIVERY_FAILED"
	ShipmentStatusHeldAtWarehouse  ShipmentStatus = "HELD_AT_WAREHOUSE"
	ShipmentStatusReturnedToSender ShipmentStatus = "RETURNED_TO_SENDER"
	ShipmentStatusLost             ShipmentStatus = "LOST"
	ShipmentStatusInvestigation    ShipmentStatus = "INVESTIGATION"
	ShipmentStatusReplaced         ShipmentStatus = "REPLACED"
	ShipmentStatusCancelled        ShipmentStatus = "CANCELLED"
)

// ValidShipmentTransitions defines allowed status transitions
var ValidShipmentTransitions = map[ShipmentStatus][]ShipmentStatus{
	ShipmentStatusPending: {ShipmentStatusProcessing, ShipmentStatusCancelled},
	ShipmentStatusProcessing: {ShipmentStatusPickupScheduled, ShipmentStatusShipped, ShipmentStatusCancelled},
	ShipmentStatusPickupScheduled: {ShipmentStatusShipped, ShipmentStatusPickupFailed, ShipmentStatusCancelled},
	ShipmentStatusPickupFailed: {ShipmentStatusPickupScheduled, ShipmentStatusCancelled},
	ShipmentStatusShipped: {ShipmentStatusInTransit, ShipmentStatusDelivered, ShipmentStatusLost, ShipmentStatusInvestigation},
	ShipmentStatusInTransit: {ShipmentStatusOutForDelivery, ShipmentStatusDelivered, ShipmentStatusHeldAtWarehouse, ShipmentStatusReturnedToSender, ShipmentStatusLost, ShipmentStatusInvestigation},
	ShipmentStatusOutForDelivery: {ShipmentStatusDelivered, ShipmentStatusDeliveryFailed, ShipmentStatusHeldAtWarehouse},
	ShipmentStatusDeliveryFailed: {ShipmentStatusOutForDelivery, ShipmentStatusHeldAtWarehouse, ShipmentStatusReturnedToSender},
	ShipmentStatusHeldAtWarehouse: {ShipmentStatusOutForDelivery, ShipmentStatusReturnedToSender, ShipmentStatusDelivered},
	ShipmentStatusReturnedToSender: {ShipmentStatusReplaced, ShipmentStatusCancelled},
	ShipmentStatusInvestigation: {ShipmentStatusLost, ShipmentStatusDelivered, ShipmentStatusInTransit, ShipmentStatusReplaced},
	ShipmentStatusLost: {ShipmentStatusReplaced},
	ShipmentStatusDelivered: {ShipmentStatusInvestigation}, // For disputed deliveries
}

// AdminOnlyTransitions defines transitions that require admin
var AdminOnlyTransitions = map[ShipmentStatus][]ShipmentStatus{
	ShipmentStatusProcessing: {ShipmentStatusCancelled},
	ShipmentStatusPickupScheduled: {ShipmentStatusCancelled},
	ShipmentStatusPickupFailed: {ShipmentStatusCancelled},
	ShipmentStatusShipped: {ShipmentStatusLost, ShipmentStatusInvestigation},
	ShipmentStatusInTransit: {ShipmentStatusLost, ShipmentStatusInvestigation},
	ShipmentStatusReturnedToSender: {ShipmentStatusReplaced, ShipmentStatusCancelled},
	ShipmentStatusInvestigation: {ShipmentStatusLost, ShipmentStatusDelivered, ShipmentStatusInTransit, ShipmentStatusReplaced},
	ShipmentStatusLost: {ShipmentStatusReplaced},
	ShipmentStatusDelivered: {ShipmentStatusInvestigation},
	ShipmentStatusCancelled: {ShipmentStatusProcessing},
}

// IsValidTransition checks if a status transition is allowed
func (s ShipmentStatus) IsValidTransition(next ShipmentStatus) bool {
	allowed, exists := ValidShipmentTransitions[s]
	if !exists {
		return false
	}
	for _, status := range allowed {
		if status == next {
			return true
		}
	}
	return false
}

// RequiresAdmin checks if transition requires admin privileges
func (s ShipmentStatus) RequiresAdmin(next ShipmentStatus) bool {
	adminOnly, exists := AdminOnlyTransitions[s]
	if !exists {
		return false
	}
	for _, status := range adminOnly {
		if status == next {
			return true
		}
	}
	return false
}

// IsFinalStatus checks if the status is terminal
func (s ShipmentStatus) IsFinalStatus() bool {
	switch s {
	case ShipmentStatusDelivered, ShipmentStatusCancelled, ShipmentStatusReplaced:
		return true
	}
	return false
}

// IsProblematic checks if status indicates a problem
func (s ShipmentStatus) IsProblematic() bool {
	switch s {
	case ShipmentStatusPickupFailed, ShipmentStatusDeliveryFailed, 
		ShipmentStatusHeldAtWarehouse, ShipmentStatusReturnedToSender,
		ShipmentStatusLost, ShipmentStatusInvestigation:
		return true
	}
	return false
}

// RequiresAction checks if status requires attention
func (s ShipmentStatus) RequiresAction() bool {
	switch s {
	case ShipmentStatusPickupFailed, ShipmentStatusDeliveryFailed,
		ShipmentStatusLost, ShipmentStatusInvestigation:
		return true
	}
	return false
}

// ShippingProvider represents a courier company
type ShippingProvider struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Code      string    `json:"code" db:"code"`
	LogoURL   string    `json:"logo_url" db:"logo_url"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ShippingService represents a service type offered by a provider
type ShippingService struct {
	ID          int       `json:"id" db:"id"`
	ProviderID  int       `json:"provider_id" db:"provider_id"`
	ServiceCode string    `json:"service_code" db:"service_code"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	ETD         string    `json:"etd" db:"etd"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Shipment represents a shipment record for an order
type Shipment struct {
	ID                  int            `json:"id" db:"id"`
	OrderID             int            `json:"order_id" db:"order_id"`
	ProviderCode        string         `json:"provider_code" db:"provider_code"`
	ProviderName        string         `json:"provider_name" db:"provider_name"`
	ServiceCode         string         `json:"service_code" db:"service_code"`
	ServiceName         string         `json:"service_name" db:"service_name"`
	Cost                float64        `json:"cost" db:"cost"`
	ETD                 string         `json:"etd" db:"etd"`
	Weight              int            `json:"weight" db:"weight"`
	TrackingNumber      string         `json:"tracking_number" db:"tracking_number"`
	Status              ShipmentStatus `json:"status" db:"status"`
	OriginCityID        string         `json:"origin_city_id" db:"origin_city_id"`
	OriginCityName      string         `json:"origin_city_name" db:"origin_city_name"`
	DestinationCityID   string         `json:"destination_city_id" db:"destination_city_id"`
	DestinationCityName string         `json:"destination_city_name" db:"destination_city_name"`
	
	// Biteship integration fields
	BiteshipDraftOrderID string         `json:"biteship_draft_order_id,omitempty" db:"biteship_draft_order_id"`
	BiteshipOrderID      string         `json:"biteship_order_id,omitempty" db:"biteship_order_id"`
	BiteshipTrackingID   string         `json:"biteship_tracking_id,omitempty" db:"biteship_tracking_id"`
	BiteshipWaybillID    string         `json:"biteship_waybill_id,omitempty" db:"biteship_waybill_id"`
	
	// Timestamps
	ShippedAt           *time.Time     `json:"shipped_at,omitempty" db:"shipped_at"`
	DeliveredAt         *time.Time     `json:"delivered_at,omitempty" db:"delivered_at"`
	CreatedAt           time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at" db:"updated_at"`
	
	// Pickup control
	PickupScheduledAt   *time.Time     `json:"pickup_scheduled_at,omitempty" db:"pickup_scheduled_at"`
	PickupDeadline      *time.Time     `json:"pickup_deadline,omitempty" db:"pickup_deadline"`
	PickupAttempts      int            `json:"pickup_attempts" db:"pickup_attempts"`
	LastPickupAttemptAt *time.Time     `json:"last_pickup_attempt_at,omitempty" db:"last_pickup_attempt_at"`
	PickupNotes         string         `json:"pickup_notes,omitempty" db:"pickup_notes"`
	
	// Tracking control
	LastTrackingUpdate  *time.Time     `json:"last_tracking_update,omitempty" db:"last_tracking_update"`
	DaysWithoutUpdate   int            `json:"days_without_update" db:"days_without_update"`
	TrackingStale       bool           `json:"tracking_stale" db:"tracking_stale"`
	
	// Investigation
	InvestigationOpenedAt *time.Time   `json:"investigation_opened_at,omitempty" db:"investigation_opened_at"`
	InvestigationReason   string       `json:"investigation_reason,omitempty" db:"investigation_reason"`
	MarkedLostAt          *time.Time   `json:"marked_lost_at,omitempty" db:"marked_lost_at"`
	LostReason            string       `json:"lost_reason,omitempty" db:"lost_reason"`
	
	// Delivery control
	DeliveryAttempts      int          `json:"delivery_attempts" db:"delivery_attempts"`
	LastDeliveryAttemptAt *time.Time   `json:"last_delivery_attempt_at,omitempty" db:"last_delivery_attempt_at"`
	DeliveryNotes         string       `json:"delivery_notes,omitempty" db:"delivery_notes"`
	RecipientNameConfirmed string      `json:"recipient_name_confirmed,omitempty" db:"recipient_name_confirmed"`
	DeliveryPhotoURL      string       `json:"delivery_photo_url,omitempty" db:"delivery_photo_url"`
	
	// Reship tracking
	ReshipCount           int          `json:"reship_count" db:"reship_count"`
	OriginalShipmentID    *int         `json:"original_shipment_id,omitempty" db:"original_shipment_id"`
	IsReplacement         bool         `json:"is_replacement" db:"is_replacement"`
	ReshipReason          string       `json:"reship_reason,omitempty" db:"reship_reason"`
	ReplacedByShipmentID  *int         `json:"replaced_by_shipment_id,omitempty" db:"replaced_by_shipment_id"`
	ReshipCost            float64      `json:"reship_cost" db:"reship_cost"`
	ReshipCostBearer      string       `json:"reship_cost_bearer,omitempty" db:"reship_cost_bearer"`
	
	// Admin control
	RequiresAdminAction   bool         `json:"requires_admin_action" db:"requires_admin_action"`
	AdminActionReason     string       `json:"admin_action_reason,omitempty" db:"admin_action_reason"`
	StatusMetadata        map[string]any `json:"status_metadata,omitempty" db:"status_metadata"`
	
	// Related data
	TrackingHistory     []TrackingEvent `json:"tracking_history,omitempty" db:"-"`
}

// TrackingEvent represents a tracking history event
type TrackingEvent struct {
	ID          int            `json:"id" db:"id"`
	ShipmentID  int            `json:"shipment_id" db:"shipment_id"`
	Status      string         `json:"status" db:"status"`
	Description string         `json:"description" db:"description"`
	Location    string         `json:"location" db:"location"`
	EventTime   *time.Time     `json:"event_time" db:"event_time"`
	RawData     map[string]any `json:"raw_data,omitempty" db:"raw_data"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
}

// UserAddress represents a saved user address
type UserAddress struct {
	ID            int       `json:"id" db:"id"`
	UserID        *int      `json:"user_id,omitempty" db:"user_id"`
	Label         string    `json:"label" db:"label"`
	RecipientName string    `json:"recipient_name" db:"recipient_name"`
	Phone         string    `json:"phone" db:"phone"`
	ProvinceID    string    `json:"province_id" db:"province_id"`
	ProvinceName  string    `json:"province_name" db:"province_name"`
	CityID        string    `json:"city_id" db:"city_id"`
	CityName      string    `json:"city_name" db:"city_name"`
	DistrictID    string    `json:"district_id" db:"district_id"` // Kecamatan ID for shipping API
	District      string    `json:"district" db:"district"`       // Kecamatan name
	Subdistrict   string    `json:"subdistrict" db:"subdistrict"` // Kelurahan name
	PostalCode    string    `json:"postal_code" db:"postal_code"`
	FullAddress   string    `json:"full_address" db:"full_address"`
	IsDefault     bool      `json:"is_default" db:"is_default"`
	IsActive      bool      `json:"is_active" db:"is_active"`
	// Biteship native fields
	AreaID        string    `json:"area_id" db:"area_id"`         // Biteship area_id for shipping
	AreaName      string    `json:"area_name" db:"area_name"`     // Full area name from Biteship
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// ShippingAddressSnapshot stores address at time of order
type ShippingAddressSnapshot struct {
	RecipientName string `json:"recipient_name"`
	Phone         string `json:"phone"`
	ProvinceID    string `json:"province_id"`
	ProvinceName  string `json:"province_name"`
	CityID        string `json:"city_id"`
	CityName      string `json:"city_name"`
	DistrictID    string `json:"district_id"`   // Kecamatan ID for shipping API
	District      string `json:"district"`      // Kecamatan name
	Subdistrict   string `json:"subdistrict"`   // Kelurahan name
	PostalCode    string `json:"postal_code"`
	FullAddress   string `json:"full_address"`
}

// Subdistrict represents a kecamatan in Indonesia
type Subdistrict struct {
	ID          int       `json:"id" db:"id"`
	CityID      string    `json:"city_id" db:"city_id"`
	Name        string    `json:"name" db:"name"`
	PostalCodes []string  `json:"postal_codes" db:"postal_codes"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}
