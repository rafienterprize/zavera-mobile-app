package handler

import (
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
	"zavera/dto"
	"zavera/repository"
	"zavera/service"

	"github.com/gin-gonic/gin"
)

// CorePaymentHandler handles Tokopedia-style VA payment endpoints
type CorePaymentHandler struct {
	corePaymentService service.CorePaymentService
	rateLimiter        *RateLimiter
}

// RateLimiter implements simple rate limiting for status check
type RateLimiter struct {
	mu       sync.Mutex
	requests map[string]time.Time
	interval time.Duration
}

func NewRateLimiter(interval time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string]time.Time),
		interval: interval,
	}
}

func (r *RateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	lastRequest, exists := r.requests[key]
	if exists && time.Since(lastRequest) < r.interval {
		return false
	}
	r.requests[key] = time.Now()
	return true
}

// NewCorePaymentHandler creates a new CorePaymentHandler
func NewCorePaymentHandler(corePaymentService service.CorePaymentService) *CorePaymentHandler {
	return &CorePaymentHandler{
		corePaymentService: corePaymentService,
		rateLimiter:        NewRateLimiter(5 * time.Second), // 5 second rate limit
	}
}

// CreateVAPaymentRequest represents the request to create VA payment
type CreateVAPaymentRequest struct {
	OrderID       int    `json:"order_id" binding:"required"`
	PaymentMethod string `json:"payment_method" binding:"required,oneof=bca_va bri_va mandiri_va permata_va bni_va gopay qris credit_card"`
}

// CreateVAPayment creates a VA payment via Midtrans Core API
// POST /api/payments/core/create
func (h *CorePaymentHandler) CreateVAPayment(c *gin.Context) {
	log.Printf("💳 CreateVAPayment called")

	var req CreateVAPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ Bind error: %v", err)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	log.Printf("📦 Request: order_id=%d, payment_method=%s", req.OrderID, req.PaymentMethod)

	// Create VA payment
	response, err := h.corePaymentService.CreateVAPayment(req.OrderID, req.PaymentMethod)
	if err != nil {
		log.Printf("❌ CreateVAPayment error: %v", err)
		
		status := http.StatusInternalServerError
		errorCode := "payment_creation_failed"
		
		switch err {
		case service.ErrOrderNotFound:
			status = http.StatusNotFound
			errorCode = "order_not_found"
		case service.ErrOrderNotPendingPayment:
			status = http.StatusBadRequest
			errorCode = "order_not_pending"
		case service.ErrPaymentMethodInvalid:
			status = http.StatusBadRequest
			errorCode = "invalid_payment_method"
		case repository.ErrPaymentExpired:
			status = http.StatusGone
			errorCode = "payment_expired"
		case service.ErrMidtransTimeout:
			status = http.StatusGatewayTimeout
			errorCode = "payment_gateway_timeout"
		case service.ErrMidtransAPIError:
			status = http.StatusBadGateway
			errorCode = "payment_gateway_error"
		}

		c.JSON(status, dto.ErrorResponse{
			Error:   errorCode,
			Message: err.Error(),
		})
		return
	}

	log.Printf("✅ VA Payment created: payment_id=%d, va=%s", response.PaymentID, response.VANumber)
	c.JSON(http.StatusCreated, response)
}

// GetPaymentDetails gets payment details for an order
// GET /api/payments/core/:order_id
func (h *CorePaymentHandler) GetPaymentDetails(c *gin.Context) {
	orderIDStr := c.Param("order_id")
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_order_id",
			Message: "Order ID must be a number",
		})
		return
	}

	log.Printf("🔍 GetPaymentDetails: order_id=%d", orderID)

	// Get payment details
	response, err := h.corePaymentService.GetPaymentByOrderID(orderID)
	if err != nil {
		log.Printf("❌ GetPaymentDetails error: %v", err)
		
		status := http.StatusInternalServerError
		errorCode := "payment_retrieval_failed"
		
		switch err {
		case service.ErrOrderNotFound:
			status = http.StatusNotFound
			errorCode = "order_not_found"
		case repository.ErrPaymentNotFound:
			status = http.StatusNotFound
			errorCode = "payment_not_found"
		case repository.ErrPaymentExpired:
			status = http.StatusGone
			errorCode = "payment_expired"
		}

		c.JSON(status, dto.ErrorResponse{
			Error:   errorCode,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CheckPaymentStatusRequest represents the request to check payment status
type CheckPaymentStatusRequest struct {
	PaymentID int `json:"payment_id" binding:"required"`
}

// CheckPaymentStatus checks current payment status from database
// POST /api/payments/core/check
func (h *CorePaymentHandler) CheckPaymentStatus(c *gin.Context) {
	var req CheckPaymentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Rate limiting - max 1 request per 5 seconds per user
	userIDRaw, _ := c.Get("user_id")
	var userID int
	switch v := userIDRaw.(type) {
	case float64:
		userID = int(v)
	case int:
		userID = v
	default:
		userID = 0 // Fallback for rate limiting key
	}
	rateLimitKey := "status_check_" + strconv.Itoa(userID) + "_" + strconv.Itoa(req.PaymentID)
	
	if !h.rateLimiter.Allow(rateLimitKey) {
		c.JSON(http.StatusTooManyRequests, dto.ErrorResponse{
			Error:   "rate_limited",
			Message: "Silakan tunggu beberapa detik sebelum cek status lagi",
		})
		return
	}

	log.Printf("🔍 CheckPaymentStatus: payment_id=%d", req.PaymentID)

	response, err := h.corePaymentService.CheckPaymentStatus(req.PaymentID)
	if err != nil {
		log.Printf("❌ CheckPaymentStatus error: %v", err)
		
		if err == repository.ErrPaymentNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "payment_not_found",
				Message: "Pembayaran tidak ditemukan",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "status_check_failed",
			Message: "Gagal memeriksa status pembayaran",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CoreWebhook handles Midtrans Core API webhook
// POST /api/webhook/midtrans/core
func (h *CorePaymentHandler) CoreWebhook(c *gin.Context) {
	log.Printf("🔔 Core Webhook received from IP: %s", c.ClientIP())
	log.Printf("🔔 Webhook Headers: %+v", c.Request.Header)

	var notification service.CoreWebhookNotification
	if err := c.ShouldBindJSON(&notification); err != nil {
		log.Printf("❌ Webhook parse error: %v", err)
		// Always return 200 to prevent Midtrans retries
		c.JSON(http.StatusOK, gin.H{
			"status":  "error",
			"message": "Invalid request format",
		})
		return
	}

	log.Printf("📦 Webhook payload: order_id=%s, status=%s, tx_id=%s, fraud=%s",
		notification.OrderID, notification.TransactionStatus, notification.TransactionID, notification.FraudStatus)
	log.Printf("📦 Full notification: %+v", notification)

	// Process webhook
	err := h.corePaymentService.ProcessCoreWebhook(notification)
	if err != nil {
		log.Printf("❌ Webhook processing error: %v", err)
		// Always return 200 to prevent Midtrans retries on non-recoverable errors
		c.JSON(http.StatusOK, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	log.Printf("✅ Webhook processed successfully for order: %s", notification.OrderID)
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// GetPendingOrders returns pending orders for Menunggu Pembayaran tab
// GET /api/pembelian/pending
func (h *CorePaymentHandler) GetPendingOrders(c *gin.Context) {
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	// Convert user_id from float64 (JWT) to int
	var userID int
	switch v := userIDRaw.(type) {
	case float64:
		userID = int(v)
	case int:
		userID = v
	default:
		log.Printf("❌ GetPendingOrders: invalid user_id type: %T", userIDRaw)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "invalid_user",
			Message: "Invalid user session",
		})
		return
	}

	log.Printf("📋 GetPendingOrders: user_id=%d", userID)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	response, err := h.corePaymentService.GetPendingOrders(userID, page, pageSize)
	if err != nil {
		log.Printf("❌ GetPendingOrders error: %v", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "fetch_failed",
			Message: "Gagal memuat pesanan",
		})
		return
	}

	log.Printf("✅ GetPendingOrders: found %d orders", len(response.Orders))
	c.JSON(http.StatusOK, response)
}

// GetTransactionHistory returns transaction history for Daftar Transaksi tab (Tokopedia-style)
// GET /api/pembelian/history?filter=all|ongoing|completed|failed
func (h *CorePaymentHandler) GetTransactionHistory(c *gin.Context) {
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	// Convert user_id from float64 (JWT) to int
	var userID int
	switch v := userIDRaw.(type) {
	case float64:
		userID = int(v)
	case int:
		userID = v
	default:
		log.Printf("❌ GetTransactionHistory: invalid user_id type: %T", userIDRaw)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "invalid_user",
			Message: "Invalid user session",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	filter := c.DefaultQuery("filter", "all") // all, ongoing, completed, failed

	// Validate filter
	validFilters := map[string]bool{"all": true, "ongoing": true, "completed": true, "failed": true}
	if !validFilters[filter] {
		filter = "all"
	}

	log.Printf("📋 GetTransactionHistory: user_id=%d, filter=%s", userID, filter)

	response, err := h.corePaymentService.GetTransactionHistoryWithFilter(userID, filter, page, pageSize)
	if err != nil {
		log.Printf("❌ GetTransactionHistory error: %v", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "fetch_failed",
			Message: "Gagal memuat riwayat transaksi",
		})
		return
	}

	log.Printf("✅ GetTransactionHistory: found %d orders", len(response.Orders))
	c.JSON(http.StatusOK, response)
}
