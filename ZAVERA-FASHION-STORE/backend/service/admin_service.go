package service

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"zavera/dto"
	"zavera/models"
	"zavera/repository"
)

var (
	ErrAdminActionFailed    = errors.New("admin action failed")
	ErrOrderNotCancellable  = errors.New("order cannot be cancelled")
	ErrOrderNotRefundable   = errors.New("order cannot be refunded")
	ErrOrderNotReshippable  = errors.New("order cannot be reshipped")
	ErrPaymentNotReconcilable = errors.New("payment cannot be reconciled")
)

type AdminService interface {
	// Force actions
	ForceCancel(orderCode string, req *dto.ForceCancelRequest, admin *AdminContext) (*dto.AdminActionResponse, error)
	ForceRefund(orderCode string, req *dto.ForceRefundRequest, admin *AdminContext) (*dto.AdminActionResponse, error)
	ForceReship(orderCode string, req *dto.ForceReshipRequest, admin *AdminContext) (*dto.AdminActionResponse, error)
	ReconcilePayment(paymentID int, req *dto.ReconcilePaymentRequest, admin *AdminContext) (*dto.AdminActionResponse, error)
	
	// Dashboard
	GetDashboardMetrics() (*dto.DashboardMetricsResponse, error)
	
	// Audit
	GetAuditLogs(targetType string, targetID int) ([]*models.AdminAuditLog, error)
	GetRecentAuditLogs(limit int) ([]*models.AdminAuditLog, error)
}

type AdminContext struct {
	UserID    int
	Email     string
	IP        string
	UserAgent string
}

type adminService struct {
	orderRepo    repository.OrderRepository
	paymentRepo  repository.PaymentRepository
	refundRepo   repository.RefundRepository
	auditRepo    repository.AdminAuditRepository
	shippingRepo repository.ShippingRepository
	refundSvc    RefundService
	db           *sql.DB
}

func NewAdminService(
	orderRepo repository.OrderRepository,
	paymentRepo repository.PaymentRepository,
	refundRepo repository.RefundRepository,
	auditRepo repository.AdminAuditRepository,
	shippingRepo repository.ShippingRepository,
	refundSvc RefundService,
	db *sql.DB,
) AdminService {
	return &adminService{
		orderRepo:    orderRepo,
		paymentRepo:  paymentRepo,
		refundRepo:   refundRepo,
		auditRepo:    auditRepo,
		shippingRepo: shippingRepo,
		refundSvc:    refundSvc,
		db:           db,
	}
}

// ForceCancel cancels an order with admin override
func (s *adminService) ForceCancel(orderCode string, req *dto.ForceCancelRequest, admin *AdminContext) (*dto.AdminActionResponse, error) {
	// Check idempotency
	if req.IdempotencyKey != "" {
		existing, err := s.auditRepo.FindByIdempotencyKey(req.IdempotencyKey)
		if err == nil && existing != nil {
			return &dto.AdminActionResponse{
				Success:    existing.Success,
				Message:    "Action already processed (idempotency)",
				AuditLogID: existing.ID,
			}, nil
		}
	}

	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	
	// Ensure rollback on error
	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	// Get order with lock within transaction
	query := `
		SELECT id, order_code, user_id, customer_name, customer_email, customer_phone,
		       subtotal, shipping_cost, tax, discount, total_amount, status,
		       COALESCE(stock_reserved, true) as stock_reserved,
		       COALESCE(resi, '') as resi,
		       COALESCE(origin_city, 'Semarang') as origin_city,
		       COALESCE(destination_city, '') as destination_city,
		       notes, metadata, created_at, updated_at, 
		       paid_at, shipped_at, delivered_at, completed_at, cancelled_at
		FROM orders
		WHERE order_code = $1
		FOR UPDATE
	`
	
	var order models.Order
	var metadataJSON []byte
	err = tx.QueryRow(query, orderCode).Scan(
		&order.ID, &order.OrderCode, &order.UserID, &order.CustomerName,
		&order.CustomerEmail, &order.CustomerPhone, &order.Subtotal, &order.ShippingCost,
		&order.Tax, &order.Discount, &order.TotalAmount, &order.Status, &order.StockReserved,
		&order.Resi, &order.OriginCity, &order.DestinationCity,
		&order.Notes, &metadataJSON, &order.CreatedAt, &order.UpdatedAt,
		&order.PaidAt, &order.ShippedAt, &order.DeliveredAt, &order.CompletedAt, &order.CancelledAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return s.logFailedAction(tx, admin, models.AdminActionForceCancel, "order", 0, orderCode, req.IdempotencyKey, "Order not found")
		}
		return nil, fmt.Errorf("failed to query order: %w", err)
	}
	
	// Parse metadata
	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &order.Metadata)
	}

	// Capture state before
	stateBefore := s.captureOrderState(&order)

	// Check if cancellable
	if order.Status.IsFinalStatus() {
		return s.logFailedAction(tx, admin, models.AdminActionForceCancel, "order", order.ID, orderCode, req.IdempotencyKey, 
			fmt.Sprintf("Order in final status: %s", order.Status))
	}

	// Cancel order
	if err := s.orderRepo.UpdateStatusTx(tx, order.ID, models.OrderStatusCancelled); err != nil {
		return s.logFailedAction(tx, admin, models.AdminActionForceCancel, "order", order.ID, orderCode, req.IdempotencyKey, err.Error())
	}

	// Restore stock if requested
	if req.RestoreStock && order.StockReserved {
		if err := s.orderRepo.RestoreStockTx(tx, order.ID); err != nil {
			log.Printf("⚠️ Failed to restore stock: %v", err)
		}
	}

	// Capture state after
	stateAfter := map[string]any{
		"status":         models.OrderStatusCancelled,
		"stock_restored": req.RestoreStock,
	}

	// Create audit log
	// Use nil for admin_user_id if not available (will be NULL in database)
	var adminUserID *int
	if admin.UserID > 0 {
		adminUserID = &admin.UserID
	}
	
	auditLog := &models.AdminAuditLog{
		AdminUserID:    adminUserID,
		AdminEmail:     admin.Email,
		AdminIP:        admin.IP,
		AdminUserAgent: admin.UserAgent,
		ActionType:     models.AdminActionForceCancel,
		ActionDetail:   fmt.Sprintf("Force cancelled order: %s. Reason: %s", orderCode, req.Reason),
		TargetType:     "order",
		TargetID:       order.ID,
		TargetCode:     orderCode,
		StateBefore:    stateBefore,
		StateAfter:     stateAfter,
		Success:        true,
		IdempotencyKey: req.IdempotencyKey,
	}

	if err := s.auditRepo.CreateWithTx(tx, auditLog); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	committed = true

	// Record status change after commit (non-critical, don't fail if this errors)
	if err := s.orderRepo.RecordStatusChange(order.ID, order.Status, models.OrderStatusCancelled, admin.Email, req.Reason); err != nil {
		log.Printf("⚠️ Failed to record status change: %v", err)
	}

	log.Printf("✅ Admin force cancel: order %s by %s", orderCode, admin.Email)

	return &dto.AdminActionResponse{
		Success:     true,
		Message:     fmt.Sprintf("Order %s cancelled successfully", orderCode),
		AuditLogID:  auditLog.ID,
		StateBefore: stateBefore,
		StateAfter:  stateAfter,
	}, nil
}

// ForceRefund processes a refund with admin override
func (s *adminService) ForceRefund(orderCode string, req *dto.ForceRefundRequest, admin *AdminContext) (*dto.AdminActionResponse, error) {
	// Check idempotency
	if req.IdempotencyKey != "" {
		existing, err := s.auditRepo.FindByIdempotencyKey(req.IdempotencyKey)
		if err == nil && existing != nil {
			return &dto.AdminActionResponse{
				Success:    existing.Success,
				Message:    "Action already processed (idempotency)",
				AuditLogID: existing.ID,
			}, nil
		}
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	order, err := s.orderRepo.FindByOrderCode(orderCode)
	if err != nil {
		return s.logFailedAction(tx, admin, models.AdminActionForceRefund, "order", 0, orderCode, req.IdempotencyKey, "Order not found")
	}

	stateBefore := s.captureOrderState(order)

	// Get payment (optional - may not exist for manually marked orders)
	payment, err := s.paymentRepo.FindByOrderID(order.ID)
	if err != nil && err != sql.ErrNoRows {
		return s.logFailedAction(tx, admin, models.AdminActionForceRefund, "order", order.ID, orderCode, req.IdempotencyKey, "Failed to check payment")
	}

	// If no payment exists, this is a manual order - skip gateway refund
	skipGateway := req.SkipGateway || payment == nil
	if payment == nil {
		log.Printf("⚠️ No payment record found for order %s - will skip gateway refund", orderCode)
	}

	// Create refund request
	refundReq := &dto.RefundRequest{
		OrderCode:      orderCode,
		RefundType:     req.RefundType,
		Reason:         req.Reason,
		Amount:         req.Amount,
		Items:          req.Items,
		IdempotencyKey: req.IdempotencyKey + "-refund",
	}

	// Create refund using standard service (will validate order status)
	refund, err := s.refundSvc.CreateRefund(refundReq, &admin.UserID)
	if err != nil {
		return s.logFailedAction(tx, admin, models.AdminActionForceRefund, "order", order.ID, orderCode, req.IdempotencyKey, err.Error())
	}

	// Process refund (skip gateway if requested or no payment exists)
	if !skipGateway {
		if err := s.refundSvc.ProcessRefund(refund.ID, admin.UserID); err != nil {
			// Log but don't fail - refund is created
			log.Printf("⚠️ Gateway refund failed: %v", err)
		}
	} else {
		// Manual reconciliation - mark as completed without gateway
		reason := "Admin skip gateway"
		if payment == nil {
			reason = "No payment record - manual order"
		}
		s.refundRepo.UpdateStatus(refund.ID, models.RefundStatusCompleted, map[string]any{
			"manual":      true,
			"admin":       admin.Email,
			"skip_reason": reason,
		})
	}

	stateAfter := map[string]any{
		"refund_code":   refund.RefundCode,
		"refund_amount": refund.RefundAmount,
		"refund_type":   refund.RefundType,
		"skip_gateway":  req.SkipGateway,
	}

	// Use nil for admin_user_id if not available
	var adminUserID *int
	if admin.UserID > 0 {
		adminUserID = &admin.UserID
	}
	
	auditLog := &models.AdminAuditLog{
		AdminUserID:    adminUserID,
		AdminEmail:     admin.Email,
		AdminIP:        admin.IP,
		AdminUserAgent: admin.UserAgent,
		ActionType:     models.AdminActionForceRefund,
		ActionDetail:   fmt.Sprintf("Force refund for order: %s. Amount: %.2f. Reason: %s", orderCode, refund.RefundAmount, req.Reason),
		TargetType:     "order",
		TargetID:       order.ID,
		TargetCode:     orderCode,
		StateBefore:    stateBefore,
		StateAfter:     stateAfter,
		Success:        true,
		IdempotencyKey: req.IdempotencyKey,
		Metadata: map[string]any{
			"refund_id":   refund.ID,
			"payment_id":  payment.ID,
		},
	}

	if err := s.auditRepo.CreateWithTx(tx, auditLog); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	log.Printf("✅ Admin force refund: order %s, refund %s by %s", orderCode, refund.RefundCode, admin.Email)

	return &dto.AdminActionResponse{
		Success:     true,
		Message:     fmt.Sprintf("Refund %s created for order %s", refund.RefundCode, orderCode),
		AuditLogID:  auditLog.ID,
		StateBefore: stateBefore,
		StateAfter:  stateAfter,
	}, nil
}

// ForceReship creates a replacement shipment
func (s *adminService) ForceReship(orderCode string, req *dto.ForceReshipRequest, admin *AdminContext) (*dto.AdminActionResponse, error) {
	if req.IdempotencyKey != "" {
		existing, err := s.auditRepo.FindByIdempotencyKey(req.IdempotencyKey)
		if err == nil && existing != nil {
			return &dto.AdminActionResponse{
				Success:    existing.Success,
				Message:    "Action already processed (idempotency)",
				AuditLogID: existing.ID,
			}, nil
		}
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	order, err := s.orderRepo.FindByOrderCode(orderCode)
	if err != nil {
		return s.logFailedAction(tx, admin, models.AdminActionForceReship, "order", 0, orderCode, req.IdempotencyKey, "Order not found")
	}

	// Get existing shipment
	shipment, err := s.shippingRepo.FindByOrderID(order.ID)
	if err != nil {
		return s.logFailedAction(tx, admin, models.AdminActionForceReship, "order", order.ID, orderCode, req.IdempotencyKey, "Shipment not found")
	}

	stateBefore := map[string]any{
		"shipment_id":      shipment.ID,
		"tracking_number":  shipment.TrackingNumber,
		"status":           shipment.Status,
		"reship_count":     0, // Will be updated
	}

	// Create replacement shipment
	newShipment := &models.Shipment{
		OrderID:             order.ID,
		ProviderCode:        shipment.ProviderCode,
		ProviderName:        shipment.ProviderName,
		ServiceCode:         shipment.ServiceCode,
		ServiceName:         shipment.ServiceName,
		Cost:                0, // No additional cost for reship
		ETD:                 shipment.ETD,
		Weight:              shipment.Weight,
		TrackingNumber:      req.NewTrackingNo,
		Status:              models.ShipmentStatusProcessing,
		OriginCityID:        shipment.OriginCityID,
		OriginCityName:      shipment.OriginCityName,
		DestinationCityID:   shipment.DestinationCityID,
		DestinationCityName: shipment.DestinationCityName,
	}

	// Insert new shipment with reship metadata
	query := `
		INSERT INTO shipments (
			order_id, provider_code, provider_name, service_code, service_name,
			cost, etd, weight, tracking_number, status,
			origin_city_id, origin_city_name, destination_city_id, destination_city_name,
			is_replacement, original_shipment_id, reship_reason
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, true, $15, $16)
		RETURNING id
	`

	var newShipmentID int
	err = tx.QueryRow(query,
		newShipment.OrderID, newShipment.ProviderCode, newShipment.ProviderName,
		newShipment.ServiceCode, newShipment.ServiceName, newShipment.Cost,
		newShipment.ETD, newShipment.Weight, newShipment.TrackingNumber,
		newShipment.Status, newShipment.OriginCityID, newShipment.OriginCityName,
		newShipment.DestinationCityID, newShipment.DestinationCityName,
		shipment.ID, req.Reason,
	).Scan(&newShipmentID)

	if err != nil {
		return s.logFailedAction(tx, admin, models.AdminActionForceReship, "order", order.ID, orderCode, req.IdempotencyKey, err.Error())
	}

	// Update original shipment reship count
	updateQuery := `UPDATE shipments SET reship_count = reship_count + 1 WHERE id = $1`
	tx.Exec(updateQuery, shipment.ID)

	stateAfter := map[string]any{
		"new_shipment_id":    newShipmentID,
		"new_tracking":       req.NewTrackingNo,
		"original_shipment":  shipment.ID,
		"is_replacement":     true,
	}

	// Use nil for admin_user_id if not available
	var adminUserID *int
	if admin.UserID > 0 {
		adminUserID = &admin.UserID
	}
	
	auditLog := &models.AdminAuditLog{
		AdminUserID:    adminUserID,
		AdminEmail:     admin.Email,
		AdminIP:        admin.IP,
		AdminUserAgent: admin.UserAgent,
		ActionType:     models.AdminActionForceReship,
		ActionDetail:   fmt.Sprintf("Force reship for order: %s. Reason: %s", orderCode, req.Reason),
		TargetType:     "shipment",
		TargetID:       newShipmentID,
		TargetCode:     orderCode,
		StateBefore:    stateBefore,
		StateAfter:     stateAfter,
		Success:        true,
		IdempotencyKey: req.IdempotencyKey,
	}

	if err := s.auditRepo.CreateWithTx(tx, auditLog); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	log.Printf("✅ Admin force reship: order %s, new shipment %d by %s", orderCode, newShipmentID, admin.Email)

	return &dto.AdminActionResponse{
		Success:     true,
		Message:     fmt.Sprintf("Replacement shipment created for order %s", orderCode),
		AuditLogID:  auditLog.ID,
		StateBefore: stateBefore,
		StateAfter:  stateAfter,
	}, nil
}

// ReconcilePayment manually reconciles a payment
func (s *adminService) ReconcilePayment(paymentID int, req *dto.ReconcilePaymentRequest, admin *AdminContext) (*dto.AdminActionResponse, error) {
	if req.IdempotencyKey != "" {
		existing, err := s.auditRepo.FindByIdempotencyKey(req.IdempotencyKey)
		if err == nil && existing != nil {
			return &dto.AdminActionResponse{
				Success:    existing.Success,
				Message:    "Action already processed (idempotency)",
				AuditLogID: existing.ID,
			}, nil
		}
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Get payment
	payment, err := s.paymentRepo.FindByOrderID(paymentID)
	if err != nil {
		// Try by payment ID directly
		query := `SELECT id, order_id, status, amount FROM payments WHERE id = $1`
		var p struct {
			ID      int
			OrderID int
			Status  string
			Amount  float64
		}
		err = s.db.QueryRow(query, paymentID).Scan(&p.ID, &p.OrderID, &p.Status, &p.Amount)
		if err != nil {
			return s.logFailedAction(tx, admin, models.AdminActionReconcilePayment, "payment", paymentID, "", req.IdempotencyKey, "Payment not found")
		}
		payment = &models.Payment{ID: p.ID, OrderID: p.OrderID, Status: models.PaymentStatus(p.Status), Amount: p.Amount}
	}

	order, _ := s.orderRepo.FindByID(payment.OrderID)
	orderCode := ""
	if order != nil {
		orderCode = order.OrderCode
	}

	stateBefore := map[string]any{
		"payment_status": payment.Status,
		"order_status":   "",
	}
	if order != nil {
		stateBefore["order_status"] = order.Status
	}

	var newPaymentStatus models.PaymentStatus
	var newOrderStatus models.OrderStatus

	switch req.Action {
	case "MARK_PAID":
		newPaymentStatus = models.PaymentStatusSuccess
		newOrderStatus = models.OrderStatusPaid
	case "MARK_FAILED":
		newPaymentStatus = models.PaymentStatusFailed
		newOrderStatus = models.OrderStatusFailed
	case "MARK_EXPIRED":
		newPaymentStatus = models.PaymentStatusExpired
		newOrderStatus = models.OrderStatusExpired
	case "SYNC_GATEWAY":
		// This would trigger a gateway sync - simplified here
		newPaymentStatus = payment.Status
		newOrderStatus = order.Status
	default:
		return s.logFailedAction(tx, admin, models.AdminActionReconcilePayment, "payment", paymentID, orderCode, req.IdempotencyKey, "Invalid action")
	}

	// Update payment
	updatePayment := `
		UPDATE payments SET status = $1, is_reconciled = true, reconciled_at = NOW(),
		transaction_id = COALESCE(NULLIF($2, ''), transaction_id)
		WHERE id = $3
	`
	_, err = tx.Exec(updatePayment, newPaymentStatus, req.TransactionID, payment.ID)
	if err != nil {
		return s.logFailedAction(tx, admin, models.AdminActionReconcilePayment, "payment", paymentID, orderCode, req.IdempotencyKey, err.Error())
	}

	// Update order if needed
	if order != nil && newOrderStatus != order.Status {
		if err := s.orderRepo.UpdateStatusTx(tx, order.ID, newOrderStatus); err != nil {
			log.Printf("⚠️ Failed to update order status: %v", err)
		}

		// Restore stock if failed/expired
		if newOrderStatus.RequiresStockRestore() && order.StockReserved {
			s.orderRepo.RestoreStockTx(tx, order.ID)
		}
	}

	stateAfter := map[string]any{
		"payment_status": newPaymentStatus,
		"order_status":   newOrderStatus,
		"reconciled":     true,
		"action":         req.Action,
	}

	// Use nil for admin_user_id if not available
	var adminUserID *int
	if admin.UserID > 0 {
		adminUserID = &admin.UserID
	}
	
	auditLog := &models.AdminAuditLog{
		AdminUserID:    adminUserID,
		AdminEmail:     admin.Email,
		AdminIP:        admin.IP,
		AdminUserAgent: admin.UserAgent,
		ActionType:     models.AdminActionReconcilePayment,
		ActionDetail:   fmt.Sprintf("Reconcile payment %d: %s. Reason: %s", paymentID, req.Action, req.Reason),
		TargetType:     "payment",
		TargetID:       payment.ID,
		TargetCode:     orderCode,
		StateBefore:    stateBefore,
		StateAfter:     stateAfter,
		Success:        true,
		IdempotencyKey: req.IdempotencyKey,
	}

	if err := s.auditRepo.CreateWithTx(tx, auditLog); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	log.Printf("✅ Admin reconcile payment: %d, action %s by %s", paymentID, req.Action, admin.Email)

	return &dto.AdminActionResponse{
		Success:     true,
		Message:     fmt.Sprintf("Payment %d reconciled: %s", paymentID, req.Action),
		AuditLogID:  auditLog.ID,
		StateBefore: stateBefore,
		StateAfter:  stateAfter,
	}, nil
}

func (s *adminService) GetAuditLogs(targetType string, targetID int) ([]*models.AdminAuditLog, error) {
	return s.auditRepo.FindByTarget(targetType, targetID)
}

func (s *adminService) GetRecentAuditLogs(limit int) ([]*models.AdminAuditLog, error) {
	return s.auditRepo.FindRecent(limit)
}

// GetDashboardMetrics returns real-time dashboard metrics
func (s *adminService) GetDashboardMetrics() (*dto.DashboardMetricsResponse, error) {
	metrics := &dto.DashboardMetricsResponse{}

	// Total revenue (sum of PAID/COMPLETED orders)
	err := s.db.QueryRow(`
		SELECT COALESCE(SUM(total_amount), 0)
		FROM orders
		WHERE status IN ('PAID', 'PACKING', 'SHIPPED', 'DELIVERED', 'COMPLETED')
	`).Scan(&metrics.TotalRevenue)
	if err != nil {
		return nil, err
	}

	// Orders today count
	err = s.db.QueryRow(`
		SELECT COUNT(*)
		FROM orders
		WHERE DATE(created_at) = CURRENT_DATE
	`).Scan(&metrics.OrdersToday)
	if err != nil {
		metrics.OrdersToday = 0
	}

	// Orders shipped count
	err = s.db.QueryRow(`
		SELECT COUNT(*)
		FROM orders
		WHERE status = 'SHIPPED'
	`).Scan(&metrics.OrdersShipped)
	if err != nil {
		metrics.OrdersShipped = 0
	}

	// Orders pending count
	err = s.db.QueryRow(`
		SELECT COUNT(*)
		FROM orders
		WHERE status = 'PENDING'
	`).Scan(&metrics.OrdersPending)
	if err != nil {
		metrics.OrdersPending = 0
	}

	// Orders packing count
	err = s.db.QueryRow(`
		SELECT COUNT(*)
		FROM orders
		WHERE status = 'PACKING'
	`).Scan(&metrics.OrdersPacking)
	if err != nil {
		metrics.OrdersPacking = 0
	}

	// Low stock products (stock < 10)
	lowStockQuery := `
		SELECT id, name, stock, price, category
		FROM products
		WHERE stock < 10 AND is_active = true
		ORDER BY stock ASC
		LIMIT 10
	`
	rows, err := s.db.Query(lowStockQuery)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var p dto.LowStockProduct
			rows.Scan(&p.ID, &p.Name, &p.Stock, &p.Price, &p.Category)
			metrics.LowStockProducts = append(metrics.LowStockProducts, p)
		}
	}

	// Recent orders
	recentQuery := `
		SELECT order_code, customer_name, total_amount, status, created_at
		FROM orders
		ORDER BY created_at DESC
		LIMIT 5
	`
	rows, err = s.db.Query(recentQuery)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var o dto.RecentOrderSummary
			var createdAt sql.NullTime
			rows.Scan(&o.OrderCode, &o.CustomerName, &o.TotalAmount, &o.Status, &createdAt)
			if createdAt.Valid {
				o.CreatedAt = dto.FormatTime(createdAt.Time)
			}
			metrics.RecentOrders = append(metrics.RecentOrders, o)
		}
	}

	return metrics, nil
}

// Helper methods
func (s *adminService) captureOrderState(order *models.Order) map[string]any {
	state := map[string]any{
		"id":             order.ID,
		"order_code":     order.OrderCode,
		"status":         order.Status,
		"total_amount":   order.TotalAmount,
		"stock_reserved": order.StockReserved,
	}
	
	// Serialize to ensure clean JSON
	jsonBytes, _ := json.Marshal(state)
	var result map[string]any
	json.Unmarshal(jsonBytes, &result)
	return result
}

func (s *adminService) logFailedAction(tx *sql.Tx, admin *AdminContext, actionType models.AdminActionType, targetType string, targetID int, targetCode, idempotencyKey, errorMsg string) (*dto.AdminActionResponse, error) {
	// Use nil for admin_user_id if not available
	var adminUserID *int
	if admin.UserID > 0 {
		adminUserID = &admin.UserID
	}
	
	auditLog := &models.AdminAuditLog{
		AdminUserID:    adminUserID,
		AdminEmail:     admin.Email,
		AdminIP:        admin.IP,
		AdminUserAgent: admin.UserAgent,
		ActionType:     actionType,
		ActionDetail:   fmt.Sprintf("Failed: %s", errorMsg),
		TargetType:     targetType,
		TargetID:       targetID,
		TargetCode:     targetCode,
		Success:        false,
		ErrorMessage:   errorMsg,
		IdempotencyKey: idempotencyKey,
	}

	// Create audit log but don't commit - let caller handle transaction
	if err := s.auditRepo.CreateWithTx(tx, auditLog); err != nil {
		log.Printf("⚠️ Failed to create audit log: %v", err)
	}

	return &dto.AdminActionResponse{
		Success:    false,
		Message:    errorMsg,
		AuditLogID: auditLog.ID,
	}, fmt.Errorf("%s: %s", actionType, errorMsg)
}

