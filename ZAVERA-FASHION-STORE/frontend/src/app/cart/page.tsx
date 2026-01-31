"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { useCart } from "@/context/CartContext";
import { useAuth } from "@/context/AuthContext";
import { useToast } from "@/components/ui/Toast";
import Image from "next/image";
import Link from "next/link";
import { motion, AnimatePresence } from "framer-motion";
import { ConfirmModal } from "@/components/ui/Modal";
import Button from "@/components/ui/Button";

export default function CartPage() {
  const { cart, removeFromCart, updateQuantity, getTotalPrice, clearCart, isLoading, validateCart, refreshCart } = useCart();
  const { isAuthenticated, isLoading: authLoading } = useAuth();
  const { showToast } = useToast();
  const router = useRouter();
  const [itemToRemove, setItemToRemove] = useState<{ id: number; name: string } | null>(null);
  const [showClearModal, setShowClearModal] = useState(false);
  const [cartChanges, setCartChanges] = useState<any[]>([]);
  const [showChangesNotification, setShowChangesNotification] = useState(false);

  // Redirect to login if not authenticated
  useEffect(() => {
    if (!authLoading && !isAuthenticated) {
      showToast("Silakan login untuk melihat keranjang", "warning");
      router.push("/login");
    }
  }, [isAuthenticated, authLoading, router, showToast]);

  // Auto-refresh cart every 10 seconds (faster validation on cart page)
  useEffect(() => {
    if (!isAuthenticated) return;

    const interval = setInterval(async () => {
      console.log("🔄 Auto-validating cart...");
      const validation = await validateCart();
      
      if (validation && !validation.valid && validation.changes.length > 0) {
        setCartChanges(validation.changes);
        setShowChangesNotification(true);
        
        // Refresh cart to get latest data
        await refreshCart();
        
        // Show toast for each change
        validation.changes.forEach((change: any) => {
          if (change.change_type === "price_changed") {
            showToast(
              `${change.product_name}: Price changed from Rp ${change.old_price?.toLocaleString()} to Rp ${change.new_price?.toLocaleString()}`,
              "warning"
            );
          } else if (change.change_type === "stock_insufficient") {
            showToast(
              `${change.product_name}: Only ${change.current_stock} items available`,
              "warning"
            );
          } else if (change.change_type === "product_unavailable") {
            showToast(
              `${change.product_name}: Product is no longer available`,
              "error"
            );
          }
        });
      }
    }, 10000); // 10 seconds - faster for better UX

    return () => clearInterval(interval);
  }, [isAuthenticated, validateCart, refreshCart, showToast]);

  const handleRemoveItem = (id: number, name: string) => {
    setItemToRemove({ id, name });
  };

  const confirmRemove = () => {
    if (itemToRemove) {
      removeFromCart(itemToRemove.id);
      showToast(`${itemToRemove.name} removed from cart`, "info");
      setItemToRemove(null);
    }
  };

  const handleClearCart = () => {
    clearCart();
    showToast("Cart cleared", "info");
    setShowClearModal(false);
  };

  // Show loading while checking auth
  if (authLoading || isLoading) {
    return (
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-24">
        <div className="text-center py-12">
          <div className="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
          <p className="text-gray-600">Memuat keranjang...</p>
        </div>
      </div>
    );
  }

  // Don't render if not authenticated (will redirect)
  if (!isAuthenticated) {
    return null;
  }

  if (cart.length === 0) {
    return (
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-24">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="text-center py-12"
        >
          <div className="w-24 h-24 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-6">
            <svg className="w-12 h-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z" />
            </svg>
          </div>
          <h1 className="text-3xl font-serif font-bold mb-4">Your Cart is Empty</h1>
          <p className="text-gray-600 mb-8 max-w-md mx-auto">
            Looks like you haven&apos;t added anything to your cart yet. Start exploring our collection!
          </p>
          <Link
            href="/"
            className="inline-block bg-primary text-white px-8 py-3 hover:bg-gray-800 transition"
          >
            Explore Collection
          </Link>
        </motion.div>
      </div>
    );
  }

  return (
    <>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-24">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
        >
          <div className="flex items-center justify-between mb-8">
            <h1 className="text-3xl font-serif font-bold">Shopping Cart</h1>
            <button
              onClick={() => setShowClearModal(true)}
              className="text-sm text-gray-500 hover:text-red-600 transition-colors"
            >
              Clear All
            </button>
          </div>

          {/* Changes Notification */}
          {showChangesNotification && cartChanges.length > 0 && (
            <motion.div
              initial={{ opacity: 0, y: -20 }}
              animate={{ opacity: 1, y: 0 }}
              className="mb-6 p-4 bg-yellow-50 border border-yellow-200 rounded-lg"
            >
              <div className="flex items-start gap-3">
                <svg className="w-5 h-5 text-yellow-600 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                </svg>
                <div className="flex-1">
                  <h3 className="font-medium text-yellow-900 mb-1">Cart Updated</h3>
                  <p className="text-sm text-yellow-800 mb-2">
                    Some items in your cart have been updated by the admin:
                  </p>
                  <ul className="text-sm text-yellow-800 space-y-1">
                    {cartChanges.map((change, idx) => (
                      <li key={idx}>• {change.message}</li>
                    ))}
                  </ul>
                </div>
                <button
                  onClick={() => setShowChangesNotification(false)}
                  className="text-yellow-600 hover:text-yellow-800"
                >
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
            </motion.div>
          )}

          <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
            {/* Cart Items */}
            <div className="lg:col-span-2 space-y-4">
              <AnimatePresence mode="popLayout">
                {cart.map((item, index) => (
                  <motion.div
                    key={`${item.id}-${item.selectedSize}`}
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, x: -100 }}
                    transition={{ delay: index * 0.05 }}
                    className="flex gap-4 p-4 bg-white border border-gray-100 rounded-lg hover:shadow-md transition-shadow"
                  >
                    <Link href={`/product/${item.id}`} className="w-24 h-24 relative bg-gray-50 flex-shrink-0 rounded overflow-hidden">
                      <Image
                        src={item.image_url || '/placeholder.jpg'}
                        alt={item.name}
                        fill
                        className="object-cover hover:scale-105 transition-transform"
                      />
                    </Link>

                    <div className="flex-1 min-w-0">
                      <Link href={`/product/${item.id}`}>
                        <h3 className="font-medium mb-1 hover:text-primary transition-colors truncate">
                          {item.name}
                        </h3>
                      </Link>
                      <p className="text-sm text-gray-500 mb-2">Size: {item.selectedSize}</p>
                      <p className="text-primary font-medium">
                        Rp {item.price.toLocaleString("id-ID")}
                      </p>

                      <div className="flex items-center gap-3 mt-3">
                        <button
                          onClick={() => updateQuantity(item.id, item.quantity - 1, item.selectedSize)}
                          className="w-8 h-8 border border-gray-200 rounded hover:border-primary hover:text-primary transition-colors flex items-center justify-center"
                        >
                          −
                        </button>
                        <span className="w-8 text-center font-medium">{item.quantity}</span>
                        <button
                          onClick={() => updateQuantity(item.id, item.quantity + 1, item.selectedSize)}
                          className="w-8 h-8 border border-gray-200 rounded hover:border-primary hover:text-primary transition-colors flex items-center justify-center"
                        >
                          +
                        </button>
                      </div>
                    </div>

                    <div className="text-right flex flex-col justify-between">
                      <p className="font-semibold">
                        Rp {(item.price * item.quantity).toLocaleString("id-ID")}
                      </p>
                      <button
                        onClick={() => handleRemoveItem(item.id, item.name)}
                        className="text-sm text-gray-400 hover:text-red-600 transition-colors"
                      >
                        Remove
                      </button>
                    </div>
                  </motion.div>
                ))}
              </AnimatePresence>

              {/* Continue Shopping Link */}
              <Link
                href="/"
                className="inline-flex items-center gap-2 text-sm text-gray-600 hover:text-primary transition-colors mt-4"
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
                </svg>
                Continue Shopping
              </Link>
            </div>

            {/* Order Summary */}
            <div className="lg:col-span-1">
              <div className="bg-gray-50 rounded-lg p-6 sticky top-24">
                <h2 className="text-xl font-semibold mb-6">Order Summary</h2>

                <div className="space-y-3 mb-6">
                  <div className="flex justify-between text-sm">
                    <span className="text-gray-600">Subtotal ({cart.reduce((acc, item) => acc + item.quantity, 0)} items)</span>
                    <span>Rp {getTotalPrice().toLocaleString("id-ID")}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-gray-600">Shipping</span>
                    <span className="text-gray-500">Calculated at checkout</span>
                  </div>
                </div>

                <div className="border-t pt-4 mb-6">
                  <div className="flex justify-between font-bold text-lg">
                    <span>Total</span>
                    <span>Rp {getTotalPrice().toLocaleString("id-ID")}</span>
                  </div>
                  <p className="text-xs text-gray-500 mt-1">Tax included</p>
                </div>

                <Link href="/checkout">
                  <Button fullWidth size="lg">
                    Proceed to Checkout
                  </Button>
                </Link>

                {/* Trust Badges */}
                <div className="mt-6 pt-4 border-t">
                  <div className="flex items-center justify-center gap-4 text-gray-400">
                    <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                    </svg>
                    <span className="text-xs">Secure Checkout</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </motion.div>
      </div>

      {/* Remove Item Modal */}
      <ConfirmModal
        isOpen={!!itemToRemove}
        onClose={() => setItemToRemove(null)}
        onConfirm={confirmRemove}
        title="Remove Item"
        message={`Are you sure you want to remove "${itemToRemove?.name}" from your cart?`}
        confirmText="Remove"
        variant="danger"
      />

      {/* Clear Cart Modal */}
      <ConfirmModal
        isOpen={showClearModal}
        onClose={() => setShowClearModal(false)}
        onConfirm={handleClearCart}
        title="Clear Cart"
        message="Are you sure you want to remove all items from your cart?"
        confirmText="Clear All"
        variant="danger"
      />
    </>
  );
}
