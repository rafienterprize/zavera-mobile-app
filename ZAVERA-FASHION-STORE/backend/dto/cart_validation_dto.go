package dto

// CartValidationRequest validates cart items against current product data
type CartValidationRequest struct {
	Items []CartValidationItem `json:"items" binding:"required"`
}

type CartValidationItem struct {
	CartItemID int `json:"cart_item_id" binding:"required"`
	ProductID  int `json:"product_id" binding:"required"`
	Quantity   int `json:"quantity" binding:"required,min=1"`
}

// CartValidationResponse returns validation results
type CartValidationResponse struct {
	Valid   bool                       `json:"valid"`
	Changes []CartItemChange           `json:"changes,omitempty"`
	Cart    *CartResponse              `json:"cart"`
	Message string                     `json:"message,omitempty"`
}

// CartItemChange describes what changed in a cart item
type CartItemChange struct {
	CartItemID   int     `json:"cart_item_id"`
	ProductID    int     `json:"product_id"`
	ProductName  string  `json:"product_name"`
	ChangeType   string  `json:"change_type"` // "price_changed", "weight_changed", "stock_insufficient", "product_unavailable"
	OldValue     string  `json:"old_value,omitempty"`
	NewValue     string  `json:"new_value,omitempty"`
	OldPrice     float64 `json:"old_price,omitempty"`
	NewPrice     float64 `json:"new_price,omitempty"`
	OldWeight    int     `json:"old_weight,omitempty"`
	NewWeight    int     `json:"new_weight,omitempty"`
	CurrentStock int     `json:"current_stock,omitempty"`
	Message      string  `json:"message"`
}
