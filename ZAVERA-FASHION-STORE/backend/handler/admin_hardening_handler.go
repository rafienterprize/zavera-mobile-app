package handler

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"
	"zavera/dto"
	"zavera/models"
	"zavera/service"

	"github.com/gin-gonic/gin"
)

type AdminHardeningHandler struct {
	adminService         service.AdminService
	refundService        service.RefundService
	recoveryService      service.PaymentRecoveryService
	reconciliationService service.ReconciliationService
	db                   *sql.DB
}

func NewAdminHardeningHandler(
	adminService service.AdminService,
	refundService service.RefundService,
	recoveryService service.PaymentRecoveryService,
	reconciliationService service.ReconciliationService,
	db *sql.DB,
) *AdminHardeningHandler {
	return &AdminHardeningHandler{
		adminService:         adminService,
		refundService:        refundService,
		recoveryService:      recoveryService,
		reconciliationService: reconciliationService,
		db:                   db,
	}
}

// Helper to extract admin context from request
func (h *AdminHardeningHandler) getAdminContext(c *gin.Context) *service.AdminContext {
	// Get user from auth middleware
	userID, _ := c.Get("user_id")
	email, _ := c.Get("user_email")

	uid := 0
	if id, ok := userID.(int); ok {
		uid = id
	}

	emailStr := ""
	if e, ok := email.(string); ok {
		emailStr = e
	}

	return &service.AdminContext{
		UserID:    uid,
		Email:     emailStr,
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}
}

// ============================================
// FORCE ACTIONS
// ============================================

// ForceCancel godoc
// @Summary Force cancel an order
// @Description Admin force cancels an order with audit logging
// @Tags admin
// @Accept json
// @Produce json
// @Param code path string true "Order code"
// @Param request body dto.ForceCancelRequest true "Cancel request"
// @Success 200 {object} dto.AdminActionResponse
// @Router /api/admin/orders/{code}/force-cancel [post]
func (h *AdminHardeningHandler) ForceCancel(c *gin.Context) {
	orderCode := c.Param("code")
	
	var req dto.ForceCancelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ ForceCancel binding error for %s: %v", orderCode, err)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	log.Printf("📝 ForceCancel request for %s: reason=%s, restore_stock=%v", orderCode, req.Reason, req.RestoreStock)

	admin := h.getAdminContext(c)
	resp, err := h.adminService.ForceCancel(orderCode, &req, admin)
	if err != nil {
		log.Printf("❌ ForceCancel service error for %s: %v", orderCode, err)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "force_cancel_failed",
			Message: err.Error(),
		})
		return
	}

	log.Printf("✅ ForceCancel success for %s", orderCode)
	c.JSON(http.StatusOK, resp)
}

// ForceRefund godoc
// @Summary Force refund an order
// @Description Admin force refunds an order with audit logging
// @Tags admin
// @Accept json
// @Produce json
// @Param code path string true "Order code"
// @Param request body dto.ForceRefundRequest true "Refund request"
// @Success 200 {object} dto.AdminActionResponse
// @Router /api/admin/orders/{code}/refund [post]
func (h *AdminHardeningHandler) ForceRefund(c *gin.Context) {
	orderCode := c.Param("code")
	
	var req dto.ForceRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	admin := h.getAdminContext(c)
	resp, err := h.adminService.ForceRefund(orderCode, &req, admin)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "force_refund_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ForceReship godoc
// @Summary Force reship an order
// @Description Admin creates a replacement shipment with audit logging
// @Tags admin
// @Accept json
// @Produce json
// @Param code path string true "Order code"
// @Param request body dto.ForceReshipRequest true "Reship request"
// @Success 200 {object} dto.AdminActionResponse
// @Router /api/admin/orders/{code}/reship [post]
func (h *AdminHardeningHandler) ForceReship(c *gin.Context) {
	orderCode := c.Param("code")
	
	var req dto.ForceReshipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	admin := h.getAdminContext(c)
	resp, err := h.adminService.ForceReship(orderCode, &req, admin)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "force_reship_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ReconcilePayment godoc
// @Summary Reconcile a payment
// @Description Admin manually reconciles a payment with audit logging
// @Tags admin
// @Accept json
// @Produce json
// @Param id path int true "Payment ID"
// @Param request body dto.ReconcilePaymentRequest true "Reconcile request"
// @Success 200 {object} dto.AdminActionResponse
// @Router /api/admin/payments/{id}/reconcile [post]
func (h *AdminHardeningHandler) ReconcilePayment(c *gin.Context) {
	paymentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_payment_id",
			Message: "Payment ID must be a number",
		})
		return
	}
	
	var req dto.ReconcilePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	admin := h.getAdminContext(c)
	resp, err := h.adminService.ReconcilePayment(paymentID, &req, admin)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "reconcile_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ============================================
// REFUND ENDPOINTS
// ============================================

// CreateRefund godoc
// @Summary Create a refund
// @Description Create a new refund request
// @Tags admin
// @Accept json
// @Produce json
// @Param request body dto.RefundRequest true "Refund request"
// @Success 200 {object} dto.RefundResponse
// @Router /api/admin/refunds [post]
func (h *AdminHardeningHandler) CreateRefund(c *gin.Context) {
	var req dto.RefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	admin := h.getAdminContext(c)
	refund, err := h.refundService.CreateRefund(&req, &admin.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "refund_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, h.mapRefundToResponse(refund))
}

// GetRefund godoc
// @Summary Get refund details
// @Description Get refund by ID
// @Tags admin
// @Produce json
// @Param id path int true "Refund ID"
// @Success 200 {object} dto.RefundResponse
// @Router /api/admin/refunds/{id} [get]
func (h *AdminHardeningHandler) GetRefund(c *gin.Context) {
	idStr := c.Param("id")
	refundID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_refund_id",
			Message: "Refund ID must be a number",
		})
		return
	}
	
	refund, err := h.refundService.GetRefund(refundID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "refund_not_found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, h.mapRefundToResponse(refund))
}

// ProcessRefund godoc
// @Summary Process a pending refund
// @Description Process refund with payment gateway
// @Tags admin
// @Produce json
// @Param id path int true "Refund ID"
// @Success 200 {object} dto.RefundResponse
// @Router /api/admin/refunds/{id}/process [post]
func (h *AdminHardeningHandler) ProcessRefund(c *gin.Context) {
	idStr := c.Param("id")
	refundID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_refund_id",
			Message: "Refund ID must be a number",
		})
		return
	}
	
	refund, err := h.refundService.GetRefund(refundID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "refund_not_found",
			Message: err.Error(),
		})
		return
	}

	admin := h.getAdminContext(c)
	if err := h.refundService.ProcessRefund(refund.ID, admin.UserID); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "process_failed",
			Message: err.Error(),
		})
		return
	}

	// Get updated refund
	refund, _ = h.refundService.GetRefund(refund.ID)
	c.JSON(http.StatusOK, h.mapRefundToResponse(refund))
}

// ============================================
// PAYMENT RECOVERY ENDPOINTS
// ============================================

// SyncPayment godoc
// @Summary Sync payment status with gateway
// @Description Sync a single payment with Midtrans
// @Tags admin
// @Produce json
// @Param id path int true "Payment ID"
// @Success 200 {object} dto.PaymentSyncResponse
// @Router /api/admin/payments/{id}/sync [post]
func (h *AdminHardeningHandler) SyncPayment(c *gin.Context) {
	paymentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_payment_id",
			Message: "Payment ID must be a number",
		})
		return
	}

	resp, err := h.recoveryService.SyncPaymentStatus(paymentID, "manual")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "sync_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetStuckPayments godoc
// @Summary Get stuck payments
// @Description Get payments stuck in pending state
// @Tags admin
// @Produce json
// @Param hours query int false "Hours threshold (default 2)"
// @Success 200 {array} dto.StuckPaymentResponse
// @Router /api/admin/payments/stuck [get]
func (h *AdminHardeningHandler) GetStuckPayments(c *gin.Context) {
	hours := 2
	if h := c.Query("hours"); h != "" {
		if parsed, err := strconv.Atoi(h); err == nil {
			hours = parsed
		}
	}

	stuck, err := h.recoveryService.FindStuckPayments(hours)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "query_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stuck)
}

// RunPaymentSync godoc
// @Summary Run payment sync job
// @Description Sync all pending payments with gateway
// @Tags admin
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/admin/payments/sync-all [post]
func (h *AdminHardeningHandler) RunPaymentSync(c *gin.Context) {
	synced, errors, err := h.recoveryService.SyncAllPendingPayments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "sync_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"synced": synced,
		"errors": errors,
	})
}

// ============================================
// RECONCILIATION ENDPOINTS
// ============================================

// RunReconciliation godoc
// @Summary Run daily reconciliation
// @Description Run reconciliation for a specific date
// @Tags admin
// @Accept json
// @Produce json
// @Param request body dto.ReconciliationRequest true "Reconciliation request"
// @Success 200 {object} dto.ReconciliationSummary
// @Router /api/admin/reconciliation/run [post]
func (h *AdminHardeningHandler) RunReconciliation(c *gin.Context) {
	var req dto.ReconciliationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_date",
			Message: "Date must be in YYYY-MM-DD format",
		})
		return
	}

	admin := h.getAdminContext(c)
	summary, err := h.reconciliationService.RunDailyReconciliation(date, admin.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "reconciliation_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetReconciliation godoc
// @Summary Get reconciliation summary
// @Description Get reconciliation summary for a date
// @Tags admin
// @Produce json
// @Param date query string true "Date (YYYY-MM-DD)"
// @Success 200 {object} dto.ReconciliationSummary
// @Router /api/admin/reconciliation [get]
func (h *AdminHardeningHandler) GetReconciliation(c *gin.Context) {
	dateStr := c.Query("date")
	if dateStr == "" {
		dateStr = time.Now().Add(-24 * time.Hour).Format("2006-01-02")
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_date",
			Message: "Date must be in YYYY-MM-DD format",
		})
		return
	}

	summary, err := h.reconciliationService.GetReconciliationSummary(date)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Message: "No reconciliation found for this date",
		})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetMismatches godoc
// @Summary Get unresolved mismatches
// @Description Get all unresolved payment/order mismatches
// @Tags admin
// @Produce json
// @Success 200 {array} dto.MismatchDetail
// @Router /api/admin/reconciliation/mismatches [get]
func (h *AdminHardeningHandler) GetMismatches(c *gin.Context) {
	mismatches, err := h.reconciliationService.GetUnresolvedMismatches()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "query_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, mismatches)
}

// ============================================
// AUDIT LOG ENDPOINTS
// ============================================

// GetAuditLogs godoc
// @Summary Get audit logs
// @Description Get recent admin audit logs with pagination
// @Tags admin
// @Produce json
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Limit (default 50)"
// @Success 200 {object} map[string]interface{}
// @Router /api/admin/audit-logs [get]
func (h *AdminHardeningHandler) GetAuditLogs(c *gin.Context) {
	page := 1
	limit := 50
	
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	// Get total count
	var total int
	err := h.db.QueryRow("SELECT COUNT(*) FROM admin_audit_log").Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "query_failed",
			Message: err.Error(),
		})
		return
	}

	// Get paginated logs
	offset := (page - 1) * limit
	var logs []*models.AdminAuditLog
	
	rows, err := h.db.Query(`
		SELECT id, admin_user_id, admin_email, admin_ip, admin_user_agent,
		       action_type, action_detail, target_type, target_id, target_code,
		       state_before, state_after, success, error_message, 
		       idempotency_key, metadata, created_at
		FROM admin_audit_log
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "query_failed",
			Message: err.Error(),
		})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var log models.AdminAuditLog
		var stateBefore, stateAfter, metadata sql.NullString
		var errorMessage sql.NullString
		var idempotencyKey sql.NullString
		
		err := rows.Scan(
			&log.ID, &log.AdminUserID, &log.AdminEmail, &log.AdminIP, &log.AdminUserAgent,
			&log.ActionType, &log.ActionDetail, &log.TargetType, &log.TargetID, &log.TargetCode,
			&stateBefore, &stateAfter, &log.Success, &errorMessage,
			&idempotencyKey, &metadata, &log.CreatedAt,
		)
		
		if err == nil {
			// Parse JSON fields - they're stored as JSONB in database
			// For simplicity, we'll leave them as empty maps if parsing fails
			log.StateBefore = make(map[string]any)
			log.StateAfter = make(map[string]any)
			log.Metadata = make(map[string]any)
			
			if errorMessage.Valid {
				log.ErrorMessage = errorMessage.String
			}
			if idempotencyKey.Valid {
				log.IdempotencyKey = idempotencyKey.String
			}
			logs = append(logs, &log)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetDashboardMetrics godoc
// @Summary Get dashboard metrics
// @Description Get real-time dashboard metrics
// @Tags admin
// @Produce json
// @Success 200 {object} dto.DashboardMetricsResponse
// @Router /api/admin/dashboard/metrics [get]
func (h *AdminHardeningHandler) GetDashboardMetrics(c *gin.Context) {
	metrics, err := h.adminService.GetDashboardMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "query_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// Helper to map refund to response
func (h *AdminHardeningHandler) mapRefundToResponse(refund *models.Refund) dto.RefundResponse {
	resp := dto.RefundResponse{
		ID:             refund.ID,
		RefundCode:     refund.RefundCode,
		RefundType:     string(refund.RefundType),
		Reason:         string(refund.Reason),
		OriginalAmount: refund.OriginalAmount,
		RefundAmount:   refund.RefundAmount,
		ShippingRefund: refund.ShippingRefund,
		ItemsRefund:    refund.ItemsRefund,
		Status:         string(refund.Status),
		GatewayStatus:  refund.GatewayStatus,
		CreatedAt:      refund.CreatedAt,
	}

	if refund.CompletedAt != nil {
		resp.CompletedAt = refund.CompletedAt
	}

	for _, item := range refund.Items {
		resp.Items = append(resp.Items, dto.RefundItemResponse{
			ProductID:     item.ProductID,
			ProductName:   item.ProductName,
			Quantity:      item.Quantity,
			PricePerUnit:  item.PricePerUnit,
			RefundAmount:  item.RefundAmount,
			StockRestored: item.StockRestored,
		})
	}

	return resp
}
