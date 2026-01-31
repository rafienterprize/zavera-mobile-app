"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";

/**
 * Redirect page - /account/orders now redirects to /account/pembelian?tab=history
 * This page is deprecated in favor of the unified Tokopedia-style Pembelian page
 */
export default function OrderHistoryRedirect() {
  const router = useRouter();

  useEffect(() => {
    // Redirect to the new unified Pembelian page with history tab
    router.replace("/account/pembelian?tab=history");
  }, [router]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="text-center">
        <div className="animate-spin rounded-full h-8 w-8 border-t-2 border-b-2 border-primary mx-auto mb-4"></div>
        <p className="text-gray-600">Mengalihkan ke Daftar Transaksi...</p>
      </div>
    </div>
  );
}
