package service

import (
	"database/sql"
	"fmt"
	"zavera/dto"
	"zavera/repository"
)

type AdminDashboardService interface {
	GetExecutiveMetrics(period string) (*dto.ExecutiveMetrics, error)
	GetPaymentMonitor() (*dto.PaymentMonitor, error)
	GetInventoryAlerts() (*dto.InventoryAlerts, error)
	GetCustomerInsights() (*dto.CustomerInsights, error)
	GetConversionFunnel(period string) (*dto.ConversionFunnel, error)
	GetRevenueChart(period string) (*dto.RevenueChart, error)
}

type adminDashboardService struct {
	db          *sql.DB
	orderRepo   repository.OrderRepository
	productRepo repository.ProductRepository
}

func NewAdminDashboardService(
	db *sql.DB,
	orderRepo repository.OrderRepository,
	productRepo repository.ProductRepository,
) AdminDashboardService {
	return &adminDashboardService{
		db:          db,
		orderRepo:   orderRepo,
		productRepo: productRepo,
	}
}

// GetExecutiveMetrics returns high-level business metrics
func (s *adminDashboardService) GetExecutiveMetrics(period string) (*dto.ExecutiveMetrics, error) {
	dateFilter := s.getPeriodFilter(period)
	
	metrics := &dto.ExecutiveMetrics{
		PaymentMethods: []dto.PaymentMethodStat{},
		TopProducts:    []dto.TopProductStat{},
	}

	// GMV (Gross Merchandise Value) - total order value excluding cancelled/failed
	err := s.db.QueryRow(fmt.Sprintf(`
		SELECT 
			COALESCE(SUM(total_amount), 0) as gmv,
			COUNT(*) as total_orders,
			COALESCE(AVG(total_amount), 0) as avg_order_value
		FROM orders
		WHERE status NOT IN ('CANCELLED', 'FAILED', 'EXPIRED')
		%s
	`, dateFilter)).Scan(&metrics.GMV, &metrics.TotalOrders, &metrics.AvgOrderValue)
	if err != nil {
		return nil, fmt.Errorf("failed to get GMV: %w", err)
	}

	// Revenue (actually paid orders only)
	err = s.db.QueryRow(fmt.Sprintf(`
		SELECT 
			COALESCE(SUM(total_amount), 0) as revenue,
			COUNT(*) as paid_orders
		FROM orders
		WHERE status IN ('PAID', 'PACKING', 'SHIPPED', 'DELIVERED', 'COMPLETED')
		%s
	`, dateFilter)).Scan(&metrics.Revenue, &metrics.PaidOrders)
	if err != nil {
		return nil, fmt.Errorf("failed to get revenue: %w", err)
	}

	// Pending revenue (orders awaiting payment)
	err = s.db.QueryRow(fmt.Sprintf(`
		SELECT COALESCE(SUM(total_amount), 0)
		FROM orders
		WHERE status = 'PENDING'
		%s
	`, dateFilter)).Scan(&metrics.PendingRevenue)
	if err != nil {
		metrics.PendingRevenue = 0
	}

	// Calculate conversion rate
	if metrics.TotalOrders > 0 {
		metrics.ConversionRate = float64(metrics.PaidOrders) / float64(metrics.TotalOrders) * 100
	}

	// Get payment method breakdown
	rows, err := s.db.Query(fmt.Sprintf(`
		SELECT 
			COALESCE(op.payment_method, 'UNKNOWN') as method,
			COUNT(*) as count,
			COALESCE(SUM(o.total_amount), 0) as amount
		FROM orders o
		LEFT JOIN order_payments op ON o.id = op.order_id
		WHERE o.status IN ('PAID', 'PACKING', 'SHIPPED', 'DELIVERED', 'COMPLETED')
		%s
		GROUP BY op.payment_method
	`, dateFilter))
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var stat dto.PaymentMethodStat
			if err := rows.Scan(&stat.Method, &stat.Count, &stat.Amount); err == nil {
				metrics.PaymentMethods = append(metrics.PaymentMethods, stat)
			}
		}
	}

	// Top selling products
	rows, err = s.db.Query(fmt.Sprintf(`
		SELECT 
			oi.product_id,
			oi.product_name,
			SUM(oi.quantity) as total_sold,
			COALESCE(SUM(oi.subtotal), 0) as revenue
		FROM order_items oi
		JOIN orders o ON oi.order_id = o.id
		WHERE o.status IN ('PAID', 'PACKING', 'SHIPPED', 'DELIVERED', 'COMPLETED')
		%s
		GROUP BY oi.product_id, oi.product_name
		ORDER BY total_sold DESC
		LIMIT 10
	`, dateFilter))
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var stat dto.TopProductStat
			if err := rows.Scan(&stat.ProductID, &stat.ProductName, &stat.TotalSold, &stat.Revenue); err == nil {
				metrics.TopProducts = append(metrics.TopProducts, stat)
			}
		}
	}

	return metrics, nil
}

// GetPaymentMonitor returns real-time payment monitoring data
func (s *adminDashboardService) GetPaymentMonitor() (*dto.PaymentMonitor, error) {
	monitor := &dto.PaymentMonitor{
		StuckPayments:     []dto.StuckPayment{},
		MethodPerformance: []dto.PaymentMethodPerformance{},
	}

	// Pending payments (awaiting customer payment)
	err := s.db.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(total_amount), 0)
		FROM orders
		WHERE status = 'PENDING'
		AND created_at > NOW() - INTERVAL '24 hours'
	`).Scan(&monitor.PendingCount, &monitor.PendingAmount)
	if err != nil {
		monitor.PendingCount = 0
		monitor.PendingAmount = 0
	}

	// Expiring soon (< 1 hour remaining)
	err = s.db.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(total_amount), 0)
		FROM orders
		WHERE status = 'PENDING'
		AND created_at < NOW() - INTERVAL '23 hours'
		AND created_at > NOW() - INTERVAL '24 hours'
	`).Scan(&monitor.ExpiringSoonCount, &monitor.ExpiringSoonAmount)
	if err != nil {
		monitor.ExpiringSoonCount = 0
		monitor.ExpiringSoonAmount = 0
	}

	// Stuck payments (pending in gateway but not updated)
	// Only show PENDING orders (not EXPIRED - those are already final)
	rows, err := s.db.Query(`
		SELECT 
			op.id,
			o.order_code,
			op.payment_method,
			COALESCE(op.bank, 'N/A'),
			o.total_amount,
			op.created_at,
			EXTRACT(EPOCH FROM (NOW() - op.created_at))/3600 as hours_pending
		FROM order_payments op
		JOIN orders o ON op.order_id = o.id
		WHERE op.payment_status = 'PENDING'
		AND o.status = 'PENDING'
		AND op.created_at < NOW() - INTERVAL '1 hour'
		ORDER BY op.created_at ASC
		LIMIT 50
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var sp dto.StuckPayment
			if err := rows.Scan(&sp.PaymentID, &sp.OrderCode, &sp.PaymentMethod, &sp.Bank, 
				&sp.Amount, &sp.CreatedAt, &sp.HoursPending); err == nil {
				monitor.StuckPayments = append(monitor.StuckPayments, sp)
			}
		}
	}

	// Today's payment stats
	err = s.db.QueryRow(`
		SELECT 
			COUNT(*) as total_paid,
			COALESCE(SUM(total_amount), 0) as total_amount
		FROM orders
		WHERE status IN ('PAID', 'PACKING', 'SHIPPED', 'DELIVERED', 'COMPLETED')
		AND DATE(paid_at) = CURRENT_DATE
	`).Scan(&monitor.TodayPaidCount, &monitor.TodayPaidAmount)
	if err != nil {
		monitor.TodayPaidCount = 0
		monitor.TodayPaidAmount = 0
	}

	// Payment method performance today
	rows, err = s.db.Query(`
		SELECT 
			op.payment_method,
			COUNT(*) as count,
			COALESCE(AVG(EXTRACT(EPOCH FROM (op.paid_at - op.created_at))/60), 0) as avg_time_minutes
		FROM order_payments op
		WHERE op.payment_status = 'PAID'
		AND DATE(op.paid_at) = CURRENT_DATE
		GROUP BY op.payment_method
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var perf dto.PaymentMethodPerformance
			if err := rows.Scan(&perf.Method, &perf.Count, &perf.AvgTimeMinutes); err == nil {
				monitor.MethodPerformance = append(monitor.MethodPerformance, perf)
			}
		}
	}

	return monitor, nil
}

// GetInventoryAlerts returns low stock and out of stock products
func (s *adminDashboardService) GetInventoryAlerts() (*dto.InventoryAlerts, error) {
	alerts := &dto.InventoryAlerts{
		OutOfStock: []dto.ProductStockAlert{},
		LowStock:   []dto.ProductStockAlert{},
		FastMoving: []dto.FastMovingProduct{},
	}

	// Out of stock
	rows, err := s.db.Query(`
		SELECT id, name, stock, price, category
		FROM products
		WHERE stock = 0 AND is_active = true
		ORDER BY name
		LIMIT 50
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var alert dto.ProductStockAlert
			if err := rows.Scan(&alert.ProductID, &alert.ProductName, &alert.Stock, &alert.Price, &alert.Category); err == nil {
				alert.Severity = "CRITICAL"
				alerts.OutOfStock = append(alerts.OutOfStock, alert)
			}
		}
	}

	// Low stock (< 10 units)
	rows, err = s.db.Query(`
		SELECT id, name, stock, price, category
		FROM products
		WHERE stock > 0 AND stock < 10 AND is_active = true
		ORDER BY stock ASC, name
		LIMIT 50
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var alert dto.ProductStockAlert
			if err := rows.Scan(&alert.ProductID, &alert.ProductName, &alert.Stock, &alert.Price, &alert.Category); err == nil {
				if alert.Stock <= 3 {
					alert.Severity = "HIGH"
				} else {
					alert.Severity = "MEDIUM"
				}
				alerts.LowStock = append(alerts.LowStock, alert)
			}
		}
	}

	// Fast moving (sold > 10 in last 7 days, stock < 20)
	rows, err = s.db.Query(`
		SELECT 
			p.id, p.name, p.stock, p.price, p.category,
			COUNT(oi.id) as orders_count,
			SUM(oi.quantity) as total_sold
		FROM products p
		JOIN order_items oi ON p.id = oi.product_id
		JOIN orders o ON oi.order_id = o.id
		WHERE o.status IN ('PAID', 'PACKING', 'SHIPPED', 'DELIVERED', 'COMPLETED')
		AND o.created_at > NOW() - INTERVAL '7 days'
		AND p.stock < 20
		AND p.is_active = true
		GROUP BY p.id, p.name, p.stock, p.price, p.category
		HAVING SUM(oi.quantity) > 10
		ORDER BY total_sold DESC
		LIMIT 20
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var fm dto.FastMovingProduct
			if err := rows.Scan(&fm.ProductID, &fm.ProductName, &fm.Stock, &fm.Price, &fm.Category, 
				&fm.OrdersCount, &fm.TotalSold); err == nil {
				fm.DaysOfStock = float64(fm.Stock) / (float64(fm.TotalSold) / 7.0)
				alerts.FastMoving = append(alerts.FastMoving, fm)
			}
		}
	}

	return alerts, nil
}

// GetCustomerInsights returns customer analytics
func (s *adminDashboardService) GetCustomerInsights() (*dto.CustomerInsights, error) {
	insights := &dto.CustomerInsights{
		Segments:     []dto.CustomerSegment{},
		TopCustomers: []dto.TopCustomer{},
	}

	// Total customers
	err := s.db.QueryRow(`SELECT COUNT(DISTINCT id) FROM users WHERE role = 'customer'`).Scan(&insights.TotalCustomers)
	if err != nil {
		insights.TotalCustomers = 0
	}

	// Active customers (ordered in last 30 days)
	err = s.db.QueryRow(`
		SELECT COUNT(DISTINCT customer_email)
		FROM orders
		WHERE created_at > NOW() - INTERVAL '30 days'
		AND status NOT IN ('CANCELLED', 'FAILED', 'EXPIRED')
	`).Scan(&insights.ActiveCustomers)
	if err != nil {
		insights.ActiveCustomers = 0
	}

	// New customers (registered in last 30 days)
	err = s.db.QueryRow(`
		SELECT COUNT(*)
		FROM users
		WHERE role = 'customer'
		AND created_at > NOW() - INTERVAL '30 days'
	`).Scan(&insights.NewCustomers)
	if err != nil {
		insights.NewCustomers = 0
	}

	// RFM Segmentation (simplified)
	rows, err := s.db.Query(`
		WITH customer_stats AS (
			SELECT 
				customer_email,
				COUNT(*) as frequency,
				COALESCE(SUM(total_amount), 0) as monetary,
				MAX(created_at) as recency
			FROM orders
			WHERE status IN ('PAID', 'PACKING', 'SHIPPED', 'DELIVERED', 'COMPLETED')
			GROUP BY customer_email
		)
		SELECT 
			CASE 
				WHEN recency > NOW() - INTERVAL '30 days' AND frequency >= 3 AND monetary >= 1000000 THEN 'VIP'
				WHEN recency > NOW() - INTERVAL '30 days' AND frequency >= 2 THEN 'LOYAL'
				WHEN recency > NOW() - INTERVAL '30 days' THEN 'ACTIVE'
				WHEN recency > NOW() - INTERVAL '90 days' THEN 'AT_RISK'
				ELSE 'DORMANT'
			END as segment,
			COUNT(*) as count,
			COALESCE(AVG(monetary), 0) as avg_value
		FROM customer_stats
		GROUP BY segment
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var seg dto.CustomerSegment
			if err := rows.Scan(&seg.Segment, &seg.Count, &seg.AvgValue); err == nil {
				insights.Segments = append(insights.Segments, seg)
			}
		}
	}

	// Top customers by revenue
	rows, err = s.db.Query(`
		SELECT 
			customer_email,
			customer_name,
			COUNT(*) as total_orders,
			COALESCE(SUM(total_amount), 0) as total_spent,
			MAX(created_at) as last_order
		FROM orders
		WHERE status IN ('PAID', 'PACKING', 'SHIPPED', 'DELIVERED', 'COMPLETED')
		GROUP BY customer_email, customer_name
		ORDER BY total_spent DESC
		LIMIT 20
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var tc dto.TopCustomer
			if err := rows.Scan(&tc.Email, &tc.Name, &tc.TotalOrders, &tc.TotalSpent, &tc.LastOrder); err == nil {
				insights.TopCustomers = append(insights.TopCustomers, tc)
			}
		}
	}

	return insights, nil
}

// GetConversionFunnel returns conversion funnel metrics
func (s *adminDashboardService) GetConversionFunnel(period string) (*dto.ConversionFunnel, error) {
	dateFilter := s.getPeriodFilter(period)
	
	funnel := &dto.ConversionFunnel{
		DropOffs: []dto.FunnelDropOff{},
	}

	// Orders created
	err := s.db.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*) FROM orders WHERE 1=1 %s
	`, dateFilter)).Scan(&funnel.OrdersCreated)
	if err != nil {
		funnel.OrdersCreated = 0
	}

	// Orders paid
	err = s.db.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM orders 
		WHERE status IN ('PAID', 'PACKING', 'SHIPPED', 'DELIVERED', 'COMPLETED')
		%s
	`, dateFilter)).Scan(&funnel.OrdersPaid)
	if err != nil {
		funnel.OrdersPaid = 0
	}

	// Orders shipped
	err = s.db.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM orders 
		WHERE status IN ('SHIPPED', 'DELIVERED', 'COMPLETED')
		%s
	`, dateFilter)).Scan(&funnel.OrdersShipped)
	if err != nil {
		funnel.OrdersShipped = 0
	}

	// Orders delivered
	err = s.db.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM orders 
		WHERE status IN ('DELIVERED', 'COMPLETED')
		%s
	`, dateFilter)).Scan(&funnel.OrdersDelivered)
	if err != nil {
		funnel.OrdersDelivered = 0
	}

	// Orders completed
	err = s.db.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM orders 
		WHERE status = 'COMPLETED'
		%s
	`, dateFilter)).Scan(&funnel.OrdersCompleted)
	if err != nil {
		funnel.OrdersCompleted = 0
	}

	// Calculate conversion rates
	if funnel.OrdersCreated > 0 {
		funnel.PaymentRate = float64(funnel.OrdersPaid) / float64(funnel.OrdersCreated) * 100
		funnel.FulfillmentRate = float64(funnel.OrdersShipped) / float64(funnel.OrdersCreated) * 100
		funnel.DeliveryRate = float64(funnel.OrdersDelivered) / float64(funnel.OrdersCreated) * 100
		funnel.CompletionRate = float64(funnel.OrdersCompleted) / float64(funnel.OrdersCreated) * 100
	}

	// Drop-off analysis
	if funnel.OrdersCreated > 0 {
		funnel.DropOffs = append(funnel.DropOffs, dto.FunnelDropOff{
			Stage:      "Payment",
			Count:      funnel.OrdersCreated - funnel.OrdersPaid,
			Percentage: float64(funnel.OrdersCreated-funnel.OrdersPaid) / float64(funnel.OrdersCreated) * 100,
		})
	}
	if funnel.OrdersPaid > 0 {
		funnel.DropOffs = append(funnel.DropOffs, dto.FunnelDropOff{
			Stage:      "Fulfillment",
			Count:      funnel.OrdersPaid - funnel.OrdersShipped,
			Percentage: float64(funnel.OrdersPaid-funnel.OrdersShipped) / float64(funnel.OrdersPaid) * 100,
		})
	}
	if funnel.OrdersShipped > 0 {
		funnel.DropOffs = append(funnel.DropOffs, dto.FunnelDropOff{
			Stage:      "Delivery",
			Count:      funnel.OrdersShipped - funnel.OrdersDelivered,
			Percentage: float64(funnel.OrdersShipped-funnel.OrdersDelivered) / float64(funnel.OrdersShipped) * 100,
		})
	}

	return funnel, nil
}

// GetRevenueChart returns revenue data for charting
func (s *adminDashboardService) GetRevenueChart(period string) (*dto.RevenueChart, error) {
	chart := &dto.RevenueChart{
		DataPoints: []dto.RevenueDataPoint{},
	}

	var groupBy string
	var dateFormat string
	var intervalDays int

	switch period {
	case "7days":
		groupBy = "DATE(created_at)"
		dateFormat = "YYYY-MM-DD"
		intervalDays = 7
	case "30days":
		groupBy = "DATE(created_at)"
		dateFormat = "YYYY-MM-DD"
		intervalDays = 30
	case "90days":
		groupBy = "DATE_TRUNC('week', created_at)"
		dateFormat = "YYYY-MM-DD"
		intervalDays = 90
	case "year":
		groupBy = "DATE_TRUNC('month', created_at)"
		dateFormat = "YYYY-MM"
		intervalDays = 365
	default:
		groupBy = "DATE(created_at)"
		dateFormat = "YYYY-MM-DD"
		intervalDays = 7
	}

	rows, err := s.db.Query(fmt.Sprintf(`
		SELECT 
			TO_CHAR(%s, '%s') as date,
			COUNT(*) as orders,
			COALESCE(SUM(total_amount), 0) as revenue
		FROM orders
		WHERE status IN ('PAID', 'PACKING', 'SHIPPED', 'DELIVERED', 'COMPLETED')
		AND created_at > NOW() - INTERVAL '%d days'
		GROUP BY %s
		ORDER BY %s ASC
	`, groupBy, dateFormat, intervalDays, groupBy, groupBy))
	
	if err != nil {
		return chart, nil // Return empty chart instead of error
	}
	defer rows.Close()

	for rows.Next() {
		var dp dto.RevenueDataPoint
		if err := rows.Scan(&dp.Date, &dp.Orders, &dp.Revenue); err == nil {
			chart.DataPoints = append(chart.DataPoints, dp)
		}
	}

	return chart, nil
}

// Helper to get period filter
func (s *adminDashboardService) getPeriodFilter(period string) string {
	switch period {
	case "today":
		return "AND DATE(created_at) = CURRENT_DATE"
	case "yesterday":
		return "AND DATE(created_at) = CURRENT_DATE - INTERVAL '1 day'"
	case "week":
		return "AND created_at > NOW() - INTERVAL '7 days'"
	case "last_week":
		return "AND created_at > NOW() - INTERVAL '14 days' AND created_at <= NOW() - INTERVAL '7 days'"
	case "month":
		return "AND created_at > NOW() - INTERVAL '30 days'"
	case "last_month":
		return "AND created_at > NOW() - INTERVAL '60 days' AND created_at <= NOW() - INTERVAL '30 days'"
	case "year":
		return "AND created_at > NOW() - INTERVAL '365 days'"
	case "last_year":
		return "AND created_at > NOW() - INTERVAL '730 days' AND created_at <= NOW() - INTERVAL '365 days'"
	default:
		return "AND DATE(created_at) = CURRENT_DATE"
	}
}
