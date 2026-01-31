package service

import (
	"fmt"
	"strings"
	"zavera/models"
	"zavera/repository"
)

type VariantService struct {
	variantRepo *repository.VariantRepository
	productRepo repository.ProductRepository
}

func NewVariantService(variantRepo *repository.VariantRepository, productRepo repository.ProductRepository) *VariantService {
	return &VariantService{
		variantRepo: variantRepo,
		productRepo: productRepo,
	}
}

func (s *VariantService) CreateVariant(variant *models.ProductVariant) error {
	// Validate product exists
	_, err := s.productRepo.FindByID(variant.ProductID)
	if err != nil {
		return fmt.Errorf("product not found")
	}

	// Generate SKU if not provided
	if variant.SKU == "" {
		variant.SKU, err = s.generateSKU(variant)
		if err != nil {
			return err
		}
	}

	// Generate variant name if not provided
	if variant.VariantName == "" {
		variant.VariantName = s.generateVariantName(variant)
	}

	return s.variantRepo.Create(variant)
}

func (s *VariantService) UpdateVariant(variant *models.ProductVariant) error {
	existing, err := s.variantRepo.GetByID(variant.ID)
	if err != nil {
		return fmt.Errorf("variant not found")
	}

	// Prevent SKU change if variant has orders (simplified check)
	if existing.SKU != variant.SKU {
		return fmt.Errorf("cannot change SKU of existing variant")
	}

	return s.variantRepo.Update(variant)
}

func (s *VariantService) GetVariant(id int) (*models.ProductVariant, error) {
	return s.variantRepo.GetByID(id)
}

func (s *VariantService) GetVariantBySKU(sku string) (*models.ProductVariant, error) {
	return s.variantRepo.GetBySKU(sku)
}

func (s *VariantService) GetProductVariants(productID int) ([]models.ProductVariant, error) {
	return s.variantRepo.GetByProductID(productID)
}

func (s *VariantService) DeleteVariant(id int) error {
	return s.variantRepo.Delete(id)
}

func (s *VariantService) BulkGenerateVariants(productID int, sizes, colors []string, basePrice float64, stockPerVariant, weight, length, width, height int) error {
	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		return fmt.Errorf("product not found")
	}

	variants := []models.ProductVariant{}
	position := 0

	for _, size := range sizes {
		for _, color := range colors {
			sku := fmt.Sprintf("%s-%s-%s", s.sanitizeSKU(product.Name), s.sanitizeSKU(size), s.sanitizeSKU(color))
			variantName := fmt.Sprintf("%s - %s", size, color)

			variant := models.ProductVariant{
				ProductID:         productID,
				SKU:               sku,
				VariantName:       variantName,
				Size:              &size,
				Color:             &color,
				StockQuantity:     stockPerVariant,
				LowStockThreshold: 5,
				IsActive:          true,
				IsDefault:         position == 0,
				Position:          position,
			}

			if basePrice > 0 {
				variant.Price = &basePrice
			}

			// Apply default dimensions if provided
			// Create new variables to avoid pointer issues
			if weight > 0 {
				w := weight
				variant.WeightGrams = &w
			}
			if length > 0 {
				l := length
				variant.LengthCm = &l
			}
			if width > 0 {
				w := width
				variant.WidthCm = &w
			}
			if height > 0 {
				h := height
				variant.HeightCm = &h
			}

			variants = append(variants, variant)
			position++
		}
	}

	return s.variantRepo.BulkCreate(variants)
}

func (s *VariantService) AddVariantImage(image *models.VariantImage) error {
	// Validate variant exists
	_, err := s.variantRepo.GetByID(image.VariantID)
	if err != nil {
		return fmt.Errorf("variant not found")
	}

	return s.variantRepo.AddImage(image)
}

func (s *VariantService) GetVariantImages(variantID int) ([]models.VariantImage, error) {
	return s.variantRepo.GetVariantImages(variantID)
}

func (s *VariantService) DeleteVariantImage(imageID int) error {
	return s.variantRepo.DeleteImage(imageID)
}

func (s *VariantService) SetPrimaryImage(variantID, imageID int) error {
	return s.variantRepo.SetPrimaryImage(variantID, imageID)
}

func (s *VariantService) ReorderImages(variantID int, imageIDs []int) error {
	for i, imageID := range imageIDs {
		err := s.variantRepo.UpdateImagePosition(imageID, i)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *VariantService) CheckAvailability(variantID, quantity int) (bool, error) {
	available, err := s.variantRepo.GetAvailableStock(variantID)
	if err != nil {
		return false, err
	}
	return available >= quantity, nil
}

func (s *VariantService) ReserveStock(variantID, customerID int, sessionID string, quantity int) (int, error) {
	// Check availability first
	available, err := s.variantRepo.GetAvailableStock(variantID)
	if err != nil {
		return 0, err
	}

	if available < quantity {
		return 0, fmt.Errorf("insufficient stock: available %d, requested %d", available, quantity)
	}

	// Reserve with 15 minute timeout
	return s.variantRepo.ReserveStock(variantID, customerID, sessionID, quantity, 15)
}

func (s *VariantService) CompleteReservation(reservationID, orderID int) error {
	return s.variantRepo.CompleteReservation(reservationID, orderID)
}

func (s *VariantService) CancelReservation(reservationID int) error {
	return s.variantRepo.CancelReservation(reservationID)
}

func (s *VariantService) UpdateStock(variantID, quantity int) error {
	if quantity < 0 {
		return fmt.Errorf("stock quantity cannot be negative")
	}
	return s.variantRepo.UpdateStock(variantID, quantity)
}

func (s *VariantService) AdjustStock(variantID, delta int) error {
	return s.variantRepo.AdjustStock(variantID, delta)
}

func (s *VariantService) GetLowStockVariants() ([]models.LowStockVariant, error) {
	return s.variantRepo.GetLowStockVariants()
}

func (s *VariantService) GetStockSummary(productID int) ([]models.VariantStockSummary, error) {
	return s.variantRepo.GetStockSummary(productID)
}

func (s *VariantService) GetVariantAttributes() ([]models.VariantAttribute, error) {
	return s.variantRepo.GetAttributes()
}

func (s *VariantService) GetProductWithVariants(productID int) (*models.ProductWithVariants, error) {
	return s.variantRepo.GetProductWithVariants(productID)
}

func (s *VariantService) GetAvailableOptions(productID int) (map[string][]string, error) {
	options := make(map[string][]string)

	fields := []string{"size", "color", "material", "pattern", "fit", "sleeve"}
	for _, field := range fields {
		values, err := s.variantRepo.GetUniqueValues(productID, field)
		if err == nil && len(values) > 0 {
			options[field] = values
		}
	}

	return options, nil
}

func (s *VariantService) FindVariantByAttributes(productID int, size, color string) (*models.ProductVariant, error) {
	filters := map[string]interface{}{
		"product_id": productID,
		"is_active":  true,
	}

	if size != "" {
		filters["size"] = size
	}
	if color != "" {
		filters["color"] = color
	}

	variants, err := s.variantRepo.Search(filters)
	if err != nil {
		return nil, err
	}

	if len(variants) == 0 {
		return nil, fmt.Errorf("variant not found")
	}

	return &variants[0], nil
}

func (s *VariantService) CleanupExpiredReservations() error {
	return s.variantRepo.CleanExpiredReservations()
}

// Helper functions
func (s *VariantService) generateSKU(variant *models.ProductVariant) (string, error) {
	product, err := s.productRepo.FindByID(variant.ProductID)
	if err != nil {
		return "", err
	}

	parts := []string{s.sanitizeSKU(product.Name)}

	if variant.Size != nil && *variant.Size != "" {
		parts = append(parts, s.sanitizeSKU(*variant.Size))
	}
	if variant.Color != nil && *variant.Color != "" {
		parts = append(parts, s.sanitizeSKU(*variant.Color))
	}

	baseSKU := strings.Join(parts, "-")

	// Check uniqueness
	_, err = s.variantRepo.GetBySKU(baseSKU)
	if err == nil {
		// SKU exists, append number
		for i := 1; i < 1000; i++ {
			testSKU := fmt.Sprintf("%s-%d", baseSKU, i)
			_, err = s.variantRepo.GetBySKU(testSKU)
			if err != nil {
				return testSKU, nil
			}
		}
		return "", fmt.Errorf("could not generate unique SKU")
	}

	return baseSKU, nil
}

func (s *VariantService) sanitizeSKU(input string) string {
	// Convert to uppercase, replace spaces and special chars
	result := strings.ToUpper(input)
	result = strings.ReplaceAll(result, " ", "-")
	result = strings.ReplaceAll(result, "_", "-")

	// Remove non-alphanumeric except dash
	var cleaned strings.Builder
	for _, r := range result {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
			cleaned.WriteRune(r)
		}
	}

	// Remove consecutive dashes
	result = cleaned.String()
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}

	// Trim dashes
	result = strings.Trim(result, "-")

	// Limit length
	if len(result) > 50 {
		result = result[:50]
	}

	return result
}

func (s *VariantService) generateVariantName(variant *models.ProductVariant) string {
	parts := []string{}

	if variant.Size != nil && *variant.Size != "" {
		parts = append(parts, *variant.Size)
	}
	if variant.Color != nil && *variant.Color != "" {
		parts = append(parts, *variant.Color)
	}
	if variant.Material != nil && *variant.Material != "" {
		parts = append(parts, *variant.Material)
	}
	if variant.Pattern != nil && *variant.Pattern != "" {
		parts = append(parts, *variant.Pattern)
	}

	if len(parts) == 0 {
		return "Default"
	}

	return strings.Join(parts, " - ")
}
