# Implementation Plan

## Phase 1: Database Schema Extensions

- [x] 1. Create database migration for ZAVERA commerce upgrade


  - [x] 1.1 Add PACKING and REFUNDED to order_status enum


    - Extend existing order_status enum type
    - Update ValidOrderTransitions in models/models.go
    - _Requirements: 1.2_
  - [x] 1.2 Add missing columns to orders table

    - Add resi VARCHAR(100)
    - Add delivered_at TIMESTAMP
    - Add origin_city VARCHAR(100) DEFAULT 'Semarang'
    - Add destination_city VARCHAR(100)
    - _Requirements: 1.5, 3.4_

  - [x] 1.3 Create stock_movements table


    - product_id, order_id, type (RESERVE/RELEASE/DEDUCT), quantity, balance_after, created_at
    - Add indexes on product_id and order_id

    - _Requirements: 6.5_
  - [x] 1.4 Create shipping_snapshots table

    - order_id (unique), courier, service, cost, etd, origin_city_id, destination_city_id, weight, rajaongkir_raw_json
    - _Requirements: 2.3_
  - [x] 1.5 Create email_templates table

    - template_key, subject_template, html_template, is_active
    - Seed with ORDER_CREATED, PAYMENT_SUCCESS, ORDER_SHIPPED, ORDER_DELIVERED templates
    - _Requirements: 4.1, 4.2, 4.3, 4.4_

  - [x] 1.6 Create email_logs table


    - order_id, template_key, recipient_email, subject, status, error_message, sent_at
    - _Requirements: 4.5_


- [x] 2. Checkpoint - Verify migration applies cleanly

  - Ensure all tests pass, ask the user if questions arise.

## Phase 2: Order State Machine Enhancement

- [x] 3. Update order state machine in models/models.go




  - [x] 3.1 Add OrderStatusPacking and OrderStatusRefunded constants

    - Add PACKING between PAID and SHIPPED
    - Add REFUNDED as terminal state
    - _Requirements: 1.2_

  - [x] 3.2 Update ValidOrderTransitions map




    - PAID → PACKING (not directly to SHIPPED)
    - PACKING → SHIPPED
    - DELIVERED → COMPLETED or REFUNDED
    - COMPLETED → REFUNDED
    - _Requirements: 1.2_
  - [x] 3.3 Write property test for order state machine transitions


    - **Property 1: Order State Machine Validity**
    - **Validates: Requirements 1.2, 1.3**

- [x] 4. Update order service for new states


  - [x] 4.1 Add MarkAsPacking method to OrderService

    - Validate transition from PAID to PACKING
    - Record status change in audit log
    - _Requirements: 5.3_

  - [x] 4.2 Add MarkAsDelivered method to OrderService


    - Update delivered_at timestamp
    - Validate transition from SHIPPED to DELIVERED
    - _Requirements: 1.5_

  - [x] 4.3 Add MarkAsRefunded method to OrderService


    - Validate transition from DELIVERED/COMPLETED to REFUNDED
    - _Requirements: 7.4_
  - [x] 4.4 Write property test for status change audit logging


    - **Property 3: Status Change Audit Logging**
    - **Validates: Requirements 1.4, 8.1**

## Phase 3: Stock Movement System

- [x] 5. Create stock movement repository and service


  - [x] 5.1 Create StockMovement model in models/models.go

    - ID, ProductID, OrderID, Type, Quantity, BalanceAfter, CreatedAt
    - _Requirements: 6.5_
  - [x] 5.2 Create stock_repository.go


    - RecordMovement(productID, orderID int, movementType string, quantity int) error
    - GetMovementsByProduct(productID int) ([]StockMovement, error)
    - GetMovementsByOrder(orderID int) ([]StockMovement, error)
    - _Requirements: 6.5_
  - [x] 5.3 Update order_repository.go Create method to record RESERVE movement


    - After stock deduction, insert into stock_movements with type RESERVE
    - _Requirements: 6.1, 6.5_

  - [x] 5.4 Update order_repository.go RestoreStock method to record RELEASE movement

    - After stock restoration, insert into stock_movements with type RELEASE
    - _Requirements: 6.2, 6.3, 6.5_
  - [x] 5.5 Write property test for stock reservation


    - **Property 15: Stock Reservation on Checkout**
    - **Validates: Requirements 6.1**
  - [x] 5.6 Write property test for stock release


    - **Property 16: Stock Release on Failure**
    - **Validates: Requirements 6.2, 6.3, 7.3**
  - [x] 5.7 Write property test for stock movement logging


    - **Property 17: Stock Movement Logging**
    - **Validates: Requirements 6.5**

- [x] 6. Checkpoint - Verify stock system works

  - Ensure all tests pass, ask the user if questions arise.

## Phase 4: Shipping Snapshot System

- [x] 7. Create shipping snapshot repository and update checkout


  - [x] 7.1 Create ShippingSnapshot model in models/shipping.go

    - ID, OrderID, Courier, Service, Cost, ETD, OriginCityID, DestinationCityID, Weight, RajaOngkirRawJSON, CreatedAt
    - _Requirements: 2.3_
  - [x] 7.2 Add CreateShippingSnapshot to shipping_repository.go


    - Insert shipping snapshot with raw API response
    - _Requirements: 2.3_
  - [x] 7.3 Add GetShippingSnapshotByOrderID to shipping_repository.go

    - Retrieve snapshot for validation
    - _Requirements: 2.5_
  - [x] 7.4 Update checkout_service.go to create shipping snapshot


    - After getting rates from Kommerce API, store raw response
    - Create snapshot record with all required fields
    - _Requirements: 2.3_
  - [x] 7.5 Add shipping cost validation in checkout


    - Before creating order, verify cost matches snapshot
    - Reject if mismatch detected
    - _Requirements: 9.2_
  - [x] 7.6 Write property test for shipping snapshot storage


    - **Property 7: Shipping Snapshot Storage**
    - **Validates: Requirements 2.3**
  - [x] 7.7 Write property test for shipping data integrity


    - **Property 8: Shipping Data Integrity**
    - **Validates: Requirements 2.4, 2.5, 11.1, 11.2**

## Phase 5: Manual Resi Generator

- [x] 8. Implement resi generation service


  - [x] 8.1 Create resi_service.go


    - GenerateResi(orderID int, courierCode string) (string, error)
    - Format: ZVR-{COURIER}-{YYYYMMDD}-{ORDERID}-{RANDOM4}
    - Verify uniqueness before returning
    - _Requirements: 3.1, 3.2_
  - [x] 8.2 Add IsResiLocked method

    - Return true if order status >= SHIPPED
    - _Requirements: 3.3_

  - [x] 8.3 Add UpdateResi to order_repository.go

    - Check IsResiLocked before update
    - Return error if locked
    - _Requirements: 3.3, 3.4_
  - [x] 8.4 Update fulfillment_service.go MarkShipped to generate resi


    - Call GenerateResi before marking as shipped
    - Store resi in orders.resi column
    - _Requirements: 5.4_
  - [x] 8.5 Write property test for resi format and uniqueness


    - **Property 9: Resi Format and Uniqueness**
    - **Validates: Requirements 3.1, 3.2, 3.4**
  - [x] 8.6 Write property test for resi immutability


    - **Property 10: Resi Immutability After Shipping**
    - **Validates: Requirements 3.3**

- [x] 9. Checkpoint - Verify resi system works

  - Ensure all tests pass, ask the user if questions arise.

## Phase 6: Transactional Email System

- [x] 10. Create email service


  - [x] 10.1 Create email_service.go


    - SendOrderCreated(order *Order, items []OrderItem) error
    - SendPaymentSuccess(order *Order) error
    - SendOrderShipped(order *Order, shipment *Shipment) error
    - SendOrderDelivered(order *Order) error
    - _Requirements: 4.1, 4.2, 4.3, 4.4_
  - [x] 10.2 Create HTML email templates

    - ORDER_CREATED: Product list, shipping address, ongkir, total, payment instructions
    - PAYMENT_SUCCESS: Order confirmation
    - ORDER_SHIPPED: Courier, service, resi, tracking link
    - ORDER_DELIVERED: Delivery confirmation
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_

  - [x] 10.3 Create email_repository.go


    - CreateEmailLog(log *EmailLog) error
    - GetEmailLogsByOrder(orderID int) ([]EmailLog, error)
    - _Requirements: 4.5_
  - [x] 10.4 Integrate email sending into order lifecycle


    - Call SendOrderCreated after order creation
    - Call SendPaymentSuccess after payment confirmation
    - Call SendOrderShipped after marking as shipped
    - Call SendOrderDelivered after marking as delivered
    - _Requirements: 4.1, 4.2, 4.3, 4.4_
  - [x] 10.5 Write property test for email content completeness


    - **Property 11: Email Content Completeness**
    - **Validates: Requirements 4.1, 4.2, 4.3, 4.4, 4.5**

## Phase 7: Admin Fulfillment Panel Enhancement

- [x] 11. Update admin fulfillment handlers


  - [x] 11.1 Add PackOrder endpoint to admin_order_handler.go


    - POST /api/admin/orders/:code/pack
    - Validate order status is PAID
    - Call MarkAsPacking
    - Create audit log
    - _Requirements: 5.3_
  - [x] 11.2 Update ShipOrder endpoint in fulfillment_handler.go

    - Validate order status is PACKING
    - Generate resi using resi_service
    - Call MarkAsShipped
    - Trigger shipping email
    - Create audit log
    - _Requirements: 5.4_
  - [x] 11.3 Add GetOrderActions endpoint

    - Return available actions based on order status
    - Pack button enabled only for PAID orders
    - Ship button enabled only for PACKING orders
    - _Requirements: 5.5_
  - [x] 11.4 Update CancelOrder to validate status

    - Customer: only PENDING allowed
    - Admin: only < SHIPPED allowed
    - _Requirements: 5.6, 7.1, 7.2_
  - [x] 11.5 Write property test for admin pack action


    - **Property 12: Admin Pack Action**
    - **Validates: Requirements 5.3**
  - [x] 11.6 Write property test for admin ship action


    - **Property 13: Admin Ship Action**
    - **Validates: Requirements 5.4**
  - [x] 11.7 Write property test for cancel permission validation


    - **Property 14: Cancel Permission Validation**
    - **Validates: Requirements 5.6, 7.1, 7.2**

- [x] 12. Checkpoint - Verify admin fulfillment works

  - Ensure all tests pass, ask the user if questions arise.

## Phase 8: Data Validation and Dashboard

- [x] 13. Implement server-side validation


  - [x] 13.1 Add ValidateOrderTotals function


    - Recalculate subtotal from items
    - Verify shipping_cost matches snapshot
    - Verify total = subtotal + shipping - discount
    - _Requirements: 9.1, 9.2, 9.3_
  - [x] 13.2 Update checkout to reject frontend price manipulation

    - Ignore frontend-submitted totals
    - Always recalculate server-side
    - _Requirements: 9.3, 11.5_
  - [x] 13.3 Write property test for server-side total calculation


    - **Property 20: Server-Side Total Calculation**
    - **Validates: Requirements 9.1, 9.3, 11.3**

- [x] 14. Implement dashboard metrics


  - [x] 14.1 Add GetDashboardMetrics to admin_service.go


    - Total revenue (sum of PAID/COMPLETED orders)
    - Orders today count
    - Orders shipped count
    - Orders pending count
    - Low stock products (stock < 10)
    - _Requirements: 10.1, 10.2, 10.3_
  - [x] 14.2 Add dashboard endpoint to admin handler


    - GET /api/admin/dashboard/metrics
    - Query database in real-time (no caching)
    - _Requirements: 10.4_
  - [x] 14.3 Write property test for dashboard metrics accuracy


    - **Property 21: Dashboard Metrics Accuracy**
    - **Validates: Requirements 10.1, 10.2, 10.3**

## Phase 9: Audit Log Enhancement

- [x] 15. Verify and enhance audit logging


  - [x] 15.1 Verify admin_audit_log table has immutability trigger


    - Check prevent_audit_update trigger exists
    - Test that UPDATE/DELETE fail
    - _Requirements: 8.2_
  - [x] 15.2 Add comprehensive audit logging to all admin actions

    - Log state_before and state_after for all modifications
    - Include IP address and user agent
    - _Requirements: 8.1_
  - [x] 15.3 Write property test for audit log immutability


    - **Property 19: Audit Log Immutability**
    - **Validates: Requirements 8.2**

## Phase 10: Refund System Integration

- [x] 16. Integrate refund with order lifecycle


  - [x] 16.1 Update refund_service.go to update order status


    - After successful refund, call MarkAsRefunded
    - Restore stock via stock movement system
    - _Requirements: 7.4_
  - [x] 16.2 Add refund eligibility check


    - Only DELIVERED or COMPLETED orders can be refunded
    - _Requirements: 7.4_
  - [x] 16.3 Write property test for refund processing


    - **Property 18: Refund Processing**
    - **Validates: Requirements 7.4**

- [x] 17. Final Checkpoint - Complete system verification

  - Ensure all tests pass, ask the user if questions arise.
