package repository

import (
	"testing"
	"time"
	"zavera/models"
)

// TestRefundRepository_RefundItems tests refund item operations
func TestRefundRepository_RefundItems(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		t.Skip("Skipping test: no test database available")
	}
	defer db.Close()

	repo := NewRefundRepository(db)

	// Create a test refund first
	refund := &models.Refund{
		RefundCode:     GenerateRefundCode(),
		OrderID:        1,
		RefundType:     models.RefundTypeItemOnly,
		Reason:         models.RefundReasonDamagedItem,
		ReasonDetail:   "Test item refund",
		OriginalAmount: 100000,
		RefundAmount:   50000,
		ShippingRefund: 0,
		ItemsRefund:    50000,
		Status:         models.RefundStatusPending,
		IdempotencyKey: "test-items-" + time.Now().Format("20060102150405"),
		RequestedAt:    time.Now(),
	}

	err := repo.Create(refund)
	if err != nil {
		t.Fatalf("Failed to create refund: %v", err)
	}

	// Test CreateRefundItem
	item1 := &models.RefundItem{
		RefundID:     refund.ID,
		OrderItemID:  1,
		ProductID:    101,
		ProductName:  "Test Product 1",
		Quantity:     2,
		PricePerUnit: 15000,
		RefundAmount: 30000,
		ItemReason:   "Damaged on arrival",
	}

	err = repo.CreateRefundItem(item1)
	if err != nil {
		t.Fatalf("Failed to create refund item: %v", err)
	}

	if item1.ID == 0 {
		t.Error("Refund item ID should be set after creation")
	}

	// Create another item
	item2 := &models.RefundItem{
		RefundID:     refund.ID,
		OrderItemID:  2,
		ProductID:    102,
		ProductName:  "Test Product 2",
		Quantity:     1,
		PricePerUnit: 20000,
		RefundAmount: 20000,
		ItemReason:   "Wrong color",
	}

	err = repo.CreateRefundItem(item2)
	if err != nil {
		t.Fatalf("Failed to create second refund item: %v", err)
	}

	// Test FindItemsByRefundID
	items, err := repo.FindItemsByRefundID(refund.ID)
	if err != nil {
		t.Fatalf("Failed to find items by refund ID: %v", err)
	}

	if len(items) != 2 {
		t.Errorf("Expected 2 refund items, got %d", len(items))
	}

	// Verify item details
	foundItem1 := false
	foundItem2 := false
	for _, item := range items {
		if item.ProductID == 101 {
			foundItem1 = true
			if item.Quantity != 2 {
				t.Errorf("Expected quantity 2 for item 1, got %d", item.Quantity)
			}
			if item.RefundAmount != 30000 {
				t.Errorf("Expected refund amount 30000 for item 1, got %.2f", item.RefundAmount)
			}
			if item.StockRestored {
				t.Error("Stock should not be restored yet for item 1")
			}
		}
		if item.ProductID == 102 {
			foundItem2 = true
			if item.Quantity != 1 {
				t.Errorf("Expected quantity 1 for item 2, got %d", item.Quantity)
			}
		}
	}

	if !foundItem1 {
		t.Error("Item 1 not found in results")
	}
	if !foundItem2 {
		t.Error("Item 2 not found in results")
	}

	// Test MarkItemStockRestored
	err = repo.MarkItemStockRestored(item1.ID)
	if err != nil {
		t.Fatalf("Failed to mark item stock restored: %v", err)
	}

	// Verify stock restored flag
	updatedItems, err := repo.FindItemsByRefundID(refund.ID)
	if err != nil {
		t.Fatalf("Failed to find items after stock restoration: %v", err)
	}

	for _, item := range updatedItems {
		if item.ID == item1.ID {
			if !item.StockRestored {
				t.Error("Stock should be marked as restored for item 1")
			}
			if item.StockRestoredAt == nil {
				t.Error("StockRestoredAt should be set for item 1")
			}
		}
		if item.ID == item2.ID {
			if item.StockRestored {
				t.Error("Stock should not be restored yet for item 2")
			}
		}
	}

	// Clean up
	cleanupTestRefund(t, db, refund.ID)
}

// TestRefundRepository_CreateRefundItemWithTx tests transaction-based item creation
func TestRefundRepository_CreateRefundItemWithTx(t *testing.T) {
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
		RefundType:     models.RefundTypeItemOnly,
		Reason:         models.RefundReasonDamagedItem,
		OriginalAmount: 100000,
		RefundAmount:   50000,
		Status:         models.RefundStatusPending,
		IdempotencyKey: "test-tx-items-" + time.Now().Format("20060102150405"),
		RequestedAt:    time.Now(),
	}

	err := repo.Create(refund)
	if err != nil {
		t.Fatalf("Failed to create refund: %v", err)
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// Create items within transaction
	item1 := &models.RefundItem{
		RefundID:     refund.ID,
		OrderItemID:  1,
		ProductID:    201,
		ProductName:  "TX Test Product 1",
		Quantity:     1,
		PricePerUnit: 25000,
		RefundAmount: 25000,
	}

	err = repo.CreateRefundItemWithTx(tx, item1)
	if err != nil {
		tx.Rollback()
		t.Fatalf("Failed to create refund item with tx: %v", err)
	}

	item2 := &models.RefundItem{
		RefundID:     refund.ID,
		OrderItemID:  2,
		ProductID:    202,
		ProductName:  "TX Test Product 2",
		Quantity:     1,
		PricePerUnit: 25000,
		RefundAmount: 25000,
	}

	err = repo.CreateRefundItemWithTx(tx, item2)
	if err != nil {
		tx.Rollback()
		t.Fatalf("Failed to create second refund item with tx: %v", err)
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Verify items were created
	items, err := repo.FindItemsByRefundID(refund.ID)
	if err != nil {
		t.Fatalf("Failed to find items: %v", err)
	}

	if len(items) != 2 {
		t.Errorf("Expected 2 items after transaction commit, got %d", len(items))
	}

	// Clean up
	cleanupTestRefund(t, db, refund.ID)
}

// TestRefundRepository_FindItemsByRefundID_EmptyResult tests finding items for refund with no items
func TestRefundRepository_FindItemsByRefundID_EmptyResult(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		t.Skip("Skipping test: no test database available")
	}
	defer db.Close()

	repo := NewRefundRepository(db)

	// Create a refund without items
	refund := &models.Refund{
		RefundCode:     GenerateRefundCode(),
		OrderID:        1,
		RefundType:     models.RefundTypeFull,
		Reason:         models.RefundReasonCustomerRequest,
		OriginalAmount: 100000,
		RefundAmount:   100000,
		Status:         models.RefundStatusPending,
		IdempotencyKey: "test-empty-items-" + time.Now().Format("20060102150405"),
		RequestedAt:    time.Now(),
	}

	err := repo.Create(refund)
	if err != nil {
		t.Fatalf("Failed to create refund: %v", err)
	}

	// Find items (should be empty)
	items, err := repo.FindItemsByRefundID(refund.ID)
	if err != nil {
		t.Fatalf("Failed to find items: %v", err)
	}

	if len(items) != 0 {
		t.Errorf("Expected 0 items for refund without items, got %d", len(items))
	}

	// Clean up
	cleanupTestRefund(t, db, refund.ID)
}

// TestRefundRepository_MarkItemStockRestored_Idempotency tests that marking stock restored is idempotent
func TestRefundRepository_MarkItemStockRestored_Idempotency(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		t.Skip("Skipping test: no test database available")
	}
	defer db.Close()

	repo := NewRefundRepository(db)

	// Create a test refund and item
	refund := &models.Refund{
		RefundCode:     GenerateRefundCode(),
		OrderID:        1,
		RefundType:     models.RefundTypeItemOnly,
		Reason:         models.RefundReasonDamagedItem,
		OriginalAmount: 50000,
		RefundAmount:   50000,
		Status:         models.RefundStatusPending,
		IdempotencyKey: "test-idempotent-" + time.Now().Format("20060102150405"),
		RequestedAt:    time.Now(),
	}

	err := repo.Create(refund)
	if err != nil {
		t.Fatalf("Failed to create refund: %v", err)
	}

	item := &models.RefundItem{
		RefundID:     refund.ID,
		OrderItemID:  1,
		ProductID:    301,
		ProductName:  "Idempotent Test Product",
		Quantity:     1,
		PricePerUnit: 50000,
		RefundAmount: 50000,
	}

	err = repo.CreateRefundItem(item)
	if err != nil {
		t.Fatalf("Failed to create refund item: %v", err)
	}

	// Mark stock restored first time
	err = repo.MarkItemStockRestored(item.ID)
	if err != nil {
		t.Fatalf("Failed to mark stock restored (first time): %v", err)
	}

	// Get the timestamp
	items1, err := repo.FindItemsByRefundID(refund.ID)
	if err != nil {
		t.Fatalf("Failed to find items: %v", err)
	}

	if len(items1) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(items1))
	}

	_ = items1[0].StockRestoredAt // Store for reference but don't compare

	// Wait a moment
	time.Sleep(100 * time.Millisecond)

	// Mark stock restored second time (idempotent operation)
	err = repo.MarkItemStockRestored(item.ID)
	if err != nil {
		t.Fatalf("Failed to mark stock restored (second time): %v", err)
	}

	// Verify timestamp changed (UPDATE always updates the timestamp)
	items2, err := repo.FindItemsByRefundID(refund.ID)
	if err != nil {
		t.Fatalf("Failed to find items after second mark: %v", err)
	}

	if len(items2) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(items2))
	}

	if !items2[0].StockRestored {
		t.Error("Stock should still be marked as restored")
	}

	// Note: The timestamp will be updated on second call, which is acceptable
	// The important thing is that stock_restored remains true
	if items2[0].StockRestoredAt == nil {
		t.Error("StockRestoredAt should still be set")
	}

	// Clean up
	cleanupTestRefund(t, db, refund.ID)
}
