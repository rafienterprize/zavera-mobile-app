package handler

import (
	"net/http"
	"zavera/dto"
	"zavera/models"
	"zavera/service"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderService    service.OrderService
	paymentService  service.PaymentService
	shippingService service.ShippingService
}

func NewOrderHandler(orderService service.OrderService, paymentService service.PaymentService, shippingService ...service.ShippingService) *OrderHandler {
	h := &OrderHandler{
		orderService:   orderService,
		paymentService: paymentService,
	}
	if len(shippingService) > 0 {
		h.shippingService = shippingService[0]
	}
	return h
}

// getSessionID gets session ID from cookie
func (h *OrderHandler) getSessionID(c *gin.Context) string {
	sessionID, _ := c.Cookie("session_id")
	return sessionID
}

// Checkout godoc
// @Summary Checkout cart
// @Description Create order from cart and get payment token
// @Tags orders
// @Accept json
// @Produce json
// @Param request body dto.CheckoutRequest true "Checkout request"
// @Success 200 {object} map[string]interface{}
// @Router /api/checkout [post]
func (h *OrderHandler) Checkout(c *gin.Context) {
	sessionID := h.getSessionID(c)
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "session_required",
			Message: "Session ID not found",
		})
		return
	}

	var req dto.CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Get user_id from JWT if authenticated (optional)
	var userID *int
	if uid, exists := c.Get("user_id"); exists {
		id := int(uid.(float64))
		userID = &id
	}

	// Create order
	checkoutResp, err := h.orderService.CreateOrder(sessionID, req, userID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "cart is empty" {
			status = http.StatusBadRequest
		}

		c.JSON(status, dto.ErrorResponse{
			Error:   "checkout_failed",
			Message: err.Error(),
		})
		return
	}

	// Create payment
	snapToken, err := h.paymentService.CreatePayment(checkoutResp.OrderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "payment_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order":      checkoutResp,
		"snap_token": snapToken,
	})
}

// GetOrder godoc
// @Summary Get order by code
// @Description Get order details by order code
// @Tags orders
// @Accept json
// @Produce json
// @Param code path string true "Order Code"
// @Success 200 {object} dto.OrderResponse
// @Router /api/orders/{code} [get]
func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderCode := c.Param("code")

	order, err := h.orderService.GetOrder(orderCode)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Message: "Order not found",
		})
		return
	}

	// SECURITY FIX: Verify access rights to this order
	
	// Check if user is authenticated
	userIDRaw, isAuthenticated := c.Get("user_id")
	
	// Helper to get user ID as int
	getUserID := func() int {
		if userIDRaw == nil {
			return 0
		}
		switch v := userIDRaw.(type) {
		case float64:
			return int(v)
		case int:
			return v
		case int64:
			return int(v)
		default:
			return 0
		}
	}
	
	// If order belongs to a registered user
	if order.UserID != nil {
		if !isAuthenticated {
			// Order belongs to a user but requester is not authenticated
			// Still allow access with email verification for order tracking
			emailParam := c.Query("email")
			if emailParam == "" || emailParam != order.CustomerEmail {
				c.JSON(http.StatusForbidden, dto.ErrorResponse{
					Error:   "access_denied",
					Message: "Please provide your email to view this order",
				})
				return
			}
		} else {
			// User is authenticated - verify ownership
			uid := getUserID()
			if *order.UserID != uid {
				// Order belongs to different user - require email verification
				emailParam := c.Query("email")
				if emailParam == "" || emailParam != order.CustomerEmail {
					c.JSON(http.StatusForbidden, dto.ErrorResponse{
						Error:   "access_denied",
						Message: "You don't have permission to view this order",
					})
					return
				}
			}
		}
	} else {
		// Guest order (user_id = NULL) - require email verification for security
		emailParam := c.Query("email")
		if emailParam == "" || emailParam != order.CustomerEmail {
			// Allow access by order code only (common e-commerce pattern)
			// But mask sensitive data for unauthenticated requests
			order.CustomerPhone = maskPhone(order.CustomerPhone)
			order.CustomerEmail = maskEmail(order.CustomerEmail)
		}
	}

	// Get shipment info if available
	if h.shippingService != nil {
		shipment, err := h.shippingService.GetShipmentByOrderCode(orderCode)
		if err == nil && shipment != nil {
			order.Shipment = &dto.ShipmentResponse{
				ID:             shipment.ID,
				OrderID:        shipment.OrderID,
				ProviderCode:   shipment.ProviderCode,
				ProviderName:   shipment.ProviderName,
				ServiceCode:    shipment.ServiceCode,
				ServiceName:    shipment.ServiceName,
				Cost:           shipment.Cost,
				ETD:            shipment.ETD,
				Status:         string(shipment.Status),
				TrackingNumber: shipment.TrackingNumber,
			}
		}
	}

	c.JSON(http.StatusOK, order)
}

// GetOrderByID godoc
// @Summary Get order by ID
// @Description Get order details by order ID (for payment flow)
// @Tags orders
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} dto.OrderResponse
// @Router /api/orders/id/{id} [get]
func (h *OrderHandler) GetOrderByID(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID := 0
	for _, ch := range orderIDStr {
		if ch >= '0' && ch <= '9' {
			orderID = orderID*10 + int(ch-'0')
		} else {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "invalid_id",
				Message: "Order ID must be a number",
			})
			return
		}
	}

	order, err := h.orderService.GetOrderByID(orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Message: "Order not found",
		})
		return
	}

	// Check if user is authenticated and owns this order
	userIDRaw, isAuthenticated := c.Get("user_id")
	if isAuthenticated && order.UserID != nil {
		var uid int
		switch v := userIDRaw.(type) {
		case float64:
			uid = int(v)
		case int:
			uid = v
		}
		if *order.UserID != uid {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Error:   "access_denied",
				Message: "You don't have permission to view this order",
			})
			return
		}
	}

	// Build response with order items
	items := h.getOrderItems(order.ID)

	c.JSON(http.StatusOK, gin.H{
		"id":             order.ID,
		"order_code":     order.OrderCode,
		"total_amount":   order.TotalAmount,
		"status":         order.Status,
		"customer_name":  order.CustomerName,
		"customer_email": order.CustomerEmail,
		"customer_phone": order.CustomerPhone,
		"items":          items,
		"created_at":     order.CreatedAt,
	})
}

// getOrderItems fetches order items for an order
func (h *OrderHandler) getOrderItems(orderID int) []map[string]interface{} {
	items, err := h.orderService.GetOrderItems(orderID)
	if err != nil {
		return []map[string]interface{}{}
	}
	
	result := make([]map[string]interface{}, len(items))
	for i, item := range items {
		result[i] = map[string]interface{}{
			"product_name": item.ProductName,
			"quantity":     item.Quantity,
			"price":        item.PricePerUnit,
			"subtotal":     item.Subtotal,
			"image_url":    item.ProductImage,
		}
	}
	return result
}

// maskPhone masks phone number for privacy
func maskPhone(phone string) string {
	if len(phone) <= 4 {
		return "****"
	}
	return phone[:2] + "****" + phone[len(phone)-2:]
}

// maskEmail masks email for privacy
func maskEmail(email string) string {
	atIndex := -1
	for i, c := range email {
		if c == '@' {
			atIndex = i
			break
		}
	}
	if atIndex <= 2 {
		return "***" + email[atIndex:]
	}
	return email[:2] + "***" + email[atIndex:]
}

// PaymentCallback godoc
// @Summary Midtrans payment callback
// @Description Handle payment notification from Midtrans
// @Tags payments
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/payment/callback [post]
func (h *OrderHandler) PaymentCallback(c *gin.Context) {
	var notification map[string]interface{}
	if err := c.ShouldBindJSON(&notification); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	orderCode, ok := notification["order_id"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_order_id",
			Message: "Order ID not found in notification",
		})
		return
	}

	transactionStatus, ok := notification["transaction_status"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_status",
			Message: "Transaction status not found",
		})
		return
	}

	transactionID, _ := notification["transaction_id"].(string)

	// Map Midtrans status to our payment status
	var paymentStatus models.PaymentStatus
	switch transactionStatus {
	case "capture", "settlement":
		paymentStatus = models.PaymentStatusSuccess
	case "pending":
		paymentStatus = models.PaymentStatusPending
	case "deny", "cancel":
		paymentStatus = models.PaymentStatusCancelled
	case "expire":
		paymentStatus = models.PaymentStatusExpired
	default:
		paymentStatus = models.PaymentStatusFailed
	}

	// Handle callback
	err := h.paymentService.HandleCallback(orderCode, paymentStatus, transactionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "callback_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Callback processed successfully"})
}
