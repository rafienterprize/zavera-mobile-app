import api from "./api";

// Types
export interface CreateVAPaymentRequest {
  order_id: number;
  payment_method: "bca_va" | "bri_va" | "mandiri_va";
}

export interface PaymentDetails {
  payment_id: number;
  order_id: number;
  order_code: string;
  payment_method: string;
  bank: string;
  bank_logo: string;
  va_number: string;
  amount: number;
  expiry_time: string;
  remaining_seconds: number;
  status: string;
  instructions: Array<{ channel: string; steps: string[] }>;
}

export interface PaymentStatusResponse {
  payment_id: number;
  status: string;
  message: string;
}

export interface PendingOrder {
  order_id: number;
  order_code: string;
  total_amount: number;
  item_count: number;
  item_summary: string;
  created_at: string;
  has_payment: boolean;
  payment_method?: string;
  bank?: string;
  bank_logo?: string;
  va_number_masked?: string;
  expiry_time?: string;
  remaining_seconds?: number;
}

export interface PendingOrdersResponse {
  orders: PendingOrder[];
  total_count: number;
  page: number;
  page_size: number;
}

export interface TransactionHistoryItem {
  order_id: number;
  order_code: string;
  total_amount: number;
  item_count: number;
  item_summary: string;
  status: string;
  payment_method?: string;
  paid_at?: string;
  created_at: string;
}

export interface TransactionHistoryResponse {
  orders: TransactionHistoryItem[];
  total_count: number;
  page: number;
  page_size: number;
}

// Core Payment API functions
export const corePaymentApi = {
  /**
   * Create VA payment via Midtrans Core API
   * Returns existing payment if one already exists (idempotent)
   */
  createVAPayment: async (request: CreateVAPaymentRequest): Promise<PaymentDetails> => {
    const response = await api.post("/payments/core/create", request);
    return response.data;
  },

  /**
   * Get payment details for an order
   * Triggers expiry check if payment is PENDING and expired
   */
  getPaymentDetails: async (orderId: number): Promise<PaymentDetails> => {
    const response = await api.get(`/payments/core/${orderId}`);
    return response.data;
  },

  /**
   * Check payment status (database only, no Midtrans API call)
   * Rate limited: max 1 request per 5 seconds
   */
  checkPaymentStatus: async (paymentId: number): Promise<PaymentStatusResponse> => {
    const response = await api.post("/payments/core/check", { payment_id: paymentId });
    return response.data;
  },

  /**
   * Get pending orders for Menunggu Pembayaran tab
   */
  getPendingOrders: async (page: number = 1, pageSize: number = 10): Promise<PendingOrdersResponse> => {
    const response = await api.get(`/pembelian/pending?page=${page}&page_size=${pageSize}`);
    return response.data;
  },

  /**
   * Get transaction history for Daftar Transaksi tab
   */
  getTransactionHistory: async (page: number = 1, pageSize: number = 10): Promise<TransactionHistoryResponse> => {
    const response = await api.get(`/pembelian/history?page=${page}&page_size=${pageSize}`);
    return response.data;
  },
};

export default corePaymentApi;
