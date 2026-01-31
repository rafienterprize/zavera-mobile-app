"use client";

import { useEffect, useState, useMemo } from "react";
import { motion } from "framer-motion";
import api from "@/lib/api";
import { Product, ProductCategory } from "@/types";
import ProductCard from "@/components/ProductCard";
import FilterPanel, { ProductFilters, ActiveFilters } from "@/components/FilterPanel";
import FilterDrawer from "@/components/FilterDrawer";

interface CategoryPageProps {
  category: ProductCategory;
  title: string;
  subtitle: string;
  bannerImage: string;
  accentColor?: string;
}

type SortOption = "newest" | "price-low" | "price-high" | "name";

export default function CategoryPage({
  category,
  title,
  subtitle,
  bannerImage,
}: CategoryPageProps) {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [sortBy, setSortBy] = useState<SortOption>("newest");
  const [viewMode, setViewMode] = useState<"grid" | "list">("grid");
  const [filters, setFilters] = useState<ProductFilters>({
    size: null,
    priceRange: null,
    subcategory: null,
  });
  const [filterDrawerOpen, setFilterDrawerOpen] = useState(false);

  useEffect(() => {
    const fetchProducts = async () => {
      try {
        const response = await api.get(`/products?category=${category}`);
        setProducts(response.data || []);
      } catch (error) {
        console.error("Failed to fetch products:", error);
        setProducts([]);
      } finally {
        setLoading(false);
      }
    };

    fetchProducts();
  }, [category]);

  // Filter and sort products
  const filteredAndSortedProducts = useMemo(() => {
    let result = [...products];

    // Apply subcategory filter
    if (filters.subcategory) {
      result = result.filter(
        (p) => p.subcategory?.toLowerCase() === filters.subcategory?.toLowerCase()
      );
    }

    // Apply price filter
    if (filters.priceRange) {
      result = result.filter(
        (p) =>
          p.price >= filters.priceRange!.min &&
          p.price <= filters.priceRange!.max
      );
    }

    // Apply size filter - only show products that have the selected size
    if (filters.size) {
      result = result.filter((p) => {
        // If product has available_sizes, check if the selected size is available
        if (p.available_sizes && p.available_sizes.length > 0) {
          return p.available_sizes.includes(filters.size!);
        }
        // If product doesn't have available_sizes, don't show it when size filter is active
        return false;
      });
    }

    // Apply sorting
    result.sort((a, b) => {
      switch (sortBy) {
        case "price-low":
          return a.price - b.price;
        case "price-high":
          return b.price - a.price;
        case "name":
          return a.name.localeCompare(b.name);
        default:
          return 0;
      }
    });

    return result;
  }, [products, filters, sortBy]);

  const handleFilterChange = (newFilters: ProductFilters) => {
    setFilters(newFilters);
  };

  const handleRemoveFilter = (type: "size" | "price" | "subcategory", value?: string) => {
    if (type === "size") {
      setFilters((prev) => ({ ...prev, size: null }));
    } else if (type === "price") {
      setFilters((prev) => ({ ...prev, priceRange: null }));
    } else if (type === "subcategory") {
      setFilters((prev) => ({ ...prev, subcategory: null }));
    }
  };

  const handleClearAllFilters = () => {
    setFilters({ size: null, priceRange: null, subcategory: null });
  };

  const hasActiveFilters =
    filters.size !== null || filters.priceRange !== null || filters.subcategory !== null;

  return (
    <div className="min-h-screen bg-white">
      {/* Hero Banner */}
      <motion.section
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ duration: 0.8 }}
        className="relative h-[40vh] md:h-[50vh] overflow-hidden"
      >
        <div
          className="absolute inset-0 bg-cover bg-center"
          style={{ backgroundImage: `url(${bannerImage})` }}
        />
        <div className="absolute inset-0 bg-gradient-to-t from-black/60 via-black/30 to-transparent" />
        <div className="absolute inset-0 flex items-end">
          <div className="max-w-7xl mx-auto px-6 lg:px-8 pb-12 w-full">
            <motion.div
              initial={{ y: 30, opacity: 0 }}
              animate={{ y: 0, opacity: 1 }}
              transition={{ delay: 0.3, duration: 0.6 }}
            >
              <h1 className="text-4xl md:text-6xl font-serif font-bold text-white mb-3">
                {title}
              </h1>
              <p className="text-lg md:text-xl text-white/90 max-w-xl">
                {subtitle}
              </p>
            </motion.div>
          </div>
        </div>
      </motion.section>

      {/* Filter & Sort Bar */}
      <section className="sticky top-[136px] z-40 bg-white border-b border-gray-100 shadow-sm">
        <div className="max-w-7xl mx-auto px-6 lg:px-8">
          <div className="flex items-center justify-between py-4">
            <div className="flex items-center space-x-4">
              {/* Mobile Filter Button */}
              <button
                onClick={() => setFilterDrawerOpen(true)}
                className="lg:hidden flex items-center gap-2 px-4 py-2 border border-gray-200 rounded-lg text-sm hover:border-gray-400 transition-colors"
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" />
                </svg>
                Filter
                {hasActiveFilters && (
                  <span className="w-5 h-5 bg-primary text-white text-xs rounded-full flex items-center justify-center">
                    {(filters.size ? 1 : 0) + (filters.priceRange ? 1 : 0) + (filters.subcategory ? 1 : 0)}
                  </span>
                )}
              </button>

              <span className="text-sm text-gray-500">
                {filteredAndSortedProducts.length} Produk
              </span>
            </div>

            <div className="flex items-center space-x-4">
              {/* Sort Dropdown */}
              <select
                value={sortBy}
                onChange={(e) => setSortBy(e.target.value as SortOption)}
                className="text-sm border border-gray-200 rounded-lg px-4 py-2 focus:outline-none focus:border-primary bg-white"
              >
                <option value="newest">Terbaru</option>
                <option value="price-low">Harga: Rendah ke Tinggi</option>
                <option value="price-high">Harga: Tinggi ke Rendah</option>
                <option value="name">Nama A-Z</option>
              </select>

              {/* View Mode Toggle */}
              <div className="hidden md:flex items-center border border-gray-200 rounded-lg overflow-hidden">
                <button
                  onClick={() => setViewMode("grid")}
                  className={`p-2 ${viewMode === "grid" ? "bg-gray-100" : "hover:bg-gray-50"}`}
                >
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
                  </svg>
                </button>
                <button
                  onClick={() => setViewMode("list")}
                  className={`p-2 ${viewMode === "list" ? "bg-gray-100" : "hover:bg-gray-50"}`}
                >
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M4 6h16M4 12h16M4 18h16" />
                  </svg>
                </button>
              </div>
            </div>
          </div>

          {/* Active Filters */}
          {hasActiveFilters && (
            <div className="pb-4">
              <ActiveFilters
                filters={filters}
                onRemoveFilter={handleRemoveFilter}
                onClearAll={handleClearAllFilters}
                category={category}
              />
            </div>
          )}
        </div>
      </section>

      {/* Main Content */}
      <section className="py-8">
        <div className="max-w-7xl mx-auto px-6 lg:px-8">
          <div className="flex gap-8">
            {/* Desktop Filter Sidebar */}
            <aside className="hidden lg:block w-64 flex-shrink-0">
              <div className="sticky top-[220px]">
                <FilterPanel
                  category={category}
                  onFilterChange={handleFilterChange}
                  activeFilters={filters}
                  productCount={filteredAndSortedProducts.length}
                />
              </div>
            </aside>

            {/* Products Grid */}
            <div className="flex-1">
              {loading ? (
                <div className={`grid gap-6 ${viewMode === "grid" ? "grid-cols-2 md:grid-cols-3" : "grid-cols-1"}`}>
                  {[...Array(8)].map((_, i) => (
                    <div key={i} className="animate-pulse">
                      <div className="aspect-[3/4] bg-gray-100 rounded-lg mb-4" />
                      <div className="h-4 bg-gray-100 rounded mb-2 w-3/4" />
                      <div className="h-4 bg-gray-100 rounded w-1/2" />
                    </div>
                  ))}
                </div>
              ) : filteredAndSortedProducts.length === 0 ? (
                <motion.div
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  className="text-center py-20"
                >
                  <div className="w-24 h-24 mx-auto mb-6 rounded-full bg-gray-100 flex items-center justify-center">
                    <svg className="w-12 h-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
                    </svg>
                  </div>
                  {hasActiveFilters ? (
                    <>
                      <h3 className="text-xl font-medium text-gray-900 mb-2">
                        Tidak Ada Produk Ditemukan
                      </h3>
                      <p className="text-gray-500 mb-8 max-w-md mx-auto">
                        Coba sesuaikan filter Anda untuk menemukan produk yang sesuai.
                      </p>
                      <button
                        onClick={handleClearAllFilters}
                        className="inline-block px-8 py-3 bg-primary text-white font-medium rounded-lg hover:bg-black transition-colors"
                      >
                        Hapus Semua Filter
                      </button>
                    </>
                  ) : (
                    <>
                      <h3 className="text-xl font-medium text-gray-900 mb-2">
                        Koleksi Segera Hadir
                      </h3>
                      <p className="text-gray-500 mb-8 max-w-md mx-auto">
                        Kami sedang menyiapkan koleksi terbaik untuk kategori ini. Jelajahi kategori lainnya atau kembali lagi nanti.
                      </p>
                      <a
                        href="/"
                        className="inline-block px-8 py-3 bg-primary text-white font-medium rounded-lg hover:bg-black transition-colors"
                      >
                        Jelajahi Koleksi Lain
                      </a>
                    </>
                  )}
                </motion.div>
              ) : (
                <motion.div
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  transition={{ duration: 0.5 }}
                  className={`grid gap-x-6 gap-y-10 ${
                    viewMode === "grid"
                      ? "grid-cols-2 md:grid-cols-3"
                      : "grid-cols-1 md:grid-cols-2"
                  }`}
                >
                  {filteredAndSortedProducts.map((product, index) => (
                    <ProductCard
                      key={product.id}
                      product={product}
                      index={index}
                      variant={category === "luxury" ? "luxury" : "default"}
                    />
                  ))}
                </motion.div>
              )}
            </div>
          </div>
        </div>
      </section>

      {/* Mobile Filter Drawer */}
      <FilterDrawer
        isOpen={filterDrawerOpen}
        onClose={() => setFilterDrawerOpen(false)}
        category={category}
        onFilterChange={handleFilterChange}
        activeFilters={filters}
        productCount={filteredAndSortedProducts.length}
      />
    </div>
  );
}
