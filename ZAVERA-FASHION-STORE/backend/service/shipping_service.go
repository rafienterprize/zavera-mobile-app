package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	"zavera/dto"
	"zavera/models"
	"zavera/repository"
)

var (
	ErrShippingNotAvailable = errors.New("shipping not available for this route")
	ErrInvalidAddress       = errors.New("invalid shipping address")
	ErrAddressNotFound      = errors.New("address not found")
	ErrShipmentNotFound     = errors.New("shipment not found")
	ErrShippingNotLocked    = errors.New("shipping must be selected before payment")
	ErrInvalidCourier       = errors.New("invalid courier or service")
)

// Store origin location (Semarang Tengah as ZAVERA warehouse location)
// Biteship uses area_id for accurate shipping calculation
const DefaultOriginCityID = "399"      // Kota Semarang (legacy, for display)
const DefaultOriginDistrictID = "5598" // Pedurungan district ID (legacy)
const DefaultOriginAreaID = "IDNP10IDNC393IDND4700" // Pedurungan, Semarang area ID for Biteship

type ShippingService interface {
	// Biteship Native Rates
	GetBiteshipRates(sessionID string, destinationAreaID string, destinationPostalCode string) (*dto.GetRatesResponse, error)
	GetBiteshipRatesForUser(sessionID string, userID *int, destinationAreaID string, destinationPostalCode string) (*dto.GetRatesResponse, error)
	
	// Legacy Rates (deprecated)
	GetShippingRates(req dto.GetShippingRatesRequest) (*dto.ShippingRatesResponse, error)
	GetCartShippingPreview(sessionID string, destinationDistrictID string, courier string) (*dto.CartShippingPreviewResponse, error)

	// Providers
	GetProviders() ([]models.ShippingProvider, error)

	// Location - Biteship area search (PRIMARY)
	SearchAreas(query string) ([]dto.AreaResponse, error)
	
	// Legacy location methods (deprecated - DO NOT USE)
	GetProvinces() ([]dto.ProvinceResponse, error)
	GetCities(provinceID string) ([]dto.CityResponse, error)
	GetDistrictsFromAPI(cityID string) ([]dto.DistrictResponse, error)
	GetSubdistrictsFromAPI(districtID string) ([]dto.SubdistrictAPIResponse, error)
	GetSubdistricts(cityID string) ([]dto.SubdistrictResponse, error)
	SearchSubdistricts(cityID string, query string) ([]dto.SubdistrictResponse, error)

	// Addresses
	CreateAddress(userID int, req dto.CreateAddressRequest) (*dto.AddressResponse, error)
	GetUserAddresses(userID int) ([]dto.AddressResponse, error)
	GetAddressByID(id int) (*dto.AddressResponse, error)
	UpdateAddress(id int, req dto.UpdateAddressRequest) (*dto.AddressResponse, error)
	DeleteAddress(id int) error
	SetDefaultAddress(userID, addressID int) error
	VerifyAddressOwnership(addressID, userID int) bool

	// Shipments
	CreateShipmentForOrder(orderID int, providerCode, serviceName string, cost float64, etd string, weight int, destCityID, destCityName string) (*models.Shipment, error)
	GetShipmentByOrderID(orderID int) (*dto.ShipmentResponse, error)
	GetShipmentByOrderCode(orderCode string) (*models.Shipment, error)
	GetShipmentByResi(resi string) (*models.Shipment, error)
	UpdateTrackingNumber(shipmentID int, trackingNumber string) error
	MarkAsShipped(shipmentID int, trackingNumber string) error
	
	// Biteship Draft Orders
	CreateDraftOrderForCheckout(orderID int, req CreateDraftOrderParams) (*dto.DraftOrderResponse, error)
	ConfirmDraftOrder(orderID int) (*dto.OrderConfirmationResponse, error)

	// Tracking
	GetTracking(shipmentID int) (*dto.ShipmentResponse, error)
	RefreshTracking(shipmentID int) error
	RunTrackingJob() error
}

// CreateDraftOrderParams contains parameters for creating a Biteship draft order
type CreateDraftOrderParams struct {
	OriginAreaID          string
	OriginAddress         string
	OriginPostalCode      string
	OriginContactName     string
	OriginContactPhone    string
	DestinationAreaID     string
	DestinationAddress    string
	DestinationPostalCode string
	DestinationContactName  string
	DestinationContactPhone string
	CourierCode           string
	CourierServiceCode    string
	Items                 []CreateDraftOrderItem
}

type shippingService struct {
	shippingRepo repository.ShippingRepository
	cartRepo     repository.CartRepository
	productRepo  repository.ProductRepository
	orderRepo    repository.OrderRepository
	biteship     *BiteshipClient
}

func NewShippingService(
	shippingRepo repository.ShippingRepository,
	cartRepo repository.CartRepository,
	productRepo repository.ProductRepository,
	orderRepo repository.OrderRepository,
) ShippingService {
	return &shippingService{
		shippingRepo: shippingRepo,
		cartRepo:     cartRepo,
		productRepo:  productRepo,
		orderRepo:    orderRepo,
		biteship:     NewBiteshipClient(),
	}
}

// ============================================
// BITESHIP AREA SEARCH
// ============================================

// SearchAreas searches for areas using Biteship API with autocomplete suggestions
func (s *shippingService) SearchAreas(query string) ([]dto.AreaResponse, error) {
	// Use enhanced search with suggestions for better autocomplete
	areas, err := s.biteship.SearchAreasWithSuggestions(query)
	if err != nil {
		// Fallback to basic search
		areas, err = s.biteship.SearchAreas(query)
		if err != nil {
			return nil, err
		}
	}

	var result []dto.AreaResponse
	for _, area := range areas {
		// Convert postal_code to string (can be int or string from API)
		postalCode := ""
		if area.PostalCode != nil {
			postalCode = fmt.Sprintf("%v", area.PostalCode)
		}
		
		// If postal_code is empty, try to extract from name
		// Format: "Menteng, Jakarta Pusat, DKI Jakarta. 10310"
		if postalCode == "" && area.Name != "" {
			// Find the last part after ". " which contains postal code
			parts := strings.Split(area.Name, ". ")
			if len(parts) > 1 {
				lastPart := parts[len(parts)-1]
				// Check if it's a valid postal code (5 digits)
				if len(lastPart) == 5 {
					postalCode = lastPart
				}
			}
		}
		
		result = append(result, dto.AreaResponse{
			AreaID:      area.ID,
			Name:        area.Name,
			PostalCode:  postalCode,
			Province:    area.AdministrativeLevel1Name,
			City:        area.AdministrativeLevel2Name,
			District:    area.AdministrativeLevel3Name,
			Subdistrict: area.AdministrativeLevel4Name,
		})
	}

	return result, nil
}

// ============================================
// BITESHIP NATIVE RATES
// ============================================

// Default origin postal code (Pedurungan, Semarang)
const DefaultOriginPostalCode = 50113

// GetBiteshipRatesForUser gets shipping rates for a specific user (PRIMARY METHOD)
// This method properly handles logged-in users by finding their cart
func (s *shippingService) GetBiteshipRatesForUser(sessionID string, userID *int, destinationAreaID string, destinationPostalCode string) (*dto.GetRatesResponse, error) {
	log.Printf("🔍 GetBiteshipRatesForUser called - sessionID: %s, userID: %v", sessionID, userID)
	
	var cart *models.Cart
	var items []models.CartItem
	var err error
	
	// If user is logged in, find their cart first
	if userID != nil && *userID > 0 {
		log.Printf("🔍 Looking for cart by user_id: %d", *userID)
		cart, err = s.cartRepo.FindByUserID(*userID)
		if err == nil && cart != nil {
			items, _ = s.cartRepo.FindItemsByCartID(cart.ID)
			log.Printf("📦 Found user's cart ID: %d with %d items", cart.ID, len(items))
		}
	}
	
	// If no cart found by user_id, try by session
	if cart == nil || len(items) == 0 {
		log.Printf("🔍 Looking for cart by session: %s", sessionID)
		cart, err = s.cartRepo.FindOrCreateBySessionID(sessionID)
		if err != nil {
			log.Printf("❌ Cart error: %v", err)
			return nil, ErrCartEmpty
		}
		if cart != nil {
			items, _ = s.cartRepo.FindItemsByCartID(cart.ID)
			log.Printf("📦 Found session cart ID: %d with %d items", cart.ID, len(items))
		}
	}
	
	if cart == nil || len(items) == 0 {
		log.Printf("❌ No cart with items found")
		return nil, ErrCartEmpty
	}
	
	// Continue with rate calculation using the found cart and items
	return s.calculateRatesForItems(items, destinationAreaID, destinationPostalCode)
}

// GetBiteshipRates gets shipping rates using Biteship postal_code (LEGACY - use GetBiteshipRatesForUser)
func (s *shippingService) GetBiteshipRates(sessionID string, destinationAreaID string, destinationPostalCode string) (*dto.GetRatesResponse, error) {
	return s.GetBiteshipRatesForUser(sessionID, nil, destinationAreaID, destinationPostalCode)
}

// calculateRatesForItems calculates shipping rates for given cart items
func (s *shippingService) calculateRatesForItems(items []models.CartItem, destinationAreaID string, destinationPostalCode string) (*dto.GetRatesResponse, error) {

	// Parse destination postal code
	destPostalCode := 0
	if destinationPostalCode != "" {
		fmt.Sscanf(destinationPostalCode, "%d", &destPostalCode)
	}
	
	// If postal code not provided, try to get from area search
	if destPostalCode == 0 && destinationAreaID != "" {
		areas, err := s.biteship.SearchAreas(destinationAreaID)
		if err == nil && len(areas) > 0 {
			for _, area := range areas {
				if area.ID == destinationAreaID {
					if pc, ok := area.PostalCode.(float64); ok {
						destPostalCode = int(pc)
					} else if pc, ok := area.PostalCode.(string); ok {
						fmt.Sscanf(pc, "%d", &destPostalCode)
					}
					break
				}
			}
		}
	}
	
	if destPostalCode == 0 {
		log.Printf("⚠️ Could not determine postal code for area_id: %s", destinationAreaID)
		return nil, fmt.Errorf("could not determine destination postal code")
	}

	// Calculate total weight and build items
	var totalWeight int
	var biteshipItems []GetRatesRequestItem
	var totalValue float64

	for _, item := range items {
		product, err := s.productRepo.FindByID(item.ProductID)
		if err != nil {
			continue
		}

		// FIX: Send total weight with quantity=1 to avoid Biteship double-counting bug
		itemWeight := product.Weight * item.Quantity
		if itemWeight == 0 {
			itemWeight = 500 * item.Quantity // Default 500g per item
		}
		totalWeight += itemWeight
		totalValue += float64(product.Price) * float64(item.Quantity)

		biteshipItems = append(biteshipItems, GetRatesRequestItem{
			Name:     product.Name,
			Value:    float64(product.Price) * float64(item.Quantity), // Total value
			Weight:   itemWeight,  // Total weight (already multiplied by quantity)
			Quantity: 1,           // Always 1 to avoid double-counting
		})
	}

	// Build Biteship request with postal_code (works better than area_id)
	// Use ALL couriers available in Biteship - let API return what's available for the route
	// Standard e-commerce practice: show ALL available couriers, let customer choose
	// Biteship courier codes: https://biteship.com/id/docs/api/couriers
	allCouriers := "jne,tiki,ninja,lion,sicepat,jnt,idexpress,rpx,wahana,pos,anteraja,sap,paxel,borzo,lalamove,grab,gojek,deliveree"
	
	biteshipReq := GetRatesRequest{
		OriginPostalCode:      DefaultOriginPostalCode,
		DestinationPostalCode: destPostalCode,
		Couriers:              allCouriers,
		Items:                 biteshipItems,
	}

	log.Printf("📦 Biteship rates request: origin_postal=%d, destination_postal=%d, weight=%dg", 
		DefaultOriginPostalCode, destPostalCode, totalWeight)

	// Call Biteship API
	rates, err := s.biteship.GetRates(biteshipReq)
	if err != nil {
		log.Printf("⚠️ Biteship rates failed: %v", err)
		
		// Try with fewer couriers (instant/same-day for short distance)
		biteshipReq.Couriers = "jne,sicepat,anteraja,jnt,tiki,pos"
		rates, err = s.biteship.GetRates(biteshipReq)
		if err != nil {
			log.Printf("❌ Biteship rates error: %v", err)
			return nil, err
		}
	}

	// Convert to response format
	var responseRates []dto.BiteshipRateResponse
	groupedRates := make(map[string][]dto.BiteshipRateResponse)

	for _, rate := range rates {
		rateResponse := dto.BiteshipRateResponse{
			CourierCode:        rate.CourierCode,
			CourierName:        rate.CourierName,
			CourierServiceCode: rate.CourierServiceCode,
			CourierServiceName: rate.CourierServiceName,
			Description:        rate.Description,
			Duration:           rate.Duration,
			Price:              int(rate.Price),
			Type:               rate.Type,
		}
		responseRates = append(responseRates, rateResponse)

		// Group by type
		rateType := rate.Type
		if rateType == "" {
			rateType = "regular"
		}
		groupedRates[rateType] = append(groupedRates[rateType], rateResponse)
	}

	// Format weight
	weightKg := fmt.Sprintf("%.1f kg", float64(totalWeight)/1000)
	if totalWeight < 1000 {
		weightKg = fmt.Sprintf("%d g", totalWeight)
	}

	log.Printf("✅ Got %d shipping rates from Biteship", len(responseRates))

	return &dto.GetRatesResponse{
		Rates:         responseRates,
		GroupedRates:  groupedRates,
		TotalWeight:   totalWeight,
		TotalWeightKg: weightKg,
		OriginCity:    "Semarang",
	}, nil
}

// ============================================
// LEGACY RATES (deprecated)
// ============================================

func (s *shippingService) GetShippingRates(req dto.GetShippingRatesRequest) (*dto.ShippingRatesResponse, error) {
	// Use Biteship API for shipping rates
	courier := req.Courier
	if courier == "" {
		courier = "jne,jnt,sicepat,tiki,anteraja" // Default couriers for Biteship
	}

	// Build Biteship rates request using postal_code (more reliable than area_id)
	biteshipReq := GetRatesRequest{
		Couriers: courier,
		Items: []GetRatesRequestItem{
			{
				Name:     "Package",
				Value:    100000, // Default value
				Weight:   req.Weight,
				Quantity: 1,
			},
		},
	}

	// Set origin postal code
	if req.OriginPostalCode != "" {
		postalCode, _ := strconv.Atoi(req.OriginPostalCode)
		if postalCode > 0 {
			biteshipReq.OriginPostalCode = postalCode
		} else {
			biteshipReq.OriginPostalCode = DefaultOriginPostalCode
		}
	} else {
		biteshipReq.OriginPostalCode = DefaultOriginPostalCode // Default: Semarang Tengah
	}

	// Set destination postal code
	if req.DestinationPostalCode != "" {
		postalCode, _ := strconv.Atoi(req.DestinationPostalCode)
		if postalCode > 0 {
			biteshipReq.DestinationPostalCode = postalCode
		}
	} else if req.DestinationAreaID != "" {
		// Try to parse area_id as postal code (fallback)
		postalCode, _ := strconv.Atoi(req.DestinationAreaID)
		if postalCode > 0 {
			biteshipReq.DestinationPostalCode = postalCode
		}
	}
	
	// Validate destination
	if biteshipReq.DestinationPostalCode == 0 {
		return nil, fmt.Errorf("%w: destination postal_code is required", ErrShippingNotAvailable)
	}

	rates, err := s.biteship.GetRates(biteshipReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrShippingNotAvailable, err)
	}

	// Get provider logos
	providers, _ := s.shippingRepo.GetActiveProviders()
	providerLogos := make(map[string]string)
	for _, p := range providers {
		providerLogos[strings.ToLower(p.Code)] = p.LogoURL
	}

	// Build response
	response := &dto.ShippingRatesResponse{
		Origin: dto.CityInfo{
			CityID: req.OriginCityID,
		},
		Destination: dto.CityInfo{
			CityID: req.DestinationCityID,
		},
		Weight: req.Weight,
		Rates:  []dto.ShippingRateResponse{},
	}

	for _, rate := range rates {
		rateResp := dto.ShippingRateResponse{
			ProviderCode: rate.CourierCode,
			ProviderName: rate.CourierName,
			ProviderLogo: providerLogos[strings.ToLower(rate.CourierCode)],
			ServiceCode:  rate.CourierServiceCode,
			ServiceName:  rate.CourierServiceName,
			Description:  rate.Description,
			Cost:         rate.Price,
			ETD:          rate.Duration,
		}
		response.Rates = append(response.Rates, rateResp)
	}

	return response, nil
}

func (s *shippingService) GetCartShippingPreview(sessionID string, destinationDistrictID string, courier string) (*dto.CartShippingPreviewResponse, error) {
	// Get cart
	cart, err := s.cartRepo.FindOrCreateBySessionID(sessionID)
	if err != nil {
		return nil, err
	}

	if len(cart.Items) == 0 {
		return nil, ErrCartEmpty
	}

	// Calculate total weight and subtotal
	var totalWeight int
	var subtotal float64

	for _, item := range cart.Items {
		product, err := s.productRepo.FindByID(item.ProductID)
		if err != nil {
			continue
		}

		// Get product weight from database (default 500g if not set)
		weight := product.Weight
		if weight <= 0 {
			weight = 500 // Fallback default
		}
		totalWeight += weight * item.Quantity
		subtotal += item.PriceSnapshot * float64(item.Quantity)
	}

	// Minimum weight 1000g (1kg)
	if totalWeight < 1000 {
		totalWeight = 1000
	}

	// Get shipping rates using district IDs
	ratesReq := dto.GetShippingRatesRequest{
		OriginCityID:          DefaultOriginCityID,
		OriginDistrictID:      DefaultOriginDistrictID,
		DestinationDistrictID: destinationDistrictID,
		Weight:                totalWeight,
		Courier:               courier,
	}

	ratesResp, err := s.GetShippingRates(ratesReq)
	if err != nil {
		return nil, err
	}

	return &dto.CartShippingPreviewResponse{
		CartSubtotal: subtotal,
		TotalWeight:  totalWeight,
		Rates:        ratesResp.Rates,
	}, nil
}

// ============================================
// PROVIDERS
// ============================================

func (s *shippingService) GetProviders() ([]models.ShippingProvider, error) {
	return s.shippingRepo.GetActiveProviders()
}

// ============================================
// LOCATION
// ============================================

func (s *shippingService) GetProvinces() ([]dto.ProvinceResponse, error) {
	// Use Biteship area search - search for common province names
	// Note: Biteship uses area search, not province list
	// Return a static list of Indonesian provinces for UI compatibility
	provinces := []dto.ProvinceResponse{
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
	
	log.Printf("✅ Returning %d provinces (static list for Biteship compatibility)", len(provinces))
	return provinces, nil
}

func (s *shippingService) GetCities(provinceID string) ([]dto.CityResponse, error) {
	// Static city data for dropdown - mapped to subdistricts table city_id
	citiesByProvince := map[string][]dto.CityResponse{
		"1": { // Bali
			{CityID: "114", Type: "Kota", CityName: "Denpasar", ProvinceID: "1"},
		},
		"3": { // Banten
			{CityID: "456", Type: "Kota", CityName: "Tangerang", ProvinceID: "3"},
			{CityID: "457", Type: "Kota", CityName: "Tangerang Selatan", ProvinceID: "3"},
		},
		"5": { // DI Yogyakarta
			{CityID: "501", Type: "Kota", CityName: "Yogyakarta", ProvinceID: "5"},
		},
		"6": { // DKI Jakarta
			{CityID: "151", Type: "Kota", CityName: "Jakarta Barat", ProvinceID: "6"},
			{CityID: "152", Type: "Kota", CityName: "Jakarta Pusat", ProvinceID: "6"},
			{CityID: "153", Type: "Kota", CityName: "Jakarta Selatan", ProvinceID: "6"},
			{CityID: "154", Type: "Kota", CityName: "Jakarta Timur", ProvinceID: "6"},
			{CityID: "155", Type: "Kota", CityName: "Jakarta Utara", ProvinceID: "6"},
		},
		"9": { // Jawa Barat
			{CityID: "22", Type: "Kota", CityName: "Bandung", ProvinceID: "9"},
			{CityID: "55", Type: "Kota", CityName: "Bekasi", ProvinceID: "9"},
			{CityID: "79", Type: "Kota", CityName: "Bogor", ProvinceID: "9"},
			{CityID: "115", Type: "Kota", CityName: "Depok", ProvinceID: "9"},
		},
		"10": { // Jawa Tengah
			{CityID: "399", Type: "Kota", CityName: "Semarang", ProvinceID: "10"},
		},
		"11": { // Jawa Timur
			{CityID: "444", Type: "Kota", CityName: "Surabaya", ProvinceID: "11"},
		},
		"28": { // Sulawesi Selatan
			{CityID: "267", Type: "Kota", CityName: "Makassar", ProvinceID: "28"},
		},
		"34": { // Sumatera Utara
			{CityID: "278", Type: "Kota", CityName: "Medan", ProvinceID: "34"},
		},
	}

	cities, ok := citiesByProvince[provinceID]
	if !ok {
		return []dto.CityResponse{}, nil
	}
	
	log.Printf("✅ Returning %d cities for province %s", len(cities), provinceID)
	return cities, nil
}

func (s *shippingService) GetSubdistricts(cityID string) ([]dto.SubdistrictResponse, error) {
	subdistricts, err := s.shippingRepo.GetSubdistrictsByCityID(cityID)
	if err != nil {
		return nil, err
	}

	var result []dto.SubdistrictResponse
	for _, sd := range subdistricts {
		result = append(result, dto.SubdistrictResponse{
			ID:          sd.ID,
			CityID:      sd.CityID,
			Name:        sd.Name,
			PostalCodes: sd.PostalCodes,
		})
	}

	return result, nil
}

// GetDistrictsFromAPI - deprecated, use SearchAreas for Biteship
func (s *shippingService) GetDistrictsFromAPI(cityID string) ([]dto.DistrictResponse, error) {
	log.Printf("⚠️ GetDistrictsFromAPI called - recommend using SearchAreas for Biteship integration")
	return []dto.DistrictResponse{}, nil
}

// GetSubdistrictsFromAPI - deprecated, use SearchAreas for Biteship
func (s *shippingService) GetSubdistrictsFromAPI(districtID string) ([]dto.SubdistrictAPIResponse, error) {
	log.Printf("⚠️ GetSubdistrictsFromAPI called - recommend using SearchAreas for Biteship integration")
	return []dto.SubdistrictAPIResponse{}, nil
}

func (s *shippingService) SearchSubdistricts(cityID string, query string) ([]dto.SubdistrictResponse, error) {
	subdistricts, err := s.shippingRepo.SearchSubdistricts(cityID, query)
	if err != nil {
		return nil, err
	}

	var result []dto.SubdistrictResponse
	for _, sd := range subdistricts {
		result = append(result, dto.SubdistrictResponse{
			ID:          sd.ID,
			CityID:      sd.CityID,
			Name:        sd.Name,
			PostalCodes: sd.PostalCodes,
		})
	}

	return result, nil
}

// ============================================
// ADDRESSES
// ============================================

func (s *shippingService) CreateAddress(userID int, req dto.CreateAddressRequest) (*dto.AddressResponse, error) {
	address := &models.UserAddress{
		UserID:        &userID,
		Label:         req.Label,
		RecipientName: req.RecipientName,
		Phone:         req.Phone,
		ProvinceID:    req.ProvinceID,
		ProvinceName:  req.ProvinceName,
		CityID:        req.CityID,
		CityName:      req.CityName,
		DistrictID:    req.DistrictID,
		District:      req.District,
		Subdistrict:   req.Subdistrict,
		PostalCode:    req.PostalCode,
		FullAddress:   req.FullAddress,
		IsDefault:     req.IsDefault,
		AreaID:        req.AreaID,
		AreaName:      req.AreaName,
	}

	err := s.shippingRepo.CreateAddress(address)
	if err != nil {
		return nil, err
	}

	return s.toAddressResponse(address), nil
}

func (s *shippingService) GetUserAddresses(userID int) ([]dto.AddressResponse, error) {
	addresses, err := s.shippingRepo.GetUserAddresses(userID)
	if err != nil {
		return nil, err
	}

	var result []dto.AddressResponse
	for _, a := range addresses {
		result = append(result, *s.toAddressResponse(&a))
	}

	return result, nil
}

func (s *shippingService) GetAddressByID(id int) (*dto.AddressResponse, error) {
	address, err := s.shippingRepo.GetAddressByID(id)
	if err != nil {
		return nil, ErrAddressNotFound
	}

	return s.toAddressResponse(address), nil
}

func (s *shippingService) UpdateAddress(id int, req dto.UpdateAddressRequest) (*dto.AddressResponse, error) {
	address, err := s.shippingRepo.GetAddressByID(id)
	if err != nil {
		return nil, ErrAddressNotFound
	}

	// Update fields
	if req.Label != "" {
		address.Label = req.Label
	}
	if req.RecipientName != "" {
		address.RecipientName = req.RecipientName
	}
	if req.Phone != "" {
		address.Phone = req.Phone
	}
	if req.ProvinceID != "" {
		address.ProvinceID = req.ProvinceID
	}
	if req.ProvinceName != "" {
		address.ProvinceName = req.ProvinceName
	}
	if req.CityID != "" {
		address.CityID = req.CityID
	}
	if req.CityName != "" {
		address.CityName = req.CityName
	}
	if req.DistrictID != "" {
		address.DistrictID = req.DistrictID
	}
	if req.District != "" {
		address.District = req.District
	}
	if req.Subdistrict != "" {
		address.Subdistrict = req.Subdistrict
	}
	if req.PostalCode != "" {
		address.PostalCode = req.PostalCode
	}
	if req.FullAddress != "" {
		address.FullAddress = req.FullAddress
	}
	if req.AreaID != "" {
		address.AreaID = req.AreaID
	}
	if req.AreaName != "" {
		address.AreaName = req.AreaName
	}
	address.IsDefault = req.IsDefault

	err = s.shippingRepo.UpdateAddress(address)
	if err != nil {
		return nil, err
	}

	return s.toAddressResponse(address), nil
}

func (s *shippingService) DeleteAddress(id int) error {
	return s.shippingRepo.DeleteAddress(id)
}

func (s *shippingService) SetDefaultAddress(userID, addressID int) error {
	return s.shippingRepo.SetDefaultAddress(userID, addressID)
}

// VerifyAddressOwnership checks if an address belongs to a user
func (s *shippingService) VerifyAddressOwnership(addressID, userID int) bool {
	address, err := s.shippingRepo.GetAddressByID(addressID)
	if err != nil {
		return false
	}
	if address.UserID == nil {
		return false
	}
	return *address.UserID == userID
}

// ============================================
// SHIPMENTS
// ============================================

func (s *shippingService) CreateShipmentForOrder(orderID int, providerCode, serviceName string, cost float64, etd string, weight int, destCityID, destCityName string) (*models.Shipment, error) {
	// Get provider info
	provider, err := s.shippingRepo.GetProviderByCode(providerCode)
	providerName := providerCode
	if err == nil {
		providerName = provider.Name
	}

	shipment := &models.Shipment{
		OrderID:             orderID,
		ProviderCode:        providerCode,
		ProviderName:        providerName,
		ServiceCode:         serviceName,
		ServiceName:         serviceName,
		Cost:                cost,
		ETD:                 etd,
		Weight:              weight,
		Status:              models.ShipmentStatusPending,
		OriginCityID:        DefaultOriginCityID,
		OriginCityName:      "Kota Semarang",
		DestinationCityID:   destCityID,
		DestinationCityName: destCityName,
	}

	err = s.shippingRepo.CreateShipment(shipment)
	if err != nil {
		return nil, err
	}

	return shipment, nil
}

func (s *shippingService) GetShipmentByOrderID(orderID int) (*dto.ShipmentResponse, error) {
	shipment, err := s.shippingRepo.GetShipmentByOrderID(orderID)
	if err != nil {
		return nil, ErrShipmentNotFound
	}

	// Get order code
	order, _ := s.orderRepo.FindByID(orderID)
	orderCode := ""
	if order != nil {
		orderCode = order.OrderCode
	}

	return s.toShipmentResponse(shipment, orderCode), nil
}

func (s *shippingService) GetShipmentByOrderCode(orderCode string) (*models.Shipment, error) {
	order, err := s.orderRepo.FindByOrderCode(orderCode)
	if err != nil {
		return nil, err
	}
	return s.shippingRepo.GetShipmentByOrderID(order.ID)
}

func (s *shippingService) GetShipmentByResi(resi string) (*models.Shipment, error) {
	return s.shippingRepo.GetShipmentByResi(resi)
}

func (s *shippingService) UpdateTrackingNumber(shipmentID int, trackingNumber string) error {
	return s.shippingRepo.UpdateShipmentTracking(shipmentID, trackingNumber)
}

func (s *shippingService) MarkAsShipped(shipmentID int, trackingNumber string) error {
	// Update shipment
	err := s.shippingRepo.MarkShipmentShipped(shipmentID, trackingNumber)
	if err != nil {
		return err
	}

	// Get shipment to update order
	shipment, err := s.shippingRepo.GetShipmentByID(shipmentID)
	if err != nil {
		return err
	}

	// Update order status to SHIPPED
	return s.orderRepo.MarkAsShipped(shipment.OrderID)
}

// ============================================
// TRACKING
// ============================================

func (s *shippingService) GetTracking(shipmentID int) (*dto.ShipmentResponse, error) {
	shipment, err := s.shippingRepo.GetShipmentByID(shipmentID)
	if err != nil {
		return nil, ErrShipmentNotFound
	}

	// Get order code
	order, _ := s.orderRepo.FindByID(shipment.OrderID)
	orderCode := ""
	if order != nil {
		orderCode = order.OrderCode
	}

	return s.toShipmentResponse(shipment, orderCode), nil
}

func (s *shippingService) RefreshTracking(shipmentID int) error {
	shipment, err := s.shippingRepo.GetShipmentByID(shipmentID)
	if err != nil {
		return ErrShipmentNotFound
	}

	// Use Biteship tracking if available
	if shipment.BiteshipTrackingID != "" {
		tracking, err := s.biteship.GetTracking(shipment.BiteshipTrackingID)
		if err != nil {
			log.Printf("⚠️ Biteship tracking failed for %s: %v", shipment.BiteshipTrackingID, err)
			return err
		}

		// Map Biteship status to our shipment status
		newStatus := MapBiteshipStatusToShipmentStatus(tracking.Status)
		if newStatus != string(shipment.Status) {
			s.shippingRepo.UpdateShipmentStatus(shipmentID, models.ShipmentStatus(newStatus))

			// If delivered, update order status
			if newStatus == "DELIVERED" {
				s.shippingRepo.MarkShipmentDelivered(shipmentID)
				s.orderRepo.UpdateStatus(shipment.OrderID, models.OrderStatusDelivered)
			}
		}

		// Add tracking events from Biteship history
		for _, h := range tracking.History {
			eventTime, _ := time.Parse(time.RFC3339, h.UpdatedAt)
			event := &models.TrackingEvent{
				ShipmentID:  shipmentID,
				Status:      h.Status,
				Description: h.Note,
				EventTime:   &eventTime,
				RawData: map[string]any{
					"status":     h.Status,
					"note":       h.Note,
					"updated_at": h.UpdatedAt,
					"source":     "biteship",
				},
			}
			s.shippingRepo.AddTrackingEvent(event)
		}

		log.Printf("✅ Updated tracking from Biteship for shipment %d - Status: %s", shipmentID, tracking.Status)
		return nil
	}

	// No Biteship tracking available
	if shipment.TrackingNumber == "" {
		return errors.New("no tracking number available")
	}

	log.Printf("⚠️ No Biteship tracking ID for shipment %d, tracking number: %s", shipmentID, shipment.TrackingNumber)
	return nil
}

// RunTrackingJob polls tracking for all shipped orders
func (s *shippingService) RunTrackingJob() error {
	shipments, err := s.shippingRepo.GetShipmentsForTracking()
	if err != nil {
		return err
	}

	log.Printf("📦 Running tracking job for %d shipments", len(shipments))

	for _, shipment := range shipments {
		err := s.RefreshTracking(shipment.ID)
		if err != nil {
			log.Printf("⚠️ Failed to refresh tracking for shipment %d: %v", shipment.ID, err)
			continue
		}
		log.Printf("✅ Updated tracking for shipment %d", shipment.ID)

		// Rate limiting - wait between API calls
		time.Sleep(500 * time.Millisecond)
	}

	return nil
}

// ============================================
// HELPERS
// ============================================

func (s *shippingService) toAddressResponse(a *models.UserAddress) *dto.AddressResponse {
	return &dto.AddressResponse{
		ID:            a.ID,
		Label:         a.Label,
		RecipientName: a.RecipientName,
		Phone:         a.Phone,
		ProvinceID:    a.ProvinceID,
		ProvinceName:  a.ProvinceName,
		CityID:        a.CityID,
		CityName:      a.CityName,
		DistrictID:    a.DistrictID,
		District:      a.District,
		Subdistrict:   a.Subdistrict,
		PostalCode:    a.PostalCode,
		FullAddress:   a.FullAddress,
		IsDefault:     a.IsDefault,
		AreaID:        a.AreaID,
		AreaName:      a.AreaName,
	}
}

func (s *shippingService) toShipmentResponse(shipment *models.Shipment, orderCode string) *dto.ShipmentResponse {
	resp := &dto.ShipmentResponse{
		ID:             shipment.ID,
		OrderID:        shipment.OrderID,
		OrderCode:      orderCode,
		ProviderCode:   shipment.ProviderCode,
		ProviderName:   shipment.ProviderName,
		ServiceCode:    shipment.ServiceCode,
		ServiceName:    shipment.ServiceName,
		Cost:           shipment.Cost,
		ETD:            shipment.ETD,
		Weight:         shipment.Weight,
		TrackingNumber: shipment.TrackingNumber,
		Status:         string(shipment.Status),
		Origin:         shipment.OriginCityName,
		Destination:    shipment.DestinationCityName,
	}

	if shipment.ShippedAt != nil {
		resp.ShippedAt = shipment.ShippedAt.Format("2006-01-02 15:04:05")
	}
	if shipment.DeliveredAt != nil {
		resp.DeliveredAt = shipment.DeliveredAt.Format("2006-01-02 15:04:05")
	}

	// Add tracking history
	for _, e := range shipment.TrackingHistory {
		eventResp := dto.TrackingEventResponse{
			Status:      e.Status,
			Description: e.Description,
			Location:    e.Location,
		}
		if e.EventTime != nil {
			eventResp.EventTime = e.EventTime.Format("2006-01-02 15:04:05")
		}
		resp.TrackingHistory = append(resp.TrackingHistory, eventResp)
	}

	return resp
}

// AddressToSnapshot converts UserAddress to snapshot for order
func AddressToSnapshot(a *models.UserAddress) models.ShippingAddressSnapshot {
	return models.ShippingAddressSnapshot{
		RecipientName: a.RecipientName,
		Phone:         a.Phone,
		ProvinceID:    a.ProvinceID,
		ProvinceName:  a.ProvinceName,
		CityID:        a.CityID,
		CityName:      a.CityName,
		DistrictID:    a.DistrictID,
		District:      a.District,
		Subdistrict:   a.Subdistrict,
		PostalCode:    a.PostalCode,
		FullAddress:   a.FullAddress,
	}
}

// GuestAddressToSnapshot converts guest address to snapshot
// Updated for Biteship-native: includes area_id and area_name
func GuestAddressToSnapshot(a *dto.GuestShippingAddress) models.ShippingAddressSnapshot {
	// Use area_name as city_name if city_name is not provided
	cityName := a.CityName
	if cityName == "" && a.AreaName != "" {
		cityName = a.AreaName
	}
	
	return models.ShippingAddressSnapshot{
		RecipientName: a.RecipientName,
		Phone:         a.Phone,
		ProvinceID:    a.ProvinceID,
		ProvinceName:  a.ProvinceName,
		CityID:        a.CityID,
		CityName:      cityName,
		DistrictID:    a.DistrictID,
		District:      a.District,
		Subdistrict:   a.Subdistrict,
		PostalCode:    a.PostalCode,
		FullAddress:   a.FullAddress,
	}
}

// SnapshotToJSON converts snapshot to JSON for storage
func SnapshotToJSON(snapshot models.ShippingAddressSnapshot) []byte {
	data, _ := json.Marshal(snapshot)
	return data
}


// ============================================
// BITESHIP DRAFT ORDERS
// ============================================

// CreateDraftOrderForCheckout creates a Biteship draft order for an order
func (s *shippingService) CreateDraftOrderForCheckout(orderID int, params CreateDraftOrderParams) (*dto.DraftOrderResponse, error) {
	// Build draft order request
	var items []CreateDraftOrderItem
	for _, item := range params.Items {
		items = append(items, CreateDraftOrderItem{
			Name:        item.Name,
			Description: item.Description,
			Value:       item.Value,
			Weight:      item.Weight,
			Quantity:    item.Quantity,
		})
	}

	req := CreateDraftOrderRequest{
		ShipperContactName:      "ZAVERA Fashion Store",
		ShipperContactPhone:     "081234567890",
		ShipperOrganization:     "ZAVERA",
		OriginContactName:       params.OriginContactName,
		OriginContactPhone:      params.OriginContactPhone,
		OriginAddress:           params.OriginAddress,
		OriginPostalCode:        params.OriginPostalCode,
		OriginAreaID:            params.OriginAreaID,
		DestinationContactName:  params.DestinationContactName,
		DestinationContactPhone: params.DestinationContactPhone,
		DestinationAddress:      params.DestinationAddress,
		DestinationPostalCode:   params.DestinationPostalCode,
		DestinationAreaID:       params.DestinationAreaID,
		CourierCode:             params.CourierCode,
		CourierServiceCode:      params.CourierServiceCode,
		DeliveryType:            "now",
		Items:                   items,
	}

	// Create draft order via Biteship API
	draftOrderID, err := s.biteship.CreateDraftOrder(req)
	if err != nil {
		log.Printf("❌ Failed to create Biteship draft order for order %d: %v", orderID, err)
		return nil, err
	}

	// Get shipment and update with draft order ID
	shipment, err := s.shippingRepo.GetShipmentByOrderID(orderID)
	if err != nil {
		log.Printf("⚠️ Shipment not found for order %d, draft order created but not linked", orderID)
	} else {
		err = s.shippingRepo.UpdateShipmentBiteshipIDs(shipment.ID, draftOrderID, "", "", "")
		if err != nil {
			log.Printf("⚠️ Failed to update shipment with draft order ID: %v", err)
		}
	}

	log.Printf("✅ Created Biteship draft order %s for order %d", draftOrderID, orderID)

	return &dto.DraftOrderResponse{
		ID:      draftOrderID,
		Success: true,
		Message: "Draft order created successfully",
	}, nil
}

// ConfirmDraftOrder confirms a Biteship draft order after payment
func (s *shippingService) ConfirmDraftOrder(orderID int) (*dto.OrderConfirmationResponse, error) {
	// Get shipment to find draft order ID
	shipment, err := s.shippingRepo.GetShipmentByOrderID(orderID)
	if err != nil {
		return nil, fmt.Errorf("shipment not found for order %d: %w", orderID, err)
	}

	if shipment.BiteshipDraftOrderID == "" {
		return nil, fmt.Errorf("no draft order ID found for order %d", orderID)
	}

	// Confirm draft order via Biteship API
	order, err := s.biteship.ConfirmDraftOrder(shipment.BiteshipDraftOrderID)
	if err != nil {
		log.Printf("❌ Failed to confirm Biteship draft order %s: %v", shipment.BiteshipDraftOrderID, err)
		return nil, err
	}

	// Update shipment with confirmed order details
	err = s.shippingRepo.UpdateShipmentBiteshipIDs(
		shipment.ID,
		"", // Don't update draft order ID
		order.ID,
		order.TrackingID,
		order.WaybillID,
	)
	if err != nil {
		log.Printf("⚠️ Failed to update shipment with order confirmation: %v", err)
	}

	// Update tracking number with waybill
	if order.WaybillID != "" {
		s.shippingRepo.UpdateShipmentTracking(shipment.ID, order.WaybillID)
	}

	log.Printf("✅ Confirmed Biteship order %s - Waybill: %s, Tracking: %s", order.ID, order.WaybillID, order.TrackingID)

	return &dto.OrderConfirmationResponse{
		Success:    true,
		OrderID:    order.ID,
		WaybillID:  order.WaybillID,
		TrackingID: order.TrackingID,
	}, nil
}

// RefreshBiteshipTracking refreshes tracking using Biteship API
func (s *shippingService) RefreshBiteshipTracking(shipmentID int) error {
	shipment, err := s.shippingRepo.GetShipmentByID(shipmentID)
	if err != nil {
		return ErrShipmentNotFound
	}

	// Use Biteship tracking ID if available
	if shipment.BiteshipTrackingID != "" {
		tracking, err := s.biteship.GetTracking(shipment.BiteshipTrackingID)
		if err != nil {
			return err
		}

		// Map Biteship status to our shipment status
		newStatus := MapBiteshipStatusToShipmentStatus(tracking.Status)
		if newStatus != string(shipment.Status) {
			s.shippingRepo.UpdateShipmentStatus(shipmentID, models.ShipmentStatus(newStatus))

			// If delivered, update order status
			if newStatus == "DELIVERED" {
				s.shippingRepo.MarkShipmentDelivered(shipmentID)
				s.orderRepo.UpdateStatus(shipment.OrderID, models.OrderStatusDelivered)
			}
		}

		// Add tracking events from Biteship history
		for _, h := range tracking.History {
			eventTime, _ := time.Parse(time.RFC3339, h.UpdatedAt)
			event := &models.TrackingEvent{
				ShipmentID:  shipmentID,
				Status:      h.Status,
				Description: h.Note,
				EventTime:   &eventTime,
				RawData: map[string]any{
					"status":     h.Status,
					"note":       h.Note,
					"updated_at": h.UpdatedAt,
					"source":     "biteship",
				},
			}
			s.shippingRepo.AddTrackingEvent(event)
		}

		return nil
	}

	// Fallback to legacy tracking if no Biteship tracking ID
	return s.RefreshTracking(shipmentID)
}

// CategorizeRate categorizes a shipping rate based on service type
func CategorizeRate(serviceType string) string {
	serviceType = strings.ToLower(serviceType)
	
	switch {
	case strings.Contains(serviceType, "same_day") || strings.Contains(serviceType, "instant"):
		return "Express"
	case strings.Contains(serviceType, "standard") || strings.Contains(serviceType, "express"):
		return "Regular"
	case strings.Contains(serviceType, "economy") || strings.Contains(serviceType, "cargo"):
		return "Economy"
	default:
		return "Regular"
	}
}
