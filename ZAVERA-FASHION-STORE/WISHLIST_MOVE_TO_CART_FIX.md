# Wishlist "Move to Cart" Fix - 27 Januari 2026

## 🐛 Issue

**Problem:** Saat klik "MOVE TO CART" di wishlist page:
- Item hilang dari wishlist ✅ (benar)
- Item TIDAK muncul di cart ❌ (bug)
- Cart tetap kosong ❌ (bug)
- Cart counter di header tidak update ❌ (bug)

**Root Cause:**
- Backend API berfungsi dengan benar (menambahkan ke cart, menghapus dari wishlist)
- Frontend WishlistContext hanya refresh wishlist setelah API call
- Frontend CartContext TIDAK di-refresh, sehingga tidak tahu ada item baru
- Circular dependency issue: WishlistContext tidak bisa import CartContext

---

## ✅ Solution

Menggunakan **Custom Event** untuk komunikasi antar context tanpa circular dependency.

### Flow:
1. User klik "MOVE TO CART" di wishlist page
2. WishlistContext call API `/wishlist/:productId/move-to-cart`
3. API berhasil → Backend menambahkan ke cart & menghapus dari wishlist
4. WishlistContext dispatch custom event `'cart-updated'`
5. CartContext listen ke event ini dan auto-refresh cart
6. Cart counter di header update otomatis
7. User bisa langsung lihat item di cart

---

## 🔧 Implementation

### 1. WishlistContext - Dispatch Event

**File:** `frontend/src/context/WishlistContext.tsx`

```typescript
// Move to cart
const moveToCart = useCallback(async (productId: number) => {
  if (!isAuthenticated) {
    showToast("Please login to move items to cart", "error");
    return;
  }

  setIsLoading(true);
  try {
    await api.post(`/wishlist/${productId}/move-to-cart`);
    
    // Refresh wishlist to remove the item
    await loadWishlist();
    
    // 🔥 NEW: Trigger cart refresh event
    window.dispatchEvent(new CustomEvent('cart-updated'));
    
    showToast("Moved to cart", "success");
  } catch (error: any) {
    console.error("Failed to move to cart:", error);
    const message = error.response?.data?.message || "Failed to move to cart";
    showToast(message, "error");
  } finally {
    setIsLoading(false);
  }
}, [isAuthenticated, loadWishlist, showToast]);
```

### 2. CartContext - Listen to Event

**File:** `frontend/src/context/CartContext.tsx`

```typescript
// 🔥 NEW: Listen for cart-updated event from wishlist
useEffect(() => {
  const handleCartUpdated = () => {
    console.log("Cart updated event received, refreshing cart...");
    refreshCart();
  };

  window.addEventListener('cart-updated', handleCartUpdated);
  
  return () => {
    window.removeEventListener('cart-updated', handleCartUpdated);
  };
}, [refreshCart]);
```

---

## 🎯 Benefits

### 1. No Circular Dependency
- WishlistContext dan CartContext tetap independent
- Tidak perlu import satu sama lain
- Clean architecture

### 2. Decoupled Communication
- Menggunakan browser native event system
- Scalable untuk future features
- Easy to debug (bisa lihat event di console)

### 3. Real-time Updates
- Cart counter di header update instantly
- Cart page auto-refresh jika dibuka
- User experience lebih smooth

### 4. Reusable Pattern
- Pattern ini bisa digunakan untuk komunikasi antar context lainnya
- Contoh: `'order-completed'`, `'product-updated'`, dll

---

## 🧪 Testing

### Manual Test Steps:

1. **Login** ke aplikasi
2. **Add produk** ke wishlist (klik heart icon)
3. **Buka wishlist page** (klik icon ❤️ di header)
4. **Klik "MOVE TO CART"** pada salah satu produk
5. **Verify:**
   - ✅ Toast notification "Moved to cart" muncul
   - ✅ Item hilang dari wishlist
   - ✅ Wishlist counter di header berkurang
   - ✅ Cart counter di header bertambah
   - ✅ Buka cart page → item muncul di cart
   - ✅ Quantity = 1
   - ✅ Size = M (default)

### Edge Cases to Test:

1. **Out of Stock Product:**
   - Button "MOVE TO CART" harus disabled
   - Tidak bisa diklik

2. **Multiple Items:**
   - Move 3 items dari wishlist ke cart
   - Semua harus muncul di cart
   - Cart counter harus +3

3. **Network Error:**
   - Matikan backend
   - Klik "MOVE TO CART"
   - Harus muncul error toast
   - Item tetap di wishlist (tidak hilang)

4. **Not Logged In:**
   - Logout
   - Coba akses wishlist
   - Harus redirect ke login

---

## 📊 Before vs After

### Before (Bug):
```
User: Klik "MOVE TO CART"
↓
API Call: POST /wishlist/123/move-to-cart ✅
↓
Backend: Add to cart ✅, Remove from wishlist ✅
↓
Frontend: Refresh wishlist ✅
↓
Frontend: Cart NOT refreshed ❌
↓
Result: Item hilang dari wishlist, tapi tidak muncul di cart ❌
```

### After (Fixed):
```
User: Klik "MOVE TO CART"
↓
API Call: POST /wishlist/123/move-to-cart ✅
↓
Backend: Add to cart ✅, Remove from wishlist ✅
↓
Frontend: Refresh wishlist ✅
↓
Frontend: Dispatch 'cart-updated' event ✅
↓
Frontend: CartContext listen & refresh cart ✅
↓
Result: Item hilang dari wishlist DAN muncul di cart ✅
```

---

## 🔍 Technical Details

### Custom Event API

```typescript
// Dispatch event (WishlistContext)
window.dispatchEvent(new CustomEvent('cart-updated'));

// Listen to event (CartContext)
window.addEventListener('cart-updated', handleCartUpdated);

// Cleanup (CartContext)
window.removeEventListener('cart-updated', handleCartUpdated);
```

### Why Custom Event?

1. **Browser Native:** Tidak perlu library tambahan
2. **Type Safe:** TypeScript support
3. **Performance:** Minimal overhead
4. **Debugging:** Bisa monitor di browser DevTools
5. **Scalable:** Bisa pass data via `detail` property

### Alternative Solutions (Not Used):

1. ❌ **Direct Import:** Circular dependency
2. ❌ **Global State (Redux/Zustand):** Overkill untuk simple case
3. ❌ **Context Composition:** Complex refactoring
4. ❌ **Redirect to Cart:** Bad UX (full page reload)
5. ✅ **Custom Event:** Simple, clean, effective

---

## 🚀 Future Enhancements

### Possible Event-Driven Features:

1. **Order Completed Event:**
   ```typescript
   window.dispatchEvent(new CustomEvent('order-completed', {
     detail: { orderId: 123, total: 500000 }
   }));
   ```

2. **Product Updated Event:**
   ```typescript
   window.dispatchEvent(new CustomEvent('product-updated', {
     detail: { productId: 456, newPrice: 299000 }
   }));
   ```

3. **Stock Alert Event:**
   ```typescript
   window.dispatchEvent(new CustomEvent('stock-low', {
     detail: { productId: 789, stock: 2 }
   }));
   ```

---

## 📝 Files Changed

1. **`frontend/src/context/WishlistContext.tsx`**
   - Added: `window.dispatchEvent(new CustomEvent('cart-updated'))`
   - Location: Inside `moveToCart` function after successful API call

2. **`frontend/src/context/CartContext.tsx`**
   - Added: New `useEffect` to listen for `'cart-updated'` event
   - Location: After global refresh function setup

---

## ✅ Verification Checklist

- [x] No TypeScript errors
- [x] No circular dependency
- [x] Event listener cleanup on unmount
- [x] Cart refreshes after move to cart
- [x] Wishlist refreshes after move to cart
- [x] Cart counter updates
- [x] Wishlist counter updates
- [x] Toast notifications work
- [x] Error handling works
- [x] Works with multiple items
- [x] Works with out of stock items (disabled)

---

## 🎉 Result

**MOVE TO CART SEKARANG BERFUNGSI DENGAN SEMPURNA!**

- ✅ Item hilang dari wishlist
- ✅ Item muncul di cart
- ✅ Cart counter update
- ✅ Wishlist counter update
- ✅ Toast notifications
- ✅ No page reload
- ✅ Smooth UX

---

**Status:** ✅ Fixed and Tested
**Last Updated:** 27 Januari 2026
