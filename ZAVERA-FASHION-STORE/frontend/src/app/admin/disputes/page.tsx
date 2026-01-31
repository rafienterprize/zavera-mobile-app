"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import {
  MessageSquare,
  Search,
  RefreshCcw,
  AlertTriangle,
  Clock,
  CheckCircle,
  XCircle,
  Eye,
  Filter,
  ChevronRight,
} from "lucide-react";
import { getOpenDisputes, Dispute } from "@/lib/adminApi";

const statusColors: Record<string, string> = {
  OPEN: "bg-amber-500/20 text-amber-400",
  INVESTIGATING: "bg-purple-500/20 text-purple-400",
  EVIDENCE_REQUIRED: "bg-blue-500/20 text-blue-400",
  PENDING_RESOLUTION: "bg-cyan-500/20 text-cyan-400",
  RESOLVED_REFUND: "bg-emerald-500/20 text-emerald-400",
  RESOLVED_RESHIP: "bg-teal-500/20 text-teal-400",
  RESOLVED_REJECTED: "bg-red-500/20 text-red-400",
  CLOSED: "bg-gray-500/20 text-gray-400",
};

const typeLabels: Record<string, string> = {
  LOST_PACKAGE: "Lost Package",
  DAMAGED_PACKAGE: "Damaged Package",
  WRONG_ITEM: "Wrong Item",
  MISSING_ITEM: "Missing Item",
  NOT_DELIVERED: "Not Delivered",
  LATE_DELIVERY: "Late Delivery",
  FAKE_DELIVERY: "Fake Delivery",
  OTHER: "Other",
};

export default function DisputesPage() {
  const [disputes, setDisputes] = useState<Dispute[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState("");

  useEffect(() => {
    loadDisputes();
  }, []);

  const loadDisputes = async () => {
    setLoading(true);
    try {
      const data = await getOpenDisputes();
      setDisputes(data);
    } catch (error) {
      console.error("Failed to load disputes:", error);
    } finally {
      setLoading(false);
    }
  };

  const filteredDisputes = disputes.filter((dispute) => {
    const matchesSearch =
      search === "" ||
      dispute.dispute_code.toLowerCase().includes(search.toLowerCase()) ||
      dispute.title.toLowerCase().includes(search.toLowerCase()) ||
      dispute.customer_email.toLowerCase().includes(search.toLowerCase());

    const matchesStatus = statusFilter === "" || dispute.status === statusFilter;

    return matchesSearch && matchesStatus;
  });

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("id-ID", {
      day: "numeric",
      month: "short",
      year: "numeric",
    });
  };

  // Group by status for summary
  const statusCounts = disputes.reduce((acc, d) => {
    acc[d.status] = (acc[d.status] || 0) + 1;
    return acc;
  }, {} as Record<string, number>);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-96">
        <div className="w-10 h-10 border-2 border-white/20 border-t-white rounded-full animate-spin" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-white">Disputes</h1>
          <p className="text-white/60 mt-1">Manage customer complaints and disputes</p>
        </div>
        <button
          onClick={loadDisputes}
          className="flex items-center gap-2 px-4 py-2 rounded-xl bg-white/10 text-white hover:bg-white/20 transition-colors"
        >
          <RefreshCcw size={18} />
          Refresh
        </button>
      </div>

      {/* Status Summary */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
        <div className="bg-neutral-900 rounded-xl border border-amber-500/20 p-4">
          <div className="flex items-center gap-2 mb-2">
            <Clock className="text-amber-400" size={18} />
            <span className="text-white/60 text-sm">Open</span>
          </div>
          <p className="text-2xl font-bold text-white">{statusCounts.OPEN || 0}</p>
        </div>

        <div className="bg-neutral-900 rounded-xl border border-purple-500/20 p-4">
          <div className="flex items-center gap-2 mb-2">
            <Search className="text-purple-400" size={18} />
            <span className="text-white/60 text-sm">Investigating</span>
          </div>
          <p className="text-2xl font-bold text-white">{statusCounts.INVESTIGATING || 0}</p>
        </div>

        <div className="bg-neutral-900 rounded-xl border border-blue-500/20 p-4">
          <div className="flex items-center gap-2 mb-2">
            <AlertTriangle className="text-blue-400" size={18} />
            <span className="text-white/60 text-sm">Evidence Required</span>
          </div>
          <p className="text-2xl font-bold text-white">{statusCounts.EVIDENCE_REQUIRED || 0}</p>
        </div>

        <div className="bg-neutral-900 rounded-xl border border-cyan-500/20 p-4">
          <div className="flex items-center gap-2 mb-2">
            <CheckCircle className="text-cyan-400" size={18} />
            <span className="text-white/60 text-sm">Pending Resolution</span>
          </div>
          <p className="text-2xl font-bold text-white">{statusCounts.PENDING_RESOLUTION || 0}</p>
        </div>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-4 top-1/2 -translate-y-1/2 text-white/40" size={20} />
          <input
            type="text"
            placeholder="Search by code, title, or email..."
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
            <option value="OPEN">Open</option>
            <option value="INVESTIGATING">Investigating</option>
            <option value="EVIDENCE_REQUIRED">Evidence Required</option>
            <option value="PENDING_RESOLUTION">Pending Resolution</option>
          </select>
        </div>
      </div>

      {/* Disputes List */}
      <div className="bg-neutral-900 rounded-2xl border border-white/10 overflow-hidden">
        {filteredDisputes.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-64">
            <MessageSquare className="text-white/20 mb-4" size={48} />
            <p className="text-white/40">No disputes found</p>
          </div>
        ) : (
          <div className="divide-y divide-white/5">
            {filteredDisputes.map((dispute) => (
              <Link
                key={dispute.id}
                href={`/admin/disputes/${dispute.id}`}
                className="block p-6 hover:bg-white/5 transition-colors"
              >
                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-3 mb-2">
                      <span className="text-white/40 text-sm font-mono">{dispute.dispute_code}</span>
                      <span
                        className={`px-2 py-0.5 rounded-lg text-xs font-medium ${
                          statusColors[dispute.status] || "bg-gray-500/20 text-gray-400"
                        }`}
                      >
                        {dispute.status.replace(/_/g, " ")}
                      </span>
                      <span className="px-2 py-0.5 rounded-lg text-xs bg-white/10 text-white/60">
                        {typeLabels[dispute.dispute_type] || dispute.dispute_type}
                      </span>
                    </div>
                    <h3 className="text-white font-medium mb-1">{dispute.title}</h3>
                    <p className="text-white/60 text-sm line-clamp-2">{dispute.description}</p>
                    <div className="flex items-center gap-4 mt-3 text-sm text-white/40">
                      <span>{dispute.customer_email}</span>
                      <span>•</span>
                      <span>{formatDate(dispute.created_at)}</span>
                    </div>
                  </div>
                  <ChevronRight className="text-white/20 flex-shrink-0" size={20} />
                </div>
              </Link>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
