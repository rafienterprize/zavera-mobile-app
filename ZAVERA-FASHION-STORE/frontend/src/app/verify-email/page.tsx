"use client";

import { useEffect, useState, Suspense } from "react";
import { useSearchParams } from "next/navigation";
import Link from "next/link";
import { motion } from "framer-motion";
import api from "@/lib/api";

function VerifyEmailContent() {
  const searchParams = useSearchParams();
  const token = searchParams.get("token");

  const [status, setStatus] = useState<"loading" | "success" | "error">("loading");
  const [message, setMessage] = useState("");

  useEffect(() => {
    if (!token) {
      setStatus("error");
      setMessage("Token verifikasi tidak ditemukan.");
      return;
    }

    verifyEmail();
  }, [token]);

  const verifyEmail = async () => {
    try {
      const response = await api.get(`/auth/verify-email?token=${token}`);
      setStatus("success");
      setMessage(response.data.message || "Email berhasil diverifikasi!");
    } catch (error: unknown) {
      setStatus("error");
      const err = error as { response?: { data?: { message?: string } } };
      setMessage(err.response?.data?.message || "Verifikasi gagal. Token mungkin sudah kadaluarsa.");
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center py-12 px-4">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="max-w-md w-full"
      >
        <div className="bg-white py-12 px-8 shadow-lg rounded-xl text-center">
          {/* Logo */}
          <Link href="/" className="inline-block mb-8">
            <h1 className="text-3xl font-serif font-bold tracking-[0.2em] text-gray-900">
              ZAVERA
            </h1>
          </Link>

          {status === "loading" && (
            <>
              <div className="flex justify-center mb-6">
                <div className="animate-spin rounded-full h-16 w-16 border-t-2 border-b-2 border-primary"></div>
              </div>
              <h2 className="text-xl font-semibold text-gray-900 mb-2">
                Memverifikasi Email...
              </h2>
              <p className="text-gray-600">Mohon tunggu sebentar</p>
            </>
          )}

          {status === "success" && (
            <>
              <div className="flex justify-center mb-6">
                <div className="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center">
                  <svg className="w-8 h-8 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                  </svg>
                </div>
              </div>
              <h2 className="text-xl font-semibold text-gray-900 mb-2">
                Verifikasi Berhasil!
              </h2>
              <p className="text-gray-600 mb-8">{message}</p>
              <Link
                href="/login"
                className="inline-block w-full py-3 px-4 bg-primary text-white rounded-lg font-medium hover:bg-primary/90 transition-colors"
              >
                Masuk ke Akun
              </Link>
            </>
          )}

          {status === "error" && (
            <>
              <div className="flex justify-center mb-6">
                <div className="w-16 h-16 bg-red-100 rounded-full flex items-center justify-center">
                  <svg className="w-8 h-8 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </div>
              </div>
              <h2 className="text-xl font-semibold text-gray-900 mb-2">
                Verifikasi Gagal
              </h2>
              <p className="text-gray-600 mb-8">{message}</p>
              <div className="space-y-3">
                <Link
                  href="/login"
                  className="inline-block w-full py-3 px-4 bg-primary text-white rounded-lg font-medium hover:bg-primary/90 transition-colors"
                >
                  Kembali ke Login
                </Link>
                <Link
                  href="/register"
                  className="inline-block w-full py-3 px-4 border border-gray-300 text-gray-700 rounded-lg font-medium hover:bg-gray-50 transition-colors"
                >
                  Daftar Ulang
                </Link>
              </div>
            </>
          )}
        </div>
      </motion.div>
    </div>
  );
}

export default function VerifyEmailPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary"></div>
      </div>
    }>
      <VerifyEmailContent />
    </Suspense>
  );
}
