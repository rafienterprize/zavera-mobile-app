package repository

import (
	"database/sql"
	"testing"
	"time"
	"zavera/models"

	_ "github.com/lib/pq"
)

// TestGenerateRefundCode tests the refund code generation
func TestGenerateRefundCode(t *testing.T) {
	code1 := GenerateRefundCode()
	code2 := GenerateRefundCode()

	// Check format
	if len(code1) < 15 {
		t.Errorf("Generated code too short: %s", code1)
	}

	// Check uniqueness
	if code1 == code2 {
		t.Errorf("Generated codes should be unique: %s == %s", code1, code2)
	}

	// Check prefix
	expectedPrefix := "RFD-" + time.Now().Format("20060102")
	if len(code1) < len(expectedPrefix) || code1[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("Code should start with %s, got: %s", expectedPrefix, code1)
	}
}

// TestRefundRepository_CreateAndFind tests basic CRUD operations
// Note: This test requires a test database connection
func TestRefundRepository_CreateAndFind(t *testing.T) {
	// Skip if no test database is available
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
		PaymentID:      nil, // Test nullable field
		RefundType:     models.RefundTypeFull,
		Reason:         models.RefundReasonCustomerRequest,
		ReasonDetail:   "Test refund",
		OriginalAmount: 100000,
		RefundAmount:   100000,
		ShippingRefund: 10000,
		ItemsRefund:    90000,
		Status:         models.RefundStatusPending,
		IdempotencyKey: "test-idempotency-key-" + time.Now().Format("20060102150405"),
		RequestedBy:    nil, // Test nullable field
		RequestedAt:    time.Now(),
	}

	// Test Create
	err := repo.Create(refund)
	if err != nil {
		t.Fatalf("Failed to create refund: %v", err)
	}

	if refund.ID == 0 {
		t.Error("Refund ID should be set after creation")
	}

	// Test FindByID
	found, err := repo.FindByID(refund.ID)
	if err != nil {
		t.Fatalf("Failed to find refund by ID: %v", err)
	}

	if found.RefundCode != refund.RefundCode {
		t.Errorf("Expected refund code %s, got %s", refund.RefundCode, found.RefundCode)
	}

	// Test FindByCode
	foundByCode, err := repo.FindByCode(refund.RefundCode)
	if err != nil {
		t.Fatalf("Failed to find refund by code: %v", err)
	}

	if foundByCode.ID != refund.ID {
		t.Errorf("Expected refund ID %d, got %d", refund.ID, foundByCode.ID)
	}

	// Test FindByIdempotencyKey
	foundByKey, err := repo.FindByIdempotencyKey(refund.IdempotencyKey)
	if err != nil {
		t.Fatalf("Failed to find refund by idempotency key: %v", err)
	}

	if foundByKey == nil {
		t.Error("Expected to find refund by idempotency key")
	} else if foundByKey.ID != refund.ID {
		t.Errorf("Expected refund ID %d, got %d", refund.ID, foundByKey.ID)
	}

	// Test FindByOrderID
	refunds, err := repo.FindByOrderID(refund.OrderID)
	if err != nil {
		t.Fatalf("Failed to find refunds by order ID: %v", err)
	}

	if len(refunds) == 0 {
		t.Error("Expected to find at least one refund for order")
	}

	// Clean up
	cleanupTestRefund(t, db, refund.ID)
}

// TestRefundRepository_UpdateStatus tests status update operations
func TestRefundRepository_UpdateStatus(t *testing.T) {
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
		IdempotencyKey: "test-status-" + time.Now().Format("20060102150405"),
		RequestedAt:    time.Now(),
	}

	err := repo.Create(refund)
	if err != nil {
		t.Fatalf("Failed to create refund: %v", err)
	}

	// Test UpdateStatus
	gatewayResponse := map[string]any{
		"status": "processing",
		"message": "Refund is being processed",
	}
	err = repo.UpdateStatus(refund.ID, models.RefundStatusProcessing, gatewayResponse)
	if err != nil {
		t.Fatalf("Failed to update status: %v", err)
	}

	// Verify status was updated
	updated, err := repo.FindByID(refund.ID)
	if err != nil {
		t.Fatalf("Failed to find updated refund: %v", err)
	}

	if updated.Status != models.RefundStatusProcessing {
		t.Errorf("Expected status %s, got %s", models.RefundStatusProcessing, updated.Status)
	}

	// Test MarkCompleted
	err = repo.MarkCompleted(refund.ID, "GATEWAY-REF-123", map[string]any{
		"status": "success",
		"refund_id": "GATEWAY-REF-123",
	})
	if err != nil {
		t.Fatalf("Failed to mark completed: %v", err)
	}

	completed, err := repo.FindByID(refund.ID)
	if err != nil {
		t.Fatalf("Failed to find completed refund: %v", err)
	}

	if completed.Status != models.RefundStatusCompleted {
		t.Errorf("Expected status %s, got %s", models.RefundStatusCompleted, completed.Status)
	}

	if completed.GatewayRefundID != "GATEWAY-REF-123" {
		t.Errorf("Expected gateway refund ID GATEWAY-REF-123, got %s", completed.GatewayRefundID)
	}

	// Clean up
	cleanupTestRefund(t, db, refund.ID)
}

// TestRefundRepository_StatusHistory tests status history recording
func TestRefundRepository_StatusHistory(t *testing.T) {
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
		IdempotencyKey: "test-history-" + time.Now().Format("20060102150405"),
		RequestedAt:    time.Now(),
	}

	err := repo.Create(refund)
	if err != nil {
		t.Fatalf("Failed to create refund: %v", err)
	}

	// Record status changes
	err = repo.RecordStatusChange(refund.ID, "", models.RefundStatusPending, "system", "Refund created")
	if err != nil {
		t.Fatalf("Failed to record initial status: %v", err)
	}

	err = repo.RecordStatusChange(refund.ID, models.RefundStatusPending, models.RefundStatusProcessing, "admin", "Processing refund")
	if err != nil {
		t.Fatalf("Failed to record status change: %v", err)
	}

	// Get status history
	history, err := repo.GetStatusHistory(refund.ID)
	if err != nil {
		t.Fatalf("Failed to get status history: %v", err)
	}

	if len(history) < 2 {
		t.Errorf("Expected at least 2 status history entries, got %d", len(history))
	}

	// Verify first entry
	if history[0].NewStatus != models.RefundStatusPending {
		t.Errorf("Expected first status to be PENDING, got %s", history[0].NewStatus)
	}

	// Verify second entry
	if len(history) > 1 {
		if history[1].NewStatus != models.RefundStatusProcessing {
			t.Errorf("Expected second status to be PROCESSING, got %s", history[1].NewStatus)
		}
		if history[1].OldStatus == nil || *history[1].OldStatus != string(models.RefundStatusPending) {
			t.Error("Expected old status to be PENDING")
		}
	}

	// Clean up
	cleanupTestRefund(t, db, refund.ID)
}

// getTestDB returns a test database connection
// Returns nil if no test database is configured
func getTestDB(t *testing.T) *sql.DB {
	// Try to connect to test database
	// You can set TEST_DATABASE_URL environment variable for CI/CD
	// For now, we'll skip if not available
	connStr := "postgres://postgres:postgres@localhost:5432/zavera_test?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Logf("Could not connect to test database: %v", err)
		return nil
	}

	// Test connection
	err = db.Ping()
	if err != nil {
		t.Logf("Could not ping test database: %v", err)
		db.Close()
		return nil
	}

	return db
}

// cleanupTestRefund removes test data
func cleanupTestRefund(t *testing.T, db *sql.DB, refundID int) {
	// Delete status history first (foreign key constraint)
	_, err := db.Exec("DELETE FROM refund_status_history WHERE refund_id = $1", refundID)
	if err != nil {
		t.Logf("Warning: Failed to cleanup status history: %v", err)
	}

	// Delete refund items
	_, err = db.Exec("DELETE FROM refund_items WHERE refund_id = $1", refundID)
	if err != nil {
		t.Logf("Warning: Failed to cleanup refund items: %v", err)
	}

	// Delete refund
	_, err = db.Exec("DELETE FROM refunds WHERE id = $1", refundID)
	if err != nil {
		t.Logf("Warning: Failed to cleanup refund: %v", err)
	}
}
