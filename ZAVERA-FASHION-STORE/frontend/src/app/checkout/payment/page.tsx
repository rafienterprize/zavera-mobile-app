"use client";

import { useState, useEffect, Suspense } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import Image from "next/image";
import { motion, AnimatePresence } from "framer-motion";
import { useAuth } from "@/context/AuthContext";
import { useToast } from "@/components/ui/Toast";
import api from "@/lib/api";
import { LoadingOverlay } from "@/components/ui/LoadingSpinner";

// Payment method categories
const PAYMENT_METHODS = {
  eWallet: [
    {
      id: "gopay",
      name: "GoPay",
      logo: "/images/payments/gopay.png",
      description: "Bayar langsung dari aplikasi GoPay",
      category: "e-wallet",
    },
    {
      id: "qris",
      name: "QRIS",
      logo: "/images/payments/qris.png",
      description: "Scan QR untuk bayar dengan GoPay, OVO, Dana, LinkAja, dll",
      category: "e-wallet",
    },
  ],
  virtualAccount: [
    {
      id: "bca_va",
      name: "BCA Virtual Account",
      bank: "bca",
      logo: "/images/banks/bca.png",
      description: "Transfer via ATM, Mobile Banking, atau Internet Banking BCA",
      category: "va",
    },
    {
      id: "bri_va",
      name: "BRI Virtual Account",
      bank: "bri",
      logo: "/images/banks/bri.png",
      description: "Transfer via ATM, Mobile Banking, atau Internet Banking BRI",
      category: "va",
    },
    {
      id: "mandiri_va",
      name: "Mandiri Virtual Account",
      bank: "mandiri",
      logo: "/images/banks/mandiri.png",
      description: "Transfer via ATM, Mobile Banking, atau Internet Banking Mandiri",
      category: "va",
    },
    {
      id: "permata_va",
      name: "Permata Virtual Account",
      bank: "permata",
      logo: "/images/banks/permata.png",
      description: "Transfer via ATM, Mobile Banking, atau Internet Banking Permata",
      category: "va",
    },
    {
      id: "bni_va",
      name: "BNI Virtual Account",
      bank: "bni",
      logo: "/images/banks/bni.png",
      description: "Transfer via ATM, Mobile Banking, atau Internet Banking BNI",
      category: "va",
    },
  ],
  creditCard: [
    {
      id: "credit_card",
      name: "Kartu Kredit / Debit",
      logo: "/images/payments/credit-card.svg",
      description: "Visa, Mastercard, JCB, American Express",
      category: "cc",
    },
  ],
};

interface OrderSummary {
  order_id: number;
  order_code: string;
  total_amount: number;
  item_count: number;
  items: Array<{
    product_name: string;
    quantity: number;
    price: number;
    image_url?: string;
  }>;
}

function PaymentSelectionContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const orderId = searchParams.get("order_id");
  const { isAuthenticated } = useAuth();
  const { showToast } = useToast();

  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [selectedMethod, setSelectedMethod] = useState<string | null>(null);
  const [order, setOrder] = useState<OrderSummary | null>(null);
  const [existingPayment, setExistingPayment] = useState<boolean>(false);

  // Load order details on mount
  useEffect(() => {
    if (!orderId) {
      showToast("Order ID tidak ditemukan", "error");
      router.push("/");
      return;
    }

    loadOrderDetails();
  }, [orderId]);

  const loadOrderDetails = async () => {
    try {
      setLoading(true);
      
      // Check if payment already exists
      try {
        const paymentRes = await api.get(`/payments/core/${orderId}`);
        if (paymentRes.data && paymentRes.data.status === "PENDING") {
          // Payment exists, redirect to VA detail page
          setExistingPayment(true);
          router.push(`/checkout/payment/detail?order_id=${orderId}`);
          return;
        }
      } catch {
        // No existing payment, continue to show selection
      }

      // Get order details by ID
      const orderRes = await api.get(`/orders/id/${orderId}`);
      setOrder({
        order_id: orderRes.data.id,
        order_code: orderRes.data.order_code,
        total_amount: orderRes.data.total_amount,
        item_count: orderRes.data.items?.length || 0,
        items: orderRes.data.items || [],
      });
    } catch (err) {
      console.error("Failed to load order:", err);
      showToast("Gagal memuat detail pesanan", "error");
    } finally {
      setLoading(false);
    }
  };

  const handlePayment = async () => {
    if (!selectedMethod || !orderId) {
      showToast("Pilih metode pembayaran", "error");
      return;
    }

    try {
      setSubmitting(true);
      
      const response = await api.post("/payments/core/create", {
        order_id: parseInt(orderId),
        payment_method: selectedMethod,
      });

      if (response.data) {
        showToast("Pembayaran berhasil dibuat", "success");
        router.push(`/checkout/payment/detail?order_id=${orderId}`);
      }
    } catch (err: unknown) {
      const axiosError = err as { response?: { data?: { message?: string; error?: string } } };
      const errorMessage = axiosError.response?.data?.message || "Gagal membuat pembayaran";
      showToast(errorMessage, "error");
    } finally {
      setSubmitting(false);
    }
  };

  if (loading || existingPayment) {
    return <LoadingOverlay message="Memuat..." />;
  }

  if (!order) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <h1 className="text-xl font-semibold text-gray-900 mb-2">Pesanan Tidak Ditemukan</h1>
          <Link href="/" className="text-primary hover:underline">Kembali ke Beranda</Link>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white border-b sticky top-0 z-40">
        <div className="max-w-3xl mx-auto px-4 py-4 flex items-center gap-4">
          <button onClick={() => router.back()} className="p-2 hover:bg-gray-100 rounded-full">
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
          </button>
          <h1 className="text-lg font-semibold">Pilih Metode Pembayaran</h1>
        </div>
      </div>

      <AnimatePresence>
        {submitting && <LoadingOverlay message="Membuat pembayaran..." />}
      </AnimatePresence>

      <div className="max-w-3xl mx-auto px-4 py-6">
        {/* Order Summary */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-5 mb-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="font-semibold text-gray-900">Ringkasan Pesanan</h2>
            <span className="text-sm text-gray-500">{order.order_code}</span>
          </div>
          
          <div className="space-y-3">
            {order.items.slice(0, 2).map((item, idx) => (
              <div key={idx} className="flex items-center gap-3">
                <div className="w-12 h-12 bg-gray-100 rounded-lg overflow-hidden relative flex-shrink-0">
                  {item.image_url && (
                    <Image src={item.image_url} alt={item.product_name} fill className="object-cover" />
                  )}
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm text-gray-900 truncate">{item.product_name}</p>
                  <p className="text-xs text-gray-500">x{item.quantity}</p>
                </div>
              </div>
            ))}
            {order.items.length > 2 && (
              <p className="text-sm text-gray-500">+{order.items.length - 2} produk lainnya</p>
            )}
          </div>

          <div className="border-t border-gray-100 mt-4 pt-4 flex justify-between items-center">
            <span className="text-gray-600">Total Pembayaran</span>
            <span className="text-xl font-bold text-primary">
              Rp {order.total_amount.toLocaleString("id-ID")}
            </span>
          </div>
        </div>

        {/* Payment Methods */}
        <div className="space-y-4">
          {/* E-Wallet / QRIS */}
          <div className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
            <div className="p-5 border-b border-gray-100">
              <h2 className="font-semibold text-gray-900">E-Wallet / QRIS</h2>
              <p className="text-sm text-gray-500 mt-1">Bayar dengan scan QR code</p>
            </div>
            <div className="divide-y divide-gray-100">
              {PAYMENT_METHODS.eWallet.map((method) => (
                <motion.button
                  key={method.id}
                  onClick={() => setSelectedMethod(method.id)}
                  className={`w-full p-5 flex items-center gap-4 text-left transition ${
                    selectedMethod === method.id ? "bg-primary/5" : "hover:bg-gray-50"
                  }`}
                  whileTap={{ scale: 0.99 }}
                >
                  <div className={`w-5 h-5 rounded-full border-2 flex items-center justify-center flex-shrink-0 ${
                    selectedMethod === method.id ? "border-primary" : "border-gray-300"
                  }`}>
                    {selectedMethod === method.id && (
                      <div className="w-3 h-3 rounded-full bg-primary" />
                    )}
                  </div>
                  <div className="w-16 h-10 bg-white rounded-lg border border-gray-100 flex items-center justify-center overflow-hidden p-1">
                    <Image src={method.logo} alt={method.name} width={48} height={32} className="object-contain" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="font-medium text-gray-900">{method.name}</p>
                    <p className="text-sm text-gray-500 mt-0.5">{method.description}</p>
                  </div>
                </motion.button>
              ))}
            </div>
          </div>

          {/* Virtual Account */}
          <div className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
            <div className="p-5 border-b border-gray-100">
              <h2 className="font-semibold text-gray-900">Virtual Account</h2>
              <p className="text-sm text-gray-500 mt-1">Bayar melalui ATM, Mobile Banking, atau Internet Banking</p>
            </div>
            <div className="divide-y divide-gray-100">
              {PAYMENT_METHODS.virtualAccount.map((bank) => (
                <motion.button
                  key={bank.id}
                  onClick={() => setSelectedMethod(bank.id)}
                  className={`w-full p-5 flex items-center gap-4 text-left transition ${
                    selectedMethod === bank.id ? "bg-primary/5" : "hover:bg-gray-50"
                  }`}
                  whileTap={{ scale: 0.99 }}
                >
                  <div className={`w-5 h-5 rounded-full border-2 flex items-center justify-center flex-shrink-0 ${
                    selectedMethod === bank.id ? "border-primary" : "border-gray-300"
                  }`}>
                    {selectedMethod === bank.id && (
                      <div className="w-3 h-3 rounded-full bg-primary" />
                    )}
                  </div>
                  <div className="w-16 h-10 bg-white rounded-lg border border-gray-100 flex items-center justify-center overflow-hidden">
                    <Image src={bank.logo} alt={bank.name} width={48} height={32} className="object-contain" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="font-medium text-gray-900">{bank.name}</p>
                    <p className="text-sm text-gray-500 mt-0.5">{bank.description}</p>
                  </div>
                </motion.button>
              ))}
            </div>
          </div>

          {/* Credit Card */}
          <div className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
            <div className="p-5 border-b border-gray-100">
              <h2 className="font-semibold text-gray-900">Kartu Kredit / Debit</h2>
              <p className="text-sm text-gray-500 mt-1">Bayar dengan kartu kredit atau debit</p>
            </div>
            <div className="divide-y divide-gray-100">
              {PAYMENT_METHODS.creditCard.map((method) => (
                <motion.button
                  key={method.id}
                  onClick={() => setSelectedMethod(method.id)}
                  className={`w-full p-5 flex items-center gap-4 text-left transition ${
                    selectedMethod === method.id ? "bg-primary/5" : "hover:bg-gray-50"
                  }`}
                  whileTap={{ scale: 0.99 }}
                >
                  <div className={`w-5 h-5 rounded-full border-2 flex items-center justify-center flex-shrink-0 ${
                    selectedMethod === method.id ? "border-primary" : "border-gray-300"
                  }`}>
                    {selectedMethod === method.id && (
                      <div className="w-3 h-3 rounded-full bg-primary" />
                    )}
                  </div>
                  <div className="w-16 h-10 bg-white rounded-lg border border-gray-100 flex items-center justify-center overflow-hidden p-1">
                    <Image src={method.logo} alt={method.name} width={48} height={32} className="object-contain" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="font-medium text-gray-900">{method.name}</p>
                    <p className="text-sm text-gray-500 mt-0.5">{method.description}</p>
                  </div>
                </motion.button>
              ))}
            </div>
          </div>
        </div>

        {/* Pay Button */}
        <div className="mt-6">
          <button
            onClick={handlePayment}
            disabled={!selectedMethod || submitting}
            className="w-full py-4 bg-primary text-white font-semibold rounded-xl hover:bg-gray-800 transition disabled:bg-gray-200 disabled:text-gray-400 disabled:cursor-not-allowed"
          >
            {submitting ? "Memproses..." : "Bayar Sekarang"}
          </button>
        </div>

        {/* Security Notice */}
        <div className="mt-4 flex items-center justify-center gap-2 text-xs text-gray-400">
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
          </svg>
          <span>Pembayaran aman via Midtrans</span>
        </div>
      </div>
    </div>
  );
}

export default function PaymentSelectionPage() {
  return (
    <Suspense fallback={<LoadingOverlay message="Memuat..." />}>
      <PaymentSelectionContent />
    </Suspense>
  );
}
