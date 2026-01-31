'use client';

import { useState, useEffect } from 'react';
import { useAuth } from '@/context/AuthContext';
import { variantApi } from '@/lib/variantApi';
import { LowStockVariant } from '@/types/variant';

export default function AdminVariantsPage() {
  const { token } = useAuth();
  const [lowStockVariants, setLowStockVariants] = useState<LowStockVariant[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadLowStockVariants();
  }, []);

  const loadLowStockVariants = async () => {
    if (!token) return;
    try {
      setLoading(true);
      const data = await variantApi.getLowStockVariants(token);
      setLowStockVariants(data);
    } catch (error) {
      console.error('Failed to load low stock variants:', error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="p-8">
        <div className="text-center">Loading...</div>
      </div>
    );
  }

  return (
    <div className="p-8">
      <div className="mb-6">
        <h1 className="text-3xl font-bold">Variant Management</h1>
        <p className="text-gray-600 mt-2">Manage product variants and stock levels</p>
      </div>

      <div className="bg-white rounded-lg shadow">
        <div className="p-6 border-b">
          <h2 className="text-xl font-semibold">Low Stock Alerts</h2>
          <p className="text-sm text-gray-600 mt-1">
            Variants with stock below threshold
          </p>
        </div>

        {lowStockVariants.length === 0 ? (
          <div className="p-8 text-center text-gray-500">
            No low stock variants found
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Product
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    SKU
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Variant
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Size
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Color
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Stock
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Available
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Threshold
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {lowStockVariants.map((variant) => (
                  <tr key={variant.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 text-sm">{variant.product_name}</td>
                    <td className="px-6 py-4 text-sm font-mono">{variant.sku}</td>
                    <td className="px-6 py-4 text-sm">{variant.variant_name}</td>
                    <td className="px-6 py-4 text-sm">{variant.size || '-'}</td>
                    <td className="px-6 py-4 text-sm">{variant.color || '-'}</td>
                    <td className="px-6 py-4 text-sm">
                      <span
                        className={`font-semibold ${
                          variant.stock_quantity === 0
                            ? 'text-red-600'
                            : variant.stock_quantity <= variant.low_stock_threshold
                            ? 'text-orange-600'
                            : 'text-green-600'
                        }`}
                      >
                        {variant.stock_quantity}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-sm">{variant.available_stock}</td>
                    <td className="px-6 py-4 text-sm text-gray-500">
                      {variant.low_stock_threshold}
                    </td>
                    <td className="px-6 py-4 text-sm">
                      <a
                        href={`/admin/products/edit/${variant.product_id}?tab=variants`}
                        className="text-blue-600 hover:text-blue-800"
                      >
                        Manage
                      </a>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
