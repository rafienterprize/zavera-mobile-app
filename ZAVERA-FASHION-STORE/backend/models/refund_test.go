package models

import (
	"testing"
	"time"
)

// TestRefundModelNullableFields verifies that nullable fields work correctly
func TestRefundModelNullableFields(t *testing.T) {
	// Test 1: Refund with all nullable fields as nil
	refund1 := Refund{
		ID:             1,
		RefundCode:     "REF-001",
		OrderID:        1,
		PaymentID:      nil, // Nullable - for manual refunds
		RefundType:     RefundTypeFull,
		Reason:         RefundReasonCustomerRequest,
		RefundAmount:   100000,
		Status:         RefundStatusPending,
		RequestedBy:    nil, // Nullable - system-initiated refunds
		ProcessedBy:    nil, // Nullable - not yet processed
		RequestedAt:    time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if refund1.PaymentID != nil {
		t.Errorf("Expected PaymentID to be nil, got %v", refund1.PaymentID)
	}
	if refund1.RequestedBy != nil {
		t.Errorf("Expected RequestedBy to be nil, got %v", refund1.RequestedBy)
	}
	if refund1.ProcessedBy != nil {
		t.Errorf("Expected ProcessedBy to be nil, got %v", refund1.ProcessedBy)
	}

	// Test 2: Refund with nullable fields set
	paymentID := 123
	requestedBy := 456
	processedBy := 789
	refund2 := Refund{
		ID:             2,
		RefundCode:     "REF-002",
		OrderID:        2,
		PaymentID:      &paymentID,
		RefundType:     RefundTypePartial,
		Reason:         RefundReasonDamagedItem,
		RefundAmount:   50000,
		Status:         RefundStatusCompleted,
		RequestedBy:    &requestedBy,
		ProcessedBy:    &processedBy,
		RequestedAt:    time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if refund2.PaymentID == nil || *refund2.PaymentID != 123 {
		t.Errorf("Expected PaymentID to be 123, got %v", refund2.PaymentID)
	}
	if refund2.RequestedBy == nil || *refund2.RequestedBy != 456 {
		t.Errorf("Expected RequestedBy to be 456, got %v", refund2.RequestedBy)
	}
	if refund2.ProcessedBy == nil || *refund2.ProcessedBy != 789 {
		t.Errorf("Expected ProcessedBy to be 789, got %v", refund2.ProcessedBy)
	}
}

// TestOrderRefundFields verifies that Order model has refund tracking fields
func TestOrderRefundFields(t *testing.T) {
	// Test 1: Order without refund
	order1 := Order{
		ID:           1,
		OrderCode:    "ORD-001",
		TotalAmount:  100000,
		Status:       OrderStatusDelivered,
		RefundStatus: nil,
		RefundAmount: 0,
		RefundedAt:   nil,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if order1.RefundStatus != nil {
		t.Errorf("Expected RefundStatus to be nil, got %v", order1.RefundStatus)
	}
	if order1.RefundAmount != 0 {
		t.Errorf("Expected RefundAmount to be 0, got %v", order1.RefundAmount)
	}
	if order1.RefundedAt != nil {
		t.Errorf("Expected RefundedAt to be nil, got %v", order1.RefundedAt)
	}

	// Test 2: Order with full refund
	refundStatus := "FULL"
	refundedAt := time.Now()
	order2 := Order{
		ID:           2,
		OrderCode:    "ORD-002",
		TotalAmount:  100000,
		Status:       OrderStatusRefunded,
		RefundStatus: &refundStatus,
		RefundAmount: 100000,
		RefundedAt:   &refundedAt,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if order2.RefundStatus == nil || *order2.RefundStatus != "FULL" {
		t.Errorf("Expected RefundStatus to be 'FULL', got %v", order2.RefundStatus)
	}
	if order2.RefundAmount != 100000 {
		t.Errorf("Expected RefundAmount to be 100000, got %v", order2.RefundAmount)
	}
	if order2.RefundedAt == nil {
		t.Errorf("Expected RefundedAt to be set, got nil")
	}

	// Test 3: Order with partial refund
	partialStatus := "PARTIAL"
	order3 := Order{
		ID:           3,
		OrderCode:    "ORD-003",
		TotalAmount:  100000,
		Status:       OrderStatusCompleted,
		RefundStatus: &partialStatus,
		RefundAmount: 50000,
		RefundedAt:   &refundedAt,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if order3.RefundStatus == nil || *order3.RefundStatus != "PARTIAL" {
		t.Errorf("Expected RefundStatus to be 'PARTIAL', got %v", order3.RefundStatus)
	}
	if order3.RefundAmount != 50000 {
		t.Errorf("Expected RefundAmount to be 50000, got %v", order3.RefundAmount)
	}
}

// TestRefundStatusHistory verifies the RefundStatusHistory model
func TestRefundStatusHistory(t *testing.T) {
	// Test 1: Initial status change (no old status)
	history1 := RefundStatusHistory{
		ID:        1,
		RefundID:  1,
		OldStatus: nil,
		NewStatus: RefundStatusPending,
		Actor:     "system",
		Reason:    "Refund created",
		CreatedAt: time.Now(),
	}

	if history1.OldStatus != nil {
		t.Errorf("Expected OldStatus to be nil, got %v", history1.OldStatus)
	}
	if history1.NewStatus != RefundStatusPending {
		t.Errorf("Expected NewStatus to be PENDING, got %v", history1.NewStatus)
	}

	// Test 2: Status change with old status
	oldStatus := string(RefundStatusPending)
	history2 := RefundStatusHistory{
		ID:        2,
		RefundID:  1,
		OldStatus: &oldStatus,
		NewStatus: RefundStatusProcessing,
		Actor:     "admin@example.com",
		Reason:    "Processing refund",
		CreatedAt: time.Now(),
	}

	if history2.OldStatus == nil || *history2.OldStatus != string(RefundStatusPending) {
		t.Errorf("Expected OldStatus to be 'PENDING', got %v", history2.OldStatus)
	}
	if history2.NewStatus != RefundStatusProcessing {
		t.Errorf("Expected NewStatus to be PROCESSING, got %v", history2.NewStatus)
	}
}
