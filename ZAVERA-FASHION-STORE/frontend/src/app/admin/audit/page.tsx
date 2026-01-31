"use client";

import { useEffect, useState } from "react";
import {
  FileText,
  Search,
  RefreshCcw,
  Filter,
  ChevronLeft,
  ChevronRight,
  CheckCircle,
  XCircle,
  User,
  Clock,
  Shield,
} from "lucide-react";
import { getAuditLogs, AuditLog } from "@/lib/adminApi";

const actionColors: Record<string, string> = {
  FORCE_CANCEL: "bg-red-500/20 text-red-400",
  FORCE_REFUND: "bg-amber-500/20 text-amber-400",
  FORCE_RESHIP: "bg-blue-500/20 text-blue-400",
  RECONCILE_PAYMENT: "bg-purple-500/20 text-purple-400",
  CREATE_REFUND: "bg-emerald-500/20 text-emerald-400",
  PROCESS_REFUND: "bg-teal-500/20 text-teal-400",
  SYNC_PAYMENT: "bg-cyan-500/20 text-cyan-400",
  RESOLVE_DISPUTE: "bg-indigo-500/20 text-indigo-400",
};

export default function AuditPage() {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [search, setSearch] = useState("");
  const [actionFilter, setActionFilter] = useState("");
  const pageSize = 50;

  useEffect(() => {
    loadLogs();
  }, [page]);

  const loadLogs = async () => {
    setLoading(true);
    try {
      const data = await getAuditLogs(page, pageSize);
      setLogs(data.logs);
      setTotal(data.total);
    } catch (error) {
      console.error("Failed to load audit logs:", error);
    } finally {
      setLoading(false);
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("id-ID", {
      day: "numeric",
      month: "short",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
    });
  };

  const filteredLogs = logs.filter((log) => {
    const matchesSearch =
      search === "" ||
      log.admin_email.toLowerCase().includes(search.toLowerCase()) ||
      log.action_detail.toLowerCase().includes(search.toLowerCase()) ||
      log.target_type.toLowerCase().includes(search.toLowerCase());

    const matchesAction = actionFilter === "" || log.action_type === actionFilter;

    return matchesSearch && matchesAction;
  });

  // Get unique action types for filter
  const actionTypes = Array.from(new Set(logs.map((l) => l.action_type)));

  if (loading && logs.length === 0) {
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
          <h1 className="text-2xl font-bold text-white">Audit Log</h1>
          <p className="text-white/60 mt-1">Immutable record of all admin actions</p>
        </div>
        <button
          onClick={loadLogs}
          className="flex items-center gap-2 px-4 py-2 rounded-xl bg-white/10 text-white hover:bg-white/20 transition-colors"
        >
          <RefreshCcw size={18} />
          Refresh
        </button>
      </div>

      {/* Info Banner */}
      <div className="bg-neutral-900 rounded-2xl border border-purple-500/20 p-6">
        <div className="flex items-start gap-4">
          <div className="p-3 rounded-xl bg-purple-500/20">
            <Shield className="text-purple-400" size={24} />
          </div>
          <div>
            <h3 className="text-white font-semibold mb-1">Forensic Audit Trail</h3>
            <p className="text-white/60 text-sm">
              This log is immutable and cannot be modified or deleted. Every admin action is recorded with full context
              including who performed it, when, and what changed.
            </p>
          </div>
        </div>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-4 top-1/2 -translate-y-1/2 text-white/40" size={20} />
          <input
            type="text"
            placeholder="Search by admin email, action, or target..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-12 pr-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white placeholder-white/40 focus:outline-none focus:border-white/30"
          />
        </div>

        <div className="relative">
          <Filter className="absolute left-4 top-1/2 -translate-y-1/2 text-white/40 pointer-events-none" size={20} />
          <select
            value={actionFilter}
            onChange={(e) => setActionFilter(e.target.value)}
            className="pl-12 pr-8 py-3 rounded-xl bg-neutral-900 border border-white/10 text-white appearance-none focus:outline-none focus:border-white/30 hover:border-white/20 transition-colors cursor-pointer min-w-[200px]"
          >
            <option value="">All Actions</option>
            {actionTypes.map((type) => (
              <option key={type} value={type}>
                {type.replace(/_/g, " ")}
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Logs List */}
      <div className="bg-neutral-900 rounded-2xl border border-white/10 overflow-hidden">
        {filteredLogs.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-64">
            <FileText className="text-white/20 mb-4" size={48} />
            <p className="text-white/40">No audit logs found</p>
          </div>
        ) : (
          <div className="divide-y divide-white/5">
            {filteredLogs.map((log) => (
              <div key={log.id} className="p-6 hover:bg-white/5 transition-colors">
                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-3 mb-2">
                      <span
                        className={`px-3 py-1 rounded-lg text-sm font-medium ${
                          actionColors[log.action_type] || "bg-gray-500/20 text-gray-400"
                        }`}
                      >
                        {log.action_type.replace(/_/g, " ")}
                      </span>
                      {log.success ? (
                        <span className="flex items-center gap-1 text-emerald-400 text-sm">
                          <CheckCircle size={14} />
                          Success
                        </span>
                      ) : (
                        <span className="flex items-center gap-1 text-red-400 text-sm">
                          <XCircle size={14} />
                          Failed
                        </span>
                      )}
                    </div>

                    <p className="text-white mb-2">{log.action_detail}</p>

                    <div className="flex flex-wrap items-center gap-4 text-sm text-white/40">
                      <span className="flex items-center gap-1">
                        <User size={14} />
                        {log.admin_email}
                      </span>
                      <span className="flex items-center gap-1">
                        <FileText size={14} />
                        {log.target_type} #{log.target_id}
                      </span>
                      <span className="flex items-center gap-1">
                        <Clock size={14} />
                        {formatDate(log.created_at)}
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Pagination */}
        {total > pageSize && (
          <div className="flex items-center justify-between px-6 py-4 border-t border-white/10">
            <p className="text-white/60 text-sm">
              Showing {(page - 1) * pageSize + 1} to {Math.min(page * pageSize, total)} of {total} logs
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
