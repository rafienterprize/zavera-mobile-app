package service

import (
	"testing"
	"zavera/dto"
	"zavera/models"
)

// Test validateRefundRequest method
func TestValidateRefundRequest(t *testing.T) {
	// Create a mock refund service (we'll only test the validation logic)
	service := &refundService{}

	tests := []struct {
		name        string
		req         *dto.RefundRequest
		order       *models.Order
		payment     *models.Payment
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid full refund for delivered order",
			req: &dto.RefundRequest{
				OrderCode:  "ORD-001",
				RefundType: "FULL",
				Reason:     "CUSTOMER_REQUEST",
			},
			order: &models.Order{
				ID:          1,
				OrderCode:   "ORD-001",
				Status:      models.OrderStatusDelivered,
				TotalAmount: 100000,
			},
			payment: &models.Payment{
				ID:     1,
				Status: models.PaymentStatusSuccess,
				Amount: 100000,
			},
			expectError: false,
		},
		{
			name: "Invalid - order not delivered",
			req: &dto.RefundRequest{
				OrderCode:  "ORD-002",
				RefundType: "FULL",
				Reason:     "CUSTOMER_REQUEST",
			},
			order: &models.Order{
				ID:        2,
				OrderCode: "ORD-002",
				Status:    models.OrderStatusPending,
			},
			payment: &models.Payment{
				ID:     2,
				Status: models.PaymentStatusSuccess,
				Amount: 100000,
			},
			expectError: true,
			errorMsg:    "order is not refundable",
		},
		{
			name: "Invalid - payment not settled",
			req: &dto.RefundRequest{
				OrderCode:  "ORD-003",
				RefundType: "FULL",
				Reason:     "CUSTOMER_REQUEST",
			},
			order: &models.Order{
				ID:        3,
				OrderCode: "ORD-003",
				Status:    models.OrderStatusDelivered,
			},
			payment: &models.Payment{
				ID:     3,
				Status: models.PaymentStatusPending,
				Amount: 100000,
			},
			expectError: true,
			errorMsg:    "payment not settled",
		},
		{
			name: "Invalid - partial refund with zero amount",
			req: &dto.RefundRequest{
				OrderCode:  "ORD-004",
				RefundType: "PARTIAL",
				Reason:     "CUSTOMER_REQUEST",
				Amount:     func() *float64 { v := 0.0; return &v }(),
			},
			order: &models.Order{
				ID:        4,
				OrderCode: "ORD-004",
				Status:    models.OrderStatusDelivered,
			},
			payment: &models.Payment{
				ID:     4,
				Status: models.PaymentStatusSuccess,
				Amount: 100000,
			},
			expectError: true,
			errorMsg:    "must be positive and non-zero",
		},
		{
			name: "Invalid - item refund with zero quantity",
			req: &dto.RefundRequest{
				OrderCode:  "ORD-005",
				RefundType: "ITEM_ONLY",
				Reason:     "CUSTOMER_REQUEST",
				Items: []dto.RefundItemRequest{
					{
						OrderItemID: 1,
						Quantity:    0,
					},
				},
			},
			order: &models.Order{
				ID:        5,
				OrderCode: "ORD-005",
				Status:    models.OrderStatusDelivered,
				Items: []models.OrderItem{
					{
						ID:           1,
						ProductID:    1,
						ProductName:  "Test Product",
						Quantity:     5,
						PricePerUnit: 10000,
					},
				},
			},
			payment: &models.Payment{
				ID:     5,
				Status: models.PaymentStatusSuccess,
				Amount: 100000,
			},
			expectError: true,
			errorMsg:    "quantity must be positive",
		},
		{
			name: "Invalid - item refund with non-existent item",
			req: &dto.RefundRequest{
				OrderCode:  "ORD-006",
				RefundType: "ITEM_ONLY",
				Reason:     "CUSTOMER_REQUEST",
				Items: []dto.RefundItemRequest{
					{
						OrderItemID: 999,
						Quantity:    1,
					},
				},
			},
			order: &models.Order{
				ID:        6,
				OrderCode: "ORD-006",
				Status:    models.OrderStatusDelivered,
				Items: []models.OrderItem{
					{
						ID:           1,
						ProductID:    1,
						ProductName:  "Test Product",
						Quantity:     5,
						PricePerUnit: 10000,
					},
				},
			},
			payment: &models.Payment{
				ID:     6,
				Status: models.PaymentStatusSuccess,
				Amount: 100000,
			},
			expectError: true,
			errorMsg:    "not found in order",
		},
		{
			name: "Invalid - item refund quantity exceeds ordered quantity",
			req: &dto.RefundRequest{
				OrderCode:  "ORD-007",
				RefundType: "ITEM_ONLY",
				Reason:     "CUSTOMER_REQUEST",
				Items: []dto.RefundItemRequest{
					{
						OrderItemID: 1,
						Quantity:    10,
					},
				},
			},
			order: &models.Order{
				ID:        7,
				OrderCode: "ORD-007",
				Status:    models.OrderStatusDelivered,
				Items: []models.OrderItem{
					{
						ID:           1,
						ProductID:    1,
						ProductName:  "Test Product",
						Quantity:     5,
						PricePerUnit: 10000,
					},
				},
			},
			payment: &models.Payment{
				ID:     7,
				Status: models.PaymentStatusSuccess,
				Amount: 100000,
			},
			expectError: true,
			errorMsg:    "exceeds ordered quantity",
		},
		{
			name: "Valid - manual refund (no payment)",
			req: &dto.RefundRequest{
				OrderCode:  "ORD-008",
				RefundType: "FULL",
				Reason:     "CUSTOMER_REQUEST",
			},
			order: &models.Order{
				ID:          8,
				OrderCode:   "ORD-008",
				Status:      models.OrderStatusDelivered,
				TotalAmount: 100000,
			},
			payment:     nil, // No payment for manual order
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateRefundRequest(tt.req, tt.order, tt.payment)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing '%s', but got no error", tt.errorMsg)
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', but got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
			}
		})
	}
}

// Test calculateRefundAmount method
func TestCalculateRefundAmount(t *testing.T) {
	service := &refundService{}

	tests := []struct {
		name            string
		order           *models.Order
		payment         *models.Payment
		req             *dto.RefundRequest
		expectedTotal   float64
		expectedShip    float64
		expectedItems   float64
		expectError     bool
	}{
		{
			name: "Full refund calculation",
			order: &models.Order{
				TotalAmount:  100000,
				ShippingCost: 10000,
				Subtotal:     90000,
			},
			payment: &models.Payment{
				Amount: 100000,
			},
			req: &dto.RefundRequest{
				RefundType: "FULL",
			},
			expectedTotal: 100000,
			expectedShip:  10000,
			expectedItems: 90000,
			expectError:   false,
		},
		{
			name: "Shipping only refund calculation",
			order: &models.Order{
				TotalAmount:  100000,
				ShippingCost: 10000,
				Subtotal:     90000,
			},
			payment: &models.Payment{
				Amount: 100000,
			},
			req: &dto.RefundRequest{
				RefundType: "SHIPPING_ONLY",
			},
			expectedTotal: 10000,
			expectedShip:  10000,
			expectedItems: 0,
			expectError:   false,
		},
		{
			name: "Partial refund calculation",
			order: &models.Order{
				TotalAmount:  100000,
				ShippingCost: 10000,
				Subtotal:     90000,
			},
			payment: &models.Payment{
				Amount: 100000,
			},
			req: &dto.RefundRequest{
				RefundType: "PARTIAL",
				Amount:     func() *float64 { v := 50000.0; return &v }(),
			},
			expectedTotal: 50000,
			expectedShip:  0,
			expectedItems: 50000,
			expectError:   false,
		},
		{
			name: "Item refund calculation",
			order: &models.Order{
				TotalAmount:  100000,
				ShippingCost: 10000,
				Subtotal:     90000,
				Items: []models.OrderItem{
					{
						ID:           1,
						ProductID:    1,
						Quantity:     3,
						PricePerUnit: 20000,
					},
					{
						ID:           2,
						ProductID:    2,
						Quantity:     2,
						PricePerUnit: 15000,
					},
				},
			},
			payment: &models.Payment{
				Amount: 100000,
			},
			req: &dto.RefundRequest{
				RefundType: "ITEM_ONLY",
				Items: []dto.RefundItemRequest{
					{
						OrderItemID: 1,
						Quantity:    2, // 2 * 20000 = 40000
					},
				},
			},
			expectedTotal: 40000,
			expectedShip:  0,
			expectedItems: 40000,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total, ship, items, err := service.calculateRefundAmount(tt.order, tt.payment, tt.req)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, but got no error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
				if total != tt.expectedTotal {
					t.Errorf("Expected total %.2f, got %.2f", tt.expectedTotal, total)
				}
				if ship != tt.expectedShip {
					t.Errorf("Expected shipping %.2f, got %.2f", tt.expectedShip, ship)
				}
				if items != tt.expectedItems {
					t.Errorf("Expected items %.2f, got %.2f", tt.expectedItems, items)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
