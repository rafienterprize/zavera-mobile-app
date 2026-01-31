"use client";

import { useState, useEffect } from "react";
import { useParams, useRouter } from "next/navigation";
import Image from "next/image";
import { motion, AnimatePresence } from "framer-motion";
import api from "@/lib/api";
import { Product } from "@/types";
import { ProductVariant } from "@/types/variant";
import { variantApi } from "@/lib/variantApi";
import { useCart } from "@/context/CartContext";
import { useAuth } from "@/context/AuthContext";
import { useToast } from "@/components/ui/Toast";
import Button from "@/components/ui/Button";
import Breadcrumb, { getProductBreadcrumbs } from "@/components/Breadcrumb";
import VariantSelector from "@/components/VariantSelector";

// Low stock threshold
const LOW_STOCK_THRESHOLD = 10;

export default function ProductPage() {
  const params = useParams();
  const router = useRouter();
  const { addToCart } = useCart();
  const { isAuthenticated } = useAuth();
  const { showToast } = useToast();
  const [product, setProduct] = useState<Product | null>(null);
  const [variants, setVariants] = useState<ProductVariant[]>([]);
  const [variantsLoading, setVariantsLoading] = useState(true);
  const [selectedVariant, setSelectedVariant] = useState<ProductVariant | null>(null);
  const [quantity, setQuantity] = useState(1);
  const [loading, setLoading] = useState(true);
  const [addingToCart, setAddingToCart] = useState(false);
  const [showAddedModal, setShowAddedModal] = useState(false);
  const [showSizeGuide, setShowSizeGuide] = useState(false);
  const [selectedImageIndex, setSelectedImageIndex] = useState(0);
  const [productImages, setProductImages] = useState<string[]>([]);

  useEffect(() => {
    const fetchProduct = async () => {
      try {
        const response = await api.get(`/products/${params.id}`);
        const productData = response.data;
        console.log('🔍 Product data loaded:', productData);
        console.log('🏷️ Brand:', productData.brand);
        console.log('🧵 Material:', productData.material);
        setProduct(productData);
        
        // Parse images - handle both single image_url and images array
        let images: string[] = [];
        if (productData.images && Array.isArray(productData.images) && productData.images.length > 0) {
          images = productData.images;
        } else if (productData.image_url) {
          images = [productData.image_url];
        }
        setProductImages(images);
        
        // Fetch variants
        setVariantsLoading(true);
        try {
          const variantsData = await variantApi.getProductVariants(Number(params.id));
          console.log('Fetched variants:', variantsData);
          console.log('Variants count:', variantsData.length);
          setVariants(variantsData);
        } catch (error) {
          console.log('No variants found, using simple product:', error);
          setVariants([]);
        } finally {
          setVariantsLoading(false);
        }
      } catch (error) {
        console.error("Failed to fetch product:", error);
        showToast("Gagal memuat produk", "error");
      } finally {
        setLoading(false);
      }
    };

    fetchProduct();
  }, [params.id, showToast]);

  const handleAddToCart = async () => {
    if (!product) return;

    // Check if user is logged in
    if (!isAuthenticated) {
      showToast("Silakan login terlebih dahulu untuk menambahkan ke keranjang", "warning");
      router.push("/login");
      return;
    }

    // If product has variants, validate variant selection
    if (variants.length > 0 && !selectedVariant) {
      showToast("Silakan pilih varian terlebih dahulu", "warning");
      return;
    }
    
    setAddingToCart(true);
    
    // Simulate a small delay for better UX
    await new Promise(resolve => setTimeout(resolve, 300));
    
    // For variant products, pass variant_id to backend
    const cartItem: any = {
      ...product,
      quantity,
      selectedSize: selectedVariant?.size,
    };
    
    // Add variant_id if variant is selected
    if (selectedVariant) {
      cartItem.variant_id = selectedVariant.id;
    }
    
    addToCart(cartItem);
    
    setAddingToCart(false);
    setShowAddedModal(true);
    showToast(`${product.name} ditambahkan ke keranjang`, "success");
  };

  const getAvailableStock = (): number => {
    if (selectedVariant) {
      console.log('Selected variant stock:', selectedVariant.available_stock);
      return selectedVariant.available_stock || 0;
    }
    console.log('Product stock (no variant selected):', product?.stock);
    console.log('Has variants?', variants.length > 0);
    console.log('Variants loading?', variantsLoading);
    return product?.stock || 0;
  };

  const getCurrentPrice = (): number => {
    if (selectedVariant?.price) {
      return selectedVariant.price;
    }
    return product?.price || 0;
  };

  if (loading) {
    return (
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        {/* Breadcrumb skeleton */}
        <div className="h-4 bg-gray-100 rounded w-48 mb-8 animate-pulse" />
        
        <div className="grid grid-cols-1 md:grid-cols-2 gap-12">
          <div className="aspect-square bg-gray-100 animate-pulse rounded-lg" />
          <div className="space-y-4">
            <div className="h-4 bg-gray-100 rounded w-24 animate-pulse" />
            <div className="h-10 bg-gray-100 rounded w-3/4 animate-pulse" />
            <div className="h-8 bg-gray-100 rounded w-1/4 animate-pulse" />
            <div className="h-24 bg-gray-100 rounded animate-pulse" />
            <div className="h-12 bg-gray-100 rounded w-1/2 animate-pulse" />
          </div>
        </div>
      </div>
    );
  }

  if (!product) {
    return (
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-24">
        <div className="text-center">
          <div className="w-24 h-24 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-6">
            <svg className="w-12 h-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <h1 className="text-2xl font-bold mb-4">Produk Tidak Ditemukan</h1>
          <p className="text-gray-600 mb-8">Maaf, produk yang Anda cari tidak tersedia.</p>
          <Button onClick={() => router.push("/")}>
            Kembali ke Beranda
          </Button>
        </div>
      </div>
    );
  }

  const availableStock = getAvailableStock();
  const currentPrice = getCurrentPrice();
  const isLowStock = availableStock > 0 && availableStock < LOW_STOCK_THRESHOLD;
  const breadcrumbItems = getProductBreadcrumbs(product.name, product.category);

  return (
    <>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        {/* Breadcrumb Navigation */}
        <Breadcrumb items={breadcrumbItems} />

        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          className="grid grid-cols-1 md:grid-cols-2 gap-12"
        >
          {/* Product Images Gallery */}
          <motion.div
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: 0.1 }}
            className="space-y-4"
          >
            {/* Main Image */}
            <div className="aspect-square relative bg-gray-50 rounded-lg overflow-hidden group">
              <Image
                src={productImages[selectedImageIndex] || product.image_url || '/placeholder.jpg'}
                alt={`${product.name} - Image ${selectedImageIndex + 1}`}
                fill
                className="object-cover transition-transform duration-500 group-hover:scale-105"
                priority
              />
              
              {/* Low Stock Badge - Only show when variant selected and low stock */}
              {selectedVariant && isLowStock && (
                <div className="absolute top-4 left-4 px-3 py-1.5 bg-amber-500 text-white text-sm font-medium rounded-full flex items-center gap-1.5 z-10">
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  Sisa {availableStock}
                </div>
              )}

              {/* REMOVED: No more overlays on image - better UX */}

              {/* Category Badge */}
              {product.category && (
                <div className="absolute top-4 right-4 px-3 py-1 bg-white/90 backdrop-blur-sm text-gray-700 text-xs font-medium uppercase tracking-wider rounded z-10">
                  {product.category}
                </div>
              )}

              {/* Image Navigation Arrows */}
              {productImages.length > 1 && (
                <>
                  <button
                    onClick={() => setSelectedImageIndex((prev) => (prev === 0 ? productImages.length - 1 : prev - 1))}
                    className="absolute left-4 top-1/2 -translate-y-1/2 w-10 h-10 bg-white/90 backdrop-blur-sm rounded-full flex items-center justify-center hover:bg-white transition-all shadow-lg z-10"
                  >
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
                    </svg>
                  </button>
                  <button
                    onClick={() => setSelectedImageIndex((prev) => (prev === productImages.length - 1 ? 0 : prev + 1))}
                    className="absolute right-4 top-1/2 -translate-y-1/2 w-10 h-10 bg-white/90 backdrop-blur-sm rounded-full flex items-center justify-center hover:bg-white transition-all shadow-lg z-10"
                  >
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                    </svg>
                  </button>
                </>
              )}

              {/* Image Counter */}
              {productImages.length > 1 && (
                <div className="absolute bottom-4 right-4 px-3 py-1.5 bg-black/70 backdrop-blur-sm text-white text-sm font-medium rounded-full z-10">
                  {selectedImageIndex + 1} / {productImages.length}
                </div>
              )}
            </div>

            {/* Thumbnail Gallery */}
            {productImages.length > 1 && (
              <div className="grid grid-cols-5 gap-2">
                {productImages.map((image, index) => (
                  <button
                    key={index}
                    onClick={() => setSelectedImageIndex(index)}
                    className={`aspect-square relative bg-gray-50 rounded-lg overflow-hidden border-2 transition-all ${
                      selectedImageIndex === index
                        ? 'border-primary ring-2 ring-primary ring-offset-2'
                        : 'border-gray-200 hover:border-gray-400'
                    }`}
                  >
                    <Image
                      src={image}
                      alt={`${product.name} thumbnail ${index + 1}`}
                      fill
                      className="object-cover"
                    />
                  </button>
                ))}
              </div>
            )}
          </motion.div>

          {/* Product Details */}
          <motion.div
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: 0.2 }}
          >
            {/* Category */}
            {product.category && (
              <p className="text-sm text-gray-500 uppercase tracking-wider mb-2">
                {product.category}
              </p>
            )}

            <h1 className="text-3xl md:text-4xl font-serif font-bold mb-4">{product.name}</h1>
            <p className="text-2xl font-semibold text-primary mb-6">
              Rp {currentPrice.toLocaleString("id-ID")}
            </p>

            <div className="border-t border-b border-gray-200 py-6 mb-6">
              <p className="text-gray-600 leading-relaxed">{product.description}</p>
            </div>

            {/* Product Details - Brand & Material */}
            {(() => {
              console.log('🔍 Checking brand/material display:');
              console.log('  - product.brand:', product.brand);
              console.log('  - product.material:', product.material);
              console.log('  - Should show?', !!(product.brand || product.material));
              return null;
            })()}
            {(product.brand || product.material) && (
              <div className="mb-6 p-4 bg-gray-50 rounded-lg border border-gray-200">
                <h3 className="text-sm font-semibold text-gray-900 mb-3 flex items-center gap-2">
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  Detail Produk
                </h3>
                <div className="grid grid-cols-2 gap-4">
                  {product.brand && (
                    <div>
                      <p className="text-xs text-gray-500 mb-1">Brand</p>
                      <p className="text-sm font-medium text-gray-900">{product.brand}</p>
                    </div>
                  )}
                  {product.material && (
                    <div>
                      <p className="text-xs text-gray-500 mb-1">Material</p>
                      <p className="text-sm font-medium text-gray-900">{product.material}</p>
                    </div>
                  )}
                </div>
              </div>
            )}

            {/* Variant Selector - Dynamic from API */}
            {variants.length > 0 ? (
              <div className="mb-6">
                <div className="flex items-center justify-between mb-3">
                  <label className="block text-sm font-medium">Pilih Varian</label>
                  <button
                    onClick={() => setShowSizeGuide(true)}
                    className="flex items-center gap-1.5 text-sm text-primary hover:text-gray-800 transition-colors font-medium"
                  >
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 7h6m0 10v-3m-3 3h.01M9 17h.01M9 14h.01M12 14h.01M15 11h.01M12 11h.01M9 11h.01M7 21h10a2 2 0 002-2V5a2 2 0 00-2-2H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
                    </svg>
                    📏 Panduan Ukuran
                  </button>
                </div>
                <VariantSelector
                  productId={product.id}
                  variants={variants}
                  basePrice={product.price}
                  onVariantChange={setSelectedVariant}
                />
                {/* Message when no variant selected */}
                {!selectedVariant && (
                  <div className="mt-3 p-3 bg-blue-50 border border-blue-200 rounded-lg">
                    <p className="text-sm text-blue-800 flex items-center gap-2">
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                      Silakan pilih ukuran dan warna untuk melihat ketersediaan stok
                    </p>
                  </div>
                )}
              </div>
            ) : null}

            {/* Quantity Selector */}
            <div className="mb-8">
              <label className="block text-sm font-medium mb-3">Jumlah</label>
              <div className="flex items-center gap-4">
                <button
                  onClick={() => setQuantity(Math.max(1, quantity - 1))}
                  disabled={quantity <= 1}
                  className="w-12 h-12 border-2 border-gray-200 rounded-lg hover:border-primary transition-colors text-lg font-medium disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  −
                </button>
                <span className="w-12 text-center text-lg font-medium">{quantity}</span>
                <button
                  onClick={() => setQuantity(Math.min(availableStock, quantity + 1))}
                  disabled={quantity >= availableStock}
                  className="w-12 h-12 border-2 border-gray-200 rounded-lg hover:border-primary transition-colors text-lg font-medium disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  +
                </button>
              </div>
              <p className="text-sm text-gray-500 mt-2">
                {!variantsLoading && variants.length > 0 && !selectedVariant ? (
                  <span className="text-blue-600 font-medium">
                    Pilih varian untuk melihat stok
                  </span>
                ) : availableStock > 0 ? (
                  <>
                    <span className={isLowStock ? "text-amber-600 font-medium" : ""}>
                      {availableStock} item tersedia
                    </span>
                    {isLowStock && " - Segera habis!"}
                  </>
                ) : (
                  <span className="text-red-500">Stok habis</span>
                )}
              </p>
            </div>

            {/* Add to Cart Button */}
            <Button
              onClick={handleAddToCart}
              disabled={availableStock === 0 || (variants.length > 0 && !selectedVariant)}
              loading={addingToCart}
              fullWidth
              size="lg"
            >
              {variants.length > 0 && !selectedVariant 
                ? "Pilih Varian Terlebih Dahulu"
                : availableStock === 0 
                ? "Stok Habis" 
                : "Tambah ke Keranjang"}
            </Button>

            {/* Trust Badges */}
            <div className="mt-8 pt-6 border-t">
              <div className="grid grid-cols-3 gap-4 text-center text-sm text-gray-500">
                <div className="flex flex-col items-center gap-2">
                  <div className="w-10 h-10 bg-gray-50 rounded-full flex items-center justify-center">
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" />
                    </svg>
                  </div>
                  <span>Gratis Ongkir</span>
                </div>
                <div className="flex flex-col items-center gap-2">
                  <div className="w-10 h-10 bg-gray-50 rounded-full flex items-center justify-center">
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                    </svg>
                  </div>
                  <span>Pembayaran Aman</span>
                </div>
                <div className="flex flex-col items-center gap-2">
                  <div className="w-10 h-10 bg-gray-50 rounded-full flex items-center justify-center">
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                    </svg>
                  </div>
                  <span>Mudah Dikembalikan</span>
                </div>
              </div>
            </div>
          </motion.div>
        </motion.div>
      </div>

      {/* Added to Cart Modal */}
      <AnimatePresence>
        {showAddedModal && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 z-50 flex items-center justify-center p-4"
          >
            <div className="absolute inset-0 bg-black/50 backdrop-blur-sm" onClick={() => setShowAddedModal(false)} />
            <motion.div
              initial={{ scale: 0.95, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              exit={{ scale: 0.95, opacity: 0 }}
              className="relative bg-white rounded-xl shadow-2xl max-w-md w-full p-6"
            >
              <div className="text-center mb-6">
                <div className="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
                  <svg className="w-8 h-8 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                  </svg>
                </div>
                <h3 className="text-xl font-bold mb-2">Berhasil Ditambahkan!</h3>
                <p className="text-gray-600">{product.name}</p>
                {selectedVariant && (
                  <p className="text-sm text-gray-500">
                    {selectedVariant.size && `Ukuran: ${selectedVariant.size}`}
                    {selectedVariant.color && ` - ${selectedVariant.color}`}
                    {` × ${quantity}`}
                  </p>
                )}
              </div>
              <div className="flex gap-3">
                <button
                  onClick={() => setShowAddedModal(false)}
                  className="flex-1 px-4 py-3 border-2 border-gray-200 font-medium rounded-lg hover:border-primary transition-colors"
                >
                  Lanjut Belanja
                </button>
                <button
                  onClick={() => router.push("/cart")}
                  className="flex-1 px-4 py-3 bg-primary text-white font-medium rounded-lg hover:bg-gray-800 transition-colors"
                >
                  Lihat Keranjang
                </button>
              </div>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* Size Guide Modal */}
      <AnimatePresence>
        {showSizeGuide && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 z-50 flex items-center justify-center p-4 overflow-y-auto"
          >
            <div className="absolute inset-0 bg-black/50 backdrop-blur-sm" onClick={() => setShowSizeGuide(false)} />
            <motion.div
              initial={{ scale: 0.95, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              exit={{ scale: 0.95, opacity: 0 }}
              className="relative bg-white rounded-xl shadow-2xl max-w-2xl w-full max-h-[90vh] overflow-y-auto"
            >
              {/* Header */}
              <div className="sticky top-0 bg-white border-b px-6 py-4 flex items-center justify-between z-10">
                <h3 className="text-xl font-bold flex items-center gap-2">
                  <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 7h6m0 10v-3m-3 3h.01M9 17h.01M9 14h.01M12 14h.01M15 11h.01M12 11h.01M9 11h.01M7 21h10a2 2 0 002-2V5a2 2 0 00-2-2H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
                  </svg>
                  Panduan Ukuran
                </h3>
                <button
                  onClick={() => setShowSizeGuide(false)}
                  className="w-8 h-8 flex items-center justify-center rounded-full hover:bg-gray-100 transition-colors"
                >
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>

              <div className="p-6 space-y-6">
                {/* Product Dimensions */}
                {selectedVariant && (
                  <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                    <h4 className="font-semibold text-gray-900 mb-3 flex items-center gap-2">
                      <svg className="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
                      </svg>
                      Dimensi Produk Terpilih
                    </h4>
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <p className="text-sm text-gray-600 mb-1">Ukuran (P × L × T)</p>
                        <p className="text-lg font-semibold text-gray-900">
                          {selectedVariant.length_cm || 30} × {selectedVariant.width_cm || 20} × {selectedVariant.height_cm || 5} cm
                        </p>
                      </div>
                      <div>
                        <p className="text-sm text-gray-600 mb-1">Berat</p>
                        <p className="text-lg font-semibold text-gray-900">
                          {selectedVariant.weight_grams ? (selectedVariant.weight_grams / 1000).toFixed(1) : '0.4'} kg
                        </p>
                      </div>
                    </div>
                    <p className="text-xs text-gray-500 mt-3">
                      💡 Dimensi ini digunakan untuk menghitung biaya pengiriman
                    </p>
                  </div>
                )}

                {/* Size Chart */}
                <div>
                  <h4 className="font-semibold text-gray-900 mb-3 flex items-center gap-2">
                    <svg className="w-5 h-5 text-gray-700" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h18M3 14h18m-9-4v8m-7 0h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
                    </svg>
                    Tabel Ukuran
                  </h4>
                  <div className="overflow-x-auto">
                    <table className="w-full text-sm border-collapse">
                      <thead>
                        <tr className="bg-gray-100">
                          <th className="border border-gray-300 px-4 py-2 text-left font-semibold">Ukuran</th>
                          <th className="border border-gray-300 px-4 py-2 text-center font-semibold">Lingkar Dada (cm)</th>
                          <th className="border border-gray-300 px-4 py-2 text-center font-semibold">Panjang Badan (cm)</th>
                          <th className="border border-gray-300 px-4 py-2 text-center font-semibold">Lebar Bahu (cm)</th>
                        </tr>
                      </thead>
                      <tbody>
                        <tr className="hover:bg-gray-50">
                          <td className="border border-gray-300 px-4 py-2 font-medium">S</td>
                          <td className="border border-gray-300 px-4 py-2 text-center">88-92</td>
                          <td className="border border-gray-300 px-4 py-2 text-center">68-70</td>
                          <td className="border border-gray-300 px-4 py-2 text-center">42-44</td>
                        </tr>
                        <tr className="hover:bg-gray-50">
                          <td className="border border-gray-300 px-4 py-2 font-medium">M</td>
                          <td className="border border-gray-300 px-4 py-2 text-center">92-96</td>
                          <td className="border border-gray-300 px-4 py-2 text-center">70-72</td>
                          <td className="border border-gray-300 px-4 py-2 text-center">44-46</td>
                        </tr>
                        <tr className="hover:bg-gray-50">
                          <td className="border border-gray-300 px-4 py-2 font-medium">L</td>
                          <td className="border border-gray-300 px-4 py-2 text-center">96-100</td>
                          <td className="border border-gray-300 px-4 py-2 text-center">72-74</td>
                          <td className="border border-gray-300 px-4 py-2 text-center">46-48</td>
                        </tr>
                        <tr className="hover:bg-gray-50">
                          <td className="border border-gray-300 px-4 py-2 font-medium">XL</td>
                          <td className="border border-gray-300 px-4 py-2 text-center">100-104</td>
                          <td className="border border-gray-300 px-4 py-2 text-center">74-76</td>
                          <td className="border border-gray-300 px-4 py-2 text-center">48-50</td>
                        </tr>
                        <tr className="hover:bg-gray-50">
                          <td className="border border-gray-300 px-4 py-2 font-medium">XXL</td>
                          <td className="border border-gray-300 px-4 py-2 text-center">104-108</td>
                          <td className="border border-gray-300 px-4 py-2 text-center">76-78</td>
                          <td className="border border-gray-300 px-4 py-2 text-center">50-52</td>
                        </tr>
                      </tbody>
                    </table>
                  </div>
                  <p className="text-xs text-gray-500 mt-2">
                    📐 Ukuran dapat bervariasi ±2cm tergantung pada bahan dan model
                  </p>
                </div>

                {/* Fit Guide */}
                <div>
                  <h4 className="font-semibold text-gray-900 mb-3 flex items-center gap-2">
                    <svg className="w-5 h-5 text-gray-700" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                    </svg>
                    Panduan Fit
                  </h4>
                  <div className="space-y-3">
                    <div className="flex gap-3 p-3 bg-gray-50 rounded-lg">
                      <div className="flex-shrink-0 w-10 h-10 bg-white rounded-full flex items-center justify-center font-bold text-gray-700">
                        S
                      </div>
                      <div>
                        <p className="font-medium text-gray-900">Slim Fit</p>
                        <p className="text-sm text-gray-600">Pas di badan, mengikuti lekuk tubuh dengan ketat</p>
                      </div>
                    </div>
                    <div className="flex gap-3 p-3 bg-gray-50 rounded-lg">
                      <div className="flex-shrink-0 w-10 h-10 bg-white rounded-full flex items-center justify-center font-bold text-gray-700">
                        R
                      </div>
                      <div>
                        <p className="font-medium text-gray-900">Regular Fit</p>
                        <p className="text-sm text-gray-600">Pas di badan dengan ruang gerak yang nyaman</p>
                      </div>
                    </div>
                    <div className="flex gap-3 p-3 bg-gray-50 rounded-lg">
                      <div className="flex-shrink-0 w-10 h-10 bg-white rounded-full flex items-center justify-center font-bold text-gray-700">
                        O
                      </div>
                      <div>
                        <p className="font-medium text-gray-900">Oversized Fit</p>
                        <p className="text-sm text-gray-600">Longgar dan lebar, memberikan tampilan kasual</p>
                      </div>
                    </div>
                  </div>
                </div>

                {/* Care Instructions */}
                <div>
                  <h4 className="font-semibold text-gray-900 mb-3 flex items-center gap-2">
                    <svg className="w-5 h-5 text-gray-700" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                    Petunjuk Perawatan
                  </h4>
                  <div className="grid grid-cols-2 gap-3">
                    <div className="flex items-start gap-2 text-sm">
                      <span className="text-lg">🧺</span>
                      <div>
                        <p className="font-medium text-gray-900">Cuci dengan Tangan</p>
                        <p className="text-gray-600">Atau mesin cuci mode gentle</p>
                      </div>
                    </div>
                    <div className="flex items-start gap-2 text-sm">
                      <span className="text-lg">🌡️</span>
                      <div>
                        <p className="font-medium text-gray-900">Air Dingin</p>
                        <p className="text-gray-600">Maksimal 30°C</p>
                      </div>
                    </div>
                    <div className="flex items-start gap-2 text-sm">
                      <span className="text-lg">🚫</span>
                      <div>
                        <p className="font-medium text-gray-900">Jangan Bleach</p>
                        <p className="text-gray-600">Hindari pemutih</p>
                      </div>
                    </div>
                    <div className="flex items-start gap-2 text-sm">
                      <span className="text-lg">👕</span>
                      <div>
                        <p className="font-medium text-gray-900">Jemur Terbalik</p>
                        <p className="text-gray-600">Hindari sinar matahari langsung</p>
                      </div>
                    </div>
                  </div>
                </div>

                {/* Measurement Tips */}
                <div className="bg-amber-50 border border-amber-200 rounded-lg p-4">
                  <h4 className="font-semibold text-gray-900 mb-2 flex items-center gap-2">
                    <svg className="w-5 h-5 text-amber-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
                    </svg>
                    Tips Mengukur
                  </h4>
                  <ul className="text-sm text-gray-700 space-y-1">
                    <li>• Gunakan pita pengukur yang fleksibel</li>
                    <li>• Ukur langsung pada tubuh, bukan pada pakaian</li>
                    <li>• Berdiri tegak dengan posisi rileks</li>
                    <li>• Jika ragu antara 2 ukuran, pilih yang lebih besar</li>
                  </ul>
                </div>
              </div>

              {/* Footer */}
              <div className="sticky bottom-0 bg-gray-50 border-t px-6 py-4">
                <button
                  onClick={() => setShowSizeGuide(false)}
                  className="w-full px-4 py-3 bg-primary text-white font-medium rounded-lg hover:bg-gray-800 transition-colors"
                >
                  Tutup
                </button>
              </div>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>
    </>
  );
}
