# Requirements Document

## Introduction

This document specifies the requirements for migrating Zavéra E-Commerce's logistics system from RajaOngkir/Kommerce API to Biteship API. The migration involves complete replacement of all shipping-related functionality including region lookup, shipping rate calculation, order creation, and shipment tracking. Biteship will become the sole logistics authority with no RajaOngkir logic remaining in the codebase.

## Glossary

- **Biteship**: Indonesian logistics aggregator API providing unified access to multiple couriers
- **area_id**: Biteship's unique identifier for geographic areas (replaces RajaOngkir's province/city/district IDs)
- **location_id**: Biteship's identifier for saved pickup/delivery locations
- **draft_order**: Biteship's pre-shipment order that can be confirmed to create actual shipment
- **waybill_id**: Tracking number/AWB assigned by courier after order confirmation
- **tracking_id**: Biteship's internal tracking identifier
- **Shipping_System**: The backend service responsible for all shipping operations
- **Logistics_Client**: The HTTP client module that communicates with Biteship API
- **Migration_System**: The system responsible for data migration and cleanup

## Requirements

### Requirement 1: RajaOngkir Purge

**User Story:** As a system maintainer, I want all RajaOngkir/Kommerce code completely removed, so that the codebase has a single logistics provider and no legacy code confusion.

#### Acceptance Criteria

1. WHEN the migration is complete THEN the Migration_System SHALL have deleted the KommerceClient struct and all its methods from service/kommerce_client.go
2. WHEN the migration is complete THEN the Migration_System SHALL have removed all KOMMERCE_* environment variables from .env and .env.example files
3. WHEN the migration is complete THEN the Migration_System SHALL have removed all rajaongkir_raw_json columns and references from database tables and models
4. WHEN the migration is complete THEN the Migration_System SHALL have updated all code comments referencing RajaOngkir to reference Biteship
5. WHEN the migration is complete THEN the Migration_System SHALL have removed all fallback province/city/district data hardcoded for RajaOngkir API failures

### Requirement 2: Biteship Client Implementation

**User Story:** As a developer, I want a robust Biteship API client, so that I can interact with all Biteship endpoints reliably.

#### Acceptance Criteria

1. WHEN the Logistics_Client is initialized THEN the Logistics_Client SHALL read TOKEN_BITESHIP and BITESHIP_BASE_URL from environment variables
2. WHEN making any API request THEN the Logistics_Client SHALL include Authorization header with format "Bearer {TOKEN_BITESHIP}"
3. WHEN the Logistics_Client calls GET /v1/maps/areas THEN the Logistics_Client SHALL return a list of areas matching the search query with area_id, name, and postal_code
4. WHEN the Logistics_Client calls GET /v1/couriers THEN the Logistics_Client SHALL return a list of available couriers with courier_code, courier_name, and available_services
5. WHEN the Logistics_Client calls POST /v1/locations THEN the Logistics_Client SHALL create a saved location and return location_id
6. WHEN the Logistics_Client calls POST /v1/rates/couriers THEN the Logistics_Client SHALL return shipping rates with courier_code, courier_service_code, price, and duration
7. WHEN the Logistics_Client calls POST /v1/draft_orders THEN the Logistics_Client SHALL create a draft order and return draft_order_id
8. WHEN the Logistics_Client calls POST /v1/draft_orders/:id/confirm THEN the Logistics_Client SHALL confirm the order and return waybill_id and tracking_id
9. WHEN the Logistics_Client calls GET /v1/trackings/:id THEN the Logistics_Client SHALL return tracking status and history
10. WHEN any Biteship API call fails THEN the Logistics_Client SHALL return a structured error with HTTP status code and error message

### Requirement 3: Area-Based Location System

**User Story:** As a customer, I want to search for my delivery location using Biteship's area search, so that I can accurately specify my address for shipping.

#### Acceptance Criteria

1. WHEN a customer searches for an area THEN the Shipping_System SHALL call Biteship /v1/maps/areas with the search query
2. WHEN Biteship returns area results THEN the Shipping_System SHALL display area_id, name (full address path), and postal_code to the customer
3. WHEN a customer selects an area THEN the Shipping_System SHALL store the area_id as the primary location identifier
4. WHEN calculating shipping rates THEN the Shipping_System SHALL use area_id for both origin and destination instead of province/city/district IDs
5. IF the area search returns no results THEN the Shipping_System SHALL display a message asking the customer to refine their search

### Requirement 4: Shipping Rate Calculation

**User Story:** As a customer, I want to see accurate shipping rates from Biteship, so that I can choose the best courier and service for my order.

#### Acceptance Criteria

1. WHEN a customer requests shipping rates THEN the Shipping_System SHALL call Biteship POST /v1/rates/couriers with origin_area_id, destination_area_id, and item details
2. WHEN Biteship returns rates THEN the Shipping_System SHALL display courier_name, service_name, price, and estimated duration
3. WHEN displaying rates THEN the Shipping_System SHALL group rates by category (Express, Regular, Economy) based on service type
4. WHEN a customer selects a shipping rate THEN the Shipping_System SHALL store the selected courier_code, courier_service_code, and price
5. WHEN the selected rate is stored THEN the Shipping_System SHALL create a shipping snapshot with biteship_raw_json containing the full API response

### Requirement 5: Draft Order and Confirmation

**User Story:** As a system, I want to create Biteship draft orders and confirm them after payment, so that shipments are only created for paid orders.

#### Acceptance Criteria

1. WHEN an order is created THEN the Shipping_System SHALL call Biteship POST /v1/draft_orders with shipper, destination, courier, and item details
2. WHEN Biteship returns draft_order_id THEN the Shipping_System SHALL store the draft_order_id in the shipments table
3. WHEN an order payment is confirmed THEN the Shipping_System SHALL call Biteship POST /v1/draft_orders/:id/confirm
4. WHEN Biteship confirms the order THEN the Shipping_System SHALL store the waybill_id and tracking_id in the shipments table
5. IF draft order creation fails THEN the Shipping_System SHALL log the error and allow manual retry by admin

### Requirement 6: Shipment Tracking

**User Story:** As a customer, I want to track my shipment using Biteship tracking, so that I can monitor delivery progress.

#### Acceptance Criteria

1. WHEN a customer requests tracking THEN the Shipping_System SHALL call Biteship GET /v1/trackings/:tracking_id
2. WHEN Biteship returns tracking data THEN the Shipping_System SHALL display current status, courier info, and tracking history
3. WHEN tracking status changes THEN the Shipping_System SHALL update the shipment status in the database
4. WHEN tracking shows DELIVERED status THEN the Shipping_System SHALL update the order status to DELIVERED
5. WHEN the tracking job runs THEN the Shipping_System SHALL poll tracking for all shipments with status IN_TRANSIT or OUT_FOR_DELIVERY

### Requirement 7: Database Migration

**User Story:** As a database administrator, I want proper schema changes for Biteship integration, so that all Biteship data is properly stored.

#### Acceptance Criteria

1. WHEN the migration runs THEN the Migration_System SHALL create biteship_locations table with columns: id, user_id, location_id, area_id, area_name, contact_name, contact_phone, address, postal_code, created_at
2. WHEN the migration runs THEN the Migration_System SHALL add columns to shipments table: biteship_draft_order_id, biteship_order_id, biteship_tracking_id, biteship_waybill_id
3. WHEN the migration runs THEN the Migration_System SHALL rename rajaongkir_raw_json to biteship_raw_json in shipping_snapshots table
4. WHEN the migration runs THEN the Migration_System SHALL create indexes on biteship_draft_order_id, biteship_tracking_id, and biteship_waybill_id columns
5. WHEN the migration runs THEN the Migration_System SHALL preserve existing order data while adding new Biteship columns

### Requirement 8: Serialization Round-Trip

**User Story:** As a developer, I want Biteship API responses to be correctly serialized and deserialized, so that data integrity is maintained.

#### Acceptance Criteria

1. WHEN Biteship API returns a response THEN the Logistics_Client SHALL parse the JSON response into Go structs
2. WHEN storing Biteship response data THEN the Shipping_System SHALL serialize the response to JSON for biteship_raw_json storage
3. WHEN retrieving stored Biteship data THEN the Shipping_System SHALL deserialize the JSON back to the original struct format
4. WHEN round-tripping Biteship data THEN the Shipping_System SHALL preserve all fields without data loss

### Requirement 9: Error Handling

**User Story:** As a system operator, I want comprehensive error handling for Biteship integration, so that failures are properly logged and recoverable.

#### Acceptance Criteria

1. WHEN Biteship API returns 401 Unauthorized THEN the Logistics_Client SHALL return ErrBiteshipUnauthorized error
2. WHEN Biteship API returns 429 Rate Limited THEN the Logistics_Client SHALL return ErrBiteshipRateLimited error and implement exponential backoff
3. WHEN Biteship API returns 5xx Server Error THEN the Logistics_Client SHALL retry up to 3 times with exponential backoff
4. WHEN any Biteship operation fails THEN the Shipping_System SHALL log the error with request details and response body
5. IF draft order confirmation fails THEN the Shipping_System SHALL mark the shipment as CONFIRMATION_FAILED and alert admin

### Requirement 10: Frontend Integration

**User Story:** As a frontend developer, I want updated API endpoints for Biteship integration, so that the checkout flow works with the new system.

#### Acceptance Criteria

1. WHEN the frontend calls GET /api/shipping/areas THEN the Shipping_System SHALL return Biteship area search results
2. WHEN the frontend calls POST /api/shipping/rates THEN the Shipping_System SHALL accept area_id instead of district_id for destination
3. WHEN the frontend displays shipping options THEN the Shipping_System SHALL return courier logos and service descriptions from Biteship data
4. WHEN the frontend submits checkout THEN the Shipping_System SHALL accept destination_area_id in the request body
5. WHEN the frontend displays tracking THEN the Shipping_System SHALL return Biteship tracking data with status and history
