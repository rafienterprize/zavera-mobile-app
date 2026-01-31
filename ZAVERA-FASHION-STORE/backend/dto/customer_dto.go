package dto

// CustomerResponse represents a customer in the list
type CustomerResponse struct {
	ID            int     `json:"id"`
	Email         string  `json:"email"`
	FirstName     string  `json:"first_name"`
	LastName      string  `json:"last_name"`
	Phone         string  `json:"phone"`
	TotalOrders   int     `json:"total_orders"`
	TotalSpent    float64 `json:"total_spent"`
	LastOrderDate string  `json:"last_order_date"`
	CreatedAt     string  `json:"created_at"`
	IsVerified    bool    `json:"is_verified"`
	Segment       string  `json:"segment"` // VIP, Regular, New
}

// CustomersResponse represents paginated customers list
type CustomersResponse struct {
	Customers []CustomerResponse `json:"customers"`
	Total     int                `json:"total"`
}

// CustomerStatsResponse represents customer statistics
type CustomerStatsResponse struct {
	TotalCustomers     int     `json:"total_customers"`
	NewThisMonth       int     `json:"new_this_month"`
	VIPCustomers       int     `json:"vip_customers"`
	TotalLifetimeValue float64 `json:"total_lifetime_value"`
}

// OrderHistoryItem represents a single order in customer history
type OrderHistoryItem struct {
	OrderCode   string  `json:"order_code"`
	TotalAmount float64 `json:"total_amount"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
}

// CustomerDetailResponse represents detailed customer information
type CustomerDetailResponse struct {
	ID            int                 `json:"id"`
	Email         string              `json:"email"`
	FirstName     string              `json:"first_name"`
	LastName      string              `json:"last_name"`
	Phone         string              `json:"phone"`
	TotalOrders   int                 `json:"total_orders"`
	TotalSpent    float64             `json:"total_spent"`
	LastOrderDate string              `json:"last_order_date"`
	CreatedAt     string              `json:"created_at"`
	IsVerified    bool                `json:"is_verified"`
	Segment       string              `json:"segment"`
	RecentOrders  []OrderHistoryItem  `json:"recent_orders"`
}
