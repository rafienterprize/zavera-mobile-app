package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
	"zavera/models"
)

var (
	ErrMidtransTimeout     = errors.New("midtrans API timeout")
	ErrMidtransAPIError    = errors.New("midtrans API error")
	ErrInvalidPaymentType  = errors.New("invalid payment type")
	ErrVANumberNotFound    = errors.New("VA number not found in response")
)

// MidtransCoreClient defines the interface for Midtrans Core API operations
type MidtransCoreClient interface {
	// ChargeVA creates a VA payment via Midtrans Core API /v2/charge
	ChargeVA(request ChargeVARequest) (*ChargeVAResponse, error)
	
	// GetTransactionStatus gets transaction status from Midtrans API
	// GET /v2/{order_id}/status
	GetTransactionStatus(orderID string) (*TransactionStatusResponse, error)
}

// TransactionStatusResponse represents the response from Midtrans status check
type TransactionStatusResponse struct {
	StatusCode        string `json:"status_code"`
	StatusMessage     string `json:"status_message"`
	TransactionID     string `json:"transaction_id"`
	OrderID           string `json:"order_id"`
	GrossAmount       string `json:"gross_amount"`
	PaymentType       string `json:"payment_type"`
	TransactionTime   string `json:"transaction_time"`
	TransactionStatus string `json:"transaction_status"`
	FraudStatus       string `json:"fraud_status,omitempty"`
	SettlementTime    string `json:"settlement_time,omitempty"`
}

// ChargeVARequest represents the request to create a VA payment
type ChargeVARequest struct {
	OrderID         string  `json:"order_id"`
	GrossAmount     float64 `json:"gross_amount"`
	PaymentMethod   models.VAPaymentMethod
	CustomerName    string
	CustomerEmail   string
	CustomerPhone   string
}

// ChargeVAResponse represents the response from Midtrans VA charge
type ChargeVAResponse struct {
	TransactionID   string                 `json:"transaction_id"`
	OrderID         string                 `json:"order_id"`
	GrossAmount     string                 `json:"gross_amount"`
	PaymentType     string                 `json:"payment_type"`
	TransactionTime string                 `json:"transaction_time"`
	TransactionStatus string              `json:"transaction_status"`
	VANumber        string                 `json:"va_number"`
	Bank            string                 `json:"bank"`
	ExpiryTime      time.Time              `json:"expiry_time"`
	RawResponse     map[string]interface{} `json:"raw_response"`
	// GoPay specific fields
	QRCodeURL       string                 `json:"qr_code_url,omitempty"`
	DeeplinkURL     string                 `json:"deeplink_url,omitempty"`
}

// MidtransChargeRequest is the actual request body sent to Midtrans
type MidtransChargeRequest struct {
	PaymentType        string                 `json:"payment_type"`
	TransactionDetails TransactionDetails     `json:"transaction_details"`
	CustomerDetails    *CustomerDetails       `json:"customer_details,omitempty"`
	BankTransfer       *BankTransferDetails   `json:"bank_transfer,omitempty"`
	EChannel           *EChannelDetails       `json:"echannel,omitempty"`
	QRIS               *QRISDetails           `json:"qris,omitempty"`
	GoPay              *GoPayDetails          `json:"gopay,omitempty"`
}

type QRISDetails struct {
	Acquirer string `json:"acquirer"`
}

type GoPayDetails struct {
	EnableCallback     bool   `json:"enable_callback"`
	CallbackURL        string `json:"callback_url,omitempty"`
}

type TransactionDetails struct {
	OrderID     string `json:"order_id"`
	GrossAmount int64  `json:"gross_amount"`
}

type CustomerDetails struct {
	FirstName string `json:"first_name,omitempty"`
	Email     string `json:"email,omitempty"`
	Phone     string `json:"phone,omitempty"`
}

type BankTransferDetails struct {
	Bank string `json:"bank"`
}

type EChannelDetails struct {
	BillInfo1 string `json:"bill_info1"`
	BillInfo2 string `json:"bill_info2"`
}

// MidtransChargeResponse is the response from Midtrans /v2/charge
type MidtransChargeResponse struct {
	StatusCode        string      `json:"status_code"`
	StatusMessage     string      `json:"status_message"`
	TransactionID     string      `json:"transaction_id"`
	OrderID           string      `json:"order_id"`
	GrossAmount       string      `json:"gross_amount"`
	PaymentType       string      `json:"payment_type"`
	TransactionTime   string      `json:"transaction_time"`
	TransactionStatus string      `json:"transaction_status"`
	FraudStatus       string      `json:"fraud_status,omitempty"`
	ExpiryTime        string      `json:"expiry_time,omitempty"`
	
	// VA-specific fields
	VANumbers         []VANumber  `json:"va_numbers,omitempty"`
	PermataVANumber   string      `json:"permata_va_number,omitempty"`
	BillKey           string      `json:"bill_key,omitempty"`
	BillerCode        string      `json:"biller_code,omitempty"`
	
	// QRIS-specific fields
	Actions           []QRISAction `json:"actions,omitempty"`
}

type VANumber struct {
	Bank     string `json:"bank"`
	VANumber string `json:"va_number"`
}

type QRISAction struct {
	Name   string `json:"name"`
	Method string `json:"method"`
	URL    string `json:"url"`
}

// GoPayAction represents GoPay action URLs (QR code, deeplink)
type GoPayAction struct {
	Name   string `json:"name"`
	Method string `json:"method"`
	URL    string `json:"url"`
}

type midtransCoreClient struct {
	serverKey  string
	baseURL    string
	httpClient *http.Client
}

// NewMidtransCoreClient creates a new Midtrans Core API client
func NewMidtransCoreClient() MidtransCoreClient {
	serverKey := os.Getenv("MIDTRANS_SERVER_KEY")
	if serverKey == "" {
		log.Println("⚠️ MIDTRANS_SERVER_KEY not set, using empty key")
	}

	baseURL := "https://api.sandbox.midtrans.com"
	if os.Getenv("MIDTRANS_ENVIRONMENT") == "production" {
		baseURL = "https://api.midtrans.com"
	}

	return &midtransCoreClient{
		serverKey: serverKey,
		baseURL:   baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second, // 30 second timeout as per requirements
		},
	}
}

// ChargeVA creates a VA payment via Midtrans Core API
func (c *midtransCoreClient) ChargeVA(request ChargeVARequest) (*ChargeVAResponse, error) {
	log.Printf("🔄 Midtrans ChargeVA: order_id=%s, method=%s, amount=%.2f",
		request.OrderID, request.PaymentMethod, request.GrossAmount)

	// Build the charge request based on payment method
	chargeReq, err := c.buildChargeRequest(request)
	if err != nil {
		return nil, err
	}

	// Marshal request body
	reqBody, err := json.Marshal(chargeReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	log.Printf("📤 Midtrans request: %s", string(reqBody))

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", c.baseURL+"/v2/charge", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(c.serverKey+":")))

	// Execute request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		if os.IsTimeout(err) {
			log.Printf("❌ Midtrans timeout: %v", err)
			return nil, ErrMidtransTimeout
		}
		log.Printf("❌ Midtrans network error: %v", err)
		return nil, fmt.Errorf("midtrans network error: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	log.Printf("📥 Midtrans response: %s", string(respBody))

	// Parse response
	var midtransResp MidtransChargeResponse
	if err := json.Unmarshal(respBody, &midtransResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for error status
	if midtransResp.StatusCode != "201" && midtransResp.StatusCode != "200" {
		log.Printf("❌ Midtrans error: %s - %s", midtransResp.StatusCode, midtransResp.StatusMessage)
		
		// Provide more helpful error messages
		var userMessage string
		switch midtransResp.StatusCode {
		case "500":
			userMessage = "Midtrans sedang mengalami gangguan. Silakan coba lagi dalam beberapa menit atau gunakan metode pembayaran lain."
		case "400":
			userMessage = "Format pembayaran tidak valid. Silakan hubungi customer service."
		case "401":
			userMessage = "Konfigurasi pembayaran bermasalah. Silakan hubungi customer service."
		case "404":
			userMessage = "Metode pembayaran tidak tersedia. Silakan pilih metode lain."
		default:
			userMessage = fmt.Sprintf("Gagal membuat pembayaran: %s", midtransResp.StatusMessage)
		}
		
		return nil, fmt.Errorf("%s", userMessage)
	}

	// Extract VA number based on bank
	vaNumber, bank, err := c.extractVANumber(request.PaymentMethod, &midtransResp)
	if err != nil {
		return nil, err
	}

	// Parse expiry time
	expiryTime, err := c.parseExpiryTime(midtransResp.ExpiryTime)
	if err != nil {
		// Default to 24 hours if parsing fails
		expiryTime = time.Now().Add(24 * time.Hour)
		log.Printf("⚠️ Failed to parse expiry time, using default 24h: %v", err)
	}

	// Build raw response map for audit
	var rawResponse map[string]interface{}
	json.Unmarshal(respBody, &rawResponse)

	result := &ChargeVAResponse{
		TransactionID:     midtransResp.TransactionID,
		OrderID:           midtransResp.OrderID,
		GrossAmount:       midtransResp.GrossAmount,
		PaymentType:       midtransResp.PaymentType,
		TransactionTime:   midtransResp.TransactionTime,
		TransactionStatus: midtransResp.TransactionStatus,
		VANumber:          vaNumber,
		Bank:              bank,
		ExpiryTime:        expiryTime,
		RawResponse:       rawResponse,
	}

	// Extract GoPay/QRIS specific URLs from actions
	if request.PaymentMethod == models.VAPaymentMethodGoPay || request.PaymentMethod == models.VAPaymentMethodQRIS {
		for _, action := range midtransResp.Actions {
			if action.Name == "generate-qr-code" {
				result.QRCodeURL = action.URL
				// For QRIS, also store QR URL as va_number for display
				if request.PaymentMethod == models.VAPaymentMethodQRIS {
					result.VANumber = midtransResp.TransactionID // Use transaction ID as reference
				}
			}
			if action.Name == "deeplink-redirect" {
				result.DeeplinkURL = action.URL
			}
		}
	}

	log.Printf("✅ Midtrans ChargeVA success: tx_id=%s, va=%s, bank=%s, expiry=%s",
		result.TransactionID, result.VANumber, result.Bank, result.ExpiryTime)

	return result, nil
}

// buildChargeRequest builds the Midtrans charge request based on payment method
func (c *midtransCoreClient) buildChargeRequest(request ChargeVARequest) (*MidtransChargeRequest, error) {
	chargeReq := &MidtransChargeRequest{
		TransactionDetails: TransactionDetails{
			OrderID:     request.OrderID,
			GrossAmount: int64(request.GrossAmount),
		},
		CustomerDetails: &CustomerDetails{
			FirstName: request.CustomerName,
			Email:     request.CustomerEmail,
			Phone:     request.CustomerPhone,
		},
	}

	switch request.PaymentMethod {
	case models.VAPaymentMethodBCA:
		chargeReq.PaymentType = "bank_transfer"
		chargeReq.BankTransfer = &BankTransferDetails{Bank: "bca"}

	case models.VAPaymentMethodBRI:
		chargeReq.PaymentType = "bank_transfer"
		chargeReq.BankTransfer = &BankTransferDetails{Bank: "bri"}

	case models.VAPaymentMethodBNI:
		chargeReq.PaymentType = "bank_transfer"
		chargeReq.BankTransfer = &BankTransferDetails{Bank: "bni"}

	case models.VAPaymentMethodPermata:
		chargeReq.PaymentType = "bank_transfer"
		chargeReq.BankTransfer = &BankTransferDetails{Bank: "permata"}

	case models.VAPaymentMethodMandiri:
		// Mandiri uses echannel (bill payment)
		chargeReq.PaymentType = "echannel"
		chargeReq.EChannel = &EChannelDetails{
			BillInfo1: "Payment for",
			BillInfo2: request.OrderID,
		}

	case models.VAPaymentMethodQRIS:
		chargeReq.PaymentType = "qris"
		chargeReq.QRIS = &QRISDetails{
			Acquirer: "gopay", // Default acquirer for QRIS
		}

	case models.VAPaymentMethodGoPay:
		chargeReq.PaymentType = "gopay"
		chargeReq.GoPay = &GoPayDetails{
			EnableCallback: true,
		}

	case models.VAPaymentMethodCreditCard:
		// Credit card via Core API requires card tokenization which needs frontend integration
		// For now, we'll return an error suggesting to use Snap or VA instead
		return nil, fmt.Errorf("credit card payment requires Snap integration. Please use Virtual Account or QRIS instead")

	default:
		return nil, fmt.Errorf("%w: %s", ErrInvalidPaymentType, request.PaymentMethod)
	}

	return chargeReq, nil
}

// extractVANumber extracts the VA number from Midtrans response based on bank
func (c *midtransCoreClient) extractVANumber(method models.VAPaymentMethod, resp *MidtransChargeResponse) (string, string, error) {
	switch method {
	case models.VAPaymentMethodBCA, models.VAPaymentMethodBRI, models.VAPaymentMethodBNI:
		// BCA, BRI, and BNI use va_numbers array
		if len(resp.VANumbers) > 0 {
			return resp.VANumbers[0].VANumber, resp.VANumbers[0].Bank, nil
		}
		return "", "", ErrVANumberNotFound

	case models.VAPaymentMethodPermata:
		// Permata uses permata_va_number field
		if resp.PermataVANumber != "" {
			return resp.PermataVANumber, "permata", nil
		}
		return "", "", ErrVANumberNotFound

	case models.VAPaymentMethodMandiri:
		// Mandiri uses bill_key + biller_code
		if resp.BillKey != "" && resp.BillerCode != "" {
			// Combine biller_code and bill_key as VA number
			vaNumber := resp.BillerCode + resp.BillKey
			return vaNumber, "mandiri", nil
		}
		return "", "", ErrVANumberNotFound

	case models.VAPaymentMethodQRIS:
		// QRIS returns QR code URL in actions array
		for _, action := range resp.Actions {
			if action.Name == "generate-qr-code" {
				return action.URL, "qris", nil
			}
		}
		// Fallback - return transaction ID as reference
		return resp.TransactionID, "qris", nil

	case models.VAPaymentMethodGoPay:
		// GoPay returns QR code URL and deeplink in actions array
		// Priority: deeplink-redirect > generate-qr-code
		var qrURL, deeplinkURL string
		for _, action := range resp.Actions {
			if action.Name == "generate-qr-code" {
				qrURL = action.URL
			}
			if action.Name == "deeplink-redirect" {
				deeplinkURL = action.URL
			}
		}
		// Return deeplink if available, otherwise QR code
		if deeplinkURL != "" {
			return deeplinkURL, "gopay", nil
		}
		if qrURL != "" {
			return qrURL, "gopay", nil
		}
		// Fallback - return transaction ID as reference
		return resp.TransactionID, "gopay", nil

	case models.VAPaymentMethodCreditCard:
		// Credit card doesn't have VA number - should not reach here
		return "", "credit_card", fmt.Errorf("credit card not supported via Core API")

	default:
		return "", "", ErrInvalidPaymentType
	}
}

// parseExpiryTime parses the expiry time from Midtrans response
func (c *midtransCoreClient) parseExpiryTime(expiryStr string) (time.Time, error) {
	if expiryStr == "" {
		return time.Time{}, errors.New("empty expiry time")
	}

	// Midtrans format: "2026-01-14 10:30:00"
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		time.RFC3339,
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, expiryStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("failed to parse expiry time: %s", expiryStr)
}

// GetTransactionStatus gets transaction status from Midtrans API
// GET /v2/{order_id}/status
func (c *midtransCoreClient) GetTransactionStatus(orderID string) (*TransactionStatusResponse, error) {
	log.Printf("🔍 Midtrans GetTransactionStatus: order_id=%s", orderID)

	// Create HTTP request
	url := fmt.Sprintf("%s/v2/%s/status", c.baseURL, orderID)
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(c.serverKey+":")))

	// Execute request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		if os.IsTimeout(err) {
			log.Printf("❌ Midtrans timeout: %v", err)
			return nil, ErrMidtransTimeout
		}
		log.Printf("❌ Midtrans network error: %v", err)
		return nil, fmt.Errorf("midtrans network error: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	log.Printf("📥 Midtrans status response: %s", string(respBody))

	// Parse response
	var statusResp TransactionStatusResponse
	if err := json.Unmarshal(respBody, &statusResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for error status (404 = transaction not found)
	if statusResp.StatusCode == "404" {
		return nil, fmt.Errorf("transaction not found: %s", orderID)
	}

	log.Printf("✅ Midtrans status: order_id=%s, status=%s", orderID, statusResp.TransactionStatus)

	return &statusResp, nil
}
