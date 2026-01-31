package models

import "time"

// OrderStatus represents the status of an order
// State machine: PENDING → PAID → PACKING → SHIPPED → DELIVERED → COMPLETED
// Terminal states: COMPLETED, CANCELLED, FAILED, EXPIRED, REFUNDED
type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "PENDING"    // Waiting for payment
	OrderStatusPaid       OrderStatus = "PAID"       // Payment confirmed
	OrderStatusPacking    OrderStatus = "PACKING"    // Being packed by warehouse
	OrderStatusShipped    OrderStatus = "SHIPPED"    // Handed to courier
	OrderStatusDelivered  OrderStatus = "DELIVERED"  // Delivered to customer
	OrderStatusCompleted  OrderStatus = "COMPLETED"  // Order complete
	OrderStatusCancelled  OrderStatus = "CANCELLED"  // Cancelled
	OrderStatusFailed     OrderStatus = "FAILED"     // Payment failed
	OrderStatusExpired    OrderStatus = "EXPIRED"    // Payment expired
	OrderStatusRefunded   OrderStatus = "REFUNDED"   // Refunded
	OrderStatusProcessing OrderStatus = "PROCESSING" // Legacy - maps to PACKING
)

// ValidOrderTransitions defines allowed status transitions
// Key: current status, Value: list of allowed next statuses
// This is the production-grade state machine for ZAVERA
var ValidOrderTransitions = map[OrderStatus][]OrderStatus{
	OrderStatusPending: {
		OrderStatusPaid,
		OrderStatusCancelled,
		OrderStatusFailed,
		OrderStatusExpired,
	},
	OrderStatusPaid: {
		OrderStatusPacking,    // Admin packs the order
		OrderStatusCancelled,  // Admin can cancel before packing
	},
	OrderStatusPacking: {
		OrderStatusShipped,    // Admin ships with resi
		OrderStatusCancelled,  // Admin can cancel before shipping
	},
	OrderStatusShipped: {
		OrderStatusDelivered,  // Courier confirms delivery
	},
	OrderStatusDelivered: {
		OrderStatusCompleted,  // Auto-complete or manual
		OrderStatusRefunded,   // Customer requests refund
	},
	OrderStatusCompleted: {
		OrderStatusRefunded,   // Post-completion refund
	},
	// Terminal states: CANCELLED, FAILED, EXPIRED, REFUNDED - no transitions allowed
}

// IsValidTransition checks if a status transition is allowed
func (s OrderStatus) IsValidTransition(next OrderStatus) bool {
	allowed, exists := ValidOrderTransitions[s]
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

// IsFinalStatus checks if the status is a terminal state
func (s OrderStatus) IsFinalStatus() bool {
	switch s {
	case OrderStatusCompleted, OrderStatusCancelled, OrderStatusFailed, OrderStatusRefunded:
		return true
	default:
		return false
	}
}

// RequiresStockRestore checks if transitioning to this status requires stock restoration
func (s OrderStatus) RequiresStockRestore() bool {
	switch s {
	case OrderStatusCancelled, OrderStatusFailed, OrderStatusExpired:
		return true
	default:
		return false
	}
}

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "PENDING"
	PaymentStatusProcessing PaymentStatus = "PROCESSING"
	PaymentStatusSuccess    PaymentStatus = "SUCCESS"
	PaymentStatusFailed     PaymentStatus = "FAILED"
	PaymentStatusExpired    PaymentStatus = "EXPIRED"
	PaymentStatusCancelled  PaymentStatus = "CANCELLED"
)

// User represents a registered user
type User struct {
	ID           int        `json:"id" db:"id"`
	Email        string     `json:"email" db:"email"`
	FirstName    string     `json:"first_name" db:"first_name"`
	Name         string     `json:"name" db:"name"`
	Phone        string     `json:"phone" db:"phone"`
	PasswordHash string     `json:"-" db:"password_hash"`
	Birthdate    *time.Time `json:"birthdate,omitempty" db:"birthdate"`
	IsVerified   bool       `json:"is_verified" db:"is_verified"`
	GoogleID     *string    `json:"google_id,omitempty" db:"google_id"`
	AuthProvider string     `json:"auth_provider" db:"auth_provider"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// EmailVerificationToken represents email verification token
type EmailVerificationToken struct {
	ID        int        `json:"id" db:"id"`
	UserID    int        `json:"user_id" db:"user_id"`
	Token     string     `json:"token" db:"token"`
	ExpiresAt time.Time  `json:"expires_at" db:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty" db:"used_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// UserSession represents a user session for refresh tokens
type UserSession struct {
	ID           int       `json:"id" db:"id"`
	UserID       int       `json:"user_id" db:"user_id"`
	RefreshToken string    `json:"refresh_token" db:"refresh_token"`
	UserAgent    string    `json:"user_agent" db:"user_agent"`
	IPAddress    string    `json:"ip_address" db:"ip_address"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// Product represents a product in the catalog
type Product struct {
	ID          int            `json:"id" db:"id"`
	Name        string         `json:"name" db:"name"`
	Slug        string         `json:"slug" db:"slug"`
	Description string         `json:"description" db:"description"`
	Price       float64        `json:"price" db:"price"`
	Stock       int            `json:"stock" db:"stock"`
	Weight      int            `json:"weight" db:"weight"` // Weight in grams
	Length      int            `json:"length" db:"length"` // Length in cm (for shipping)
	Width       int            `json:"width" db:"width"`   // Width in cm (for shipping)
	Height      int            `json:"height" db:"height"` // Height in cm (for shipping)
	IsActive    bool           `json:"is_active" db:"is_active"`
	Category    string         `json:"category" db:"category"`
	Subcategory string         `json:"subcategory" db:"subcategory"`
	Brand       string         `json:"brand" db:"brand"`          // Product brand (e.g., Nike, Adidas)
	Material    string         `json:"material" db:"material"`    // Product material (e.g., Cotton, Polyester)
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
	Images      []ProductImage `json:"images" db:"-"`
}

// ProductImage represents a product image
type ProductImage struct {
	ID           int       `json:"id" db:"id"`
	ProductID    int       `json:"product_id" db:"product_id"`
	ImageURL     string    `json:"image_url" db:"image_url"`
	IsPrimary    bool      `json:"is_primary" db:"is_primary"`
	DisplayOrder int       `json:"display_order" db:"display_order"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// Cart represents a shopping cart
type Cart struct {
	ID        int       `json:"id" db:"id"`
	UserID    *int      `json:"user_id,omitempty" db:"user_id"`
	SessionID string    `json:"session_id,omitempty" db:"session_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	Items     []CartItem `json:"items" db:"-"`
}

// CartItem represents an item in a cart
type CartItem struct {
	ID            int             `json:"id" db:"id"`
	CartID        int             `json:"cart_id" db:"cart_id"`
	ProductID     int             `json:"product_id" db:"product_id"`
	VariantID     *int            `json:"variant_id,omitempty" db:"variant_id"` // For variant products
	Quantity      int             `json:"quantity" db:"quantity"`
	PriceSnapshot float64         `json:"price_snapshot" db:"price_snapshot"`
	Metadata      map[string]any  `json:"metadata,omitempty" db:"metadata"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
	Product       *Product        `json:"product,omitempty" db:"-"`
}

// Order represents a customer order
type Order struct {
	ID              int            `json:"id" db:"id"`
	OrderCode       string         `json:"order_code" db:"order_code"`
	UserID          *int           `json:"user_id,omitempty" db:"user_id"`
	CustomerName    string         `json:"customer_name" db:"customer_name"`
	CustomerEmail   string         `json:"customer_email" db:"customer_email"`
	CustomerPhone   string         `json:"customer_phone" db:"customer_phone"`
	Subtotal        float64        `json:"subtotal" db:"subtotal"`
	ShippingCost    float64        `json:"shipping_cost" db:"shipping_cost"`
	Tax             float64        `json:"tax" db:"tax"`
	Discount        float64        `json:"discount" db:"discount"`
	TotalAmount     float64        `json:"total_amount" db:"total_amount"`
	Status          OrderStatus    `json:"status" db:"status"`
	StockReserved   bool           `json:"stock_reserved" db:"stock_reserved"`
	Resi            string         `json:"resi,omitempty" db:"resi"`
	OriginCity      string         `json:"origin_city,omitempty" db:"origin_city"`
	DestinationCity string         `json:"destination_city,omitempty" db:"destination_city"`
	Notes           string         `json:"notes,omitempty" db:"notes"`
	Metadata        map[string]any `json:"metadata,omitempty" db:"metadata"`
	// Refund tracking fields
	RefundStatus    *string        `json:"refund_status,omitempty" db:"refund_status"`
	RefundAmount    float64        `json:"refund_amount" db:"refund_amount"`
	RefundedAt      *time.Time     `json:"refunded_at,omitempty" db:"refunded_at"`
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at" db:"updated_at"`
	PaidAt          *time.Time     `json:"paid_at,omitempty" db:"paid_at"`
	ShippedAt       *time.Time     `json:"shipped_at,omitempty" db:"shipped_at"`
	DeliveredAt     *time.Time     `json:"delivered_at,omitempty" db:"delivered_at"`
	CompletedAt     *time.Time     `json:"completed_at,omitempty" db:"completed_at"`
	CancelledAt     *time.Time     `json:"cancelled_at,omitempty" db:"cancelled_at"`
	Items           []OrderItem    `json:"items" db:"-"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ID           int            `json:"id" db:"id"`
	OrderID      int            `json:"order_id" db:"order_id"`
	ProductID    int            `json:"product_id" db:"product_id"`
	VariantID    *int           `json:"variant_id,omitempty" db:"variant_id"` // For variant products
	ProductName  string         `json:"product_name" db:"product_name"`
	ProductImage string         `json:"product_image" db:"product_image"`
	Quantity     int            `json:"quantity" db:"quantity"`
	PricePerUnit float64        `json:"price_per_unit" db:"price_per_unit"`
	Subtotal     float64        `json:"subtotal" db:"subtotal"`
	Metadata     map[string]any `json:"metadata,omitempty" db:"metadata"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
}

// Payment represents a payment transaction
type Payment struct {
	ID               int            `json:"id" db:"id"`
	OrderID          int            `json:"order_id" db:"order_id"`
	PaymentMethod    string         `json:"payment_method" db:"payment_method"`
	PaymentProvider  string         `json:"payment_provider" db:"payment_provider"`
	Amount           float64        `json:"amount" db:"amount"`
	Status           PaymentStatus  `json:"status" db:"status"`
	ExternalID       string         `json:"external_id,omitempty" db:"external_id"`
	TransactionID    string         `json:"transaction_id,omitempty" db:"transaction_id"`
	ProviderResponse map[string]any `json:"provider_response,omitempty" db:"provider_response"`
	CreatedAt        time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at" db:"updated_at"`
	PaidAt           *time.Time     `json:"paid_at,omitempty" db:"paid_at"`
	ExpiredAt        *time.Time     `json:"expired_at,omitempty" db:"expired_at"`
}


// ============================================
// STOCK MOVEMENT SYSTEM
// ============================================

// StockMovementType represents the type of stock movement
type StockMovementType string

const (
	StockMovementReserve    StockMovementType = "RESERVE"    // Stock reserved at checkout
	StockMovementRelease    StockMovementType = "RELEASE"    // Stock released on cancel/expire
	StockMovementDeduct     StockMovementType = "DEDUCT"     // Stock permanently deducted
	StockMovementAdjustment StockMovementType = "ADJUSTMENT" // Manual adjustment
)

// StockMovement represents a stock operation for audit
type StockMovement struct {
	ID           int               `json:"id" db:"id"`
	ProductID    int               `json:"product_id" db:"product_id"`
	OrderID      *int              `json:"order_id,omitempty" db:"order_id"`
	MovementType StockMovementType `json:"movement_type" db:"movement_type"`
	Quantity     int               `json:"quantity" db:"quantity"`
	BalanceAfter int               `json:"balance_after" db:"balance_after"`
	Notes        string            `json:"notes,omitempty" db:"notes"`
	CreatedAt    time.Time         `json:"created_at" db:"created_at"`
}

// ============================================
// SHIPPING SNAPSHOT SYSTEM
// ============================================

// ShippingSnapshot stores Biteship response at checkout time
type ShippingSnapshot struct {
	ID                    int            `json:"id" db:"id"`
	OrderID               int            `json:"order_id" db:"order_id"`
	Courier               string         `json:"courier" db:"courier"`
	Service               string         `json:"service" db:"service"`
	Cost                  float64        `json:"cost" db:"cost"`
	ETD                   string         `json:"etd" db:"etd"`
	OriginCityID          string         `json:"origin_city_id" db:"origin_city_id"`
	OriginCityName        string         `json:"origin_city_name,omitempty" db:"origin_city_name"`
	DestinationCityID     string         `json:"destination_city_id" db:"destination_city_id"`
	DestinationCityName   string         `json:"destination_city_name,omitempty" db:"destination_city_name"`
	DestinationDistrictID string         `json:"destination_district_id,omitempty" db:"destination_district_id"`
	Weight                int            `json:"weight" db:"weight"`
	// Biteship area fields
	OriginAreaID          string         `json:"origin_area_id,omitempty" db:"origin_area_id"`
	OriginAreaName        string         `json:"origin_area_name,omitempty" db:"origin_area_name"`
	DestinationAreaID     string         `json:"destination_area_id,omitempty" db:"destination_area_id"`
	DestinationAreaName   string         `json:"destination_area_name,omitempty" db:"destination_area_name"`
	// Stores raw Biteship API response for audit purposes
	BiteshipRawJSON       map[string]any `json:"biteship_raw_json" db:"biteship_raw_json"`
	CreatedAt             time.Time      `json:"created_at" db:"created_at"`
}

// ============================================
// EMAIL SYSTEM
// ============================================

// EmailTemplate represents an email template
type EmailTemplate struct {
	ID              int       `json:"id" db:"id"`
	TemplateKey     string    `json:"template_key" db:"template_key"`
	Name            string    `json:"name" db:"name"`
	SubjectTemplate string    `json:"subject_template" db:"subject_template"`
	HTMLTemplate    string    `json:"html_template" db:"html_template"`
	IsActive        bool      `json:"is_active" db:"is_active"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// EmailLogStatus represents the status of an email log
type EmailLogStatus string

const (
	EmailLogStatusPending EmailLogStatus = "PENDING"
	EmailLogStatusSent    EmailLogStatus = "SENT"
	EmailLogStatusFailed  EmailLogStatus = "FAILED"
	EmailLogStatusRetry   EmailLogStatus = "RETRY"
)

// EmailLog represents a sent email record
type EmailLog struct {
	ID             int            `json:"id" db:"id"`
	OrderID        *int           `json:"order_id,omitempty" db:"order_id"`
	UserID         *int           `json:"user_id,omitempty" db:"user_id"`
	TemplateKey    string         `json:"template_key" db:"template_key"`
	RecipientEmail string         `json:"recipient_email" db:"recipient_email"`
	Subject        string         `json:"subject" db:"subject"`
	Status         EmailLogStatus `json:"status" db:"status"`
	ErrorMessage   string         `json:"error_message,omitempty" db:"error_message"`
	SentAt         *time.Time     `json:"sent_at,omitempty" db:"sent_at"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
}

// ============================================
// ORDER HELPER METHODS
// ============================================

// CanBeCancelledByCustomer checks if customer can cancel this order
func (o *Order) CanBeCancelledByCustomer() bool {
	return o.Status == OrderStatusPending
}

// CanBeCancelledByAdmin checks if admin can cancel this order
func (o *Order) CanBeCancelledByAdmin() bool {
	switch o.Status {
	case OrderStatusPending, OrderStatusPaid, OrderStatusPacking, OrderStatusExpired:
		return true
	default:
		return false
	}
}

// CanBeRefunded checks if order can be refunded
func (o *Order) CanBeRefunded() bool {
	switch o.Status {
	case OrderStatusDelivered, OrderStatusCompleted:
		return true
	default:
		return false
	}
}

// IsResiLocked checks if resi cannot be modified
func (o *Order) IsResiLocked() bool {
	switch o.Status {
	case OrderStatusShipped, OrderStatusDelivered, OrderStatusCompleted, OrderStatusRefunded:
		return true
	default:
		return false
	}
}
