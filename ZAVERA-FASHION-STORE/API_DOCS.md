# ZAVERA Backend API Documentation

## Base URL

```
http://localhost:8080/api
```

## Endpoints

### 1. Products

#### Get All Products

```
GET /api/products
```

**Response:**

```json
[
  {
    "id": 1,
    "name": "Classic White Tee",
    "slug": "classic-white-tee",
    "description": "Essential wardrobe staple...",
    "price": 250000,
    "stock": 50,
    "image_url": "https://images.unsplash.com/...",
    "images": ["https://...", "https://..."]
  }
]
```

#### Get Product by ID

```
GET /api/products/:id
```

---

### 2. Cart

#### Get Cart

```
GET /api/cart
```

_Automatically creates cart if doesn't exist. Uses session_id cookie._

**Response:**

```json
{
  "id": 1,
  "items": [
    {
      "id": 1,
      "product_id": 1,
      "product_name": "Classic White Tee",
      "product_image": "https://...",
      "quantity": 2,
      "price_per_unit": 250000,
      "subtotal": 500000,
      "stock": 48,
      "metadata": { "size": "M" }
    }
  ],
  "subtotal": 500000,
  "item_count": 2
}
```

#### Add to Cart

```
POST /api/cart/items
Content-Type: application/json

{
  "product_id": 1,
  "quantity": 2,
  "metadata": {"size": "M", "color": "White"}
}
```

#### Update Cart Item

```
PUT /api/cart/items/:id
Content-Type: application/json

{
  "quantity": 3
}
```

#### Remove from Cart

```
DELETE /api/cart/items/:id
```

#### Clear Cart

```
DELETE /api/cart
```

---

### 3. Checkout & Orders

#### Checkout

```
POST /api/checkout
Content-Type: application/json

{
  "customer_name": "John Doe",
  "customer_email": "john@example.com",
  "customer_phone": "08123456789",
  "notes": "Please wrap as gift"
}
```

**Response:**

```json
{
  "order": {
    "order_id": 1,
    "order_code": "WRR-1736179970",
    "total_amount": 515000,
    "status": "PENDING"
  },
  "snap_token": "abc123-midtrans-token"
}
```

#### Get Order

```
GET /api/orders/:code
```

**Response:**

```json
{
  "id": 1,
  "order_code": "WRR-1736179970",
  "customer_name": "John Doe",
  "customer_email": "john@example.com",
  "customer_phone": "08123456789",
  "subtotal": 500000,
  "shipping_cost": 15000,
  "tax": 0,
  "discount": 0,
  "total_amount": 515000,
  "status": "PENDING",
  "items": [
    {
      "product_id": 1,
      "product_name": "Classic White Tee",
      "quantity": 2,
      "price_per_unit": 250000,
      "subtotal": 500000
    }
  ],
  "created_at": "2026-01-05 21:32:50"
}
```

---

### 4. Payment Callback (Midtrans)

```
POST /api/payment/callback
Content-Type: application/json

{
  "order_id": "WRR-1736179970",
  "transaction_status": "settlement",
  "transaction_id": "ABC123"
}
```

## Order Status Flow

```
PENDING → PAID → PROCESSING → SHIPPED → DELIVERED
   ↓         ↓         ↓
FAILED   CANCELLED  CANCELLED
```

## Testing with cURL

### Test Product API

```bash
curl http://localhost:8080/api/products
```

### Test Add to Cart

```bash
curl -X POST http://localhost:8080/api/cart/items \
  -H "Content-Type: application/json" \
  -d '{"product_id":1,"quantity":2,"metadata":{"size":"M"}}'
```

### Test Checkout

```bash
curl -X POST http://localhost:8080/api/checkout \
  -H "Content-Type: application/json" \
  -d '{
    "customer_name":"Test User",
    "customer_email":"test@example.com",
    "customer_phone":"08123456789"
  }'
```

## Database Schema

### Tables

- `users` - User accounts
- `products` - Product catalog
- `product_images` - Product images (multiple per product)
- `carts` - Shopping carts (session or user-based)
- `cart_items` - Items in cart
- `orders` - Customer orders
- `order_items` - Items in order (price snapshot)
- `payments` - Payment transactions

### Key Features

- Cart supports both guest (session_id) and logged-in users (user_id)
- Price snapshot in cart_items and order_items (preserves price at time of action)
- Order status enum with proper lifecycle transitions
- Payment tracking with Midtrans integration
- JSONB metadata fields for extensibility
- Automatic `updated_at` triggers
- Full indexing on critical columns
