'use client';

import { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import { ArrowLeft, Package } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import { useDialog } from '@/hooks/useDialog';
import { getProduct } from '@/lib/adminApi';
import VariantManagerNew from '@/components/admin/VariantManagerNew';

export default function EditProductPage() {
  const params = useParams();
  const router = useRouter();
  const dialog = useDialog();
  const { user, isLoading: authLoading } = useAuth();
  const [product, setProduct] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState('basic');

  useEffect(() => {
    const loadProduct = async () => {
      console.log('=== LOAD PRODUCT START ===');
      
      // Get token from localStorage
      const token = localStorage.getItem('auth_token');
      console.log('Token:', token ? 'EXISTS' : 'NULL');
      console.log('Params ID:', params.id);
      
      if (!token) {
        console.log('No token, setting loading false');
        setLoading(false);
        return;
      }
      
      if (!params.id) {
        console.log('No params.id, setting loading false');
        setLoading(false);
        return;
      }
      
      try {
        console.log('Calling getProduct with ID:', params.id);
        const data = await getProduct(token, Number(params.id));
        console.log('Product loaded successfully:', data);
        setProduct(data);
      } catch (error: any) {
        console.error('Failed to load product:', error);
        console.error('Error response:', error.response?.data);
        console.error('Error status:', error.response?.status);
      } finally {
        console.log('Setting loading to false');
        setLoading(false);
      }
    };

    // Wait for auth to finish loading
    if (!authLoading) {
      loadProduct();
    }
  }, [authLoading, params.id]);

  if (loading || authLoading) {
    const token = typeof window !== 'undefined' ? localStorage.getItem('auth_token') : null;
    return (
      <div className="min-h-screen bg-black flex items-center justify-center">
        <div className="text-white">
          <div>Loading...</div>
          <div className="text-xs text-white/40 mt-2">Token: {token ? 'Yes' : 'No'}</div>
          <div className="text-xs text-white/40">ID: {params.id}</div>
          <div className="text-xs text-white/40">Auth Loading: {authLoading ? 'Yes' : 'No'}</div>
        </div>
      </div>
    );
  }

  if (!product) {
    const token = typeof window !== 'undefined' ? localStorage.getItem('auth_token') : null;
    return (
      <div className="min-h-screen bg-black flex items-center justify-center">
        <div className="text-white text-center">
          <div className="text-xl mb-4">Product not found</div>
          <div className="text-sm text-white/60 space-y-1">
            <div>Token: {token ? 'Available' : 'Missing'}</div>
            <div>User: {user ? user.email : 'Not logged in'}</div>
            <div>Product ID: {params.id}</div>
            <div>Product State: {product === null ? 'null' : 'undefined'}</div>
          </div>
          <Link href="/admin/products" className="mt-4 inline-block px-4 py-2 bg-white/10 rounded hover:bg-white/20">
            Back to Products
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-black">
      {/* Header */}
      <div className="border-b border-white/10 bg-neutral-900/50 backdrop-blur-sm sticky top-0 z-10">
        <div className="max-w-7xl mx-auto px-6 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <Link
                href="/admin/products"
                className="p-2 rounded-xl bg-white/5 hover:bg-white/10 transition-colors"
              >
                <ArrowLeft size={20} className="text-white" />
              </Link>
              <div>
                <h1 className="text-2xl font-bold text-white">{product.name}</h1>
                <p className="text-white/60 text-sm mt-1">Edit product details and variants</p>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b border-white/10 bg-neutral-900/30">
        <div className="max-w-7xl mx-auto px-6">
          <div className="flex gap-6">
            <button
              onClick={() => setActiveTab('basic')}
              className={`px-4 py-4 text-sm font-medium border-b-2 transition-colors ${
                activeTab === 'basic'
                  ? 'border-emerald-500 text-white'
                  : 'border-transparent text-white/60 hover:text-white'
              }`}
            >
              Basic Info
            </button>
            <button
              onClick={() => setActiveTab('variants')}
              className={`px-4 py-4 text-sm font-medium border-b-2 transition-colors flex items-center gap-2 ${
                activeTab === 'variants'
                  ? 'border-emerald-500 text-white'
                  : 'border-transparent text-white/60 hover:text-white'
              }`}
            >
              <Package size={16} />
              Variants & Stock
            </button>
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="max-w-7xl mx-auto px-6 py-8">
        {activeTab === 'basic' && (
          <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
            <h2 className="text-xl font-semibold text-white mb-6">Edit Product Information</h2>
            
            <div className="space-y-6">
              {/* Product Name */}
              <div>
                <label className="block text-white/80 text-sm font-medium mb-2">
                  Product Name <span className="text-red-400">*</span>
                </label>
                <input
                  type="text"
                  defaultValue={product.name}
                  className="w-full px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white placeholder-white/40 focus:outline-none focus:border-emerald-500"
                  placeholder="e.g., Classic Denim Jacket"
                />
              </div>

              {/* Description */}
              <div>
                <label className="block text-white/80 text-sm font-medium mb-2">Description</label>
                <textarea
                  defaultValue={product.description}
                  rows={5}
                  className="w-full px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white placeholder-white/40 focus:outline-none focus:border-emerald-500 resize-none"
                  placeholder="Describe your product..."
                />
              </div>

              {/* Category & Subcategory */}
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-white/80 text-sm font-medium mb-2">Category</label>
                  <input
                    type="text"
                    defaultValue={product.category}
                    className="w-full px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white"
                    readOnly
                  />
                </div>
                <div>
                  <label className="block text-white/80 text-sm font-medium mb-2">Subcategory</label>
                  <input
                    type="text"
                    defaultValue={product.subcategory}
                    className="w-full px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white"
                  />
                </div>
              </div>

              {/* Base Price */}
              <div>
                <label className="block text-white/80 text-sm font-medium mb-2">
                  Base Price (IDR) <span className="text-red-400">*</span>
                </label>
                <input
                  type="number"
                  defaultValue={product.price}
                  className="w-full px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white focus:outline-none focus:border-emerald-500"
                />
                <p className="text-white/40 text-xs mt-2">
                  This is the base price. Variants can have their own prices.
                </p>
              </div>

              {/* Product Images */}
              <div>
                <label className="block text-white/80 text-sm font-medium mb-2">Product Images</label>
                <div className="grid grid-cols-4 gap-3">
                  {product.images && product.images.length > 0 ? (
                    product.images.map((img: string, idx: number) => (
                      <div key={idx} className="relative group aspect-square">
                        <img
                          src={img}
                          alt={`Product ${idx + 1}`}
                          className="w-full h-full object-cover rounded-xl border border-white/10"
                        />
                        {idx === 0 && (
                          <div className="absolute top-2 left-2 px-2 py-1 rounded-lg bg-emerald-500 text-white text-xs font-medium">
                            Primary
                          </div>
                        )}
                      </div>
                    ))
                  ) : (
                    <div className="col-span-4 text-white/40 text-sm">No images uploaded</div>
                  )}
                </div>
              </div>

              {/* Product dimensions are managed at variant level */}

              {/* Save Button */}
              <div className="flex justify-end gap-3 pt-4 border-t border-white/10">
                <Link
                  href="/admin/products"
                  className="px-6 py-2.5 rounded-xl bg-white/10 text-white hover:bg-white/20 transition-colors"
                >
                  Cancel
                </Link>
                <button
                  type="button"
                  className="px-6 py-2.5 rounded-xl bg-emerald-500 text-white hover:bg-emerald-600 transition-colors"
                  onClick={async () => {
                    await dialog.alert({
                      title: 'Coming Soon',
                      message: 'Fitur save akan segera hadir',
                    });
                  }}
                >
                  Save Changes
                </button>
              </div>
            </div>
          </div>
        )}

        {activeTab === 'variants' && (
          <VariantManagerNew 
            productId={product.id} 
            productName={product.name}
            productPrice={product.price} 
          />
        )}
      </div>
    </div>
  );
}
