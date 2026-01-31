package service

import (
	"crypto/rand"
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
	"zavera/models"
	"zavera/repository"
)

var (
	ErrOrderNotPendingPayment  = errors.New("order is not awaiting payment")
	ErrPaymentMethodInvalid    = errors.New("invalid payment method")
	ErrPaymentAlreadyProcessed = errors.New("payment already processed")
)

// CorePaymentService handles Tokopedia-style VA payments via Midtrans Core API
type CorePaymentService interface {
	// CreateVAPayment creates a VA payment via Midtrans Core API
	// Returns existing payment if one already exists (idempotent)
	CreateVAPayment(orderID int, paymentMethod string) (*CorePaymentResponse, error)

	// GetPaymentByOrderID gets payment details for an order
	// Triggers expiry check if payment is PENDING and expired
	GetPaymentByOrderID(orderID int) (*CorePaymentResponse, error)

	// CheckPaymentStatus checks current payment status from database
	CheckPaymentStatus(paymentID int) (*PaymentStatusResponse, error)

	// ProcessCoreWebhook processes Midtrans Core API webhook
	ProcessCoreWebhook(notification CoreWebhookNotification) error

	// GetPendingOrders returns pending orders for Menunggu Pembayaran tab
	GetPendingOrders(userID int, page, pageSize int) (*PendingOrdersResponse, error)

	// GetTransactionHistory returns transaction history for Daftar Transaksi tab
	GetTransactionHistory(userID int, page, pageSize int) (*TransactionHistoryResponse, error)

	// GetTransactionHistoryWithFilter returns filtered transaction history (Tokopedia-style)
	// filter: "all", "ongoing", "completed", "failed"
	GetTransactionHistoryWithFilter(userID int, filter string, page, pageSize int) (*TransactionHistoryResponse, error)
}

// CorePaymentResponse represents the response for payment operations
type CorePaymentResponse struct {
	PaymentID        int                              `json:"payment_id"`
	OrderID          int                              `json:"order_id"`
	OrderCode        string                           `json:"order_code"`
	PaymentMethod    string                           `json:"payment_method"`
	Bank             string                           `json:"bank"`
	BankLogo         string                           `json:"bank_logo"`
	VANumber         string                           `json:"va_number"`
	Amount           float64                          `json:"amount"`
	ExpiryTime       time.Time                        `json:"expiry_time"`
	RemainingSeconds int                              `json:"remaining_seconds"`
	Status           string                           `json:"status"`
	Instructions     []models.PaymentInstructionGroup `json:"instructions"`
	// GoPay specific fields
	QRCodeURL        string                           `json:"qr_code_url,omitempty"`
	DeeplinkURL      string                           `json:"deeplink_url,omitempty"`
	// Order details for receipt display
	OrderDetails     *OrderDetailsForReceipt          `json:"order_details,omitempty"`
}

// OrderDetailsForReceipt contains order information for receipt display
type OrderDetailsForReceipt struct {
	Items           []OrderItemForReceipt `json:"items"`
	Subtotal        float64               `json:"subtotal"`
	ShippingCost    float64               `json:"shipping_cost"`
	Total           float64               `json:"total"`
	CustomerName    string                `json:"customer_name"`
	CustomerEmail   string                `json:"customer_email"`
	CustomerPhone   string                `json:"customer_phone"`
	ShippingAddress string                `json:"shipping_address"`
	CourierName     string                `json:"courier_name,omitempty"`
	CourierService  string                `json:"courier_service,omitempty"`
}

// OrderItemForReceipt represents an order item for receipt display
type OrderItemForReceipt struct {
	ProductName  string  `json:"product_name"`
	ProductImage string  `json:"product_image"`
	Quantity     int     `json:"quantity"`
	PricePerUnit float64 `json:"price_per_unit"`
	Subtotal     float64 `json:"subtotal"`
}

// PaymentStatusResponse represents the response for status check
type PaymentStatusResponse struct {
	PaymentID int    `json:"payment_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

// PendingOrdersResponse represents pending orders list
type PendingOrdersResponse struct {
	Orders     []PendingOrderItem `json:"orders"`
	TotalCount int                `json:"total_count"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
}

// PendingOrderItem represents a pending order
type PendingOrderItem struct {
	OrderID          int        `json:"order_id"`
	OrderCode        string     `json:"order_code"`
	TotalAmount      float64    `json:"total_amount"`
	ItemCount        int        `json:"item_count"`
	ItemSummary      string     `json:"item_summary"`
	CreatedAt        time.Time  `json:"created_at"`
	HasPayment       bool       `json:"has_payment"`
	PaymentMethod    *string    `json:"payment_method,omitempty"`
	Bank             *string    `json:"bank,omitempty"`
	BankLogo         *string    `json:"bank_logo,omitempty"`
	VANumberMasked   *string    `json:"va_number_masked,omitempty"`
	ExpiryTime       *time.Time `json:"expiry_time,omitempty"`
	RemainingSeconds *int       `json:"remaining_seconds,omitempty"`
}

// TransactionHistoryResponse represents transaction history list
type TransactionHistoryResponse struct {
	Orders     []TransactionHistoryItem `json:"orders"`
	TotalCount int                      `json:"total_count"`
	Page       int                      `json:"page"`
	PageSize   int                      `json:"page_size"`
	Filter     string                   `json:"filter,omitempty"`
}

// TransactionHistoryItem represents a transaction history item (Tokopedia-style)
type TransactionHistoryItem struct {
	OrderID        int        `json:"order_id"`
	OrderCode      string     `json:"order_code"`
	TotalAmount    float64    `json:"total_amount"`
	ItemCount      int        `json:"item_count"`
	ItemSummary    string     `json:"item_summary"`
	ProductImage   string     `json:"product_image,omitempty"`
	Status         string     `json:"status"`
	PaymentMethod  *string    `json:"payment_method,omitempty"`
	Bank           *string    `json:"bank,omitempty"`
	Resi           string     `json:"resi,omitempty"`
	CourierName    *string    `json:"courier_name,omitempty"`
	CourierService *string    `json:"courier_service,omitempty"`
	ShipmentStatus *string    `json:"shipment_status,omitempty"`
	TrackingNumber *string    `json:"tracking_number,omitempty"`
	PaidAt         *time.Time `json:"paid_at,omitempty"`
	ShippedAt      *time.Time `json:"shipped_at,omitempty"`
	DeliveredAt    *time.Time `json:"delivered_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	CancelledAt    *time.Time `json:"cancelled_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// CoreWebhookNotification represents Midtrans Core API webhook payload
type CoreWebhookNotification struct {
	TransactionTime   string     `json:"transaction_time"`
	TransactionStatus string     `json:"transaction_status"`
	TransactionID     string     `json:"transaction_id"`
	StatusMessage     string     `json:"status_message"`
	StatusCode        string     `json:"status_code"`
	SignatureKey      string     `json:"signature_key"`
	PaymentType       string     `json:"payment_type"`
	OrderID           string     `json:"order_id"`
	MerchantID        string     `json:"merchant_id"`
	GrossAmount       string     `json:"gross_amount"`
	FraudStatus       string     `json:"fraud_status"`
	Currency          string     `json:"currency"`
	VANumbers         []VANumber `json:"va_numbers,omitempty"`
	PermataVANumber   string     `json:"permata_va_number,omitempty"`
}

type corePaymentService struct {
	orderPaymentRepo repository.OrderPaymentRepository
	orderRepo        repository.OrderRepository
	midtransClient   MidtransCoreClient
	emailService     EmailService
	serverKey        string
}

// NewCorePaymentService creates a new CorePaymentService
func NewCorePaymentService(
	orderPaymentRepo repository.OrderPaymentRepository,
	orderRepo repository.OrderRepository,
	serverKey string,
	emailService EmailService,
) CorePaymentService {
	return &corePaymentService{
		orderPaymentRepo: orderPaymentRepo,
		orderRepo:        orderRepo,
		midtransClient:   NewMidtransCoreClient(),
		emailService:     emailService,
		serverKey:        serverKey,
	}
}

// CreateVAPayment creates a VA payment (idempotent - returns existing if found)
func (s *corePaymentService) CreateVAPayment(orderID int, paymentMethod string) (*CorePaymentResponse, error) {
	log.Printf("🔄 CreateVAPayment: order_id=%d, method=%s", orderID, paymentMethod)

	// Validate payment method
	method := models.VAPaymentMethod(paymentMethod)
	if !method.IsValid() {
		return nil, ErrPaymentMethodInvalid
	}

	// Get order
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to find order: %w", err)
	}

	// Check order status - must be PENDING
	if order.Status != models.OrderStatusPending {
		log.Printf("❌ Order %d not awaiting payment, status: %s", orderID, order.Status)
		return nil, ErrOrderNotPendingPayment
	}

	// Check for existing PENDING payment (idempotency)
	existingPayment, err := s.orderPaymentRepo.FindPendingByOrderID(orderID)
	if err != nil && err != repository.ErrPaymentNotFound {
		return nil, fmt.Errorf("failed to check existing payment: %w", err)
	}

	if existingPayment != nil {
		log.Printf("✅ Returning existing payment: id=%d, va=%s", existingPayment.ID, existingPayment.VANumber)
		
		// Check if expired
		if existingPayment.IsExpired() {
			// Trigger expiry handling
			if err := s.handleExpiry(existingPayment, order); err != nil {
				log.Printf("⚠️ Failed to handle expiry: %v", err)
			}
			return nil, repository.ErrPaymentExpired
		}

		return s.buildPaymentResponse(existingPayment, order)
	}

	// Generate unique Midtrans order ID: ORDER_CODE-TIMESTAMP-RANDOM
	midtransOrderID := s.generateMidtransOrderID(order.OrderCode)

	// Call Midtrans Core API to create VA
	chargeResp, err := s.midtransClient.ChargeVA(ChargeVARequest{
		OrderID:       midtransOrderID,
		GrossAmount:   order.TotalAmount,
		PaymentMethod: method,
		CustomerName:  order.CustomerName,
		CustomerEmail: order.CustomerEmail,
		CustomerPhone: order.CustomerPhone,
	})
	if err != nil {
		log.Printf("❌ Midtrans ChargeVA failed: %v", err)
		return nil, fmt.Errorf("failed to create VA payment: %w", err)
	}

	// Create payment record
	payment := &models.OrderPayment{
		OrderID:         orderID,
		PaymentMethod:   method,
		Bank:            chargeResp.Bank,
		VANumber:        chargeResp.VANumber,
		TransactionID:   chargeResp.TransactionID,
		MidtransOrderID: midtransOrderID,
		ExpiryTime:      chargeResp.ExpiryTime,
		PaymentStatus:   models.CorePaymentStatusPending,
		RawResponse:     chargeResp.RawResponse,
		QRCodeURL:       chargeResp.QRCodeURL,
		DeeplinkURL:     chargeResp.DeeplinkURL,
	}

	if err := s.orderPaymentRepo.Create(payment); err != nil {
		if err == repository.ErrPaymentAlreadyExists {
			// Race condition - another request created payment, return existing
			existingPayment, _ := s.orderPaymentRepo.FindPendingByOrderID(orderID)
			if existingPayment != nil {
				return s.buildPaymentResponse(existingPayment, order)
			}
		}
		return nil, fmt.Errorf("failed to create payment record: %w", err)
	}

	// Order status remains PENDING (no need to change to MENUNGGU_PEMBAYARAN)
	// Payment record created, order stays PENDING until payment confirmed

	log.Printf("✅ VA Payment created: id=%d, va=%s, bank=%s", payment.ID, payment.VANumber, payment.Bank)

	return s.buildPaymentResponse(payment, order)
}

// GetPaymentByOrderID gets payment details with expiry check
func (s *corePaymentService) GetPaymentByOrderID(orderID int) (*CorePaymentResponse, error) {
	log.Printf("🔍 GetPaymentByOrderID: order_id=%d", orderID)

	// Get order
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to find order: %w", err)
	}

	// Get payment
	payment, err := s.orderPaymentRepo.FindByOrderID(orderID)
	if err != nil {
		return nil, err
	}

	// Check expiry on access
	if payment.PaymentStatus == models.CorePaymentStatusPending && payment.IsExpired() {
		log.Printf("⏰ Payment %d expired on access, triggering expiry handling", payment.ID)
		if err := s.handleExpiry(payment, order); err != nil {
			log.Printf("⚠️ Failed to handle expiry: %v", err)
		}
		// Reload payment after expiry handling
		payment, _ = s.orderPaymentRepo.FindByOrderID(orderID)
	}

	return s.buildPaymentResponse(payment, order)
}

// CheckPaymentStatus checks current status from Midtrans API and syncs to database
func (s *corePaymentService) CheckPaymentStatus(paymentID int) (*PaymentStatusResponse, error) {
	log.Printf("🔍 CheckPaymentStatus: payment_id=%d", paymentID)

	// Get payment and order from database
	var payment models.OrderPayment
	var orderStatus string
	err := s.orderPaymentRepo.GetDB().QueryRow(`
		SELECT op.id, op.order_id, op.payment_status, op.midtrans_order_id, op.transaction_id, o.status
		FROM order_payments op
		JOIN orders o ON op.order_id = o.id
		WHERE op.id = $1
	`, paymentID).Scan(&payment.ID, &payment.OrderID, &payment.PaymentStatus, &payment.MidtransOrderID, &payment.TransactionID, &orderStatus)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrPaymentNotFound
		}
		return nil, err
	}

	// Check if order is cancelled by admin
	if orderStatus == "CANCELLED" {
		log.Printf("⚠️ Order is CANCELLED by admin, payment check not allowed")
		return &PaymentStatusResponse{
			PaymentID: paymentID,
			Status:    "CANCELLED",
			Message:   "Pesanan telah dibatalkan oleh admin. Silakan hubungi customer service untuk informasi lebih lanjut.",
		}, nil
	}

	// If already in final status, just return from database
	if payment.PaymentStatus.IsFinal() {
		return &PaymentStatusResponse{
			PaymentID: paymentID,
			Status:    string(payment.PaymentStatus),
			Message:   s.getStatusMessage(payment.PaymentStatus),
		}, nil
	}

	// Query Midtrans API for latest status
	midtransStatus, err := s.midtransClient.GetTransactionStatus(payment.MidtransOrderID)
	if err != nil {
		log.Printf("⚠️ Failed to get Midtrans status: %v, returning database status", err)
		// Fallback to database status if Midtrans API fails
		return &PaymentStatusResponse{
			PaymentID: paymentID,
			Status:    string(payment.PaymentStatus),
			Message:   s.getStatusMessage(payment.PaymentStatus),
		}, nil
	}

	// Map Midtrans status to our status
	newStatus := s.mapMidtransStatus(midtransStatus.TransactionStatus)
	log.Printf("📊 Midtrans status: %s -> Local status: %s", midtransStatus.TransactionStatus, newStatus)

	// If status changed, update database
	if newStatus != payment.PaymentStatus {
		log.Printf("🔄 Status changed from %s to %s, updating database", payment.PaymentStatus, newStatus)

		switch newStatus {
		case models.CorePaymentStatusPaid:
			// Update payment to PAID
			_, err = s.orderPaymentRepo.GetDB().Exec(`
				UPDATE order_payments 
				SET payment_status = 'PAID', transaction_id = $1, paid_at = NOW(), updated_at = NOW()
				WHERE id = $2
			`, midtransStatus.TransactionID, paymentID)
			if err != nil {
				log.Printf("❌ Failed to update payment to PAID: %v", err)
			}

			// Update order to PAID (using enum value from database)
			_, err = s.orderPaymentRepo.GetDB().Exec(`
				UPDATE orders SET status = 'PAID', paid_at = NOW(), updated_at = NOW()
				WHERE id = $1
			`, payment.OrderID)
			if err != nil {
				log.Printf("❌ Failed to update order to PAID: %v", err)
			}

			log.Printf("✅ Payment %d and order %d updated to PAID", paymentID, payment.OrderID)

			// Cart is already cleared during checkout - no need to clear again
			
			// Send notification to admin dashboard
			order, err := s.orderRepo.FindByID(payment.OrderID)
			if err == nil {
				log.Printf("📢 Sending payment notification for order %s", order.OrderCode)
				NotifyPaymentReceived(order.OrderCode, string(payment.PaymentMethod), order.TotalAmount)
			} else {
				log.Printf("⚠️ Failed to load order for notification: %v", err)
			}

			// Send payment success email (async)
			if s.emailService != nil {
				go func() {
					order, err := s.orderRepo.FindByID(payment.OrderID)
					if err != nil {
						log.Printf("⚠️ Failed to load order for email: %v", err)
						return
					}
					paymentMethodStr := string(payment.PaymentMethod)
					if err := s.emailService.SendPaymentSuccess(order, paymentMethodStr); err != nil {
						log.Printf("⚠️ Failed to send payment success email: %v", err)
					} else {
						log.Printf("📧 Payment success email sent for order %s", order.OrderCode)
					}
				}()
			}

		case models.CorePaymentStatusExpired:
			// Update payment to EXPIRED
			_, err = s.orderPaymentRepo.GetDB().Exec(`
				UPDATE order_payments SET payment_status = 'EXPIRED', updated_at = NOW()
				WHERE id = $1
			`, paymentID)
			if err != nil {
				log.Printf("❌ Failed to update payment to EXPIRED: %v", err)
			}

			// Update order to EXPIRED (using enum value from database)
			_, err = s.orderPaymentRepo.GetDB().Exec(`
				UPDATE orders SET status = 'EXPIRED', updated_at = NOW()
				WHERE id = $1
			`, payment.OrderID)
			if err != nil {
				log.Printf("❌ Failed to update order to EXPIRED: %v", err)
			}

			// Send notification to admin dashboard
			order, err := s.orderRepo.FindByID(payment.OrderID)
			if err == nil {
				log.Printf("📢 Sending payment expired notification for order %s", order.OrderCode)
				NotifyPaymentExpired(order.OrderCode)
			} else {
				log.Printf("⚠️ Failed to load order for notification: %v", err)
			}

		case models.CorePaymentStatusCancelled, models.CorePaymentStatusFailed:
			// Update payment status
			_, err = s.orderPaymentRepo.GetDB().Exec(`
				UPDATE order_payments SET payment_status = $1, updated_at = NOW()
				WHERE id = $2
			`, string(newStatus), paymentID)
			if err != nil {
				log.Printf("❌ Failed to update payment status: %v", err)
			}
		}
		
		// Log sync for audit
		syncLog := &models.CorePaymentSyncLog{
			PaymentID:            &paymentID,
			OrderID:              payment.OrderID,
			SyncType:             "manual_check",
			SyncStatus:           "SYNCED",
			LocalPaymentStatus:   string(payment.PaymentStatus),
			GatewayStatus:        midtransStatus.TransactionStatus,
			GatewayTransactionID: midtransStatus.TransactionID,
			HasMismatch:          true,
		}
		s.orderPaymentRepo.LogSync(syncLog)
	}

	return &PaymentStatusResponse{
		PaymentID: paymentID,
		Status:    string(newStatus),
		Message:   s.getStatusMessage(newStatus),
	}, nil
}

// mapMidtransStatus maps Midtrans transaction status to our payment status
func (s *corePaymentService) mapMidtransStatus(midtransStatus string) models.CorePaymentStatus {
	switch midtransStatus {
	case "settlement", "capture":
		return models.CorePaymentStatusPaid
	case "pending":
		return models.CorePaymentStatusPending
	case "expire":
		return models.CorePaymentStatusExpired
	case "cancel":
		return models.CorePaymentStatusCancelled
	case "deny", "failure":
		return models.CorePaymentStatusFailed
	default:
		return models.CorePaymentStatusPending
	}
}

// ProcessCoreWebhook processes Midtrans webhook with signature validation and idempotency
func (s *corePaymentService) ProcessCoreWebhook(notification CoreWebhookNotification) error {
	log.Printf("🔔 ProcessCoreWebhook: order_id=%s, status=%s", notification.OrderID, notification.TransactionStatus)

	// 1. Validate signature
	if !s.verifySignature(notification) {
		log.Printf("❌ Invalid signature for order: %s", notification.OrderID)
		return ErrInvalidSignature
	}

	// 2. Extract original order code from Midtrans order ID
	orderCode := s.extractOrderCode(notification.OrderID)
	log.Printf("📋 Extracted order code: %s from %s", orderCode, notification.OrderID)

	// 3. Get order with row lock (with retry for race condition)
	var order *models.Order
	var tx *sql.Tx
	var err error
	
	// Retry up to 3 times with 500ms delay (handles race condition with payment creation)
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		order, tx, err = s.orderRepo.FindByOrderCodeForUpdate(orderCode)
		if err == nil {
			break
		}
		if attempt < maxRetries {
			log.Printf("⏳ Order not found (attempt %d/%d), retrying in 500ms...", attempt, maxRetries)
			time.Sleep(500 * time.Millisecond)
		}
	}
	
	if err != nil {
		log.Printf("❌ Order not found after %d attempts: %s", maxRetries, orderCode)
		return fmt.Errorf("order not found: %s", orderCode)
	}
	
	defer func() {
		if tx != nil {
			tx.Rollback()
		}
	}()

	// 4. Get payment (with retry for race condition)
	var payment *models.OrderPayment
	for attempt := 1; attempt <= maxRetries; attempt++ {
		payment, err = s.orderPaymentRepo.FindByOrderID(order.ID)
		if err == nil {
			break
		}
		if attempt < maxRetries {
			log.Printf("⏳ Payment not found (attempt %d/%d), retrying in 500ms...", attempt, maxRetries)
			time.Sleep(500 * time.Millisecond)
		}
	}
	
	if err != nil {
		log.Printf("❌ Payment not found after %d attempts for order: %d", maxRetries, order.ID)
		return fmt.Errorf("payment not found")
	}

	// 5. Idempotency check - skip if already in final status
	if payment.PaymentStatus.IsFinal() {
		log.Printf("⏭️ Payment %d already final: %s, skipping", payment.ID, payment.PaymentStatus)
		tx.Commit()
		return nil
	}

	// 6. Process based on transaction status
	var processErr error
	switch notification.TransactionStatus {
	case "settlement", "capture":
		processErr = s.handleWebhookSettlement(tx, order, payment, notification)
	case "expire":
		processErr = s.handleWebhookExpire(tx, order, payment)
	case "cancel":
		processErr = s.handleWebhookCancel(tx, payment)
	case "deny":
		processErr = s.handleWebhookDeny(tx, payment)
	default:
		log.Printf("⚠️ Unhandled transaction status: %s", notification.TransactionStatus)
	}

	if processErr != nil {
		return processErr
	}

	// 7. Commit transaction
	if err := tx.Commit(); err != nil {
		log.Printf("❌ Failed to commit webhook transaction: %v", err)
		return err
	}

	// 8. Send email for successful payment (async, after commit)
	if notification.TransactionStatus == "settlement" || notification.TransactionStatus == "capture" {
		if s.emailService != nil {
			go func() {
				// Reload order to get fresh data after commit
				freshOrder, err := s.orderRepo.FindByOrderCode(orderCode)
				if err == nil {
					paymentMethodStr := string(payment.PaymentMethod)
					if err := s.emailService.SendPaymentSuccess(freshOrder, paymentMethodStr); err != nil {
						log.Printf("⚠️ Failed to send payment success email: %v", err)
					} else {
						log.Printf("✅ Payment success email sent for order %s", orderCode)
					}
				}
			}()
		}
	}

	// 9. Log sync for audit
	s.logWebhookSync(order, payment, notification)

	log.Printf("✅ Webhook processed successfully for order: %s", orderCode)
	return nil
}

// GetPendingOrders returns pending orders for Menunggu Pembayaran tab
func (s *corePaymentService) GetPendingOrders(userID int, page, pageSize int) (*PendingOrdersResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	orders, totalCount, err := s.orderPaymentRepo.GetPendingPaymentsForUser(userID, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]PendingOrderItem, 0, len(orders))
	for _, o := range orders {
		item := PendingOrderItem{
			OrderID:     o["order_id"].(int),
			OrderCode:   o["order_code"].(string),
			TotalAmount: o["total_amount"].(float64),
			ItemCount:   o["item_count"].(int),
			ItemSummary: o["item_summary"].(string),
			CreatedAt:   o["created_at"].(time.Time),
			HasPayment:  o["has_payment"].(bool),
		}

		if item.HasPayment {
			pm := o["payment_method"].(string)
			bank := o["bank"].(string)
			bankLogo := s.getBankLogo(bank)
			vaMasked := o["va_number_masked"].(string)
			expiry := o["expiry_time"].(time.Time)
			remaining := o["remaining_seconds"].(int)

			item.PaymentMethod = &pm
			item.Bank = &bank
			item.BankLogo = &bankLogo
			item.VANumberMasked = &vaMasked
			item.ExpiryTime = &expiry
			item.RemainingSeconds = &remaining
		}

		items = append(items, item)
	}

	return &PendingOrdersResponse{
		Orders:     items,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

// GetTransactionHistory returns transaction history for Daftar Transaksi tab
func (s *corePaymentService) GetTransactionHistory(userID int, page, pageSize int) (*TransactionHistoryResponse, error) {
	return s.GetTransactionHistoryWithFilter(userID, "all", page, pageSize)
}

// GetTransactionHistoryWithFilter returns filtered transaction history (Tokopedia-style)
func (s *corePaymentService) GetTransactionHistoryWithFilter(userID int, filter string, page, pageSize int) (*TransactionHistoryResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	orders, totalCount, err := s.orderPaymentRepo.GetTransactionHistoryForUserWithFilter(userID, filter, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]TransactionHistoryItem, 0, len(orders))
	for _, o := range orders {
		item := TransactionHistoryItem{
			OrderID:     o["order_id"].(int),
			OrderCode:   o["order_code"].(string),
			TotalAmount: o["total_amount"].(float64),
			ItemCount:   o["item_count"].(int),
			ItemSummary: o["item_summary"].(string),
			Status:      o["status"].(string),
			CreatedAt:   o["created_at"].(time.Time),
		}

		// Product image
		if img, ok := o["product_image"].(string); ok && img != "" {
			item.ProductImage = img
		}

		// Resi
		if resi, ok := o["resi"].(string); ok && resi != "" {
			item.Resi = resi
		}

		// Payment info
		if pm, ok := o["payment_method"].(string); ok && pm != "" {
			item.PaymentMethod = &pm
		}
		if bank, ok := o["bank"].(string); ok && bank != "" {
			item.Bank = &bank
		}

		// Timestamps
		if paidAt, ok := o["paid_at"].(time.Time); ok {
			item.PaidAt = &paidAt
		}
		if shippedAt, ok := o["shipped_at"].(time.Time); ok {
			item.ShippedAt = &shippedAt
		}
		if deliveredAt, ok := o["delivered_at"].(time.Time); ok {
			item.DeliveredAt = &deliveredAt
		}
		if completedAt, ok := o["completed_at"].(time.Time); ok {
			item.CompletedAt = &completedAt
		}
		if cancelledAt, ok := o["cancelled_at"].(time.Time); ok {
			item.CancelledAt = &cancelledAt
		}

		// Shipment info
		if courierName, ok := o["courier_name"].(string); ok && courierName != "" {
			item.CourierName = &courierName
		}
		if courierService, ok := o["courier_service"].(string); ok && courierService != "" {
			item.CourierService = &courierService
		}
		if shipmentStatus, ok := o["shipment_status"].(string); ok && shipmentStatus != "" {
			item.ShipmentStatus = &shipmentStatus
		}
		if trackingNumber, ok := o["tracking_number"].(string); ok && trackingNumber != "" {
			item.TrackingNumber = &trackingNumber
		}

		items = append(items, item)
	}

	return &TransactionHistoryResponse{
		Orders:     items,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		Filter:     filter,
	}, nil
}

// ============================================
// HELPER METHODS
// ============================================

func (s *corePaymentService) generateMidtransOrderID(orderCode string) string {
	timestamp := time.Now().Unix()
	randomSuffix := s.generateRandomSuffix(4)
	return fmt.Sprintf("%s-%d-%s", orderCode, timestamp, randomSuffix)
}

func (s *corePaymentService) generateRandomSuffix(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:length]
}

func (s *corePaymentService) extractOrderCode(midtransOrderID string) string {
	// Format: ZVR-YYYYMMDD-XXXXXXXX-TIMESTAMP-RANDOM
	// Extract: ZVR-YYYYMMDD-XXXXXXXX
	parts := strings.Split(midtransOrderID, "-")
	if len(parts) >= 3 {
		return strings.Join(parts[:3], "-")
	}
	return midtransOrderID
}

func (s *corePaymentService) verifySignature(n CoreWebhookNotification) bool {
	signatureInput := n.OrderID + n.StatusCode + n.GrossAmount + s.serverKey
	hash := sha512.New()
	hash.Write([]byte(signatureInput))
	calculated := hex.EncodeToString(hash.Sum(nil))
	
	log.Printf("🔐 Signature verification:")
	log.Printf("   Order ID: %s", n.OrderID)
	log.Printf("   Status Code: %s", n.StatusCode)
	log.Printf("   Gross Amount: %s", n.GrossAmount)
	log.Printf("   Signature Input: %s", signatureInput)
	log.Printf("   Calculated: %s", calculated)
	log.Printf("   Received: %s", n.SignatureKey)
	log.Printf("   Match: %v", calculated == n.SignatureKey)
	
	return calculated == n.SignatureKey
}

func (s *corePaymentService) getBankLogo(bank string) string {
	logos := map[string]string{
		"bca":     "/images/banks/bca.png",
		"bri":     "/images/banks/bri.png",
		"bni":     "/images/banks/bni.png",
		"mandiri": "/images/banks/mandiri.png",
		"permata": "/images/banks/permata.png",
		"gopay":   "/images/payments/gopay.png",
		"qris":    "/images/payments/qris.png",
	}
	if logo, ok := logos[bank]; ok {
		return logo
	}
	return "/images/banks/default.svg"
}

func (s *corePaymentService) getStatusMessage(status models.CorePaymentStatus) string {
	messages := map[models.CorePaymentStatus]string{
		models.CorePaymentStatusPending:   "Pembayaran belum diterima",
		models.CorePaymentStatusPaid:      "Pembayaran berhasil",
		models.CorePaymentStatusExpired:   "Pembayaran telah kadaluarsa",
		models.CorePaymentStatusCancelled: "Pembayaran dibatalkan",
		models.CorePaymentStatusFailed:    "Pembayaran gagal",
	}
	if msg, ok := messages[status]; ok {
		return msg
	}
	return "Status tidak diketahui"
}

func (s *corePaymentService) buildPaymentResponse(payment *models.OrderPayment, order *models.Order) (*CorePaymentResponse, error) {
	instructions, _ := s.orderPaymentRepo.GetBankInstructions(payment.Bank)

	// Build order details for receipt
	orderDetails := &OrderDetailsForReceipt{
		Items:         []OrderItemForReceipt{},
		Subtotal:      order.Subtotal,
		ShippingCost:  order.ShippingCost,
		Total:         order.TotalAmount,
		CustomerName:  order.CustomerName,
		CustomerEmail: order.CustomerEmail,
		CustomerPhone: order.CustomerPhone,
	}

	// Get order items
	for _, item := range order.Items {
		orderDetails.Items = append(orderDetails.Items, OrderItemForReceipt{
			ProductName:  item.ProductName,
			ProductImage: item.ProductImage,
			Quantity:     item.Quantity,
			PricePerUnit: item.PricePerUnit,
			Subtotal:     item.Subtotal,
		})
	}

	// Get shipping info from shipments table
	var courierName, courierService, shippingAddress sql.NullString
	err := s.orderPaymentRepo.GetDB().QueryRow(`
		SELECT 
			COALESCE(provider_name, ''), 
			COALESCE(service_name, ''),
			COALESCE(recipient_address, '')
		FROM shipments 
		WHERE order_id = $1
	`, order.ID).Scan(&courierName, &courierService, &shippingAddress)
	
	if err == nil {
		orderDetails.CourierName = courierName.String
		orderDetails.CourierService = courierService.String
		orderDetails.ShippingAddress = shippingAddress.String
	} else if err != sql.ErrNoRows {
		log.Printf("⚠️ Failed to get shipment info: %v", err)
	}

	return &CorePaymentResponse{
		PaymentID:        payment.ID,
		OrderID:          payment.OrderID,
		OrderCode:        order.OrderCode,
		PaymentMethod:    string(payment.PaymentMethod),
		Bank:             payment.Bank,
		BankLogo:         s.getBankLogo(payment.Bank),
		VANumber:         payment.VANumber,
		Amount:           order.TotalAmount,
		ExpiryTime:       payment.ExpiryTime,
		RemainingSeconds: payment.GetRemainingSeconds(),
		Status:           string(payment.PaymentStatus),
		Instructions:     instructions,
		QRCodeURL:        payment.QRCodeURL,
		DeeplinkURL:      payment.DeeplinkURL,
		OrderDetails:     orderDetails,
	}, nil
}

func (s *corePaymentService) handleExpiry(payment *models.OrderPayment, order *models.Order) error {
	log.Printf("⏰ Handling expiry for payment %d, order %d", payment.ID, order.ID)

	// Update payment status
	if err := s.orderPaymentRepo.UpdateToExpired(payment.ID); err != nil {
		return err
	}

	// Update order status
	s.orderRepo.UpdateStatus(order.ID, models.OrderStatusKadaluarsa)

	// Restore stock
	if order.StockReserved {
		s.orderRepo.RestoreStock(order.ID)
	}

	log.Printf("✅ Expiry handled for payment %d", payment.ID)
	return nil
}

func (s *corePaymentService) handleWebhookSettlement(tx *sql.Tx, order *models.Order, payment *models.OrderPayment, n CoreWebhookNotification) error {
	log.Printf("💰 Processing settlement for order %s", order.OrderCode)

	// Update payment to PAID
	_, err := tx.Exec(`
		UPDATE order_payments 
		SET payment_status = 'PAID', transaction_id = $1, paid_at = NOW(), updated_at = NOW()
		WHERE id = $2
	`, n.TransactionID, payment.ID)
	if err != nil {
		return err
	}

	// Update order to PAID (using enum value from database)
	_, err = tx.Exec(`
		UPDATE orders SET status = 'PAID', paid_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`, order.ID)
	if err != nil {
		return err
	}

	log.Printf("✅ Settlement processed for order %s", order.OrderCode)
	
	// Send notification to admin dashboard
	log.Printf("📢 Sending payment notification for order %s", order.OrderCode)
	NotifyPaymentReceived(order.OrderCode, string(payment.PaymentMethod), order.TotalAmount)
	log.Printf("📢 Payment notification sent")
	
	// Send payment success email (async) - AFTER transaction commit
	// We'll do this after tx.Commit() in ProcessCoreWebhook
	
	return nil
}

func (s *corePaymentService) handleWebhookExpire(tx *sql.Tx, order *models.Order, payment *models.OrderPayment) error {
	log.Printf("⏰ Processing expire for order %s", order.OrderCode)

	// Update payment to EXPIRED
	_, err := tx.Exec(`
		UPDATE order_payments SET payment_status = 'EXPIRED', updated_at = NOW()
		WHERE id = $1
	`, payment.ID)
	if err != nil {
		return err
	}

	// Update order to EXPIRED (using enum value from database)
	_, err = tx.Exec(`
		UPDATE orders SET status = 'EXPIRED', updated_at = NOW()
		WHERE id = $1
	`, order.ID)
	if err != nil {
		return err
	}

	// Restore stock
	if order.StockReserved {
		s.orderRepo.RestoreStockTx(tx, order.ID)
	}

	log.Printf("✅ Expire processed for order %s", order.OrderCode)
	
	// Send notification to admin dashboard
	NotifyPaymentExpired(order.OrderCode)
	
	return nil
}

func (s *corePaymentService) handleWebhookCancel(tx *sql.Tx, payment *models.OrderPayment) error {
	_, err := tx.Exec(`
		UPDATE order_payments SET payment_status = 'CANCELLED', updated_at = NOW()
		WHERE id = $1
	`, payment.ID)
	return err
}

func (s *corePaymentService) handleWebhookDeny(tx *sql.Tx, payment *models.OrderPayment) error {
	_, err := tx.Exec(`
		UPDATE order_payments SET payment_status = 'FAILED', updated_at = NOW()
		WHERE id = $1
	`, payment.ID)
	return err
}

func (s *corePaymentService) logWebhookSync(order *models.Order, payment *models.OrderPayment, n CoreWebhookNotification) {
	syncLog := &models.CorePaymentSyncLog{
		PaymentID:            &payment.ID,
		OrderID:              order.ID,
		OrderCode:            order.OrderCode,
		SyncType:             "webhook",
		SyncStatus:           "SYNCED",
		LocalPaymentStatus:   string(payment.PaymentStatus),
		LocalOrderStatus:     string(order.Status),
		GatewayStatus:        n.TransactionStatus,
		GatewayTransactionID: n.TransactionID,
		HasMismatch:          false,
	}
	s.orderPaymentRepo.LogSync(syncLog)
}

