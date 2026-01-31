# Design Document: Biteship Migration

## Overview

This design document describes the architecture for migrating Zavéra E-Commerce from RajaOngkir/Kommerce API to Biteship API. The migration involves replacing all shipping-related functionality while preserving the existing payment flow and order lifecycle. The design ensures clean separation between the Logistics_Client (Biteship API communication), Shipping_System (business logic), and database layer.

### Current State Analysis

**Existing Components to Replace:**
- `service/kommerce_client.go` - KommerceClient with RajaOngkir API calls
- `service/shipping_service.go` - ShippingService using KommerceClient
- `handler/shipping_handler.go` - HTTP handlers with RajaOngkir-specific endpoints
- `repository/shipping_repository.go` - Repository with rajaongkir_raw_json references
- `models/models.go` - ShippingSnapshot with RajaOngkirRawJSON field

**Existing Components to Preserve:**
- Payment flow (VA generation, payment waiting state, confirmation)
- Order lifecycle state machine
- Shipment status state machine
- Admin fulfillment panel
- Email notification system

## Architecture

### High-Level System Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              FRONTEND (Next.js)                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │ Area Search │  │ Rate Select │  │  Checkout   │  │ Order Tracking      │ │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └──────────┬──────────┘ │
└─────────┼────────────────┼────────────────┼────────────────────┼────────────┘
          │                │                │                    │
          ▼                ▼                ▼                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           API LAYER (Gin Handlers)                           │
│  ┌─────────────────────────────────────────────────────────────────────────┐│
│  │                        ShippingHandler                                   ││
│  │  GET /areas  POST /rates  POST /checkout  GET /tracking/:id             ││
│  └─────────────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────────────┘
          │                │                │                    │
          ▼                ▼                ▼                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         SERVICE LAYER                                        │
│  ┌─────────────────────────────────────────────────────────────────────────┐│
│  │                      ShippingService                                     ││
│  │  - SearchAreas()      - GetShippingRates()                              ││
│  │  - CreateDraftOrder() - ConfirmOrder()                                  ││
│  │  - GetTracking()      - RefreshTracking()                               ││
│  └──────────────────────────────┬──────────────────────────────────────────┘│
│                                 │                                            │
│  ┌──────────────────────────────▼──────────────────────────────────────────┐│
│  │                      BiteshipClient                                      ││
│  │  - GET  /v1/maps/areas           - GET  /v1/couriers                    ││
│  │  - POST /v1/locations            - POST /v1/rates/couriers              ││
│  │  - POST /v1/draft_orders         - POST /v1/draft_orders/:id/confirm    ││
│  │  - GET  /v1/trackings/:id                                               ││
│  └──────────────────────────────┬──────────────────────────────────────────┘│
└─────────────────────────────────┼───────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         EXTERNAL: Biteship API                               │
│                    https://api.biteship.com/v1/*                            │
│                    Authorization: Bearer {TOKEN_BITESHIP}                    │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                         REPOSITORY LAYER                                     │
│  ┌─────────────────────────────────────────────────────────────────────────┐│
│  │                    ShippingRepository                                    ││
│  │  - CreateBiteshipLocation()  - GetShipmentByOrderID()                   ││
│  │  - CreateShippingSnapshot()  - UpdateShipmentBiteshipIDs()              ││
│  │  - AddTrackingEvent()        - GetShipmentsForTracking()                ││
│  └──────────────────────────────┬──────────────────────────────────────────┘│
└─────────────────────────────────┼───────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         DATABASE (PostgreSQL)                                │
│  ┌───────────────┐  ┌───────────────────┐  ┌─────────────────────────────┐ │
│  │biteship_      │  │shipping_snapshots │  │shipments                    │ │
│  │locations      │  │(biteship_raw_json)│  │(biteship_draft_order_id,    │ │
│  │               │  │                   │  │ biteship_tracking_id,       │ │
│  │               │  │                   │  │ biteship_waybill_id)        │ │
│  └───────────────┘  └───────────────────┘  └─────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Service Boundaries and Responsibilities

| Component | Responsibility |
|-----------|---------------|
| **BiteshipClient** | HTTP communication with Biteship API, request/response serialization, error mapping, retry logic |
| **ShippingService** | Business logic for shipping operations, orchestrates BiteshipClient and Repository calls |
| **ShippingHandler** | HTTP request/response handling, input validation, authentication |
| **ShippingRepository** | Database operations for shipping-related tables |

## Components and Interfaces

### BiteshipClient Interface

```go
type BiteshipClient interface {
    // Area search
    SearchAreas(query string) ([]BiteshipArea, error)
    
    // Couriers
    GetCouriers() ([]BiteshipCourier, error)
    
    // Locations
    CreateLocation(req CreateLocationRequest) (*BiteshipLocation, error)
    
    // Rates
    GetRates(req GetRatesRequest) ([]BiteshipRate, error)
    
    // Orders
    CreateDraftOrder(req CreateDraftOrderRequest) (*BiteshipDraftOrder, error)
    ConfirmDraftOrder(draftOrderID string) (*BiteshipOrder, error)
    
    // Tracking
    GetTracking(trackingID string) (*BiteshipTracking, error)
}
```

### ShippingService Interface (Updated)

```go
type ShippingService interface {
    // Area-based location (NEW - replaces province/city/district)
    SearchAreas(query string) ([]dto.AreaResponse, error)
    
    // Rates (UPDATED - uses area_id)
    GetShippingRates(req dto.GetShippingRatesRequest) (*dto.ShippingRatesResponse, error)
    GetCartShippingPreview(sessionID string, destinationAreaID string) (*dto.CartShippingPreviewResponse, error)
    
    // Draft Orders (NEW)
    CreateDraftOrderForCheckout(orderID int, req dto.CreateDraftOrderRequest) (*dto.DraftOrderResponse, error)
    ConfirmDraftOrder(orderID int) (*dto.OrderConfirmationResponse, error)
    
    // Tracking (UPDATED - uses Biteship tracking_id)
    GetTracking(shipmentID int) (*dto.ShipmentResponse, error)
    RefreshTracking(shipmentID int) error
    RunTrackingJob() error
    
    // Existing (preserved)
    GetProviders() ([]models.ShippingProvider, error)
    CreateShipmentForOrder(...) (*models.Shipment, error)
    // ... address management methods
}
```

## Data Models

### Biteship API Response Types

```go
// BiteshipArea represents an area from /v1/maps/areas
type BiteshipArea struct {
    ID              string `json:"id"`               // area_id
    Name            string `json:"name"`             // Full path: "Kelurahan, Kecamatan, Kota, Provinsi"
    CountryName     string `json:"country_name"`
    CountryCode     string `json:"country_code"`
    AdministrativeLevel1 string `json:"administrative_division_level_1_name"` // Province
    AdministrativeLevel2 string `json:"administrative_division_level_2_name"` // City
    AdministrativeLevel3 string `json:"administrative_division_level_3_name"` // District
    AdministrativeLevel4 string `json:"administrative_division_level_4_name"` // Subdistrict
    PostalCode      string `json:"postal_code"`
}

// BiteshipCourier represents a courier from /v1/couriers
type BiteshipCourier struct {
    AvailableForCashOnDelivery  bool     `json:"available_for_cash_on_delivery"`
    AvailableForProofOfDelivery bool     `json:"available_for_proof_of_delivery"`
    AvailableForInstantWaybillID bool    `json:"available_for_instant_waybill_id"`
    CourierCode                 string   `json:"courier_code"`
    CourierName                 string   `json:"courier_name"`
    CourierServiceCode          string   `json:"courier_service_code"`
    CourierServiceName          string   `json:"courier_service_name"`
    Tier                        string   `json:"tier"`
    Description                 string   `json:"description"`
    ServiceType                 string   `json:"service_type"`
    ShippingType                string   `json:"shipping_type"`
    Duration                    string   `json:"duration"`
}

// BiteshipRate represents a shipping rate from /v1/rates/couriers
type BiteshipRate struct {
    CourierCode        string  `json:"courier_code"`
    CourierName        string  `json:"courier_name"`
    CourierServiceCode string  `json:"courier_service_code"`
    CourierServiceName string  `json:"courier_service_name"`
    Description        string  `json:"description"`
    Duration           string  `json:"duration"`
    ShipmentDurationRange string `json:"shipment_duration_range"`
    ShipmentDurationUnit string `json:"shipment_duration_unit"`
    ServiceType        string  `json:"service_type"`
    ShippingType       string  `json:"shipping_type"`
    Price              float64 `json:"price"`
    Type               string  `json:"type"`
}

// BiteshipDraftOrder represents a draft order response
type BiteshipDraftOrder struct {
    ID        string `json:"id"`
    Status    string `json:"status"`
    CreatedAt string `json:"created_at"`
}

// BiteshipOrder represents a confirmed order response
type BiteshipOrder struct {
    ID          string `json:"id"`
    Status      string `json:"status"`
    WaybillID   string `json:"waybill_id"`
    TrackingID  string `json:"tracking_id"`
    CourierCode string `json:"courier_code"`
    CreatedAt   string `json:"created_at"`
}

// BiteshipTracking represents tracking information
type BiteshipTracking struct {
    ID            string `json:"id"`
    WaybillID     string `json:"waybill_id"`
    CourierCode   string `json:"courier_code"`
    CourierName   string `json:"courier_name"`
    Status        string `json:"status"`
    OriginAddress string `json:"origin_address"`
    DestinationAddress string `json:"destination_address"`
    History       []BiteshipTrackingHistory `json:"history"`
}

type BiteshipTrackingHistory struct {
    Note      string `json:"note"`
    Status    string `json:"status"`
    UpdatedAt string `json:"updated_at"`
}
```

### Database Models (Updated)

```go
// ShippingSnapshot - UPDATED
type ShippingSnapshot struct {
    ID                    int            `json:"id" db:"id"`
    OrderID               int            `json:"order_id" db:"order_id"`
    Courier               string         `json:"courier" db:"courier"`
    Service               string         `json:"service" db:"service"`
    Cost                  float64        `json:"cost" db:"cost"`
    ETD                   string         `json:"etd" db:"etd"`
    OriginAreaID          string         `json:"origin_area_id" db:"origin_area_id"`
    OriginAreaName        string         `json:"origin_area_name" db:"origin_area_name"`
    DestinationAreaID     string         `json:"destination_area_id" db:"destination_area_id"`
    DestinationAreaName   string         `json:"destination_area_name" db:"destination_area_name"`
    Weight                int            `json:"weight" db:"weight"`
    BiteshipRawJSON       map[string]any `json:"biteship_raw_json" db:"biteship_raw_json"` // Renamed from rajaongkir_raw_json
    CreatedAt             time.Time      `json:"created_at" db:"created_at"`
}

// Shipment - UPDATED with Biteship fields
type Shipment struct {
    // ... existing fields ...
    
    // NEW Biteship fields
    BiteshipDraftOrderID string `json:"biteship_draft_order_id,omitempty" db:"biteship_draft_order_id"`
    BiteshipOrderID      string `json:"biteship_order_id,omitempty" db:"biteship_order_id"`
    BiteshipTrackingID   string `json:"biteship_tracking_id,omitempty" db:"biteship_tracking_id"`
    BiteshipWaybillID    string `json:"biteship_waybill_id,omitempty" db:"biteship_waybill_id"`
}

// BiteshipLocation - NEW table
type BiteshipLocation struct {
    ID           int       `json:"id" db:"id"`
    UserID       *int      `json:"user_id,omitempty" db:"user_id"`
    LocationID   string    `json:"location_id" db:"location_id"`       // Biteship location_id
    AreaID       string    `json:"area_id" db:"area_id"`               // Biteship area_id
    AreaName     string    `json:"area_name" db:"area_name"`           // Full area path
    ContactName  string    `json:"contact_name" db:"contact_name"`
    ContactPhone string    `json:"contact_phone" db:"contact_phone"`
    Address      string    `json:"address" db:"address"`
    PostalCode   string    `json:"postal_code" db:"postal_code"`
    CreatedAt    time.Time `json:"created_at" db:"created_at"`
}
```

## Shipment Lifecycle State Diagram

**DESIGN ASPECT EXISTS — NO ACTION NEEDED**

The existing shipment status state machine in `models/shipping.go` is comprehensive and does not require changes. The Biteship integration will map Biteship tracking statuses to existing ShipmentStatus values.

```
Biteship Status Mapping:
┌─────────────────────────┬──────────────────────────┐
│ Biteship Status         │ ShipmentStatus           │
├─────────────────────────┼──────────────────────────┤
│ confirmed               │ PROCESSING               │
│ allocated               │ PICKUP_SCHEDULED         │
│ picking_up              │ PICKUP_SCHEDULED         │
│ picked                  │ SHIPPED                  │
│ dropping_off            │ IN_TRANSIT               │
│ delivered               │ DELIVERED                │
│ rejected                │ DELIVERY_FAILED          │
│ courier_not_found       │ PICKUP_FAILED            │
│ returned                │ RETURNED_TO_SENDER       │
│ cancelled               │ CANCELLED                │
└─────────────────────────┴──────────────────────────┘
```

## API Interaction Flow

### Checkout Flow with Biteship

```
Customer                Frontend              ShippingHandler         ShippingService         BiteshipClient          Database
   │                       │                        │                       │                       │                    │
   │ 1. Search area        │                        │                       │                       │                    │
   │──────────────────────>│                        │                       │                       │                    │
   │                       │ GET /api/shipping/areas│                       │                       │                    │
   │                       │───────────────────────>│                       │                       │                    │
   │                       │                        │ SearchAreas()         │                       │                    │
   │                       │                        │──────────────────────>│                       │                    │
   │                       │                        │                       │ GET /v1/maps/areas    │                    │
   │                       │                        │                       │──────────────────────>│                    │
   │                       │                        │                       │<──────────────────────│                    │
   │                       │                        │<──────────────────────│                       │                    │
   │                       │<───────────────────────│                       │                       │                    │
   │<──────────────────────│                        │                       │                       │                    │
   │                       │                        │                       │                       │                    │
   │ 2. Select area        │                        │                       │                       │                    │
   │──────────────────────>│                        │                       │                       │                    │
   │                       │                        │                       │                       │                    │
   │ 3. Get shipping rates │                        │                       │                       │                    │
   │──────────────────────>│                        │                       │                       │                    │
   │                       │ POST /api/shipping/rates                       │                       │                    │
   │                       │───────────────────────>│                       │                       │                    │
   │                       │                        │ GetShippingRates()    │                       │                    │
   │                       │                        │──────────────────────>│                       │                    │
   │                       │                        │                       │ POST /v1/rates/couriers                    │
   │                       │                        │                       │──────────────────────>│                    │
   │                       │                        │                       │<──────────────────────│                    │
   │                       │                        │<──────────────────────│                       │                    │
   │                       │<───────────────────────│                       │                       │                    │
   │<──────────────────────│                        │                       │                       │                    │
   │                       │                        │                       │                       │                    │
   │ 4. Submit checkout    │                        │                       │                       │                    │
   │──────────────────────>│                        │                       │                       │                    │
   │                       │ POST /api/checkout     │                       │                       │                    │
   │                       │───────────────────────>│                       │                       │                    │
   │                       │                        │ CreateOrder()         │                       │                    │
   │                       │                        │──────────────────────>│                       │                    │
   │                       │                        │                       │                       │ Save order         │
   │                       │                        │                       │                       │───────────────────>│
   │                       │                        │                       │ POST /v1/draft_orders │                    │
   │                       │                        │                       │──────────────────────>│                    │
   │                       │                        │                       │<──────────────────────│                    │
   │                       │                        │                       │                       │ Save draft_order_id│
   │                       │                        │                       │                       │───────────────────>│
   │                       │                        │<──────────────────────│                       │                    │
   │                       │<───────────────────────│                       │                       │                    │
   │<──────────────────────│                        │                       │                       │                    │
   │                       │                        │                       │                       │                    │
   │ [Payment via Midtrans - existing flow]         │                       │                       │                    │
   │                       │                        │                       │                       │                    │
   │ 5. Payment confirmed (webhook)                 │                       │                       │                    │
   │                       │                        │ HandlePaymentCallback │                       │                    │
   │                       │                        │──────────────────────>│                       │                    │
   │                       │                        │                       │ ConfirmDraftOrder()   │                    │
   │                       │                        │                       │──────────────────────>│                    │
   │                       │                        │                       │ POST /v1/draft_orders/:id/confirm          │
   │                       │                        │                       │──────────────────────>│                    │
   │                       │                        │                       │<──────────────────────│                    │
   │                       │                        │                       │                       │ Save waybill_id,   │
   │                       │                        │                       │                       │ tracking_id        │
   │                       │                        │                       │                       │───────────────────>│
   │                       │                        │<──────────────────────│                       │                    │
```

### Tracking Flow

```
Customer/Job            ShippingService         BiteshipClient          Database
   │                          │                       │                    │
   │ GetTracking()            │                       │                    │
   │─────────────────────────>│                       │                    │
   │                          │ Get shipment          │                    │
   │                          │───────────────────────────────────────────>│
   │                          │<───────────────────────────────────────────│
   │                          │                       │                    │
   │                          │ GET /v1/trackings/:id │                    │
   │                          │──────────────────────>│                    │
   │                          │<──────────────────────│                    │
   │                          │                       │                    │
   │                          │ Map status            │                    │
   │                          │ Update shipment       │                    │
   │                          │───────────────────────────────────────────>│
   │                          │                       │                    │
   │                          │ Add tracking events   │                    │
   │                          │───────────────────────────────────────────>│
   │<─────────────────────────│                       │                    │
```

## Database Interaction Overview

### Tables and Ownership

| Table | Owner | Changes |
|-------|-------|---------|
| `biteship_locations` | ShippingRepository | NEW - stores Biteship saved locations |
| `shipping_snapshots` | ShippingRepository | MODIFIED - rename rajaongkir_raw_json to biteship_raw_json, add area_id columns |
| `shipments` | ShippingRepository | MODIFIED - add biteship_draft_order_id, biteship_order_id, biteship_tracking_id, biteship_waybill_id |
| `shipment_tracking_history` | ShippingRepository | UNCHANGED |
| `user_addresses` | ShippingRepository | MODIFIED - add area_id column |
| `shipping_providers` | ShippingRepository | UNCHANGED |

## Error Handling

### Error Types

```go
var (
    // Authentication errors
    ErrBiteshipUnauthorized = errors.New("biteship: unauthorized - check TOKEN_BITESHIP")
    
    // Rate limiting
    ErrBiteshipRateLimited = errors.New("biteship: rate limited - retry later")
    
    // API errors
    ErrBiteshipAPIFailed = errors.New("biteship: API request failed")
    ErrBiteshipInvalidRequest = errors.New("biteship: invalid request parameters")
    
    // Business errors
    ErrNoShippingRates = errors.New("no shipping rates available for this route")
    ErrAreaNotFound = errors.New("area not found")
    ErrDraftOrderFailed = errors.New("failed to create draft order")
    ErrOrderConfirmationFailed = errors.New("failed to confirm order")
    ErrTrackingNotFound = errors.New("tracking information not found")
)
```

### Retry Strategy

```go
type RetryConfig struct {
    MaxRetries     int           // 3
    InitialBackoff time.Duration // 1 second
    MaxBackoff     time.Duration // 30 seconds
    BackoffFactor  float64       // 2.0
}

// Retry on:
// - 429 Rate Limited
// - 5xx Server Errors
// - Network timeouts

// Do NOT retry on:
// - 400 Bad Request
// - 401 Unauthorized
// - 404 Not Found
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Authorization Header Format
*For any* Biteship API request, the Authorization header SHALL have the format "Bearer {TOKEN_BITESHIP}" where TOKEN_BITESHIP is the configured environment variable value.
**Validates: Requirements 2.2**

### Property 2: Area Search Response Parsing
*For any* valid Biteship area search response, parsing SHALL extract area_id, name, and postal_code fields without data loss.
**Validates: Requirements 2.3, 3.2**

### Property 3: Rate Response Parsing
*For any* valid Biteship rates response, parsing SHALL extract courier_code, courier_service_code, price, and duration fields without data loss.
**Validates: Requirements 2.6, 4.2**

### Property 4: Rate Categorization
*For any* shipping rate with service_type, the rate SHALL be categorized as Express (same_day, instant), Regular (standard, express), or Economy (economy, cargo) based on service_type value.
**Validates: Requirements 4.3**

### Property 5: Draft Order ID Persistence
*For any* successful draft order creation, the returned draft_order_id SHALL be stored in the shipments table for the corresponding order.
**Validates: Requirements 5.2**

### Property 6: Order Confirmation Persistence
*For any* successful order confirmation, the returned waybill_id and tracking_id SHALL be stored in the shipments table.
**Validates: Requirements 5.4**

### Property 7: Tracking Status Mapping
*For any* Biteship tracking status, the status SHALL be mapped to a valid ShipmentStatus enum value according to the defined mapping table.
**Validates: Requirements 6.3**

### Property 8: Delivered Status Propagation
*For any* shipment where tracking shows DELIVERED status, the corresponding order status SHALL be updated to DELIVERED.
**Validates: Requirements 6.4**

### Property 9: Biteship JSON Round-Trip
*For any* valid BiteshipRate struct, serializing to JSON and deserializing back SHALL produce an equivalent struct with all fields preserved.
**Validates: Requirements 8.4**

### Property 10: Error Status Code Mapping
*For any* Biteship API response with status code 401, the client SHALL return ErrBiteshipUnauthorized; for 429, SHALL return ErrBiteshipRateLimited.
**Validates: Requirements 9.1, 9.2**

### Property 11: Retry Behavior
*For any* Biteship API call that returns 5xx status, the client SHALL retry up to 3 times with exponential backoff before returning an error.
**Validates: Requirements 9.3**

### Property 12: Area ID Usage in Rates
*For any* shipping rate request, the request SHALL use area_id for both origin and destination parameters, not province/city/district IDs.
**Validates: Requirements 3.4, 4.1**

## Testing Strategy

### Dual Testing Approach

This implementation uses both unit tests and property-based tests:

- **Unit tests**: Verify specific examples, edge cases, and integration points
- **Property-based tests**: Verify universal properties that should hold across all inputs

### Property-Based Testing Library

**Library**: `github.com/leanovate/gopter` (Go property testing library)

**Configuration**: Each property test runs minimum 100 iterations.

### Test Categories

1. **BiteshipClient Tests**
   - Response parsing properties (areas, rates, tracking)
   - Error mapping properties
   - Authorization header property
   - Retry behavior property

2. **ShippingService Tests**
   - Rate categorization property
   - Status mapping property
   - Draft order flow properties
   - Tracking update properties

3. **Serialization Tests**
   - JSON round-trip property for all Biteship types

4. **Integration Tests**
   - End-to-end checkout flow with mocked Biteship API
   - Payment confirmation to order confirmation flow

### Test Annotation Format

Each property-based test SHALL be annotated with:
```go
// **Feature: biteship-migration, Property {number}: {property_text}**
// **Validates: Requirements X.Y**
```

---

## Design Verdict

**DESIGN IS COMPLETE — READY TO MOVE TO TASK BREAKDOWN**

All design aspects have been addressed:
- ✅ Service boundaries and responsibilities defined
- ✅ Data flow between components documented
- ✅ Shipment lifecycle state machine preserved (existing is sufficient)
- ✅ Error propagation strategy defined
- ✅ Retry & idempotency considerations included
- ✅ Separation between payment flow and shipment confirmation maintained
- ✅ Database schema changes specified
- ✅ API interaction flows documented
- ✅ Correctness properties defined
- ✅ Testing strategy specified
