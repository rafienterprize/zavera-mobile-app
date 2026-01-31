"use client";

import { useEffect, useState, Suspense, useCallback } from "react";
import { useSearchParams } from "next/navigation";
import Link from "next/link";
import { motion } from "framer-motion";
import api from "@/lib/api";
import { useDialog } from "@/context/DialogContext";

interface OrderDetails {
  id: number;
  order_code: string;
  customer_name: string;
  customer_email: string;
  total_amount: number;
  status: string;
  resi?: string;
  items: Array<{
    product_name: string;
    quantity: number;
    price_per_unit: number;
    subtotal: number;
  }>;
  shipment?: {
    provider_code: string;
    provider_name: string;
    service_name: string;
    etd: string;
    status: string;
  };
  created_at: string;
}

function OrderSuccessContent() {
  const dialog = useDialog();
  const searchParams = useSearchParams();
  const orderCode = searchParams.get("code");
  const [order, setOrder] = useState<OrderDetails | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [remainingTime, setRemainingTime] = useState<{ hours: number; minutes: number; seconds: number } | null>(null);
  const [isPaying, setIsPaying] = useState(false);

  const fetchOrder = useCallback(async () => {
    if (orderCode) {
      try {
        const token = localStorage.getItem("auth_token");
        const headers = token ? { Authorization: `Bearer ${token}` } : {};
        const res = await api.get(`/orders/${orderCode}`, { headers });
        console.log("Order data fetched:", res.data);
        console.log("Order status:", res.data.status);
        setOrder(res.data);
        setError(null);
      } catch (err: unknown) {
        console.error("Error fetching order:", err);
        const axiosErr = err as { response?: { status?: number } };
        if (axiosErr.response?.status === 403) {
          setError("Anda tidak memiliki akses ke pesanan ini");
        } else {
          setError("Gagal memuat data pesanan");
        }
      } finally {
        setLoading(false);
      }
    } else {
      setLoading(false);
    }
  }, [orderCode]);

  useEffect(() => {
    fetchOrder();
  }, [fetchOrder]);

  // Countdown timer for PENDING orders
  useEffect(() => {
    if (!order || (order.status !== "PENDING" && order.status !== "MENUNGGU_PEMBAYARAN")) return;

    const calculateRemaining = () => {
      const created = new Date(order.created_at).getTime();
      const expiry = created + 24 * 60 * 60 * 1000; // 24 hours
      const now = Date.now();
      const remaining = expiry - now;

      if (remaining <= 0) {
        setRemainingTime(null);
        return;
      }

      const hours = Math.floor(remaining / (60 * 60 * 1000));
      const minutes = Math.floor((remaining % (60 * 60 * 1000)) / (60 * 1000));
      const seconds = Math.floor((remaining % (60 * 1000)) / 1000);

      setRemainingTime({ hours, minutes, seconds });
    };

    calculateRemaining();
    const interval = setInterval(calculateRemaining, 1000);

    return () => clearInterval(interval);
  }, [order]);

  // Handle pay now for PENDING orders
  const handlePayNow = async () => {
    if (!order) return;
    
    try {
      setIsPaying(true);
      const token = localStorage.getItem("auth_token");
      
      const response = await api.post<{ snap_token: string }>(
        "/payments/initiate",
        { order_id: order.id },
        { headers: { Authorization: `Bearer ${token}` } }
      );
      
      const snapToken = response.data.snap_token;
      
      if (window.snap) {
        window.snap.pay(snapToken, {
          onSuccess: () => {
            fetchOrder(); // Refresh order data
          },
          onPending: () => {
            // Still pending
          },
          onError: async () => {
            await dialog.alert({
              title: 'Pembayaran Gagal',
              message: 'Terjadi kesalahan saat memproses pembayaran. Silakan coba lagi.',
              variant: 'error'
            });
          },
          onClose: () => {
            setIsPaying(false);
          },
        });
      }
    } catch (error) {
      console.error(error);
      await dialog.alert({
        title: 'Error',
        message: 'Gagal memproses pembayaran. Silakan coba lagi.',
        variant: 'error'
      });
    } finally {
      setIsPaying(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-pulse text-center">
          <div className="w-16 h-16 bg-gray-200 rounded-full mx-auto mb-4" />
          <div className="h-6 bg-gray-200 rounded w-48 mx-auto" />
        </div>
      </div>
    );
  }

  // Show error state if failed to load order
  if (error || !order) {
    return (
      <div className="min-h-screen bg-gradient-to-b from-red-50 to-white py-16 px-4">
        <div className="max-w-2xl mx-auto text-center">
          <div className="w-24 h-24 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-6">
            <svg className="w-12 h-12 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
          </div>
          <h1 className="text-2xl font-bold text-gray-900 mb-2">{error || "Pesanan tidak ditemukan"}</h1>
          <p className="text-gray-600 mb-6">Kode pesanan: {orderCode}</p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Link
              href="/account/pembelian?tab=history"
              className="px-8 py-3 bg-primary text-white text-center font-medium hover:bg-gray-800 transition-colors rounded-lg"
            >
              Lihat Daftar Transaksi
            </Link>
            <button
              onClick={() => fetchOrder()}
              className="px-8 py-3 border-2 border-primary text-primary text-center font-medium hover:bg-primary hover:text-white transition-colors rounded-lg"
            >
              Coba Lagi
            </button>
          </div>
        </div>
      </div>
    );
  }

  const isPending = order.status === "PENDING" || order.status === "MENUNGGU_PEMBAYARAN";
  const isPaid = order.status === "PAID" || order.status === "DIBAYAR" || order.status === "SHIPPED" || order.status === "DIKIRIM" || order.status === "DELIVERED" || order.status === "SELESAI";

  return (
    <div className={`min-h-screen py-16 px-4 ${isPending ? "bg-gradient-to-b from-amber-50 to-white" : "bg-gradient-to-b from-green-50 to-white"}`}>
      <div className="max-w-2xl mx-auto">
        {/* Status Icon & Title */}
        <motion.div
          initial={{ scale: 0 }}
          animate={{ scale: 1 }}
          transition={{ type: "spring", damping: 15, stiffness: 200 }}
          className="text-center mb-8"
        >
          {isPending ? (
            <>
              {/* Pending Icon - Yellow Clock */}
              <div className="w-24 h-24 bg-amber-100 rounded-full flex items-center justify-center mx-auto mb-6">
                <motion.svg
                  initial={{ rotate: -90 }}
                  animate={{ rotate: 0 }}
                  transition={{ duration: 0.5, delay: 0.2 }}
                  className="w-12 h-12 text-amber-600"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                  />
                </motion.svg>
              </div>
              <h1 className="text-3xl font-bold text-gray-900 mb-2">Menunggu Pembayaran</h1>
              <p className="text-gray-600">Segera selesaikan pembayaran Anda</p>
              
              {/* Countdown Timer */}
              {remainingTime ? (
                <div className="mt-4 inline-flex items-center gap-2 px-4 py-2 bg-amber-100 rounded-full">
                  <svg className="w-5 h-5 text-amber-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  <span className="text-amber-800 font-medium">
                    Sisa waktu: {remainingTime.hours}j {remainingTime.minutes}m {remainingTime.seconds}d
                  </span>
                </div>
              ) : (
                <div className="mt-4 inline-flex items-center gap-2 px-4 py-2 bg-red-100 rounded-full">
                  <span className="text-red-700 font-medium">Waktu pembayaran habis</span>
                </div>
              )}
            </>
          ) : (
            <>
              {/* Success Icon - Green Check */}
              <div className="w-24 h-24 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-6">
                <motion.svg
                  initial={{ pathLength: 0 }}
                  animate={{ pathLength: 1 }}
                  transition={{ duration: 0.5, delay: 0.2 }}
                  className="w-12 h-12 text-green-600"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <motion.path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={3}
                    d="M5 13l4 4L19 7"
                  />
                </motion.svg>
              </div>
              <h1 className="text-3xl font-bold text-gray-900 mb-2">Pembayaran Berhasil!</h1>
              <p className="text-gray-600">Terima kasih atas pembelian Anda</p>
            </>
          )}
        </motion.div>

        {/* Order Details Card */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.3 }}
          className="bg-white rounded-2xl shadow-lg border border-gray-100 overflow-hidden"
        >
          {/* Order Header */}
          <div className="bg-gray-50 px-6 py-4 border-b">
            <div className="flex justify-between items-center">
              <div>
                <p className="text-sm text-gray-500">Nomor Pesanan</p>
                <p className="font-mono font-bold text-lg">{orderCode}</p>
              </div>
              <span className={`px-3 py-1 text-sm font-medium rounded-full ${
                isPending 
                  ? "bg-amber-100 text-amber-700" 
                  : "bg-green-100 text-green-700"
              }`}>
                {order?.status === "PENDING" || order?.status === "MENUNGGU_PEMBAYARAN" ? "Menunggu Pembayaran" : 
                 order?.status === "PAID" || order?.status === "DIBAYAR" ? "Sudah Dibayar" :
                 order?.status === "SHIPPED" || order?.status === "DIKIRIM" ? "Dikirim" :
                 order?.status === "DELIVERED" || order?.status === "SELESAI" ? "Selesai" : order?.status}
              </span>
            </div>
          </div>

          {/* Pay Now Button for PENDING */}
          {isPending && remainingTime && (
            <div className="px-6 py-4 bg-amber-50 border-b">
              <button
                onClick={handlePayNow}
                disabled={isPaying}
                className="w-full py-3 px-4 bg-primary text-white rounded-lg font-medium hover:bg-primary/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
              >
                {isPaying ? (
                  <>
                    <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
                    Memproses...
                  </>
                ) : (
                  <>
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" />
                    </svg>
                    Bayar Sekarang
                  </>
                )}
              </button>
            </div>
          )}

          {/* Order Items */}
          {order && (
            <div className="px-6 py-4 border-b">
              <h3 className="font-semibold mb-3">Detail Pesanan</h3>
              <div className="space-y-3">
                {order.items?.map((item, idx) => (
                  <div key={idx} className="flex justify-between text-sm">
                    <span className="text-gray-600">
                      {item.product_name} × {item.quantity}
                    </span>
                    <span className="font-medium">
                      Rp {item.subtotal.toLocaleString("id-ID")}
                    </span>
                  </div>
                ))}
              </div>
              <div className="border-t mt-4 pt-4 flex justify-between font-bold">
                <span>Total</span>
                <span>Rp {order.total_amount.toLocaleString("id-ID")}</span>
              </div>
            </div>
          )}

          {/* Customer Info */}
          {order && (
            <div className="px-6 py-4 border-b bg-gray-50">
              <h3 className="font-semibold mb-2">Informasi Pengiriman</h3>
              <p className="text-sm text-gray-600">{order.customer_name}</p>
              <p className="text-sm text-gray-600">{order.customer_email}</p>
            </div>
          )}

          {/* Shipping & Resi Info - Only show for PAID orders */}
          {order && isPaid && (order.resi || order.shipment) && (
            <div className="px-6 py-4 border-b">
              <h3 className="font-semibold mb-3 flex items-center gap-2">
                <svg className="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" />
                </svg>
                Info Pengiriman
              </h3>
              
              {/* Courier Info */}
              {order.shipment && (
                <div className="flex items-center gap-3 mb-3 pb-3 border-b border-gray-100">
                  <div className="w-12 h-12 bg-gray-100 rounded-lg flex items-center justify-center">
                    <span className="text-xs font-bold text-gray-600">
                      {order.shipment.provider_code?.toUpperCase()}
                    </span>
                  </div>
                  <div>
                    <p className="font-medium text-gray-900">
                      {order.shipment.provider_name} - {order.shipment.service_name}
                    </p>
                    <p className="text-sm text-gray-500">
                      Estimasi: {order.shipment.etd}
                    </p>
                  </div>
                </div>
              )}
              
              {/* Resi Number */}
              {order.resi && (
                <div className="bg-blue-50 rounded-lg p-4">
                  <div className="flex items-center justify-between mb-2">
                    <div>
                      <p className="text-xs text-blue-600 font-medium mb-1">Nomor Resi</p>
                      <p className="font-mono text-xl font-bold text-gray-900 tracking-wider">
                        {order.resi}
                      </p>
                    </div>
                    <button
                      onClick={async () => {
                        navigator.clipboard.writeText(order.resi || "");
                        await dialog.alert({
                          title: 'Berhasil!',
                          message: 'Nomor resi berhasil disalin ke clipboard',
                          variant: 'success',
                          buttonText: 'OK'
                        });
                      }}
                      className="p-2 hover:bg-blue-100 rounded-lg transition-colors"
                      title="Salin Resi"
                    >
                      <svg className="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                      </svg>
                    </button>
                  </div>
                  
                  {/* Track Button */}
                  <a
                    href={`https://cekresi.com/?noresi=${order.resi}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="mt-3 flex items-center justify-center gap-2 w-full py-2.5 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
                  >
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                    </svg>
                    Lacak Pengiriman
                  </a>
                </div>
              )}
              
              {/* No Resi Yet */}
              {!order.resi && (order.status === "PAID" || order.status === "DIBAYAR") && (
                <div className="text-center py-3 text-sm text-gray-500 bg-yellow-50 rounded-lg">
                  <p>⏳ Resi akan tersedia setelah pesanan diproses</p>
                </div>
              )}
            </div>
          )}

          {/* Security Badge */}
          <div className="px-6 py-4 flex items-center gap-3 text-sm text-gray-500">
            <svg className="w-5 h-5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
            </svg>
            <span>Pembayaran aman dengan Midtrans</span>
          </div>
        </motion.div>

        {/* Action Buttons */}
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.5 }}
          className="mt-8 flex flex-col sm:flex-row gap-4 justify-center"
        >
          <Link
            href="/"
            className="px-8 py-3 bg-primary text-white text-center font-medium hover:bg-gray-800 transition-colors"
          >
            Lanjut Belanja
          </Link>
          <Link
            href="/account/pembelian?tab=history"
            className="px-8 py-3 border-2 border-primary text-primary text-center font-medium hover:bg-primary hover:text-white transition-colors"
          >
            Daftar Transaksi
          </Link>
        </motion.div>

        {/* Email Notice - Only for PAID */}
        {isPaid && (
          <motion.p
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.6 }}
            className="text-center text-sm text-gray-500 mt-6"
          >
            Email konfirmasi telah dikirim ke {order?.customer_email || "email Anda"}
          </motion.p>
        )}
      </div>
    </div>
  );
}

export default function OrderSuccessPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-pulse">Loading...</div>
      </div>
    }>
      <OrderSuccessContent />
    </Suspense>
  );
}
