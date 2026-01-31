package dto

type CreateVariantRequest struct {
	ProductID         int                    `json:"product_id" binding:"required"`
	SKU               string                 `json:"sku"`
	VariantName       string                 `json:"variant_name"`
	Size              *string                `json:"size"`
	Color             *string                `json:"color"`
	ColorHex          *string                `json:"color_hex"`
	Material          *string                `json:"material"`
	Pattern           *string                `json:"pattern"`
	Fit               *string                `json:"fit"`
	Sleeve            *string                `json:"sleeve"`
	CustomAttributes  map[string]interface{} `json:"custom_attributes"`
	Price             *float64               `json:"price"`
	CompareAtPrice    *float64               `json:"compare_at_price"`
	CostPerItem       *float64               `json:"cost_per_item"`
	StockQuantity     int                    `json:"stock_quantity"`
	LowStockThreshold int                    `json:"low_stock_threshold"`
	IsActive          bool                   `json:"is_active"`
	IsDefault         bool                   `json:"is_default"`
	WeightGrams       *int                   `json:"weight_grams"`
	Weight            *int                   `json:"weight"` // Alias for weight_grams
	LengthCm          *int                   `json:"length_cm"`
	Length            *int                   `json:"length"` // Alias for length_cm
	WidthCm           *int                   `json:"width_cm"`
	Width             *int                   `json:"width"` // Alias for width_cm
	HeightCm          *int                   `json:"height_cm"`
	Height            *int                   `json:"height"` // Alias for height_cm
	Barcode           *string                `json:"barcode"`
	Position          int                    `json:"position"`
}

type UpdateVariantRequest struct {
	SKU               string                 `json:"sku"`
	VariantName       string                 `json:"variant_name"`
	Size              *string                `json:"size"`
	Color             *string                `json:"color"`
	ColorHex          *string                `json:"color_hex"`
	Material          *string                `json:"material"`
	Pattern           *string                `json:"pattern"`
	Fit               *string                `json:"fit"`
	Sleeve            *string                `json:"sleeve"`
	CustomAttributes  map[string]interface{} `json:"custom_attributes"`
	Price             *float64               `json:"price"`
	CompareAtPrice    *float64               `json:"compare_at_price"`
	CostPerItem       *float64               `json:"cost_per_item"`
	StockQuantity     int                    `json:"stock_quantity"`
	LowStockThreshold int                    `json:"low_stock_threshold"`
	IsActive          bool                   `json:"is_active"`
	IsDefault         bool                   `json:"is_default"`
	WeightGrams       *int                   `json:"weight_grams"`
	Weight            *int                   `json:"weight"` // Alias for weight_grams
	LengthCm          *int                   `json:"length_cm"`
	Length            *int                   `json:"length"` // Alias for length_cm
	WidthCm           *int                   `json:"width_cm"`
	Width             *int                   `json:"width"` // Alias for width_cm
	HeightCm          *int                   `json:"height_cm"`
	Height            *int                   `json:"height"` // Alias for height_cm
	Barcode           *string                `json:"barcode"`
	Position          int                    `json:"position"`
}

type BulkGenerateVariantsRequest struct {
	ProductID       int      `json:"product_id" binding:"required"`
	Sizes           []string `json:"sizes" binding:"required"`
	Colors          []string `json:"colors" binding:"required"`
	BasePrice       float64  `json:"base_price"`
	StockPerVariant int      `json:"stock_per_variant"`
	Weight          int      `json:"weight"`           // Default weight in grams
	Length          int      `json:"length"`           // Default length in cm
	Width           int      `json:"width"`            // Default width in cm
	Height          int      `json:"height"`           // Default height in cm
}

type AddVariantImageRequest struct {
	VariantID int     `json:"variant_id" binding:"required"`
	ImageURL  string  `json:"image_url" binding:"required"`
	AltText   *string `json:"alt_text"`
	Position  int     `json:"position"`
	IsPrimary bool    `json:"is_primary"`
	Width     *int    `json:"width"`
	Height    *int    `json:"height"`
	Format    *string `json:"format"`
}

type ReorderImagesRequest struct {
	ImageIDs []int `json:"image_ids" binding:"required"`
}

type SetPrimaryImageRequest struct {
	ImageID int `json:"image_id" binding:"required"`
}

type UpdateVariantStockRequest struct {
	Quantity int `json:"quantity" binding:"required,min=0"`
}

type AdjustStockRequest struct {
	Delta int `json:"delta" binding:"required"`
}

type ReserveStockRequest struct {
	VariantID  int    `json:"variant_id" binding:"required"`
	Quantity   int    `json:"quantity" binding:"required,min=1"`
	CustomerID int    `json:"customer_id"`
	SessionID  string `json:"session_id"`
}

type ReserveStockResponse struct {
	ReservationID int    `json:"reservation_id"`
	ExpiresAt     string `json:"expires_at"`
	Message       string `json:"message"`
}

type CheckAvailabilityRequest struct {
	VariantID int `json:"variant_id" binding:"required"`
	Quantity  int `json:"quantity" binding:"required,min=1"`
}

type CheckAvailabilityResponse struct {
	Available      bool `json:"available"`
	AvailableStock int  `json:"available_stock"`
	RequestedStock int  `json:"requested_stock"`
}

type FindVariantRequest struct {
	ProductID int     `json:"product_id" binding:"required"`
	Size      *string `json:"size"`
	Color     *string `json:"color"`
}

type VariantSearchRequest struct {
	ProductID int     `json:"product_id"`
	Size      *string `json:"size"`
	Color     *string `json:"color"`
	IsActive  *bool   `json:"is_active"`
	LowStock  bool    `json:"low_stock"`
}
