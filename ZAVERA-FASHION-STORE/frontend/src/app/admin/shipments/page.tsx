"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import {
  Truck,
  Search,
  RefreshCcw,
  AlertTriangle,
  Clock,
  PackageX,
  Eye,
  Play,
  CheckCircle,
  XCircle,
  Package,
  MapPin,
} from "lucide-react";
import { getFulfillmentDashboard, investigateShipment, markShipmentLost, runMonitors, getShipmentsList } from "@/lib/adminApi";

interface ShipmentSummary {
  shipment_id: number;
  order_id: number;
  order_code: string;
  tracking_number: string;
  status: string;
  provider_code: string;
  days_without_update: number;
  pickup_attempts?: number;
  requires_admin?: boolean;
  alert_level?: string;
}

const statusColors: Record<string, string> = {
  PENDING: "bg-gray-500/20 text-gray-400",
  PROCESSING: "bg-blue-500/20 text-blue-400",
  PICKUP_SCHEDULED: "bg-cyan-500/20 text-cyan-400",
  PICKUP_FAILED: "bg-red-500/20 text-red-400",
  SHIPPED: "bg-purple-500/20 text-purple-400",
  IN_TRANSIT: "bg-indigo-500/20 text-indigo-400",
  OUT_FOR_DELIVERY: "bg-amber-500/20 text-amber-400",
  DELIVERED: "bg-emerald-500/20 text-emerald-400",
  DELIVERY_FAILED: "bg-red-500/20 text-red-400",
  HELD_AT_WAREHOUSE: "bg-orange-500/20 text-orange-400",
  RETURNED_TO_SENDER: "bg-pink-500/20 text-pink-400",
  LOST: "bg-red-500/20 text-red-400",
  INVESTIGATION: "bg-yellow-500/20 text-yellow-400",
  REPLACED: "bg-teal-500/20 text-teal-400",
  CANCELLED: "bg-gray-500/20 text-gray-400",
};

export default function ShipmentsPage() {
  const [loading, setLoading] = useState(true);
  const [dashboard, setDashboard] = useState<any>(null);
  const [shipments, setShipments] = useState<any[]>([]);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [statusFilter, setStatusFilter] = useState("");
  const [activeTab, setActiveTab] = useState<"stuck" | "pickup" | "alerts">("stuck");
  const [runningMonitors, setRunningMonitors] = useState(false);

  useEffect(() => {
    loadData();
  }, [page, statusFilter]);

  const loadData = async () => {
    setLoading(true);
    try {
      const [dashboardData, shipmentsData] = await Promise.all([
        getFulfillmentDashboard(),
        getShipmentsList(statusFilter, page),
      ]);
      
      setDashboard(dashboardData);
      setShipments(shipmentsData.shipments || []);
      setTotal(shipmentsData.total || 0);
    } catch (error) {
      console.error("Failed to load shipments:", error);
    } finally {
      setLoading(false);
    }
  };

  const handleRunMonitors = async () => {
    setRunningMonitors(true);
    try {
      await runMonitors();
      await loadData();
    } catch (error) {
      console.error("Failed to run monitors:", error);
    } finally {
      setRunningMonitors(false);
    }
  };

  const d = dashboard || {};
  const statusCounts = d.status_counts || {};

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
          <h1 className="text-2xl font-bold text-white">Shipments</h1>
          <p className="text-white/60 mt-1">Monitor and manage all shipments</p>
        </div>
        <div className="flex items-center gap-3">
          <button
            onClick={handleRunMonitors}
            disabled={runningMonitors}
            className="flex items-center gap-2 px-4 py-2 rounded-xl bg-purple-500/20 text-purple-400 hover:bg-purple-500/30 transition-colors disabled:opacity-50"
          >
            <Play size={18} className={runningMonitors ? "animate-pulse" : ""} />
            {runningMonitors ? "Running..." : "Run Monitors"}
          </button>
          <button
            onClick={loadData}
            className="flex items-center gap-2 px-4 py-2 rounded-xl bg-white/10 text-white hover:bg-white/20 transition-colors"
          >
            <RefreshCcw size={18} />
            Refresh
          </button>
        </div>
      </div>

      {/* Status Overview */}
      <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-3">
        {Object.entries(statusCounts).map(([status, count]) => (
          <button
            key={status}
            onClick={() => {
              setStatusFilter(status);
              setPage(1);
            }}
            className={`p-4 rounded-xl border transition-all ${
              statusFilter === status 
                ? "border-white/30 bg-white/10" 
                : "border-white/10 hover:border-white/20"
            } ${statusColors[status] || "bg-gray-500/20"}`}
          >
            <p className="text-2xl font-bold">{count as number}</p>
            <p className="text-sm opacity-80">{status.replace(/_/g, " ")}</p>
          </button>
        ))}
      </div>

      {/* Shipments List - MOVED TO TOP */}
      <div className="bg-neutral-900 rounded-2xl border border-white/10 overflow-hidden">
        <div className="p-6 border-b border-white/10">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-lg font-semibold text-white">All Shipments</h2>
              <p className="text-white/60 text-sm mt-1">
                {total} total shipments {statusFilter && `• Filtered by ${statusFilter.replace(/_/g, " ")}`}
              </p>
            </div>
            <div className="flex items-center gap-3">
              <select
                value={statusFilter}
                onChange={(e) => {
                  setStatusFilter(e.target.value);
                  setPage(1);
                }}
                className="px-4 py-2 rounded-xl bg-neutral-800 text-white border border-white/10 focus:outline-none focus:border-white/30 hover:border-white/20 transition-colors cursor-pointer"
              >
                <option value="">All Status</option>
                <option value="PENDING">Pending</option>
                <option value="PROCESSING">Processing</option>
                <option value="PICKUP_SCHEDULED">Pickup Scheduled</option>
                <option value="SHIPPED">Shipped</option>
                <option value="IN_TRANSIT">In Transit</option>
                <option value="OUT_FOR_DELIVERY">Out for Delivery</option>
                <option value="DELIVERED">Delivered</option>
                <option value="DELIVERY_FAILED">Delivery Failed</option>
              </select>
            </div>
          </div>
        </div>

        {shipments.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-16">
            <Package className="text-white/20 mb-4" size={64} />
            <p className="text-white/60 text-lg mb-2">No shipments found</p>
            <p className="text-white/40 text-sm">
              {statusFilter ? "Try changing the filter" : "Shipments will appear here once orders are shipped"}
            </p>
          </div>
        ) : (
          <>
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-white/10 bg-white/5">
                    <th className="text-left text-white/60 text-sm font-medium px-6 py-4">Order</th>
                    <th className="text-left text-white/60 text-sm font-medium px-6 py-4">Tracking</th>
                    <th className="text-left text-white/60 text-sm font-medium px-6 py-4">Courier</th>
                    <th className="text-left text-white/60 text-sm font-medium px-6 py-4">Status</th>
                    <th className="text-left text-white/60 text-sm font-medium px-6 py-4">Days</th>
                    <th className="text-left text-white/60 text-sm font-medium px-6 py-4">Created</th>
                    <th className="text-right text-white/60 text-sm font-medium px-6 py-4">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {shipments.map((shipment: any) => (
                    <tr key={shipment.id} className="border-b border-white/5 hover:bg-white/5 transition-colors">
                      <td className="px-6 py-4">
                        <Link href={`/admin/orders/${shipment.order_code}`} className="text-blue-400 hover:underline font-mono text-sm">
                          {shipment.order_code}
                        </Link>
                      </td>
                      <td className="px-6 py-4">
                        <span className="text-white/80 font-mono text-sm">
                          {shipment.tracking_number || "-"}
                        </span>
                      </td>
                      <td className="px-6 py-4">
                        <div>
                          <p className="text-white font-medium">{shipment.provider_code?.toUpperCase()}</p>
                          <p className="text-white/40 text-xs">{shipment.provider_name}</p>
                        </div>
                      </td>
                      <td className="px-6 py-4">
                        <span
                          className={`inline-flex items-center gap-1.5 px-3 py-1 rounded-lg text-sm font-medium ${
                            statusColors[shipment.status] || "bg-gray-500/20 text-gray-400"
                          }`}
                        >
                          {shipment.status.replace(/_/g, " ")}
                        </span>
                      </td>
                      <td className="px-6 py-4">
                        <span className={`text-sm font-medium ${
                          shipment.days_without_update >= 7 ? "text-red-400" :
                          shipment.days_without_update >= 3 ? "text-amber-400" : "text-white/60"
                        }`}>
                          {shipment.days_without_update || 0} days
                        </span>
                      </td>
                      <td className="px-6 py-4">
                        <span className="text-white/60 text-sm">
                          {new Date(shipment.created_at).toLocaleDateString("id-ID", {
                            day: "numeric",
                            month: "short",
                            year: "numeric",
                          })}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-right">
                        <Link
                          href={`/admin/orders/${shipment.order_code}`}
                          className="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-blue-500/20 text-blue-400 hover:bg-blue-500/30 transition-colors text-sm font-medium"
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

            {/* Pagination */}
            {total > 50 && (
              <div className="flex items-center justify-between px-6 py-4 border-t border-white/10 bg-white/5">
                <p className="text-white/60 text-sm">
                  Showing {(page - 1) * 50 + 1} to {Math.min(page * 50, total)} of {total} shipments
                </p>
                <div className="flex items-center gap-2">
                  <button
                    onClick={() => setPage((p) => Math.max(1, p - 1))}
                    disabled={page === 1}
                    className="px-4 py-2 rounded-lg bg-white/10 text-white disabled:opacity-50 disabled:cursor-not-allowed hover:bg-white/20 transition-colors text-sm font-medium"
                  >
                    Previous
                  </button>
                  <span className="text-white px-4 text-sm">Page {page} of {Math.ceil(total / 50)}</span>
                  <button
                    onClick={() => setPage((p) => p + 1)}
                    disabled={page * 50 >= total}
                    className="px-4 py-2 rounded-lg bg-white/10 text-white disabled:opacity-50 disabled:cursor-not-allowed hover:bg-white/20 transition-colors text-sm font-medium"
                  >
                    Next
                  </button>
                </div>
              </div>
            )}
          </>
        )}
      </div>

      {/* Problem Summary */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="bg-neutral-900 rounded-2xl border border-amber-500/20 p-6">
          <div className="flex items-center gap-3 mb-2">
            <Clock className="text-amber-400" size={24} />
            <span className="text-white/60">Stuck Shipments</span>
          </div>
          <p className="text-3xl font-bold text-white">{d.stuck_shipments || 0}</p>
          <p className="text-white/40 text-sm mt-1">No update 7+ days</p>
        </div>

        <div className="bg-neutral-900 rounded-2xl border border-red-500/20 p-6">
          <div className="flex items-center gap-3 mb-2">
            <PackageX className="text-red-400" size={24} />
            <span className="text-white/60">Lost Packages</span>
          </div>
          <p className="text-3xl font-bold text-white">{d.lost_packages || 0}</p>
          <p className="text-white/40 text-sm mt-1">Requires resolution</p>
        </div>

        <div className="bg-neutral-900 rounded-2xl border border-orange-500/20 p-6">
          <div className="flex items-center gap-3 mb-2">
            <Truck className="text-orange-400" size={24} />
            <span className="text-white/60">Pickup Failures</span>
          </div>
          <p className="text-3xl font-bold text-white">{d.pickup_failures || 0}</p>
          <p className="text-white/40 text-sm mt-1">Courier issues</p>
        </div>

        <div className="bg-neutral-900 rounded-2xl border border-yellow-500/20 p-6">
          <div className="flex items-center gap-3 mb-2">
            <Search className="text-yellow-400" size={24} />
            <span className="text-white/60">Under Investigation</span>
          </div>
          <p className="text-3xl font-bold text-white">{d.under_investigation || 0}</p>
          <p className="text-white/40 text-sm mt-1">Being reviewed</p>
        </div>
      </div>

      {/* Alerts Section */}
      <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-lg font-semibold text-white">Recent Alerts</h2>
          <div className="flex items-center gap-2">
            <span className="text-white/40 text-sm">{d.unresolved_alerts || 0} unresolved</span>
            {d.critical_alerts > 0 && (
              <span className="px-2 py-1 rounded-lg bg-red-500/20 text-red-400 text-sm">
                {d.critical_alerts} critical
              </span>
            )}
          </div>
        </div>

        {d.recent_alerts?.length > 0 ? (
          <div className="space-y-3">
            {d.recent_alerts.map((alert: any) => (
              <div
                key={alert.id}
                className={`p-4 rounded-xl border ${
                  alert.alert_level === "urgent"
                    ? "border-red-500/30 bg-red-500/10"
                    : alert.alert_level === "critical"
                    ? "border-amber-500/30 bg-amber-500/10"
                    : "border-blue-500/30 bg-blue-500/10"
                }`}
              >
                <div className="flex items-start justify-between">
                  <div className="flex items-start gap-3">
                    <AlertTriangle
                      className={
                        alert.alert_level === "urgent"
                          ? "text-red-400"
                          : alert.alert_level === "critical"
                          ? "text-amber-400"
                          : "text-blue-400"
                      }
                      size={20}
                    />
                    <div>
                      <p className="text-white font-medium">{alert.title}</p>
                      <p className="text-white/60 text-sm mt-0.5">{alert.description}</p>
                      <p className="text-white/40 text-xs mt-2">Shipment #{alert.shipment_id}</p>
                    </div>
                  </div>
                  <Link
                    href={`/admin/shipments/${alert.shipment_id}`}
                    className="px-3 py-1.5 rounded-lg bg-white/10 text-white text-sm hover:bg-white/20 transition-colors"
                  >
                    View
                  </Link>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="text-center py-12">
            <CheckCircle className="mx-auto text-emerald-400 mb-4" size={48} />
            <p className="text-white/60">No alerts at this time</p>
          </div>
        )}
      </div>

      {/* Performance Metrics */}
      <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
        <h2 className="text-lg font-semibold text-white mb-6">Performance</h2>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div>
            <p className="text-white/60 text-sm mb-1">Avg Delivery Time</p>
            <p className="text-2xl font-bold text-white">{d.avg_delivery_days?.toFixed(1) || "N/A"} days</p>
          </div>
          <div>
            <p className="text-white/60 text-sm mb-1">Delivery Failures</p>
            <p className="text-2xl font-bold text-white">{d.delivery_failures || 0}</p>
          </div>
          <div>
            <p className="text-white/60 text-sm mb-1">Pending Resolution</p>
            <p className="text-2xl font-bold text-white">{d.pending_resolution || 0}</p>
          </div>
        </div>
      </div>
    </div>
  );
}
