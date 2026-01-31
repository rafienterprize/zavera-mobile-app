"use client";

import { useEffect, useState } from "react";
import { useDialog } from "@/hooks/useDialog";

export default function DebugMidtransPage() {
  const dialog = useDialog();
  const [snapStatus, setSnapStatus] = useState<string>("Checking...");
  const [clientKey, setClientKey] = useState<string>("");

  useEffect(() => {
    // Check client key
    const key = process.env.NEXT_PUBLIC_MIDTRANS_CLIENT_KEY || "NOT SET";
    setClientKey(key);

    // Check snap availability
    const checkSnap = () => {
      if (typeof window !== "undefined" && window.snap) {
        setSnapStatus("✅ Loaded and ready");
      } else {
        setSnapStatus("⏳ Not loaded yet...");
        setTimeout(checkSnap, 500);
      }
    };
    checkSnap();
  }, []);

  const testSnapPay = async () => {
    if (window.snap) {
      // This will fail but we can see the error
      try {
        window.snap.pay("test-invalid-token", {
          onSuccess: () => {
            console.log("Success");
          },
          onPending: () => {
            console.log("Pending");
          },
          onError: async (result?: unknown) => {
            console.log("Expected error (invalid token):", result);
            await dialog.alert({
              title: 'Test Berhasil!',
              message: 'Snap.pay berfungsi! Error ini memang diharapkan karena token invalid.',
            });
          },
          onClose: () => {
            console.log("Popup closed");
          },
        });
      } catch (e) {
        console.error("Error calling snap.pay:", e);
        await dialog.alert({
          title: 'Error',
          message: "Error: " + (e as Error).message,
        });
      }
    } else {
      await dialog.alert({
        title: 'Error',
        message: 'Snap belum dimuat!',
      });
    }
  };

  return (
    <div className="max-w-2xl mx-auto p-8 pt-24">
      <h1 className="text-2xl font-bold mb-8">Midtrans Debug Page</h1>
      
      <div className="space-y-4">
        <div className="p-4 bg-gray-100 rounded">
          <p className="font-medium">Client Key:</p>
          <code className="text-sm break-all">{clientKey}</code>
        </div>
        
        <div className="p-4 bg-gray-100 rounded">
          <p className="font-medium">Snap Status:</p>
          <p>{snapStatus}</p>
        </div>
        
        <div className="p-4 bg-gray-100 rounded">
          <p className="font-medium">API URL:</p>
          <code className="text-sm">{process.env.NEXT_PUBLIC_API_URL || "NOT SET"}</code>
        </div>

        <button
          onClick={testSnapPay}
          className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
        >
          Test snap.pay() (will show error - expected)
        </button>
        
        <div className="p-4 bg-yellow-50 border border-yellow-200 rounded">
          <p className="font-medium text-yellow-800">Important:</p>
          <p className="text-sm text-yellow-700">
            Make sure your Client Key (frontend) and Server Key (backend) are from the SAME Midtrans account.
            They should have matching prefixes like Mid-client-XXX and Mid-server-XXX where XXX is similar.
          </p>
        </div>
      </div>
    </div>
  );
}
