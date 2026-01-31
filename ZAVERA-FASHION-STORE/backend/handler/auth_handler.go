package handler

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"zavera/dto"
	"zavera/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService service.AuthService
	cartService service.CartService
}

func NewAuthHandler(authService service.AuthService, cartService service.CartService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		cartService: cartService,
	}
}

// Register godoc
// @Summary Register new user
// @Description Register a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Register request"
// @Success 201 {object} map[string]interface{}
// @Router /api/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	user, err := h.authService.Register(req)
	if err != nil {
		status := http.StatusInternalServerError
		if err == service.ErrUserExists {
			status = http.StatusConflict
		}
		c.JSON(status, dto.ErrorResponse{
			Error:   "registration_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Registration successful. Please check your email to verify your account.",
		"user_id": user.ID,
	})
}

// Login godoc
// @Summary Login user
// @Description Login with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login request"
// @Success 200 {object} dto.AuthResponse
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	response, err := h.authService.Login(req)
	if err != nil {
		status := http.StatusUnauthorized
		errorCode := "login_failed"

		if err == service.ErrEmailNotVerified {
			status = http.StatusForbidden
			errorCode = "email_not_verified"
		}

		c.JSON(status, dto.ErrorResponse{
			Error:   errorCode,
			Message: err.Error(),
		})
		return
	}

	// Link guest cart to user account on login
	sessionID, _ := c.Cookie("session_id")
	if sessionID != "" && h.cartService != nil {
		h.cartService.LinkCartToUser(sessionID, response.User.ID)
	}

	c.JSON(http.StatusOK, response)
}

// GoogleLogin godoc
// @Summary Login with Google
// @Description Login or register with Google OAuth
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.GoogleLoginRequest true "Google login request"
// @Success 200 {object} dto.AuthResponse
// @Router /api/auth/google [post]
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	var req dto.GoogleLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	response, err := h.authService.GoogleLogin(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "google_login_failed",
			Message: err.Error(),
		})
		return
	}

	// Link guest cart to user account on login
	sessionID, _ := c.Cookie("session_id")
	if sessionID != "" && h.cartService != nil {
		h.cartService.LinkCartToUser(sessionID, response.User.ID)
	}

	c.JSON(http.StatusOK, response)
}

// VerifyEmail godoc
// @Summary Verify email
// @Description Verify user email with token
// @Tags auth
// @Accept json
// @Produce json
// @Param token query string true "Verification token"
// @Success 200 {object} map[string]string
// @Router /api/auth/verify-email [get]
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Token is required",
		})
		return
	}

	err := h.authService.VerifyEmail(token)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "verification_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Email verified successfully. You can now login.",
	})
}

// ResendVerification godoc
// @Summary Resend verification email
// @Description Resend email verification link
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.ResendVerificationRequest true "Resend request"
// @Success 200 {object} map[string]string
// @Router /api/auth/resend-verification [post]
func (h *AuthHandler) ResendVerification(c *gin.Context) {
	var req dto.ResendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	err := h.authService.ResendVerification(req.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "resend_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Verification email sent successfully.",
	})
}

// GetMe godoc
// @Summary Get current user
// @Description Get current authenticated user info
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.UserResponse
// @Router /api/auth/me [get]
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	user, err := h.authService.GetUserByID(int(userID.(float64)))
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Message: "User not found",
		})
		return
	}

	var birthdate *string
	if user.Birthdate != nil {
		bd := user.Birthdate.Format("2006-01-02")
		birthdate = &bd
	}

	c.JSON(http.StatusOK, dto.UserResponse{
		ID:           user.ID,
		Email:        user.Email,
		FirstName:    user.FirstName,
		Name:         user.Name,
		Phone:        user.Phone,
		Birthdate:    birthdate,
		IsVerified:   user.IsVerified,
		AuthProvider: user.AuthProvider,
		CreatedAt:    user.CreatedAt.Format("2006-01-02 15:04:05"),
	})
}

// GetUserOrders godoc
// @Summary Get user orders
// @Description Get order history for authenticated user
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} dto.UserOrdersResponse
// @Router /api/user/orders [get]
func (h *AuthHandler) GetUserOrders(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	response, err := h.authService.GetUserOrders(int(userID.(float64)), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "fetch_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// AuthMiddleware validates JWT token and verifies user exists in database
func (h *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "unauthorized",
				Message: "Authorization header required",
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		claims, err := h.authService.ValidateJWT(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Extract user_id from claims
		userIDFloat, ok := (*claims)["user_id"].(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid token claims",
			})
			c.Abort()
			return
		}
		userID := int(userIDFloat)

		// Verify user still exists in database
		exists, err := h.authService.UserExists(userID)
		if err != nil || !exists {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "session_expired",
				Message: "Sesi Anda telah berakhir. Silakan login kembali.",
			})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", (*claims)["user_id"])
		c.Set("email", (*claims)["email"])
		c.Next()
	}
}


// OptionalAuthMiddleware validates JWT token if present, but doesn't require it
// Also verifies user exists in database
func (h *AuthHandler) OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		claims, err := h.authService.ValidateJWT(parts[1])
		if err != nil {
			c.Next()
			return
		}

		// Extract user_id and verify user exists
		userIDFloat, ok := (*claims)["user_id"].(float64)
		if !ok {
			c.Next()
			return
		}
		userID := int(userIDFloat)

		// Verify user still exists in database
		exists, err := h.authService.UserExists(userID)
		if err != nil || !exists {
			// User doesn't exist, treat as unauthenticated (don't set user context)
			c.Next()
			return
		}

		// Set user info in context
		c.Set("user_id", (*claims)["user_id"])
		c.Set("email", (*claims)["email"])
		c.Next()
	}
}

// AdminMiddleware checks if user is admin (Google-locked)
func (h *AuthHandler) AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		email, exists := c.Get("email")
		if !exists {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "unauthorized",
				Message: "User not authenticated",
			})
			c.Abort()
			return
		}

		// Get admin email from env (default to pemberani073@gmail.com)
		adminEmail := os.Getenv("ADMIN_GOOGLE_EMAIL")
		if adminEmail == "" {
			adminEmail = "pemberani073@gmail.com"
		}

		// Check if user is admin
		userEmail, ok := email.(string)
		if !ok || userEmail != adminEmail {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Error:   "forbidden",
				Message: "Admin access required. Only authorized administrators can access this resource.",
			})
			c.Abort()
			return
		}

		// Set admin flag
		c.Set("is_admin", true)
		c.Set("user_email", userEmail)
		c.Next()
	}
}
