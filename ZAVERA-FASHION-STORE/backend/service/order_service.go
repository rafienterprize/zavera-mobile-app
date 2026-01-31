package service

import (
	"errors"
	"fmt"
	"zavera/dto"
	"zavera/models"
	"zavera/repository"
)

var (
	ErrOrderNotFound       = errors.New("order not found")
	ErrInvalidTransition   = errors.New("invalid status transition")
	ErrOrderAlreadyFinal   = errors.New("order is in final state, cannot be modified")
	ErrCartEmpty           = errors.New("cart is empty")
	ErrInsufficientStock   = errors.New("insufficient stock")
)

type OrderService interface {
	CreateOrder(sessionID string, req dto.CheckoutRequest, userID *int) (*dto.CheckoutResponse, error)
	GetOrder(orderCode string) (*dto.OrderResponse, error)
	GetOrderByID(orderID int) (*models.Order, error)
	GetOrderItems(orderID int) ([]models.OrderItem, error)
	UpdateOrderStatus(orderCode string, status models.OrderStatus) error
	
	// State machine methods
	MarkAsPaid(orderCode string) error
	MarkAsPacking(orderCode string) error
	MarkAsShippedWithResi(orderCode string, resi string) error
	MarkAsShipped(orderCode string) error
	MarkAsDelivered(orderCode string) error
	MarkAsCompleted(orderCode string) error
	MarkAsRefunded(orderCode string) error
	CancelOrder(orderCode string, reason string) error
	CancelOrderByCustomer(orderCode string) error
	CancelOrderByAdmin(orderCode string, reason string) error
	ExpireOrder(orderCode string) error
	FailOrder(orderCode string, reason string) error
	
	// Validation
	ValidateOrderTotals(order *models.Order) error
}

type orderService struct {
	orderRepo   repository.OrderRepository
	cartRepo    repository.CartRepository
	productRepo repository.ProductRepository
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	cartRepo repository.CartRepository,
	productRepo repository.ProductRepository,
) OrderService {
	return &orderService{
		orderRepo:   orderRepo,
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

func (s *orderService) CreateOrder(sessionID string, req dto.CheckoutRequest, userID *int) (*dto.CheckoutResponse, error) {
	// Get cart
	cart, err := s.cartRepo.FindOrCreateBySessionID(sessionID)
	if err != nil {
		return nil, err
	}

	if len(cart.Items) == 0 {
		return nil, ErrCartEmpty
	}

	// Validate stock and calculate subtotal
	var subtotal float64
	var orderItems []models.OrderItem

	for _, item := range cart.Items {
		product, err := s.productRepo.FindByID(item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("product not found: %w", err)
		}

		if product.Stock < item.Quantity {
			return nil, fmt.Errorf("%w for product: %s", ErrInsufficientStock, product.Name)
		}

		itemSubtotal := item.PriceSnapshot * float64(item.Quantity)
		subtotal += itemSubtotal

		orderItem := models.OrderItem{
			ProductID:    item.ProductID,
			ProductName:  product.Name,
			Quantity:     item.Quantity,
			PricePerUnit: item.PriceSnapshot,
			Subtotal:     itemSubtotal,
			Metadata:     item.Metadata,
		}

		orderItems = append(orderItems, orderItem)
	}

	// Calculate totals
	shippingCost := 15000.0 // Fixed shipping for now
	tax := 0.0
	discount := 0.0
	totalAmount := subtotal + shippingCost + tax - discount

	// Create order (stock is reserved atomically in repository)
	order := &models.Order{
		UserID:        userID, // Link to user if authenticated
		CustomerName:  req.CustomerName,
		CustomerEmail: req.CustomerEmail,
		CustomerPhone: req.CustomerPhone,
		Subtotal:      subtotal,
		ShippingCost:  shippingCost,
		Tax:           tax,
		Discount:      discount,
		TotalAmount:   totalAmount,
		Status:        models.OrderStatusPending,
		Notes:         req.Notes,
	}

	err = s.orderRepo.Create(order, orderItems)
	if err != nil {
		return nil, err
	}

	// Clear cart after successful order creation
	s.cartRepo.ClearCart(cart.ID)

	// Return response
	response := &dto.CheckoutResponse{
		OrderID:     order.ID,
		OrderCode:   order.OrderCode,
		TotalAmount: order.TotalAmount,
		Status:      string(order.Status),
	}

	return response, nil
}

func (s *orderService) GetOrder(orderCode string) (*dto.OrderResponse, error) {
	order, err := s.orderRepo.FindByOrderCode(orderCode)
	if err != nil {
		return nil, ErrOrderNotFound
	}

	return s.toOrderResponse(order), nil
}

// GetOrderByID returns an order by ID
func (s *orderService) GetOrderByID(orderID int) (*models.Order, error) {
	return s.orderRepo.FindByID(orderID)
}

// GetOrderItems returns order items for an order
func (s *orderService) GetOrderItems(orderID int) ([]models.OrderItem, error) {
	return s.orderRepo.GetOrderItems(orderID)
}

// UpdateOrderStatus is a generic status update with validation
func (s *orderService) UpdateOrderStatus(orderCode string, status models.OrderStatus) error {
	order, err := s.orderRepo.FindByOrderCode(orderCode)
	if err != nil {
		return ErrOrderNotFound
	}

	// Check if order is in final state
	if order.Status.IsFinalStatus() {
		return ErrOrderAlreadyFinal
	}

	// Validate transition
	if !order.Status.IsValidTransition(status) {
		return fmt.Errorf("%w: cannot transition from %s to %s", 
			ErrInvalidTransition, order.Status, status)
	}

	// Handle stock restoration for failure states
	if status.RequiresStockRestore() && order.StockReserved {
		if err := s.orderRepo.RestoreStock(order.ID); err != nil {
			return fmt.Errorf("failed to restore stock: %w", err)
		}
	}

	// Update status based on type
	var updateErr error
	switch status {
	case models.OrderStatusPaid:
		updateErr = s.orderRepo.MarkAsPaid(order.ID)
	case models.OrderStatusShipped:
		updateErr = s.orderRepo.MarkAsShipped(order.ID)
	case models.OrderStatusCompleted:
		updateErr = s.orderRepo.MarkAsCompleted(order.ID)
	case models.OrderStatusCancelled:
		updateErr = s.orderRepo.MarkAsCancelled(order.ID)
	case models.OrderStatusExpired:
		updateErr = s.orderRepo.MarkAsExpired(order.ID)
	default:
		updateErr = s.orderRepo.UpdateStatus(order.ID, status)
	}

	if updateErr != nil {
		return updateErr
	}

	// Record status change (non-blocking)
	s.orderRepo.RecordStatusChange(order.ID, order.Status, status, "system", "")

	return nil
}

// MarkAsPaid transitions order from PENDING to PAID
func (s *orderService) MarkAsPaid(orderCode string) error {
	order, err := s.orderRepo.FindByOrderCode(orderCode)
	if err != nil {
		return ErrOrderNotFound
	}

	// Idempotency: already paid
	if order.Status == models.OrderStatusPaid {
		return nil
	}

	// Validate transition
	if !order.Status.IsValidTransition(models.OrderStatusPaid) {
		return fmt.Errorf("%w: cannot mark as paid from status %s", 
			ErrInvalidTransition, order.Status)
	}

	err = s.orderRepo.MarkAsPaid(order.ID)
	if err != nil {
		return err
	}

	s.orderRepo.RecordStatusChange(order.ID, order.Status, models.OrderStatusPaid, "payment", "Payment successful")
	return nil
}

// MarkAsShipped transitions order from PAID/PROCESSING to SHIPPED
func (s *orderService) MarkAsShipped(orderCode string) error {
	order, err := s.orderRepo.FindByOrderCode(orderCode)
	if err != nil {
		return ErrOrderNotFound
	}

	// Idempotency: already shipped
	if order.Status == models.OrderStatusShipped {
		return nil
	}

	// Validate transition
	if !order.Status.IsValidTransition(models.OrderStatusShipped) {
		return fmt.Errorf("%w: cannot ship order from status %s", 
			ErrInvalidTransition, order.Status)
	}

	err = s.orderRepo.MarkAsShipped(order.ID)
	if err != nil {
		return err
	}

	s.orderRepo.RecordStatusChange(order.ID, order.Status, models.OrderStatusShipped, "admin", "Order shipped")
	return nil
}

// MarkAsPacking transitions order from PAID to PACKING
func (s *orderService) MarkAsPacking(orderCode string) error {
	order, err := s.orderRepo.FindByOrderCode(orderCode)
	if err != nil {
		return ErrOrderNotFound
	}

	// Idempotency: already packing
	if order.Status == models.OrderStatusPacking {
		return nil
	}

	// Validate transition (PAID -> PACKING)
	if order.Status != models.OrderStatusPaid {
		return fmt.Errorf("%w: can only pack orders with PAID status, current: %s", 
			ErrInvalidTransition, order.Status)
	}

	err = s.orderRepo.MarkAsPacking(order.ID)
	if err != nil {
		return err
	}

	s.orderRepo.RecordStatusChange(order.ID, order.Status, models.OrderStatusPacking, "admin", "Order being packed")
	return nil
}

// MarkAsShippedWithResi transitions order from PACKING to SHIPPED with resi
func (s *orderService) MarkAsShippedWithResi(orderCode string, resi string) error {
	order, err := s.orderRepo.FindByOrderCode(orderCode)
	if err != nil {
		return ErrOrderNotFound
	}

	// Idempotency: already shipped
	if order.Status == models.OrderStatusShipped {
		return nil
	}

	// Validate transition (PACKING -> SHIPPED)
	if order.Status != models.OrderStatusPacking {
		return fmt.Errorf("%w: can only ship orders with PACKING status, current: %s", 
			ErrInvalidTransition, order.Status)
	}

	// Validate resi
	if resi == "" {
		return errors.New("resi is required for shipping")
	}

	err = s.orderRepo.MarkAsShippedWithResi(order.ID, resi)
	if err != nil {
		return err
	}

	s.orderRepo.RecordStatusChange(order.ID, order.Status, models.OrderStatusShipped, "admin", fmt.Sprintf("Order shipped with resi: %s", resi))
	return nil
}

// MarkAsDelivered transitions order from SHIPPED to DELIVERED
func (s *orderService) MarkAsDelivered(orderCode string) error {
	order, err := s.orderRepo.FindByOrderCode(orderCode)
	if err != nil {
		return ErrOrderNotFound
	}

	// Idempotency: already delivered
	if order.Status == models.OrderStatusDelivered {
		return nil
	}

	// Validate transition (SHIPPED -> DELIVERED)
	if order.Status != models.OrderStatusShipped {
		return fmt.Errorf("%w: can only mark delivered orders with SHIPPED status, current: %s", 
			ErrInvalidTransition, order.Status)
	}

	err = s.orderRepo.MarkAsDelivered(order.ID)
	if err != nil {
		return err
	}

	s.orderRepo.RecordStatusChange(order.ID, order.Status, models.OrderStatusDelivered, "system", "Order delivered")
	return nil
}

// MarkAsRefunded transitions order from DELIVERED/COMPLETED to REFUNDED
func (s *orderService) MarkAsRefunded(orderCode string) error {
	order, err := s.orderRepo.FindByOrderCode(orderCode)
	if err != nil {
		return ErrOrderNotFound
	}

	// Idempotency: already refunded
	if order.Status == models.OrderStatusRefunded {
		return nil
	}

	// Validate transition (DELIVERED/COMPLETED -> REFUNDED)
	if order.Status != models.OrderStatusDelivered && order.Status != models.OrderStatusCompleted {
		return fmt.Errorf("%w: can only refund DELIVERED or COMPLETED orders, current: %s", 
			ErrInvalidTransition, order.Status)
	}

	err = s.orderRepo.MarkAsRefunded(order.ID)
	if err != nil {
		return err
	}

	s.orderRepo.RecordStatusChange(order.ID, order.Status, models.OrderStatusRefunded, "admin", "Order refunded")
	return nil
}

// CancelOrderByCustomer allows customer to cancel only PENDING orders
func (s *orderService) CancelOrderByCustomer(orderCode string) error {
	order, err := s.orderRepo.FindByOrderCode(orderCode)
	if err != nil {
		return ErrOrderNotFound
	}

	// Customer can only cancel PENDING orders
	if order.Status != models.OrderStatusPending {
		return fmt.Errorf("%w: customers can only cancel orders with PENDING status", ErrInvalidTransition)
	}

	return s.CancelOrder(orderCode, "Cancelled by customer")
}

// CancelOrderByAdmin allows admin to cancel orders before SHIPPED
func (s *orderService) CancelOrderByAdmin(orderCode string, reason string) error {
	order, err := s.orderRepo.FindByOrderCode(orderCode)
	if err != nil {
		return ErrOrderNotFound
	}

	// Admin can cancel PENDING, PAID, or PACKING orders
	if !order.CanBeCancelledByAdmin() {
		return fmt.Errorf("%w: admin can only cancel orders before shipping, current: %s", ErrInvalidTransition, order.Status)
	}

	return s.CancelOrder(orderCode, reason)
}

// ValidateOrderTotals validates that order totals are calculated correctly
func (s *orderService) ValidateOrderTotals(order *models.Order) error {
	// Recalculate subtotal from items
	var calculatedSubtotal float64
	for _, item := range order.Items {
		calculatedSubtotal += item.PricePerUnit * float64(item.Quantity)
	}

	// Check subtotal matches
	if calculatedSubtotal != order.Subtotal {
		return fmt.Errorf("subtotal mismatch: calculated %.2f, stored %.2f", calculatedSubtotal, order.Subtotal)
	}

	// Check total calculation
	expectedTotal := order.Subtotal + order.ShippingCost + order.Tax - order.Discount
	if expectedTotal != order.TotalAmount {
		return fmt.Errorf("total mismatch: calculated %.2f, stored %.2f", expectedTotal, order.TotalAmount)
	}

	return nil
}

// MarkAsCompleted transitions order from SHIPPED/DELIVERED to COMPLETED
func (s *orderService) MarkAsCompleted(orderCode string) error {
	order, err := s.orderRepo.FindByOrderCode(orderCode)
	if err != nil {
		return ErrOrderNotFound
	}

	// Idempotency: already completed
	if order.Status == models.OrderStatusCompleted {
		return nil
	}

	// Validate transition
	if !order.Status.IsValidTransition(models.OrderStatusCompleted) {
		return fmt.Errorf("%w: cannot complete order from status %s", 
			ErrInvalidTransition, order.Status)
	}

	err = s.orderRepo.MarkAsCompleted(order.ID)
	if err != nil {
		return err
	}

	s.orderRepo.RecordStatusChange(order.ID, order.Status, models.OrderStatusCompleted, "system", "Order completed")
	return nil
}

// CancelOrder cancels an order and restores stock if needed
func (s *orderService) CancelOrder(orderCode string, reason string) error {
	order, err := s.orderRepo.FindByOrderCode(orderCode)
	if err != nil {
		return ErrOrderNotFound
	}

	// Idempotency: already cancelled
	if order.Status == models.OrderStatusCancelled {
		return nil
	}

	// Check if order is in final state
	if order.Status.IsFinalStatus() {
		return ErrOrderAlreadyFinal
	}

	// Validate transition
	if !order.Status.IsValidTransition(models.OrderStatusCancelled) {
		return fmt.Errorf("%w: cannot cancel order from status %s", 
			ErrInvalidTransition, order.Status)
	}

	// Restore stock if reserved
	if order.StockReserved {
		if err := s.orderRepo.RestoreStock(order.ID); err != nil {
			return fmt.Errorf("failed to restore stock: %w", err)
		}
	}

	err = s.orderRepo.MarkAsCancelled(order.ID)
	if err != nil {
		return err
	}

	s.orderRepo.RecordStatusChange(order.ID, order.Status, models.OrderStatusCancelled, "system", reason)
	return nil
}

// ExpireOrder expires a pending order and restores stock
func (s *orderService) ExpireOrder(orderCode string) error {
	order, err := s.orderRepo.FindByOrderCode(orderCode)
	if err != nil {
		return ErrOrderNotFound
	}

	// Idempotency: already expired
	if order.Status == models.OrderStatusExpired {
		return nil
	}

	// Only pending orders can expire
	if order.Status != models.OrderStatusPending {
		return fmt.Errorf("%w: only pending orders can expire", ErrInvalidTransition)
	}

	// Restore stock
	if order.StockReserved {
		if err := s.orderRepo.RestoreStock(order.ID); err != nil {
			return fmt.Errorf("failed to restore stock: %w", err)
		}
	}

	err = s.orderRepo.MarkAsExpired(order.ID)
	if err != nil {
		return err
	}

	s.orderRepo.RecordStatusChange(order.ID, order.Status, models.OrderStatusExpired, "system", "Payment expired")
	return nil
}

// FailOrder marks an order as failed and restores stock
func (s *orderService) FailOrder(orderCode string, reason string) error {
	order, err := s.orderRepo.FindByOrderCode(orderCode)
	if err != nil {
		return ErrOrderNotFound
	}

	// Idempotency: already failed
	if order.Status == models.OrderStatusFailed {
		return nil
	}

	// Only pending orders can fail
	if order.Status != models.OrderStatusPending {
		return fmt.Errorf("%w: only pending orders can fail", ErrInvalidTransition)
	}

	// Restore stock
	if order.StockReserved {
		if err := s.orderRepo.RestoreStock(order.ID); err != nil {
			return fmt.Errorf("failed to restore stock: %w", err)
		}
	}

	err = s.orderRepo.UpdateStatus(order.ID, models.OrderStatusFailed)
	if err != nil {
		return err
	}

	s.orderRepo.RecordStatusChange(order.ID, order.Status, models.OrderStatusFailed, "system", reason)
	return nil
}

func (s *orderService) toOrderResponse(order *models.Order) *dto.OrderResponse {
	response := &dto.OrderResponse{
		ID:            order.ID,
		OrderCode:     order.OrderCode,
		UserID:        order.UserID,
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
		Items:         []dto.OrderItemResponse{},
		CreatedAt:     order.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	// Add timestamps
	if order.PaidAt != nil {
		response.PaidAt = order.PaidAt.Format("2006-01-02 15:04:05")
	}
	if order.ShippedAt != nil {
		response.ShippedAt = order.ShippedAt.Format("2006-01-02 15:04:05")
	}
	if order.DeliveredAt != nil {
		response.DeliveredAt = order.DeliveredAt.Format("2006-01-02 15:04:05")
	}

	for _, item := range order.Items {
		itemResponse := dto.OrderItemResponse{
			ProductID:    item.ProductID,
			ProductName:  item.ProductName,
			ProductImage: item.ProductImage,
			Quantity:     item.Quantity,
			PricePerUnit: item.PricePerUnit,
			Subtotal:     item.Subtotal,
		}
		response.Items = append(response.Items, itemResponse)
	}

	return response
}
