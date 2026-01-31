package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"zavera/dto"
	"zavera/models"
	"zavera/service"

	"github.com/gin-gonic/gin"
)

// AdminRefundHandler handles admin refund operations
type AdminRefundHandler struct {
	refundService service.RefundService
}

// NewAdminRefundHandler creates a new admin refund handler
func NewAdminRefundHandler(refundService service.RefundService) *AdminRefundHandler {
	return &AdminRefundHandler{
		refundService: refundService,
	}
}

// Helper function to safely convert user_id from context to int
// Handles both int and float64 types (JWT parsing returns float64)
func getUserIDFromContext(c *gin.Context) (int, error) {
	adminID, exists := c.Get("user_id")
	if !exists {
		return 0, fmt.Errorf("user_id not found in context")
	}

	switch v := adminID.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("invalid user_id type: %T", adminID)
	}
}

// CreateRefund handles POST /admin/refunds
// Creates a new refund for an order
// Validates: Requirements 14.1, 14.6
func (h *AdminRefundHandler) CreateRefund(c *gin.Context) {
	var req dto.RefundRequest
	
	// Validate request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.RefundErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: "Invalid request body",
			Details: map[string]interface{}{
				"validation_error": err.Error(),
			},
		})
		return
	}

	// Get admin user ID from context (set by auth middleware)
	adminIDInt, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.RefundErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "Authentication required",
		})
		return
	}

	requestedBy := &adminIDInt

	// Create refund
	refund, err := h.refundService.CreateRefund(&req, requestedBy)
	if err != nil {
		h.handleError(c, err, req.OrderCode)
		return
	}

	// Format response
	response := h.formatRefundResponse(refund)
	
	c.JSON(http.StatusCreated, response)
}

// ProcessRefund handles POST /admin/refunds/:id/process
// Processes a pending refund through the payment gateway
// Validates: Requirements 14.2, 14.6
func (h *AdminRefundHandler) ProcessRefund(c *gin.Context) {
	refundID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.RefundErrorResponse{
			Error:   "INVALID_REFUND_ID",
			Message: "Refund ID must be a valid integer",
		})
		return
	}

	// Get admin user ID from context
	adminIDInt, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.RefundErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "Authentication required",
		})
		return
	}

	// Process refund
	err = h.refundService.ProcessRefund(refundID, adminIDInt)
	if err != nil {
		h.handleProcessError(c, err, refundID)
		return
	}

	// Get updated refund details
	refund, err := h.refundService.GetRefund(refundID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.RefundErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Failed to retrieve updated refund details",
		})
		return
	}

	// Format success response
	response := dto.RefundSuccessResponse{
		Success:         true,
		Message:         "Refund processed successfully",
		RefundCode:      refund.RefundCode,
		GatewayRefundID: refund.GatewayRefundID,
	}

	c.JSON(http.StatusOK, response)
}

// RetryRefund handles POST /admin/refunds/:id/retry
// Retries a failed refund
// Validates: Requirements 14.2, 14.6
func (h *AdminRefundHandler) RetryRefund(c *gin.Context) {
	refundID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.RefundErrorResponse{
			Error:   "INVALID_REFUND_ID",
			Message: "Refund ID must be a valid integer",
		})
		return
	}

	// Get admin user ID from context
	adminIDInt, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.RefundErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "Authentication required",
		})
		return
	}

	// Get refund to check status
	refund, err := h.refundService.GetRefund(refundID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.RefundErrorResponse{
			Error:   "REFUND_NOT_FOUND",
			Message: fmt.Sprintf("Refund with ID %d not found", refundID),
		})
		return
	}

	// Verify refund is in FAILED status
	if refund.Status != models.RefundStatusFailed {
		c.JSON(http.StatusConflict, dto.RefundErrorResponse{
			Error:   "INVALID_STATUS",
			Message: fmt.Sprintf("Cannot retry refund in status %s, must be FAILED", refund.Status),
			Details: map[string]interface{}{
				"current_status": string(refund.Status),
			},
		})
		return
	}

	// Retry by processing again (uses same idempotency key)
	err = h.refundService.ProcessRefund(refundID, adminIDInt)
	if err != nil {
		h.handleProcessError(c, err, refundID)
		return
	}

	// Get updated refund details
	refund, err = h.refundService.GetRefund(refundID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.RefundErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Failed to retrieve updated refund details",
		})
		return
	}

	// Format success response
	response := dto.RefundSuccessResponse{
		Success:         true,
		Message:         "Refund retry completed successfully",
		RefundCode:      refund.RefundCode,
		GatewayRefundID: refund.GatewayRefundID,
	}

	c.JSON(http.StatusOK, response)
}

// MarkRefundCompleted handles POST /admin/refunds/:id/mark-completed
// Manually marks a refund as completed after manual bank transfer
func (h *AdminRefundHandler) MarkRefundCompleted(c *gin.Context) {
	refundID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.RefundErrorResponse{
			Error:   "INVALID_REFUND_ID",
			Message: "Refund ID must be a valid integer",
		})
		return
	}

	// Get admin user ID from context
	adminIDInt, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.RefundErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "Authentication required",
		})
		return
	}

	// Get confirmation note from request body
	var req struct {
		Note string `json:"note" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.RefundErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: "Note is required to confirm manual refund completion",
		})
		return
	}

	// Mark refund as completed
	err = h.refundService.MarkRefundCompletedManually(refundID, adminIDInt, req.Note)
	if err != nil {
		h.handleProcessError(c, err, refundID)
		return
	}

	// Get updated refund details
	refund, err := h.refundService.GetRefund(refundID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.RefundErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Failed to retrieve updated refund details",
		})
		return
	}

	// Format success response
	response := dto.RefundSuccessResponse{
		Success:         true,
		Message:         "Refund marked as completed successfully",
		RefundCode:      refund.RefundCode,
		GatewayRefundID: refund.GatewayRefundID,
	}

	c.JSON(http.StatusOK, response)
}

// GetRefund handles GET /admin/refunds/:id
// Retrieves detailed refund information including items and status history
// Validates: Requirements 14.3, 14.7
func (h *AdminRefundHandler) GetRefund(c *gin.Context) {
	refundID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.RefundErrorResponse{
			Error:   "INVALID_REFUND_ID",
			Message: "Refund ID must be a valid integer",
		})
		return
	}

	// Get refund with items
	refund, err := h.refundService.GetRefund(refundID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.RefundErrorResponse{
			Error:   "REFUND_NOT_FOUND",
			Message: fmt.Sprintf("Refund with ID %d not found", refundID),
		})
		return
	}

	// Format response with all details
	response := h.formatRefundResponse(refund)
	
	c.JSON(http.StatusOK, response)
}

// GetOrderRefunds handles GET /admin/orders/:code/refunds
// Lists all refunds for a specific order
// Validates: Requirements 14.4, 14.7
func (h *AdminRefundHandler) GetOrderRefunds(c *gin.Context) {
	orderCode := c.Param("code")
	if orderCode == "" {
		c.JSON(http.StatusBadRequest, dto.RefundErrorResponse{
			Error:   "INVALID_ORDER_CODE",
			Message: "Order code is required",
		})
		return
	}

	// Get refunds by order code
	refunds, err := h.refundService.GetRefundsByOrderCode(orderCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.RefundErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Failed to retrieve refunds",
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		})
		return
	}

	// Format response
	responses := make([]dto.RefundResponse, 0, len(refunds))
	for _, refund := range refunds {
		responses = append(responses, h.formatRefundResponse(refund))
	}

	c.JSON(http.StatusOK, responses)
}

// ListRefunds handles GET /admin/refunds
// Lists all refunds with pagination
// Validates: Requirements 14.4, 14.7
func (h *AdminRefundHandler) ListRefunds(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Parse filter parameters
	status := c.Query("status")
	orderCode := c.Query("order_code")

	fmt.Printf("📋 ListRefunds called: page=%d, pageSize=%d, status=%s, orderCode=%s\n", page, pageSize, status, orderCode)

	// Get all refunds from service
	refunds, totalCount, err := h.refundService.ListRefunds(page, pageSize, status, orderCode)
	if err != nil {
		fmt.Printf("❌ ListRefunds error: %v\n", err)
		c.JSON(http.StatusInternalServerError, dto.RefundErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Failed to retrieve refunds",
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		})
		return
	}

	fmt.Printf("✅ ListRefunds found %d refunds (total: %d)\n", len(refunds), totalCount)

	// Format response
	responses := make([]dto.RefundResponse, 0, len(refunds))
	for _, refund := range refunds {
		responses = append(responses, h.formatRefundResponse(refund))
	}

	fmt.Printf("✅ ListRefunds returning %d formatted responses\n", len(responses))

	c.JSON(http.StatusOK, dto.RefundListResponse{
		Refunds:    responses,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	})
}

// Helper methods

// formatRefundResponse converts a refund model to a response DTO
// Validates: Requirements 14.7
func (h *AdminRefundHandler) formatRefundResponse(refund *models.Refund) dto.RefundResponse {
	response := dto.RefundResponse{
		ID:              refund.ID,
		RefundCode:      refund.RefundCode,
		OrderID:         refund.OrderID,
		PaymentID:       refund.PaymentID,
		RefundType:      string(refund.RefundType),
		Reason:          string(refund.Reason),
		ReasonDetail:    refund.ReasonDetail,
		OriginalAmount:  refund.OriginalAmount,
		RefundAmount:    refund.RefundAmount,
		ShippingRefund:  refund.ShippingRefund,
		ItemsRefund:     refund.ItemsRefund,
		Status:          string(refund.Status),
		GatewayRefundID: refund.GatewayRefundID,
		GatewayStatus:   refund.GatewayStatus,
		IdempotencyKey:  refund.IdempotencyKey,
		ProcessedBy:     refund.ProcessedBy,
		ProcessedAt:     refund.ProcessedAt,
		RequestedBy:     refund.RequestedBy,
		RequestedAt:     refund.RequestedAt,
		CompletedAt:     refund.CompletedAt,
		CreatedAt:       refund.CreatedAt,
		UpdatedAt:       refund.UpdatedAt,
	}

	// Add order code from refund service
	response.OrderCode = h.refundService.GetOrderCodeForRefund(refund.OrderID)

	// Format refund items
	if len(refund.Items) > 0 {
		response.Items = make([]dto.RefundItemResponse, len(refund.Items))
		for i, item := range refund.Items {
			response.Items[i] = dto.RefundItemResponse{
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
			}
		}
	}

	// TODO: Add status history
	// response.StatusHistory = ...

	return response
}

// handleError handles refund creation errors
// Validates: Requirements 1.3, 7.6, 15.7
func (h *AdminRefundHandler) handleError(c *gin.Context, err error, orderCode string) {
	switch {
	case err == service.ErrRefundNotFound:
		c.JSON(http.StatusNotFound, dto.RefundErrorResponse{
			Error:   "REFUND_NOT_FOUND",
			Message: "Refund not found",
		})
	
	case err == service.ErrRefundAlreadyExists:
		c.JSON(http.StatusConflict, dto.RefundErrorResponse{
			Error:   "REFUND_ALREADY_EXISTS",
			Message: "A refund already exists for this order",
			Details: map[string]interface{}{
				"order_code": orderCode,
			},
		})
	
	case err == service.ErrRefundNotRefundable:
		c.JSON(http.StatusBadRequest, dto.RefundErrorResponse{
			Error:   "ORDER_NOT_REFUNDABLE",
			Message: err.Error(),
			Details: map[string]interface{}{
				"order_code": orderCode,
			},
		})
	
	case err == service.ErrRefundAmountExceeds:
		c.JSON(http.StatusBadRequest, dto.RefundErrorResponse{
			Error:   "REFUND_AMOUNT_EXCEEDS_BALANCE",
			Message: err.Error(),
			Details: map[string]interface{}{
				"order_code": orderCode,
			},
		})
	
	case err == service.ErrPaymentNotSettled:
		c.JSON(http.StatusBadRequest, dto.RefundErrorResponse{
			Error:   "PAYMENT_NOT_SETTLED",
			Message: err.Error(),
			Details: map[string]interface{}{
				"order_code": orderCode,
			},
		})
	
	case err == service.ErrIdempotencyConflict:
		c.JSON(http.StatusConflict, dto.RefundErrorResponse{
			Error:   "IDEMPOTENCY_CONFLICT",
			Message: "Idempotency key already used",
		})
	
	case err == sql.ErrNoRows:
		c.JSON(http.StatusNotFound, dto.RefundErrorResponse{
			Error:   "ORDER_NOT_FOUND",
			Message: fmt.Sprintf("Order %s not found", orderCode),
			Details: map[string]interface{}{
				"order_code": orderCode,
			},
		})
	
	default:
		// Check for validation errors (string contains checks)
		errMsg := err.Error()
		if contains(errMsg, "not found") {
			c.JSON(http.StatusNotFound, dto.RefundErrorResponse{
				Error:   "RESOURCE_NOT_FOUND",
				Message: errMsg,
			})
		} else if contains(errMsg, "must be positive") || contains(errMsg, "exceeds") || contains(errMsg, "invalid") {
			c.JSON(http.StatusBadRequest, dto.RefundErrorResponse{
				Error:   "VALIDATION_ERROR",
				Message: errMsg,
				Details: map[string]interface{}{
					"order_code": orderCode,
				},
			})
		} else {
			c.JSON(http.StatusInternalServerError, dto.RefundErrorResponse{
				Error:   "INTERNAL_ERROR",
				Message: "Failed to create refund",
				Details: map[string]interface{}{
					"error": errMsg,
				},
			})
		}
	}
}

// handleProcessError handles refund processing errors
func (h *AdminRefundHandler) handleProcessError(c *gin.Context, err error, refundID int) {
	switch {
	case err == service.ErrRefundNotFound:
		c.JSON(http.StatusNotFound, dto.RefundErrorResponse{
			Error:   "REFUND_NOT_FOUND",
			Message: fmt.Sprintf("Refund with ID %d not found", refundID),
		})
	
	case err == service.ErrRefundAlreadyFinal:
		c.JSON(http.StatusConflict, dto.RefundErrorResponse{
			Error:   "REFUND_ALREADY_FINAL",
			Message: "Refund is already in a final state (COMPLETED or FAILED)",
			Details: map[string]interface{}{
				"refund_id": refundID,
			},
		})
	
	default:
		errMsg := err.Error()
		if contains(errMsg, "midtrans") || contains(errMsg, "gateway") {
			c.JSON(http.StatusBadGateway, dto.RefundErrorResponse{
				Error:   "GATEWAY_ERROR",
				Message: "Payment gateway error",
				Details: map[string]interface{}{
					"error":     errMsg,
					"refund_id": refundID,
				},
			})
		} else if contains(errMsg, "cannot be processed") {
			c.JSON(http.StatusConflict, dto.RefundErrorResponse{
				Error:   "INVALID_STATUS",
				Message: errMsg,
				Details: map[string]interface{}{
					"refund_id": refundID,
				},
			})
		} else {
			c.JSON(http.StatusInternalServerError, dto.RefundErrorResponse{
				Error:   "INTERNAL_ERROR",
				Message: "Failed to process refund",
				Details: map[string]interface{}{
					"error":     errMsg,
					"refund_id": refundID,
				},
			})
		}
	}
}

// contains checks if a string contains a substring (case-insensitive helper)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
			(s[:len(substr)] == substr || 
			 s[len(s)-len(substr):] == substr ||
			 containsMiddle(s, substr))))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
