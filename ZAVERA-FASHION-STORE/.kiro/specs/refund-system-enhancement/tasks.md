# Implementation Plan: Refund System Enhancement

## Overview

This implementation plan breaks down the refund system enhancement into discrete, manageable tasks. Each task builds on previous work and includes validation through code execution. The plan follows professional backend engineering standards to ensure the refund system works reliably, similar to Tokopedia/Shopee return systems.

## Tasks

- [x] 1. Database Schema Updates and Migrations
  - Create migration files to update database schema
  - Make `requested_by` and `payment_id` nullable in refunds table
  - Add refund tracking fields to orders table (`refund_status`, `refund_amount`, `refunded_at`)
  - Create `refund_status_history` table for audit trail
  - Add necessary indexes for performance
  - Test migrations on local database
  - _Requirements: 1.1, 1.2, 5.4, 5.5, 5.7, 10.2_

- [x] 2. Update Refund Models and DTOs
  - [x] 2.1 Update Go models with nullable fields
    - Modify `Refund` struct to use `*int` for `requested_by`, `payment_id`, `processed_by`
    - Add refund fields to `Order` struct (`RefundStatus`, `RefundAmount`, `RefundedAt`)
    - Create `RefundStatusHistory` model
    - _Requirements: 1.1, 1.2, 5.4, 5.5, 10.2_
  
  - [x] 2.2 Create comprehensive DTOs for refund operations
    - Create `RefundRequest` DTO with validation tags
    - Create `RefundItemRequest` DTO for item-specific refunds
    - Create `RefundResponse` DTO with all fields
    - Create `MidtransRefundRequest` and `MidtransRefundResponse` DTOs
    - _Requirements: 8.1, 8.2, 8.3, 8.4, 14.1_

- [x] 3. Implement Core Refund Repository Methods
  - [x] 3.1 Implement CRUD operations with transaction support
    - Implement `Create` and `CreateWithTx` methods
    - Implement `FindByID`, `FindByCode`, `FindByOrderID` methods
    - Implement `FindByIdempotencyKey` for idempotency checking
    - Add proper error handling and logging
    - _Requirements: 1.4, 9.1, 9.6_
  
  - [x] 3.2 Implement status management methods
    - Implement `UpdateStatus` method
    - Implement `MarkCompleted` method with gateway response storage
    - Implement `MarkFailed` method with error message storage
    - _Requirements: 2.3, 2.4, 2.8_
  
  - [x] 3.3 Implement refund items and status history methods
    - Implement `CreateRefundItem` and `CreateRefundItemWithTx` methods
    - Implement `FindItemsByRefundID` method
    - Implement `MarkItemStockRestored` method
    - Implement `RecordStatusChange` for audit trail
    - Implement `GetStatusHistory` method
    - _Requirements: 6.4, 10.2, 10.3_

- [x] 4. Implement Refund Service Core Logic
  - [x] 4.1 Implement refund validation logic
    - Create `validateRefundRequest` method checking order status, payment status
    - Validate refund amount is positive and within refundable balance
    - Validate order items exist and quantities are valid
    - Return descriptive error messages for all validation failures
    - _Requirements: 15.1, 15.2, 15.3, 15.4, 15.5, 15.6, 15.7_
  
  - [x] 4.2 Implement refund amount calculation logic
    - Create `calculateRefundAmount` method for all refund types
    - Implement FULL refund calculation (total amount)
    - Implement SHIPPING_ONLY refund calculation (shipping cost only)
    - Implement PARTIAL refund calculation (specified amount)
    - Implement ITEM_ONLY refund calculation (sum of item quantities × prices)
    - Calculate refundable balance (total - sum of completed refunds)
    - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5, 8.6, 8.7_
  
  - [x] 4.3 Implement CreateRefund method with transaction safety
    - Check idempotency key and return existing refund if found
    - Validate order exists and is refundable
    - Get payment record (handle NULL for manual orders)
    - Start database transaction with row-level locking (`SELECT ... FOR UPDATE`)
    - Check existing refunds within locked transaction
    - Calculate refund amounts
    - Create refund record and refund items
    - Record initial status change
    - Commit transaction or rollback on error
    - _Requirements: 1.4, 1.5, 1.6, 1.7, 9.1, 9.3_
  
  - [x] 4.4 Implement manual refund handling
    - Create `createManualRefund` method for orders without payment
    - Skip gateway processing for manual refunds
    - Set status to COMPLETED immediately
    - Set gateway_refund_id to "MANUAL_REFUND"
    - Use order amounts for refund calculation
    - _Requirements: 2.7, 13.2, 13.3, 13.4, 13.5_

- [x] 5. Implement Midtrans Gateway Integration
  - [x] 5.1 Implement ProcessMidtransRefund method
    - Build Midtrans API request with order code, refund amount, reason
    - Set proper authentication headers (Basic Auth with server key)
    - Make HTTP POST request to `/v2/{order_code}/refund` endpoint
    - Set 30-second timeout for API call
    - Parse response and handle success (200/201) and error responses
    - Store complete gateway response for audit
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.6, 2.8_
  
  - [x] 5.2 Implement ProcessRefund method
    - Validate refund can be processed (status is PENDING)
    - Update status to PROCESSING
    - Call ProcessMidtransRefund for gateway processing
    - Handle success: mark COMPLETED, store gateway refund ID
    - Handle failure: mark FAILED, store error message, allow retry
    - Record all status changes in audit trail
    - _Requirements: 2.3, 2.4, 7.2, 7.3, 7.4, 7.5_
  
  - [x] 5.3 Implement CheckMidtransRefundStatus method
    - Make GET request to Midtrans status endpoint
    - Parse and return refund status from gateway
    - Use for manual status verification if needed
    - _Requirements: 2.1, 2.2_

- [x] 6. Implement Order Status and Stock Management
  - [x] 6.1 Implement updateOrderRefundStatus method
    - Calculate total refunded amount from all completed refunds
    - Update order's `refund_amount` field
    - Set `refund_status` to FULL if total equals order total, else PARTIAL
    - Set `refunded_at` timestamp on first refund
    - Update order status to REFUNDED if full refund
    - Record order status change in audit trail
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 5.7_
  
  - [x] 6.2 Implement restoreRefundedStock method
    - For FULL refunds: restore stock for all order items
    - For ITEM_ONLY refunds: restore stock only for refunded items
    - For SHIPPING_ONLY refunds: skip stock restoration
    - Increment product stock by refunded quantity
    - Mark refund items as `stock_restored = true`
    - Handle stock restoration failures gracefully (log error, don't fail refund)
    - Ensure idempotency (don't restore stock twice for same items)
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5, 6.6, 6.7_

- [x] 7. Implement Admin Refund API Endpoints
  - [x] 7.1 Create admin refund handler
    - Implement POST `/admin/refunds` endpoint for creating refunds
    - Implement POST `/admin/refunds/:id/process` endpoint for processing refunds
    - Implement POST `/admin/refunds/:id/retry` endpoint for retrying failed refunds
    - Implement GET `/admin/refunds/:id` endpoint for refund details
    - Implement GET `/admin/orders/:code/refunds` endpoint for listing order refunds
    - Implement GET `/admin/refunds` endpoint for listing all refunds with pagination
    - Add authentication middleware (require valid JWT token)
    - Add authorization middleware (require admin role)
    - _Requirements: 14.1, 14.2, 14.3, 14.4, 14.6_
  
  - [x] 7.2 Implement request validation and error handling
    - Validate request body against DTO schemas
    - Return 400 Bad Request for validation errors
    - Return 401 Unauthorized for missing/invalid auth
    - Return 403 Forbidden for insufficient permissions
    - Return 404 Not Found for non-existent resources
    - Return 409 Conflict for idempotency key conflicts
    - Return 500 Internal Server Error for unexpected errors
    - Use consistent error response format with error code, message, details
    - _Requirements: 1.3, 7.6, 15.7_
  
  - [x] 7.3 Implement response formatting
    - Include all refund fields in response
    - Include refund items array
    - Include status history array
    - Format timestamps in ISO 8601 format
    - Format amounts with 2 decimal places
    - _Requirements: 14.7_

- [x] 8. Implement Customer Refund API Endpoints
  - [x] 8.1 Create customer refund handler
    - Implement GET `/customer/orders/:code/refunds` endpoint
    - Implement GET `/customer/refunds/:code` endpoint
    - Add authentication middleware (require valid JWT token)
    - Verify customer owns the order before returning refund data
    - _Requirements: 14.5, 14.6_
  
  - [x] 8.2 Implement customer-friendly response formatting
    - Include refund status with human-readable labels
    - Include estimated refund timeline based on payment method
    - Include refund amount and breakdown
    - Include processing and completion dates
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 12.5_

- [x] 9. Update Admin Panel UI for Refund Management
  - [x] 9.1 Add refund button to order detail page
    - Show "Refund" button for orders with status DELIVERED or COMPLETED
    - Disable button if order already fully refunded
    - Add click handler to open refund modal
    - _Requirements: 4.1_
  
  - [x] 9.2 Create refund modal component
    - Add refund type selector (FULL, PARTIAL, SHIPPING_ONLY, ITEM_ONLY)
    - Add reason dropdown with predefined options
    - Add reason detail textarea
    - Add amount input for PARTIAL refunds
    - Add item selector with quantity inputs for ITEM_ONLY refunds
    - Add validation before submission
    - Show loading state during processing
    - Show success message with gateway refund ID
    - Show error message with retry option on failure
    - _Requirements: 4.2, 4.3, 4.4, 4.6, 4.7_
  
  - [x] 9.3 Add refund history section to order detail page
    - Display all refunds for the order
    - Show refund code, type, amount, status, timestamps
    - Show gateway refund ID or "MANUAL_REFUND" indicator
    - Show refund items for partial refunds
    - Show status history timeline
    - Add "Retry" button for failed refunds
    - _Requirements: 4.8, 13.6_

- [x] 10. Update Customer Portal UI for Refund Visibility
  - [x] 10.1 Update purchase history page (pembelian)
    - Add refund status badge to refunded orders
    - Add refund information card showing amount, status, timeline
    - Show estimated refund completion time based on payment method
    - Add link to view full refund details
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.8_
  
  - [x] 10.2 Create refund details section for order detail page
    - Create new component to display refund information
    - Show refund status with color-coded badges
    - Show refund amount breakdown (items + shipping)
    - Show refund timeline with status-specific messages
    - Show processing and completion dates
    - Show refund items for partial refunds
    - Add payment-method-specific timeline messages
    - _Requirements: 3.5, 3.6, 3.7, 3.8, 12.5, 12.6, 12.7_

- [x] 11. Implement Notification System
  - [x] 11.1 Create email notification templates
    - Create "Refund Created" email template
    - Create "Refund Processing" email template with timeline
    - Create "Refund Completed" email template
    - Create "Refund Failed" email template with support contact
    - Include order code, refund amount, status in all templates
    - Include link to view refund details in all templates
    - _Requirements: 11.1, 11.2, 11.3, 11.4, 11.5, 11.6_
  
  - [x] 11.2 Implement notification sending logic
    - Send email when refund is created
    - Send email when refund status changes to PROCESSING
    - Send email when refund status changes to COMPLETED
    - Send email when refund status changes to FAILED
    - Handle email sending failures gracefully (log error, don't fail refund)
    - _Requirements: 11.1, 11.2, 11.3, 11.4_

- [x] 12. Implement Comprehensive Error Handling
  - [x] 12.1 Add error handling for all failure scenarios
    - Handle database errors with transaction rollback
    - Handle gateway timeout errors with retry capability
    - Handle gateway validation errors with descriptive messages
    - Handle foreign key constraint violations with descriptive errors
    - Handle concurrent refund attempts with proper locking
    - Log all errors with sufficient context (refund code, order code, stack trace)
    - _Requirements: 1.3, 7.1, 7.2, 7.3, 7.4, 7.7_
  
  - [x] 12.2 Implement retry mechanism for failed refunds
    - Allow admin to retry failed refunds via API endpoint
    - Use same idempotency key for all retry attempts
    - Update status from FAILED to PROCESSING on retry
    - Call gateway API again with same parameters
    - Handle success and failure appropriately
    - _Requirements: 2.5, 7.5, 9.5_

- [x] 13. Write Unit Tests for Core Functionality
  - [x] 13.1 Write unit tests for refund validation
    - Test order status validation (reject non-DELIVERED/COMPLETED orders)
    - Test payment status validation (reject non-SUCCESS payments)
    - Test refund amount validation (reject zero/negative amounts)
    - Test refund amount exceeds balance validation
    - Test item existence validation
    - Test item quantity validation
    - Test validation error messages are descriptive
    - _Requirements: 15.1, 15.2, 15.3, 15.4, 15.5, 15.6, 15.7_
  
  - [x] 13.2 Write unit tests for refund amount calculations
    - Test FULL refund calculation equals order total
    - Test SHIPPING_ONLY refund calculation equals shipping cost
    - Test PARTIAL refund calculation with specified amount
    - Test ITEM_ONLY refund calculation as sum of item prices
    - Test refundable balance calculation with existing refunds
    - Test refund breakdown (items_refund + shipping_refund)
    - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5, 8.7_
  
  - [x] 13.3 Write unit tests for order status updates
    - Test full refund updates order status to REFUNDED
    - Test partial refund keeps order status unchanged
    - Test refund_status field set to FULL or PARTIAL
    - Test refund_amount field updated correctly
    - Test refunded_at timestamp set on first refund
    - Test multiple refunds aggregate correctly
    - _Requirements: 5.1, 5.2, 5.4, 5.5, 5.6, 5.7_
  
  - [x] 13.4 Write unit tests for stock restoration
    - Test full refund restores all item stock
    - Test item refund restores only selected item stock
    - Test shipping refund does not restore stock
    - Test stock_restored flag set after restoration
    - Test stock restoration idempotency (no double restoration)
    - Test stock restoration failure doesn't fail refund
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5, 6.6, 6.7_
  
  - [x] 13.5 Write unit tests for manual refund handling
    - Test manual refund skips gateway processing
    - Test manual refund sets status to COMPLETED immediately
    - Test manual refund sets gateway_refund_id to "MANUAL_REFUND"
    - Test manual refund uses order amounts for calculation
    - Test manual refund still updates order status and restores stock
    - _Requirements: 2.7, 13.2, 13.3, 13.4, 13.5, 13.7_

- [x] 14. Write Integration Tests for API Endpoints
  - [x] 14.1 Write integration tests for admin refund endpoints
    - Test POST /admin/refunds creates refund successfully
    - Test POST /admin/refunds validates authentication
    - Test POST /admin/refunds validates authorization (admin role)
    - Test POST /admin/refunds/:id/process processes refund
    - Test POST /admin/refunds/:id/retry retries failed refund
    - Test GET /admin/refunds/:id returns refund details with items and history
    - Test GET /admin/orders/:code/refunds lists all order refunds
    - Test GET /admin/refunds lists all refunds with pagination
    - _Requirements: 14.1, 14.2, 14.3, 14.4, 14.6, 14.7_
  
  - [x] 14.2 Write integration tests for customer refund endpoints
    - Test GET /customer/orders/:code/refunds returns customer's refunds
    - Test GET /customer/orders/:code/refunds validates authentication
    - Test GET /customer/orders/:code/refunds validates customer owns order
    - Test GET /customer/refunds/:code returns refund details
    - _Requirements: 14.5, 14.6_
  
  - [x] 14.3 Write integration tests for error scenarios
    - Test 400 Bad Request for invalid refund requests
    - Test 401 Unauthorized for missing auth token
    - Test 403 Forbidden for non-admin users
    - Test 404 Not Found for non-existent orders/refunds
    - Test 409 Conflict for duplicate idempotency keys
    - Test error response format is consistent
    - _Requirements: 7.6, 15.7_

- [x] 15. Write Property-Based Tests for Correctness Properties
  - [x] 15.1 Write property test for foreign key validation (Property 1)
    - **Property 1: Foreign Key Validation**
    - **Validates: Requirements 1.1, 1.3**
  
  - [x] 15.2 Write property test for nullable payment ID handling (Property 2)
    - **Property 2: Nullable Payment ID Handling**
    - **Validates: Requirements 1.2**
  
  - [x] 15.3 Write property test for transaction atomicity (Property 3)
    - **Property 3: Transaction Atomicity**
    - **Validates: Requirements 1.4, 1.5, 7.1**
  
  - [x] 15.4 Write property test for concurrent refund prevention (Property 4)
    - **Property 4: Concurrent Refund Prevention**
    - **Validates: Requirements 1.6, 9.3, 9.5**
  
  - [x] 15.5 Write property test for refund amount calculations (Properties 29-32)
    - **Property 29: Full Refund Amount Calculation**
    - **Property 30: Shipping Only Refund Amount Calculation**
    - **Property 31: Partial Refund Amount Calculation**
    - **Property 32: Item Refund Amount Calculation**
    - **Validates: Requirements 8.1, 8.2, 8.3, 8.4, 8.7**
  
  - [x] 15.6 Write property test for stock restoration (Properties 21-26)
    - **Property 21: Full Refund Stock Restoration**
    - **Property 22: Partial Refund Stock Restoration**
    - **Property 23: Stock Restoration Flag**
    - **Property 26: Stock Restoration Idempotency**
    - **Validates: Requirements 6.1, 6.2, 6.3, 6.4, 6.7**
  
  - [x] 15.7 Write property test for idempotency (Property 34)
    - **Property 34: Idempotency Key Checking**
    - **Validates: Requirements 9.1, 9.2**
  
  - [x] 15.8 Write property test for order status updates (Properties 16-20)
    - **Property 16: Full Refund Order Status Update**
    - **Property 17: Partial Refund Order Status Preservation**
    - **Property 19: Refund Amount Aggregation**
    - **Validates: Requirements 5.1, 5.2, 5.4, 5.5, 5.6**

- [x] 16. Test End-to-End Refund Flows
  - [x] 16.1 Test complete full refund flow
    - Admin creates full refund for delivered order
    - System validates order and payment
    - System creates refund record in database
    - System calls Midtrans API successfully
    - System updates order status to REFUNDED
    - System restores product stock
    - System sends notification to customer
    - Customer sees refund in purchase history
    - _Requirements: All requirements_
  
  - [x] 16.2 Test complete manual refund flow
    - Admin creates refund for order without payment
    - System detects no payment record
    - System creates manual refund (skips gateway)
    - System sets status to COMPLETED immediately
    - System updates order status
    - System restores product stock
    - System sends notification to customer
    - _Requirements: 2.7, 13.2, 13.3, 13.4, 13.5, 13.7_
  
  - [x] 16.3 Test refund failure and retry flow
    - Admin creates refund
    - Midtrans API returns error
    - System marks refund as FAILED
    - Admin clicks retry button
    - System retries with same idempotency key
    - Midtrans API succeeds on retry
    - System marks refund as COMPLETED
    - _Requirements: 2.4, 2.5, 7.2, 7.3, 7.4, 7.5_

- [x] 17. Performance Testing and Optimization
  - [x] 17.1 Test concurrent refund handling
    - Simulate multiple admins creating refunds for same order simultaneously
    - Verify row-level locking prevents race conditions
    - Verify only one refund is created
    - Measure transaction duration under load
    - _Requirements: 1.6, 9.3_
  
  - [x] 17.2 Test database query performance
    - Test refund listing query performance with 10,000+ refunds
    - Test order refund lookup performance
    - Verify indexes are used correctly
    - Optimize slow queries if needed
    - _Requirements: Performance considerations_
  
  - [x] 17.3 Test Midtrans API timeout handling
    - Simulate API timeout scenarios
    - Verify 30-second timeout is enforced
    - Verify refund marked as FAILED on timeout
    - Verify retry works after timeout
    - _Requirements: 7.2_

- [x] 18. Security Testing and Hardening
  - [x] 18.1 Test authentication and authorization
    - Test all endpoints require valid JWT token
    - Test admin endpoints require admin role
    - Test customer endpoints verify order ownership
    - Test expired tokens are rejected
    - _Requirements: 14.6_
  
  - [x] 18.2 Test input validation and SQL injection prevention
    - Test all user inputs are sanitized
    - Test SQL injection attempts are blocked
    - Test XSS attempts are blocked
    - Test malformed JSON is rejected
    - _Requirements: Security considerations_
  
  - [x] 18.3 Test rate limiting
    - Test refund creation rate limit (10 per minute per admin)
    - Verify rate limit returns 429 Too Many Requests
    - Verify rate limit resets after time window
    - _Requirements: Security considerations_

- [x] 19. Monitoring and Logging Setup
  - [x] 19.1 Implement comprehensive logging
    - Log all refund operations at INFO level
    - Log all gateway API calls at DEBUG level
    - Log all errors with stack traces at ERROR level
    - Log all audit trail events at INFO level
    - Include refund code, order code, user ID in all logs
    - _Requirements: 7.7, 10.1, 10.2, 10.3, 10.4, 10.5_
  
  - [x] 19.2 Set up metrics and alerts
    - Track refund creation rate metric
    - Track refund success/failure rate metric
    - Track gateway API response time metric
    - Track database transaction duration metric
    - Set up alert for gateway failure rate > 5%
    - Set up alert for refund processing time > 10 seconds
    - Set up alert for stock restoration failures
    - _Requirements: Monitoring and alerting_

- [x] 20. Documentation and Deployment
  - [x] 20.1 Write API documentation
    - Document all refund endpoints with request/response examples
    - Document error codes and messages
    - Document authentication requirements
    - Document rate limits
    - Create Postman collection for testing
    - _Requirements: All API requirements_
  
  - [x] 20.2 Write deployment guide
    - Document database migration steps
    - Document environment variables needed
    - Document Midtrans configuration
    - Document rollback procedures
    - Create deployment checklist
    - _Requirements: Implementation notes_
  
  - [x] 20.3 Deploy to staging environment
    - Run database migrations on staging
    - Deploy backend code to staging
    - Deploy frontend code to staging
    - Test with Midtrans sandbox
    - Verify all functionality works end-to-end
    - _Requirements: All requirements_
  
  - [x] 20.4 Deploy to production
    - Schedule maintenance window
    - Run database migrations on production
    - Deploy backend code to production
    - Deploy frontend code to production
    - Switch to Midtrans production
    - Monitor for errors
    - Verify refunds work correctly
    - _Requirements: All requirements_

## Notes

- Each task references specific requirements for traceability
- Tasks build incrementally - complete in order for best results
- Test tasks validate correctness at each stage
- All database operations use transactions for safety
- All gateway calls include proper error handling
- All user inputs are validated before processing
- All operations are logged for audit trail
