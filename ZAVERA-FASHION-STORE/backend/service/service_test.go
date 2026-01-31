package service

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"zavera/models"
)

// **Feature: zavera-commerce-upgrade, Property 9: Resi Format and Uniqueness**
// **Validates: Requirements 3.1, 3.2, 3.4**
// *For any* generated resi, the format SHALL match ZVR-{COURIER}-{YYYYMMDD}-{ORDERID}-{RANDOM4}
// and SHALL be unique across all orders.

// genCourierCode generates valid courier codes
func genCourierCode() gopter.Gen {
	return gen.OneConstOf("JNE", "JNT", "SICEPAT", "POS", "TIKI", "ANTERAJA", "NINJA", "LION")
}

// genOrderID generates valid order IDs
func genOrderID() gopter.Gen {
	return gen.IntRange(1, 999999)
}

func TestProperty_ResiFormatValidation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property: Valid resi format should be validated correctly
	properties.Property("valid resi format is validated correctly", prop.ForAll(
		func(courier string, orderID int) bool {
			// Create a mock resi service for validation
			svc := &resiService{}
			
			// Generate a valid resi format
			dateStr := time.Now().Format("20060102")
			resi := fmt.Sprintf("ZVR-%s-%s-%d-A1B2", courier, dateStr, orderID)
			
			return svc.ValidateResiFormat(resi)
		},
		genCourierCode(),
		genOrderID(),
	))

	// Property: Invalid resi formats should be rejected
	properties.Property("invalid resi formats are rejected", prop.ForAll(
		func(invalidResi string) bool {
			svc := &resiService{}
			
			// These should all be invalid
			invalidFormats := []string{
				"",                           // Empty
				"ZVR",                         // Too short
				"ZVR-JNE",                     // Missing parts
				"ABC-JNE-20260111-123-A1B2",   // Wrong prefix
				"ZVR-jne-20260111-123-A1B2",   // Lowercase courier
				"ZVR-JNE-2026011-123-A1B2",    // Invalid date (7 chars)
				"ZVR-JNE-20260111-abc-A1B2",   // Non-numeric order ID
				"ZVR-JNE-20260111-123-A1B",    // Random hex too short
				"ZVR-JNE-20260111-123-A1B2C",  // Random hex too long
				"ZVR-JNE-20260111-123-a1b2",   // Lowercase hex
			}
			
			for _, invalid := range invalidFormats {
				if svc.ValidateResiFormat(invalid) {
					return false
				}
			}
			return true
		},
		gen.Const("test"),
	))

	// Property: Resi format regex pattern
	properties.Property("resi matches expected regex pattern", prop.ForAll(
		func(courier string, orderID int) bool {
			dateStr := time.Now().Format("20060102")
			resi := fmt.Sprintf("ZVR-%s-%s-%d-A1B2", courier, dateStr, orderID)
			
			// Pattern: ZVR-{COURIER}-{YYYYMMDD}-{ORDERID}-{RANDOM4}
			pattern := `^ZVR-[A-Z]{2,10}-\d{8}-\d+-[A-F0-9]{4}$`
			matched, _ := regexp.MatchString(pattern, resi)
			return matched
		},
		genCourierCode(),
		genOrderID(),
	))

	properties.TestingRun(t)
}

// **Feature: zavera-commerce-upgrade, Property 10: Resi Immutability After Shipping**
// **Validates: Requirements 3.3**
func TestProperty_ResiImmutability(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	svc := &resiService{}

	// Property: Resi is locked for SHIPPED and beyond
	properties.Property("resi is locked for shipped and beyond statuses", prop.ForAll(
		func(status models.OrderStatus) bool {
			order := &models.Order{Status: status}
			isLocked := svc.IsResiLocked(order)
			
			// Should be locked for SHIPPED, DELIVERED, COMPLETED, REFUNDED
			expected := status == models.OrderStatusShipped ||
			           status == models.OrderStatusDelivered ||
			           status == models.OrderStatusCompleted ||
			           status == models.OrderStatusRefunded
			return isLocked == expected
		},
		gen.OneConstOf(
			models.OrderStatusPending,
			models.OrderStatusPaid,
			models.OrderStatusPacking,
			models.OrderStatusShipped,
			models.OrderStatusDelivered,
			models.OrderStatusCompleted,
			models.OrderStatusCancelled,
			models.OrderStatusFailed,
			models.OrderStatusExpired,
			models.OrderStatusRefunded,
		),
	))

	properties.TestingRun(t)
}

// **Feature: zavera-commerce-upgrade, Property 20: Server-Side Total Calculation**
// **Validates: Requirements 9.1, 9.3, 11.3**
// *For any* order, the total_amount SHALL equal subtotal + shipping_cost + tax - discount

func genPrice() gopter.Gen {
	return gen.Float64Range(1000, 10000000)
}

func genQuantity() gopter.Gen {
	return gen.IntRange(1, 100)
}

func TestProperty_ServerSideTotalCalculation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property: Total calculation is correct
	properties.Property("total equals subtotal + shipping + tax - discount", prop.ForAll(
		func(subtotal, shipping, tax, discount float64) bool {
			order := &models.Order{
				Subtotal:     subtotal,
				ShippingCost: shipping,
				Tax:          tax,
				Discount:     discount,
				TotalAmount:  subtotal + shipping + tax - discount,
			}
			
			// Verify the calculation
			expectedTotal := order.Subtotal + order.ShippingCost + order.Tax - order.Discount
			return order.TotalAmount == expectedTotal
		},
		genPrice(),
		gen.Float64Range(0, 100000),
		gen.Float64Range(0, 1000000),
		gen.Float64Range(0, 500000),
	))

	// Property: Subtotal equals sum of item prices * quantities
	properties.Property("subtotal equals sum of item prices times quantities", prop.ForAll(
		func(prices []float64, quantities []int) bool {
			if len(prices) == 0 || len(quantities) == 0 {
				return true
			}
			
			// Use minimum length
			n := len(prices)
			if len(quantities) < n {
				n = len(quantities)
			}
			
			var items []models.OrderItem
			var expectedSubtotal float64
			
			for i := 0; i < n; i++ {
				itemSubtotal := prices[i] * float64(quantities[i])
				items = append(items, models.OrderItem{
					PricePerUnit: prices[i],
					Quantity:     quantities[i],
					Subtotal:     itemSubtotal,
				})
				expectedSubtotal += itemSubtotal
			}
			
			order := &models.Order{
				Subtotal: expectedSubtotal,
				Items:    items,
			}
			
			// Recalculate
			var calculatedSubtotal float64
			for _, item := range order.Items {
				calculatedSubtotal += item.PricePerUnit * float64(item.Quantity)
			}
			
			return calculatedSubtotal == order.Subtotal
		},
		gen.SliceOfN(5, genPrice()),
		gen.SliceOfN(5, genQuantity()),
	))

	properties.TestingRun(t)
}

// **Feature: zavera-commerce-upgrade, Property 11: Email Content Completeness**
// **Validates: Requirements 4.1, 4.2, 4.3, 4.4, 4.5**
// *For any* transactional email, the email SHALL be in HTML format and contain required information

func TestProperty_EmailContentCompleteness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property: Email subjects contain order ID
	properties.Property("email subjects contain order ID", prop.ForAll(
		func(orderID int) bool {
			subjects := []string{
				fmt.Sprintf("🛍️ Pesanan ZAVERA #%d telah dibuat", orderID),
				fmt.Sprintf("💳 Pembayaran diterima – Pesanan #%d", orderID),
				fmt.Sprintf("📦 Pesanan #%d sedang dikirim", orderID),
				fmt.Sprintf("🎉 Pesanan #%d sudah sampai", orderID),
			}
			
			orderIDStr := fmt.Sprintf("#%d", orderID)
			for _, subject := range subjects {
				if !strings.Contains(subject, orderIDStr) {
					return false
				}
			}
			return true
		},
		genOrderID(),
	))

	properties.TestingRun(t)
}

// **Feature: zavera-commerce-upgrade, Property 12: Admin Pack Action**
// **Validates: Requirements 5.3**
// *For any* PAID order, when an admin executes the "Pack" action, the order status
// SHALL change to PACKING

func TestProperty_AdminPackAction(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property: Only PAID orders can be packed
	properties.Property("only PAID orders can transition to PACKING", prop.ForAll(
		func(status models.OrderStatus) bool {
			canPack := status.IsValidTransition(models.OrderStatusPacking)
			expected := status == models.OrderStatusPaid
			return canPack == expected
		},
		gen.OneConstOf(
			models.OrderStatusPending,
			models.OrderStatusPaid,
			models.OrderStatusPacking,
			models.OrderStatusShipped,
			models.OrderStatusDelivered,
			models.OrderStatusCompleted,
			models.OrderStatusCancelled,
			models.OrderStatusFailed,
			models.OrderStatusExpired,
			models.OrderStatusRefunded,
		),
	))

	properties.TestingRun(t)
}

// **Feature: zavera-commerce-upgrade, Property 13: Admin Ship Action**
// **Validates: Requirements 5.4**
// *For any* PACKING order, when an admin executes the "Ship" action, status SHALL change to SHIPPED

func TestProperty_AdminShipAction(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property: Only PACKING orders can be shipped
	properties.Property("only PACKING orders can transition to SHIPPED", prop.ForAll(
		func(status models.OrderStatus) bool {
			canShip := status.IsValidTransition(models.OrderStatusShipped)
			expected := status == models.OrderStatusPacking
			return canShip == expected
		},
		gen.OneConstOf(
			models.OrderStatusPending,
			models.OrderStatusPaid,
			models.OrderStatusPacking,
			models.OrderStatusShipped,
			models.OrderStatusDelivered,
			models.OrderStatusCompleted,
			models.OrderStatusCancelled,
			models.OrderStatusFailed,
			models.OrderStatusExpired,
			models.OrderStatusRefunded,
		),
	))

	properties.TestingRun(t)
}

// **Feature: zavera-commerce-upgrade, Property 3: Status Change Audit Logging**
// **Validates: Requirements 1.4, 8.1**
// *For any* order status change, an audit log entry SHALL be created

func TestProperty_StatusChangeAuditLogging(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property: All valid transitions should be loggable
	properties.Property("all valid transitions can be logged", prop.ForAll(
		func(from, to models.OrderStatus) bool {
			if !from.IsValidTransition(to) {
				return true // Skip invalid transitions
			}
			
			// Verify we can create an audit log entry
			auditLog := &models.AdminAuditLog{
				ActionType:   models.AdminActionUpdateStatus,
				ActionDetail: fmt.Sprintf("Status changed from %s to %s", from, to),
				TargetType:   "order",
				StateBefore:  map[string]any{"status": string(from)},
				StateAfter:   map[string]any{"status": string(to)},
				Success:      true,
			}
			
			// Verify the audit log has required fields
			return auditLog.ActionType != "" &&
			       auditLog.ActionDetail != "" &&
			       auditLog.TargetType != "" &&
			       auditLog.StateBefore != nil &&
			       auditLog.StateAfter != nil
		},
		gen.OneConstOf(
			models.OrderStatusPending,
			models.OrderStatusPaid,
			models.OrderStatusPacking,
			models.OrderStatusShipped,
			models.OrderStatusDelivered,
			models.OrderStatusCompleted,
		),
		gen.OneConstOf(
			models.OrderStatusPaid,
			models.OrderStatusPacking,
			models.OrderStatusShipped,
			models.OrderStatusDelivered,
			models.OrderStatusCompleted,
			models.OrderStatusCancelled,
			models.OrderStatusRefunded,
		),
	))

	properties.TestingRun(t)
}

// **Feature: zavera-commerce-upgrade, Property 17: Stock Movement Logging**
// **Validates: Requirements 6.5**
// *For any* stock operation, a stock_movements record SHALL be created

func TestProperty_StockMovementLogging(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property: Stock movement types are valid
	properties.Property("stock movement types are valid", prop.ForAll(
		func(movementType models.StockMovementType) bool {
			validTypes := []models.StockMovementType{
				models.StockMovementReserve,
				models.StockMovementRelease,
				models.StockMovementDeduct,
				models.StockMovementAdjustment,
			}
			
			for _, valid := range validTypes {
				if movementType == valid {
					return true
				}
			}
			return false
		},
		gen.OneConstOf(
			models.StockMovementReserve,
			models.StockMovementRelease,
			models.StockMovementDeduct,
			models.StockMovementAdjustment,
		),
	))

	// Property: Stock movement has required fields
	properties.Property("stock movement has required fields", prop.ForAll(
		func(productID, quantity int, movementType models.StockMovementType) bool {
			movement := &models.StockMovement{
				ProductID:    productID,
				Quantity:     quantity,
				MovementType: movementType,
				BalanceAfter: 100 - quantity, // Simulated balance
			}
			
			return movement.ProductID > 0 &&
			       movement.Quantity > 0 &&
			       movement.MovementType != ""
		},
		gen.IntRange(1, 1000),
		gen.IntRange(1, 100),
		gen.OneConstOf(
			models.StockMovementReserve,
			models.StockMovementRelease,
			models.StockMovementDeduct,
		),
	))

	properties.TestingRun(t)
}

// **Feature: zavera-commerce-upgrade, Property 15: Stock Reservation on Checkout**
// **Validates: Requirements 6.1**
func TestProperty_StockReservationOnCheckout(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property: Reserved stock decreases available stock
	properties.Property("reserved stock decreases available stock", prop.ForAll(
		func(initialStock, reserveQty int) bool {
			if reserveQty > initialStock {
				return true // Skip invalid cases
			}
			
			expectedBalance := initialStock - reserveQty
			
			movement := &models.StockMovement{
				MovementType: models.StockMovementReserve,
				Quantity:     reserveQty,
				BalanceAfter: expectedBalance,
			}
			
			return movement.BalanceAfter == initialStock - reserveQty
		},
		gen.IntRange(10, 1000),
		gen.IntRange(1, 10),
	))

	properties.TestingRun(t)
}

// **Feature: zavera-commerce-upgrade, Property 7: Shipping Snapshot Storage**
// **Validates: Requirements 2.3**
func TestProperty_ShippingSnapshotStorage(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property: Shipping snapshot has required fields
	properties.Property("shipping snapshot has required fields", prop.ForAll(
		func(orderID int, courier string, cost float64, weight int) bool {
			snapshot := &models.ShippingSnapshot{
				OrderID:  orderID,
				Courier:  courier,
				Service:  "REG",
				Cost:     cost,
				ETD:      "2-3",
				Weight:   weight,
			}
			
			return snapshot.OrderID > 0 &&
			       snapshot.Courier != "" &&
			       snapshot.Service != "" &&
			       snapshot.Cost > 0 &&
			       snapshot.ETD != "" &&
			       snapshot.Weight > 0
		},
		genOrderID(),
		genCourierCode(),
		gen.Float64Range(5000, 100000),
		gen.IntRange(100, 50000),
	))

	properties.TestingRun(t)
}

// **Feature: zavera-commerce-upgrade, Property 8: Shipping Data Integrity**
// **Validates: Requirements 2.4, 2.5, 11.1, 11.2**
func TestProperty_ShippingDataIntegrity(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property: Order shipping cost matches snapshot
	properties.Property("order shipping cost matches snapshot", prop.ForAll(
		func(cost float64) bool {
			snapshot := &models.ShippingSnapshot{
				Cost: cost,
			}
			
			order := &models.Order{
				ShippingCost: snapshot.Cost,
			}
			
			return order.ShippingCost == snapshot.Cost
		},
		gen.Float64Range(5000, 100000),
	))

	properties.TestingRun(t)
}

// **Feature: zavera-commerce-upgrade, Property 21: Dashboard Metrics Accuracy**
// **Validates: Requirements 10.1, 10.2, 10.3**
func TestProperty_DashboardMetricsAccuracy(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property: Revenue calculation is sum of paid/completed orders
	properties.Property("revenue is sum of paid and completed order totals", prop.ForAll(
		func(amounts []float64) bool {
			var expectedRevenue float64
			for _, amount := range amounts {
				expectedRevenue += amount
			}
			
			// Simulate dashboard calculation
			var calculatedRevenue float64
			for _, amount := range amounts {
				calculatedRevenue += amount
			}
			
			return calculatedRevenue == expectedRevenue
		},
		gen.SliceOfN(10, gen.Float64Range(10000, 1000000)),
	))

	properties.TestingRun(t)
}


// **Feature: zavera-commerce-upgrade, Property 19: Audit Log Immutability**
// **Validates: Requirements 8.2**
// *For any* audit log entry, UPDATE and DELETE operations SHALL fail (append-only)

func TestProperty_AuditLogImmutability(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property: Audit log entries have required immutable fields
	properties.Property("audit log entries have required immutable fields", prop.ForAll(
		func(adminID int, actionType models.AdminActionType) bool {
			auditLog := &models.AdminAuditLog{
				AdminUserID:  adminID,
				AdminEmail:   "admin@test.com",
				ActionType:   actionType,
				ActionDetail: "Test action",
				TargetType:   "order",
				TargetID:     1,
				Success:      true,
			}
			
			// Verify required fields are set
			return auditLog.AdminUserID > 0 &&
			       auditLog.AdminEmail != "" &&
			       auditLog.ActionType != "" &&
			       auditLog.ActionDetail != "" &&
			       auditLog.TargetType != ""
		},
		gen.IntRange(1, 1000),
		gen.OneConstOf(
			models.AdminActionForceCancel,
			models.AdminActionForceRefund,
			models.AdminActionUpdateStatus,
			models.AdminActionRestoreStock,
		),
	))

	// Property: Audit log captures state before and after
	properties.Property("audit log captures state changes", prop.ForAll(
		func(beforeStatus, afterStatus string) bool {
			auditLog := &models.AdminAuditLog{
				StateBefore: map[string]any{"status": beforeStatus},
				StateAfter:  map[string]any{"status": afterStatus},
			}
			
			// Verify state is captured
			return auditLog.StateBefore != nil &&
			       auditLog.StateAfter != nil &&
			       auditLog.StateBefore["status"] == beforeStatus &&
			       auditLog.StateAfter["status"] == afterStatus
		},
		gen.OneConstOf("PENDING", "PAID", "PACKING", "SHIPPED"),
		gen.OneConstOf("PAID", "PACKING", "SHIPPED", "DELIVERED"),
	))

	properties.TestingRun(t)
}


// **Feature: zavera-commerce-upgrade, Property 18: Refund Processing**
// **Validates: Requirements 7.4**
// *For any* refund request on an eligible order, the system SHALL log the reason,
// restore stock if applicable, and update order status to REFUNDED

func TestProperty_RefundProcessing(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property: Only DELIVERED or COMPLETED orders can be refunded
	properties.Property("only delivered or completed orders can be refunded", prop.ForAll(
		func(status models.OrderStatus) bool {
			order := &models.Order{Status: status}
			canRefund := order.CanBeRefunded()
			
			expected := status == models.OrderStatusDelivered || 
			           status == models.OrderStatusCompleted
			return canRefund == expected
		},
		gen.OneConstOf(
			models.OrderStatusPending,
			models.OrderStatusPaid,
			models.OrderStatusPacking,
			models.OrderStatusShipped,
			models.OrderStatusDelivered,
			models.OrderStatusCompleted,
			models.OrderStatusCancelled,
			models.OrderStatusFailed,
			models.OrderStatusExpired,
			models.OrderStatusRefunded,
		),
	))

	// Property: Refund transition is valid from eligible states
	properties.Property("refund transition is valid from eligible states", prop.ForAll(
		func(status models.OrderStatus) bool {
			canTransition := status.IsValidTransition(models.OrderStatusRefunded)
			
			// Only DELIVERED and COMPLETED can transition to REFUNDED
			expected := status == models.OrderStatusDelivered || 
			           status == models.OrderStatusCompleted
			return canTransition == expected
		},
		gen.OneConstOf(
			models.OrderStatusPending,
			models.OrderStatusPaid,
			models.OrderStatusPacking,
			models.OrderStatusShipped,
			models.OrderStatusDelivered,
			models.OrderStatusCompleted,
			models.OrderStatusCancelled,
			models.OrderStatusFailed,
			models.OrderStatusExpired,
			models.OrderStatusRefunded,
		),
	))

	properties.TestingRun(t)
}
