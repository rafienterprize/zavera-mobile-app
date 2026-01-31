package repository

import (
	"testing"
	"time"
	"zavera/models"
)

// TestRefundRepository_CompleteWorkflow tests the complete refund workflow
// including refund creation, items, and status history
func TestRefundRepository_CompleteWorkflow(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		t.Skip("Skipping test: no test database available")
	}
	defer db.Close()

	repo := NewRefundRepository(db)

	// Step 1: Create a refund
	refund := &models.Refund{
		RefundCode:     GenerateRefundCode(),
		OrderID:        1,
		RefundType:     models.RefundTypeItemOnly,
		Reason:         models.RefundReasonDamagedItem,
		ReasonDetail:   "Complete workflow test",
		OriginalAmount: 100000,
		RefundAmount:   50000,
		ShippingRefund: 0,
		ItemsRefund:    50000,
		Status:         models.RefundStatusPending,
		IdempotencyKey: "test-workflow-" + time.Now().Format("20060102150405"),
		RequestedAt:    time.Now(),
	}

	err := repo.Create(refund)
	if err != nil {
		t.Fatalf("Failed to create refund: %v", err)
	}

	// Step 2: Record initial status
	err = repo.RecordStatusChange(refund.ID, "", models.RefundStatusPending, "SYSTEM", "Refund created")
	if err != nil {
		t.Fatalf("Failed to record initial status: %v", err)
	}

	// Step 3: Add refund items
	item1 := &models.RefundItem{
		RefundID:     refund.ID,
		OrderItemID:  1,
		ProductID:    101,
		ProductName:  "Damaged Product",
		Quantity:     2,
		PricePerUnit: 25000,
		RefundAmount: 50000,
		ItemReason:   "Damaged on arrival",
	}

	err = repo.CreateRefundItem(item1)
	if err != nil {
		t.Fatalf("Failed to create refund item: %v", err)
	}

	// Step 4: Update status to PROCESSING
	err = repo.UpdateStatus(refund.ID, models.RefundStatusProcessing, map[string]any{
		"message": "Processing refund",
	})
	if err != nil {
		t.Fatalf("Failed to update status to PROCESSING: %v", err)
	}

	err = repo.RecordStatusChange(refund.ID, models.RefundStatusPending, models.RefundStatusProcessing, "admin@example.com", "Admin started processing")
	if err != nil {
		t.Fatalf("Failed to record status change to PROCESSING: %v", err)
	}

	// Step 5: Mark as completed
	err = repo.MarkCompleted(refund.ID, "GATEWAY-REF-WORKFLOW-123", map[string]any{
		"status":     "success",
		"refund_id":  "GATEWAY-REF-WORKFLOW-123",
		"message":    "Refund processed successfully",
		"gateway":    "midtrans",
		"timestamp":  time.Now().Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("Failed to mark refund as completed: %v", err)
	}

	err = repo.RecordStatusChange(refund.ID, models.RefundStatusProcessing, models.RefundStatusCompleted, "SYSTEM", "Gateway refund successful")
	if err != nil {
		t.Fatalf("Failed to record status change to COMPLETED: %v", err)
	}

	// Step 6: Mark stock as restored
	err = repo.MarkItemStockRestored(item1.ID)
	if err != nil {
		t.Fatalf("Failed to mark stock restored: %v", err)
	}

	// Verification: Check final refund state
	finalRefund, err := repo.FindByID(refund.ID)
	if err != nil {
		t.Fatalf("Failed to find final refund: %v", err)
	}

	if finalRefund.Status != models.RefundStatusCompleted {
		t.Errorf("Expected final status COMPLETED, got %s", finalRefund.Status)
	}

	if finalRefund.GatewayRefundID != "GATEWAY-REF-WORKFLOW-123" {
		t.Errorf("Expected gateway refund ID GATEWAY-REF-WORKFLOW-123, got %s", finalRefund.GatewayRefundID)
	}

	if finalRefund.CompletedAt == nil {
		t.Error("CompletedAt should be set")
	}

	// Verification: Check refund items
	items, err := repo.FindItemsByRefundID(refund.ID)
	if err != nil {
		t.Fatalf("Failed to find refund items: %v", err)
	}

	if len(items) != 1 {
		t.Errorf("Expected 1 refund item, got %d", len(items))
	}

	if len(items) > 0 {
		if !items[0].StockRestored {
			t.Error("Stock should be marked as restored")
		}
		if items[0].StockRestoredAt == nil {
			t.Error("StockRestoredAt should be set")
		}
	}

	// Verification: Check status history
	history, err := repo.GetStatusHistory(refund.ID)
	if err != nil {
		t.Fatalf("Failed to get status history: %v", err)
	}

	if len(history) != 3 {
		t.Errorf("Expected 3 status history entries, got %d", len(history))
	}

	// Verify status progression
	expectedStatuses := []models.RefundStatus{
		models.RefundStatusPending,
		models.RefundStatusProcessing,
		models.RefundStatusCompleted,
	}

	for i, h := range history {
		if i < len(expectedStatuses) {
			if h.NewStatus != expectedStatuses[i] {
				t.Errorf("Expected status %s at position %d, got %s", expectedStatuses[i], i, h.NewStatus)
			}
		}
	}

	// Verify actors
	if history[0].Actor != "SYSTEM" {
		t.Errorf("Expected first actor to be SYSTEM, got %s", history[0].Actor)
	}

	if history[1].Actor != "admin@example.com" {
		t.Errorf("Expected second actor to be admin@example.com, got %s", history[1].Actor)
	}

	if history[2].Actor != "SYSTEM" {
		t.Errorf("Expected third actor to be SYSTEM, got %s", history[2].Actor)
	}

	// Clean up
	cleanupTestRefund(t, db, refund.ID)
}

// TestRefundRepository_TransactionWorkflow tests refund creation with transaction
func TestRefundRepository_TransactionWorkflow(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		t.Skip("Skipping test: no test database available")
	}
	defer db.Close()

	repo := NewRefundRepository(db)

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// Create refund within transaction
	refund := &models.Refund{
		RefundCode:     GenerateRefundCode(),
		OrderID:        1,
		RefundType:     models.RefundTypeItemOnly,
		Reason:         models.RefundReasonDamagedItem,
		OriginalAmount: 75000,
		RefundAmount:   75000,
		Status:         models.RefundStatusPending,
		IdempotencyKey: "test-tx-workflow-" + time.Now().Format("20060102150405"),
		RequestedAt:    time.Now(),
	}

	err = repo.CreateWithTx(tx, refund)
	if err != nil {
		tx.Rollback()
		t.Fatalf("Failed to create refund with tx: %v", err)
	}

	// Create refund items within same transaction
	items := []*models.RefundItem{
		{
			RefundID:     refund.ID,
			OrderItemID:  1,
			ProductID:    201,
			ProductName:  "TX Product 1",
			Quantity:     1,
			PricePerUnit: 50000,
			RefundAmount: 50000,
		},
		{
			RefundID:     refund.ID,
			OrderItemID:  2,
			ProductID:    202,
			ProductName:  "TX Product 2",
			Quantity:     1,
			PricePerUnit: 25000,
			RefundAmount: 25000,
		},
	}

	for _, item := range items {
		err = repo.CreateRefundItemWithTx(tx, item)
		if err != nil {
			tx.Rollback()
			t.Fatalf("Failed to create refund item with tx: %v", err)
		}
	}

	// Update status within transaction
	err = repo.UpdateStatusWithTx(tx, refund.ID, models.RefundStatusProcessing, map[string]any{
		"message": "Processing in transaction",
	})
	if err != nil {
		tx.Rollback()
		t.Fatalf("Failed to update status with tx: %v", err)
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Verify everything was committed
	foundRefund, err := repo.FindByID(refund.ID)
	if err != nil {
		t.Fatalf("Failed to find refund after commit: %v", err)
	}

	if foundRefund.Status != models.RefundStatusProcessing {
		t.Errorf("Expected status PROCESSING, got %s", foundRefund.Status)
	}

	foundItems, err := repo.FindItemsByRefundID(refund.ID)
	if err != nil {
		t.Fatalf("Failed to find items after commit: %v", err)
	}

	if len(foundItems) != 2 {
		t.Errorf("Expected 2 items after commit, got %d", len(foundItems))
	}

	// Clean up
	cleanupTestRefund(t, db, refund.ID)
}

// TestRefundRepository_RollbackWorkflow tests transaction rollback
func TestRefundRepository_RollbackWorkflow(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		t.Skip("Skipping test: no test database available")
	}
	defer db.Close()

	repo := NewRefundRepository(db)

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// Create refund within transaction
	refund := &models.Refund{
		RefundCode:     GenerateRefundCode(),
		OrderID:        1,
		RefundType:     models.RefundTypeFull,
		Reason:         models.RefundReasonCustomerRequest,
		OriginalAmount: 100000,
		RefundAmount:   100000,
		Status:         models.RefundStatusPending,
		IdempotencyKey: "test-rollback-" + time.Now().Format("20060102150405"),
		RequestedAt:    time.Now(),
	}

	err = repo.CreateWithTx(tx, refund)
	if err != nil {
		tx.Rollback()
		t.Fatalf("Failed to create refund with tx: %v", err)
	}

	refundID := refund.ID

	// Rollback transaction
	err = tx.Rollback()
	if err != nil {
		t.Fatalf("Failed to rollback transaction: %v", err)
	}

	// Verify refund was not created
	_, err = repo.FindByID(refundID)
	if err == nil {
		t.Error("Expected error when finding rolled back refund, got nil")
	}

	// Verify by idempotency key
	foundByKey, err := repo.FindByIdempotencyKey(refund.IdempotencyKey)
	if err != nil {
		t.Fatalf("Unexpected error finding by idempotency key: %v", err)
	}

	if foundByKey != nil {
		t.Error("Expected nil when finding rolled back refund by idempotency key")
	}
}
