"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import {
  Plus,
  Search,
  Edit,
  Trash2,
  Package,
  RefreshCcw,
  ChevronLeft,
  ChevronRight,
  AlertTriangle,
  Check,
  X,
} from "lucide-react";
import {
  getAdminProducts,
  updateStock,
  deleteProduct,
  AdminProduct,
} from "@/lib/adminApi";
import { useDialog } from "@/context/DialogContext";

export default function AdminProductsPage() {
  const dialog = useDialog();
  const [products, setProducts] = useState<AdminProduct[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [showModal, setShowModal] = useState<"stock" | null>(null);
  const [selectedProduct, setSelectedProduct] = useState<AdminProduct | null>(null);
  const [stockAdjustment, setStockAdjustment] = useState({ quantity: 0, reason: "" });
  const [saving, setSaving] = useState(false);
  const pageSize = 20;

  useEffect(() => {
    loadProducts();
  }, [page]);

  const loadProducts = async () => {
    setLoading(true);
    try {
      const result = await getAdminProducts(page, pageSize, "", true);
      setProducts(result.products);
      setTotal(result.total);
    } catch (error) {
      console.error("Failed to load products:", error);
    } finally {
      setLoading(false);
    }
  };

  const filteredProducts = products.filter(
    (p) =>
      search === "" ||
      p.name.toLowerCase().includes(search.toLowerCase()) ||
      p.category.toLowerCase().includes(search.toLowerCase())
  );

  const handleStockUpdate = async () => {
    if (!selectedProduct) return;
    setSaving(true);
    try {
      await updateStock(selectedProduct.id, stockAdjustment.quantity, stockAdjustment.reason);
      setShowModal(null);
      setSelectedProduct(null);
      setStockAdjustment({ quantity: 0, reason: "" });
      loadProducts();
    } catch (error) {
      console.error("Failed to update stock:", error);
      await dialog.alert({
        title: 'Error',
        message: 'Gagal mengupdate stok produk',
        variant: 'error'
      });
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (product: AdminProduct) => {
    const confirmed = await dialog.confirm({
      title: 'Hapus Produk',
      message: `Apakah Anda yakin ingin menghapus "${product.name}"?`,
      variant: 'danger',
      confirmText: 'Ya, Hapus',
      cancelText: 'Batal'
    });
    
    if (!confirmed) return;
    
    try {
      await deleteProduct(product.id);
      loadProducts();
    } catch (error) {
      console.error("Failed to delete product:", error);
      await dialog.alert({
        title: 'Error',
        message: 'Gagal menghapus produk',
        variant: 'error'
      });
    }
  };

  const openEditModal = (product: AdminProduct) => {
    // Redirect to edit page instead of modal
    window.location.href = `/admin/products/edit/${product.id}`;
  };

  const openStockModal = (product: AdminProduct) => {
    setSelectedProduct(product);
    setStockAdjustment({ quantity: 0, reason: "" });
    setShowModal("stock");
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      minimumFractionDigits: 0,
    }).format(amount);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-white">Products</h1>
          <p className="text-white/60 mt-1">Manage your product catalog</p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={loadProducts}
            className="flex items-center gap-2 px-4 py-2 rounded-xl bg-white/10 text-white hover:bg-white/20 transition-colors"
          >
            <RefreshCcw size={18} />
            Refresh
          </button>
          <Link
            href="/admin/products/add"
            className="flex items-center gap-2 px-4 py-2 rounded-xl bg-emerald-500 text-white hover:bg-emerald-600 transition-colors"
          >
            <Plus size={18} />
            Add Product
          </Link>
        </div>
      </div>

      {/* Search */}
      <div className="relative">
        <Search className="absolute left-4 top-1/2 -translate-y-1/2 text-white/40" size={20} />
        <input
          type="text"
          placeholder="Search products..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-full pl-12 pr-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white placeholder-white/40 focus:outline-none focus:border-white/30"
        />
      </div>

      {/* Products Table */}
      <div className="bg-neutral-900 rounded-2xl border border-white/10 overflow-hidden">
        {loading ? (
          <div className="flex items-center justify-center h-64">
            <div className="w-10 h-10 border-2 border-white/20 border-t-white rounded-full animate-spin" />
          </div>
        ) : filteredProducts.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-64">
            <Package className="text-white/20 mb-4" size={48} />
            <p className="text-white/40">No products found</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-white/10">
                  <th className="text-left text-white/60 text-sm font-medium px-6 py-4">Product</th>
                  <th className="text-left text-white/60 text-sm font-medium px-6 py-4">Category</th>
                  <th className="text-left text-white/60 text-sm font-medium px-6 py-4">Price</th>
                  <th className="text-left text-white/60 text-sm font-medium px-6 py-4">Stock</th>
                  <th className="text-left text-white/60 text-sm font-medium px-6 py-4">Status</th>
                  <th className="text-right text-white/60 text-sm font-medium px-6 py-4">Actions</th>
                </tr>
              </thead>
              <tbody>
                {filteredProducts.map((product) => (
                  <tr key={product.id} className="border-b border-white/5 hover:bg-white/5 transition-colors">
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-3">
                        <div className="w-12 h-12 rounded-lg bg-white/10 overflow-hidden">
                          {product.images[0] ? (
                            <img
                              src={product.images[0].image_url}
                              alt={product.name}
                              className="w-full h-full object-cover"
                            />
                          ) : (
                            <div className="w-full h-full flex items-center justify-center">
                              <Package className="text-white/40" size={20} />
                            </div>
                          )}
                        </div>
                        <div>
                          <p className="text-white font-medium">{product.name}</p>
                          <p className="text-white/40 text-sm">{product.weight}g</p>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <span className="px-2 py-1 rounded-lg bg-white/10 text-white/80 text-sm">
                        {product.category}
                      </span>
                    </td>
                    <td className="px-6 py-4">
                      <p className="text-white font-medium">{formatCurrency(product.price)}</p>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-2">
                        {product.stock === 0 ? (
                          <span className="text-white/40 text-sm flex items-center gap-1">
                            <Package size={14} />
                            Variants
                          </span>
                        ) : (
                          <>
                            <span
                              className={`font-medium ${
                                product.stock < 10
                                  ? "text-amber-400"
                                  : "text-white"
                              }`}
                            >
                              {product.stock}
                            </span>
                            {product.stock < 10 && (
                              <AlertTriangle className="text-amber-400" size={16} />
                            )}
                          </>
                        )}
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <span
                        className={`inline-flex items-center gap-1 px-2 py-1 rounded-lg text-sm ${
                          product.is_active
                            ? "bg-emerald-500/20 text-emerald-400"
                            : "bg-red-500/20 text-red-400"
                        }`}
                      >
                        {product.is_active ? <Check size={14} /> : <X size={14} />}
                        {product.is_active ? "Active" : "Inactive"}
                      </span>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center justify-end gap-2">
                        <button
                          onClick={() => openStockModal(product)}
                          className="p-2 rounded-lg bg-blue-500/20 text-blue-400 hover:bg-blue-500/30 transition-colors"
                          title="Update Stock"
                        >
                          <Package size={16} />
                        </button>
                        <button
                          onClick={() => openEditModal(product)}
                          className="p-2 rounded-lg bg-white/10 text-white hover:bg-white/20 transition-colors"
                          title="Edit"
                        >
                          <Edit size={16} />
                        </button>
                        <button
                          onClick={() => handleDelete(product)}
                          className="p-2 rounded-lg bg-red-500/20 text-red-400 hover:bg-red-500/30 transition-colors"
                          title="Delete"
                        >
                          <Trash2 size={16} />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {/* Pagination */}
        {total > pageSize && (
          <div className="flex items-center justify-between px-6 py-4 border-t border-white/10">
            <p className="text-white/60 text-sm">
              Showing {(page - 1) * pageSize + 1} to {Math.min(page * pageSize, total)} of {total} products
            </p>
            <div className="flex items-center gap-2">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
                className="p-2 rounded-lg bg-white/10 text-white disabled:opacity-50 hover:bg-white/20 transition-colors"
              >
                <ChevronLeft size={20} />
              </button>
              <span className="text-white px-4">Page {page}</span>
              <button
                onClick={() => setPage((p) => p + 1)}
                disabled={page * pageSize >= total}
                className="p-2 rounded-lg bg-white/10 text-white disabled:opacity-50 hover:bg-white/20 transition-colors"
              >
                <ChevronRight size={20} />
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Create Modal - REMOVED, now using separate page */}

      {/* Stock Update Modal */}
      {showModal === "stock" && selectedProduct && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6 w-full max-w-md">
            <h3 className="text-xl font-bold text-white mb-2">Update Stock</h3>
            <p className="text-white/60 mb-6">{selectedProduct.name}</p>

            <div className="p-4 rounded-xl bg-white/5 mb-6">
              <p className="text-white/60 text-sm">Current Stock</p>
              <p className="text-2xl font-bold text-white">{selectedProduct.stock}</p>
            </div>

            <div className="space-y-4">
              <div>
                <label className="block text-white/60 text-sm mb-2">
                  Adjustment (+ to add, - to subtract)
                </label>
                <input
                  type="number"
                  value={stockAdjustment.quantity}
                  onChange={(e) =>
                    setStockAdjustment({ ...stockAdjustment, quantity: Number(e.target.value) })
                  }
                  className="w-full px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white focus:outline-none focus:border-white/30"
                />
              </div>
              <div>
                <label className="block text-white/60 text-sm mb-2">Reason</label>
                <input
                  type="text"
                  value={stockAdjustment.reason}
                  onChange={(e) =>
                    setStockAdjustment({ ...stockAdjustment, reason: e.target.value })
                  }
                  placeholder="e.g., Restock, Damage, Adjustment"
                  className="w-full px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white placeholder-white/40 focus:outline-none focus:border-white/30"
                />
              </div>

              <div className="p-4 rounded-xl bg-blue-500/10 border border-blue-500/20">
                <p className="text-blue-400 text-sm">
                  New Stock: {selectedProduct.stock + stockAdjustment.quantity}
                </p>
              </div>
            </div>

            <div className="flex gap-3 mt-6">
              <button
                onClick={() => {
                  setShowModal(null);
                  setSelectedProduct(null);
                }}
                className="flex-1 px-4 py-3 rounded-xl bg-white/10 text-white hover:bg-white/20 transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleStockUpdate}
                disabled={saving || stockAdjustment.quantity === 0}
                className="flex-1 px-4 py-3 rounded-xl bg-blue-500 text-white font-medium hover:bg-blue-600 transition-colors disabled:opacity-50"
              >
                {saving ? "Updating..." : "Update Stock"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
