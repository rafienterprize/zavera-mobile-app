"use client";

import React, {
  createContext,
  useContext,
  useState,
  useEffect,
  useCallback,
} from "react";
import api from "@/lib/api";
import { useAuth } from "./AuthContext";
import { useToast } from "@/components/ui/Toast";

interface WishlistItem {
  id: number;
  product_id: number;
  product_name: string;
  product_image: string;
  product_price: number;
  product_stock: number;
  is_available: boolean;
  added_at: string;
}

interface WishlistResponse {
  items: WishlistItem[];
  count: number;
}

interface WishlistContextType {
  wishlist: WishlistItem[];
  wishlistCount: number;
  isLoading: boolean;
  addToWishlist: (productId: number) => Promise<void>;
  removeFromWishlist: (productId: number) => Promise<void>;
  moveToCart: (productId: number) => Promise<void>;
  isInWishlist: (productId: number) => boolean;
  refreshWishlist: () => Promise<void>;
}

const WishlistContext = createContext<WishlistContextType | undefined>(undefined);

export function WishlistProvider({ children }: { children: React.ReactNode }) {
  const [wishlist, setWishlist] = useState<WishlistItem[]>([]);
  const [wishlistCount, setWishlistCount] = useState(0);
  const [isLoading, setIsLoading] = useState(false);
  const { isAuthenticated } = useAuth();
  const { showToast } = useToast();

  // Load wishlist from backend
  const loadWishlist = useCallback(async () => {
    if (!isAuthenticated) {
      setWishlist([]);
      setWishlistCount(0);
      return;
    }

    try {
      const response = await api.get<WishlistResponse>("/wishlist");
      if (response.data) {
        setWishlist(response.data.items || []);
        setWishlistCount(response.data.count || 0);
      }
    } catch (error) {
      console.error("Failed to load wishlist:", error);
      setWishlist([]);
      setWishlistCount(0);
    }
  }, [isAuthenticated]);

  // Load wishlist on mount and when auth changes
  useEffect(() => {
    loadWishlist();
  }, [loadWishlist]);

  // Refresh wishlist
  const refreshWishlist = useCallback(async () => {
    setIsLoading(true);
    await loadWishlist();
    setIsLoading(false);
  }, [loadWishlist]);

  // Add to wishlist
  const addToWishlist = useCallback(async (productId: number) => {
    if (!isAuthenticated) {
      showToast("Please login to add items to wishlist", "error");
      return;
    }

    setIsLoading(true);
    try {
      const response = await api.post<WishlistResponse>("/wishlist", {
        product_id: productId,
      });
      
      if (response.data) {
        setWishlist(response.data.items || []);
        setWishlistCount(response.data.count || 0);
        showToast("Added to wishlist", "success");
      }
    } catch (error: any) {
      console.error("Failed to add to wishlist:", error);
      const message = error.response?.data?.message || "Failed to add to wishlist";
      showToast(message, "error");
    } finally {
      setIsLoading(false);
    }
  }, [isAuthenticated, showToast]);

  // Remove from wishlist
  const removeFromWishlist = useCallback(async (productId: number) => {
    if (!isAuthenticated) {
      return;
    }

    setIsLoading(true);
    try {
      const response = await api.delete<WishlistResponse>(`/wishlist/${productId}`);
      
      if (response.data) {
        setWishlist(response.data.items || []);
        setWishlistCount(response.data.count || 0);
        showToast("Removed from wishlist", "success");
      }
    } catch (error: any) {
      console.error("Failed to remove from wishlist:", error);
      const message = error.response?.data?.message || "Failed to remove from wishlist";
      showToast(message, "error");
    } finally {
      setIsLoading(false);
    }
  }, [isAuthenticated, showToast]);

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
      
      // Trigger cart refresh event
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

  // Check if product is in wishlist
  const isInWishlist = useCallback((productId: number): boolean => {
    return wishlist.some(item => item.product_id === productId);
  }, [wishlist]);

  return (
    <WishlistContext.Provider
      value={{
        wishlist,
        wishlistCount,
        isLoading,
        addToWishlist,
        removeFromWishlist,
        moveToCart,
        isInWishlist,
        refreshWishlist,
      }}
    >
      {children}
    </WishlistContext.Provider>
  );
}

export function useWishlist() {
  const context = useContext(WishlistContext);
  if (!context) {
    throw new Error("useWishlist must be used within WishlistProvider");
  }
  return context;
}
