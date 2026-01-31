package service

import (
	"errors"
	"zavera/dto"
	"zavera/models"
	"zavera/repository"
)

type WishlistService interface {
	GetWishlist(userID int) (*dto.WishlistResponse, error)
	AddToWishlist(userID int, productID int) (*dto.WishlistResponse, error)
	RemoveFromWishlist(userID int, productID int) (*dto.WishlistResponse, error)
	MoveToCart(userID int, productID int, sessionID string) (*dto.CartResponse, error)
	IsInWishlist(userID int, productID int) (bool, error)
}

type wishlistService struct {
	wishlistRepo repository.WishlistRepository
	productRepo  repository.ProductRepository
	cartRepo     repository.CartRepository
}

func NewWishlistService(
	wishlistRepo repository.WishlistRepository,
	productRepo repository.ProductRepository,
	cartRepo repository.CartRepository,
) WishlistService {
	return &wishlistService{
		wishlistRepo: wishlistRepo,
		productRepo:  productRepo,
		cartRepo:     cartRepo,
	}
}

// GetWishlist returns user's wishlist with product details
func (s *wishlistService) GetWishlist(userID int) (*dto.WishlistResponse, error) {
	items, err := s.wishlistRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	response := &dto.WishlistResponse{
		Items: []dto.WishlistItemResponse{},
		Count: 0,
	}

	for _, item := range items {
		// Get product details
		product, err := s.productRepo.FindByID(item.ProductID)
		if err != nil {
			// Product might have been deleted, skip it
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

		itemResponse := dto.WishlistItemResponse{
			ID:           item.ID,
			ProductID:    item.ProductID,
			ProductName:  product.Name,
			ProductImage: primaryImage,
			ProductPrice: product.Price,
			ProductStock: product.Stock,
			IsAvailable:  product.IsActive && product.Stock > 0,
			AddedAt:      item.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		response.Items = append(response.Items, itemResponse)
	}

	response.Count = len(response.Items)
	return response, nil
}

// AddToWishlist adds a product to user's wishlist
func (s *wishlistService) AddToWishlist(userID int, productID int) (*dto.WishlistResponse, error) {
	// Validate product exists
	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		return nil, errors.New("product not found")
	}

	if !product.IsActive {
		return nil, errors.New("product is not available")
	}

	// Add to wishlist
	_, err = s.wishlistRepo.Add(userID, productID)
	if err != nil {
		return nil, err
	}

	// Return updated wishlist
	return s.GetWishlist(userID)
}

// RemoveFromWishlist removes a product from user's wishlist
func (s *wishlistService) RemoveFromWishlist(userID int, productID int) (*dto.WishlistResponse, error) {
	err := s.wishlistRepo.Remove(userID, productID)
	if err != nil {
		return nil, err
	}

	// Return updated wishlist
	return s.GetWishlist(userID)
}

// MoveToCart moves a product from wishlist to cart
func (s *wishlistService) MoveToCart(userID int, productID int, sessionID string) (*dto.CartResponse, error) {
	// Validate product exists and is available
	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		return nil, errors.New("product not found")
	}

	if !product.IsActive || product.Stock <= 0 {
		return nil, errors.New("product is not available")
	}

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

	// Add to cart with quantity 1
	cartItem := &models.CartItem{
		CartID:        cart.ID,
		ProductID:     productID,
		Quantity:      1,
		PriceSnapshot: product.Price,
		Metadata:      map[string]any{"selected_size": "M"},
	}

	err = s.cartRepo.AddItem(cartItem)
	if err != nil {
		return nil, err
	}

	// Remove from wishlist
	s.wishlistRepo.Remove(userID, productID)

	// Return updated cart
	cart, err = s.cartRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	return s.toCartResponse(cart)
}

// IsInWishlist checks if a product is in user's wishlist
func (s *wishlistService) IsInWishlist(userID int, productID int) (bool, error) {
	return s.wishlistRepo.IsInWishlist(userID, productID)
}

// Helper to convert cart to response
func (s *wishlistService) toCartResponse(cart *models.Cart) (*dto.CartResponse, error) {
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
