"use client";

import { Suspense, useEffect, useState } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import Link from "next/link";
import { motion } from "framer-motion";

function OrderPendingContent() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const orderCode = searchParams.get("code");
  const [isChecking, setIsChecking] = useState(false);

  // Poll order status every 3 seconds
  useEffect(() => {
    if (!orderCode) return;

    const checkOrderStatus = async () => {
      try {
        setIsChecking(true);
        const response = await fetch(`http://localhost:8080/api/orders/${orderCode}`);
        
        if (response.ok) {
          const order = await response.json();
          
          // If order is paid, redirect to success page
          if (order.status === 'PAID' || order.status === 'PROCESSING') {
            console.log('✅ Order paid, redirecting to success page');
            router.push(`/order-success?code=${orderCode}`);
          }
          // If order is cancelled/expired/failed, redirect to failed page
          else if (order.status === 'CANCELLED' || order.status === 'EXPIRED' || order.status === 'FAILED') {
            console.log('❌ Order failed, redirecting to failed page');
            router.push(`/order-failed?code=${orderCode}`);
          }
        }
      } catch (error) {
        console.error('Error checking order status:', error);
      } finally {
        setIsChecking(false);
      }
    };

    // Check immediately on mount
    checkOrderStatus();

    // Then poll every 3 seconds
    const interval = setInterval(checkOrderStatus, 3000);

    // Cleanup on unmount
    return () => clearInterval(interval);
  }, [orderCode, router]);

  return (
    <div className="min-h-screen bg-gradient-to-b from-amber-50 to-white py-16 px-4">
      <div className="max-w-2xl mx-auto">
        {/* Pending Icon */}
        <motion.div
          initial={{ scale: 0 }}
          animate={{ scale: 1 }}
          transition={{ type: "spring", damping: 15, stiffness: 200 }}
          className="text-center mb-8"
        >
          <div className="w-24 h-24 bg-amber-100 rounded-full flex items-center justify-center mx-auto mb-6">
            <motion.svg
              animate={{ rotate: 360 }}
              transition={{ duration: 2, repeat: Infinity, ease: "linear" }}
              className="w-12 h-12 text-amber-600"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
            </motion.svg>
          </div>
          <h1 className="text-3xl font-bold text-gray-900 mb-2">Payment Pending</h1>
          <p className="text-gray-600">Please complete your payment</p>
        </motion.div>

        {/* Info Card */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.3 }}
          className="bg-white rounded-2xl shadow-lg border border-gray-100 overflow-hidden"
        >
          <div className="px-6 py-4 border-b bg-amber-50">
            <div className="flex items-center gap-3">
              <svg className="w-6 h-6 text-amber-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span className="font-medium text-amber-800">
                {isChecking ? 'Checking payment status...' : 'Waiting for Payment'}
              </span>
              {isChecking && (
                <motion.div
                  animate={{ rotate: 360 }}
                  transition={{ duration: 1, repeat: Infinity, ease: "linear" }}
                  className="w-4 h-4 border-2 border-amber-600 border-t-transparent rounded-full"
                />
              )}
            </div>
          </div>

          <div className="px-6 py-6">
            <div className="mb-6">
              <p className="text-sm text-gray-500 mb-1">Order Number</p>
              <p className="font-mono font-bold text-xl">{orderCode}</p>
            </div>

            <div className="bg-gray-50 rounded-lg p-4 mb-6">
              <h3 className="font-semibold mb-3">Next Steps:</h3>
              <ol className="space-y-2 text-sm text-gray-600">
                <li className="flex items-start gap-2">
                  <span className="w-5 h-5 bg-primary text-white rounded-full flex items-center justify-center text-xs flex-shrink-0 mt-0.5">1</span>
                  <span>Complete the payment through your selected payment method</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="w-5 h-5 bg-primary text-white rounded-full flex items-center justify-center text-xs flex-shrink-0 mt-0.5">2</span>
                  <span>You will receive a confirmation email once payment is verified</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="w-5 h-5 bg-primary text-white rounded-full flex items-center justify-center text-xs flex-shrink-0 mt-0.5">3</span>
                  <span>Your order will be processed and shipped</span>
                </li>
              </ol>
            </div>

            <div className="flex items-center gap-2 text-sm text-gray-500 bg-blue-50 rounded-lg p-3">
              <svg className="w-5 h-5 text-blue-600 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span>Payment must be completed within 24 hours or your order will be cancelled</span>
            </div>
          </div>

          {/* Security Badge */}
          <div className="px-6 py-4 border-t flex items-center gap-3 text-sm text-gray-500">
            <svg className="w-5 h-5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
            </svg>
            <span>Secured by Midtrans Payment Gateway</span>
          </div>
        </motion.div>

        {/* Action Buttons */}
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.5 }}
          className="mt-8 flex flex-col sm:flex-row gap-4 justify-center"
        >
          <Link
            href="/"
            className="px-8 py-3 bg-primary text-white text-center font-medium hover:bg-gray-800 transition-colors"
          >
            Continue Shopping
          </Link>
          <Link
            href={`/order-status?code=${orderCode}`}
            className="px-8 py-3 border-2 border-primary text-primary text-center font-medium hover:bg-primary hover:text-white transition-colors"
          >
            Check Status
          </Link>
        </motion.div>
      </div>
    </div>
  );
}

export default function OrderPendingPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-pulse">Loading...</div>
      </div>
    }>
      <OrderPendingContent />
    </Suspense>
  );
}
