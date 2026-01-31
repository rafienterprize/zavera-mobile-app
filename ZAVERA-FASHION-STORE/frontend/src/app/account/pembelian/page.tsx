"use client";

import { useState, useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import Image from "next/image";
import { motion, AnimatePresence } from "framer-motion";
import { useAuth } from "@/context/AuthContext";
import { useToast } from "@/components/ui/Toast";
import api from "@/lib/api";
import LoadingSpinner from "@/components/ui/LoadingSpinner";
import { Suspense } from "react";

// Bank logos
const BANK_LOGOS: Record<string, string> = {
  bca: "/images/banks/bca.png",
  bri: "/images/banks/bri.png",
  bni: "/images/banks/bni.png",
  mandiri: "/images/banks/mandiri.png",
  permata: "/images/banks/permata.png",
  gopay: "/images/payments/gopay.png",
  qris: "/images/payments/qris.png",
};

// Status styles - Tokopedia-style
const STATUS_CONFIG: Record<string, { bg: string; text: string; label: string }> = {
  // Berlangsung (Ongoing)
  PAID: { bg: "bg-blue-100", text: "text-blue-700", label: "Dibayar" },
  DIBAYAR: { bg: "bg-blue-100", text: "text-blue-700", label: "Dibayar" },
  PACKING: { bg: "bg-blue-100", text: "text-blue-700", label: "Dikemas" },
  DIPROSES: { bg: "bg-blue-100", text: "text-blue-700", label: "Diproses" },
  PROCESSING: { bg: "bg-blue-100", text: "text-blue-700", label: "Diproses" },
  SHIPPED: { bg: "bg-violet-100", text: "text-violet-700", label: "Dikirim" },
  DIKIRIM: { bg: "bg-violet-100", text: "text-violet-700", label: "Dikirim" },
  // Selesai (Completed)
  DELIVERED: { bg: "bg-emerald-100", text: "text-emerald-700", label: "Terkirim" },
  TERKIRIM: { bg: "bg-emerald-100", text: "text-emerald-700", label: "Terkirim" },
  COMPLETED: { bg: "bg-emerald-100", text: "text-emerald-700", label: "Selesai" },
  SELESAI: { bg: "bg-emerald-100", text: "text-emerald-700", label: "Selesai" },
  // Tidak Berhasil (Failed)
  CANCELLED: { bg: "bg-gray-100", text: "text-gray-600", label: "Dibatalkan" },
  DIBATALKAN: { bg: "bg-gray-100", text: "text-gray-600", label: "Dibatalkan" },
  EXPIRED: { bg: "bg-red-100", text: "text-red-600", label: "Kadaluarsa" },
  KADALUARSA: { bg: "bg-red-100", text: "text-red-600", label: "Kadaluarsa" },
  FAILED: { bg: "bg-red-100", text: "text-red-600", label: "Gagal" },
  REFUNDED: { bg: "bg-orange-100", text: "text-orange-700", label: "Dikembalikan" },
  // Menunggu
  PENDING: { bg: "bg-amber-100", text: "text-amber-700", label: "Menunggu" },
  MENUNGGU_PEMBAYARAN: { bg: "bg-amber-100", text: "text-amber-700", label: "Menunggu Pembayaran" },
};

// Filter tabs - Tokopedia style
const FILTER_TABS = [
  { key: "all", label: "Semua" },
  { key: "ongoing", label: "Berlangsung" },
  { key: "completed", label: "Berhasil" },
  { key: "failed", label: "Tidak Berhasil" },
];

interface PendingOrder {
  order_id: number;
  order_code: string;
  total_amount: number;
  item_count: number;
  item_summary: string;
  created_at: string;
  has_payment: boolean;
  payment_method?: string;
  bank?: string;
  bank_logo?: string;
  va_number_masked?: string;
  expiry_time?: string;
  remaining_seconds?: number;
}

interface TransactionHistoryItem {
  order_id: number;
  order_code: string;
  total_amount: number;
  item_count: number;
  item_summary: string;
  product_image?: string;
  status: string;
  payment_method?: string;
  bank?: string;
  resi?: string;
  courier_name?: string;
  courier_service?: string;
  shipment_status?: string;
  tracking_number?: string;
  paid_at?: string;
  shipped_at?: string;
  delivered_at?: string;
  completed_at?: string;
  cancelled_at?: string;
  created_at: string;
}

// Mini Countdown for pending payments
const MiniCountdown = ({ expiryTime }: { expiryTime: string }) => {
  const [remaining, setRemaining] = useState(0);

  useEffect(() => {
    const calculateRemaining = () => {
      const expiry = new Date(expiryTime).getTime();
      const now = Date.now();
      return Math.max(0, Math.floor((expiry - now) / 1000));
    };

    setRemaining(calculateRemaining());
    const interval = setInterval(() => {
      setRemaining(calculateRemaining());
    }, 1000);

    return () => clearInterval(interval);
  }, [expiryTime]);

  if (remaining <= 0) {
    return <span className="text-xs text-red-600 font-medium">Waktu habis</span>;
  }

  const hours = Math.floor(remaining / 3600);
  const minutes = Math.floor((remaining % 3600) / 60);
  const seconds = remaining % 60;

  return (
    <div className="flex items-center gap-1">
      <div className="px-1.5 py-0.5 bg-primary text-white text-xs font-mono rounded">
        {hours.toString().padStart(2, "0")}
      </div>
      <span className="text-primary text-xs">:</span>
      <div className="px-1.5 py-0.5 bg-primary text-white text-xs font-mono rounded">
        {minutes.toString().padStart(2, "0")}
      </div>
      <span className="text-primary text-xs">:</span>
      <div className="px-1.5 py-0.5 bg-primary text-white text-xs font-mono rounded">
        {seconds.toString().padStart(2, "0")}
      </div>
    </div>
  );
};

// Status Badge - Tokopedia style
const StatusBadge = ({ status }: { status: string }) => {
  const config = STATUS_CONFIG[status] || { bg: "bg-gray-100", text: "text-gray-600", label: status };
  return (
    <span className={`inline-flex items-center px-2.5 py-1 text-xs font-medium rounded-full ${config.bg} ${config.text}`}>
      {config.label}
    </span>
  );
};

// Pending Order Card
const PendingOrderCard = ({ order, onClick }: { order: PendingOrder; onClick: () => void }) => {
  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      className="bg-white rounded-xl border border-gray-200 overflow-hidden hover:shadow-md transition cursor-pointer"
      onClick={onClick}
    >
      <div className="p-4 border-b border-gray-100 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="w-9 h-9 bg-primary rounded-full flex items-center justify-center">
            <svg className="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <div>
            <p className="font-medium text-gray-900 text-sm">{order.order_code}</p>
            <p className="text-xs text-gray-500">
              {new Date(order.created_at).toLocaleDateString("id-ID", { day: "numeric", month: "short", year: "numeric" })}
            </p>
          </div>
        </div>
        {order.has_payment && order.expiry_time && <MiniCountdown expiryTime={order.expiry_time} />}
      </div>

      <div className="p-4">
        <h3 className="font-medium text-gray-900">{order.item_summary}</h3>
        {order.item_count > 1 && <p className="text-sm text-gray-500 mt-0.5">+{order.item_count - 1} produk lainnya</p>}

        {order.has_payment && order.bank && (
          <div className="flex items-center gap-3 mt-3 p-3 bg-gray-50 rounded-lg">
            <div className="w-12 h-8 bg-white rounded border flex items-center justify-center p-1">
              <Image
                src={BANK_LOGOS[order.bank] || "/images/banks/default.svg"}
                alt={order.bank.toUpperCase()}
                width={40}
                height={24}
                className="object-contain"
              />
            </div>
            <div className="flex-1">
              <p className="text-sm font-medium text-gray-900">{order.bank.toUpperCase()} VA</p>
              <p className="text-xs text-gray-500 font-mono">{order.va_number_masked}</p>
            </div>
          </div>
        )}

        <div className="flex items-center justify-between mt-4 pt-3 border-t border-gray-100">
          <div>
            <p className="text-xs text-gray-500">Total</p>
            <p className="text-lg font-bold text-gray-900">Rp{order.total_amount.toLocaleString("id-ID")}</p>
          </div>
          <button className="px-5 py-2.5 bg-primary text-white text-sm font-medium rounded-lg hover:bg-primary/90 transition">
            {order.has_payment ? "Bayar" : "Pilih Pembayaran"}
          </button>
        </div>
      </div>
    </motion.div>
  );
};


// Transaction History Card - Tokopedia style
const TransactionCard = ({ order }: { order: TransactionHistoryItem }) => {
  const router = useRouter();
  const { showToast } = useToast();

  const isOngoing = ["PAID", "DIBAYAR", "PACKING", "DIPROSES", "PROCESSING", "SHIPPED", "DIKIRIM"].includes(order.status);
  const isCompleted = ["DELIVERED", "TERKIRIM", "COMPLETED", "SELESAI"].includes(order.status);
  const isFailed = ["CANCELLED", "DIBATALKAN", "EXPIRED", "KADALUARSA", "FAILED", "REFUNDED"].includes(order.status);

  const copyResi = () => {
    if (order.resi) {
      navigator.clipboard.writeText(order.resi);
      showToast("Nomor resi disalin!", "success");
    }
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      className="bg-white rounded-xl border border-gray-200 overflow-hidden"
    >
      {/* Header */}
      <div className="px-4 py-3 border-b border-gray-100 flex items-center justify-between bg-gray-50">
        <div className="flex items-center gap-3">
          <svg className="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z" />
          </svg>
          <span className="text-sm text-gray-600">Belanja</span>
          <span className="text-sm text-gray-400">
            {new Date(order.created_at).toLocaleDateString("id-ID", { day: "numeric", month: "short", year: "numeric" })}
          </span>
          <StatusBadge status={order.status} />
          <span className="text-xs text-gray-400 font-mono">{order.order_code}</span>
        </div>
      </div>

      {/* Content */}
      <div className="p-4">
        <div className="flex gap-4">
          {/* Product Image */}
          <div className="w-16 h-16 bg-gray-100 rounded-lg overflow-hidden flex-shrink-0">
            {order.product_image ? (
              <Image
                src={order.product_image}
                alt={order.item_summary}
                width={64}
                height={64}
                className="w-full h-full object-cover"
              />
            ) : (
              <div className="w-full h-full flex items-center justify-center">
                <svg className="w-6 h-6 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                </svg>
              </div>
            )}
          </div>

          {/* Product Info */}
          <div className="flex-1 min-w-0">
            <h3 className="font-medium text-gray-900">
              {order.item_summary}
            </h3>
            {order.item_count > 1 && (
              <p className="text-sm text-gray-500 mt-0.5">
                +{order.item_count - 1} produk lainnya
              </p>
            )}
            <p className="text-sm text-gray-500 mt-0.5">
              {order.item_count} {order.item_count === 1 ? "barang" : "barang"}
              {order.payment_method && ` • ${order.bank?.toUpperCase() || order.payment_method}`}
            </p>
            
            {/* Shipping Info for ongoing orders */}
            {isOngoing && order.courier_name && (
              <p className="text-sm text-gray-500 mt-1">
                {order.courier_name} {order.courier_service && `- ${order.courier_service}`}
              </p>
            )}
          </div>

          {/* Price */}
          <div className="text-right flex-shrink-0">
            <p className="text-xs text-gray-500">Total Belanja</p>
            <p className="font-bold text-gray-900">Rp{order.total_amount.toLocaleString("id-ID")}</p>
          </div>
        </div>

        {/* Resi Info - Show for shipped orders */}
        {order.resi && (isOngoing || isCompleted) && (
          <div className="mt-4 p-3 bg-blue-50 rounded-lg border border-blue-100">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-xs text-blue-600 font-medium">Nomor Resi</p>
                <p className="font-mono text-sm font-bold text-gray-900">{order.resi}</p>
              </div>
              <button
                onClick={(e) => { e.stopPropagation(); copyResi(); }}
                className="p-2 hover:bg-blue-100 rounded-lg transition"
              >
                <svg className="w-4 h-4 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                </svg>
              </button>
            </div>
          </div>
        )}

        {/* Refund notice for refunded orders */}
        {order.status === "REFUNDED" && (
          <div className="mt-4 p-3 bg-orange-50 rounded-lg border border-orange-100">
            <div className="flex items-start gap-2">
              <svg className="w-5 h-5 text-orange-600 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <div className="flex-1">
                <p className="text-sm font-medium text-orange-800">Pembayaran Dikembalikan</p>
                <p className="text-sm text-orange-700 mt-1">
                  Dana sudah dikembalikan ke metode pembayaran yang kamu pakai.{" "}
                  <Link href={`/orders/${order.order_code}`} className="font-medium text-orange-800 hover:underline">
                    Lihat Detail Refund
                  </Link>
                </p>
              </div>
            </div>
          </div>
        )}

        {/* Actions */}
        <div className="mt-4 pt-3 border-t border-gray-100 flex items-center justify-end gap-3">
          <Link
            href={`/orders/${order.order_code}`}
            className="text-sm font-medium text-primary hover:underline"
          >
            Lihat Detail Transaksi
          </Link>
          
          {/* Track button for shipped orders */}
          {order.resi && ["SHIPPED", "DIKIRIM"].includes(order.status) && (
            <a
              href={`https://cekresi.com/?noresi=${order.resi}`}
              target="_blank"
              rel="noopener noreferrer"
              className="px-4 py-2 bg-primary text-white text-sm font-medium rounded-lg hover:bg-primary/90 transition"
            >
              Lacak
            </a>
          )}

          {/* Buy again for completed/failed */}
          {(isCompleted || isFailed) && (
            <button
              onClick={() => router.push("/")}
              className="px-4 py-2 bg-primary text-white text-sm font-medium rounded-lg hover:bg-primary/90 transition"
            >
              Beli Lagi
            </button>
          )}

          {/* More options */}
          <button className="p-2 hover:bg-gray-100 rounded-lg transition">
            <svg className="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
            </svg>
          </button>
        </div>
      </div>
    </motion.div>
  );
};

// Empty State
const EmptyState = ({ type }: { type: "pending" | "history" }) => (
  <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="text-center py-16">
    <div className="w-20 h-20 mx-auto mb-4 bg-gray-100 rounded-full flex items-center justify-center">
      <svg className="w-10 h-10 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z" />
      </svg>
    </div>
    <h3 className="text-lg font-medium text-gray-900 mb-1">
      {type === "pending" ? "Tidak Ada Pesanan Menunggu" : "Belum Ada Transaksi"}
    </h3>
    <p className="text-gray-500 mb-6">
      {type === "pending" ? "Semua pesanan sudah dibayar" : "Mulai belanja untuk melihat transaksi"}
    </p>
    <Link href="/" className="inline-flex items-center gap-2 px-6 py-3 bg-primary text-white font-medium rounded-lg hover:bg-primary/90 transition">
      Mulai Belanja
    </Link>
  </motion.div>
);


// Main Page Content
function PembelianContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { isAuthenticated, isLoading: authLoading } = useAuth();
  const { showToast } = useToast();

  // Get initial tab from URL
  const initialTab = searchParams.get("tab") === "history" ? "history" : "pending";
  const initialFilter = searchParams.get("filter") || "all";

  const [activeTab, setActiveTab] = useState<"pending" | "history">(initialTab as "pending" | "history");
  const [filter, setFilter] = useState(initialFilter);
  const [loading, setLoading] = useState(true);
  const [pendingOrders, setPendingOrders] = useState<PendingOrder[]>([]);
  const [historyOrders, setHistoryOrders] = useState<TransactionHistoryItem[]>([]);
  const [pendingTotal, setPendingTotal] = useState(0);
  const [historyTotal, setHistoryTotal] = useState(0);
  const [page, setPage] = useState(1);
  const pageSize = 10;

  useEffect(() => {
    if (!authLoading && !isAuthenticated) {
      router.push("/login?redirect=/account/pembelian");
    }
  }, [authLoading, isAuthenticated, router]);

  useEffect(() => {
    if (isAuthenticated) {
      if (activeTab === "pending") {
        loadPendingOrders();
      } else {
        loadTransactionHistory();
      }
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [activeTab, filter, page, isAuthenticated]);

  // Update URL when tab/filter changes
  useEffect(() => {
    const params = new URLSearchParams();
    if (activeTab === "history") {
      params.set("tab", "history");
      if (filter !== "all") params.set("filter", filter);
    }
    const newUrl = params.toString() ? `?${params.toString()}` : "/account/pembelian";
    window.history.replaceState({}, "", newUrl);
  }, [activeTab, filter]);

  const loadPendingOrders = async () => {
    try {
      setLoading(true);
      const response = await api.get(`/pembelian/pending?page=${page}&page_size=${pageSize}`);
      setPendingOrders(response.data.orders || []);
      setPendingTotal(response.data.total_count || 0);
    } catch (err) {
      console.error("Failed to load pending orders:", err);
      showToast("Gagal memuat pesanan", "error");
    } finally {
      setLoading(false);
    }
  };

  const loadTransactionHistory = async () => {
    try {
      setLoading(true);
      const response = await api.get(`/pembelian/history?page=${page}&page_size=${pageSize}&filter=${filter}`);
      setHistoryOrders(response.data.orders || []);
      setHistoryTotal(response.data.total_count || 0);
    } catch (err) {
      console.error("Failed to load transaction history:", err);
      showToast("Gagal memuat riwayat transaksi", "error");
    } finally {
      setLoading(false);
    }
  };

  const handleTabChange = (tab: "pending" | "history") => {
    setActiveTab(tab);
    setPage(1);
    if (tab === "pending") setFilter("all");
  };

  const handleFilterChange = (newFilter: string) => {
    setFilter(newFilter);
    setPage(1);
  };

  const handlePendingOrderClick = (order: PendingOrder) => {
    // If order has payment, go to payment detail
    // If order doesn't have payment, go to payment selection
    if (order.has_payment) {
      router.push(`/checkout/payment/detail?order_id=${order.order_id}`);
    } else {
      router.push(`/checkout/payment?order_id=${order.order_id}`);
    }
  };

  if (authLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <LoadingSpinner />
      </div>
    );
  }

  const totalPages = Math.ceil((activeTab === "pending" ? pendingTotal : historyTotal) / pageSize);

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white border-b">
        <div className="max-w-4xl mx-auto px-4 py-4 flex items-center gap-3">
          <Link href="/" className="p-2 hover:bg-gray-100 rounded-lg transition">
            <svg className="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M15 19l-7-7 7-7" />
            </svg>
          </Link>
          <h1 className="text-lg font-semibold text-gray-900">Pembelian</h1>
        </div>
      </div>

      {/* Main Tabs */}
      <div className="bg-white border-b sticky top-0 z-40">
        <div className="max-w-4xl mx-auto px-4">
          <div className="flex">
            <button
              onClick={() => handleTabChange("pending")}
              className={`relative flex-1 py-4 text-sm font-medium transition ${
                activeTab === "pending" ? "text-primary" : "text-gray-500 hover:text-gray-700"
              }`}
            >
              <span className="flex items-center justify-center gap-2">
                Menunggu Pembayaran
                {pendingTotal > 0 && (
                  <span className="px-2 py-0.5 bg-red-500 text-white text-xs rounded-full">{pendingTotal}</span>
                )}
              </span>
              {activeTab === "pending" && (
                <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-primary transition-all duration-300" />
              )}
            </button>
            <button
              onClick={() => handleTabChange("history")}
              className={`relative flex-1 py-4 text-sm font-medium transition ${
                activeTab === "history" ? "text-primary" : "text-gray-500 hover:text-gray-700"
              }`}
            >
              Daftar Transaksi
              {activeTab === "history" && (
                <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-primary transition-all duration-300" />
              )}
            </button>
          </div>
        </div>
      </div>

      {/* Filter Tabs - Only show for Daftar Transaksi */}
      {activeTab === "history" && (
        <div className="bg-white border-b">
          <div className="max-w-4xl mx-auto px-4">
            <div className="flex items-center gap-2 py-3 overflow-x-auto">
              <span className="text-sm text-gray-500 mr-2">Status</span>
              {FILTER_TABS.map((tab) => (
                <button
                  key={tab.key}
                  onClick={() => handleFilterChange(tab.key)}
                  className={`px-4 py-1.5 text-sm font-medium rounded-full border transition whitespace-nowrap ${
                    filter === tab.key
                      ? "bg-primary text-white border-primary"
                      : "bg-white text-gray-600 border-gray-300 hover:border-gray-400"
                  }`}
                >
                  {tab.label}
                </button>
              ))}
              {filter !== "all" && (
                <button
                  onClick={() => handleFilterChange("all")}
                  className="text-sm text-primary hover:underline ml-2"
                >
                  Reset Filter
                </button>
              )}
            </div>
          </div>
        </div>
      )}

      {/* Content */}
      <div className="max-w-4xl mx-auto px-4 py-6">
        {loading ? (
          <div className="flex items-center justify-center py-16">
            <LoadingSpinner />
          </div>
        ) : (
          <AnimatePresence mode="wait">
            {activeTab === "pending" ? (
              <motion.div
                key="pending"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                className="space-y-4"
              >
                {pendingOrders.length === 0 ? (
                  <EmptyState type="pending" />
                ) : (
                  pendingOrders.map((order) => (
                    <PendingOrderCard key={order.order_id} order={order} onClick={() => handlePendingOrderClick(order)} />
                  ))
                )}
              </motion.div>
            ) : (
              <motion.div
                key="history"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                className="space-y-4"
              >
                {historyOrders.length === 0 ? (
                  <EmptyState type="history" />
                ) : (
                  historyOrders.map((order) => <TransactionCard key={order.order_id} order={order} />)
                )}
              </motion.div>
            )}
          </AnimatePresence>
        )}

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="flex items-center justify-center gap-2 mt-8">
            <button
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page === 1}
              className="p-2 border rounded-lg disabled:opacity-30 hover:bg-gray-50 transition"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
              </svg>
            </button>
            <div className="flex items-center gap-1">
              {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
                let pageNum = i + 1;
                if (totalPages > 5) {
                  if (page <= 3) pageNum = i + 1;
                  else if (page >= totalPages - 2) pageNum = totalPages - 4 + i;
                  else pageNum = page - 2 + i;
                }
                return (
                  <button
                    key={pageNum}
                    onClick={() => setPage(pageNum)}
                    className={`w-10 h-10 rounded-lg text-sm font-medium transition ${
                      page === pageNum ? "bg-primary text-white" : "hover:bg-gray-100"
                    }`}
                  >
                    {pageNum}
                  </button>
                );
              })}
            </div>
            <button
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              disabled={page >= totalPages}
              className="p-2 border rounded-lg disabled:opacity-30 hover:bg-gray-50 transition"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
              </svg>
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

export default function PembelianPage() {
  return (
    <Suspense fallback={<div className="min-h-screen flex items-center justify-center"><LoadingSpinner /></div>}>
      <PembelianContent />
    </Suspense>
  );
}
