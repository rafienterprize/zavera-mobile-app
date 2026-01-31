package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
	"zavera/models"
	"zavera/repository"
)

var (
	ErrResiLocked       = errors.New("resi cannot be modified after order is shipped")
	ErrResiExists       = errors.New("resi already exists")
	ErrInvalidResiFormat = errors.New("invalid resi format")
)

// ResiService handles airway bill (resi) generation
type ResiService interface {
	// GenerateResi generates a unique resi for an order
	// Format: ZVR-{COURIER}-{YYYYMMDD}-{ORDERID}-{RANDOM4}
	GenerateResi(orderID int, courierCode string) (string, error)
	
	// ValidateResiFormat checks if a resi matches the expected format
	ValidateResiFormat(resi string) bool
	
	// IsResiLocked checks if the resi for an order cannot be modified
	IsResiLocked(order *models.Order) bool
	
	// ParseResi extracts components from a resi string
	ParseResi(resi string) (*ResiComponents, error)
}

// ResiComponents represents the parsed components of a resi
type ResiComponents struct {
	Prefix    string // ZVR
	Courier   string // JNE, JNT, etc.
	Date      string // YYYYMMDD
	OrderID   string // Order ID
	RandomHex string // 4 random hex characters
}

type resiService struct {
	orderRepo repository.OrderRepository
}

// NewResiService creates a new resi service
func NewResiService(orderRepo repository.OrderRepository) ResiService {
	return &resiService{
		orderRepo: orderRepo,
	}
}

// GenerateResi generates a unique resi for an order
// Format: ZVR-{COURIER}-{YYYYMMDD}-{ORDERID}-{RANDOM4}
// Example: ZVR-JNE-20260111-98312-A7KD
func (s *resiService) GenerateResi(orderID int, courierCode string) (string, error) {
	// Normalize courier code to uppercase
	courier := strings.ToUpper(courierCode)
	
	// Get current date
	dateStr := time.Now().Format("20060102")
	
	// Try up to 10 times to generate a unique resi
	for attempt := 0; attempt < 10; attempt++ {
		// Generate 2 random bytes = 4 hex characters
		randomBytes := make([]byte, 2)
		_, err := rand.Read(randomBytes)
		if err != nil {
			return "", fmt.Errorf("failed to generate random bytes: %w", err)
		}
		randomHex := strings.ToUpper(hex.EncodeToString(randomBytes))
		
		// Build resi string
		resi := fmt.Sprintf("ZVR-%s-%s-%d-%s", courier, dateStr, orderID, randomHex)
		
		// Check if resi already exists
		exists, err := s.orderRepo.IsResiExists(resi)
		if err != nil {
			return "", fmt.Errorf("failed to check resi existence: %w", err)
		}
		
		if !exists {
			return resi, nil
		}
	}
	
	return "", fmt.Errorf("failed to generate unique resi after 10 attempts")
}

// ValidateResiFormat checks if a resi matches the expected format
// Format: ZVR-{COURIER}-{YYYYMMDD}-{ORDERID}-{RANDOM4}
func (s *resiService) ValidateResiFormat(resi string) bool {
	parts := strings.Split(resi, "-")
	if len(parts) != 5 {
		return false
	}
	
	// Check prefix
	if parts[0] != "ZVR" {
		return false
	}
	
	// Check courier (should be uppercase letters)
	courier := parts[1]
	if len(courier) < 2 || len(courier) > 10 {
		return false
	}
	for _, c := range courier {
		if c < 'A' || c > 'Z' {
			return false
		}
	}
	
	// Check date format (YYYYMMDD)
	dateStr := parts[2]
	if len(dateStr) != 8 {
		return false
	}
	_, err := time.Parse("20060102", dateStr)
	if err != nil {
		return false
	}
	
	// Check order ID (should be numeric)
	orderIDStr := parts[3]
	for _, c := range orderIDStr {
		if c < '0' || c > '9' {
			return false
		}
	}
	
	// Check random hex (should be 4 uppercase hex characters)
	randomHex := parts[4]
	if len(randomHex) != 4 {
		return false
	}
	for _, c := range randomHex {
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	
	return true
}

// IsResiLocked checks if the resi for an order cannot be modified
func (s *resiService) IsResiLocked(order *models.Order) bool {
	return order.IsResiLocked()
}

// ParseResi extracts components from a resi string
func (s *resiService) ParseResi(resi string) (*ResiComponents, error) {
	if !s.ValidateResiFormat(resi) {
		return nil, ErrInvalidResiFormat
	}
	
	parts := strings.Split(resi, "-")
	return &ResiComponents{
		Prefix:    parts[0],
		Courier:   parts[1],
		Date:      parts[2],
		OrderID:   parts[3],
		RandomHex: parts[4],
	}, nil
}

// GetTrackingURL returns the tracking URL for a resi based on courier
func GetTrackingURL(courier, resi string) string {
	courier = strings.ToLower(courier)
	
	switch courier {
	case "jne":
		return fmt.Sprintf("https://www.jne.co.id/id/tracking/trace/%s", resi)
	case "jnt", "j&t":
		return fmt.Sprintf("https://www.jet.co.id/track/%s", resi)
	case "sicepat":
		return fmt.Sprintf("https://www.sicepat.com/checkAwb/%s", resi)
	case "pos":
		return fmt.Sprintf("https://www.posindonesia.co.id/id/tracking/%s", resi)
	case "tiki":
		return fmt.Sprintf("https://www.tiki.id/id/tracking/%s", resi)
	case "anteraja":
		return fmt.Sprintf("https://anteraja.id/tracking/%s", resi)
	case "ninja":
		return fmt.Sprintf("https://www.ninjaxpress.co/id-id/tracking?id=%s", resi)
	case "lion":
		return fmt.Sprintf("https://www.lionparcel.com/tracking/%s", resi)
	default:
		// Generic tracking URL
		return fmt.Sprintf("https://cekresi.com/?noresi=%s", resi)
	}
}
