# 🎯 Website Logic Reference - WAJIB IKUTI!

## ⚠️ CRITICAL: Logic yang HARUS sama dengan website

Ini adalah logic dari website yang **WAJIB** diikuti di mobile app. Jangan sampai beda!

---

## 1. 🔐 AUTHENTICATION LOGIC

### Login Flow
```typescript
// Website: AuthContext.tsx
const login = async (email: string, password: string) => {
  const response = await api.post("/auth/login", { email, password });
  const { access_token, user: userData } = response.data;
  
  // 1. Save token to localStorage
  localStorage.setItem("auth_token", access_token);
  
  // 2. Set user state
  setUser(userData);
  
  // 3. IMPORTANT: Refresh cart after login (100ms delay)
  setTimeout(() => triggerCartRefresh(), 100);
  
  return userData;
};
```

**Mobile app HARUS:**
- Save token ke SharedPreferences
- Set user state
- **REFRESH CART setelah login** (delay 100ms)
- Return user data untuk redirect logic

### Google Login Flow
```typescript
const loginWithGoogle = async (idToken: string) => {
  const response = await api.post("/auth/google", { id_token: idToken });
  const { access_token, user: userData } = response.data;
  
  localStorage.setItem("auth_token", access_token);
  setUser(userData);
  
  // IMPORTANT: Refresh cart after login
  setTimeout(() => triggerCartRefresh(), 100);
};
```

### Logout Flow
```typescript
const logout = () => {
  // 1. Remove token
  localStorage.removeItem("auth_token");
  
  // 2. IMPORTANT: Clear cart on logout
  localStorage.removeItem("zavera_cart");
  
  // 3. Clear user state
  setUser(null);
  
  // 4. Trigger cart refresh to clear cart state
  setTimeout(() => triggerCartRefresh(), 100);
};
```

**Mobile app HARUS:**
- Remove token
- **Clear cart data**
- Clear user state
- **Refresh cart state** (delay 100ms)

### Auto-Logout on 401
```typescript
// Website: api.ts interceptor
api.interceptors.response.use(
  (response) => response,
  (error) => {
    const errorCode = error.response?.data?.error;
    
    // If session expired or unauthorized
    if (errorCode === "session_expired" || 
        (error.response?.status === 401 && errorCode === "unauthorized")) {
      // Clear auth data
      localStorage.removeItem("auth_token");
      localStorage.removeItem("user");
      
      // Redirect to login
      window.location.href = "/login?session_expired=true";
    }
    return Promise.reject(error);
  }
);
```

**Mobile app HARUS:**
- Intercept 401 responses
- Check error code: `session_expired` atau `unauthorized`
- Clear token & user data
- Navigate to login screen dengan parameter `session_expired=true`

---

## 2. 🛒 CART LOGIC (PALING PENTING!)

### Cart Initialization
```typescript
// Website: CartContext.tsx
useEffect(() => {
  const initCart = async () => {
    setIsLoading(true);
    
    // 1. Try to load from backend FIRST
    const loadedFromBackend = await loadCartFromBackend();
    
    // 2. Fallback to localStorage only if backend fails
    if (!loadedFromBackend) {
      const savedCart = localStorage.getItem("zavera_cart");
      if (savedCart) {
        const parsed = JSON.parse(savedCart);
        setCart(validCart);
      }
    }
    
    setIsLoading(false);
  };
  
  initCart();
}, []);
```

**Mobile app HARUS:**
1. Load dari backend API DULU
2. Fallback ke local storage HANYA jika backend gagal
3. Validate cart items (check price, quantity valid)

### Load Cart from Backend
```typescript
const loadCartFromBackend = async () => {
  const token = localStorage.getItem("auth_token");
  
  // IMPORTANT: If no token, clear cart!
  if (!token) {
    setCart([]);
    localStorage.removeItem("zavera_cart");
    return true; // Prevent localStorage fallback
  }
  
  try {
    const response = await api.get("/cart");
    
    if (response.data && response.data.items) {
      const backendItems = response.data.items.map(convertBackendItem);
      setCart(backendItems);
      // Save to localStorage as backup
      localStorage.setItem("zavera_cart", JSON.stringify(backendItems));
      return true;
    }
    
    // Empty cart
    setCart([]);
    localStorage.removeItem("zavera_cart");
    return true;
  } catch (error) {
    // If unauthorized, clear cart
    setCart([]);
    localStorage.removeItem("zavera_cart");
    return true;
  }
};
```

**Mobile app HARUS:**
- Check token DULU
- **Jika tidak ada token → CLEAR CART** (jangan load dari local)
- Load dari `/cart` endpoint
- Convert backend items ke frontend format
- Save ke local storage sebagai backup
- Handle error dengan clear cart

### Backend Cart Item Format
```typescript
interface BackendCartItem {
  id: number;                    // Cart item ID (bukan product ID!)
  product_id: number;
  product_name: string;
  product_image: string;
  quantity: number;
  price_per_unit: number;
  subtotal: number;
  stock: number;
  metadata?: {
    selected_size?: string;
  };
}

interface BackendCartResponse {
  id: number;                    // Cart ID
  items: BackendCartItem[];
  subtotal: number;
  item_count: number;
}
```

**PENTING:** Backend return `id` untuk cart item (bukan product ID). Simpan ini sebagai `cartItemId` untuk update/delete!

### Convert Backend to Frontend
```typescript
const convertBackendItem = (item: BackendCartItem): CartItem => ({
  id: item.product_id,           // Product ID
  name: item.product_name,
  price: item.price_per_unit,
  image_url: item.product_image,
  quantity: item.quantity,
  selectedSize: item.metadata?.selected_size || "M",
  stock: item.stock,
  cartItemId: item.id,           // IMPORTANT: Save backend cart item ID!
});
```

### Add to Cart Logic
```typescript
const addToCart = async (item: CartItem) => {
  // 1. Check if user is logged in
  const token = localStorage.getItem("auth_token");
  if (!token) {
    console.log("User must be logged in to add items to cart");
    return; // Don't add if not logged in!
  }
  
  // 2. Calculate NEW TOTAL quantity (existing + new)
  const existingItem = cart.find(
    (i) => i.id === item.id && i.selectedSize === item.selectedSize
  );
  const newTotalQuantity = existingItem 
    ? existingItem.quantity + item.quantity 
    : item.quantity;
  
  // 3. Optimistically update local state
  setCart((prev) => {
    const existing = prev.find(
      (i) => i.id === item.id && i.selectedSize === item.selectedSize
    );
    if (existing) {
      return prev.map((i) =>
        i.id === item.id && i.selectedSize === item.selectedSize
          ? { ...i, quantity: i.quantity + item.quantity }
          : i
      );
    }
    return [...prev, item];
  });
  
  // 4. Sync to backend - send TOTAL quantity (backend will SET, not ADD)
  try {
    const payload = {
      product_id: item.id,
      quantity: newTotalQuantity,  // IMPORTANT: Send TOTAL, not increment!
      metadata: {
        selected_size: item.selectedSize || "M",
      },
      variant_id: item.variant_id, // If variant product
    };
    
    const response = await api.post("/cart/items", payload);
    
    // 5. Update cart with backend response (source of truth)
    if (response.data && response.data.items) {
      const backendItems = response.data.items.map(convertBackendItem);
      setCart(backendItems);
    }
  } catch (error) {
    console.error("Failed to add to backend cart:", error);
    // Keep local state as fallback
  }
};
```

**Mobile app HARUS:**
1. **Check login DULU** - jangan add jika belum login!
2. Calculate TOTAL quantity (existing + new)
3. Update local state optimistically
4. **Send TOTAL quantity ke backend** (bukan increment!)
5. Update cart dengan response backend

### Remove from Cart Logic
```typescript
const removeFromCart = async (id: number, selectedSize?: string) => {
  // 1. Find cart item to get backend ID
  const itemToRemove = cart.find(
    (item) => item.id === id && (!selectedSize || item.selectedSize === selectedSize)
  );
  
  if (!itemToRemove?.cartItemId) {
    console.error("Cannot remove: cartItemId not found");
    // Just remove from local state
    setCart((prev) => prev.filter(...));
    return;
  }
  
  // 2. Delete from backend FIRST (use cartItemId, not product id!)
  try {
    const response = await api.delete(`/cart/items/${itemToRemove.cartItemId}`);
    
    // 3. Update cart with backend response (source of truth)
    if (response.data && response.data.items) {
      const backendItems = response.data.items.map(convertBackendItem);
      setCart(backendItems);
      localStorage.setItem("zavera_cart", JSON.stringify(backendItems));
    } else {
      // Empty cart
      setCart([]);
      localStorage.removeItem("zavera_cart");
    }
  } catch (error) {
    console.error("Failed to remove from backend cart:", error);
    // On error, reload cart from backend to ensure consistency
    await loadCartFromBackend();
  }
};
```

**Mobile app HARUS:**
1. Find item by product ID + size
2. **Use `cartItemId` untuk delete** (bukan product ID!)
3. Delete dari backend DULU
4. Update cart dengan response backend
5. Jika error, reload cart dari backend

### Update Quantity Logic
```typescript
const updateQuantity = async (id: number, quantity: number, selectedSize?: string) => {
  // 1. If quantity <= 0, remove item
  if (quantity <= 0) {
    removeFromCart(id, selectedSize);
    return;
  }
  
  // 2. Find cart item to get backend ID
  const itemToUpdate = cart.find(
    (item) => item.id === id && (!selectedSize || item.selectedSize === selectedSize)
  );
  
  // 3. Optimistically update local state
  setCart((prev) =>
    prev.map((item) =>
      item.id === id && (!selectedSize || item.selectedSize === selectedSize)
        ? { ...item, quantity }
        : item
    )
  );
  
  // 4. Sync to backend (use cartItemId!)
  if (itemToUpdate?.cartItemId) {
    try {
      const response = await api.put(`/cart/items/${itemToUpdate.cartItemId}`, {
        quantity,
      });
      
      if (response.data && response.data.items) {
        const backendItems = response.data.items.map(convertBackendItem);
        setCart(backendItems);
      }
    } catch (error) {
      console.error("Failed to update backend cart:", error);
    }
  }
};
```

### Clear Cart Logic
```typescript
const clearCart = async () => {
  // 1. Clear local state
  setCart([]);
  localStorage.removeItem("zavera_cart");
  
  // 2. Clear backend cart
  try {
    await api.delete("/cart");
  } catch (error) {
    console.error("Failed to clear backend cart:", error);
  }
};
```

### Cart Validation
```typescript
const validateCart = async () => {
  const token = localStorage.getItem("auth_token");
  if (!token) {
    return null;
  }
  
  try {
    const response = await api.get("/cart/validate");
    return response.data; // { valid, changes, cart, message }
  } catch (error) {
    console.error("Failed to validate cart:", error);
    return null;
  }
};
```

**Call validate BEFORE checkout!**

### Listen for Cart Updates
```typescript
// Website listens for 'cart-updated' event from wishlist
useEffect(() => {
  const handleCartUpdated = () => {
    refreshCart();
  };
  
  window.addEventListener('cart-updated', handleCartUpdated);
  
  return () => {
    window.removeEventListener('cart-updated', handleCartUpdated);
  };
}, [refreshCart]);
```

**Mobile app:** Use EventBus atau Stream untuk trigger cart refresh dari wishlist.

---

## 3. ❤️ WISHLIST LOGIC

### Move to Cart from Wishlist
```typescript
// After moving item to cart, trigger cart refresh
await api.post(`/wishlist/${productId}/move-to-cart`);

// Dispatch event to refresh cart
window.dispatchEvent(new Event('cart-updated'));
```

**Mobile app HARUS:**
- Call `/wishlist/:productId/move-to-cart` endpoint
- Trigger cart refresh event
- Refresh wishlist

---

## 4. 🚚 SHIPPING LOGIC

### Area Search (Biteship)
```typescript
// Website: api.ts
export async function searchAreas(query: string): Promise<BiteshipArea[]> {
  const response = await api.get("/shipping/areas", {
    params: { q: query },
  });
  return response.data.areas || [];
}
```

**Mobile app HARUS:**
- Use `/shipping/areas?q=query` untuk search
- Return list of areas dengan `area_id`, `name`, `postal_code`

### Get Shipping Rates
```typescript
export async function getShippingRatesByAreaId(
  originAreaId: string,
  destinationAreaId: string,
  weight: number
) {
  const response = await api.post("/shipping/rates", {
    origin_area_id: originAreaId,
    destination_area_id: destinationAreaId,
    weight,
  });
  return response.data;
}
```

---

## 5. 💳 CHECKOUT LOGIC

### Checkout Flow
1. **Validate cart** - Call `/cart/validate`
2. **Select shipping** - Get rates, user selects courier
3. **Create order** - POST `/checkout/shipping` with:
   ```json
   {
     "shipping_address_id": 123,
     "courier_code": "jne",
     "courier_service": "REG"
   }
   ```
4. **Create payment** - POST `/payments/initiate` or `/payments/core/create`
5. **Show payment** - Display Snap token or VA details
6. **Track order** - GET `/orders/:code`

---

## 6. 🔄 STATE MANAGEMENT

### Website menggunakan React Context:
- `AuthContext` - User & token
- `CartContext` - Cart items & operations
- `WishlistContext` - Wishlist items

**Mobile app bisa pakai:**
- Provider (seperti website)
- Riverpod
- Bloc
- GetX

**Yang penting:** Logic sama dengan website!

---

## 7. 📱 MOBILE-SPECIFIC CONSIDERATIONS

### Token Storage
- Website: `localStorage`
- Mobile: `SharedPreferences` (Flutter) atau `AsyncStorage` (React Native)

### API Base URL
- Website: `process.env.NEXT_PUBLIC_API_URL`
- Mobile: Hardcoded atau config file (ganti dengan IP laptop untuk testing)

### Session Expired Handling
- Website: Redirect ke `/login?session_expired=true`
- Mobile: Navigate ke LoginScreen dengan parameter

---

## ✅ CHECKLIST IMPLEMENTASI

### Authentication
- [ ] Login flow dengan cart refresh
- [ ] Google login dengan cart refresh
- [ ] Logout dengan clear cart
- [ ] Auto-logout on 401
- [ ] Session expired handling

### Cart
- [ ] Load dari backend FIRST
- [ ] Clear cart jika tidak login
- [ ] Add to cart dengan total quantity
- [ ] Remove dengan cartItemId
- [ ] Update quantity dengan cartItemId
- [ ] Clear cart (local + backend)
- [ ] Validate cart before checkout
- [ ] Listen for cart-updated events

### Wishlist
- [ ] Move to cart dengan cart refresh

### Shipping
- [ ] Area search (Biteship)
- [ ] Get shipping rates

### Checkout
- [ ] Validate cart
- [ ] Create order
- [ ] Create payment (Snap + VA)
- [ ] Track order

---

## 🚨 COMMON MISTAKES - JANGAN SAMPAI!

### ❌ SALAH:
```dart
// Add to cart tanpa check login
await addToCart(product);

// Remove dengan product ID
await api.delete('/cart/items/$productId');

// Add quantity increment
await api.post('/cart/items', {
  'product_id': productId,
  'quantity': 1, // SALAH! Harus total quantity
});

// Load cart dari local storage dulu
final cart = await loadFromLocal();
```

### ✅ BENAR:
```dart
// Check login dulu
if (!isLoggedIn) {
  showLoginDialog();
  return;
}
await addToCart(product);

// Remove dengan cartItemId
await api.delete('/cart/items/${item.cartItemId}');

// Add dengan total quantity
final existingQty = getExistingQuantity(productId);
await api.post('/cart/items', {
  'product_id': productId,
  'quantity': existingQty + 1, // Total quantity
});

// Load dari backend dulu
final cart = await loadFromBackend();
if (cart == null) {
  cart = await loadFromLocal(); // Fallback
}
```

---

## 📞 NEED HELP?

Kalau ada yang bingung atau tidak yakin, **TANYA DULU** sebelum implement!

Lebih baik tanya daripada salah implement dan kena komplain.

**Files to reference:**
- `ZAVERA-FASHION-STORE/frontend/src/context/CartContext.tsx`
- `ZAVERA-FASHION-STORE/frontend/src/context/AuthContext.tsx`
- `ZAVERA-FASHION-STORE/frontend/src/lib/api.ts`

---

**INGAT:** Logic mobile app HARUS 100% sama dengan website. Jangan ada perbedaan!
