package models

import "time"

// BiteshipLocation represents a saved Biteship location
type BiteshipLocation struct {
	ID           int       `json:"id" db:"id"`
	UserID       *int      `json:"user_id,omitempty" db:"user_id"`
	LocationID   string    `json:"location_id" db:"location_id"`       // Biteship location_id
	AreaID       string    `json:"area_id" db:"area_id"`               // Biteship area_id
	AreaName     string    `json:"area_name" db:"area_name"`           // Full area path
	ContactName  string    `json:"contact_name" db:"contact_name"`
	ContactPhone string    `json:"contact_phone" db:"contact_phone"`
	Address      string    `json:"address" db:"address"`
	PostalCode   string    `json:"postal_code" db:"postal_code"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}
