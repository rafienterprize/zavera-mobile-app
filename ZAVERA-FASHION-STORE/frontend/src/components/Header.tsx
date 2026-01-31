"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useCart } from "@/context/CartContext";
import { useWishlist } from "@/context/WishlistContext";
import { useAuth } from "@/context/AuthContext";
import { motion, AnimatePresence } from "framer-motion";
import { useState, useEffect, useRef } from "react";

const categories = [
  {
    name: "WANITA",
    href: "/wanita",
    subcategories: ["Dress", "Tops", "Bottoms", "Outerwear"],
  },
  {
    name: "PRIA",
    href: "/pria",
    subcategories: ["Shirts", "Pants", "Jackets", "Suits"],
  },
  {
    name: "SPORTS",
    href: "/sports",
    subcategories: ["Activewear", "Footwear", "Accessories"],
  },
  {
    name: "ANAK",
    href: "/anak",
    subcategories: ["Boys", "Girls", "Baby"],
  },
  {
    name: "LUXURY",
    href: "/luxury",
    subcategories: ["Designer", "Premium", "Limited Edition"],
  },
  {
    name: "BEAUTY",
    href: "/beauty",
    subcategories: ["Skincare", "Makeup", "Fragrance"],
  },
];

export default function Header() {
  const { getTotalItems } = useCart();
  const { wishlistCount } = useWishlist();
  const { user, isAuthenticated, logout } = useAuth();
  const pathname = usePathname();
  const [mounted, setMounted] = useState(false);
  const [isScrolled, setIsScrolled] = useState(false);
  const [activeMenu, setActiveMenu] = useState<string | null>(null);
  const [searchOpen, setSearchOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [userMenuOpen, setUserMenuOpen] = useState(false);
  const userMenuRef = useRef<HTMLDivElement>(null);

  // Check if we're on homepage
  const isHomePage = pathname === "/";

  // Handle hydration
  useEffect(() => {
    setMounted(true);
  }, []);

  const itemCount = mounted ? getTotalItems() : 0;

  useEffect(() => {
    const handleScroll = () => {
      setIsScrolled(window.scrollY > 50);
    };
    window.addEventListener("scroll", handleScroll);
    return () => window.removeEventListener("scroll", handleScroll);
  }, []);

  // Close user menu when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (userMenuRef.current && !userMenuRef.current.contains(event.target as Node)) {
        setUserMenuOpen(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  // Determine if header should be transparent (only on homepage and not scrolled)
  const isTransparent = isHomePage && !isScrolled;

  const handleLogout = () => {
    logout();
    setUserMenuOpen(false);
  };

  return (
    <>
      {/* Top Banner - Fixed */}
      <div className="fixed top-0 left-0 right-0 z-50 bg-primary text-white text-center py-2 text-xs tracking-wider">
        FREE SHIPPING FOR ORDERS OVER Rp 500.000 | USE CODE: ZAVERA2024
      </div>

      <motion.header
        className={`fixed top-8 left-0 right-0 z-50 transition-all duration-500 ${
          isTransparent
            ? "bg-black/20 backdrop-blur-sm"
            : "bg-white/98 backdrop-blur-xl shadow-lg"
        }`}
        initial={{ y: -100 }}
        animate={{ y: 0 }}
        transition={{ duration: 0.6 }}
      >
        <nav className="max-w-7xl mx-auto px-6 lg:px-8">
          {/* Main Header Row */}
          <div className={`flex justify-between items-center h-16 transition-colors duration-500 ${
            isTransparent ? "border-b border-white/20" : "border-b border-gray-100"
          }`}>
            {/* Logo */}
            <Link
              href="/"
              className={`text-2xl font-serif font-bold tracking-[0.2em] hover:opacity-80 transition-all duration-500 ${
                isTransparent ? "text-white" : "text-gray-900"
              }`}
            >
              ZAVERA
            </Link>

            {/* Search Bar - Center */}
            <div className="hidden lg:flex flex-1 max-w-md mx-8">
              <div className="relative w-full">
                <input
                  type="text"
                  placeholder="Cari produk, tren, dan merek..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className={`w-full px-4 py-2 rounded-full text-sm focus:outline-none transition-all duration-500 ${
                    isTransparent 
                      ? "bg-white/20 backdrop-blur-sm border border-white/30 text-white placeholder-white/70 focus:bg-white/30 focus:border-white/50"
                      : "bg-gray-50 border border-gray-200 text-gray-900 placeholder-gray-500 focus:border-primary focus:ring-1 focus:ring-primary"
                  }`}
                />
                <button className={`absolute right-3 top-1/2 -translate-y-1/2 p-1 rounded-full transition-colors ${
                  isTransparent ? "hover:bg-white/20" : "hover:bg-gray-100"
                }`}>
                  <svg className={`w-4 h-4 transition-colors duration-500 ${isTransparent ? "text-white/80" : "text-gray-500"}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                  </svg>
                </button>
              </div>
            </div>

            {/* Right Icons */}
            <div className="flex items-center space-x-5">
              {/* Mobile Search */}
              <button
                className={`lg:hidden p-2 rounded-full transition-colors ${
                  isTransparent ? "hover:bg-white/20 text-white" : "hover:bg-gray-100 text-gray-700"
                }`}
                onClick={() => setSearchOpen(!searchOpen)}
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
              </button>

              {/* User Account */}
              <div className="relative" ref={userMenuRef}>
                <button 
                  onClick={() => setUserMenuOpen(!userMenuOpen)}
                  className={`p-2 rounded-full transition-colors ${
                    isTransparent ? "hover:bg-white/20 text-white" : "hover:bg-gray-100 text-gray-700"
                  }`}
                >
                  {isAuthenticated ? (
                    <div className="w-5 h-5 bg-primary rounded-full flex items-center justify-center">
                      <span className="text-xs text-white font-medium">
                        {user?.first_name?.charAt(0).toUpperCase()}
                      </span>
                    </div>
                  ) : (
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                    </svg>
                  )}
                </button>

                {/* User Dropdown Menu */}
                <AnimatePresence>
                  {userMenuOpen && (
                    <motion.div
                      initial={{ opacity: 0, y: 10 }}
                      animate={{ opacity: 1, y: 0 }}
                      exit={{ opacity: 0, y: 10 }}
                      className="absolute right-0 top-full mt-2 w-56 bg-white rounded-lg shadow-xl border border-gray-100 py-2 z-50"
                    >
                      {isAuthenticated ? (
                        <>
                          <div className="px-4 py-3 border-b border-gray-100">
                            <p className="text-sm font-medium text-gray-900">{user?.first_name}</p>
                            <p className="text-xs text-gray-500 truncate">{user?.email}</p>
                          </div>
                          <Link
                            href="/account/pembelian"
                            onClick={() => setUserMenuOpen(false)}
                            className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
                          >
                            <span className="flex items-center gap-2">
                              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z" />
                              </svg>
                              Pembelian Saya
                            </span>
                          </Link>
                          <Link
                            href="/account/addresses"
                            onClick={() => setUserMenuOpen(false)}
                            className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
                          >
                            <span className="flex items-center gap-2">
                              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                              </svg>
                              Daftar Alamat
                            </span>
                          </Link>
                          <button
                            onClick={handleLogout}
                            className="w-full text-left px-4 py-2 text-sm text-red-600 hover:bg-red-50 transition-colors"
                          >
                            <span className="flex items-center gap-2">
                              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
                              </svg>
                              Keluar
                            </span>
                          </button>
                        </>
                      ) : (
                        <>
                          <Link
                            href="/login"
                            onClick={() => setUserMenuOpen(false)}
                            className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
                          >
                            Masuk
                          </Link>
                          <Link
                            href="/register"
                            onClick={() => setUserMenuOpen(false)}
                            className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
                          >
                            Daftar
                          </Link>
                        </>
                      )}
                    </motion.div>
                  )}
                </AnimatePresence>
              </div>

              {/* Wishlist */}
              <Link href="/wishlist" className={`relative p-2 rounded-full transition-colors group ${
                isTransparent ? "hover:bg-white/20 text-white" : "hover:bg-gray-100 text-gray-700"
              }`}>
                <svg className="w-5 h-5 transition-transform group-hover:scale-110" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
                </svg>
                {wishlistCount > 0 && (
                  <motion.span
                    className="absolute -top-1 -right-1 bg-red-500 text-white text-xs font-medium rounded-full h-5 w-5 flex items-center justify-center"
                    initial={{ scale: 0 }}
                    animate={{ scale: 1 }}
                    transition={{ type: "spring", stiffness: 500, damping: 30 }}
                  >
                    {wishlistCount}
                  </motion.span>
                )}
              </Link>

              {/* Cart */}
              <Link href="/cart" className={`relative p-2 rounded-full transition-colors group ${
                isTransparent ? "hover:bg-white/20 text-white" : "hover:bg-gray-100 text-gray-700"
              }`}>
                <svg
                  className="w-5 h-5 transition-transform group-hover:scale-110"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={1.5}
                    d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z"
                  />
                </svg>
                {itemCount > 0 && (
                  <motion.span
                    className="absolute -top-1 -right-1 bg-primary text-white text-xs font-medium rounded-full h-5 w-5 flex items-center justify-center"
                    initial={{ scale: 0 }}
                    animate={{ scale: 1 }}
                    transition={{ type: "spring", stiffness: 500, damping: 30 }}
                  >
                    {itemCount}
                  </motion.span>
                )}
              </Link>
            </div>
          </div>

          {/* Category Navigation Row */}
          <div className="hidden md:flex items-center justify-center space-x-8 h-12">
            {categories.map((category) => (
              <div
                key={category.name}
                className="relative"
                onMouseEnter={() => setActiveMenu(category.name)}
                onMouseLeave={() => setActiveMenu(null)}
              >
                <Link
                  href={category.href}
                  className={`text-sm font-medium tracking-wider transition-colors py-3 ${
                    isTransparent
                      ? activeMenu === category.name
                        ? "text-white"
                        : "text-white/90 hover:text-white"
                      : activeMenu === category.name
                        ? "text-primary"
                        : "text-gray-700 hover:text-primary"
                  }`}
                >
                  {category.name}
                </Link>

                {/* Mega Menu Dropdown */}
                <AnimatePresence>
                  {activeMenu === category.name && (
                    <motion.div
                      initial={{ opacity: 0, y: 10 }}
                      animate={{ opacity: 1, y: 0 }}
                      exit={{ opacity: 0, y: 10 }}
                      transition={{ duration: 0.2 }}
                      className="absolute top-full left-1/2 -translate-x-1/2 pt-2"
                    >
                      <div className="bg-white shadow-xl rounded-lg border border-gray-100 py-4 px-6 min-w-[200px]">
                        <div className="space-y-2">
                          {category.subcategories.map((sub) => (
                            <Link
                              key={sub}
                              href={`${category.href}?sub=${sub.toLowerCase()}`}
                              className="block text-sm text-gray-600 hover:text-primary hover:pl-2 transition-all"
                            >
                              {sub}
                            </Link>
                          ))}
                        </div>
                        <div className="mt-4 pt-4 border-t border-gray-100">
                          <Link
                            href={category.href}
                            className="text-sm font-medium text-primary hover:underline"
                          >
                            Lihat Semua →
                          </Link>
                        </div>
                      </div>
                    </motion.div>
                  )}
                </AnimatePresence>
              </div>
            ))}
          </div>
        </nav>

        {/* Mobile Search Dropdown */}
        <AnimatePresence>
          {searchOpen && (
            <motion.div
              initial={{ height: 0, opacity: 0 }}
              animate={{ height: "auto", opacity: 1 }}
              exit={{ height: 0, opacity: 0 }}
              className={`lg:hidden overflow-hidden ${
                isTransparent ? "border-t border-white/20 bg-black/30 backdrop-blur-md" : "border-t border-gray-100 bg-white"
              }`}
            >
              <div className="p-4">
                <input
                  type="text"
                  placeholder="Cari produk..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className={`w-full px-4 py-3 rounded-lg text-sm focus:outline-none ${
                    isTransparent 
                      ? "bg-white/20 border border-white/30 text-white placeholder-white/70"
                      : "bg-gray-50 border border-gray-200 text-gray-900 focus:border-primary"
                  }`}
                  autoFocus
                />
              </div>
            </motion.div>
          )}
        </AnimatePresence>

        {/* Mobile Category Menu */}
        <div className={`md:hidden overflow-x-auto ${
          isTransparent ? "border-t border-white/20" : "border-t border-gray-100"
        }`}>
          <div className="flex space-x-6 px-4 py-3">
            {categories.map((category) => (
              <Link
                key={category.name}
                href={category.href}
                className={`text-xs font-medium tracking-wider whitespace-nowrap transition-colors ${
                  isTransparent ? "text-white/90 hover:text-white" : "text-gray-700 hover:text-primary"
                }`}
              >
                {category.name}
              </Link>
            ))}
          </div>
        </div>
      </motion.header>

      {/* Spacer for fixed header + banner (only needed on non-homepage) */}
      {!isHomePage && <div className="h-[140px] md:h-[136px]" />}
    </>
  );
}
