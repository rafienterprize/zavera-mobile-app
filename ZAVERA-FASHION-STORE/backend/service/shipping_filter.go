package service

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"zavera/dto"
)

// ============================================
// SHIPPING SERVICE FILTER FOR FASHION E-COMMERCE
// ============================================
// This filter ensures ZAVERA checkout behaves like Shopee/Tokopedia/Zalora
// for fashion parcels, NOT like a cargo booking app.
//
// RULES:
// 1. Hide cargo/trucking services for weight <= 5kg
// 2. Sort by: Regular > Economy > Express > Same Day
// 3. Convert ETD to real dates (e.g., "Tiba 12 - 13 Jan")
// 4. Block absurd prices (> 5x REG price)
// 5. Group by category: REGULER, EXPRESS, SAME DAY

// ShippingRate represents a shipping rate for filtering
// This is a generic type used by the filter, compatible with Biteship rates
type ShippingRate struct {
	ProviderCode string
	ProviderName string
	ServiceCode  string
	ServiceName  string
	Description  string
	Cost         float64
	ETD          string
	Category     dto.ShippingCategory
}

// ServiceClassification defines how a service should be handled
type ServiceClassification struct {
	Category    dto.ShippingCategory
	DisplayName string
	Show        bool   // Whether to show in checkout
	Reason      string // Why hidden (for logging)
	Priority    int    // Sort priority (lower = show first)
}

// CargoServiceCodes - Services that should NEVER appear for fashion e-commerce
// These are trucking, cargo, bulk freight services
var CargoServiceCodes = map[string]bool{
	// TIKI Cargo/Trucking Services
	"T15":   true, // Motor Di Bawah 150cc
	"T25":   true, // Motor Di Bawah 250cc
	"T60":   true, // Motor Di Atas 500cc
	"TRC":   true, // Trucking
	"TRX":   true, // Trucking Express
	"SRP":   true, // Special cargo
	"DAT":   true, // Document/cargo service
	
	// JNE Cargo/Trucking Services
	"JTR":   true, // JNE Trucking
	"TRUCK": true, // Trucking
	"CARGO": true, // Cargo
	"FTL":   true, // Full Truck Load
	"LTL":   true, // Less Than Truck Load
	
	// SiCepat Cargo Services
	"GOKIL": true, // Cargo service
	"KARGO": true, // Cargo
	
	// J&T Cargo Services
	"JSD":   true, // J&T Super Deal (bulk)
	
	// POS Cargo Services
	"Pos Jumbo Ekonomi": true,
	"Paket Jumbo":       true,
}

// ExpressServiceCodes - Fast delivery services (same day, next day)
var ExpressServiceCodes = map[string]bool{
	// JNE Express
	"YES":    true, // Yakin Esok Sampai
	"CTCYES": true, // City Courier YES
	"CTCSPS": true, // City Courier Super Speed
	"SS":     true, // Super Speed
	
	// TIKI Express
	"ONS": true, // Over Night Service
	
	// SiCepat Express
	"SIUNT": true, // SiCepat Untung (express)
	"BEST":  true, // Best Service (express)
	
	// J&T Express
	"EZ": true, // J&T Express
	
	// POS Express
	"Pos Nextday": true,
	"Q9":          true, // Q9 Express
}

// SameDayServiceCodes - Same day delivery services
var SameDayServiceCodes = map[string]bool{
	// TIKI Same Day
	"SDS": true, // Same Day Service
	
	// SiCepat Same Day
	"HALU": true, // Same Day
	
	// POS Same Day
	"Pos Sameday": true,
}

// EconomyServiceCodes - Budget/economy services
var EconomyServiceCodes = map[string]bool{
	// JNE Economy
	"OKE": true, // Ongkos Kirim Ekonomis
	
	// TIKI Economy
	"ECO": true, // Economy
	
	// SiCepat Economy
	"SIGESIT": true, // SiCepat Gesit (economy)
	
	// POS Economy
	"Pos Reguler": true, // Actually regular but cheaper
}

// SimplifiedCourierNames - Clean courier names for display
var SimplifiedCourierNames = map[string]string{
	"jne":      "JNE",
	"jnt":      "J&T Express",
	"tiki":     "TIKI",
	"sicepat":  "SiCepat",
	"anteraja": "AnterAja",
	"pos":      "POS Indonesia",
}

// ServiceDisplayNames - Friendly service names
var ServiceDisplayNames = map[string]string{
	// JNE
	"REG":    "Regular",
	"YES":    "Yakin Esok Sampai",
	"OKE":    "Ongkos Kirim Ekonomis",
	"CTCYES": "City Courier Express",
	"CTCSPS": "City Courier Super Speed",
	
	// TIKI
	"SDS": "Same Day Service",
	"ONS": "Over Night Service",
	"ECO": "Economy",
	
	// SiCepat
	"BEST":   "Best Service",
	"HALU":   "Same Day",
	"SIUNT":  "SiCepat Untung",
	
	// J&T
	"EZ": "Express",
	
	// POS
	"Pos Nextday": "Nextday",
	"Pos Sameday": "Sameday",
	"Pos Reguler": "Reguler",
}

// ClassifyService determines the category and visibility of a shipping service
func ClassifyService(serviceCode string, courierCode string, weight int) ServiceClassification {
	serviceUpper := strings.ToUpper(serviceCode)
	
	// 1. Check if it's a cargo/trucking service - ALWAYS HIDE
	if CargoServiceCodes[serviceCode] || CargoServiceCodes[serviceUpper] {
		return ServiceClassification{
			Category:    dto.ShippingCategoryCargo,
			DisplayName: serviceCode,
			Show:        false,
			Reason:      "Cargo/trucking service not suitable for fashion e-commerce",
		}
	}
	
	// 2. Check for cargo keywords in service code
	cargoKeywords := []string{"TRUCK", "CARGO", "KARGO", "FTL", "LTL", "JUMBO", "T15", "T25", "T60", "TRC", "TRX", "SRP", "JTR"}
	for _, keyword := range cargoKeywords {
		if strings.Contains(serviceUpper, keyword) {
			return ServiceClassification{
				Category:    dto.ShippingCategoryCargo,
				DisplayName: serviceCode,
				Show:        false,
				Reason:      "Contains cargo keyword: " + keyword,
			}
		}
	}
	
	// 3. Weight-based filtering: if weight < 10kg, hide cargo services
	// For fashion e-commerce, typical parcel is 0.5-3kg
	if weight < 10000 { // 10kg in grams
		// Additional check for high-weight services
		if strings.Contains(strings.ToLower(serviceCode), "motor") ||
			strings.Contains(strings.ToLower(serviceCode), "bawah") {
			return ServiceClassification{
				Category:    dto.ShippingCategoryCargo,
				DisplayName: serviceCode,
				Show:        false,
				Reason:      "Motor/vehicle shipping service",
			}
		}
	}
	
	// 4. Classify as Express
	if ExpressServiceCodes[serviceCode] || ExpressServiceCodes[serviceUpper] {
		return ServiceClassification{
			Category:    dto.ShippingCategoryExpress,
			DisplayName: GetServiceDisplayName(serviceCode),
			Show:        true,
			Reason:      "",
		}
	}
	
	// 5. Classify as Economy
	if EconomyServiceCodes[serviceCode] || EconomyServiceCodes[serviceUpper] {
		return ServiceClassification{
			Category:    dto.ShippingCategoryEconomy,
			DisplayName: GetServiceDisplayName(serviceCode),
			Show:        true,
			Reason:      "",
		}
	}
	
	// 6. Default to Regular for unknown services
	return ServiceClassification{
		Category:    dto.ShippingCategoryRegular,
		DisplayName: GetServiceDisplayName(serviceCode),
		Show:        true,
		Reason:      "",
	}
}

// GetServiceDisplayName returns a friendly display name for a service
func GetServiceDisplayName(serviceCode string) string {
	if name, ok := ServiceDisplayNames[serviceCode]; ok {
		return name
	}
	return serviceCode
}

// GetSimplifiedCourierName returns a clean courier name
func GetSimplifiedCourierName(courierCode string) string {
	code := strings.ToLower(courierCode)
	if name, ok := SimplifiedCourierNames[code]; ok {
		return name
	}
	return strings.ToUpper(courierCode)
}

// FilterShippingRates filters and classifies shipping rates for fashion e-commerce
// Returns only services suitable for fashion parcels (typically < 10kg)
func FilterShippingRates(rates []ShippingRate, weight int) []ShippingRate {
	var filtered []ShippingRate
	
	// Track services per courier to limit options
	courierServiceCount := make(map[string]int)
	maxServicesPerCourier := 3 // Show max 3 services per courier
	
	for _, rate := range rates {
		classification := ClassifyService(rate.ServiceCode, rate.ProviderCode, weight)
		
		// Skip cargo/trucking services
		if !classification.Show {
			continue
		}
		
		// Limit services per courier
		courierCode := strings.ToLower(rate.ProviderCode)
		if courierServiceCount[courierCode] >= maxServicesPerCourier {
			continue
		}
		
		// Add category to rate
		rate.Category = classification.Category
		rate.ProviderName = GetSimplifiedCourierName(rate.ProviderCode)
		
		filtered = append(filtered, rate)
		courierServiceCount[courierCode]++
	}
	
	return filtered
}

// FilterShippingRatesDTO filters and classifies shipping rates for DTO response
func FilterShippingRatesDTO(rates []dto.ShippingRateResponse, weight int) []dto.ShippingRateResponse {
	var filtered []dto.ShippingRateResponse
	
	// Track services per courier to limit options
	courierServiceCount := make(map[string]int)
	maxServicesPerCourier := 3 // Show max 3 services per courier
	
	for _, rate := range rates {
		classification := ClassifyService(rate.ServiceCode, rate.ProviderCode, weight)
		
		// Skip cargo/trucking services
		if !classification.Show {
			continue
		}
		
		// Limit services per courier
		courierCode := strings.ToLower(rate.ProviderCode)
		if courierServiceCount[courierCode] >= maxServicesPerCourier {
			continue
		}
		
		// Update rate with classification
		rate.ShippingCategory = classification.Category
		rate.ProviderName = GetSimplifiedCourierName(rate.ProviderCode)
		rate.ServiceName = classification.DisplayName
		
		filtered = append(filtered, rate)
		courierServiceCount[courierCode]++
	}
	
	return filtered
}

// ============================================
// ENHANCED FILTERING FOR TOKOPEDIA/SHOPEE STYLE
// ============================================

// AllowedServicesForFashion - Only these services should appear for weight <= 5kg
// Note: REG is used by multiple couriers, so we just need one entry
var AllowedServicesForFashion = map[string]bool{
	// Common services across couriers
	"REG":  true, // Regular (JNE, SiCepat, TIKI, AnterAja)
	"YES":  true, // JNE Yakin Esok Sampai
	"OKE":  true, // JNE Ongkos Kirim Ekonomis
	"EZ":   true, // J&T Express
	"BEST": true, // SiCepat Best
	"HALU": true, // SiCepat Same Day
	"SIUNT": true, // SiCepat Untung
	"ECO":  true, // TIKI Economy
	"ONS":  true, // TIKI Over Night Service
	"SDS":  true, // TIKI Same Day Service
	"Pos Reguler": true,
	"Pos Nextday": true,
	"Pos Sameday": true,
}

// CargoKeywords - Keywords that indicate cargo/trucking services
var CargoKeywords = []string{
	"TRC", "TRUCK", "CARGO", "KARGO", "FREIGHT", "BULK",
	"T15", "T25", "T60", "JTR", "FTL", "LTL", "JUMBO",
	"MOTOR", "PALLET", "HEAVY",
}

// IsCargoService checks if a service is cargo/trucking based on code and name
func IsCargoService(serviceCode, serviceName string, weight int) bool {
	codeUpper := strings.ToUpper(serviceCode)
	nameUpper := strings.ToUpper(serviceName)
	
	// Check explicit cargo codes
	if CargoServiceCodes[serviceCode] || CargoServiceCodes[codeUpper] {
		return true
	}
	// Check cargo keywords in code
	for _, keyword := range CargoKeywords {
		if strings.Contains(codeUpper, keyword) {
			return true
		}
		if strings.Contains(nameUpper, keyword) {
			return true
		}
	}
	
	// For weight <= 5kg, be more strict
	if weight <= 5000 {
		// Check if code starts with "T" (TIKI trucking pattern)
		if strings.HasPrefix(codeUpper, "T") && len(codeUpper) <= 3 {
			// T15, T25, T60, TRC, TRX are trucking
			if codeUpper != "TIK" && codeUpper != "TKI" { // Not TIKI abbreviation
				return true
			}
		}
	}
	
	return false
}

// GetCategoryPriority returns sort priority for a category
// Lower number = show first
func GetCategoryPriority(category dto.ShippingCategory) int {
	switch category {
	case dto.ShippingCategoryRegular:
		return 1
	case dto.ShippingCategoryEconomy:
		return 2
	case dto.ShippingCategoryExpress:
		return 3
	case "SameDay":
		return 4
	case dto.ShippingCategoryCargo:
		return 99 // Should be hidden anyway
	default:
		return 5
	}
}

// ConvertETDToDate converts ETD string (e.g., "2-3") to actual date string
// Returns format: "Tiba 12 - 13 Jan"
func ConvertETDToDate(etd string) string {
	now := time.Now()
	
	// Parse ETD - can be "2", "2-3", "1-2 hari", etc.
	etd = strings.TrimSpace(etd)
	etd = strings.ReplaceAll(etd, "hari", "")
	etd = strings.ReplaceAll(etd, "HARI", "")
	etd = strings.ReplaceAll(etd, " ", "")
	
	// Handle "0" as same day
	if etd == "0" {
		return "Tiba hari ini"
	}
	
	var minDays, maxDays int
	
	if strings.Contains(etd, "-") {
		parts := strings.Split(etd, "-")
		if len(parts) >= 2 {
			minDays, _ = strconv.Atoi(parts[0])
			maxDays, _ = strconv.Atoi(parts[1])
		}
	} else {
		days, _ := strconv.Atoi(etd)
		minDays = days
		maxDays = days
	}
	
	// Default if parsing failed
	if minDays == 0 && maxDays == 0 {
		return etd // Return original if can't parse
	}
	
	// Calculate dates
	minDate := now.AddDate(0, 0, minDays)
	maxDate := now.AddDate(0, 0, maxDays)
	
	// Format dates in Indonesian
	months := []string{"", "Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}
	
	if minDays == maxDays {
		return fmt.Sprintf("Tiba %d %s", minDate.Day(), months[minDate.Month()])
	}
	
	// Check if same month
	if minDate.Month() == maxDate.Month() {
		return fmt.Sprintf("Tiba %d - %d %s", minDate.Day(), maxDate.Day(), months[minDate.Month()])
	}
	
	return fmt.Sprintf("Tiba %d %s - %d %s", minDate.Day(), months[minDate.Month()], maxDate.Day(), months[maxDate.Month()])
}

// GetETDDays extracts the minimum days from ETD string for sorting
func GetETDDays(etd string) int {
	etd = strings.TrimSpace(etd)
	etd = strings.ReplaceAll(etd, "hari", "")
	etd = strings.ReplaceAll(etd, "HARI", "")
	etd = strings.ReplaceAll(etd, " ", "")
	
	if etd == "0" {
		return 0
	}
	
	if strings.Contains(etd, "-") {
		parts := strings.Split(etd, "-")
		if len(parts) >= 1 {
			days, _ := strconv.Atoi(parts[0])
			return days
		}
	}
	
	days, _ := strconv.Atoi(etd)
	return days
}

// FilteredShippingRate extends ShippingRate with additional display info
type FilteredShippingRate struct {
	dto.ShippingRateResponse
	ETADate       string  `json:"eta_date"`        // "Tiba 12 - 13 Jan"
	ETDDays       int     `json:"etd_days"`        // For sorting
	IsAbsurdPrice bool    `json:"is_absurd_price"` // Price > 5x REG price
	OriginCity    string  `json:"origin_city"`     // "Semarang"
	TotalWeight   int     `json:"total_weight"`    // Weight in grams
}

// EnhancedShippingResponse is the new response format like Tokopedia
type EnhancedShippingResponse struct {
	CartSubtotal    float64                           `json:"cart_subtotal"`
	TotalWeight     int                               `json:"total_weight"`
	TotalWeightKg   string                            `json:"total_weight_kg"`   // "1.2 kg"
	OriginCity      string                            `json:"origin_city"`       // "Semarang"
	DestinationCity string                            `json:"destination_city"`  // From address
	GroupedRates    map[string][]FilteredShippingRate `json:"grouped_rates"`     // Grouped by category
	Rates           []FilteredShippingRate            `json:"rates"`             // Flat list, sorted
	RegularMinPrice float64                           `json:"regular_min_price"` // For absurd price check
}

// FilterAndEnhanceRates applies all filtering rules and enhances rates for display
func FilterAndEnhanceRates(rates []ShippingRate, weight int, originCity, destCity string) *EnhancedShippingResponse {
	var filtered []FilteredShippingRate
	var regularMinPrice float64 = -1
	
	// First pass: filter and find regular min price
	for _, rate := range rates {
		// Skip cargo services
		if IsCargoService(rate.ServiceCode, rate.ServiceName, weight) {
			fmt.Printf("🚫 Filtered out cargo service: %s %s\n", rate.ProviderCode, rate.ServiceCode)
			continue
		}
		
		classification := ClassifyService(rate.ServiceCode, rate.ProviderCode, weight)
		if !classification.Show {
			continue
		}
		
		// Track regular min price for absurd price check
		if classification.Category == dto.ShippingCategoryRegular {
			if regularMinPrice < 0 || rate.Cost < regularMinPrice {
				regularMinPrice = rate.Cost
			}
		}
		
		enhanced := FilteredShippingRate{
			ShippingRateResponse: dto.ShippingRateResponse{
				ProviderCode:     rate.ProviderCode,
				ProviderName:     GetSimplifiedCourierName(rate.ProviderCode),
				ServiceCode:      rate.ServiceCode,
				ServiceName:      classification.DisplayName,
				Description:      rate.Description,
				Cost:             rate.Cost,
				ETD:              rate.ETD,
				ShippingCategory: classification.Category,
			},
			ETADate:     ConvertETDToDate(rate.ETD),
			ETDDays:     GetETDDays(rate.ETD),
			OriginCity:  originCity,
			TotalWeight: weight,
		}
		
		filtered = append(filtered, enhanced)
	}
	
	// Second pass: mark absurd prices (> 5x regular min price)
	if regularMinPrice > 0 {
		absurdThreshold := regularMinPrice * 5
		for i := range filtered {
			if filtered[i].Cost > absurdThreshold {
				filtered[i].IsAbsurdPrice = true
				fmt.Printf("💰 Marked absurd price: %s %s = Rp %.0f (threshold: Rp %.0f)\n",
					filtered[i].ProviderCode, filtered[i].ServiceCode, filtered[i].Cost, absurdThreshold)
			}
		}
	}
	
	// Sort by: Category priority, then by price within category
	sort.Slice(filtered, func(i, j int) bool {
		priI := GetCategoryPriority(filtered[i].ShippingCategory)
		priJ := GetCategoryPriority(filtered[j].ShippingCategory)
		
		if priI != priJ {
			return priI < priJ
		}
		
		// Within same category, sort by price
		return filtered[i].Cost < filtered[j].Cost
	})
	
	// Group by category
	grouped := make(map[string][]FilteredShippingRate)
	for _, rate := range filtered {
		// Skip absurd prices from main list (can be shown with "Tampilkan layanan kargo" button)
		if rate.IsAbsurdPrice {
			continue
		}
		
		category := string(rate.ShippingCategory)
		grouped[category] = append(grouped[category], rate)
	}
	
	// Build final list (excluding absurd prices)
	var finalRates []FilteredShippingRate
	for _, rate := range filtered {
		if !rate.IsAbsurdPrice {
			finalRates = append(finalRates, rate)
		}
	}
	
	// Format weight
	weightKg := fmt.Sprintf("%.1f kg", float64(weight)/1000)
	if weight < 1000 {
		weightKg = fmt.Sprintf("%d g", weight)
	}
	
	return &EnhancedShippingResponse{
		TotalWeight:     weight,
		TotalWeightKg:   weightKg,
		OriginCity:      originCity,
		DestinationCity: destCity,
		GroupedRates:    grouped,
		Rates:           finalRates,
		RegularMinPrice: regularMinPrice,
	}
}

// ClassifyServiceEnhanced is an enhanced version with Same Day support
func ClassifyServiceEnhanced(serviceCode string, courierCode string, weight int) ServiceClassification {
	serviceUpper := strings.ToUpper(serviceCode)
	
	// 1. Check if it's a cargo/trucking service - ALWAYS HIDE for weight <= 5kg
	if weight <= 5000 && IsCargoService(serviceCode, "", weight) {
		return ServiceClassification{
			Category:    dto.ShippingCategoryCargo,
			DisplayName: serviceCode,
			Show:        false,
			Reason:      "Cargo service hidden for fashion e-commerce (weight <= 5kg)",
			Priority:    99,
		}
	}
	
	// 2. Check for Same Day services
	if SameDayServiceCodes[serviceCode] || SameDayServiceCodes[serviceUpper] {
		return ServiceClassification{
			Category:    "SameDay",
			DisplayName: GetServiceDisplayName(serviceCode),
			Show:        true,
			Reason:      "",
			Priority:    4,
		}
	}
	
	// 3. Check for Express services
	if ExpressServiceCodes[serviceCode] || ExpressServiceCodes[serviceUpper] {
		return ServiceClassification{
			Category:    dto.ShippingCategoryExpress,
			DisplayName: GetServiceDisplayName(serviceCode),
			Show:        true,
			Reason:      "",
			Priority:    3,
		}
	}
	
	// 4. Check for Economy services
	if EconomyServiceCodes[serviceCode] || EconomyServiceCodes[serviceUpper] {
		return ServiceClassification{
			Category:    dto.ShippingCategoryEconomy,
			DisplayName: GetServiceDisplayName(serviceCode),
			Show:        true,
			Reason:      "",
			Priority:    2,
		}
	}
	
	// 5. Default to Regular
	return ServiceClassification{
		Category:    dto.ShippingCategoryRegular,
		DisplayName: GetServiceDisplayName(serviceCode),
		Show:        true,
		Reason:      "",
		Priority:    1,
	}
}
