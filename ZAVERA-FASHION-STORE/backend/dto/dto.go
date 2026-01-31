package dto

// ProductResponse represents the product API response
type ProductResponse struct {
	ID             int      `json:"id"`
	Name           string   `json:"name"`
	Slug           string   `json:"slug"`
	Description    string   `json:"description"`
	Price          float64  `json:"price"`
	Stock          int      `json:"stock"`
	Weight         int      `json:"weight"` // Weight in grams
	Length         int      `json:"length"` // Length in cm
	Width          int      `json:"width"`  // Width in cm
	Height         int      `json:"height"` // Height in cm
	ImageURL       string   `json:"image_url"`
	Images         []string `json:"images,omitempty"`
	Category       string   `json:"category"`
	Subcategory    string   `json:"subcategory,omitempty"`
	Brand          string   `json:"brand,omitempty"`          // Product brand (e.g., Nike, Adidas)
	Material       string   `json:"material,omitempty"`       // Product material (e.g., Cotton, Polyester)
	AvailableSizes []string `json:"available_sizes,omitempty"` // Sizes from active variants
}

// AddToCartRequest represents the request to add item to cart
type AddToCartRequest struct {
	ProductID int                    `json:"product_id" binding:"required"`
	VariantID *int                   `json:"variant_id,omitempty"` // Optional: for variant products
	Quantity  int                    `json:"quantity" binding:"required,min=1"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateCartItemRequest represents the request to update cart item quantity
type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" binding:"required,min=0"`
}

// CartResponse represents the cart API response
type CartResponse struct {
	ID        int              `json:"id"`
	Items     []CartItemResponse `json:"items"`
	Subtotal  float64          `json:"subtotal"`
	ItemCount int              `json:"item_count"`
}

// CartItemResponse represents a cart item in API response
type CartItemResponse struct {
	ID            int                    `json:"id"`
	ProductID     int                    `json:"product_id"`
	ProductName   string                 `json:"product_name"`
	ProductImage  string                 `json:"product_image"`
	Quantity      int                    `json:"quantity"`
	PricePerUnit  float64                `json:"price_per_unit"`
	Subtotal      float64                `json:"subtotal"`
	Stock         int                    `json:"stock"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// CheckoutRequest represents the checkout request
type CheckoutRequest struct {
	CustomerName  string `json:"customer_name" binding:"required"`
	CustomerEmail string `json:"customer_email" binding:"required,email"`
	CustomerPhone string `json:"customer_phone" binding:"required"`
	Notes         string `json:"notes,omitempty"`
}

// CheckoutResponse represents the checkout response
type CheckoutResponse struct {
	OrderID      int     `json:"order_id"`
	OrderCode    string  `json:"order_code"`
	TotalAmount  float64 `json:"total_amount"`
	Status       string  `json:"status"`
}

// OrderResponse represents the order API response
type OrderResponse struct {
	ID            int                 `json:"id"`
	OrderCode     string              `json:"order_code"`
	UserID        *int                `json:"user_id,omitempty"`
	CustomerName  string              `json:"customer_name"`
	CustomerEmail string              `json:"customer_email"`
	CustomerPhone string              `json:"customer_phone"`
	Subtotal      float64             `json:"subtotal"`
	ShippingCost  float64             `json:"shipping_cost"`
	Tax           float64             `json:"tax"`
	Discount      float64             `json:"discount"`
	TotalAmount   float64             `json:"total_amount"`
	Status        string              `json:"status"`
	Resi          string              `json:"resi,omitempty"`
	Items         []OrderItemResponse `json:"items"`
	CreatedAt     string              `json:"created_at"`
	PaidAt        string              `json:"paid_at,omitempty"`
	ShippedAt     string              `json:"shipped_at,omitempty"`
	DeliveredAt   string              `json:"delivered_at,omitempty"`
	// Shipping details
	Shipment *ShipmentResponse `json:"shipment,omitempty"`
}

// OrderItemResponse represents an order item in API response
type OrderItemResponse struct {
	ProductID    int     `json:"product_id"`
	ProductName  string  `json:"product_name"`
	ProductImage string  `json:"product_image,omitempty"`
	Quantity     int     `json:"quantity"`
	PricePerUnit float64 `json:"price_per_unit"`
	Subtotal     float64 `json:"subtotal"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// InitiatePaymentRequest represents the payment initiation request
type InitiatePaymentRequest struct {
	OrderID int `json:"order_id" binding:"required"`
}

// InitiatePaymentResponse represents the payment initiation response
type InitiatePaymentResponse struct {
	SnapToken string `json:"snap_token"`
}

// MidtransNotification represents Midtrans webhook notification payload
type MidtransNotification struct {
	TransactionTime   string `json:"transaction_time"`
	TransactionStatus string `json:"transaction_status"`
	TransactionID     string `json:"transaction_id"`
	StatusMessage     string `json:"status_message"`
	StatusCode        string `json:"status_code"`
	SignatureKey      string `json:"signature_key"`
	PaymentType       string `json:"payment_type"`
	OrderID           string `json:"order_id"`
	MerchantID        string `json:"merchant_id"`
	GrossAmount       string `json:"gross_amount"`
	FraudStatus       string `json:"fraud_status"`
	Currency          string `json:"currency"`
}


// ============================================
// AUTH DTOs
// ============================================

// RegisterRequest represents user registration request
type RegisterRequest struct {
	FirstName string `json:"first_name" binding:"required,min=2,max=100"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	Birthdate string `json:"birthdate" binding:"required"` // Format: YYYY-MM-DD
}

// LoginRequest represents user login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// GoogleLoginRequest represents Google OAuth login request
type GoogleLoginRequest struct {
	IDToken string `json:"id_token" binding:"required"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	User        UserResponse `json:"user"`
	AccessToken string       `json:"access_token"`
}

// UserResponse represents user data in API response
type UserResponse struct {
	ID           int     `json:"id"`
	Email        string  `json:"email"`
	FirstName    string  `json:"first_name"`
	Name         string  `json:"name,omitempty"`
	Phone        string  `json:"phone,omitempty"`
	Birthdate    *string `json:"birthdate,omitempty"`
	IsVerified   bool    `json:"is_verified"`
	AuthProvider string  `json:"auth_provider"`
	CreatedAt    string  `json:"created_at"`
}

// VerifyEmailRequest represents email verification request
type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

// ResendVerificationRequest represents resend verification email request
type ResendVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// UserOrdersResponse represents user's order history
type UserOrdersResponse struct {
	Orders     []OrderResponse `json:"orders"`
	TotalCount int             `json:"total_count"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
}
