"use client";

import { useEffect } from "react";
import { motion, AnimatePresence } from "framer-motion";
import FilterPanel, { ProductFilters } from "./FilterPanel";
import { ProductCategory } from "@/types";

interface FilterDrawerProps {
  isOpen: boolean;
  onClose: () => void;
  category: ProductCategory;
  onFilterChange: (filters: ProductFilters) => void;
  activeFilters: ProductFilters;
  productCount: number;
}

export default function FilterDrawer({
  isOpen,
  onClose,
  category,
  onFilterChange,
  activeFilters,
  productCount,
}: FilterDrawerProps) {
  // Prevent body scroll when drawer is open
  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = "hidden";
    } else {
      document.body.style.overflow = "unset";
    }
    return () => {
      document.body.style.overflow = "unset";
    };
  }, [isOpen]);

  return (
    <AnimatePresence>
      {isOpen && (
        <>
          {/* Backdrop */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={onClose}
            className="fixed inset-0 bg-black/50 z-50 lg:hidden"
          />

          {/* Drawer */}
          <motion.div
            initial={{ x: "-100%" }}
            animate={{ x: 0 }}
            exit={{ x: "-100%" }}
            transition={{ type: "spring", damping: 25, stiffness: 300 }}
            className="fixed inset-y-0 left-0 w-[85%] max-w-sm bg-white z-50 lg:hidden shadow-xl"
          >
            {/* Header */}
            <div className="flex items-center justify-between p-4 border-b border-gray-100">
              <h2 className="text-lg font-semibold">Filter</h2>
              <button
                onClick={onClose}
                className="p-2 hover:bg-gray-100 rounded-full transition-colors"
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            {/* Filter Content */}
            <div className="overflow-y-auto h-[calc(100%-140px)] p-4">
              <FilterPanel
                category={category}
                onFilterChange={onFilterChange}
                activeFilters={activeFilters}
                productCount={productCount}
              />
            </div>

            {/* Footer */}
            <div className="absolute bottom-0 left-0 right-0 p-4 bg-white border-t border-gray-100">
              <button
                onClick={onClose}
                className="w-full py-3 bg-primary text-white font-medium rounded-lg hover:bg-gray-800 transition-colors"
              >
                Tampilkan {productCount} Produk
              </button>
            </div>
          </motion.div>
        </>
      )}
    </AnimatePresence>
  );
}
