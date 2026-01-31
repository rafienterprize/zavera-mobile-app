package handler

import (
	"net/http"
	"strconv"
	"zavera/dto"
	"zavera/service"

	"github.com/gin-gonic/gin"
)

type FulfillmentHandler struct {
	fulfillmentSvc service.FulfillmentService
	disputeSvc     service.DisputeService
	monitorSvc     service.ShipmentMonitorService
}

func NewFulfillmentHandler(
	fulfillmentSvc service.FulfillmentService,
	disputeSvc service.DisputeService,
	monitorSvc service.ShipmentMonitorService,
) *FulfillmentHandler {
	return &FulfillmentHandler{
		fulfillmentSvc: fulfillmentSvc,
		disputeSvc:     disputeSvc,
		monitorSvc:     monitorSvc,
	}
}

// ============================================
// SHIPMENT CONTROL ENDPOINTS
// ============================================

// InvestigateShipment opens investigation for a shipment
// POST /api/admin/shipments/:id/investigate
func (h *FulfillmentHandler) InvestigateShipment(c *gin.Context) {
	shipmentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid shipment ID"})
		return
	}

	var req dto.InvestigateShipmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminEmail := c.GetString("user_email")
	if adminEmail == "" {
		adminEmail = "admin@system"
	}

	if err := h.fulfillmentSvc.OpenInvestigation(shipmentID, &req, adminEmail); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Investigation opened",
	})
}

// MarkLost marks a shipment as lost
// POST /api/admin/shipments/:id/mark-lost
func (h *FulfillmentHandler) MarkLost(c *gin.Context) {
	shipmentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid shipment ID"})
		return
	}

	var req dto.MarkLostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminEmail := c.GetString("user_email")
	if adminEmail == "" {
		adminEmail = "admin@system"
	}

	if err := h.fulfillmentSvc.MarkLost(shipmentID, &req, adminEmail); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Shipment marked as lost",
	})
}


// Reship creates a replacement shipment
// POST /api/admin/shipments/:id/reship
func (h *FulfillmentHandler) Reship(c *gin.Context) {
	shipmentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid shipment ID"})
		return
	}

	var req dto.ReshipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminEmail := c.GetString("user_email")
	if adminEmail == "" {
		adminEmail = "admin@system"
	}

	newShipment, err := h.fulfillmentSvc.CreateReship(shipmentID, &req, adminEmail)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"message":         "Replacement shipment created",
		"new_shipment_id": newShipment.ID,
	})
}

// OverrideStatus allows admin to override shipment status
// POST /api/admin/shipments/:id/override-status
func (h *FulfillmentHandler) OverrideStatus(c *gin.Context) {
	shipmentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid shipment ID"})
		return
	}

	var req dto.OverrideStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminEmail := c.GetString("user_email")
	if adminEmail == "" {
		adminEmail = "admin@system"
	}

	if err := h.fulfillmentSvc.OverrideStatus(shipmentID, &req, adminEmail); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Status overridden",
	})
}

// GetShipmentDetails returns enhanced shipment details
// GET /api/admin/shipments/:id/details
func (h *FulfillmentHandler) GetShipmentDetails(c *gin.Context) {
	shipmentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid shipment ID"})
		return
	}

	details, err := h.fulfillmentSvc.GetEnhancedShipment(shipmentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, details)
}

// GetStuckShipments returns shipments without tracking updates
// GET /api/admin/shipments/stuck
func (h *FulfillmentHandler) GetStuckShipments(c *gin.Context) {
	days := 7
	if d := c.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil {
			days = parsed
		}
	}

	stuck, err := h.fulfillmentSvc.GetStuckShipments(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stuck_shipments": stuck,
		"count":           len(stuck),
		"threshold_days":  days,
	})
}

// GetPickupFailures returns shipments with pickup failures
// GET /api/admin/shipments/pickup-failures
func (h *FulfillmentHandler) GetPickupFailures(c *gin.Context) {
	failures, err := h.fulfillmentSvc.GetPickupFailures()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pickup_failures": failures,
		"count":           len(failures),
	})
}

// GetFulfillmentDashboard returns fulfillment overview
// GET /api/admin/fulfillment/dashboard
func (h *FulfillmentHandler) GetFulfillmentDashboard(c *gin.Context) {
	dashboard, err := h.fulfillmentSvc.GetFulfillmentDashboard()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dashboard)
}

// GetShipmentsList returns paginated list of shipments
// GET /api/admin/shipments
func (h *FulfillmentHandler) GetShipmentsList(c *gin.Context) {
	status := c.Query("status")
	page := 1
	pageSize := 50
	
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	
	shipments, total, err := h.fulfillmentSvc.GetShipmentsList(status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"shipments": shipments,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}


// ============================================
// DISPUTE ENDPOINTS
// ============================================

// CreateDispute creates a new dispute
// POST /api/admin/disputes
func (h *FulfillmentHandler) CreateDispute(c *gin.Context) {
	var req dto.CreateDisputeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetInt("user_id")
	var customerUserID *int
	if userID > 0 {
		customerUserID = &userID
	}

	dispute, err := h.disputeSvc.CreateDispute(&req, customerUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, h.disputeSvc.ToDisputeResponse(dispute))
}

// GetDispute returns a dispute by ID
// GET /api/admin/disputes/:id
func (h *FulfillmentHandler) GetDispute(c *gin.Context) {
	disputeID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dispute ID"})
		return
	}

	dispute, err := h.disputeSvc.GetDispute(disputeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, h.disputeSvc.ToDisputeResponse(dispute))
}

// GetDisputeByCode returns a dispute by code
// GET /api/admin/disputes/code/:code
func (h *FulfillmentHandler) GetDisputeByCode(c *gin.Context) {
	code := c.Param("code")

	dispute, err := h.disputeSvc.GetDisputeByCode(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, h.disputeSvc.ToDisputeResponse(dispute))
}

// GetOpenDisputes returns all open disputes
// GET /api/admin/disputes/open
func (h *FulfillmentHandler) GetOpenDisputes(c *gin.Context) {
	disputes, err := h.disputeSvc.GetOpenDisputes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var responses []*dto.DisputeResponse
	for _, d := range disputes {
		responses = append(responses, h.disputeSvc.ToDisputeResponse(d))
	}

	c.JSON(http.StatusOK, gin.H{
		"disputes": responses,
		"count":    len(responses),
	})
}

// StartInvestigation starts investigation on a dispute
// POST /api/admin/disputes/:id/investigate
func (h *FulfillmentHandler) StartDisputeInvestigation(c *gin.Context) {
	disputeID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dispute ID"})
		return
	}

	adminID := c.GetInt("user_id")
	if adminID == 0 {
		adminID = 1 // Default admin
	}

	if err := h.disputeSvc.StartInvestigation(disputeID, adminID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Investigation started",
	})
}

// RequestEvidence requests evidence from customer
// POST /api/admin/disputes/:id/request-evidence
func (h *FulfillmentHandler) RequestEvidence(c *gin.Context) {
	disputeID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dispute ID"})
		return
	}

	var req struct {
		Message string `json:"message" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminID := c.GetInt("user_id")
	if adminID == 0 {
		adminID = 1
	}

	if err := h.disputeSvc.RequestEvidence(disputeID, req.Message, adminID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Evidence requested",
	})
}


// ResolveDispute resolves a dispute
// POST /api/admin/disputes/:id/resolve
func (h *FulfillmentHandler) ResolveDispute(c *gin.Context) {
	disputeID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dispute ID"})
		return
	}

	var req dto.ResolveDisputeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminID := c.GetInt("user_id")
	if adminID == 0 {
		adminID = 1
	}

	if err := h.disputeSvc.ResolveDispute(disputeID, &req, adminID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Dispute resolved",
	})
}

// CloseDispute closes a resolved dispute
// POST /api/admin/disputes/:id/close
func (h *FulfillmentHandler) CloseDispute(c *gin.Context) {
	disputeID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dispute ID"})
		return
	}

	adminID := c.GetInt("user_id")
	if adminID == 0 {
		adminID = 1
	}

	if err := h.disputeSvc.CloseDispute(disputeID, adminID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Dispute closed",
	})
}

// AddDisputeMessage adds a message to a dispute
// POST /api/admin/disputes/:id/messages
func (h *FulfillmentHandler) AddDisputeMessage(c *gin.Context) {
	disputeID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dispute ID"})
		return
	}

	var req dto.AddDisputeMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminID := c.GetInt("user_id")
	adminEmail := c.GetString("user_email")
	if adminEmail == "" {
		adminEmail = "Admin"
	}

	var senderID *int
	if adminID > 0 {
		senderID = &adminID
	}

	if err := h.disputeSvc.AddMessage(disputeID, &req, "admin", senderID, adminEmail); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Message added",
	})
}

// GetDisputeMessages returns messages for a dispute
// GET /api/admin/disputes/:id/messages
func (h *FulfillmentHandler) GetDisputeMessages(c *gin.Context) {
	disputeID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dispute ID"})
		return
	}

	includeInternal := c.Query("include_internal") == "true"

	messages, err := h.disputeSvc.GetMessages(disputeID, includeInternal)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	var responses []dto.DisputeMessageResponse
	for _, m := range messages {
		responses = append(responses, dto.DisputeMessageResponse{
			ID:             m.ID,
			SenderType:     m.SenderType,
			SenderName:     m.SenderName,
			Message:        m.Message,
			AttachmentURLs: m.AttachmentURLs,
			IsInternal:     m.IsInternal,
			CreatedAt:      m.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": responses,
		"count":    len(responses),
	})
}

// ============================================
// MONITORING ENDPOINTS
// ============================================

// RunMonitoringJob manually triggers monitoring detectors
// POST /api/admin/fulfillment/run-monitors
func (h *FulfillmentHandler) RunMonitoringJob(c *gin.Context) {
	if err := h.monitorSvc.RunAllDetectors(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Monitoring detectors executed",
	})
}

// SchedulePickup schedules courier pickup
// POST /api/admin/shipments/:id/schedule-pickup
func (h *FulfillmentHandler) SchedulePickup(c *gin.Context) {
	shipmentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid shipment ID"})
		return
	}

	var req dto.SchedulePickupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminEmail := c.GetString("user_email")
	if adminEmail == "" {
		adminEmail = "admin@system"
	}

	if err := h.fulfillmentSvc.SchedulePickup(shipmentID, &req, adminEmail); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Pickup scheduled",
	})
}

// MarkShipped marks a shipment as shipped
// POST /api/admin/shipments/:id/mark-shipped
func (h *FulfillmentHandler) MarkShipped(c *gin.Context) {
	shipmentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid shipment ID"})
		return
	}

	var req dto.MarkShippedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminEmail := c.GetString("user_email")
	if adminEmail == "" {
		adminEmail = "admin@system"
	}

	if err := h.fulfillmentSvc.MarkShipped(shipmentID, &req, adminEmail); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Shipment marked as shipped",
	})
}
