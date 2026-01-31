"use client";

import React, {
  createContext,
  useContext,
  useState,
  useEffect,
  useCallback,
  useRef,
} from "react";
import { CartItem, Product } from "@/types";
import api from "@/lib/api";

interface BackendCartItem {
  id: number;
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
  id: number;
  items: BackendCartItem[];
  subtotal: number;
  item_count: number;
}

interface CartContextType {
  cart: CartItem[];
  addToCart: (item: CartItem | (Product & { quantity?: number; selectedSize?: string })) => void;
  removeFromCart: (id: number, selectedSize?: string) => void;
  updateQuantity: (id: number, quantity: number, selectedSize?: string) => void;
  clearCart: () => void;
  getTotalPrice: () => number;
  getTotalItems: () => number;
  syncCartToBackend: () => Promise<void>;
  refreshCart: () => Promise<void>;
  validateCart: () => Promise<CartValidationResult | null>;
  isLoading: boolean;
}

interface CartValidationResult {
  valid: boolean;
  changes: CartItemChange[];
  cart: any;
  message: string;
}

interface CartItemChange {
  cart_item_id: number;
  product_id: number;
  product_name: string;
  change_type: string;
  old_price?: number;
  new_price?: number;
  old_weight?: number;
  new_weight?: number;
  current_stock?: number;
  message: string;
}

const CartContext = createContext<CartContextType | undefined>(undefined);

// Export refreshCart function for use after login
let globalRefreshCart: (() => Promise<void>) | null = null;

export const triggerCartRefresh = () => {
  if (globalRefreshCart) {
    globalRefreshCart();
  }
};

export function CartProvider({ children }: { children: React.ReactNode }) {
  const [cart, setCart] = useState<CartItem[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isHydrated, setIsHydrated] = useState(false);
  const isInitialLoad = useRef(true);

  // Convert backend cart item to frontend cart item
  const convertBackendItem = (item: BackendCartItem): CartItem => ({
    id: item.product_id,
    name: item.product_name,
    price: item.price_per_unit,
    image_url: item.product_image,
    description: "",
    category: "wanita" as const,
    quantity: item.quantity,
    selectedSize: item.metadata?.selected_size || "M",
    stock: item.stock,
    // Keep backend item ID for updates
    cartItemId: item.id,
  });

  // Load cart from backend API
  const loadCartFromBackend = useCallback(async () => {
    // Check if user is logged in
    const token = localStorage.getItem("auth_token");
    console.log("🛒 loadCartFromBackend: token exists?", !!token);
    
    if (!token) {
      // No user logged in, clear cart
      console.log("🛒 loadCartFromBackend: No token, clearing cart");
      setCart([]);
      localStorage.removeItem("zavera_cart");
      return true; // Return true to prevent localStorage fallback
    }
    
    try {
      console.log("🛒 loadCartFromBackend: Fetching cart from API...");
      const response = await api.get<BackendCartResponse>("/cart");
      console.log("🛒 loadCartFromBackend: API response:", response.data);
      
      if (response.data && response.data.items) {
        const backendItems = response.data.items.map(convertBackendItem);
        console.log("🛒 loadCartFromBackend: Converted items:", backendItems);
        setCart(backendItems);
        // Also save to localStorage as backup
        localStorage.setItem("zavera_cart", JSON.stringify(backendItems));
        return true;
      }
      // Empty cart from backend
      console.log("🛒 loadCartFromBackend: Empty cart from backend");
      setCart([]);
      localStorage.removeItem("zavera_cart");
      return true;
    } catch (error) {
      console.log("🛒 loadCartFromBackend: Failed to load cart from backend:", error);
      // If unauthorized, clear cart
      setCart([]);
      localStorage.removeItem("zavera_cart");
      return true;
    }
  }, []);

  // Load cart on mount - try backend first, fallback to localStorage
  useEffect(() => {
    const initCart = async () => {
      setIsLoading(true);
      
      // Try to load from backend first
      const loadedFromBackend = await loadCartFromBackend();
      
      if (!loadedFromBackend) {
        // Fallback to localStorage
        const savedCart = localStorage.getItem("zavera_cart");
        if (savedCart) {
          try {
            const parsed = JSON.parse(savedCart);
            const validCart = parsed.filter((item: CartItem) => 
              item && 
              typeof item.id === 'number' && 
              typeof item.price === 'number' && 
              !isNaN(item.price) &&
              typeof item.quantity === 'number' &&
              !isNaN(item.quantity)
            );
            setCart(validCart);
          } catch (e) {
            console.error("Failed to parse cart:", e);
            setCart([]);
          }
        }
      }
      
      setIsHydrated(true);
      setIsLoading(false);
      isInitialLoad.current = false;
    };

    initCart();
  }, [loadCartFromBackend]);

  // Refresh cart from backend (call after login)
  const refreshCart = useCallback(async () => {
    setIsLoading(true);
    await loadCartFromBackend();
    setIsLoading(false);
  }, [loadCartFromBackend]);

  // Set global refresh function
  useEffect(() => {
    globalRefreshCart = refreshCart;
    return () => {
      globalRefreshCart = null;
    };
  }, [refreshCart]);

  // Listen for cart-updated event from wishlist
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

  // Save cart to localStorage whenever it changes (as backup)
  useEffect(() => {
    if (isHydrated && !isInitialLoad.current) {
      localStorage.setItem("zavera_cart", JSON.stringify(cart));
    }
  }, [cart, isHydrated]);

  // Sync entire local cart to backend
  const syncCartToBackend = useCallback(async () => {
    if (cart.length === 0) return;

    setIsLoading(true);
    let successCount = 0;
    const failedItems: number[] = [];
    
    try {
      // First, clear the backend cart
      try {
        await api.delete("/cart");
      } catch (e) {
        console.log("Cart already empty or error clearing:", e);
      }

      // Then add each item to the backend cart
      for (const item of cart) {
        try {
          const payload: any = {
            product_id: item.id,
            quantity: item.quantity,
            metadata: {
              selected_size: item.selectedSize || "M",
            },
          };
          
          // Add variant_id if present
          if ((item as any).variant_id) {
            payload.variant_id = (item as any).variant_id;
          }
          
          await api.post("/cart/items", payload);
          successCount++;
        } catch (error) {
          console.error(`Failed to add item ${item.name} (ID: ${item.id}) to backend cart:`, error);
          failedItems.push(item.id);
        }
      }

      // Remove failed items from local cart
      if (failedItems.length > 0) {
        setCart(prev => prev.filter(i => !failedItems.includes(i.id)));
        console.log(`Removed ${failedItems.length} invalid items from cart`);
      }

      // If no items were synced successfully, throw error
      if (successCount === 0 && cart.length > 0) {
        setCart([]);
        localStorage.removeItem("zavera_cart");
        throw new Error("Cart items are no longer available. Please add items again.");
      }

      console.log(`✅ Cart synced: ${successCount}/${cart.length} items`);
    } catch (error) {
      console.error("Failed to sync cart to backend:", error);
      throw error;
    } finally {
      setIsLoading(false);
    }
  }, [cart]);

  // Add to cart - sync to backend immediately
  const addToCart = useCallback(async (item: CartItem | (Product & { quantity?: number; selectedSize?: string })) => {
    // Check if user is logged in
    const token = localStorage.getItem("auth_token");
    if (!token) {
      // User not logged in, don't add to cart
      console.log("User must be logged in to add items to cart");
      return;
    }
    
    const cartItem: CartItem = {
      ...item,
      quantity: item.quantity || 1,
      selectedSize: item.selectedSize || "M",
    };

    if (typeof cartItem.price !== 'number' || isNaN(cartItem.price)) {
      console.error("Invalid price for cart item:", cartItem);
      return;
    }

    // Calculate the new total quantity (existing + new)
    const existingItem = cart.find(
      (i) => i.id === cartItem.id && i.selectedSize === cartItem.selectedSize
    );
    const newTotalQuantity = existingItem 
      ? existingItem.quantity + cartItem.quantity 
      : cartItem.quantity;

    // Optimistically update local state
    setCart((prev) => {
      const existing = prev.find(
        (i) => i.id === cartItem.id && i.selectedSize === cartItem.selectedSize
      );
      if (existing) {
        return prev.map((i) =>
          i.id === cartItem.id && i.selectedSize === cartItem.selectedSize
            ? { ...i, quantity: i.quantity + cartItem.quantity }
            : i
        );
      }
      return [...prev, cartItem];
    });

    // Sync to backend - send the TOTAL quantity (backend will SET, not ADD)
    try {
      const payload: any = {
        product_id: cartItem.id,
        quantity: newTotalQuantity,
        metadata: {
          selected_size: cartItem.selectedSize || "M",
        },
      };
      
      // Add variant_id if present (for variant products)
      if ((cartItem as any).variant_id) {
        payload.variant_id = (cartItem as any).variant_id;
      }
      
      const response = await api.post<BackendCartResponse>("/cart/items", payload);
      
      // Update cart with backend response
      if (response.data && response.data.items) {
        const backendItems = response.data.items.map(convertBackendItem);
        setCart(backendItems);
      }
    } catch (error) {
      console.error("Failed to add to backend cart:", error);
      // Keep local state as fallback
    }
  }, [cart]);

  // Remove from cart - sync to backend
  const removeFromCart = useCallback(async (id: number, selectedSize?: string) => {
    // Find the cart item to get its backend ID
    const itemToRemove = cart.find(
      (item) => item.id === id && (!selectedSize || item.selectedSize === selectedSize)
    );

    console.log("🗑️ Removing item:", { id, selectedSize, cartItemId: itemToRemove?.cartItemId });

    if (!itemToRemove?.cartItemId) {
      console.error("❌ Cannot remove: cartItemId not found for item", { id, selectedSize });
      // Just remove from local state if no backend ID
      setCart((prev) =>
        prev.filter(
          (item) =>
            !(item.id === id && (!selectedSize || item.selectedSize === selectedSize))
        )
      );
      return;
    }

    // Sync to backend FIRST before updating local state
    try {
      console.log(`🔄 Calling DELETE /cart/items/${itemToRemove.cartItemId}`);
      const response = await api.delete<BackendCartResponse>(`/cart/items/${itemToRemove.cartItemId}`);
      console.log("✅ Backend delete successful, response:", response.data);
      
      // Update cart with backend response (this is the source of truth)
      if (response.data && response.data.items) {
        const backendItems = response.data.items.map(convertBackendItem);
        setCart(backendItems);
        localStorage.setItem("zavera_cart", JSON.stringify(backendItems));
        console.log("✅ Cart updated from backend, new item count:", backendItems.length);
      } else {
        // Empty cart
        setCart([]);
        localStorage.removeItem("zavera_cart");
        console.log("✅ Cart is now empty");
      }
    } catch (error) {
      console.error("❌ Failed to remove from backend cart:", error);
      // On error, reload cart from backend to ensure consistency
      await loadCartFromBackend();
    }
  }, [cart, loadCartFromBackend]);

  // Update quantity - sync to backend
  const updateQuantity = useCallback(async (
    id: number,
    quantity: number,
    selectedSize?: string
  ) => {
    if (quantity <= 0) {
      removeFromCart(id, selectedSize);
      return;
    }

    // Find the cart item to get its backend ID
    const itemToUpdate = cart.find(
      (item) => item.id === id && (!selectedSize || item.selectedSize === selectedSize)
    );

    // Optimistically update local state
    setCart((prev) =>
      prev.map((item) =>
        item.id === id && (!selectedSize || item.selectedSize === selectedSize)
          ? { ...item, quantity }
          : item
      )
    );

    // Sync to backend if we have the backend item ID
    if (itemToUpdate?.cartItemId) {
      try {
        const response = await api.put<BackendCartResponse>(`/cart/items/${itemToUpdate.cartItemId}`, {
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
  }, [cart, removeFromCart]);

  // Clear cart - sync to backend
  const clearCart = useCallback(async () => {
    setCart([]);
    localStorage.removeItem("zavera_cart");
    
    try {
      await api.delete("/cart");
    } catch (error) {
      console.error("Failed to clear backend cart:", error);
    }
  }, []);

  const getTotalItems = useCallback(() => {
    if (!isHydrated) return 0;
    return cart.reduce((total, item) => {
      const quantity = typeof item.quantity === 'number' && !isNaN(item.quantity) ? item.quantity : 0;
      return total + quantity;
    }, 0);
  }, [cart, isHydrated]);

  const getTotalPrice = useCallback(() => {
    if (!isHydrated) return 0;
    return cart.reduce((total, item) => {
      const price = typeof item.price === 'number' && !isNaN(item.price) ? item.price : 0;
      const quantity = typeof item.quantity === 'number' && !isNaN(item.quantity) ? item.quantity : 0;
      return total + (price * quantity);
    }, 0);
  }, [cart, isHydrated]);

  // Validate cart against current product data
  const validateCart = useCallback(async (): Promise<CartValidationResult | null> => {
    const token = localStorage.getItem("auth_token");
    if (!token) {
      return null;
    }

    try {
      const response = await api.get<CartValidationResult>("/cart/validate");
      return response.data;
    } catch (error) {
      console.error("Failed to validate cart:", error);
      return null;
    }
  }, []);

  return (
    <CartContext.Provider
      value={{
        cart: isHydrated ? cart : [],
        addToCart,
        removeFromCart,
        updateQuantity,
        clearCart,
        getTotalPrice,
        getTotalItems,
        syncCartToBackend,
        refreshCart,
        validateCart,
        isLoading,
      }}
    >
      {children}
    </CartContext.Provider>
  );
}

export function useCart() {
  const context = useContext(CartContext);
  if (!context) {
    throw new Error("useCart must be used within CartProvider");
  }
  return context;
}
