package dto

import "time"

// ============================================
// ADMIN PRODUCT DTOs
// ============================================

type CreateProductRequest struct {
	Name        string   `json:"name" binding:"required"`
	Slug        string   `json:"slug"`
	Description string   `json:"description"`
	Price       float64  `json:"price" binding:"required,gt=0"`
	Stock       int      `json:"stock" binding:"gte=0"`
	Weight      int      `json:"weight"` // in grams, default 500
	Length      int      `json:"length"` // in cm, default 30
	Width       int      `json:"width"`  // in cm, default 20
	Height      int      `json:"height"` // in cm, default 5
	Category    string   `json:"category" binding:"required"`
	Subcategory string   `json:"subcategory"`
	Brand       string   `json:"brand"`    // Product brand (e.g., Nike, Adidas)
	Material    string   `json:"material"` // Product material (e.g., Cotton, Polyester)
	IsActive    *bool    `json:"is_active"`
	Images      []string `json:"images"` // URLs
}

type UpdateProductRequest struct {
	Name        *string  `json:"name"`
	Slug        *string  `json:"slug"`
	Description *string  `json:"description"`
	Price       *float64 `json:"price"`
	Stock       *int     `json:"stock"`
	Weight      *int     `json:"weight"`
	Length      *int     `json:"length"`
	Width       *int     `json:"width"`
	Height      *int     `json:"height"`
	Category    *string  `json:"category"`
	Subcategory *string  `json:"subcategory"`
	Brand       *string  `json:"brand"`
	Material    *string  `json:"material"`
	IsActive    *bool    `json:"is_active"`
}

type UpdateStockRequest struct {
	Quantity int    `json:"quantity" binding:"required"` // positive = add, negative = subtract
	Reason   string `json:"reason"`                      // "restock", "adjustment", "damage", etc.
}

type AddProductImageRequest struct {
	ImageURL     string `json:"image_url" binding:"required"`
	IsPrimary    bool   `json:"is_primary"`
	DisplayOrder int    `json:"display_order"`
}

type AdminProductResponse struct {
	ID          int                   `json:"id"`
	Name        string                `json:"name"`
	Slug        string                `json:"slug"`
	Description string                `json:"description"`
	Price       float64               `json:"price"`
	Stock       int                   `json:"stock"`
	Weight      int                   `json:"weight"`
	Length      int                   `json:"length"`
	Width       int                   `json:"width"`
	Height      int                   `json:"height"`
	Category    string                `json:"category"`
	Subcategory string                `json:"subcategory"`
	Brand       string                `json:"brand"`
	Material    string                `json:"material"`
	IsActive    bool                  `json:"is_active"`
	Images      []ProductImageResponse `json:"images"`
	CreatedAt   string                `json:"created_at"`
	UpdatedAt   string                `json:"updated_at"`
}

type ProductImageResponse struct {
	ID           int    `json:"id"`
	ImageURL     string `json:"image_url"`
	IsPrimary    bool   `json:"is_primary"`
	DisplayOrder int    `json:"display_order"`
}

// ============================================
// ADMIN ORDER DTOs
// ============================================

type AdminOrderFilter struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Status   string `json:"status"`
	Search   string `json:"search"` // order_code, customer_name, email
	DateFrom string `json:"date_from"`
	DateTo   string `json:"date_to"`
}

type AdminOrderResponse struct {
	ID            int                    `json:"id"`
	OrderCode     string                 `json:"order_code"`
	CustomerName  string                 `json:"customer_name"`
	CustomerEmail string                 `json:"customer_email"`
	CustomerPhone string                 `json:"customer_phone"`
	Subtotal      float64                `json:"subtotal"`
	ShippingCost  float64                `json:"shipping_cost"`
	Tax           float64                `json:"tax"`
	Discount      float64                `json:"discount"`
	TotalAmount   float64                `json:"total_amount"`
	Status        string                 `json:"status"`
	Resi          string                 `json:"resi,omitempty"`
	Items         []OrderItemResponse    `json:"items"`
	Payment       *AdminPaymentInfo      `json:"payment,omitempty"`
	Shipment      *AdminShipmentInfo     `json:"shipment,omitempty"`
	CreatedAt     string                 `json:"created_at"`
	UpdatedAt     string                 `json:"updated_at"`
	PaidAt        *string                `json:"paid_at,omitempty"`
	ShippedAt     *string                `json:"shipped_at,omitempty"`
}

type AdminPaymentInfo struct {
	ID              int     `json:"id"`
	Status          string  `json:"status"`
	PaymentMethod   string  `json:"payment_method"`
	PaymentProvider string  `json:"payment_provider"`
	Amount          float64 `json:"amount"`
	TransactionID   string  `json:"transaction_id,omitempty"`
	PaidAt          *string `json:"paid_at,omitempty"`
}

type AdminShipmentInfo struct {
	ID              int     `json:"id"`
	ProviderCode    string  `json:"provider_code"`
	ProviderName    string  `json:"provider_name"`
	ServiceCode     string  `json:"service_code"`
	ServiceName     string  `json:"service_name"`
	TrackingNumber  string  `json:"tracking_number,omitempty"`
	Status          string  `json:"status"`
	Cost            float64 `json:"cost"`
	ETD             string  `json:"etd"`
	Weight          int     `json:"weight"`
	OriginCity      string  `json:"origin_city"`
	DestinationCity string  `json:"destination_city"`
	ShippedAt       *string `json:"shipped_at,omitempty"`
	DeliveredAt     *string `json:"delivered_at,omitempty"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required"`
	Reason string `json:"reason"`
}

// ShipOrderRequest for shipping an order
type ShipOrderRequest struct {
	Resi string `json:"resi"` // Optional - will be auto-generated if empty
}

// CancelOrderRequest for cancelling an order
type CancelOrderRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// OrderAction represents an available action for an order
type OrderAction struct {
	Action      string `json:"action"`
	Label       string `json:"label"`
	Enabled     bool   `json:"enabled"`
	Description string `json:"description,omitempty"`
}

type OrderStatsResponse struct {
	TotalOrders     int     `json:"total_orders"`
	TotalRevenue    float64 `json:"total_revenue"`
	PendingOrders   int     `json:"pending_orders"`
	PaidOrders      int     `json:"paid_orders"`
	ProcessingOrders int    `json:"processing_orders"`
	ShippedOrders   int     `json:"shipped_orders"`
	DeliveredOrders int     `json:"delivered_orders"`
	CancelledOrders int     `json:"cancelled_orders"`
	TodayOrders     int     `json:"today_orders"`
	TodayRevenue    float64 `json:"today_revenue"`
}

// ============================================
// ADMIN DASHBOARD DTOs
// ============================================

type AdminDashboardStats struct {
	Orders    OrderStatsResponse    `json:"orders"`
	Products  ProductStatsResponse  `json:"products"`
	Revenue   RevenueStatsResponse  `json:"revenue"`
}

type ProductStatsResponse struct {
	TotalProducts   int `json:"total_products"`
	ActiveProducts  int `json:"active_products"`
	LowStockProducts int `json:"low_stock_products"` // stock < 10
	OutOfStock      int `json:"out_of_stock"`
}

type RevenueStatsResponse struct {
	Today     float64 `json:"today"`
	ThisWeek  float64 `json:"this_week"`
	ThisMonth float64 `json:"this_month"`
	Total     float64 `json:"total"`
}

// Helper to format time
func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func FormatTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02 15:04:05")
	return &s
}


// ============================================
// DASHBOARD METRICS DTOs
// ============================================

// DashboardMetricsResponse contains all dashboard metrics
type DashboardMetricsResponse struct {
	TotalRevenue      float64              `json:"total_revenue"`
	OrdersToday       int                  `json:"orders_today"`
	OrdersShipped     int                  `json:"orders_shipped"`
	OrdersPending     int                  `json:"orders_pending"`
	OrdersPacking     int                  `json:"orders_packing"`
	LowStockProducts  []LowStockProduct    `json:"low_stock_products"`
	RecentOrders      []RecentOrderSummary `json:"recent_orders"`
}

// LowStockProduct represents a product with low stock
type LowStockProduct struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Stock    int     `json:"stock"`
	Price    float64 `json:"price"`
	Category string  `json:"category"`
}

// RecentOrderSummary represents a recent order for dashboard
type RecentOrderSummary struct {
	OrderCode    string  `json:"order_code"`
	CustomerName string  `json:"customer_name"`
	TotalAmount  float64 `json:"total_amount"`
	Status       string  `json:"status"`
	CreatedAt    string  `json:"created_at"`
}

// ============================================
// EXECUTIVE DASHBOARD DTOs
// ============================================

// ExecutiveMetrics contains high-level business metrics
type ExecutiveMetrics struct {
	GMV              float64              `json:"gmv"`                // Gross Merchandise Value
	Revenue          float64              `json:"revenue"`            // Actually paid
	PendingRevenue   float64              `json:"pending_revenue"`    // Awaiting payment
	TotalOrders      int                  `json:"total_orders"`
	PaidOrders       int                  `json:"paid_orders"`
	AvgOrderValue    float64              `json:"avg_order_value"`
	ConversionRate   float64              `json:"conversion_rate"`    // % of orders that get paid
	PaymentMethods   []PaymentMethodStat  `json:"payment_methods"`
	TopProducts      []TopProductStat     `json:"top_products"`
}

type PaymentMethodStat struct {
	Method string  `json:"method"`
	Count  int     `json:"count"`
	Amount float64 `json:"amount"`
}

type TopProductStat struct {
	ProductID   int     `json:"product_id"`
	ProductName string  `json:"product_name"`
	TotalSold   int     `json:"total_sold"`
	Revenue     float64 `json:"revenue"`
}

// ============================================
// PAYMENT MONITOR DTOs
// ============================================

// PaymentMonitor contains real-time payment monitoring data
type PaymentMonitor struct {
	PendingCount        int                        `json:"pending_count"`
	PendingAmount       float64                    `json:"pending_amount"`
	ExpiringSoonCount   int                        `json:"expiring_soon_count"`
	ExpiringSoonAmount  float64                    `json:"expiring_soon_amount"`
	StuckPayments       []StuckPayment             `json:"stuck_payments"`
	TodayPaidCount      int                        `json:"today_paid_count"`
	TodayPaidAmount     float64                    `json:"today_paid_amount"`
	MethodPerformance   []PaymentMethodPerformance `json:"method_performance"`
}

type StuckPayment struct {
	PaymentID     int       `json:"payment_id"`
	OrderCode     string    `json:"order_code"`
	PaymentMethod string    `json:"payment_method"`
	Bank          string    `json:"bank"`
	Amount        float64   `json:"amount"`
	CreatedAt     time.Time `json:"created_at"`
	HoursPending  float64   `json:"hours_pending"`
}

type PaymentMethodPerformance struct {
	Method         string  `json:"method"`
	Count          int     `json:"count"`
	AvgTimeMinutes float64 `json:"avg_time_minutes"`
}

// ============================================
// INVENTORY ALERTS DTOs
// ============================================

// InventoryAlerts contains stock alerts
type InventoryAlerts struct {
	OutOfStock  []ProductStockAlert  `json:"out_of_stock"`
	LowStock    []ProductStockAlert  `json:"low_stock"`
	FastMoving  []FastMovingProduct  `json:"fast_moving"`
}

type ProductStockAlert struct {
	ProductID   int     `json:"product_id"`
	ProductName string  `json:"product_name"`
	Stock       int     `json:"stock"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
	Severity    string  `json:"severity"` // CRITICAL, HIGH, MEDIUM
}

type FastMovingProduct struct {
	ProductID   int     `json:"product_id"`
	ProductName string  `json:"product_name"`
	Stock       int     `json:"stock"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
	OrdersCount int     `json:"orders_count"`
	TotalSold   int     `json:"total_sold"`
	DaysOfStock float64 `json:"days_of_stock"` // Estimated days until out of stock
}

// ============================================
// CUSTOMER INSIGHTS DTOs
// ============================================

// CustomerInsights contains customer analytics
type CustomerInsights struct {
	TotalCustomers  int               `json:"total_customers"`
	ActiveCustomers int               `json:"active_customers"`
	NewCustomers    int               `json:"new_customers"`
	Segments        []CustomerSegment `json:"segments"`
	TopCustomers    []TopCustomer     `json:"top_customers"`
}

type CustomerSegment struct {
	Segment  string  `json:"segment"`  // VIP, LOYAL, ACTIVE, AT_RISK, DORMANT
	Count    int     `json:"count"`
	AvgValue float64 `json:"avg_value"`
}

type TopCustomer struct {
	Email       string    `json:"email"`
	Name        string    `json:"name"`
	TotalOrders int       `json:"total_orders"`
	TotalSpent  float64   `json:"total_spent"`
	LastOrder   time.Time `json:"last_order"`
}

// ============================================
// CONVERSION FUNNEL DTOs
// ============================================

// ConversionFunnel contains conversion funnel metrics
type ConversionFunnel struct {
	OrdersCreated    int              `json:"orders_created"`
	OrdersPaid       int              `json:"orders_paid"`
	OrdersShipped    int              `json:"orders_shipped"`
	OrdersDelivered  int              `json:"orders_delivered"`
	OrdersCompleted  int              `json:"orders_completed"`
	PaymentRate      float64          `json:"payment_rate"`
	FulfillmentRate  float64          `json:"fulfillment_rate"`
	DeliveryRate     float64          `json:"delivery_rate"`
	CompletionRate   float64          `json:"completion_rate"`
	DropOffs         []FunnelDropOff  `json:"drop_offs"`
}

type FunnelDropOff struct {
	Stage      string  `json:"stage"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

// ============================================
// REVENUE CHART DTOs
// ============================================

// RevenueChart contains revenue data for charting
type RevenueChart struct {
	DataPoints []RevenueDataPoint `json:"data_points"`
}

type RevenueDataPoint struct {
	Date    string  `json:"date"`
	Orders  int     `json:"orders"`
	Revenue float64 `json:"revenue"`
}

// ============================================
// SYSTEM HEALTH DTOs
// ============================================

// SystemHealth contains system health metrics
type SystemHealth struct {
	WebhookSuccessRate    float64 `json:"webhook_success_rate"`
	PaymentGatewayLatency int     `json:"payment_gateway_latency"`
	BackgroundJobsHealthy bool    `json:"background_jobs_healthy"`
	LastTrackingUpdate    string  `json:"last_tracking_update"`
}

// CourierPerformance contains courier performance metrics
type CourierPerformance struct {
	CourierName     string  `json:"courier_name"`
	Delivered       int     `json:"delivered"`
	Failed          int     `json:"failed"`
	AvgDeliveryDays float64 `json:"avg_delivery_days"`
	SuccessRate     float64 `json:"success_rate"`
}
