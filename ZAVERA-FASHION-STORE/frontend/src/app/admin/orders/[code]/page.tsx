"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import {
  ArrowLeft,
  Package,
  Truck,
  CreditCard,
  RefreshCcw,
  XCircle,
  CheckCircle,
  Clock,
  AlertTriangle,
  User,
  Mail,
  Phone,
  MapPin,
  FileText,
  PackageCheck,
  Send,
} from "lucide-react";
import api from "@/lib/api";
import { forceCancel, forceRefund, forceReship } from "@/lib/adminApi";
import ConfirmDialog from "@/components/ui/ConfirmDialog";

interface OrderDetail {
  id: number;
  order_code: string;
  customer_name: string;
  customer_email: string;
  customer_phone: string;
  subtotal: number;
  shipping_cost: number;
  tax: number;
  discount: number;
  total_amount: number;
  status: string;
  resi?: string;
  refund_status?: string;
  refund_amount?: number;
  created_at: string;
  items: {
    id: number;
    product_id: number;
    product_name: string;
    quantity: number;
    price_per_unit: number;
    subtotal: number;
    image_url?: string;
  }[];
  payment?: {
    id: number;
    status: string;
    payment_method?: string;
    payment_provider?: string;
    paid_at?: string;
  };
  shipment?: {
    id: number;
    tracking_number: string;
    status: string;
    provider_name: string;
    shipped_at?: string;
    delivered_at?: string;
  };
}

interface RefundDetail {
  id: number;
  refund_code: string;
  order_code: string;
  refund_type: string;
  reason: string;
  reason_detail: string;
  original_amount: number;
  refund_amount: number;
  shipping_refund: number;
  items_refund: number;
  status: string;
  gateway_refund_id?: string;
  processed_by?: number;
  processed_at?: string;
  requested_by?: number;
  requested_at: string;
  completed_at?: string;
  items?: {
    id: number;
    order_item_id: number;
    product_name: string;
    quantity: number;
    price_per_unit: number;
    subtotal: number;
    stock_restored: boolean;
  }[];
  status_history?: {
    id: number;
    old_status: string;
    new_status: string;
    actor: string;
    reason: string;
    created_at: string;
  }[];
}

interface RefundItemSelection {
  order_item_id: number;
  product_name: string;
  quantity: number;
  max_quantity: number;
  price_per_unit: number;
}

export default function OrderDetailPage() {
  const params = useParams();
  const router = useRouter();
  const orderCode = params.code as string;

  const [order, setOrder] = useState<OrderDetail | null>(null);
  const [refunds, setRefunds] = useState<RefundDetail[]>([]);
  const [loading, setLoading] = useState(true);
  const [refundsLoading, setRefundsLoading] = useState(false);
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const [showModal, setShowModal] = useState<string | null>(null);
  const [actionReason, setActionReason] = useState("");
  const [resiInput, setResiInput] = useState("");
  const [resiError, setResiError] = useState("");
  
  // Refund modal state
  const [refundType, setRefundType] = useState<'FULL' | 'PARTIAL' | 'SHIPPING_ONLY' | 'ITEM_ONLY'>('FULL');
  const [refundReason, setRefundReason] = useState('');
  const [refundReasonDetail, setRefundReasonDetail] = useState('');
  const [refundAmount, setRefundAmount] = useState<string>('');
  const [selectedItems, setSelectedItems] = useState<RefundItemSelection[]>([]);
  const [refundError, setRefundError] = useState('');
  
  // Toast notification state
  const [showToast, setShowToast] = useState(false);
  const [toastMessage, setToastMessage] = useState('');
  const [toastType, setToastType] = useState<'success' | 'error'>('success');
  
  // Confirm dialog state
  const [showConfirm, setShowConfirm] = useState(false);
  const [confirmConfig, setConfirmConfig] = useState<{
    title: string;
    message: string;
    onConfirm: () => void;
    variant?: 'danger' | 'warning' | 'info';
  }>({
    title: '',
    message: '',
    onConfirm: () => {},
  });
  
  // Note input modal state
  const [showNoteModal, setShowNoteModal] = useState(false);
  const [noteInput, setNoteInput] = useState('');
  const [noteModalConfig, setNoteModalConfig] = useState<{
    title: string;
    message: string;
    placeholder: string;
    onConfirm: (note: string) => void;
  }>({
    title: '',
    message: '',
    placeholder: '',
    onConfirm: () => {},
  });

  useEffect(() => {
    loadOrder();
    loadRefunds();
  }, [orderCode]);

  const loadOrder = async () => {
    try {
      const token = localStorage.getItem("auth_token");
      const response = await api.get(`/admin/orders/${orderCode}`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      setOrder(response.data);
    } catch (error) {
      console.error("Failed to load order:", error);
    } finally {
      setLoading(false);
    }
  };

  const loadRefunds = async () => {
    setRefundsLoading(true);
    try {
      const token = localStorage.getItem("auth_token");
      const response = await api.get(`/admin/orders/${orderCode}/refunds`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      setRefunds(response.data || []);
    } catch (error) {
      console.error("Failed to load refunds:", error);
      setRefunds([]);
    } finally {
      setRefundsLoading(false);
    }
  };

  const handleForceCancel = async () => {
    if (!actionReason.trim()) return;
    setActionLoading("cancel");
    try {
      console.log("Cancelling order:", orderCode, "Reason:", actionReason);
      const response = await forceCancel(orderCode, actionReason);
      console.log("Cancel response:", response);
      setShowModal(null);
      setActionReason("");
      loadOrder();
    } catch (error: any) {
      console.error("Failed to cancel:", error);
      console.error("Error response:", error.response?.data);
      const errorMsg = error.response?.data?.message || error.response?.data?.error || "Failed to cancel order";
      showErrorToast(errorMsg);
    } finally {
      setActionLoading(null);
    }
  };

  const handleForceRefund = async () => {
    if (!refundReason.trim()) {
      setRefundError('Please select a reason');
      return;
    }
    
    // Validate based on refund type
    if (refundType === 'PARTIAL') {
      const amount = parseFloat(refundAmount);
      if (isNaN(amount) || amount <= 0) {
        setRefundError('Please enter a valid refund amount');
        return;
      }
      const refundableBalance = order!.total_amount - (order!.refund_amount || 0);
      if (amount > refundableBalance) {
        setRefundError(`Amount exceeds refundable balance: ${formatCurrency(refundableBalance)}`);
        return;
      }
    }
    
    if (refundType === 'ITEM_ONLY') {
      if (selectedItems.length === 0) {
        setRefundError('Please select at least one item');
        return;
      }
      const hasInvalidQty = selectedItems.some(item => item.quantity <= 0 || item.quantity > item.max_quantity);
      if (hasInvalidQty) {
        setRefundError('Invalid item quantities');
        return;
      }
    }
    
    setActionLoading("refund");
    setRefundError('');
    
    try {
      const token = localStorage.getItem("auth_token");
      const payload: any = {
        order_code: orderCode,
        refund_type: refundType,
        reason: refundReason,
        reason_detail: refundReasonDetail || undefined,
        idempotency_key: `${orderCode}-${Date.now()}`,
      };
      
      if (refundType === 'PARTIAL') {
        payload.amount = parseFloat(refundAmount);
      }
      
      if (refundType === 'ITEM_ONLY') {
        payload.items = selectedItems.map(item => ({
          order_item_id: item.order_item_id,
          quantity: item.quantity,
          reason: refundReason,
        }));
      }
      
      const response = await api.post('/admin/refunds', payload, {
        headers: { Authorization: `Bearer ${token}` },
      });
      
      // Only process if refund is in PENDING status
      // Manual refunds (orders without payment) are auto-completed
      if (response.data.status === 'PENDING') {
        try {
          const processResponse = await api.post(`/admin/refunds/${response.data.id}/process`, {}, {
            headers: { Authorization: `Bearer ${token}` },
          });
          console.log('✅ Refund processed:', processResponse.data);
        } catch (processError: any) {
          console.error('⚠️ Refund process error:', processError);
          const errorMsg = processError.response?.data?.message || processError.message || '';
          
          // Check if it's a manual processing required error (Error 418)
          if (errorMsg.includes('MANUAL_PROCESSING_REQUIRED') || errorMsg.includes('manual bank transfer')) {
            console.log('⚠️ Manual processing required - showing notification');
            setRefundError('MANUAL_PROCESSING_REQUIRED: Automatic refund failed. Please process manual bank transfer to customer and mark refund as completed after transfer is done.');
            
            // Still close modal and reload to show the refund in PENDING state
            setShowModal(null);
            await Promise.all([loadOrder(), loadRefunds()]);
            return; // Don't throw error, just show the message
          }
          
          // For other errors, throw to be caught by outer catch
          throw processError;
        }
      }
      
      console.log('✅ Refund created successfully:', response.data);
      
      setShowModal(null);
      resetRefundForm();
      
      // Force reload order and refunds to get updated status
      await Promise.all([loadOrder(), loadRefunds()]);
      
      // Show success message
      showSuccessToast('Refund berhasil diproses!');
    } catch (error: any) {
      console.error("Failed to refund:", error);
      const errorMsg = error.response?.data?.message || error.response?.data?.error || error.message || "Failed to process refund";
      setRefundError(errorMsg);
    } finally {
      setActionLoading(null);
    }
  };
  
  const resetRefundForm = () => {
    setRefundType('FULL');
    setRefundReason('');
    setRefundReasonDetail('');
    setRefundAmount('');
    setSelectedItems([]);
    setRefundError('');
  };
  
  const showSuccessToast = (message: string) => {
    setToastMessage(message);
    setToastType('success');
    setShowToast(true);
    setTimeout(() => setShowToast(false), 3000);
  };
  
  const showErrorToast = (message: string) => {
    setToastMessage(message);
    setToastType('error');
    setShowToast(true);
    setTimeout(() => setShowToast(false), 5000);
  };
  
  const handleRetryRefund = async (refundId: number) => {
    setConfirmConfig({
      title: 'Retry Refund',
      message: 'Apakah Anda yakin ingin mencoba ulang refund yang gagal ini?',
      variant: 'warning',
      onConfirm: async () => {
        setShowConfirm(false);
        setActionLoading(`retry-${refundId}`);
        try {
          const token = localStorage.getItem("auth_token");
          await api.post(`/admin/refunds/${refundId}/retry`, {}, {
            headers: { Authorization: `Bearer ${token}` },
          });
          loadRefunds();
          loadOrder();
          showSuccessToast('Refund retry berhasil!');
        } catch (error: any) {
          console.error("Failed to retry refund:", error);
          const errorMsg = error.response?.data?.message || error.message || "Failed to retry refund";
          showErrorToast(`Gagal retry: ${errorMsg}`);
        } finally {
          setActionLoading(null);
        }
      }
    });
    setShowConfirm(true);
  };
  
  const handleMarkRefundCompleted = async (refundId: number) => {
    setConfirmConfig({
      title: 'Mark Refund as Completed',
      message: 'Apakah Anda sudah melakukan transfer manual ke customer? Pastikan transfer sudah berhasil sebelum menandai refund sebagai completed.',
      variant: 'warning',
      onConfirm: async () => {
        setShowConfirm(false);
        
        // Show note input modal instead of native prompt
        setNoteModalConfig({
          title: 'Masukkan Catatan Konfirmasi',
          message: 'Masukkan detail transfer manual yang sudah dilakukan:',
          placeholder: `Contoh: Transfer manual via BCA ke rekening customer pada ${new Date().toLocaleDateString('id-ID')}`,
          onConfirm: async (note: string) => {
            if (!note || note.trim() === '') {
              showErrorToast('Catatan konfirmasi diperlukan');
              return;
            }
            
            setShowNoteModal(false);
            setActionLoading(`complete-${refundId}`);
            try {
              const token = localStorage.getItem("auth_token");
              await api.post(`/admin/refunds/${refundId}/mark-completed`, {
                note: note.trim()
              }, {
                headers: { Authorization: `Bearer ${token}` },
              });
              loadRefunds();
              loadOrder();
              showSuccessToast('Refund berhasil ditandai sebagai completed!');
            } catch (error: any) {
              console.error("Failed to mark refund as completed:", error);
              const errorMsg = error.response?.data?.message || error.message || "Failed to mark refund as completed";
              showErrorToast(`Gagal: ${errorMsg}`);
            } finally {
              setActionLoading(null);
            }
          }
        });
        setNoteInput('');
        setShowNoteModal(true);
      }
    });
    setShowConfirm(true);
  };
  
  const initializeItemSelection = () => {
    if (!order) return;
    setSelectedItems(order.items.map(item => ({
      order_item_id: item.id,
      product_name: item.product_name,
      quantity: item.quantity,
      max_quantity: item.quantity,
      price_per_unit: item.price_per_unit,
    })));
  };
  
  const updateItemQuantity = (orderItemId: number, quantity: number) => {
    setSelectedItems(prev => prev.map(item => 
      item.order_item_id === orderItemId 
        ? { ...item, quantity: Math.max(0, Math.min(quantity, item.max_quantity)) }
        : item
    ));
  };
  
  const calculateItemRefundTotal = () => {
    return selectedItems
      .filter(item => item.quantity > 0)
      .reduce((sum, item) => sum + (item.quantity * item.price_per_unit), 0);
  };

  const handleForceReship = async () => {
    if (!actionReason.trim()) return;
    setActionLoading("reship");
    try {
      await forceReship(orderCode, actionReason);
      setShowModal(null);
      setActionReason("");
      loadOrder();
    } catch (error) {
      console.error("Failed to reship:", error);
      showErrorToast("Gagal membuat reship");
    } finally {
      setActionLoading(null);
    }
  };

  // Process order (PAID -> PACKING)
  const handlePackOrder = async () => {
    setActionLoading("pack");
    try {
      const token = localStorage.getItem("auth_token");
      await api.post(`/admin/orders/${orderCode}/pack`, {}, {
        headers: { Authorization: `Bearer ${token}` },
      });
      loadOrder();
    } catch (error) {
      console.error("Failed to pack order:", error);
      showErrorToast("Gagal memproses pesanan");
    } finally {
      setActionLoading(null);
    }
  };

  // Validate resi format
  const validateResi = (resi: string): string => {
    // If empty, allow it (will auto-generate)
    if (!resi || resi.trim() === "") {
      return "";
    }
    
    const trimmed = resi.trim();
    if (trimmed.length < 8) {
      return "Nomor resi minimal 8 karakter";
    }
    // Allow alphanumeric and dash/hyphen
    if (!/^[A-Za-z0-9-]+$/.test(trimmed)) {
      return "Nomor resi hanya boleh huruf, angka, dan tanda strip (-)";
    }
    return "";
  };

  // Generate resi from Biteship (before shipping)
  const handleGenerateResi = async () => {
    setResiError("");
    setActionLoading("generate_resi");
    try {
      const token = localStorage.getItem("auth_token");
      const response = await api.post(`/admin/orders/${orderCode}/generate-resi`, {}, {
        headers: { Authorization: `Bearer ${token}` },
      });
      
      const generatedResi = response.data?.resi || response.data?.waybill_id;
      
      if (generatedResi) {
        // Set resi to input field so admin can see and edit
        setResiInput(generatedResi);
        showSuccessToast(`✅ Resi berhasil di-generate: ${generatedResi}`);
      } else {
        setResiError("Gagal generate resi dari Biteship");
      }
    } catch (error: any) {
      console.error("Failed to generate resi:", error);
      const msg = error.response?.data?.error || error.response?.data?.message || "Gagal generate resi dari Biteship";
      setResiError(msg);
    } finally {
      setActionLoading(null);
    }
  };

  // Ship order with resi (PACKING -> SHIPPED)
  const handleShipOrder = async () => {
    // Validate resi is provided
    if (!resiInput.trim()) {
      setResiError("Nomor resi harus diisi. Klik 'Generate dari Biteship' atau input manual.");
      return;
    }
    
    const error = validateResi(resiInput);
    if (error) {
      setResiError(error);
      return;
    }
    
    setResiError("");
    setActionLoading("ship");
    try {
      const token = localStorage.getItem("auth_token");
      const response = await api.post(`/admin/orders/${orderCode}/ship`, { 
        resi: resiInput.trim()
      }, {
        headers: { Authorization: `Bearer ${token}` },
      });
      
      showSuccessToast(`✅ Pesanan dikirim dengan resi: ${resiInput.trim()}`);
      setShowModal(null);
      setResiInput("");
      loadOrder();
    } catch (error: any) {
      console.error("Failed to ship order:", error);
      const msg = error.response?.data?.error || error.response?.data?.message || "Gagal mengirim pesanan";
      setResiError(msg);
    } finally {
      setActionLoading(null);
    }
  };

  // Mark as delivered (SHIPPED -> DELIVERED)
  const handleDeliverOrder = async () => {
    setActionLoading("deliver");
    try {
      const token = localStorage.getItem("auth_token");
      await api.post(`/admin/orders/${orderCode}/deliver`, {}, {
        headers: { Authorization: `Bearer ${token}` },
      });
      loadOrder();
    } catch (error) {
      console.error("Failed to deliver order:", error);
      showErrorToast("Gagal menandai pesanan selesai");
    } finally {
      setActionLoading(null);
    }
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      minimumFractionDigits: 0,
    }).format(amount);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("id-ID", {
      day: "numeric",
      month: "long",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  // Format payment method display (e.g., "BCA VA", "GoPay", "QRIS")
  const formatPaymentMethod = (payment?: OrderDetail["payment"]) => {
    if (!payment) return "N/A";
    
    const method = payment.payment_method?.toLowerCase() || "";
    const provider = payment.payment_provider?.toLowerCase() || "";
    
    // Bank Transfer (VA)
    if (method === "bank_transfer" || method === "va") {
      const bankName = provider.toUpperCase();
      return `${bankName} VA`;
    }
    
    // E-Wallet
    if (method === "gopay" || provider === "gopay") return "GoPay";
    if (method === "shopeepay" || provider === "shopeepay") return "ShopeePay";
    if (method === "dana" || provider === "dana") return "DANA";
    if (method === "ovo" || provider === "ovo") return "OVO";
    
    // QRIS
    if (method === "qris") return "QRIS";
    
    // Credit Card
    if (method === "credit_card" || method === "cc") return "Credit Card";
    
    // Fallback
    if (provider) return provider.toUpperCase();
    if (method) return method.replace(/_/g, " ").toUpperCase();
    
    return "N/A";
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-96">
        <div className="w-10 h-10 border-2 border-white/20 border-t-white rounded-full animate-spin" />
      </div>
    );
  }

  if (!order) {
    return (
      <div className="flex flex-col items-center justify-center h-96">
        <AlertTriangle className="text-amber-400 mb-4" size={48} />
        <p className="text-white text-lg">Order not found</p>
        <Link href="/admin/orders" className="mt-4 text-white/60 hover:text-white">
          Back to Orders
        </Link>
      </div>
    );
  }

  const canCancel = ["PENDING", "PAID", "PACKING"].includes(order.status);
  const canRefund = ["DELIVERED", "COMPLETED"].includes(order.status) && order.refund_status !== "FULL"; // Refund only after delivery and not fully refunded
  const canReship = ["SHIPPED", "DELIVERED"].includes(order.status);
  const canPack = order.status === "PAID";
  const canShip = order.status === "PACKING";
  const canDeliver = order.status === "SHIPPED";
  
  // Calculate refundable balance
  const refundableBalance = order.total_amount - (order.refund_amount || 0);
  
  // Stuck payment detection - Only PENDING orders that need action
  // EXPIRED orders are already final, no action needed
  const isStuckPayment = order.status === "PENDING" && order.payment?.status === "PENDING";
  const canMarkAsPaid = order.status === "PENDING" && order.payment?.status === "PENDING";

  // Mark order as paid manually
  const handleMarkAsPaid = async () => {
    if (!actionReason.trim()) return;
    setActionLoading("mark_paid");
    try {
      const token = localStorage.getItem("auth_token");
      console.log("🔄 Sending mark as paid request:", {
        orderCode,
        status: "PAID",
        reason: actionReason
      });
      
      const response = await api.patch(`/admin/orders/${orderCode}/status`, {
        status: "PAID",
        reason: actionReason
      }, {
        headers: { Authorization: `Bearer ${token}` },
      });
      
      console.log("✅ Mark as paid response:", response.data);
      setShowModal(null);
      setActionReason("");
      loadOrder();
    } catch (error: any) {
      console.error("❌ Failed to mark as paid:", error);
      console.error("Error response:", error.response?.data);
      showErrorToast(`Gagal mengupdate status pembayaran: ${error.response?.data?.error || error.message}`);
    } finally {
      setActionLoading(null);
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <button
          onClick={() => router.back()}
          className="p-2 rounded-xl bg-white/10 text-white hover:bg-white/20 transition-colors"
        >
          <ArrowLeft size={20} />
        </button>
        <div className="flex-1">
          <h1 className="text-2xl font-bold text-white">{order.order_code}</h1>
          <p className="text-white/60 mt-1">Created {formatDate(order.created_at)}</p>
        </div>
      </div>

      {/* Stuck Payment Alert Banner - Only for PENDING orders */}
      {isStuckPayment && (
        <div className="flex items-center gap-3 p-4 rounded-xl bg-amber-500/10 border border-amber-500/30">
          <AlertTriangle className="text-amber-400 flex-shrink-0" size={20} />
          <div className="flex-1">
            <p className="text-amber-400 font-semibold text-sm">
              Payment Pending - Waiting for customer payment
            </p>
            <p className="text-white/60 text-xs mt-0.5">
              Verify if customer has paid before taking action
            </p>
          </div>
          <div className="flex gap-2">
            <a 
              href="https://dashboard.midtrans.com" 
              target="_blank" 
              rel="noopener noreferrer"
              className="px-3 py-1.5 rounded-lg bg-blue-500/20 text-blue-400 hover:bg-blue-500/30 transition-colors text-xs font-medium"
            >
              Check Midtrans
            </a>
            <button
              onClick={() => {
                const phone = order.customer_phone.replace(/^0/, '62');
                const message = encodeURIComponent(`Halo ${order.customer_name}, apakah Anda sudah melakukan pembayaran untuk order ${order.order_code} sebesar ${formatCurrency(order.total_amount)}?`);
                window.open(`https://wa.me/${phone}?text=${message}`, '_blank');
              }}
              className="px-3 py-1.5 rounded-lg bg-emerald-500/20 text-emerald-400 hover:bg-emerald-500/30 transition-colors text-xs font-medium"
            >
              WhatsApp
            </button>
          </div>
        </div>
      )}

      {/* EXPIRED Order Info Banner */}
      {order.status === "EXPIRED" && (
        <div className="flex items-center gap-3 p-4 rounded-xl bg-neutral-800 border border-white/10">
          <XCircle className="text-white/40 flex-shrink-0" size={20} />
          <div className="flex-1">
            <p className="text-white/60 font-semibold text-sm">
              Order Expired - Payment not completed
            </p>
            <p className="text-white/40 text-xs mt-0.5">
              This order has expired and is automatically closed. No action needed.
            </p>
          </div>
        </div>
      )}

      {/* Status Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {/* Order Status */}
        <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
          <div className="flex items-center gap-3 mb-4">
            <div className="p-2 rounded-xl bg-blue-500/20">
              <Package className="text-blue-400" size={20} />
            </div>
            <span className="text-white/60">Order Status</span>
          </div>
          <p className="text-2xl font-bold text-white">{order.status}</p>
          {order.refund_status && (
            <p className="text-amber-400 text-sm mt-2">Refund: {order.refund_status}</p>
          )}
        </div>

        {/* Payment Status */}
        <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
          <div className="flex items-center gap-3 mb-4">
            <div className="p-2 rounded-xl bg-emerald-500/20">
              <CreditCard className="text-emerald-400" size={20} />
            </div>
            <span className="text-white/60">Payment</span>
          </div>
          <p className="text-2xl font-bold text-white">{formatPaymentMethod(order.payment)}</p>
          <p className="text-white/60 text-sm mt-1">Status: {order.payment?.status || "N/A"}</p>
          {order.payment?.paid_at && (
            <p className="text-white/40 text-sm mt-2">Paid {formatDate(order.payment.paid_at)}</p>
          )}
        </div>

        {/* Shipment Status */}
        <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
          <div className="flex items-center gap-3 mb-4">
            <div className="p-2 rounded-xl bg-purple-500/20">
              <Truck className="text-purple-400" size={20} />
            </div>
            <span className="text-white/60">Shipment</span>
          </div>
          <p className="text-2xl font-bold text-white">{order.shipment?.status || order.status}</p>
          {order.resi && (
            <div className="mt-3 p-3 bg-white/5 rounded-lg border border-white/10">
              <div className="flex items-center justify-between mb-2">
                <p className="text-white/60 text-xs font-semibold">Nomor Resi</p>
                <button
                  onClick={() => {
                    navigator.clipboard.writeText(order.resi!);
                    showSuccessToast('✅ Resi berhasil di-copy!');
                  }}
                  className="px-2 py-1 rounded bg-purple-500/20 text-purple-400 hover:bg-purple-500/30 transition-colors text-xs font-medium"
                >
                  Copy
                </button>
              </div>
              <p className="text-white font-mono tracking-wider text-lg mb-2">{order.resi}</p>
              <p className="text-white/40 text-xs mb-2">
                Kurir: {order.shipment?.provider_name || 'N/A'}
              </p>
              {order.shipment?.shipped_at && (
                <p className="text-white/40 text-xs">
                  Dikirim: {formatDate(order.shipment.shipped_at)}
                </p>
              )}
            </div>
          )}
          {order.shipment?.tracking_number && !order.resi && (
            <p className="text-white/40 text-sm mt-2">{order.shipment.tracking_number}</p>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Order Items */}
        <div className="lg:col-span-2 bg-neutral-900 rounded-2xl border border-white/10 p-6">
          <h2 className="text-lg font-semibold text-white mb-6">Order Items</h2>
          <div className="space-y-4">
            {order.items.map((item, index) => (
              <div key={index} className="flex items-center gap-4 p-4 rounded-xl bg-white/5">
                <div className="w-16 h-16 rounded-lg bg-white/10 flex items-center justify-center">
                  {item.image_url ? (
                    <img src={item.image_url} alt={item.product_name} className="w-full h-full object-cover rounded-lg" />
                  ) : (
                    <Package className="text-white/40" size={24} />
                  )}
                </div>
                <div className="flex-1">
                  <p className="text-white font-medium">{item.product_name}</p>
                  <p className="text-white/60 text-sm">Qty: {item.quantity}</p>
                </div>
                <div className="text-right">
                  <p className="text-white font-medium">{formatCurrency(item.subtotal)}</p>
                  <p className="text-white/40 text-sm">{formatCurrency(item.price_per_unit)} each</p>
                </div>
              </div>
            ))}
          </div>

          {/* Order Summary */}
          <div className="mt-6 pt-6 border-t border-white/10 space-y-3">
            <div className="flex justify-between text-white/60">
              <span>Subtotal</span>
              <span>{formatCurrency(order.subtotal)}</span>
            </div>
            <div className="flex justify-between text-white/60">
              <span>Shipping</span>
              <span>{formatCurrency(order.shipping_cost)}</span>
            </div>
            {order.tax > 0 && (
              <div className="flex justify-between text-white/60">
                <span>Tax</span>
                <span>{formatCurrency(order.tax)}</span>
              </div>
            )}
            {order.discount > 0 && (
              <div className="flex justify-between text-emerald-400">
                <span>Discount</span>
                <span>-{formatCurrency(order.discount)}</span>
              </div>
            )}
            <div className="flex justify-between text-white text-lg font-bold pt-3 border-t border-white/10">
              <span>Total</span>
              <span>{formatCurrency(order.total_amount)}</span>
            </div>
          </div>
        </div>

        {/* Sidebar - Customer Info & Actions */}
        <div className="space-y-6">
          {/* Customer Info */}
          <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
            <h2 className="text-lg font-semibold text-white mb-6">Customer</h2>
            <div className="space-y-4">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-lg bg-white/10">
                  <User className="text-white/60" size={18} />
                </div>
                <div>
                  <p className="text-white/40 text-sm">Name</p>
                  <p className="text-white">{order.customer_name}</p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-lg bg-white/10">
                  <Mail className="text-white/60" size={18} />
                </div>
                <div>
                  <p className="text-white/40 text-sm">Email</p>
                  <p className="text-white">{order.customer_email}</p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-lg bg-white/10">
                  <Phone className="text-white/60" size={18} />
                </div>
                <div>
                  <p className="text-white/40 text-sm">Phone</p>
                  <p className="text-white">{order.customer_phone}</p>
                </div>
              </div>
            </div>
          </div>

          {/* Stuck Payment Actions Card */}
          {isStuckPayment && (
            <div className="bg-neutral-900 rounded-2xl border border-red-500/30 p-6">
              <h2 className="text-lg font-semibold text-red-400 mb-4 flex items-center gap-2">
                <AlertTriangle size={20} />
                Payment Actions
              </h2>
              
              {/* Verification Checklist */}
              <div className="bg-white/5 rounded-xl p-4 mb-4">
                <p className="text-white/80 text-sm font-semibold mb-3">Verification Steps:</p>
                <div className="space-y-2 text-xs text-white/60">
                  <div className="flex items-start gap-2">
                    <span className="text-blue-400">1.</span>
                    <span>Check Midtrans dashboard</span>
                  </div>
                  <div className="flex items-start gap-2">
                    <span className="text-blue-400">2.</span>
                    <span>Verify bank statement</span>
                  </div>
                  <div className="flex items-start gap-2">
                    <span className="text-blue-400">3.</span>
                    <span>Contact customer</span>
                  </div>
                  <div className="flex items-start gap-2">
                    <span className="text-blue-400">4.</span>
                    <span>Verify amount: <span className="text-emerald-400 font-semibold">{formatCurrency(order.total_amount)}</span></span>
                  </div>
                </div>
              </div>

              {/* Action Buttons */}
              <div className="space-y-3">
                <button
                  onClick={() => setShowModal("cancel")}
                  className="w-full flex items-center justify-center gap-2 px-4 py-3 rounded-xl bg-red-500 text-white hover:bg-red-600 transition-colors font-semibold"
                >
                  <XCircle size={18} />
                  Cancel Order
                </button>
                <button
                  onClick={() => setShowModal("mark_paid")}
                  className="w-full flex items-center justify-center gap-2 px-4 py-3 rounded-xl bg-amber-500 text-black hover:bg-amber-600 transition-colors font-semibold"
                >
                  <CheckCircle size={18} />
                  Confirm Payment
                </button>
                <p className="text-amber-400 text-xs text-center font-semibold">⚠️ Only if customer HAS PAID</p>
              </div>
            </div>
          )}

          {/* Order Processing Actions */}
          {!isStuckPayment && (
            <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
              <h2 className="text-lg font-semibold text-white mb-4">Order Actions</h2>
              <div className="space-y-3">
                {canMarkAsPaid && (
                  <button
                    onClick={() => setShowModal("mark_paid")}
                    className="w-full flex items-center justify-center gap-2 px-4 py-3 rounded-xl bg-amber-500/20 text-amber-400 hover:bg-amber-500/30 transition-colors font-medium border border-amber-500/30"
                  >
                    <CheckCircle size={18} />
                    Confirm Payment
                  </button>
                )}
                {canPack && (
                  <button
                    onClick={handlePackOrder}
                    disabled={actionLoading === "pack"}
                    className="w-full flex items-center justify-center gap-2 px-4 py-3 rounded-xl bg-emerald-500/20 text-emerald-400 hover:bg-emerald-500/30 transition-colors disabled:opacity-50 font-medium"
                  >
                    <PackageCheck size={18} />
                    {actionLoading === "pack" ? "Processing..." : "Proses Pesanan"}
                  </button>
                )}
                {canShip && (
                  <button
                    onClick={() => setShowModal("ship")}
                    className="w-full flex items-center justify-center gap-2 px-4 py-3 rounded-xl bg-blue-500/20 text-blue-400 hover:bg-blue-500/30 transition-colors font-medium"
                  >
                    <Send size={18} />
                    Kirim Pesanan
                  </button>
                )}
                {canDeliver && (
                  <button
                    onClick={handleDeliverOrder}
                    disabled={actionLoading === "deliver"}
                    className="w-full flex items-center justify-center gap-2 px-4 py-3 rounded-xl bg-green-500/20 text-green-400 hover:bg-green-500/30 transition-colors disabled:opacity-50 font-medium"
                  >
                    <CheckCircle size={18} />
                    {actionLoading === "deliver" ? "Processing..." : "Tandai Selesai"}
                  </button>
                )}
                
                {/* Admin Force Actions */}
                {(canCancel || canRefund || canReship) && (
                  <>
                    <div className="border-t border-white/10 my-3"></div>
                    <p className="text-white/40 text-xs mb-2">Admin Actions:</p>
                  </>
                )}
                {canCancel && !isStuckPayment && (
                  <button
                    onClick={() => setShowModal("cancel")}
                    className="w-full flex items-center justify-center gap-2 px-4 py-2 rounded-xl bg-red-500/20 text-red-400 hover:bg-red-500/30 transition-colors text-sm"
                  >
                    <XCircle size={16} />
                    Cancel
                  </button>
                )}
                {canRefund && (
                  <button
                    onClick={() => {
                      setShowModal("refund");
                      if (refundType === 'ITEM_ONLY') {
                        initializeItemSelection();
                      }
                    }}
                    className="w-full flex items-center justify-center gap-2 px-4 py-2 rounded-xl bg-amber-500/20 text-amber-400 hover:bg-amber-500/30 transition-colors text-sm"
                  >
                    <RefreshCcw size={16} />
                    Refund
                  </button>
                )}
                {canReship && (
                  <button
                    onClick={() => setShowModal("reship")}
                    className="w-full flex items-center justify-center gap-2 px-4 py-2 rounded-xl bg-purple-500/20 text-purple-400 hover:bg-purple-500/30 transition-colors text-sm"
                  >
                    <Truck size={16} />
                    Reship
                  </button>
                )}
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Refund History Section */}
      {refunds.length > 0 && (
        <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
          <h2 className="text-lg font-semibold text-white mb-6 flex items-center gap-2">
            <RefreshCcw size={20} />
            Refund History
          </h2>
          
          {refundsLoading ? (
            <div className="flex items-center justify-center py-8">
              <div className="w-8 h-8 border-2 border-white/20 border-t-white rounded-full animate-spin" />
            </div>
          ) : (
            <div className="space-y-4">
              {refunds.map((refund) => (
                <div key={refund.id} className="p-4 rounded-xl bg-white/5 border border-white/10">
                  <div className="flex items-start justify-between mb-3">
                    <div>
                      <div className="flex items-center gap-2 mb-1">
                        <span className="text-white font-semibold">{refund.refund_code}</span>
                        <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${
                          refund.status === 'COMPLETED' ? 'bg-emerald-500/20 text-emerald-400' :
                          refund.status === 'PROCESSING' ? 'bg-blue-500/20 text-blue-400' :
                          refund.status === 'FAILED' ? 'bg-red-500/20 text-red-400' :
                          'bg-amber-500/20 text-amber-400'
                        }`}>
                          {refund.status}
                        </span>
                        <span className="px-2 py-0.5 rounded-full text-xs font-medium bg-purple-500/20 text-purple-400">
                          {refund.refund_type}
                        </span>
                      </div>
                      <p className="text-white/60 text-sm">
                        Reason: {refund.reason}
                        {refund.reason_detail && ` - ${refund.reason_detail}`}
                      </p>
                    </div>
                    <div className="text-right">
                      <p className="text-white font-bold text-lg">{formatCurrency(refund.refund_amount)}</p>
                      {refund.gateway_refund_id && (
                        <p className="text-white/40 text-xs mt-1">
                          {refund.gateway_refund_id === 'MANUAL_REFUND' ? (
                            <span className="text-amber-400 font-semibold">MANUAL REFUND</span>
                          ) : (
                            `Gateway ID: ${refund.gateway_refund_id}`
                          )}
                        </p>
                      )}
                    </div>
                  </div>
                  
                  {/* Refund Breakdown */}
                  {(refund.items_refund > 0 || refund.shipping_refund > 0) && (
                    <div className="flex gap-4 mb-3 text-sm">
                      {refund.items_refund > 0 && (
                        <div className="text-white/60">
                          Items: <span className="text-white">{formatCurrency(refund.items_refund)}</span>
                        </div>
                      )}
                      {refund.shipping_refund > 0 && (
                        <div className="text-white/60">
                          Shipping: <span className="text-white">{formatCurrency(refund.shipping_refund)}</span>
                        </div>
                      )}
                    </div>
                  )}
                  
                  {/* Refund Items */}
                  {refund.items && refund.items.length > 0 && (
                    <div className="mb-3 p-3 rounded-lg bg-white/5">
                      <p className="text-white/60 text-xs font-semibold mb-2">Refunded Items:</p>
                      <div className="space-y-1">
                        {refund.items.map((item) => (
                          <div key={item.id} className="flex justify-between text-sm">
                            <span className="text-white/80">
                              {item.product_name} × {item.quantity}
                            </span>
                            <span className="text-white">{formatCurrency(item.subtotal)}</span>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                  
                  {/* Timestamps */}
                  <div className="flex flex-wrap gap-4 text-xs text-white/40 mb-3">
                    <div>Requested: {formatDate(refund.requested_at)}</div>
                    {refund.processed_at && <div>Processed: {formatDate(refund.processed_at)}</div>}
                    {refund.completed_at && <div>Completed: {formatDate(refund.completed_at)}</div>}
                  </div>
                  
                  {/* Status History */}
                  {refund.status_history && refund.status_history.length > 0 && (
                    <div className="pt-3 border-t border-white/10">
                      <p className="text-white/60 text-xs font-semibold mb-2">Status History:</p>
                      <div className="space-y-1">
                        {refund.status_history.map((history) => (
                          <div key={history.id} className="flex items-center gap-2 text-xs">
                            <span className="text-white/40">{formatDate(history.created_at)}</span>
                            <span className="text-white/60">→</span>
                            <span className="text-white">{history.new_status}</span>
                            <span className="text-white/40">by {history.actor}</span>
                            {history.reason && <span className="text-white/40">({history.reason})</span>}
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                  
                  {/* Retry Button for Failed Refunds */}
                  {refund.status === 'FAILED' && (
                    <button
                      onClick={() => handleRetryRefund(refund.id)}
                      disabled={actionLoading === `retry-${refund.id}`}
                      className="mt-3 w-full flex items-center justify-center gap-2 px-4 py-2 rounded-lg bg-amber-500/20 text-amber-400 hover:bg-amber-500/30 transition-colors text-sm font-medium disabled:opacity-50"
                    >
                      <RefreshCcw size={14} />
                      {actionLoading === `retry-${refund.id}` ? 'Retrying...' : 'Retry Refund'}
                    </button>
                  )}
                  
                  {/* Mark as Completed Button for Pending Refunds (Manual Processing) */}
                  {refund.status === 'PENDING' && (
                    <button
                      onClick={() => handleMarkRefundCompleted(refund.id)}
                      disabled={actionLoading === `complete-${refund.id}`}
                      className="mt-3 w-full flex items-center justify-center gap-2 px-4 py-2 rounded-lg bg-emerald-500/20 text-emerald-400 hover:bg-emerald-500/30 transition-colors text-sm font-medium disabled:opacity-50"
                    >
                      <CheckCircle size={14} />
                      {actionLoading === `complete-${refund.id}` ? 'Processing...' : 'Mark as Completed'}
                    </button>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Action Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4 overflow-y-auto">
          <div className={`bg-neutral-900 rounded-2xl border border-white/10 p-6 w-full my-8 ${
            showModal === "mark_paid" ? "max-w-2xl" : showModal === "refund" ? "max-w-3xl" : "max-w-md"
          }`}>
            <h3 className="text-xl font-bold text-white mb-4">
              {showModal === "cancel" && "Cancel Order"}
              {showModal === "refund" && "Process Refund"}
              {showModal === "reship" && "Create Reship"}
              {showModal === "ship" && "Kirim Pesanan"}
              {showModal === "mark_paid" && "⚠️ Confirm Payment Received"}
            </h3>
            
            {/* Refund Modal Content */}
            {showModal === "refund" ? (
              <div className="space-y-4">
                <p className="text-white/60 mb-4">
                  Process a refund for this order. Refundable balance: <span className="text-emerald-400 font-semibold">{formatCurrency(refundableBalance)}</span>
                </p>
                
                {/* Refund Type Selection */}
                <div>
                  <label className="text-white text-sm font-medium mb-2 block">
                    Refund Type <span className="text-red-400">*</span>
                  </label>
                  <div className="grid grid-cols-2 gap-2">
                    {(['FULL', 'PARTIAL', 'SHIPPING_ONLY', 'ITEM_ONLY'] as const).map((type) => (
                      <button
                        key={type}
                        onClick={() => {
                          setRefundType(type);
                          if (type === 'ITEM_ONLY') {
                            initializeItemSelection();
                          }
                        }}
                        className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                          refundType === type
                            ? 'bg-amber-500 text-black'
                            : 'bg-white/10 text-white hover:bg-white/20'
                        }`}
                      >
                        {type.replace(/_/g, ' ')}
                      </button>
                    ))}
                  </div>
                  <p className="text-white/40 text-xs mt-1">
                    {refundType === 'FULL' && 'Refund entire order amount'}
                    {refundType === 'PARTIAL' && 'Refund a specific amount'}
                    {refundType === 'SHIPPING_ONLY' && 'Refund only shipping cost'}
                    {refundType === 'ITEM_ONLY' && 'Refund specific items'}
                  </p>
                </div>
                
                {/* Partial Amount Input */}
                {refundType === 'PARTIAL' && (
                  <div>
                    <label className="text-white text-sm font-medium mb-2 block">
                      Refund Amount <span className="text-red-400">*</span>
                    </label>
                    <input
                      type="number"
                      value={refundAmount}
                      onChange={(e) => setRefundAmount(e.target.value)}
                      placeholder="Enter amount"
                      min="0"
                      max={refundableBalance}
                      step="1000"
                      className="w-full p-3 rounded-xl bg-white/5 border border-white/10 text-white placeholder-white/40 focus:outline-none focus:border-amber-500"
                    />
                    <p className="text-white/40 text-xs mt-1">
                      Maximum: {formatCurrency(refundableBalance)}
                    </p>
                  </div>
                )}
                
                {/* Item Selection */}
                {refundType === 'ITEM_ONLY' && (
                  <div>
                    <label className="text-white text-sm font-medium mb-2 block">
                      Select Items <span className="text-red-400">*</span>
                    </label>
                    <div className="space-y-2 max-h-60 overflow-y-auto">
                      {selectedItems.map((item) => (
                        <div key={item.order_item_id} className="p-3 rounded-lg bg-white/5 border border-white/10">
                          <div className="flex items-center justify-between mb-2">
                            <span className="text-white text-sm font-medium">{item.product_name}</span>
                            <span className="text-white/60 text-sm">{formatCurrency(item.price_per_unit)} each</span>
                          </div>
                          <div className="flex items-center gap-2">
                            <button
                              onClick={() => updateItemQuantity(item.order_item_id, item.quantity - 1)}
                              className="px-2 py-1 rounded bg-white/10 text-white hover:bg-white/20"
                            >
                              -
                            </button>
                            <input
                              type="number"
                              value={item.quantity}
                              onChange={(e) => updateItemQuantity(item.order_item_id, parseInt(e.target.value) || 0)}
                              min="0"
                              max={item.max_quantity}
                              className="w-20 p-2 rounded bg-white/5 border border-white/10 text-white text-center"
                            />
                            <button
                              onClick={() => updateItemQuantity(item.order_item_id, item.quantity + 1)}
                              className="px-2 py-1 rounded bg-white/10 text-white hover:bg-white/20"
                            >
                              +
                            </button>
                            <span className="text-white/60 text-sm">/ {item.max_quantity} max</span>
                            <span className="ml-auto text-white font-medium">
                              {formatCurrency(item.quantity * item.price_per_unit)}
                            </span>
                          </div>
                        </div>
                      ))}
                    </div>
                    <div className="mt-2 p-3 rounded-lg bg-amber-500/10 border border-amber-500/20">
                      <div className="flex justify-between text-sm">
                        <span className="text-white/80">Total Refund Amount:</span>
                        <span className="text-amber-400 font-bold">{formatCurrency(calculateItemRefundTotal())}</span>
                      </div>
                    </div>
                  </div>
                )}
                
                {/* Reason Selection */}
                <div>
                  <label className="text-white text-sm font-medium mb-2 block">
                    Reason <span className="text-red-400">*</span>
                  </label>
                  <select
                    value={refundReason}
                    onChange={(e) => setRefundReason(e.target.value)}
                    className="w-full p-3 rounded-xl bg-white/5 border border-white/10 text-white focus:outline-none focus:border-amber-500"
                  >
                    <option value="">Select reason...</option>
                    <option value="CUSTOMER_REQUEST">Customer Request</option>
                    <option value="DEFECTIVE_PRODUCT">Defective Product</option>
                    <option value="WRONG_ITEM">Wrong Item Sent</option>
                    <option value="NOT_AS_DESCRIBED">Not As Described</option>
                    <option value="DAMAGED_IN_SHIPPING">Damaged in Shipping</option>
                    <option value="LATE_DELIVERY">Late Delivery</option>
                    <option value="DUPLICATE_ORDER">Duplicate Order</option>
                    <option value="OTHER">Other</option>
                  </select>
                </div>
                
                {/* Reason Detail */}
                <div>
                  <label className="text-white text-sm font-medium mb-2 block">
                    Additional Details
                  </label>
                  <textarea
                    value={refundReasonDetail}
                    onChange={(e) => setRefundReasonDetail(e.target.value)}
                    placeholder="Enter additional details (optional)"
                    className="w-full p-3 rounded-xl bg-white/5 border border-white/10 text-white placeholder-white/40 focus:outline-none focus:border-amber-500 resize-none h-24"
                  />
                </div>
                
                {/* Error Message */}
                {refundError && (
                  <div className="p-3 rounded-lg bg-red-500/20 border border-red-500/30">
                    <p className="text-red-400 text-sm">{refundError}</p>
                  </div>
                )}
                
                {/* Refund Summary */}
                <div className="p-4 rounded-xl bg-white/5 border border-white/10">
                  <p className="text-white/60 text-sm font-semibold mb-2">Refund Summary:</p>
                  <div className="space-y-1 text-sm">
                    <div className="flex justify-between">
                      <span className="text-white/60">Type:</span>
                      <span className="text-white">{refundType.replace(/_/g, ' ')}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-white/60">Amount:</span>
                      <span className="text-white font-bold">
                        {refundType === 'FULL' && formatCurrency(order!.total_amount)}
                        {refundType === 'PARTIAL' && formatCurrency(parseFloat(refundAmount) || 0)}
                        {refundType === 'SHIPPING_ONLY' && formatCurrency(order!.shipping_cost)}
                        {refundType === 'ITEM_ONLY' && formatCurrency(calculateItemRefundTotal())}
                      </span>
                    </div>
                    {order!.payment && (
                      <div className="flex justify-between">
                        <span className="text-white/60">Payment Method:</span>
                        <span className="text-white">{formatPaymentMethod(order!.payment)}</span>
                      </div>
                    )}
                  </div>
                </div>
              </div>
            ) : showModal === "ship" ? (
              <div>
                {/* Auto-Generate Info Banner */}
                <div className="mb-4 p-4 rounded-xl bg-blue-500/10 border border-blue-500/30">
                  <div className="flex items-start gap-3">
                    <div className="p-2 rounded-lg bg-blue-500/20 flex-shrink-0">
                      <Send className="text-blue-400" size={18} />
                    </div>
                    <div>
                      <p className="text-blue-400 font-semibold text-sm mb-1">
                        ✨ Auto-Generate Resi dari Biteship
                      </p>
                      <p className="text-white/60 text-xs">
                        Klik tombol &quot;Generate dari Biteship&quot; untuk mendapatkan nomor resi otomatis dari API Biteship. 
                        Atau input manual jika sudah punya resi dari kurir.
                      </p>
                    </div>
                  </div>
                </div>

                <div className="mb-4">
                  <label className="text-white text-sm font-medium mb-2 block">
                    Nomor Resi <span className="text-red-400">*</span>
                  </label>
                  
                  {/* Generate Button */}
                  <button
                    onClick={handleGenerateResi}
                    disabled={actionLoading === "generate_resi" || !!resiInput}
                    className="w-full mb-3 flex items-center justify-center gap-2 px-4 py-3 rounded-xl bg-gradient-to-r from-blue-500 to-purple-500 text-white font-medium hover:from-blue-600 hover:to-purple-600 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {actionLoading === "generate_resi" ? (
                      <>
                        <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                        Generating dari Biteship...
                      </>
                    ) : (
                      <>
                        <Send size={16} />
                        Generate Resi dari Biteship
                      </>
                    )}
                  </button>
                  
                  <input
                    type="text"
                    value={resiInput}
                    onChange={(e) => {
                      const val = e.target.value.toUpperCase().replace(/[^A-Z0-9]/g, "");
                      setResiInput(val);
                      setResiError("");
                    }}
                    placeholder="Klik 'Generate dari Biteship' atau input manual"
                    className={`w-full p-4 rounded-xl bg-white/5 border text-white placeholder-white/40 focus:outline-none font-mono tracking-wider ${
                      resiError ? "border-red-500" : "border-white/10 focus:border-white/30"
                    }`}
                  />
                  {resiError && (
                    <p className="text-red-400 text-sm mt-2">{resiError}</p>
                  )}
                  <p className="text-white/40 text-xs mt-2">
                    {resiInput ? (
                      `✅ Resi: ${resiInput} - Klik "Confirm" untuk kirim pesanan`
                    ) : (
                      "💡 Klik tombol 'Generate dari Biteship' untuk mendapatkan resi otomatis"
                    )}
                  </p>
                </div>
              </div>
            ) : showModal === "mark_paid" ? (
              <div className="space-y-4">
                <p className="text-white/60 mb-4">
                  <span className="text-amber-400 font-semibold">
                    ONLY use this if customer has ACTUALLY PAID. This will manually update order status to PAID after you verify payment.
                  </span>
                </p>
                {/* 2 Column Layout for Warnings */}
                <div className="grid md:grid-cols-2 gap-4">
                  {/* CRITICAL WARNING */}
                  <div className="p-4 rounded-xl bg-red-500/20 border-2 border-red-500">
                    <p className="text-red-400 font-bold text-sm mb-2">🚨 CRITICAL WARNING</p>
                    <p className="text-white text-xs mb-2">
                      This action should <strong>ONLY</strong> be used when:
                    </p>
                    <ul className="text-white text-xs space-y-1 list-disc list-inside mb-2">
                      <li><strong>Customer HAS ACTUALLY PAID</strong> (verified via bank/Midtrans)</li>
                      <li>Payment amount matches order total: <strong>{formatCurrency(order.total_amount)}</strong></li>
                      <li>You have proof of payment (screenshot/bank statement)</li>
                    </ul>
                    <p className="text-amber-400 text-xs font-semibold">
                      ⚠️ If customer has NOT paid, use &quot;Cancel Order&quot; instead!
                    </p>
                  </div>

                  {/* Verification Checklist */}
                  <div className="p-4 rounded-xl bg-amber-500/10 border border-amber-500/20">
                    <p className="text-amber-400 text-sm font-medium mb-2">✅ Verification Checklist:</p>
                    <ul className="text-white/60 text-xs space-y-1.5">
                      <li className="flex items-start gap-2">
                        <span className="text-amber-400 mt-0.5">☐</span>
                        <span>Checked Midtrans dashboard - payment status is &quot;Settlement&quot;</span>
                      </li>
                      <li className="flex items-start gap-2">
                        <span className="text-amber-400 mt-0.5">☐</span>
                        <span>Verified bank statement shows incoming transfer</span>
                      </li>
                      <li className="flex items-start gap-2">
                        <span className="text-amber-400 mt-0.5">☐</span>
                        <span>Confirmed amount matches: {formatCurrency(order.total_amount)}</span>
                      </li>
                      <li className="flex items-start gap-2">
                        <span className="text-amber-400 mt-0.5">☐</span>
                        <span>Contacted customer and received payment proof</span>
                      </li>
                    </ul>
                  </div>
                </div>

                {/* Verification Details Input */}
                <div>
                  <label className="text-white text-sm font-medium mb-2 block">
                    Verification Details <span className="text-red-400">*</span>
                  </label>
                  <textarea
                    value={actionReason}
                    onChange={(e) => setActionReason(e.target.value)}
                    placeholder="REQUIRED: Document your verification (e.g., 'Verified via BCA internet banking. Transfer received on 2026-01-13 at 14:30 WIB. Amount: Rp 914,000. Sender name: Sebastian Alexander. Screenshot saved.')"
                    className="w-full p-3 rounded-xl bg-white/5 border border-amber-500/50 text-white placeholder-white/40 focus:outline-none focus:border-amber-500 resize-none h-24 text-sm"
                  />
                  <p className="text-white/40 text-xs mt-1">
                    Include: verification method, date/time, amount, sender name, and where proof is stored
                  </p>
                </div>
              </div>
            ) : (
              <div>
                <p className="text-white/60 mb-4">
                  {showModal === "cancel" && "This will cancel the order and restore stock."}
                  {showModal === "reship" && "This will create a replacement shipment."}
                </p>
                <textarea
                  value={actionReason}
                  onChange={(e) => setActionReason(e.target.value)}
                  placeholder="Enter reason..."
                  className="w-full p-4 rounded-xl bg-white/5 border border-white/10 text-white placeholder-white/40 focus:outline-none focus:border-white/30 resize-none h-32"
                />
              </div>
            )}
            
            <div className="flex gap-3 mt-6">
              <button
                onClick={() => {
                  setShowModal(null);
                  setActionReason("");
                  setResiInput("");
                  setResiError("");
                  if (showModal === "refund") {
                    resetRefundForm();
                  }
                }}
                className="flex-1 px-4 py-3 rounded-xl bg-white/10 text-white hover:bg-white/20 transition-colors"
              >
                Batal
              </button>
              <button
                onClick={() => {
                  if (showModal === "cancel") handleForceCancel();
                  if (showModal === "refund") handleForceRefund();
                  if (showModal === "reship") handleForceReship();
                  if (showModal === "ship") handleShipOrder();
                  if (showModal === "mark_paid") handleMarkAsPaid();
                }}
                disabled={
                  (showModal === "ship" ? resiInput.trim().length < 8 : 
                   showModal === "refund" ? !refundReason.trim() :
                   !actionReason.trim()) || 
                  actionLoading !== null
                }
                className={`flex-1 px-4 py-3 rounded-xl font-semibold transition-colors disabled:opacity-50 ${
                  showModal === "cancel"
                    ? "bg-red-500 text-white hover:bg-red-600"
                    : showModal === "refund"
                    ? "bg-amber-500 text-black hover:bg-amber-600"
                    : showModal === "ship"
                    ? "bg-blue-500 text-white hover:bg-blue-600"
                    : showModal === "mark_paid"
                    ? "bg-amber-500 text-black hover:bg-amber-600"
                    : "bg-purple-500 text-white hover:bg-purple-600"
                }`}
              >
                {actionLoading ? "Processing..." : 
                 showModal === "mark_paid" ? "✅ Yes, Customer Has Paid" : 
                 showModal === "refund" ? "Process Refund" :
                 "Confirm"}
              </button>
            </div>
          </div>
        </div>
      )}
      
      {/* Toast Notification */}
      {showToast && (
        <div className="fixed bottom-4 right-4 z-50 animate-slide-up">
          <div className={`px-6 py-4 rounded-xl shadow-lg flex items-center gap-3 ${
            toastType === 'success' 
              ? 'bg-emerald-500 text-white' 
              : 'bg-red-500 text-white'
          }`}>
            {toastType === 'success' ? (
              <CheckCircle size={24} />
            ) : (
              <XCircle size={24} />
            )}
            <span className="font-medium">{toastMessage}</span>
          </div>
        </div>
      )}
      
      {/* Confirm Dialog */}
      {showConfirm && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6 max-w-md w-full">
            <h3 className="text-xl font-bold text-white mb-4">{confirmConfig.title}</h3>
            <p className="text-white/80 mb-6">{confirmConfig.message}</p>
            <div className="flex gap-3">
              <button
                onClick={() => setShowConfirm(false)}
                className="flex-1 px-4 py-3 rounded-xl bg-white/10 text-white hover:bg-white/20 transition-colors"
              >
                Batal
              </button>
              <button
                onClick={confirmConfig.onConfirm}
                className={`flex-1 px-4 py-3 rounded-xl font-semibold transition-colors ${
                  confirmConfig.variant === 'danger'
                    ? 'bg-red-500 text-white hover:bg-red-600'
                    : confirmConfig.variant === 'warning'
                    ? 'bg-amber-500 text-black hover:bg-amber-600'
                    : 'bg-blue-500 text-white hover:bg-blue-600'
                }`}
              >
                Ya, lanjutkan
              </button>
            </div>
          </div>
        </div>
      )}
      
      {/* Note Input Modal */}
      {showNoteModal && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6 max-w-lg w-full">
            <h3 className="text-xl font-bold text-white mb-4">{noteModalConfig.title}</h3>
            <p className="text-white/80 mb-4">{noteModalConfig.message}</p>
            <textarea
              value={noteInput}
              onChange={(e) => setNoteInput(e.target.value)}
              placeholder={noteModalConfig.placeholder}
              className="w-full p-4 rounded-xl bg-white/5 border border-white/10 text-white placeholder-white/40 focus:outline-none focus:border-amber-500 resize-none h-32 mb-4"
              autoFocus
            />
            <div className="flex gap-3">
              <button
                onClick={() => {
                  setShowNoteModal(false);
                  setNoteInput('');
                }}
                className="flex-1 px-4 py-3 rounded-xl bg-white/10 text-white hover:bg-white/20 transition-colors"
              >
                Batal
              </button>
              <button
                onClick={() => noteModalConfig.onConfirm(noteInput)}
                disabled={!noteInput.trim()}
                className="flex-1 px-4 py-3 rounded-xl bg-amber-500 text-black hover:bg-amber-600 transition-colors font-semibold disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Konfirmasi
              </button>
            </div>
          </div>
        </div>
      )}
      
      {/* Confirm Dialog */}
      <ConfirmDialog
        isOpen={showConfirm}
        title={confirmConfig.title}
        message={confirmConfig.message}
        confirmText="Ya, Lanjutkan"
        cancelText="Batal"
        variant={confirmConfig.variant}
        onConfirm={confirmConfig.onConfirm}
        onCancel={() => setShowConfirm(false)}
      />
    </div>
  );
}
