package service

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
	"zavera/dto"
	"zavera/models"
	"zavera/repository"
)

type AdminOrderService interface {
	GetAllOrdersAdmin(filter dto.AdminOrderFilter) ([]dto.AdminOrderResponse, int, error)
	GetOrderDetailAdmin(orderCode string) (*dto.AdminOrderResponse, error)
	UpdateOrderStatusAdmin(orderCode string, status string, reason string, adminEmail string) error
	GetOrderStats() (*dto.OrderStatsResponse, error)
	PackOrder(orderCode string, adminEmail string) error
	GenerateResiOnly(orderCode string, adminEmail string) (string, error)
	ShipOrder(orderCode string, resi string, adminEmail string) (string, error)
	DeliverOrder(orderCode string, adminEmail string) error
	GetOrderActions(orderCode string) ([]dto.OrderAction, error)
	CancelOrderAdmin(orderCode string, reason string, adminEmail string) error
}

type adminOrderService struct {
	db              *sql.DB
	orderRepo       repository.OrderRepository
	paymentRepo     repository.PaymentRepository
	shippingRepo    repository.ShippingRepository
	shippingService ShippingService
	resiService     ResiService
	auditRepo       repository.AdminAuditRepository
	emailService    EmailService
}

func NewAdminOrderService(
	db *sql.DB,
	orderRepo repository.OrderRepository,
	paymentRepo repository.PaymentRepository,
	shippingRepo repository.ShippingRepository,
	emailRepo repository.EmailRepository,
	shippingService ShippingService,
) AdminOrderService {
	var emailSvc EmailService
	if emailRepo != nil {
		emailSvc = NewEmailService(emailRepo)
	}
	return &adminOrderService{
		db:              db,
		orderRepo:       orderRepo,
		paymentRepo:     paymentRepo,
		shippingRepo:    shippingRepo,
		shippingService: shippingService,
		resiService:     NewResiService(orderRepo),
		emailService:    emailSvc,
	}
}

func (s *adminOrderService) GetAllOrdersAdmin(filter dto.AdminOrderFilter) ([]dto.AdminOrderResponse, int, error) {
	offset := (filter.Page - 1) * filter.PageSize

	// Build WHERE clause
	whereClauses := []string{}
	args := []interface{}{}
	argIndex := 1

	if filter.Status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, filter.Status)
		argIndex++
	}

	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		whereClauses = append(whereClauses, fmt.Sprintf(
			"(order_code ILIKE $%d OR customer_name ILIKE $%d OR customer_email ILIKE $%d)",
			argIndex, argIndex, argIndex,
		))
		args = append(args, searchPattern)
		argIndex++
	}

	if filter.DateFrom != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, filter.DateFrom)
		argIndex++
	}

	if filter.DateTo != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, filter.DateTo+" 23:59:59")
		argIndex++
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM orders %s", whereClause)
	var total int
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get orders
	query := fmt.Sprintf(`
		SELECT id, order_code, customer_name, customer_email, customer_phone,
		       subtotal, shipping_cost, tax, discount, total_amount, status,
		       created_at, updated_at, paid_at, shipped_at
		FROM orders
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, filter.PageSize, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []dto.AdminOrderResponse
	for rows.Next() {
		var o dto.AdminOrderResponse
		var createdAt, updatedAt sql.NullTime
		var paidAt, shippedAt sql.NullTime

		err := rows.Scan(
			&o.ID, &o.OrderCode, &o.CustomerName, &o.CustomerEmail, &o.CustomerPhone,
			&o.Subtotal, &o.ShippingCost, &o.Tax, &o.Discount, &o.TotalAmount, &o.Status,
			&createdAt, &updatedAt, &paidAt, &shippedAt,
		)
		if err != nil {
			continue
		}

		if createdAt.Valid {
			o.CreatedAt = dto.FormatTime(createdAt.Time)
		}
		if updatedAt.Valid {
			o.UpdatedAt = dto.FormatTime(updatedAt.Time)
		}
		o.PaidAt = dto.FormatTimePtr(&paidAt.Time)
		o.ShippedAt = dto.FormatTimePtr(&shippedAt.Time)

		// Load items
		o.Items = s.getOrderItems(o.ID)

		orders = append(orders, o)
	}

	return orders, total, nil
}

func (s *adminOrderService) GetOrderDetailAdmin(orderCode string) (*dto.AdminOrderResponse, error) {
	order, err := s.orderRepo.FindByOrderCode(orderCode)
	if err != nil {
		return nil, ErrOrderNotFound
	}

	response := &dto.AdminOrderResponse{
		ID:            order.ID,
		OrderCode:     order.OrderCode,
		CustomerName:  order.CustomerName,
		CustomerEmail: order.CustomerEmail,
		CustomerPhone: order.CustomerPhone,
		Subtotal:      order.Subtotal,
		ShippingCost:  order.ShippingCost,
		Tax:           order.Tax,
		Discount:      order.Discount,
		TotalAmount:   order.TotalAmount,
		Status:        string(order.Status),
		Resi:          order.Resi,
		CreatedAt:     dto.FormatTime(order.CreatedAt),
		UpdatedAt:     dto.FormatTime(order.UpdatedAt),
	}

	if order.PaidAt != nil {
		response.PaidAt = dto.FormatTimePtr(order.PaidAt)
	}
	if order.ShippedAt != nil {
		response.ShippedAt = dto.FormatTimePtr(order.ShippedAt)
	}

	// Load items
	for _, item := range order.Items {
		response.Items = append(response.Items, dto.OrderItemResponse{
			ProductID:    item.ProductID,
			ProductName:  item.ProductName,
			ProductImage: item.ProductImage,
			Quantity:     item.Quantity,
			PricePerUnit: item.PricePerUnit,
			Subtotal:     item.Subtotal,
		})
	}

	// Load payment info - try order_payments first (Core API), then payments (Snap)
	var paymentLoaded bool
	
	// Try order_payments table (Core API - VA/QRIS)
	var opID int
	var opStatus, opMethod, opBank, opTransactionID string
	var opPaidAt sql.NullTime
	err = s.db.QueryRow(`
		SELECT id, payment_status, payment_method, bank, transaction_id, paid_at
		FROM order_payments WHERE order_id = $1 ORDER BY created_at DESC LIMIT 1
	`, order.ID).Scan(&opID, &opStatus, &opMethod, &opBank, &opTransactionID, &opPaidAt)
	if err == nil {
		response.Payment = &dto.AdminPaymentInfo{
			ID:              opID,
			Status:          opStatus,
			PaymentMethod:   opMethod,
			PaymentProvider: opBank,
			TransactionID:   opTransactionID,
			Amount:          order.TotalAmount,
		}
		if opPaidAt.Valid {
			paidAtStr := opPaidAt.Time.Format("2006-01-02 15:04:05")
			response.Payment.PaidAt = &paidAtStr
		}
		paymentLoaded = true
	}
	
	// Fallback to payments table (Snap)
	if !paymentLoaded {
		payment, err := s.paymentRepo.FindByOrderID(order.ID)
		if err == nil && payment != nil {
			response.Payment = &dto.AdminPaymentInfo{
				ID:              payment.ID,
				Status:          string(payment.Status),
				PaymentMethod:   payment.PaymentMethod,
				PaymentProvider: payment.PaymentProvider,
				Amount:          payment.Amount,
				TransactionID:   payment.TransactionID,
			}
			if payment.PaidAt != nil {
				response.Payment.PaidAt = dto.FormatTimePtr(payment.PaidAt)
			}
		}
	}

	// Load shipment info
	shipment, err := s.shippingRepo.FindByOrderID(order.ID)
	if err == nil && shipment != nil {
		response.Shipment = &dto.AdminShipmentInfo{
			ID:              shipment.ID,
			ProviderCode:    shipment.ProviderCode,
			ProviderName:    shipment.ProviderName,
			ServiceCode:     shipment.ServiceCode,
			ServiceName:     shipment.ServiceName,
			TrackingNumber:  shipment.TrackingNumber,
			Status:          string(shipment.Status),
			Cost:            shipment.Cost,
			ETD:             shipment.ETD,
			Weight:          shipment.Weight,
			OriginCity:      shipment.OriginCityName,
			DestinationCity: shipment.DestinationCityName,
		}
		if shipment.ShippedAt != nil {
			response.Shipment.ShippedAt = dto.FormatTimePtr(shipment.ShippedAt)
		}
		if shipment.DeliveredAt != nil {
			response.Shipment.DeliveredAt = dto.FormatTimePtr(shipment.DeliveredAt)
		}
	}

	return response, nil
}

func (s *adminOrderService) UpdateOrderStatusAdmin(orderCode string, status string, reason string, adminEmail string) error {
	order, err := s.orderRepo.FindByOrderCode(orderCode)
	if err != nil {
		return ErrOrderNotFound
	}

	newStatus := models.OrderStatus(status)

	// Validate transition
	if !order.Status.IsValidTransition(newStatus) {
		return fmt.Errorf("%w: cannot transition from %s to %s",
			ErrInvalidTransition, order.Status, newStatus)
	}

	// Handle stock restoration for failure states
	if newStatus.RequiresStockRestore() && order.StockReserved {
		if err := s.orderRepo.RestoreStock(order.ID); err != nil {
			return fmt.Errorf("failed to restore stock: %w", err)
		}
	}

	// Update status
	var updateErr error
	switch newStatus {
	case models.OrderStatusPaid:
		updateErr = s.orderRepo.MarkAsPaid(order.ID)
		// Also update payment status if exists
		if updateErr == nil {
			// Try to update payment record to PAID (ignore error if no payment exists)
			s.db.Exec(`
				UPDATE order_payments 
				SET payment_status = 'PAID', paid_at = NOW(), updated_at = NOW()
				WHERE order_id = $1 AND payment_status = 'PENDING'
			`, order.ID)
		}
	case models.OrderStatusShipped:
		updateErr = s.orderRepo.MarkAsShipped(order.ID)
	case models.OrderStatusCompleted:
		updateErr = s.orderRepo.MarkAsCompleted(order.ID)
	case models.OrderStatusCancelled:
		updateErr = s.orderRepo.MarkAsCancelled(order.ID)
	case models.OrderStatusExpired:
		updateErr = s.orderRepo.MarkAsExpired(order.ID)
	default:
		updateErr = s.orderRepo.UpdateStatus(order.ID, newStatus)
	}

	if updateErr != nil {
		return updateErr
	}

	// Record status change
	changedBy := "admin"
	if adminEmail != "" {
		changedBy = adminEmail
	}
	s.orderRepo.RecordStatusChange(order.ID, order.Status, newStatus, changedBy, reason)

	return nil
}

func (s *adminOrderService) GetOrderStats() (*dto.OrderStatsResponse, error) {
	stats := &dto.OrderStatsResponse{}

	// Total orders and revenue
	err := s.db.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(total_amount), 0)
		FROM orders
		WHERE status NOT IN ('CANCELLED', 'FAILED', 'EXPIRED')
	`).Scan(&stats.TotalOrders, &stats.TotalRevenue)
	if err != nil {
		return nil, err
	}

	// Orders by status
	statusQuery := `
		SELECT status, COUNT(*)
		FROM orders
		GROUP BY status
	`
	rows, err := s.db.Query(statusQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int
		rows.Scan(&status, &count)

		switch status {
		case "PENDING":
			stats.PendingOrders = count
		case "PAID":
			stats.PaidOrders = count
		case "PROCESSING":
			stats.ProcessingOrders = count
		case "SHIPPED":
			stats.ShippedOrders = count
		case "DELIVERED":
			stats.DeliveredOrders = count
		case "CANCELLED":
			stats.CancelledOrders = count
		}
	}

	// Today's orders and revenue
	today := time.Now().Format("2006-01-02")
	err = s.db.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(total_amount), 0)
		FROM orders
		WHERE DATE(created_at) = $1
		AND status NOT IN ('CANCELLED', 'FAILED', 'EXPIRED')
	`, today).Scan(&stats.TodayOrders, &stats.TodayRevenue)
	if err != nil {
		// Non-critical, continue
		stats.TodayOrders = 0
		stats.TodayRevenue = 0
	}

	return stats, nil
}

func (s *adminOrderService) getOrderItems(orderID int) []dto.OrderItemResponse {
	query := `
		SELECT oi.product_id, oi.product_name, oi.quantity, oi.price_per_unit, oi.subtotal,
		       COALESCE(
		           (SELECT image_url FROM product_images WHERE product_id = oi.product_id ORDER BY is_primary DESC, display_order ASC LIMIT 1),
		           ''
		       ) as product_image
		FROM order_items oi
		WHERE oi.order_id = $1
	`

	rows, err := s.db.Query(query, orderID)
	if err != nil {
		return []dto.OrderItemResponse{}
	}
	defer rows.Close()

	var items []dto.OrderItemResponse
	for rows.Next() {
		var item dto.OrderItemResponse
		err := rows.Scan(
			&item.ProductID, &item.ProductName, &item.Quantity,
			&item.PricePerUnit, &item.Subtotal, &item.ProductImage,
		)
		if err != nil {
			continue
		}
		items = append(items, item)
	}

	return items
}


// PackOrder marks an order as being packed (PAID -> PACKING)
func (s *adminOrderService) PackOrder(orderCode string, adminEmail string) error {
	order, err := s.orderRepo.FindByOrderCode(orderCode)
	if err != nil {
		return ErrOrderNotFound
	}

	// Validate status - can only pack PAID orders
	if order.Status != models.OrderStatusPaid {
		return fmt.Errorf("%w: can only pack orders with PAID status, current: %s",
			ErrInvalidTransition, order.Status)
	}

	// Update to PACKING
	err = s.orderRepo.MarkAsPacking(order.ID)
	if err != nil {
		return err
	}

	// Record status change
	changedBy := "admin"
	if adminEmail != "" {
		changedBy = adminEmail
	}
	s.orderRepo.RecordStatusChange(order.ID, order.Status, models.OrderStatusPacking, changedBy, "Order packed by admin")

	return nil
}

// GenerateResiOnly generates resi from Biteship WITHOUT shipping the order yet
// This allows admin to see the resi before confirming shipment
func (s *adminOrderService) GenerateResiOnly(orderCode string, adminEmail string) (string, error) {
	order, err := s.orderRepo.FindByOrderCode(orderCode)
	if err != nil {
		return "", ErrOrderNotFound
	}

	// Validate status - can only generate resi for PACKING orders
	if order.Status != models.OrderStatusPacking {
		return "", fmt.Errorf("%w: can only generate resi for orders with PACKING status, current: %s",
			ErrInvalidTransition, order.Status)
	}

	// Get shipment to check for draft order
	shipment, err := s.shippingRepo.FindByOrderID(order.ID)
	if err != nil {
		return "", fmt.Errorf("shipment not found for order")
	}

	// Try to get resi from Biteship draft order
	if shipment.BiteshipDraftOrderID != "" {
		log.Printf("🚀 Generating resi from Biteship for order %s (draft: %s)", orderCode, shipment.BiteshipDraftOrderID)
		
		// Confirm draft order to get waybill (resi)
		confirmation, err := s.shippingService.ConfirmDraftOrder(order.ID)
		if err != nil {
			log.Printf("⚠️ Failed to confirm Biteship draft order: %v", err)
			// Fallback to manual resi generation
			resi, genErr := s.resiService.GenerateResi(order.ID, shipment.ProviderCode)
			if genErr != nil {
				return "", fmt.Errorf("failed to generate fallback resi: %w", genErr)
			}
			log.Printf("⚠️ Using fallback manual resi: %s", resi)
			return resi, nil
		}

		// Got resi from Biteship!
		resi := confirmation.WaybillID
		if resi == "" {
			log.Printf("⚠️ Biteship confirmation succeeded but no waybill_id returned")
			// Fallback to manual resi
			resi, genErr := s.resiService.GenerateResi(order.ID, shipment.ProviderCode)
			if genErr != nil {
				return "", fmt.Errorf("failed to generate fallback resi: %w", genErr)
			}
			return resi, nil
		}

		log.Printf("✅ Got resi from Biteship: %s (Tracking: %s)", resi, confirmation.TrackingID)
		
		// Optionally save resi to shipment for reference (but don't ship yet)
		// This allows admin to see the resi before confirming
		// The actual shipping happens in ShipOrder method
		
		return resi, nil
	}

	// No Biteship draft order - generate manual resi
	log.Printf("⚠️ No Biteship draft order found for order %s, generating manual resi", orderCode)
	resi, err := s.resiService.GenerateResi(order.ID, shipment.ProviderCode)
	if err != nil {
		return "", fmt.Errorf("failed to generate manual resi: %w", err)
	}
	return resi, nil
}

// ShipOrder marks an order as shipped with resi (PACKING -> SHIPPED)
func (s *adminOrderService) ShipOrder(orderCode string, resi string, adminEmail string) (string, error) {
	order, err := s.orderRepo.FindByOrderCode(orderCode)
	if err != nil {
		return "", ErrOrderNotFound
	}

	// Validate status - can only ship PACKING orders
	if order.Status != models.OrderStatusPacking {
		return "", fmt.Errorf("%w: can only ship orders with PACKING status, current: %s",
			ErrInvalidTransition, order.Status)
	}

	// Get shipment to determine courier
	shipment, err := s.shippingRepo.FindByOrderID(order.ID)
	if err != nil {
		return "", fmt.Errorf("shipment not found for order")
	}

	// Validate resi is provided
	if resi == "" {
		return "", fmt.Errorf("nomor resi harus diisi")
	}

	// Validate resi format
	resi = strings.TrimSpace(resi)
	if len(resi) < 8 {
		return "", fmt.Errorf("nomor resi tidak valid: minimal 8 karakter")
	}
	// Resi must be alphanumeric with optional dashes/hyphens
	for _, c := range resi {
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return "", fmt.Errorf("nomor resi tidak valid: hanya boleh huruf, angka, dan tanda strip (-)")
		}
	}

	// Update order with resi and status
	err = s.orderRepo.MarkAsShippedWithResi(order.ID, resi)
	if err != nil {
		return "", err
	}

	// Update shipment
	s.shippingRepo.MarkShipmentShipped(shipment.ID, resi)

	// Record status change
	changedBy := "admin"
	if adminEmail != "" {
		changedBy = adminEmail
	}
	s.orderRepo.RecordStatusChange(order.ID, order.Status, models.OrderStatusShipped, changedBy, fmt.Sprintf("Order shipped with resi: %s", resi))

	// Send ORDER_SHIPPED email (Tokopedia-style: only when admin actually ships)
	if s.emailService != nil {
		go func() {
			// Reload order to get updated data with resi
			updatedOrder, err := s.orderRepo.FindByOrderCode(orderCode)
			if err != nil {
				return
			}
			// Get shipping address from metadata
			shippingAddr := ""
			if updatedOrder.Metadata != nil {
				if addr, ok := updatedOrder.Metadata["shipping_address_snapshot"].(string); ok {
					shippingAddr = addr
				}
			}
			s.emailService.SendOrderShipped(updatedOrder, shipment, shippingAddr)
		}()
	}

	log.Printf("✅ Order %s shipped with resi: %s", orderCode, resi)
	return resi, nil
}

// DeliverOrder marks an order as delivered (SHIPPED -> DELIVERED)
func (s *adminOrderService) DeliverOrder(orderCode string, adminEmail string) error {
	order, err := s.orderRepo.FindByOrderCode(orderCode)
	if err != nil {
		return ErrOrderNotFound
	}

	// Validate status - can only deliver SHIPPED orders
	if order.Status != models.OrderStatusShipped {
		return fmt.Errorf("%w: can only deliver orders with SHIPPED status, current: %s",
			ErrInvalidTransition, order.Status)
	}

	// Update order status to DELIVERED
	err = s.orderRepo.UpdateStatus(order.ID, models.OrderStatusDelivered)
	if err != nil {
		return err
	}

	// Update shipment status
	shipment, _ := s.shippingRepo.FindByOrderID(order.ID)
	if shipment != nil {
		s.shippingRepo.UpdateShipmentStatus(shipment.ID, "DELIVERED")
	}

	// Record status change
	changedBy := "admin"
	if adminEmail != "" {
		changedBy = adminEmail
	}
	s.orderRepo.RecordStatusChange(order.ID, order.Status, models.OrderStatusDelivered, changedBy, "Order marked as delivered")

	return nil
}

// GetOrderActions returns available actions for an order based on its status
func (s *adminOrderService) GetOrderActions(orderCode string) ([]dto.OrderAction, error) {
	order, err := s.orderRepo.FindByOrderCode(orderCode)
	if err != nil {
		return nil, ErrOrderNotFound
	}

	var actions []dto.OrderAction

	switch order.Status {
	case models.OrderStatusPending:
		// PENDING = waiting for payment
		actions = append(actions, dto.OrderAction{
			Action:      "mark_paid",
			Label:       "Mark as Paid",
			Enabled:     true,
			Description: "Manually mark this order as paid",
		})
		actions = append(actions, dto.OrderAction{
			Action:      "cancel",
			Label:       "Cancel Order",
			Enabled:     true,
			Description: "Cancel this pending order",
		})

	case models.OrderStatusPaid:
		actions = append(actions, dto.OrderAction{
			Action:      "pack",
			Label:       "Pack Order",
			Enabled:     true,
			Description: "Mark order as being packed",
		})
		actions = append(actions, dto.OrderAction{
			Action:      "cancel",
			Label:       "Cancel Order",
			Enabled:     true,
			Description: "Cancel this order (will restore stock)",
		})

	case models.OrderStatusPacking:
		actions = append(actions, dto.OrderAction{
			Action:      "ship",
			Label:       "Ship Order",
			Enabled:     true,
			Description: "Mark order as shipped (will generate resi)",
		})
		actions = append(actions, dto.OrderAction{
			Action:      "cancel",
			Label:       "Cancel Order",
			Enabled:     true,
			Description: "Cancel this order (will restore stock)",
		})

	case models.OrderStatusShipped:
		actions = append(actions, dto.OrderAction{
			Action:      "deliver",
			Label:       "Mark Delivered",
			Enabled:     true,
			Description: "Mark order as delivered",
		})

	case models.OrderStatusDelivered:
		actions = append(actions, dto.OrderAction{
			Action:      "complete",
			Label:       "Complete Order",
			Enabled:     true,
			Description: "Mark order as completed",
		})
		actions = append(actions, dto.OrderAction{
			Action:      "refund",
			Label:       "Process Refund",
			Enabled:     true,
			Description: "Process refund for this order",
		})

	case models.OrderStatusCompleted:
		actions = append(actions, dto.OrderAction{
			Action:      "refund",
			Label:       "Process Refund",
			Enabled:     true,
			Description: "Process refund for this completed order",
		})
	}

	return actions, nil
}

// CancelOrderAdmin cancels an order (admin can cancel before SHIPPED)
func (s *adminOrderService) CancelOrderAdmin(orderCode string, reason string, adminEmail string) error {
	order, err := s.orderRepo.FindByOrderCode(orderCode)
	if err != nil {
		return ErrOrderNotFound
	}

	// Admin can only cancel orders before SHIPPED
	if !order.CanBeCancelledByAdmin() {
		return fmt.Errorf("%w: admin can only cancel orders before shipping, current: %s",
			ErrInvalidTransition, order.Status)
	}

	// Restore stock if reserved
	if order.StockReserved {
		if err := s.orderRepo.RestoreStock(order.ID); err != nil {
			return fmt.Errorf("failed to restore stock: %w", err)
		}
	}

	// Update status
	err = s.orderRepo.MarkAsCancelled(order.ID)
	if err != nil {
		return err
	}

	// Record status change
	changedBy := "admin"
	if adminEmail != "" {
		changedBy = adminEmail
	}
	s.orderRepo.RecordStatusChange(order.ID, order.Status, models.OrderStatusCancelled, changedBy, reason)

	return nil
}
