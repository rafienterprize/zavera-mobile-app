package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type VariantAttributes map[string]interface{}

func (va VariantAttributes) Value() (driver.Value, error) {
	return json.Marshal(va)
}

func (va *VariantAttributes) Scan(value interface{}) error {
	if value == nil {
		*va = make(VariantAttributes)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, va)
}

type ProductVariant struct {
	ID                 int                `json:"id"`
	ProductID          int                `json:"product_id"`
	SKU                string             `json:"sku"`
	VariantName        string             `json:"variant_name"`
	Size               *string            `json:"size,omitempty"`
	Color              *string            `json:"color,omitempty"`
	ColorHex           *string            `json:"color_hex,omitempty"`
	Material           *string            `json:"material,omitempty"`
	Pattern            *string            `json:"pattern,omitempty"`
	Fit                *string            `json:"fit,omitempty"`
	Sleeve             *string            `json:"sleeve,omitempty"`
	CustomAttributes   VariantAttributes  `json:"custom_attributes,omitempty"`
	Price              *float64           `json:"price,omitempty"`
	CompareAtPrice     *float64           `json:"compare_at_price,omitempty"`
	CostPerItem        *float64           `json:"cost_per_item,omitempty"`
	StockQuantity      int                `json:"stock_quantity"`
	ReservedStock      int                `json:"reserved_stock"`
	LowStockThreshold  int                `json:"low_stock_threshold"`
	IsActive           bool               `json:"is_active"`
	IsDefault          bool               `json:"is_default"`
	WeightGrams        *int               `json:"weight_grams,omitempty"`
	LengthCm           *int               `json:"length_cm,omitempty"`
	WidthCm            *int               `json:"width_cm,omitempty"`
	HeightCm           *int               `json:"height_cm,omitempty"`
	Barcode            *string            `json:"barcode,omitempty"`
	Position           int                `json:"position"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
	Images             []VariantImage     `json:"images,omitempty"`
	AvailableStock     *int               `json:"available_stock,omitempty"`
}

type VariantImage struct {
	ID        int       `json:"id"`
	VariantID int       `json:"variant_id"`
	ImageURL  string    `json:"image_url"`
	AltText   *string   `json:"alt_text,omitempty"`
	Position  int       `json:"position"`
	IsPrimary bool      `json:"is_primary"`
	Width     *int      `json:"width,omitempty"`
	Height    *int      `json:"height,omitempty"`
	Format    *string   `json:"format,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type VariantAttribute struct {
	ID          int             `json:"id"`
	Name        string          `json:"name"`
	DisplayName string          `json:"display_name"`
	Type        string          `json:"type"`
	Options     json.RawMessage `json:"options,omitempty"`
	SortOrder   int             `json:"sort_order"`
	IsActive    bool            `json:"is_active"`
	CreatedAt   time.Time       `json:"created_at"`
}

type StockReservation struct {
	ID         int       `json:"id"`
	VariantID  int       `json:"variant_id"`
	CustomerID *int      `json:"customer_id,omitempty"`
	SessionID  *string   `json:"session_id,omitempty"`
	Quantity   int       `json:"quantity"`
	ReservedAt time.Time `json:"reserved_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	Status     string    `json:"status"`
	OrderID    *int      `json:"order_id,omitempty"`
}

type LowStockVariant struct {
	ID                int     `json:"id"`
	ProductID         int     `json:"product_id"`
	ProductName       string  `json:"product_name"`
	SKU               string  `json:"sku"`
	VariantName       string  `json:"variant_name"`
	Size              *string `json:"size,omitempty"`
	Color             *string `json:"color,omitempty"`
	StockQuantity     int     `json:"stock_quantity"`
	LowStockThreshold int     `json:"low_stock_threshold"`
	AvailableStock    int     `json:"available_stock"`
}

type VariantStockSummary struct {
	VariantID         int     `json:"variant_id"`
	ProductID         int     `json:"product_id"`
	ProductName       string  `json:"product_name"`
	SKU               string  `json:"sku"`
	VariantName       string  `json:"variant_name"`
	StockQuantity     int     `json:"stock_quantity"`
	ReservedQuantity  int     `json:"reserved_quantity"`
	AvailableQuantity int     `json:"available_quantity"`
}

type ProductWithVariants struct {
	Product
	Variants   []ProductVariant `json:"variants"`
	PriceRange *PriceRange      `json:"price_range,omitempty"`
}

type PriceRange struct {
	MinPrice float64 `json:"min_price"`
	MaxPrice float64 `json:"max_price"`
}
