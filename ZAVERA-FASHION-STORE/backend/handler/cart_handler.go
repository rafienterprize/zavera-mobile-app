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

type CartHandler struct {
	cartService service.CartService
}

func NewCartHandler(cartService service.CartService) *CartHandler {
	return &CartHandler{
		cartService: cartService,
	}
}

// getOrCreateSessionID gets session ID from cookie or creates new one
func (h *CartHandler) getOrCreateSessionID(c *gin.Context) string {
	sessionID, err := c.Cookie("session_id")
	if err != nil || sessionID == "" {
		sessionID = uuid.New().String()
		c.SetCookie("session_id", sessionID, 86400*30, "/", "", false, true)
	}
	return sessionID
}

// getUserID gets user ID from context if logged in
func (h *CartHandler) getUserID(c *gin.Context) *int {
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

// GetCart godoc
// @Summary Get user's cart
// @Description Get current cart with all items (persisted for logged-in users)
// @Tags cart
// @Accept json
// @Produce json
// @Success 200 {object} dto.CartResponse
// @Router /api/cart [get]
func (h *CartHandler) GetCart(c *gin.Context) {
	sessionID := h.getOrCreateSessionID(c)
	userID := h.getUserID(c)

	var cart *dto.CartResponse
	var err error

	if userID != nil {
		// Logged-in user: get cart linked to user account
		cart, err = h.cartService.GetCartForUser(*userID, sessionID)
	} else {
		// Guest: get cart by session
		cart, err = h.cartService.GetCart(sessionID)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, cart)
}

// AddToCart godoc
// @Summary Add item to cart
// @Description Add product to cart with quantity (persisted for logged-in users)
// @Tags cart
// @Accept json
// @Produce json
// @Param request body dto.AddToCartRequest true "Add to cart request"
// @Success 200 {object} dto.CartResponse
// @Router /api/cart/items [post]
func (h *CartHandler) AddToCart(c *gin.Context) {
	sessionID := h.getOrCreateSessionID(c)
	userID := h.getUserID(c)
	
	// Debug logging
	if userID != nil {
		log.Printf("🛒 AddToCart - SessionID: %s, UserID: %d", sessionID, *userID)
	} else {
		log.Printf("🛒 AddToCart - SessionID: %s, UserID: nil (guest)", sessionID)
	}

	var req dto.AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}
	
	log.Printf("🛒 AddToCart - ProductID: %d, Quantity: %d", req.ProductID, req.Quantity)

	var cart *dto.CartResponse
	var err error

	if userID != nil {
		// Logged-in user: add to user's cart
		cart, err = h.cartService.AddToCartForUser(*userID, sessionID, req)
	} else {
		// Guest: add to session cart
		cart, err = h.cartService.AddToCart(sessionID, req)
	}

	if err != nil {
		log.Printf("❌ AddToCart error: %v", err)
		status := http.StatusInternalServerError
		if err.Error() == "product not found" {
			status = http.StatusNotFound
		} else if err.Error() == "insufficient stock" {
			status = http.StatusBadRequest
		}

		c.JSON(status, dto.ErrorResponse{
			Error:   "add_to_cart_failed",
			Message: err.Error(),
		})
		return
	}
	
	log.Printf("✅ AddToCart success - Cart has %d items", len(cart.Items))

	c.JSON(http.StatusOK, cart)
}

// UpdateCartItem godoc
// @Summary Update cart item quantity
// @Description Update quantity of item in cart
// @Tags cart
// @Accept json
// @Produce json
// @Param id path int true "Cart Item ID"
// @Param request body dto.UpdateCartItemRequest true "Update request"
// @Success 200 {object} dto.CartResponse
// @Router /api/cart/items/{id} [put]
func (h *CartHandler) UpdateCartItem(c *gin.Context) {
	sessionID := h.getOrCreateSessionID(c)
	userID := h.getUserID(c)

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid item ID",
		})
		return
	}

	var req dto.UpdateCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	var cart *dto.CartResponse

	if userID != nil {
		cart, err = h.cartService.UpdateCartItemForUser(*userID, sessionID, id, req.Quantity)
	} else {
		cart, err = h.cartService.UpdateCartItem(sessionID, id, req.Quantity)
	}

	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "cart item not found" || err.Error() == "cart not found" {
			status = http.StatusNotFound
		} else if err.Error() == "insufficient stock" {
			status = http.StatusBadRequest
		}

		c.JSON(status, dto.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, cart)
}

// RemoveFromCart godoc
// @Summary Remove item from cart
// @Description Remove item from cart
// @Tags cart
// @Accept json
// @Produce json
// @Param id path int true "Cart Item ID"
// @Success 200 {object} dto.CartResponse
// @Router /api/cart/items/{id} [delete]
func (h *CartHandler) RemoveFromCart(c *gin.Context) {
	sessionID := h.getOrCreateSessionID(c)
	userID := h.getUserID(c)

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid item ID",
		})
		return
	}

	log.Printf("🗑️ RemoveFromCart - ItemID: %d, SessionID: %s, UserID: %v", id, sessionID, userID)

	// Verify item belongs to this user's/session's cart
	var cart *dto.CartResponse
	if userID != nil {
		cart, err = h.cartService.GetCartForUser(*userID, sessionID)
	} else {
		cart, err = h.cartService.GetCart(sessionID)
	}

	if err != nil {
		log.Printf("❌ RemoveFromCart - Failed to get cart: %v", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	// Check if item exists in this cart
	itemFound := false
	for _, item := range cart.Items {
		if item.ID == id {
			itemFound = true
			log.Printf("✅ RemoveFromCart - Item found in cart: ProductID=%d, Quantity=%d", item.ProductID, item.Quantity)
			break
		}
	}

	if !itemFound {
		log.Printf("❌ RemoveFromCart - Item %d not found in cart (cart has %d items)", id, len(cart.Items))
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "forbidden",
			Message: "Item does not belong to your cart",
		})
		return
	}

	if userID != nil {
		cart, err = h.cartService.RemoveFromCartForUser(*userID, sessionID, id)
	} else {
		cart, err = h.cartService.RemoveFromCart(sessionID, id)
	}

	if err != nil {
		log.Printf("❌ RemoveFromCart - Failed to remove item: %v", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "remove_failed",
			Message: err.Error(),
		})
		return
	}

	log.Printf("✅ RemoveFromCart - Success! Cart now has %d items", len(cart.Items))
	c.JSON(http.StatusOK, cart)
}

// ClearCart godoc
// @Summary Clear cart
// @Description Remove all items from cart
// @Tags cart
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/cart [delete]
func (h *CartHandler) ClearCart(c *gin.Context) {
	sessionID := h.getOrCreateSessionID(c)

	err := h.cartService.ClearCart(sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "clear_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cart cleared successfully"})
}

// ValidateCart godoc
// @Summary Validate cart items
// @Description Validate cart items against current product data (price, stock, availability)
// @Tags cart
// @Accept json
// @Produce json
// @Success 200 {object} dto.CartValidationResponse
// @Router /api/cart/validate [get]
func (h *CartHandler) ValidateCart(c *gin.Context) {
	sessionID := h.getOrCreateSessionID(c)
	userID := h.getUserID(c)

	if userID == nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Please login to validate cart",
		})
		return
	}

	validation, err := h.cartService.ValidateCart(*userID, sessionID)
	if err != nil {
		log.Printf("❌ ValidateCart error: %v", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "validation_failed",
			Message: err.Error(),
		})
		return
	}

	if !validation.Valid {
		log.Printf("⚠️ Cart validation found %d changes", len(validation.Changes))
	} else {
		log.Printf("✅ Cart validation passed")
	}

	c.JSON(http.StatusOK, validation)
}
