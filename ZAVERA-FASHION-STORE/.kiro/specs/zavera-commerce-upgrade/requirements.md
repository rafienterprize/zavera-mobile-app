# Requirements Document

## Introduction

This document specifies the requirements for upgrading ZAVERA from a basic checkout website into a production-grade e-commerce platform comparable to Tokopedia, Shopee, and Zalora. The upgrade encompasses transactional email, shipping fulfillment, manual airway bill generation, complete order lifecycle management, stock locking, refund/cancel operations, audit logging, and admin-customer synchronization. All operations must be database-driven, consistent, and tamper-proof.

## Glossary

- **ZAVERA**: The e-commerce platform being upgraded
- **Order**: A customer purchase transaction containing products, shipping, and payment information
- **Resi**: Airway bill / tracking number for shipments
- **RajaOngkir**: Indonesian shipping cost API provider
- **Stock Lock**: Temporary reservation of inventory during checkout
- **Audit Log**: Immutable record of all system actions for accountability
- **ETA**: Estimated Time of Arrival for shipments
- **Ongkir**: Shipping cost (Indonesian term)

## Requirements

### Requirement 1: Order Lifecycle State Machine

**User Story:** As a system administrator, I want orders to follow a strict state machine, so that order processing is predictable and auditable.

#### Acceptance Criteria

1. WHEN an order is created THEN the Order_System SHALL set the initial status to PENDING_PAYMENT
2. WHEN an order status changes THEN the Order_System SHALL validate that the transition follows the allowed state machine: PENDING_PAYMENT → PAID → PACKING → SHIPPED → DELIVERED → COMPLETED, with CANCELLED and REFUNDED as terminal states reachable from specific states
3. WHEN an invalid state transition is attempted THEN the Order_System SHALL reject the transition and return an error message
4. WHEN any order status changes THEN the Order_System SHALL create an audit log entry with before and after states
5. WHEN an order is created THEN the Order_System SHALL store: user_id, status, subtotal, shipping_cost, total, courier, service, resi, origin_city, destination_city, weight, paid_at, shipped_at, and delivered_at timestamps

### Requirement 2: RajaOngkir Shipping Integration

**User Story:** As a customer, I want to see accurate shipping costs and delivery estimates from real courier data, so that I can make informed purchasing decisions.

#### Acceptance Criteria

1. WHEN a customer requests shipping options THEN the Shipping_System SHALL query RajaOngkir API with origin (Semarang), destination (user city), and weight (sum of product.weight * qty)
2. WHEN RajaOngkir returns shipping options THEN the Shipping_System SHALL display courier name, service type (REG/YES/ONS), price, and ETA extracted from the API response
3. WHEN a shipping option is selected THEN the Shipping_System SHALL store a shipping_snapshot containing order_id, courier, service, cost, etd, and rajaongkir_raw_json
4. WHEN displaying ETA THEN the Shipping_System SHALL use only the value returned from RajaOngkir API without modification
5. WHEN the order shipping cost is validated THEN the Shipping_System SHALL verify it matches the stored RajaOngkir snapshot exactly

### Requirement 3: Manual Resi Generator

**User Story:** As an admin, I want to generate unique airway bill numbers for orders, so that shipments can be tracked even without RajaOngkir Pro integration.

#### Acceptance Criteria

1. WHEN an admin generates a resi THEN the Resi_System SHALL create a unique identifier in format: ZVR-{COURIER}-{YYYYMMDD}-{ORDERID}-{RANDOM4}
2. WHEN a resi is generated THEN the Resi_System SHALL verify uniqueness against all existing resi values before saving
3. WHEN an order has status SHIPPED or beyond THEN the Resi_System SHALL prevent any modification to the resi field
4. WHEN a resi is saved THEN the Resi_System SHALL store it in the orders.resi column

### Requirement 4: Transactional Email System

**User Story:** As a customer, I want to receive HTML email notifications at key order milestones, so that I stay informed about my purchase status.

#### Acceptance Criteria

1. WHEN an order is created THEN the Email_System SHALL send an HTML email with subject "🛍️ Pesanan ZAVERA #{{order_id}} telah dibuat" containing product list, shipping address, ongkir, total, and payment instructions
2. WHEN payment is confirmed THEN the Email_System SHALL send an HTML email with subject "💳 Pembayaran diterima – Pesanan #{{order_id}}"
3. WHEN an order is shipped THEN the Email_System SHALL send an HTML email with subject "📦 Pesanan #{{order_id}} sedang dikirim" containing courier, service, resi, and tracking link
4. WHEN an order is delivered THEN the Email_System SHALL send an HTML email with subject "🎉 Pesanan #{{order_id}} sudah sampai"
5. WHEN sending any transactional email THEN the Email_System SHALL use HTML format with proper styling

### Requirement 5: Admin Fulfillment Panel

**User Story:** As an admin, I want a fulfillment panel to manage orders through their lifecycle, so that I can efficiently process customer orders.

#### Acceptance Criteria

1. WHEN an admin views the fulfillment panel THEN the Admin_System SHALL display all orders with filtering by status
2. WHEN an admin opens an order THEN the Admin_System SHALL display complete order details including products, customer info, and shipping information
3. WHEN an admin clicks "Pack" on a PAID order THEN the Admin_System SHALL change status to PACKING and log the action
4. WHEN an admin clicks "Ship" on a PACKING order THEN the Admin_System SHALL generate resi, change status to SHIPPED, and trigger the shipping email
5. WHEN an order status is PAID THEN the Admin_System SHALL enable the shipping button; otherwise the button SHALL be disabled
6. WHEN an admin cancels an order THEN the Admin_System SHALL verify the order status is less than SHIPPED before allowing cancellation

### Requirement 6: Stock Locking System

**User Story:** As a system operator, I want stock to be reserved during checkout and managed based on payment outcomes, so that inventory remains accurate.

#### Acceptance Criteria

1. WHEN a customer initiates checkout THEN the Stock_System SHALL reserve (lock) the requested quantity for each product
2. WHEN payment fails or times out after 30 minutes THEN the Stock_System SHALL release the reserved stock back to available inventory
3. WHEN an order is cancelled THEN the Stock_System SHALL restore the reserved stock to available inventory
4. WHEN payment is confirmed THEN the Stock_System SHALL permanently deduct the reserved stock from inventory
5. WHEN any stock movement occurs THEN the Stock_System SHALL record it in stock_movements table with product_id, order_id, type (reserve/release/deduct), qty, and timestamp

### Requirement 7: Cancel and Refund Operations

**User Story:** As a customer or admin, I want to cancel orders and process refunds under appropriate conditions, so that mistakes can be corrected fairly.

#### Acceptance Criteria

1. WHEN a customer requests cancellation THEN the Cancel_System SHALL allow it only if order status is PENDING_PAYMENT
2. WHEN an admin requests cancellation THEN the Cancel_System SHALL allow it only if order status is before SHIPPED
3. WHEN a cancellation is processed THEN the Cancel_System SHALL reverse stock reservations and update order status to CANCELLED
4. WHEN a refund is processed THEN the Refund_System SHALL log the reason, reverse stock, and update order status to REFUNDED

### Requirement 8: Audit Log System

**User Story:** As a system auditor, I want all actions logged immutably, so that no admin actions can be hidden or tampered with.

#### Acceptance Criteria

1. WHEN any data modification occurs THEN the Audit_System SHALL create a log entry with actor_type (admin/system/customer), actor_id, action, table, row_id, before state, after state, IP address, and timestamp
2. WHEN an audit log entry is created THEN the Audit_System SHALL prevent any modification or deletion of that entry
3. WHEN querying audit logs THEN the Audit_System SHALL return complete history for any specified table and row_id

### Requirement 9: Data Consistency Validation

**User Story:** As a system operator, I want automatic validation of order totals and shipping costs, so that no manual price injection can occur.

#### Acceptance Criteria

1. WHEN an order total is calculated THEN the Validation_System SHALL compute it server-side from product prices and quantities
2. WHEN an order is saved THEN the Validation_System SHALL verify shipping_cost matches the RajaOngkir snapshot exactly
3. WHEN any price data is submitted from frontend THEN the Validation_System SHALL recalculate and validate against backend data before accepting

### Requirement 10: Admin Dashboard Metrics

**User Story:** As an admin, I want to see real-time business metrics on the dashboard, so that I can monitor store performance.

#### Acceptance Criteria

1. WHEN an admin views the dashboard THEN the Dashboard_System SHALL display total revenue calculated from database
2. WHEN an admin views the dashboard THEN the Dashboard_System SHALL display orders created today, orders shipped, and orders pending from database queries
3. WHEN an admin views the dashboard THEN the Dashboard_System SHALL display products with low stock levels
4. WHEN displaying any metric THEN the Dashboard_System SHALL query the database in real-time without caching stale data

### Requirement 11: No Fake Data Rule

**User Story:** As a system architect, I want all displayed data to come from authoritative sources, so that customers and admins see accurate information.

#### Acceptance Criteria

1. WHEN displaying ETA THEN the Display_System SHALL use only values from RajaOngkir API responses
2. WHEN displaying shipping cost THEN the Display_System SHALL use only values from RajaOngkir API responses
3. WHEN displaying order totals THEN the Display_System SHALL use only server-calculated values
4. WHEN displaying order status THEN the Display_System SHALL use only backend-stored values
5. WHEN processing any transaction THEN the Backend_System SHALL validate all data server-side without trusting frontend-submitted values
