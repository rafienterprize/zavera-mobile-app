"use client";

import { motion } from "framer-motion";
import Link from "next/link";
import Image from "next/image";

const categories = [
  {
    name: "Wanita",
    href: "/wanita",
    image: "https://images.unsplash.com/photo-1483985988355-763728e1935b?w=800&q=80",
    description: "Koleksi Elegan",
  },
  {
    name: "Pria",
    href: "/pria",
    image: "https://images.unsplash.com/photo-1617137968427-85924c800a22?w=800&q=80",
    description: "Gaya Maskulin",
  },
  {
    name: "Sports",
    href: "/sports",
    image: "https://images.unsplash.com/photo-1571019613454-1cb2f99b2d8b?w=800&q=80",
    description: "Activewear Premium",
  },
  {
    name: "Anak",
    href: "/anak",
    image: "https://images.unsplash.com/photo-1519238263530-99bdd11df2ea?w=800&q=80",
    description: "Fashion Si Kecil",
  },
  {
    name: "Luxury",
    href: "/luxury",
    image: "https://images.unsplash.com/photo-1584917865442-de89df76afd3?w=800&q=80",
    description: "Koleksi Eksklusif",
  },
  {
    name: "Beauty",
    href: "/beauty",
    image: "https://images.unsplash.com/photo-1596462502278-27bfdc403348?w=800&q=80",
    description: "Perawatan Premium",
  },
];

export default function CategoryGrid() {
  return (
    <section className="py-16 bg-gray-50">
      <div className="max-w-7xl mx-auto px-6 lg:px-8">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="text-center mb-12"
        >
          <h2 className="text-3xl md:text-4xl font-serif font-bold mb-3">
            Jelajahi Kategori
          </h2>
          <p className="text-gray-600 max-w-xl mx-auto">
            Temukan koleksi fashion terbaik untuk setiap gaya dan kebutuhan Anda
          </p>
        </motion.div>

        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
          {categories.map((category, index) => (
            <motion.div
              key={category.name}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: index * 0.1 }}
            >
              <Link
                href={category.href}
                className="group block relative aspect-[3/4] rounded-xl overflow-hidden"
              >
                <Image
                  src={category.image}
                  alt={category.name}
                  fill
                  className="object-cover transition-transform duration-500 group-hover:scale-110"
                />
                <div className="absolute inset-0 bg-gradient-to-t from-black/70 via-black/20 to-transparent" />
                <div className="absolute inset-0 flex flex-col items-center justify-end p-4 text-center">
                  <h3 className="text-white font-semibold text-lg mb-1">
                    {category.name}
                  </h3>
                  <p className="text-white/80 text-xs">
                    {category.description}
                  </p>
                </div>
                <div className="absolute inset-0 border-2 border-transparent group-hover:border-white/30 rounded-xl transition-colors" />
              </Link>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}
