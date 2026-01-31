"use client";

import { useEffect } from "react";

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error("Application error:", error);
  }, [error]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="text-center p-8">
        <h2 className="text-2xl font-bold text-gray-900 mb-4">
          Terjadi Kesalahan
        </h2>
        <p className="text-gray-600 mb-6">
          Maaf, terjadi kesalahan saat memuat halaman ini.
        </p>
        <button
          onClick={() => reset()}
          className="px-6 py-3 bg-black text-white rounded-lg hover:bg-gray-800 transition-colors"
        >
          Coba Lagi
        </button>
      </div>
    </div>
  );
}
