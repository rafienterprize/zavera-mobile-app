package handler

import (
	"net/http"
	"zavera/dto"
	"zavera/service"

	"github.com/gin-gonic/gin"
)

type TrackingHandler struct {
	shippingService service.ShippingService
	orderService    service.OrderService
}

func NewTrackingHandler(shippingService service.ShippingService, orderService service.OrderService) *TrackingHandler {
	return &TrackingHandler{
		shippingService: shippingService,
		orderService:    orderService,
	}
}

// GetTrackingByResi gets tracking information by resi number
// GET /api/tracking/:resi
func (h *TrackingHandler) GetTrackingByResi(c *gin.Context) {
	resi := c.Param("resi")
	if resi == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_resi",
			Message: "Nomor resi tidak valid",
		})
		return
	}

	// Get shipment by resi (tracking_number)
	shipment, err := h.shippingService.GetShipmentByResi(resi)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "tracking_not_found",
			Message: "Nomor resi tidak ditemukan",
		})
		return
	}

	// Get order details
	order, err := h.orderService.GetOrderByID(shipment.OrderID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "order_not_found",
			Message: "Pesanan tidak ditemukan",
		})
		return
	}

	// Build tracking history
	var history []dto.TrackingHistoryResponse
	for _, event := range shipment.TrackingHistory {
		updatedAt := ""
		if event.EventTime != nil {
			updatedAt = event.EventTime.Format("2006-01-02T15:04:05Z07:00")
		}
		history = append(history, dto.TrackingHistoryResponse{
			Note:      event.Description,
			Status:    event.Status,
			UpdatedAt: updatedAt,
		})
	}

	// Build response
	response := dto.TrackingResponse{
		OrderCode:   order.OrderCode,
		Resi:        shipment.TrackingNumber,
		CourierName: shipment.ProviderName,
		Status:      string(shipment.Status),
		Origin:      shipment.OriginCityName,
		Destination: shipment.DestinationCityName,
		History:     history,
	}

	c.JSON(http.StatusOK, response)
}
