"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import {
  DollarSign,
  RefreshCcw,
  ShoppingBag,
  Truck,
  Clock,
  PackageX,
  AlertTriangle,
  MessageSquare,
  FileText,
  Activity,
  ArrowUpRight,
  Package,
  TrendingUp,
  TrendingDown,
  Users,
  CreditCard,
  BarChart3,
  AlertCircle,
  CheckCircle,
} from "lucide-react";
import {
  getExecutiveMetrics,
  getPaymentMonitor,
  getInventoryAlerts,
  getCustomerInsights,
  getConversionFunnel,
  getRevenueChart,
  getFulfillmentDashboard,
  getOpenDisputes,
  getSystemHealth,
  getCourierPerformance,
  ExecutiveMetrics,
  PaymentMonitor,
  InventoryAlerts,
  CustomerInsights,
  ConversionFunnel,
  RevenueChart,
} from "@/lib/adminApi";

interface StatCardProps {
  title: string;
  value: string | number;
  subtitle?: string;
  icon: React.ReactNode;
  color: "emerald" | "amber" | "red" | "blue" | "purple" | "cyan";
  href?: string;
  trend?: { value: number; isPositive: boolean };
}

function StatCard({ title, value, subtitle, icon, color, href, trend }: StatCardProps) {
  const colorClasses = {
    emerald: "from-emerald-500/20 to-emerald-500/5 border-emerald-500/20 text-emerald-400",
    amber: "from-amber-500/20 to-amber-500/5 border-amber-500/20 text-amber-400",
    red: "from-red-500/20 to-red-500/5 border-red-500/20 text-red-400",
    blue: "from-blue-500/20 to-blue-500/5 border-blue-500/20 text-blue-400",
    purple: "from-purple-500/20 to-purple-500/5 border-purple-500/20 text-purple-400",
    cyan: "from-cyan-500/20 to-cyan-500/5 border-cyan-500/20 text-cyan-400",
  };

  const content = (
    <div
      className={`relative overflow-hidden rounded-2xl bg-gradient-to-br ${colorClasses[color]} border p-6 transition-all duration-300 hover:scale-[1.02]`}
    >
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <p className="text-white/60 text-sm font-medium mb-1">{title}</p>
          <p className="text-3xl font-bold text-white">{value}</p>
          {subtitle && <p className="text-white/40 text-sm mt-1">{subtitle}</p>}
          {trend && (
            <div className={`flex items-center gap-1 mt-2 text-sm ${trend.isPositive ? "text-emerald-400" : "text-red-400"}`}>
              {trend.isPositive ? <TrendingUp size={16} /> : <TrendingDown size={16} />}
              <span>{Math.abs(trend.value).toFixed(1)}%</span>
            </div>
          )}
        </div>
        <div className={`p-3 rounded-xl bg-white/10 ${colorClasses[color]}`}>{icon}</div>
      </div>
      {href && (
        <div className="absolute bottom-4 right-4 text-white/40 text-sm flex items-center gap-1">
          View all <ArrowUpRight size={14} />
        </div>
      )}
    </div>
  );

  if (href) {
    return (
      <Link href={href} className="group">
        {content}
      </Link>
    );
  }

  return content;
}

export default function AdminDashboard() {
  const [loading, setLoading] = useState(true);
  const [period, setPeriod] = useState("month"); // Changed from "today" to "month"
  const [executive, setExecutive] = useState<ExecutiveMetrics | null>(null);
  const [payments, setPayments] = useState<PaymentMonitor | null>(null);
  const [inventory, setInventory] = useState<InventoryAlerts | null>(null);
  const [customers, setCustomers] = useState<CustomerInsights | null>(null);
  const [funnel, setFunnel] = useState<ConversionFunnel | null>(null);
  const [chart, setChart] = useState<RevenueChart | null>(null);
  const [fulfillment, setFulfillment] = useState<any>(null);
  const [disputes, setDisputes] = useState<any[]>([]);
  const [systemHealth, setSystemHealth] = useState<any>(null);
  const [previousMetrics, setPreviousMetrics] = useState<ExecutiveMetrics | null>(null);
  const [courierPerformance, setCourierPerformance] = useState<any[]>([]);

  useEffect(() => {
    loadDashboard();
  }, [period]);

  const loadDashboard = async () => {
    setLoading(true);
    try {
      // Get previous period for comparison
      const previousPeriod = getPreviousPeriod(period);
      
      const [execData, paymentData, inventoryData, customerData, funnelData, chartData, fulfillmentData, disputeData, prevExecData, healthData, courierData] =
        await Promise.all([
          getExecutiveMetrics(period).catch(() => null),
          getPaymentMonitor().catch(() => null),
          getInventoryAlerts().catch(() => null),
          getCustomerInsights().catch(() => null),
          getConversionFunnel(period).catch(() => null),
          getRevenueChart("7days").catch(() => null),
          getFulfillmentDashboard().catch(() => null),
          getOpenDisputes().catch(() => []),
          getExecutiveMetrics(previousPeriod).catch(() => null),
          getSystemHealth().catch(() => null),
          getCourierPerformance().catch(() => []),
        ]);

      setExecutive(execData);
      setPayments(paymentData);
      setInventory(inventoryData);
      setCustomers(customerData);
      setFunnel(funnelData);
      setChart(chartData);
      setFulfillment(fulfillmentData);
      setDisputes(disputeData);
      setPreviousMetrics(prevExecData);
      setSystemHealth(healthData);
      setCourierPerformance(courierData);
    } catch (error) {
      console.error("Failed to load dashboard:", error);
    } finally {
      setLoading(false);
    }
  };

  const getPreviousPeriod = (current: string) => {
    // Map current period to previous period for comparison
    const periodMap: Record<string, string> = {
      today: "yesterday",
      week: "last_week",
      month: "last_month",
      year: "last_year",
    };
    return periodMap[current] || "yesterday";
  };

  const calculateGrowth = (current: number, previous: number) => {
    if (!previous || previous === 0) return { value: 0, isPositive: true };
    const growth = ((current - previous) / previous) * 100;
    return { value: growth, isPositive: growth >= 0 };
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(amount);
  };

  const formatNumber = (num: number) => {
    return new Intl.NumberFormat("id-ID").format(num);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-96">
        <div className="w-10 h-10 border-2 border-white/20 border-t-white rounded-full animate-spin" />
      </div>
    );
  }

  const criticalAlerts =
    (payments?.stuck_payments?.length || 0) +
    (inventory?.out_of_stock?.length || 0) +
    (fulfillment?.stuck_shipments || 0) +
    (fulfillment?.lost_packages || 0);

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-white">Executive Dashboard</h1>
          <p className="text-white/60 mt-1">Real-time business intelligence & operational metrics</p>
        </div>
        <div className="flex items-center gap-3">
          <select
            value={period}
            onChange={(e) => setPeriod(e.target.value)}
            className="px-4 py-2 rounded-xl bg-neutral-900 text-white border border-white/10 focus:outline-none focus:border-white/30 hover:border-white/20 transition-colors cursor-pointer"
          >
            <option value="today">Today</option>
            <option value="week">This Week</option>
            <option value="month">This Month</option>
            <option value="year">This Year</option>
          </select>
          <button
            onClick={loadDashboard}
            className="flex items-center gap-2 px-4 py-2 rounded-xl bg-white/10 text-white hover:bg-white/20 transition-colors"
          >
            <RefreshCcw size={18} />
            Refresh
          </button>
        </div>
      </div>

      {/* Info Banner - Show if no data */}
      {!loading && executive?.total_orders === 0 && (
        <div className="p-4 rounded-2xl bg-blue-500/10 border border-blue-500/20">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-xl bg-blue-500/20">
              <AlertCircle className="text-blue-400" size={20} />
            </div>
            <div className="flex-1">
              <p className="text-blue-400 font-semibold">No Data for Selected Period</p>
              <p className="text-white/60 text-sm">
                No orders found for &quot;{period}&quot;. Try selecting a different time period or create some test orders.
              </p>
            </div>
            <Link
              href="/admin/orders"
              className="px-4 py-2 rounded-xl bg-blue-500 text-white font-medium hover:bg-blue-600 transition-colors text-sm"
            >
              View All Orders
            </Link>
          </div>
        </div>
      )}

      {/* Critical Alerts Banner */}
      {criticalAlerts > 0 && (
        <div className="p-4 rounded-2xl bg-red-500/10 border border-red-500/20 animate-pulse">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-xl bg-red-500/20">
              <AlertTriangle className="text-red-400" size={24} />
            </div>
            <div className="flex-1">
              <p className="text-red-400 font-semibold">Critical Attention Required</p>
              <p className="text-white/60 text-sm">
                {criticalAlerts} critical issues detected: {payments?.stuck_payments?.length || 0} stuck payments,{" "}
                {inventory?.out_of_stock?.length || 0} out of stock, {fulfillment?.stuck_shipments || 0} delayed shipments
              </p>
            </div>
            <div className="flex gap-2">
              {payments?.stuck_payments && payments.stuck_payments.length > 0 && (
                <Link
                  href={`/admin/orders/${payments.stuck_payments[0].order_code}`}
                  className="px-4 py-2 rounded-xl bg-amber-500 text-white font-medium hover:bg-amber-600 transition-colors text-sm"
                >
                  Check Stuck Payment
                </Link>
              )}
              {inventory?.out_of_stock && inventory.out_of_stock.length > 0 && (
                <Link
                  href="/admin/products"
                  className="px-4 py-2 rounded-xl bg-red-500 text-white font-medium hover:bg-red-600 transition-colors text-sm"
                >
                  Restock Items
                </Link>
              )}
              {fulfillment?.stuck_shipments > 0 && (
                <Link
                  href="/admin/shipments"
                  className="px-4 py-2 rounded-xl bg-purple-500 text-white font-medium hover:bg-purple-600 transition-colors text-sm"
                >
                  Check Shipments
                </Link>
              )}
            </div>
          </div>
        </div>
      )}

      {/* Executive KPIs - TOP ROW */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="GMV (Gross Merchandise Value)"
          value={formatCurrency(executive?.gmv || 0)}
          subtitle={`${formatNumber(executive?.total_orders || 0)} total orders`}
          icon={<BarChart3 size={24} />}
          color="purple"
          trend={previousMetrics ? calculateGrowth(executive?.gmv || 0, previousMetrics.gmv) : undefined}
        />

        <StatCard
          title="Revenue (Paid Orders)"
          value={formatCurrency(executive?.revenue || 0)}
          subtitle={`${formatNumber(executive?.paid_orders || 0)} paid orders`}
          icon={<DollarSign size={24} />}
          color="emerald"
          trend={previousMetrics ? calculateGrowth(executive?.revenue || 0, previousMetrics.revenue) : undefined}
        />

        <StatCard
          title="Pending Revenue"
          value={formatCurrency(executive?.pending_revenue || 0)}
          subtitle={`Conversion: ${executive?.conversion_rate?.toFixed(1) || 0}%`}
          icon={<Clock size={24} />}
          color={executive?.pending_revenue && executive.pending_revenue > 0 ? "amber" : "blue"}
          trend={previousMetrics ? calculateGrowth(executive?.conversion_rate || 0, previousMetrics.conversion_rate) : undefined}
        />

        <StatCard
          title="Avg Order Value"
          value={formatCurrency(executive?.avg_order_value || 0)}
          subtitle="Per transaction"
          icon={<TrendingUp size={24} />}
          color="cyan"
          trend={previousMetrics ? calculateGrowth(executive?.avg_order_value || 0, previousMetrics.avg_order_value) : undefined}
        />
      </div>

      {/* Payment Monitor - REAL-TIME */}
      <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h2 className="text-lg font-semibold text-white flex items-center gap-2">
              <CreditCard size={20} />
              Payment Monitor (Real-time)
            </h2>
            <p className="text-white/60 text-sm mt-1">Live payment status tracking</p>
          </div>
          <Link href="/admin/orders?status=PENDING" className="text-blue-400 hover:text-blue-300 text-sm flex items-center gap-1">
            View all <ArrowUpRight size={14} />
          </Link>
        </div>

        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
          <div className="bg-white/5 rounded-xl p-4 text-center">
            <p className="text-2xl font-bold text-blue-400">{payments?.pending_count || 0}</p>
            <p className="text-white/60 text-sm">Pending</p>
            <p className="text-white/40 text-xs mt-1">{formatCurrency(payments?.pending_amount || 0)}</p>
          </div>
          <div className="bg-white/5 rounded-xl p-4 text-center">
            <p className="text-2xl font-bold text-amber-400">{payments?.expiring_soon_count || 0}</p>
            <p className="text-white/60 text-sm">Expiring Soon</p>
            <p className="text-white/40 text-xs mt-1">&lt; 1 hour left</p>
          </div>
          <div className="bg-white/5 rounded-xl p-4 text-center">
            <p className="text-2xl font-bold text-red-400">{payments?.stuck_payments?.length || 0}</p>
            <p className="text-white/60 text-sm">Stuck</p>
            <p className="text-white/40 text-xs mt-1">&gt; 1 hour pending</p>
          </div>
          <div className="bg-white/5 rounded-xl p-4 text-center">
            <p className="text-2xl font-bold text-emerald-400">{payments?.today_paid_count || 0}</p>
            <p className="text-white/60 text-sm">Paid Today</p>
            <p className="text-white/40 text-xs mt-1">{formatCurrency(payments?.today_paid_amount || 0)}</p>
          </div>
        </div>

        {/* Stuck Payments Alert */}
        {payments?.stuck_payments && payments.stuck_payments.length > 0 && (
          <div className="bg-red-500/10 border border-red-500/20 rounded-xl p-4">
            <div className="flex items-center gap-2 mb-3">
              <AlertCircle className="text-red-400" size={18} />
              <p className="text-red-400 font-semibold">Stuck Payments Detected - Action Required</p>
              <p className="text-white/60 text-xs ml-auto">
                Payments pending &gt; 1 hour need manual verification
              </p>
            </div>
            <div className="space-y-2 max-h-48 overflow-y-auto">
              {payments.stuck_payments.slice(0, 5).map((sp) => (
                <div key={sp.payment_id} className="flex items-center justify-between text-sm bg-white/5 rounded-lg p-3 hover:bg-white/10 transition-colors">
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <p className="text-white font-medium">{sp.order_code}</p>
                      <span className="px-2 py-0.5 rounded text-xs bg-amber-500/20 text-amber-400">
                        {sp.payment_method}
                      </span>
                    </div>
                    <p className="text-white/60 text-xs mt-1">
                      {sp.bank} • Pending {sp.hours_pending.toFixed(1)} hours
                    </p>
                  </div>
                  <div className="text-right flex items-center gap-3">
                    <div>
                      <p className="text-white font-semibold">{formatCurrency(sp.amount)}</p>
                      <p className="text-red-400 text-xs">{sp.hours_pending.toFixed(1)}h stuck</p>
                    </div>
                    <Link
                      href={`/admin/orders/${sp.order_code}`}
                      className="px-3 py-1.5 rounded-lg bg-amber-500 text-white text-xs font-medium hover:bg-amber-600 transition-colors"
                    >
                      Check Order
                    </Link>
                  </div>
                </div>
              ))}
            </div>
            <div className="mt-4 pt-4 border-t border-white/10">
              <p className="text-white/60 text-xs mb-2">
                <strong>What to do:</strong> Check if customer has paid via bank/e-wallet. If paid, manually verify and update order status.
              </p>
              <div className="flex gap-2">
                <Link
                  href="/admin/orders?status=PENDING"
                  className="px-4 py-2 rounded-lg bg-white/10 text-white text-sm hover:bg-white/20 transition-colors"
                >
                  View All Pending Orders
                </Link>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* System Health Monitor - NEW */}
      <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h2 className="text-lg font-semibold text-white flex items-center gap-2">
              <Activity size={20} />
              System Health Monitor
            </h2>
            <p className="text-white/60 text-sm mt-1">Real-time system performance & reliability</p>
          </div>
          <div className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-emerald-500/20">
            <div className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
            <span className="text-emerald-400 text-sm font-medium">All Systems Operational</span>
          </div>
        </div>

        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="bg-white/5 rounded-xl p-4">
            <p className="text-white/60 text-sm mb-2">Webhook Success Rate</p>
            <div className="flex items-end gap-2">
              <p className="text-2xl font-bold text-emerald-400">{systemHealth?.webhook_success_rate || 0}%</p>
              <p className="text-white/40 text-xs mb-1">last 24h</p>
            </div>
            <div className="mt-2 h-1.5 bg-white/10 rounded-full overflow-hidden">
              <div 
                className="h-full bg-emerald-500 rounded-full" 
                style={{ width: `${systemHealth?.webhook_success_rate || 0}%` }}
              />
            </div>
          </div>

          <div className="bg-white/5 rounded-xl p-4">
            <p className="text-white/60 text-sm mb-2">Payment Gateway</p>
            <div className="flex items-end gap-2">
              <p className="text-2xl font-bold text-blue-400">{systemHealth?.payment_gateway_latency || 0}ms</p>
              <p className="text-white/40 text-xs mb-1">avg latency</p>
            </div>
            <p className="text-emerald-400 text-xs mt-2">✓ Healthy</p>
          </div>

          <div className="bg-white/5 rounded-xl p-4">
            <p className="text-white/60 text-sm mb-2">Background Jobs</p>
            <div className="flex items-center gap-2 mt-2">
              {systemHealth?.background_jobs_healthy ? (
                <>
                  <CheckCircle className="text-emerald-400" size={20} />
                  <span className="text-emerald-400 text-sm">All Running</span>
                </>
              ) : (
                <>
                  <AlertCircle className="text-red-400" size={20} />
                  <span className="text-red-400 text-sm">Issues Detected</span>
                </>
              )}
            </div>
            <p className="text-white/40 text-xs mt-2">Payment expiry, tracking updates</p>
          </div>

          <div className="bg-white/5 rounded-xl p-4">
            <p className="text-white/60 text-sm mb-2">Last Tracking Update</p>
            <p className="text-white text-sm mt-2">
              {systemHealth?.last_tracking_update 
                ? new Date(systemHealth.last_tracking_update).toLocaleTimeString('id-ID')
                : 'N/A'}
            </p>
            <p className="text-white/40 text-xs mt-2">Shipment monitor active</p>
          </div>
        </div>
      </div>

      {/* Revenue Chart Visualization - NEW */}
      {chart && chart.data_points && chart.data_points.length > 0 && (
        <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
          <h2 className="text-lg font-semibold text-white mb-6">Revenue Trend (Last 7 Days)</h2>
          <div className="space-y-3">
            {chart.data_points.map((point, idx) => {
              const maxRevenue = Math.max(...chart.data_points.map(p => p.revenue));
              const widthPercent = maxRevenue > 0 ? (point.revenue / maxRevenue) * 100 : 0;
              
              return (
                <div key={idx} className="flex items-center gap-4">
                  <div className="w-24 text-white/60 text-sm">{point.date}</div>
                  <div className="flex-1">
                    <div className="flex items-center gap-3">
                      <div className="flex-1 h-8 bg-white/5 rounded-lg overflow-hidden">
                        <div 
                          className="h-full bg-gradient-to-r from-emerald-500 to-emerald-400 rounded-lg flex items-center px-3 transition-all duration-500"
                          style={{ width: `${widthPercent}%` }}
                        >
                          {widthPercent > 20 && (
                            <span className="text-white text-sm font-medium">
                              {formatCurrency(point.revenue)}
                            </span>
                          )}
                        </div>
                      </div>
                      {widthPercent <= 20 && (
                        <span className="text-white text-sm font-medium w-32 text-right">
                          {formatCurrency(point.revenue)}
                        </span>
                      )}
                      <span className="text-white/40 text-sm w-20 text-right">{point.orders} orders</span>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* Refund & Dispute Intelligence - NEW */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-lg font-semibold text-white flex items-center gap-2">
              <MessageSquare size={20} />
              Disputes & Refunds
            </h2>
            <Link href="/admin/disputes" className="text-blue-400 hover:text-blue-300 text-sm flex items-center gap-1">
              View all <ArrowUpRight size={14} />
            </Link>
          </div>

          <div className="grid grid-cols-2 gap-4 mb-6">
            <div className="bg-amber-500/10 border border-amber-500/20 rounded-xl p-4">
              <p className="text-amber-400 text-2xl font-bold">{disputes.length || 0}</p>
              <p className="text-white/60 text-sm">Open Disputes</p>
            </div>
            <div className="bg-red-500/10 border border-red-500/20 rounded-xl p-4">
              <p className="text-red-400 text-2xl font-bold">
                {disputes.filter((d: any) => d.status === 'PENDING_RESOLUTION').length || 0}
              </p>
              <p className="text-white/60 text-sm">Need Resolution</p>
            </div>
          </div>

          {disputes.length > 0 ? (
            <div className="space-y-2">
              {disputes.slice(0, 3).map((dispute: any) => (
                <Link
                  key={dispute.id}
                  href={`/admin/disputes/${dispute.id}`}
                  className="block p-3 rounded-xl bg-white/5 hover:bg-white/10 transition-colors"
                >
                  <div className="flex items-center justify-between">
                    <div className="flex-1 min-w-0">
                      <p className="text-white text-sm font-medium truncate">{dispute.title}</p>
                      <p className="text-white/40 text-xs mt-0.5">{dispute.dispute_code}</p>
                    </div>
                    <span className="px-2 py-1 rounded text-xs bg-amber-500/20 text-amber-400 ml-2">
                      {dispute.status}
                    </span>
                  </div>
                </Link>
              ))}
            </div>
          ) : (
            <div className="text-center py-8">
              <CheckCircle className="mx-auto text-emerald-400 mb-2" size={32} />
              <p className="text-white/60 text-sm">No open disputes</p>
            </div>
          )}
        </div>

        {/* Courier Performance - NEW */}
        <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-lg font-semibold text-white flex items-center gap-2">
              <Truck size={20} />
              Courier Performance
            </h2>
            <Link href="/admin/shipments" className="text-blue-400 hover:text-blue-300 text-sm flex items-center gap-1">
              Details <ArrowUpRight size={14} />
            </Link>
          </div>

          <div className="space-y-3">
            {courierPerformance && courierPerformance.length > 0 ? (
              courierPerformance.map((courier) => (
                <div key={courier.courier_name} className="p-4 rounded-xl bg-white/5">
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-white font-medium">{courier.courier_name}</span>
                    <span className={`text-sm font-semibold ${
                      courier.success_rate >= 98 ? 'text-emerald-400' : 
                      courier.success_rate >= 95 ? 'text-amber-400' : 'text-red-400'
                    }`}>
                      {courier.success_rate.toFixed(1)}%
                    </span>
                  </div>
                  <div className="flex items-center justify-between text-xs text-white/60">
                    <span>{courier.delivered} delivered</span>
                    <span>{courier.failed} failed</span>
                    <span>{courier.avg_delivery_days.toFixed(1)} days avg</span>
                  </div>
                  <div className="mt-2 h-1.5 bg-white/10 rounded-full overflow-hidden">
                    <div 
                      className={`h-full rounded-full ${
                        courier.success_rate >= 98 ? 'bg-emerald-500' : 
                        courier.success_rate >= 95 ? 'bg-amber-500' : 'bg-red-500'
                      }`}
                      style={{ width: `${courier.success_rate}%` }}
                    />
                  </div>
                </div>
              ))
            ) : (
              <div className="text-center py-8">
                <Truck className="mx-auto text-white/20 mb-2" size={32} />
                <p className="text-white/60 text-sm">No courier data available</p>
                <p className="text-white/40 text-xs mt-1">Data will appear after shipments are delivered</p>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Conversion Funnel */}
      <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
        <h2 className="text-lg font-semibold text-white mb-6">Conversion Funnel</h2>
        <div className="space-y-3">
          <div className="relative">
            <div className="flex items-center justify-between mb-1">
              <span className="text-white/80">Orders Created</span>
              <span className="text-white font-semibold">{formatNumber(funnel?.orders_created || 0)}</span>
            </div>
            <div className="h-3 bg-white/10 rounded-full overflow-hidden">
              <div className="h-full bg-blue-500" style={{ width: "100%" }} />
            </div>
          </div>

          <div className="relative">
            <div className="flex items-center justify-between mb-1">
              <span className="text-white/80">Orders Paid</span>
              <span className="text-emerald-400 font-semibold">
                {formatNumber(funnel?.orders_paid || 0)} ({funnel?.payment_rate?.toFixed(1) || 0}%)
              </span>
            </div>
            <div className="h-3 bg-white/10 rounded-full overflow-hidden">
              <div className="h-full bg-emerald-500" style={{ width: `${funnel?.payment_rate || 0}%` }} />
            </div>
          </div>

          <div className="relative">
            <div className="flex items-center justify-between mb-1">
              <span className="text-white/80">Orders Shipped</span>
              <span className="text-purple-400 font-semibold">
                {formatNumber(funnel?.orders_shipped || 0)} ({funnel?.fulfillment_rate?.toFixed(1) || 0}%)
              </span>
            </div>
            <div className="h-3 bg-white/10 rounded-full overflow-hidden">
              <div className="h-full bg-purple-500" style={{ width: `${funnel?.fulfillment_rate || 0}%` }} />
            </div>
          </div>

          <div className="relative">
            <div className="flex items-center justify-between mb-1">
              <span className="text-white/80">Orders Delivered</span>
              <span className="text-cyan-400 font-semibold">
                {formatNumber(funnel?.orders_delivered || 0)} ({funnel?.delivery_rate?.toFixed(1) || 0}%)
              </span>
            </div>
            <div className="h-3 bg-white/10 rounded-full overflow-hidden">
              <div className="h-full bg-cyan-500" style={{ width: `${funnel?.delivery_rate || 0}%` }} />
            </div>
          </div>
        </div>
      </div>

      {/* Inventory & Customer Insights Row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Inventory Alerts */}
        <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-lg font-semibold text-white flex items-center gap-2">
              <Package size={20} />
              Inventory Alerts
            </h2>
            <Link href="/admin/products" className="text-blue-400 hover:text-blue-300 text-sm flex items-center gap-1">
              Manage <ArrowUpRight size={14} />
            </Link>
          </div>

          <div className="space-y-4">
            <div className="bg-red-500/10 border border-red-500/20 rounded-xl p-4">
              <p className="text-red-400 font-semibold mb-2">Out of Stock: {inventory?.out_of_stock?.length || 0}</p>
              {inventory?.out_of_stock && inventory.out_of_stock.length > 0 && (
                <div className="space-y-2 max-h-32 overflow-y-auto">
                  {inventory.out_of_stock.slice(0, 3).map((item) => (
                    <div key={item.product_id} className="text-sm text-white/80">
                      • {item.product_name} ({item.category})
                    </div>
                  ))}
                </div>
              )}
            </div>

            <div className="bg-amber-500/10 border border-amber-500/20 rounded-xl p-4">
              <p className="text-amber-400 font-semibold mb-2">Low Stock: {inventory?.low_stock?.length || 0}</p>
              {inventory?.low_stock && inventory.low_stock.length > 0 && (
                <div className="space-y-2 max-h-32 overflow-y-auto">
                  {inventory.low_stock.slice(0, 3).map((item) => (
                    <div key={item.product_id} className="text-sm text-white/80 flex justify-between">
                      <span>• {item.product_name}</span>
                      <span className="text-amber-400">{item.stock} left</span>
                    </div>
                  ))}
                </div>
              )}
            </div>

            <div className="bg-blue-500/10 border border-blue-500/20 rounded-xl p-4">
              <p className="text-blue-400 font-semibold mb-2">Fast Moving: {inventory?.fast_moving?.length || 0}</p>
              {inventory?.fast_moving && inventory.fast_moving.length > 0 && (
                <div className="space-y-2 max-h-32 overflow-y-auto">
                  {inventory.fast_moving.slice(0, 3).map((item) => (
                    <div key={item.product_id} className="text-sm text-white/80 flex justify-between">
                      <span>• {item.product_name}</span>
                      <span className="text-blue-400">{item.days_of_stock.toFixed(1)} days left</span>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Customer Insights */}
        <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-lg font-semibold text-white flex items-center gap-2">
              <Users size={20} />
              Customer Insights
            </h2>
            <Link href="/admin/customers" className="text-blue-400 hover:text-blue-300 text-sm flex items-center gap-1">
              View all <ArrowUpRight size={14} />
            </Link>
          </div>

          <div className="grid grid-cols-3 gap-4 mb-6">
            <div className="bg-white/5 rounded-xl p-4 text-center">
              <p className="text-2xl font-bold text-purple-400">{formatNumber(customers?.total_customers || 0)}</p>
              <p className="text-white/60 text-sm">Total</p>
            </div>
            <div className="bg-white/5 rounded-xl p-4 text-center">
              <p className="text-2xl font-bold text-emerald-400">{formatNumber(customers?.active_customers || 0)}</p>
              <p className="text-white/60 text-sm">Active</p>
            </div>
            <div className="bg-white/5 rounded-xl p-4 text-center">
              <p className="text-2xl font-bold text-blue-400">{formatNumber(customers?.new_customers || 0)}</p>
              <p className="text-white/60 text-sm">New (30d)</p>
            </div>
          </div>

          {/* Customer Segments */}
          <div className="space-y-2">
            <p className="text-white/80 font-medium mb-3">Customer Segments (RFM)</p>
            {customers?.segments?.map((seg) => (
              <div key={seg.segment} className="flex items-center justify-between text-sm bg-white/5 rounded-lg p-3">
                <span className="text-white/80">{seg.segment}</span>
                <div className="text-right">
                  <p className="text-white font-medium">{seg.count} customers</p>
                  <p className="text-white/60 text-xs">{formatCurrency(seg.avg_value)} avg</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Bottom Section */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Top Products */}
        <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
          <h2 className="text-lg font-semibold text-white mb-6">Top Selling Products</h2>
          <div className="space-y-3">
            {executive?.top_products?.slice(0, 5).map((product, idx) => (
              <div key={product.product_id} className="flex items-center gap-4 bg-white/5 rounded-xl p-4">
                <div className="w-8 h-8 rounded-full bg-purple-500/20 flex items-center justify-center text-purple-400 font-bold">
                  {idx + 1}
                </div>
                <div className="flex-1">
                  <p className="text-white font-medium">{product.product_name}</p>
                  <p className="text-white/60 text-sm">{product.total_sold} sold</p>
                </div>
                <div className="text-right">
                  <p className="text-emerald-400 font-semibold">{formatCurrency(product.revenue)}</p>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Payment Methods */}
        <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
          <h2 className="text-lg font-semibold text-white mb-6">Payment Methods</h2>
          <div className="space-y-3">
            {executive?.payment_methods?.map((method) => (
              <div key={method.method} className="flex items-center justify-between bg-white/5 rounded-xl p-4">
                <div>
                  <p className="text-white font-medium">{method.method || "Unknown"}</p>
                  <p className="text-white/60 text-sm">{method.count} transactions</p>
                </div>
                <div className="text-right">
                  <p className="text-emerald-400 font-semibold">{formatCurrency(method.amount)}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Quick Actions */}
      <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
        <h2 className="text-lg font-semibold text-white mb-6">Quick Actions</h2>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
          <Link href="/admin/orders" className="p-4 rounded-xl bg-white/5 hover:bg-white/10 transition-colors text-center">
            <FileText className="mx-auto text-blue-400 mb-2" size={24} />
            <p className="text-white font-medium">Orders</p>
          </Link>
          <Link href="/admin/products" className="p-4 rounded-xl bg-white/5 hover:bg-white/10 transition-colors text-center">
            <Package className="mx-auto text-purple-400 mb-2" size={24} />
            <p className="text-white font-medium">Products</p>
          </Link>
          <Link href="/admin/shipments" className="p-4 rounded-xl bg-white/5 hover:bg-white/10 transition-colors text-center">
            <Truck className="mx-auto text-amber-400 mb-2" size={24} />
            <p className="text-white font-medium">Shipments</p>
          </Link>
          <Link href="/admin/audit" className="p-4 rounded-xl bg-white/5 hover:bg-white/10 transition-colors text-center">
            <Activity className="mx-auto text-emerald-400 mb-2" size={24} />
            <p className="text-white font-medium">Audit Logs</p>
          </Link>
        </div>
      </div>
    </div>
  );
}
