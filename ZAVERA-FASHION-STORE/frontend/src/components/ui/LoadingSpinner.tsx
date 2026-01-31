"use client";

import { motion } from "framer-motion";

interface LoadingSpinnerProps {
  size?: "sm" | "md" | "lg";
  className?: string;
}

export default function LoadingSpinner({ size = "md", className = "" }: LoadingSpinnerProps) {
  const sizeClasses = {
    sm: "w-4 h-4",
    md: "w-6 h-6",
    lg: "w-10 h-10",
  };

  return (
    <motion.div
      className={`${sizeClasses[size]} ${className}`}
      animate={{ rotate: 360 }}
      transition={{ duration: 1, repeat: Infinity, ease: "linear" }}
    >
      <svg className="w-full h-full" viewBox="0 0 24 24" fill="none">
        <circle
          className="opacity-25"
          cx="12"
          cy="12"
          r="10"
          stroke="currentColor"
          strokeWidth="3"
        />
        <path
          className="opacity-75"
          fill="currentColor"
          d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
        />
      </svg>
    </motion.div>
  );
}

// Full page loading overlay
export function LoadingOverlay({ message = "Processing..." }: { message?: string }) {
  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
      className="fixed inset-0 z-[100] flex flex-col items-center justify-center bg-white/90 backdrop-blur-sm"
    >
      <LoadingSpinner size="lg" className="text-primary mb-4" />
      <p className="text-gray-600 font-medium">{message}</p>
    </motion.div>
  );
}

// Skeleton loader for products
export function ProductSkeleton() {
  return (
    <div className="animate-pulse">
      <div className="aspect-[3/4] bg-gray-200 rounded-lg mb-4" />
      <div className="h-4 bg-gray-200 rounded mb-2 w-3/4" />
      <div className="h-4 bg-gray-200 rounded w-1/2" />
    </div>
  );
}

// Skeleton loader for cart items
export function CartItemSkeleton() {
  return (
    <div className="flex gap-4 animate-pulse">
      <div className="w-24 h-24 bg-gray-200 rounded" />
      <div className="flex-1">
        <div className="h-4 bg-gray-200 rounded mb-2 w-3/4" />
        <div className="h-3 bg-gray-200 rounded mb-2 w-1/2" />
        <div className="h-4 bg-gray-200 rounded w-1/4" />
      </div>
    </div>
  );
}
