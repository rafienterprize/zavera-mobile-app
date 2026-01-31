'use client';

import { useState, useEffect } from 'react';
import { Package } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import { useDialog } from '@/hooks/useDialog';
import { variantApi } from '@/lib/variantApi';
import { ProductVariant, CreateVariantRequest, VariantStockSummary } from '@/types/variant';

// Color options with hex codes
const COLOR_OPTIONS = [
  { name: 'Black', hex: '#000000' },
  { name: 'White', hex: '#FFFFFF' },
  { name: 'Navy', hex: '#000080' },
  { name: 'Red', hex: '#FF0000' },
  { name: 'Blue', hex: '#0000FF' },
  { name: 'Green', hex: '#008000' },
  { name: 'Gray', hex: '#808080' },
  { name: 'Pink', hex: '#FFC0CB' },
  { name: 'Yellow', hex: '#FFFF00' },
  { name: 'Brown', hex: '#8B4513' },
  { name: 'Beige', hex: '#F5F5DC' },
  { name: 'Orange', hex: '#FFA500' },
];

interface VariantManagerProps {
  productId: number;
  productPrice: number;
}

export default function VariantManager({ productId, productPrice }: VariantManagerProps) {
  const { token } = useAuth();
  const dialog = useDialog();
  const [variants, setVariants] = useState<ProductVariant[]>([]);
  const [stockSummary, setStockSummary] = useState<VariantStockSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [showBulkGenerator, setShowBulkGenerator] = useState(false);
  const [showAddForm, setShowAddForm] = useState(false);
  const [editingVariant, setEditingVariant] = useState<ProductVariant | null>(null);

  // Bulk generator state
  const [bulkSizes, setBulkSizes] = useState<string[]>(['S', 'M', 'L', 'XL']);
  const [bulkColors, setBulkColors] = useState<string[]>(['Black', 'White', 'Navy']);
  const [bulkStock, setBulkStock] = useState(10);
  const [bulkPrice, setBulkPrice] = useState(productPrice);
  const [bulkWeight, setBulkWeight] = useState(400); // Default 400g for clothing
  const [bulkLength, setBulkLength] = useState(30); // Default 30cm
  const [bulkWidth, setBulkWidth] = useState(20); // Default 20cm
  const [bulkHeight, setBulkHeight] = useState(5); // Default 5cm (folded)

  // Form state
  const [formData, setFormData] = useState<CreateVariantRequest>({
    product_id: productId,
    sku: '',
    variant_name: '',
    size: '',
    color: '',
    color_hex: '',
    stock_quantity: 0,
    low_stock_threshold: 5,
    is_active: true,
    is_default: false,
    position: 0,
  });

  const loadVariants = async () => {
    try {
      const data = await variantApi.getProductVariants(productId);
      setVariants(data);
    } catch (error) {
      console.error('Failed to load variants:', error);
    } finally {
      setLoading(false);
    }
  };

  const loadStockSummary = async () => {
    if (!token) return;
    try {
      const data = await variantApi.getStockSummary(token, productId);
      setStockSummary(data);
    } catch (error) {
      console.error('Failed to load stock summary:', error);
    }
  };

  useEffect(() => {
    loadVariants();
    if (token) {
      loadStockSummary();
    }
  }, [productId, token]);

  const handleBulkGenerate = async () => {
    if (!token) return;
    try {
      await variantApi.bulkGenerateVariants(token, {
        product_id: productId,
        sizes: bulkSizes,
        colors: bulkColors,
        base_price: bulkPrice,
        stock_per_variant: bulkStock,
        weight: bulkWeight,
        length: bulkLength,
        width: bulkWidth,
        height: bulkHeight,
      });
      await dialog.alert({
        title: 'Berhasil!',
        message: 'Variants berhasil dibuat!',
      });
      setShowBulkGenerator(false);
      loadVariants();
      loadStockSummary();
    } catch (error) {
      await dialog.alert({
        title: 'Error',
        message: 'Gagal membuat variants',
      });
      console.error(error);
    }
  };

  const handleCreateVariant = async () => {
    if (!token) return;
    try {
      await variantApi.createVariant(token, formData);
      await dialog.alert({
        title: 'Berhasil!',
        message: 'Variant berhasil dibuat!',
      });
      setShowAddForm(false);
      resetForm();
      loadVariants();
      loadStockSummary();
    } catch (error) {
      await dialog.alert({
        title: 'Error',
        message: 'Gagal membuat variant',
      });
      console.error(error);
    }
  };

  const handleUpdateVariant = async () => {
    console.log('🔄 handleUpdateVariant called');
    console.log('📝 Token:', token ? 'EXISTS' : 'NULL');
    console.log('📝 editingVariant:', editingVariant);
    console.log('📝 formData:', formData);
    
    if (!token || !editingVariant) {
      console.error('❌ Missing token or editingVariant');
      return;
    }
    
    try {
      console.log('🚀 Calling variantApi.updateVariant...');
      await variantApi.updateVariant(token, editingVariant.id, formData);
      console.log('✅ Update successful!');
      await dialog.alert({
        title: 'Berhasil!',
        message: 'Variant berhasil diupdate!',
      });
      setEditingVariant(null);
      resetForm();
      loadVariants();
      loadStockSummary();
    } catch (error) {
      console.error('❌ Update failed:', error);
      await dialog.alert({
        title: 'Error',
        message: 'Gagal mengupdate variant',
      });
      console.error(error);
    }
  };

  const handleDeleteVariant = async (id: number) => {
    if (!token) return;
    
    const confirmed = await dialog.confirm({
      title: 'Hapus Variant',
      message: 'Apakah Anda yakin ingin menghapus variant ini? Tindakan ini tidak dapat dibatalkan.',
    });
    
    if (!confirmed) return;
    
    try {
      await variantApi.deleteVariant(token, id);
      await dialog.alert({
        title: 'Berhasil!',
        message: 'Variant berhasil dihapus!',
      });
      loadVariants();
      loadStockSummary();
    } catch (error) {
      await dialog.alert({
        title: 'Error',
        message: 'Gagal menghapus variant. Mungkin variant ini memiliki order yang sedang berjalan.',
      });
      console.error(error);
    }
  };

  const handleUpdateStock = async (variantId: number, newStock: number) => {
    if (!token) return;
    try {
      await variantApi.updateStock(token, variantId, newStock);
      loadVariants();
      loadStockSummary();
    } catch (error) {
      await dialog.alert({
        title: 'Error',
        message: 'Gagal mengupdate stok',
      });
      console.error(error);
    }
  };

  const resetForm = () => {
    setFormData({
      product_id: productId,
      sku: '',
      variant_name: '',
      size: '',
      color: '',
      color_hex: '',
      stock_quantity: 0,
      low_stock_threshold: 5,
      is_active: true,
      is_default: false,
      position: 0,
    });
  };

  const startEdit = (variant: ProductVariant) => {
    console.log('📝 Editing variant:', variant);
    console.log('📦 Dimensions:', {
      weight_grams: variant.weight_grams,
      length_cm: variant.length_cm,
      width_cm: variant.width_cm,
      height_cm: variant.height_cm
    });
    
    setEditingVariant(variant);
    setFormData({
      product_id: variant.product_id,
      sku: variant.sku,
      variant_name: variant.variant_name,
      size: variant.size || '',
      color: variant.color || '',
      color_hex: variant.color_hex || '',
      stock_quantity: variant.stock_quantity,
      low_stock_threshold: variant.low_stock_threshold,
      is_active: variant.is_active,
      is_default: variant.is_default,
      position: variant.position,
      price: variant.price,
      weight: variant.weight_grams || variant.weight,
      length: variant.length_cm || variant.length,
      width: variant.width_cm || variant.width,
      height: variant.height_cm || variant.height,
    });
    setShowAddForm(true);
  };

  if (loading) {
    return <div className="p-4">Loading variants...</div>;
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold text-white">Product Variants</h2>
        <div className="space-x-2">
          <button
            onClick={() => setShowBulkGenerator(!showBulkGenerator)}
            className="px-4 py-2 bg-purple-600 text-white rounded-xl hover:bg-purple-700 transition-colors"
          >
            Bulk Generate
          </button>
          <button
            onClick={() => {
              setShowAddForm(!showAddForm);
              setEditingVariant(null);
              resetForm();
            }}
            className="px-4 py-2 bg-emerald-600 text-white rounded-xl hover:bg-emerald-700 transition-colors"
          >
            Add Variant
          </button>
        </div>
      </div>

      {showBulkGenerator && (
        <div className="bg-neutral-900 p-6 rounded-2xl border border-white/10">
          <h3 className="text-lg font-semibold text-white mb-4">Bulk Generate Variants</h3>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-white/80 mb-2">Sizes (comma-separated)</label>
              <input
                type="text"
                value={bulkSizes.join(', ')}
                onChange={(e) => setBulkSizes(e.target.value.split(',').map((s) => s.trim()))}
                className="w-full px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white focus:outline-none focus:border-emerald-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-white/80 mb-2">Colors (comma-separated)</label>
              <input
                type="text"
                value={bulkColors.join(', ')}
                onChange={(e) => setBulkColors(e.target.value.split(',').map((s) => s.trim()))}
                className="w-full px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white focus:outline-none focus:border-emerald-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-white/80 mb-2">Stock per Variant</label>
              <input
                type="number"
                value={bulkStock}
                onChange={(e) => setBulkStock(parseInt(e.target.value))}
                className="w-full px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white focus:outline-none focus:border-emerald-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-white/80 mb-2">Base Price (optional)</label>
              <input
                type="number"
                value={bulkPrice}
                onChange={(e) => setBulkPrice(parseFloat(e.target.value))}
                className="w-full px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white focus:outline-none focus:border-emerald-500"
              />
            </div>
          </div>

          {/* Bulk Dimensions */}
          <div className="mt-4 p-4 bg-blue-500/10 rounded-xl border border-blue-500/20">
            <h4 className="text-sm font-semibold text-blue-400 mb-3">📦 Default Dimensions (Applied to All Variants)</h4>
            <p className="text-xs text-white/60 mb-4">
              Dimensi ini akan diterapkan ke semua variant yang di-generate. Bisa diedit per-variant nanti.
            </p>
            <div className="grid grid-cols-4 gap-4">
              <div>
                <label className="block text-sm font-medium text-white/80 mb-2">Weight (g)</label>
                <input
                  type="number"
                  value={bulkWeight}
                  onChange={(e) => setBulkWeight(parseInt(e.target.value))}
                  className="w-full px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white focus:outline-none focus:border-emerald-500"
                  placeholder="400"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-white/80 mb-2">Length (cm)</label>
                <input
                  type="number"
                  value={bulkLength}
                  onChange={(e) => setBulkLength(parseInt(e.target.value))}
                  className="w-full px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white focus:outline-none focus:border-emerald-500"
                  placeholder="30"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-white/80 mb-2">Width (cm)</label>
                <input
                  type="number"
                  value={bulkWidth}
                  onChange={(e) => setBulkWidth(parseInt(e.target.value))}
                  className="w-full px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white focus:outline-none focus:border-emerald-500"
                  placeholder="20"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-white/80 mb-2">Height (cm)</label>
                <input
                  type="number"
                  value={bulkHeight}
                  onChange={(e) => setBulkHeight(parseInt(e.target.value))}
                  className="w-full px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white focus:outline-none focus:border-emerald-500"
                  placeholder="5"
                />
              </div>
            </div>
          </div>

          <div className="mt-6 flex justify-end space-x-3">
            <button
              onClick={() => setShowBulkGenerator(false)}
              className="px-6 py-2.5 rounded-xl bg-white/10 text-white hover:bg-white/20 transition-colors"
            >
              Cancel
            </button>
            <button
              onClick={handleBulkGenerate}
              className="px-6 py-2.5 bg-purple-600 text-white rounded-xl hover:bg-purple-700 transition-colors"
            >
              Generate {bulkSizes.length * bulkColors.length} Variants
            </button>
          </div>
        </div>
      )}

      {showAddForm && (
        <div className="bg-neutral-900 p-6 rounded-2xl border border-white/10">
          <h3 className="text-lg font-semibold text-white mb-4">
            {editingVariant ? 'Edit Variant' : 'Add New Variant'}
          </h3>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-white/80 mb-2">SKU</label>
              <input
                type="text"
                value={formData.sku}
                onChange={(e) => setFormData({ ...formData, sku: e.target.value })}
                className="w-full px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white placeholder-white/40 focus:outline-none focus:border-emerald-500"
                placeholder="Auto-generated if empty"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-white/80 mb-2">Variant Name</label>
              <input
                type="text"
                value={formData.variant_name}
                onChange={(e) => setFormData({ ...formData, variant_name: e.target.value })}
                className="w-full px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white placeholder-white/40 focus:outline-none focus:border-emerald-500"
                placeholder="Auto-generated if empty"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-white/80 mb-2">Size</label>
              <input
                type="text"
                value={formData.size}
                onChange={(e) => setFormData({ ...formData, size: e.target.value })}
                className="w-full px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white focus:outline-none focus:border-emerald-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-white/80 mb-2">Color</label>
              <select
                value={formData.color}
                onChange={(e) => {
                  console.log('🎨 Color changed to:', e.target.value);
                  const selectedColor = COLOR_OPTIONS.find(c => c.name === e.target.value);
                  console.log('🎨 Selected color object:', selectedColor);
                  const newFormData = { 
                    ...formData, 
                    color: e.target.value,
                    color_hex: selectedColor?.hex || ''
                  };
                  console.log('🎨 New formData:', newFormData);
                  setFormData(newFormData);
                }}
                className="w-full px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white focus:outline-none focus:border-emerald-500"
              >
                <option value="">Select Color</option>
                {COLOR_OPTIONS.map((color) => (
                  <option key={color.name} value={color.name} className="bg-neutral-900">
                    {color.name}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-white/80 mb-2">Color Hex (Auto-filled)</label>
              <input
                type="text"
                value={formData.color_hex}
                onChange={(e) => setFormData({ ...formData, color_hex: e.target.value })}
                className="w-full px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white placeholder-white/40 focus:outline-none focus:border-emerald-500"
                placeholder="#000000"
                readOnly
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-white/80 mb-2">Price Override</label>
              <input
                type="number"
                step="0.01"
                value={formData.price || ''}
                onChange={(e) =>
                  setFormData({ ...formData, price: e.target.value ? parseFloat(e.target.value) : undefined })
                }
                className="w-full px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white placeholder-white/40 focus:outline-none focus:border-emerald-500"
                placeholder="Leave empty to use product price"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-white/80 mb-2">Stock Quantity</label>
              <input
                type="number"
                value={formData.stock_quantity}
                onChange={(e) => setFormData({ ...formData, stock_quantity: parseInt(e.target.value) })}
                className="w-full px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white focus:outline-none focus:border-emerald-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-white/80 mb-2">Low Stock Threshold</label>
              <input
                type="number"
                value={formData.low_stock_threshold}
                onChange={(e) => setFormData({ ...formData, low_stock_threshold: parseInt(e.target.value) })}
                className="w-full px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white focus:outline-none focus:border-emerald-500"
              />
            </div>
          </div>

          {/* Shipping Dimensions */}
          <div className="mt-4 p-4 bg-white/5 rounded-xl border border-white/10">
            <h4 className="text-sm font-semibold text-white mb-3">📦 Shipping Dimensions</h4>
            <p className="text-xs text-white/60 mb-4">
              Digunakan untuk kalkulasi ongkir. Berat dijumlahkan, dimensi pakai yang terbesar (P x L) dan tinggi ditumpuk.
            </p>
            <div className="grid grid-cols-4 gap-4">
              <div>
                <label className="block text-sm font-medium text-white/80 mb-2">Weight (g)</label>
                <input
                  type="number"
                  value={formData.weight || ''}
                  onChange={(e) => setFormData({ ...formData, weight: e.target.value ? parseInt(e.target.value) : undefined })}
                  className="w-full px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white placeholder-white/40 focus:outline-none focus:border-emerald-500"
                  placeholder="500"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-white/80 mb-2">Length (cm)</label>
                <input
                  type="number"
                  value={formData.length || ''}
                  onChange={(e) => setFormData({ ...formData, length: e.target.value ? parseInt(e.target.value) : undefined })}
                  className="w-full px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white placeholder-white/40 focus:outline-none focus:border-emerald-500"
                  placeholder="30"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-white/80 mb-2">Width (cm)</label>
                <input
                  type="number"
                  value={formData.width || ''}
                  onChange={(e) => setFormData({ ...formData, width: e.target.value ? parseInt(e.target.value) : undefined })}
                  className="w-full px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white placeholder-white/40 focus:outline-none focus:border-emerald-500"
                  placeholder="20"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-white/80 mb-2">Height (cm)</label>
                <input
                  type="number"
                  value={formData.height || ''}
                  onChange={(e) => setFormData({ ...formData, height: e.target.value ? parseInt(e.target.value) : undefined })}
                  className="w-full px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white placeholder-white/40 focus:outline-none focus:border-emerald-500"
                  placeholder="5"
                />
              </div>
            </div>
          </div>

          {/* Active & Default Checkboxes */}
          <div className="mt-4 flex items-center space-x-6">
            <label className="flex items-center text-white">
              <input
                type="checkbox"
                checked={formData.is_active}
                onChange={(e) => setFormData({ ...formData, is_active: e.target.checked })}
                className="mr-2 w-4 h-4 rounded border-white/20 bg-white/5 text-emerald-500 focus:ring-emerald-500"
              />
              Active
            </label>
            <label className="flex items-center text-white">
              <input
                type="checkbox"
                checked={formData.is_default}
                onChange={(e) => setFormData({ ...formData, is_default: e.target.checked })}
                className="mr-2 w-4 h-4 rounded border-white/20 bg-white/5 text-emerald-500 focus:ring-emerald-500"
              />
              Default
            </label>
          </div>

          <div className="mt-6 flex justify-end space-x-3">
            <button
              onClick={() => {
                setShowAddForm(false);
                setEditingVariant(null);
                resetForm();
              }}
              className="px-6 py-2.5 rounded-xl bg-white/10 text-white hover:bg-white/20 transition-colors"
            >
              Cancel
            </button>
            <button
              onClick={editingVariant ? handleUpdateVariant : handleCreateVariant}
              className="px-6 py-2.5 bg-emerald-600 text-white rounded-xl hover:bg-emerald-700 transition-colors"
            >
              {editingVariant ? 'Update' : 'Create'} Variant
            </button>
          </div>
        </div>
      )}

      <div className="bg-neutral-900 rounded-2xl border border-white/10 overflow-hidden">
        <table className="w-full">
          <thead className="bg-neutral-800/50 border-b border-white/10">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium text-white/60 uppercase tracking-wider">SKU</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-white/60 uppercase tracking-wider">Variant</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-white/60 uppercase tracking-wider">Size</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-white/60 uppercase tracking-wider">Color</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-white/60 uppercase tracking-wider">Price</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-white/60 uppercase tracking-wider">Stock</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-white/60 uppercase tracking-wider">Available</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-white/60 uppercase tracking-wider">Status</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-white/60 uppercase tracking-wider">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-white/5">
            {variants.map((variant) => {
              const summary = stockSummary.find((s) => s.variant_id === variant.id);
              return (
                <tr key={variant.id} className="hover:bg-white/5 transition-colors">
                  <td className="px-4 py-3 text-sm font-mono text-white/80">{variant.sku}</td>
                  <td className="px-4 py-3 text-sm text-white">{variant.variant_name}</td>
                  <td className="px-4 py-3 text-sm text-white">{variant.size || '-'}</td>
                  <td className="px-4 py-3 text-sm">
                    <div className="flex items-center space-x-2">
                      {variant.color_hex && (
                        <div
                          className="w-5 h-5 rounded border-2 border-white/20"
                          style={{ backgroundColor: variant.color_hex }}
                        />
                      )}
                      <span className="text-white">{variant.color || '-'}</span>
                    </div>
                  </td>
                  <td className="px-4 py-3 text-sm text-white">
                    {variant.price ? `Rp ${variant.price.toLocaleString()}` : '-'}
                  </td>
                  <td className="px-4 py-3 text-sm">
                    <input
                      type="number"
                      value={variant.stock_quantity}
                      onChange={(e) => handleUpdateStock(variant.id, parseInt(e.target.value))}
                      className="w-20 px-3 py-1.5 rounded-lg bg-white/5 border border-white/10 text-white focus:outline-none focus:border-emerald-500"
                    />
                  </td>
                  <td className="px-4 py-3 text-sm text-white">
                    {summary ? summary.available_quantity : variant.available_stock || 0}
                  </td>
                  <td className="px-4 py-3 text-sm">
                    <span
                      className={`px-2 py-1 rounded-lg text-xs font-medium ${
                        variant.is_active ? 'bg-emerald-500/20 text-emerald-400' : 'bg-white/10 text-white/60'
                      }`}
                    >
                      {variant.is_active ? 'Active' : 'Inactive'}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm space-x-3">
                    <button
                      onClick={() => startEdit(variant)}
                      className="text-blue-400 hover:text-blue-300 font-medium transition-colors"
                    >
                      Edit
                    </button>
                    <button
                      onClick={() => handleDeleteVariant(variant.id)}
                      className="text-red-400 hover:text-red-300 font-medium transition-colors"
                    >
                      Delete
                    </button>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
        {variants.length === 0 && (
          <div className="p-12 text-center text-white/40">
            <Package size={48} className="mx-auto mb-3 opacity-50" />
            <p className="text-lg">No variants found</p>
            <p className="text-sm mt-1">Create variants to manage stock by size, color, etc.</p>
          </div>
        )}
      </div>
    </div>
  );
}
