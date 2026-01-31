package handler

import (
	"fmt"
	"net/http"
	"zavera/dto"
	"zavera/models"
	"zavera/service"

	"github.com/gin-gonic/gin"
)

type CustomerRefundHandler struct {
	refundService service.RefundService
	orderService  service.OrderService
}

func NewCustomerRefundHandler(refundService service.RefundService, orderService service.OrderService) *CustomerRefundHandler {
	return &CustomerRefundHandler{
		refundService: refundService,
		orderService:  orderService,
	}
}

// Helper function to safely extract user ID from context
// Handles both int and float64 types (JWT tokens store numbers as float64)
func getCustomerUserIDFromContext(c *gin.Context) (int, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, fmt.Errorf("user_id not found in context")
	}
	
	// Try int first
	if id, ok := userID.(int); ok {
		return id, nil
	}
	
	// Try float64 (JWT tokens store numbers as float64)
	if id, ok := userID.(float64); ok {
		return int(id), nil
	}
	
	return 0, fmt.Errorf("user_id has invalid type: %T", userID)
}

// GetOrderRefunds returns all refunds for a customer's order
// GET /customer/orders/:code/refunds
func (h *CustomerRefundHandler) GetOrderRefunds(c *gin.Context) {
	orderCode := c.Param("code")
	
	// Get user ID from JWT token (set by AuthMiddleware)
	customerUserID, err := getCustomerUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "UNAUTHORIZED",
			"message": "Authentication required",
		})
		return
	}
	
	// Get order to verify customer ownership
	orderResp, err := h.orderService.GetOrder(orderCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "ORDER_NOT_FOUND",
			"message": fmt.Sprintf("Order not found: %s", orderCode),
		})
		return
	}
	
	// Get full order model to verify ownership
	order, err := h.orderService.GetOrderByID(orderResp.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "ORDER_NOT_FOUND",
			"message": "Order not found",
		})
		return
	}
	
	// Verify customer owns the order
	if order.UserID == nil || *order.UserID != customerUserID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "FORBIDDEN",
			"message": "You do not have permission to view refunds for this order",
		})
		return
	}
	
	// Get refunds for the order
	refunds, err := h.refundService.GetRefundsByOrder(orderResp.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "INTERNAL_ERROR",
			"message": "Failed to retrieve refunds",
		})
		return
	}
	
	// Convert to customer-friendly response
	customerRefunds := make([]dto.CustomerRefundResponse, 0, len(refunds))
	for _, refund := range refunds {
		customerRefund := h.toCustomerRefundResponse(refund)
		customerRefund.OrderCode = orderCode
		customerRefunds = append(customerRefunds, customerRefund)
	}
	
	c.JSON(http.StatusOK, dto.CustomerRefundListResponse{
		Refunds: customerRefunds,
		Count:   len(customerRefunds),
	})
}

// GetRefundByCode returns refund details by refund code
// GET /customer/refunds/:code
func (h *CustomerRefundHandler) GetRefundByCode(c *gin.Context) {
	refundCode := c.Param("code")
	
	// Get user ID from JWT token (set by AuthMiddleware)
	customerUserID, err := getCustomerUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "UNAUTHORIZED",
			"message": "Authentication required",
		})
		return
	}
	
	// Get refund
	refund, err := h.refundService.GetRefundByCode(refundCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "REFUND_NOT_FOUND",
			"message": fmt.Sprintf("Refund not found: %s", refundCode),
		})
		return
	}
	
	// Get order to verify customer ownership
	order, err := h.orderService.GetOrderByID(refund.OrderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "ORDER_NOT_FOUND",
			"message": "Order not found for this refund",
		})
		return
	}
	
	// Verify customer owns the order
	if order.UserID == nil || *order.UserID != customerUserID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "FORBIDDEN",
			"message": "You do not have permission to view this refund",
		})
		return
	}
	
	// Convert to customer-friendly response
	customerRefund := h.toCustomerRefundResponse(refund)
	customerRefund.OrderCode = order.OrderCode
	
	c.JSON(http.StatusOK, customerRefund)
}

// toCustomerRefundResponse converts a refund model to customer-friendly response
func (h *CustomerRefundHandler) toCustomerRefundResponse(refund *models.Refund) dto.CustomerRefundResponse {
	// Get status label
	statusLabel := h.getStatusLabel(refund.Status)
	
	// Get timeline message
	timeline := h.getTimelineMessage(refund.Status, refund.PaymentID)
	
	// Convert refund items
	items := make([]dto.RefundItemResponse, 0, len(refund.Items))
	for _, item := range refund.Items {
		items = append(items, dto.RefundItemResponse{
			ID:              item.ID,
			RefundID:        item.RefundID,
			OrderItemID:     item.OrderItemID,
			ProductID:       item.ProductID,
			ProductName:     item.ProductName,
			Quantity:        item.Quantity,
			PricePerUnit:    item.PricePerUnit,
			RefundAmount:    item.RefundAmount,
			ItemReason:      item.ItemReason,
			StockRestored:   item.StockRestored,
			StockRestoredAt: item.StockRestoredAt,
			CreatedAt:       item.CreatedAt,
		})
	}
	
	return dto.CustomerRefundResponse{
		RefundCode:     refund.RefundCode,
		OrderCode:      "", // Will be filled by caller if needed
		RefundType:     string(refund.RefundType),
		RefundAmount:   refund.RefundAmount,
		ShippingRefund: refund.ShippingRefund,
		ItemsRefund:    refund.ItemsRefund,
		Status:         string(refund.Status),
		StatusLabel:    statusLabel,
		Timeline:       timeline,
		RequestedAt:    refund.RequestedAt,
		ProcessedAt:    refund.ProcessedAt,
		CompletedAt:    refund.CompletedAt,
		Items:          items,
	}
}

// getStatusLabel returns customer-friendly status label
func (h *CustomerRefundHandler) getStatusLabel(status models.RefundStatus) string {
	switch status {
	case models.RefundStatusPending:
		return "Refund Pending"
	case models.RefundStatusProcessing:
		return "Refund in Progress"
	case models.RefundStatusCompleted:
		return "Refund Completed"
	case models.RefundStatusFailed:
		return "Refund Failed"
	default:
		return string(status)
	}
}

// getTimelineMessage returns customer-friendly timeline message based on status and payment method
func (h *CustomerRefundHandler) getTimelineMessage(status models.RefundStatus, paymentID *int) string {
	switch status {
	case models.RefundStatusPending:
		return "Your refund request has been received and is awaiting processing."
	case models.RefundStatusProcessing:
		// Payment method specific timelines
		if paymentID == nil {
			// Manual refund
			return "Your refund is being processed manually. Please contact support for details."
		}
		// Default timeline for electronic payments
		return "Refund in progress - funds will arrive in 3-7 business days depending on your payment method."
	case models.RefundStatusCompleted:
		if paymentID == nil {
			return "Refund completed - processed manually. Please contact support if you have questions."
		}
		return "Refund completed - funds have been returned to your payment method."
	case models.RefundStatusFailed:
		return "Refund failed - please contact support for assistance."
	default:
		return "Refund status: " + string(status)
	}
}
