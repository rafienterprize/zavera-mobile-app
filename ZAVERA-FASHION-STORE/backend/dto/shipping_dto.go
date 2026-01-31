package dto

// ============================================
// SHIPPING CATEGORY TYPES
// ============================================

// ShippingCategory represents the category of shipping service
type ShippingCategory string

const (
	ShippingCategoryExpress ShippingCategory = "Express"
	ShippingCategoryRegular ShippingCategory = "Regular"
	ShippingCategoryEconomy ShippingCategory = "Economy"
	ShippingCategorySameDay ShippingCategory = "SameDay"
	ShippingCategoryCargo   ShippingCategory = "Cargo" // Hidden from UI
)

// ============================================
// SHIPPING DTOs
// ============================================

// GetShippingRatesRequest represents request to get shipping rates
// Updated for Biteship API - uses area_id or postal_code
type GetShippingRatesRequest struct {
	// Legacy fields for display
	OriginCityID          string `json:"origin_city_id"`
	DestinationCityID     string `json:"destination_city_id"`
	
	// Legacy district IDs (kept for backward compatibility)
	OriginDistrictID      string `json:"origin_district_id"`
	DestinationDistrictID string `json:"destination_district_id"`
	
	// Biteship area-based fields (preferred)
	OriginAreaID          string `json:"origin_area_id"`
	DestinationAreaID     string `json:"destination_area_id"`
	OriginPostalCode      string `json:"origin_postal_code"`
	DestinationPostalCode string `json:"destination_postal_code"`
	
	Weight                int    `json:"weight" binding:"required,min=1"` // in grams
	Courier               string `json:"courier"`                         // optional: filter by specific courier
}

// ShippingRateResponse represents a single shipping rate option
type ShippingRateResponse struct {
	ProviderCode     string           `json:"provider_code"`
	ProviderName     string           `json:"provider_name"`
	ProviderLogo     string           `json:"provider_logo"`
	ServiceCode      string           `json:"service_code"`
	ServiceName      string           `json:"service_name"`
	Description      string           `json:"description"`
	Cost             float64          `json:"cost"`
	ETD              string           `json:"etd"`               // e.g., "1-2"
	ETADate          string           `json:"eta_date"`          // e.g., "Tiba 12 - 13 Jan"
	ShippingCategory ShippingCategory `json:"shipping_category"` // Express, Regular, Economy, SameDay
	IsAbsurdPrice    bool             `json:"is_absurd_price"`   // Price > 5x REG price
}

// ShippingRatesResponse represents list of available shipping rates
type ShippingRatesResponse struct {
	Origin      CityInfo               `json:"origin"`
	Destination CityInfo               `json:"destination"`
	Weight      int                    `json:"weight"`
	Rates       []ShippingRateResponse `json:"rates"`
}

// CityInfo represents city information
type CityInfo struct {
	CityID       string `json:"city_id"`
	CityName     string `json:"city_name"`
	ProvinceName string `json:"province_name,omitempty"`
	PostalCode   string `json:"postal_code,omitempty"`
}

// SelectShippingRequest represents request to select shipping for checkout
type SelectShippingRequest struct {
	AddressID    int    `json:"address_id"`    // Use saved address
	ProviderCode string `json:"provider_code" binding:"required"`
	ServiceCode  string `json:"service_code" binding:"required"`
	// For guest checkout without saved address
	ShippingAddress *GuestShippingAddress `json:"shipping_address,omitempty"`
}

// GuestShippingAddress for guest checkout
// Updated for Biteship-native: uses area_id + postal_code instead of city_id/district_id
type GuestShippingAddress struct {
	RecipientName string `json:"recipient_name" binding:"required"`
	Phone         string `json:"phone" binding:"required"`
	
	// Biteship-native fields (preferred)
	AreaID        string `json:"area_id"`        // Biteship area ID
	AreaName      string `json:"area_name"`      // Full area name from Biteship
	PostalCode    string `json:"postal_code" binding:"required"` // Required for shipping calculation
	
	// Legacy fields (kept for backward compatibility)
	ProvinceID    string `json:"province_id"`
	ProvinceName  string `json:"province_name"`
	CityID        string `json:"city_id"`
	CityName      string `json:"city_name"`
	DistrictID    string `json:"district_id"`   // Kecamatan ID
	District      string `json:"district"`      // Kecamatan name
	Subdistrict   string `json:"subdistrict"`   // Kelurahan name
	
	FullAddress   string `json:"full_address" binding:"required"`
}

// ShippingSelectionResponse represents response after selecting shipping
type ShippingSelectionResponse struct {
	OrderCode       string                 `json:"order_code"`
	ShippingCost    float64                `json:"shipping_cost"`
	TotalAmount     float64                `json:"total_amount"`
	Provider        string                 `json:"provider"`
	Service         string                 `json:"service"`
	ETD             string                 `json:"etd"`
	ShippingAddress ShippingAddressDisplay `json:"shipping_address"`
	ShippingLocked  bool                   `json:"shipping_locked"`
}

// ShippingAddressDisplay for displaying address
type ShippingAddressDisplay struct {
	RecipientName string `json:"recipient_name"`
	Phone         string `json:"phone"`
	FullAddress   string `json:"full_address"`
	CityName      string `json:"city_name"`
	ProvinceName  string `json:"province_name"`
	PostalCode    string `json:"postal_code"`
}

// ============================================
// ADDRESS DTOs
// ============================================

// CreateAddressRequest represents request to create a new address
type CreateAddressRequest struct {
	Label         string `json:"label"`
	RecipientName string `json:"recipient_name" binding:"required"`
	Phone         string `json:"phone" binding:"required"`
	ProvinceID    string `json:"province_id"`
	ProvinceName  string `json:"province_name"`
	CityID        string `json:"city_id"`
	CityName      string `json:"city_name"`
	DistrictID    string `json:"district_id"`   // Kecamatan ID for shipping calculation
	District      string `json:"district"`      // Kecamatan name
	Subdistrict   string `json:"subdistrict"`   // Kelurahan name
	PostalCode    string `json:"postal_code"`
	FullAddress   string `json:"full_address" binding:"required"`
	IsDefault     bool   `json:"is_default"`
	// Biteship native fields
	AreaID        string `json:"area_id"`       // Biteship area_id for shipping
	AreaName      string `json:"area_name"`     // Full area name from Biteship
}

// UpdateAddressRequest represents request to update an address
type UpdateAddressRequest struct {
	Label         string `json:"label"`
	RecipientName string `json:"recipient_name"`
	Phone         string `json:"phone"`
	ProvinceID    string `json:"province_id"`
	ProvinceName  string `json:"province_name"`
	CityID        string `json:"city_id"`
	CityName      string `json:"city_name"`
	DistrictID    string `json:"district_id"`   // Kecamatan ID for shipping calculation
	District      string `json:"district"`      // Kecamatan name
	Subdistrict   string `json:"subdistrict"`   // Kelurahan name
	PostalCode    string `json:"postal_code"`
	FullAddress   string `json:"full_address"`
	IsDefault     bool   `json:"is_default"`
	// Biteship native fields
	AreaID        string `json:"area_id"`       // Biteship area_id for shipping
	AreaName      string `json:"area_name"`     // Full area name from Biteship
}

// AddressResponse represents address in API response
type AddressResponse struct {
	ID            int    `json:"id"`
	Label         string `json:"label"`
	RecipientName string `json:"recipient_name"`
	Phone         string `json:"phone"`
	ProvinceID    string `json:"province_id"`
	ProvinceName  string `json:"province_name"`
	CityID        string `json:"city_id"`
	CityName      string `json:"city_name"`
	DistrictID    string `json:"district_id"`   // Kecamatan ID for shipping calculation
	District      string `json:"district"`      // Kecamatan name
	Subdistrict   string `json:"subdistrict"`   // Kelurahan name
	PostalCode    string `json:"postal_code"`
	FullAddress   string `json:"full_address"`
	IsDefault     bool   `json:"is_default"`
	// Biteship native fields
	AreaID        string `json:"area_id"`       // Biteship area_id for shipping
	AreaName      string `json:"area_name"`     // Full area name from Biteship
}

// ============================================
// SHIPMENT DTOs
// ============================================

// ShipmentResponse represents shipment in API response
type ShipmentResponse struct {
	ID              int                       `json:"id"`
	OrderID         int                       `json:"order_id"`
	OrderCode       string                    `json:"order_code"`
	ProviderCode    string                    `json:"provider_code"`
	ProviderName    string                    `json:"provider_name"`
	ServiceCode     string                    `json:"service_code"`
	ServiceName     string                    `json:"service_name"`
	Cost            float64                   `json:"cost"`
	ETD             string                    `json:"etd"`
	Weight          int                       `json:"weight"`
	TrackingNumber  string                    `json:"tracking_number"`
	Status          string                    `json:"status"`
	Origin          string                    `json:"origin"`
	Destination     string                    `json:"destination"`
	ShippedAt       string                    `json:"shipped_at,omitempty"`
	DeliveredAt     string                    `json:"delivered_at,omitempty"`
	TrackingHistory []TrackingEventResponse   `json:"tracking_history,omitempty"`
}

// TrackingEventResponse represents tracking event in API response
type TrackingEventResponse struct {
	Status      string `json:"status"`
	Description string `json:"description"`
	Location    string `json:"location"`
	EventTime   string `json:"event_time"`
}

// UpdateTrackingRequest represents request to update tracking number
type UpdateTrackingRequest struct {
	TrackingNumber string `json:"tracking_number" binding:"required"`
}

// ============================================
// LOCATION DTOs (for Kommerce API)
// ============================================

// ProvinceResponse represents province data
type ProvinceResponse struct {
	ProvinceID   string `json:"province_id"`
	ProvinceName string `json:"province_name"`
}

// CityResponse represents city data
type CityResponse struct {
	CityID       string `json:"city_id"`
	CityName     string `json:"city_name"`
	ProvinceID   string `json:"province_id"`
	ProvinceName string `json:"province_name"`
	Type         string `json:"type"` // Kabupaten/Kota
	PostalCode   string `json:"postal_code"`
}

// SubdistrictResponse represents subdistrict (kecamatan) data
type SubdistrictResponse struct {
	ID          int      `json:"id"`
	CityID      string   `json:"city_id"`
	Name        string   `json:"name"`
	PostalCodes []string `json:"postal_codes"`
}

// DistrictResponse represents district (kecamatan) data from Kommerce API (legacy)
// For Biteship integration, use AreaResponse instead
type DistrictResponse struct {
	DistrictID   string `json:"district_id"`
	DistrictName string `json:"district_name"`
	CityID       string `json:"city_id"`
}

// SubdistrictAPIResponse represents subdistrict (kelurahan) data from Kommerce API (legacy)
// For Biteship integration, use AreaResponse instead
type SubdistrictAPIResponse struct {
	SubdistrictID   string `json:"subdistrict_id"`
	SubdistrictName string `json:"subdistrict_name"`
	DistrictID      string `json:"district_id"`
	PostalCode      string `json:"postal_code"`
}

// ============================================
// CHECKOUT WITH SHIPPING DTOs
// ============================================

// CheckoutWithShippingRequest represents checkout request with shipping
// Updated for Biteship-native: accepts courier_code/courier_service_code OR provider_code/service_code
type CheckoutWithShippingRequest struct {
	CustomerName    string                `json:"customer_name" binding:"required"`
	CustomerEmail   string                `json:"customer_email" binding:"required,email"`
	CustomerPhone   string                `json:"customer_phone" binding:"required"`
	Notes           string                `json:"notes,omitempty"`
	AddressID       *int                  `json:"address_id,omitempty"`
	ShippingAddress *GuestShippingAddress `json:"shipping_address,omitempty"`
	
	// Biteship-native fields (preferred)
	CourierCode        string `json:"courier_code"`         // e.g., "jne", "jnt", "sicepat"
	CourierServiceCode string `json:"courier_service_code"` // e.g., "reg", "oke", "yes"
	
	// Legacy fields (kept for backward compatibility)
	ProviderCode    string `json:"provider_code"`
	ServiceCode     string `json:"service_code"`
}

// CheckoutWithShippingResponse represents checkout response with shipping details
type CheckoutWithShippingResponse struct {
	OrderID         int                    `json:"order_id"`
	OrderCode       string                 `json:"order_code"`
	Subtotal        float64                `json:"subtotal"`
	ShippingCost    float64                `json:"shipping_cost"`
	TotalAmount     float64                `json:"total_amount"`
	Status          string                 `json:"status"`
	ShippingLocked  bool                   `json:"shipping_locked"`
	Provider        string                 `json:"provider"`
	Service         string                 `json:"service"`
	ETD             string                 `json:"etd"`
	ShippingAddress ShippingAddressDisplay `json:"shipping_address"`
}

// CartShippingPreviewRequest for previewing shipping cost from cart
type CartShippingPreviewRequest struct {
	DestinationCityID     string `json:"destination_city_id"`                         // Legacy, for display
	DestinationDistrictID string `json:"destination_district_id" binding:"required"`  // Required for API
	Courier               string `json:"courier,omitempty"`                           // optional filter
}

// CartShippingPreviewResponse for cart shipping preview (Tokopedia/Shopee style)
type CartShippingPreviewResponse struct {
	CartSubtotal    float64                           `json:"cart_subtotal"`
	TotalWeight     int                               `json:"total_weight"`      // in grams
	TotalWeightKg   string                            `json:"total_weight_kg"`   // "1.2 kg"
	OriginCity      string                            `json:"origin_city"`       // "Semarang"
	DestinationCity string                            `json:"destination_city"`  // From address
	GroupedRates    map[string][]ShippingRateResponse `json:"grouped_rates"`     // Grouped by category: REGULER, EXPRESS, SAME DAY
	Rates           []ShippingRateResponse            `json:"rates"`             // Flat list, sorted by priority
	RegularMinPrice float64                           `json:"regular_min_price"` // For absurd price reference
}


// ============================================
// TRACKING DTOs
// ============================================

// TrackingResponse represents tracking information for a shipment
type TrackingResponse struct {
	OrderCode   string                     `json:"order_code"`
	Resi        string                     `json:"resi"`
	CourierName string                     `json:"courier_name"`
	Status      string                     `json:"status"`
	Origin      string                     `json:"origin"`
	Destination string                     `json:"destination"`
	History     []TrackingHistoryResponse  `json:"history"`
}

// TrackingHistoryResponse represents a single tracking event
type TrackingHistoryResponse struct {
	Note      string `json:"note"`
	Status    string `json:"status"`
	UpdatedAt string `json:"updated_at"`
}
