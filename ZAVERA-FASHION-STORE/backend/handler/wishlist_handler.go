package handler

import (
	"log"
	"net/http"
	"strconv"
	"zavera/dto"
	"zavera/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WishlistHandler struct {
	wishlistService service.WishlistService
}

func NewWishlistHandler(wishlistService service.WishlistService) *WishlistHandler {
	return &WishlistHandler{
		wishlistService: wishlistService,
	}
}

// getUserID gets user ID from context (required for wishlist)
func (h *WishlistHandler) getUserID(c *gin.Context) *int {
	userID, exists := c.Get("user_id")
	if !exists {
		return nil
	}

	// Convert to int (JWT stores as float64)
	switch v := userID.(type) {
	case int:
		return &v
	case int64:
		id := int(v)
		return &id
	case float64:
		id := int(v)
		return &id
	case uint:
		id := int(v)
		return &id
	default:
		return nil
	}
}

// getOrCreateSessionID gets session ID from cookie or creates new one
func (h *WishlistHandler) getOrCreateSessionID(c *gin.Context) string {
	sessionID, err := c.Cookie("session_id")
	if err != nil || sessionID == "" {
		sessionID = uuid.New().String()
		c.SetCookie("session_id", sessionID, 86400*30, "/", "", false, true)
	}
	return sessionID
}

// GetWishlist godoc
// @Summary Get user's wishlist
// @Description Get current wishlist with all items (requires authentication)
// @Tags wishlist
// @Accept json
// @Produce json
// @Success 200 {object} dto.WishlistResponse
// @Router /api/wishlist [get]
func (h *WishlistHandler) GetWishlist(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Please login to view wishlist",
		})
		return
	}

	log.Printf("💝 GetWishlist - UserID: %d", *userID)

	wishlist, err := h.wishlistService.GetWishlist(*userID)
	if err != nil {
		log.Printf("❌ GetWishlist error: %v", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	log.Printf("✅ GetWishlist success - %d items", wishlist.Count)
	c.JSON(http.StatusOK, wishlist)
}

// AddToWishlist godoc
// @Summary Add item to wishlist
// @Description Add product to wishlist (requires authentication)
// @Tags wishlist
// @Accept json
// @Produce json
// @Param request body dto.AddToWishlistRequest true "Add to wishlist request"
// @Success 200 {object} dto.WishlistResponse
// @Router /api/wishlist [post]
func (h *WishlistHandler) AddToWishlist(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Please login to add items to wishlist",
		})
		return
	}

	var req dto.AddToWishlistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	log.Printf("💝 AddToWishlist - UserID: %d, ProductID: %d", *userID, req.ProductID)

	wishlist, err := h.wishlistService.AddToWishlist(*userID, req.ProductID)
	if err != nil {
		log.Printf("❌ AddToWishlist error: %v", err)
		status := http.StatusInternalServerError
		if err.Error() == "product not found" {
			status = http.StatusNotFound
		} else if err.Error() == "product is not available" {
			status = http.StatusBadRequest
		}

		c.JSON(status, dto.ErrorResponse{
			Error:   "add_to_wishlist_failed",
			Message: err.Error(),
		})
		return
	}

	log.Printf("✅ AddToWishlist success - Wishlist has %d items", wishlist.Count)
	c.JSON(http.StatusOK, wishlist)
}

// RemoveFromWishlist godoc
// @Summary Remove item from wishlist
// @Description Remove product from wishlist by product ID (requires authentication)
// @Tags wishlist
// @Accept json
// @Produce json
// @Param productId path int true "Product ID"
// @Success 200 {object} dto.WishlistResponse
// @Router /api/wishlist/{productId} [delete]
func (h *WishlistHandler) RemoveFromWishlist(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Please login to remove items from wishlist",
		})
		return
	}

	productIDStr := c.Param("productId")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid product ID",
		})
		return
	}

	log.Printf("💔 RemoveFromWishlist - UserID: %d, ProductID: %d", *userID, productID)

	wishlist, err := h.wishlistService.RemoveFromWishlist(*userID, productID)
	if err != nil {
		log.Printf("❌ RemoveFromWishlist error: %v", err)
		status := http.StatusInternalServerError
		if err.Error() == "wishlist item not found" {
			status = http.StatusNotFound
		}

		c.JSON(status, dto.ErrorResponse{
			Error:   "remove_failed",
			Message: err.Error(),
		})
		return
	}

	log.Printf("✅ RemoveFromWishlist success - Wishlist has %d items", wishlist.Count)
	c.JSON(http.StatusOK, wishlist)
}

// MoveToCart godoc
// @Summary Move wishlist item to cart
// @Description Move product from wishlist to cart (requires authentication)
// @Tags wishlist
// @Accept json
// @Produce json
// @Param productId path int true "Product ID"
// @Success 200 {object} dto.CartResponse
// @Router /api/wishlist/{productId}/move-to-cart [post]
func (h *WishlistHandler) MoveToCart(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Please login to move items to cart",
		})
		return
	}

	productIDStr := c.Param("productId")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid product ID",
		})
		return
	}

	sessionID := h.getOrCreateSessionID(c)
	log.Printf("🛒 MoveToCart - UserID: %d, ProductID: %d, SessionID: %s", *userID, productID, sessionID)

	cart, err := h.wishlistService.MoveToCart(*userID, productID, sessionID)
	if err != nil {
		log.Printf("❌ MoveToCart error: %v", err)
		status := http.StatusInternalServerError
		if err.Error() == "product not found" {
			status = http.StatusNotFound
		} else if err.Error() == "product is not available" {
			status = http.StatusBadRequest
		}

		c.JSON(status, dto.ErrorResponse{
			Error:   "move_to_cart_failed",
			Message: err.Error(),
		})
		return
	}

	log.Printf("✅ MoveToCart success - Cart has %d items", len(cart.Items))
	c.JSON(http.StatusOK, cart)
}
