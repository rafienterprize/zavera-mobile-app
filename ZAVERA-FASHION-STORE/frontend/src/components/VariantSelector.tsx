'use client';

import { useState, useEffect } from 'react';
import { ProductVariant, AvailableOptions } from '@/types/variant';
import { variantApi } from '@/lib/variantApi';

interface VariantSelectorProps {
  productId: number;
  variants: ProductVariant[];
  basePrice: number;
  onVariantChange: (variant: ProductVariant | null) => void;
}

export default function VariantSelector({
  productId,
  variants,
  basePrice,
  onVariantChange,
}: VariantSelectorProps) {
  const [selectedSize, setSelectedSize] = useState<string | null>(null);
  const [selectedColor, setSelectedColor] = useState<string | null>(null);
  const [selectedVariant, setSelectedVariant] = useState<ProductVariant | null>(null);
  const [availableOptions, setAvailableOptions] = useState<AvailableOptions>({});

  useEffect(() => {
    loadAvailableOptions();
  }, [productId]);

  useEffect(() => {
    findMatchingVariant();
  }, [selectedSize, selectedColor]);

  const loadAvailableOptions = async () => {
    try {
      const options = await variantApi.getAvailableOptions(productId);
      setAvailableOptions(options);
    } catch (error) {
      console.error('Failed to load options:', error);
    }
  };

  const findMatchingVariant = async () => {
    if (!selectedSize && !selectedColor) {
      setSelectedVariant(null);
      onVariantChange(null);
      return;
    }

    try {
      const variant = await variantApi.findVariant(productId, selectedSize || undefined, selectedColor || undefined);
      setSelectedVariant(variant);
      onVariantChange(variant);
    } catch (error) {
      setSelectedVariant(null);
      onVariantChange(null);
    }
  };

  const isVariantAvailable = (size?: string, color?: string): boolean => {
    const variant = variants.find(
      (v) =>
        v.is_active &&
        (!size || v.size === size) &&
        (!color || v.color === color) &&
        (v.available_stock || 0) > 0
    );
    return !!variant;
  };

  const getPrice = (): number => {
    if (selectedVariant?.price) {
      return selectedVariant.price;
    }
    return basePrice;
  };

  const getAvailableStock = (): number => {
    if (selectedVariant) {
      return selectedVariant.available_stock || 0;
    }
    return 0;
  };

  const getPriceRange = (): { min: number; max: number } | null => {
    const activePrices = variants
      .filter((v) => v.is_active)
      .map((v) => v.price || basePrice);

    if (activePrices.length === 0) return null;

    const min = Math.min(...activePrices);
    const max = Math.max(...activePrices);

    return min !== max ? { min, max } : null;
  };

  const priceRange = getPriceRange();

  return (
    <div className="space-y-6">
      {/* Price Display */}
      <div className="text-3xl font-bold">
        {selectedVariant ? (
          <span>Rp {getPrice().toLocaleString()}</span>
        ) : priceRange ? (
          <span>
            Rp {priceRange.min.toLocaleString()} - Rp {priceRange.max.toLocaleString()}
          </span>
        ) : (
          <span>Rp {basePrice.toLocaleString()}</span>
        )}
      </div>

      {/* Size Selector */}
      {availableOptions.size && availableOptions.size.length > 0 && (
        <div>
          <label className="block text-sm font-medium mb-3">
            Size {selectedSize && <span className="text-gray-500">: {selectedSize}</span>}
          </label>
          <div className="flex flex-wrap gap-2">
            {availableOptions.size.map((size) => {
              const available = isVariantAvailable(size, selectedColor || undefined);
              const isSelected = selectedSize === size;

              return (
                <button
                  key={size}
                  onClick={() => setSelectedSize(isSelected ? null : size)}
                  disabled={!available}
                  className={`px-4 py-2 border rounded-lg font-medium transition-all ${
                    isSelected
                      ? 'border-black bg-black text-white'
                      : available
                      ? 'border-gray-300 hover:border-black'
                      : 'border-gray-200 text-gray-400 cursor-not-allowed line-through'
                  }`}
                >
                  {size}
                </button>
              );
            })}
          </div>
        </div>
      )}

      {/* Color Selector */}
      {availableOptions.color && availableOptions.color.length > 0 && (
        <div>
          <label className="block text-sm font-medium mb-3">
            Color {selectedColor && <span className="text-gray-500">: {selectedColor}</span>}
          </label>
          <div className="flex flex-wrap gap-3">
            {availableOptions.color.map((color) => {
              const available = isVariantAvailable(selectedSize || undefined, color);
              const isSelected = selectedColor === color;
              const variant = variants.find((v) => v.color === color);
              const colorHex = variant?.color_hex;

              return (
                <button
                  key={color}
                  onClick={() => setSelectedColor(isSelected ? null : color)}
                  disabled={!available}
                  className={`relative group ${!available ? 'cursor-not-allowed opacity-50' : ''}`}
                  title={color}
                >
                  <div
                    className={`w-12 h-12 rounded-full border-2 transition-all ${
                      isSelected
                        ? 'border-black scale-110'
                        : available
                        ? 'border-gray-300 hover:border-black'
                        : 'border-gray-200'
                    }`}
                    style={{
                      backgroundColor: colorHex || '#ccc',
                    }}
                  />
                  {!available && (
                    <div className="absolute inset-0 flex items-center justify-center">
                      <div className="w-full h-0.5 bg-red-500 rotate-45" />
                    </div>
                  )}
                  <div className="absolute -bottom-6 left-1/2 transform -translate-x-1/2 text-xs whitespace-nowrap opacity-0 group-hover:opacity-100 transition-opacity">
                    {color}
                  </div>
                </button>
              );
            })}
          </div>
        </div>
      )}

      {/* Stock Status */}
      {selectedVariant && (
        <div className="text-sm">
          {getAvailableStock() > 0 ? (
            <span className="text-green-600 font-medium">
              {getAvailableStock()} in stock
            </span>
          ) : (
            <span className="text-red-600 font-medium">Out of stock</span>
          )}
        </div>
      )}

      {/* Selection Required Message */}
      {!selectedVariant && (availableOptions.size || availableOptions.color) && (
        <div className="text-sm text-gray-500">
          Please select {availableOptions.size && 'size'}
          {availableOptions.size && availableOptions.color && ' and '}
          {availableOptions.color && 'color'}
        </div>
      )}

      {/* Variant Details */}
      {selectedVariant && (
        <div className="text-sm text-gray-600 space-y-1">
          <div>SKU: {selectedVariant.sku}</div>
          {selectedVariant.material && <div>Material: {selectedVariant.material}</div>}
          {selectedVariant.pattern && <div>Pattern: {selectedVariant.pattern}</div>}
          {selectedVariant.fit && <div>Fit: {selectedVariant.fit}</div>}
        </div>
      )}
    </div>
  );
}
