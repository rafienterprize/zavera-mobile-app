"use client";

import { useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { ProductCategory } from "@/types";

export interface ProductFilters {
  size: string | null; // Changed from sizes: string[] to single size
  priceRange: { min: number; max: number } | null;
  subcategory: string | null;
}

interface FilterPanelProps {
  category: ProductCategory;
  onFilterChange: (filters: ProductFilters) => void;
  activeFilters: ProductFilters;
  productCount: number;
}

const SIZES = ["XS", "S", "M", "L", "XL", "XXL"];

// Mapping subcategory: Display Label (ID) -> Database Value (EN)
const SUBCATEGORY_MAPPING: Record<ProductCategory, Record<string, string>> = {
  wanita: {
    "Dress": "Dress",
    "Atasan": "Tops",
    "Bawahan": "Bottoms",
    "Outerwear": "Outerwear",
    "Aksesoris": "Accessories"
  },
  pria: {
    "Atasan": "Tops",
    "Kemeja": "Shirts",
    "Celana": "Bottoms",
    "Jaket": "Outerwear",
    "Jas": "Suits",
    "Sepatu": "Footwear"
  },
  anak: {
    "Anak Laki-laki": "Boys",
    "Anak Perempuan": "Girls",
    "Bayi": "Baby",
    "Sepatu": "Footwear"
  },
  sports: {
    "Pakaian Olahraga": "Activewear",
    "Sepatu": "Footwear",
    "Jaket": "Outerwear",
    "Aksesoris": "Accessories"
  },
  luxury: {
    "Aksesoris": "Accessories",
    "Outerwear": "Outerwear"
  },
  beauty: {
    "Perawatan Kulit": "Skincare",
    "Makeup": "Makeup",
    "Parfum": "Fragrance"
  }
};

const SUBCATEGORIES: Record<ProductCategory, string[]> = {
  wanita: ["Dress", "Atasan", "Bawahan", "Outerwear", "Aksesoris"],
  pria: ["Atasan", "Kemeja", "Celana", "Jaket", "Jas", "Sepatu"],
  anak: ["Anak Laki-laki", "Anak Perempuan", "Bayi", "Sepatu"],
  sports: ["Pakaian Olahraga", "Sepatu", "Jaket", "Aksesoris"],
  luxury: ["Aksesoris", "Outerwear"],
  beauty: ["Perawatan Kulit", "Makeup", "Parfum"],
};

const PRICE_RANGES = [
  { label: "Semua Harga", min: 0, max: Infinity },
  { label: "Di bawah Rp 100.000", min: 0, max: 100000 },
  { label: "Rp 100.000 - Rp 300.000", min: 100000, max: 300000 },
  { label: "Rp 300.000 - Rp 500.000", min: 300000, max: 500000 },
  { label: "Rp 500.000 - Rp 1.000.000", min: 500000, max: 1000000 },
  { label: "Di atas Rp 1.000.000", min: 1000000, max: Infinity },
];

interface FilterSectionProps {
  title: string;
  children: React.ReactNode;
  defaultOpen?: boolean;
}

function FilterSection({ title, children, defaultOpen = true }: FilterSectionProps) {
  const [isOpen, setIsOpen] = useState(defaultOpen);

  return (
    <div className="border-b border-gray-100 py-4">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center justify-between w-full text-left"
      >
        <span className="text-sm font-medium text-gray-900">{title}</span>
        <svg
          className={`w-4 h-4 text-gray-500 transition-transform ${isOpen ? "rotate-180" : ""}`}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </button>
      <AnimatePresence>
        {isOpen && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: "auto", opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.2 }}
            className="overflow-hidden"
          >
            <div className="pt-3">{children}</div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}

export default function FilterPanel({
  category,
  onFilterChange,
  activeFilters,
  productCount,
}: FilterPanelProps) {
  const subcategories = SUBCATEGORIES[category] || [];
  const subcategoryMapping = SUBCATEGORY_MAPPING[category] || {};

  const handleSizeChange = (size: string | null) => {
    // If clicking the same size, deselect it (set to null)
    // Otherwise, select the new size
    const newSize = activeFilters.size === size ? null : size;
    onFilterChange({ ...activeFilters, size: newSize });
  };

  const handlePriceChange = (min: number, max: number) => {
    if (min === 0 && max === Infinity) {
      onFilterChange({ ...activeFilters, priceRange: null });
    } else {
      onFilterChange({ ...activeFilters, priceRange: { min, max } });
    }
  };

  const handleSubcategoryChange = (displayLabel: string | null) => {
    // Convert display label (ID) to database value (EN)
    const dbValue = displayLabel ? subcategoryMapping[displayLabel] : null;
    onFilterChange({ ...activeFilters, subcategory: dbValue });
  };

  // Get display label from database value
  const getDisplayLabel = (dbValue: string | null): string | null => {
    if (!dbValue) return null;
    const entry = Object.entries(subcategoryMapping).find(([_, val]) => val === dbValue);
    return entry ? entry[0] : dbValue;
  };

  const clearAllFilters = () => {
    onFilterChange({ size: null, priceRange: null, subcategory: null });
  };

  const hasActiveFilters =
    activeFilters.size !== null ||
    activeFilters.priceRange !== null ||
    activeFilters.subcategory !== null;

  return (
    <div className="bg-white rounded-lg border border-gray-100 p-4">
      {/* Header */}
      <div className="flex items-center justify-between pb-4 border-b border-gray-100">
        <div>
          <h3 className="text-sm font-semibold text-gray-900">Filter</h3>
          <p className="text-xs text-gray-500 mt-0.5">{productCount} produk</p>
        </div>
        {hasActiveFilters && (
          <button
            onClick={clearAllFilters}
            className="text-xs text-primary hover:underline"
          >
            Hapus Semua
          </button>
        )}
      </div>

      {/* Subcategory Filter */}
      {subcategories.length > 0 && (
        <FilterSection title="Kategori">
          <div className="space-y-2">
            <label 
              className={`flex items-center gap-2 cursor-pointer px-3 py-2 rounded-lg transition-colors ${
                activeFilters.subcategory === null
                  ? 'bg-black text-white' 
                  : 'hover:bg-gray-50'
              }`}
            >
              <input
                type="radio"
                name="subcategory"
                checked={activeFilters.subcategory === null}
                onChange={() => handleSubcategoryChange(null)}
                className="hidden"
              />
              <span className={`text-sm ${
                activeFilters.subcategory === null 
                  ? 'text-white font-medium' 
                  : 'text-gray-600'
              }`}>
                Semua
              </span>
            </label>
            {subcategories.map((displayLabel) => {
              const dbValue = subcategoryMapping[displayLabel];
              const isSelected = activeFilters.subcategory === dbValue;
              return (
                <label 
                  key={displayLabel} 
                  className={`flex items-center gap-2 cursor-pointer px-3 py-2 rounded-lg transition-colors ${
                    isSelected 
                      ? 'bg-black text-white' 
                      : 'hover:bg-gray-50'
                  }`}
                >
                  <input
                    type="radio"
                    name="subcategory"
                    checked={isSelected}
                    onChange={() => handleSubcategoryChange(displayLabel)}
                    className="hidden"
                  />
                  <span className={`text-sm ${isSelected ? 'text-white font-medium' : 'text-gray-600'}`}>
                    {displayLabel}
                  </span>
                </label>
              );
            })}
          </div>
        </FilterSection>
      )}

      {/* Size Filter */}
      <FilterSection title="Ukuran">
        <div className="flex flex-wrap gap-2">
          {SIZES.map((size) => (
            <button
              key={size}
              onClick={() => handleSizeChange(size)}
              className={`px-3 py-1.5 text-sm border rounded transition-colors ${
                activeFilters.size === size
                  ? "border-primary bg-primary text-white"
                  : "border-gray-200 text-gray-600 hover:border-gray-400"
              }`}
            >
              {size}
            </button>
          ))}
        </div>
      </FilterSection>

      {/* Price Filter */}
      <FilterSection title="Harga">
        <div className="space-y-2">
          {PRICE_RANGES.map((range, index) => {
            const isSelected =
              (range.min === 0 && range.max === Infinity && activeFilters.priceRange === null) ||
              (activeFilters.priceRange?.min === range.min &&
                activeFilters.priceRange?.max === range.max);
            return (
              <label key={index} className="flex items-center gap-2 cursor-pointer group">
                <input
                  type="radio"
                  name="priceRange"
                  checked={isSelected}
                  onChange={() => handlePriceChange(range.min, range.max)}
                  className="w-4 h-4 text-primary border-gray-300 focus:ring-primary"
                />
                <span className="text-sm text-gray-600 group-hover:text-gray-900">
                  {range.label}
                </span>
              </label>
            );
          })}
        </div>
      </FilterSection>
    </div>
  );
}

// Active Filters Tags Component
interface ActiveFiltersProps {
  filters: ProductFilters;
  onRemoveFilter: (type: "size" | "price" | "subcategory", value?: string) => void;
  onClearAll: () => void;
  category: ProductCategory;
}

export function ActiveFilters({ filters, onRemoveFilter, onClearAll, category }: ActiveFiltersProps) {
  const hasFilters =
    filters.size !== null || filters.priceRange !== null || filters.subcategory !== null;

  if (!hasFilters) return null;

  const subcategoryMapping = SUBCATEGORY_MAPPING[category] || {};
  
  // Get display label from database value
  const getSubcategoryDisplayLabel = (dbValue: string): string => {
    const entry = Object.entries(subcategoryMapping).find(([_, val]) => val === dbValue);
    return entry ? entry[0] : dbValue;
  };

  const getPriceLabel = () => {
    if (!filters.priceRange) return null;
    const range = PRICE_RANGES.find(
      (r) => r.min === filters.priceRange?.min && r.max === filters.priceRange?.max
    );
    return range?.label || `Rp ${filters.priceRange.min.toLocaleString()} - Rp ${filters.priceRange.max.toLocaleString()}`;
  };

  return (
    <div className="flex flex-wrap items-center gap-2">
      <span className="text-sm text-gray-500">Filter aktif:</span>
      
      {filters.subcategory && (
        <button
          onClick={() => onRemoveFilter("subcategory")}
          className="inline-flex items-center gap-1 px-3 py-1 bg-gray-100 text-gray-700 text-sm rounded-full hover:bg-gray-200 transition-colors"
        >
          {getSubcategoryDisplayLabel(filters.subcategory)}
          <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      )}

      {filters.size && (
        <button
          onClick={() => onRemoveFilter("size")}
          className="inline-flex items-center gap-1 px-3 py-1 bg-gray-100 text-gray-700 text-sm rounded-full hover:bg-gray-200 transition-colors"
        >
          Ukuran: {filters.size}
          <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      )}

      {filters.priceRange && (
        <button
          onClick={() => onRemoveFilter("price")}
          className="inline-flex items-center gap-1 px-3 py-1 bg-gray-100 text-gray-700 text-sm rounded-full hover:bg-gray-200 transition-colors"
        >
          {getPriceLabel()}
          <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      )}

      <button
        onClick={onClearAll}
        className="text-sm text-primary hover:underline ml-2"
      >
        Hapus Semua
      </button>
    </div>
  );
}
