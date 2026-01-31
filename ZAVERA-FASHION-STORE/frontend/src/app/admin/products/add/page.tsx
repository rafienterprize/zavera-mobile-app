"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Upload, X, Package, Plus, Trash2 } from "lucide-react";
import { createProduct } from "@/lib/adminApi";
import { variantApi } from "@/lib/variantApi";
import { useDialog } from "@/hooks/useDialog";
import { AlertDialog, ConfirmDialog } from "@/components/Dialog";

// Category mapping: Display Label (ID) -> Database Value (EN) for subcategories
// MUST match exactly with FilterPanel.tsx SUBCATEGORY_MAPPING
const CATEGORIES = {
  wanita: { 
    label: 'Wanita', 
    subcategories: [
      { label: 'Dress', value: 'Dress' },
      { label: 'Atasan', value: 'Tops' },
      { label: 'Bawahan', value: 'Bottoms' },
      { label: 'Outerwear', value: 'Outerwear' },
      { label: 'Aksesoris', value: 'Accessories' }
    ]
  },
  pria: { 
    label: 'Pria', 
    subcategories: [
      { label: 'Atasan', value: 'Tops' },
      { label: 'Kemeja', value: 'Shirts' },
      { label: 'Celana', value: 'Bottoms' },
      { label: 'Jaket', value: 'Outerwear' },
      { label: 'Jas', value: 'Suits' },
      { label: 'Sepatu', value: 'Footwear' }
    ]
  },
  anak: { 
    label: 'Anak', 
    subcategories: [
      { label: 'Anak Laki-laki', value: 'Boys' },
      { label: 'Anak Perempuan', value: 'Girls' },
      { label: 'Bayi', value: 'Baby' },
      { label: 'Sepatu', value: 'Footwear' }
    ]
  },
  sports: { 
    label: 'Sports', 
    subcategories: [
      { label: 'Pakaian Olahraga', value: 'Activewear' },
      { label: 'Sepatu', value: 'Footwear' },
      { label: 'Jaket', value: 'Outerwear' },
      { label: 'Aksesoris', value: 'Accessories' }
    ]
  },
  luxury: { 
    label: 'Luxury', 
    subcategories: [
      { label: 'Aksesoris', value: 'Accessories' },
      { label: 'Outerwear', value: 'Outerwear' }
    ]
  },
  beauty: { 
    label: 'Beauty', 
    subcategories: [
      { label: 'Perawatan Kulit', value: 'Skincare' },
      { label: 'Makeup', value: 'Makeup' },
      { label: 'Parfum', value: 'Fragrance' }
    ]
  },
};

const SIZES = ['XS', 'S', 'M', 'L', 'XL', 'XXL', 'XXXL'];
const COLORS = [
  { name: 'Black', hex: '#000000' }, { name: 'White', hex: '#FFFFFF' },
  { name: 'Navy', hex: '#000080' }, { name: 'Red', hex: '#FF0000' },
  { name: 'Blue', hex: '#0000FF' }, { name: 'Green', hex: '#008000' },
  { name: 'Gray', hex: '#808080' }, { name: 'Pink', hex: '#FFC0CB' },
];

export default function AddProductPage() {
  const router = useRouter();
  const dialog = useDialog();
  const [formData, setFormData] = useState({
    name: "", description: "", price: "", category: "wanita",
    subcategory: "", brand: "", material: "", pattern: "",
  });
  const [uploadedImages, setUploadedImages] = useState<string[]>([]);
  const [variants, setVariants] = useState<any[]>([]);
  const [uploading, setUploading] = useState(false);
  const [saving, setSaving] = useState(false);

  const handleImageUpload = async (files: FileList) => {
    if (files.length === 0) return;
    setUploading(true);
    try {
      const token = localStorage.getItem("auth_token");
      const uploadPromises = Array.from(files).map(async (file) => {
        const fd = new FormData();
        fd.append("image", file);
        const res = await fetch("http://localhost:8080/api/admin/products/upload-image", {
          method: "POST", headers: { Authorization: `Bearer ${token}` }, body: fd,
        });
        if (!res.ok) throw new Error("Upload failed");
        const data = await res.json();
        return data.image_url;
      });
      const urls = await Promise.all(uploadPromises);
      setUploadedImages((prev) => [...prev, ...urls]);
    } catch (error) {
      console.error("Failed:", error);
      await dialog.alert({
        title: 'Error',
        message: 'Gagal mengupload gambar',
        variant: 'error',
      });
    } finally {
      setUploading(false);
    }
  };

  const addVariant = () => {
    setVariants([...variants, {
      size: 'M', color: 'Black', color_hex: '#000000',
      stock: 10, price: Number(formData.price) || 0,
      weight: 400, length: 70, width: 45, height: 3
    }]);
  };

  const updateVariant = (index: number, field: string, value: any) => {
    const newVariants = [...variants];
    newVariants[index] = { ...newVariants[index], [field]: value };
    setVariants(newVariants);
  };

  const updateVariantColor = (index: number, colorName: string) => {
    const color = COLORS.find(c => c.name === colorName);
    const newVariants = [...variants];
    newVariants[index] = { 
      ...newVariants[index], 
      color: colorName,
      color_hex: color?.hex || '#000000'
    };
    setVariants(newVariants);
    console.log('🎨 Updated variant color:', newVariants[index]);
  };

  const removeVariant = (index: number) => {
    setVariants(variants.filter((_, i) => i !== index));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    // Validation
    if (!formData.name.trim()) {
      await dialog.alert({
        title: 'Validasi Error',
        message: 'Nama produk harus diisi',
        variant: 'error',
      });
      return;
    }
    
    if (!formData.price || Number(formData.price) <= 0) {
      await dialog.alert({
        title: 'Validasi Error',
        message: 'Harga harus lebih besar dari 0',
        variant: 'error',
      });
      return;
    }
    
    if (!formData.category) {
      await dialog.alert({
        title: 'Validasi Error',
        message: 'Kategori harus dipilih',
        variant: 'error',
      });
      return;
    }
    
    if (uploadedImages.length === 0) {
      await dialog.alert({
        title: 'Validasi Error',
        message: 'Minimal upload 1 gambar produk',
        variant: 'error',
      });
      return;
    }
    
    if (variants.length === 0) {
      await dialog.alert({
        title: 'Validasi Error',
        message: 'Minimal tambahkan 1 variant produk',
        variant: 'error',
      });
      return;
    }

    setSaving(true);
    
    let productId: number | null = null;
    
    try {
      const token = localStorage.getItem("auth_token");
      
      // Only send fields that backend expects
      const productData = {
        name: formData.name,
        description: formData.description || "",
        price: Number(formData.price),
        stock: 0,
        weight: 400,
        length: 70,
        width: 45,
        height: 3,
        category: formData.category,
        subcategory: formData.subcategory || "",
        brand: formData.brand || "",
        material: formData.material || "",
        images: uploadedImages,
      };
      
      console.log('=== CREATING PRODUCT ===');
      console.log('Product data:', JSON.stringify(productData, null, 2));
      
      // Try to create product first
      const productRes = await createProduct(productData);
      console.log('📦 Full product response:', productRes);
      console.log('📦 Response data:', productRes.data);
      console.log('📦 Response data.data:', productRes.data?.data);
      
      // Extract product ID from response
      // Backend returns: { success: true, message: "...", data: { id: X, ... } }
      // Axios wraps it, so: productRes.data = backend response
      productId = productRes.data?.data?.id;
      
      console.log('✅ Product created successfully! ID:', productId);
      console.log('✅ Product ID type:', typeof productId);
      console.log('✅ Full response structure:', JSON.stringify(productRes.data, null, 2));

      // Validate product ID before creating variants
      if (!productId || typeof productId !== 'number') {
        console.error('❌ Invalid product ID:', productId);
        console.error('❌ Response structure:', {
          hasData: !!productRes.data,
          hasDataData: !!productRes.data?.data,
          dataKeys: productRes.data ? Object.keys(productRes.data) : [],
          dataDataKeys: productRes.data?.data ? Object.keys(productRes.data.data) : [],
        });
        await dialog.alert({
          title: 'Error',
          message: `Product ID tidak valid: ${productId}. Tidak bisa membuat variant. Silakan cek console untuk detail.`,
          variant: 'error',
        });
        return;
      }

      // Only create variants if product was created successfully
      let variantSuccessCount = 0;
      let variantFailCount = 0;
      
      for (let i = 0; i < variants.length; i++) {
        const variant = variants[i];
        console.log(`\n=== Creating variant ${i + 1}/${variants.length} ===`);
        console.log('Variant data:', variant);
        console.log('Product ID:', productId);
        
        // Prepare variant payload with correct types
        const variantPayload: any = {
          product_id: productId, 
          stock_quantity: variant.stock || 0, 
          is_active: true, 
          low_stock_threshold: 5, 
          position: i,
        };
        
        // Add optional fields only if they have values
        if (variant.size) variantPayload.size = variant.size;
        if (variant.color) variantPayload.color = variant.color;
        if (variant.color_hex) variantPayload.color_hex = variant.color_hex;
        if (variant.price) variantPayload.price = variant.price;
        if (variant.weight) variantPayload.weight_grams = variant.weight;
        
        console.log('📦 Variant payload:', JSON.stringify(variantPayload, null, 2));
        console.log('📦 Payload keys:', Object.keys(variantPayload));
        console.log('📦 product_id type:', typeof variantPayload.product_id);
        console.log('📦 product_id value:', variantPayload.product_id);
        
        try {
          const variantRes = await variantApi.createVariant(token!, variantPayload);
          console.log(`✅ Variant ${i + 1} created successfully:`, variantRes);
          variantSuccessCount++;
        } catch (variantError: any) {
          console.error(`❌ Failed to create variant ${i + 1}:`, variantError);
          console.error('Variant data:', variant);
          console.error('Variant payload:', variantPayload);
          console.error('Error response:', variantError.response?.data);
          console.error('Error status:', variantError.response?.status);
          console.error('Error message:', variantError.message);
          variantFailCount++;
        }
      }

      console.log(`✅ Variants processed: ${variantSuccessCount} success, ${variantFailCount} failed`);

      // Show appropriate message based on variant creation results
      if (variantFailCount > 0) {
        await dialog.alert({
          title: 'Produk Dibuat dengan Peringatan',
          message: `Produk berhasil dibuat, tetapi ${variantFailCount} dari ${variants.length} variant gagal dibuat. Anda bisa menambahkan variant nanti di halaman edit produk.`,
          variant: 'warning',
        });
      } else {
        await dialog.alert({
          title: 'Berhasil!',
          message: 'Produk dan semua variant berhasil dibuat!',
          variant: 'success',
        });
      }
      
      router.push("/admin/products");
      
    } catch (error: any) {
      console.error("=== CREATE PRODUCT ERROR ===");
      console.error("Error:", error);
      console.error("Error response:", error.response);
      console.error("Error data:", error.response?.data);
      console.error("Error status:", error.response?.status);
      console.error("Error message:", error.message);
      
      // Parse error message for better UX
      let errorMsg = error.response?.data?.message || error.response?.data?.error || error.message;
      let errorTitle = 'Error';
      
      // Handle specific error cases
      if (errorMsg.includes('slug already exists') || errorMsg.includes('Produk dengan nama yang sama sudah ada')) {
        errorTitle = 'Produk Sudah Ada';
        errorMsg = `Produk dengan nama "${formData.name}" sudah ada di database.\n\nSilakan gunakan nama yang berbeda atau edit produk yang sudah ada.`;
      } else if (errorMsg.includes('duplicate')) {
        errorTitle = 'Data Duplikat';
        errorMsg = 'Produk dengan data yang sama sudah ada. Silakan periksa kembali data produk Anda.';
      } else if (errorMsg.includes('invalid')) {
        errorTitle = 'Data Tidak Valid';
        errorMsg = `Data produk tidak valid: ${errorMsg}`;
      }
      
      await dialog.alert({
        title: errorTitle,
        message: errorMsg,
        variant: 'error',
      });
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="min-h-screen bg-black">
      {/* Header */}
      <div className="border-b border-white/10 bg-neutral-900/50 backdrop-blur-sm sticky top-0 z-10">
        <div className="max-w-7xl mx-auto px-6 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <Link href="/admin/products" className="p-2 rounded-xl bg-white/5 hover:bg-white/10">
                <ArrowLeft size={20} className="text-white" />
              </Link>
              <div>
                <h1 className="text-2xl font-bold text-white">Add New Product</h1>
                <p className="text-white/60 text-sm mt-1">Fill in all product information</p>
              </div>
            </div>
            <div className="flex gap-3">
              <Link href="/admin/products" className="px-6 py-2.5 rounded-xl bg-white/10 text-white hover:bg-white/20">
                Cancel
              </Link>
              <button onClick={handleSubmit} disabled={saving || !formData.name || !formData.price || uploadedImages.length === 0}
                className="px-6 py-2.5 rounded-xl bg-emerald-500 text-white hover:bg-emerald-600 disabled:opacity-50">
                {saving ? "Creating..." : "Create Product"}
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="max-w-7xl mx-auto px-6 py-8">
        <form onSubmit={handleSubmit} className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Left Column - Main Info */}
          <div className="lg:col-span-2 space-y-6">
            {/* Basic Information */}
            <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
              <h2 className="text-lg font-semibold text-white mb-4">Basic Information</h2>
              <div className="space-y-4">
                <div>
                  <label className="block text-white/80 text-sm font-medium mb-2">
                    Product Name <span className="text-red-400">*</span>
                  </label>
                  <input type="text" value={formData.name} onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    placeholder="e.g., Classic Denim Jacket"
                    className="w-full px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white placeholder-white/40 focus:outline-none focus:border-emerald-500" required />
                </div>

                <div>
                  <label className="block text-white/80 text-sm font-medium mb-2">Description</label>
                  <textarea value={formData.description} onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                    placeholder="Describe your product..." rows={5}
                    className="w-full px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white placeholder-white/40 focus:outline-none focus:border-emerald-500 resize-none" />
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-white/80 text-sm font-medium mb-2">
                      Category <span className="text-red-400">*</span>
                    </label>
                    <select value={formData.category} onChange={(e) => setFormData({ ...formData, category: e.target.value, subcategory: '' })}
                      className="w-full px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white focus:outline-none focus:border-emerald-500" required>
                      {Object.entries(CATEGORIES).map(([key, cat]) => (
                        <option key={key} value={key}>{cat.label}</option>
                      ))}
                    </select>
                  </div>

                  <div>
                    <label className="block text-white/80 text-sm font-medium mb-2">
                      Subcategory <span className="text-red-400">*</span>
                    </label>
                    <select value={formData.subcategory} onChange={(e) => setFormData({ ...formData, subcategory: e.target.value })}
                      className="w-full px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white focus:outline-none focus:border-emerald-500" required>
                      <option value="">Pilih subcategory</option>
                      {CATEGORIES[formData.category as keyof typeof CATEGORIES]?.subcategories.map((sub) => (
                        <option key={sub.value} value={sub.value}>{sub.label}</option>
                      ))}
                    </select>
                  </div>
                </div>

                <div className="grid grid-cols-3 gap-4">
                  <div>
                    <label className="block text-white/80 text-sm font-medium mb-2">
                      Base Price (IDR) <span className="text-red-400">*</span>
                    </label>
                    <input type="number" value={formData.price} onChange={(e) => setFormData({ ...formData, price: e.target.value })}
                      placeholder="0" className="w-full px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white focus:outline-none focus:border-emerald-500" required />
                  </div>
                  <div>
                    <label className="block text-white/80 text-sm font-medium mb-2">Brand</label>
                    <input type="text" value={formData.brand} onChange={(e) => setFormData({ ...formData, brand: e.target.value })}
                      placeholder="e.g., Nike" className="w-full px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white focus:outline-none focus:border-emerald-500" />
                  </div>
                  <div>
                    <label className="block text-white/80 text-sm font-medium mb-2">Material</label>
                    <input type="text" value={formData.material} onChange={(e) => setFormData({ ...formData, material: e.target.value })}
                      placeholder="e.g., Cotton" className="w-full px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white focus:outline-none focus:border-emerald-500" />
                  </div>
                </div>
              </div>
            </div>

            {/* Product Variants */}
            <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
              <div className="flex items-center justify-between mb-4">
                <div>
                  <h2 className="text-lg font-semibold text-white flex items-center gap-2">
                    <Package size={20} />
                    Product Variants
                  </h2>
                  <p className="text-white/60 text-sm mt-1">Each variant has its own stock, price, and dimensions</p>
                </div>
                <button type="button" onClick={addVariant}
                  className="px-4 py-2 bg-emerald-500 text-white rounded-lg hover:bg-emerald-600 flex items-center gap-2">
                  <Plus size={16} /> Add Variant
                </button>
              </div>

              {variants.length === 0 ? (
                <div className="text-center py-8 text-white/40">
                  <Package size={48} className="mx-auto mb-3 opacity-50" />
                  <p>No variants yet. Click &quot;Add Variant&quot; to create one.</p>
                </div>
              ) : (
                <div className="space-y-4">
                  {variants.map((variant, index) => (
                    <div key={index} className="bg-white/5 rounded-xl p-4 border border-white/10">
                      <div className="flex items-start justify-between mb-4">
                        <h3 className="text-white font-medium">Variant #{index + 1}</h3>
                        <button type="button" onClick={() => removeVariant(index)}
                          className="p-1.5 text-red-400 hover:bg-red-500/10 rounded">
                          <Trash2 size={16} />
                        </button>
                      </div>

                      <div className="grid grid-cols-4 gap-3">
                        <div>
                          <label className="block text-white/60 text-xs mb-1">Size</label>
                          <select value={variant.size} onChange={(e) => updateVariant(index, 'size', e.target.value)}
                            className="w-full px-3 py-2 rounded-lg bg-white/5 border border-white/10 text-white text-sm">
                            {SIZES.map(s => <option key={s} value={s}>{s}</option>)}
                          </select>
                        </div>
                        <div>
                          <label className="block text-white/60 text-xs mb-1">Color</label>
                          <select 
                            value={variant.color} 
                            onChange={(e) => updateVariantColor(index, e.target.value)}
                            className="w-full px-3 py-2 rounded-lg bg-white/5 border border-white/10 text-white text-sm">
                            {COLORS.map(c => <option key={c.name} value={c.name}>{c.name}</option>)}
                          </select>
                        </div>
                        <div>
                          <label className="block text-white/60 text-xs mb-1">Stock</label>
                          <input type="number" value={variant.stock} onChange={(e) => updateVariant(index, 'stock', parseInt(e.target.value))}
                            className="w-full px-3 py-2 rounded-lg bg-white/5 border border-white/10 text-white text-sm" />
                        </div>
                        <div>
                          <label className="block text-white/60 text-xs mb-1">Price (IDR)</label>
                          <input type="number" value={variant.price} onChange={(e) => updateVariant(index, 'price', parseFloat(e.target.value))}
                            className="w-full px-3 py-2 rounded-lg bg-white/5 border border-white/10 text-white text-sm" />
                        </div>
                      </div>

                      <div className="grid grid-cols-4 gap-3 mt-3">
                        <div>
                          <label className="block text-white/60 text-xs mb-1">Weight (g)</label>
                          <input type="number" value={variant.weight} onChange={(e) => updateVariant(index, 'weight', parseInt(e.target.value))}
                            className="w-full px-3 py-2 rounded-lg bg-white/5 border border-white/10 text-white text-sm" />
                        </div>
                        <div>
                          <label className="block text-white/60 text-xs mb-1">Length (cm)</label>
                          <input type="number" value={variant.length} onChange={(e) => updateVariant(index, 'length', parseInt(e.target.value))}
                            className="w-full px-3 py-2 rounded-lg bg-white/5 border border-white/10 text-white text-sm" />
                        </div>
                        <div>
                          <label className="block text-white/60 text-xs mb-1">Width (cm)</label>
                          <input type="number" value={variant.width} onChange={(e) => updateVariant(index, 'width', parseInt(e.target.value))}
                            className="w-full px-3 py-2 rounded-lg bg-white/5 border border-white/10 text-white text-sm" />
                        </div>
                        <div>
                          <label className="block text-white/60 text-xs mb-1">Height (cm)</label>
                          <input type="number" value={variant.height} onChange={(e) => updateVariant(index, 'height', parseInt(e.target.value))}
                            className="w-full px-3 py-2 rounded-lg bg-white/5 border border-white/10 text-white text-sm" />
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>

          {/* Right Column - Images */}
          <div className="space-y-6">
            <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
              <h2 className="text-lg font-semibold text-white mb-4">
                Product Images <span className="text-red-400">*</span>
              </h2>

              <input type="file" accept="image/*" multiple onChange={(e) => e.target.files && handleImageUpload(e.target.files)}
                disabled={uploading} className="hidden" id="file-upload" />
              <label htmlFor="file-upload" className="block border-2 border-dashed border-white/20 rounded-xl p-8 text-center cursor-pointer hover:border-emerald-500 transition-colors">
                <Upload className="mx-auto text-emerald-400 mb-2" size={32} />
                <p className="text-white font-medium mb-1">{uploading ? "Uploading..." : "Upload Images"}</p>
                <p className="text-white/40 text-xs">Drag & drop or click</p>
              </label>

              {uploadedImages.length > 0 && (
                <div className="mt-4 space-y-2">
                  <p className="text-white/60 text-sm">Uploaded ({uploadedImages.length})</p>
                  <div className="grid grid-cols-2 gap-3">
                    {uploadedImages.map((url, index) => (
                      <div key={index} className="relative group">
                        <img src={url} alt={`Product ${index + 1}`} className="w-full h-32 object-cover rounded-xl" />
                        <button type="button" onClick={() => setUploadedImages(uploadedImages.filter((_, i) => i !== index))}
                          className="absolute top-2 right-2 p-1.5 rounded-full bg-red-500 text-white opacity-0 group-hover:opacity-100 transition-opacity">
                          <X size={16} />
                        </button>
                        {index === 0 && (
                          <div className="absolute bottom-2 left-2 px-2 py-1 rounded-lg bg-emerald-500 text-white text-xs font-medium">
                            Primary
                          </div>
                        )}
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </div>
        </form>
      </div>

      {/* Custom Dialog Components */}
      <AlertDialog
        isOpen={dialog.showAlert}
        onClose={dialog.closeAlert}
        title={dialog.alertConfig.title}
        message={dialog.alertConfig.message}
        variant={dialog.alertConfig.variant}
        buttonText={dialog.alertConfig.buttonText}
      />

      <ConfirmDialog
        isOpen={dialog.showConfirm}
        onClose={dialog.closeConfirm}
        onConfirm={dialog.confirmConfig.onConfirm}
        title={dialog.confirmConfig.title}
        message={dialog.confirmConfig.message}
        confirmText={dialog.confirmConfig.confirmText}
        cancelText={dialog.confirmConfig.cancelText}
        variant={dialog.confirmConfig.variant}
      />
    </div>
  );
}
