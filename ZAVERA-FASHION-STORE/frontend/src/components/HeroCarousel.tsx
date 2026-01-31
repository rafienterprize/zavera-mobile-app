"use client";

import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import { ChevronLeft, ChevronRight } from "lucide-react";

interface Slide {
  id: number;
  image: string;
  title: string;
  subtitle: string;
  description: string;
  ctaText: string;
  ctaLink: string;
  textPosition: "left" | "center" | "right";
}

const slides: Slide[] = [
  {
    id: 1,
    image: "https://images.unsplash.com/photo-1483985988355-763728e1935b?w=1600&q=80",
    title: "Refresh Your Style",
    subtitle: "Up to 60% Off +",
    description: "Voucher 20%",
    ctaText: "SHOP NOW",
    ctaLink: "/wanita",
    textPosition: "right",
  },
  {
    id: 2,
    image: "https://images.unsplash.com/photo-1490481651871-ab68de25d43d?w=1600&q=80",
    title: "New Season Collection",
    subtitle: "Exclusive Deals",
    description: "Get up to 50% off on selected items",
    ctaText: "DISCOVER NOW",
    ctaLink: "/pria",
    textPosition: "left",
  },
  {
    id: 3,
    image: "https://images.unsplash.com/photo-1441984904996-e0b6ba687e04?w=1600&q=80",
    title: "Luxury Fashion",
    subtitle: "Premium Quality",
    description: "Designer brands at your fingertips",
    ctaText: "EXPLORE LUXURY",
    ctaLink: "/luxury",
    textPosition: "center",
  },
];

export default function HeroCarousel() {
  const [currentSlide, setCurrentSlide] = useState(0);
  const [isAutoPlaying, setIsAutoPlaying] = useState(true);

  const nextSlide = useCallback(() => {
    setCurrentSlide((prev) => (prev + 1) % slides.length);
  }, []);

  const prevSlide = useCallback(() => {
    setCurrentSlide((prev) => (prev - 1 + slides.length) % slides.length);
  }, []);

  const goToSlide = (index: number) => {
    setCurrentSlide(index);
    setIsAutoPlaying(false);
    // Resume auto-play after 5 seconds
    setTimeout(() => setIsAutoPlaying(true), 5000);
  };

  // Auto-play functionality
  useEffect(() => {
    if (!isAutoPlaying) return;

    const interval = setInterval(() => {
      nextSlide();
    }, 5000); // Change slide every 5 seconds

    return () => clearInterval(interval);
  }, [isAutoPlaying, nextSlide]);

  const getTextAlignment = (position: string) => {
    switch (position) {
      case "left":
        return "items-start text-left";
      case "right":
        return "items-end text-right";
      case "center":
        return "items-center text-center";
      default:
        return "items-start text-left";
    }
  };

  return (
    <div className="relative w-full h-[400px] md:h-[500px] lg:h-[600px] overflow-hidden bg-gray-100">
      {/* Slides */}
      <div className="relative w-full h-full">
        {slides.map((slide, index) => (
          <div
            key={slide.id}
            className={`absolute inset-0 transition-opacity duration-700 ease-in-out ${
              index === currentSlide ? "opacity-100 z-10" : "opacity-0 z-0"
            }`}
          >
            {/* Background Image */}
            <div
              className="absolute inset-0 bg-cover bg-center"
              style={{ backgroundImage: `url('${slide.image}')` }}
            />

            {/* Overlay */}
            <div className="absolute inset-0 bg-black/20" />

            {/* Content */}
            <div className="relative h-full max-w-7xl mx-auto px-6 lg:px-8">
              <div className={`h-full flex flex-col justify-center ${getTextAlignment(slide.textPosition)}`}>
                <div className="max-w-2xl">
                  {/* Subtitle */}
                  <p className="text-white/90 text-sm md:text-base font-medium tracking-wider mb-2 animate-fade-in-up">
                    {slide.subtitle}
                  </p>

                  {/* Title */}
                  <h1 className="text-white text-4xl md:text-5xl lg:text-6xl font-bold mb-4 animate-fade-in-up animation-delay-100">
                    {slide.title}
                  </h1>

                  {/* Description */}
                  <p className="text-white/90 text-lg md:text-xl mb-8 animate-fade-in-up animation-delay-200">
                    {slide.description}
                  </p>

                  {/* CTA Button */}
                  <Link
                    href={slide.ctaLink}
                    className="inline-block px-8 py-4 bg-white text-black font-semibold text-sm tracking-wider hover:bg-gray-100 transition-all duration-300 transform hover:scale-105 animate-fade-in-up animation-delay-300"
                  >
                    {slide.ctaText} →
                  </Link>
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Navigation Arrows */}
      <button
        onClick={() => {
          prevSlide();
          setIsAutoPlaying(false);
        }}
        className="absolute left-4 top-1/2 -translate-y-1/2 z-20 p-2 md:p-3 bg-white/80 hover:bg-white rounded-full transition-all duration-300 group"
        aria-label="Previous slide"
      >
        <ChevronLeft className="w-5 h-5 md:w-6 md:h-6 text-black group-hover:scale-110 transition-transform" />
      </button>

      <button
        onClick={() => {
          nextSlide();
          setIsAutoPlaying(false);
        }}
        className="absolute right-4 top-1/2 -translate-y-1/2 z-20 p-2 md:p-3 bg-white/80 hover:bg-white rounded-full transition-all duration-300 group"
        aria-label="Next slide"
      >
        <ChevronRight className="w-5 h-5 md:w-6 md:h-6 text-black group-hover:scale-110 transition-transform" />
      </button>

      {/* Dots Indicator */}
      <div className="absolute bottom-6 left-1/2 -translate-x-1/2 z-20 flex gap-2">
        {slides.map((_, index) => (
          <button
            key={index}
            onClick={() => goToSlide(index)}
            className={`transition-all duration-300 rounded-full ${
              index === currentSlide
                ? "w-8 h-2 bg-white"
                : "w-2 h-2 bg-white/50 hover:bg-white/75"
            }`}
            aria-label={`Go to slide ${index + 1}`}
          />
        ))}
      </div>

      {/* Progress Bar (optional) */}
      {isAutoPlaying && (
        <div className="absolute bottom-0 left-0 right-0 h-1 bg-white/20 z-20">
          <div
            className="h-full bg-white transition-all duration-[5000ms] ease-linear"
            style={{
              width: currentSlide === slides.length - 1 ? "100%" : "0%",
            }}
            key={currentSlide}
          />
        </div>
      )}
    </div>
  );
}
