# ZAVERA E-Commerce - Production-Ready Backend

## ✅ BACKEND ARCHITECTURE COMPLETED

### 📁 Folder Structure

```
backend/
├── config/
│   └── database.go          # PostgreSQL connection with database/sql
├── models/
│   └── models.go            # Domain models (User, Product, Cart, Order, Payment)
├── dto/
│   └── dto.go               # Request/Response DTOs
├── repository/
│   ├── product_repository.go    # Product data access
│   ├── cart_repository.go       # Cart data access
│   ├── order_repository.go      # Order data access
│   └── payment_repository.go    # Payment data access
├── service/
│   ├── product_service.go       # Product business logic
│   ├── cart_service.go          # Cart business logic (stock validation)
│   ├── order_service.go         # Order lifecycle management
│   └── payment_service.go       # Midtrans integration
├── handler/
│   ├── product_handler.go       # Product HTTP handlers
│   ├── cart_handler.go          # Cart HTTP handlers
│   └── order_handler.go         # Order & Payment HTTP handlers
├── routes/
│   └── routes.go            # Route setup with dependency injection
└── main.go                  # Application entry point
```

### 🗄️ Database Schema

**Tables:** users, products, product_images, carts, cart_items, orders, order_items, payments

**Key Features:**

- Order status enum: PENDING → PAID → PROCESSING → SHIPPED → DELIVERED
- Payment status enum: PENDING → SUCCESS | FAILED | EXPIRED
- Cart supports guest (session_id) and logged-in users (user_id)
- Price snapshot in cart_items and order_items
- JSONB metadata for extensibility
- Full indexing on critical columns

### 🚀 REST API Endpoints

#### Products

- `GET /api/products` - Get all products
- `GET /api/products/:id` - Get product by ID

#### Cart

- `GET /api/cart` - Get cart (auto-creates with session)
- `POST /api/cart/items` - Add to cart
- `PUT /api/cart/items/:id` - Update quantity
- `DELETE /api/cart/items/:id` - Remove item
- `DELETE /api/cart` - Clear cart

#### Orders

- `POST /api/checkout` - Create order & get payment token
- `GET /api/orders/:code` - Get order details

#### Payments

- `POST /api/payment/callback` - Midtrans webhook

### 🔐 Data Flow

1. **Add to Cart**

   - Client → Handler validates request
   - Service checks product stock
   - Repository creates/updates cart_item with price snapshot
   - Returns updated cart

2. **Checkout**

   - Handler gets cart from session
   - Service validates all items have stock
   - Repository creates order (PENDING) + order_items
   - Service deducts stock
   - Payment service creates Midtrans snap token
   - Returns order + snap_token

3. **Payment Callback**
   - Midtrans → Handler receives notification
   - Service updates payment status
   - If SUCCESS: Order status PENDING → PAID
   - If FAILED: Order status → FAILED

### ✅ Production Best Practices

- **Separation of Concerns**: Handler → Service → Repository pattern
- **Single Source of Truth**: Cart and pricing always in backend database
- **Stock Management**: Stock check before cart add, stock deduction on checkout
- **Price Integrity**: Price snapshot prevents price change exploits
- **Order Lifecycle**: Proper status transitions with validation
- **Session Management**: UUID session cookies for guest carts
- **Error Handling**: Proper HTTP status codes and error messages
- **CORS**: Configured for localhost:3000 with credentials

### 📝 How to Run

1. **Start Database:**

   ```bash
   # Already running with schema applied
   ```

2. **Start Backend:**

   ```bash
   cd backend
   go run main.go
   # Server on http://localhost:8080
   ```

3. **Start Frontend:**
   ```bash
   cd frontend
   npm run dev
   # App on http://localhost:3000
   ```

Or use batch files:

- `start-backend.bat` - Starts Go backend
- `cd frontend && npm run dev` - Starts Next.js frontend

### 🧪 Testing

See [API_DOCS.md](./API_DOCS.md) for full API documentation and cURL examples.

Quick test:

```bash
curl http://localhost:8080/api/products
```

### 🎯 Order Status Lifecycle

```
User adds to cart
     ↓
User clicks checkout → Order created (PENDING)
     ↓
Midtrans payment page shown
     ↓
User pays → Callback received → Order (PAID)
     ↓
Admin processes → Order (PROCESSING)
     ↓
Admin ships → Order (SHIPPED)
     ↓
Customer receives → Order (DELIVERED)

Failure paths:
- Payment fails → Order (FAILED)
- User cancels → Order (CANCELLED)
- Admin cancels → Order (CANCELLED)
```

### 📦 Dependencies

```
- gin-gonic/gin - HTTP framework
- lib/pq - PostgreSQL driver
- google/uuid - Session ID generation
- midtrans-go - Payment gateway
- gin-contrib/cors - CORS middleware
- joho/godotenv - Environment variables
```

### 🔧 Configuration

`.env` file:

```
DB_HOST=localhost
DB_PORT=5432
DB_NAME=zavera_db
DB_USER=postgres
DB_PASSWORD=Yan2692009

MIDTRANS_SERVER_KEY=your-server-key
MIDTRANS_CLIENT_KEY=your-client-key
MIDTRANS_ENV=sandbox
```

---

## 🎨 Frontend Integration (Next Steps)

Frontend needs to be updated to use new backend cart API instead of localStorage:

### Required Changes in `frontend/src/context/CartContext.tsx`:

1. Replace localStorage with API calls:

   - `addToCart` → `POST /api/cart/items`
   - `removeFromCart` → `DELETE /api/cart/items/:id`
   - `updateQuantity` → `PUT /api/cart/items/:id`
   - Load cart on mount → `GET /api/cart`

2. Handle session cookies (browser handles automatically)

3. Update checkout flow to use `/api/checkout` endpoint

### Example API Integration:

```typescript
const addToCart = async (product: Product, quantity: number, size: string) => {
  const response = await axios.post(
    "http://localhost:8080/api/cart/items",
    {
      product_id: product.id,
      quantity,
      metadata: { size },
    },
    { withCredentials: true }
  );

  setCart(response.data);
};
```

---

**Backend is now production-ready with proper architecture, data integrity, and order management! 🚀**
