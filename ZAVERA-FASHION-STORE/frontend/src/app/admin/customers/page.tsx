"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { Search, Users, TrendingUp, ShoppingBag, DollarSign, Filter, Download, UserCheck, UserX, Clock } from "lucide-react";
import api from "@/lib/api";

interface Customer {
  id: number;
  email: string;
  first_name: string;
  last_name: string;
  phone: string;
  total_orders: number;
  total_spent: number;
  last_order_date: string;
  created_at: string;
  is_verified: boolean;
  segment: "VIP" | "Regular" | "New";
}

interface CustomerStats {
  total_customers: number;
  new_this_month: number;
  vip_customers: number;
  total_lifetime_value: number;
}

export default function CustomersPage() {
  const [customers, setCustomers] = useState<Customer[]>([]);
  const [stats, setStats] = useState<CustomerStats>({
    total_customers: 0,
    new_this_month: 0,
    vip_customers: 0,
    total_lifetime_value: 0,
  });
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState("");
  const [filterSegment, setFilterSegment] = useState<string>("all");
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);

  useEffect(() => {
    loadCustomers();
  }, [page, filterSegment, searchQuery]);

  const loadCustomers = async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams({
        page: page.toString(),
        limit: "20",
        ...(searchQuery && { search: searchQuery }),
        ...(filterSegment !== "all" && { segment: filterSegment }),
      });

      const res = await api.get(`/admin/customers?${params}`);
      setCustomers(res.data.customers || []);
      setStats(res.data.stats || stats);
      setTotalPages(res.data.total_pages || 1);
    } catch (err) {
      console.error("Failed to load customers:", err);
    } finally {
      setLoading(false);
    }
  };

  const exportCustomers = async () => {
    try {
      const res = await api.get("/admin/customers/export", { responseType: "blob" });
      const url = window.URL.createObjectURL(new Blob([res.data]));
      const link = document.createElement("a");
      link.href = url;
      link.setAttribute("download", `customers-${new Date().toISOString().split("T")[0]}.csv`);
      document.body.appendChild(link);
      link.click();
      link.remove();
    } catch (err) {
      console.error("Failed to export:", err);
    }
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      minimumFractionDigits: 0,
    }).format(amount);
  };

  const formatDate = (date: string) => {
    return new Date(date).toLocaleDateString("id-ID", {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  };

  const getSegmentColor = (segment: string) => {
    switch (segment) {
      case "VIP":
        return "bg-purple-500/20 text-purple-400 border-purple-500/30";
      case "Regular":
        return "bg-blue-500/20 text-blue-400 border-blue-500/30";
      case "New":
        return "bg-emerald-500/20 text-emerald-400 border-emerald-500/30";
      default:
        return "bg-gray-500/20 text-gray-400 border-gray-500/30";
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-white">Customers</h1>
          <p className="text-white/60 mt-1">Manage your customer base</p>
        </div>
        <button
          onClick={exportCustomers}
          className="flex items-center gap-2 px-4 py-2 rounded-xl bg-white/10 text-white hover:bg-white/20 transition-colors"
        >
          <Download size={18} />
          Export CSV
        </button>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="bg-gradient-to-br from-blue-500/20 to-blue-500/5 border border-blue-500/20 rounded-2xl p-6">
          <div className="flex items-start justify-between">
            <div>
              <p className="text-white/60 text-sm font-medium mb-1">Total Customers</p>
              <p className="text-3xl font-bold text-white">{stats.total_customers}</p>
            </div>
            <div className="p-3 rounded-xl bg-white/10">
              <Users className="text-blue-400" size={24} />
            </div>
          </div>
        </div>

        <div className="bg-gradient-to-br from-emerald-500/20 to-emerald-500/5 border border-emerald-500/20 rounded-2xl p-6">
          <div className="flex items-start justify-between">
            <div>
              <p className="text-white/60 text-sm font-medium mb-1">New This Month</p>
              <p className="text-3xl font-bold text-white">{stats.new_this_month}</p>
            </div>
            <div className="p-3 rounded-xl bg-white/10">
              <TrendingUp className="text-emerald-400" size={24} />
            </div>
          </div>
        </div>

        <div className="bg-gradient-to-br from-purple-500/20 to-purple-500/5 border border-purple-500/20 rounded-2xl p-6">
          <div className="flex items-start justify-between">
            <div>
              <p className="text-white/60 text-sm font-medium mb-1">VIP Customers</p>
              <p className="text-3xl font-bold text-white">{stats.vip_customers}</p>
              <p className="text-white/40 text-sm mt-1">5+ orders</p>
            </div>
            <div className="p-3 rounded-xl bg-white/10">
              <UserCheck className="text-purple-400" size={24} />
            </div>
          </div>
        </div>

        <div className="bg-gradient-to-br from-amber-500/20 to-amber-500/5 border border-amber-500/20 rounded-2xl p-6">
          <div className="flex items-start justify-between">
            <div>
              <p className="text-white/60 text-sm font-medium mb-1">Lifetime Value</p>
              <p className="text-2xl font-bold text-white">{formatCurrency(stats.total_lifetime_value)}</p>
            </div>
            <div className="p-3 rounded-xl bg-white/10">
              <DollarSign className="text-amber-400" size={24} />
            </div>
          </div>
        </div>
      </div>

      {/* Filters */}
      <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
        <div className="flex flex-col lg:flex-row gap-4">
          {/* Search */}
          <div className="flex-1 relative">
            <Search className="absolute left-4 top-1/2 -translate-y-1/2 text-white/40" size={20} />
            <input
              type="text"
              placeholder="Search by name, email, or phone..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-12 pr-4 py-3 bg-white/5 border border-white/10 rounded-xl text-white placeholder-white/40 focus:outline-none focus:border-white/30"
            />
          </div>

          {/* Segment Filter */}
          <div className="flex items-center gap-2">
            <Filter className="text-white/40" size={20} />
            <select
              value={filterSegment}
              onChange={(e) => setFilterSegment(e.target.value)}
              className="px-4 py-3 bg-neutral-900 border border-white/10 rounded-xl text-white focus:outline-none focus:border-white/30 hover:border-white/20 transition-colors cursor-pointer"
            >
              <option value="all">All Segments</option>
              <option value="VIP">VIP</option>
              <option value="Regular">Regular</option>
              <option value="New">New</option>
            </select>
          </div>
        </div>
      </div>

      {/* Customer List */}
      <div className="bg-neutral-900 rounded-2xl border border-white/10 overflow-hidden">
        {loading ? (
          <div className="flex items-center justify-center py-12">
            <div className="w-8 h-8 border-2 border-white/20 border-t-white rounded-full animate-spin" />
          </div>
        ) : customers.length === 0 ? (
          <div className="text-center py-12">
            <Users className="mx-auto text-white/20 mb-3" size={48} />
            <p className="text-white/40">No customers found</p>
          </div>
        ) : (
          <>
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="bg-white/5 border-b border-white/10">
                  <tr>
                    <th className="px-6 py-4 text-left text-sm font-medium text-white/60">Customer</th>
                    <th className="px-6 py-4 text-left text-sm font-medium text-white/60">Segment</th>
                    <th className="px-6 py-4 text-left text-sm font-medium text-white/60">Orders</th>
                    <th className="px-6 py-4 text-left text-sm font-medium text-white/60">Total Spent</th>
                    <th className="px-6 py-4 text-left text-sm font-medium text-white/60">Last Order</th>
                    <th className="px-6 py-4 text-left text-sm font-medium text-white/60">Joined</th>
                    <th className="px-6 py-4 text-right text-sm font-medium text-white/60">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-white/10">
                  {customers.map((customer) => (
                    <tr key={customer.id} className="hover:bg-white/5 transition-colors">
                      <td className="px-6 py-4">
                        <div>
                          <p className="font-medium text-white">
                            {customer.first_name} {customer.last_name}
                          </p>
                          <p className="text-sm text-white/60">{customer.email}</p>
                          {customer.phone && <p className="text-sm text-white/40">{customer.phone}</p>}
                        </div>
                      </td>
                      <td className="px-6 py-4">
                        <span className={`px-3 py-1 rounded-lg text-xs font-medium border ${getSegmentColor(customer.segment)}`}>
                          {customer.segment}
                        </span>
                      </td>
                      <td className="px-6 py-4">
                        <div className="flex items-center gap-2">
                          <ShoppingBag className="text-white/40" size={16} />
                          <span className="text-white font-medium">{customer.total_orders}</span>
                        </div>
                      </td>
                      <td className="px-6 py-4">
                        <span className="text-white font-medium">{formatCurrency(customer.total_spent)}</span>
                      </td>
                      <td className="px-6 py-4">
                        <span className="text-white/60 text-sm">
                          {customer.last_order_date ? formatDate(customer.last_order_date) : "-"}
                        </span>
                      </td>
                      <td className="px-6 py-4">
                        <span className="text-white/60 text-sm">{formatDate(customer.created_at)}</span>
                      </td>
                      <td className="px-6 py-4 text-right">
                        <Link
                          href={`/admin/customers/${customer.id}`}
                          className="text-blue-400 hover:text-blue-300 text-sm font-medium"
                        >
                          View Details
                        </Link>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {/* Pagination */}
            {totalPages > 1 && (
              <div className="px-6 py-4 border-t border-white/10 flex items-center justify-between">
                <button
                  onClick={() => setPage(p => Math.max(1, p - 1))}
                  disabled={page === 1}
                  className="px-4 py-2 rounded-lg bg-white/5 text-white disabled:opacity-50 disabled:cursor-not-allowed hover:bg-white/10 transition-colors"
                >
                  Previous
                </button>
                <span className="text-white/60">
                  Page {page} of {totalPages}
                </span>
                <button
                  onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                  disabled={page === totalPages}
                  className="px-4 py-2 rounded-lg bg-white/5 text-white disabled:opacity-50 disabled:cursor-not-allowed hover:bg-white/10 transition-colors"
                >
                  Next
                </button>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}
