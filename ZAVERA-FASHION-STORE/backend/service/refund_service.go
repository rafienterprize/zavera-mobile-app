package service

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"zavera/dto"
	"zavera/models"
	"zavera/repository"
)

var (
	ErrRefundNotFound       = errors.New("refund not found")
	ErrRefundAlreadyExists  = errors.New("refund already exists for this order")
	ErrRefundNotRefundable  = errors.New("order is not refundable")
	ErrRefundAmountExceeds  = errors.New("refund amount exceeds refundable amount")
	ErrRefundAlreadyFinal   = errors.New("refund is already in final state")
	ErrPaymentNotSettled    = errors.New("payment not settled, cannot refund")
	ErrIdempotencyConflict  = errors.New("idempotency key already used")
)

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

// Helper function to create string pointer only if not empty
func stringPtrIfNotEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

type RefundService interface {
	// Core refund operations
	CreateRefund(req *dto.RefundRequest, requestedBy *int) (*models.Refund, error)
	ProcessRefund(refundID int, processedBy int) error
	GetRefund(refundID int) (*models.Refund, error)
	GetRefundByCode(code string) (*models.Refund, error)
	GetRefundsByOrder(orderID int) ([]*models.Refund, error)
	GetRefundsByOrderCode(orderCode string) ([]*models.Refund, error)
	ListRefunds(page, pageSize int, status, orderCode string) ([]*models.Refund, int, error)
	GetOrderCodeForRefund(orderID int) string
	MarkRefundCompletedManually(refundID int, processedBy int, note string) error
	
	// Specific refund types
	FullRefund(orderCode string, reason models.RefundReason, detail string, requestedBy *int, idempotencyKey string) (*models.Refund, error)
	PartialRefund(orderCode string, amount float64, reason models.RefundReason, detail string, requestedBy *int, idempotencyKey string) (*models.Refund, error)
	ShippingOnlyRefund(orderCode string, reason models.RefundReason, detail string, requestedBy *int, idempotencyKey string) (*models.Refund, error)
	ItemRefund(orderCode string, items []dto.RefundItemRequest, reason models.RefundReason, detail string, requestedBy *int, idempotencyKey string) (*models.Refund, error)
	
	// Gateway operations
	ProcessMidtransRefund(refund *models.Refund) (*dto.MidtransRefundResponse, error)
	CheckMidtransRefundStatus(refundCode string) (*dto.MidtransRefundResponse, error)
}

type refundService struct {
	refundRepo  repository.RefundRepository
	orderRepo   repository.OrderRepository
	paymentRepo repository.PaymentRepository
	auditRepo   repository.AdminAuditRepository
	serverKey   string
	baseURL     string
}

func NewRefundService(
	refundRepo repository.RefundRepository,
	orderRepo repository.OrderRepository,
	paymentRepo repository.PaymentRepository,
	auditRepo repository.AdminAuditRepository,
) RefundService {
	baseURL := "https://api.sandbox.midtrans.com"
	if os.Getenv("MIDTRANS_ENVIRONMENT") == "production" {
		baseURL = "https://api.midtrans.com"
	}
	
	return &refundService{
		refundRepo:  refundRepo,
		orderRepo:   orderRepo,
		paymentRepo: paymentRepo,
		auditRepo:   auditRepo,
		serverKey:   os.Getenv("MIDTRANS_SERVER_KEY"),
		baseURL:     baseURL,
	}
}

func (s *refundService) CreateRefund(req *dto.RefundRequest, requestedBy *int) (*models.Refund, error) {
	// Check idempotency
	if req.IdempotencyKey != "" {
		existing, err := s.refundRepo.FindByIdempotencyKey(req.IdempotencyKey)
		if err == nil && existing != nil {
			log.Printf("Idempotency hit for refund: %s", req.IdempotencyKey)
			return existing, nil
		}
	}

	// Get order
	order, err := s.orderRepo.FindByOrderCode(req.OrderCode)
	if err != nil {
		return nil, fmt.Errorf("order not found: %s", req.OrderCode)
	}

	// Get payment (optional for manually marked orders)
	// Try Snap payment first (table: payments)
	payment, err := s.paymentRepo.FindByOrderID(order.ID)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("❌ Error finding Snap payment for order %s (ID: %d): %v", req.OrderCode, order.ID, err)
		return nil, fmt.Errorf("failed to check payment for order %s: %w", req.OrderCode, err)
	}
	
	// If Snap payment found but not SUCCESS, check Core API payment as alternative
	if payment != nil && payment.Status != models.PaymentStatusSuccess {
		log.Printf("🔍 Snap payment found but status is %s, checking Core API payment for order %s", payment.Status, req.OrderCode)
		
		// Query order_payments table directly
		var corePaymentID int
		var corePaymentStatus string
		var corePaymentAmount float64
		var corePaymentMethod string
		
		query := `
			SELECT id, payment_status, 
			       CASE 
			           WHEN payment_method = 'bca_va' THEN 'bank_transfer'
			           WHEN payment_method = 'bri_va' THEN 'bank_transfer'
			           WHEN payment_method = 'bni_va' THEN 'bank_transfer'
			           WHEN payment_method = 'mandiri_va' THEN 'bank_transfer'
			           WHEN payment_method = 'permata_va' THEN 'bank_transfer'
			           WHEN payment_method = 'qris' THEN 'qris'
			           WHEN payment_method = 'gopay' THEN 'gopay'
			           ELSE 'other'
			       END as payment_method
			FROM order_payments 
			WHERE order_id = $1
			ORDER BY created_at DESC
			LIMIT 1
		`
		
		err = s.paymentRepo.GetDB().QueryRow(query, order.ID).Scan(&corePaymentID, &corePaymentStatus, &corePaymentMethod)
		if err == nil && corePaymentStatus == "PAID" {
			// Found Core API payment with PAID status - use this instead of Snap payment
			corePaymentAmount = order.TotalAmount
			
			payment = &models.Payment{
				ID:            0, // Signal to use NULL for payment_id in refunds table
				OrderID:       order.ID,
				PaymentMethod: corePaymentMethod,
				Amount:        corePaymentAmount,
				Status:        models.PaymentStatusSuccess, // Map PAID to SUCCESS
			}
			
			log.Printf("✅ Using Core API payment instead: CorePaymentID=%d, Status=PAID, Amount=%.2f, Method=%s", 
				corePaymentID, payment.Amount, payment.PaymentMethod)
		} else {
			log.Printf("⚠️ No PAID Core API payment found, using Snap payment with status %s", payment.Status)
		}
	}
	
	// If no Snap payment found, check Core API payment (table: order_payments)
	if payment == nil {
		log.Printf("🔍 No Snap payment found, checking Core API payment for order %s", req.OrderCode)
		
		// Query order_payments table directly
		var corePaymentID int
		var corePaymentStatus string
		var corePaymentAmount float64
		var corePaymentMethod string
		
		query := `
			SELECT id, payment_status, 
			       CASE 
			           WHEN payment_method = 'bca_va' THEN 'bank_transfer'
			           WHEN payment_method = 'bri_va' THEN 'bank_transfer'
			           WHEN payment_method = 'bni_va' THEN 'bank_transfer'
			           WHEN payment_method = 'mandiri_va' THEN 'bank_transfer'
			           WHEN payment_method = 'permata_va' THEN 'bank_transfer'
			           WHEN payment_method = 'qris' THEN 'qris'
			           WHEN payment_method = 'gopay' THEN 'gopay'
			           ELSE 'other'
			       END as payment_method
			FROM order_payments 
			WHERE order_id = $1
			ORDER BY created_at DESC
			LIMIT 1
		`
		
		err = s.paymentRepo.GetDB().QueryRow(query, order.ID).Scan(&corePaymentID, &corePaymentStatus, &corePaymentMethod)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("❌ Error finding Core API payment for order %s: %v", req.OrderCode, err)
			return nil, fmt.Errorf("failed to check Core API payment for order %s: %w", req.OrderCode, err)
		}
		
		if err == nil {
			// Found Core API payment - convert to payment model for compatibility
			corePaymentAmount = order.TotalAmount // Use order total as payment amount
			
			// Map Core API status to payment status
			var mappedStatus models.PaymentStatus
			switch corePaymentStatus {
			case "PAID":
				mappedStatus = models.PaymentStatusSuccess
			case "PENDING":
				mappedStatus = models.PaymentStatusPending
			case "EXPIRED":
				mappedStatus = models.PaymentStatusExpired
			case "FAILED":
				mappedStatus = models.PaymentStatusFailed
			default:
				mappedStatus = models.PaymentStatusPending
			}
			
			// IMPORTANT: Use ID=0 for Core API payments to signal that payment_id should be NULL
			// Core API payments are in order_payments table, not payments table
			// Foreign key constraint requires payment_id to reference payments table
			payment = &models.Payment{
				ID:            0, // Signal to use NULL for payment_id in refunds table
				OrderID:       order.ID,
				PaymentMethod: corePaymentMethod,
				Amount:        corePaymentAmount,
				Status:        mappedStatus,
			}
			
			log.Printf("✅ Found Core API payment for order %s: CorePaymentID=%d, Status=%s, Amount=%.2f, Method=%s (will use NULL payment_id for refund)", 
				req.OrderCode, corePaymentID, payment.Status, payment.Amount, payment.PaymentMethod)
		}
	} else {
		log.Printf("✅ Found Snap payment for order %s: ID=%d, Status=%s, Amount=%.2f", req.OrderCode, payment.ID, payment.Status, payment.Amount)
	}
	
	if payment == nil {
		log.Printf("⚠️ No payment record found (neither Snap nor Core API) for order %s - will create manual refund", req.OrderCode)
	}

	// Validate refund request
	if err := s.validateRefundRequest(req, order, payment); err != nil {
		return nil, err
	}

	// If no payment exists, this is a manually marked order
	// Create refund record but skip gateway processing
	if payment == nil {
		log.Printf("⚠️ No payment record for order %s - creating manual refund", req.OrderCode)
		return s.createManualRefund(req, requestedBy, order)
	}

	// Calculate refund amount
	refundAmount, shippingRefund, itemsRefund, err := s.calculateRefundAmount(order, payment, req)
	if err != nil {
		return nil, err
	}

	// START TRANSACTION with row lock to prevent race condition
	tx, err := s.refundRepo.GetDB().Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Lock order row to prevent concurrent refunds
	var lockedOrderID int
	err = tx.QueryRow(`SELECT id FROM orders WHERE id = $1 FOR UPDATE`, order.ID).Scan(&lockedOrderID)
	if err != nil {
		return nil, fmt.Errorf("failed to lock order: %w", err)
	}

	// Check existing refunds WITH LOCK HELD - this is now safe from race conditions
	var totalRefunded float64
	err = tx.QueryRow(`
		SELECT COALESCE(SUM(refund_amount), 0) 
		FROM refunds 
		WHERE order_id = $1 
		AND status IN ('COMPLETED', 'PROCESSING')
	`, order.ID).Scan(&totalRefunded)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing refunds: %w", err)
	}

	refundableAmount := payment.Amount - totalRefunded
	if refundAmount > refundableAmount {
		return nil, fmt.Errorf("%w: requested %.2f, available %.2f", ErrRefundAmountExceeds, refundAmount, refundableAmount)
	}

	// Create refund record
	// For Core API payments (payment.ID == 0), use NULL for payment_id
	var paymentIDPtr *int
	if payment.ID > 0 {
		paymentIDPtr = &payment.ID
	} else {
		paymentIDPtr = nil // Core API payment - use NULL
	}
	
	refund := &models.Refund{
		RefundCode:     repository.GenerateRefundCode(),
		OrderID:        order.ID,
		PaymentID:      paymentIDPtr,
		RefundType:     models.RefundType(req.RefundType),
		Reason:         models.RefundReason(req.Reason),
		ReasonDetail:   req.ReasonDetail,
		OriginalAmount: payment.Amount,
		RefundAmount:   refundAmount,
		ShippingRefund: shippingRefund,
		ItemsRefund:    itemsRefund,
		Status:         models.RefundStatusPending,
		IdempotencyKey: stringPtrIfNotEmpty(req.IdempotencyKey),
		RequestedBy:    requestedBy,
	}

	// Create refund within transaction
	if err := s.refundRepo.CreateWithTx(tx, refund); err != nil {
		return nil, fmt.Errorf("failed to create refund: %w", err)
	}

	// Create refund items if partial/item refund
	if req.RefundType == "PARTIAL" || req.RefundType == "ITEM_ONLY" {
		for _, item := range req.Items {
			orderItem := s.findOrderItem(order.Items, item.OrderItemID)
			if orderItem == nil {
				continue
			}

			refundItem := &models.RefundItem{
				RefundID:     refund.ID,
				OrderItemID:  item.OrderItemID,
				ProductID:    orderItem.ProductID,
				ProductName:  orderItem.ProductName,
				Quantity:     item.Quantity,
				PricePerUnit: orderItem.PricePerUnit,
				RefundAmount: float64(item.Quantity) * orderItem.PricePerUnit,
				ItemReason:   item.Reason,
			}
			if err := s.refundRepo.CreateRefundItemWithTx(tx, refundItem); err != nil {
				return nil, fmt.Errorf("failed to create refund item: %w", err)
			}
		}
	}

	// Record status change within transaction
	if err := s.refundRepo.RecordStatusChangeWithTx(tx, refund.ID, "", models.RefundStatusPending, "system", "Refund created"); err != nil {
		log.Printf("⚠️ Failed to record status change: %v", err)
		// Don't fail the whole refund creation if status history fails
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	log.Printf("✅ Refund created: %s for order %s, amount: %.2f", refund.RefundCode, order.OrderCode, refundAmount)
	return refund, nil
}

func (s *refundService) ProcessRefund(refundID int, processedBy int) error {
	refund, err := s.refundRepo.FindByID(refundID)
	if err != nil {
		return ErrRefundNotFound
	}

	if refund.Status.IsFinalStatus() {
		return ErrRefundAlreadyFinal
	}

	if !refund.Status.CanProcess() {
		return fmt.Errorf("refund cannot be processed in status: %s", refund.Status)
	}

	// Update status to processing
	s.refundRepo.UpdateStatus(refundID, models.RefundStatusProcessing, nil)
	s.refundRepo.RecordStatusChange(refundID, refund.Status, models.RefundStatusProcessing, fmt.Sprintf("user:%d", processedBy), "Processing started")

	// Process with Midtrans
	resp, err := s.ProcessMidtransRefund(refund)
	if err != nil {
		log.Printf("❌ Midtrans refund failed: %v", err)
		
		// Check if error is 418 (settlement time issue)
		// For this case, we keep status as PENDING with note for manual processing
		errorMsg := err.Error()
		if strings.Contains(errorMsg, "payment provider requires additional settlement time") || 
		   strings.Contains(errorMsg, "Payment Provider doesn't allow refund within this time") {
			log.Printf("⚠️ Error 418 detected - keeping as PENDING for manual processing")
			
			// Keep as PENDING but add note for manual processing
			approvalNote := "⚠️ REQUIRES MANUAL PROCESSING: Automatic refund failed due to payment provider settlement time. Admin should process manual bank transfer and then mark as completed."
			s.refundRepo.UpdateStatus(refundID, models.RefundStatusPending, nil)
			s.refundRepo.RecordStatusChange(refundID, models.RefundStatusProcessing, models.RefundStatusPending, 
				fmt.Sprintf("user:%d", processedBy), approvalNote)
			
			// Return specific error for frontend to show manual processing option
			return fmt.Errorf("MANUAL_PROCESSING_REQUIRED: Automatic refund failed. Please process manual bank transfer to customer and mark refund as completed after transfer is done")
		}
		
		// For other errors, mark as failed
		s.refundRepo.MarkFailed(refundID, err.Error(), nil)
		s.refundRepo.RecordStatusChange(refundID, models.RefundStatusProcessing, models.RefundStatusFailed, "system", err.Error())
		return err
	}

	// Mark as completed
	gatewayResponse := map[string]any{
		"status_code":    resp.StatusCode,
		"status_message": resp.StatusMessage,
		"refund_key":     resp.RefundKey,
		"refund_amount":  resp.RefundAmount,
	}
	
	s.refundRepo.MarkCompleted(refundID, fmt.Sprintf("%d", resp.RefundChargebackID), gatewayResponse)
	s.refundRepo.RecordStatusChange(refundID, models.RefundStatusProcessing, models.RefundStatusCompleted, "system", "Refund completed via Midtrans")

	// Update order refund status
	s.updateOrderRefundStatus(refund.OrderID)

	// Restore stock for refunded items
	s.restoreRefundedStock(refund)

	log.Printf("✅ Refund completed: %s, gateway ID: %d", refund.RefundCode, resp.RefundChargebackID)
	return nil
}

// MarkRefundCompletedManually marks a refund as completed after manual bank transfer
func (s *refundService) MarkRefundCompletedManually(refundID int, processedBy int, note string) error {
	refund, err := s.refundRepo.FindByID(refundID)
	if err != nil {
		return ErrRefundNotFound
	}

	// Only allow marking PENDING refunds as completed
	if refund.Status != models.RefundStatusPending {
		return fmt.Errorf("can only mark PENDING refunds as completed, current status: %s", refund.Status)
	}

	// Mark as completed with manual gateway ID
	gatewayResponse := map[string]any{
		"manual_completion": true,
		"processed_by":      processedBy,
		"note":              note,
		"completed_at":      time.Now(),
	}
	
	s.refundRepo.MarkCompleted(refundID, "MANUAL_BANK_TRANSFER", gatewayResponse)
	s.refundRepo.RecordStatusChange(refundID, models.RefundStatusPending, models.RefundStatusCompleted, 
		fmt.Sprintf("user:%d", processedBy), fmt.Sprintf("Manual refund completed: %s", note))

	// Update order refund status
	s.updateOrderRefundStatus(refund.OrderID)

	// Restore stock for refunded items
	s.restoreRefundedStock(refund)

	log.Printf("✅ Refund manually completed: %s by user %d", refund.RefundCode, processedBy)
	return nil
}

func (s *refundService) GetRefund(refundID int) (*models.Refund, error) {
	refund, err := s.refundRepo.FindByID(refundID)
	if err != nil {
		return nil, ErrRefundNotFound
	}
	
	// Load items
	items, _ := s.refundRepo.FindItemsByRefundID(refundID)
	refund.Items = items
	
	return refund, nil
}

func (s *refundService) GetRefundByCode(code string) (*models.Refund, error) {
	refund, err := s.refundRepo.FindByCode(code)
	if err != nil {
		return nil, ErrRefundNotFound
	}
	
	items, _ := s.refundRepo.FindItemsByRefundID(refund.ID)
	refund.Items = items
	
	return refund, nil
}

func (s *refundService) GetRefundsByOrder(orderID int) ([]*models.Refund, error) {
	return s.refundRepo.FindByOrderID(orderID)
}

func (s *refundService) GetRefundsByOrderCode(orderCode string) ([]*models.Refund, error) {
	// Get order first
	order, err := s.orderRepo.FindByOrderCode(orderCode)
	if err != nil {
		return nil, fmt.Errorf("order not found: %s", orderCode)
	}
	
	// Get refunds for this order
	refunds, err := s.refundRepo.FindByOrderID(order.ID)
	if err != nil {
		return nil, err
	}
	
	// Load items for each refund
	for _, refund := range refunds {
		items, _ := s.refundRepo.FindItemsByRefundID(refund.ID)
		refund.Items = items
	}
	
	return refunds, nil
}

func (s *refundService) ListRefunds(page, pageSize int, status, orderCode string) ([]*models.Refund, int, error) {
	// Get refunds with pagination and filters
	refunds, totalCount, err := s.refundRepo.FindAll(page, pageSize, status, orderCode)
	if err != nil {
		return nil, 0, err
	}
	
	// Load items for each refund
	for _, refund := range refunds {
		items, _ := s.refundRepo.FindItemsByRefundID(refund.ID)
		refund.Items = items
	}
	
	return refunds, totalCount, nil
}

func (s *refundService) GetOrderCodeForRefund(orderID int) string {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return ""
	}
	return order.OrderCode
}

// Convenience methods for specific refund types
func (s *refundService) FullRefund(orderCode string, reason models.RefundReason, detail string, requestedBy *int, idempotencyKey string) (*models.Refund, error) {
	req := &dto.RefundRequest{
		OrderCode:      orderCode,
		RefundType:     "FULL",
		Reason:         string(reason),
		ReasonDetail:   detail,
		IdempotencyKey: idempotencyKey,
	}
	return s.CreateRefund(req, requestedBy)
}

func (s *refundService) PartialRefund(orderCode string, amount float64, reason models.RefundReason, detail string, requestedBy *int, idempotencyKey string) (*models.Refund, error) {
	req := &dto.RefundRequest{
		OrderCode:      orderCode,
		RefundType:     "PARTIAL",
		Reason:         string(reason),
		ReasonDetail:   detail,
		Amount:         &amount,
		IdempotencyKey: idempotencyKey,
	}
	return s.CreateRefund(req, requestedBy)
}

func (s *refundService) ShippingOnlyRefund(orderCode string, reason models.RefundReason, detail string, requestedBy *int, idempotencyKey string) (*models.Refund, error) {
	req := &dto.RefundRequest{
		OrderCode:      orderCode,
		RefundType:     "SHIPPING_ONLY",
		Reason:         string(reason),
		ReasonDetail:   detail,
		IdempotencyKey: idempotencyKey,
	}
	return s.CreateRefund(req, requestedBy)
}

func (s *refundService) ItemRefund(orderCode string, items []dto.RefundItemRequest, reason models.RefundReason, detail string, requestedBy *int, idempotencyKey string) (*models.Refund, error) {
	req := &dto.RefundRequest{
		OrderCode:      orderCode,
		RefundType:     "ITEM_ONLY",
		Reason:         string(reason),
		ReasonDetail:   detail,
		Items:          items,
		IdempotencyKey: idempotencyKey,
	}
	return s.CreateRefund(req, requestedBy)
}

// ProcessMidtransRefund calls Midtrans refund API
func (s *refundService) ProcessMidtransRefund(refund *models.Refund) (*dto.MidtransRefundResponse, error) {
	// Check if we should skip Midtrans refund API (for development/testing)
	skipMidtransRefund := os.Getenv("SKIP_MIDTRANS_REFUND") == "true"
	
	if skipMidtransRefund {
		log.Printf("⚠️ SKIP_MIDTRANS_REFUND=true - Bypassing Midtrans refund API for testing")
		log.Printf("   Refund Code: %s", refund.RefundCode)
		log.Printf("   Amount: %.2f", refund.RefundAmount)
		log.Printf("   ⚠️ This should ONLY be used in development/testing!")
		
		// Return mock successful response
		return &dto.MidtransRefundResponse{
			StatusCode:          "200",
			StatusMessage:       "Success (Development Mode - Midtrans API Bypassed)",
			RefundChargebackID:  999999, // Mock ID
			RefundAmount:        fmt.Sprintf("%.2f", refund.RefundAmount),
			RefundKey:           refund.RefundCode,
		}, nil
	}
	
	order, err := s.orderRepo.FindByID(refund.OrderID)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	// Get payment to get the correct external_id (with timestamp)
	// Try Snap payment first
	payment, err := s.paymentRepo.FindByOrderID(order.ID)
	
	var orderIDForMidtrans string
	
	if err != nil || payment == nil {
		// No Snap payment found, check Core API payment
		log.Printf("🔍 No Snap payment found, checking Core API payment for refund")
		
		var midtransOrderID string
		query := `SELECT midtrans_order_id FROM order_payments WHERE order_id = $1 ORDER BY created_at DESC LIMIT 1`
		err = s.paymentRepo.GetDB().QueryRow(query, order.ID).Scan(&midtransOrderID)
		
		if err != nil {
			return nil, fmt.Errorf("payment not found for order (neither Snap nor Core API): %w", err)
		}
		
		orderIDForMidtrans = midtransOrderID
		log.Printf("✅ Using Core API midtrans_order_id: %s", orderIDForMidtrans)
	} else {
		// Use Snap payment external_id
		orderIDForMidtrans = payment.ExternalID
		if orderIDForMidtrans == "" {
			// Fallback to order code if external_id is empty (shouldn't happen)
			orderIDForMidtrans = order.OrderCode
			log.Printf("⚠️ Payment external_id is empty, using order code: %s", orderIDForMidtrans)
		}
		log.Printf("✅ Using Snap payment external_id: %s", orderIDForMidtrans)
	}

	// Midtrans refund endpoint
	url := fmt.Sprintf("%s/v2/%s/refund", s.baseURL, orderIDForMidtrans)
	
	log.Printf("🔄 Calling Midtrans Refund API:")
	log.Printf("   URL: %s", url)
	log.Printf("   Order: %s", order.OrderCode)
	log.Printf("   Midtrans Order ID: %s", orderIDForMidtrans)
	log.Printf("   Refund Code: %s", refund.RefundCode)
	log.Printf("   Amount: %.2f", refund.RefundAmount)

	reqBody := dto.MidtransRefundRequest{
		RefundKey: refund.RefundCode,
		Amount:    refund.RefundAmount,
		Reason:    string(refund.Reason),
	}

	jsonBody, _ := json.Marshal(reqBody)
	log.Printf("   Request Body: %s", string(jsonBody))
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	// Set auth header
	auth := base64.StdEncoding.EncodeToString([]byte(s.serverKey + ":"))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ Midtrans API request failed: %v", err)
		return nil, fmt.Errorf("midtrans request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("   Response Status: %d", resp.StatusCode)
	log.Printf("   Response Body: %s", string(body))
	
	var refundResp dto.MidtransRefundResponse
	if err := json.Unmarshal(body, &refundResp); err != nil {
		log.Printf("❌ Failed to parse Midtrans response: %v", err)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check status
	if refundResp.StatusCode != "200" && refundResp.StatusCode != "201" {
		log.Printf("❌ Midtrans refund failed: %s - %s", refundResp.StatusCode, refundResp.StatusMessage)
		
		// Map Midtrans error codes to user-friendly messages
		userFriendlyError := s.mapMidtransRefundError(refundResp.StatusCode, refundResp.StatusMessage)
		return nil, fmt.Errorf("%s", userFriendlyError)
	}

	log.Printf("✅ Midtrans refund successful!")
	log.Printf("   Refund Chargeback ID: %d", refundResp.RefundChargebackID)
	log.Printf("   Status: %s", refundResp.StatusMessage)
	
	return &refundResp, nil
}

// mapMidtransRefundError maps Midtrans error codes to user-friendly messages
func (s *refundService) mapMidtransRefundError(statusCode, originalMessage string) string {
	switch statusCode {
	case "418":
		// Payment provider doesn't allow refund within this time
		// Check if order is old enough (> 24 hours) - if yes, suggest manual refund
		return "Refund cannot be processed automatically. The payment provider requires additional settlement time. For orders older than 24 hours, please contact support for manual refund processing, or try again in a few hours."
	
	case "404":
		// Transaction doesn't exist
		return "Transaction not found in payment gateway. The payment may have been cancelled or expired. Please verify the transaction status before attempting a refund."
	
	case "412":
		// Transaction already refunded
		return "This transaction has already been refunded. Please check the refund history to avoid duplicate refunds."
	
	case "413":
		// Refund amount exceeds transaction amount
		return "Refund amount exceeds the original transaction amount. Please verify the refund amount and try again."
	
	case "500":
		// Internal server error from Midtrans
		return "Payment gateway is experiencing technical difficulties. Please try again in a few minutes or contact support if the issue persists."
	
	case "503":
		// Service unavailable
		return "Payment gateway is temporarily unavailable. Please try again in a few minutes."
	
	case "401":
		// Unauthorized
		return "Payment gateway authentication failed. Please contact technical support to verify API credentials."
	
	case "400":
		// Bad request
		return "Invalid refund request. Please verify all refund details and try again. If the issue persists, contact support."
	
	default:
		// Unknown error - return original message with context
		return fmt.Sprintf("Refund failed: %s. Please contact support if you need assistance. (Error code: %s)", originalMessage, statusCode)
	}
}

func (s *refundService) CheckMidtransRefundStatus(refundCode string) (*dto.MidtransRefundResponse, error) {
	refund, err := s.refundRepo.FindByCode(refundCode)
	if err != nil {
		return nil, ErrRefundNotFound
	}

	order, err := s.orderRepo.FindByID(refund.OrderID)
	if err != nil {
		return nil, fmt.Errorf("order not found")
	}

	url := fmt.Sprintf("%s/v2/%s/status", s.baseURL, order.OrderCode)
	
	req, _ := http.NewRequest("GET", url, nil)
	auth := base64.StdEncoding.EncodeToString([]byte(s.serverKey + ":"))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	
	var statusResp dto.MidtransRefundResponse
	json.Unmarshal(body, &statusResp)

	return &statusResp, nil
}

// Helper methods

// validateRefundRequest validates all refund request requirements
// Validates: Requirements 15.1, 15.2, 15.3, 15.4, 15.5, 15.6, 15.7
func (s *refundService) validateRefundRequest(req *dto.RefundRequest, order *models.Order, payment *models.Payment) error {
	// Requirement 15.1: Verify order status is DELIVERED or COMPLETED
	if !s.isOrderRefundable(order) {
		return fmt.Errorf("%w: order status is %s, must be DELIVERED or COMPLETED", ErrRefundNotRefundable, order.Status)
	}

	// Requirement 15.2: Verify payment status is SUCCESS (if payment exists)
	if payment != nil && payment.Status != models.PaymentStatusSuccess {
		return fmt.Errorf("%w: payment status is %s, must be SUCCESS", ErrPaymentNotSettled, payment.Status)
	}

	// Requirement 15.3: Verify refund amount is positive and non-zero
	switch req.RefundType {
	case "PARTIAL":
		if req.Amount == nil || *req.Amount <= 0 {
			return fmt.Errorf("refund amount must be positive and non-zero, got: %v", req.Amount)
		}
	case "ITEM_ONLY":
		if len(req.Items) == 0 {
			return fmt.Errorf("item refund requires at least one item")
		}
		// Validate each item
		for i, item := range req.Items {
			// Requirement 15.6: Verify refund quantities are positive
			if item.Quantity <= 0 {
				return fmt.Errorf("item %d: quantity must be positive, got: %d", i, item.Quantity)
			}
			
			// Requirement 15.5: Verify all specified order items exist
			orderItem := s.findOrderItem(order.Items, item.OrderItemID)
			if orderItem == nil {
				return fmt.Errorf("item %d: order item ID %d not found in order %s", i, item.OrderItemID, order.OrderCode)
			}
			
			// Requirement 15.6: Verify refund quantities do not exceed ordered quantities
			if item.Quantity > orderItem.Quantity {
				return fmt.Errorf("item %d: refund quantity %d exceeds ordered quantity %d for product %s", 
					i, item.Quantity, orderItem.Quantity, orderItem.ProductName)
			}
		}
	}

	// Note: Requirement 15.4 (refund amount does not exceed refundable balance) 
	// is validated in CreateRefund after calculating the refund amount and checking existing refunds

	return nil
}

func (s *refundService) isOrderRefundable(order *models.Order) bool {
	// Can refund delivered or completed orders
	switch order.Status {
	case models.OrderStatusDelivered, models.OrderStatusCompleted:
		return true
	}
	return false
}

// calculateRefundAmount calculates refund amounts based on refund type
// Validates: Requirements 8.1, 8.2, 8.3, 8.4, 8.5, 8.6, 8.7
func (s *refundService) calculateRefundAmount(order *models.Order, payment *models.Payment, req *dto.RefundRequest) (total, shipping, items float64, err error) {
	switch req.RefundType {
	case "FULL":
		// Requirement 8.1: FULL refund = entire order total amount
		// Requirement 8.7: Show items_refund and shipping_refund separately
		return payment.Amount, order.ShippingCost, order.Subtotal, nil
		
	case "SHIPPING_ONLY":
		// Requirement 8.2: SHIPPING_ONLY refund = only shipping cost
		return order.ShippingCost, order.ShippingCost, 0, nil
		
	case "PARTIAL":
		// Requirement 8.3: PARTIAL refund = specified amount up to refundable balance
		if req.Amount == nil || *req.Amount <= 0 {
			return 0, 0, 0, fmt.Errorf("partial refund requires positive amount")
		}
		// Note: Validation against refundable balance happens in CreateRefund
		return *req.Amount, 0, *req.Amount, nil
		
	case "ITEM_ONLY":
		// Requirement 8.4: ITEM_ONLY refund = sum of (quantity × price_per_unit) for selected items
		itemsTotal := 0.0
		for _, item := range req.Items {
			orderItem := s.findOrderItem(order.Items, item.OrderItemID)
			if orderItem != nil {
				itemsTotal += float64(item.Quantity) * orderItem.PricePerUnit
			}
		}
		if itemsTotal <= 0 {
			return 0, 0, 0, fmt.Errorf("item refund amount must be positive")
		}
		return itemsTotal, 0, itemsTotal, nil
	}
	
	return 0, 0, 0, fmt.Errorf("invalid refund type: %s", req.RefundType)
}

func (s *refundService) findOrderItem(items []models.OrderItem, itemID int) *models.OrderItem {
	for _, item := range items {
		if item.ID == itemID {
			return &item
		}
	}
	return nil
}

func (s *refundService) updateOrderRefundStatus(orderID int) {
	refunds, err := s.refundRepo.FindByOrderID(orderID)
	if err != nil {
		log.Printf("⚠️ Failed to find refunds for order %d: %v", orderID, err)
		return
	}
	
	totalRefunded := 0.0
	for _, r := range refunds {
		if r.Status == models.RefundStatusCompleted {
			totalRefunded += r.RefundAmount
		}
	}

	log.Printf("📊 Updating order %d refund status: total refunded = %.2f", orderID, totalRefunded)

	// Get order to check if fully refunded
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		log.Printf("⚠️ Failed to find order %d: %v", orderID, err)
		return
	}

	s.updateOrderRefundStatusWithAmount(orderID, order.TotalAmount, totalRefunded)
}

func (s *refundService) updateOrderRefundStatusWithAmount(orderID int, orderTotal, totalRefunded float64) {
	log.Printf("📊 Updating order %d refund status: total refunded = %.2f, order total = %.2f", orderID, totalRefunded, orderTotal)

	// Update order refund columns
	query := `
		UPDATE orders SET 
			refund_status = CASE WHEN $1 >= total_amount THEN 'FULL' ELSE 'PARTIAL' END,
			refund_amount = $1,
			refunded_at = NOW()
		WHERE id = $2
	`
	result, err := s.refundRepo.GetDB().Exec(query, totalRefunded, orderID)
	if err != nil {
		log.Printf("⚠️ Failed to update order refund status: %v", err)
		return
	}
	
	rowsAffected, _ := result.RowsAffected()
	refundStatus := "PARTIAL"
	if totalRefunded >= orderTotal {
		refundStatus = "FULL"
	}
	
	log.Printf("✅ Updated order %d refund status: %d rows affected, refund_status=%s, refund_amount=%.2f", 
		orderID, rowsAffected, refundStatus, totalRefunded)

	// If fully refunded, update order status to REFUNDED
	if totalRefunded >= orderTotal {
		order, _ := s.orderRepo.FindByID(orderID)
		if order != nil {
			s.orderRepo.MarkAsRefunded(orderID)
			s.orderRepo.RecordStatusChange(orderID, order.Status, models.OrderStatusRefunded, "system", "Full refund completed")
			log.Printf("✅ Order %d marked as REFUNDED", orderID)
		}
	}
}

func (s *refundService) restoreRefundedStock(refund *models.Refund) {
	if refund.RefundType == models.RefundTypeFull {
		// Restore all stock via order
		s.orderRepo.RestoreStock(refund.OrderID)
		return
	}

	// Restore specific items
	items, _ := s.refundRepo.FindItemsByRefundID(refund.ID)
	for _, item := range items {
		if item.StockRestored {
			continue
		}
		
		query := `UPDATE products SET stock = stock + $1 WHERE id = $2`
		_, err := s.refundRepo.GetDB().Exec(query, item.Quantity, item.ProductID)
		if err == nil {
			s.refundRepo.MarkItemStockRestored(item.ID)
		}
	}
}

// createManualRefund creates a refund for orders without payment records (manually marked as paid)
// Validates: Requirements 2.7, 13.2, 13.3, 13.4, 13.5
func (s *refundService) createManualRefund(req *dto.RefundRequest, requestedBy *int, order *models.Order) (*models.Refund, error) {
	// Check existing refunds to prevent over-refunding
	existingRefunds, _ := s.refundRepo.FindByOrderID(order.ID)
	totalRefunded := 0.0
	for _, r := range existingRefunds {
		if r.Status == models.RefundStatusCompleted || r.Status == models.RefundStatusProcessing {
			totalRefunded += r.RefundAmount
		}
	}
	
	// Calculate refund amount based on order totals (Requirement 13.5)
	var refundAmount, shippingRefund, itemsRefund float64
	
	switch req.RefundType {
	case "FULL":
		refundAmount = order.TotalAmount
		shippingRefund = order.ShippingCost
		itemsRefund = order.Subtotal
	case "SHIPPING_ONLY":
		refundAmount = order.ShippingCost
		shippingRefund = order.ShippingCost
		itemsRefund = 0
	case "PARTIAL":
		if req.Amount == nil || *req.Amount <= 0 {
			return nil, fmt.Errorf("partial refund requires positive amount")
		}
		// Validate amount doesn't exceed order total
		if *req.Amount > order.TotalAmount {
			return nil, fmt.Errorf("%w: requested %.2f, order total %.2f", ErrRefundAmountExceeds, *req.Amount, order.TotalAmount)
		}
		refundAmount = *req.Amount
		shippingRefund = 0
		itemsRefund = refundAmount
	case "ITEM_ONLY":
		// Calculate item refund amount
		itemsTotal := 0.0
		for _, item := range req.Items {
			orderItem := s.findOrderItem(order.Items, item.OrderItemID)
			if orderItem != nil {
				itemsTotal += float64(item.Quantity) * orderItem.PricePerUnit
			}
		}
		if itemsTotal <= 0 {
			return nil, fmt.Errorf("item refund amount must be positive")
		}
		refundAmount = itemsTotal
		shippingRefund = 0
		itemsRefund = itemsTotal
	default:
		return nil, fmt.Errorf("invalid refund type: %s", req.RefundType)
	}

	// Validate refund amount doesn't exceed available balance
	refundableAmount := order.TotalAmount - totalRefunded
	if refundAmount > refundableAmount {
		return nil, fmt.Errorf("%w: requested %.2f, available %.2f (already refunded %.2f)", 
			ErrRefundAmountExceeds, refundAmount, refundableAmount, totalRefunded)
	}

	// START TRANSACTION for manual refund
	tx, err := s.refundRepo.GetDB().Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Create refund record
	// Requirement 13.3: Set status to COMPLETED immediately
	// Requirement 13.4: Set gateway_refund_id to "MANUAL_REFUND"
	refund := &models.Refund{
		RefundCode:      repository.GenerateRefundCode(),
		OrderID:         order.ID,
		PaymentID:       nil, // Requirement 2.7: NULL for manual refunds
		RefundType:      models.RefundType(req.RefundType),
		OriginalAmount:  order.TotalAmount,
		RefundAmount:    refundAmount,
		ShippingRefund:  shippingRefund,
		ItemsRefund:     itemsRefund,
		Reason:          models.RefundReason(req.Reason),
		ReasonDetail:    req.ReasonDetail,
		Status:          models.RefundStatusCompleted, // Requirement 13.3: Auto-complete
		RequestedBy:     requestedBy,
		IdempotencyKey:  stringPtrIfNotEmpty(req.IdempotencyKey),
		ProcessedBy:     requestedBy,
		GatewayRefundID: stringPtr("MANUAL_REFUND"), // Requirement 13.4
	}

	if err := s.refundRepo.CreateWithTx(tx, refund); err != nil {
		return nil, fmt.Errorf("failed to create manual refund: %w", err)
	}

	// Create refund items if item refund
	if req.RefundType == "ITEM_ONLY" {
		for _, item := range req.Items {
			orderItem := s.findOrderItem(order.Items, item.OrderItemID)
			if orderItem == nil {
				continue
			}

			refundItem := &models.RefundItem{
				RefundID:     refund.ID,
				OrderItemID:  item.OrderItemID,
				ProductID:    orderItem.ProductID,
				ProductName:  orderItem.ProductName,
				Quantity:     item.Quantity,
				PricePerUnit: orderItem.PricePerUnit,
				RefundAmount: float64(item.Quantity) * orderItem.PricePerUnit,
				ItemReason:   item.Reason,
			}
			if err := s.refundRepo.CreateRefundItemWithTx(tx, refundItem); err != nil {
				return nil, fmt.Errorf("failed to create refund item: %w", err)
			}
		}
	}

	// Record status change within transaction
	if err := s.refundRepo.RecordStatusChangeWithTx(tx, refund.ID, "", models.RefundStatusCompleted, "system", "Manual refund created and completed"); err != nil {
		log.Printf("⚠️ Failed to record status change: %v", err)
		// Don't fail the whole refund creation if status history fails
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit manual refund transaction: %w", err)
	}

	log.Printf("✅ Manual refund created: %s for order %s (amount: %.2f) - no gateway processing", 
		refund.RefundCode, order.OrderCode, refundAmount)

	// Update order status AFTER commit (Requirement 13.7: Still update order status for manual refunds)
	// Pass the refund amount directly to avoid race condition
	s.updateOrderRefundStatusWithAmount(order.ID, order.TotalAmount, refundAmount)

	// Restore stock (Requirement 13.7: Still restore stock for manual refunds)
	s.restoreRefundedStock(refund)
	
	return refund, nil
}
