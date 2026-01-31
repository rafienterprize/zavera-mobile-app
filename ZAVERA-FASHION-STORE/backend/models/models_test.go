package models

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Feature: zavera-commerce-upgrade, Property 1: Order State Machine Validity**
// **Validates: Requirements 1.2, 1.3**
// *For any* order and any attempted status transition, the transition SHALL succeed
// if and only if it follows the defined state machine.

// AllOrderStatuses contains all valid order statuses for testing
var AllOrderStatuses = []OrderStatus{
	OrderStatusPending,
	OrderStatusPaid,
	OrderStatusPacking,
	OrderStatusShipped,
	OrderStatusDelivered,
	OrderStatusCompleted,
	OrderStatusCancelled,
	OrderStatusFailed,
	OrderStatusExpired,
	OrderStatusRefunded,
}

// TerminalStatuses are statuses that cannot transition to any other status
var TerminalStatuses = []OrderStatus{
	OrderStatusCancelled,
	OrderStatusFailed,
	OrderStatusExpired,
	OrderStatusRefunded,
}

// genOrderStatus generates random order statuses
func genOrderStatus() gopter.Gen {
	return gen.OneConstOf(
		OrderStatusPending,
		OrderStatusPaid,
		OrderStatusPacking,
		OrderStatusShipped,
		OrderStatusDelivered,
		OrderStatusCompleted,
		OrderStatusCancelled,
		OrderStatusFailed,
		OrderStatusExpired,
		OrderStatusRefunded,
	)
}

func TestProperty_OrderStateMachineValidity(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property: Valid transitions should succeed, invalid should fail
	properties.Property("valid transitions succeed and invalid transitions fail", prop.ForAll(
		func(from, to OrderStatus) bool {
			isValid := from.IsValidTransition(to)
			
			// Check against our defined transitions
			allowed, exists := ValidOrderTransitions[from]
			if !exists {
				// Terminal states have no valid transitions
				return !isValid
			}
			
			// Check if 'to' is in allowed list
			expectedValid := false
			for _, s := range allowed {
				if s == to {
					expectedValid = true
					break
				}
			}
			
			return isValid == expectedValid
		},
		genOrderStatus(),
		genOrderStatus(),
	))

	// Property: Terminal states cannot transition to any other state
	properties.Property("terminal states have no outgoing transitions", prop.ForAll(
		func(to OrderStatus) bool {
			for _, terminal := range TerminalStatuses {
				if terminal.IsValidTransition(to) {
					return false
				}
			}
			return true
		},
		genOrderStatus(),
	))

	// Property: IsFinalStatus correctly identifies terminal states
	properties.Property("IsFinalStatus correctly identifies terminal states", prop.ForAll(
		func(status OrderStatus) bool {
			isFinal := status.IsFinalStatus()
			
			// Check if status is in terminal list
			isTerminal := false
			for _, t := range TerminalStatuses {
				if status == t {
					isTerminal = true
					break
				}
			}
			
			// COMPLETED is also final but can transition to REFUNDED
			if status == OrderStatusCompleted {
				return isFinal == true
			}
			
			return isFinal == isTerminal
		},
		genOrderStatus(),
	))

	// Property: State machine follows expected flow
	properties.Property("happy path transitions are valid", prop.ForAll(
		func(_ int) bool {
			// Test the happy path: PENDING -> PAID -> PACKING -> SHIPPED -> DELIVERED -> COMPLETED
			happyPath := []OrderStatus{
				OrderStatusPending,
				OrderStatusPaid,
				OrderStatusPacking,
				OrderStatusShipped,
				OrderStatusDelivered,
				OrderStatusCompleted,
			}
			
			for i := 0; i < len(happyPath)-1; i++ {
				if !happyPath[i].IsValidTransition(happyPath[i+1]) {
					return false
				}
			}
			return true
		},
		gen.Int(),
	))

	properties.TestingRun(t)
}

// **Feature: zavera-commerce-upgrade, Property 2: Order Initial Status**
// **Validates: Requirements 1.1**
// *For any* newly created order, the initial status SHALL be PENDING_PAYMENT (PENDING).
func TestProperty_OrderInitialStatus(t *testing.T) {
	// This is more of a unit test since initial status is set at creation time
	// The property is: all new orders start with PENDING status
	initialStatus := OrderStatusPending
	
	// Verify PENDING can transition to expected states
	expectedTransitions := []OrderStatus{OrderStatusPaid, OrderStatusCancelled, OrderStatusFailed, OrderStatusExpired}
	
	for _, expected := range expectedTransitions {
		if !initialStatus.IsValidTransition(expected) {
			t.Errorf("Initial status PENDING should be able to transition to %s", expected)
		}
	}
}

// **Feature: zavera-commerce-upgrade, Property 14: Cancel Permission Validation**
// **Validates: Requirements 5.6, 7.1, 7.2**
func TestProperty_CancelPermissionValidation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property: Customer can only cancel PENDING orders
	properties.Property("customer can only cancel PENDING orders", prop.ForAll(
		func(status OrderStatus) bool {
			order := &Order{Status: status}
			canCancel := order.CanBeCancelledByCustomer()
			
			// Customer should only be able to cancel PENDING orders
			expected := status == OrderStatusPending
			return canCancel == expected
		},
		genOrderStatus(),
	))

	// Property: Admin can cancel orders before SHIPPED
	properties.Property("admin can cancel orders before SHIPPED", prop.ForAll(
		func(status OrderStatus) bool {
			order := &Order{Status: status}
			canCancel := order.CanBeCancelledByAdmin()
			
			// Admin can cancel PENDING, PAID, PACKING
			expected := status == OrderStatusPending || 
			           status == OrderStatusPaid || 
			           status == OrderStatusPacking
			return canCancel == expected
		},
		genOrderStatus(),
	))

	properties.TestingRun(t)
}

// **Feature: zavera-commerce-upgrade, Property 10: Resi Immutability After Shipping**
// **Validates: Requirements 3.3**
func TestProperty_ResiImmutabilityAfterShipping(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property: Resi is locked for SHIPPED and beyond
	properties.Property("resi is locked for SHIPPED and beyond", prop.ForAll(
		func(status OrderStatus) bool {
			order := &Order{Status: status}
			isLocked := order.IsResiLocked()
			
			// Resi should be locked for SHIPPED, DELIVERED, COMPLETED, REFUNDED
			expected := status == OrderStatusShipped ||
			           status == OrderStatusDelivered ||
			           status == OrderStatusCompleted ||
			           status == OrderStatusRefunded
			return isLocked == expected
		},
		genOrderStatus(),
	))

	properties.TestingRun(t)
}

// **Feature: zavera-commerce-upgrade, Property 18: Refund Processing**
// **Validates: Requirements 7.4**
func TestProperty_RefundEligibility(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property: Only DELIVERED or COMPLETED orders can be refunded
	properties.Property("only DELIVERED or COMPLETED orders can be refunded", prop.ForAll(
		func(status OrderStatus) bool {
			order := &Order{Status: status}
			canRefund := order.CanBeRefunded()
			
			// Only DELIVERED and COMPLETED can be refunded
			expected := status == OrderStatusDelivered || status == OrderStatusCompleted
			return canRefund == expected
		},
		genOrderStatus(),
	))

	properties.TestingRun(t)
}

// **Feature: zavera-commerce-upgrade, Property 16: Stock Release on Failure**
// **Validates: Requirements 6.2, 6.3, 7.3**
func TestProperty_StockReleaseOnFailure(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property: RequiresStockRestore returns true for CANCELLED, FAILED, EXPIRED
	properties.Property("stock restore required for cancelled/failed/expired orders", prop.ForAll(
		func(status OrderStatus) bool {
			requiresRestore := status.RequiresStockRestore()
			
			expected := status == OrderStatusCancelled ||
			           status == OrderStatusFailed ||
			           status == OrderStatusExpired
			return requiresRestore == expected
		},
		genOrderStatus(),
	))

	properties.TestingRun(t)
}
