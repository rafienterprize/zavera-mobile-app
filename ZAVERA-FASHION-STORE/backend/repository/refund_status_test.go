package repository

import (
	"testing"
	"time"
	"zavera/models"
)

// TestRefundRepository_MarkCompleted_StoresGatewayResponse tests that MarkCompleted
// stores the gateway refund ID and complete gateway response
// Validates: Requirements 2.3, 2.8
func TestRefundRepository_MarkCompleted_StoresGatewayResponse(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		t.Skip("Skipping test: no test database available")
	}
	defer db.Close()

	repo := NewRefundRepository(db)

	// Create a test refund
	refund := &models.Refund{
		RefundCode:     GenerateRefundCode(),
		OrderID:        1,
		RefundType:     models.RefundTypeFull,
		Reason:         models.RefundReasonCustomerRequest,
		OriginalAmount: 100000,
		RefundAmount:   100000,
		Status:         models.RefundStatusProcessing,
		IdempotencyKey: "test-mark-completed-" + generateRandomString(8),
		RequestedAt:    mustParseTime("2024-01-15T10:00:00Z"),
	}

	err := repo.Create(refund)
	if err != nil {
		t.Fatalf("Failed to create refund: %v", err)
	}
	defer cleanupTestRefund(t, db, refund.ID)

	// Test MarkCompleted with gateway response
	gatewayRefundID := "MIDTRANS-REF-12345"
	gatewayResponse := map[string]any{
		"status_code":          "200",
		"status_message":       "Success, refund is processed",
		"refund_chargeback_id": 12345,
		"refund_amount":        "100000.00",
		"refund_key":           refund.RefundCode,
	}

	err = repo.MarkCompleted(refund.ID, gatewayRefundID, gatewayResponse)
	if err != nil {
		t.Fatalf("Failed to mark refund as completed: %v", err)
	}

	// Verify the refund was marked as completed
	completed, err := repo.FindByID(refund.ID)
	if err != nil {
		t.Fatalf("Failed to find completed refund: %v", err)
	}

	// Verify status is COMPLETED
	if completed.Status != models.RefundStatusCompleted {
		t.Errorf("Expected status COMPLETED, got %s", completed.Status)
	}

	// Verify gateway refund ID is stored (Requirement 2.3)
	if completed.GatewayRefundID != gatewayRefundID {
		t.Errorf("Expected gateway refund ID %s, got %s", gatewayRefundID, completed.GatewayRefundID)
	}

	// Verify gateway status is set to success
	if completed.GatewayStatus != "success" {
		t.Errorf("Expected gateway status 'success', got %s", completed.GatewayStatus)
	}

	// Verify gateway response is stored (Requirement 2.8)
	if completed.GatewayResponse == nil {
		t.Error("Expected gateway response to be stored, got nil")
	} else {
		if statusCode, ok := completed.GatewayResponse["status_code"].(string); !ok || statusCode != "200" {
			t.Errorf("Expected status_code '200' in gateway response, got %v", completed.GatewayResponse["status_code"])
		}
		if statusMsg, ok := completed.GatewayResponse["status_message"].(string); !ok || statusMsg != "Success, refund is processed" {
			t.Errorf("Expected status_message in gateway response, got %v", completed.GatewayResponse["status_message"])
		}
	}

	// Verify completed_at timestamp is set
	if completed.CompletedAt == nil {
		t.Error("Expected completed_at timestamp to be set")
	}
}

// TestRefundRepository_MarkFailed_StoresErrorMessage tests that MarkFailed
// stores the error message and gateway response
// Validates: Requirements 2.4, 2.8
func TestRefundRepository_MarkFailed_StoresErrorMessage(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		t.Skip("Skipping test: no test database available")
	}
	defer db.Close()

	repo := NewRefundRepository(db)

	// Create a test refund
	refund := &models.Refund{
		RefundCode:     GenerateRefundCode(),
		OrderID:        1,
		RefundType:     models.RefundTypeFull,
		Reason:         models.RefundReasonCustomerRequest,
		ReasonDetail:   "Customer requested refund",
		OriginalAmount: 100000,
		RefundAmount:   100000,
		Status:         models.RefundStatusProcessing,
		IdempotencyKey: "test-mark-failed-" + generateRandomString(8),
		RequestedAt:    mustParseTime("2024-01-15T10:00:00Z"),
	}

	err := repo.Create(refund)
	if err != nil {
		t.Fatalf("Failed to create refund: %v", err)
	}
	defer cleanupTestRefund(t, db, refund.ID)

	// Test MarkFailed with error message and gateway response
	errorMessage := "Merchant cannot modify the status of the transaction"
	gatewayResponse := map[string]any{
		"status_code":    "412",
		"status_message": errorMessage,
		"id":             refund.RefundCode,
	}

	err = repo.MarkFailed(refund.ID, errorMessage, gatewayResponse)
	if err != nil {
		t.Fatalf("Failed to mark refund as failed: %v", err)
	}

	// Verify the refund was marked as failed
	failed, err := repo.FindByID(refund.ID)
	if err != nil {
		t.Fatalf("Failed to find failed refund: %v", err)
	}

	// Verify status is FAILED (Requirement 2.4)
	if failed.Status != models.RefundStatusFailed {
		t.Errorf("Expected status FAILED, got %s", failed.Status)
	}

	// Verify gateway status is set to failed
	if failed.GatewayStatus != "failed" {
		t.Errorf("Expected gateway status 'failed', got %s", failed.GatewayStatus)
	}

	// Verify error message is appended to reason_detail (Requirement 2.4)
	expectedDetail := "Customer requested refund | Error: " + errorMessage
	if failed.ReasonDetail != expectedDetail {
		t.Errorf("Expected reason_detail '%s', got '%s'", expectedDetail, failed.ReasonDetail)
	}

	// Verify gateway response is stored (Requirement 2.8)
	if failed.GatewayResponse == nil {
		t.Error("Expected gateway response to be stored, got nil")
	} else {
		if statusCode, ok := failed.GatewayResponse["status_code"].(string); !ok || statusCode != "412" {
			t.Errorf("Expected status_code '412' in gateway response, got %v", failed.GatewayResponse["status_code"])
		}
		if statusMsg, ok := failed.GatewayResponse["status_message"].(string); !ok || statusMsg != errorMessage {
			t.Errorf("Expected status_message in gateway response, got %v", failed.GatewayResponse["status_message"])
		}
	}
}

// TestRefundRepository_UpdateStatus_StoresGatewayResponse tests that UpdateStatus
// stores the complete gateway response
// Validates: Requirement 2.8
func TestRefundRepository_UpdateStatus_StoresGatewayResponse(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		t.Skip("Skipping test: no test database available")
	}
	defer db.Close()

	repo := NewRefundRepository(db)

	// Create a test refund
	refund := &models.Refund{
		RefundCode:     GenerateRefundCode(),
		OrderID:        1,
		RefundType:     models.RefundTypeFull,
		Reason:         models.RefundReasonCustomerRequest,
		OriginalAmount: 100000,
		RefundAmount:   100000,
		Status:         models.RefundStatusPending,
		IdempotencyKey: "test-update-status-" + generateRandomString(8),
		RequestedAt:    mustParseTime("2024-01-15T10:00:00Z"),
	}

	err := repo.Create(refund)
	if err != nil {
		t.Fatalf("Failed to create refund: %v", err)
	}
	defer cleanupTestRefund(t, db, refund.ID)

	// Test UpdateStatus with gateway response
	gatewayResponse := map[string]any{
		"status":  "processing",
		"message": "Refund is being processed by payment gateway",
		"timestamp": "2024-01-15T10:05:00Z",
	}

	err = repo.UpdateStatus(refund.ID, models.RefundStatusProcessing, gatewayResponse)
	if err != nil {
		t.Fatalf("Failed to update status: %v", err)
	}

	// Verify the status was updated
	updated, err := repo.FindByID(refund.ID)
	if err != nil {
		t.Fatalf("Failed to find updated refund: %v", err)
	}

	// Verify status is PROCESSING
	if updated.Status != models.RefundStatusProcessing {
		t.Errorf("Expected status PROCESSING, got %s", updated.Status)
	}

	// Verify gateway response is stored (Requirement 2.8)
	if updated.GatewayResponse == nil {
		t.Error("Expected gateway response to be stored, got nil")
	} else {
		if status, ok := updated.GatewayResponse["status"].(string); !ok || status != "processing" {
			t.Errorf("Expected status 'processing' in gateway response, got %v", updated.GatewayResponse["status"])
		}
		if message, ok := updated.GatewayResponse["message"].(string); !ok || message != "Refund is being processed by payment gateway" {
			t.Errorf("Expected message in gateway response, got %v", updated.GatewayResponse["message"])
		}
	}
}

// Helper functions

func generateRandomString(length int) string {
	return randomHex(length / 2)
}

func mustParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}
