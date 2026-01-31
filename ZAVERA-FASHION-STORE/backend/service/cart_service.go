package service

import (
	"errors"
	"fmt"
	"zavera/dto"
	"zavera/models"
	"zavera/repository"
)

type CartService interface {
	GetCart(sessionID string) (*dto.CartResponse, error)
	GetCartForUser(userID int, sessionID string) (*dto.CartResponse, error)
	AddToCart(sessionID string, req dto.AddToCartRequest) (*dto.CartResponse, error)
	AddToCartForUser(userID int, sessionID string, req dto.AddToCartRequest) (*dto.CartResponse, error)
	UpdateCartItem(sessionID string, itemID int, quantity int) (*dto.CartResponse, error)
	UpdateCartItemForUser(userID int, sessionID string, itemID int, quantity int) (*dto.CartResponse, error)
	RemoveFromCart(sessionID string, itemID int) (*dto.CartResponse, error)
	RemoveFromCartForUser(userID int, sessionID string, itemID int) (*dto.CartResponse, error)
	ClearCart(sessionID string) error
	ClearCartForUser(userID int) error
	// Cart-User linking
	LinkCartToUser(sessionID string, userID int) error
	// Cart validation
	ValidateCart(userID int, sessionID string) (*dto.CartValidationResponse, error)
}

type cartService struct {
	cartRepo    repository.CartRepository
	productRepo repository.ProductRepository
}

func NewCartService(cartRepo repository.CartRepository, productRepo repository.ProductRepository) CartService {
	return &cartService{
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

func (s *cartService) GetCart(sessionID string) (*dto.CartResponse, error) {
	cart, err := s.cartRepo.FindOrCreateBySessionID(sessionID)
	if err != nil {
		return nil, err
	}

	return s.toCartResponse(cart)
}

func (s *cartService) AddToCart(sessionID string, req dto.AddToCartRequest) (*dto.CartResponse, error) {
	// Get or create cart
	cart, err := s.cartRepo.FindOrCreateBySessionID(sessionID)
	if err != nil {
		return nil, err
	}

	// Validate product and stock
	product, err := s.productRepo.FindByID(req.ProductID)
	if err != nil {
		return nil, errors.New("product not found")
	}

	// Resolve variant_id from metadata if not provided
	// For now, we'll let the cart item store metadata and resolve variant_id at checkout
	// This is a temporary solution until we add GetVariantsByProductID to repository
	variantID := req.VariantID

	// Stock validation:
	// - If product.Stock > 0: Simple product, check product stock
	// - If product.Stock = 0: Variant-based product, skip check here (variant stock checked at checkout)
	if product.Stock > 0 && product.Stock < req.Quantity {
		return nil, errors.New("insufficient stock")
	}

	// Create cart item
	cartItem := &models.CartItem{
		CartID:        cart.ID,
		ProductID:     req.ProductID,
		VariantID:     variantID,
		Quantity:      req.Quantity,
		PriceSnapshot: product.Price,
		Metadata:      req.Metadata,
	}

	err = s.cartRepo.AddItem(cartItem)
	if err != nil {
		return nil, err
	}

	// Return updated cart
	return s.GetCart(sessionID)
}

func (s *cartService) UpdateCartItem(sessionID string, itemID int, quantity int) (*dto.CartResponse, error) {
	// Get cart
	cart, err := s.cartRepo.FindOrCreateBySessionID(sessionID)
	if err != nil {
		return nil, err
	}

	// Find item in cart
	var targetItem *models.CartItem
	for i := range cart.Items {
		if cart.Items[i].ID == itemID {
			targetItem = &cart.Items[i]
			break
		}
	}

	if targetItem == nil {
		return nil, errors.New("cart item not found")
	}

	// If quantity is 0, delete item
	if quantity == 0 {
		return s.RemoveFromCart(sessionID, itemID)
	}

	// Validate stock
	product, err := s.productRepo.FindByID(targetItem.ProductID)
	if err != nil {
		return nil, err
	}

	// Stock validation:
	// - If product.Stock > 0: Simple product, check product stock
	// - If product.Stock = 0: Variant-based product, skip check here
	if product.Stock > 0 && product.Stock < quantity {
		return nil, errors.New("insufficient stock")
	}

	// Update quantity
	targetItem.Quantity = quantity
	err = s.cartRepo.UpdateItem(targetItem)
	if err != nil {
		return nil, err
	}

	// Return updated cart
	return s.GetCart(sessionID)
}

func (s *cartService) RemoveFromCart(sessionID string, itemID int) (*dto.CartResponse, error) {
	err := s.cartRepo.DeleteItem(itemID)
	if err != nil {
		return nil, err
	}

	return s.GetCart(sessionID)
}

func (s *cartService) ClearCart(sessionID string) error {
	cart, err := s.cartRepo.FindOrCreateBySessionID(sessionID)
	if err != nil {
		return err
	}

	return s.cartRepo.ClearCart(cart.ID)
}

// ClearCartForUser clears cart for logged-in user
func (s *cartService) ClearCartForUser(userID int) error {
	cart, err := s.cartRepo.FindByUserID(userID)
	if err != nil {
		return nil // No cart to clear
	}
	return s.cartRepo.ClearCart(cart.ID)
}

// GetCartForUser gets cart for logged-in user (prioritizes user_id over session)
func (s *cartService) GetCartForUser(userID int, sessionID string) (*dto.CartResponse, error) {
	// Try to find user's cart first
	cart, err := s.cartRepo.FindByUserID(userID)
	if err != nil {
		// No user cart exists, check if there's a guest cart to link
		guestCart, guestErr := s.cartRepo.FindOrCreateBySessionID(sessionID)
		if guestErr != nil {
			return nil, guestErr
		}
		// Link guest cart to user
		s.cartRepo.LinkCartToUser(guestCart.ID, userID)
		cart = guestCart
	}

	return s.toCartResponse(cart)
}

// AddToCartForUser adds item to cart for logged-in user
func (s *cartService) AddToCartForUser(userID int, sessionID string, req dto.AddToCartRequest) (*dto.CartResponse, error) {
	// Get or create user's cart
	cart, err := s.cartRepo.FindByUserID(userID)
	if err != nil {
		// Create new cart linked to user
		cart, err = s.cartRepo.FindOrCreateBySessionID(sessionID)
		if err != nil {
			return nil, err
		}
		s.cartRepo.LinkCartToUser(cart.ID, userID)
	}

	// Validate product and stock
	product, err := s.productRepo.FindByID(req.ProductID)
	if err != nil {
		return nil, errors.New("product not found")
	}

	// Resolve variant_id from metadata if not provided
	// For now, we'll let the cart item store metadata and resolve variant_id at checkout
	variantID := req.VariantID

	// Stock validation:
	// - If product.Stock > 0: Simple product, check product stock
	// - If product.Stock = 0: Variant-based product, skip check here
	if product.Stock > 0 && product.Stock < req.Quantity {
		return nil, errors.New("insufficient stock")
	}

	// Create cart item
	cartItem := &models.CartItem{
		CartID:        cart.ID,
		ProductID:     req.ProductID,
		VariantID:     variantID,
		Quantity:      req.Quantity,
		PriceSnapshot: product.Price,
		Metadata:      req.Metadata,
	}

	err = s.cartRepo.AddItem(cartItem)
	if err != nil {
		return nil, err
	}

	return s.GetCartForUser(userID, sessionID)
}

// UpdateCartItemForUser updates cart item for logged-in user
func (s *cartService) UpdateCartItemForUser(userID int, sessionID string, itemID int, quantity int) (*dto.CartResponse, error) {
	cart, err := s.cartRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("cart not found")
	}

	// Find item in cart
	var targetItem *models.CartItem
	for i := range cart.Items {
		if cart.Items[i].ID == itemID {
			targetItem = &cart.Items[i]
			break
		}
	}

	if targetItem == nil {
		return nil, errors.New("cart item not found")
	}

	if quantity == 0 {
		return s.RemoveFromCartForUser(userID, sessionID, itemID)
	}

	// Validate stock
	product, err := s.productRepo.FindByID(targetItem.ProductID)
	if err != nil {
		return nil, err
	}

	// Stock validation:
	// - If product.Stock > 0: Simple product, check product stock
	// - If product.Stock = 0: Variant-based product, skip check here
	if product.Stock > 0 && product.Stock < quantity {
		return nil, errors.New("insufficient stock")
	}

	targetItem.Quantity = quantity
	err = s.cartRepo.UpdateItem(targetItem)
	if err != nil {
		return nil, err
	}

	return s.GetCartForUser(userID, sessionID)
}

// RemoveFromCartForUser removes item from cart for logged-in user
func (s *cartService) RemoveFromCartForUser(userID int, sessionID string, itemID int) (*dto.CartResponse, error) {
	err := s.cartRepo.DeleteItem(itemID)
	if err != nil {
		return nil, err
	}

	return s.GetCartForUser(userID, sessionID)
}

// LinkCartToUser links guest cart to user account on login
// This merges guest cart items into user's existing cart if any
func (s *cartService) LinkCartToUser(sessionID string, userID int) error {
	// Get guest cart
	guestCart, err := s.cartRepo.FindOrCreateBySessionID(sessionID)
	if err != nil {
		return err
	}

	// Check if user already has a cart
	userCart, err := s.cartRepo.FindByUserID(userID)
	if err != nil {
		// No existing user cart, just link the guest cart
		return s.cartRepo.LinkCartToUser(guestCart.ID, userID)
	}

	// User has existing cart, merge guest cart into it
	if len(guestCart.Items) > 0 {
		return s.cartRepo.MergeGuestCartToUser(guestCart.ID, userCart.ID)
	}

	return nil
}

func (s *cartService) toCartResponse(cart *models.Cart) (*dto.CartResponse, error) {
	response := &dto.CartResponse{
		ID:    cart.ID,
		Items: []dto.CartItemResponse{},
	}

	var subtotal float64
	var itemCount int

	for _, item := range cart.Items {
		// Get product details
		product, err := s.productRepo.FindByID(item.ProductID)
		if err != nil {
			continue
		}

		// Get primary image
		primaryImage := ""
		if len(product.Images) > 0 {
			for _, img := range product.Images {
				if img.IsPrimary {
					primaryImage = img.ImageURL
					break
				}
			}
			if primaryImage == "" {
				primaryImage = product.Images[0].ImageURL
			}
		}

		itemResponse := dto.CartItemResponse{
			ID:           item.ID,
			ProductID:    item.ProductID,
			ProductName:  product.Name,
			ProductImage: primaryImage,
			Quantity:     item.Quantity,
			PricePerUnit: item.PriceSnapshot,
			Subtotal:     item.PriceSnapshot * float64(item.Quantity),
			Stock:        product.Stock,
			Metadata:     item.Metadata,
		}

		response.Items = append(response.Items, itemResponse)
		subtotal += itemResponse.Subtotal
		itemCount += item.Quantity
	}

	response.Subtotal = subtotal
	response.ItemCount = itemCount

	return response, nil
}

// ValidateCart validates cart items against current product data
// Returns changes if any product price, weight, or stock has changed
func (s *cartService) ValidateCart(userID int, sessionID string) (*dto.CartValidationResponse, error) {
	// Get current cart
	cart, err := s.GetCartForUser(userID, sessionID)
	if err != nil {
		return nil, err
	}

	if len(cart.Items) == 0 {
		return &dto.CartValidationResponse{
			Valid:   true,
			Changes: []dto.CartItemChange{},
			Cart:    cart,
			Message: "Cart is empty",
		}, nil
	}

	var changes []dto.CartItemChange
	hasChanges := false

	// Check each cart item against current product data
	for _, item := range cart.Items {
		product, err := s.productRepo.FindByID(item.ProductID)
		if err != nil {
			// Product no longer exists
			changes = append(changes, dto.CartItemChange{
				CartItemID:  item.ID,
				ProductID:   item.ProductID,
				ProductName: item.ProductName,
				ChangeType:  "product_unavailable",
				Message:     "Product is no longer available",
			})
			hasChanges = true
			continue
		}

		// Stock validation:
		// - If product.Stock > 0: Simple product, validate stock
		// - If product.Stock = 0: Variant product, skip validation (stock in variants)
		if product.Stock > 0 {
			// Simple product - check stock
			if product.Stock < item.Quantity {
				changes = append(changes, dto.CartItemChange{
					CartItemID:   item.ID,
					ProductID:    item.ProductID,
					ProductName:  item.ProductName,
					ChangeType:   "stock_insufficient",
					CurrentStock: product.Stock,
					Message:      fmt.Sprintf("Only %d items available", product.Stock),
				})
				hasChanges = true
			}
		}
		// Note: For variant products (product.Stock = 0), stock will be validated at checkout

		// Check price changes
		if product.Price != item.PricePerUnit {
			changes = append(changes, dto.CartItemChange{
				CartItemID:  item.ID,
				ProductID:   item.ProductID,
				ProductName: item.ProductName,
				ChangeType:  "price_changed",
				OldPrice:    item.PricePerUnit,
				NewPrice:    product.Price,
				Message:     "Price has changed",
			})
			hasChanges = true
		}

		// Check weight changes
		if product.Weight != 0 {
			// Get old weight from cart item metadata or assume it hasn't changed
			// For now, we'll just notify if weight is different from what we expect
			// This is informational only
		}
	}

	// If there are changes, update cart with current prices
	if hasChanges {
		// Reload cart to get updated data
		cart, _ = s.GetCartForUser(userID, sessionID)
	}

	return &dto.CartValidationResponse{
		Valid:   !hasChanges,
		Changes: changes,
		Cart:    cart,
		Message: getMessage(hasChanges, len(changes)),
	}, nil
}

func getMessage(hasChanges bool, changeCount int) string {
	if !hasChanges {
		return "Cart is valid and ready for checkout"
	}
	if changeCount == 1 {
		return "1 item in your cart has changed"
	}
	return fmt.Sprintf("%d items in your cart have changed", changeCount)
}
