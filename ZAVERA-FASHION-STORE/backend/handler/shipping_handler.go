package handler

import (
	"log"
	"net/http"
	"strconv"
	"zavera/dto"
	"zavera/service"

	"github.com/gin-gonic/gin"
)

type ShippingHandler struct {
	shippingService service.ShippingService
}

func NewShippingHandler(shippingService service.ShippingService) *ShippingHandler {
	return &ShippingHandler{
		shippingService: shippingService,
	}
}

// getUserIDAsInt converts user_id from context to int
func (h *ShippingHandler) getUserIDAsInt(userID interface{}) int {
	switch v := userID.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case uint:
		return int(v)
	default:
		return 0
	}
}

// getSessionID gets session ID from cookie (consistent with cart handler)
func (h *ShippingHandler) getSessionID(c *gin.Context) string {
	sessionID, err := c.Cookie("session_id")
	if err != nil || sessionID == "" {
		// Fallback to header for API clients
		sessionID = c.GetHeader("X-Session-ID")
	}
	return sessionID
}

// ============================================
// BITESHIP NATIVE SHIPPING RATES
// ============================================

// GetBiteshipRates returns shipping rates using Biteship postal_code
// POST /api/shipping/rates
func (h *ShippingHandler) GetShippingRates(c *gin.Context) {
	sessionID := h.getSessionID(c)
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "session_required",
			Message: "Session ID is required",
		})
		return
	}

	// Get user ID if logged in
	var userID *int
	if id, exists := c.Get("user_id"); exists {
		uid := h.getUserIDAsInt(id)
		if uid > 0 {
			userID = &uid
		}
	}

	var req dto.GetRatesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "destination_area_id or destination_postal_code is required",
		})
		return
	}

	if req.DestinationAreaID == "" && req.DestinationPostalCode == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "destination_area_id or destination_postal_code is required",
		})
		return
	}

	log.Printf("📦 GetBiteshipRates: session=%s, user_id=%v, destination_area_id=%s, destination_postal_code=%s", 
		sessionID, userID, req.DestinationAreaID, req.DestinationPostalCode)

	rates, err := h.shippingService.GetBiteshipRatesForUser(sessionID, userID, req.DestinationAreaID, req.DestinationPostalCode)
	if err != nil {
		log.Printf("❌ GetBiteshipRates error: %v", err)
		if err == service.ErrCartEmpty {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "cart_empty",
				Message: "Cart is empty",
			})
			return
		}
		c.JSON(http.StatusServiceUnavailable, dto.ErrorResponse{
			Error:   "shipping_unavailable",
			Message: "Unable to fetch shipping rates. Please try again.",
		})
		return
	}

	c.JSON(http.StatusOK, rates)
}

// GetCartShippingPreview returns shipping rates for current cart
// GET /api/shipping/preview?destination_district_id=xxx&courier=jne
// NOTE: Uses district_id (kecamatan) for shipping calculation. For Biteship, use area_id instead.
func (h *ShippingHandler) GetCartShippingPreview(c *gin.Context) {
	sessionID := h.getSessionID(c)
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "session_required",
			Message: "Session ID is required",
		})
		return
	}

	// Support both old (city_id) and new (district_id) parameters for backward compatibility
	destinationDistrictID := c.Query("destination_district_id")
	if destinationDistrictID == "" {
		// Fallback to old parameter name
		destinationDistrictID = c.Query("destination_city_id")
	}
	
	if destinationDistrictID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "destination_district_id is required for accurate shipping calculation",
		})
		return
	}

	courier := c.Query("courier")

	preview, err := h.shippingService.GetCartShippingPreview(sessionID, destinationDistrictID, courier)
	if err != nil {
		if err == service.ErrCartEmpty {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "cart_empty",
				Message: "Cart is empty",
			})
			return
		}
		c.JSON(http.StatusServiceUnavailable, dto.ErrorResponse{
			Error:   "shipping_unavailable",
			Message: "Unable to fetch shipping rates. Please try again.",
		})
		return
	}

	c.JSON(http.StatusOK, preview)
}

// ============================================
// PROVIDERS
// ============================================

// GetProviders returns list of shipping providers
// GET /api/shipping/providers
func (h *ShippingHandler) GetProviders(c *gin.Context) {
	providers, err := h.shippingService.GetProviders()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "server_error",
			Message: "Failed to fetch providers",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"providers": providers})
}

// ============================================
// LOCATION
// ============================================

// SearchAreas searches for areas using Biteship API
// GET /api/shipping/areas?q=xxx
func (h *ShippingHandler) SearchAreas(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Search query (q) is required",
		})
		return
	}

	areas, err := h.shippingService.SearchAreas(query)
	if err != nil {
		log.Printf("⚠️ Biteship area search failed: %v", err)
		c.JSON(http.StatusServiceUnavailable, dto.ErrorResponse{
			Error:   "service_unavailable",
			Message: "Unable to search areas. Please try again.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"areas": areas})
}

// Fallback provinces data when API is unavailable (rate limited)
var fallbackProvinces = []dto.ProvinceResponse{
	{ProvinceID: "1", ProvinceName: "Bali"},
	{ProvinceID: "2", ProvinceName: "Bangka Belitung"},
	{ProvinceID: "3", ProvinceName: "Banten"},
	{ProvinceID: "4", ProvinceName: "Bengkulu"},
	{ProvinceID: "5", ProvinceName: "DI Yogyakarta"},
	{ProvinceID: "6", ProvinceName: "DKI Jakarta"},
	{ProvinceID: "7", ProvinceName: "Gorontalo"},
	{ProvinceID: "8", ProvinceName: "Jambi"},
	{ProvinceID: "9", ProvinceName: "Jawa Barat"},
	{ProvinceID: "10", ProvinceName: "Jawa Tengah"},
	{ProvinceID: "11", ProvinceName: "Jawa Timur"},
	{ProvinceID: "12", ProvinceName: "Kalimantan Barat"},
	{ProvinceID: "13", ProvinceName: "Kalimantan Selatan"},
	{ProvinceID: "14", ProvinceName: "Kalimantan Tengah"},
	{ProvinceID: "15", ProvinceName: "Kalimantan Timur"},
	{ProvinceID: "16", ProvinceName: "Kalimantan Utara"},
	{ProvinceID: "17", ProvinceName: "Kepulauan Riau"},
	{ProvinceID: "18", ProvinceName: "Lampung"},
	{ProvinceID: "19", ProvinceName: "Maluku"},
	{ProvinceID: "20", ProvinceName: "Maluku Utara"},
	{ProvinceID: "21", ProvinceName: "Nanggroe Aceh Darussalam"},
	{ProvinceID: "22", ProvinceName: "Nusa Tenggara Barat"},
	{ProvinceID: "23", ProvinceName: "Nusa Tenggara Timur"},
	{ProvinceID: "24", ProvinceName: "Papua"},
	{ProvinceID: "25", ProvinceName: "Papua Barat"},
	{ProvinceID: "26", ProvinceName: "Riau"},
	{ProvinceID: "27", ProvinceName: "Sulawesi Barat"},
	{ProvinceID: "28", ProvinceName: "Sulawesi Selatan"},
	{ProvinceID: "29", ProvinceName: "Sulawesi Tengah"},
	{ProvinceID: "30", ProvinceName: "Sulawesi Tenggara"},
	{ProvinceID: "31", ProvinceName: "Sulawesi Utara"},
	{ProvinceID: "32", ProvinceName: "Sumatera Barat"},
	{ProvinceID: "33", ProvinceName: "Sumatera Selatan"},
	{ProvinceID: "34", ProvinceName: "Sumatera Utara"},
}

// Fallback cities for Jawa Tengah (province_id=10) for testing
var fallbackCitiesJawaTengah = []dto.CityResponse{
	{CityID: "398", Type: "Kabupaten", CityName: "Semarang", ProvinceID: "10"},
	{CityID: "399", Type: "Kota", CityName: "Semarang", ProvinceID: "10"},
	{CityID: "163", Type: "Kabupaten", CityName: "Demak", ProvinceID: "10"},
	{CityID: "209", Type: "Kabupaten", CityName: "Kendal", ProvinceID: "10"},
	{CityID: "427", Type: "Kota", CityName: "Surakarta (Solo)", ProvinceID: "10"},
	{CityID: "196", Type: "Kabupaten", CityName: "Karanganyar", ProvinceID: "10"},
	{CityID: "445", Type: "Kabupaten", CityName: "Tegal", ProvinceID: "10"},
	{CityID: "446", Type: "Kota", CityName: "Tegal", ProvinceID: "10"},
	{CityID: "349", Type: "Kabupaten", CityName: "Pekalongan", ProvinceID: "10"},
	{CityID: "350", Type: "Kota", CityName: "Pekalongan", ProvinceID: "10"},
}

// GetProvinces returns list of provinces
// GET /api/shipping/provinces
func (h *ShippingHandler) GetProvinces(c *gin.Context) {
	provinces, err := h.shippingService.GetProvinces()
	if err != nil {
		// Use fallback data when API is rate limited
		log.Printf("⚠️ Using fallback provinces data due to API error: %v", err)
		c.JSON(http.StatusOK, gin.H{"provinces": fallbackProvinces})
		return
	}

	c.JSON(http.StatusOK, gin.H{"provinces": provinces})
}

// GetCities returns list of cities
// GET /api/shipping/cities?province_id=xxx
func (h *ShippingHandler) GetCities(c *gin.Context) {
	provinceID := c.Query("province_id")

	cities, err := h.shippingService.GetCities(provinceID)
	if err != nil {
		// Use fallback data for Jawa Tengah when API is rate limited
		if provinceID == "10" {
			log.Printf("⚠️ Using fallback cities data for Jawa Tengah due to API error: %v", err)
			c.JSON(http.StatusOK, gin.H{"cities": fallbackCitiesJawaTengah})
			return
		}
		c.JSON(http.StatusServiceUnavailable, dto.ErrorResponse{
			Error:   "service_unavailable",
			Message: "Unable to fetch cities. Please try again.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"cities": cities})
}

// GetSubdistricts returns list of subdistricts (kecamatan) for a city
// GET /api/shipping/subdistricts?city_id=xxx
// NOTE: This reads from local database. For Biteship, use /api/shipping/areas instead.
func (h *ShippingHandler) GetSubdistricts(c *gin.Context) {
	cityID := c.Query("city_id")
	if cityID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "city_id is required",
		})
		return
	}

	query := c.Query("q") // Optional search query

	var subdistricts []dto.SubdistrictResponse
	var err error

	if query != "" {
		subdistricts, err = h.shippingService.SearchSubdistricts(cityID, query)
	} else {
		subdistricts, err = h.shippingService.GetSubdistricts(cityID)
	}

	if err != nil {
		c.JSON(http.StatusServiceUnavailable, dto.ErrorResponse{
			Error:   "service_unavailable",
			Message: "Unable to fetch subdistricts. Please try again.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"subdistricts": subdistricts})
}

// GetDistricts returns list of districts (kecamatan) from Kommerce API (legacy)
// GET /api/shipping/districts?city_id=xxx
// NOTE: For Biteship integration, use /api/shipping/areas instead
func (h *ShippingHandler) GetDistricts(c *gin.Context) {
	cityID := c.Query("city_id")
	if cityID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "city_id is required",
		})
		return
	}

	districts, err := h.shippingService.GetDistrictsFromAPI(cityID)
	if err != nil {
		// Fallback districts for Kota Semarang (city_id=399) for testing
		if cityID == "399" {
			log.Printf("⚠️ Using fallback districts data for Kota Semarang due to API error: %v", err)
			fallbackDistricts := []dto.DistrictResponse{
				{DistrictID: "5765", DistrictName: "Banyumanik"},
				{DistrictID: "5766", DistrictName: "Candisari"},
				{DistrictID: "5767", DistrictName: "Gajahmungkur"},
				{DistrictID: "5768", DistrictName: "Gayamsari"},
				{DistrictID: "5769", DistrictName: "Genuk"},
				{DistrictID: "5770", DistrictName: "Gunungpati"},
				{DistrictID: "5771", DistrictName: "Mijen"},
				{DistrictID: "5772", DistrictName: "Ngaliyan"},
				{DistrictID: "5773", DistrictName: "Pedurungan"},
				{DistrictID: "5774", DistrictName: "Semarang Barat"},
				{DistrictID: "5775", DistrictName: "Semarang Selatan"},
				{DistrictID: "5776", DistrictName: "Semarang Tengah"},
				{DistrictID: "5777", DistrictName: "Semarang Timur"},
				{DistrictID: "5778", DistrictName: "Semarang Utara"},
				{DistrictID: "5779", DistrictName: "Tembalang"},
				{DistrictID: "5780", DistrictName: "Tugu"},
			}
			c.JSON(http.StatusOK, gin.H{"districts": fallbackDistricts})
			return
		}
		c.JSON(http.StatusServiceUnavailable, dto.ErrorResponse{
			Error:   "service_unavailable",
			Message: "Unable to fetch districts. Please try again.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"districts": districts})
}

// GetSubdistrictsAPI returns list of subdistricts (kelurahan) from Kommerce API (legacy)
// GET /api/shipping/kelurahan?district_id=xxx
// NOTE: For Biteship integration, use /api/shipping/areas instead
func (h *ShippingHandler) GetSubdistrictsAPI(c *gin.Context) {
	districtID := c.Query("district_id")
	if districtID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "district_id is required",
		})
		return
	}

	subdistricts, err := h.shippingService.GetSubdistrictsFromAPI(districtID)
	if err != nil {
		// Fallback kelurahan for Pedurungan (district_id=5773) for testing
		if districtID == "5773" {
			log.Printf("⚠️ Using fallback kelurahan data for Pedurungan due to API error: %v", err)
			fallbackKelurahan := []dto.SubdistrictAPIResponse{
				{SubdistrictID: "80401", SubdistrictName: "Gemah", PostalCode: "50191"},
				{SubdistrictID: "80402", SubdistrictName: "Kalicari", PostalCode: "50198"},
				{SubdistrictID: "80403", SubdistrictName: "Muktiharjo Kidul", PostalCode: "50192"},
				{SubdistrictID: "80404", SubdistrictName: "Palebon", PostalCode: "50199"},
				{SubdistrictID: "80405", SubdistrictName: "Pedurungan Kidul", PostalCode: "50193"},
				{SubdistrictID: "80406", SubdistrictName: "Pedurungan Lor", PostalCode: "50194"},
				{SubdistrictID: "80407", SubdistrictName: "Pedurungan Tengah", PostalCode: "50195"},
				{SubdistrictID: "80408", SubdistrictName: "Penggaron Kidul", PostalCode: "50196"},
				{SubdistrictID: "80409", SubdistrictName: "Plamongan Sari", PostalCode: "50197"},
				{SubdistrictID: "80410", SubdistrictName: "Tlogomulyo", PostalCode: "50190"},
				{SubdistrictID: "80411", SubdistrictName: "Tlogosari Kulon", PostalCode: "50196"},
				{SubdistrictID: "80412", SubdistrictName: "Tlogosari Wetan", PostalCode: "50196"},
			}
			c.JSON(http.StatusOK, gin.H{"subdistricts": fallbackKelurahan})
			return
		}
		c.JSON(http.StatusServiceUnavailable, dto.ErrorResponse{
			Error:   "service_unavailable",
			Message: "Unable to fetch kelurahan. Please try again.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"subdistricts": subdistricts})
}

// ============================================
// ADDRESSES (Protected routes)
// ============================================

// CreateAddress creates a new address for user
// POST /api/user/addresses
func (h *ShippingHandler) CreateAddress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	var req dto.CreateAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Convert userID to int (JWT stores as float64)
	uid := h.getUserIDAsInt(userID)
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid user ID",
		})
		return
	}

	address, err := h.shippingService.CreateAddress(uid, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "server_error",
			Message: "Failed to create address",
		})
		return
	}

	c.JSON(http.StatusCreated, address)
}

// GetUserAddresses returns user's saved addresses
// GET /api/user/addresses
func (h *ShippingHandler) GetUserAddresses(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	uid := h.getUserIDAsInt(userID)
	addresses, err := h.shippingService.GetUserAddresses(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "server_error",
			Message: "Failed to fetch addresses",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"addresses": addresses})
}

// GetAddress returns a specific address
// GET /api/user/addresses/:id
func (h *ShippingHandler) GetAddress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid address ID",
		})
		return
	}

	address, err := h.shippingService.GetAddressByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Message: "Address not found",
		})
		return
	}

	// Verify ownership - convert userID to int for comparison
	uid := h.getUserIDAsInt(userID)
	if !h.shippingService.VerifyAddressOwnership(id, uid) {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "forbidden",
			Message: "You don't have permission to access this address",
		})
		return
	}

	c.JSON(http.StatusOK, address)
}

// UpdateAddress updates an address
// PUT /api/user/addresses/:id
func (h *ShippingHandler) UpdateAddress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid address ID",
		})
		return
	}

	// Verify ownership before update
	uid := h.getUserIDAsInt(userID)
	if !h.shippingService.VerifyAddressOwnership(id, uid) {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "forbidden",
			Message: "You don't have permission to update this address",
		})
		return
	}

	var req dto.UpdateAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	address, err := h.shippingService.UpdateAddress(id, req)
	if err != nil {
		if err == service.ErrAddressNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "not_found",
				Message: "Address not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "server_error",
			Message: "Failed to update address",
		})
		return
	}

	c.JSON(http.StatusOK, address)
}

// DeleteAddress deletes an address
// DELETE /api/user/addresses/:id
func (h *ShippingHandler) DeleteAddress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid address ID",
		})
		return
	}

	// Verify ownership before delete
	uid := h.getUserIDAsInt(userID)
	if !h.shippingService.VerifyAddressOwnership(id, uid) {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "forbidden",
			Message: "You don't have permission to delete this address",
		})
		return
	}

	err = h.shippingService.DeleteAddress(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "server_error",
			Message: "Failed to delete address",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Address deleted"})
}

// SetDefaultAddress sets an address as default
// POST /api/user/addresses/:id/default
func (h *ShippingHandler) SetDefaultAddress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	idStr := c.Param("id")
	addressID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid address ID",
		})
		return
	}

	uid := h.getUserIDAsInt(userID)
	err = h.shippingService.SetDefaultAddress(uid, addressID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "server_error",
			Message: "Failed to set default address",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Default address updated"})
}

// ============================================
// SHIPMENTS & TRACKING
// ============================================

// GetShipmentByOrder returns shipment for an order
// GET /api/orders/:code/shipment
func (h *ShippingHandler) GetShipmentByOrder(c *gin.Context) {
	orderCode := c.Param("code")
	if orderCode == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Order code is required",
		})
		return
	}

	// This would need order lookup first - simplified for now
	// In real implementation, get order by code then get shipment
	c.JSON(http.StatusNotImplemented, dto.ErrorResponse{
		Error:   "not_implemented",
		Message: "Use /api/shipments/:id endpoint",
	})
}

// GetShipment returns shipment details with tracking
// GET /api/shipments/:id
func (h *ShippingHandler) GetShipment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid shipment ID",
		})
		return
	}

	shipment, err := h.shippingService.GetTracking(id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Message: "Shipment not found",
		})
		return
	}

	c.JSON(http.StatusOK, shipment)
}

// RefreshTracking refreshes tracking info from courier API
// POST /api/shipments/:id/refresh
func (h *ShippingHandler) RefreshTracking(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid shipment ID",
		})
		return
	}

	err = h.shippingService.RefreshTracking(id)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, dto.ErrorResponse{
			Error:   "tracking_unavailable",
			Message: "Unable to refresh tracking. Please try again.",
		})
		return
	}

	// Return updated shipment
	shipment, _ := h.shippingService.GetTracking(id)
	c.JSON(http.StatusOK, shipment)
}

// ============================================
// ADMIN ENDPOINTS
// ============================================

// AdminUpdateTracking updates tracking number (admin only)
// PUT /api/admin/shipments/:id/tracking
func (h *ShippingHandler) AdminUpdateTracking(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid shipment ID",
		})
		return
	}

	var req dto.UpdateTrackingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	err = h.shippingService.UpdateTrackingNumber(id, req.TrackingNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "server_error",
			Message: "Failed to update tracking number",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tracking number updated"})
}

// AdminMarkShipped marks shipment as shipped (admin only)
// POST /api/admin/shipments/:id/ship
func (h *ShippingHandler) AdminMarkShipped(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid shipment ID",
		})
		return
	}

	var req dto.MarkShippedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	err = h.shippingService.MarkAsShipped(id, req.TrackingNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "server_error",
			Message: "Failed to mark as shipped",
		})
		return
	}

	// Return updated shipment
	shipment, _ := h.shippingService.GetTracking(id)
	c.JSON(http.StatusOK, shipment)
}

// AdminRunTrackingJob manually triggers tracking job (admin only)
// POST /api/admin/shipping/tracking-job
func (h *ShippingHandler) AdminRunTrackingJob(c *gin.Context) {
	err := h.shippingService.RunTrackingJob()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "job_failed",
			Message: "Tracking job failed: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tracking job completed"})
}
