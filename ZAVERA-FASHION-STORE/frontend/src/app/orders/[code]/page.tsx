"use client";

import { useEffect, useState, useCallback } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import Image from "next/image";
import { motion } from "framer-motion";
import { useAuth } from "@/context/AuthContext";
import { useToast } from "@/components/ui/Toast";
import api from "@/lib/api";
import LoadingSpinner from "@/components/ui/LoadingSpinner";

// SVG Icons for timeline (professional, minimal style)
const TimelineIcons = {
  placed: (
    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth={1.5}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M9 12h3.75M9 15h3.75M9 18h3.75m3 .75H18a2.25 2.25 0 002.25-2.25V6.108c0-1.135-.845-2.098-1.976-2.192a48.424 48.424 0 00-1.123-.08m-5.801 0c-.065.21-.1.433-.1.664 0 .414.336.75.75.75h4.5a.75.75 0 00.75-.75 2.25 2.25 0 00-.1-.664m-5.8 0A2.251 2.251 0 0113.5 2.25H15c1.012 0 1.867.668 2.15 1.586m-5.8 0c-.376.023-.75.05-1.124.08C9.095 4.01 8.25 4.973 8.25 6.108V8.25m0 0H4.875c-.621 0-1.125.504-1.125 1.125v11.25c0 .621.504 1.125 1.125 1.125h9.75c.621 0 1.125-.504 1.125-1.125V9.375c0-.621-.504-1.125-1.125-1.125H8.25zM6.75 12h.008v.008H6.75V12zm0 3h.008v.008H6.75V15zm0 3h.008v.008H6.75V18z" />
    </svg>
  ),
  paid: (
    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth={1.5}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M2.25 8.25h19.5M2.25 9h19.5m-16.5 5.25h6m-6 2.25h3m-3.75 3h15a2.25 2.25 0 002.25-2.25V6.75A2.25 2.25 0 0019.5 4.5h-15a2.25 2.25 0 00-2.25 2.25v10.5A2.25 2.25 0 004.5 19.5z" />
    </svg>
  ),
  processing: (
    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth={1.5}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M21 7.5l-9-5.25L3 7.5m18 0l-9 5.25m9-5.25v9l-9 5.25M3 7.5l9 5.25M3 7.5v9l9 5.25m0-9v9" />
    </svg>
  ),
  shipped: (
    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth={1.5}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M8.25 18.75a1.5 1.5 0 01-3 0m3 0a1.5 1.5 0 00-3 0m3 0h6m-9 0H3.375a1.125 1.125 0 01-1.125-1.125V14.25m17.25 4.5a1.5 1.5 0 01-3 0m3 0a1.5 1.5 0 00-3 0m3 0h1.125c.621 0 1.129-.504 1.09-1.124a17.902 17.902 0 00-3.213-9.193 2.056 2.056 0 00-1.58-.86H14.25M16.5 18.75h-2.25m0-11.177v-.958c0-.568-.422-1.048-.987-1.106a48.554 48.554 0 00-10.026 0 1.106 1.106 0 00-.987 1.106v7.635m12-6.677v6.677m0 4.5v-4.5m0 0h-12" />
    </svg>
  ),
  delivered: (
    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth={1.5}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>
  ),
};

// Order status timeline steps
const ORDER_STEPS = [
  { key: "placed", label: "Pesanan Dibuat" },
  { key: "paid", label: "Pembayaran Diterima" },
  { key: "processing", label: "Diproses" },
  { key: "shipped", label: "Dikirim" },
  { key: "delivered", label: "Selesai" },
];

// Status mapping to step index
const STATUS_TO_STEP: Record<string, number> = {
  PENDING: 0,
  MENUNGGU_PEMBAYARAN: 0,
  PAID: 1,
  PACKING: 2,
  PROCESSING: 2,
  SHIPPED: 3,
  DELIVERED: 4,
  COMPLETED: 4,
  CANCELLED: -1,
  EXPIRED: -1,
  KADALUARSA: -1,
  FAILED: -1,
};

interface OrderItem {
  product_id: number;
  product_name: string;
  product_image?: string;
  quantity: number;
  price_per_unit: number;
  subtotal: number;
  size?: string;
}

interface TrackingEvent {
  status: string;
  description: string;
  location?: string;
  event_time: string;
}

interface RefundItem {
  product_id: number;
  product_name: string;
  quantity: number;
  price_per_unit: number;
  subtotal: number;
}

interface RefundStatusHistory {
  status: string;
  changed_at: string;
  changed_by?: string;
  notes?: string;
}

interface Refund {
  id: number;
  refund_code: string;
  order_id: number;
  order_code: string;
  refund_type: string;
  refund_amount: number;
  items_refund: number;
  shipping_refund: number;
  reason: string;
  reason_detail?: string;
  status: string;
  gateway_refund_id?: string;
  gateway_response?: string;
  error_message?: string;
  requested_at: string;
  processed_at?: string;
  completed_at?: string;
  failed_at?: string;
  items?: RefundItem[];
  status_history?: RefundStatusHistory[];
}

interface OrderDetail {
  id: number;
  order_code: string;
  customer_name: string;
  customer_email: string;
  customer_phone: string;
  shipping_address?: string;
  subtotal: number;
  shipping_cost: number;
  total_amount: number;
  status: string;
  resi?: string;
  items: OrderItem[];
  shipment?: {
    id: number;
    order_id: number;
    provider_code: string;
    provider_name: string;
    service_code: string;
    service_name: string;
    cost: number;
    etd: string;
    status: string;
    tracking_number?: string;
    shipped_at?: string;
    delivered_at?: string;
    tracking_history?: TrackingEvent[];
  };
  payment?: {
    method: string;
    bank?: string;
    va_number?: string;
    paid_at?: string;
  };
  refund_status?: string;
  refund_amount?: number;
  refunded_at?: string;
  created_at: string;
  paid_at?: string;
  shipped_at?: string;
  delivered_at?: string;
}

// Format price
const formatPrice = (price: number) => {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
  }).format(price);
};

// Format date
const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString("id-ID", {
    day: "numeric",
    month: "long",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
};

const formatShortDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString("id-ID", {
    day: "numeric",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
};

// Order Timeline Component (Tokopedia/Shopee style)
const OrderTimeline = ({ status, createdAt, paidAt, shippedAt, deliveredAt, refundStatus }: {
  status: string;
  createdAt: string;
  paidAt?: string;
  shippedAt?: string;
  deliveredAt?: string;
  refundStatus?: string;
}) => {
  const currentStep = STATUS_TO_STEP[status] ?? 0;
  const isCancelled = currentStep === -1;
  const isRefunded = status === "REFUNDED" || refundStatus === "FULL" || refundStatus === "PARTIAL";

  const timestamps: Record<string, string | undefined> = {
    placed: createdAt,
    paid: paidAt,
    processing: paidAt, // Same as paid for now
    shipped: shippedAt,
    delivered: deliveredAt,
  };

  // Show refund status if order is refunded
  if (isRefunded) {
    return (
      <div className="bg-orange-50 border border-orange-200 rounded-xl p-6 text-center">
        <div className="w-16 h-16 bg-orange-100 rounded-full flex items-center justify-center mx-auto mb-4">
          <svg className="w-8 h-8 text-orange-600" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth={1.5}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
          </svg>
        </div>
        <h3 className="text-lg font-semibold text-orange-700">Pembayaran Dikembalikan</h3>
        <p className="text-orange-600 text-sm mt-1">
          {refundStatus === "FULL" 
            ? "Dana sudah dikembalikan ke metode pembayaran yang kamu pakai." 
            : "Dana sebagian sudah dikembalikan ke metode pembayaran yang kamu pakai."}
        </p>
      </div>
    );
  }

  if (isCancelled) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-xl p-6 text-center">
        <div className="w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4">
          <svg className="w-8 h-8 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth={1.5}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M9.75 9.75l4.5 4.5m0-4.5l-4.5 4.5M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
        <h3 className="text-lg font-semibold text-red-700">Pesanan Dibatalkan</h3>
        <p className="text-red-600 text-sm mt-1">
          {status === "EXPIRED" || status === "KADALUARSA" 
            ? "Pembayaran tidak diterima dalam batas waktu" 
            : "Pesanan ini telah dibatalkan"}
        </p>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-xl border p-6">
      <h3 className="font-semibold text-gray-900 mb-6">Status Pesanan</h3>
      <div className="relative">
        {/* Progress line */}
        <div className="absolute top-5 left-6 right-6 h-0.5 bg-gray-200">
          <div 
            className="h-full bg-primary transition-all duration-500"
            style={{ width: `${(currentStep / (ORDER_STEPS.length - 1)) * 100}%` }}
          />
        </div>
        
        {/* Steps */}
        <div className="relative flex justify-between">
          {ORDER_STEPS.map((step, index) => {
            const isCompleted = index <= currentStep;
            const isCurrent = index === currentStep;
            const timestamp = timestamps[step.key];
            const IconComponent = TimelineIcons[step.key as keyof typeof TimelineIcons];
            
            return (
              <div key={step.key} className="flex flex-col items-center" style={{ width: "20%" }}>
                <div 
                  className={`w-10 h-10 rounded-full flex items-center justify-center z-10 transition-all border-2 ${
                    isCompleted 
                      ? "bg-primary border-primary text-white" 
                      : "bg-white border-gray-200 text-gray-400"
                  } ${isCurrent ? "ring-4 ring-primary/20" : ""}`}
                >
                  {IconComponent}
                </div>
                <p className={`text-xs mt-2 text-center font-medium ${
                  isCompleted ? "text-gray-900" : "text-gray-400"
                }`}>
                  {step.label}
                </p>
                {timestamp && isCompleted && (
                  <p className="text-[10px] text-gray-400 mt-1 text-center">
                    {formatShortDate(timestamp)}
                  </p>
                )}
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
};


// Tracking History Component (Shopee style)
const TrackingHistory = ({ resi, shipment }: { resi?: string; shipment?: OrderDetail["shipment"] }) => {
  const { showToast } = useToast();
  
  // Mock tracking history if not available from API
  const trackingHistory: TrackingEvent[] = shipment?.tracking_history || [];

  const copyResi = () => {
    if (resi) {
      navigator.clipboard.writeText(resi);
      showToast("Nomor resi disalin!", "success");
    }
  };

  if (!resi && !shipment) {
    return null;
  }

  return (
    <div className="bg-white rounded-xl border overflow-hidden">
      <div className="p-4 border-b bg-gray-50">
        <div className="flex items-center justify-between">
          <h3 className="font-semibold text-gray-900 flex items-center gap-2">
            <svg className="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
            Info Pengiriman
          </h3>
        </div>
      </div>

      <div className="p-4">
        {/* Courier Info */}
        {shipment && (
          <div className="flex items-center gap-4 pb-4 border-b mb-4">
            <div className="w-14 h-14 bg-gray-100 rounded-lg flex items-center justify-center">
              <span className="text-sm font-bold text-gray-600">
                {shipment.provider_code?.toUpperCase()}
              </span>
            </div>
            <div className="flex-1">
              <p className="font-medium text-gray-900">
                {shipment.provider_name} - {shipment.service_name}
              </p>
              <p className="text-sm text-gray-500">Estimasi: {shipment.etd}</p>
            </div>
          </div>
        )}

        {/* Resi Number */}
        {resi && (
          <div className="bg-blue-50 rounded-lg p-4 mb-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-xs text-blue-600 font-medium mb-1">Nomor Resi</p>
                <p className="font-mono text-lg font-bold text-gray-900 tracking-wider">{resi}</p>
              </div>
              <button
                onClick={copyResi}
                className="p-2 hover:bg-blue-100 rounded-lg transition"
                title="Salin Resi"
              >
                <svg className="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                </svg>
              </button>
            </div>
            
            {/* Track Button */}
            <a
              href={`https://cekresi.com/?noresi=${resi}`}
              target="_blank"
              rel="noopener noreferrer"
              className="mt-3 flex items-center justify-center gap-2 w-full py-2.5 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
              Lacak Pengiriman
            </a>
          </div>
        )}

        {/* Tracking History Timeline (Shopee style) */}
        {trackingHistory.length > 0 && (
          <div className="mt-4">
            <h4 className="text-sm font-medium text-gray-700 mb-3">Riwayat Pengiriman</h4>
            <div className="space-y-0">
              {trackingHistory.map((event, index) => (
                <div key={index} className="flex gap-3">
                  {/* Timeline dot and line */}
                  <div className="flex flex-col items-center">
                    <div className={`w-3 h-3 rounded-full ${
                      index === 0 ? "bg-primary" : "bg-gray-300"
                    }`} />
                    {index < trackingHistory.length - 1 && (
                      <div className="w-0.5 h-full min-h-[40px] bg-gray-200" />
                    )}
                  </div>
                  
                  {/* Event content */}
                  <div className="pb-4 flex-1">
                    <p className={`text-sm font-medium ${
                      index === 0 ? "text-primary" : "text-gray-700"
                    }`}>
                      {event.status}
                    </p>
                    <p className="text-sm text-gray-600">{event.description}</p>
                    {event.location && (
                      <p className="text-xs text-gray-400">{event.location}</p>
                    )}
                    <p className="text-xs text-gray-400 mt-1">
                      {formatShortDate(event.event_time)}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* No tracking yet */}
        {!resi && shipment?.status === "PENDING" && (
          <div className="text-center py-4 text-sm text-gray-500">
            <p>⏳ Resi akan tersedia setelah pesanan dikirim</p>
          </div>
        )}
      </div>
    </div>
  );
};


// Refund Status Config
const REFUND_STATUS_CONFIG: Record<string, { bg: string; text: string; label: string; icon: JSX.Element }> = {
  PENDING: {
    bg: "bg-amber-100",
    text: "text-amber-700",
    label: "Menunggu Proses",
    icon: (
      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
    ),
  },
  PROCESSING: {
    bg: "bg-blue-100",
    text: "text-blue-700",
    label: "Sedang Diproses",
    icon: (
      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
      </svg>
    ),
  },
  COMPLETED: {
    bg: "bg-emerald-100",
    text: "text-emerald-700",
    label: "Selesai",
    icon: (
      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
    ),
  },
  FAILED: {
    bg: "bg-red-100",
    text: "text-red-700",
    label: "Gagal",
    icon: (
      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
    ),
  },
};

// Refund Type Labels
const REFUND_TYPE_LABELS: Record<string, string> = {
  FULL: "Pengembalian Penuh",
  PARTIAL: "Pengembalian Sebagian",
  SHIPPING_ONLY: "Pengembalian Ongkir",
  ITEM_ONLY: "Pengembalian Produk",
};

// Payment method timeline estimates (in days)
const REFUND_TIMELINE: Record<string, { min: number; max: number }> = {
  bank_transfer: { min: 1, max: 3 },
  bca: { min: 1, max: 3 },
  bni: { min: 1, max: 3 },
  bri: { min: 1, max: 3 },
  mandiri: { min: 1, max: 3 },
  permata: { min: 1, max: 3 },
  gopay: { min: 0, max: 1 },
  qris: { min: 0, max: 1 },
  credit_card: { min: 5, max: 14 },
};

// Refund Details Component
const RefundDetails = ({ refunds, paymentMethod }: { refunds: Refund[]; paymentMethod?: string }) => {
  if (!refunds || refunds.length === 0) return null;

  const getTimelineEstimate = (method?: string) => {
    const timeline = REFUND_TIMELINE[method?.toLowerCase() || "bank_transfer"] || { min: 1, max: 3 };
    if (timeline.min === 0 && timeline.max === 1) {
      return "dalam 24 jam";
    }
    if (timeline.min === timeline.max) {
      return `${timeline.min} hari kerja`;
    }
    return `${timeline.min}-${timeline.max} hari kerja`;
  };

  return (
    <div className="bg-white rounded-xl border overflow-hidden">
      <div className="p-4 border-b bg-orange-50">
        <div className="flex items-center gap-2">
          <svg className="w-5 h-5 text-orange-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
          </svg>
          <h3 className="font-semibold text-gray-900">Informasi Pengembalian Dana</h3>
        </div>
      </div>

      <div className="divide-y">
        {refunds.map((refund) => {
          const statusConfig = REFUND_STATUS_CONFIG[refund.status] || REFUND_STATUS_CONFIG.PENDING;
          const isCompleted = refund.status === "COMPLETED";
          const isFailed = refund.status === "FAILED";
          const isProcessing = refund.status === "PROCESSING";

          return (
            <div key={refund.id} className="p-4">
              {/* Refund Header */}
              <div className="flex items-start justify-between mb-3">
                <div>
                  <p className="text-sm text-gray-500 mb-1">Kode Refund</p>
                  <p className="font-mono text-sm font-medium text-gray-900">{refund.refund_code}</p>
                </div>
                <div className={`flex items-center gap-1.5 px-3 py-1.5 rounded-full ${statusConfig.bg} ${statusConfig.text}`}>
                  {statusConfig.icon}
                  <span className="text-sm font-medium">{statusConfig.label}</span>
                </div>
              </div>

              {/* Refund Amount */}
              <div className="bg-gray-50 rounded-lg p-4 mb-3">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm text-gray-600">Jumlah Pengembalian</span>
                  <span className="text-xl font-bold text-gray-900">{formatPrice(refund.refund_amount)}</span>
                </div>
                <div className="flex items-center justify-between text-sm">
                  <span className="text-gray-500">Tipe</span>
                  <span className="text-gray-700 font-medium">{REFUND_TYPE_LABELS[refund.refund_type] || refund.refund_type}</span>
                </div>
                {refund.items_refund > 0 && (
                  <div className="flex items-center justify-between text-sm mt-1">
                    <span className="text-gray-500">Produk</span>
                    <span className="text-gray-700">{formatPrice(refund.items_refund)}</span>
                  </div>
                )}
                {refund.shipping_refund > 0 && (
                  <div className="flex items-center justify-between text-sm mt-1">
                    <span className="text-gray-500">Ongkir</span>
                    <span className="text-gray-700">{formatPrice(refund.shipping_refund)}</span>
                  </div>
                )}
              </div>

              {/* Status Messages */}
              {isCompleted && (
                <div className="bg-emerald-50 border border-emerald-200 rounded-lg p-3 mb-3">
                  <div className="flex items-start gap-2">
                    <svg className="w-5 h-5 text-emerald-600 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                    <div className="flex-1">
                      <p className="text-sm font-medium text-emerald-800">Pengembalian Dana Berhasil</p>
                      <p className="text-sm text-emerald-700 mt-1">
                        Dana telah dikembalikan ke metode pembayaran kamu pada{" "}
                        {refund.completed_at && formatDate(refund.completed_at)}
                      </p>
                      {refund.gateway_refund_id && refund.gateway_refund_id !== "MANUAL_REFUND" && (
                        <p className="text-xs text-emerald-600 mt-1 font-mono">ID: {refund.gateway_refund_id}</p>
                      )}
                    </div>
                  </div>
                </div>
              )}

              {isProcessing && (
                <div className="bg-blue-50 border border-blue-200 rounded-lg p-3 mb-3">
                  <div className="flex items-start gap-2">
                    <svg className="w-5 h-5 text-blue-600 flex-shrink-0 mt-0.5 animate-spin" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                    </svg>
                    <div className="flex-1">
                      <p className="text-sm font-medium text-blue-800">Sedang Diproses</p>
                      <p className="text-sm text-blue-700 mt-1">
                        Pengembalian dana sedang diproses. Dana akan masuk ke rekening kamu {getTimelineEstimate(paymentMethod)}.
                      </p>
                    </div>
                  </div>
                </div>
              )}

              {isFailed && (
                <div className="bg-red-50 border border-red-200 rounded-lg p-3 mb-3">
                  <div className="flex items-start gap-2">
                    <svg className="w-5 h-5 text-red-600 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                    <div className="flex-1">
                      <p className="text-sm font-medium text-red-800">Pengembalian Gagal</p>
                      <p className="text-sm text-red-700 mt-1">
                        {refund.error_message || "Terjadi kesalahan saat memproses pengembalian dana. Tim kami akan segera menghubungi kamu."}
                      </p>
                    </div>
                  </div>
                </div>
              )}

              {/* Refund Reason */}
              <div className="mb-3">
                <p className="text-sm text-gray-500 mb-1">Alasan Pengembalian</p>
                <p className="text-sm text-gray-900">{refund.reason}</p>
                {refund.reason_detail && (
                  <p className="text-sm text-gray-600 mt-1">{refund.reason_detail}</p>
                )}
              </div>

              {/* Refunded Items (for ITEM_ONLY refunds) */}
              {refund.items && refund.items.length > 0 && (
                <div className="mb-3">
                  <p className="text-sm text-gray-500 mb-2">Produk yang Dikembalikan</p>
                  <div className="space-y-2">
                    {refund.items.map((item, idx) => (
                      <div key={idx} className="flex items-center justify-between text-sm bg-gray-50 rounded p-2">
                        <span className="text-gray-700">{item.product_name}</span>
                        <span className="text-gray-600">
                          {item.quantity} x {formatPrice(item.price_per_unit)}
                        </span>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Timeline */}
              <div className="text-xs text-gray-500 space-y-1">
                <div className="flex items-center justify-between">
                  <span>Dibuat</span>
                  <span>{formatShortDate(refund.requested_at)}</span>
                </div>
                {refund.processed_at && (
                  <div className="flex items-center justify-between">
                    <span>Diproses</span>
                    <span>{formatShortDate(refund.processed_at)}</span>
                  </div>
                )}
                {refund.completed_at && (
                  <div className="flex items-center justify-between">
                    <span>Selesai</span>
                    <span>{formatShortDate(refund.completed_at)}</span>
                  </div>
                )}
                {refund.failed_at && (
                  <div className="flex items-center justify-between">
                    <span>Gagal</span>
                    <span>{formatShortDate(refund.failed_at)}</span>
                  </div>
                )}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};


// Main Order Detail Page
export default function OrderDetailPage() {
  const params = useParams();
  const router = useRouter();
  const orderCode = params.code as string;
  const { isAuthenticated, isLoading: authLoading } = useAuth();
  const { showToast } = useToast();

  const [order, setOrder] = useState<OrderDetail | null>(null);
  const [refunds, setRefunds] = useState<Refund[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchOrder = useCallback(async () => {
    if (!orderCode) return;
    
    try {
      setLoading(true);
      const token = localStorage.getItem("auth_token");
      const headers = token ? { Authorization: `Bearer ${token}` } : {};
      
      // Fetch order details
      const orderRes = await api.get(`/orders/${orderCode}`, { headers });
      setOrder(orderRes.data);
      
      // Fetch refunds if order has refund status
      if (orderRes.data.refund_status || orderRes.data.status === "REFUNDED") {
        try {
          const refundsRes = await api.get(`/customer/orders/${orderCode}/refunds`, { headers });
          setRefunds(refundsRes.data.refunds || []);
        } catch (refundErr) {
          console.error("Error fetching refunds:", refundErr);
          // Don't fail the whole page if refunds fail to load
        }
      }
      
      setError(null);
    } catch (err: unknown) {
      console.error("Error fetching order:", err);
      const axiosErr = err as { response?: { status?: number } };
      if (axiosErr.response?.status === 403) {
        setError("Anda tidak memiliki akses ke pesanan ini");
      } else if (axiosErr.response?.status === 404) {
        setError("Pesanan tidak ditemukan");
      } else {
        setError("Gagal memuat data pesanan");
      }
    } finally {
      setLoading(false);
    }
  }, [orderCode]);

  useEffect(() => {
    if (!authLoading && !isAuthenticated) {
      router.push(`/login?redirect=/orders/${orderCode}`);
      return;
    }
    if (isAuthenticated) {
      fetchOrder();
    }
  }, [authLoading, isAuthenticated, orderCode, router, fetchOrder]);

  if (authLoading || loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <LoadingSpinner />
      </div>
    );
  }

  if (error || !order) {
    return (
      <div className="min-h-screen bg-gray-50 py-8">
        <div className="max-w-3xl mx-auto px-4">
          <div className="bg-white rounded-xl shadow-sm p-12 text-center">
            <div className="w-20 h-20 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-6">
              <svg className="w-10 h-10 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
            </div>
            <h2 className="text-xl font-semibold text-gray-900 mb-2">{error || "Pesanan tidak ditemukan"}</h2>
            <p className="text-gray-600 mb-6">Kode pesanan: {orderCode}</p>
            <div className="flex gap-3 justify-center">
              <Link
                href="/account/pembelian?tab=history"
                className="px-6 py-2.5 bg-primary text-white rounded-lg font-medium hover:bg-primary/90 transition"
              >
                Kembali ke Daftar Transaksi
              </Link>
              <button
                onClick={fetchOrder}
                className="px-6 py-2.5 border border-gray-300 rounded-lg font-medium hover:bg-gray-50 transition"
              >
                Coba Lagi
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white border-b sticky top-0 z-40">
        <div className="max-w-3xl mx-auto px-4 py-4 flex items-center gap-3">
          <button 
            onClick={() => router.back()}
            className="p-2 hover:bg-gray-100 rounded-lg transition"
          >
            <svg className="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M15 19l-7-7 7-7" />
            </svg>
          </button>
          <div>
            <h1 className="text-lg font-semibold text-gray-900">Detail Pesanan</h1>
            <p className="text-sm text-gray-500 font-mono">{order.order_code}</p>
          </div>
        </div>
      </div>

      <div className="max-w-3xl mx-auto px-4 py-6 space-y-4">
        {/* Order Timeline */}
        <motion.div
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
        >
          <OrderTimeline 
            status={order.status}
            createdAt={order.created_at}
            paidAt={order.paid_at}
            shippedAt={order.shipped_at}
            deliveredAt={order.delivered_at}
            refundStatus={order.refund_status}
          />
        </motion.div>

        {/* Tracking Info - Only show for shipped/delivered orders that are NOT refunded */}
        {(order.resi || order.shipment) && order.status !== "REFUNDED" && !order.refund_status && (
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.1 }}
          >
            <TrackingHistory resi={order.resi} shipment={order.shipment} />
          </motion.div>
        )}

        {/* Refund Information - Show if order has refunds */}
        {refunds.length > 0 && (
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.15 }}
          >
            <RefundDetails refunds={refunds} paymentMethod={order.payment?.bank || order.payment?.method} />
          </motion.div>
        )}

        {/* Order Items */}
        <motion.div
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2 }}
          className="bg-white rounded-xl border overflow-hidden"
        >
          <div className="p-4 border-b bg-gray-50">
            <h3 className="font-semibold text-gray-900">Detail Produk</h3>
          </div>
          <div className="divide-y">
            {order.items?.map((item, idx) => (
              <div key={idx} className="p-4 flex gap-4">
                <div className="w-20 h-20 bg-gray-100 rounded-lg overflow-hidden flex-shrink-0 relative">
                  {item.product_image ? (
                    <Image
                      src={item.product_image}
                      alt={item.product_name}
                      fill
                      className="object-cover"
                    />
                  ) : (
                    <div className="w-full h-full flex items-center justify-center">
                      <svg className="w-8 h-8 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                      </svg>
                    </div>
                  )}
                </div>
                <div className="flex-1 min-w-0">
                  <h4 className="font-medium text-gray-900">{item.product_name}</h4>
                  <p className="text-sm text-gray-500">
                    {item.quantity} x {formatPrice(item.price_per_unit)}
                    {item.size && ` • Ukuran: ${item.size}`}
                  </p>
                </div>
                <div className="text-right flex-shrink-0">
                  <p className="font-semibold text-gray-900">{formatPrice(item.subtotal)}</p>
                </div>
              </div>
            ))}
          </div>
        </motion.div>

        {/* Payment Summary */}
        <motion.div
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.3 }}
          className="bg-white rounded-xl border overflow-hidden"
        >
          <div className="p-4 border-b bg-gray-50">
            <h3 className="font-semibold text-gray-900">Ringkasan Pembayaran</h3>
          </div>
          <div className="p-4 space-y-3">
            {order.payment && (
              <div className="flex justify-between text-sm pb-3 border-b">
                <span className="text-gray-600">Metode Pembayaran</span>
                <span className="font-medium text-gray-900">
                  {order.payment.bank?.toUpperCase() || order.payment.method}
                </span>
              </div>
            )}
            <div className="flex justify-between text-sm">
              <span className="text-gray-600">Subtotal Produk</span>
              <span className="text-gray-900">{formatPrice(order.subtotal)}</span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-gray-600">Ongkos Kirim</span>
              <span className="text-gray-900">{formatPrice(order.shipping_cost)}</span>
            </div>
            <div className="flex justify-between font-semibold pt-3 border-t">
              <span>Total Pembayaran</span>
              <span className="text-primary text-lg">{formatPrice(order.total_amount)}</span>
            </div>
          </div>
        </motion.div>

        {/* Shipping Address */}
        {order.shipping_address && (
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.4 }}
            className="bg-white rounded-xl border overflow-hidden"
          >
            <div className="p-4 border-b bg-gray-50">
              <h3 className="font-semibold text-gray-900">Alamat Pengiriman</h3>
            </div>
            <div className="p-4">
              <p className="font-medium text-gray-900">{order.customer_name}</p>
              <p className="text-sm text-gray-600 mt-1">{order.customer_phone}</p>
              <p className="text-sm text-gray-600 mt-2">{order.shipping_address}</p>
            </div>
          </motion.div>
        )}

        {/* Order Info */}
        <motion.div
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.5 }}
          className="bg-white rounded-xl border overflow-hidden"
        >
          <div className="p-4 border-b bg-gray-50">
            <h3 className="font-semibold text-gray-900">Info Pesanan</h3>
          </div>
          <div className="p-4 space-y-2 text-sm">
            <div className="flex justify-between">
              <span className="text-gray-600">Nomor Pesanan</span>
              <span className="font-mono text-gray-900">{order.order_code}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-gray-600">Tanggal Pemesanan</span>
              <span className="text-gray-900">{formatDate(order.created_at)}</span>
            </div>
            {order.paid_at && (
              <div className="flex justify-between">
                <span className="text-gray-600">Tanggal Pembayaran</span>
                <span className="text-gray-900">{formatDate(order.paid_at)}</span>
              </div>
            )}
          </div>
        </motion.div>

        {/* Actions */}
        <motion.div
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.6 }}
          className="flex gap-3"
        >
          <Link
            href="/account/pembelian?tab=history"
            className="flex-1 py-3 text-center border border-gray-300 rounded-lg font-medium text-gray-700 hover:bg-gray-50 transition"
          >
            Kembali
          </Link>
          <Link
            href="/"
            className="flex-1 py-3 text-center bg-primary text-white rounded-lg font-medium hover:bg-primary/90 transition"
          >
            Belanja Lagi
          </Link>
        </motion.div>
      </div>
    </div>
  );
}
