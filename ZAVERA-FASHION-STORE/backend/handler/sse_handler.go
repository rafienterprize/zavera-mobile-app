package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"zavera/service"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID    int    `json:"user_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	jwt.StandardClaims
}

// SSEHandler handles Server-Sent Events for admin notifications
type SSEHandler struct {
	broker *service.SSEBroker
}

// NewSSEHandler creates a new SSE handler
func NewSSEHandler(broker *service.SSEBroker) *SSEHandler {
	return &SSEHandler{
		broker: broker,
	}
}

// HandleAdminSSE handles SSE connections for admin dashboard
// GET /api/admin/events
func HandleAdminSSE(c *gin.Context) {
	// Get broker instance
	broker := service.GetSSEBroker()
	
	// Validate JWT token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		log.Println("❌ SSE: No Authorization header")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized - no token"})
		return
	}
	
	// Extract token (format: "Bearer <token>")
	var token string
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	} else {
		log.Println("❌ SSE: Invalid Authorization header format")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized - invalid token format"})
		return
	}
	
	// Validate JWT token
	claims, err := validateJWTTokenSSE(token)
	if err != nil {
		log.Printf("❌ SSE: Invalid token: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized - invalid token"})
		return
	}
	
	// Check if user is admin
	adminEmail := os.Getenv("ADMIN_GOOGLE_EMAIL")
	if adminEmail == "" {
		adminEmail = "pemberani073@gmail.com"
	}
	
	if claims.Email != adminEmail {
		log.Printf("❌ SSE: User %s is not admin", claims.Email)
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden - admin only"})
		return
	}
	
	adminID := claims.UserID
	username := claims.FirstName
	if username == "" {
		username = "admin"
	}
	
	log.Printf("🔌 SSE: Admin %s (%d) connecting...", username, adminID)
	
	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // Disable nginx buffering
	
	// Create new SSE client
	client := broker.NewClient(adminID, username)
	
	// Register client with broker
	broker.RegisterClient(client)
	
	// Ensure client is unregistered when connection closes
	defer broker.UnregisterClient(client)
	
	// Get response writer flusher
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		log.Println("❌ SSE: Streaming not supported")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming not supported"})
		return
	}
	
	// Send initial connection event
	welcomeEvent := service.SSEEvent{
		ID:    fmt.Sprintf("welcome-%d", time.Now().Unix()),
		Event: "connected",
		Data: map[string]interface{}{
			"message":   "Connected to admin notification stream",
			"timestamp": time.Now().Format(time.RFC3339),
			"admin":     username,
		},
		Retry: 3000,
	}
	
	fmt.Fprint(c.Writer, service.FormatSSEMessage(welcomeEvent))
	flusher.Flush()
	
	log.Printf("✅ SSE: Admin %s connected successfully", username)
	
	// Send test notification to verify SSE is working
	go func() {
		time.Sleep(2 * time.Second)
		service.BroadcastNotification(service.AdminNotification{
			ID:       fmt.Sprintf("test-%d", time.Now().Unix()),
			Type:     "system",
			Title:    "🎉 SSE Connected",
			Message:  fmt.Sprintf("Welcome %s! Real-time notifications are now active.", username),
			Severity: service.SeverityInfo,
			Data:     map[string]interface{}{"admin": username},
			Timestamp: time.Now(),
			Read:     false,
		})
	}()
	
	// Create context with timeout for keepalive
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()
	
	// Keepalive ticker (every 30 seconds)
	keepaliveTicker := time.NewTicker(30 * time.Second)
	defer keepaliveTicker.Stop()
	
	// Event loop
	for {
		select {
		case <-ctx.Done():
			// Client disconnected
			log.Printf("🔌 SSE: Client %s disconnected (context done)", username)
			return
			
		case <-c.Request.Context().Done():
			// HTTP request context cancelled
			log.Printf("🔌 SSE: Client %s disconnected (request cancelled)", username)
			return
			
		case event, ok := <-client.Events:
			if !ok {
				// Channel closed
				log.Printf("🔌 SSE: Client %s channel closed", username)
				return
			}
			
			// Send event to client
			msg := service.FormatSSEMessage(event)
			fmt.Fprint(c.Writer, msg)
			flusher.Flush()
			
		case <-keepaliveTicker.C:
			// Send keepalive comment
			fmt.Fprint(c.Writer, service.SendKeepAlive())
			flusher.Flush()
		}
	}
}

// validateJWTTokenSSE validates JWT token and returns claims
func validateJWTTokenSSE(tokenString string) (*JWTClaims, error) {
	// Get JWT secret from env
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-this-in-production"
	}
	
	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	
	if err != nil {
		return nil, err
	}
	
	// Extract claims
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}
	
	return nil, fmt.Errorf("invalid token claims")
}
