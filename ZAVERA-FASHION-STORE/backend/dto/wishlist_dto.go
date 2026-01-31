package dto

// WishlistResponse represents the wishlist API response
type WishlistResponse struct {
	Items []WishlistItemResponse `json:"items"`
	Count int                    `json:"count"`
}

// WishlistItemResponse represents a wishlist item in API response
type WishlistItemResponse struct {
	ID           int     `json:"id"`
	ProductID    int     `json:"product_id"`
	ProductName  string  `json:"product_name"`
	ProductImage string  `json:"product_image"`
	ProductPrice float64 `json:"product_price"`
	ProductStock int     `json:"product_stock"`
	IsAvailable  bool    `json:"is_available"`
	AddedAt      string  `json:"added_at"`
}

// AddToWishlistRequest represents the request to add item to wishlist
type AddToWishlistRequest struct {
	ProductID int `json:"product_id" binding:"required"`
}

// MoveToCartRequest represents the request to move wishlist item to cart
type MoveToCartRequest struct {
	ProductID int `json:"product_id" binding:"required"`
}
