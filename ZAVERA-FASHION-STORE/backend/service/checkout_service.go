package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"zavera/dto"
	"zavera/models"
	"zavera/repository"
)

var (
	ErrShippingRequired = errors.New("shipping selection is required")
)

// CheckoutService handles the complete checkout flow with shipping
type CheckoutService interface {
	// Full checkout with shipping
	CheckoutWithShipping(sessionID string, req dto.CheckoutWithShippingRequest, userID *int) (*dto.CheckoutWithShippingResponse, error)
	
	// Get shipping rates for cart (uses district ID for accurate pricing)
	GetCartShippingOptions(sessionID string, destinationDistrictID string, courier string) (*dto.CartShippingPreviewResponse, error)
}

type checkoutService struct {
	orderRepo    repository.OrderRepository
	cartRepo     repository.CartRepository
	productRepo  repository.ProductRepository
	shippingRepo repository.ShippingRepository
	emailRepo    repository.EmailRepository
	biteship     *BiteshipClient
	emailService EmailService
}

func NewCheckoutService(
	orderRepo repository.OrderRepository,
	cartRepo repository.CartRepository,
	productRepo repository.ProductRepository,
	shippingRepo repository.ShippingRepository,
	emailRepo repository.EmailRepository,
) CheckoutService {
	// Create email service
	var emailSvc EmailService
	if emailRepo != nil {
		emailSvc = NewEmailService(emailRepo)
	}
	
	return &checkoutService{
		orderRepo:    orderRepo,
		cartRepo:     cartRepo,
		productRepo:  productRepo,
		shippingRepo: shippingRepo,
		biteship:     NewBiteshipClient(),
		emailService: emailSvc,
	}
}

// CheckoutWithShipping creates an order with shipping selection
func (s *checkoutService) CheckoutWithShipping(sessionID string, req dto.CheckoutWithShippingRequest, userID *int) (*dto.CheckoutWithShippingResponse, error) {
	// 1. Get cart - prioritize user_id over session_id for logged-in users
	var cart *models.Cart
	var err error
	
	if userID != nil && *userID > 0 {
		// Try to find cart by user_id first
		cart, err = s.cartRepo.FindByUserID(*userID)
		if err != nil || cart == nil || len(cart.Items) == 0 {
			// Fallback to session
			cart, err = s.cartRepo.FindOrCreateBySessionID(sessionID)
		}
	} else {
		cart, err = s.cartRepo.FindOrCreateBySessionID(sessionID)
	}
	
	if err != nil {
		return nil, err
	}

	if len(cart.Items) == 0 {
		return nil, ErrCartEmpty
	}

	// Resolve courier codes (support both Biteship-native and legacy fields)
	courierCode := req.CourierCode
	if courierCode == "" {
		courierCode = req.ProviderCode
	}
	courierServiceCode := req.CourierServiceCode
	if courierServiceCode == "" {
		courierServiceCode = req.ServiceCode
	}
	
	// Validate courier selection
	if courierCode == "" || courierServiceCode == "" {
		return nil, fmt.Errorf("courier selection is required")
	}

	// 2. Resolve shipping address
	var addressSnapshot models.ShippingAddressSnapshot
	var destinationPostalCode string
	var destinationAreaID string
	var destinationAreaName string
	// Legacy fields for backward compatibility
	var destinationCityID string
	var destinationCityName string

	if req.AddressID != nil && *req.AddressID > 0 {
		// Use saved address
		address, err := s.shippingRepo.GetAddressByID(*req.AddressID)
		if err != nil {
			return nil, ErrAddressNotFound
		}
		addressSnapshot = AddressToSnapshot(address)
		destinationPostalCode = address.PostalCode
		destinationCityID = address.CityID
		destinationCityName = address.CityName
	} else if req.ShippingAddress != nil {
		// Use guest address (Biteship-native)
		addressSnapshot = GuestAddressToSnapshot(req.ShippingAddress)
		destinationPostalCode = req.ShippingAddress.PostalCode
		destinationAreaID = req.ShippingAddress.AreaID
		destinationAreaName = req.ShippingAddress.AreaName
		destinationCityID = req.ShippingAddress.CityID
		destinationCityName = req.ShippingAddress.CityName
		
		// Use area_name as city_name if not provided
		if destinationCityName == "" && destinationAreaName != "" {
			destinationCityName = destinationAreaName
		}
	} else {
		return nil, ErrInvalidAddress
	}

	// Validate postal code is provided (required for Biteship)
	if destinationPostalCode == "" {
		return nil, fmt.Errorf("%w: postal code is required for shipping calculation", ErrInvalidAddress)
	}

	// 3. Calculate cart totals, weight, and build items with dimensions
	var subtotal float64
	var totalWeight int
	var orderItems []models.OrderItem
	var biteshipItems []GetRatesRequestItem

	for _, item := range cart.Items {
		product, err := s.productRepo.FindByID(item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("product not found: %w", err)
		}

		// Stock validation:
		// - If product.Stock > 0: Simple product, validate stock
		// - If product.Stock = 0: Variant product, skip validation here (stock in variants)
		// Note: For variant products, stock will be validated and deducted at variant level during order creation
		if product.Stock > 0 && product.Stock < item.Quantity {
			return nil, fmt.Errorf("%w for product: %s", ErrInsufficientStock, product.Name)
		}

		itemSubtotal := item.PriceSnapshot * float64(item.Quantity)
		subtotal += itemSubtotal

		// Get product weight from database (default 500g if not set)
		productWeight := product.Weight
		if productWeight <= 0 {
			productWeight = 500 // Fallback default
		}
		totalWeight += productWeight * item.Quantity

		// Build Biteship item with dimensions
		// FIX: Send total weight with quantity=1 to avoid Biteship double-counting bug
		totalItemWeight := productWeight * item.Quantity
		biteshipItem := GetRatesRequestItem{
			Name:     product.Name,
			Value:    item.PriceSnapshot * float64(item.Quantity), // Total value
			Weight:   totalItemWeight,  // Total weight (not per-item)
			Quantity: 1,                // Always 1 to avoid double-counting
		}
		
		// Add dimensions if available (for volumetric weight calculation)
		// Note: Dimensions should be for the COMBINED package, not per-item
		if product.Length > 0 {
			biteshipItem.Length = product.Length
		}
		if product.Width > 0 {
			biteshipItem.Width = product.Width
		}
		if product.Height > 0 {
			biteshipItem.Height = product.Height
		}
		
		biteshipItems = append(biteshipItems, biteshipItem)

		// Get product image
		productImage := ""
		if len(product.Images) > 0 {
			for _, img := range product.Images {
				if img.IsPrimary {
					productImage = img.ImageURL
					break
				}
			}
			if productImage == "" {
				productImage = product.Images[0].ImageURL
			}
		}

		orderItem := models.OrderItem{
			ProductID:    item.ProductID,
			VariantID:    item.VariantID, // Copy variant_id from cart item
			ProductName:  product.Name,
			ProductImage: productImage,
			Quantity:     item.Quantity,
			PricePerUnit: item.PriceSnapshot,
			Subtotal:     itemSubtotal,
			Metadata:     item.Metadata,
		}

		orderItems = append(orderItems, orderItem)
	}

	// Minimum weight 1000g
	if totalWeight < 1000 {
		totalWeight = 1000
	}

	// 4. Get shipping rate from Biteship API using postal_code
	destPostalCode, _ := strconv.Atoi(destinationPostalCode)
	if destPostalCode == 0 {
		destPostalCode = 10110 // Default Jakarta if no postal code
	}
	
	biteshipReq := GetRatesRequest{
		OriginPostalCode:      DefaultOriginPostalCode, // Pedurungan, Semarang 50113
		DestinationPostalCode: destPostalCode,
		Couriers:              courierCode,
		Items:                 biteshipItems, // Send individual items with dimensions
	}

	biteshipRates, err := s.biteship.GetRates(biteshipReq)
	
	// Find selected service
	var selectedRate *BiteshipRate
	var providerName, serviceName string
	var shippingCost float64
	var etd string
	
	if err != nil {
		// Fallback: use dummy shipping rate when API fails
		log.Printf("⚠️ Using fallback shipping rate due to API error: %v", err)
		providerName = courierCode
		serviceName = courierServiceCode
		shippingCost = 15000 // Default shipping cost
		etd = "2-3 days"
	} else {
		for i, rate := range biteshipRates {
			if rate.CourierServiceCode == courierServiceCode || rate.CourierCode == courierCode {
				selectedRate = &biteshipRates[i]
				break
			}
		}
		
		if selectedRate == nil && len(biteshipRates) > 0 {
			// Use first available rate if exact match not found
			selectedRate = &biteshipRates[0]
		}
		
		if selectedRate != nil {
			providerName = selectedRate.CourierName
			serviceName = selectedRate.CourierServiceName
			shippingCost = selectedRate.Price
			etd = selectedRate.Duration
		} else {
			return nil, ErrInvalidCourier
		}
	}

	// 5. Calculate totals
	tax := 0.0
	discount := 0.0
	totalAmount := subtotal + shippingCost + tax - discount

	// 6. Create order with shipping locked
	addressJSON, _ := json.Marshal(addressSnapshot)

	order := &models.Order{
		UserID:        userID,
		CustomerName:  req.CustomerName,
		CustomerEmail: req.CustomerEmail,
		CustomerPhone: req.CustomerPhone,
		Subtotal:      subtotal,
		ShippingCost:  shippingCost,
		Tax:           tax,
		Discount:      discount,
		TotalAmount:   totalAmount,
		Status:        models.OrderStatusPending,
		Notes:         req.Notes,
		Metadata: map[string]any{
			"shipping_courier_code":     courierCode,
			"shipping_service_code":     courierServiceCode,
			"shipping_locked":           true,
			"total_weight":              totalWeight,
			"shipping_address_snapshot": string(addressJSON),
			"destination_postal_code":   destinationPostalCode,
			"destination_area_id":       destinationAreaID,
			"destination_area_name":     destinationAreaName,
			"destination_city_id":       destinationCityID,
			"destination_city_name":     destinationCityName,
			"shipping_source":           "biteship",
		},
	}

	err = s.orderRepo.Create(order, orderItems)
	if err != nil {
		return nil, err
	}

	// Send notification to admin dashboard
	customerName := order.CustomerName
	if customerName == "" {
		customerName = "Customer"
	}
	NotifyOrderCreated(order.OrderCode, customerName, order.TotalAmount)

	// 7. Create shipment record (status: PENDING until payment)
	shipment := &models.Shipment{
		OrderID:             order.ID,
		ProviderCode:        courierCode,
		ProviderName:        providerName,
		ServiceCode:         courierServiceCode,
		ServiceName:         serviceName,
		Cost:                shippingCost,
		ETD:                 etd,
		Weight:              totalWeight,
		Status:              models.ShipmentStatusPending,
		OriginCityID:        DefaultOriginCityID,
		OriginCityName:      "Kota Semarang",
		DestinationCityID:   destinationCityID,
		DestinationCityName: destinationCityName,
	}

	s.shippingRepo.CreateShipment(shipment)

	// 7b. Create Biteship draft order for auto-resi generation
	// This allows admin to auto-generate resi when shipping the order
	log.Printf("📦 Creating Biteship draft order for order %d", order.ID)
	
	// Build draft order items
	var draftItems []CreateDraftOrderItem
	for _, item := range cart.Items {
		product, err := s.productRepo.FindByID(item.ProductID)
		if err != nil {
			continue
		}
		
		// Get product dimensions
		length := product.Length
		width := product.Width
		height := product.Height
		weight := product.Weight
		if weight == 0 {
			weight = 500 // Default 500g
		}
		
		draftItems = append(draftItems, CreateDraftOrderItem{
			Name:        product.Name,
			Description: product.Description,
			Value:       float64(product.Price),
			Length:      length,
			Width:       width,
			Height:      height,
			Weight:      weight,
			Quantity:    item.Quantity,
		})
	}
	
	// Create draft order params
	// IMPORTANT: Use postal_code instead of area_id for better compatibility
	draftParams := CreateDraftOrderParams{
		OriginAreaID:            "", // Leave empty, use postal_code instead
		OriginAddress:           "Jl. Pedurungan Tengah, Pedurungan, Semarang",
		OriginPostalCode:        fmt.Sprintf("%d", DefaultOriginPostalCode),
		OriginContactName:       "ZAVERA Fashion Store",
		OriginContactPhone:      "081234567890",
		DestinationAreaID:       "", // Leave empty, use postal_code instead
		DestinationAddress:      addressSnapshot.FullAddress,
		DestinationPostalCode:   destinationPostalCode,
		DestinationContactName:  addressSnapshot.RecipientName,
		DestinationContactPhone: addressSnapshot.Phone,
		CourierCode:             courierCode,
		CourierServiceCode:      courierServiceCode,
		Items:                   draftItems,
	}
	
	// Create draft order via shipping service
	shippingSvc := NewShippingService(s.shippingRepo, s.cartRepo, s.productRepo, s.orderRepo)
	draftResp, err := shippingSvc.CreateDraftOrderForCheckout(order.ID, draftParams)
	if err != nil {
		log.Printf("⚠️ Failed to create Biteship draft order: %v (will fallback to manual resi)", err)
		// Don't fail checkout if draft order fails - admin can still input resi manually
	} else {
		log.Printf("✅ Created Biteship draft order: %s for order %d", draftResp.ID, order.ID)
	}

	// 7c. Create shipping snapshot for audit (stores raw API response)
	shippingSnapshot := &models.ShippingSnapshot{
		OrderID:               order.ID,
		Courier:               courierCode,
		Service:               courierServiceCode,
		Cost:                  shippingCost,
		ETD:                   etd,
		OriginCityID:          DefaultOriginCityID,
		OriginCityName:        "Kota Semarang",
		DestinationCityID:     destinationCityID,
		DestinationCityName:   destinationCityName,
		DestinationDistrictID: destinationPostalCode,
		Weight:                totalWeight,
		BiteshipRawJSON: map[string]any{
			"courier_code":   courierCode,
			"courier_name":   providerName,
			"service_code":   courierServiceCode,
			"service_name":   serviceName,
			"cost":           shippingCost,
			"etd":            etd,
			"weight":         totalWeight,
			"source":         "biteship",
			"snapshot_time":  order.CreatedAt,
		},
	}
	s.shippingRepo.CreateShippingSnapshot(shippingSnapshot)

	// 9. Clear cart after successful checkout
	s.cartRepo.ClearCart(cart.ID)
	
	// 10. NO email for order created - Tokopedia style
	// Email only sent after payment success (legal transaction)
	// Customer can see order status in their account page

	// 11. Return response
	return &dto.CheckoutWithShippingResponse{
		OrderID:        order.ID,
		OrderCode:      order.OrderCode,
		Subtotal:       subtotal,
		ShippingCost:   shippingCost,
		TotalAmount:    totalAmount,
		Status:         string(order.Status),
		ShippingLocked: true,
		Provider:       providerName,
		Service:        serviceName,
		ETD:            etd,
		ShippingAddress: dto.ShippingAddressDisplay{
			RecipientName: addressSnapshot.RecipientName,
			Phone:         addressSnapshot.Phone,
			FullAddress:   addressSnapshot.FullAddress,
			CityName:      addressSnapshot.CityName,
			ProvinceName:  addressSnapshot.ProvinceName,
			PostalCode:    addressSnapshot.PostalCode,
		},
	}, nil
}

// GetCartShippingOptions returns shipping options for current cart using Biteship
func (s *checkoutService) GetCartShippingOptions(sessionID string, destinationPostalCode string, courier string) (*dto.CartShippingPreviewResponse, error) {
	fmt.Printf("🛒 GetCartShippingOptions - SessionID: %s, DestPostalCode: %s\n", sessionID, destinationPostalCode)
	
	// Get cart
	cart, err := s.cartRepo.FindOrCreateBySessionID(sessionID)
	if err != nil {
		fmt.Printf("❌ Cart error: %v\n", err)
		return nil, err
	}

	fmt.Printf("📦 Cart ID: %d, Items count: %d\n", cart.ID, len(cart.Items))

	if len(cart.Items) == 0 {
		fmt.Println("❌ Cart is empty!")
		return nil, ErrCartEmpty
	}

	// Calculate totals and build items with dimensions
	var subtotal float64
	var totalWeight int
	var biteshipItems []GetRatesRequestItem

	for _, item := range cart.Items {
		product, err := s.productRepo.FindByID(item.ProductID)
		if err != nil {
			continue
		}

		subtotal += item.PriceSnapshot * float64(item.Quantity)
		
		// Get product weight from database (default 500g if not set)
		productWeight := product.Weight
		if productWeight <= 0 {
			productWeight = 500 // Fallback default
		}
		totalWeight += productWeight * item.Quantity
		
		// Add item with dimensions to Biteship request
		// FIX: Send total weight with quantity=1 to avoid Biteship double-counting bug
		totalItemWeight := productWeight * item.Quantity
		biteshipItem := GetRatesRequestItem{
			Name:     product.Name,
			Value:    item.PriceSnapshot * float64(item.Quantity), // Total value
			Weight:   totalItemWeight,  // Total weight (not per-item)
			Quantity: 1,                // Always 1 to avoid double-counting
		}
		
		// Add dimensions if available (for volumetric weight calculation)
		// Note: Dimensions should be for the COMBINED package, not per-item
		if product.Length > 0 {
			biteshipItem.Length = product.Length
		}
		if product.Width > 0 {
			biteshipItem.Width = product.Width
		}
		if product.Height > 0 {
			biteshipItem.Height = product.Height
		}
		
		biteshipItems = append(biteshipItems, biteshipItem)
	}

	// Minimum weight
	if totalWeight < 1000 {
		totalWeight = 1000
	}

	// Get rates using Biteship API with postal_code
	if courier == "" {
		courier = "jne,jnt,sicepat,tiki,anteraja"
	}

	// Parse destination postal code
	destPostalCode, _ := strconv.Atoi(destinationPostalCode)
	if destPostalCode == 0 {
		fmt.Printf("❌ Invalid postal code: %s\n", destinationPostalCode)
		return nil, fmt.Errorf("invalid postal code")
	}

	fmt.Printf("📡 Getting Biteship rates - Origin: %d (Pedurungan, Semarang), Dest: %d, Weight: %d, Items: %d, Courier: %s\n", DefaultOriginPostalCode, destPostalCode, totalWeight, len(biteshipItems), courier)
	
	// Debug: Print each item details
	for i, item := range biteshipItems {
		fmt.Printf("  Item %d: %s - Weight: %dg, Dimensions: %dx%dx%d cm, Qty: %d\n", 
			i+1, item.Name, item.Weight, item.Length, item.Width, item.Height, item.Quantity)
	}

	biteshipReq := GetRatesRequest{
		OriginPostalCode:      DefaultOriginPostalCode, // Pedurungan, Semarang 50113
		DestinationPostalCode: destPostalCode,
		Couriers:              courier,
		Items:                 biteshipItems, // Send individual items with dimensions
	}

	biteshipRates, err := s.biteship.GetRates(biteshipReq)
	if err != nil {
		fmt.Printf("❌ Biteship GetRates error: %v\n", err)
		return nil, err
	}

	fmt.Printf("✅ Got %d rates from Biteship API\n", len(biteshipRates))
	
	// Debug: Print first few rates
	for i, rate := range biteshipRates {
		if i < 5 {
			fmt.Printf("  Rate %d: %s %s - Rp %.0f (%s)\n", 
				i+1, rate.CourierName, rate.CourierServiceName, rate.Price, rate.Duration)
		}
	}

	// Get provider logos
	providers, _ := s.shippingRepo.GetActiveProviders()
	providerLogos := make(map[string]string)
	for _, p := range providers {
		providerLogos[p.Code] = p.LogoURL
	}

	// Convert to DTO response
	var rateResponses []dto.ShippingRateResponse
	groupedRates := make(map[string][]dto.ShippingRateResponse)
	var regularMinPrice float64 = -1

	for _, rate := range biteshipRates {
		// Categorize rate
		category := CategorizeRate(rate.ServiceType)
		
		rateDTO := dto.ShippingRateResponse{
			ProviderCode:     rate.CourierCode,
			ProviderName:     rate.CourierName,
			ProviderLogo:     providerLogos[rate.CourierCode],
			ServiceCode:      rate.CourierServiceCode,
			ServiceName:      rate.CourierServiceName,
			Description:      rate.Description,
			Cost:             rate.Price,
			ETD:              rate.Duration,
			ShippingCategory: dto.ShippingCategory(category),
		}
		
		// Track regular min price
		if category == "Regular" {
			if regularMinPrice < 0 || rate.Price < regularMinPrice {
				regularMinPrice = rate.Price
			}
		}
		
		rateResponses = append(rateResponses, rateDTO)
		
		// Group by category
		groupedRates[category] = append(groupedRates[category], rateDTO)
	}

	// Format weight for display
	weightKg := fmt.Sprintf("%.1f kg", float64(totalWeight)/1000)
	if totalWeight < 1000 {
		weightKg = fmt.Sprintf("%d g", totalWeight)
	}

	return &dto.CartShippingPreviewResponse{
		CartSubtotal:    subtotal,
		TotalWeight:     totalWeight,
		TotalWeightKg:   weightKg,
		OriginCity:      "Semarang",
		GroupedRates:    groupedRates,
		Rates:           rateResponses,
		RegularMinPrice: regularMinPrice,
	}, nil
}
