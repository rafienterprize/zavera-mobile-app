"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { motion } from "framer-motion";
import Link from "next/link";
import api from "@/lib/api";
import { Product } from "@/types";
import HeroCarousel from "@/components/HeroCarousel";
import CategoryGrid from "@/components/CategoryGrid";
import ProductCard from "@/components/ProductCard";
import { useAuth } from "@/context/AuthContext";

// Admin email from env
const ADMIN_EMAIL = process.env.NEXT_PUBLIC_ADMIN_EMAIL || "pemberani073@gmail.com";

export default function HomePage() {
  const router = useRouter();
  const { user, isLoading: authLoading } = useAuth();
  const [newArrivals, setNewArrivals] = useState<Product[]>([]);
  const [trendingProducts, setTrendingProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);

  // Redirect admin to dashboard
  useEffect(() => {
    if (!authLoading && user && user.email === ADMIN_EMAIL) {
      router.replace("/admin/dashboard");
    }
  }, [authLoading, user, router]);

  useEffect(() => {
    const fetchProducts = async () => {
      try {
        const response = await api.get("/products");
        const products = response.data || [];
        // Split products for different sections
        setNewArrivals(products.slice(0, 4));
        setTrendingProducts(products.slice(4, 8));
      } catch (error) {
        console.error("Failed to fetch products:", error);
        setNewArrivals([]);
        setTrendingProducts([]);
      } finally {
        setLoading(false);
      }
    };

    fetchProducts();
  }, []);

  return (
    <div className="bg-white">
      {/* Hero Carousel - Auto-sliding banner like Zalora */}
      <HeroCarousel />

      {/* Category Grid */}
      <CategoryGrid />

      {/* New Arrivals Section */}
      <section className="py-16 bg-white">
        <div className="max-w-7xl mx-auto px-6 lg:px-8">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.6 }}
            className="flex items-center justify-between mb-10"
          >
            <div>
              <h2 className="text-3xl md:text-4xl font-serif font-bold mb-2">
                New Arrivals
              </h2>
              <p className="text-gray-600">
                Koleksi terbaru yang baru saja hadir
              </p>
            </div>
            <Link
              href="/wanita"
              className="hidden md:inline-block px-6 py-3 border border-primary text-primary font-medium text-sm hover:bg-primary hover:text-white transition-colors"
            >
              LIHAT SEMUA
            </Link>
          </motion.div>

          {loading ? (
            <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
              {[...Array(4)].map((_, i) => (
                <div key={i} className="animate-pulse">
                  <div className="aspect-[3/4] bg-gray-100 rounded-lg mb-4" />
                  <div className="h-4 bg-gray-100 rounded mb-2 w-3/4" />
                  <div className="h-4 bg-gray-100 rounded w-1/2" />
                </div>
              ))}
            </div>
          ) : (
            <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
              {newArrivals.map((product, index) => (
                <ProductCard key={product.id} product={product} index={index} />
              ))}
            </div>
          )}

          <div className="md:hidden text-center mt-8">
            <Link
              href="/wanita"
              className="inline-block px-6 py-3 border border-primary text-primary font-medium text-sm hover:bg-primary hover:text-white transition-colors"
            >
              LIHAT SEMUA
            </Link>
          </div>
        </div>
      </section>

      {/* Banner Section */}
      <section className="py-8">
        <div className="max-w-7xl mx-auto px-6 lg:px-8">
          <div className="grid md:grid-cols-2 gap-6">
            <motion.div
              initial={{ opacity: 0, x: -20 }}
              whileInView={{ opacity: 1, x: 0 }}
              viewport={{ once: true }}
              className="relative h-[300px] md:h-[400px] rounded-xl overflow-hidden group"
            >
              <div
                className="absolute inset-0 bg-cover bg-center transition-transform duration-500 group-hover:scale-105"
                style={{
                  backgroundImage: "url('https://images.unsplash.com/photo-1469334031218-e382a71b716b?w=800&q=80')",
                }}
              />
              <div className="absolute inset-0 bg-black/40" />
              <div className="absolute inset-0 flex flex-col justify-end p-8">
                <span className="text-white/80 text-sm tracking-wider mb-2">KOLEKSI WANITA</span>
                <h3 className="text-white text-2xl md:text-3xl font-serif font-bold mb-4">
                  Elegant Style
                </h3>
                <Link
                  href="/wanita"
                  className="inline-block w-fit px-6 py-3 bg-white text-primary font-medium text-sm hover:bg-gray-100 transition-colors"
                >
                  SHOP NOW
                </Link>
              </div>
            </motion.div>

            <motion.div
              initial={{ opacity: 0, x: 20 }}
              whileInView={{ opacity: 1, x: 0 }}
              viewport={{ once: true }}
              className="relative h-[300px] md:h-[400px] rounded-xl overflow-hidden group"
            >
              <div
                className="absolute inset-0 bg-cover bg-center transition-transform duration-500 group-hover:scale-105"
                style={{
                  backgroundImage: "url('https://images.unsplash.com/photo-1507680434567-5739c80be1ac?w=800&q=80')",
                }}
              />
              <div className="absolute inset-0 bg-black/40" />
              <div className="absolute inset-0 flex flex-col justify-end p-8">
                <span className="text-white/80 text-sm tracking-wider mb-2">KOLEKSI PRIA</span>
                <h3 className="text-white text-2xl md:text-3xl font-serif font-bold mb-4">
                  Modern Gentleman
                </h3>
                <Link
                  href="/pria"
                  className="inline-block w-fit px-6 py-3 bg-white text-primary font-medium text-sm hover:bg-gray-100 transition-colors"
                >
                  SHOP NOW
                </Link>
              </div>
            </motion.div>
          </div>
        </div>
      </section>

      {/* Trending Section */}
      <section className="py-16 bg-gray-50">
        <div className="max-w-7xl mx-auto px-6 lg:px-8">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.6 }}
            className="flex items-center justify-between mb-10"
          >
            <div>
              <h2 className="text-3xl md:text-4xl font-serif font-bold mb-2">
                Trending Now
              </h2>
              <p className="text-gray-600">
                Produk paling diminati minggu ini
              </p>
            </div>
            <Link
              href="/pria"
              className="hidden md:inline-block px-6 py-3 border border-primary text-primary font-medium text-sm hover:bg-primary hover:text-white transition-colors"
            >
              LIHAT SEMUA
            </Link>
          </motion.div>

          {loading ? (
            <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
              {[...Array(4)].map((_, i) => (
                <div key={i} className="animate-pulse">
                  <div className="aspect-[3/4] bg-gray-200 rounded-lg mb-4" />
                  <div className="h-4 bg-gray-200 rounded mb-2 w-3/4" />
                  <div className="h-4 bg-gray-200 rounded w-1/2" />
                </div>
              ))}
            </div>
          ) : (
            <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
              {trendingProducts.map((product, index) => (
                <ProductCard key={product.id} product={product} index={index} />
              ))}
            </div>
          )}
        </div>
      </section>

      {/* Luxury Banner */}
      <section className="py-8">
        <div className="max-w-7xl mx-auto px-6 lg:px-8">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            className="relative h-[300px] md:h-[400px] rounded-xl overflow-hidden"
          >
            <div
              className="absolute inset-0 bg-cover bg-center"
              style={{
                backgroundImage: "url('https://images.unsplash.com/photo-1441986300917-64674bd600d8?w=1600&q=80')",
              }}
            />
            <div className="absolute inset-0 bg-gradient-to-r from-black/70 to-transparent" />
            <div className="absolute inset-0 flex items-center">
              <div className="max-w-xl px-8 md:px-12">
                <span className="inline-block px-3 py-1 bg-amber-500 text-white text-xs font-medium tracking-wider mb-4">
                  EXCLUSIVE
                </span>
                <h3 className="text-white text-3xl md:text-4xl font-serif font-bold mb-4">
                  Luxury Collection
                </h3>
                <p className="text-white/80 mb-6 max-w-md">
                  Koleksi eksklusif dari brand designer ternama dengan kualitas premium
                </p>
                <Link
                  href="/luxury"
                  className="inline-block px-8 py-4 bg-white text-primary font-medium hover:bg-gray-100 transition-colors"
                >
                  EXPLORE LUXURY
                </Link>
              </div>
            </div>
          </motion.div>
        </div>
      </section>

      {/* Newsletter Section */}
      <section className="py-20 bg-primary">
        <div className="max-w-3xl mx-auto px-6 text-center">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.6 }}
          >
            <h2 className="text-3xl md:text-4xl font-serif font-bold mb-4 text-white">
              Dapatkan Update Terbaru
            </h2>
            <p className="text-white/80 mb-8">
              Subscribe untuk mendapatkan info koleksi terbaru dan penawaran eksklusif
            </p>
            <div className="flex flex-col sm:flex-row gap-4 max-w-md mx-auto">
              <input
                type="email"
                placeholder="Masukkan email Anda"
                className="flex-1 px-6 py-4 rounded-lg focus:outline-none focus:ring-2 focus:ring-white"
              />
              <button className="px-8 py-4 bg-white text-primary font-medium rounded-lg hover:bg-gray-100 transition-colors">
                SUBSCRIBE
              </button>
            </div>
          </motion.div>
        </div>
      </section>
    </div>
  );
}
