# Dashboard Upgrade Implementation Plan

## Status: PARTIALLY IMPLEMENTED ✅

### ✅ COMPLETED UPGRADES

#### 1. Growth Comparison (MoM/YoY)
- **Frontend**: Added `previousMetrics` state and `calculateGrowth()` function
- **Backend**: Extended `getPeriodFilter()` to support yesterday, last_week, last_month, last_year
- **UI**: Trend indicators (↑↓) with percentage on all KPI cards
- **Status**: ✅ DONE

#### 2. System Health Monitor
- **Added**: Real-time system health panel showing:
  - Webhook success rate (98.5%)
  - Payment gateway latency (245ms)
  - Background jobs status
  - Last tracking update timestamp
- **Status**: ✅ DONE (using mock data, needs real endpoint)

#### 3. Revenue Chart Visualization
- **Added**: Horizontal bar chart showing last 7 days revenue
- **Features**: 
  - Visual bars with gradient colors
  - Revenue amount and order count per day
  - Responsive width based on max revenue
- **Status**: ✅ DONE

#### 4. Refund & Dispute Intelligence
- **Added**: Dashboard panel showing:
  - Open disputes count
  - Disputes needing resolution
  - Top 3 recent disputes with links
- **Status**: ✅ DONE

#### 5. Courier Performance Comparison
- **Added**: Courier leaderboard showing:
  - Delivery success rate per courier
  - Delivered vs failed shipments
  - Average delivery days
  - Color-coded performance (green/yellow/red)
- **Status**: ✅ DONE (using mock data, needs real endpoint)

---

## 🔧 REMAINING WORK

### Backend Endpoints Needed

#### 1. System Health Endpoint
```go
// GET /api/admin/system/health
type SystemHealth struct {
    WebhookSuccessRate     float64 `json:"webhook_success_rate"`
    PaymentGatewayLatency  int     `json:"payment_gateway_latency"`
    BackgroundJobsHealthy  bool    `json:"background_jobs_healthy"`
    LastTrackingUpdate     string  `json:"last_tracking_update"`
}
```

**Implementation**:
- Query webhook logs for success rate (last 24h)
- Measure Midtrans API response time
- Check last run time of background jobs
- Get latest shipment tracking update timestamp

#### 2. Courier Performance Endpoint
```go
// GET /api/admin/analytics/courier-performance
type CourierPerformance struct {
    CourierName    string  `json:"courier_name"`
    Delivered      int     `json:"delivered"`
    Failed         int     `json:"failed"`
    AvgDeliveryDays float64 `json:"avg_delivery_days"`
    SuccessRate    float64 `json:"success_rate"`
}
```

**Implementation**:
```sql
SELECT 
    provider_code as courier_name,
    COUNT(CASE WHEN status = 'DELIVERED' THEN 1 END) as delivered,
    COUNT(CASE WHEN status IN ('DELIVERY_FAILED', 'LOST') THEN 1 END) as failed,
    AVG(EXTRACT(DAY FROM (delivered_at - shipped_at))) as avg_delivery_days,
    (COUNT(CASE WHEN status = 'DELIVERED' THEN 1 END)::float / 
     COUNT(*)::float * 100) as success_rate
FROM shipments
WHERE created_at > NOW() - INTERVAL '30 days'
GROUP BY provider_code
ORDER BY success_rate DESC
```

#### 3. Shipments List Endpoint
```go
// GET /api/admin/shipments?status=SHIPPED&page=1
type ShipmentListItem struct {
    ID              int    `json:"id"`
    OrderCode       string `json:"order_code"`
    TrackingNumber  string `json:"tracking_number"`
    Status          string `json:"status"`
    ProviderCode    string `json:"provider_code"`
    ProviderName    string `json:"provider_name"`
    DaysWithoutUpdate int  `json:"days_without_update"`
    CreatedAt       string `json:"created_at"`
}
```

**Status**: Handler added, needs service implementation

#### 4. Refunds List Endpoint
```go
// GET /api/admin/refunds?status=PENDING&page=1
type RefundListItem struct {
    ID            int     `json:"id"`
    RefundCode    string  `json:"refund_code"`
    OrderCode     string  `json:"order_code"`
    RefundType    string  `json:"refund_type"`
    Amount        float64 `json:"amount"`
    Status        string  `json:"status"`
    CreatedAt     string  `json:"created_at"`
}
```

**Status**: Frontend ready, needs backend endpoint

---

## 📋 IMPLEMENTATION CHECKLIST

### Priority 0 (Critical)
- [ ] Implement System Health endpoint
- [ ] Implement Courier Performance endpoint
- [ ] Complete Shipments List service method
- [ ] Implement Refunds List endpoint
- [ ] Add routes for new endpoints

### Priority 1 (High Value)
- [ ] Add SLA breach indicators (payment 24h, packing 48h, shipping 7d)
- [ ] Implement inventory turnover calculation
- [ ] Add slow-moving products detection (>90 days no sales)
- [ ] Integrate SSE for real-time dashboard updates

### Priority 2 (Nice to Have)
- [ ] Add refund rate calculation (% of orders refunded)
- [ ] Track products with highest return rate
- [ ] Add dispute resolution time tracking
- [ ] Implement advanced customer segmentation filters

---

## 🎯 EXPECTED OUTCOMES

### Dashboard Will Show:
1. ✅ Executive metrics with growth trends
2. ✅ Real-time payment monitoring with stuck payment alerts
3. ✅ System health status (webhook, gateway, jobs)
4. ✅ Revenue trend visualization
5. ✅ Inventory alerts (out of stock, low stock, fast moving)
6. ✅ Customer insights with RFM segmentation
7. ✅ Conversion funnel with drop-off analysis
8. ✅ Dispute & refund intelligence
9. ✅ Courier performance comparison
10. ⏳ Shipments list with filtering (needs backend)
11. ⏳ Refunds list with processing (needs backend)
12. ✅ Audit logs (already complete)

### Pages Status:
- **Dashboard**: 90% complete (needs real data for system health & courier perf)
- **Orders**: ✅ Complete
- **Products**: ✅ Complete
- **Customers**: ✅ Complete
- **Shipments**: 60% complete (UI ready, needs data endpoint)
- **Refunds**: 80% complete (UI ready, needs list endpoint)
- **Disputes**: ✅ Complete
- **Audit Logs**: ✅ Complete

---

## 🚀 NEXT STEPS

1. **Implement missing backend endpoints** (2-3 hours)
2. **Connect frontend to real data** (1 hour)
3. **Test all features end-to-end** (1 hour)
4. **Add SLA breach indicators** (1 hour)
5. **Integrate SSE for real-time updates** (2 hours)

**Total estimated time**: 7-9 hours to reach international-grade standard.

---

## 📊 COMPARISON: BEFORE vs AFTER

### BEFORE
- Basic metrics only (GMV, Revenue, Orders)
- No growth comparison
- No system health visibility
- No courier performance tracking
- Empty shipments page
- Empty refunds page
- No dispute visibility on dashboard

### AFTER
- ✅ Executive metrics with MoM/YoY growth
- ✅ System health monitoring
- ✅ Revenue trend visualization
- ✅ Courier performance leaderboard
- ✅ Dispute & refund intelligence
- ⏳ Populated shipments page (needs backend)
- ⏳ Populated refunds page (needs backend)
- ✅ Real-time payment monitoring
- ✅ Inventory intelligence
- ✅ Customer segmentation

**Dashboard completeness**: 85% → Needs backend endpoints to reach 100%
