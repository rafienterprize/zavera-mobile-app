package handler

import (
	"log"
	"net/http"
	"zavera/dto"
	"zavera/service"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentService service.PaymentService
}

func NewPaymentHandler(paymentService service.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

// InitiatePayment godoc
// @Summary Initiate payment for an existing order
// @Description Creates a Midtrans Snap token for the given order
// @Tags payments
// @Accept json
// @Produce json
// @Param request body dto.InitiatePaymentRequest true "Initiate payment request"
// @Success 200 {object} dto.InitiatePaymentResponse
// @Router /api/payments/initiate [post]
func (h *PaymentHandler) InitiatePayment(c *gin.Context) {
	var req dto.InitiatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ InitiatePayment bind error: %v", err)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	log.Printf("💳 InitiatePayment request: order_id=%d", req.OrderID)

	// Create snap token via payment service
	snapToken, err := h.paymentService.InitiatePayment(req.OrderID)
	if err != nil {
		status := http.StatusInternalServerError

		switch err.Error() {
		case "order not found":
			status = http.StatusNotFound
		case "order is not in pending status":
			status = http.StatusConflict
		case "payment already exists for this order":
			status = http.StatusConflict
		}

		c.JSON(status, dto.ErrorResponse{
			Error:   "payment_initiation_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.InitiatePaymentResponse{
		SnapToken: snapToken,
	})
}

// Webhook godoc
// @Summary Midtrans payment webhook
// @Description Handle payment notification from Midtrans (webhook/notification URL)
// @Tags payments
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/payments/webhook [post]
func (h *PaymentHandler) Webhook(c *gin.Context) {
	// Log raw request for debugging
	log.Printf("🔔 Webhook received from IP: %s", c.ClientIP())
	
	var notification dto.MidtransNotification
	if err := c.ShouldBindJSON(&notification); err != nil {
		log.Printf("❌ Webhook parse error: %v", err)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Log webhook payload for debugging
	log.Printf("📦 Webhook payload: order_id=%s, transaction_status=%s, transaction_id=%s, payment_type=%s",
		notification.OrderID, notification.TransactionStatus, notification.TransactionID, notification.PaymentType)

	// Process webhook via payment service
	// This handles signature verification, idempotency, and order/stock updates
	err := h.paymentService.ProcessWebhook(notification)
	if err != nil {
		log.Printf("❌ Webhook processing error: %v", err)
		// Log the error but still return 200 to Midtrans to avoid retries
		// on non-recoverable errors (like invalid signature)
		// Midtrans recommends always returning 200
		c.JSON(http.StatusOK, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	log.Printf("✅ Webhook processed successfully for order: %s", notification.OrderID)
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Webhook processed successfully",
	})
}
