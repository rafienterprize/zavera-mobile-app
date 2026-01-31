import api from "./api";

// Helper to get auth header
const getAuthHeader = () => {
  const token = typeof window !== "undefined" ? localStorage.getItem("auth_token") : null;
  return token ? { Authorization: `Bearer ${token}` } : {};
};

// ============================================
// PRODUCTS (ADMIN)
// ============================================

export interface AdminProduct {
  id: number;
  name: string;
  slug: string;
  description: string;
  price: number;
  stock: number;
  weight: number;
  category: string;
  subcategory: string;
  is_active: boolean;
  images: ProductImage[];
  created_at: string;
  updated_at: string;
}

export interface ProductImage {
  id: number;
  image_url: string;
  is_primary: boolean;
  display_order: number;
}

export const getAdminProducts = async (
  page = 1,
  pageSize = 20,
  category = "",
  includeInactive = true
): Promise<{ products: AdminProduct[]; total: number }> => {
  const params = new URLSearchParams({
    page: page.toString(),
    page_size: pageSize.toString(),
    include_inactive: includeInactive.toString(),
  });
  if (category) params.append("category", category);

  const response = await api.get(`/admin/products?${params.toString()}`, {
    headers: getAuthHeader(),
  });
  return {
    products: response.data.products || [],
    total: response.data.total_count || 0,
  };
};

export const getProduct = async (token: string, id: number): Promise<AdminProduct> => {
  const response = await api.get(`/products/${id}`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  return response.data;
};

export const createProduct = async (data: {
  name: string;
  description?: string;
  price: number;
  stock: number;
  weight?: number;
  category: string;
  subcategory?: string;
  images?: string[];
}) => {
  return api.post("/admin/products", data, { headers: getAuthHeader() });
};

export const updateProduct = async (
  id: number,
  data: {
    name?: string;
    description?: string;
    price?: number;
    stock?: number;
    weight?: number;
    category?: string;
    subcategory?: string;
    is_active?: boolean;
  }
) => {
  return api.put(`/admin/products/${id}`, data, { headers: getAuthHeader() });
};

export const updateStock = async (id: number, quantity: number, reason?: string) => {
  return api.patch(`/admin/products/${id}/stock`, { quantity, reason }, { headers: getAuthHeader() });
};

export const deleteProduct = async (id: number) => {
  return api.delete(`/admin/products/${id}`, { headers: getAuthHeader() });
};

export const addProductImage = async (productId: number, imageUrl: string, isPrimary = false) => {
  return api.post(
    `/admin/products/${productId}/images`,
    { image_url: imageUrl, is_primary: isPrimary },
    { headers: getAuthHeader() }
  );
};

export const deleteProductImage = async (productId: number, imageId: number) => {
  return api.delete(`/admin/products/${productId}/images/${imageId}`, { headers: getAuthHeader() });
};

// ============================================
// ORDER STATS
// ============================================

export interface OrderStats {
  total_orders: number;
  total_revenue: number;
  pending_orders: number;
  paid_orders: number;
  processing_orders: number;
  shipped_orders: number;
  delivered_orders: number;
  cancelled_orders: number;
  today_orders: number;
  today_revenue: number;
}

export const getOrderStats = async (): Promise<OrderStats> => {
  const response = await api.get("/admin/orders/stats", { headers: getAuthHeader() });
  return response.data;
};

// ============================================
// DASHBOARD
// ============================================

export interface DashboardStats {
  financial: {
    total_paid_today: number;
    total_refunded_today: number;
    stuck_payments: number;
    reconciliation_mismatches: number;
  };
  fulfillment: {
    in_transit: number;
    delayed: number;
    lost: number;
    pickup_failures: number;
  };
  disputes: {
    open: number;
    investigating: number;
    evidence_required: number;
  };
}

export const getDashboardStats = async (): Promise<DashboardStats> => {
  // Aggregate from multiple endpoints
  const [fulfillment, stuckPayments, reconciliation] = await Promise.all([
    api.get("/admin/fulfillment/dashboard", { headers: getAuthHeader() }).catch(() => ({ data: {} })),
    api.get("/admin/payments/stuck", { headers: getAuthHeader() }).catch(() => ({ data: { stuck_payments: [] } })),
    api.get("/admin/reconciliation/mismatches", { headers: getAuthHeader() }).catch(() => ({ data: { mismatches: [] } })),
  ]);

  const fd = fulfillment.data || {};
  
  return {
    financial: {
      total_paid_today: 0, // Will be calculated from orders
      total_refunded_today: 0,
      stuck_payments: stuckPayments.data?.stuck_payments?.length || 0,
      reconciliation_mismatches: reconciliation.data?.mismatches?.length || 0,
    },
    fulfillment: {
      in_transit: fd.status_counts?.IN_TRANSIT || 0,
      delayed: fd.stuck_shipments || 0,
      lost: fd.lost_packages || 0,
      pickup_failures: fd.pickup_failures || 0,
    },
    disputes: {
      open: fd.open_disputes || 0,
      investigating: 0,
      evidence_required: 0,
    },
  };
};

// ============================================
// ORDERS
// ============================================

export interface Order {
  id: number;
  order_code: string;
  customer_name: string;
  customer_email: string;
  customer_phone: string;
  total_amount: number;
  status: string;
  payment_status?: string;
  shipment_status?: string;
  refund_status?: string;
  created_at: string;
  items: OrderItem[];
}

export interface OrderItem {
  product_id: number;
  product_name: string;
  quantity: number;
  price_per_unit: number;
  subtotal: number;
}

export const getOrders = async (page = 1, limit = 20, status = "", search = ""): Promise<{ orders: Order[]; total: number }> => {
  const params = new URLSearchParams({
    page: page.toString(),
    page_size: limit.toString(),
  });
  if (status) params.append("status", status);
  if (search) params.append("search", search);
  
  const response = await api.get(`/admin/orders?${params.toString()}`, {
    headers: getAuthHeader(),
  });
  return {
    orders: response.data.orders || [],
    total: response.data.total_count || 0,
  };
};

export const getOrderByCode = async (code: string): Promise<Order> => {
  const response = await api.get(`/orders/${code}`, { headers: getAuthHeader() });
  return response.data;
};

// ============================================
// FORCE ACTIONS
// ============================================

export const forceCancel = async (orderCode: string, reason: string) => {
  return api.post(`/admin/orders/${orderCode}/force-cancel`, { 
    reason,
    restore_stock: true 
  }, { headers: getAuthHeader() });
};

export const forceRefund = async (orderCode: string, data: { refund_type: string; reason: string; amount?: number }) => {
  return api.post(`/admin/orders/${orderCode}/refund`, data, { headers: getAuthHeader() });
};

export const forceReship = async (orderCode: string, reason: string) => {
  return api.post(`/admin/orders/${orderCode}/reship`, { reason }, { headers: getAuthHeader() });
};

// ============================================
// REFUNDS
// ============================================

export interface Refund {
  id: number;
  refund_code: string;
  order_code: string;
  refund_type: string;
  original_amount: number;
  refund_amount: number;
  status: string;
  gateway_status?: string;
  reason: string;
  created_at: string;
  processed_at?: string;
}

export const getRefunds = async (): Promise<Refund[]> => {
  // Would need a list refunds endpoint
  return [];
};

export const getRefundByCode = async (code: string): Promise<Refund> => {
  const response = await api.get(`/admin/refunds/${code}`, { headers: getAuthHeader() });
  return response.data;
};

export const processRefund = async (code: string) => {
  return api.post(`/admin/refunds/${code}/process`, {}, { headers: getAuthHeader() });
};

export const createRefund = async (data: {
  order_code: string;
  refund_type: string;
  reason: string;
  amount?: number;
  items?: { order_item_id: number; quantity: number }[];
}) => {
  return api.post("/admin/refunds", data, { headers: getAuthHeader() });
};

// ============================================
// SHIPMENTS
// ============================================

export interface Shipment {
  id: number;
  order_id: number;
  order_code?: string;
  tracking_number: string;
  status: string;
  provider_code: string;
  provider_name: string;
  days_without_update: number;
  requires_admin_action: boolean;
  admin_action_reason?: string;
  created_at: string;
  shipped_at?: string;
  delivered_at?: string;
}

export const getShipments = async (): Promise<Shipment[]> => {
  // Would aggregate from stuck + pickup failures
  const [stuck, pickupFailures] = await Promise.all([
    api.get("/admin/shipments/stuck?days=0", { headers: getAuthHeader() }).catch(() => ({ data: { stuck_shipments: [] } })),
    api.get("/admin/shipments/pickup-failures", { headers: getAuthHeader() }).catch(() => ({ data: { pickup_failures: [] } })),
  ]);
  
  return [...(stuck.data?.stuck_shipments || []), ...(pickupFailures.data?.pickup_failures || [])];
};

export const getShipmentDetails = async (id: number) => {
  const response = await api.get(`/admin/shipments/${id}/details`, { headers: getAuthHeader() });
  return response.data;
};

export const investigateShipment = async (id: number, reason: string) => {
  return api.post(`/admin/shipments/${id}/investigate`, { reason }, { headers: getAuthHeader() });
};

export const markShipmentLost = async (id: number, data: { reason: string; create_dispute?: boolean }) => {
  return api.post(`/admin/shipments/${id}/mark-lost`, data, { headers: getAuthHeader() });
};

export const reshipShipment = async (id: number, data: { reason: string; cost_bearer: string }) => {
  return api.post(`/admin/shipments/${id}/reship`, data, { headers: getAuthHeader() });
};

export const overrideShipmentStatus = async (id: number, data: { new_status: string; reason: string; bypass_validation?: boolean }) => {
  return api.post(`/admin/shipments/${id}/override-status`, data, { headers: getAuthHeader() });
};

// ============================================
// DISPUTES
// ============================================

export interface Dispute {
  id: number;
  dispute_code: string;
  order_id: number;
  order_code?: string;
  shipment_id?: number;
  dispute_type: string;
  status: string;
  title: string;
  description: string;
  customer_email: string;
  customer_claim?: string;
  evidence_urls?: string[];
  resolution?: string;
  resolution_notes?: string;
  resolution_amount?: number;
  created_at: string;
  resolved_at?: string;
  messages?: DisputeMessage[];
}

export interface DisputeMessage {
  id: number;
  sender_type: string;
  sender_name?: string;
  message: string;
  attachment_urls?: string[];
  is_internal: boolean;
  created_at: string;
}

export const getOpenDisputes = async (): Promise<Dispute[]> => {
  const response = await api.get("/admin/disputes/open", { headers: getAuthHeader() });
  return response.data?.disputes || [];
};

export const getSystemHealth = async () => {
  const response = await api.get("/admin/system/health", { headers: getAuthHeader() });
  return response.data;
};

export const getCourierPerformance = async () => {
  const response = await api.get("/admin/analytics/courier-performance", { headers: getAuthHeader() });
  return response.data;
};

export const getShipmentsList = async (status?: string, page: number = 1) => {
  const params = new URLSearchParams();
  if (status) params.append("status", status);
  params.append("page", page.toString());
  
  const response = await api.get(`/admin/shipments?${params.toString()}`, { headers: getAuthHeader() });
  return response.data;
};

export const getDisputeById = async (id: number): Promise<Dispute> => {
  const response = await api.get(`/admin/disputes/${id}`, { headers: getAuthHeader() });
  return response.data;
};

export const getDisputeByCode = async (code: string): Promise<Dispute> => {
  const response = await api.get(`/admin/disputes/code/${code}`, { headers: getAuthHeader() });
  return response.data;
};

export const startDisputeInvestigation = async (id: number) => {
  return api.post(`/admin/disputes/${id}/investigate`, {}, { headers: getAuthHeader() });
};

export const requestDisputeEvidence = async (id: number, message: string) => {
  return api.post(`/admin/disputes/${id}/request-evidence`, { message }, { headers: getAuthHeader() });
};

export const resolveDispute = async (id: number, data: {
  resolution: string;
  resolution_notes: string;
  resolution_amount?: number;
  create_refund?: boolean;
  create_reship?: boolean;
}) => {
  return api.post(`/admin/disputes/${id}/resolve`, data, { headers: getAuthHeader() });
};

export const closeDispute = async (id: number) => {
  return api.post(`/admin/disputes/${id}/close`, {}, { headers: getAuthHeader() });
};

export const addDisputeMessage = async (id: number, data: { message: string; is_internal?: boolean }) => {
  return api.post(`/admin/disputes/${id}/messages`, data, { headers: getAuthHeader() });
};

export const getDisputeMessages = async (id: number, includeInternal = true): Promise<DisputeMessage[]> => {
  const response = await api.get(`/admin/disputes/${id}/messages?include_internal=${includeInternal}`, {
    headers: getAuthHeader(),
  });
  return response.data?.messages || [];
};

// ============================================
// AUDIT
// ============================================

export interface AuditLog {
  id: number;
  admin_email: string;
  action_type: string;
  action_detail: string;
  target_type: string;
  target_id: number;
  success: boolean;
  created_at: string;
}

export const getAuditLogs = async (page = 1, limit = 50): Promise<{ logs: AuditLog[]; total: number }> => {
  const response = await api.get(`/admin/audit-logs?page=${page}&limit=${limit}`, { headers: getAuthHeader() });
  return {
    logs: response.data?.logs || [],
    total: response.data?.total || 0,
  };
};

// ============================================
// RECONCILIATION
// ============================================

export const runReconciliation = async () => {
  return api.post("/admin/reconciliation/run", {}, { headers: getAuthHeader() });
};

export const getReconciliation = async (date?: string) => {
  const url = date ? `/admin/reconciliation?date=${date}` : "/admin/reconciliation";
  const response = await api.get(url, { headers: getAuthHeader() });
  return response.data;
};

// ============================================
// PAYMENT RECOVERY
// ============================================

export const syncPayment = async (paymentId: number) => {
  return api.post(`/admin/payments/${paymentId}/sync`, {}, { headers: getAuthHeader() });
};

export const getStuckPayments = async () => {
  const response = await api.get("/admin/payments/stuck", { headers: getAuthHeader() });
  return response.data?.stuck_payments || [];
};

export const runPaymentSync = async () => {
  return api.post("/admin/payments/sync-all", {}, { headers: getAuthHeader() });
};

// ============================================
// MONITORING
// ============================================

export const runMonitors = async () => {
  return api.post("/admin/fulfillment/run-monitors", {}, { headers: getAuthHeader() });
};

export const getFulfillmentDashboard = async () => {
  const response = await api.get("/admin/fulfillment/dashboard", { headers: getAuthHeader() });
  return response.data;
};

// ============================================
// EXECUTIVE DASHBOARD (P0)
// ============================================

export interface ExecutiveMetrics {
  gmv: number;
  revenue: number;
  pending_revenue: number;
  total_orders: number;
  paid_orders: number;
  avg_order_value: number;
  conversion_rate: number;
  payment_methods: Array<{ method: string; count: number; amount: number }>;
  top_products: Array<{ product_id: number; product_name: string; total_sold: number; revenue: number }>;
}

export interface PaymentMonitor {
  pending_count: number;
  pending_amount: number;
  expiring_soon_count: number;
  expiring_soon_amount: number;
  stuck_payments: Array<{
    payment_id: number;
    order_code: string;
    payment_method: string;
    bank: string;
    amount: number;
    created_at: string;
    hours_pending: number;
  }>;
  today_paid_count: number;
  today_paid_amount: number;
  method_performance: Array<{ method: string; count: number; avg_time_minutes: number }>;
}

export interface InventoryAlerts {
  out_of_stock: Array<{
    product_id: number;
    product_name: string;
    stock: number;
    price: number;
    category: string;
    severity: string;
  }>;
  low_stock: Array<{
    product_id: number;
    product_name: string;
    stock: number;
    price: number;
    category: string;
    severity: string;
  }>;
  fast_moving: Array<{
    product_id: number;
    product_name: string;
    stock: number;
    price: number;
    category: string;
    orders_count: number;
    total_sold: number;
    days_of_stock: number;
  }>;
}

export interface CustomerInsights {
  total_customers: number;
  active_customers: number;
  new_customers: number;
  segments: Array<{ segment: string; count: number; avg_value: number }>;
  top_customers: Array<{
    email: string;
    name: string;
    total_orders: number;
    total_spent: number;
    last_order: string;
  }>;
}

export interface ConversionFunnel {
  orders_created: number;
  orders_paid: number;
  orders_shipped: number;
  orders_delivered: number;
  orders_completed: number;
  payment_rate: number;
  fulfillment_rate: number;
  delivery_rate: number;
  completion_rate: number;
  drop_offs: Array<{ stage: string; count: number; percentage: number }>;
}

export interface RevenueChart {
  data_points: Array<{ date: string; orders: number; revenue: number }>;
}

export const getExecutiveMetrics = async (period = "today"): Promise<ExecutiveMetrics> => {
  const response = await api.get(`/admin/dashboard/executive?period=${period}`, { headers: getAuthHeader() });
  return response.data;
};

export const getPaymentMonitor = async (): Promise<PaymentMonitor> => {
  const response = await api.get("/admin/dashboard/payments", { headers: getAuthHeader() });
  return response.data;
};

export const getInventoryAlerts = async (): Promise<InventoryAlerts> => {
  const response = await api.get("/admin/dashboard/inventory", { headers: getAuthHeader() });
  return response.data;
};

export const getCustomerInsights = async (): Promise<CustomerInsights> => {
  const response = await api.get("/admin/dashboard/customers", { headers: getAuthHeader() });
  return response.data;
};

export const getConversionFunnel = async (period = "today"): Promise<ConversionFunnel> => {
  const response = await api.get(`/admin/dashboard/funnel?period=${period}`, { headers: getAuthHeader() });
  return response.data;
};

export const getRevenueChart = async (period = "7days"): Promise<RevenueChart> => {
  const response = await api.get(`/admin/dashboard/revenue-chart?period=${period}`, { headers: getAuthHeader() });
  return response.data;
};
