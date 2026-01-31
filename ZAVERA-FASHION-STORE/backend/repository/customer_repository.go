package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
	"zavera/models"
)

// CustomerWithStats represents a customer with order statistics
type CustomerWithStats struct {
	models.User
	TotalOrders   int
	TotalSpent    float64
	LastOrderDate *time.Time
}

// CustomerStats represents overall customer statistics
type CustomerStats struct {
	TotalCustomers     int
	NewThisMonth       int
	VIPCustomers       int
	TotalLifetimeValue float64
}

// UserOrderStats represents order statistics for a specific user
type UserOrderStats struct {
	TotalOrders   int
	TotalSpent    float64
	LastOrderDate *time.Time
}

// GetCustomersWithStats returns customers with their order statistics
func (r *userRepository) GetCustomersWithStats(offset, limit int, search, segment string) ([]CustomerWithStats, int, error) {
	// Build WHERE clause
	whereConditions := []string{"u.is_active = true"}
	args := []interface{}{}
	argCount := 1

	if search != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("(u.email ILIKE $%d OR u.first_name ILIKE $%d OR u.phone ILIKE $%d)", argCount, argCount, argCount))
		args = append(args, "%"+search+"%")
		argCount++
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// Get total count
	countQuery := fmt.Sprintf(`
		SELECT COUNT(DISTINCT u.id)
		FROM users u
		WHERE %s
	`, whereClause)

	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get customers with stats
	query := fmt.Sprintf(`
		SELECT 
			u.id, u.email, u.first_name, u.name, u.phone, u.is_verified, u.created_at,
			COALESCE(COUNT(DISTINCT o.id), 0) as total_orders,
			COALESCE(SUM(o.total_amount), 0) as total_spent,
			MAX(o.created_at) as last_order_date
		FROM users u
		LEFT JOIN orders o ON u.id = o.user_id AND o.status NOT IN ('CANCELLED', 'EXPIRED')
		WHERE %s
		GROUP BY u.id, u.email, u.first_name, u.name, u.phone, u.is_verified, u.created_at
		ORDER BY u.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argCount, argCount+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var customers []CustomerWithStats
	for rows.Next() {
		var c CustomerWithStats
		var lastOrderDate sql.NullTime

		err := rows.Scan(
			&c.ID, &c.Email, &c.FirstName, &c.Name, &c.Phone, &c.IsVerified, &c.CreatedAt,
			&c.TotalOrders, &c.TotalSpent, &lastOrderDate,
		)
		if err != nil {
			continue
		}

		if lastOrderDate.Valid {
			c.LastOrderDate = &lastOrderDate.Time
		}

		// Apply segment filter
		seg := "New"
		if c.TotalOrders >= 5 {
			seg = "VIP"
		} else if c.TotalOrders > 0 {
			seg = "Regular"
		}

		if segment != "" && segment != seg {
			continue
		}

		customers = append(customers, c)
	}

	return customers, total, nil
}

// GetCustomerStats returns overall customer statistics
func (r *userRepository) GetCustomerStats() (*CustomerStats, error) {
	query := `
		SELECT 
			COUNT(DISTINCT u.id) as total_customers,
			COUNT(DISTINCT CASE WHEN u.created_at >= DATE_TRUNC('month', CURRENT_DATE) THEN u.id END) as new_this_month,
			COUNT(DISTINCT CASE WHEN order_counts.total_orders >= 5 THEN u.id END) as vip_customers,
			COALESCE(SUM(order_totals.total_spent), 0) as total_lifetime_value
		FROM users u
		LEFT JOIN (
			SELECT user_id, COUNT(*) as total_orders
			FROM orders
			WHERE status NOT IN ('CANCELLED', 'EXPIRED')
			GROUP BY user_id
		) order_counts ON u.id = order_counts.user_id
		LEFT JOIN (
			SELECT user_id, SUM(total_amount) as total_spent
			FROM orders
			WHERE status NOT IN ('CANCELLED', 'EXPIRED')
			GROUP BY user_id
		) order_totals ON u.id = order_totals.user_id
		WHERE u.is_active = true
	`

	var stats CustomerStats
	err := r.db.QueryRow(query).Scan(
		&stats.TotalCustomers,
		&stats.NewThisMonth,
		&stats.VIPCustomers,
		&stats.TotalLifetimeValue,
	)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

// GetUserOrderStats returns order statistics for a specific user
func (r *userRepository) GetUserOrderStats(userID int) (*UserOrderStats, error) {
	query := `
		SELECT 
			COALESCE(COUNT(*), 0) as total_orders,
			COALESCE(SUM(total_amount), 0) as total_spent,
			MAX(created_at) as last_order_date
		FROM orders
		WHERE user_id = $1 AND status NOT IN ('CANCELLED', 'EXPIRED')
	`

	var stats UserOrderStats
	var lastOrderDate sql.NullTime

	err := r.db.QueryRow(query, userID).Scan(
		&stats.TotalOrders,
		&stats.TotalSpent,
		&lastOrderDate,
	)
	if err != nil {
		return nil, err
	}

	if lastOrderDate.Valid {
		stats.LastOrderDate = &lastOrderDate.Time
	}

	return &stats, nil
}
