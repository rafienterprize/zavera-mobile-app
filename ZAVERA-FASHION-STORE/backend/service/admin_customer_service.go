package service

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
	"time"
	"zavera/dto"
	"zavera/repository"
)

type adminCustomerService struct {
	db        *sql.DB
	userRepo  repository.UserRepository
	orderRepo repository.OrderRepository
}

func NewAdminCustomerService(db *sql.DB, userRepo repository.UserRepository, orderRepo repository.OrderRepository) *adminCustomerService {
	return &adminCustomerService{
		db:        db,
		userRepo:  userRepo,
		orderRepo: orderRepo,
	}
}

func (s *adminCustomerService) GetCustomers(page, limit int, search, segment string) (*dto.CustomersResponse, error) {
	offset := (page - 1) * limit

	// Simple query to get customers with basic stats
	query := `
		SELECT 
			u.id, u.email, u.first_name, u.phone, u.created_at, u.is_verified,
			COALESCE(COUNT(o.id), 0) as total_orders,
			COALESCE(SUM(o.total_amount), 0) as total_spent,
			MAX(o.created_at) as last_order_date
		FROM users u
		LEFT JOIN orders o ON u.id = o.user_id AND o.status IN ('PAID', 'PACKING', 'SHIPPED', 'DELIVERED', 'COMPLETED')
		WHERE u.role = 'customer'
	`

	args := []interface{}{}
	if search != "" {
		query += " AND (u.email ILIKE $1 OR u.first_name ILIKE $1)"
		args = append(args, "%"+search+"%")
	}

	query += " GROUP BY u.id, u.email, u.first_name, u.phone, u.created_at, u.is_verified"
	query += " ORDER BY u.created_at DESC"
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var customerList []dto.CustomerResponse
	for rows.Next() {
		var c dto.CustomerResponse
		var createdAt time.Time
		var lastOrderDate sql.NullTime

		err := rows.Scan(
			&c.ID, &c.Email, &c.FirstName, &c.Phone, &createdAt, &c.IsVerified,
			&c.TotalOrders, &c.TotalSpent, &lastOrderDate,
		)
		if err != nil {
			continue
		}

		c.CreatedAt = createdAt.Format(time.RFC3339)
		if lastOrderDate.Valid {
			c.LastOrderDate = lastOrderDate.Time.Format(time.RFC3339)
		}

		// Determine segment
		c.Segment = "New"
		if c.TotalOrders >= 5 {
			c.Segment = "VIP"
		} else if c.TotalOrders > 0 {
			c.Segment = "Regular"
		}

		customerList = append(customerList, c)
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM users WHERE role = 'customer'"
	var total int
	s.db.QueryRow(countQuery).Scan(&total)

	return &dto.CustomersResponse{
		Customers: customerList,
		Total:     total,
	}, nil
}

func (s *adminCustomerService) GetCustomerStats() (*dto.CustomerStatsResponse, error) {
	stats := &dto.CustomerStatsResponse{}

	// Total customers
	s.db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'customer'").Scan(&stats.TotalCustomers)

	// New this month
	s.db.QueryRow(`
		SELECT COUNT(*) FROM users 
		WHERE role = 'customer' 
		AND created_at >= DATE_TRUNC('month', CURRENT_DATE)
	`).Scan(&stats.NewThisMonth)

	// VIP customers (5+ orders)
	s.db.QueryRow(`
		SELECT COUNT(DISTINCT user_id) 
		FROM orders 
		WHERE status IN ('PAID', 'PACKING', 'SHIPPED', 'DELIVERED', 'COMPLETED')
		GROUP BY user_id
		HAVING COUNT(*) >= 5
	`).Scan(&stats.VIPCustomers)

	// Total lifetime value
	s.db.QueryRow(`
		SELECT COALESCE(SUM(total_amount), 0) 
		FROM orders 
		WHERE status IN ('PAID', 'PACKING', 'SHIPPED', 'DELIVERED', 'COMPLETED')
	`).Scan(&stats.TotalLifetimeValue)

	return stats, nil
}

func (s *adminCustomerService) GetCustomerDetail(userID int) (*dto.CustomerDetailResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	// Get customer orders
	var orders []dto.OrderHistoryItem
	rows, err := s.db.Query(`
		SELECT order_code, total_amount, status, created_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 10
	`, userID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var order dto.OrderHistoryItem
			var createdAt time.Time
			rows.Scan(&order.OrderCode, &order.TotalAmount, &order.Status, &createdAt)
			order.CreatedAt = createdAt.Format(time.RFC3339)
			orders = append(orders, order)
		}
	}

	// Get stats
	var totalOrders int
	var totalSpent float64
	var lastOrderDate sql.NullTime
	s.db.QueryRow(`
		SELECT 
			COUNT(*), 
			COALESCE(SUM(total_amount), 0),
			MAX(created_at)
		FROM orders
		WHERE user_id = $1
		AND status IN ('PAID', 'PACKING', 'SHIPPED', 'DELIVERED', 'COMPLETED')
	`, userID).Scan(&totalOrders, &totalSpent, &lastOrderDate)

	seg := "New"
	if totalOrders >= 5 {
		seg = "VIP"
	} else if totalOrders > 0 {
		seg = "Regular"
	}

	lastOrder := ""
	if lastOrderDate.Valid {
		lastOrder = lastOrderDate.Time.Format(time.RFC3339)
	}

	return &dto.CustomerDetailResponse{
		ID:            user.ID,
		Email:         user.Email,
		FirstName:     user.FirstName,
		LastName:      "", // User model doesn't have LastName
		Phone:         user.Phone,
		TotalOrders:   totalOrders,
		TotalSpent:    totalSpent,
		LastOrderDate: lastOrder,
		CreatedAt:     user.CreatedAt.Format(time.RFC3339),
		IsVerified:    user.IsVerified,
		Segment:       seg,
		RecentOrders:  orders,
	}, nil
}

func (s *adminCustomerService) ExportCustomers() ([]byte, error) {
	rows, err := s.db.Query(`
		SELECT 
			u.id, u.email, u.first_name, u.phone, u.created_at,
			COALESCE(COUNT(o.id), 0) as total_orders,
			COALESCE(SUM(o.total_amount), 0) as total_spent
		FROM users u
		LEFT JOIN orders o ON u.id = o.user_id AND o.status IN ('PAID', 'PACKING', 'SHIPPED', 'DELIVERED', 'COMPLETED')
		WHERE u.role = 'customer'
		GROUP BY u.id, u.email, u.first_name, u.phone, u.created_at
		ORDER BY u.created_at DESC
		LIMIT 10000
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	writer.Write([]string{"ID", "Email", "First Name", "Phone", "Total Orders", "Total Spent", "Segment", "Joined Date"})

	// Write data
	for rows.Next() {
		var id int
		var email, firstName, phone string
		var createdAt time.Time
		var totalOrders int
		var totalSpent float64

		rows.Scan(&id, &email, &firstName, &phone, &createdAt, &totalOrders, &totalSpent)

		seg := "New"
		if totalOrders >= 5 {
			seg = "VIP"
		} else if totalOrders > 0 {
			seg = "Regular"
		}

		writer.Write([]string{
			fmt.Sprintf("%d", id),
			email,
			firstName,
			"",
			phone,
			fmt.Sprintf("%d", totalOrders),
			fmt.Sprintf("%.2f", totalSpent),
			seg,
			createdAt.Format("2006-01-02"),
		})
	}

	writer.Flush()
	return buf.Bytes(), nil
}
