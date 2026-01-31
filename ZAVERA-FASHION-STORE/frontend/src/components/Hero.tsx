"use client";

import { motion } from "framer-motion";
import Link from "next/link";
import Image from "next/image";

export default function Hero() {
  return (
    <section className="relative h-screen w-full overflow-hidden -mt-8">
      {/* Background Image */}
      <div className="absolute inset-0">
        <Image
          src="https://images.unsplash.com/photo-1490481651871-ab68de25d43d?w=1920&q=80"
          alt="Fashion"
          fill
          className="object-cover"
          priority
        />
        <div className="absolute inset-0 bg-gradient-to-r from-black/70 via-black/40 to-transparent" />
      </div>

      {/* Content */}
      <div className="relative h-full flex items-center">
        <div className="max-w-7xl mx-auto px-6 lg:px-8 w-full">
          <motion.div
            initial={{ opacity: 0, x: -30 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.8, delay: 0.2 }}
            className="max-w-2xl"
          >
            <motion.span
              className="inline-block text-sm font-medium tracking-[0.3em] text-white/80 mb-4"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ duration: 0.8, delay: 0.3 }}
            >
              NEW COLLECTION 2026
            </motion.span>

            <motion.h1
              className="text-5xl md:text-6xl lg:text-7xl font-serif font-bold text-white mb-6 leading-tight"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.8, delay: 0.4 }}
            >
              Fashion
              <br />
              <span className="text-white/90">Starts Here</span>
            </motion.h1>

            <motion.p
              className="text-lg text-gray-200 mb-8 max-w-lg font-light leading-relaxed"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ duration: 0.8, delay: 0.6 }}
            >
              Personal style dimulai dari sini, mix & match item andalan biar makin percaya diri.
            </motion.p>

            <motion.div
              className="flex flex-col sm:flex-row items-start gap-4"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.8, delay: 0.8 }}
            >
              <Link
                href="/wanita"
                className="px-8 py-4 bg-white text-primary font-medium tracking-wide hover:bg-gray-100 transition-all duration-300"
              >
                BELANJA SEKARANG
              </Link>
              <Link
                href="/luxury"
                className="px-8 py-4 border-2 border-white text-white font-medium tracking-wide hover:bg-white hover:text-primary transition-all duration-300"
              >
                KOLEKSI LUXURY
              </Link>
            </motion.div>
          </motion.div>
        </div>
      </div>

      {/* Promo Banner */}
      <motion.div
        className="absolute bottom-0 left-0 right-0 bg-white/95 backdrop-blur-sm py-4"
        initial={{ y: 100 }}
        animate={{ y: 0 }}
        transition={{ duration: 0.6, delay: 1 }}
      >
        <div className="max-w-7xl mx-auto px-6 lg:px-8 flex items-center justify-between">
          <p className="text-sm text-gray-600">
            <span className="font-medium text-primary">Voucher selalu ada</span> - Cek & klaim yang tersedia
          </p>
          <Link href="/wanita" className="text-sm font-medium text-primary hover:underline">
            Lihat Pilihan →
          </Link>
        </div>
      </motion.div>
    </section>
  );
}
