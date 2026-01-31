package models

import "time"

// VAPaymentMethod represents the VA payment method types
type VAPaymentMethod string

const (
	VAPaymentMethodBCA        VAPaymentMethod = "bca_va"
	VAPaymentMethodBRI        VAPaymentMethod = "bri_va"
	VAPaymentMethodMandiri    VAPaymentMethod = "mandiri_va"
	VAPaymentMethodPermata    VAPaymentMethod = "permata_va"
	VAPaymentMethodBNI        VAPaymentMethod = "bni_va"
	VAPaymentMethodQRIS       VAPaymentMethod = "qris"
	VAPaymentMethodGoPay      VAPaymentMethod = "gopay"
	VAPaymentMethodCreditCard VAPaymentMethod = "credit_card"
)

// IsValid checks if the payment method is valid
func (m VAPaymentMethod) IsValid() bool {
	switch m {
	case VAPaymentMethodBCA, VAPaymentMethodBRI, VAPaymentMethodMandiri,
		VAPaymentMethodPermata, VAPaymentMethodBNI, VAPaymentMethodQRIS, 
		VAPaymentMethodGoPay, VAPaymentMethodCreditCard:
		return true
	}
	return false
}

// GetBank returns the bank code for the payment method
func (m VAPaymentMethod) GetBank() string {
	switch m {
	case VAPaymentMethodBCA:
		return "bca"
	case VAPaymentMethodBRI:
		return "bri"
	case VAPaymentMethodMandiri:
		return "mandiri"
	case VAPaymentMethodPermata:
		return "permata"
	case VAPaymentMethodBNI:
		return "bni"
	case VAPaymentMethodQRIS:
		return "qris"
	case VAPaymentMethodGoPay:
		return "gopay"
	case VAPaymentMethodCreditCard:
		return "credit_card"
	}
	return ""
}

// GetDisplayName returns the display name for the payment method
func (m VAPaymentMethod) GetDisplayName() string {
	switch m {
	case VAPaymentMethodBCA:
		return "BCA Virtual Account"
	case VAPaymentMethodBRI:
		return "BRI Virtual Account"
	case VAPaymentMethodMandiri:
		return "Mandiri Virtual Account"
	case VAPaymentMethodPermata:
		return "Permata Virtual Account"
	case VAPaymentMethodBNI:
		return "BNI Virtual Account"
	case VAPaymentMethodQRIS:
		return "QRIS"
	case VAPaymentMethodGoPay:
		return "GoPay"
	case VAPaymentMethodCreditCard:
		return "Kartu Kredit / Debit"
	}
	return ""
}

// IsVA checks if the payment method is a Virtual Account
func (m VAPaymentMethod) IsVA() bool {
	switch m {
	case VAPaymentMethodBCA, VAPaymentMethodBRI, VAPaymentMethodMandiri,
		VAPaymentMethodPermata, VAPaymentMethodBNI:
		return true
	}
	return false
}

// IsQRIS checks if the payment method is QRIS
func (m VAPaymentMethod) IsQRIS() bool {
	return m == VAPaymentMethodQRIS
}

// IsGoPay checks if the payment method is GoPay
func (m VAPaymentMethod) IsGoPay() bool {
	return m == VAPaymentMethodGoPay
}

// IsEWallet checks if the payment method is an e-wallet (GoPay, QRIS)
func (m VAPaymentMethod) IsEWallet() bool {
	return m == VAPaymentMethodGoPay || m == VAPaymentMethodQRIS
}

// IsCreditCard checks if the payment method is Credit Card
func (m VAPaymentMethod) IsCreditCard() bool {
	return m == VAPaymentMethodCreditCard
}

// CorePaymentStatus represents the status of a Core API payment
type CorePaymentStatus string

const (
	CorePaymentStatusPending   CorePaymentStatus = "PENDING"
	CorePaymentStatusPaid      CorePaymentStatus = "PAID"
	CorePaymentStatusExpired   CorePaymentStatus = "EXPIRED"
	CorePaymentStatusCancelled CorePaymentStatus = "CANCELLED"
	CorePaymentStatusFailed    CorePaymentStatus = "FAILED"
)

// IsFinal checks if the payment status is a final state
func (s CorePaymentStatus) IsFinal() bool {
	switch s {
	case CorePaymentStatusPaid, CorePaymentStatusExpired, CorePaymentStatusCancelled, CorePaymentStatusFailed:
		return true
	}
	return false
}

// OrderPayment represents a VA payment record for an order
// Payment method is IMMUTABLE after creation
type OrderPayment struct {
	ID              int               `json:"id" db:"id"`
	OrderID         int               `json:"order_id" db:"order_id"`
	PaymentMethod   VAPaymentMethod   `json:"payment_method" db:"payment_method"`
	Bank            string            `json:"bank" db:"bank"`
	VANumber        string            `json:"va_number" db:"va_number"`
	TransactionID   string            `json:"transaction_id,omitempty" db:"transaction_id"`
	MidtransOrderID string            `json:"midtrans_order_id" db:"midtrans_order_id"`
	ExpiryTime      time.Time         `json:"expiry_time" db:"expiry_time"`
	PaymentStatus   CorePaymentStatus `json:"payment_status" db:"payment_status"`
	RawResponse     map[string]any    `json:"raw_response,omitempty" db:"raw_response"`
	// GoPay specific fields
	QRCodeURL       string            `json:"qr_code_url,omitempty" db:"qr_code_url"`
	DeeplinkURL     string            `json:"deeplink_url,omitempty" db:"deeplink_url"`
	CreatedAt       time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at" db:"updated_at"`
	PaidAt          *time.Time        `json:"paid_at,omitempty" db:"paid_at"`
}

// IsExpired checks if the payment has expired based on expiry_time
func (p *OrderPayment) IsExpired() bool {
	return time.Now().After(p.ExpiryTime)
}

// GetRemainingSeconds returns the remaining seconds until expiry
// Returns 0 if already expired
func (p *OrderPayment) GetRemainingSeconds() int {
	remaining := time.Until(p.ExpiryTime)
	if remaining <= 0 {
		return 0
	}
	return int(remaining.Seconds())
}

// GetMaskedVANumber returns the VA number with only last 4 digits visible
func (p *OrderPayment) GetMaskedVANumber() string {
	if len(p.VANumber) <= 4 {
		return p.VANumber
	}
	return "****" + p.VANumber[len(p.VANumber)-4:]
}

// CanBeProcessed checks if the payment can still be processed
func (p *OrderPayment) CanBeProcessed() bool {
	return p.PaymentStatus == CorePaymentStatusPending && !p.IsExpired()
}

// BankPaymentInstruction represents a single payment instruction step
type BankPaymentInstruction struct {
	ID          int       `json:"id" db:"id"`
	Bank        string    `json:"bank" db:"bank"`
	Channel     string    `json:"channel" db:"channel"`
	StepOrder   int       `json:"step_order" db:"step_order"`
	Instruction string    `json:"instruction" db:"instruction"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// PaymentInstructionGroup represents grouped instructions by channel
type PaymentInstructionGroup struct {
	Channel string   `json:"channel"`
	Steps   []string `json:"steps"`
}

// CorePaymentSyncLog represents a payment sync audit log entry
type CorePaymentSyncLog struct {
	ID                   int       `json:"id" db:"id"`
	PaymentID            *int      `json:"payment_id,omitempty" db:"payment_id"`
	OrderID              int       `json:"order_id" db:"order_id"`
	OrderCode            string    `json:"order_code" db:"order_code"`
	SyncType             string    `json:"sync_type" db:"sync_type"` // webhook, expiry_check, manual
	SyncStatus           string    `json:"sync_status" db:"sync_status"` // SYNCED, FAILED, SKIPPED
	LocalPaymentStatus   string    `json:"local_payment_status,omitempty" db:"local_payment_status"`
	LocalOrderStatus     string    `json:"local_order_status,omitempty" db:"local_order_status"`
	GatewayStatus        string    `json:"gateway_status,omitempty" db:"gateway_status"`
	GatewayTransactionID string    `json:"gateway_transaction_id,omitempty" db:"gateway_transaction_id"`
	HasMismatch          bool      `json:"has_mismatch" db:"has_mismatch"`
	ErrorMessage         string    `json:"error_message,omitempty" db:"error_message"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
}

// OrderStatusForPayment represents order statuses relevant to payment flow
const (
	OrderStatusMenungguPembayaran OrderStatus = "MENUNGGU_PEMBAYARAN"
	OrderStatusDibayar            OrderStatus = "DIBAYAR"
	OrderStatusDibatalkan         OrderStatus = "DIBATALKAN"
	OrderStatusKadaluarsa         OrderStatus = "KADALUARSA"
)

// ValidPaymentOrderTransitions defines allowed order status transitions for payment flow
var ValidPaymentOrderTransitions = map[OrderStatus][]OrderStatus{
	OrderStatusMenungguPembayaran: {
		OrderStatusDibayar,     // Webhook: settlement
		OrderStatusKadaluarsa,  // Webhook: expire OR on-access expiry
		OrderStatusDibatalkan,  // User cancel (before payment selection only)
	},
	// DIBAYAR, DIBATALKAN, KADALUARSA are final states for payment flow
}

// IsValidPaymentTransition checks if an order status transition is valid for payment flow
func IsValidPaymentTransition(from, to OrderStatus) bool {
	allowed, exists := ValidPaymentOrderTransitions[from]
	if !exists {
		return false
	}
	for _, status := range allowed {
		if status == to {
			return true
		}
	}
	return false
}
