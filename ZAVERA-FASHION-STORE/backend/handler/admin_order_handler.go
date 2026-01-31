package handler

import (
	"net/http"
	"strconv"
	"zavera/dto"
	"zavera/service"

	"github.com/gin-gonic/gin"
)

type AdminOrderHandler struct {
	orderService service.AdminOrderService
}

func NewAdminOrderHandler(orderService service.AdminOrderService) *AdminOrderHandler {
	return &AdminOrderHandler{
		orderService: orderService,
	}
}

// GetAllOrders returns all orders for admin with filtering
// GET /api/admin/orders
func (h *AdminOrderHandler) GetAllOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")
	search := c.Query("search")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	filter := dto.AdminOrderFilter{
		Page:     page,
		PageSize: pageSize,
		Status:   status,
		Search:   search,
		DateFrom: dateFrom,
		DateTo:   dateTo,
	}

	orders, total, err := h.orderService.GetAllOrdersAdmin(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "server_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"orders":      orders,
		"total_count": total,
		"page":        page,
		"page_size":   pageSize,
	})
}

// GetOrderDetail returns detailed order info including payment and shipment
// GET /api/admin/orders/:code
func (h *AdminOrderHandler) GetOrderDetail(c *gin.Context) {
	orderCode := c.Param("code")

	order, err := h.orderService.GetOrderDetailAdmin(orderCode)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Message: "Order not found",
		})
		return
	}

	c.JSON(http.StatusOK, order)
}

// UpdateOrderStatus updates order status (admin override)
// PATCH /api/admin/orders/:code/status
func (h *AdminOrderHandler) UpdateOrderStatus(c *gin.Context) {
	orderCode := c.Param("code")

	var req dto.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Get admin context
	adminEmail, _ := c.Get("user_email")
	email := ""
	if e, ok := adminEmail.(string); ok {
		email = e
	}

	err := h.orderService.UpdateOrderStatusAdmin(orderCode, req.Status, req.Reason, email)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order status updated"})
}

// GetOrderStats returns order statistics for dashboard
// GET /api/admin/orders/stats
func (h *AdminOrderHandler) GetOrderStats(c *gin.Context) {
	stats, err := h.orderService.GetOrderStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "server_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// PackOrder marks an order as being packed
// POST /api/admin/orders/:code/pack
func (h *AdminOrderHandler) PackOrder(c *gin.Context) {
	orderCode := c.Param("code")

	// Get admin context
	adminEmail, _ := c.Get("user_email")
	email := ""
	if e, ok := adminEmail.(string); ok {
		email = e
	}

	err := h.orderService.PackOrder(orderCode, email)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "pack_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order marked as packing",
		"status":  "PACKING",
	})
}

// GenerateResi generates resi from Biteship without shipping the order yet
// POST /api/admin/orders/:code/generate-resi
func (h *AdminOrderHandler) GenerateResi(c *gin.Context) {
	orderCode := c.Param("code")

	// Get admin context
	adminEmail, _ := c.Get("user_email")
	email := ""
	if e, ok := adminEmail.(string); ok {
		email = e
	}

	// Call service to generate resi
	resi, err := h.orderService.GenerateResiOnly(orderCode, email)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "generate_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Resi generated successfully",
		"resi":       resi,
		"waybill_id": resi,
	})
}

// ShipOrder marks an order as shipped with resi
// POST /api/admin/orders/:code/ship
func (h *AdminOrderHandler) ShipOrder(c *gin.Context) {
	orderCode := c.Param("code")

	var req dto.ShipOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Resi is required",
		})
		return
	}

	// Validate resi is provided
	if req.Resi == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "resi_required",
			Message: "Nomor resi harus diisi. Gunakan endpoint /generate-resi untuk mendapatkan resi dari Biteship.",
		})
		return
	}

	// Get admin context
	adminEmail, _ := c.Get("user_email")
	email := ""
	if e, ok := adminEmail.(string); ok {
		email = e
	}

	resi, err := h.orderService.ShipOrder(orderCode, req.Resi, email)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "ship_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order shipped successfully",
		"status":  "SHIPPED",
		"resi":    resi,
	})
}

// GetOrderActions returns available actions for an order based on its status
// GET /api/admin/orders/:code/actions
func (h *AdminOrderHandler) GetOrderActions(c *gin.Context) {
	orderCode := c.Param("code")

	actions, err := h.orderService.GetOrderActions(orderCode)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Message: "Order not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order_code": orderCode,
		"actions":    actions,
	})
}

// CancelOrder cancels an order (admin)
// POST /api/admin/orders/:code/cancel
func (h *AdminOrderHandler) CancelOrder(c *gin.Context) {
	orderCode := c.Param("code")

	var req dto.CancelOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Reason is required",
		})
		return
	}

	// Get admin context
	adminEmail, _ := c.Get("user_email")
	email := ""
	if e, ok := adminEmail.(string); ok {
		email = e
	}

	err := h.orderService.CancelOrderAdmin(orderCode, req.Reason, email)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "cancel_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order cancelled successfully",
		"status":  "CANCELLED",
	})
}

// DeliverOrder marks an order as delivered
// POST /api/admin/orders/:code/deliver
func (h *AdminOrderHandler) DeliverOrder(c *gin.Context) {
	orderCode := c.Param("code")

	// Get admin context
	adminEmail, _ := c.Get("user_email")
	email := ""
	if e, ok := adminEmail.(string); ok {
		email = e
	}

	err := h.orderService.DeliverOrder(orderCode, email)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "deliver_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order marked as delivered",
		"status":  "DELIVERED",
	})
}
