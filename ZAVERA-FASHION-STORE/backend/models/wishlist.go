package models

import "time"

// Wishlist represents a user's wishlist
type Wishlist struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	ProductID int       `json:"product_id" db:"product_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	Product   *Product  `json:"product,omitempty" db:"-"`
}
