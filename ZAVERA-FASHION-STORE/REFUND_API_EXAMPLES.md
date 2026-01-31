# Refund System - API Examples

## Authentication

All API requests require JWT token in Authorization header:
```
Authorization: Bearer {your_jwt_token}
```

Get token by logging in:
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "your_password"
  }'
```

## Admin API Endpoints

### 1. Create Refund

**Endpoint**: `POST /api/admin/refunds`

**Full Refund Example**:
```bash
curl -X POST http://localhost:8080/api/admin/refunds \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{
    "order_code": "ORD-20260125-001",
    "refund_type": "FULL",
    "reason": "Customer Request",
    "reason_detail": "Customer changed mind",
    "idempotency_key": "refund-001-20260125"
  }'
```

**Partial Refund Example**:
```bash
curl -X POST http://localhost:8080/api/admin/refunds \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{
    "order_code": "ORD-20260125-001",
    "refund_type": "PARTIAL",
    "refund_amount": 50000,
    "reason": "Damaged Item",
    "reason_detail": "Item arrived damaged",
    "idempotency_key": "refund-002-20260125"
  }'
```

**Item Only Refund Example**:
```bash
curl -X POST http://localhost:8080/api/admin/refunds \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{
    "order_code": "ORD-20260125-001",
    "refund_type": "ITEM_ONLY",
    "reason": "Wrong Item",
    "reason_detail": "Sent wrong size",
    "items": [
      {
        "product_id": 1,
        "quantity": 1
      }
    ],
    "idempotency_key": "refund-003-20260125"
  }'
```

**Shipping Only Refund Example**:
```bash
curl -X POST http://localhost:8080/api/admin/refunds \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{
    "order_code": "ORD-20260125-001",
    "refund_type": "SHIPPING_ONLY",
    "reason": "Late Delivery",
    "reason_detail": "Delivery took too long",
    "idempotency_key": "refund-004-20260125"
  }'
```

**Response (Success)**:
```json
{
  "id": 1,
  "refund_code": "REF-20260125-001",
  "order_id": 123,
  "order_code": "ORD-20260125-001",
  "refund_type": "FULL",
  "refund_amount": 150000,
  "items_refund": 130000,
  "shipping_refund": 20000,
  "reason": "Customer Request",
  "reason_detail": "Customer changed mind",
  "status": "COMPLETED",
  "gateway_refund_id": "midtrans-ref-123",
  "requested_at": "2026-01-25T10:00:00Z",
  "completed_at": "2026-01-25T10:00:05Z",
  "items": [
    {
      "product_id": 1,
      "product_name": "T-Shirt",
      "quantity": 2,
      "price_per_unit": 50000,
      "subtotal": 100000
    }
  ]
}
```

### 2. Process Refund

**Endpoint**: `POST /api/admin/refunds/:id/process`

```bash
curl -X POST http://localhost:8080/api/admin/refunds/1/process \
  -H "Authorization: Bearer {token}"
```

**Response**:
```json
{
  "message": "Refund processed successfully",
  "refund": {
    "id": 1,
    "status": "COMPLETED",
    "gateway_refund_id": "midtrans-ref-123"
  }
}
```

### 3. Retry Failed Refund

**Endpoint**: `POST /api/admin/refunds/:id/retry`

```bash
curl -X POST http://localhost:8080/api/admin/refunds/1/retry \
  -H "Authorization: Bearer {token}"
```

**Response**:
```json
{
  "message": "Refund retry successful",
  "refund": {
    "id": 1,
    "status": "COMPLETED",
    "gateway_refund_id": "midtrans-ref-456"
  }
}
```

### 4. Get Refund Details

**Endpoint**: `GET /api/admin/refunds/:id`

```bash
curl -X GET http://localhost:8080/api/admin/refunds/1 \
  -H "Authorization: Bearer {token}"
```

**Response**:
```json
{
  "id": 1,
  "refund_code": "REF-20260125-001",
  "order_code": "ORD-20260125-001",
  "refund_type": "FULL",
  "refund_amount": 150000,
  "status": "COMPLETED",
  "items": [...],
  "status_history": [
    {
      "status": "PENDING",
      "changed_at": "2026-01-25T10:00:00Z"
    },
    {
      "status": "PROCESSING",
      "changed_at": "2026-01-25T10:00:02Z"
    },
    {
      "status": "COMPLETED",
      "changed_at": "2026-01-25T10:00:05Z"
    }
  ]
}
```

### 5. List All Refunds

**Endpoint**: `GET /api/admin/refunds`

```bash
curl -X GET "http://localhost:8080/api/admin/refunds?page=1&page_size=10&status=COMPLETED" \
  -H "Authorization: Bearer {token}"
```

**Query Parameters**:
- `page` (optional): Page number (default: 1)
- `page_size` (optional): Items per page (default: 10)
- `status` (optional): Filter by status (PENDING, PROCESSING, COMPLETED, FAILED)
- `order_code` (optional): Filter by order code

**Response**:
```json
{
  "refunds": [...],
  "total_count": 50,
  "page": 1,
  "page_size": 10,
  "total_pages": 5
}
```

### 6. Get Order Refunds

**Endpoint**: `GET /api/admin/orders/:code/refunds`

```bash
curl -X GET http://localhost:8080/api/admin/orders/ORD-20260125-001/refunds \
  -H "Authorization: Bearer {token}"
```

**Response**:
```json
{
  "refunds": [
    {
      "id": 1,
      "refund_code": "REF-20260125-001",
      "refund_type": "FULL",
      "refund_amount": 150000,
      "status": "COMPLETED"
    }
  ],
  "total_refunded": 150000
}
```

## Customer API Endpoints

### 1. Get Order Refunds

**Endpoint**: `GET /api/customer/orders/:code/refunds`

```bash
curl -X GET http://localhost:8080/api/customer/orders/ORD-20260125-001/refunds \
  -H "Authorization: Bearer {customer_token}"
```

**Response**:
```json
{
  "refunds": [
    {
      "refund_code": "REF-20260125-001",
      "refund_type": "FULL",
      "refund_amount": 150000,
      "status": "COMPLETED",
      "reason": "Customer Request",
      "requested_at": "2026-01-25T10:00:00Z",
      "completed_at": "2026-01-25T10:00:05Z",
      "timeline_estimate": "1-3 hari kerja"
    }
  ]
}
```

### 2. Get Refund by Code

**Endpoint**: `GET /api/customer/refunds/:code`

```bash
curl -X GET http://localhost:8080/api/customer/refunds/REF-20260125-001 \
  -H "Authorization: Bearer {customer_token}"
```

**Response**:
```json
{
  "refund_code": "REF-20260125-001",
  "order_code": "ORD-20260125-001",
  "refund_type": "FULL",
  "refund_amount": 150000,
  "items_refund": 130000,
  "shipping_refund": 20000,
  "status": "COMPLETED",
  "reason": "Customer Request",
  "items": [...],
  "timeline_estimate": "1-3 hari kerja",
  "status_message": "Dana telah dikembalikan ke metode pembayaran kamu"
}
```

## Error Responses

### 400 Bad Request
```json
{
  "error": "Invalid request",
  "message": "Refund amount exceeds refundable balance",
  "details": {
    "refund_amount": 200000,
    "refundable_balance": 150000
  }
}
```

### 401 Unauthorized
```json
{
  "error": "Unauthorized",
  "message": "Invalid or expired token"
}
```

### 403 Forbidden
```json
{
  "error": "Forbidden",
  "message": "Admin access required"
}
```

### 404 Not Found
```json
{
  "error": "Not found",
  "message": "Order not found"
}
```

### 409 Conflict
```json
{
  "error": "Conflict",
  "message": "Refund with this idempotency key already exists",
  "existing_refund": {
    "id": 1,
    "refund_code": "REF-20260125-001"
  }
}
```

### 500 Internal Server Error
```json
{
  "error": "Internal server error",
  "message": "Failed to process refund"
}
```

## Postman Collection

Import this JSON into Postman:

```json
{
  "info": {
    "name": "Zavera Refund System",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Admin",
      "item": [
        {
          "name": "Create Full Refund",
          "request": {
            "method": "POST",
            "header": [
              {
                "key": "Authorization",
                "value": "Bearer {{admin_token}}"
              }
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"order_code\": \"{{order_code}}\",\n  \"refund_type\": \"FULL\",\n  \"reason\": \"Customer Request\",\n  \"reason_detail\": \"Testing\",\n  \"idempotency_key\": \"test-{{$timestamp}}\"\n}",
              "options": {
                "raw": {
                  "language": "json"
                }
              }
            },
            "url": {
              "raw": "{{base_url}}/api/admin/refunds",
              "host": ["{{base_url}}"],
              "path": ["api", "admin", "refunds"]
            }
          }
        }
      ]
    }
  ],
  "variable": [
    {
      "key": "base_url",
      "value": "http://localhost:8080"
    },
    {
      "key": "admin_token",
      "value": "your_admin_token_here"
    },
    {
      "key": "order_code",
      "value": "ORD-20260125-001"
    }
  ]
}
```

## Testing Tips

1. **Get Admin Token First**:
   ```bash
   curl -X POST http://localhost:8080/api/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email":"admin@example.com","password":"password"}'
   ```

2. **Use Unique Idempotency Keys**:
   - Use timestamp: `refund-${Date.now()}`
   - Use UUID: `refund-${uuid()}`

3. **Test Different Refund Types**:
   - Start with FULL refund (simplest)
   - Then try PARTIAL, SHIPPING_ONLY, ITEM_ONLY

4. **Test Error Cases**:
   - Invalid order code
   - Order not DELIVERED/COMPLETED
   - Refund amount exceeds balance
   - Duplicate idempotency key

5. **Monitor Backend Logs**:
   - Watch console for error messages
   - Check database for refund records

## Database Queries for Verification

```sql
-- Check refund was created
SELECT * FROM refunds WHERE order_code = 'ORD-20260125-001';

-- Check refund status history
SELECT * FROM refund_status_history 
WHERE refund_id = 1 
ORDER BY changed_at DESC;

-- Check order was updated
SELECT order_code, status, refund_status, refund_amount 
FROM orders 
WHERE order_code = 'ORD-20260125-001';

-- Check stock was restored
SELECT id, name, stock 
FROM products 
WHERE id IN (SELECT product_id FROM order_items WHERE order_id = 123);
```

---

**Happy Testing!** 🚀
