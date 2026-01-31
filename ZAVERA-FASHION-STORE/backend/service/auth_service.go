package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/smtp"
	"os"
	"strconv"
	"time"
	"zavera/dto"
	"zavera/models"
	"zavera/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"
)

var (
	ErrUserExists          = errors.New("user with this email already exists")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrEmailNotVerified    = errors.New("email not verified")
	ErrInvalidToken        = errors.New("invalid or expired token")
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidGoogleToken  = errors.New("invalid Google token")
)

type AuthService interface {
	Register(req dto.RegisterRequest) (*models.User, error)
	Login(req dto.LoginRequest) (*dto.AuthResponse, error)
	GoogleLogin(req dto.GoogleLoginRequest) (*dto.AuthResponse, error)
	VerifyEmail(token string) error
	ResendVerification(email string) error
	GetUserByID(userID int) (*models.User, error)
	GetUserOrders(userID int, page, pageSize int) (*dto.UserOrdersResponse, error)
	GenerateJWT(user *models.User) (string, error)
	ValidateJWT(tokenString string) (*jwt.MapClaims, error)
	UserExists(userID int) (bool, error)
}

type authService struct {
	userRepo     repository.UserRepository
	shippingRepo repository.ShippingRepository
}

func NewAuthService(userRepo repository.UserRepository, shippingRepo ...repository.ShippingRepository) AuthService {
	svc := &authService{userRepo: userRepo}
	if len(shippingRepo) > 0 {
		svc.shippingRepo = shippingRepo[0]
	}
	return svc
}

func (s *authService) Register(req dto.RegisterRequest) (*models.User, error) {
	// Check if user exists
	existingUser, err := s.userRepo.FindByEmail(req.Email)
	if err == nil && existingUser != nil {
		return nil, ErrUserExists
	}

	// Parse birthdate
	birthdate, err := time.Parse("2006-01-02", req.Birthdate)
	if err != nil {
		return nil, fmt.Errorf("invalid birthdate format, use YYYY-MM-DD")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		Email:        req.Email,
		FirstName:    req.FirstName,
		Name:         req.FirstName,
		PasswordHash: string(hashedPassword),
		Birthdate:    &birthdate,
		IsVerified:   false,
		AuthProvider: "local",
	}

	err = s.userRepo.Create(user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Send notification to admin
	NotifyUserRegistered(user.Email, user.FirstName)

	// Generate verification token and send email
	err = s.sendVerificationEmail(user)
	if err != nil {
		// Log error but don't fail registration
		fmt.Printf("Failed to send verification email: %v\n", err)
	}

	return user, nil
}

func (s *authService) Login(req dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Check if email is verified
	if !user.IsVerified {
		return nil, ErrEmailNotVerified
	}

	// Generate JWT
	token, err := s.GenerateJWT(user)
	if err != nil {
		return nil, err
	}

	// Send notification to admin (skip for admin login)
	adminEmail := os.Getenv("ADMIN_GOOGLE_EMAIL")
	if adminEmail == "" {
		adminEmail = "pemberani073@gmail.com"
	}
	if user.Email != adminEmail {
		NotifyUserLogin(user.Email, user.FirstName)
	}

	return &dto.AuthResponse{
		User:        s.toUserResponse(user),
		AccessToken: token,
	}, nil
}


func (s *authService) GoogleLogin(req dto.GoogleLoginRequest) (*dto.AuthResponse, error) {
	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	
	// Validate Google ID token
	payload, err := idtoken.Validate(context.Background(), req.IDToken, googleClientID)
	if err != nil {
		fmt.Printf("Google token validation failed: %v\n", err)
		fmt.Printf("Client ID used: %s\n", googleClientID)
		return nil, fmt.Errorf("invalid Google token: %w", err)
	}

	email, _ := payload.Claims["email"].(string)
	name, _ := payload.Claims["name"].(string)
	googleID := payload.Subject

	fmt.Printf("Google login - Email: %s, Name: %s, GoogleID: %s\n", email, name, googleID)

	if email == "" {
		return nil, errors.New("email not found in Google token")
	}

	// Check if user exists by Google ID
	user, err := s.userRepo.FindByGoogleID(googleID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if user == nil {
		// Check if user exists by email
		user, err = s.userRepo.FindByEmail(email)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}

		if user == nil {
			// Create new user
			gid := googleID
			user = &models.User{
				Email:        email,
				FirstName:    name,
				Name:         name,
				GoogleID:     &gid,
				IsVerified:   true,
				AuthProvider: "google",
			}
			err = s.userRepo.Create(user)
			if err != nil {
				return nil, fmt.Errorf("failed to create user: %w", err)
			}
		} else {
			// Link Google account to existing user
			gid := googleID
			user.GoogleID = &gid
			user.IsVerified = true
			s.userRepo.Update(user)
		}
	}

	// Generate JWT
	token, err := s.GenerateJWT(user)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		User:        s.toUserResponse(user),
		AccessToken: token,
	}, nil
}

func (s *authService) VerifyEmail(token string) error {
	// Find token
	verificationToken, err := s.userRepo.FindVerificationToken(token)
	if err != nil {
		fmt.Printf("Token not found in database: %s\n", token)
		return ErrInvalidToken
	}

	// Check if user is already verified (e.g., via Google login)
	user, err := s.userRepo.FindByID(verificationToken.UserID)
	if err == nil && user != nil && user.IsVerified {
		// User already verified, mark token as used and return success
		s.userRepo.MarkTokenAsUsed(verificationToken.ID)
		return nil // Return success since user is already verified
	}

	// Check if already used
	if verificationToken.UsedAt != nil {
		fmt.Printf("Token already used at: %v\n", verificationToken.UsedAt)
		return ErrInvalidToken
	}

	// Check if expired
	if time.Now().After(verificationToken.ExpiresAt) {
		fmt.Printf("Token expired at: %v, current time: %v\n", verificationToken.ExpiresAt, time.Now())
		return ErrInvalidToken
	}

	// Mark user as verified
	err = s.userRepo.MarkAsVerified(verificationToken.UserID)
	if err != nil {
		return err
	}

	// Mark token as used
	s.userRepo.MarkTokenAsUsed(verificationToken.ID)

	return nil
}

func (s *authService) ResendVerification(email string) error {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return ErrUserNotFound
	}

	if user.IsVerified {
		return errors.New("email already verified")
	}

	// Delete old tokens
	s.userRepo.DeleteExpiredTokens(user.ID)

	// Send new verification email
	return s.sendVerificationEmail(user)
}

func (s *authService) GetUserByID(userID int) (*models.User, error) {
	return s.userRepo.FindByID(userID)
}

// UserExists checks if a user exists in the database by ID
func (s *authService) UserExists(userID int) (bool, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return false, nil // User not found, return false without error
	}
	return user != nil, nil
}

func (s *authService) GetUserOrders(userID int, page, pageSize int) (*dto.UserOrdersResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	orders, totalCount, err := s.userRepo.FindOrdersByUserID(userID, page, pageSize)
	if err != nil {
		return nil, err
	}

	var orderResponses []dto.OrderResponse
	for _, order := range orders {
		orderResp := s.toOrderResponse(&order)
		
		// Get shipment info if available
		if s.shippingRepo != nil {
			shipment, err := s.shippingRepo.GetShipmentByOrderID(order.ID)
			if err == nil && shipment != nil {
				orderResp.Shipment = &dto.ShipmentResponse{
					ID:              shipment.ID,
					OrderID:         shipment.OrderID,
					ProviderCode:    shipment.ProviderCode,
					ProviderName:    shipment.ProviderName,
					ServiceCode:     shipment.ServiceCode,
					ServiceName:     shipment.ServiceName,
					Cost:            shipment.Cost,
					ETD:             shipment.ETD,
					Status:          string(shipment.Status),
					TrackingNumber:  shipment.TrackingNumber,
				}
			}
		}
		
		orderResponses = append(orderResponses, orderResp)
	}

	return &dto.UserOrdersResponse{
		Orders:     orderResponses,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

func (s *authService) GenerateJWT(user *models.User) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", errors.New("JWT_SECRET environment variable is required")
	}

	expiryHours := 24
	if envExpiry := os.Getenv("JWT_EXPIRY_HOURS"); envExpiry != "" {
		if parsed, err := strconv.Atoi(envExpiry); err == nil {
			expiryHours = parsed
		}
	}

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * time.Duration(expiryHours)).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func (s *authService) ValidateJWT(tokenString string) (*jwt.MapClaims, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, errors.New("JWT_SECRET environment variable is required")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &claims, nil
	}

	return nil, errors.New("invalid token")
}


// Helper functions
func (s *authService) sendVerificationEmail(user *models.User) error {
	// Generate token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return err
	}
	token := hex.EncodeToString(tokenBytes)

	// Save token
	verificationToken := &models.EmailVerificationToken{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	err := s.userRepo.CreateVerificationToken(verificationToken)
	if err != nil {
		return err
	}

	// Build verification URL
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	verifyURL := fmt.Sprintf("%s/verify-email?token=%s", frontendURL, token)

	// Send email
	return s.sendEmail(user.Email, "Verifikasi Email ZAVERA", s.buildVerificationEmailBody(user.FirstName, verifyURL))
}

func (s *authService) sendEmail(to, subject, body string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USERNAME")
	smtpPass := os.Getenv("SMTP_PASSWORD")
	smtpFrom := os.Getenv("SMTP_FROM")

	if smtpHost == "" || smtpUser == "" || smtpPass == "" {
		return errors.New("SMTP not configured")
	}

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)

	msg := fmt.Sprintf("From: ZAVERA <%s>\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s", smtpFrom, to, subject, body)

	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	return smtp.SendMail(addr, auth, smtpFrom, []string{to}, []byte(msg))
}

func (s *authService) buildVerificationEmailBody(name, verifyURL string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin: 0; padding: 0; font-family: 'Helvetica Neue', Arial, sans-serif; background-color: #f5f5f5;">
    <table width="100%%" cellpadding="0" cellspacing="0" style="background-color: #f5f5f5; padding: 40px 20px;">
        <tr>
            <td align="center">
                <table width="600" cellpadding="0" cellspacing="0" style="background-color: #ffffff; border-radius: 8px; overflow: hidden; box-shadow: 0 4px 6px rgba(0,0,0,0.1);">
                    <!-- Header -->
                    <tr>
                        <td style="background-color: #1a1a1a; padding: 30px; text-align: center;">
                            <h1 style="color: #ffffff; margin: 0; font-size: 28px; letter-spacing: 4px;">ZAVERA</h1>
                        </td>
                    </tr>
                    <!-- Content -->
                    <tr>
                        <td style="padding: 40px 30px;">
                            <h2 style="color: #1a1a1a; margin: 0 0 20px; font-size: 24px;">Halo, %s!</h2>
                            <p style="color: #666666; font-size: 16px; line-height: 1.6; margin: 0 0 30px;">
                                Terima kasih telah mendaftar di ZAVERA. Untuk menyelesaikan pendaftaran, silakan verifikasi email Anda dengan mengklik tombol di bawah ini.
                            </p>
                            <table width="100%%" cellpadding="0" cellspacing="0">
                                <tr>
                                    <td align="center">
                                        <a href="%s" style="display: inline-block; background-color: #1a1a1a; color: #ffffff; text-decoration: none; padding: 15px 40px; border-radius: 4px; font-size: 16px; font-weight: 500; letter-spacing: 1px;">VERIFIKASI EMAIL</a>
                                    </td>
                                </tr>
                            </table>
                            <p style="color: #999999; font-size: 14px; line-height: 1.6; margin: 30px 0 0;">
                                Link ini akan kadaluarsa dalam 24 jam. Jika Anda tidak mendaftar di ZAVERA, abaikan email ini.
                            </p>
                        </td>
                    </tr>
                    <!-- Footer -->
                    <tr>
                        <td style="background-color: #f9f9f9; padding: 20px 30px; text-align: center; border-top: 1px solid #eeeeee;">
                            <p style="color: #999999; font-size: 12px; margin: 0;">
                                © 2024 ZAVERA. All rights reserved.
                            </p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>
`, name, verifyURL)
}

func (s *authService) toUserResponse(user *models.User) dto.UserResponse {
	var birthdate *string
	if user.Birthdate != nil {
		bd := user.Birthdate.Format("2006-01-02")
		birthdate = &bd
	}

	return dto.UserResponse{
		ID:           user.ID,
		Email:        user.Email,
		FirstName:    user.FirstName,
		Name:         user.Name,
		Phone:        user.Phone,
		Birthdate:    birthdate,
		IsVerified:   user.IsVerified,
		AuthProvider: user.AuthProvider,
		CreatedAt:    user.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (s *authService) toOrderResponse(order *models.Order) dto.OrderResponse {
	response := dto.OrderResponse{
		ID:            order.ID,
		OrderCode:     order.OrderCode,
		CustomerName:  order.CustomerName,
		CustomerEmail: order.CustomerEmail,
		CustomerPhone: order.CustomerPhone,
		Subtotal:      order.Subtotal,
		ShippingCost:  order.ShippingCost,
		Tax:           order.Tax,
		Discount:      order.Discount,
		TotalAmount:   order.TotalAmount,
		Status:        string(order.Status),
		Resi:          order.Resi,
		Items:         []dto.OrderItemResponse{},
		CreatedAt:     order.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	// Add timestamps
	if order.PaidAt != nil {
		response.PaidAt = order.PaidAt.Format("2006-01-02 15:04:05")
	}
	if order.ShippedAt != nil {
		response.ShippedAt = order.ShippedAt.Format("2006-01-02 15:04:05")
	}
	if order.DeliveredAt != nil {
		response.DeliveredAt = order.DeliveredAt.Format("2006-01-02 15:04:05")
	}

	for _, item := range order.Items {
		itemResponse := dto.OrderItemResponse{
			ProductID:    item.ProductID,
			ProductName:  item.ProductName,
			ProductImage: item.ProductImage,
			Quantity:     item.Quantity,
			PricePerUnit: item.PricePerUnit,
			Subtotal:     item.Subtotal,
		}
		response.Items = append(response.Items, itemResponse)
	}

	return response
}
