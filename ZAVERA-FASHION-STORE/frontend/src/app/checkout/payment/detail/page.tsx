"use client";

import { useState, useEffect, useCallback, Suspense } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import Image from "next/image";
import { motion, AnimatePresence } from "framer-motion";
import { useToast } from "@/components/ui/Toast";
import api from "@/lib/api";
import { LoadingOverlay } from "@/components/ui/LoadingSpinner";

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

// Bank instructions
const BANK_INSTRUCTIONS: Record<string, Array<{ channel: string; steps: string[] }>> = {
  bca: [
    {
      channel: "ATM BCA",
      steps: [
        "Masukkan kartu ATM dan PIN",
        "Pilih menu Transaksi Lainnya",
        "Pilih Transfer ke BCA Virtual Account",
        "Masukkan nomor Virtual Account",
        "Konfirmasi pembayaran",
      ],
    },
    {
      channel: "m-BCA (BCA mobile)",
      steps: [
        "Login ke aplikasi m-BCA",
        "Pilih m-Transfer > BCA Virtual Account",
        "Masukkan nomor Virtual Account",
        "Masukkan PIN m-BCA",
        "Konfirmasi pembayaran",
      ],
    },
    {
      channel: "KlikBCA (Internet Banking)",
      steps: [
        "Login ke KlikBCA",
        "Pilih Transfer Dana > Transfer ke BCA Virtual Account",
        "Masukkan nomor Virtual Account",
        "Masukkan respon KeyBCA",
        "Konfirmasi pembayaran",
      ],
    },
  ],
  bri: [
    {
      channel: "ATM BRI",
      steps: [
        "Masukkan kartu ATM dan PIN",
        "Pilih menu Transaksi Lain > Pembayaran > Lainnya > BRIVA",
        "Masukkan nomor Virtual Account",
        "Konfirmasi pembayaran",
      ],
    },
    {
      channel: "BRImo (Mobile Banking)",
      steps: [
        "Login ke aplikasi BRImo",
        "Pilih BRIVA",
        "Masukkan nomor Virtual Account",
        "Konfirmasi pembayaran",
      ],
    },
  ],
  bni: [
    {
      channel: "ATM BNI",
      steps: [
        "Masukkan kartu ATM dan PIN",
        "Pilih Menu Lain > Transfer > Rekening Tabungan",
        "Pilih Virtual Account Billing",
        "Masukkan nomor Virtual Account",
        "Konfirmasi pembayaran",
      ],
    },
    {
      channel: "BNI Mobile Banking",
      steps: [
        "Login ke aplikasi BNI Mobile Banking",
        "Pilih Transfer > Virtual Account Billing",
        "Masukkan nomor Virtual Account",
        "Konfirmasi pembayaran",
      ],
    },
  ],
  mandiri: [
    {
      channel: "ATM Mandiri",
      steps: [
        "Masukkan kartu ATM dan PIN",
        "Pilih Bayar/Beli > Multipayment",
        "Masukkan kode perusahaan: 70012",
        "Masukkan nomor Virtual Account",
        "Konfirmasi pembayaran",
      ],
    },
    {
      channel: "Livin' by Mandiri",
      steps: [
        "Login ke aplikasi Livin'",
        "Pilih Bayar > Multipayment",
        "Pilih penyedia jasa: Midtrans",
        "Masukkan nomor Virtual Account",
        "Konfirmasi pembayaran",
      ],
    },
  ],
  permata: [
    {
      channel: "ATM Permata/Alto",
      steps: [
        "Masukkan kartu ATM dan PIN",
        "Pilih Transaksi Lainnya > Pembayaran > Pembayaran Lainnya",
        "Pilih Virtual Account",
        "Masukkan nomor Virtual Account",
        "Konfirmasi pembayaran",
      ],
    },
    {
      channel: "PermataMobile X",
      steps: [
        "Login ke aplikasi PermataMobile X",
        "Pilih Bayar Tagihan > Virtual Account",
        "Masukkan nomor Virtual Account",
        "Konfirmasi pembayaran",
      ],
    },
  ],
  gopay: [
    {
      channel: "Aplikasi Gojek",
      steps: [
        "Buka aplikasi Gojek di smartphone Anda",
        "Tap menu 'Bayar' atau scan QR code",
        "Jika scan QR, arahkan kamera ke QR code pembayaran",
        "Periksa detail pembayaran dan tap 'Konfirmasi & Bayar'",
        "Masukkan PIN GoPay Anda",
        "Pembayaran selesai",
      ],
    },
    {
      channel: "Aplikasi GoPay",
      steps: [
        "Buka aplikasi GoPay di smartphone Anda",
        "Tap 'Scan QR' atau 'Bayar'",
        "Scan QR code pembayaran",
        "Periksa detail pembayaran",
        "Masukkan PIN GoPay Anda",
        "Pembayaran selesai",
      ],
    },
  ],
  qris: [
    {
      channel: "E-Wallet (GoPay, OVO, Dana, dll)",
      steps: [
        "Buka aplikasi e-wallet pilihan Anda",
        "Pilih menu 'Scan' atau 'Bayar'",
        "Scan QR code pembayaran",
        "Periksa detail pembayaran",
        "Masukkan PIN atau konfirmasi pembayaran",
        "Pembayaran selesai",
      ],
    },
  ],
};

interface PaymentDetails {
  payment_id: number;
  order_id: number;
  order_code: string;
  payment_method: string;
  bank: string;
  bank_logo: string;
  va_number: string;
  amount: number;
  expiry_time: string;
  remaining_seconds: number;
  status: string;
  instructions: Array<{ channel: string; steps: string[] }>;
  // GoPay specific fields
  qr_code_url?: string;
  deeplink_url?: string;
  // Order details for receipt
  order_details?: {
    items: Array<{
      product_name: string;
      product_image: string;
      quantity: number;
      price_per_unit: number;
      subtotal: number;
    }>;
    subtotal: number;
    shipping_cost: number;
    total: number;
    customer_name: string;
    customer_email: string;
    customer_phone: string;
    shipping_address: string;
    courier_name?: string;
    courier_service?: string;
  };
}


// Elegant Countdown Timer - Zavera Style
const CountdownTimer = ({ expiryTime, onExpire }: { expiryTime: string; onExpire: () => void }) => {
  const [remaining, setRemaining] = useState(0);

  useEffect(() => {
    const calculateRemaining = () => {
      const expiry = new Date(expiryTime).getTime();
      const now = Date.now();
      return Math.max(0, Math.floor((expiry - now) / 1000));
    };

    setRemaining(calculateRemaining());

    const interval = setInterval(() => {
      const newRemaining = calculateRemaining();
      setRemaining(newRemaining);
      if (newRemaining <= 0) {
        onExpire();
        clearInterval(interval);
      }
    }, 1000);

    return () => clearInterval(interval);
  }, [expiryTime, onExpire]);

  if (remaining <= 0) {
    return (
      <div className="text-center">
        <span className="text-red-600 font-medium">Waktu pembayaran telah habis</span>
      </div>
    );
  }

  const hours = Math.floor(remaining / 3600);
  const minutes = Math.floor((remaining % 3600) / 60);
  const seconds = remaining % 60;

  return (
    <div className="flex items-center justify-center gap-3">
      <div className="text-center">
        <div className="w-16 h-16 bg-primary text-white rounded-lg flex items-center justify-center mb-1">
          <span className="text-2xl font-bold font-mono">{hours.toString().padStart(2, "0")}</span>
        </div>
        <span className="text-xs text-muted uppercase tracking-wider">Jam</span>
      </div>
      <span className="text-2xl font-light text-muted mb-5">:</span>
      <div className="text-center">
        <div className="w-16 h-16 bg-primary text-white rounded-lg flex items-center justify-center mb-1">
          <span className="text-2xl font-bold font-mono">{minutes.toString().padStart(2, "0")}</span>
        </div>
        <span className="text-xs text-muted uppercase tracking-wider">Menit</span>
      </div>
      <span className="text-2xl font-light text-muted mb-5">:</span>
      <div className="text-center">
        <div className="w-16 h-16 bg-primary text-white rounded-lg flex items-center justify-center mb-1">
          <span className="text-2xl font-bold font-mono">{seconds.toString().padStart(2, "0")}</span>
        </div>
        <span className="text-xs text-muted uppercase tracking-wider">Detik</span>
      </div>
    </div>
  );
};

// Elegant Instruction Accordion
const InstructionAccordion = ({ instructions, bank }: { instructions: Array<{ channel: string; steps: string[] }>; bank: string }) => {
  const [openIndex, setOpenIndex] = useState<number | null>(0);
  const displayInstructions = instructions.length > 0 ? instructions : BANK_INSTRUCTIONS[bank] || [];

  return (
    <div className="space-y-3">
      {displayInstructions.map((instruction, idx) => (
        <div key={idx} className="border border-accent rounded-lg overflow-hidden">
          <button
            onClick={() => setOpenIndex(openIndex === idx ? null : idx)}
            className="w-full px-5 py-4 flex items-center justify-between bg-secondary hover:bg-accent/30 transition"
          >
            <span className="font-medium text-primary">{instruction.channel}</span>
            <motion.svg
              animate={{ rotate: openIndex === idx ? 180 : 0 }}
              transition={{ duration: 0.2 }}
              className="w-5 h-5 text-muted"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M19 9l-7 7-7-7" />
            </motion.svg>
          </button>
          <AnimatePresence>
            {openIndex === idx && (
              <motion.div
                initial={{ height: 0, opacity: 0 }}
                animate={{ height: "auto", opacity: 1 }}
                exit={{ height: 0, opacity: 0 }}
                transition={{ duration: 0.2 }}
                className="overflow-hidden"
              >
                <div className="px-5 py-4 bg-white border-t border-accent">
                  <ol className="space-y-3">
                    {instruction.steps.map((step, stepIdx) => (
                      <li key={stepIdx} className="flex gap-4">
                        <span className="flex-shrink-0 w-7 h-7 bg-primary text-white rounded-full flex items-center justify-center text-sm font-medium">
                          {stepIdx + 1}
                        </span>
                        <span className="text-muted pt-0.5 leading-relaxed">{step}</span>
                      </li>
                    ))}
                  </ol>
                </div>
              </motion.div>
            )}
          </AnimatePresence>
        </div>
      ))}
    </div>
  );
};

// Order Summary Component (Receipt Style)
const OrderSummary = ({ orderDetails }: { orderDetails: PaymentDetails['order_details'] }) => {
  if (!orderDetails) return null;
  
  return (
    <div className="bg-white rounded-2xl shadow-sm border border-accent overflow-hidden">
      {/* Header */}
      <div className="p-6 border-b border-accent bg-secondary/30">
        <h3 className="text-lg font-serif font-bold text-primary">Ringkasan Belanja</h3>
      </div>
      
      {/* Items */}
      <div className="p-6 space-y-4">
        {orderDetails.items.map((item, idx) => (
          <div key={idx} className="flex gap-3">
            <div className="w-16 h-16 bg-gray-100 rounded-lg overflow-hidden flex-shrink-0">
              {item.product_image ? (
                <Image
                  src={item.product_image}
                  alt={item.product_name}
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
            <div className="flex-1 min-w-0">
              <p className="font-medium text-gray-900 truncate">{item.product_name}</p>
              <p className="text-sm text-muted">{item.quantity}x @ Rp{item.price_per_unit.toLocaleString("id-ID")}</p>
            </div>
            <div className="text-right flex-shrink-0">
              <p className="font-medium text-gray-900">Rp{item.subtotal.toLocaleString("id-ID")}</p>
            </div>
          </div>
        ))}
        
        {/* Breakdown */}
        <div className="space-y-2 pt-4 border-t border-accent">
          <div className="flex justify-between text-sm">
            <span className="text-muted">Subtotal ({orderDetails.items.length} barang)</span>
            <span className="text-gray-900">Rp{orderDetails.subtotal.toLocaleString("id-ID")}</span>
          </div>
          <div className="flex justify-between text-sm">
            <span className="text-muted">Ongkos Kirim</span>
            <span className="text-gray-900">Rp{orderDetails.shipping_cost.toLocaleString("id-ID")}</span>
          </div>
          {orderDetails.courier_name && (
            <p className="text-xs text-muted">
              {orderDetails.courier_name} - {orderDetails.courier_service}
            </p>
          )}
        </div>
        
        {/* Total */}
        <div className="pt-4 border-t-2 border-primary">
          <div className="flex items-center justify-between">
            <span className="text-lg font-bold text-primary">Total Pembayaran</span>
            <span className="text-2xl font-serif font-bold text-primary">
              Rp{orderDetails.total.toLocaleString("id-ID")}
            </span>
          </div>
        </div>
        
        {/* Shipping Address */}
        {orderDetails.shipping_address && (
          <div className="pt-4 border-t border-accent">
            <p className="text-xs text-muted uppercase tracking-wider mb-2">Alamat Pengiriman</p>
            <p className="text-sm text-gray-900 font-medium">{orderDetails.customer_name}</p>
            <p className="text-sm text-muted">{orderDetails.customer_phone}</p>
            <p className="text-sm text-muted mt-1">{orderDetails.shipping_address}</p>
          </div>
        )}
      </div>
    </div>
  );
};


function VADetailContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const orderId = searchParams.get("order_id");
  const { showToast } = useToast();

  const [loading, setLoading] = useState(true);
  const [checking, setChecking] = useState(false);
  const [payment, setPayment] = useState<PaymentDetails | null>(null);
  const [copied, setCopied] = useState(false);
  const [copiedAmount, setCopiedAmount] = useState(false);
  const [lastCheckTime, setLastCheckTime] = useState(0);
  const [autoCheckEnabled, setAutoCheckEnabled] = useState(true);

  // Prevent back navigation to checkout page
  useEffect(() => {
    // Replace history state to prevent back to checkout
    if (typeof window !== 'undefined') {
      window.history.pushState(null, '', window.location.href);
      
      const handlePopState = () => {
        window.history.pushState(null, '', window.location.href);
        showToast("Gunakan tombol di halaman untuk navigasi", "info");
      };
      
      window.addEventListener('popstate', handlePopState);
      
      return () => {
        window.removeEventListener('popstate', handlePopState);
      };
    }
  }, [showToast]);

  // Auto-check payment status every 10 seconds
  useEffect(() => {
    if (!payment || !autoCheckEnabled || payment.status === 'PAID' || payment.status === 'EXPIRED') {
      return;
    }

    console.log('🔄 Auto-check payment status enabled');
    
    const interval = setInterval(async () => {
      console.log('⏰ Auto-checking payment status...');
      try {
        const response = await api.post("/payments/core/check", {
          payment_id: payment.payment_id,
        });

        if (response.data.status === "PAID") {
          console.log('✅ Payment confirmed as PAID');
          showToast("Pembayaran berhasil!", "success");
          setAutoCheckEnabled(false); // Stop auto-check
          router.push(`/order-success?code=${payment.order_code}`);
        } else if (response.data.status === "EXPIRED") {
          console.log('⏰ Payment expired');
          setPayment(prev => prev ? { ...prev, status: "EXPIRED" } : null);
          setAutoCheckEnabled(false); // Stop auto-check
        } else {
          console.log('⏳ Payment still pending');
        }
      } catch (error) {
        console.error('❌ Auto-check error:', error);
      }
    }, 10000); // Check every 10 seconds

    return () => {
      console.log('🛑 Auto-check stopped');
      clearInterval(interval);
    };
  }, [payment, autoCheckEnabled, router, showToast]);

  useEffect(() => {
    if (!orderId) {
      showToast("Order ID tidak ditemukan", "error");
      router.push("/");
      return;
    }
    loadPaymentDetails();
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [orderId]);

  const loadPaymentDetails = async () => {
    try {
      setLoading(true);
      const response = await api.get(`/payments/core/${orderId}`);
      setPayment(response.data);
    } catch (err: unknown) {
      const axiosError = err as { response?: { status?: number } };
      if (axiosError.response?.status === 404) {
        router.push(`/checkout/payment?order_id=${orderId}`);
      } else if (axiosError.response?.status === 410) {
        showToast("Pembayaran telah kadaluarsa", "error");
      } else {
        showToast("Gagal memuat detail pembayaran", "error");
      }
    } finally {
      setLoading(false);
    }
  };

  const copyToClipboard = async (text: string, type: "va" | "amount") => {
    try {
      await navigator.clipboard.writeText(text);
      if (type === "va") {
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
      } else {
        setCopiedAmount(true);
        setTimeout(() => setCopiedAmount(false), 2000);
      }
      showToast("Berhasil disalin", "success");
    } catch {
      showToast("Gagal menyalin", "error");
    }
  };

  const checkPaymentStatus = async () => {
    if (!payment) return;

    const now = Date.now();
    if (now - lastCheckTime < 5000) {
      showToast("Tunggu beberapa detik sebelum cek status lagi", "warning");
      return;
    }

    try {
      setChecking(true);
      setLastCheckTime(now);
      
      const response = await api.post("/payments/core/check", {
        payment_id: payment.payment_id,
      });

      if (response.data.status === "PAID") {
        showToast("Pembayaran berhasil!", "success");
        router.push(`/order-success?code=${payment.order_code}`);
      } else if (response.data.status === "EXPIRED") {
        showToast("Pembayaran telah kadaluarsa", "error");
        setPayment(prev => prev ? { ...prev, status: "EXPIRED" } : null);
      } else if (response.data.status === "CANCELLED") {
        showToast("Pesanan telah dibatalkan oleh admin", "error");
        setPayment(prev => prev ? { ...prev, status: "CANCELLED" } : null);
        setAutoCheckEnabled(false); // Stop auto-check
      } else {
        showToast(response.data.message || "Pembayaran belum diterima", "info");
      }
    } catch (error: any) {
      if (error.response?.data?.message) {
        showToast(error.response.data.message, "error");
      } else {
        showToast("Gagal memeriksa status pembayaran", "error");
      }
    } finally {
      setChecking(false);
    }
  };

  const handleExpire = useCallback(() => {
    setPayment(prev => prev ? { ...prev, status: "EXPIRED" } : null);
    showToast("Waktu pembayaran telah habis", "error");
  }, [showToast]);

  if (loading) {
    return <LoadingOverlay message="Memuat detail pembayaran..." />;
  }

  if (!payment) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-secondary">
        <div className="text-center">
          <h1 className="text-xl font-serif font-bold text-primary mb-2">Pembayaran Tidak Ditemukan</h1>
          <Link href="/" className="text-muted hover:text-primary transition">Kembali ke Beranda</Link>
        </div>
      </div>
    );
  }

  const isExpired = payment.status === "EXPIRED";
  const isPaid = payment.status === "PAID";
  const isCancelled = payment.status === "CANCELLED";


  return (
    <div className="min-h-screen bg-secondary">
      {/* Elegant Header */}
      <div className="bg-white border-b border-accent">
        <div className="max-w-7xl mx-auto px-6 py-4 flex items-center gap-4">
          <Link 
            href="/account/pembelian" 
            className="w-10 h-10 flex items-center justify-center rounded-full hover:bg-accent transition"
          >
            <svg className="w-5 h-5 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M15 19l-7-7 7-7" />
            </svg>
          </Link>
          <h1 className="text-lg font-serif font-bold text-primary">Detail Pembayaran</h1>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-6 py-8">
        {/* 2 Column Layout */}
        <div className="grid lg:grid-cols-2 gap-6">
          {/* LEFT COLUMN - Payment Info */}
          <div className="space-y-6">
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="bg-white rounded-2xl shadow-sm border border-accent overflow-hidden"
            >
              {/* Status Header */}
              <div className="p-8 text-center border-b border-accent bg-gradient-to-b from-white to-secondary/50">
                {/* Status Icon */}
                <div className={`w-16 h-16 mx-auto mb-4 rounded-full flex items-center justify-center ${
                  isPaid ? "bg-primary/10" : 
                  isCancelled ? "bg-orange-100" :
                  isExpired ? "bg-red-100" : "bg-primary"
                }`}>
                  {isPaid ? (
                    <svg className="w-8 h-8 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                    </svg>
                  ) : isCancelled ? (
                    <svg className="w-8 h-8 text-orange-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                    </svg>
                  ) : isExpired ? (
                    <svg className="w-8 h-8 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  ) : (
                    <svg className="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                  )}
                </div>

                <h2 className="text-xl font-serif font-bold text-primary mb-1">
                  {isPaid ? "Pembayaran Berhasil" : 
                   isCancelled ? "Pesanan Dibatalkan" :
                   isExpired ? "Pembayaran Kadaluarsa" : "Menunggu Pembayaran"}
                </h2>
                <p className="text-muted text-sm">
                  {isPaid 
                    ? "Terima kasih, pembayaran Anda telah dikonfirmasi"
                    : isCancelled
                      ? "Pesanan ini telah dibatalkan oleh admin. Silakan hubungi customer service untuk informasi lebih lanjut."
                      : isExpired 
                        ? "Waktu pembayaran telah berakhir"
                        : `Selesaikan pembayaran sebelum ${new Date(payment.expiry_time).toLocaleDateString("id-ID", {
                            day: "numeric",
                            month: "long",
                            year: "numeric",
                          })}, ${new Date(payment.expiry_time).toLocaleTimeString("id-ID", {
                            hour: "2-digit",
                            minute: "2-digit",
                          })} WIB`
                  }
                </p>

                {/* Countdown Timer */}
                {!isExpired && !isPaid && !isCancelled && (
                  <div className="mt-6">
                    <CountdownTimer expiryTime={payment.expiry_time} onExpire={handleExpire} />
                  </div>
                )}
              </div>

              {/* Payment Details */}
              <div className="p-8 space-y-6">
                {/* Bank/Payment Method Info */}
                <div className="p-5 bg-secondary rounded-xl">
                  <div className="flex items-center gap-4 mb-4">
                    <div className="w-16 h-10 bg-white rounded-lg border border-accent flex items-center justify-center p-2">
                      <Image
                        src={BANK_LOGOS[payment.bank] || payment.bank_logo || "/images/banks/default.svg"}
                        alt={payment.bank.toUpperCase()}
                        width={48}
                        height={32}
                        className="object-contain"
                      />
                    </div>
                    <div>
                      <p className="font-medium text-primary">
                        {payment.bank === "gopay" ? "GoPay" : 
                         payment.bank === "qris" ? "QRIS" : 
                         `${payment.bank.toUpperCase()} Virtual Account`}
                      </p>
                      <p className="text-sm text-muted">{payment.order_code}</p>
                    </div>
                  </div>

                  <div className="space-y-4">
                {/* QRIS QR Code - Zavera Theme Design */}
                {payment.bank === "qris" && payment.qr_code_url && (
                  <div className="flex justify-center">
                    <div className="w-full max-w-xs">
                      {/* QRIS Header - Zavera Theme */}
                      <div className="bg-primary p-4 rounded-t-xl">
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-3">
                            <Image src="/images/payments/qris.png" alt="QRIS" width={45} height={18} className="object-contain brightness-0 invert" />
                            <div className="text-white text-[11px] leading-tight">
                              <p className="font-semibold">QR Code Standar</p>
                              <p className="opacity-80">Pembayaran Nasional</p>
                            </div>
                          </div>
                          <div className="bg-white/10 backdrop-blur rounded px-2 py-1">
                            <span className="text-white text-[10px] font-bold tracking-wider">GPN</span>
                          </div>
                        </div>
                        {/* Merchant Info */}
                        <div className="mt-3 pt-3 border-t border-white/20">
                          <p className="text-white font-serif text-sm tracking-wide">ZAVERA Fashion Store</p>
                          <p className="text-white/60 text-[10px] mt-0.5">NMID: {payment.order_code}</p>
                        </div>
                      </div>
                      
                      {/* QR Code Container */}
                      <div className="bg-white p-5 border-x border-accent">
                        <Image
                          src={payment.qr_code_url}
                          alt="QRIS QR Code"
                          width={200}
                          height={200}
                          className="mx-auto"
                        />
                      </div>
                      
                      {/* QRIS Footer - Zavera Theme */}
                      <div className="bg-secondary p-3 rounded-b-xl border border-t-0 border-accent">
                        <p className="text-[10px] text-muted text-center">
                          Dicetak oleh: <span className="font-medium text-primary">{payment.order_code}</span>
                        </p>
                        <p className="text-[10px] text-muted/70 text-center mt-1">
                          Scan menggunakan aplikasi e-wallet favorit Anda
                        </p>
                      </div>
                    </div>
                  </div>
                )}

                {/* GoPay QR Code */}
                {payment.bank === "gopay" && payment.qr_code_url && (
                  <div className="text-center">
                    <p className="text-xs text-muted uppercase tracking-wider mb-3">Scan QR Code</p>
                    <div className="inline-block p-4 bg-white rounded-xl border border-accent">
                      <Image
                        src={payment.qr_code_url}
                        alt="QR Code"
                        width={200}
                        height={200}
                        className="mx-auto"
                      />
                    </div>
                    {payment.deeplink_url && (
                      <div className="mt-4">
                        <a
                          href={payment.deeplink_url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="inline-flex items-center gap-2 px-6 py-3 bg-[#00AA13] text-white font-medium rounded-xl hover:bg-[#009911] transition"
                        >
                          <Image src="/images/payments/gopay.png" alt="GoPay" width={24} height={24} />
                          Bayar dengan GoPay
                        </a>
                        <p className="text-xs text-muted mt-2">Atau scan QR code di atas</p>
                      </div>
                    )}
                  </div>
                )}

                {/* VA Number - Only show for VA payments */}
                {payment.bank !== "gopay" && payment.bank !== "qris" && (
                  <div>
                    <p className="text-xs text-muted uppercase tracking-wider mb-2">Nomor Virtual Account</p>
                    <div className="flex items-center gap-3">
                      <span className="flex-1 text-xl font-bold text-primary tracking-widest break-all" style={{ fontFamily: 'Consolas, "Courier New", monospace', letterSpacing: '0.1em' }}>
                        {payment.va_number}
                      </span>
                      <button
                        onClick={() => copyToClipboard(payment.va_number, "va")}
                        className={`flex-shrink-0 px-4 py-2 text-sm font-medium rounded-lg transition whitespace-nowrap ${
                          copied
                            ? "bg-primary text-white"
                            : "bg-white border border-accent text-primary hover:bg-accent"
                        }`}
                      >
                        {copied ? "Tersalin!" : "Salin"}
                      </button>
                    </div>
                    <p className="text-xs text-muted mt-1 italic">Pastikan menyalin nomor dengan benar (angka 0 bukan huruf O)</p>
                  </div>
                )}

                {/* Amount */}
                <div className="pt-4 border-t border-accent">
                  <p className="text-xs text-muted uppercase tracking-wider mb-2">Total Pembayaran</p>
                  <div className="flex items-center gap-3">
                    <span className="text-2xl font-serif font-bold text-primary">
                      Rp{payment.amount.toLocaleString("id-ID")}
                    </span>
                  </div>
                  <p className="text-xs text-muted mt-1">Pastikan nominal transfer sesuai hingga digit terakhir</p>
                </div>
              </div>
            </div>

            {/* Important Notes */}
            {!isExpired && !isPaid && !isCancelled && (
              <div className="p-5 bg-amber-50 rounded-xl border border-amber-200">
                <div className="flex gap-3">
                  <svg className="w-5 h-5 text-amber-600 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  <div className="text-sm text-amber-800 space-y-2">
                    {payment.bank === "gopay" ? (
                      <>
                        <p><span className="font-medium">Penting:</span> Scan QR code atau klik tombol &quot;Bayar dengan GoPay&quot;</p>
                        <p>Pastikan saldo GoPay Anda mencukupi</p>
                      </>
                    ) : payment.bank === "qris" ? (
                      <>
                        <p><span className="font-medium">Penting:</span> Scan QR code menggunakan aplikasi e-wallet Anda</p>
                        <p>Mendukung GoPay, OVO, Dana, LinkAja, dan lainnya</p>
                      </>
                    ) : (
                      <>
                        <p><span className="font-medium">Penting:</span> Transfer hanya bisa dilakukan dari bank yang sama ({payment.bank.toUpperCase()})</p>
                        <p>Pastikan nominal transfer sesuai hingga digit terakhir</p>
                      </>
                    )}
                  </div>
                </div>
              </div>
            )}

            {/* Cancelled Order Notice */}
            {isCancelled && (
              <div className="p-5 bg-orange-50 rounded-xl border border-orange-200">
                <div className="flex gap-3">
                  <svg className="w-5 h-5 text-orange-600 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                  </svg>
                  <div className="text-sm text-orange-800 space-y-2">
                    <p><span className="font-medium">Pesanan Dibatalkan</span></p>
                    <p>Pesanan ini telah dibatalkan oleh admin. Jika Anda sudah melakukan pembayaran, dana akan dikembalikan dalam 3-7 hari kerja.</p>
                    <p className="mt-3">Untuk informasi lebih lanjut, silakan hubungi customer service kami:</p>
                    <div className="mt-2 space-y-1">
                      <p>📧 Email: support@zavera.com</p>
                      <p>📱 WhatsApp: +62 812-3456-7890</p>
                    </div>
                  </div>
                </div>
              </div>
            )}
              </div>
            </motion.div>
          </div>

          {/* RIGHT COLUMN - Order Summary + Instructions + Actions */}
          <div className="space-y-6">
            {/* Order Summary (Receipt) */}
            {payment.order_details && (
              <OrderSummary orderDetails={payment.order_details} />
            )}

            {/* Payment Instructions */}
            {!isExpired && !isPaid && !isCancelled && (
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.1 }}
                className="bg-white rounded-2xl shadow-sm border border-accent overflow-hidden"
              >
                <div className="p-6">
                  <h3 className="text-lg font-serif font-bold text-primary mb-4">Cara Pembayaran</h3>
                  <InstructionAccordion instructions={payment.instructions} bank={payment.bank} />
                </div>
              </motion.div>
            )}

            {/* Action Buttons */}
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.2 }}
              className="bg-white rounded-2xl shadow-sm border border-accent overflow-hidden"
            >
              <div className="p-6">
                {!isExpired && !isPaid && !isCancelled ? (
                  <div className="space-y-3">
                    <button
                      onClick={checkPaymentStatus}
                      disabled={checking}
                      className="w-full py-4 px-6 bg-primary text-white font-medium rounded-lg hover:bg-primary/90 transition disabled:bg-accent disabled:text-muted"
                    >
                      {checking ? "Memeriksa..." : "Cek Status Pembayaran"}
                    </button>
                    <Link
                      href={`/orders/${payment.order_code}`}
                      className="block w-full py-4 px-6 border-2 border-primary text-primary font-medium rounded-lg text-center hover:bg-primary hover:text-white transition"
                    >
                      Lihat Detail Pesanan
                    </Link>
                  </div>
                ) : (
                  <div className="space-y-3">
                    <Link
                      href={`/orders/${payment.order_code}`}
                      className="block w-full py-4 px-6 bg-primary text-white font-medium rounded-lg text-center hover:bg-primary/90 transition"
                    >
                      Lihat Detail Pesanan
                    </Link>
                    {isCancelled && (
                      <Link
                        href="/"
                        className="block w-full py-4 px-6 border-2 border-primary text-primary font-medium rounded-lg text-center hover:bg-primary hover:text-white transition"
                      >
                        Kembali Berbelanja
                      </Link>
                    )}
                    {(isExpired || isPaid) && (
                      <Link
                        href="/"
                        className="block w-full py-4 px-6 border-2 border-primary text-primary font-medium rounded-lg text-center hover:bg-primary hover:text-white transition"
                      >
                        Kembali ke Beranda
                      </Link>
                    )}
                  </div>
                )}
              </div>
            </motion.div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default function VADetailPage() {
  return (
    <Suspense fallback={<LoadingOverlay message="Memuat detail pembayaran..." />}>
      <VADetailContent />
    </Suspense>
  );
}
