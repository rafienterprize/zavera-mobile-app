"use client";

import { motion } from "framer-motion";
import Image from "next/image";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useWishlist } from "@/context/WishlistContext";
import { useAuth } from "@/context/AuthContext";
import { useEffect } from "react";

export default function WishlistPage() {
  const router = useRouter();
  const { isAuthenticated, isLoading: authLoading } = useAuth();
  const { wishlist, wishlistCount, isLoading, removeFromWishlist, moveToCart } = useWishlist();

  // Redirect to login if not authenticated
  useEffect(() => {
    if (!authLoading && !isAuthenticated) {
      router.push("/login?redirect=/wishlist");
    }
  }, [isAuthenticated, authLoading, router]);

  if (authLoading || !isAuthenticated) {
    return (
      <div className="min-h-screen bg-neutral-900 flex items-center justify-center">
        <div className="text-white text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-white mx-auto mb-4"></div>
          <p>Loading...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-neutral-900 text-white">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        {/* Header */}
        <motion.div
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          className="mb-8"
        >
          <h1 className="text-4xl font-bold mb-2">My Wishlist</h1>
          <p className="text-gray-400">
            {wishlistCount} {wishlistCount === 1 ? "item" : "items"} saved
          </p>
        </motion.div>

        {/* Empty State */}
        {wishlistCount === 0 && !isLoading && (
          <motion.div
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            className="text-center py-20"
          >
            <div className="mb-6">
              <svg
                className="w-24 h-24 mx-auto text-gray-600"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={1.5}
                  d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z"
                />
              </svg>
            </div>
            <h2 className="text-2xl font-semibold mb-2">Your wishlist is empty</h2>
            <p className="text-gray-400 mb-8">
              Save your favorite items to buy them later
            </p>
            <Link
              href="/"
              className="inline-block px-8 py-3 bg-white text-primary font-medium hover:bg-gray-100 transition-colors"
            >
              EXPLORE PRODUCTS
            </Link>
          </motion.div>
        )}

        {/* Loading State */}
        {isLoading && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {[...Array(4)].map((_, i) => (
              <div key={i} className="animate-pulse">
                <div className="bg-gray-800 aspect-[3/4] rounded-lg mb-4"></div>
                <div className="h-4 bg-gray-800 rounded mb-2"></div>
                <div className="h-4 bg-gray-800 rounded w-2/3"></div>
              </div>
            ))}
          </div>
        )}

        {/* Wishlist Grid */}
        {wishlistCount > 0 && !isLoading && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {wishlist.map((item, index) => (
              <motion.div
                key={item.id}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: index * 0.1 }}
                className="group relative"
              >
                <Link href={`/product/${item.product_id}`}>
                  <div className="relative aspect-[3/4] overflow-hidden rounded-lg mb-4">
                    <Image
                      src={item.product_image || "https://images.unsplash.com/photo-1441986300917-64674bd600d8?w=800&q=80"}
                      alt={item.product_name}
                      fill
                      className="object-cover transition-transform duration-700 group-hover:scale-110"
                    />
                    
                    {/* Overlay on hover */}
                    <div className="absolute inset-0 bg-black/0 group-hover:bg-black/20 transition-colors duration-300" />

                    {/* Stock badge */}
                    {!item.is_available && (
                      <div className="absolute inset-0 bg-black/50 flex items-center justify-center">
                        <span className="px-4 py-2 bg-white text-primary font-medium text-sm">
                          OUT OF STOCK
                        </span>
                      </div>
                    )}
                  </div>
                </Link>

                <div className="space-y-2">
                  <Link href={`/product/${item.product_id}`}>
                    <h3 className="font-medium text-sm tracking-wide group-hover:text-white transition-colors line-clamp-2">
                      {item.product_name}
                    </h3>
                  </Link>
                  <p className="text-sm font-semibold text-white">
                    Rp {item.product_price.toLocaleString("id-ID")}
                  </p>

                  {/* Action Buttons */}
                  <div className="flex gap-2 pt-2">
                    <button
                      onClick={() => moveToCart(item.product_id)}
                      disabled={!item.is_available}
                      className="flex-1 px-4 py-2 bg-white text-primary text-sm font-medium hover:bg-gray-100 transition-colors disabled:bg-gray-700 disabled:text-gray-400 disabled:cursor-not-allowed"
                    >
                      {item.is_available ? "MOVE TO CART" : "UNAVAILABLE"}
                    </button>
                    <button
                      onClick={() => removeFromWishlist(item.product_id)}
                      className="px-4 py-2 bg-neutral-800 text-white hover:bg-neutral-700 transition-colors border border-white/10"
                      title="Remove from wishlist"
                    >
                      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                      </svg>
                    </button>
                  </div>
                </div>
              </motion.div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
