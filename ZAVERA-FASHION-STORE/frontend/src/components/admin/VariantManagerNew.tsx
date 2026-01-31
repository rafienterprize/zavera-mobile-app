'use client';

import { useState, useEffect } from 'react';
import { Package, Plus, Trash2 } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import { useDialog } from '@/hooks/useDialog';
import { AlertDialog } from '@/components/Dialog';
import { variantApi } from '@/lib/variantApi';
import { ProductVariant } from '@/types/variant';

// Size and Color options (same as create product)
const SIZES = ['XS', 'S', 'M', 'L', 'XL', 'XXL', 'XXXL'];
const COLORS = [
  { name: 'Black', hex: '#000000' },
  { name: 'White', hex: '#FFFFFF' },
  { name: 'Navy', hex: '#000080' },
  { name: 'Red', hex: '#FF0000' },
  { name: 'Blue', hex: '#0000FF' },
  { name: 'Green', hex: '#008000' },
  { name: 'Gray', hex: '#808080' },
  { name: 'Pink', hex: '#FFC0CB' },
];

interface VariantManagerProps {
  productId: number;
  productName: string;
  productPrice: number;
}

interface VariantFormData {
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

export default function VariantManagerNew({ productId, productName, productPrice }: VariantManagerProps) {
  const { token } = useAuth();
  const dialog = useDialog();
  const [variants, setVariants] = useState<ProductVariant[]>([]);
  const [localVariants, setLocalVariants] = useState<VariantFormData[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    loadVariants();
  }, [productId]);

  const loadVariants = async () => {
    try {
      const data = await variantApi.getProductVariants(productId);
      setVariants(data);
      
      // Convert to local format for editing
      const localData = data.map(v => ({
        size: v.size || 'M',
        color: v.color || 'Black',
        color_hex: v.color_hex || '#000000',
        stock: v.stock_quantity || 0,
        price: v.price || productPrice,
        weight: v.weight_grams || 400,
        length: v.length_cm || 70,
        width: v.width_cm || 45,
        height: v.height_cm || 3,
      }));
      setLocalVariants(localData);
    } catch (error) {
      console.error('Failed to load variants:', error);
    } finally {
      setLoading(false);
    }
  };

  const addVariant = () => {
    setLocalVariants([...localVariants, {
      size: 'M',
      color: 'Black',
      color_hex: '#000000',
      stock: 10,
      price: productPrice,
      weight: 400,
      length: 70,
      width: 45,
      height: 3,
    }]);
  };

  const removeVariant = (index: number) => {
    setLocalVariants(localVariants.filter((_, i) => i !== index));
  };

  const updateVariant = (index: number, field: string, value: any) => {
    const newVariants = [...localVariants];
    newVariants[index] = { ...newVariants[index], [field]: value };
    setLocalVariants(newVariants);
  };

  const updateVariantColor = (index: number, colorName: string) => {
    const color = COLORS.find(c => c.name === colorName);
    const newVariants = [...localVariants];
    newVariants[index] = {
      ...newVariants[index],
      color: colorName,
      color_hex: color?.hex || '#000000'
    };
    setLocalVariants(newVariants);
  };

  const handleSave = async () => {
    if (!token) return;
    
    setSaving(true);
    try {
      // Delete all existing variants
      for (const variant of variants) {
        try {
          await variantApi.deleteVariant(token, variant.id);
        } catch (error) {
          console.error('Failed to delete variant:', error);
        }
      }

      // Create new variants
      let successCount = 0;
      let failCount = 0;

      for (let i = 0; i < localVariants.length; i++) {
        const variant = localVariants[i];
        try {
          await variantApi.createVariant(token, {
            product_id: productId,
            size: variant.size,
            color: variant.color,
            color_hex: variant.color_hex,
            stock_quantity: variant.stock,
            price: variant.price,
            weight_grams: variant.weight,
            is_active: true,
            low_stock_threshold: 5,
            position: i,
          });
          successCount++;
        } catch (error) {
          console.error('Failed to create variant:', error);
          failCount++;
        }
      }

      if (failCount > 0) {
        await dialog.alert({
          title: 'Partially Saved',
          message: `${successCount} variants saved, ${failCount} failed.`,
          variant: 'warning',
        });
      } else {
        await dialog.alert({
          title: 'Success!',
          message: 'All variants saved successfully!',
          variant: 'success',
        });
      }

      loadVariants();
    } catch (error) {
      await dialog.alert({
        title: 'Error',
        message: 'Failed to save variants',
        variant: 'error',
      });
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
        <div className="text-white/60">Loading variants...</div>
      </div>
    );
  }

  return (
    <>
      <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h2 className="text-lg font-semibold text-white flex items-center gap-2">
              <Package size={20} />
              Product Variants
            </h2>
            <p className="text-white/60 text-sm mt-1">Each variant has its own stock, price, and dimensions</p>
          </div>
          <div className="flex gap-2">
            <button
              type="button"
              onClick={addVariant}
              className="px-4 py-2 bg-emerald-500 text-white rounded-lg hover:bg-emerald-600 flex items-center gap-2"
            >
              <Plus size={16} /> Add Variant
            </button>
            {localVariants.length > 0 && (
              <button
                type="button"
                onClick={handleSave}
                disabled={saving}
                className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50"
              >
                {saving ? 'Saving...' : 'Save Changes'}
              </button>
            )}
          </div>
        </div>

        {localVariants.length === 0 ? (
          <div className="text-center py-8 text-white/40">
            <Package size={48} className="mx-auto mb-3 opacity-50" />
            <p>No variants yet. Click &quot;Add Variant&quot; to create one.</p>
          </div>
        ) : (
          <div className="space-y-4">
            {localVariants.map((variant, index) => (
              <div key={index} className="bg-white/5 rounded-xl p-4 border border-white/10">
                <div className="flex items-start justify-between mb-4">
                  <h3 className="text-white font-medium">Variant #{index + 1}</h3>
                  <button
                    type="button"
                    onClick={() => removeVariant(index)}
                    className="p-1.5 text-red-400 hover:bg-red-500/10 rounded"
                  >
                    <Trash2 size={16} />
                  </button>
                </div>

                <div className="grid grid-cols-4 gap-3">
                  <div>
                    <label className="block text-white/60 text-xs mb-1">Size</label>
                    <select
                      value={variant.size}
                      onChange={(e) => updateVariant(index, 'size', e.target.value)}
                      className="w-full px-3 py-2 rounded-lg bg-white/5 border border-white/10 text-white text-sm"
                    >
                      {SIZES.map(s => <option key={s} value={s}>{s}</option>)}
                    </select>
                  </div>
                  <div>
                    <label className="block text-white/60 text-xs mb-1">Color</label>
                    <select
                      value={variant.color}
                      onChange={(e) => updateVariantColor(index, e.target.value)}
                      className="w-full px-3 py-2 rounded-lg bg-white/5 border border-white/10 text-white text-sm"
                    >
                      {COLORS.map(c => <option key={c.name} value={c.name}>{c.name}</option>)}
                    </select>
                  </div>
                  <div>
                    <label className="block text-white/60 text-xs mb-1">Stock</label>
                    <input
                      type="number"
                      value={variant.stock}
                      onChange={(e) => updateVariant(index, 'stock', parseInt(e.target.value))}
                      className="w-full px-3 py-2 rounded-lg bg-white/5 border border-white/10 text-white text-sm"
                    />
                  </div>
                  <div>
                    <label className="block text-white/60 text-xs mb-1">Price (IDR)</label>
                    <input
                      type="number"
                      value={variant.price}
                      onChange={(e) => updateVariant(index, 'price', parseFloat(e.target.value))}
                      className="w-full px-3 py-2 rounded-lg bg-white/5 border border-white/10 text-white text-sm"
                      placeholder="0 = use product price"
                    />
                  </div>
                </div>

                <div className="grid grid-cols-4 gap-3 mt-3">
                  <div>
                    <label className="block text-white/60 text-xs mb-1">Weight (g)</label>
                    <input
                      type="number"
                      value={variant.weight}
                      onChange={(e) => updateVariant(index, 'weight', parseInt(e.target.value))}
                      className="w-full px-3 py-2 rounded-lg bg-white/5 border border-white/10 text-white text-sm"
                    />
                  </div>
                  <div>
                    <label className="block text-white/60 text-xs mb-1">Length (cm)</label>
                    <input
                      type="number"
                      value={variant.length}
                      onChange={(e) => updateVariant(index, 'length', parseInt(e.target.value))}
                      className="w-full px-3 py-2 rounded-lg bg-white/5 border border-white/10 text-white text-sm"
                    />
                  </div>
                  <div>
                    <label className="block text-white/60 text-xs mb-1">Width (cm)</label>
                    <input
                      type="number"
                      value={variant.width}
                      onChange={(e) => updateVariant(index, 'width', parseInt(e.target.value))}
                      className="w-full px-3 py-2 rounded-lg bg-white/5 border border-white/10 text-white text-sm"
                    />
                  </div>
                  <div>
                    <label className="block text-white/60 text-xs mb-1">Height (cm)</label>
                    <input
                      type="number"
                      value={variant.height}
                      onChange={(e) => updateVariant(index, 'height', parseInt(e.target.value))}
                      className="w-full px-3 py-2 rounded-lg bg-white/5 border border-white/10 text-white text-sm"
                    />
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Dialog Component */}
      <AlertDialog
        isOpen={dialog.showAlert}
        onClose={dialog.closeAlert}
        title={dialog.alertConfig.title}
        message={dialog.alertConfig.message}
        variant={dialog.alertConfig.variant}
        buttonText={dialog.alertConfig.buttonText}
      />
    </>
  );
}
