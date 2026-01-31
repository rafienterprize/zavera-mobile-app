package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// ============================================
// BITESHIP ERROR TYPES
// ============================================

var (
	ErrBiteshipUnauthorized   = errors.New("biteship: unauthorized - check TOKEN_BITESHIP")
	ErrBiteshipRateLimited    = errors.New("biteship: rate limited - retry later")
	ErrBiteshipAPIFailed      = errors.New("biteship: API request failed")
	ErrBiteshipInvalidRequest = errors.New("biteship: invalid request parameters")
	ErrBiteshipAreaNotFound   = errors.New("biteship: area not found")
	ErrBiteshipNoRates        = errors.New("biteship: no shipping rates available")
	ErrBiteshipDraftFailed    = errors.New("biteship: failed to create draft order")
	ErrBiteshipConfirmFailed  = errors.New("biteship: failed to confirm order")
	ErrBiteshipTrackingNotFound = errors.New("biteship: tracking not found")
)

// ============================================
// BITESHIP API RESPONSE TYPES
// ============================================

// BiteshipArea represents an area from /v1/maps/areas
type BiteshipArea struct {
	ID                       string      `json:"id"`
	Name                     string      `json:"name"`
	CountryName              string      `json:"country_name"`
	CountryCode              string      `json:"country_code"`
	AdministrativeLevel1Name string      `json:"administrative_division_level_1_name"`
	AdministrativeLevel2Name string      `json:"administrative_division_level_2_name"`
	AdministrativeLevel3Name string      `json:"administrative_division_level_3_name"`
	AdministrativeLevel4Name string      `json:"administrative_division_level_4_name"`
	PostalCode               interface{} `json:"postal_code"` // Can be string or int
}

// BiteshipAreasResponse represents the response from /v1/maps/areas
type BiteshipAreasResponse struct {
	Success bool           `json:"success"`
	Areas   []BiteshipArea `json:"areas"`
}

// BiteshipCourier represents a courier from /v1/couriers
type BiteshipCourier struct {
	AvailableForCashOnDelivery   bool   `json:"available_for_cash_on_delivery"`
	AvailableForProofOfDelivery  bool   `json:"available_for_proof_of_delivery"`
	AvailableForInstantWaybillID bool   `json:"available_for_instant_waybill_id"`
	CourierCode                  string `json:"courier_code"`
	CourierName                  string `json:"courier_name"`
	CourierServiceCode           string `json:"courier_service_code"`
	CourierServiceName           string `json:"courier_service_name"`
	Tier                         string `json:"tier"`
	Description                  string `json:"description"`
	ServiceType                  string `json:"service_type"`
	ShippingType                 string `json:"shipping_type"`
	Duration                     string `json:"duration"`
}

// BiteshipCouriersResponse represents the response from /v1/couriers
type BiteshipCouriersResponse struct {
	Success  bool              `json:"success"`
	Couriers []BiteshipCourier `json:"couriers"`
}

// BiteshipRate represents a shipping rate from /v1/rates/couriers
type BiteshipRate struct {
	CourierCode           string  `json:"courier_code"`
	CourierName           string  `json:"courier_name"`
	CourierServiceCode    string  `json:"courier_service_code"`
	CourierServiceName    string  `json:"courier_service_name"`
	Description           string  `json:"description"`
	Duration              string  `json:"duration"`
	ShipmentDurationRange string  `json:"shipment_duration_range"`
	ShipmentDurationUnit  string  `json:"shipment_duration_unit"`
	ServiceType           string  `json:"service_type"`
	ShippingType          string  `json:"shipping_type"`
	Price                 float64 `json:"price"`
	Type                  string  `json:"type"`
}

// BiteshipRatesResponse represents the response from /v1/rates/couriers
type BiteshipRatesResponse struct {
	Success  bool           `json:"success"`
	Origin   BiteshipArea   `json:"origin"`
	Destination BiteshipArea `json:"destination"`
	Pricing  []BiteshipRate `json:"pricing"`
}

// BiteshipLocation represents a saved location
type BiteshipLocationResponse struct {
	Success bool   `json:"success"`
	ID      string `json:"id"`
}

// BiteshipDraftOrder represents a draft order response
type BiteshipDraftOrder struct {
	Success bool   `json:"success"`
	ID      string `json:"id"`
	Status  string `json:"status"`
}

// BiteshipDraftOrderDetail represents detailed draft order information
type BiteshipDraftOrderDetail struct {
	Success     bool   `json:"success"`
	ID          string `json:"id"`
	Status      string `json:"status"`
	OrderID     string `json:"order_id"`
	WaybillID   string `json:"waybill_id"`
	TrackingID  string `json:"tracking_id"`
	CourierCode string `json:"courier_code"`
	Courier     struct {
		WaybillID  string `json:"waybill_id"`
		TrackingID string `json:"tracking_id"`
	} `json:"courier"`
}

// BiteshipOrder represents a confirmed order response
type BiteshipOrder struct {
	Success    bool   `json:"success"`
	ID         string `json:"id"`
	Status     string `json:"status"`
	WaybillID  string `json:"waybill_id"`
	TrackingID string `json:"tracking_id"`
	CourierCode string `json:"courier_code"`
}

// BiteshipTrackingHistory represents a tracking history entry
type BiteshipTrackingHistory struct {
	Note      string `json:"note"`
	Status    string `json:"status"`
	UpdatedAt string `json:"updated_at"`
}

// BiteshipTracking represents tracking information
type BiteshipTracking struct {
	Success            bool                      `json:"success"`
	ID                 string                    `json:"id"`
	WaybillID          string                    `json:"waybill_id"`
	CourierCode        string                    `json:"courier_code"`
	CourierName        string                    `json:"courier_name"`
	Status             string                    `json:"status"`
	OriginAddress      string                    `json:"origin_address"`
	DestinationAddress string                    `json:"destination_address"`
	History            []BiteshipTrackingHistory `json:"history"`
}

// BiteshipAPIError represents an API error response
type BiteshipAPIError struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    int    `json:"code"`
}


// ============================================
// BITESHIP CLIENT
// ============================================

// RetryConfig defines retry behavior for API calls
type RetryConfig struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	BackoffFactor  float64
}

// DefaultRetryConfig returns the default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     30 * time.Second,
		BackoffFactor:  2.0,
	}
}

// BiteshipClient handles communication with Biteship API
type BiteshipClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
	retry      RetryConfig
}

// NewBiteshipClient creates a new Biteship API client
func NewBiteshipClient() *BiteshipClient {
	baseURL := strings.TrimRight(os.Getenv("BITESHIP_BASE_URL"), "/")
	if baseURL == "" {
		baseURL = "https://api.biteship.com"
	}

	token := strings.TrimSpace(os.Getenv("TOKEN_BITESHIP"))

	return &BiteshipClient{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		retry: DefaultRetryConfig(),
	}
}

// doRequest performs an HTTP request with retry logic
func (c *BiteshipClient) doRequest(method, endpoint string, body interface{}) ([]byte, error) {
	if c.token == "" {
		return nil, ErrBiteshipUnauthorized
	}

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to marshal request body", ErrBiteshipInvalidRequest)
		}
		log.Printf("📤 Biteship API Request [%s %s]: %s", method, endpoint, string(jsonBody))
		reqBody = bytes.NewBuffer(jsonBody)
	}

	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)
	
	var lastErr error
	backoff := c.retry.InitialBackoff

	for attempt := 0; attempt <= c.retry.MaxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("🔄 Biteship API retry attempt %d/%d after %v", attempt, c.retry.MaxRetries, backoff)
			time.Sleep(backoff)
			backoff = time.Duration(math.Min(float64(backoff)*c.retry.BackoffFactor, float64(c.retry.MaxBackoff)))
			
			// Reset body reader for retry
			if body != nil {
				jsonBody, _ := json.Marshal(body)
				reqBody = bytes.NewBuffer(jsonBody)
			}
		}

		req, err := http.NewRequest(method, url, reqBody)
		if err != nil {
			lastErr = fmt.Errorf("%w: %v", ErrBiteshipAPIFailed, err)
			continue
		}

		// Set headers - Authorization with Bearer token
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("%w: %v", ErrBiteshipAPIFailed, err)
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("%w: failed to read response", ErrBiteshipAPIFailed)
			continue
		}

		log.Printf("📡 Biteship API [%s %s] Status: %d", method, endpoint, resp.StatusCode)

		// Handle specific status codes
		switch resp.StatusCode {
		case 200, 201:
			return respBody, nil
		case 401:
			return nil, ErrBiteshipUnauthorized
		case 429:
			lastErr = ErrBiteshipRateLimited
			continue // Retry on rate limit
		case 400:
			var apiErr BiteshipAPIError
			json.Unmarshal(respBody, &apiErr)
			log.Printf("❌ Biteship API 400 Error: %s (Response: %s)", apiErr.Error, string(respBody))
			return nil, fmt.Errorf("%w: %s", ErrBiteshipInvalidRequest, apiErr.Error)
		case 404:
			return nil, fmt.Errorf("%w: resource not found", ErrBiteshipAPIFailed)
		default:
			if resp.StatusCode >= 500 {
				lastErr = fmt.Errorf("%w: server error %d", ErrBiteshipAPIFailed, resp.StatusCode)
				continue // Retry on 5xx errors
			}
			return nil, fmt.Errorf("%w: status %d - %s", ErrBiteshipAPIFailed, resp.StatusCode, string(respBody))
		}
	}

	return nil, lastErr
}

// doGet performs a GET request
func (c *BiteshipClient) doGet(endpoint string) ([]byte, error) {
	return c.doRequest("GET", endpoint, nil)
}

// doPost performs a POST request
func (c *BiteshipClient) doPost(endpoint string, body interface{}) ([]byte, error) {
	return c.doRequest("POST", endpoint, body)
}


// ============================================
// BITESHIP API METHODS
// ============================================

// SearchAreas searches for areas using Biteship API
// GET /v1/maps/areas?countries=ID&input={query}
func (c *BiteshipClient) SearchAreas(query string) ([]BiteshipArea, error) {
	if query == "" {
		return nil, fmt.Errorf("%w: search query is required", ErrBiteshipInvalidRequest)
	}

	endpoint := fmt.Sprintf("/v1/maps/areas?countries=ID&input=%s", url.QueryEscape(query))
	
	respBody, err := c.doGet(endpoint)
	if err != nil {
		return nil, err
	}

	var response BiteshipAreasResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("%w: failed to parse areas response", ErrBiteshipAPIFailed)
	}

	if !response.Success {
		return nil, ErrBiteshipAreaNotFound
	}

	log.Printf("✅ Found %d areas for query: %s", len(response.Areas), query)
	return response.Areas, nil
}

// SearchAreasWithSuggestions searches areas with Biteship API
// IMPROVED: Only returns results that CONTAIN the search query (like Google Maps/Tokopedia)
// If user types "bubut", only show areas with "bubut" in the name, NOT fuzzy matches like "Bunut"
func (c *BiteshipClient) SearchAreasWithSuggestions(query string) ([]BiteshipArea, error) {
	if query == "" || len(query) < 3 {
		return nil, fmt.Errorf("%w: search query must be at least 3 characters", ErrBiteshipInvalidRequest)
	}

	query = strings.TrimSpace(query)
	queryLower := strings.ToLower(query)
	
	// Build search queries - original + completions
	searchQueries := []string{query}
	
	// Add common completions for partial words
	completions := getQueryCompletions(query)
	searchQueries = append(searchQueries, completions...)
	
	// Remove duplicates
	seen := make(map[string]bool)
	uniqueQueries := []string{}
	for _, q := range searchQueries {
		qLower := strings.ToLower(q)
		if !seen[qLower] && q != "" {
			seen[qLower] = true
			uniqueQueries = append(uniqueQueries, q)
		}
	}
	
	// Query Biteship API and collect all results
	var allAreas []BiteshipArea
	seenAreas := make(map[string]bool) // Use name+postal as key to avoid duplicates
	
	for i, searchQuery := range uniqueQueries {
		if i >= 5 { // Limit to 5 queries for performance
			break
		}
		
		areas, err := c.SearchAreas(searchQuery)
		if err != nil {
			continue
		}
		
		for _, area := range areas {
			// Use name as unique key (includes postal code)
			key := area.Name
			if !seenAreas[key] {
				seenAreas[key] = true
				allAreas = append(allAreas, area)
			}
		}
	}

	// CRITICAL: Filter results to only include areas that CONTAIN the search query
	// This is how Google Maps, Tokopedia, Gojek work - exact substring matching
	filteredAreas := filterAreasBySubstring(allAreas, queryLower)
	
	// If we have exact matches, use them
	if len(filteredAreas) > 0 {
		// Sort by relevance
		sortAreasByRelevance(filteredAreas, queryLower)
		log.Printf("✅ Found %d EXACT matching areas for: %s", len(filteredAreas), query)
		return filteredAreas, nil
	}
	
	// If no exact matches, return all results but sorted by relevance
	// This handles cases where user might have typo or searching for something unusual
	sortAreasByRelevance(allAreas, queryLower)
	log.Printf("⚠️ No exact matches for '%s', returning %d fuzzy results", query, len(allAreas))
	return allAreas, nil
}

// filterAreasBySubstring filters areas to only include those containing the query
// This ensures "bubut" only shows areas with "bubut" in the name, not "Bunut"
func filterAreasBySubstring(areas []BiteshipArea, query string) []BiteshipArea {
	var filtered []BiteshipArea
	
	for _, area := range areas {
		// Check if query appears in any part of the area name
		nameLower := strings.ToLower(area.Name)
		districtLower := strings.ToLower(area.AdministrativeLevel3Name)
		subdistrictLower := strings.ToLower(area.AdministrativeLevel4Name)
		cityLower := strings.ToLower(area.AdministrativeLevel2Name)
		provinceLower := strings.ToLower(area.AdministrativeLevel1Name)
		
		// Check if query is contained in any field
		if strings.Contains(nameLower, query) ||
		   strings.Contains(districtLower, query) ||
		   strings.Contains(subdistrictLower, query) ||
		   strings.Contains(cityLower, query) ||
		   strings.Contains(provinceLower, query) {
			filtered = append(filtered, area)
		}
	}
	
	return filtered
}

// sortAreasByRelevance sorts areas so most relevant results appear first
// Priority: exact match > starts with > contains
func sortAreasByRelevance(areas []BiteshipArea, query string) {
	query = strings.ToLower(query)
	
	// Sort using bubble sort (simple and works for small arrays)
	for i := 0; i < len(areas); i++ {
		for j := i + 1; j < len(areas); j++ {
			scoreI := getRelevanceScore(areas[i], query)
			scoreJ := getRelevanceScore(areas[j], query)
			if scoreJ > scoreI {
				areas[i], areas[j] = areas[j], areas[i]
			}
		}
	}
}

// getRelevanceScore returns a score for how relevant an area is to the query
// Higher score = more relevant = appears first in results
func getRelevanceScore(area BiteshipArea, query string) int {
	nameLower := strings.ToLower(area.Name)
	subdistrictLower := strings.ToLower(area.AdministrativeLevel4Name) // Kelurahan
	districtLower := strings.ToLower(area.AdministrativeLevel3Name)    // Kecamatan
	cityLower := strings.ToLower(area.AdministrativeLevel2Name)        // Kota/Kabupaten
	provinceLower := strings.ToLower(area.AdministrativeLevel1Name)    // Provinsi
	
	// TIER 1: Exact match (highest priority)
	if strings.EqualFold(area.AdministrativeLevel4Name, query) { // Kelurahan exact
		return 200
	}
	if strings.EqualFold(area.AdministrativeLevel3Name, query) { // Kecamatan exact
		return 195
	}
	if strings.EqualFold(area.AdministrativeLevel2Name, query) { // Kota exact
		return 190
	}
	
	// TIER 2: Starts with query (very high priority)
	if strings.HasPrefix(subdistrictLower, query) { // Kelurahan starts with
		return 150
	}
	if strings.HasPrefix(districtLower, query) { // Kecamatan starts with
		return 145
	}
	if strings.HasPrefix(cityLower, query) { // Kota starts with
		return 140
	}
	
	// TIER 3: Contains query at word boundary
	// e.g., "bubut" in "Jalan Bubut" should rank higher than "bubutan"
	words := strings.Fields(nameLower)
	for _, word := range words {
		if strings.HasPrefix(word, query) {
			return 120
		}
	}
	
	// TIER 4: Contains query anywhere
	if strings.Contains(subdistrictLower, query) {
		return 100
	}
	if strings.Contains(districtLower, query) {
		return 95
	}
	if strings.Contains(cityLower, query) {
		return 90
	}
	if strings.Contains(nameLower, query) {
		return 80
	}
	if strings.Contains(provinceLower, query) {
		return 70
	}
	
	// TIER 5: No direct match - lowest priority (fuzzy results from Biteship)
	return 10
}

// getQueryCompletions returns possible completions for partial queries
func getQueryCompletions(query string) []string {
	query = strings.ToLower(strings.TrimSpace(query))
	var completions []string
	
	// Common Indonesian city/district name completions
	prefixMap := map[string][]string{
		// Major cities
		"jak":      {"jakarta", "jakarta pusat", "jakarta selatan", "jakarta barat", "jakarta timur", "jakarta utara"},
		"jaka":     {"jakarta", "jakarta pusat", "jakarta selatan", "jakarta barat", "jakarta timur", "jakarta utara"},
		"jakar":    {"jakarta"},
		"jakart":   {"jakarta"},
		"band":     {"bandung"},
		"bandu":    {"bandung"},
		"bandun":   {"bandung"},
		"sura":     {"surabaya", "surakarta"},
		"surab":    {"surabaya"},
		"suraba":   {"surabaya"},
		"surabay":  {"surabaya"},
		"sema":     {"semarang"},
		"semar":    {"semarang"},
		"semara":   {"semarang"},
		"semaran":  {"semarang"},
		"yogy":     {"yogyakarta"},
		"yogya":    {"yogyakarta"},
		"yogyak":   {"yogyakarta"},
		"meda":     {"medan"},
		"medan":    {"medan"},
		"maka":     {"makassar"},
		"makas":    {"makassar"},
		"makass":   {"makassar"},
		"makasar":  {"makassar"},
		"makassa":  {"makassar"},
		"pale":     {"palembang"},
		"palem":    {"palembang"},
		"palemb":   {"palembang"},
		"tang":     {"tangerang"},
		"tange":    {"tangerang"},
		"tanger":   {"tangerang"},
		"tangera":  {"tangerang"},
		"tangeran": {"tangerang"},
		"depo":     {"depok"},
		"depok":    {"depok"},
		"beka":     {"bekasi"},
		"bekas":    {"bekasi"},
		"bekasi":   {"bekasi"},
		"bogo":     {"bogor"},
		"bogor":    {"bogor"},
		"mala":     {"malang"},
		"malan":    {"malang"},
		"malang":   {"malang"},
		"solo":     {"surakarta", "solo"},
		"bali":     {"bali", "denpasar"},
		"denp":     {"denpasar"},
		"denpa":    {"denpasar"},
		"denpas":   {"denpasar"},
		"denpasa":  {"denpasar"},
		"bata":     {"batam"},
		"batam":    {"batam"},
		"peka":     {"pekanbaru", "pekalongan"},
		"pekan":    {"pekanbaru", "pekalongan"},
		"pont":     {"pontianak"},
		"ponti":    {"pontianak"},
		"pontia":   {"pontianak"},
		"pontian":  {"pontianak"},
		"banj":     {"banjarmasin"},
		"banja":    {"banjarmasin"},
		"banjar":   {"banjarmasin"},
		"banjarm":  {"banjarmasin"},
		"sama":     {"samarinda"},
		"samar":    {"samarinda"},
		"samari":   {"samarinda"},
		"samarin":  {"samarinda"},
		"mana":     {"manado"},
		"manad":    {"manado"},
		"manado":   {"manado"},
		"pada":     {"padang"},
		"padan":    {"padang"},
		"padang":   {"padang"},
		"lamp":     {"lampung", "bandar lampung"},
		"lampu":    {"lampung", "bandar lampung"},
		"lampun":   {"lampung", "bandar lampung"},
		"aceh":     {"banda aceh", "aceh"},
		"kend":     {"kendari"},
		"kenda":    {"kendari"},
		"kendar":   {"kendari"},
		"kendari":  {"kendari"},
		"palu":     {"palu"},
		"ambo":     {"ambon"},
		"ambon":    {"ambon"},
		"jaya":     {"jayapura"},
		"jayap":    {"jayapura"},
		"jayapu":   {"jayapura"},
		"jayapur":  {"jayapura"},
		"kupa":     {"kupang"},
		"kupan":    {"kupang"},
		"kupang":   {"kupang"},
		"mata":     {"mataram"},
		"matar":    {"mataram"},
		"matara":   {"mataram"},
		"mataram":  {"mataram"},
		"cire":     {"cirebon"},
		"cireb":    {"cirebon"},
		"cirebo":   {"cirebon"},
		"cirebon":  {"cirebon"},
		"tasi":     {"tasikmalaya"},
		"tasik":    {"tasikmalaya"},
		"tasikm":   {"tasikmalaya"},
		"suka":     {"sukabumi"},
		"sukab":    {"sukabumi"},
		"sukabu":   {"sukabumi"},
		"sukabum":  {"sukabumi"},
		"kara":     {"karawang"},
		"karaw":    {"karawang"},
		"karawa":   {"karawang"},
		"karawan":  {"karawang"},
		"purw":     {"purwokerto", "purwakarta"},
		"purwo":    {"purwokerto"},
		"purwok":   {"purwokerto"},
		"purwoke":  {"purwokerto"},
		"purwoker": {"purwokerto"},
		"tega":     {"tegal"},
		"tegal":    {"tegal"},
		"mage":     {"magelang"},
		"magel":    {"magelang"},
		"magela":   {"magelang"},
		"magelan":  {"magelang"},
		"kudu":     {"kudus"},
		"kudus":    {"kudus"},
		"jepa":     {"jepara"},
		"jepar":    {"jepara"},
		"jepara":   {"jepara"},
		"dema":     {"demak"},
		"demak":    {"demak"},
		"keda":     {"kediri"},
		"kedar":    {"kediri"},
		"kedari":   {"kediri"},
		"kediri":   {"kediri"},
		"mojo":     {"mojokerto"},
		"mojok":    {"mojokerto"},
		"mojoke":   {"mojokerto"},
		"mojoker":  {"mojokerto"},
		"mojokert": {"mojokerto"},
		"pasu":     {"pasuruan"},
		"pasur":    {"pasuruan"},
		"pasuru":   {"pasuruan"},
		"pasurua":  {"pasuruan"},
		"prob":     {"probolinggo"},
		"probo":    {"probolinggo"},
		"probol":   {"probolinggo"},
		"proboli":  {"probolinggo"},
		"probolin": {"probolinggo"},
		"jemb":     {"jember"},
		"jembe":    {"jember"},
		"jember":   {"jember"},
		"bany":     {"banyuwangi", "banyumas"},
		"banyu":    {"banyuwangi", "banyumas"},
		"banyuw":   {"banyuwangi"},
		"banyuwa":  {"banyuwangi"},
		"banyuwan": {"banyuwangi"},
		"madi":     {"madiun"},
		"madiu":    {"madiun"},
		"madiun":   {"madiun"},
		"gres":     {"gresik"},
		"gresi":    {"gresik"},
		"gresik":   {"gresik"},
		"tuba":     {"tuban"},
		"tuban":    {"tuban"},
		"sera":     {"serang"},
		"seran":    {"serang"},
		"serang":   {"serang"},
		"cile":     {"cilegon"},
		"cileg":    {"cilegon"},
		"cilego":   {"cilegon"},
		"cilegon":  {"cilegon"},
		"cima":     {"cimahi"},
		"cimah":    {"cimahi"},
		"cimahi":   {"cimahi"},
		"garu":     {"garut"},
		"garut":    {"garut"},
		"suba":     {"subang"},
		"suban":    {"subang"},
		"subang":   {"subang"},
		"cian":     {"cianjur"},
		"cianj":    {"cianjur"},
		"cianju":   {"cianjur"},
		"cianjur":  {"cianjur"},
		// Districts
		"pedu":     {"pedurungan"},
		"pedur":    {"pedurungan"},
		"peduru":   {"pedurungan"},
		"pedurun":  {"pedurungan"},
		"pedurung": {"pedurungan"},
		"temb":     {"tembalang"},
		"temba":    {"tembalang"},
		"tembal":   {"tembalang"},
		"tembala":  {"tembalang"},
		"tembalan": {"tembalang"},
		"ngal":     {"ngaliyan"},
		"ngali":    {"ngaliyan"},
		"ngaliy":   {"ngaliyan"},
		"ngaliya":  {"ngaliyan"},
		"mije":     {"mijen"},
		"mijen":    {"mijen"},
		"gunu":     {"gunungpati"},
		"gunung":   {"gunungpati"},
		"gunungp":  {"gunungpati"},
		"gunungpa": {"gunungpati"},
		"ment":     {"menteng"},
		"mente":    {"menteng"},
		"menten":   {"menteng"},
		"menteng":  {"menteng"},
		"gamb":     {"gambir"},
		"gambi":    {"gambir"},
		"gambir":   {"gambir"},
		"keba":     {"kebayoran"},
		"kebay":    {"kebayoran"},
		"kebayo":   {"kebayoran"},
		"kebayor":  {"kebayoran"},
		"tebe":     {"tebet"},
		"tebet":    {"tebet"},
		"cila":     {"cilandak"},
		"cilan":    {"cilandak"},
		"ciland":   {"cilandak"},
		"cilanda":  {"cilandak"},
		"cilandak": {"cilandak"},
	}
	
	if matches, ok := prefixMap[query]; ok {
		completions = append(completions, matches...)
	}
	
	return completions
}

// GetCouriers fetches available couriers
// GET /v1/couriers
func (c *BiteshipClient) GetCouriers() ([]BiteshipCourier, error) {
	endpoint := "/v1/couriers"
	
	respBody, err := c.doGet(endpoint)
	if err != nil {
		return nil, err
	}

	var response BiteshipCouriersResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("%w: failed to parse couriers response", ErrBiteshipAPIFailed)
	}

	if !response.Success {
		return nil, fmt.Errorf("%w: failed to get couriers", ErrBiteshipAPIFailed)
	}

	log.Printf("✅ Found %d couriers", len(response.Couriers))
	return response.Couriers, nil
}

// CreateLocationRequest represents a request to create a location
type CreateLocationRequest struct {
	Name        string `json:"name"`
	ContactName string `json:"contact_name"`
	ContactPhone string `json:"contact_phone"`
	Address     string `json:"address"`
	Note        string `json:"note,omitempty"`
	PostalCode  string `json:"postal_code"`
	Latitude    float64 `json:"latitude,omitempty"`
	Longitude   float64 `json:"longitude,omitempty"`
	AreaID      string `json:"area_id"`
}

// CreateLocation creates a saved location
// POST /v1/locations
func (c *BiteshipClient) CreateLocation(req CreateLocationRequest) (string, error) {
	endpoint := "/v1/locations"
	
	respBody, err := c.doPost(endpoint, req)
	if err != nil {
		return "", err
	}

	var response BiteshipLocationResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return "", fmt.Errorf("%w: failed to parse location response", ErrBiteshipAPIFailed)
	}

	if !response.Success {
		return "", fmt.Errorf("%w: failed to create location", ErrBiteshipAPIFailed)
	}

	log.Printf("✅ Created location with ID: %s", response.ID)
	return response.ID, nil
}

// GetRatesRequest represents a request to get shipping rates
type GetRatesRequest struct {
	OriginAreaID         string                `json:"origin_area_id,omitempty"`
	DestinationAreaID    string                `json:"destination_area_id,omitempty"`
	OriginPostalCode     int                   `json:"origin_postal_code,omitempty"`
	DestinationPostalCode int                  `json:"destination_postal_code,omitempty"`
	Couriers             string                `json:"couriers,omitempty"`
	Items                []GetRatesRequestItem `json:"items"`
}

// GetRatesRequestItem represents an item in the rates request
type GetRatesRequestItem struct {
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Value       float64 `json:"value"`
	Length      int     `json:"length,omitempty"`
	Width       int     `json:"width,omitempty"`
	Height      int     `json:"height,omitempty"`
	Weight      int     `json:"weight"`
	Quantity    int     `json:"quantity"`
}

// GetRates fetches shipping rates
// POST /v1/rates/couriers
func (c *BiteshipClient) GetRates(req GetRatesRequest) ([]BiteshipRate, error) {
	// Validate - need either area_id or postal_code
	hasOrigin := req.OriginAreaID != "" || req.OriginPostalCode > 0
	hasDestination := req.DestinationAreaID != "" || req.DestinationPostalCode > 0
	
	if !hasOrigin || !hasDestination {
		return nil, fmt.Errorf("%w: origin and destination (area_id or postal_code) are required", ErrBiteshipInvalidRequest)
	}

	endpoint := "/v1/rates/couriers"
	
	respBody, err := c.doPost(endpoint, req)
	if err != nil {
		return nil, err
	}

	log.Printf("📥 Biteship Rates Response: %s", string(respBody))

	var response BiteshipRatesResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		log.Printf("❌ Failed to parse rates response: %v", err)
		return nil, fmt.Errorf("%w: failed to parse rates response", ErrBiteshipAPIFailed)
	}

	if !response.Success {
		log.Printf("❌ Biteship rates not successful")
		return nil, ErrBiteshipNoRates
	}

	log.Printf("✅ Found %d shipping rates", len(response.Pricing))
	return response.Pricing, nil
}

// CreateDraftOrderRequest represents a request to create a draft order
type CreateDraftOrderRequest struct {
	ShipperContactName  string                        `json:"shipper_contact_name"`
	ShipperContactPhone string                        `json:"shipper_contact_phone"`
	ShipperContactEmail string                        `json:"shipper_contact_email,omitempty"`
	ShipperOrganization string                        `json:"shipper_organization,omitempty"`
	OriginContactName   string                        `json:"origin_contact_name"`
	OriginContactPhone  string                        `json:"origin_contact_phone"`
	OriginAddress       string                        `json:"origin_address"`
	OriginNote          string                        `json:"origin_note,omitempty"`
	OriginPostalCode    string                        `json:"origin_postal_code"`
	OriginAreaID        string                        `json:"origin_area_id,omitempty"` // Optional - use postal_code if not provided
	DestinationContactName  string                    `json:"destination_contact_name"`
	DestinationContactPhone string                    `json:"destination_contact_phone"`
	DestinationContactEmail string                    `json:"destination_contact_email,omitempty"`
	DestinationAddress      string                    `json:"destination_address"`
	DestinationNote         string                    `json:"destination_note,omitempty"`
	DestinationPostalCode   string                    `json:"destination_postal_code"`
	DestinationAreaID       string                    `json:"destination_area_id,omitempty"` // Optional - use postal_code if not provided
	CourierCode             string                    `json:"courier_code"`
	CourierServiceCode      string                    `json:"courier_service_code"`
	DeliveryType            string                    `json:"delivery_type"`
	DeliveryDate            string                    `json:"delivery_date,omitempty"`
	DeliveryTime            string                    `json:"delivery_time,omitempty"`
	OrderNote               string                    `json:"order_note,omitempty"`
	Items                   []CreateDraftOrderItem    `json:"items"`
}

// CreateDraftOrderItem represents an item in the draft order
type CreateDraftOrderItem struct {
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Value       float64 `json:"value"`
	Length      int     `json:"length,omitempty"`
	Width       int     `json:"width,omitempty"`
	Height      int     `json:"height,omitempty"`
	Weight      int     `json:"weight"`
	Quantity    int     `json:"quantity"`
}

// CreateDraftOrder creates a draft order
// POST /v1/draft_orders
func (c *BiteshipClient) CreateDraftOrder(req CreateDraftOrderRequest) (string, error) {
	endpoint := "/v1/draft_orders"
	
	respBody, err := c.doPost(endpoint, req)
	if err != nil {
		return "", err
	}

	var response BiteshipDraftOrder
	if err := json.Unmarshal(respBody, &response); err != nil {
		return "", fmt.Errorf("%w: failed to parse draft order response", ErrBiteshipAPIFailed)
	}

	if !response.Success {
		return "", ErrBiteshipDraftFailed
	}

	log.Printf("✅ Created draft order with ID: %s", response.ID)
	return response.ID, nil
}

// ConfirmDraftOrder confirms a draft order
// POST /v1/draft_orders/{id}/confirm
func (c *BiteshipClient) ConfirmDraftOrder(draftOrderID string) (*BiteshipOrder, error) {
	if draftOrderID == "" {
		return nil, fmt.Errorf("%w: draft order ID is required", ErrBiteshipInvalidRequest)
	}

	// First, check draft order status
	draftOrder, err := c.GetDraftOrder(draftOrderID)
	if err != nil {
		log.Printf("⚠️ Failed to get draft order status: %v", err)
		// Continue anyway, try to confirm
	} else {
		log.Printf("📋 Draft order status: %s", draftOrder.Status)
		
		// If already placed/confirmed, try to get the order details
		if draftOrder.Status == "placed" || draftOrder.Status == "confirmed" {
			log.Printf("⚠️ Draft order already in '%s' status, checking for existing order...", draftOrder.Status)
			
			// If we have waybill_id, return it
			if draftOrder.WaybillID != "" {
				log.Printf("✅ Found existing waybill: %s", draftOrder.WaybillID)
				return &BiteshipOrder{
					Success:     true,
					ID:          draftOrder.OrderID,
					Status:      draftOrder.Status,
					WaybillID:   draftOrder.WaybillID,
					TrackingID:  draftOrder.TrackingID,
					CourierCode: draftOrder.CourierCode,
				}, nil
			}
			
			// If placed but no waybill yet, return error - need to wait
			if draftOrder.Status == "placed" && draftOrder.WaybillID == "" {
				return nil, fmt.Errorf("draft order is placed but waybill not yet generated - please wait or contact Biteship support")
			}
		}
	}

	endpoint := fmt.Sprintf("/v1/draft_orders/%s/confirm", draftOrderID)
	
	respBody, err := c.doPost(endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response BiteshipOrder
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("%w: failed to parse order confirmation response", ErrBiteshipAPIFailed)
	}

	if !response.Success {
		return nil, ErrBiteshipConfirmFailed
	}

	log.Printf("✅ Confirmed order - Waybill: %s, Tracking: %s", response.WaybillID, response.TrackingID)
	return &response, nil
}

// GetDraftOrder retrieves draft order details
// GET /v1/draft_orders/{id}
func (c *BiteshipClient) GetDraftOrder(draftOrderID string) (*BiteshipDraftOrderDetail, error) {
	if draftOrderID == "" {
		return nil, fmt.Errorf("%w: draft order ID is required", ErrBiteshipInvalidRequest)
	}

	endpoint := fmt.Sprintf("/v1/draft_orders/%s", draftOrderID)
	
	respBody, err := c.doGet(endpoint)
	if err != nil {
		return nil, err
	}

	var response BiteshipDraftOrderDetail
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("%w: failed to parse draft order response", ErrBiteshipAPIFailed)
	}

	if !response.Success {
		return nil, fmt.Errorf("failed to get draft order details")
	}

	return &response, nil
}

// GetTracking fetches tracking information
// GET /v1/trackings/{tracking_id}
func (c *BiteshipClient) GetTracking(trackingID string) (*BiteshipTracking, error) {
	if trackingID == "" {
		return nil, fmt.Errorf("%w: tracking ID is required", ErrBiteshipInvalidRequest)
	}

	endpoint := fmt.Sprintf("/v1/trackings/%s", trackingID)
	
	respBody, err := c.doGet(endpoint)
	if err != nil {
		if errors.Is(err, ErrBiteshipAPIFailed) {
			return nil, ErrBiteshipTrackingNotFound
		}
		return nil, err
	}

	var response BiteshipTracking
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("%w: failed to parse tracking response", ErrBiteshipAPIFailed)
	}

	if !response.Success {
		return nil, ErrBiteshipTrackingNotFound
	}

	log.Printf("✅ Retrieved tracking for %s - Status: %s", trackingID, response.Status)
	return &response, nil
}

// MapBiteshipStatusToShipmentStatus maps Biteship tracking status to our shipment status
func MapBiteshipStatusToShipmentStatus(biteshipStatus string) string {
	status := strings.ToLower(biteshipStatus)

	switch {
	case strings.Contains(status, "delivered"):
		return "DELIVERED"
	case strings.Contains(status, "dropping_off"), strings.Contains(status, "out_for_delivery"):
		return "OUT_FOR_DELIVERY"
	case strings.Contains(status, "in_transit"), strings.Contains(status, "on_process"):
		return "IN_TRANSIT"
	case strings.Contains(status, "picked"), strings.Contains(status, "allocated"):
		return "SHIPPED"
	case strings.Contains(status, "picking_up"):
		return "PICKUP_SCHEDULED"
	case strings.Contains(status, "confirmed"):
		return "PROCESSING"
	case strings.Contains(status, "returned"):
		return "RETURNED_TO_SENDER"
	case strings.Contains(status, "rejected"), strings.Contains(status, "courier_not_found"):
		return "PICKUP_FAILED"
	case strings.Contains(status, "cancelled"):
		return "CANCELLED"
	default:
		return "IN_TRANSIT"
	}
}
