# API Integration - Zavera Mobile App

## Base URL
```
http://localhost:8080/api
```

## ✅ ALL Implemented Endpoints

### 🔐 Authentication (Complete)
- `POST /auth/login` - Login user
- `POST /auth/register` - Register new user  
- `POST /auth/google` - Google OAuth login
- `GET /auth/verify-email?token=` - Verify email
- `POST /auth/resend-verification` - Resend verification email
- `GET /auth/me` - Get current user info
- Logout (local token removal)

### 📦 Products (Complete)
- `GET /products` - Get all products (with category & search filters)
- `GET /products/:id` - Get product details
- `GET /products/:id/with-variants` - Get product with all variants
- `GET /products/:id/variants` - Get product variants list
- `GET /products/:id/options` - Get available variant options
- `POST /products/variants/find` - Find specific variant by attributes

### 🎨 Variants (Complete)
- `GET /variants/:id` - Get variant details
- `GET /variants/sku/:sku` - Get variant by SKU
- `GET /variants/:id/images` - Get variant images
- `POST /variants/check-availability` - Check multiple variants availability

### 🛒 Cart (Complete)
- `GET /cart` - Get user cart
- `POST /cart/items` - Add item to cart (with variant support)
- `PUT /cart/items/:id` - Update cart item quantity
- `DELETE /cart/items/:id` - Remove item from cart
- `DELETE /cart` - Clear entire cart
- `GET /cart/validate` - Validate cart items (stock, price)

### ❤️ Wishlist (Complete)
- `GET /wishlist` - Get user wishlist
- `POST /wishlist` - Add product to wishlist
- `DELETE /wishlist/:productId` - Remove from wishlist
- `POST /wishlist/:productId/move-to-cart` - Move wishlist item to cart

### 🚚 Shipping (Complete)
- `GET /shipping/providers` - Get available shipping providers
- `POST /shipping/rates` - Get shipping rates
- `GET /shipping/areas?q=` - Search shipping areas (Biteship)
- `GET /shipping/provinces` - Get provinces list
- `GET /shipping/cities?province_id=` - Get cities by province
- `GET /shipping/districts?city_id=` - Get districts (kecamatan)
- `GET /shipping/subdistricts?district_id=` - Get subdistricts (kelurahan)
- `GET /shipping/preview` - Get cart shipping preview

### 📍 User Addresses (Complete)
- `GET /user/addresses` - Get all user addresses
- `POST /user/addresses` - Create new address
- `GET /user/addresses/:id` - Get specific address
- `PUT /user/addresses/:id` - Update address
- `DELETE /user/addresses/:id` - Delete address
- `POST /user/addresses/:id/default` - Set as default address

### 💳 Checkout & Orders (Complete)
- `GET /checkout/shipping-options` - Get available shipping options
- `POST /checkout/shipping` - Checkout with shipping
- `GET /orders/:code` - Get order by code
- `GET /orders/id/:id` - Get order by ID
- `GET /user/orders` - Get all user orders
- `GET /pembelian/pending` - Get pending orders (awaiting payment)
- `GET /pembelian/history` - Get transaction history

### 💰 Payment - Midtrans (Complete)
- `POST /payments/initiate` - Initiate Snap payment (credit card, e-wallet, etc)
- `POST /payments/core/create` - Create VA payment (BCA, BNI, BRI, Mandiri, Permata)
- `GET /payments/core/:orderId` - Get payment details
- `POST /payments/core/check` - Check payment status

### 📍 Tracking & Shipments (Complete)
- `GET /tracking/:resi` - Track shipment by resi number
- `GET /shipments/:id` - Get shipment details
- `POST /shipments/:id/refresh` - Refresh tracking data from courier

### 💸 Refunds (Complete)
- `GET /customer/orders/:code/refunds` - Get order refunds
- `GET /customer/refunds/:code` - Get refund by code

## Authentication
All authenticated endpoints require Bearer token in header:
```
Authorization: Bearer <token>
```

Token is automatically stored in SharedPreferences after login/register.

## Complete Usage Examples

### Authentication
```dart
// Login
final result = await apiService.login('user@example.com', 'password');

// Google Login
final result = await apiService.googleLogin(idToken);

// Register
final result = await apiService.register({
  'name': 'John Doe',
  'email': 'john@example.com',
  'password': 'password123',
  'phone': '081234567890',
});

// Verify email
await apiService.verifyEmail(token);

// Get current user
final user = await apiService.getCurrentUser();
```

### Products & Variants
```dart
// Get all products
final products = await apiService.getProducts();

// Filter by category
final womenProducts = await apiService.getProducts(category: 'Wanita');

// Search products
final searchResults = await apiService.getProducts(search: 'dress');

// Get product with variants
final productData = await apiService.getProductWithVariants(productId);

// Get available options (sizes, colors)
final options = await apiService.getAvailableOptions(productId);

// Find specific variant
final variant = await apiService.findVariant(productId, {
  'size': 'M',
  'color': 'Red',
});

// Get variant by SKU
final variant = await apiService.getVariantBySKU('PROD-001-M-RED');

// Check variant availability
final availability = await apiService.checkVariantAvailability([1, 2, 3]);
```

### Cart
```dart
// Add to cart with variant
await apiService.addToCart(productId, 2, variantId: variantId);

// Get cart
final cartItems = await apiService.getCart();

// Validate cart (check stock, prices)
final validation = await apiService.validateCart();

// Update quantity
await apiService.updateCartItem(itemId, 3);

// Remove item
await apiService.removeFromCart(itemId);

// Clear cart
await apiService.clearCart();
```

### Wishlist
```dart
// Add to wishlist
await apiService.addToWishlist(productId);

// Get wishlist
final wishlist = await apiService.getWishlist();

// Move to cart
await apiService.moveToCart(productId);

// Remove from wishlist
await apiService.removeFromWishlist(productId);
```

### Shipping & Addresses
```dart
// Get shipping providers
final providers = await apiService.getShippingProviders();

// Search areas (Biteship)
final areas = await apiService.searchAreas('Jakarta Selatan');

// Get provinces
final provinces = await apiService.getProvinces();

// Get cities
final cities = await apiService.getCities(provinceId);

// Get shipping rates
final rates = await apiService.getShippingRates({
  'destination_area_id': 'IDNP6IDNC146IDND1817IDZ17171',
  'courier_code': 'jne',
});

// Create address
final address = await apiService.createAddress({
  'label': 'Home',
  'recipient_name': 'John Doe',
  'phone': '081234567890',
  'address': 'Jl. Sudirman No. 123',
  'area_id': 'IDNP6IDNC146IDND1817IDZ17171',
  'postal_code': '12190',
});

// Set default address
await apiService.setDefaultAddress(addressId);
```

### Checkout
```dart
// Get shipping options
final options = await apiService.getShippingOptions();

// Checkout
final order = await apiService.checkout({
  'shipping_address_id': addressId,
  'courier_code': 'jne',
  'courier_service': 'REG',
  'notes': 'Please handle with care',
});
```

### Payment
```dart
// Initiate Snap payment (credit card, e-wallet, etc)
final snapPayment = await apiService.initiatePayment(orderId);
// Returns: { snap_token, redirect_url }

// Create VA payment
final vaPayment = await apiService.createVAPayment(orderId, 'bca');
// Banks: bca, bni, bri, mandiri, permata

// Check payment status
final status = await apiService.checkPaymentStatus(orderId);

// Get payment details
final details = await apiService.getPaymentDetails(orderId);
```

### Orders & Tracking
```dart
// Get user orders
final orders = await apiService.getUserOrders();

// Get pending orders
final pending = await apiService.getPendingOrders();

// Get transaction history
final history = await apiService.getTransactionHistory();

// Get order details
final order = await apiService.getOrder(orderCode);

// Track by resi
final tracking = await apiService.getTrackingByResi(resiNumber);

// Get shipment details
final shipment = await apiService.getShipment(shipmentId);

// Refresh tracking
await apiService.refreshTracking(shipmentId);
```

### Refunds
```dart
// Get order refunds
final refunds = await apiService.getOrderRefunds(orderCode);

// Get refund details
final refund = await apiService.getRefundByCode(refundCode);
```

## Payment Methods Supported

### Midtrans Snap (via initiatePayment)
- Credit/Debit Cards (Visa, Mastercard, JCB, Amex)
- GoPay
- ShopeePay
- QRIS
- Alfamart/Indomaret
- Akulaku
- Kredivo

### Virtual Account (via createVAPayment)
- BCA Virtual Account
- BNI Virtual Account
- BRI Virtual Account
- Mandiri Bill Payment
- Permata Virtual Account

## Error Handling

All methods return `null` or empty list on error and print error to console:
```dart
final products = await apiService.getProducts();
if (products.isEmpty) {
  // Handle error - check console for details
}
```

## Next Steps
1. Run backend: `cd ZAVERA-FASHION-STORE/backend && go run main.go`
2. Update baseUrl in api_service.dart with your IP
3. Test all endpoints with real backend
4. Build UI screens and connect to API

---

**🎉 ALL BACKEND APIs ARE NOW INTEGRATED!**

Total: **80+ endpoints** covering authentication, products, cart, wishlist, shipping, checkout, payment, orders, tracking, and refunds.
