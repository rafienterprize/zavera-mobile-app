package dto

// ============================================
// BITESHIP NATIVE DTOs
// ============================================

// AreaResponse - Area from Biteship /v1/maps/areas search
type AreaResponse struct {
	AreaID      string `json:"area_id"`
	Name        string `json:"name"`
	PostalCode  string `json:"postal_code"`
	Province    string `json:"province,omitempty"`
	City        string `json:"city,omitempty"`
	District    string `json:"district,omitempty"`
	Subdistrict string `json:"subdistrict,omitempty"`
}

// BiteshipAreaResponse - Area from /v1/maps/areas search
type BiteshipAreaResponse struct {
	ID                                string `json:"id"`
	Name                              string `json:"name"`
	CountryName                       string `json:"country_name"`
	CountryCode                       string `json:"country_code"`
	AdministrativeDivisionLevel1Name  string `json:"administrative_division_level_1_name"`
	AdministrativeDivisionLevel1Type  string `json:"administrative_division_level_1_type"`
	AdministrativeDivisionLevel2Name  string `json:"administrative_division_level_2_name"`
	AdministrativeDivisionLevel2Type  string `json:"administrative_division_level_2_type"`
	AdministrativeDivisionLevel3Name  string `json:"administrative_division_level_3_name"`
	AdministrativeDivisionLevel3Type  string `json:"administrative_division_level_3_type"`
	PostalCode                        int    `json:"postal_code"`
}

// DraftOrderResponse - Response from creating draft order
type DraftOrderResponse struct {
	ID        string `json:"id"`
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	OrderID   string `json:"order_id,omitempty"`
}

// OrderConfirmationResponse - Response from confirming draft order
type OrderConfirmationResponse struct {
	Success       bool   `json:"success"`
	Message       string `json:"message,omitempty"`
	OrderID       string `json:"order_id"`
	WaybillID     string `json:"waybill_id"`
	TrackingID    string `json:"tracking_id"`
	CourierCode   string `json:"courier_code"`
	CourierName   string `json:"courier_name"`
	Price         int    `json:"price"`
}

// BiteshipRateRequest - Request for /v1/rates/couriers
type BiteshipRateRequest struct {
	OriginAreaID      string              `json:"origin_area_id"`
	DestinationAreaID string              `json:"destination_area_id"`
	Couriers          string              `json:"couriers,omitempty"`
	Items             []BiteshipRateItem  `json:"items"`
}

// BiteshipRateItem - Item in rate request
type BiteshipRateItem struct {
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Value       float64 `json:"value"`
	Length      int     `json:"length,omitempty"`
	Width       int     `json:"width,omitempty"`
	Height      int     `json:"height,omitempty"`
	Weight      int     `json:"weight"`
	Quantity    int     `json:"quantity"`
}

// BiteshipRateResponse - Rate from Biteship API
type BiteshipRateResponse struct {
	CourierCode        string `json:"courier_code"`
	CourierName        string `json:"courier_name"`
	CourierServiceCode string `json:"courier_service_code"`
	CourierServiceName string `json:"courier_service_name"`
	Description        string `json:"description"`
	Duration           string `json:"duration"`
	Price              int    `json:"price"`
	Type               string `json:"type"` // instant, same_day, express, regular, economy
}

// BiteshipShippingAddress - Address with Biteship area_id
type BiteshipShippingAddress struct {
	RecipientName string `json:"recipient_name"`
	Phone         string `json:"phone"`
	AreaID        string `json:"area_id"`
	AreaName      string `json:"area_name"`
	PostalCode    string `json:"postal_code"`
	FullAddress   string `json:"full_address"`
}

// GetRatesRequest - Frontend request for shipping rates
type GetRatesRequest struct {
	DestinationAreaID   string `json:"destination_area_id"`
	DestinationPostalCode string `json:"destination_postal_code"`
}

// GetRatesResponse - Response with shipping rates
type GetRatesResponse struct {
	Rates         []BiteshipRateResponse            `json:"rates"`
	GroupedRates  map[string][]BiteshipRateResponse `json:"grouped_rates"`
	TotalWeight   int                               `json:"total_weight"`
	TotalWeightKg string                            `json:"total_weight_kg"`
	OriginCity    string                            `json:"origin_city"`
}

// CheckoutWithBiteshipRequest - Checkout request with Biteship shipping
type CheckoutWithBiteshipRequest struct {
	CustomerName       string                  `json:"customer_name" binding:"required"`
	CustomerEmail      string                  `json:"customer_email" binding:"required,email"`
	CustomerPhone      string                  `json:"customer_phone" binding:"required"`
	CourierCode        string                  `json:"courier_code" binding:"required"`
	CourierServiceCode string                  `json:"courier_service_code" binding:"required"`
	ShippingAddress    BiteshipShippingAddress `json:"shipping_address" binding:"required"`
	Notes              string                  `json:"notes,omitempty"`
}
