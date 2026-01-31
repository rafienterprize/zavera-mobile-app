"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import {
  Search,
  Filter,
  ChevronLeft,
  ChevronRight,
  Eye,
  RefreshCcw,
  XCircle,
  Truck,
  Package,
  Clock,
  CheckCircle,
  AlertCircle,
} from "lucide-react";
import api from "@/lib/api";

interface Order {
  id: number;
  order_code: string;
  customer_name: string;
  customer_email: string;
  total_amount: number;
  status: string;
  created_at: string;
}

const statusColors: Record<string, string> = {
  PENDING: "bg-amber-500/20 text-amber-400",
  PAID: "bg-emerald-500/20 text-emerald-400",
  PROCESSING: "bg-blue-500/20 text-blue-400",
  SHIPPED: "bg-purple-500/20 text-purple-400",
  DELIVERED: "bg-emerald-500/20 text-emerald-400",
  CANCELLED: "bg-red-500/20 text-red-400",
  REFUNDED: "bg-gray-500/20 text-gray-400",
};

const statusIcons: Record<string, React.ReactNode> = {
  PENDING: <Clock size={14} />,
  PAID: <CheckCircle size={14} />,
  PROCESSING: <Package size={14} />,
  SHIPPED: <Truck size={14} />,
  DELIVERED: <CheckCircle size={14} />,
  CANCELLED: <XCircle size={14} />,
  REFUNDED: <RefreshCcw size={14} />,
};

export default function OrdersPage() {
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState("");
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const pageSize = 20;

  useEffect(() => {
    loadOrders();
  }, [page, statusFilter]);

  const loadOrders = async () => {
    setLoading(true);
    try {
      const token = localStorage.getItem("auth_token");
      const params = new URLSearchParams({
        page: page.toString(),
        page_size: pageSize.toString(),
      });
      if (statusFilter) params.append("status", statusFilter);
      if (search) params.append("search", search);
      
      const response = await api.get(`/admin/orders?${params.toString()}`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      setOrders(response.data.orders || []);
      setTotal(response.data.total_count || 0);
    } catch (error) {
      console.error("Failed to load orders:", error);
    } finally {
      setLoading(false);
    }
  };

  const filteredOrders = orders.filter((order) => {
    const matchesSearch =
      search === "" ||
      order.order_code.toLowerCase().includes(search.toLowerCase()) ||
      order.customer_name.toLowerCase().includes(search.toLowerCase()) ||
      order.customer_email.toLowerCase().includes(search.toLowerCase());

    const matchesStatus = statusFilter === "" || order.status === statusFilter;

    return matchesSearch && matchesStatus;
  });

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      minimumFractionDigits: 0,
    }).format(amount);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("id-ID", {
      day: "numeric",
      month: "short",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-white">Orders</h1>
          <p className="text-white/60 mt-1">Manage and track all orders</p>
        </div>
        <button
          onClick={loadOrders}
          className="flex items-center gap-2 px-4 py-2 rounded-xl bg-white/10 text-white hover:bg-white/20 transition-colors"
        >
          <RefreshCcw size={18} />
          Refresh
        </button>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-4 top-1/2 -translate-y-1/2 text-white/40" size={20} />
          <input
            type="text"
            placeholder="Search by order code, customer name, or email..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-12 pr-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white placeholder-white/40 focus:outline-none focus:border-white/30"
          />
        </div>

        <div className="relative">
          <Filter className="absolute left-4 top-1/2 -translate-y-1/2 text-white/40 pointer-events-none" size={20} />
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="pl-12 pr-8 py-3 rounded-xl bg-neutral-900 border border-white/10 text-white appearance-none focus:outline-none focus:border-white/30 hover:border-white/20 transition-colors cursor-pointer min-w-[180px]"
          >
            <option value="">All Status</option>
            <option value="PENDING">Pending</option>
            <option value="PAID">Paid</option>
            <option value="PROCESSING">Processing</option>
            <option value="SHIPPED">Shipped</option>
            <option value="DELIVERED">Delivered</option>
            <option value="CANCELLED">Cancelled</option>
            <option value="REFUNDED">Refunded</option>
          </select>
        </div>
      </div>

      {/* Table */}
      <div className="bg-neutral-900 rounded-2xl border border-white/10 overflow-hidden">
        {loading ? (
          <div className="flex items-center justify-center h-64">
            <div className="w-10 h-10 border-2 border-white/20 border-t-white rounded-full animate-spin" />
          </div>
        ) : filteredOrders.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-64">
            <Package className="text-white/20 mb-4" size={48} />
            <p className="text-white/40">No orders found</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-white/10">
                  <th className="text-left text-white/60 text-sm font-medium px-6 py-4">Order</th>
                  <th className="text-left text-white/60 text-sm font-medium px-6 py-4">Customer</th>
                  <th className="text-left text-white/60 text-sm font-medium px-6 py-4">Amount</th>
                  <th className="text-left text-white/60 text-sm font-medium px-6 py-4">Status</th>
                  <th className="text-left text-white/60 text-sm font-medium px-6 py-4">Date</th>
                  <th className="text-right text-white/60 text-sm font-medium px-6 py-4">Actions</th>
                </tr>
              </thead>
              <tbody>
                {filteredOrders.map((order) => (
                  <tr key={order.id} className="border-b border-white/5 hover:bg-white/5 transition-colors">
                    <td className="px-6 py-4">
                      <p className="text-white font-medium">{order.order_code}</p>
                    </td>
                    <td className="px-6 py-4">
                      <p className="text-white">{order.customer_name}</p>
                      <p className="text-white/40 text-sm">{order.customer_email}</p>
                    </td>
                    <td className="px-6 py-4">
                      <p className="text-white font-medium">{formatCurrency(order.total_amount)}</p>
                    </td>
                    <td className="px-6 py-4">
                      <span
                        className={`inline-flex items-center gap-1.5 px-3 py-1 rounded-lg text-sm font-medium ${
                          statusColors[order.status] || "bg-gray-500/20 text-gray-400"
                        }`}
                      >
                        {statusIcons[order.status]}
                        {order.status}
                      </span>
                    </td>
                    <td className="px-6 py-4">
                      <p className="text-white/60 text-sm">{formatDate(order.created_at)}</p>
                    </td>
                    <td className="px-6 py-4 text-right">
                      <Link
                        href={`/admin/orders/${order.order_code}`}
                        className="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-white/10 text-white hover:bg-white/20 transition-colors"
                      >
                        <Eye size={16} />
                        View
                      </Link>
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
              Showing {(page - 1) * pageSize + 1} to {Math.min(page * pageSize, total)} of {total} orders
            </p>
            <div className="flex items-center gap-2">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
                className="p-2 rounded-lg bg-white/10 text-white disabled:opacity-50 disabled:cursor-not-allowed hover:bg-white/20 transition-colors"
              >
                <ChevronLeft size={20} />
              </button>
              <span className="text-white px-4">Page {page}</span>
              <button
                onClick={() => setPage((p) => p + 1)}
                disabled={page * pageSize >= total}
                className="p-2 rounded-lg bg-white/10 text-white disabled:opacity-50 disabled:cursor-not-allowed hover:bg-white/20 transition-colors"
              >
                <ChevronRight size={20} />
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
