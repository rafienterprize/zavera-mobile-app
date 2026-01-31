package handler

import (
	"log"
	"net/http"
	"strconv"
	"zavera/dto"
	"zavera/models"
	"zavera/service"

	"github.com/gin-gonic/gin"
)

type VariantHandler struct {
	variantService *service.VariantService
}

func NewVariantHandler(variantService *service.VariantService) *VariantHandler {
	return &VariantHandler{variantService: variantService}
}

func (h *VariantHandler) CreateVariant(c *gin.Context) {
	var req dto.CreateVariantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ Variant JSON Binding Error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	log.Printf("✅ Received variant creation request:")
	log.Printf("   Product ID: %d", req.ProductID)
	log.Printf("   Size: %v", req.Size)
	log.Printf("   Color: %v", req.Color)
	log.Printf("   ColorHex: %v", req.ColorHex)
	log.Printf("   Stock: %d", req.StockQuantity)
	log.Printf("   Price: %v", req.Price)
	log.Printf("   Weight: %v", req.WeightGrams)

	variant := &models.ProductVariant{
		ProductID:         req.ProductID,
		SKU:               req.SKU,
		VariantName:       req.VariantName,
		Size:              req.Size,
		Color:             req.Color,
		ColorHex:          req.ColorHex,
		Material:          req.Material,
		Pattern:           req.Pattern,
		Fit:               req.Fit,
		Sleeve:            req.Sleeve,
		CustomAttributes:  req.CustomAttributes,
		Price:             req.Price,
		CompareAtPrice:    req.CompareAtPrice,
		CostPerItem:       req.CostPerItem,
		StockQuantity:     req.StockQuantity,
		LowStockThreshold: req.LowStockThreshold,
		IsActive:          req.IsActive,
		IsDefault:         req.IsDefault,
		WeightGrams:       req.WeightGrams,
		LengthCm:          req.LengthCm,
		WidthCm:           req.WidthCm,
		HeightCm:          req.HeightCm,
		Barcode:           req.Barcode,
		Position:          req.Position,
	}

	// Handle alias fields (weight, length, width, height)
	if req.Weight != nil && req.WeightGrams == nil {
		variant.WeightGrams = req.Weight
	}
	if req.Length != nil && req.LengthCm == nil {
		variant.LengthCm = req.Length
	}
	if req.Width != nil && req.WidthCm == nil {
		variant.WidthCm = req.Width
	}
	if req.Height != nil && req.HeightCm == nil {
		variant.HeightCm = req.Height
	}

	log.Printf("🔧 Creating variant in database...")
	if err := h.variantService.CreateVariant(variant); err != nil {
		log.Printf("❌ Variant creation failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "create_failed",
			"message": err.Error(),
		})
		return
	}

	log.Printf("✅ Variant created successfully! ID: %d", variant.ID)
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Variant created successfully",
		"data":    variant,
	})
}

func (h *VariantHandler) UpdateVariant(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid variant ID"})
		return
	}

	var req dto.UpdateVariantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	variant, err := h.variantService.GetVariant(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Variant not found"})
		return
	}

	variant.SKU = req.SKU
	variant.VariantName = req.VariantName
	variant.Size = req.Size
	variant.Color = req.Color
	variant.ColorHex = req.ColorHex
	variant.Material = req.Material
	variant.Pattern = req.Pattern
	variant.Fit = req.Fit
	variant.Sleeve = req.Sleeve
	variant.CustomAttributes = req.CustomAttributes
	variant.Price = req.Price
	variant.CompareAtPrice = req.CompareAtPrice
	variant.CostPerItem = req.CostPerItem
	variant.StockQuantity = req.StockQuantity
	variant.LowStockThreshold = req.LowStockThreshold
	variant.IsActive = req.IsActive
	variant.IsDefault = req.IsDefault
	variant.WeightGrams = req.WeightGrams
	variant.LengthCm = req.LengthCm
	variant.WidthCm = req.WidthCm
	variant.HeightCm = req.HeightCm
	variant.Barcode = req.Barcode
	variant.Position = req.Position

	// Handle alias fields (weight, length, width, height)
	if req.Weight != nil && req.WeightGrams == nil {
		variant.WeightGrams = req.Weight
	}
	if req.Length != nil && req.LengthCm == nil {
		variant.LengthCm = req.Length
	}
	if req.Width != nil && req.WidthCm == nil {
		variant.WidthCm = req.Width
	}
	if req.Height != nil && req.HeightCm == nil {
		variant.HeightCm = req.Height
	}

	if err := h.variantService.UpdateVariant(variant); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, variant)
}

func (h *VariantHandler) GetVariant(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid variant ID"})
		return
	}

	variant, err := h.variantService.GetVariant(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Variant not found"})
		return
	}

	c.JSON(http.StatusOK, variant)
}

func (h *VariantHandler) GetVariantBySKU(c *gin.Context) {
	sku := c.Param("sku")
	if sku == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SKU is required"})
		return
	}

	variant, err := h.variantService.GetVariantBySKU(sku)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Variant not found"})
		return
	}

	c.JSON(http.StatusOK, variant)
}

func (h *VariantHandler) GetProductVariants(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	variants, err := h.variantService.GetProductVariants(productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, variants)
}

func (h *VariantHandler) DeleteVariant(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("❌ DeleteVariant: Invalid ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid variant ID"})
		return
	}

	log.Printf("🗑️ DeleteVariant: Attempting to delete variant ID %d", id)

	if err := h.variantService.DeleteVariant(id); err != nil {
		log.Printf("❌ DeleteVariant error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("✅ DeleteVariant: Successfully deleted variant ID %d", id)
	c.JSON(http.StatusOK, gin.H{"message": "Variant deleted successfully"})
}

func (h *VariantHandler) BulkGenerateVariants(c *gin.Context) {
	var req dto.BulkGenerateVariantsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ BulkGenerate bind error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.Sizes) == 0 || len(req.Colors) == 0 {
		log.Printf("❌ BulkGenerate validation error: sizes=%v, colors=%v", req.Sizes, req.Colors)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Sizes and colors are required"})
		return
	}

	// Check if product already has variants
	existingVariants, _ := h.variantService.GetProductVariants(req.ProductID)
	if len(existingVariants) > 0 {
		log.Printf("⚠️ BulkGenerate: Product %d already has %d variants", req.ProductID, len(existingVariants))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Product already has variants. Please delete existing variants first or add new variants manually.",
			"existing_count": len(existingVariants),
		})
		return
	}

	log.Printf("🔧 BulkGenerate request: product_id=%d, sizes=%v, colors=%v, price=%.2f, stock=%d, weight=%d, length=%d, width=%d, height=%d",
		req.ProductID, req.Sizes, req.Colors, req.BasePrice, req.StockPerVariant, req.Weight, req.Length, req.Width, req.Height)

	err := h.variantService.BulkGenerateVariants(
		req.ProductID,
		req.Sizes,
		req.Colors,
		req.BasePrice,
		req.StockPerVariant,
		req.Weight,
		req.Length,
		req.Width,
		req.Height,
	)

	if err != nil {
		log.Printf("❌ BulkGenerate service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	variants, _ := h.variantService.GetProductVariants(req.ProductID)
	log.Printf("✅ BulkGenerate success: generated %d variants", len(variants))
	c.JSON(http.StatusCreated, gin.H{
		"message":  "Variants generated successfully",
		"count":    len(variants),
		"variants": variants,
	})
}

func (h *VariantHandler) AddVariantImage(c *gin.Context) {
	var req dto.AddVariantImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	image := &models.VariantImage{
		VariantID: req.VariantID,
		ImageURL:  req.ImageURL,
		AltText:   req.AltText,
		Position:  req.Position,
		IsPrimary: req.IsPrimary,
		Width:     req.Width,
		Height:    req.Height,
		Format:    req.Format,
	}

	if err := h.variantService.AddVariantImage(image); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, image)
}

func (h *VariantHandler) GetVariantImages(c *gin.Context) {
	variantID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid variant ID"})
		return
	}

	images, err := h.variantService.GetVariantImages(variantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, images)
}

func (h *VariantHandler) DeleteVariantImage(c *gin.Context) {
	imageID, err := strconv.Atoi(c.Param("imageId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image ID"})
		return
	}

	if err := h.variantService.DeleteVariantImage(imageID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Image deleted successfully"})
}

func (h *VariantHandler) SetPrimaryImage(c *gin.Context) {
	variantID, err := strconv.Atoi(c.Param("variantId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid variant ID"})
		return
	}

	var req dto.SetPrimaryImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.variantService.SetPrimaryImage(variantID, req.ImageID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Primary image set successfully"})
}

func (h *VariantHandler) ReorderImages(c *gin.Context) {
	variantID, err := strconv.Atoi(c.Param("variantId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid variant ID"})
		return
	}

	var req dto.ReorderImagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.variantService.ReorderImages(variantID, req.ImageIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Images reordered successfully"})
}

func (h *VariantHandler) CheckAvailability(c *gin.Context) {
	var req dto.CheckAvailabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	available, err := h.variantService.CheckAvailability(req.VariantID, req.Quantity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	variant, _ := h.variantService.GetVariant(req.VariantID)
	availableStock := 0
	if variant != nil && variant.AvailableStock != nil {
		availableStock = *variant.AvailableStock
	}

	c.JSON(http.StatusOK, dto.CheckAvailabilityResponse{
		Available:      available,
		AvailableStock: availableStock,
		RequestedStock: req.Quantity,
	})
}

func (h *VariantHandler) ReserveStock(c *gin.Context) {
	var req dto.ReserveStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reservationID, err := h.variantService.ReserveStock(
		req.VariantID,
		req.CustomerID,
		req.SessionID,
		req.Quantity,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ReserveStockResponse{
		ReservationID: reservationID,
		ExpiresAt:     "15 minutes from now",
		Message:       "Stock reserved successfully",
	})
}

func (h *VariantHandler) UpdateStock(c *gin.Context) {
	variantID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid variant ID"})
		return
	}

	var req dto.UpdateVariantStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.variantService.UpdateStock(variantID, req.Quantity); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	variant, _ := h.variantService.GetVariant(variantID)
	c.JSON(http.StatusOK, gin.H{
		"message": "Stock updated successfully",
		"variant": variant,
	})
}

func (h *VariantHandler) AdjustStock(c *gin.Context) {
	variantID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid variant ID"})
		return
	}

	var req dto.AdjustStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.variantService.AdjustStock(variantID, req.Delta); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	variant, _ := h.variantService.GetVariant(variantID)
	c.JSON(http.StatusOK, gin.H{
		"message": "Stock adjusted successfully",
		"variant": variant,
	})
}

func (h *VariantHandler) GetLowStockVariants(c *gin.Context) {
	variants, err := h.variantService.GetLowStockVariants()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, variants)
}

func (h *VariantHandler) GetStockSummary(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	summary, err := h.variantService.GetStockSummary(productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

func (h *VariantHandler) GetVariantAttributes(c *gin.Context) {
	attributes, err := h.variantService.GetVariantAttributes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, attributes)
}

func (h *VariantHandler) GetProductWithVariants(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	product, err := h.variantService.GetProductWithVariants(productID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, product)
}

func (h *VariantHandler) GetAvailableOptions(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	options, err := h.variantService.GetAvailableOptions(productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, options)
}

func (h *VariantHandler) FindVariant(c *gin.Context) {
	var req dto.FindVariantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	size := ""
	color := ""
	if req.Size != nil {
		size = *req.Size
	}
	if req.Color != nil {
		color = *req.Color
	}

	variant, err := h.variantService.FindVariantByAttributes(req.ProductID, size, color)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Variant not found"})
		return
	}

	c.JSON(http.StatusOK, variant)
}
