# Debug: Keranjang Kosong di Checkout

## Masalah

User melihat "Keranjang Kosong" di halaman checkout padahal di halaman cart ada barang (2 items).

## Kemungkinan Penyebab

### 1. Token Expired atau Hilang
- User sudah login dan add to cart
- Saat pindah ke checkout, token expired
- CartContext clear cart karena tidak ada token

### 2. Cart Tidak Ter-sync
- Cart di localStorage berbeda dengan backend
- Backend cart kosong tapi localStorage ada isi
- CartContext prioritas backend, jadi cart jadi kosong

### 3. Race Condition
- Cart masih loading saat checkout page render
- Kondisi `cart.length === 0` terpenuhi sebelum cart selesai load

## Cara Debug

### Step 1: Buka Browser Console (F12)

Saat Anda klik "Proceed to Checkout" dan muncul "Keranjang Kosong", lihat console log:

```
🛒 loadCartFromBackend: token exists? true/false
🛒 loadCartFromBackend: Fetching cart from API...
🛒 loadCartFromBackend: API response: {...}
🛒 loadCartFromBackend: Converted items: [...]
🛒 Checkout: Cart is empty
🛒 isAuthenticated: true/false
🛒 user: {...}
🛒 cart state: []
```

### Step 2: Analisa Log

#### Scenario A: Token Hilang
```
🛒 loadCartFromBackend: token exists? false
🛒 loadCartFromBackend: No token, clearing cart
🛒 Checkout: Cart is empty
🛒 isAuthenticated: false
```

**Solusi**: User harus login ulang

#### Scenario B: Backend Cart Kosong
```
🛒 loadCartFromBackend: token exists? true
🛒 loadCartFromBackend: Fetching cart from API...
🛒 loadCartFromBackend: API response: {items: []}
🛒 loadCartFromBackend: Empty cart from backend
🛒 Checkout: Cart is empty
```

**Solusi**: Cart di backend kosong, perlu sync dari localStorage

#### Scenario C: API Error
```
🛒 loadCartFromBackend: token exists? true
🛒 loadCartFromBackend: Fetching cart from API...
🛒 loadCartFromBackend: Failed to load cart from backend: Error: ...
🛒 Checkout: Cart is empty
```

**Solusi**: Backend error, perlu cek backend logs

### Step 3: Cek localStorage

Di console, ketik:
```javascript
localStorage.getItem('zavera_cart')
localStorage.getItem('auth_token')
```

**Harusnya**:
- `zavera_cart`: Ada array dengan 2 items
- `auth_token`: Ada token string

### Step 4: Cek Backend API

Di console, ketik:
```javascript
fetch('http://localhost:8080/api/cart', {
  headers: {
    'Authorization': 'Bearer ' + localStorage.getItem('auth_token')
  }
}).then(r => r.json()).then(console.log)
```

**Harusnya**: Return cart dengan items

## Solusi Sementara

### Quick Fix 1: Refresh Page
```
1. Di halaman cart, refresh (F5)
2. Klik "Proceed to Checkout" lagi
```

### Quick Fix 2: Re-add Items
```
1. Kembali ke halaman produk
2. Add to cart lagi
3. Proceed to checkout
```

### Quick Fix 3: Clear dan Login Ulang
```
1. Logout
2. Clear browser cache (Ctrl+Shift+Del)
3. Login lagi
4. Add to cart
5. Proceed to checkout
```

## Solusi Permanent (Developer)

### Fix 1: Jangan Clear Cart Jika Token Hilang

**File**: `frontend/src/context/CartContext.tsx`

**Masalah**: Saat token hilang, cart langsung di-clear
```typescript
if (!token) {
  setCart([]);  // ❌ Terlalu agresif
  return true;
}
```

**Solusi**: Fallback ke localStorage
```typescript
if (!token) {
  // Fallback to localStorage instead of clearing
  const savedCart = localStorage.getItem("zavera_cart");
  if (savedCart) {
    try {
      const parsed = JSON.parse(savedCart);
      setCart(parsed);
    } catch (e) {
      setCart([]);
    }
  }
  return false; // Allow localStorage fallback
}
```

### Fix 2: Add Loading State di Checkout

**File**: `frontend/src/app/checkout/page.tsx`

**Masalah**: Render "Keranjang Kosong" sebelum cart selesai load

**Solusi**: Tambah loading state
```typescript
const { cart, isLoading } = useCart();

if (isLoading) {
  return <LoadingSpinner message="Memuat keranjang..." />;
}

if (cart.length === 0) {
  return <EmptyCartView />;
}
```

### Fix 3: Auto-sync Cart Saat Checkout

**File**: `frontend/src/app/checkout/page.tsx`

**Solusi**: Sync cart dari localStorage ke backend saat mount
```typescript
useEffect(() => {
  const syncCart = async () => {
    if (isAuthenticated && cart.length === 0) {
      // Try to sync from localStorage
      const savedCart = localStorage.getItem("zavera_cart");
      if (savedCart) {
        await syncCartToBackend();
        await refreshCart();
      }
    }
  };
  syncCart();
}, []);
```

## Files Modified

1. ✅ `frontend/src/app/checkout/page.tsx`
   - Tambah console.log untuk debugging

2. ✅ `frontend/src/context/CartContext.tsx`
   - Tambah console.log untuk debugging

## Testing

```bash
cd frontend
npm run dev
```

### Test Flow:
1. Login
2. Add 2 items to cart
3. Go to cart page - verify 2 items shown
4. Open console (F12)
5. Click "Proceed to Checkout"
6. Check console logs
7. Screenshot dan kirim logs

## Expected Logs (Normal Flow)

```
🛒 loadCartFromBackend: token exists? true
🛒 loadCartFromBackend: Fetching cart from API...
🛒 loadCartFromBackend: API response: {
  id: 1,
  items: [
    {id: 1, product_name: "Hip Hop Jeans", quantity: 1, ...},
    {id: 2, product_name: "Hip Hop Jeans", quantity: 2, ...}
  ],
  subtotal: 990000,
  item_count: 3
}
🛒 loadCartFromBackend: Converted items: [
  {id: 46, name: "Hip Hop Jeans", quantity: 1, ...},
  {id: 46, name: "Hip Hop Jeans", quantity: 2, ...}
]
```

Checkout page should show cart items, NOT "Keranjang Kosong"

## Next Steps

1. ✅ Tambah logging (DONE)
2. ⏳ User test dan kirim console logs
3. ⏳ Analisa logs untuk identify root cause
4. ⏳ Implement permanent fix based on logs

---

**Status**: 🔍 Debugging
**Priority**: HIGH (Blocking checkout)
**Need**: Console logs from user
