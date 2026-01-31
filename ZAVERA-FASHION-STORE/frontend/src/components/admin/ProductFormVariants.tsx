'use client';

import { useState } from 'react';
import { Plus, X, Package, DollarSign, Ruler } from 'lucide-react';

const SIZES = ['XS', 'S', 'M', 'L', 'XL', 'XXL', 'XXXL'];
const COLORS = [
  { name: 'Black', hex: '#000000' },
  { name: 'White', hex: '#FFFFFF' },
  { name: 'Navy', hex: '#000080' },
  { name: 'Red', hex: '#FF0000' },
  { name: 'Blue', hex: '#0000FF' },
  { name: 'Green', hex: '#008000' },
  { name: 'Yellow', hex: '#FFFF00' },
  { name: 'Pink', hex: '#FFC0CB' },
  { name: 'Gray', hex: '#808080' },
  { name: 'Brown', hex: '#A52A2A' },
];

interface Variant {
  size: string;
  color: string;
  color_hex: string;
  stock: number;
  price: number;
  weight: number;
  length: number;
  width: number;
  height: number;
}

interface Props {
  formData: any;
  setFormData: (data: any) => void;
}

export default function ProductFormVariants({ formData, setFormData }: Props) {
  const [selectedSizes, setSelectedSizes] = useState<string[]>([]);
  const [selectedColors, setSelectedColors] = useState<string[]>([]);
  const [showBulkGenerator, setShowBulkGenerator] = useState(true);
  const [bulkStock, setBulkStock] = useState(10);

  const toggleSize = (size: string) => {
    setSelectedSizes((prev) =>
      prev.includes(size) ? prev.filter((s) => s !== size) : [...prev, size]
    );
  };

  const toggleColor = (colorName: string) => {
    setSelectedColors((prev) =>
      prev.includes(colorName) ? prev.filter((c) => c !== colorName) : [...prev, colorName]
    );
  };

  const generateVariants = () => {
    const variants: Variant[] = [];
    selectedSizes.forEach((size) => {
      selectedColors.forEach((colorName) => {
        const color = COLORS.find((c) => c.name === colorName);
        if (!color) return;

        // Default dimensions based on size
        const dimensions = getSizeDimensions(size);

        variants.push({
          size,
          color: color.name,
          color_hex: color.hex,
          stock: bulkStock,
          price: formData.base_price,
          ...dimensions,
        });
      });
    });

    setFormData({ ...formData, variants });
    setShowBulkGenerator(false);
  };

  const getSizeDimensions = (size: string) => {
    const dimensionMap: Record<string, any> = {
      XS: { weight: 300, length: 60, width: 40, height: 3 },
      S: { weight: 350, length: 65, width: 42, height: 3 },
      M: { weight: 400, length: 70, width: 45, height: 3 },
      L: { weight: 450, length: 75, width: 48, height: 3 },
      XL: { weight: 500, length: 80, width: 50, height: 3 },
      XXL: { weight: 550, length: 85, width: 52, height: 3 },
      XXXL: { weight: 600, length: 90, width: 55, height: 3 },
    };
    return dimensionMap[size] || { weight: 400, length: 70, width: 45, height: 3 };
  };

  const updateVariant = (index: number, field: string, value: any) => {
    const newVariants = [...formData.variants];
    newVariants[index] = { ...newVariants[index], [field]: value };
    setFormData({ ...formData, variants: newVariants });
  };

  const removeVariant = (index: number) => {
    setFormData({
      ...formData,
      variants: formData.variants.filter((_: any, i: number) => i !== index),
    });
  };

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold">Product Variants</h2>
        <p className="text-gray-600 mt-1">
          Create variants with different sizes, colors, prices, and dimensions
        </p>
      </div>

      {/* Bulk Generator */}
      {showBulkGenerator && formData.variants.length === 0 && (
        <div className="bg-gradient-to-br from-emerald-50 to-blue-50 border-2 border-emerald-200 rounded-xl p-6">
          <h3 className="text-lg font-semibold mb-4 flex items-center gap-2">
            <Package className="text-emerald-600" />
            Quick Variant Generator
          </h3>

          {/* Size Selection */}
          <div className="mb-6">
            <label className="block text-sm font-medium mb-3">
              Select Sizes <span className="text-red-500">*</span>
            </label>
            <div className="flex flex-wrap gap-2">
              {SIZES.map((size) => (
                <button
                  key={size}
                  onClick={() => toggleSize(size)}
                  className={`px-4 py-2 rounded-lg font-medium transition-all ${
                    selectedSizes.includes(size)
                      ? 'bg-emerald-500 text-white shadow-md scale-105'
                      : 'bg-white border-2 border-gray-200 hover:border-emerald-300'
                  }`}
                >
                  {size}
                </button>
              ))}
            </div>
            <p className="text-sm text-gray-500 mt-2">Selected: {selectedSizes.length} sizes</p>
          </div>

          {/* Color Selection */}
          <div className="mb-6">
            <label className="block text-sm font-medium mb-3">
              Select Colors <span className="text-red-500">*</span>
            </label>
            <div className="grid grid-cols-5 gap-3">
              {COLORS.map((color) => (
                <button
                  key={color.name}
                  onClick={() => toggleColor(color.name)}
                  className={`relative p-3 rounded-lg border-2 transition-all ${
                    selectedColors.includes(color.name)
                      ? 'border-emerald-500 shadow-md scale-105'
                      : 'border-gray-200 hover:border-gray-300'
                  }`}
                >
                  <div
                    className="w-full h-12 rounded mb-2"
                    style={{ backgroundColor: color.hex }}
                  />
                  <p className="text-xs font-medium text-center">{color.name}</p>
                  {selectedColors.includes(color.name) && (
                    <div className="absolute top-1 right-1 w-5 h-5 bg-emerald-500 rounded-full flex items-center justify-center">
                      <span className="text-white text-xs">✓</span>
                    </div>
                  )}
                </button>
              ))}
            </div>
            <p className="text-sm text-gray-500 mt-2">Selected: {selectedColors.length} colors</p>
          </div>

          {/* Default Stock */}
          <div className="mb-6">
            <label className="block text-sm font-medium mb-2">Default Stock per Variant</label>
            <input
              type="number"
              value={bulkStock}
              onChange={(e) => setBulkStock(parseInt(e.target.value))}
              className="w-full px-4 py-3 border rounded-lg"
              min="0"
            />
          </div>

          {/* Generate Button */}
          <div className="flex items-center justify-between bg-white rounded-lg p-4">
            <div>
              <p className="font-semibold">
                Will generate: {selectedSizes.length} × {selectedColors.length} ={' '}
                <span className="text-emerald-600">{selectedSizes.length * selectedColors.length} variants</span>
              </p>
              <p className="text-sm text-gray-500">Each variant will have customizable dimensions</p>
            </div>
            <button
              onClick={generateVariants}
              disabled={selectedSizes.length === 0 || selectedColors.length === 0}
              className="px-6 py-3 bg-emerald-500 text-white rounded-lg font-medium hover:bg-emerald-600 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Generate Variants
            </button>
          </div>
        </div>
      )}

      {/* Variants List */}
      {formData.variants.length > 0 && (
        <div>
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold">
              Variants ({formData.variants.length})
            </h3>
            <button
              onClick={() => setShowBulkGenerator(true)}
              className="text-sm text-emerald-600 hover:text-emerald-700"
            >
              + Add More Variants
            </button>
          </div>

          <div className="space-y-4">
            {formData.variants.map((variant: Variant, index: number) => (
              <div key={index} className="bg-white border-2 border-gray-200 rounded-xl p-6 hover:border-emerald-300 transition-colors">
                <div className="flex items-start justify-between mb-4">
                  <div className="flex items-center gap-3">
                    <div
                      className="w-12 h-12 rounded-lg border-2"
                      style={{ backgroundColor: variant.color_hex }}
                    />
                    <div>
                      <h4 className="font-semibold text-lg">
                        {variant.size} - {variant.color}
                      </h4>
                      <p className="text-sm text-gray-500">Variant #{index + 1}</p>
                    </div>
                  </div>
                  <button
                    onClick={() => removeVariant(index)}
                    className="p-2 text-red-500 hover:bg-red-50 rounded-lg"
                  >
                    <X size={20} />
                  </button>
                </div>

                <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                  {/* Stock */}
                  <div>
                    <label className="block text-xs font-medium mb-1 flex items-center gap-1">
                      <Package size={14} />
                      Stock
                    </label>
                    <input
                      type="number"
                      value={variant.stock}
                      onChange={(e) => updateVariant(index, 'stock', parseInt(e.target.value))}
                      className="w-full px-3 py-2 border rounded-lg text-sm"
                      min="0"
                    />
                  </div>

                  {/* Price */}
                  <div>
                    <label className="block text-xs font-medium mb-1 flex items-center gap-1">
                      <DollarSign size={14} />
                      Price (IDR)
                    </label>
                    <input
                      type="number"
                      value={variant.price}
                      onChange={(e) => updateVariant(index, 'price', parseFloat(e.target.value))}
                      className="w-full px-3 py-2 border rounded-lg text-sm"
                      min="0"
                    />
                  </div>

                  {/* Weight */}
                  <div>
                    <label className="block text-xs font-medium mb-1 flex items-center gap-1">
                      <Ruler size={14} />
                      Weight (g)
                    </label>
                    <input
                      type="number"
                      value={variant.weight}
                      onChange={(e) => updateVariant(index, 'weight', parseInt(e.target.value))}
                      className="w-full px-3 py-2 border rounded-lg text-sm"
                      min="0"
                    />
                  </div>

                  {/* Length */}
                  <div>
                    <label className="block text-xs font-medium mb-1">Length (cm)</label>
                    <input
                      type="number"
                      value={variant.length}
                      onChange={(e) => updateVariant(index, 'length', parseInt(e.target.value))}
                      className="w-full px-3 py-2 border rounded-lg text-sm"
                      min="0"
                    />
                  </div>

                  {/* Width */}
                  <div>
                    <label className="block text-xs font-medium mb-1">Width (cm)</label>
                    <input
                      type="number"
                      value={variant.width}
                      onChange={(e) => updateVariant(index, 'width', parseInt(e.target.value))}
                      className="w-full px-3 py-2 border rounded-lg text-sm"
                      min="0"
                    />
                  </div>

                  {/* Height */}
                  <div>
                    <label className="block text-xs font-medium mb-1">Height (cm)</label>
                    <input
                      type="number"
                      value={variant.height}
                      onChange={(e) => updateVariant(index, 'height', parseInt(e.target.value))}
                      className="w-full px-3 py-2 border rounded-lg text-sm"
                      min="0"
                    />
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Info */}
      <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
        <h4 className="font-semibold text-blue-900 mb-2">💡 Variant Tips:</h4>
        <ul className="text-sm text-blue-800 space-y-1">
          <li>• Each size can have different dimensions for accurate shipping costs</li>
          <li>• Set different prices per variant if needed (e.g., XL costs more)</li>
          <li>• Stock is tracked separately for each variant</li>
          <li>• Dimensions affect shipping cost calculation via Biteship</li>
        </ul>
      </div>
    </div>
  );
}
