"use client";

import { useState, useRef, useCallback, useEffect } from "react";
import { useRouter } from "next/navigation";
import Image from "next/image";
import Link from "next/link";
import { AnimatePresence, motion } from "framer-motion";
import { useCart } from "@/context/CartContext";
import { useAuth } from "@/context/AuthContext";
import { useToast } from "@/components/ui/Toast";
import api from "@/lib/api";
import { LoadingOverlay } from "@/components/ui/LoadingSpinner";

// ============================================
// BITESHIP AREA TYPES - Direct from API
// ============================================
interface BiteshipArea {
  area_id: string;                               // area_id - PRIMARY identifier
  name: string;                                  // Full area name from Biteship
  postal_code: string;                           // Postal code as string
  province: string;                              // Province name
  city: string;                                  // City name
  district: string;                              // District name
}

// Shipping rate from Biteship /v1/rates/couriers
interface ShippingRate {
  courier_code: string;
  courier_name: string;
  courier_service_code: string;
  courier_service_name: string;
  description: string;
  duration: string;
  price: number;
  type: string;
}

// Address stored with Biteship area_id ONLY
// NO province_id, city_id, district_id
interface ShippingAddress {
  recipient_name: string;
  phone: string;
  area_id: string;        // PRIMARY location identifier
  area_name: string;      // Full display name from Biteship (unchanged)
  postal_code: string;
  full_address: string;   // Detailed address (street, house number, RT/RW)
}

// Courier logos
const COURIER_LOGOS: Record<string, string> = {
  jne: "/images/couriers/jne.png",
  jnt: "/images/couriers/jnt.png",
  sicepat: "/images/couriers/sicepat.jpg",
  anteraja: "/images/couriers/anteraja.jpg",
  pos: "/images/couriers/pos.svg",
  tiki: "/images/couriers/tiki.jpg",
  ninja: "/images/couriers/ninja.png",
  lion: "/images/couriers/lion.png",
  rpx: "/images/couriers/rpx.png",
  sap: "/images/couriers/sap.png",
  idexpress: "/images/couriers/idexpress.png",
  ide: "/images/couriers/idexpress.png",
};

// Payment methods
const PAYMENT_METHODS = {
  eWallet: [
    { id: "gopay", name: "GoPay", logo: "/images/payments/gopay.png", desc: "Bayar langsung dari GoPay" },
    { id: "qris", name: "QRIS", logo: "/images/payments/qris.png", desc: "GoPay, OVO, Dana, dll" },
  ],
  virtualAccount: [
    { id: "bca_va", name: "BCA Virtual Account", logo: "/images/banks/bca.png" },
    { id: "bri_va", name: "BRI Virtual Account", logo: "/images/banks/bri.png" },
    { id: "mandiri_va", name: "Mandiri Virtual Account", logo: "/images/banks/mandiri.png" },
    { id: "permata_va", name: "Permata Virtual Account", logo: "/images/banks/permata.png" },
    { id: "bni_va", name: "BNI Virtual Account", logo: "/images/banks/bni.png" },
  ],
};

// Shipping type styles
const SHIPPING_TYPE_STYLES: Record<string, { bg: string; text: string; label: string }> = {
  instant: { bg: "bg-rose-50", text: "text-rose-700", label: "Instant" },
  same_day: { bg: "bg-amber-50", text: "text-amber-700", label: "Same Day" },
  express: { bg: "bg-blue-50", text: "text-blue-700", label: "Express" },
  regular: { bg: "bg-stone-100", text: "text-stone-600", label: "Regular" },
  economy: { bg: "bg-gray-100", text: "text-gray-600", label: "Economy" },
};

const ShippingTypeBadge = ({ type }: { type: string }) => {
  const style = SHIPPING_TYPE_STYLES[type] || SHIPPING_TYPE_STYLES.regular;
  return <span className={`px-2 py-0.5 rounded text-xs font-medium ${style.bg} ${style.text}`}>{style.label}</span>;
};

const CourierLogo = ({ code }: { code: string }) => {
  const logoPath = COURIER_LOGOS[code.toLowerCase()];
  if (logoPath) {
    return (
      <div className="w-full h-full flex items-center justify-center bg-white rounded overflow-hidden">
        <img src={logoPath} alt={code.toUpperCase()} className="max-w-full max-h-full object-contain" />
      </div>
    );
  }
  return (
    <div className="w-full h-full bg-secondary rounded flex items-center justify-center">
      <span className="text-xs font-bold text-muted">{code.toUpperCase()}</span>
    </div>
  );
};

export default function CheckoutPage() {
  const router = useRouter();
  const { cart, getTotalPrice, clearCart, syncCartToBackend, validateCart, refreshCart } = useCart();
  const { user, isAuthenticated } = useAuth();
  const { showToast } = useToast();

  // UI States
  const [showAddressPanel, setShowAddressPanel] = useState(false);
  const [showAddressList, setShowAddressList] = useState(false);
  const [showAreaSearch, setShowAreaSearch] = useState(false);
  const [showShippingPanel, setShowShippingPanel] = useState(false);
  const [showPaymentPanel, setShowPaymentPanel] = useState(false);
  const [loading, setLoading] = useState(false);
  const [loadingMessage, setLoadingMessage] = useState("");

  // ============================================
  // SAVED ADDRESSES STATE
  // ============================================
  interface SavedAddress {
    id: number;
    label: string;
    recipient_name: string;
    phone: string;
    province_name: string;
    city_name: string;
    district: string;
    subdistrict: string;
    postal_code: string;
    full_address: string;
    is_default: boolean;
    area_id?: string;
    area_name?: string;
  }
  const [savedAddresses, setSavedAddresses] = useState<SavedAddress[]>([]);
  const [loadingAddresses, setLoadingAddresses] = useState(false);

  // ============================================
  // BITESHIP AREA AUTOCOMPLETE STATE
  // ============================================
  const [areaQuery, setAreaQuery] = useState("");
  const [areaResults, setAreaResults] = useState<BiteshipArea[]>([]);
  const [loadingAreas, setLoadingAreas] = useState(false);
  const [selectedArea, setSelectedArea] = useState<BiteshipArea | null>(null);

  // Shipping Rates from Biteship
  const [shippingRates, setShippingRates] = useState<ShippingRate[]>([]);
  const [groupedRates, setGroupedRates] = useState<Record<string, ShippingRate[]>>({});
  const [selectedRate, setSelectedRate] = useState<ShippingRate | null>(null);
  const [loadingRates, setLoadingRates] = useState(false);
  const [shippingMeta, setShippingMeta] = useState({ totalWeight: 0, totalWeightKg: "0 g", originCity: "Semarang" });

  // Payment
  const [selectedPayment, setSelectedPayment] = useState<string | null>(null);

  // Address State - Biteship native (area_id only, NO province/city/district IDs)
  const [addressSaved, setAddressSaved] = useState(false);
  const [guestEmail, setGuestEmail] = useState("");
  const [address, setAddress] = useState<ShippingAddress>({
    recipient_name: "",
    phone: "",
    area_id: "",
    area_name: "",
    postal_code: "",
    full_address: "",
  });

  const isProcessingRef = useRef(false);
  
  // ============================================
  // AUTOCOMPLETE REFS - For debounce & cancellation
  // ============================================
  const searchTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const abortControllerRef = useRef<AbortController | null>(null);

  // ============================================
  // LOAD SAVED ADDRESSES & DEFAULT ADDRESS
  // ============================================
  const loadSavedAddresses = useCallback(async () => {
    if (!isAuthenticated) return;
    setLoadingAddresses(true);
    try {
      const res = await api.get("/user/addresses");
      const addresses: SavedAddress[] = res.data.addresses || [];
      setSavedAddresses(addresses);
      
      // Auto-select default address if exists and no address selected yet
      const defaultAddr = addresses.find(a => a.is_default);
      if (defaultAddr && !addressSaved) {
        selectSavedAddress(defaultAddr);
      }
    } catch (err) {
      console.error("Failed to load addresses:", err);
    } finally {
      setLoadingAddresses(false);
    }
  }, [isAuthenticated, addressSaved]);

  // Select a saved address
  const selectSavedAddress = async (addr: SavedAddress) => {
    const areaName = addr.area_name || `${addr.subdistrict}, ${addr.district}, ${addr.city_name}, ${addr.province_name}`;
    setAddress({
      recipient_name: addr.recipient_name,
      phone: addr.phone,
      area_id: addr.area_id || "",
      area_name: areaName,
      postal_code: addr.postal_code,
      full_address: addr.full_address,
    });
    setAddressSaved(true);
    setShowAddressList(false);
    
    // Load shipping rates
    if (addr.area_id || addr.postal_code) {
      await loadShippingRates(addr.area_id || "", addr.postal_code);
    }
  };

  // Load addresses on mount
  useEffect(() => {
    if (isAuthenticated) {
      loadSavedAddresses();
    }
  }, [isAuthenticated, loadSavedAddresses]);

  // Pre-fill user data (only if no default address)
  useEffect(() => {
    if (isAuthenticated && user && !addressSaved) {
      setAddress(prev => ({ ...prev, recipient_name: prev.recipient_name || user.first_name || "", phone: prev.phone || user.phone || "" }));
    }
  }, [isAuthenticated, user, addressSaved]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (searchTimeoutRef.current) clearTimeout(searchTimeoutRef.current);
      if (abortControllerRef.current) abortControllerRef.current.abort();
    };
  }, []);

  // ============================================
  // AUTO-REFRESH CART & SHIPPING RATES
  // E-commerce standard: validate every 10 seconds on checkout
  // ============================================
  useEffect(() => {
    if (!isAuthenticated) return;

    const interval = setInterval(async () => {
      console.log("🔄 Auto-validating cart on checkout...");
      const validation = await validateCart();
      
      if (validation && !validation.valid && validation.changes.length > 0) {
        // Refresh cart to get latest data
        await refreshCart();
        
        // Show notification for each change
        validation.changes.forEach((change: any) => {
          if (change.change_type === "price_changed") {
            showToast(
              `${change.product_name}: Harga berubah dari Rp ${change.old_price?.toLocaleString()} ke Rp ${change.new_price?.toLocaleString()}`,
              "warning"
            );
          } else if (change.change_type === "stock_insufficient") {
            showToast(
              `${change.product_name}: Hanya ${change.current_stock} items tersedia`,
              "warning"
            );
          } else if (change.change_type === "product_unavailable") {
            showToast(
              `${change.product_name}: Product tidak tersedia lagi`,
              "error"
            );
          }
        });

        // Reload shipping rates if address is set
        if (address.area_id && shippingRates.length > 0) {
          console.log("🔄 Reloading shipping rates due to cart changes...");
          await loadShippingRates(address.area_id, address.postal_code);
        }
      }
    }, 10000); // 10 seconds - faster on checkout page

    return () => clearInterval(interval);
  }, [isAuthenticated, validateCart, refreshCart, showToast, address.area_id, address.postal_code, shippingRates.length]);

  // ============================================
  // BITESHIP AREA AUTOCOMPLETE IMPLEMENTATION
  // GET /v1/maps/areas?input={query}
  // ============================================
  const searchAreas = useCallback(async (query: string, signal: AbortSignal) => {
    // Minimum 3 characters required
    if (query.length < 3) {
      setAreaResults([]);
      setLoadingAreas(false);
      return;
    }

    try {
      const res = await api.get(`/shipping/areas?q=${encodeURIComponent(query)}`, { signal });
      // Render results EXACTLY as returned by Biteship - DO NOT modify
      setAreaResults(res.data.areas || []);
    } catch (err: unknown) {
      // Ignore abort errors
      if (err instanceof Error && err.name === 'AbortError') return;
      if (err instanceof Error && err.name === 'CanceledError') return;
      console.error("Area search failed:", err);
      setAreaResults([]);
    } finally {
      setLoadingAreas(false);
    }
  }, []);

  // Autocomplete handler with debounce (300ms) and request cancellation
  const handleAreaQueryChange = useCallback((query: string) => {
    setAreaQuery(query);
    
    // Clear previous timeout
    if (searchTimeoutRef.current) {
      clearTimeout(searchTimeoutRef.current);
    }
    
    // Cancel previous request
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }

    // Clear results if query too short
    if (query.length < 3) {
      setAreaResults([]);
      setLoadingAreas(false);
      return;
    }

    // Show loading immediately
    setLoadingAreas(true);

    // Debounce: 300ms delay before API call
    searchTimeoutRef.current = setTimeout(() => {
      // Create new AbortController for this request
      abortControllerRef.current = new AbortController();
      searchAreas(query, abortControllerRef.current.signal);
    }, 300);
  }, [searchAreas]);

  // Select area from autocomplete results
  // Store ONLY: area_id, postal_code, area_name (full name from Biteship)
  const selectArea = (area: BiteshipArea) => {
    setSelectedArea(area);
    setAddress(prev => ({
      ...prev,
      area_id: area.area_id,
      area_name: area.name,
      postal_code: area.postal_code,
    }));
    setShowAreaSearch(false);
    setAreaQuery("");
    setAreaResults([]);
  };

  // ============================================
  // BITESHIP SHIPPING RATES
  // ============================================
  const loadShippingRates = async (destinationAreaId: string, postalCode: string) => {
    setLoadingRates(true);
    setShippingRates([]);
    setGroupedRates({});

    try {
      // DON'T sync cart to backend here - it causes duplicate items!
      // Cart is already synced when items are added via addToCart()
      // Syncing here will re-add items from localStorage which may not have variant_id
      
      // Just refresh cart from backend to get latest data
      await refreshCart();
      // Small delay to ensure database commit
      await new Promise(resolve => setTimeout(resolve, 100));
      
      const res = await api.post("/shipping/rates", {
        destination_area_id: destinationAreaId,
        destination_postal_code: postalCode,
      });

      const rates: ShippingRate[] = res.data.rates || [];
      setShippingRates(rates);

      // Group by type
      const grouped: Record<string, ShippingRate[]> = {};
      rates.forEach(rate => {
        const type = rate.type || "regular";
        if (!grouped[type]) grouped[type] = [];
        grouped[type].push(rate);
      });
      setGroupedRates(grouped);

      setShippingMeta({
        totalWeight: res.data.total_weight || 0,
        totalWeightKg: res.data.total_weight_kg || "0 g",
        originCity: res.data.origin_city || "Semarang",
      });
    } catch (err) {
      console.error("Failed to load shipping rates:", err);
      showToast("Gagal memuat ongkir", "error");
    } finally {
      setLoadingRates(false);
    }
  };

  // ============================================
  // ADDRESS HANDLERS
  // ============================================
  const handleAddressChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const { name, value } = e.target;
    setAddress(prev => ({ ...prev, [name]: value }));
  };

  const saveAddress = async () => {
    // Validation
    if (!address.recipient_name.trim()) {
      showToast("Nama penerima wajib diisi", "error");
      return;
    }
    if (!address.phone.trim()) {
      showToast("Nomor telepon wajib diisi", "error");
      return;
    }
    if (!address.area_id) {
      showToast("Pilih area pengiriman", "error");
      return;
    }
    if (!address.full_address.trim()) {
      showToast("Alamat lengkap wajib diisi", "error");
      return;
    }
    if (!isAuthenticated && !guestEmail.trim()) {
      showToast("Email wajib diisi", "error");
      return;
    }
    if (!isAuthenticated && guestEmail) {
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      if (!emailRegex.test(guestEmail)) {
        showToast("Format email tidak valid", "error");
        return;
      }
    }

    // Validation passed - save address
    setAddressSaved(true);
    setShowAddressPanel(false);

    // Load shipping rates using Biteship postal_code
    // Note: loadShippingRates will sync cart automatically
    await loadShippingRates(address.area_id, address.postal_code);
  };

  const selectShipping = (rate: ShippingRate) => {
    setSelectedRate(rate);
    setShowShippingPanel(false);
  };

  // ============================================
  // CHECKOUT & PAYMENT
  // ============================================
  const handlePayment = useCallback(async () => {
    if (isProcessingRef.current || loading) return;
    if (!addressSaved) {
      showToast("Tambahkan alamat pengiriman", "error");
      return;
    }
    if (!selectedRate) {
      showToast("Pilih metode pengiriman", "error");
      return;
    }
    if (!selectedPayment) {
      showToast("Pilih metode pembayaran", "error");
      return;
    }

    isProcessingRef.current = true;
    setLoading(true);
    setLoadingMessage("Memproses pesanan...");

    try {
      // Cart already synced when loading shipping rates, no need to sync again
      // syncCartToBackend clears cart first which causes race condition

      const headers: Record<string, string> = {};
      const token = localStorage.getItem("auth_token");
      if (token) headers["Authorization"] = `Bearer ${token}`;

      const customerEmail = user?.email || guestEmail;
      if (!customerEmail) {
        showToast("Email diperlukan", "error");
        setLoading(false);
        isProcessingRef.current = false;
        return;
      }

      // Create order with Biteship shipping
      const orderRes = await api.post("/checkout/shipping", {
        customer_name: address.recipient_name,
        customer_email: customerEmail,
        customer_phone: address.phone,
        courier_code: selectedRate.courier_code,
        courier_service_code: selectedRate.courier_service_code,
        shipping_address: {
          recipient_name: address.recipient_name,
          phone: address.phone,
          area_id: address.area_id,
          area_name: address.area_name,
          postal_code: address.postal_code,
          full_address: address.full_address,
        },
      }, { headers });

      // Create payment
      setLoadingMessage("Membuat pembayaran...");
      try {
        await api.post("/payments/core/create", {
          order_id: orderRes.data.order_id,
          payment_method: selectedPayment,
        });

        // Clear cart from context
        clearCart();
        
        setLoading(false);
        isProcessingRef.current = false;
        
        // Use replace instead of push to prevent back navigation
        router.replace(`/checkout/payment/detail?order_id=${orderRes.data.order_id}`);
      } catch (paymentErr: unknown) {
        // Payment creation failed, but order is already created
        // Redirect to payment selection page where user can retry
        const axiosError = paymentErr as { response?: { data?: { message?: string } } };
        const errorMessage = axiosError.response?.data?.message || "Gagal membuat pembayaran";
        
        console.log("⚠️ Payment creation failed:", errorMessage, "- redirecting to payment page");
        showToast(`${errorMessage}. Silakan pilih metode pembayaran lagi.`, "warning");
        
        setLoading(false);
        isProcessingRef.current = false;
        
        // Redirect to payment selection page
        router.replace(`/checkout/payment?order_id=${orderRes.data.order_id}`);
      }
    } catch (err: unknown) {
      const axiosError = err as { response?: { data?: { message?: string } } };
      const errorMessage = axiosError.response?.data?.message || "Checkout gagal";
      showToast(errorMessage, "error");
      setLoading(false);
      isProcessingRef.current = false;
    }
  }, [address, selectedRate, selectedPayment, addressSaved, loading, user, guestEmail, clearCart, router, showToast]);

  // Calculations
  const subtotal = getTotalPrice();
  const shippingCost = selectedRate?.price || 0;
  const total = subtotal + shippingCost;

  // Empty cart view
  if (cart.length === 0) {
    console.log("🛒 Checkout: Cart is empty");
    console.log("🛒 isAuthenticated:", isAuthenticated);
    console.log("🛒 user:", user);
    console.log("🛒 cart state:", cart);
    
    return (
      <div className="min-h-screen flex items-center justify-center bg-secondary">
        <div className="text-center">
          <div className="w-20 h-20 mx-auto mb-4 bg-secondary rounded-full flex items-center justify-center">
            <svg className="w-10 h-10 text-muted" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z" />
            </svg>
          </div>
          <h1 className="text-xl font-semibold text-primary mb-2">Keranjang Kosong</h1>
          <p className="text-muted mb-6">Belum ada produk di keranjang</p>
          <Link href="/" className="inline-flex items-center gap-2 px-6 py-3 bg-primary text-white rounded-full hover:bg-gray-800 transition">
            Mulai Belanja
          </Link>
        </div>
      </div>
    );
  }

  // ============================================
  // MAIN RENDER
  // ============================================
  return (
    <>
      <AnimatePresence>{loading && <LoadingOverlay message={loadingMessage} />}</AnimatePresence>

      <div className="min-h-screen bg-secondary">
        {/* Header */}
        <div className="bg-white border-b border-accent shadow-sm sticky top-0 z-40">
          <div className="max-w-7xl mx-auto px-4 lg:px-8">
            <div className="h-16 flex items-center justify-between">
              <div className="flex items-center gap-6">
                <Link href="/" className="text-2xl font-serif tracking-wider text-primary">ZAVERA</Link>
                <div className="hidden md:flex items-center gap-3 text-muted">
                  <div className="w-px h-5 bg-accent" />
                  <span className="text-base font-medium text-primary">Checkout</span>
                </div>
              </div>
              <div className="flex items-center gap-2 text-sm text-muted">
                <svg className="w-5 h-5 text-primary" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clipRule="evenodd" />
                </svg>
                <span className="hidden sm:inline font-medium">Transaksi Aman</span>
              </div>
            </div>
          </div>
        </div>

        <div className="max-w-7xl mx-auto px-4 lg:px-8 py-6">
          <div className="flex flex-col lg:flex-row gap-6">
            {/* LEFT COLUMN */}
            <div className="flex-1 space-y-4">
              {/* Address Card */}
              <div className="bg-white rounded-xl shadow-sm border border-accent overflow-hidden">
                <div className="px-5 py-4 border-b border-accent flex items-center justify-between">
                  <h2 className="font-medium text-primary flex items-center gap-3">
                    <span className="w-7 h-7 rounded-full bg-primary text-white text-xs flex items-center justify-center">
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                      </svg>
                    </span>
                    <span className="uppercase tracking-wide text-sm">Alamat Pengiriman</span>
                  </h2>
                  <button onClick={() => isAuthenticated && savedAddresses.length > 0 ? setShowAddressList(true) : setShowAddressPanel(true)}
                    className="text-sm font-medium text-primary hover:text-primary/80 px-4 py-2 rounded-xl hover:bg-primary/5 transition-all">
                    {addressSaved ? "Ganti" : (isAuthenticated && savedAddresses.length > 0 ? "Pilih" : "Tambah")}
                  </button>
                </div>
                <div className="p-5">
                  {loadingAddresses ? (
                    <div className="flex items-center gap-3 py-2">
                      <div className="animate-spin w-5 h-5 border-2 border-primary border-t-transparent rounded-full" />
                      <span className="text-sm text-muted">Memuat alamat...</span>
                    </div>
                  ) : addressSaved ? (
                    <div className="space-y-1.5">
                      <p className="font-semibold text-primary">{address.recipient_name} <span className="font-normal text-muted">({address.phone})</span></p>
                      <p className="text-sm text-muted leading-relaxed">{address.full_address}</p>
                      <p className="text-sm text-primary font-medium">{address.area_name} - {address.postal_code}</p>
                    </div>
                  ) : (
                    <div className="flex items-center gap-4 text-muted py-2">
                      <div className="w-11 h-11 rounded-full bg-secondary flex items-center justify-center">
                        <svg className="w-5 h-5 text-muted" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M12 4v16m8-8H4" />
                        </svg>
                      </div>
                      <p className="text-sm">{isAuthenticated && savedAddresses.length > 0 ? "Pilih alamat pengiriman" : "Tambahkan alamat pengiriman untuk melanjutkan"}</p>
                    </div>
                  )}
                </div>
              </div>

              {/* Products Card */}
              <div className="bg-white rounded-xl shadow-sm border border-accent overflow-hidden">
                <div className="px-5 py-4 border-b border-accent flex items-center gap-3">
                  <div className="w-7 h-7 rounded-xl bg-primary flex items-center justify-center">
                    <svg className="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" />
                    </svg>
                  </div>
                  <h2 className="font-medium text-primary">ZAVERA Store</h2>
                  <span className="text-xs text-muted bg-secondary px-2.5 py-1 rounded-full">{cart.length} Barang</span>
                </div>

                <div className="divide-y divide-accent">
                  {cart.map((item) => (
                    <div key={`${item.id}-${item.selectedSize}`} className="p-5 flex gap-4">
                      <div className="w-20 h-20 bg-secondary rounded-xl overflow-hidden relative flex-shrink-0 border border-accent">
                        <Image src={item.image_url || '/placeholder.jpg'} alt={item.name} fill className="object-cover" />
                      </div>
                      <div className="flex-1 min-w-0">
                        <p className="font-medium text-primary line-clamp-2 text-sm">{item.name}</p>
                        <div className="mt-2 flex items-center gap-2 text-xs text-muted">
                          <span className="px-2 py-0.5 bg-secondary rounded">Ukuran: {item.selectedSize}</span>
                          <span className="text-accent">|</span>
                          <span>Qty: {item.quantity}</span>
                        </div>
                        <p className="mt-2 font-semibold text-primary">Rp {(item.price * item.quantity).toLocaleString("id-ID")}</p>
                      </div>
                    </div>
                  ))}
                </div>

                {/* Shipping Selection */}
                <div className="border-t border-accent">
                  <div className="px-5 py-4 flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <svg className="w-5 h-5 text-muted" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" />
                      </svg>
                      <span className="text-sm font-medium text-primary">Pilih Pengiriman</span>
                    </div>
                    {!addressSaved && (
                      <span className="text-xs text-amber-700 bg-amber-50 px-3 py-1.5 rounded-full font-medium">Isi alamat dulu</span>
                    )}
                  </div>

                  {addressSaved && (
                    <div className="px-5 pb-5">
                      {selectedRate ? (
                        <button onClick={() => setShowShippingPanel(true)}
                          className="w-full p-4 border-2 border-primary bg-primary/5 rounded-xl flex items-center gap-4 hover:bg-primary/10 transition-all">
                          <div className="w-14 h-10 flex-shrink-0">
                            <CourierLogo code={selectedRate.courier_code} />
                          </div>
                          <div className="flex-1 text-left">
                            <div className="flex items-center gap-2">
                              <span className="font-semibold text-primary">{selectedRate.courier_name}</span>
                              <ShippingTypeBadge type={selectedRate.type} />
                            </div>
                            <p className="text-sm text-muted">{selectedRate.courier_service_name} • {selectedRate.duration}</p>
                          </div>
                          <div className="text-right">
                            <p className="font-bold text-primary">Rp {selectedRate.price.toLocaleString("id-ID")}</p>
                            <p className="text-xs text-muted">Ubah</p>
                          </div>
                        </button>
                      ) : (
                        <button onClick={() => setShowShippingPanel(true)}
                          className="w-full p-4 border-2 border-dashed border-accent rounded-xl flex items-center justify-center gap-2 text-muted hover:border-primary hover:text-primary transition-all">
                          {loadingRates ? (
                            <>
                              <div className="animate-spin w-5 h-5 border-2 border-primary border-t-transparent rounded-full" />
                              <span>Memuat opsi pengiriman...</span>
                            </>
                          ) : (
                            <>
                              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 5l7 7-7 7" />
                              </svg>
                              <span>Pilih Pengiriman</span>
                            </>
                          )}
                        </button>
                      )}
                    </div>
                  )}
                </div>
              </div>
            </div>

            {/* RIGHT COLUMN - Payment & Summary */}
            <div className="w-full lg:w-[380px] space-y-4">
              {/* Payment Methods */}
              <div className="bg-white rounded-xl shadow-sm border border-accent overflow-hidden">
                <div className="px-5 py-4 border-b border-accent flex items-center justify-between">
                  <h2 className="font-medium text-primary">Metode Pembayaran</h2>
                  {selectedPayment && (
                    <button onClick={() => setShowPaymentPanel(true)} className="text-sm font-medium text-primary hover:text-primary/80 transition-all">
                      Lihat Semua
                    </button>
                  )}
                </div>

                <div className="p-4 space-y-1.5">
                  {[...PAYMENT_METHODS.eWallet, ...PAYMENT_METHODS.virtualAccount.slice(0, 2)].map((method) => (
                    <button key={method.id} onClick={() => setSelectedPayment(method.id)}
                      className={`w-full p-3 rounded-xl border-2 flex items-center gap-3 transition-all ${
                        selectedPayment === method.id ? "border-primary bg-primary/5" : "border-accent hover:border-primary/30 hover:bg-secondary"
                      }`}>
                      <div className="w-12 h-8 bg-white rounded-lg border border-accent flex items-center justify-center overflow-hidden p-1">
                        <Image src={method.logo} alt={method.name} width={40} height={28} className="object-contain" />
                      </div>
                      <span className="flex-1 text-left text-sm font-medium text-primary">{method.name}</span>
                      <div className={`w-5 h-5 rounded-full border-2 flex items-center justify-center transition-all ${
                        selectedPayment === method.id ? "border-primary bg-primary" : "border-accent"
                      }`}>
                        {selectedPayment === method.id && (
                          <svg className="w-3 h-3 text-white" fill="currentColor" viewBox="0 0 20 20">
                            <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                          </svg>
                        )}
                      </div>
                    </button>
                  ))}

                  <button onClick={() => setShowPaymentPanel(true)}
                    className="w-full p-2.5 text-sm text-primary font-medium hover:bg-primary/5 rounded-xl transition-all flex items-center justify-center gap-1">
                    Lihat Semua Metode Pembayaran
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 5l7 7-7 7" />
                    </svg>
                  </button>
                </div>
              </div>

              {/* Order Summary */}
              <div className="bg-white rounded-xl shadow-sm border border-accent overflow-hidden sticky top-24">
                <div className="px-5 py-4 border-b border-accent">
                  <h2 className="font-medium text-primary">Ringkasan Belanja</h2>
                </div>

                <div className="p-5 space-y-3">
                  <div className="flex justify-between text-sm">
                    <span className="text-muted">Total Harga ({cart.length} barang)</span>
                    <span className="text-primary font-medium">Rp {subtotal.toLocaleString("id-ID")}</span>
                  </div>
                  {shippingMeta.totalWeight > 0 && (
                    <div className="flex justify-between text-sm">
                      <span className="text-muted">Total Berat</span>
                      <span className="text-primary">{shippingMeta.totalWeightKg}</span>
                    </div>
                  )}
                  <div className="flex justify-between text-sm">
                    <span className="text-muted">Total Ongkos Kirim</span>
                    <span className="text-primary font-medium">{selectedRate ? `Rp ${shippingCost.toLocaleString("id-ID")}` : "-"}</span>
                  </div>

                  <div className="border-t border-accent pt-4 mt-4">
                    <div className="flex justify-between items-center">
                      <span className="font-medium text-primary">Total Tagihan</span>
                      <span className="text-xl font-serif font-bold text-primary">Rp {total.toLocaleString("id-ID")}</span>
                    </div>
                  </div>
                </div>

                <div className="p-5 pt-0">
                  <button onClick={handlePayment}
                    disabled={!addressSaved || !selectedRate || !selectedPayment || loading}
                    className="w-full py-4 bg-primary text-white font-semibold rounded-xl hover:bg-primary/90 transition-all disabled:bg-accent disabled:text-muted disabled:cursor-not-allowed flex items-center justify-center gap-2 shadow-sm">
                    {loading ? (
                      <>
                        <div className="animate-spin w-5 h-5 border-2 border-white border-t-transparent rounded-full" />
                        <span>Memproses...</span>
                      </>
                    ) : (
                      <>
                        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                        <span>Bayar Sekarang</span>
                      </>
                    )}
                  </button>

                  <p className="text-xs text-muted text-center mt-3">
                    Dengan melanjutkan pembayaran, kamu menyetujui S&K
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* ============================================ */}
      {/* ADDRESS PANEL - Biteship Native */}
      {/* ============================================ */}
      <AnimatePresence>
        {showAddressPanel && (
          <>
            <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}
              className="fixed inset-0 bg-black/50 z-50" onClick={() => setShowAddressPanel(false)} />
            <motion.div initial={{ x: "100%" }} animate={{ x: 0 }} exit={{ x: "100%" }}
              transition={{ type: "tween", duration: 0.3 }}
              className="fixed right-0 top-0 h-full w-full max-w-md bg-white z-50 shadow-2xl flex flex-col">
              <div className="px-5 py-4 border-b border-accent flex items-center gap-3">
                <button onClick={() => setShowAddressPanel(false)} className="w-10 h-10 flex items-center justify-center hover:bg-secondary rounded-full transition">
                  <svg className="w-5 h-5 text-muted" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
                <h2 className="font-semibold text-primary">Alamat Pengiriman</h2>
              </div>

              <div className="flex-1 overflow-y-auto p-5 space-y-4">
                {/* Recipient Name */}
                <div>
                  <label className="block text-xs font-medium text-muted uppercase tracking-wide mb-2">Nama Penerima *</label>
                  <input type="text" name="recipient_name" value={address.recipient_name} onChange={handleAddressChange}
                    placeholder="Masukkan nama lengkap"
                    className="w-full px-4 py-3 bg-secondary border border-accent rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary transition" />
                </div>

                {/* Email for guest */}
                {!isAuthenticated && (
                  <div>
                    <label className="block text-xs font-medium text-muted uppercase tracking-wide mb-2">Email *</label>
                    <input type="email" value={guestEmail} onChange={(e) => setGuestEmail(e.target.value)}
                      placeholder="email@example.com"
                      className="w-full px-4 py-3 bg-secondary border border-accent rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary transition" />
                  </div>
                )}

                {/* Phone */}
                <div>
                  <label className="block text-xs font-medium text-muted uppercase tracking-wide mb-2">Nomor Telepon *</label>
                  <input type="tel" name="phone" value={address.phone} onChange={handleAddressChange}
                    placeholder="08xxxxxxxxxx"
                    className="w-full px-4 py-3 bg-secondary border border-accent rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary transition" />
                </div>

                {/* BITESHIP AREA AUTOCOMPLETE - Single input, NO dropdowns */}
                <div>
                  <label className="block text-xs font-medium text-muted uppercase tracking-wide mb-2">Cari Kecamatan / Area Pengiriman *</label>
                  <button onClick={() => setShowAreaSearch(true)}
                    className="w-full px-4 py-3 bg-secondary border border-accent rounded-xl text-left flex items-center justify-between hover:border-primary/30 transition">
                    {address.area_id ? (
                      <div className="flex-1 min-w-0">
                        <span className="text-primary font-medium block truncate">{address.area_name}</span>
                        <span className="text-xs text-muted">Kode Pos: {address.postal_code}</span>
                      </div>
                    ) : (
                      <span className="text-muted">Cari Kecamatan / Area Pengiriman</span>
                    )}
                    <svg className="w-5 h-5 text-muted flex-shrink-0 ml-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                    </svg>
                  </button>
                  {address.area_id && (
                    <p className="text-xs text-green-600 mt-1 flex items-center gap-1">
                      <svg className="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
                        <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                      </svg>
                      Area terverifikasi Biteship
                    </p>
                  )}
                </div>

                {/* Detailed Address - Street, house number, RT/RW */}
                <div>
                  <label className="block text-xs font-medium text-muted uppercase tracking-wide mb-2">Alamat Detail *</label>
                  <textarea name="full_address" value={address.full_address} onChange={handleAddressChange} rows={3}
                    placeholder="Nama jalan, nomor rumah, RT/RW, patokan"
                    className="w-full px-4 py-3 bg-secondary border border-accent rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary transition resize-none" />
                  <p className="text-xs text-muted mt-1">Contoh: Jl. Pemuda No. 123, RT 01/RW 02, dekat Masjid Al-Ikhlas</p>
                </div>
              </div>

              <div className="p-5 border-t border-accent bg-white">
                <button onClick={saveAddress} className="w-full py-3.5 bg-primary text-white font-semibold rounded-xl hover:bg-primary/90 transition">
                  Simpan Alamat
                </button>
              </div>
            </motion.div>
          </>
        )}
      </AnimatePresence>

      {/* ============================================ */}
      {/* BITESHIP AREA AUTOCOMPLETE PANEL */}
      {/* Single input field - NO dropdowns */}
      {/* ============================================ */}
      <AnimatePresence>
        {showAreaSearch && (
          <>
            <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}
              className="fixed inset-0 bg-black/50 z-[60]" onClick={() => setShowAreaSearch(false)} />
            <motion.div initial={{ x: "100%" }} animate={{ x: 0 }} exit={{ x: "100%" }}
              transition={{ type: "tween", duration: 0.3 }}
              className="fixed right-0 top-0 h-full w-full max-w-md bg-white z-[60] shadow-2xl flex flex-col">
              <div className="px-5 py-4 border-b border-accent flex items-center gap-3">
                <button onClick={() => setShowAreaSearch(false)} className="w-10 h-10 flex items-center justify-center hover:bg-secondary rounded-full">
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
                  </svg>
                </button>
                <h2 className="font-semibold text-primary">Cari Kecamatan / Area Pengiriman</h2>
              </div>

              {/* Single autocomplete input - NO dropdowns */}
              <div className="p-4 border-b">
                <div className="relative">
                  <svg className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-muted" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                  </svg>
                  <input
                    type="text"
                    value={areaQuery}
                    onChange={(e) => handleAreaQueryChange(e.target.value)}
                    placeholder="Cari Kecamatan / Area Pengiriman"
                    className="w-full pl-12 pr-4 py-3 bg-secondary border-0 rounded-xl focus:outline-none focus:ring-2 focus:ring-primary/20"
                    autoFocus
                  />
                  {loadingAreas && (
                    <div className="absolute right-4 top-1/2 -translate-y-1/2">
                      <div className="animate-spin w-5 h-5 border-2 border-primary border-t-transparent rounded-full" />
                    </div>
                  )}
                </div>
                <p className="text-xs text-muted mt-2">Contoh: Pedurungan Semarang, Menteng Jakarta, 50131</p>
              </div>

              {/* Autocomplete results - rendered EXACTLY as returned by Biteship */}
              <div className="flex-1 overflow-y-auto">
                {areaQuery.length < 3 ? (
                  <div className="text-center py-12 text-muted">
                    <svg className="w-12 h-12 mx-auto mb-3 text-accent" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                    </svg>
                    <p>Ketik minimal 3 karakter untuk mencari</p>
                    <p className="text-xs mt-2 text-muted/70">Hasil akan muncul otomatis saat mengetik</p>
                  </div>
                ) : areaResults.length === 0 && !loadingAreas ? (
                  <div className="text-center py-12 text-muted">
                    <svg className="w-10 h-10 mx-auto mb-3 text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                    </svg>
                    <p className="font-medium">Tidak ditemukan untuk &quot;{areaQuery}&quot;</p>
                    <p className="text-sm mt-2">Coba kata kunci lain atau nama lengkap area</p>
                  </div>
                ) : (
                  areaResults.map((area, idx) => (
                    <button key={`${area.area_id}-${idx}`} onClick={() => selectArea(area)}
                      className="w-full px-5 py-4 text-left hover:bg-secondary flex items-start gap-3 border-b border-gray-50 transition">
                      <div className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center flex-shrink-0 mt-0.5">
                        <svg className="w-4 h-4 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                        </svg>
                      </div>
                      <div className="flex-1 min-w-0">
                        <p className="font-medium text-primary">{area.name}</p>
                        <p className="text-xs text-primary mt-1 font-medium">Kode Pos: {area.postal_code}</p>
                      </div>
                      {selectedArea?.area_id === area.area_id && (
                        <svg className="w-5 h-5 text-primary flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                          <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                        </svg>
                      )}
                    </button>
                  ))
                )}
              </div>
            </motion.div>
          </>
        )}
      </AnimatePresence>

      {/* ============================================ */}
      {/* SHIPPING PANEL - Biteship Rates */}
      {/* ============================================ */}
      <AnimatePresence>
        {showShippingPanel && (
          <>
            <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}
              className="fixed inset-0 bg-black/50 z-50" onClick={() => setShowShippingPanel(false)} />
            <motion.div initial={{ x: "100%" }} animate={{ x: 0 }} exit={{ x: "100%" }}
              transition={{ type: "tween", duration: 0.3 }}
              className="fixed right-0 top-0 h-full w-full max-w-lg bg-white z-50 shadow-2xl flex flex-col">
              <div className="px-5 py-4 border-b border-accent flex items-center gap-3">
                <button onClick={() => setShowShippingPanel(false)} className="w-10 h-10 flex items-center justify-center hover:bg-secondary rounded-full">
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
                <h2 className="font-semibold text-primary">Pilih Pengiriman</h2>
              </div>

              {/* Shipping Meta */}
              <div className="px-5 py-3 bg-secondary border-b flex items-center gap-4 text-xs text-muted">
                <span className="flex items-center gap-1">
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                  </svg>
                  Dari {shippingMeta.originCity}
                </span>
                <span className="flex items-center gap-1">
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 6l3 1m0 0l-3 9a5.002 5.002 0 006.001 0M6 7l3 9M6 7l6-2m6 2l3-1m-3 1l-3 9a5.002 5.002 0 006.001 0M18 7l3 9m-3-9l-6-2m0-2v2m0 16V5m0 16H9m3 0h3" />
                  </svg>
                  Berat: {shippingMeta.totalWeightKg}
                </span>
              </div>

              <div className="flex-1 overflow-y-auto">
                {loadingRates ? (
                  <div className="flex items-center justify-center py-12">
                    <div className="animate-spin w-8 h-8 border-2 border-primary border-t-transparent rounded-full" />
                  </div>
                ) : shippingRates.length === 0 ? (
                  <div className="text-center py-12 text-muted">
                    <p>Tidak ada layanan tersedia</p>
                    <button onClick={() => loadShippingRates(address.area_id, address.postal_code)} className="text-primary hover:underline text-sm mt-2">
                      Coba lagi
                    </button>
                  </div>
                ) : (
                  <div className="divide-y divide-gray-100">
                    {Object.entries(groupedRates).map(([type, rates]) => (
                      <div key={type} className="py-2">
                        <div className="px-5 py-2 sticky top-0 bg-white">
                          <span className={`text-xs font-bold tracking-wider uppercase ${
                            type === 'instant' ? 'text-rose-600' :
                            type === 'same_day' ? 'text-amber-600' :
                            type === 'express' ? 'text-blue-600' :
                            'text-gray-600'
                          }`}>{SHIPPING_TYPE_STYLES[type]?.label || type.toUpperCase()}</span>
                        </div>
                        {rates.map((rate, idx) => {
                          const isSelected = selectedRate?.courier_code === rate.courier_code &&
                                           selectedRate?.courier_service_code === rate.courier_service_code;
                          return (
                            <button key={`${rate.courier_code}-${rate.courier_service_code}-${idx}`}
                              onClick={() => selectShipping(rate)}
                              className={`w-full px-5 py-4 text-left transition ${
                                isSelected ? "bg-primary/5 border-l-4 border-primary" : "hover:bg-secondary border-l-4 border-transparent"
                              }`}>
                              <div className="flex items-start gap-3">
                                <div className="w-12 h-8 flex-shrink-0">
                                  <CourierLogo code={rate.courier_code} />
                                </div>
                                <div className="flex-1 min-w-0">
                                  <div className="flex items-center gap-2 flex-wrap">
                                    <span className="font-medium text-primary">{rate.courier_name}</span>
                                    <ShippingTypeBadge type={rate.type} />
                                  </div>
                                  <p className="text-sm text-muted mt-0.5">{rate.courier_service_name}</p>
                                  <p className="text-sm text-muted mt-1">{rate.duration}</p>
                                </div>
                                <div className="flex items-center gap-3 flex-shrink-0">
                                  <p className="font-bold text-primary">Rp {rate.price.toLocaleString("id-ID")}</p>
                                  <div className={`w-5 h-5 rounded-full border-2 flex items-center justify-center ${
                                    isSelected ? "border-primary bg-primary" : "border-accent"
                                  }`}>
                                    {isSelected && (
                                      <svg className="w-3 h-3 text-white" fill="currentColor" viewBox="0 0 20 20">
                                        <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                                      </svg>
                                    )}
                                  </div>
                                </div>
                              </div>
                            </button>
                          );
                        })}
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </motion.div>
          </>
        )}
      </AnimatePresence>

      {/* ============================================ */}
      {/* PAYMENT PANEL */}
      {/* ============================================ */}
      <AnimatePresence>
        {showPaymentPanel && (
          <>
            <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}
              className="fixed inset-0 bg-black/50 z-50" onClick={() => setShowPaymentPanel(false)} />
            <motion.div initial={{ x: "100%" }} animate={{ x: 0 }} exit={{ x: "100%" }}
              transition={{ type: "tween", duration: 0.3 }}
              className="fixed right-0 top-0 h-full w-full max-w-md bg-white z-50 shadow-2xl flex flex-col">
              <div className="px-5 py-4 border-b border-accent flex items-center gap-3">
                <button onClick={() => setShowPaymentPanel(false)} className="w-10 h-10 flex items-center justify-center hover:bg-secondary rounded-full">
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
                <h2 className="font-semibold text-primary">Metode Pembayaran</h2>
              </div>

              <div className="flex-1 overflow-y-auto">
                {/* E-Wallet / QRIS */}
                <div className="py-3">
                  <div className="px-5 py-2">
                    <span className="text-xs font-bold text-muted tracking-wider">E-WALLET / QRIS</span>
                  </div>
                  {PAYMENT_METHODS.eWallet.map(method => (
                    <button key={method.id}
                      onClick={() => { setSelectedPayment(method.id); setShowPaymentPanel(false); }}
                      className={`w-full px-5 py-4 flex items-center gap-4 transition ${
                        selectedPayment === method.id ? "bg-primary/5" : "hover:bg-secondary"
                      }`}>
                      <div className="w-16 h-10 bg-white rounded-lg border border-accent flex items-center justify-center overflow-hidden p-1.5">
                        <Image src={method.logo} alt={method.name} width={52} height={36} className="object-contain" />
                      </div>
                      <div className="flex-1 text-left">
                        <p className="font-medium text-primary">{method.name}</p>
                        <p className="text-xs text-muted">{method.desc}</p>
                      </div>
                      <div className={`w-5 h-5 rounded-full border-2 flex items-center justify-center ${
                        selectedPayment === method.id ? "border-primary bg-primary" : "border-accent"
                      }`}>
                        {selectedPayment === method.id && (
                          <svg className="w-3 h-3 text-white" fill="currentColor" viewBox="0 0 20 20">
                            <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                          </svg>
                        )}
                      </div>
                    </button>
                  ))}
                </div>

                {/* Virtual Account */}
                <div className="py-3 border-t border-accent">
                  <div className="px-5 py-2">
                    <span className="text-xs font-bold text-muted tracking-wider">VIRTUAL ACCOUNT</span>
                  </div>
                  {PAYMENT_METHODS.virtualAccount.map(method => (
                    <button key={method.id}
                      onClick={() => { setSelectedPayment(method.id); setShowPaymentPanel(false); }}
                      className={`w-full px-5 py-4 flex items-center gap-4 transition ${
                        selectedPayment === method.id ? "bg-primary/5" : "hover:bg-secondary"
                      }`}>
                      <div className="w-16 h-10 bg-white rounded-lg border border-accent flex items-center justify-center overflow-hidden p-1.5">
                        <Image src={method.logo} alt={method.name} width={52} height={36} className="object-contain" />
                      </div>
                      <span className="flex-1 text-left font-medium text-primary">{method.name}</span>
                      <div className={`w-5 h-5 rounded-full border-2 flex items-center justify-center ${
                        selectedPayment === method.id ? "border-primary bg-primary" : "border-accent"
                      }`}>
                        {selectedPayment === method.id && (
                          <svg className="w-3 h-3 text-white" fill="currentColor" viewBox="0 0 20 20">
                            <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                          </svg>
                        )}
                      </div>
                    </button>
                  ))}
                </div>
              </div>
            </motion.div>
          </>
        )}
      </AnimatePresence>

      {/* ============================================ */}
      {/* ADDRESS LIST PANEL (Saved Addresses) */}
      {/* ============================================ */}
      <AnimatePresence>
        {showAddressList && (
          <>
            <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}
              className="fixed inset-0 bg-black/50 z-50" onClick={() => setShowAddressList(false)} />
            <motion.div initial={{ x: "100%" }} animate={{ x: 0 }} exit={{ x: "100%" }}
              transition={{ type: "tween", duration: 0.3 }}
              className="fixed right-0 top-0 h-full w-full max-w-md bg-white z-50 shadow-2xl flex flex-col">
              <div className="px-5 py-4 border-b border-accent flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <button onClick={() => setShowAddressList(false)} className="w-10 h-10 flex items-center justify-center hover:bg-secondary rounded-full">
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                  <h2 className="font-semibold text-primary">Pilih Alamat</h2>
                </div>
                <button onClick={() => { setShowAddressList(false); setShowAddressPanel(true); }}
                  className="text-sm font-medium text-primary hover:bg-primary/5 px-4 py-2 rounded-lg transition">
                  + Alamat Baru
                </button>
              </div>

              <div className="flex-1 overflow-y-auto">
                {savedAddresses.length === 0 ? (
                  <div className="text-center py-12 text-muted">
                    <p>Belum ada alamat tersimpan</p>
                    <button onClick={() => { setShowAddressList(false); setShowAddressPanel(true); }}
                      className="mt-4 text-primary font-medium hover:underline">
                      Tambah Alamat Baru
                    </button>
                  </div>
                ) : (
                  <div className="divide-y divide-gray-100">
                    {savedAddresses.map(addr => {
                      const isSelected = address.full_address === addr.full_address && address.recipient_name === addr.recipient_name;
                      return (
                        <button key={addr.id} onClick={() => selectSavedAddress(addr)}
                          className={`w-full px-5 py-4 text-left transition ${isSelected ? "bg-primary/5 border-l-4 border-primary" : "hover:bg-secondary border-l-4 border-transparent"}`}>
                          <div className="flex items-start gap-3">
                            <div className="flex-1 min-w-0">
                              <div className="flex items-center gap-2 mb-1">
                                <span className="font-medium text-primary">{addr.label || "Alamat"}</span>
                                {addr.is_default && (
                                  <span className="px-2 py-0.5 bg-primary/10 text-primary text-xs font-medium rounded">Utama</span>
                                )}
                              </div>
                              <p className="font-semibold text-gray-900">{addr.recipient_name}</p>
                              <p className="text-sm text-gray-600">{addr.phone}</p>
                              <p className="text-sm text-gray-600 mt-1 line-clamp-2">{addr.full_address}</p>
                              <p className="text-sm text-gray-500 mt-0.5">
                                {addr.area_name || `${addr.subdistrict}, ${addr.district}, ${addr.city_name}`}
                                {addr.postal_code && ` - ${addr.postal_code}`}
                              </p>
                            </div>
                            <div className={`w-5 h-5 rounded-full border-2 flex items-center justify-center flex-shrink-0 mt-1 ${
                              isSelected ? "border-primary bg-primary" : "border-accent"
                            }`}>
                              {isSelected && (
                                <svg className="w-3 h-3 text-white" fill="currentColor" viewBox="0 0 20 20">
                                  <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                                </svg>
                              )}
                            </div>
                          </div>
                        </button>
                      );
                    })}
                  </div>
                )}
              </div>

              {/* Manage addresses link */}
              <div className="p-4 border-t bg-gray-50">
                <Link href="/account/addresses" className="block text-center text-sm text-primary font-medium hover:underline">
                  Kelola Daftar Alamat
                </Link>
              </div>
            </motion.div>
          </>
        )}
      </AnimatePresence>
    </>
  );
}
