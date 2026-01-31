"use client";

import { useState, useEffect, useCallback, useRef } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { motion, AnimatePresence } from "framer-motion";
import { useAuth } from "@/context/AuthContext";
import { useToast } from "@/components/ui/Toast";
import { useDialog } from "@/context/DialogContext";
import api from "@/lib/api";

// ============================================
// TYPES
// ============================================
interface BiteshipArea {
  area_id: string;
  name: string;
  postal_code: string;
  province: string;
  city: string;
  district: string;
}

interface Address {
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

interface AddressForm {
  label: string;
  recipient_name: string;
  phone: string;
  area_id: string;
  area_name: string;
  postal_code: string;
  full_address: string;
  is_default: boolean;
}

const initialForm: AddressForm = {
  label: "",
  recipient_name: "",
  phone: "",
  area_id: "",
  area_name: "",
  postal_code: "",
  full_address: "",
  is_default: false,
};

export default function AddressesPage() {
  const router = useRouter();
  const { user, isAuthenticated, isLoading: authLoading } = useAuth();
  const { showToast } = useToast();
  const dialog = useDialog();

  const [addresses, setAddresses] = useState<Address[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [form, setForm] = useState<AddressForm>(initialForm);
  const [saving, setSaving] = useState(false);

  // Area search
  const [showAreaSearch, setShowAreaSearch] = useState(false);
  const [areaQuery, setAreaQuery] = useState("");
  const [areaResults, setAreaResults] = useState<BiteshipArea[]>([]);
  const [loadingAreas, setLoadingAreas] = useState(false);
  const searchTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const abortControllerRef = useRef<AbortController | null>(null);

  // Redirect if not authenticated
  useEffect(() => {
    if (!authLoading && !isAuthenticated) {
      router.push("/login?redirect=/account/addresses");
    }
  }, [authLoading, isAuthenticated, router]);

  // Load addresses
  const loadAddresses = useCallback(async () => {
    try {
      const res = await api.get("/user/addresses");
      setAddresses(res.data.addresses || []);
    } catch (err) {
      console.error("Failed to load addresses:", err);
      showToast("Gagal memuat alamat", "error");
    } finally {
      setLoading(false);
    }
  }, [showToast]);

  useEffect(() => {
    if (isAuthenticated) {
      loadAddresses();
    }
  }, [isAuthenticated, loadAddresses]);


  // Area search
  const searchAreas = useCallback(async (query: string, signal: AbortSignal) => {
    if (query.length < 3) {
      setAreaResults([]);
      setLoadingAreas(false);
      return;
    }
    try {
      const res = await api.get(`/shipping/areas?q=${encodeURIComponent(query)}`, { signal });
      setAreaResults(res.data.areas || []);
    } catch (err: unknown) {
      if (err instanceof Error && (err.name === 'AbortError' || err.name === 'CanceledError')) return;
      setAreaResults([]);
    } finally {
      setLoadingAreas(false);
    }
  }, []);

  const handleAreaQueryChange = useCallback((query: string) => {
    setAreaQuery(query);
    if (searchTimeoutRef.current) clearTimeout(searchTimeoutRef.current);
    if (abortControllerRef.current) abortControllerRef.current.abort();
    if (query.length < 3) {
      setAreaResults([]);
      setLoadingAreas(false);
      return;
    }
    setLoadingAreas(true);
    searchTimeoutRef.current = setTimeout(() => {
      abortControllerRef.current = new AbortController();
      searchAreas(query, abortControllerRef.current.signal);
    }, 300);
  }, [searchAreas]);

  const selectArea = (area: BiteshipArea) => {
    setForm(prev => ({
      ...prev,
      area_id: area.area_id,
      area_name: area.name,
      postal_code: area.postal_code,
    }));
    setShowAreaSearch(false);
    setAreaQuery("");
    setAreaResults([]);
  };

  // Form handlers
  const openAddForm = () => {
    setForm({ ...initialForm, recipient_name: user?.first_name || "", phone: user?.phone || "" });
    setEditingId(null);
    setShowForm(true);
  };

  const openEditForm = (addr: Address) => {
    setForm({
      label: addr.label,
      recipient_name: addr.recipient_name,
      phone: addr.phone,
      area_id: addr.area_id || "",
      area_name: addr.area_name || `${addr.subdistrict}, ${addr.district}, ${addr.city_name}, ${addr.province_name}`,
      postal_code: addr.postal_code,
      full_address: addr.full_address,
      is_default: addr.is_default,
    });
    setEditingId(addr.id);
    setShowForm(true);
  };

  const closeForm = () => {
    setShowForm(false);
    setEditingId(null);
    setForm(initialForm);
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const { name, value, type } = e.target;
    if (type === "checkbox") {
      setForm(prev => ({ ...prev, [name]: (e.target as HTMLInputElement).checked }));
    } else {
      setForm(prev => ({ ...prev, [name]: value }));
    }
  };

  const saveAddress = async () => {
    if (!form.recipient_name.trim()) { showToast("Nama penerima wajib diisi", "error"); return; }
    if (!form.phone.trim()) { showToast("Nomor telepon wajib diisi", "error"); return; }
    if (!form.area_id) { showToast("Pilih area pengiriman", "error"); return; }
    if (!form.full_address.trim()) { showToast("Alamat lengkap wajib diisi", "error"); return; }

    setSaving(true);
    try {
      const payload = {
        label: form.label || "Alamat",
        recipient_name: form.recipient_name,
        phone: form.phone,
        province_name: form.area_name.split(", ").pop() || "",
        city_name: form.area_name.split(", ")[1] || "",
        district: form.area_name.split(", ")[0] || "",
        subdistrict: form.area_name.split(", ")[0] || "",
        postal_code: form.postal_code,
        full_address: form.full_address,
        is_default: form.is_default,
        area_id: form.area_id,
        area_name: form.area_name,
      };

      if (editingId) {
        await api.put(`/user/addresses/${editingId}`, payload);
        showToast("Alamat berhasil diperbarui", "success");
      } else {
        await api.post("/user/addresses", payload);
        showToast("Alamat berhasil ditambahkan", "success");
      }
      closeForm();
      loadAddresses();
    } catch (err) {
      console.error("Failed to save address:", err);
      showToast("Gagal menyimpan alamat", "error");
    } finally {
      setSaving(false);
    }
  };

  const deleteAddress = async (id: number) => {
    const confirmed = await dialog.confirm({
      title: 'Hapus Alamat',
      message: 'Apakah Anda yakin ingin menghapus alamat ini?',
      variant: 'danger',
      confirmText: 'Ya, Hapus',
      cancelText: 'Batal'
    });
    
    if (!confirmed) return;
    
    try {
      await api.delete(`/user/addresses/${id}`);
      showToast("Alamat berhasil dihapus", "success");
      loadAddresses();
    } catch (err) {
      console.error("Failed to delete address:", err);
      showToast("Gagal menghapus alamat", "error");
    }
  };

  const setDefault = async (id: number) => {
    try {
      await api.post(`/user/addresses/${id}/default`);
      showToast("Alamat utama berhasil diubah", "success");
      loadAddresses();
    } catch (err) {
      console.error("Failed to set default:", err);
      showToast("Gagal mengubah alamat utama", "error");
    }
  };

  if (authLoading || loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="animate-spin w-8 h-8 border-2 border-primary border-t-transparent rounded-full" />
      </div>
    );
  }


  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white border-b shadow-sm sticky top-0 z-40">
        <div className="max-w-4xl mx-auto px-4">
          <div className="h-16 flex items-center justify-between">
            <div className="flex items-center gap-4">
              <Link href="/account/pembelian" className="w-10 h-10 flex items-center justify-center hover:bg-gray-100 rounded-full">
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
                </svg>
              </Link>
              <h1 className="text-lg font-semibold">Daftar Alamat</h1>
            </div>
            <button onClick={openAddForm} className="flex items-center gap-2 px-4 py-2 bg-primary text-white rounded-lg hover:bg-gray-800 transition text-sm font-medium">
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
              Tambah Alamat
            </button>
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="max-w-4xl mx-auto px-4 py-6">
        {addresses.length === 0 ? (
          <div className="bg-white rounded-xl p-12 text-center">
            <div className="w-16 h-16 mx-auto mb-4 bg-gray-100 rounded-full flex items-center justify-center">
              <svg className="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
              </svg>
            </div>
            <h3 className="text-lg font-medium text-gray-900 mb-2">Belum ada alamat tersimpan</h3>
            <p className="text-gray-500 mb-6">Tambahkan alamat untuk mempermudah checkout</p>
            <button onClick={openAddForm} className="px-6 py-3 bg-primary text-white rounded-lg hover:bg-gray-800 transition">
              Tambah Alamat Pertama
            </button>
          </div>
        ) : (
          <div className="space-y-4">
            {addresses.map(addr => (
              <div key={addr.id} className={`bg-white rounded-xl p-5 border-2 transition ${addr.is_default ? "border-primary" : "border-transparent hover:border-gray-200"}`}>
                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-2">
                      <span className="font-medium text-gray-900">{addr.label || "Alamat"}</span>
                      {addr.is_default && (
                        <span className="px-2 py-0.5 bg-primary/10 text-primary text-xs font-medium rounded">Utama</span>
                      )}
                    </div>
                    <p className="font-semibold text-gray-900">{addr.recipient_name}</p>
                    <p className="text-gray-600 text-sm">{addr.phone}</p>
                    <p className="text-gray-600 text-sm mt-2 leading-relaxed">{addr.full_address}</p>
                    <p className="text-gray-500 text-sm mt-1">
                      {addr.area_name || `${addr.subdistrict}, ${addr.district}, ${addr.city_name}, ${addr.province_name}`}
                      {addr.postal_code && ` - ${addr.postal_code}`}
                    </p>
                  </div>
                  <div className="flex items-center gap-2">
                    <button onClick={() => openEditForm(addr)} className="p-2 hover:bg-gray-100 rounded-lg transition" title="Edit">
                      <svg className="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                      </svg>
                    </button>
                    <button onClick={() => deleteAddress(addr.id)} className="p-2 hover:bg-red-50 rounded-lg transition" title="Hapus">
                      <svg className="w-5 h-5 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                      </svg>
                    </button>
                  </div>
                </div>
                {!addr.is_default && (
                  <button onClick={() => setDefault(addr.id)} className="mt-4 text-sm text-primary font-medium hover:underline">
                    Jadikan Alamat Utama
                  </button>
                )}
              </div>
            ))}
          </div>
        )}
      </div>


      {/* Add/Edit Form Modal */}
      <AnimatePresence>
        {showForm && (
          <>
            <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}
              className="fixed inset-0 bg-black/50 z-50" onClick={closeForm} />
            <motion.div initial={{ y: "100%" }} animate={{ y: 0 }} exit={{ y: "100%" }}
              transition={{ type: "tween", duration: 0.3 }}
              className="fixed bottom-0 left-0 right-0 bg-white z-50 rounded-t-2xl max-h-[90vh] overflow-hidden flex flex-col">
              <div className="px-5 py-4 border-b flex items-center justify-between">
                <h2 className="font-semibold text-lg">{editingId ? "Edit Alamat" : "Tambah Alamat Baru"}</h2>
                <button onClick={closeForm} className="w-10 h-10 flex items-center justify-center hover:bg-gray-100 rounded-full">
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>

              <div className="flex-1 overflow-y-auto p-5 space-y-4">
                {/* Label */}
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Label Alamat</label>
                  <input type="text" name="label" value={form.label} onChange={handleChange}
                    placeholder="Contoh: Rumah, Kantor" className="w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-primary/20 focus:border-primary" />
                </div>

                {/* Recipient */}
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Nama Penerima *</label>
                  <input type="text" name="recipient_name" value={form.recipient_name} onChange={handleChange}
                    placeholder="Nama lengkap penerima" className="w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-primary/20 focus:border-primary" />
                </div>

                {/* Phone */}
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Nomor Telepon *</label>
                  <input type="tel" name="phone" value={form.phone} onChange={handleChange}
                    placeholder="08xxxxxxxxxx" className="w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-primary/20 focus:border-primary" />
                </div>

                {/* Area Search */}
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Kecamatan/Kelurahan *</label>
                  {form.area_id ? (
                    <div className="flex items-center gap-2">
                      <div className="flex-1 px-4 py-3 bg-gray-50 border rounded-lg text-sm">{form.area_name}</div>
                      <button onClick={() => { setForm(prev => ({ ...prev, area_id: "", area_name: "", postal_code: "" })); setShowAreaSearch(true); }}
                        className="px-4 py-3 text-primary font-medium hover:bg-primary/5 rounded-lg">Ganti</button>
                    </div>
                  ) : (
                    <button onClick={() => setShowAreaSearch(true)}
                      className="w-full px-4 py-3 border rounded-lg text-left text-gray-400 hover:border-primary transition">
                      Cari kecamatan atau kelurahan...
                    </button>
                  )}
                </div>

                {/* Full Address */}
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Alamat Lengkap *</label>
                  <textarea name="full_address" value={form.full_address} onChange={handleChange} rows={3}
                    placeholder="Nama jalan, nomor rumah, RT/RW, patokan" className="w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-primary/20 focus:border-primary resize-none" />
                </div>

                {/* Default checkbox */}
                <label className="flex items-center gap-3 cursor-pointer">
                  <input type="checkbox" name="is_default" checked={form.is_default} onChange={handleChange}
                    className="w-5 h-5 rounded border-gray-300 text-primary focus:ring-primary" />
                  <span className="text-sm text-gray-700">Jadikan alamat utama</span>
                </label>
              </div>

              <div className="p-5 border-t bg-white">
                <button onClick={saveAddress} disabled={saving}
                  className="w-full py-4 bg-primary text-white rounded-xl font-medium hover:bg-gray-800 transition disabled:opacity-50">
                  {saving ? "Menyimpan..." : (editingId ? "Simpan Perubahan" : "Simpan Alamat")}
                </button>
              </div>
            </motion.div>
          </>
        )}
      </AnimatePresence>


      {/* Area Search Modal */}
      <AnimatePresence>
        {showAreaSearch && (
          <>
            <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}
              className="fixed inset-0 bg-black/50 z-[60]" onClick={() => setShowAreaSearch(false)} />
            <motion.div initial={{ y: "100%" }} animate={{ y: 0 }} exit={{ y: "100%" }}
              transition={{ type: "tween", duration: 0.3 }}
              className="fixed bottom-0 left-0 right-0 bg-white z-[60] rounded-t-2xl max-h-[80vh] overflow-hidden flex flex-col">
              <div className="px-5 py-4 border-b">
                <div className="flex items-center gap-3">
                  <button onClick={() => setShowAreaSearch(false)} className="w-10 h-10 flex items-center justify-center hover:bg-gray-100 rounded-full">
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
                    </svg>
                  </button>
                  <input type="text" value={areaQuery} onChange={(e) => handleAreaQueryChange(e.target.value)}
                    placeholder="Cari kecamatan atau kelurahan..." autoFocus
                    className="flex-1 px-4 py-3 bg-gray-100 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary/20" />
                </div>
              </div>

              <div className="flex-1 overflow-y-auto">
                {loadingAreas ? (
                  <div className="flex items-center justify-center py-12">
                    <div className="animate-spin w-6 h-6 border-2 border-primary border-t-transparent rounded-full" />
                  </div>
                ) : areaQuery.length < 3 ? (
                  <div className="text-center py-12 text-gray-500">
                    <p>Ketik minimal 3 karakter untuk mencari</p>
                  </div>
                ) : areaResults.length === 0 ? (
                  <div className="text-center py-12 text-gray-500">
                    <p>Tidak ditemukan hasil untuk &quot;{areaQuery}&quot;</p>
                  </div>
                ) : (
                  <div className="divide-y">
                    {areaResults.map((area, idx) => (
                      <button key={`${area.area_id}-${idx}`} onClick={() => selectArea(area)}
                        className="w-full px-5 py-4 text-left hover:bg-gray-50 transition">
                        <p className="font-medium text-gray-900">{area.name}</p>
                        {area.postal_code && <p className="text-sm text-gray-500 mt-0.5">Kode Pos: {area.postal_code}</p>}
                      </button>
                    ))}
                  </div>
                )}
              </div>
            </motion.div>
          </>
        )}
      </AnimatePresence>
    </div>
  );
}
