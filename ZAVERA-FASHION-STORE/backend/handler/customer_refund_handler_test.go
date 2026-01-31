package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"zavera/dto"
	"zavera/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock RefundService
type MockRefundService struct {
	mock.Mock
}

func (m *MockRefundService) CreateRefund(req *dto.RefundRequest, requestedBy *int) (*models.Refund, error) {
	args := m.Called(req, requestedBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Refund), args.Error(1)
}

func (m *MockRefundService) ProcessRefund(refundID int, processedBy int) error {
	args := m.Called(refundID, processedBy)
	return args.Error(0)
}

func (m *MockRefundService) GetRefund(refundID int) (*models.Refund, error) {
	args := m.Called(refundID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Refund), args.Error(1)
}

func (m *MockRefundService) GetRefundByCode(code string) (*models.Refund, error) {
	args := m.Called(code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Refund), args.Error(1)
}

func (m *MockRefundService) GetRefundsByOrder(orderID int) ([]*models.Refund, error) {
	args := m.Called(orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Refund), args.Error(1)
}

func (m *MockRefundService) FullRefund(orderCode string, reason models.RefundReason, detail string, requestedBy *int, idempotencyKey string) (*models.Refund, error) {
	args := m.Called(orderCode, reason, detail, requestedBy, idempotencyKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Refund), args.Error(1)
}

func (m *MockRefundService) PartialRefund(orderCode string, amount float64, reason models.RefundReason, detail string, requestedBy *int, idempotencyKey string) (*models.Refund, error) {
	args := m.Called(orderCode, amount, reason, detail, requestedBy, idempotencyKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Refund), args.Error(1)
}

func (m *MockRefundService) ShippingOnlyRefund(orderCode string, reason models.RefundReason, detail string, requestedBy *int, idempotencyKey string) (*models.Refund, error) {
	args := m.Called(orderCode, reason, detail, requestedBy, idempotencyKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Refund), args.Error(1)
}

func (m *MockRefundService) ItemRefund(orderCode string, items []dto.RefundItemRequest, reason models.RefundReason, detail string, requestedBy *int, idempotencyKey string) (*models.Refund, error) {
	args := m.Called(orderCode, items, reason, detail, requestedBy, idempotencyKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Refund), args.Error(1)
}

func (m *MockRefundService) ProcessMidtransRefund(refund *models.Refund) (*dto.MidtransRefundResponse, error) {
	args := m.Called(refund)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.MidtransRefundResponse), args.Error(1)
}

func (m *MockRefundService) CheckMidtransRefundStatus(refundCode string) (*dto.MidtransRefundResponse, error) {
	args := m.Called(refundCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.MidtransRefundResponse), args.Error(1)
}

// Mock OrderService
type MockOrderService struct {
	mock.Mock
}

func (m *MockOrderService) CreateOrder(sessionID string, req dto.CheckoutRequest, userID *int) (*dto.CheckoutResponse, error) {
	args := m.Called(sessionID, req, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.CheckoutResponse), args.Error(1)
}

func (m *MockOrderService) GetOrder(orderCode string) (*dto.OrderResponse, error) {
	args := m.Called(orderCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.OrderResponse), args.Error(1)
}

func (m *MockOrderService) GetOrderByID(orderID int) (*models.Order, error) {
	args := m.Called(orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Order), args.Error(1)
}

func (m *MockOrderService) GetOrderItems(orderID int) ([]models.OrderItem, error) {
	args := m.Called(orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.OrderItem), args.Error(1)
}

func (m *MockOrderService) UpdateOrderStatus(orderCode string, status models.OrderStatus) error {
	args := m.Called(orderCode, status)
	return args.Error(0)
}

func (m *MockOrderService) MarkAsPaid(orderCode string) error {
	args := m.Called(orderCode)
	return args.Error(0)
}

func (m *MockOrderService) MarkAsPacking(orderCode string) error {
	args := m.Called(orderCode)
	return args.Error(0)
}

func (m *MockOrderService) MarkAsShippedWithResi(orderCode string, resi string) error {
	args := m.Called(orderCode, resi)
	return args.Error(0)
}

func (m *MockOrderService) MarkAsShipped(orderCode string) error {
	args := m.Called(orderCode)
	return args.Error(0)
}

func (m *MockOrderService) MarkAsDelivered(orderCode string) error {
	args := m.Called(orderCode)
	return args.Error(0)
}

func (m *MockOrderService) MarkAsCompleted(orderCode string) error {
	args := m.Called(orderCode)
	return args.Error(0)
}

func (m *MockOrderService) MarkAsRefunded(orderCode string) error {
	args := m.Called(orderCode)
	return args.Error(0)
}

func (m *MockOrderService) CancelOrder(orderCode string, reason string) error {
	args := m.Called(orderCode, reason)
	return args.Error(0)
}

func (m *MockOrderService) CancelOrderByCustomer(orderCode string) error {
	args := m.Called(orderCode)
	return args.Error(0)
}

func (m *MockOrderService) CancelOrderByAdmin(orderCode string, reason string) error {
	args := m.Called(orderCode, reason)
	return args.Error(0)
}

func (m *MockOrderService) ExpireOrder(orderCode string) error {
	args := m.Called(orderCode)
	return args.Error(0)
}

func (m *MockOrderService) FailOrder(orderCode string, reason string) error {
	args := m.Called(orderCode, reason)
	return args.Error(0)
}

func (m *MockOrderService) ValidateOrderTotals(order *models.Order) error {
	args := m.Called(order)
	return args.Error(0)
}

// Test GetOrderRefunds - Success
func TestGetOrderRefunds_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockRefundService := new(MockRefundService)
	mockOrderService := new(MockOrderService)
	handler := NewCustomerRefundHandler(mockRefundService, mockOrderService)
	
	// Setup test data
	orderCode := "ORD-2024-001"
	orderID := 1
	userID := 123
	paymentID := 1
	
	orderResp := &dto.OrderResponse{
		ID:        orderID,
		OrderCode: orderCode,
	}
	
	order := &models.Order{
		ID:        orderID,
		OrderCode: orderCode,
		UserID:    &userID,
	}
	
	refunds := []*models.Refund{
		{
			ID:             1,
			RefundCode:     "REF-2024-001",
			OrderID:        orderID,
			PaymentID:      &paymentID,
			RefundType:     models.RefundTypeFull,
			Reason:         models.RefundReasonCustomerRequest,
			RefundAmount:   100000,
			ShippingRefund: 10000,
			ItemsRefund:    90000,
			Status:         models.RefundStatusCompleted,
			RequestedAt:    time.Now(),
			Items:          []models.RefundItem{},
		},
	}
	
	// Setup expectations
	mockOrderService.On("GetOrder", orderCode).Return(orderResp, nil)
	mockOrderService.On("GetOrderByID", orderID).Return(order, nil)
	mockRefundService.On("GetRefundsByOrder", orderID).Return(refunds, nil)
	
	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "code", Value: orderCode}}
	c.Set("user_id", userID)
	
	// Execute
	handler.GetOrderRefunds(c)
	
	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response dto.CustomerRefundListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 1, response.Count)
	assert.Equal(t, "REF-2024-001", response.Refunds[0].RefundCode)
	assert.Equal(t, orderCode, response.Refunds[0].OrderCode)
	assert.Equal(t, "Refund Completed", response.Refunds[0].StatusLabel)
	
	mockOrderService.AssertExpectations(t)
	mockRefundService.AssertExpectations(t)
}

// Test GetOrderRefunds - Unauthorized (no user_id in context)
func TestGetOrderRefunds_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockRefundService := new(MockRefundService)
	mockOrderService := new(MockOrderService)
	handler := NewCustomerRefundHandler(mockRefundService, mockOrderService)
	
	// Create request without user_id
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "code", Value: "ORD-2024-001"}}
	
	// Execute
	handler.GetOrderRefunds(c)
	
	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "UNAUTHORIZED", response["error"])
}

// Test GetOrderRefunds - Forbidden (customer doesn't own order)
func TestGetOrderRefunds_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockRefundService := new(MockRefundService)
	mockOrderService := new(MockOrderService)
	handler := NewCustomerRefundHandler(mockRefundService, mockOrderService)
	
	// Setup test data
	orderCode := "ORD-2024-001"
	orderID := 1
	userID := 123
	differentUserID := 456
	
	orderResp := &dto.OrderResponse{
		ID:        orderID,
		OrderCode: orderCode,
	}
	
	order := &models.Order{
		ID:        orderID,
		OrderCode: orderCode,
		UserID:    &differentUserID, // Different user owns this order
	}
	
	// Setup expectations
	mockOrderService.On("GetOrder", orderCode).Return(orderResp, nil)
	mockOrderService.On("GetOrderByID", orderID).Return(order, nil)
	
	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "code", Value: orderCode}}
	c.Set("user_id", userID)
	
	// Execute
	handler.GetOrderRefunds(c)
	
	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "FORBIDDEN", response["error"])
	
	mockOrderService.AssertExpectations(t)
}

// Test GetRefundByCode - Success
func TestGetRefundByCode_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockRefundService := new(MockRefundService)
	mockOrderService := new(MockOrderService)
	handler := NewCustomerRefundHandler(mockRefundService, mockOrderService)
	
	// Setup test data
	refundCode := "REF-2024-001"
	orderCode := "ORD-2024-001"
	orderID := 1
	userID := 123
	paymentID := 1
	
	refund := &models.Refund{
		ID:             1,
		RefundCode:     refundCode,
		OrderID:        orderID,
		PaymentID:      &paymentID,
		RefundType:     models.RefundTypeFull,
		Reason:         models.RefundReasonCustomerRequest,
		RefundAmount:   100000,
		ShippingRefund: 10000,
		ItemsRefund:    90000,
		Status:         models.RefundStatusCompleted,
		RequestedAt:    time.Now(),
		Items:          []models.RefundItem{},
	}
	
	order := &models.Order{
		ID:        orderID,
		OrderCode: orderCode,
		UserID:    &userID,
	}
	
	// Setup expectations
	mockRefundService.On("GetRefundByCode", refundCode).Return(refund, nil)
	mockOrderService.On("GetOrderByID", orderID).Return(order, nil)
	
	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "code", Value: refundCode}}
	c.Set("user_id", userID)
	
	// Execute
	handler.GetRefundByCode(c)
	
	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response dto.CustomerRefundResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, refundCode, response.RefundCode)
	assert.Equal(t, orderCode, response.OrderCode)
	assert.Equal(t, "Refund Completed", response.StatusLabel)
	assert.Contains(t, response.Timeline, "funds have been returned")
	
	mockRefundService.AssertExpectations(t)
	mockOrderService.AssertExpectations(t)
}

// Test GetRefundByCode - Refund Not Found
func TestGetRefundByCode_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockRefundService := new(MockRefundService)
	mockOrderService := new(MockOrderService)
	handler := NewCustomerRefundHandler(mockRefundService, mockOrderService)
	
	refundCode := "REF-2024-999"
	
	// Setup expectations
	mockRefundService.On("GetRefundByCode", refundCode).Return(nil, errors.New("not found"))
	
	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "code", Value: refundCode}}
	c.Set("user_id", 123)
	
	// Execute
	handler.GetRefundByCode(c)
	
	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "REFUND_NOT_FOUND", response["error"])
	
	mockRefundService.AssertExpectations(t)
}

// Test timeline messages for different statuses
func TestGetTimelineMessage(t *testing.T) {
	handler := &CustomerRefundHandler{}
	paymentID := 1
	
	tests := []struct {
		name      string
		status    models.RefundStatus
		paymentID *int
		expected  string
	}{
		{
			name:      "Pending status",
			status:    models.RefundStatusPending,
			paymentID: &paymentID,
			expected:  "Your refund request has been received and is awaiting processing.",
		},
		{
			name:      "Processing with payment",
			status:    models.RefundStatusProcessing,
			paymentID: &paymentID,
			expected:  "Refund in progress - funds will arrive in 3-7 business days depending on your payment method.",
		},
		{
			name:      "Processing manual refund",
			status:    models.RefundStatusProcessing,
			paymentID: nil,
			expected:  "Your refund is being processed manually. Please contact support for details.",
		},
		{
			name:      "Completed with payment",
			status:    models.RefundStatusCompleted,
			paymentID: &paymentID,
			expected:  "Refund completed - funds have been returned to your payment method.",
		},
		{
			name:      "Completed manual refund",
			status:    models.RefundStatusCompleted,
			paymentID: nil,
			expected:  "Refund completed - processed manually. Please contact support if you have questions.",
		},
		{
			name:      "Failed status",
			status:    models.RefundStatusFailed,
			paymentID: &paymentID,
			expected:  "Refund failed - please contact support for assistance.",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.getTimelineMessage(tt.status, tt.paymentID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test status labels
func TestGetStatusLabel(t *testing.T) {
	handler := &CustomerRefundHandler{}
	
	tests := []struct {
		status   models.RefundStatus
		expected string
	}{
		{models.RefundStatusPending, "Refund Pending"},
		{models.RefundStatusProcessing, "Refund in Progress"},
		{models.RefundStatusCompleted, "Refund Completed"},
		{models.RefundStatusFailed, "Refund Failed"},
	}
	
	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			result := handler.getStatusLabel(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}
