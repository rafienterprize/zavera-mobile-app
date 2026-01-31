"use client";

import Script from "next/script";

export default function MidtransScript() {
  const midtransClientKey = process.env.NEXT_PUBLIC_MIDTRANS_CLIENT_KEY;

  return (
    <Script
      id="midtrans-snap"
      src="https://app.sandbox.midtrans.com/snap/snap.js"
      data-client-key={midtransClientKey}
      strategy="afterInteractive"
      onLoad={() => {
        console.log("✅ Midtrans Snap script loaded, client key:", midtransClientKey?.substring(0, 15) + "...");
      }}
      onError={(e) => {
        console.error("❌ Failed to load Midtrans Snap script:", e);
      }}
    />
  );
}
