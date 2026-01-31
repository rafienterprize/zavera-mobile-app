package handler

import (
	"log"
	"net/http"
	"zavera/dto"
	"zavera/service"

	"github.com/gin-gonic/gin"
)

type CheckoutHandler struct {
	checkoutService service.CheckoutService
	shippingService service.ShippingService
}

func NewCheckoutHandler(checkoutService service.CheckoutService, shippingService service.ShippingService) *CheckoutHandler {
	return &CheckoutHandler{
		checkoutService: checkoutService,
		shippingService: shippingService,
	}
}

// getSessionID gets session ID from cookie (consistent with cart handler)
func (h *CheckoutHandler) getSessionID(c *gin.Context) string {
	sessionID, err := c.Cookie("session_id")
	if err != nil || sessionID == "" {
		// Fallback to header for API clients
		sessionID = c.GetHeader("X-Session-ID")
	}
	return sessionID
}

// CheckoutWithShipping handles checkout with shipping selection
// POST /api/checkout/shipping
func (h *CheckoutHandler) CheckoutWithShipping(c *gin.Context) {
	log.Printf("🛒 CheckoutWithShipping called")
	
	sessionID := h.getSessionID(c)
	log.Printf("📋 Session ID: %s", sessionID)
	
	if sessionID == "" {
		log.Printf("❌ No session ID")
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "session_required",
			Message: "Session ID is required",
		})
		return
	}

	var req dto.CheckoutWithShippingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ Bind error: %v", err)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	log.Printf("📦 Request: customer=%s, email=%s, courier=%s/%s, service=%s/%s", 
		req.CustomerName, req.CustomerEmail, req.CourierCode, req.ProviderCode, req.CourierServiceCode, req.ServiceCode)

	// Validate shipping address is provided
	if req.AddressID == nil && req.ShippingAddress == nil {
		log.Printf("❌ No shipping address")
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "address_required",
			Message: "Shipping address is required. Provide address_id or shipping_address.",
		})
		return
	}

	// Get user ID if authenticated
	var userID *int
	if id, exists := c.Get("user_id"); exists {
		switch v := id.(type) {
		case int:
			userID = &v
		case int64:
			uid := int(v)
			userID = &uid
		case float64:
			uid := int(v)
			userID = &uid
		case uint:
			uid := int(v)
			userID = &uid
		}
	}

	log.Printf("👤 User ID: %v", userID)

	response, err := h.checkoutService.CheckoutWithShipping(sessionID, req, userID)
	if err != nil {
		log.Printf("❌ Checkout error: %v", err)
		switch err {
		case service.ErrCartEmpty:
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "cart_empty",
				Message: "Your cart is empty",
			})
		case service.ErrAddressNotFound:
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "address_not_found",
				Message: "Shipping address not found",
			})
		case service.ErrInvalidAddress:
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "invalid_address",
				Message: "Invalid shipping address",
			})
		case service.ErrInvalidCourier:
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "invalid_courier",
				Message: "Selected courier service is not available",
			})
		default:
			// Check for insufficient stock
			if err.Error() != "" {
				c.JSON(http.StatusBadRequest, dto.ErrorResponse{
					Error:   "checkout_failed",
					Message: err.Error(),
				})
				return
			}
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:   "server_error",
				Message: "Failed to process checkout",
			})
		}
		return
	}

	log.Printf("✅ Checkout success: order_id=%d, order_code=%s", response.OrderID, response.OrderCode)
	c.JSON(http.StatusOK, response)
}

// GetShippingOptions returns available shipping options for cart
// GET /api/checkout/shipping-options?destination_postal_code=xxx&courier=jne
// NOTE: Uses postal_code for Biteship shipping calculation
func (h *CheckoutHandler) GetShippingOptions(c *gin.Context) {
	sessionID := h.getSessionID(c)
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "session_required",
			Message: "Session ID is required",
		})
		return
	}

	// Support postal_code (new Biteship) and district_id (legacy) parameters
	destinationPostalCode := c.Query("destination_postal_code")
	if destinationPostalCode == "" {
		// Fallback to old parameter names
		destinationPostalCode = c.Query("destination_district_id")
	}
	if destinationPostalCode == "" {
		destinationPostalCode = c.Query("destination_city_id")
	}
	
	if destinationPostalCode == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "destination_postal_code is required for shipping calculation",
		})
		return
	}

	courier := c.Query("courier")

	options, err := h.checkoutService.GetCartShippingOptions(sessionID, destinationPostalCode, courier)
	if err != nil {
		if err == service.ErrCartEmpty {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "cart_empty",
				Message: "Your cart is empty",
			})
			return
		}
		// Use fallback shipping rates when API is rate limited
		log.Printf("⚠️ Using fallback shipping rates due to API error: %v", err)
		fallbackRates := map[string]interface{}{
			"rates": []map[string]interface{}{
				{
					"provider_code":     "jne",
					"provider_name":     "JNE",
					"service_code":      "REG",
					"service_name":      "Reguler",
					"cost":              15000,
					"etd":               "2-3",
					"shipping_category": "Regular",
				},
				{
					"provider_code":     "jne",
					"provider_name":     "JNE",
					"service_code":      "YES",
					"service_name":      "Yakin Esok Sampai",
					"cost":              25000,
					"etd":               "1",
					"shipping_category": "Express",
				},
				{
					"provider_code":     "jnt",
					"provider_name":     "J&T Express",
					"service_code":      "EZ",
					"service_name":      "J&T EZ",
					"cost":              12000,
					"etd":               "3-5",
					"shipping_category": "Economy",
				},
				{
					"provider_code":     "sicepat",
					"provider_name":     "SiCepat",
					"service_code":      "REG",
					"service_name":      "Reguler",
					"cost":              14000,
					"etd":               "2-3",
					"shipping_category": "Regular",
				},
			},
			"grouped_rates": map[string][]map[string]interface{}{
				"Regular": {
					{"provider_code": "jne", "provider_name": "JNE", "service_code": "REG", "service_name": "Reguler", "cost": 15000, "etd": "2-3", "shipping_category": "Regular"},
					{"provider_code": "sicepat", "provider_name": "SiCepat", "service_code": "REG", "service_name": "Reguler", "cost": 14000, "etd": "2-3", "shipping_category": "Regular"},
				},
				"Express": {
					{"provider_code": "jne", "provider_name": "JNE", "service_code": "YES", "service_name": "Yakin Esok Sampai", "cost": 25000, "etd": "1", "shipping_category": "Express"},
				},
				"Economy": {
					{"provider_code": "jnt", "provider_name": "J&T Express", "service_code": "EZ", "service_name": "J&T EZ", "cost": 12000, "etd": "3-5", "shipping_category": "Economy"},
				},
			},
			"total_weight":    500,
			"total_weight_kg": "500 g",
			"origin_city":     "Semarang",
		}
		c.JSON(http.StatusOK, fallbackRates)
		return
	}

	c.JSON(http.StatusOK, options)
}
