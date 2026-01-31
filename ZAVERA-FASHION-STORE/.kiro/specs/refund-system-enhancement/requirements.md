# Requirements Document: Refund System Enhancement

## Introduction

The Refund System Enhancement addresses critical issues in the current refund implementation, including database constraint violations, payment gateway integration gaps, and poor customer experience. This enhancement ensures refunds are processed reliably, money actually returns to customers, and both customers and admins have clear visibility into refund status and details.

## Glossary

- **Refund_System**: The complete system responsible for processing, tracking, and managing refunds
- **Payment_Gateway**: Midtrans payment service that processes actual money transfers
- **Admin_Panel**: Web interface used by administrators to manage orders and refunds
- **Customer_Portal**: Web interface used by customers to view order and refund status
- **Database**: PostgreSQL database storing refund, order, and payment records
- **Foreign_Key_Constraint**: Database rule ensuring referential integrity between tables
- **Refund_Status**: Current state of a refund (PENDING, PROCESSING, COMPLETED, FAILED)
- **Order_Status**: Current state of an order (PENDING, PAID, DELIVERED, REFUNDED, etc.)
- **Stock_Restoration**: Process of returning product inventory when items are refunded
- **Idempotency_Key**: Unique identifier preventing duplicate refund processing
- **Gateway_Refund_ID**: Unique identifier from Midtrans for a processed refund
- **Audit_Trail**: Historical record of all refund actions and status changes
- **VA_Payment**: Virtual Account payment method (BCA, BNI, Mandiri, etc.)
- **E_Wallet**: Electronic wallet payment method (GoPay, ShopeePay, DANA, OVO)
- **QRIS**: Quick Response Code Indonesian Standard payment method
- **Manual_Refund**: Refund for orders without payment records (manually marked as paid)

## Requirements

### Requirement 1: Database Integrity and Constraint Handling

**User Story:** As a system administrator, I want the refund system to handle database constraints correctly, so that refund operations never fail due to foreign key violations.

#### Acceptance Criteria

1. WHEN creating a refund record, THE Refund_System SHALL validate that the `requested_by` user exists before insertion
2. WHEN creating a refund for a manual order (no payment record), THE Refund_System SHALL allow NULL values for `payment_id` field
3. IF a foreign key constraint would be violated, THEN THE Refund_System SHALL return a descriptive error message identifying the missing reference
4. WHEN a refund is created, THE Refund_System SHALL use a database transaction to ensure atomicity
5. IF any part of refund creation fails, THEN THE Refund_System SHALL rollback all changes and restore the original state
6. WHEN checking for existing refunds, THE Refund_System SHALL use row-level locking to prevent race conditions
7. THE Refund_System SHALL validate that the order exists before creating a refund record

### Requirement 2: Payment Gateway Integration

**User Story:** As a customer, I want my money to actually return to my payment method when a refund is processed, so that I receive my funds back.

#### Acceptance Criteria

1. WHEN a refund is processed, THE Refund_System SHALL call the Midtrans refund API with the correct order code and refund amount
2. WHEN calling Midtrans API, THE Refund_System SHALL include proper authentication headers using the server key
3. WHEN Midtrans returns a successful response (status code 200 or 201), THE Refund_System SHALL store the gateway refund ID
4. IF Midtrans returns an error response, THEN THE Refund_System SHALL mark the refund as FAILED and store the error message
5. WHEN a refund fails at the gateway, THE Refund_System SHALL allow retry attempts by administrators
6. FOR ALL payment methods (VA, E_Wallet, QRIS, Credit Card), THE Refund_System SHALL use the same Midtrans refund endpoint
7. WHEN processing a Manual_Refund (no payment record), THE Refund_System SHALL skip gateway processing and mark as COMPLETED immediately
8. THE Refund_System SHALL store the complete gateway response for audit purposes

### Requirement 3: Refund Status Tracking and Visibility

**User Story:** As a customer, I want to see clear refund status and details in my purchase history, so that I know when and how much money I will receive back.

#### Acceptance Criteria

1. WHEN viewing purchase history, THE Customer_Portal SHALL display refund status for refunded orders
2. WHEN displaying a refunded order, THE Customer_Portal SHALL show the refund amount
3. WHEN displaying a refunded order, THE Customer_Portal SHALL show the refund processing date
4. WHEN displaying a refunded order, THE Customer_Portal SHALL show the estimated refund completion timeline
5. WHEN a refund is in PROCESSING status, THE Customer_Portal SHALL display "Refund in progress - funds will arrive in 3-7 business days"
6. WHEN a refund is COMPLETED, THE Customer_Portal SHALL display "Refund completed - funds returned to your payment method"
7. WHEN a refund is FAILED, THE Customer_Portal SHALL display "Refund failed - please contact support"
8. WHEN viewing order details, THE Customer_Portal SHALL show a dedicated refund information section

### Requirement 4: Admin Refund Processing Workflow

**User Story:** As an administrator, I want a safe and validated refund processing workflow, so that I can process refunds without errors and with proper verification.

#### Acceptance Criteria

1. WHEN viewing an order eligible for refund, THE Admin_Panel SHALL display a "Refund" action button
2. WHEN clicking the refund button, THE Admin_Panel SHALL display a confirmation modal with order details
3. WHEN processing a refund, THE Admin_Panel SHALL require the administrator to enter a reason
4. WHEN submitting a refund request, THE Admin_Panel SHALL validate that the refund amount does not exceed the refundable amount
5. WHEN a refund is processing, THE Admin_Panel SHALL disable the refund button to prevent duplicate submissions
6. WHEN a refund completes successfully, THE Admin_Panel SHALL display a success message with the gateway refund ID
7. IF a refund fails, THEN THE Admin_Panel SHALL display the error message and offer a retry option
8. WHEN viewing refund history, THE Admin_Panel SHALL show all refund attempts with timestamps and status

### Requirement 5: Order Status Management

**User Story:** As a system, I want to maintain correct order status throughout the refund lifecycle, so that order state accurately reflects refund operations.

#### Acceptance Criteria

1. WHEN a full refund is completed, THE Refund_System SHALL update the order status to REFUNDED
2. WHEN a partial refund is completed, THE Refund_System SHALL keep the order status as DELIVERED or COMPLETED
3. WHEN updating order status, THE Refund_System SHALL record the status change in the audit trail
4. THE Refund_System SHALL update the order's `refund_status` field to indicate FULL or PARTIAL refund
5. THE Refund_System SHALL update the order's `refund_amount` field with the total refunded amount
6. WHEN multiple refunds exist for an order, THE Refund_System SHALL calculate the total refunded amount correctly
7. THE Refund_System SHALL update the order's `refunded_at` timestamp when the first refund completes

### Requirement 6: Stock Restoration

**User Story:** As a system, I want to restore product stock when items are refunded, so that inventory remains accurate.

#### Acceptance Criteria

1. WHEN a full refund is completed, THE Refund_System SHALL restore stock for all order items
2. WHEN a partial item refund is completed, THE Refund_System SHALL restore stock only for the refunded items
3. WHEN restoring stock, THE Refund_System SHALL increment the product stock by the refunded quantity
4. WHEN stock restoration completes, THE Refund_System SHALL mark the refund items as `stock_restored = true`
5. IF stock restoration fails, THEN THE Refund_System SHALL log the error but not fail the entire refund
6. WHEN a shipping-only refund is processed, THE Refund_System SHALL not restore any stock
7. THE Refund_System SHALL prevent duplicate stock restoration for the same refund items

### Requirement 7: Error Handling and Recovery

**User Story:** As a system, I want to handle errors gracefully and provide recovery mechanisms, so that temporary failures do not result in permanent data inconsistency.

#### Acceptance Criteria

1. WHEN a database error occurs during refund creation, THE Refund_System SHALL rollback the transaction
2. WHEN a gateway API call fails due to network timeout, THE Refund_System SHALL mark the refund as FAILED with retry capability
3. WHEN a gateway API call fails due to invalid request, THE Refund_System SHALL mark the refund as FAILED and log the validation error
4. IF the database update succeeds but gateway call fails, THEN THE Refund_System SHALL mark the refund as FAILED and allow retry
5. WHEN retrying a failed refund, THE Refund_System SHALL use the same idempotency key to prevent duplicate gateway charges
6. WHEN an error occurs, THE Refund_System SHALL provide clear error messages to administrators
7. THE Refund_System SHALL log all errors with sufficient context for debugging

### Requirement 8: Refund Types and Calculations

**User Story:** As an administrator, I want to process different types of refunds with correct amount calculations, so that customers receive the appropriate refund amount.

#### Acceptance Criteria

1. WHEN processing a FULL refund, THE Refund_System SHALL refund the entire order total amount
2. WHEN processing a SHIPPING_ONLY refund, THE Refund_System SHALL refund only the shipping cost
3. WHEN processing a PARTIAL refund, THE Refund_System SHALL refund the specified amount up to the refundable balance
4. WHEN processing an ITEM_ONLY refund, THE Refund_System SHALL calculate the refund amount as sum of (quantity × price_per_unit) for selected items
5. WHEN calculating refundable amount, THE Refund_System SHALL subtract any previously completed refunds
6. THE Refund_System SHALL prevent refund amounts that exceed the original payment amount
7. WHEN displaying refund breakdown, THE Refund_System SHALL show items_refund and shipping_refund separately

### Requirement 9: Idempotency and Duplicate Prevention

**User Story:** As a system, I want to prevent duplicate refund processing, so that customers are not refunded multiple times for the same order.

#### Acceptance Criteria

1. WHEN creating a refund with an idempotency key, THE Refund_System SHALL check if the key already exists
2. IF an idempotency key already exists, THEN THE Refund_System SHALL return the existing refund record without creating a new one
3. WHEN processing concurrent refund requests for the same order, THE Refund_System SHALL use row-level locking to prevent race conditions
4. THE Refund_System SHALL validate that the total refunded amount does not exceed the order total
5. WHEN multiple refunds are requested simultaneously, THE Refund_System SHALL process them sequentially using database locks
6. THE Refund_System SHALL store the idempotency key with each refund record for audit purposes

### Requirement 10: Audit Trail and Compliance

**User Story:** As a system administrator, I want complete audit trails for all refund operations, so that I can track who did what and when for compliance purposes.

#### Acceptance Criteria

1. WHEN a refund is created, THE Refund_System SHALL record the requesting user ID and timestamp
2. WHEN a refund status changes, THE Refund_System SHALL record the status change in the refund_status_history table
3. WHEN recording status changes, THE Refund_System SHALL include the actor (user or system), old status, new status, and reason
4. WHEN an administrator processes a refund, THE Refund_System SHALL record the processed_by user ID and processed_at timestamp
5. WHEN a refund completes, THE Refund_System SHALL record the completed_at timestamp
6. THE Refund_System SHALL store the complete gateway response for each refund attempt
7. WHEN viewing audit logs, THE Admin_Panel SHALL display all refund actions in chronological order

### Requirement 11: Notification System

**User Story:** As a customer, I want to receive notifications when my refund status changes, so that I stay informed about my refund progress.

#### Acceptance Criteria

1. WHEN a refund is created, THE Refund_System SHALL send an email notification to the customer
2. WHEN a refund status changes to PROCESSING, THE Refund_System SHALL send a notification with estimated completion time
3. WHEN a refund status changes to COMPLETED, THE Refund_System SHALL send a notification confirming funds have been returned
4. WHEN a refund status changes to FAILED, THE Refund_System SHALL send a notification asking the customer to contact support
5. WHEN sending notifications, THE Refund_System SHALL include the order code, refund amount, and refund status
6. THE Refund_System SHALL include a link to view refund details in all notification emails

### Requirement 12: Payment Method Specific Handling

**User Story:** As a system, I want to handle different payment methods correctly during refunds, so that refunds work for all payment types.

#### Acceptance Criteria

1. WHEN refunding a VA_Payment, THE Refund_System SHALL return funds to the customer's bank account
2. WHEN refunding an E_Wallet payment, THE Refund_System SHALL return funds to the customer's wallet balance
3. WHEN refunding a QRIS payment, THE Refund_System SHALL return funds according to the QRIS provider's method
4. WHEN refunding a credit card payment, THE Refund_System SHALL return funds to the customer's credit card
5. WHEN displaying refund timeline, THE Refund_System SHALL show payment-method-specific estimated completion times
6. FOR VA_Payment refunds, THE Refund_System SHALL display "3-7 business days" as the estimated timeline
7. FOR E_Wallet refunds, THE Refund_System SHALL display "1-3 business days" as the estimated timeline

### Requirement 13: Manual Refund Handling

**User Story:** As an administrator, I want to process refunds for manually marked orders (without payment records), so that I can handle edge cases where payment was received outside the system.

#### Acceptance Criteria

1. WHEN creating a refund for an order without a payment record, THE Refund_System SHALL allow the refund creation
2. WHEN processing a Manual_Refund, THE Refund_System SHALL skip Midtrans gateway processing
3. WHEN creating a Manual_Refund, THE Refund_System SHALL set the status to COMPLETED immediately
4. WHEN creating a Manual_Refund, THE Refund_System SHALL set gateway_refund_id to "MANUAL_REFUND"
5. WHEN creating a Manual_Refund, THE Refund_System SHALL use the order total amounts for refund calculation
6. WHEN displaying a Manual_Refund, THE Admin_Panel SHALL clearly indicate it was processed manually
7. THE Refund_System SHALL still update order status and restore stock for Manual_Refunds

### Requirement 14: Refund API Endpoints

**User Story:** As a frontend developer, I want well-defined API endpoints for refund operations, so that I can integrate refund functionality into the UI.

#### Acceptance Criteria

1. THE Refund_System SHALL provide a POST /admin/refunds endpoint for creating refunds
2. THE Refund_System SHALL provide a POST /admin/refunds/:id/process endpoint for processing pending refunds
3. THE Refund_System SHALL provide a GET /admin/refunds/:id endpoint for retrieving refund details
4. THE Refund_System SHALL provide a GET /admin/orders/:code/refunds endpoint for listing order refunds
5. THE Refund_System SHALL provide a GET /customer/orders/:code/refunds endpoint for customers to view their refunds
6. WHEN calling refund endpoints, THE Refund_System SHALL validate authentication and authorization
7. WHEN returning refund data, THE Refund_System SHALL include all refund items and status history

### Requirement 15: Refund Validation Rules

**User Story:** As a system, I want to validate refund requests before processing, so that invalid refunds are rejected early.

#### Acceptance Criteria

1. WHEN validating a refund request, THE Refund_System SHALL verify the order status is DELIVERED or COMPLETED
2. WHEN validating a refund request, THE Refund_System SHALL verify the payment status is SUCCESS (if payment exists)
3. WHEN validating a refund request, THE Refund_System SHALL verify the refund amount is positive and non-zero
4. WHEN validating a refund request, THE Refund_System SHALL verify the refund amount does not exceed the refundable balance
5. WHEN validating an item refund, THE Refund_System SHALL verify all specified order items exist
6. WHEN validating an item refund, THE Refund_System SHALL verify refund quantities do not exceed ordered quantities
7. IF any validation fails, THEN THE Refund_System SHALL return a descriptive error message without creating a refund record
