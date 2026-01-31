"use client";

import { motion } from "framer-motion";
import Link from "next/link";
import Image from "next/image";
import { useRouter } from "next/navigation";
import { Product } from "@/types";
import { useWishlist } from "@/context/WishlistContext";
import { useAuth } from "@/context/AuthContext";

interface ProductCardProps {
  product: Product;
  index: number;
  variant?: "default" | "luxury";
}

export default function ProductCard({ product, index, variant = "default" }: ProductCardProps) {
  const router = useRouter();
  const { isAuthenticated } = useAuth();
  const { addToWishlist, removeFromWishlist, isInWishlist } = useWishlist();
  
  const inWishlist = isInWishlist(product.id);

  // Redirect ke halaman detail produk untuk pilih ukuran & jumlah
  const handleQuickView = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    router.push(`/product/${product.id}`);
  };

  // Handle wishlist toggle
  const handleWishlistToggle = async (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    
    if (!isAuthenticated) {
      router.push("/login?redirect=/products");
      return;
    }

    if (inWishlist) {
      await removeFromWishlist(product.id);
    } else {
      await addToWishlist(product.id);
    }
  };

  const isLuxury = variant === "luxury" || product.category === "luxury";
  
  // Get primary image or first image from images array, fallback to image_url or placeholder
  const getProductImage = () => {
    console.log("Product data:", product);
    console.log("Product images:", product.images);
    console.log("Product image_url:", product.image_url);
    
    // If images array exists and has items, use first one
    if (product.images && product.images.length > 0) {
      const imageUrl = product.images[0];
      console.log("Using image from images array:", imageUrl);
      return imageUrl;
    }
    
    // Fallback to image_url or placeholder
    const fallbackUrl = product.image_url || "https://images.unsplash.com/photo-1441986300917-64674bd600d8?w=800&q=80";
    console.log("Using fallback image:", fallbackUrl);
    return fallbackUrl;
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true }}
      transition={{ duration: 0.5, delay: index * 0.1 }}
      className="group"
    >
      <Link href={`/product/${product.id}`}>
        <div className={`relative aspect-[3/4] overflow-hidden mb-4 rounded-lg ${isLuxury ? "ring-1 ring-amber-200" : ""}`}>
          <Image
            src={getProductImage()}
            alt={product.name}
            fill
            className="object-cover transition-transform duration-700 group-hover:scale-110"
          />

          {/* Hover Overlay */}
          <div className="absolute inset-0 bg-black/0 group-hover:bg-black/20 transition-colors duration-300" />

          {/* Quick Actions - Appears on Hover */}
          <div className="absolute bottom-0 left-0 right-0 p-4 opacity-0 group-hover:opacity-100 transition-all duration-300 translate-y-2 group-hover:translate-y-0">
            <button
              onClick={handleQuickView}
              className={`w-full py-3 font-medium text-sm tracking-wide transition-colors ${
                isLuxury
                  ? "bg-amber-500 text-white hover:bg-amber-600"
                  : "bg-white text-primary hover:bg-gray-100"
              }`}
            >
              LIHAT DETAIL
            </button>
          </div>

          {/* Badges */}
          <div className="absolute top-3 left-3 flex flex-col gap-2">
            {isLuxury && (
              <span className="px-2 py-1 bg-amber-500 text-white text-xs font-medium tracking-wider">
                LUXURY
              </span>
            )}
            {/* Low stock badge - only show for simple products (not variant-based) */}
            {product.stock > 0 && product.stock < 10 && (
              <span className="px-2 py-1 bg-red-500 text-white text-xs font-medium tracking-wider">
                SISA {product.stock}
              </span>
            )}
          </div>

          {/* REMOVED: SOLD OUT overlay for product cards
              Reason: For variant-based products, product.stock = 0 is normal
              User needs to click into product detail to see variant availability
          */}

          {/* Wishlist Button */}
          <button
            onClick={handleWishlistToggle}
            className={`absolute top-3 right-3 p-2 rounded-full opacity-0 group-hover:opacity-100 transition-all ${
              inWishlist 
                ? "bg-red-500 text-white hover:bg-red-600" 
                : "bg-white/90 text-gray-700 hover:bg-white"
            }`}
          >
            <svg 
              className="w-4 h-4" 
              fill={inWishlist ? "currentColor" : "none"} 
              stroke="currentColor" 
              viewBox="0 0 24 24"
            >
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
            </svg>
          </button>
        </div>

        <div className="space-y-1">
          {product.category && (
            <p className="text-xs text-gray-400 uppercase tracking-wider">
              {product.category}
            </p>
          )}
          <h3 className="font-medium text-sm tracking-wide group-hover:text-primary transition-colors line-clamp-2">
            {product.name}
          </h3>
          <p className={`text-sm font-semibold ${isLuxury ? "text-amber-600" : "text-gray-900"}`}>
            Rp {product.price.toLocaleString("id-ID")}
          </p>
        </div>
      </Link>
    </motion.div>
  );
}
