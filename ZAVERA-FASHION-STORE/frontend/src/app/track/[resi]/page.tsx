"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { Package, Truck, CheckCircle, Clock, MapPin, ArrowLeft } from "lucide-react";
import api from "@/lib/api";

interface TrackingHistory {
  note: string;
  status: string;
  updated_at: string;
}

interface TrackingData {
  order_code: string;
  resi: string;
  courier_name: string;
  status: string;
  origin: string;
  destination: string;
  history: TrackingHistory[];
}

export default function TrackingPage() {
  const params = useParams();
  const resi = params.resi as string;

  const [tracking, setTracking] = useState<TrackingData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    loadTracking();
  }, [resi]);

  const loadTracking = async () => {
    try {
      const response = await api.get(`/tracking/${resi}`);
      setTracking(response.data);
    } catch (error: any) {
      console.error("Failed to load tracking:", error);
      setError(error.response?.data?.message || "Gagal memuat data tracking");
    } finally {
      setLoading(false);
    }
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

  const getStatusIcon = (status: string) => {
    const statusLower = status.toLowerCase();
    if (statusLower.includes("delivered") || statusLower.includes("selesai")) {
      return <CheckCircle className="text-emerald-400" size={24} />;
    }
    if (statusLower.includes("transit") || statusLower.includes("perjalanan")) {
      return <Truck className="text-blue-400" size={24} />;
    }
    if (statusLower.includes("picked") || statusLower.includes("diambil")) {
      return <Package className="text-purple-400" size={24} />;
    }
    return <Clock className="text-amber-400" size={24} />;
  };

  const getStatusColor = (status: string) => {
    const statusLower = status.toLowerCase();
    if (statusLower.includes("delivered") || statusLower.includes("selesai")) {
      return "bg-emerald-500/20 text-emerald-400 border-emerald-500/30";
    }
    if (statusLower.includes("transit") || statusLower.includes("perjalanan")) {
      return "bg-blue-500/20 text-blue-400 border-blue-500/30";
    }
    if (statusLower.includes("picked") || statusLower.includes("diambil")) {
      return "bg-purple-500/20 text-purple-400 border-purple-500/30";
    }
    return "bg-amber-500/20 text-amber-400 border-amber-500/30";
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-neutral-950 via-neutral-900 to-neutral-950 flex items-center justify-center">
        <div className="w-10 h-10 border-2 border-white/20 border-t-white rounded-full animate-spin" />
      </div>
    );
  }

  if (error || !tracking) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-neutral-950 via-neutral-900 to-neutral-950 flex items-center justify-center p-4">
        <div className="max-w-md w-full bg-neutral-900 rounded-2xl border border-white/10 p-8 text-center">
          <Package className="text-white/40 mx-auto mb-4" size={48} />
          <h1 className="text-2xl font-bold text-white mb-2">Tracking Tidak Ditemukan</h1>
          <p className="text-white/60 mb-6">
            {error || "Nomor resi tidak ditemukan atau belum tersedia"}
          </p>
          <Link
            href="/"
            className="inline-flex items-center gap-2 px-6 py-3 rounded-xl bg-white text-black hover:bg-white/90 transition-colors font-semibold"
          >
            <ArrowLeft size={18} />
            Kembali ke Beranda
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-neutral-950 via-neutral-900 to-neutral-950 py-12 px-4">
      <div className="max-w-4xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <Link
            href="/"
            className="inline-flex items-center gap-2 text-white/60 hover:text-white transition-colors mb-4"
          >
            <ArrowLeft size={18} />
            Kembali
          </Link>
          <h1 className="text-3xl font-bold text-white mb-2">Lacak Pesanan</h1>
          <p className="text-white/60">Pantau status pengiriman pesanan Anda</p>
        </div>

        {/* Tracking Info Card */}
        <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6 mb-6">
          <div className="grid md:grid-cols-2 gap-6">
            <div>
              <p className="text-white/60 text-sm mb-1">Nomor Resi</p>
              <p className="text-white font-mono text-lg font-bold">{tracking.resi}</p>
            </div>
            <div>
              <p className="text-white/60 text-sm mb-1">Kurir</p>
              <p className="text-white text-lg font-semibold">{tracking.courier_name}</p>
            </div>
            <div>
              <p className="text-white/60 text-sm mb-1">Kode Pesanan</p>
              <p className="text-white font-medium">{tracking.order_code}</p>
            </div>
            <div>
              <p className="text-white/60 text-sm mb-1">Status</p>
              <span className={`inline-block px-3 py-1 rounded-full text-sm font-medium border ${getStatusColor(tracking.status)}`}>
                {tracking.status}
              </span>
            </div>
          </div>

          {/* Origin & Destination */}
          <div className="mt-6 pt-6 border-t border-white/10 grid md:grid-cols-2 gap-6">
            <div className="flex items-start gap-3">
              <div className="p-2 rounded-lg bg-blue-500/20 flex-shrink-0">
                <MapPin className="text-blue-400" size={18} />
              </div>
              <div>
                <p className="text-white/60 text-xs mb-1">Asal</p>
                <p className="text-white text-sm">{tracking.origin}</p>
              </div>
            </div>
            <div className="flex items-start gap-3">
              <div className="p-2 rounded-lg bg-purple-500/20 flex-shrink-0">
                <MapPin className="text-purple-400" size={18} />
              </div>
              <div>
                <p className="text-white/60 text-xs mb-1">Tujuan</p>
                <p className="text-white text-sm">{tracking.destination}</p>
              </div>
            </div>
          </div>
        </div>

        {/* Tracking History */}
        <div className="bg-neutral-900 rounded-2xl border border-white/10 p-6">
          <h2 className="text-xl font-bold text-white mb-6">Riwayat Pengiriman</h2>
          
          {tracking.history && tracking.history.length > 0 ? (
            <div className="space-y-4">
              {tracking.history.map((item, index) => (
                <div key={index} className="flex gap-4">
                  {/* Timeline */}
                  <div className="flex flex-col items-center">
                    <div className="p-2 rounded-lg bg-white/10">
                      {getStatusIcon(item.status)}
                    </div>
                    {index < tracking.history.length - 1 && (
                      <div className="w-0.5 h-full bg-white/10 my-2" />
                    )}
                  </div>

                  {/* Content */}
                  <div className="flex-1 pb-6">
                    <div className="bg-white/5 rounded-xl p-4 border border-white/10">
                      <p className="text-white font-medium mb-1">{item.note}</p>
                      <p className="text-white/60 text-sm">{formatDate(item.updated_at)}</p>
                      <span className={`inline-block mt-2 px-2 py-1 rounded text-xs font-medium ${getStatusColor(item.status)}`}>
                        {item.status}
                      </span>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-8">
              <Clock className="text-white/40 mx-auto mb-3" size={48} />
              <p className="text-white/60">Belum ada riwayat pengiriman</p>
              <p className="text-white/40 text-sm mt-1">
                Riwayat akan muncul setelah paket diambil kurir
              </p>
            </div>
          )}
        </div>

        {/* Help Section */}
        <div className="mt-6 bg-blue-500/10 border border-blue-500/30 rounded-xl p-4">
          <p className="text-blue-400 text-sm">
            💡 <strong>Tips:</strong> Simpan nomor resi ini untuk melacak pesanan Anda kapan saja.
            Tracking akan diupdate secara otomatis oleh kurir.
          </p>
        </div>
      </div>
    </div>
  );
}
