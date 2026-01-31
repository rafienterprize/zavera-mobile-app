# Implementation Plan

## Phase 1: Database Migration & Schema Changes

- [x] 1. Create database migration script for Biteship integration

  - [x] 1.1 Create biteship_locations table


    - Columns: id, user_id, location_id, area_id, area_name, contact_name, contact_phone, address, postal_code, created_at
    - Add foreign key to users table
    - Add index on user_id and location_id
    - _Requirements: 7.1_

  - [x] 1.2 Add Biteship columns to shipments table
    - Add columns: biteship_draft_order_id, biteship_order_id, biteship_tracking_id, biteship_waybill_id
    - All columns VARCHAR(100), nullable
    - _Requirements: 7.2_
  - [x] 1.3 Create indexes on new shipments columns

    - Index on biteship_draft_order_id
    - Index on biteship_tracking_id
    - Index on biteship_waybill_id
    - _Requirements: 7.4_

  - [x] 1.4 Rename rajaongkir_raw_json to biteship_raw_json in shipping_snapshots
    - Use ALTER TABLE RENAME COLUMN
    - Preserve existing data
    - _Requirements: 7.3_
  - [x] 1.5 Add area_id columns to shipping_snapshots

    - Add origin_area_id, origin_area_name, destination_area_id, destination_area_name
    - _Requirements: 7.5_
  - [x] 1.6 Add area_id column to user_addresses table







    - Add area_id VARCHAR(100) column
    - _Requirements: 3.3_

## Phase 2: BiteshipClient Implementation

- [x] 2. Implement BiteshipClient struct and initialization


  - [x] 2.1 Create service/biteship_client.go with BiteshipClient struct
    - Read TOKEN_BITESHIP and BITESHIP_BASE_URL from environment
    - Initialize HTTP client with 30s timeout
    - Define error types: ErrBiteshipUnauthorized, ErrBiteshipRateLimited, ErrBiteshipAPIFailed
    - _Requirements: 2.1, 2.10_
  - [ ]* 2.2 Write property test for environment variable initialization
    - **Property 1: Authorization Header Format**
    - **Validates: Requirements 2.2**
  - [x] 2.3 Implement doRequest helper with retry logic
    - Add Authorization header with Bearer token
    - Implement exponential backoff for 429 and 5xx errors
    - Max 3 retries, initial backoff 1s, max backoff 30s
    - _Requirements: 9.2, 9.3_
  - [ ]* 2.4 Write property test for retry behavior
    - **Property 11: Retry Behavior**
    - **Validates: Requirements 9.3**

- [x] 3. Implement Biteship API response types
  - [x] 3.1 Create dto/biteship_dto.go with all response structs
    - BiteshipArea, BiteshipCourier, BiteshipRate
    - BiteshipDraftOrder, BiteshipOrder, BiteshipTracking
    - BiteshipTrackingHistory, BiteshipAPIError
    - _Requirements: 2.3, 2.4, 2.6, 2.7, 2.8, 2.9_
  - [ ]* 3.2 Write property test for JSON round-trip
    - **Property 9: Biteship JSON Round-Trip**
    - **Validates: Requirements 8.4**

- [x] 4. Implement BiteshipClient API methods
  - [x] 4.1 Implement SearchAreas method
    - GET /v1/maps/areas?countries=ID&input={query}
    - Parse response into []BiteshipArea
    - _Requirements: 2.3_
  - [ ]* 4.2 Write property test for area search response parsing
    - **Property 2: Area Search Response Parsing**
    - **Validates: Requirements 2.3, 3.2**
  - [x] 4.3 Implement GetCouriers method
    - GET /v1/couriers
    - Parse response into []BiteshipCourier
    - _Requirements: 2.4_
  - [x] 4.4 Implement CreateLocation method
    - POST /v1/locations
    - Return location_id from response
    - _Requirements: 2.5_
  - [x] 4.5 Implement GetRates method
    - POST /v1/rates/couriers with origin_area_id, destination_area_id, items
    - Parse response into []BiteshipRate
    - _Requirements: 2.6_
  - [ ]* 4.6 Write property test for rate response parsing
    - **Property 3: Rate Response Parsing**
    - **Validates: Requirements 2.6, 4.2**
  - [x] 4.7 Implement CreateDraftOrder method
    - POST /v1/draft_orders with shipper, destination, courier, items
    - Return draft_order_id from response
    - _Requirements: 2.7_
  - [x] 4.8 Implement ConfirmDraftOrder method
    - POST /v1/draft_orders/{id}/confirm
    - Return waybill_id and tracking_id from response
    - _Requirements: 2.8_
  - [x] 4.9 Implement GetTracking method
    - GET /v1/trackings/{tracking_id}
    - Parse response into BiteshipTracking
    - _Requirements: 2.9_
  - [ ]* 4.10 Write property test for error status code mapping
    - **Property 10: Error Status Code Mapping**
    - **Validates: Requirements 9.1, 9.2**

- [x] 5. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

## Phase 3: Model Updates

- [x] 6. Update Go models for Biteship integration
  - [x] 6.1 Update models/shipping.go ShippingSnapshot struct
    - Rename RajaOngkirRawJSON to BiteshipRawJSON
    - Add OriginAreaID, OriginAreaName, DestinationAreaID, DestinationAreaName fields
    - _Requirements: 7.3, 7.5_
  - [x] 6.2 Update models/shipping.go Shipment struct
    - Add BiteshipDraftOrderID, BiteshipOrderID, BiteshipTrackingID, BiteshipWaybillID fields
    - _Requirements: 7.2_
  - [x] 6.3 Create models/biteship.go with BiteshipLocation struct
    - ID, UserID, LocationID, AreaID, AreaName, ContactName, ContactPhone, Address, PostalCode, CreatedAt
    - _Requirements: 7.1_
  - [x] 6.4 Add Biteship status mapping function
    - MapBiteshipStatusToShipmentStatus function
    - Map: confirmed→PROCESSING, picked→SHIPPED, delivered→DELIVERED, etc.
    - _Requirements: 6.3_
  - [ ]* 6.5 Write property test for tracking status mapping
    - **Property 7: Tracking Status Mapping**
    - **Validates: Requirements 6.3**

## Phase 4: Repository Layer Updates

- [x] 7. Update ShippingRepository for Biteship
  - [x] 7.1 Update CreateShippingSnapshot to use biteship_raw_json
    - Change column name in INSERT/UPDATE queries
    - Use BiteshipRawJSON field from model
    - _Requirements: 7.3_
  - [x] 7.2 Update GetShippingSnapshotByOrderID to use biteship_raw_json
    - Change column name in SELECT query
    - Map to BiteshipRawJSON field
    - _Requirements: 7.3_
  - [x] 7.3 Add UpdateShipmentBiteshipIDs method
    - Update biteship_draft_order_id, biteship_order_id, biteship_tracking_id, biteship_waybill_id
    - _Requirements: 5.2, 5.4_
  - [ ]* 7.4 Write property test for draft order ID persistence
    - **Property 5: Draft Order ID Persistence**
    - **Validates: Requirements 5.2**
  - [ ]* 7.5 Write property test for order confirmation persistence
    - **Property 6: Order Confirmation Persistence**
    - **Validates: Requirements 5.4**
  - [x] 7.6 Add CreateBiteshipLocation method
    - INSERT into biteship_locations table
    - Return created location with ID
    - _Requirements: 7.1_
  - [x] 7.7 Add GetBiteshipLocationsByUserID method
    - SELECT from biteship_locations WHERE user_id = ?
    - _Requirements: 7.1_
  - [x] 7.8 Update GetShipmentsForTracking to include Biteship tracking_id
    - Add biteship_tracking_id to SELECT
    - Filter by shipments with biteship_tracking_id IS NOT NULL
    - _Requirements: 6.5_

- [x] 8. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

## Phase 5: Service Layer Updates

- [x] 9. Update ShippingService for Biteship integration
  - [x] 9.1 Replace KommerceClient with BiteshipClient in ShippingService
    - Update NewShippingService to create BiteshipClient
    - Remove kommerce field, add biteship field
    - _Requirements: 2.1_
  - [x] 9.2 Implement SearchAreas method in ShippingService
    - Call BiteshipClient.SearchAreas
    - Transform to dto.AreaResponse slice
    - _Requirements: 3.1, 3.2_
  - [x] 9.3 Update GetShippingRates to use area_id
    - Accept destination_area_id instead of district_id
    - Call BiteshipClient.GetRates with area_ids
    - _Requirements: 3.4, 4.1_
  - [ ]* 9.4 Write property test for area ID usage in rates
    - **Property 12: Area ID Usage in Rates**
    - **Validates: Requirements 3.4, 4.1**
  - [x] 9.5 Implement rate categorization logic
    - Categorize by service_type: same_day/instant→Express, standard/express→Regular, economy/cargo→Economy
    - _Requirements: 4.3_
  - [ ]* 9.6 Write property test for rate categorization
    - **Property 4: Rate Categorization**
    - **Validates: Requirements 4.3**
  - [x] 9.7 Update GetCartShippingPreview to use area_id
    - Accept destinationAreaID parameter
    - Call updated GetShippingRates
    - _Requirements: 4.1_
  - [x] 9.8 Implement CreateDraftOrderForCheckout method
    - Build draft order request from order data
    - Call BiteshipClient.CreateDraftOrder
    - Store draft_order_id via repository
    - _Requirements: 5.1, 5.2_
  - [x] 9.9 Implement ConfirmDraftOrder method
    - Call BiteshipClient.ConfirmDraftOrder
    - Store waybill_id and tracking_id via repository
    - _Requirements: 5.3, 5.4_
  - [x] 9.10 Update RefreshTracking to use Biteship
    - Get biteship_tracking_id from shipment
    - Call BiteshipClient.GetTracking
    - Map status and update shipment
    - _Requirements: 6.1, 6.2, 6.3_
  - [ ]* 9.11 Write property test for delivered status propagation
    - **Property 8: Delivered Status Propagation**
    - **Validates: Requirements 6.4**
  - [x] 9.12 Update RunTrackingJob to use Biteship tracking
    - Filter shipments with biteship_tracking_id
    - Call RefreshTracking for each
    - _Requirements: 6.5_

- [x] 10. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

## Phase 6: Handler Layer Updates

- [x] 11. Update ShippingHandler for Biteship endpoints
  - [x] 11.1 Add SearchAreas handler
    - GET /api/shipping/areas?q={query}
    - Call ShippingService.SearchAreas
    - Return []AreaResponse
    - _Requirements: 10.1_
  - [x] 11.2 Update GetShippingRates handler
    - Accept destination_area_id in request body
    - Call updated ShippingService.GetShippingRates
    - _Requirements: 10.2_
  - [x] 11.3 Update GetCartShippingPreview handler
    - Accept destination_area_id query parameter
    - Call updated ShippingService.GetCartShippingPreview
    - _Requirements: 10.2_
  - [x] 11.4 Update tracking handlers to use Biteship data
    - Return Biteship tracking data with status and history
    - _Requirements: 10.5_

- [x] 12. Update CheckoutHandler for Biteship integration
  - [x] 12.1 Update checkout request DTO
    - Add destination_area_id field
    - Make district_id optional for backward compatibility
    - _Requirements: 10.4_
  - [x] 12.2 Update ProcessCheckout to create draft order
    - After order creation, call ShippingService.CreateDraftOrderForCheckout
    - Store draft_order_id in shipment
    - _Requirements: 5.1, 5.2_
  - [x] 12.3 Update payment confirmation handler
    - After payment confirmed, call ShippingService.ConfirmDraftOrder
    - Store waybill_id and tracking_id
    - _Requirements: 5.3, 5.4_

- [x] 13. Update routes registration
  - [x] 13.1 Add new /api/shipping/areas route
    - Register GET handler for area search
    - _Requirements: 10.1_
  - [x] 13.2 Remove deprecated RajaOngkir-specific routes
    - Remove /api/shipping/districts if no longer needed
    - Remove /api/shipping/kelurahan if no longer needed
    - _Requirements: 1.5_

- [x] 14. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

## Phase 7: RajaOngkir/Kommerce Cleanup

- [x] 15. Remove RajaOngkir/Kommerce code
  - [x] 15.1 Delete service/kommerce_client.go
    - Remove entire file
    - _Requirements: 1.1_
  - [x] 15.2 Remove KOMMERCE_* environment variables
    - Remove from backend/.env
    - Remove from backend/.env.example
    - _Requirements: 1.2_
  - [x] 15.3 Remove fallback province/city/district data from handlers
    - Remove fallbackProvinces, fallbackCitiesJawaTengah from shipping_handler.go
    - Remove fallback district/kelurahan data
    - _Requirements: 1.5_
  - [x] 15.4 Update code comments
    - Replace "RajaOngkir" references with "Biteship" in comments
    - Update handler comments for new endpoints
    - _Requirements: 1.4_
  - [x] 15.5 Remove deprecated location endpoints
    - Remove GetProvinces, GetCities, GetDistricts, GetSubdistrictsAPI handlers if replaced
    - Or update them to use Biteship area search
    - _Requirements: 1.5_

## Phase 8: Frontend Integration Updates

- [x] 16. Update frontend shipping types
  - [x] 16.1 Update frontend/src/types/shipping.ts
    - Add BiteshipArea type with area_id, name, postal_code
    - Update ShippingAddress to use area_id
    - _Requirements: 10.1, 10.4_
  - [x] 16.2 Update frontend/src/lib/api.ts
    - Add searchAreas function calling GET /api/shipping/areas
    - Update getShippingRates to use area_id
    - _Requirements: 10.1, 10.2_

- [x] 17. Update checkout page for area-based selection
  - [x] 17.1 Replace province/city/district selectors with area search
    - Add area search input with autocomplete
    - Display area results with full path and postal code
    - Store selected area_id
    - _Requirements: 3.1, 3.2, 3.3_
  - [x] 17.2 Update shipping rate request
    - Send destination_area_id instead of district_id
    - _Requirements: 10.2_
  - [x] 17.3 Update checkout submission
    - Include destination_area_id in checkout request
    - _Requirements: 10.4_

- [x] 18. Update order tracking display
  - [x] 18.1 Update tracking page to display Biteship data
    - Show waybill_id as tracking number
    - Display tracking history from Biteship
    - _Requirements: 10.5_

## Phase 9: Final Testing & Verification

- [x] 19. Integration testing
  - [ ]* 19.1 Write integration test for area search flow
    - Mock Biteship API response
    - Verify area results displayed correctly
    - _Requirements: 3.1, 3.2_
  - [ ]* 19.2 Write integration test for checkout flow
    - Mock Biteship draft order and confirmation
    - Verify draft_order_id and waybill_id stored
    - _Requirements: 5.1, 5.2, 5.3, 5.4_
  - [ ]* 19.3 Write integration test for tracking flow
    - Mock Biteship tracking response
    - Verify status mapping and history display
    - _Requirements: 6.1, 6.2, 6.3, 6.4_

- [x] 20. Final Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.
