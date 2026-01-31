# ZAVERA Shipping API Documentation

## Overview

The shipping system integrates with Kommerce APIs (RajaOngkir-like) to provide real-time shipping rates and tracking.

## Environment Variables

```env
KOMMERCE_COST_BASE_URL=https://rajaongkir.komerce.id/api/v1
KOMMERCE_DELIVERY_BASE_URL=https://api.collaborator.komerce.id
KOMMERCE_COST_API_KEY=your_cost_api_key
KOMMERCE_DELIVERY_API_KEY=your_delivery_api_key
ENABLE_TRACKING_JOB=true
```

## Database Migration

Run the shipping migration:
```bash
migrate_shipping.bat
```

---

## API Endpoints

### Location APIs

#### Get Provinces
```
GET /api/shipping/provinces
```

Response:
```json
{
  "provinces": [
    { "province_id": "1", "province_name": "Bali" },
    { "province_id": "2", "province_name": "Bangka Belitung" }
  ]
}
```

#### Get Cities
```
GET /api/shipping/cities?province_id=6
```

Response:
```json
{
  "cities": [
    {
      "city_id": "151",
      "city_name": "Jakarta Barat",
      "province_id": "6",
      "province_name": "DKI Jakarta",
      "type": "Kota",
      "postal_code": "11220"
    }
  ]
}
```

---

### Shipping Rates

#### Get Shipping Rates
```
POST /api/shipping/rates
Content-Type: application/json

{
  "origin_city_id": "152",
  "destination_city_id": "444",
  "weight": 1000,
  "courier": "jne:jnt:sicepat"
}
```

Response:
```json
{
  "origin": { "city_id": "152" },
  "destination": { "city_id": "444" },
  "weight": 1000,
  "rates": [
    {
      "provider_code": "jne",
      "provider_name": "JNE Express",
      "provider_logo": "https://...",
      "service_code": "REG",
      "service_name": "REG",
      "description": "Layanan Reguler",
      "cost": 18000,
      "etd": "2-3"
    }
  ]
}
```

#### Get Cart Shipping Preview
```
GET /api/shipping/preview?destination_city_id=444&courier=jne
X-Session-ID: your-session-id
```

Response:
```json
{
  "cart_subtotal": 599000,
  "total_weight": 1000,
  "rates": [...]
}
```

---

### Checkout with Shipping

#### Get Shipping Options for Cart
```
GET /api/checkout/shipping-options?destination_city_id=444
X-Session-ID: your-session-id
```

#### Checkout with Shipping Selection
```
POST /api/checkout/shipping
X-Session-ID: your-session-id
Content-Type: application/json

{
  "customer_name": "John Doe",
  "customer_email": "john@example.com",
  "customer_phone": "08123456789",
  "provider_code": "jne",
  "service_code": "REG",
  "shipping_address": {
    "recipient_name": "John Doe",
    "phone": "08123456789",
    "city_id": "444",
    "city_name": "Surabaya",
    "province_name": "Jawa Timur",
    "postal_code": "60111",
    "full_address": "Jl. Raya No. 123"
  }
}
```

Or with saved address:
```json
{
  "customer_name": "John Doe",
  "customer_email": "john@example.com",
  "customer_phone": "08123456789",
  "provider_code": "jne",
  "service_code": "REG",
  "address_id": 1
}
```

Response:
```json
{
  "order_id": 1,
  "order_code": "ZVR-20260109-ABC123",
  "subtotal": 599000,
  "shipping_cost": 18000,
  "total_amount": 617000,
  "status": "PENDING",
  "shipping_locked": true,
  "provider": "JNE Express",
  "service": "REG",
  "etd": "2-3",
  "shipping_address": {
    "recipient_name": "John Doe",
    "phone": "08123456789",
    "full_address": "Jl. Raya No. 123",
    "city_name": "Surabaya",
    "province_name": "Jawa Timur",
    "postal_code": "60111"
  }
}
```

---

### User Addresses (Protected)

#### Create Address
```
POST /api/user/addresses
Authorization: Bearer <token>
Content-Type: application/json

{
  "label": "Home",
  "recipient_name": "John Doe",
  "phone": "08123456789",
  "province_id": "11",
  "province_name": "Jawa Timur",
  "city_id": "444",
  "city_name": "Surabaya",
  "district": "Gubeng",
  "postal_code": "60111",
  "full_address": "Jl. Raya No. 123",
  "is_default": true
}
```

#### Get User Addresses
```
GET /api/user/addresses
Authorization: Bearer <token>
```

#### Update Address
```
PUT /api/user/addresses/:id
Authorization: Bearer <token>
```

#### Delete Address
```
DELETE /api/user/addresses/:id
Authorization: Bearer <token>
```

#### Set Default Address
```
POST /api/user/addresses/:id/default
Authorization: Bearer <token>
```

---

### Shipment & Tracking

#### Get Shipment Details
```
GET /api/shipments/:id
```

Response:
```json
{
  "id": 1,
  "order_id": 1,
  "order_code": "ZVR-20260109-ABC123",
  "provider_code": "jne",
  "provider_name": "JNE Express",
  "service_code": "REG",
  "service_name": "REG",
  "cost": 18000,
  "etd": "2-3",
  "weight": 1000,
  "tracking_number": "JNE123456789",
  "status": "IN_TRANSIT",
  "origin": "Jakarta Pusat",
  "destination": "Surabaya",
  "shipped_at": "2026-01-09 10:00:00",
  "tracking_history": [
    {
      "status": "Package received at origin",
      "description": "Package received at origin",
      "location": "Jakarta",
      "event_time": "2026-01-09 10:00:00"
    }
  ]
}
```

#### Refresh Tracking
```
POST /api/shipments/:id/refresh
```

---

### Admin Endpoints

#### Update Tracking Number
```
PUT /api/admin/shipments/:id/tracking
Content-Type: application/json

{
  "tracking_number": "JNE123456789"
}
```

#### Mark as Shipped
```
POST /api/admin/shipments/:id/ship
Content-Type: application/json

{
  "tracking_number": "JNE123456789"
}
```

#### Run Tracking Job Manually
```
POST /api/admin/shipping/tracking-job
```

---

## Checkout Flow

1. **Add items to cart**
2. **Get shipping options** - `GET /api/checkout/shipping-options?destination_city_id=xxx`
3. **Select shipping & checkout** - `POST /api/checkout/shipping`
4. **Initiate payment** - `POST /api/payments/initiate`
5. **Complete payment** (Midtrans handles this)
6. **Payment webhook** updates order to PAID, shipment to PROCESSING
7. **Admin ships order** - `POST /api/admin/shipments/:id/ship`
8. **Auto tracking** polls courier API and updates status
9. **Order delivered** - status updated automatically

## Shipment Status Flow

```
PENDING → PROCESSING → SHIPPED → IN_TRANSIT → OUT_FOR_DELIVERY → DELIVERED
                                                              ↘ RETURNED
                                                              ↘ FAILED
```

## Supported Couriers

- JNE (jne)
- J&T Express (jnt)
- SiCepat (sicepat)
- Pos Indonesia (pos)
- TIKI (tiki)
- Anteraja (anteraja)
- Ninja Express (ninja)
- Lion Parcel (lion)
- ID Express (ide)
- SAP Express (sap)
