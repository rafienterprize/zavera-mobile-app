package service

import (
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"
	"zavera/dto"
	"zavera/models"
	"zavera/repository"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
)

var (
	ErrPaymentNotFound    = errors.New("payment not found")
	ErrPaymentAlreadyPaid = errors.New("payment already completed")
	ErrInvalidSignature   = errors.New("invalid webhook signature")
	ErrOrderNotPending    = errors.New("order is not in pending status")
)

type PaymentService interface {
	InitiatePayment(orderID int) (string, error)
	ProcessWebhook(notification dto.MidtransNotification) error
	CreatePayment(orderID int) (string, error)
	HandleCallback(orderCode string, status models.PaymentStatus, transactionID string) error
}

type paymentService struct {
	paymentRepo  repository.PaymentRepository
	orderRepo    repository.OrderRepository
	shippingRepo repository.ShippingRepository
	resiService  ResiService
	emailService EmailService
	snapClient   snap.Client
	serverKey    string
}

func NewPaymentService(
	paymentRepo repository.PaymentRepository,
	orderRepo repository.OrderRepository,
	shippingRepo repository.ShippingRepository,
	emailRepo repository.EmailRepository,
) PaymentService {
	serverKey := os.Getenv("MIDTRANS_SERVER_KEY")
	var s snap.Client
	env := midtrans.Sandbox
	if os.Getenv("MIDTRANS_ENVIRONMENT") == "production" {
		env = midtrans.Production
	}
	s.New(serverKey, env)

	// Initialize resi service
	resiSvc := NewResiService(orderRepo)

	// Initialize email service
	var emailSvc EmailService
	if emailRepo != nil {
		emailSvc = NewEmailService(emailRepo)
	}

	return &paymentService{
		paymentRepo:  paymentRepo,
		orderRepo:    orderRepo,
		shippingRepo: shippingRepo,
		resiService:  resiSvc,
		emailService: emailSvc,
		snapClient:   s,
		serverKey:    serverKey,
	}
}

func (s *paymentService) InitiatePayment(orderID int) (string, error) {
	log.Printf("🔍 InitiatePayment called for orderID: %d", orderID)
	
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		log.Printf("❌ Order not found: %v", err)
		if err == sql.ErrNoRows {
			return "", ErrOrderNotFound
		}
		return "", err
	}

	log.Printf("📦 Order found: code=%s, status=%s, total=%.2f", order.OrderCode, order.Status, order.TotalAmount)

	if order.Status != models.OrderStatusPending {
		log.Printf("❌ Order not pending: %s", order.Status)
		return "", ErrOrderNotPending
	}

	existingPayment, err := s.paymentRepo.FindByOrderID(order.ID)
	if err == nil && existingPayment != nil {
		log.Printf("💳 Existing payment found: status=%s", existingPayment.Status)
		if existingPayment.Status == models.PaymentStatusPending {
			if token, ok := existingPayment.ProviderResponse["token"].(string); ok && token != "" {
				log.Printf("✅ Returning existing token (reuse)")
				return token, nil
			}
			// Token not found or empty, need to generate new one with unique ID
			log.Printf("⚠️ Existing payment has no valid token, generating new one")
		}
		if existingPayment.Status == models.PaymentStatusSuccess {
			return "", ErrPaymentAlreadyPaid
		}
	}

	// Generate unique transaction ID to avoid "order_id already taken" error
	// Format: ORDER_CODE-TIMESTAMP to make it unique for each payment attempt
	uniqueOrderID := fmt.Sprintf("%s-%d", order.OrderCode, time.Now().Unix())

	// Calculate items total to ensure it matches order total
	var itemsTotal int64
	var items []midtrans.ItemDetails
	for _, item := range order.Items {
		itemPrice := int64(item.PricePerUnit)
		itemsTotal += itemPrice * int64(item.Quantity)
		items = append(items, midtrans.ItemDetails{
			ID:    fmt.Sprintf("PROD-%d", item.ProductID),
			Name:  truncateString(item.ProductName, 50),
			Price: itemPrice,
			Qty:   int32(item.Quantity),
		})
	}

	if order.ShippingCost > 0 {
		shippingCost := int64(order.ShippingCost)
		itemsTotal += shippingCost
		items = append(items, midtrans.ItemDetails{
			ID: "SHIPPING", Name: "Shipping Cost",
			Price: shippingCost, Qty: 1,
		})
	}

	if order.Tax > 0 {
		tax := int64(order.Tax)
		itemsTotal += tax
		items = append(items, midtrans.ItemDetails{
			ID: "TAX", Name: "Tax",
			Price: tax, Qty: 1,
		})
	}

	// Use calculated items total as gross amount to avoid mismatch
	grossAmount := itemsTotal
	log.Printf("💰 Calculated gross amount: %d (order total: %.2f)", grossAmount, order.TotalAmount)

	req := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  uniqueOrderID, // Use unique ID to avoid "already taken" error
			GrossAmt: grossAmount,
		},
		CustomerDetail: &midtrans.CustomerDetails{
			FName: order.CustomerName,
			Email: order.CustomerEmail,
			Phone: order.CustomerPhone,
		},
		EnabledPayments: snap.AllSnapPaymentType,
		Items:           &items,
	}

	log.Printf("🚀 Calling Midtrans CreateTransaction with uniqueOrderID: %s", uniqueOrderID)
	snapResp, midtransErr := s.snapClient.CreateTransaction(req)
	if midtransErr != nil && !reflect.ValueOf(midtransErr).IsNil() {
		log.Printf("❌ Midtrans error: %v", midtransErr)
		return "", fmt.Errorf("failed to create payment: %v", midtransErr)
	}

	if snapResp == nil || snapResp.Token == "" {
		log.Printf("❌ Empty token from Midtrans")
		return "", fmt.Errorf("failed to create payment: empty token")
	}

	log.Printf("✅ Midtrans token created: %s", snapResp.Token)

	payment := &models.Payment{
		OrderID:         order.ID,
		PaymentMethod:   "midtrans",
		PaymentProvider: "midtrans_snap",
		Amount:          order.TotalAmount,
		Status:          models.PaymentStatusPending,
		ExternalID:      uniqueOrderID, // Store unique ID for webhook matching
		ProviderResponse: map[string]any{
			"token":         snapResp.Token,
			"redirect_url":  snapResp.RedirectURL,
			"order_code":    order.OrderCode, // Keep original order code for reference
			"unique_id":     uniqueOrderID,
		},
	}

	if existingPayment != nil {
		s.paymentRepo.UpdateToken(existingPayment.ID, snapResp.Token, snapResp.RedirectURL)
	} else {
		s.paymentRepo.Create(payment)
	}

	return snapResp.Token, nil
}

// ProcessWebhook - HARDENED VERSION with row locking to prevent race conditions
func (s *paymentService) ProcessWebhook(notification dto.MidtransNotification) error {
	log.Printf("Processing webhook for order: %s, status: %s", 
		notification.OrderID, notification.TransactionStatus)

	// 1. Verify signature FIRST (before any DB operations)
	if !s.verifySignature(notification) {
		log.Printf("Invalid signature for order: %s", notification.OrderID)
		return ErrInvalidSignature
	}

	// 2. Extract original order code from uniqueOrderID
	// Format: ZVR-YYYYMMDD-XXXXXXXX-TIMESTAMP or ZVR-YYYYMMDD-XXXXXXXX
	orderCode := notification.OrderID
	// If it contains timestamp suffix (more than 3 parts), extract original order code
	parts := strings.Split(notification.OrderID, "-")
	if len(parts) > 3 {
		// Reconstruct original order code: ZVR-YYYYMMDD-XXXXXXXX
		orderCode = strings.Join(parts[:3], "-")
		log.Printf("Extracted order code: %s from uniqueID: %s", orderCode, notification.OrderID)
	}

	// 3. Get order WITH ROW LOCK to prevent race condition
	order, tx, err := s.orderRepo.FindByOrderCodeForUpdate(orderCode)
	if err != nil {
		log.Printf("Order not found: %s (original: %s)", orderCode, notification.OrderID)
		return fmt.Errorf("order not found: %s", orderCode)
	}
	// Ensure transaction is handled properly
	defer func() {
		if tx != nil {
			tx.Rollback() // Will be no-op if already committed
		}
	}()

	// 4. Get payment (or create if not exists)
	payment, err := s.paymentRepo.FindByOrderID(order.ID)
	if err != nil {
		log.Printf("⚠️ Payment not found for order: %d, creating new payment record", order.ID)
		// Create payment record from webhook data WITHIN TRANSACTION
		payment = &models.Payment{
			OrderID:         order.ID,
			PaymentMethod:   notification.PaymentType,
			PaymentProvider: "midtrans",
			Amount:          order.TotalAmount,
			Status:          models.PaymentStatusPending,
			ExternalID:      notification.OrderID,
			TransactionID:   notification.TransactionID,
			ProviderResponse: map[string]any{
				"order_id":           notification.OrderID,
				"transaction_id":     notification.TransactionID,
				"payment_type":       notification.PaymentType,
				"transaction_status": notification.TransactionStatus,
			},
		}
		
		// Create payment within the same transaction
		query := `
			INSERT INTO payments (order_id, payment_method, payment_provider, amount, status, external_id, transaction_id, provider_response, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
			RETURNING id, created_at, updated_at
		`
		providerResponseJSON, _ := json.Marshal(payment.ProviderResponse)
		err = tx.QueryRow(query, 
			payment.OrderID, payment.PaymentMethod, payment.PaymentProvider, 
			payment.Amount, payment.Status, payment.ExternalID, payment.TransactionID,
			providerResponseJSON,
		).Scan(&payment.ID, &payment.CreatedAt, &payment.UpdatedAt)
		
		if err != nil {
			log.Printf("❌ Failed to create payment record: %v", err)
			return fmt.Errorf("failed to create payment: %w", err)
		}
		log.Printf("✅ Payment record created from webhook: ID=%d", payment.ID)
	}

	// 5. Map Midtrans status
	newStatus := s.mapMidtransStatus(notification.TransactionStatus, notification.FraudStatus)
	log.Printf("Mapped status: %s -> %s", notification.TransactionStatus, newStatus)

	// 6. IDEMPOTENCY: skip if already processed (now with lock held - safe!)
	if s.isFinalPaymentStatus(payment.Status) {
		log.Printf("Payment already final: %s, skipping", payment.Status)
		tx.Commit() // Release lock
		return nil
	}

	// 6. Process based on status (all within transaction)
	var processErr error
	switch newStatus {
	case models.PaymentStatusSuccess:
		processErr = s.handleSuccessTx(tx, order, payment, notification)
	case models.PaymentStatusExpired:
		processErr = s.handleExpiredTx(tx, order, payment, notification)
	case models.PaymentStatusCancelled:
		processErr = s.handleCancelledTx(tx, order, payment, notification)
	case models.PaymentStatusFailed:
		processErr = s.handleFailedTx(tx, order, payment, notification)
	default:
		// Just update payment status for pending
		s.paymentRepo.UpdateStatusWithResponse(payment.ID, newStatus, map[string]any{
			"transaction_id": notification.TransactionID,
		})
	}

	if processErr != nil {
		return processErr
	}

	// 7. Commit transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit webhook transaction: %v", err)
		return err
	}

	// 8. Log to payment_sync_log for audit trail (non-critical, outside transaction)
	s.logPaymentSync(order, payment, notification, newStatus)

	return nil
}

// handleSuccessTx handles successful payment within a transaction
func (s *paymentService) handleSuccessTx(tx *sql.Tx, order *models.Order, payment *models.Payment, n dto.MidtransNotification) error {
	// IDEMPOTENCY
	if order.Status == models.OrderStatusPaid {
		log.Printf("Order %s already PAID", order.OrderCode)
		return nil
	}

	if order.Status != models.OrderStatusPending {
		log.Printf("Order %s not PENDING, current: %s", order.OrderCode, order.Status)
		return nil
	}

	// Update ORDER status to PAID (within transaction)
	err := s.orderRepo.MarkAsPaidTx(tx, order.ID)
	if err != nil {
		log.Printf("Failed to mark order as paid: %v", err)
		return err
	}
	log.Printf("✅ Order %s marked as PAID", order.OrderCode)

	// Update PAYMENT status (within transaction)
	_, err = tx.Exec(`
		UPDATE payments 
		SET status = $1, transaction_id = $2, payment_method = $3, 
		    paid_at = NOW(), updated_at = NOW()
		WHERE id = $4
	`, models.PaymentStatusSuccess, n.TransactionID, n.PaymentType, payment.ID)
	if err != nil {
		log.Printf("Failed to update payment: %v", err)
		return err
	}

	// DON'T auto-generate resi here - resi should only be generated when admin ships the order
	// Resi generation moved to admin ship order endpoint
	
	// Update SHIPMENT status to PROCESSING (within same transaction!)
	_, err = tx.Exec(`
		UPDATE shipments 
		SET status = 'PROCESSING', updated_at = NOW() 
		WHERE order_id = $1 AND status = 'PENDING'
	`, order.ID)
	if err != nil {
		log.Printf("⚠️ Failed to update shipment status: %v", err)
		// Don't fail the whole transaction for this - shipment might not exist yet
	} else {
		log.Printf("📦 Shipment for order %d updated to PROCESSING", order.ID)
	}

	// Record history (within transaction)
	_, _ = tx.Exec(`
		INSERT INTO order_status_history (order_id, from_status, to_status, changed_by, reason)
		VALUES ($1, $2, $3, 'webhook', 'Payment success')
	`, order.ID, order.Status, models.OrderStatusPaid)

	// Send payment success email (async, outside transaction)
	if s.emailService != nil {
		go func() {
			// Reload order to get updated data
			updatedOrder, err := s.orderRepo.FindByID(order.ID)
			if err != nil {
				log.Printf("⚠️ Failed to reload order for email: %v", err)
				return
			}

			// Send payment success email only
			// Tokopedia-style: Email shipped is sent when admin actually ships the order
			if err := s.emailService.SendPaymentSuccess(updatedOrder, n.PaymentType); err != nil {
				log.Printf("⚠️ Failed to send payment success email: %v", err)
			} else {
				log.Printf("📧 Payment success email sent to %s", updatedOrder.CustomerEmail)
			}

			// NOTE: Email shipped will be sent by admin when they ship the order
			// Resi is pre-generated but customer only gets notified when actually shipped
		}()
	}

	return nil
}

// handleSuccess - legacy non-transactional version (kept for backward compatibility)
func (s *paymentService) handleSuccess(order *models.Order, payment *models.Payment, n dto.MidtransNotification) error {
	// IDEMPOTENCY
	if order.Status == models.OrderStatusPaid {
		log.Printf("Order %s already PAID", order.OrderCode)
		return nil
	}

	if order.Status != models.OrderStatusPending {
		log.Printf("Order %s not PENDING, current: %s", order.OrderCode, order.Status)
		return nil
	}

	// Update ORDER status to PAID
	err := s.orderRepo.MarkAsPaid(order.ID)
	if err != nil {
		log.Printf("Failed to mark order as paid: %v", err)
		return err
	}
	log.Printf("✅ Order %s marked as PAID", order.OrderCode)

	// Update PAYMENT status
	s.paymentRepo.MarkAsPaidWithDetails(payment.ID, n.TransactionID, n.PaymentType, map[string]any{
		"transaction_id":     n.TransactionID,
		"transaction_status": n.TransactionStatus,
		"payment_type":       n.PaymentType,
		"fraud_status":       n.FraudStatus,
	})

	// Update SHIPMENT status to PROCESSING (payment received, preparing shipment)
	// This is done via direct SQL since we don't have shippingRepo here
	s.updateShipmentToProcessing(order.ID)

	// Record history
	s.orderRepo.RecordStatusChange(order.ID, order.Status, models.OrderStatusPaid, "webhook", "Payment success")

	return nil
}

// updateShipmentToProcessing updates shipment status when payment is successful
func (s *paymentService) updateShipmentToProcessing(orderID int) {
	// Direct update since we don't have shippingRepo injected
	// In production, this should be done via event/message queue
	query := `UPDATE shipments SET status = 'PROCESSING', updated_at = NOW() WHERE order_id = $1 AND status = 'PENDING'`
	_, err := s.paymentRepo.GetDB().Exec(query, orderID)
	if err != nil {
		log.Printf("⚠️ Failed to update shipment status: %v", err)
	} else {
		log.Printf("📦 Shipment for order %d updated to PROCESSING", orderID)
	}
}

func (s *paymentService) handleExpired(order *models.Order, payment *models.Payment, n dto.MidtransNotification) error {
	if order.Status != models.OrderStatusPending {
		return nil
	}

	// Restore stock
	if order.StockReserved {
		s.orderRepo.RestoreStock(order.ID)
	}

	// Update order
	s.orderRepo.UpdateStatus(order.ID, models.OrderStatusExpired)
	log.Printf("Order %s marked as EXPIRED", order.OrderCode)

	// Update payment
	s.paymentRepo.UpdateStatusWithResponse(payment.ID, models.PaymentStatusExpired, map[string]any{
		"transaction_status": n.TransactionStatus,
	})

	return nil
}

// handleExpiredTx handles expired payment within a transaction
func (s *paymentService) handleExpiredTx(tx *sql.Tx, order *models.Order, payment *models.Payment, n dto.MidtransNotification) error {
	if order.Status != models.OrderStatusPending {
		return nil
	}

	// Restore stock within transaction
	if order.StockReserved {
		s.orderRepo.RestoreStockTx(tx, order.ID)
	}

	// Update order
	s.orderRepo.UpdateStatusTx(tx, order.ID, models.OrderStatusExpired)
	log.Printf("Order %s marked as EXPIRED", order.OrderCode)

	// Update payment
	_, _ = tx.Exec(`
		UPDATE payments SET status = $1, expired_at = NOW(), updated_at = NOW()
		WHERE id = $2
	`, models.PaymentStatusExpired, payment.ID)

	return nil
}

func (s *paymentService) handleCancelled(order *models.Order, payment *models.Payment, n dto.MidtransNotification) error {
	if order.Status != models.OrderStatusPending {
		return nil
	}

	// Restore stock
	if order.StockReserved {
		s.orderRepo.RestoreStock(order.ID)
	}

	// Update order
	s.orderRepo.MarkAsCancelled(order.ID)
	log.Printf("Order %s marked as CANCELLED", order.OrderCode)

	// Update payment
	s.paymentRepo.UpdateStatusWithResponse(payment.ID, models.PaymentStatusCancelled, map[string]any{
		"transaction_status": n.TransactionStatus,
	})

	return nil
}

// handleCancelledTx handles cancelled payment within a transaction
func (s *paymentService) handleCancelledTx(tx *sql.Tx, order *models.Order, payment *models.Payment, n dto.MidtransNotification) error {
	if order.Status != models.OrderStatusPending {
		return nil
	}

	// Restore stock within transaction
	if order.StockReserved {
		s.orderRepo.RestoreStockTx(tx, order.ID)
	}

	// Update order
	_, _ = tx.Exec(`
		UPDATE orders SET status = $1, cancelled_at = NOW(), updated_at = NOW()
		WHERE id = $2
	`, models.OrderStatusCancelled, order.ID)
	log.Printf("Order %s marked as CANCELLED", order.OrderCode)

	// Update payment
	_, _ = tx.Exec(`
		UPDATE payments SET status = $1, updated_at = NOW()
		WHERE id = $2
	`, models.PaymentStatusCancelled, payment.ID)

	return nil
}

func (s *paymentService) handleFailed(order *models.Order, payment *models.Payment, n dto.MidtransNotification) error {
	if order.Status != models.OrderStatusPending {
		return nil
	}

	// Restore stock
	if order.StockReserved {
		s.orderRepo.RestoreStock(order.ID)
	}

	// Update order
	s.orderRepo.UpdateStatus(order.ID, models.OrderStatusFailed)
	log.Printf("Order %s marked as FAILED", order.OrderCode)

	// Update payment
	s.paymentRepo.UpdateStatusWithResponse(payment.ID, models.PaymentStatusFailed, map[string]any{
		"transaction_status": n.TransactionStatus,
	})

	return nil
}

// handleFailedTx handles failed payment within a transaction
func (s *paymentService) handleFailedTx(tx *sql.Tx, order *models.Order, payment *models.Payment, n dto.MidtransNotification) error {
	if order.Status != models.OrderStatusPending {
		return nil
	}

	// Restore stock within transaction
	if order.StockReserved {
		s.orderRepo.RestoreStockTx(tx, order.ID)
	}

	// Update order
	s.orderRepo.UpdateStatusTx(tx, order.ID, models.OrderStatusFailed)
	log.Printf("Order %s marked as FAILED", order.OrderCode)

	// Update payment
	_, _ = tx.Exec(`
		UPDATE payments SET status = $1, updated_at = NOW()
		WHERE id = $2
	`, models.PaymentStatusFailed, payment.ID)

	return nil
}

// logPaymentSync logs payment sync for audit trail
func (s *paymentService) logPaymentSync(order *models.Order, payment *models.Payment, n dto.MidtransNotification, newStatus models.PaymentStatus) {
	query := `
		INSERT INTO payment_sync_log (
			payment_id, order_id, order_code, sync_type, sync_status,
			local_payment_status, local_order_status, gateway_status,
			gateway_transaction_id, has_mismatch
		) VALUES ($1, $2, $3, 'webhook', 'SYNCED', $4, $5, $6, $7, false)
	`
	_, err := s.paymentRepo.GetDB().Exec(query,
		payment.ID, order.ID, order.OrderCode,
		payment.Status, order.Status, n.TransactionStatus, n.TransactionID,
	)
	if err != nil {
		log.Printf("⚠️ Failed to log payment sync: %v", err)
	}
}

func (s *paymentService) verifySignature(n dto.MidtransNotification) bool {
	signatureInput := n.OrderID + n.StatusCode + n.GrossAmount + s.serverKey
	hash := sha512.New()
	hash.Write([]byte(signatureInput))
	calculated := hex.EncodeToString(hash.Sum(nil))
	isValid := calculated == n.SignatureKey
	if !isValid {
		log.Printf("Signature mismatch: expected %s, got %s", calculated, n.SignatureKey)
	}
	return isValid
}

func (s *paymentService) mapMidtransStatus(transactionStatus, fraudStatus string) models.PaymentStatus {
	switch transactionStatus {
	case "capture":
		if fraudStatus == "accept" {
			return models.PaymentStatusSuccess
		}
		return models.PaymentStatusPending
	case "settlement":
		return models.PaymentStatusSuccess
	case "pending":
		return models.PaymentStatusPending
	case "deny", "failure":
		return models.PaymentStatusFailed
	case "cancel":
		return models.PaymentStatusCancelled
	case "expire":
		return models.PaymentStatusExpired
	default:
		return models.PaymentStatusPending
	}
}

func (s *paymentService) isFinalPaymentStatus(status models.PaymentStatus) bool {
	switch status {
	case models.PaymentStatusSuccess, models.PaymentStatusFailed, 
		 models.PaymentStatusExpired, models.PaymentStatusCancelled:
		return true
	}
	return false
}

func truncateString(str string, maxLen int) string {
	if len(str) <= maxLen {
		return str
	}
	return str[:maxLen-3] + "..."
}

// Legacy
func (s *paymentService) CreatePayment(orderID int) (string, error) {
	return s.InitiatePayment(orderID)
}

func (s *paymentService) HandleCallback(orderCode string, status models.PaymentStatus, txID string) error {
	order, err := s.orderRepo.FindByOrderCode(orderCode)
	if err != nil {
		return err
	}
	payment, _ := s.paymentRepo.FindByOrderID(order.ID)
	if payment != nil && s.isFinalPaymentStatus(payment.Status) {
		return nil
	}

	if status == models.PaymentStatusSuccess {
		s.orderRepo.MarkAsPaid(order.ID)
		if payment != nil {
			s.paymentRepo.MarkAsPaid(payment.ID, txID)
		}
	}
	return nil
}
