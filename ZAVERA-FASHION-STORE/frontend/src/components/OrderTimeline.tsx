"use client";

import { motion } from "framer-motion";

export type OrderStatus = 
  | "PENDING" 
  | "PAID" 
  | "PROCESSING" 
  | "SHIPPED" 
  | "DELIVERED" 
  | "COMPLETED" 
  | "CANCELLED" 
  | "FAILED" 
  | "EXPIRED";

interface TimelineStep {
  key: OrderStatus;
  label: string;
  icon: React.ReactNode;
}

const TIMELINE_STEPS: TimelineStep[] = [
  {
    key: "PENDING",
    label: "Menunggu Pembayaran",
    icon: (
      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
    ),
  },
  {
    key: "PAID",
    label: "Dibayar",
    icon: (
      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
    ),
  },
  {
    key: "PROCESSING",
    label: "Diproses",
    icon: (
      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
      </svg>
    ),
  },
  {
    key: "SHIPPED",
    label: "Dikirim",
    icon: (
      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16V6a1 1 0 00-1-1H4a1 1 0 00-1 1v10a1 1 0 001 1h1m8-1a1 1 0 01-1 1H9m4-1V8a1 1 0 011-1h2.586a1 1 0 01.707.293l3.414 3.414a1 1 0 01.293.707V16a1 1 0 01-1 1h-1m-6-1a1 1 0 001 1h1M5 17a2 2 0 104 0m-4 0a2 2 0 114 0m6 0a2 2 0 104 0m-4 0a2 2 0 114 0" />
      </svg>
    ),
  },
  {
    key: "DELIVERED",
    label: "Terkirim",
    icon: (
      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
      </svg>
    ),
  },
];

// Status color mapping
export const STATUS_COLORS: Record<OrderStatus, string> = {
  PENDING: "bg-yellow-100 text-yellow-800 border-yellow-200",
  PAID: "bg-blue-100 text-blue-800 border-blue-200",
  PROCESSING: "bg-purple-100 text-purple-800 border-purple-200",
  SHIPPED: "bg-indigo-100 text-indigo-800 border-indigo-200",
  DELIVERED: "bg-green-100 text-green-800 border-green-200",
  COMPLETED: "bg-green-100 text-green-800 border-green-200",
  CANCELLED: "bg-red-100 text-red-800 border-red-200",
  FAILED: "bg-red-100 text-red-800 border-red-200",
  EXPIRED: "bg-gray-100 text-gray-800 border-gray-200",
};

// Status labels in Indonesian
export const STATUS_LABELS: Record<OrderStatus, string> = {
  PENDING: "Menunggu Pembayaran",
  PAID: "Dibayar",
  PROCESSING: "Diproses",
  SHIPPED: "Dikirim",
  DELIVERED: "Terkirim",
  COMPLETED: "Selesai",
  CANCELLED: "Dibatalkan",
  FAILED: "Gagal",
  EXPIRED: "Kadaluarsa",
};

interface OrderTimelineProps {
  status: OrderStatus;
  createdAt?: string;
  compact?: boolean;
}

// Get the index of a status in the timeline
export function getStatusIndex(status: OrderStatus): number {
  const index = TIMELINE_STEPS.findIndex((step) => step.key === status);
  // For completed, use delivered index
  if (status === "COMPLETED") return TIMELINE_STEPS.length - 1;
  return index;
}

// Check if a step is completed based on current status
export function isStepCompleted(stepKey: OrderStatus, currentStatus: OrderStatus): boolean {
  const stepIndex = getStatusIndex(stepKey);
  const currentIndex = getStatusIndex(currentStatus);
  
  // Handle terminal states
  if (["CANCELLED", "FAILED", "EXPIRED"].includes(currentStatus)) {
    return false;
  }
  
  return stepIndex <= currentIndex;
}

// Check if a step is the current step
export function isCurrentStep(stepKey: OrderStatus, currentStatus: OrderStatus): boolean {
  if (currentStatus === "COMPLETED" && stepKey === "DELIVERED") return true;
  return stepKey === currentStatus;
}

export default function OrderTimeline({ status, compact = false }: OrderTimelineProps) {
  // Don't show timeline for cancelled/failed/expired orders
  if (["CANCELLED", "FAILED", "EXPIRED"].includes(status)) {
    return (
      <div className="flex items-center gap-2 py-2">
        <div className={`w-8 h-8 rounded-full flex items-center justify-center ${
          status === "CANCELLED" || status === "FAILED" 
            ? "bg-red-100 text-red-600" 
            : "bg-gray-100 text-gray-600"
        }`}>
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </div>
        <span className={`text-sm font-medium ${
          status === "CANCELLED" || status === "FAILED" 
            ? "text-red-600" 
            : "text-gray-600"
        }`}>
          {STATUS_LABELS[status]}
        </span>
      </div>
    );
  }

  if (compact) {
    return (
      <div className="flex items-center gap-1">
        {TIMELINE_STEPS.map((step, index) => {
          const completed = isStepCompleted(step.key, status);
          const current = isCurrentStep(step.key, status);
          
          return (
            <div key={step.key} className="flex items-center">
              <motion.div
                initial={{ scale: 0.8, opacity: 0 }}
                animate={{ scale: 1, opacity: 1 }}
                transition={{ delay: index * 0.1 }}
                className={`w-6 h-6 rounded-full flex items-center justify-center transition-colors ${
                  completed
                    ? current
                      ? "bg-primary text-white"
                      : "bg-green-500 text-white"
                    : "bg-gray-200 text-gray-400"
                }`}
                title={step.label}
              >
                {completed && !current ? (
                  <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
                  </svg>
                ) : (
                  <span className="text-xs">{index + 1}</span>
                )}
              </motion.div>
              {index < TIMELINE_STEPS.length - 1 && (
                <div className={`w-4 h-0.5 ${
                  isStepCompleted(TIMELINE_STEPS[index + 1].key, status)
                    ? "bg-green-500"
                    : "bg-gray-200"
                }`} />
              )}
            </div>
          );
        })}
      </div>
    );
  }

  return (
    <div className="py-4">
      <div className="flex items-center justify-between">
        {TIMELINE_STEPS.map((step, index) => {
          const completed = isStepCompleted(step.key, status);
          const current = isCurrentStep(step.key, status);
          
          return (
            <div key={step.key} className="flex-1 flex flex-col items-center relative">
              {/* Connector line */}
              {index > 0 && (
                <div
                  className={`absolute top-4 right-1/2 w-full h-0.5 -translate-y-1/2 ${
                    completed ? "bg-green-500" : "bg-gray-200"
                  }`}
                  style={{ width: "calc(100% - 2rem)", left: "calc(-50% + 1rem)" }}
                />
              )}
              
              {/* Step circle */}
              <motion.div
                initial={{ scale: 0.8, opacity: 0 }}
                animate={{ scale: 1, opacity: 1 }}
                transition={{ delay: index * 0.1 }}
                className={`relative z-10 w-8 h-8 rounded-full flex items-center justify-center transition-all ${
                  completed
                    ? current
                      ? "bg-primary text-white ring-4 ring-primary/20"
                      : "bg-green-500 text-white"
                    : "bg-gray-200 text-gray-400"
                }`}
              >
                {completed && !current ? (
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                  </svg>
                ) : (
                  step.icon
                )}
              </motion.div>
              
              {/* Step label */}
              <span className={`mt-2 text-xs text-center ${
                current ? "text-primary font-medium" : completed ? "text-gray-700" : "text-gray-400"
              }`}>
                {step.label}
              </span>
            </div>
          );
        })}
      </div>
    </div>
  );
}

// Status Badge Component
interface StatusBadgeProps {
  status: OrderStatus;
  size?: "sm" | "md";
}

export function StatusBadge({ status, size = "md" }: StatusBadgeProps) {
  const sizeClasses = size === "sm" 
    ? "px-2 py-0.5 text-xs" 
    : "px-2.5 py-1 text-sm";
  
  return (
    <span className={`inline-flex items-center rounded-full font-medium border ${STATUS_COLORS[status]} ${sizeClasses}`}>
      {STATUS_LABELS[status]}
    </span>
  );
}
