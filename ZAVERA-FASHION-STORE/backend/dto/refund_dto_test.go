package dto

import (
	"encoding/json"
	"testing"
	"time"
)

// TestRefundRequestDTO tests RefundRequest DTO marshaling
func TestRefundRequestDTO(t *testing.T) {
	amount := 100000.0
	req := RefundRequest{
		OrderCode:      "ORD-2024-001",
		RefundType:     "FULL",
		Reason:         "CUSTOMER_REQUEST",
		ReasonDetail:   "Customer changed mind",
		Amount:         &amount,
		IdempotencyKey: "test-key-123",
	}

	// Test JSON marshaling
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal RefundRequest: %v", err)
	}

	// Test JSON unmarshaling
	var decoded RefundRequest
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal RefundRequest: %v", err)
	}

	// Verify fields
	if decoded.OrderCode != req.OrderCode {
		t.Errorf("Expected OrderCode %s, got %s", req.OrderCode, decoded.OrderCode)
	}
	if decoded.RefundType != req.RefundType {
		t.Errorf("Expected RefundType %s, got %s", req.RefundType, decoded.RefundType)
	}
}

// TestRefundItemRequestDTO tests RefundItemRequest DTO
func TestRefundItemRequestDTO(t *testing.T) {
	item := RefundItemRequest{
		OrderItemID: 1,
		Quantity:    2,
		Reason:      "Damaged item",
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("Failed to marshal RefundItemRequest: %v", err)
	}

	var decoded RefundItemRequest
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal RefundItemRequest: %v", err)
	}

	if decoded.OrderItemID != item.OrderItemID {
		t.Errorf("Expected OrderItemID %d, got %d", item.OrderItemID, decoded.OrderItemID)
	}
	if decoded.Quantity != item.Quantity {
		t.Errorf("Expected Quantity %d, got %d", item.Quantity, decoded.Quantity)
	}
}

// TestRefundResponseDTO tests RefundResponse DTO with all fields
func TestRefundResponseDTO(t *testing.T) {
	now := time.Now()
	paymentID := 123
	processedBy := 456
	requestedBy := 789

	resp := RefundResponse{
		ID:              1,
		RefundCode:      "REF-2024-001",
		OrderCode:       "ORD-2024-001",
		OrderID:         100,
		PaymentID:       &paymentID,
		RefundType:      "FULL",
		Reason:          "CUSTOMER_REQUEST",
		ReasonDetail:    "Customer changed mind",
		OriginalAmount:  150000.0,
		RefundAmount:    150000.0,
		ShippingRefund:  15000.0,
		ItemsRefund:     135000.0,
		Status:          "COMPLETED",
		GatewayRefundID: "MIDTRANS-REF-123",
		GatewayStatus:   "refund",
		IdempotencyKey:  "test-key-123",
		ProcessedBy:     &processedBy,
		ProcessedAt:     &now,
		RequestedBy:     &requestedBy,
		RequestedAt:     now,
		CompletedAt:     &now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal RefundResponse: %v", err)
	}

	var decoded RefundResponse
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal RefundResponse: %v", err)
	}

	if decoded.RefundCode != resp.RefundCode {
		t.Errorf("Expected RefundCode %s, got %s", resp.RefundCode, decoded.RefundCode)
	}
	if decoded.RefundAmount != resp.RefundAmount {
		t.Errorf("Expected RefundAmount %.2f, got %.2f", resp.RefundAmount, decoded.RefundAmount)
	}
	if *decoded.PaymentID != *resp.PaymentID {
		t.Errorf("Expected PaymentID %d, got %d", *resp.PaymentID, *decoded.PaymentID)
	}
}

// TestMidtransRefundRequestDTO tests MidtransRefundRequest DTO
func TestMidtransRefundRequestDTO(t *testing.T) {
	req := MidtransRefundRequest{
		RefundKey: "REF-2024-001",
		Amount:    100000.0,
		Reason:    "CUSTOMER_REQUEST",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal MidtransRefundRequest: %v", err)
	}

	var decoded MidtransRefundRequest
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal MidtransRefundRequest: %v", err)
	}

	if decoded.RefundKey != req.RefundKey {
		t.Errorf("Expected RefundKey %s, got %s", req.RefundKey, decoded.RefundKey)
	}
	if decoded.Amount != req.Amount {
		t.Errorf("Expected Amount %.2f, got %.2f", req.Amount, decoded.Amount)
	}
}

// TestMidtransRefundResponseDTO tests MidtransRefundResponse DTO
func TestMidtransRefundResponseDTO(t *testing.T) {
	resp := MidtransRefundResponse{
		StatusCode:         "200",
		StatusMessage:      "Success, refund is processed",
		RefundChargebackID: 12345,
		RefundAmount:       "100000.00",
		RefundKey:          "REF-2024-001",
		TransactionID:      "TXN-123",
		OrderID:            "ORD-2024-001",
		GrossAmount:        "100000.00",
		Currency:           "IDR",
		PaymentType:        "bank_transfer",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal MidtransRefundResponse: %v", err)
	}

	var decoded MidtransRefundResponse
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal MidtransRefundResponse: %v", err)
	}

	if decoded.StatusCode != resp.StatusCode {
		t.Errorf("Expected StatusCode %s, got %s", resp.StatusCode, decoded.StatusCode)
	}
	if decoded.RefundChargebackID != resp.RefundChargebackID {
		t.Errorf("Expected RefundChargebackID %d, got %d", resp.RefundChargebackID, decoded.RefundChargebackID)
	}
}

// TestCustomerRefundResponseDTO tests CustomerRefundResponse DTO
func TestCustomerRefundResponseDTO(t *testing.T) {
	now := time.Now()
	resp := CustomerRefundResponse{
		RefundCode:     "REF-2024-001",
		OrderCode:      "ORD-2024-001",
		RefundType:     "FULL",
		RefundAmount:   150000.0,
		ShippingRefund: 15000.0,
		ItemsRefund:    135000.0,
		Status:         "COMPLETED",
		StatusLabel:    "Refund completed",
		Timeline:       "Funds returned to your payment method",
		RequestedAt:    now,
		CompletedAt:    &now,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal CustomerRefundResponse: %v", err)
	}

	var decoded CustomerRefundResponse
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal CustomerRefundResponse: %v", err)
	}

	if decoded.RefundCode != resp.RefundCode {
		t.Errorf("Expected RefundCode %s, got %s", resp.RefundCode, decoded.RefundCode)
	}
	if decoded.StatusLabel != resp.StatusLabel {
		t.Errorf("Expected StatusLabel %s, got %s", resp.StatusLabel, decoded.StatusLabel)
	}
}

// TestRefundListResponseDTO tests RefundListResponse DTO
func TestRefundListResponseDTO(t *testing.T) {
	now := time.Now()
	resp := RefundListResponse{
		Refunds: []RefundResponse{
			{
				ID:             1,
				RefundCode:     "REF-2024-001",
				OrderCode:      "ORD-2024-001",
				OrderID:        100,
				RefundType:     "FULL",
				Reason:         "CUSTOMER_REQUEST",
				OriginalAmount: 150000.0,
				RefundAmount:   150000.0,
				ShippingRefund: 15000.0,
				ItemsRefund:    135000.0,
				Status:         "COMPLETED",
				RequestedAt:    now,
				CreatedAt:      now,
				UpdatedAt:      now,
			},
		},
		TotalCount: 1,
		Page:       1,
		PageSize:   10,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal RefundListResponse: %v", err)
	}

	var decoded RefundListResponse
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal RefundListResponse: %v", err)
	}

	if decoded.TotalCount != resp.TotalCount {
		t.Errorf("Expected TotalCount %d, got %d", resp.TotalCount, decoded.TotalCount)
	}
	if len(decoded.Refunds) != len(resp.Refunds) {
		t.Errorf("Expected %d refunds, got %d", len(resp.Refunds), len(decoded.Refunds))
	}
}

// TestRefundSuccessResponseDTO tests RefundSuccessResponse DTO
func TestRefundSuccessResponseDTO(t *testing.T) {
	resp := RefundSuccessResponse{
		Success:         true,
		Message:         "Refund processed successfully",
		RefundCode:      "REF-2024-001",
		GatewayRefundID: "MIDTRANS-REF-123",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal RefundSuccessResponse: %v", err)
	}

	var decoded RefundSuccessResponse
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal RefundSuccessResponse: %v", err)
	}

	if decoded.Success != resp.Success {
		t.Errorf("Expected Success %v, got %v", resp.Success, decoded.Success)
	}
	if decoded.RefundCode != resp.RefundCode {
		t.Errorf("Expected RefundCode %s, got %s", resp.RefundCode, decoded.RefundCode)
	}
}

// TestRefundErrorResponseDTO tests RefundErrorResponse DTO
func TestRefundErrorResponseDTO(t *testing.T) {
	resp := RefundErrorResponse{
		Error:   "REFUND_AMOUNT_EXCEEDS_BALANCE",
		Message: "Refund amount exceeds refundable balance",
		Details: map[string]interface{}{
			"order_code":        "ORD-2024-001",
			"requested_amount":  1000000.0,
			"refundable_balance": 850000.0,
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal RefundErrorResponse: %v", err)
	}

	var decoded RefundErrorResponse
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal RefundErrorResponse: %v", err)
	}

	if decoded.Error != resp.Error {
		t.Errorf("Expected Error %s, got %s", resp.Error, decoded.Error)
	}
	if decoded.Details["order_code"] != resp.Details["order_code"] {
		t.Errorf("Expected order_code %s, got %s", resp.Details["order_code"], decoded.Details["order_code"])
	}
}
