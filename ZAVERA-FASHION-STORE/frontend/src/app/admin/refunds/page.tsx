"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { RefreshCcw, Search, Filter, CheckCircle, Clock, XCircle, AlertTriangle } from "lucide-react";
import api from "@/lib/api";

interface Refund {
  id: number;
  refund_code: string;
  order_code: string;
  order_id: number;
  refund_type: string;
  reason: string;
  reason_detail: string;
  original_amount: number;
  refund_amount: number;
  shipping_refund: number;
  items_refund: number;
  status: string;
  gateway_refund_id?: string;
  gateway_status?: string;
  requested_at: string;
  processed_at?: string;
  completed_at?: string;
  created_at: string;
}

export default function RefundsPage() {
  const router = useRouter();
  const [refunds, setRefunds] = useState<Refund[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState("ALL");

  useEffect(() => {
    loadRefunds();
  }, []);

  const loadRefunds = async () => {
    setLoading(true);
    try {
      const token = localStorage.getItem("auth_token");
      
      // Call the backend list refunds endpoint directly
      const response = await api.get('/admin/refunds', {
        headers: { Authorization: `Bearer ${token}` },
        params: {
          page: 1,
          page_size: 100, // Get all refunds for now
        }
      });
      
      setRefunds(response.data.refunds || []);
    } catch (error) {
      console.error("Failed to load refunds:", error);
      setRefunds([]);
    } finally {
      setLoading(false);
    }
  };

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

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'COMPLETED':
        return <CheckCircle className="text-emerald-400" size={20} />;
      case 'PROCESSING':
        return <Clock className="text-blue-400" size={20} />;
      case 'PENDING':
        return <AlertTriangle className="text-amber-400" size={20} />;
      case 'FAILED':
        return <XCircle className="text-red-400" size={20} />;
      default:
        return <Clock className="text-white/40" size={20} />;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'COMPLETED':
        return 'bg-emerald-500/20 text-emerald-400 border-emerald-500/30';
      case 'PROCESSING':
        return 'bg-blue-500/20 text-blue-400 border-blue-500/30';
      case 'PENDING':
        return 'bg-amber-500/20 text-amber-400 border-amber-500/30';
      case 'FAILED':
        return 'bg-red-500/20 text-red-400 border-red-500/30';
      default:
        return 'bg-white/10 text-white/60 border-white/20';
    }
  };

  const filteredRefunds = refunds.filter(refund => {
    const matchesSearch = 
      refund.refund_code.toLowerCase().includes(searchQuery.toLowerCase()) ||
      refund.order_code.toLowerCase().includes(searchQuery.toLowerCase());
    
    const matchesStatus = statusFilter === "ALL" || refund.status === statusFilter;
    
    return matchesSearch && matchesStatus;
  });

  const stats = {
    pending: refunds.filter(r => r.status === 'PENDING').length,
    processing: refunds.filter(r => r.status === 'PROCESSING').length,
    completed: refunds.filter(r => r.status === 'COMPLETED').length,
    failed: refunds.filter(r => r.status === 'FAILED').length,
    totalAmount: refunds
      .filter(r => r.status === 'COMPLETED')
      .reduce((sum, r) => sum + r.refund_amount, 0),
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-white">Refunds</h1>
        <p className="text-white/60 mt-1">Manage refund requests and processing</p>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-5 gap-4">
        <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
          <div className="flex items-center gap-3 mb-2">
            <Clock className="text-amber-400" size={20} />
            <span className="text-white/60 text-sm">Pending</span>
          </div>
          <p className="text-3xl font-bold text-white">{stats.pending}</p>
        </div>

        <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
          <div className="flex items-center gap-3 mb-2">
            <RefreshCcw className="text-blue-400" size={20} />
            <span className="text-white/60 text-sm">Processing</span>
          </div>
          <p className="text-3xl font-bold text-white">{stats.processing}</p>
        </div>

        <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
          <div className="flex items-center gap-3 mb-2">
            <CheckCircle className="text-emerald-400" size={20} />
            <span className="text-white/60 text-sm">Completed</span>
          </div>
          <p className="text-3xl font-bold text-white">{stats.completed}</p>
        </div>

        <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
          <div className="flex items-center gap-3 mb-2">
            <XCircle className="text-red-400" size={20} />
            <span className="text-white/60 text-sm">Failed</span>
          </div>
          <p className="text-3xl font-bold text-white">{stats.failed}</p>
        </div>

        <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
          <div className="flex items-center gap-3 mb-2">
            <span className="text-white/60 text-sm">Total Refunded</span>
          </div>
          <p className="text-2xl font-bold text-emerald-400">{formatCurrency(stats.totalAmount)}</p>
        </div>
      </div>

      {/* Filters */}
      <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
        <div className="flex flex-col md:flex-row gap-4">
          {/* Search */}
          <div className="flex-1 relative">
            <Search className="absolute left-4 top-1/2 -translate-y-1/2 text-white/40" size={20} />
            <input
              type="text"
              placeholder="Search by refund code or order code..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-12 pr-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white placeholder-white/40 focus:outline-none focus:border-white/30"
            />
          </div>

          {/* Status Filter */}
          <div className="flex items-center gap-2">
            <Filter className="text-white/40" size={20} />
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              className="px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white focus:outline-none focus:border-white/30"
            >
              <option value="ALL">All Status</option>
              <option value="PENDING">Pending</option>
              <option value="PROCESSING">Processing</option>
              <option value="COMPLETED">Completed</option>
              <option value="FAILED">Failed</option>
            </select>
          </div>

          {/* Refresh Button */}
          <button
            onClick={loadRefunds}
            className="px-4 py-3 rounded-xl bg-white/10 text-white hover:bg-white/20 transition-colors flex items-center gap-2"
          >
            <RefreshCcw size={20} />
            Refresh
          </button>
        </div>
      </div>

      {/* Refunds List */}
      <div className="bg-neutral-900 rounded-2xl border border-white/10 overflow-hidden">
        {loading ? (
          <div className="flex items-center justify-center py-12">
            <div className="w-10 h-10 border-2 border-white/20 border-t-white rounded-full animate-spin" />
          </div>
        ) : filteredRefunds.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12">
            <RefreshCcw className="text-white/20 mb-4" size={48} />
            <p className="text-white/60 text-lg">No refunds found</p>
            <p className="text-white/40 text-sm mt-2">
              Refunds are created when you process a refund from an order detail page or when a dispute is resolved with refund.
            </p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-white/5 border-b border-white/10">
                <tr>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-white/80">Refund Code</th>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-white/80">Order Code</th>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-white/80">Type</th>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-white/80">Amount</th>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-white/80">Status</th>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-white/80">Gateway ID</th>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-white/80">Date</th>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-white/80">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-white/10">
                {filteredRefunds.map((refund) => (
                  <tr key={refund.id} className="hover:bg-white/5 transition-colors">
                    <td className="px-6 py-4">
                      <span className="text-white font-medium">{refund.refund_code}</span>
                    </td>
                    <td className="px-6 py-4">
                      <button
                        onClick={() => router.push(`/admin/orders/${refund.order_code}`)}
                        className="text-blue-400 hover:text-blue-300 font-medium"
                      >
                        {refund.order_code}
                      </button>
                    </td>
                    <td className="px-6 py-4">
                      <span className="px-2 py-1 rounded-full text-xs font-medium bg-purple-500/20 text-purple-400">
                        {refund.refund_type}
                      </span>
                    </td>
                    <td className="px-6 py-4">
                      <span className="text-white font-semibold">{formatCurrency(refund.refund_amount)}</span>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-2">
                        {getStatusIcon(refund.status)}
                        <span className={`px-2 py-1 rounded-full text-xs font-medium border ${getStatusColor(refund.status)}`}>
                          {refund.status}
                        </span>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      {refund.gateway_refund_id ? (
                        <span className={`text-xs font-mono ${
                          refund.gateway_refund_id === 'MANUAL_REFUND' 
                            ? 'text-amber-400' 
                            : 'text-white/60'
                        }`}>
                          {refund.gateway_refund_id}
                        </span>
                      ) : (
                        <span className="text-white/40 text-xs">-</span>
                      )}
                    </td>
                    <td className="px-6 py-4">
                      <span className="text-white/60 text-sm">{formatDate(refund.created_at)}</span>
                    </td>
                    <td className="px-6 py-4">
                      <button
                        onClick={() => router.push(`/admin/orders/${refund.order_code}`)}
                        className="px-3 py-1.5 rounded-lg bg-white/10 text-white hover:bg-white/20 transition-colors text-sm"
                      >
                        View Order
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Info Box */}
      <div className="bg-blue-500/10 border border-blue-500/30 rounded-2xl p-6">
        <div className="flex items-start gap-3">
          <AlertTriangle className="text-blue-400 flex-shrink-0 mt-1" size={20} />
          <div>
            <h3 className="text-blue-400 font-semibold mb-2">How Refunds Work</h3>
            <ul className="text-white/60 text-sm space-y-1">
              <li>• Refunds are created from order detail page or dispute resolution</li>
              <li>• PENDING refunds need to be processed manually</li>
              <li>• Processing sends refund request to Midtrans gateway</li>
              <li>• COMPLETED refunds have been successfully processed</li>
              <li>• MANUAL_REFUND indicates orders without payment gateway</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  );
}
