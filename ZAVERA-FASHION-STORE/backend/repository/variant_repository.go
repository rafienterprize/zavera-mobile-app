package repository

import (
	"database/sql"
	"fmt"
	"log"
	"zavera/models"
)

type VariantRepository struct {
	db *sql.DB
}

func NewVariantRepository(db *sql.DB) *VariantRepository {
	return &VariantRepository{db: db}
}

func (r *VariantRepository) Create(variant *models.ProductVariant) error {
	query := `
		INSERT INTO product_variants (
			product_id, sku, variant_name, size, color, color_hex,
			material, pattern, fit, sleeve, custom_attributes,
			price, compare_at_price, cost_per_item,
			stock_quantity, reserved_stock, low_stock_threshold,
			is_active, is_default, weight_grams, length_cm, width_cm, height_cm,
			barcode, position
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRow(
		query,
		variant.ProductID, variant.SKU, variant.VariantName,
		variant.Size, variant.Color, variant.ColorHex,
		variant.Material, variant.Pattern, variant.Fit, variant.Sleeve,
		variant.CustomAttributes, variant.Price, variant.CompareAtPrice,
		variant.CostPerItem, variant.StockQuantity, variant.ReservedStock,
		variant.LowStockThreshold, variant.IsActive, variant.IsDefault,
		variant.WeightGrams, variant.LengthCm, variant.WidthCm, variant.HeightCm,
		variant.Barcode, variant.Position,
	).Scan(&variant.ID, &variant.CreatedAt, &variant.UpdatedAt)
}

func (r *VariantRepository) Update(variant *models.ProductVariant) error {
	query := `
		UPDATE product_variants SET
			sku = $1, variant_name = $2, size = $3, color = $4, color_hex = $5,
			material = $6, pattern = $7, fit = $8, sleeve = $9, custom_attributes = $10,
			price = $11, compare_at_price = $12, cost_per_item = $13,
			stock_quantity = $14, low_stock_threshold = $15,
			is_active = $16, is_default = $17, weight_grams = $18, 
			length_cm = $19, width_cm = $20, height_cm = $21,
			barcode = $22, position = $23, updated_at = CURRENT_TIMESTAMP
		WHERE id = $24
		RETURNING updated_at`

	return r.db.QueryRow(
		query,
		variant.SKU, variant.VariantName, variant.Size, variant.Color, variant.ColorHex,
		variant.Material, variant.Pattern, variant.Fit, variant.Sleeve,
		variant.CustomAttributes, variant.Price, variant.CompareAtPrice,
		variant.CostPerItem, variant.StockQuantity, variant.LowStockThreshold,
		variant.IsActive, variant.IsDefault, variant.WeightGrams,
		variant.LengthCm, variant.WidthCm, variant.HeightCm,
		variant.Barcode, variant.Position, variant.ID,
	).Scan(&variant.UpdatedAt)
}

func (r *VariantRepository) GetByID(id int) (*models.ProductVariant, error) {
	variant := &models.ProductVariant{}
	query := `
		SELECT id, product_id, sku, variant_name, size, color, color_hex,
			material, pattern, fit, sleeve, custom_attributes,
			price, compare_at_price, cost_per_item,
			stock_quantity, reserved_stock, low_stock_threshold,
			is_active, is_default, weight_grams, length_cm, width_cm, height_cm,
			barcode, position,
			created_at, updated_at,
			get_available_stock(id) as available_stock
		FROM product_variants
		WHERE id = $1`

	err := r.db.QueryRow(query, id).Scan(
		&variant.ID, &variant.ProductID, &variant.SKU, &variant.VariantName,
		&variant.Size, &variant.Color, &variant.ColorHex,
		&variant.Material, &variant.Pattern, &variant.Fit, &variant.Sleeve,
		&variant.CustomAttributes, &variant.Price, &variant.CompareAtPrice,
		&variant.CostPerItem, &variant.StockQuantity, &variant.ReservedStock,
		&variant.LowStockThreshold, &variant.IsActive, &variant.IsDefault,
		&variant.WeightGrams, &variant.LengthCm, &variant.WidthCm, &variant.HeightCm,
		&variant.Barcode, &variant.Position,
		&variant.CreatedAt, &variant.UpdatedAt, &variant.AvailableStock,
	)

	if err != nil {
		return nil, err
	}

	variant.Images, _ = r.GetVariantImages(variant.ID)
	return variant, nil
}

func (r *VariantRepository) GetBySKU(sku string) (*models.ProductVariant, error) {
	variant := &models.ProductVariant{}
	query := `
		SELECT id, product_id, sku, variant_name, size, color, color_hex,
			material, pattern, fit, sleeve, custom_attributes,
			price, compare_at_price, cost_per_item,
			stock_quantity, reserved_stock, low_stock_threshold,
			is_active, is_default, weight_grams, length_cm, width_cm, height_cm,
			barcode, position,
			created_at, updated_at,
			get_available_stock(id) as available_stock
		FROM product_variants
		WHERE sku = $1`

	err := r.db.QueryRow(query, sku).Scan(
		&variant.ID, &variant.ProductID, &variant.SKU, &variant.VariantName,
		&variant.Size, &variant.Color, &variant.ColorHex,
		&variant.Material, &variant.Pattern, &variant.Fit, &variant.Sleeve,
		&variant.CustomAttributes, &variant.Price, &variant.CompareAtPrice,
		&variant.CostPerItem, &variant.StockQuantity, &variant.ReservedStock,
		&variant.LowStockThreshold, &variant.IsActive, &variant.IsDefault,
		&variant.WeightGrams, &variant.LengthCm, &variant.WidthCm, &variant.HeightCm,
		&variant.Barcode, &variant.Position,
		&variant.CreatedAt, &variant.UpdatedAt, &variant.AvailableStock,
	)

	if err != nil {
		return nil, err
	}

	variant.Images, _ = r.GetVariantImages(variant.ID)
	return variant, nil
}

func (r *VariantRepository) GetByProductID(productID int) ([]models.ProductVariant, error) {
	query := `
		SELECT id, product_id, sku, variant_name, size, color, color_hex,
			material, pattern, fit, sleeve, custom_attributes,
			price, compare_at_price, cost_per_item,
			stock_quantity, reserved_stock, low_stock_threshold,
			is_active, is_default, weight_grams, length_cm, width_cm, height_cm, 
			barcode, position,
			created_at, updated_at,
			get_available_stock(id) as available_stock
		FROM product_variants
		WHERE product_id = $1
		ORDER BY position ASC, id ASC`

	rows, err := r.db.Query(query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	variants := []models.ProductVariant{}
	for rows.Next() {
		var v models.ProductVariant
		err := rows.Scan(
			&v.ID, &v.ProductID, &v.SKU, &v.VariantName,
			&v.Size, &v.Color, &v.ColorHex,
			&v.Material, &v.Pattern, &v.Fit, &v.Sleeve,
			&v.CustomAttributes, &v.Price, &v.CompareAtPrice,
			&v.CostPerItem, &v.StockQuantity, &v.ReservedStock,
			&v.LowStockThreshold, &v.IsActive, &v.IsDefault,
			&v.WeightGrams, &v.LengthCm, &v.WidthCm, &v.HeightCm,
			&v.Barcode, &v.Position,
			&v.CreatedAt, &v.UpdatedAt, &v.AvailableStock,
		)
		if err != nil {
			continue
		}
		v.Images, _ = r.GetVariantImages(v.ID)
		variants = append(variants, v)
	}

	return variants, nil
}

func (r *VariantRepository) Delete(id int) error {
	// Check if variant has orders
	var orderCount int
	err := r.db.QueryRow("SELECT COUNT(*) FROM order_items WHERE variant_id = $1", id).Scan(&orderCount)
	if err != nil {
		log.Printf("❌ Delete variant %d: Error checking orders: %v", id, err)
		return fmt.Errorf("failed to check variant orders: %w", err)
	}
	
	log.Printf("🔍 Delete variant %d: Found %d orders", id, orderCount)
	
	if orderCount > 0 {
		return fmt.Errorf("cannot delete variant: it has %d existing order(s). Variants with orders cannot be deleted to maintain data integrity", orderCount)
	}

	result, err := r.db.Exec("DELETE FROM product_variants WHERE id = $1", id)
	if err != nil {
		log.Printf("❌ Delete variant %d: Delete failed: %v", id, err)
		return fmt.Errorf("failed to delete variant: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	log.Printf("✅ Delete variant %d: Deleted %d rows", id, rowsAffected)
	
	return nil
}

func (r *VariantRepository) BulkCreate(variants []models.ProductVariant) error {
	if len(variants) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO product_variants (
			product_id, sku, variant_name, size, color, color_hex,
			material, pattern, fit, sleeve, custom_attributes,
			price, compare_at_price, cost_per_item,
			stock_quantity, reserved_stock, low_stock_threshold,
			is_active, is_default, weight_grams, length_cm, width_cm, height_cm, barcode, position
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25)
		RETURNING id, created_at, updated_at`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := range variants {
		v := &variants[i]
		err = stmt.QueryRow(
			v.ProductID, v.SKU, v.VariantName,
			v.Size, v.Color, v.ColorHex,
			v.Material, v.Pattern, v.Fit, v.Sleeve,
			v.CustomAttributes, v.Price, v.CompareAtPrice,
			v.CostPerItem, v.StockQuantity, v.ReservedStock,
			v.LowStockThreshold, v.IsActive, v.IsDefault,
			v.WeightGrams, v.LengthCm, v.WidthCm, v.HeightCm, v.Barcode, v.Position,
		).Scan(&v.ID, &v.CreatedAt, &v.UpdatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Variant Images
func (r *VariantRepository) AddImage(image *models.VariantImage) error {
	query := `
		INSERT INTO variant_images (variant_id, image_url, alt_text, position, is_primary, width, height, format)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at`

	return r.db.QueryRow(
		query, image.VariantID, image.ImageURL, image.AltText,
		image.Position, image.IsPrimary, image.Width, image.Height, image.Format,
	).Scan(&image.ID, &image.CreatedAt)
}

func (r *VariantRepository) GetVariantImages(variantID int) ([]models.VariantImage, error) {
	query := `
		SELECT id, variant_id, image_url, alt_text, position, is_primary, width, height, format, created_at
		FROM variant_images
		WHERE variant_id = $1
		ORDER BY position ASC, id ASC`

	rows, err := r.db.Query(query, variantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	images := []models.VariantImage{}
	for rows.Next() {
		var img models.VariantImage
		err := rows.Scan(
			&img.ID, &img.VariantID, &img.ImageURL, &img.AltText,
			&img.Position, &img.IsPrimary, &img.Width, &img.Height,
			&img.Format, &img.CreatedAt,
		)
		if err != nil {
			continue
		}
		images = append(images, img)
	}

	return images, nil
}

func (r *VariantRepository) DeleteImage(imageID int) error {
	_, err := r.db.Exec("DELETE FROM variant_images WHERE id = $1", imageID)
	return err
}

func (r *VariantRepository) UpdateImagePosition(imageID, position int) error {
	_, err := r.db.Exec("UPDATE variant_images SET position = $1 WHERE id = $2", position, imageID)
	return err
}

func (r *VariantRepository) SetPrimaryImage(variantID, imageID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE variant_images SET is_primary = false WHERE variant_id = $1", variantID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE variant_images SET is_primary = true WHERE id = $1 AND variant_id = $2", imageID, variantID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Stock Management
func (r *VariantRepository) ReserveStock(variantID, customerID int, sessionID string, quantity, timeoutMinutes int) (int, error) {
	var reservationID int
	query := `SELECT reserve_stock($1, $2, $3, $4, $5)`
	err := r.db.QueryRow(query, variantID, customerID, sessionID, quantity, timeoutMinutes).Scan(&reservationID)
	return reservationID, err
}

func (r *VariantRepository) CompleteReservation(reservationID, orderID int) error {
	_, err := r.db.Exec("SELECT complete_reservation($1, $2)", reservationID, orderID)
	return err
}

func (r *VariantRepository) CancelReservation(reservationID int) error {
	_, err := r.db.Exec("SELECT cancel_reservation($1)", reservationID)
	return err
}

func (r *VariantRepository) CleanExpiredReservations() error {
	_, err := r.db.Exec("SELECT clean_expired_reservations()")
	return err
}

func (r *VariantRepository) GetAvailableStock(variantID int) (int, error) {
	var stock int
	err := r.db.QueryRow("SELECT get_available_stock($1)", variantID).Scan(&stock)
	return stock, err
}

func (r *VariantRepository) UpdateStock(variantID, quantity int) error {
	query := `UPDATE product_variants SET stock_quantity = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := r.db.Exec(query, quantity, variantID)
	return err
}

func (r *VariantRepository) AdjustStock(variantID, delta int) error {
	query := `
		UPDATE product_variants 
		SET stock_quantity = stock_quantity + $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2 AND stock_quantity + $1 >= 0`
	result, err := r.db.Exec(query, delta, variantID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("insufficient stock or variant not found")
	}
	return nil
}

// Low Stock
func (r *VariantRepository) GetLowStockVariants() ([]models.LowStockVariant, error) {
	query := `SELECT id, product_id, product_name, sku, variant_name, size, color, 
		stock_quantity, low_stock_threshold, available_stock 
		FROM low_stock_variants`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	variants := []models.LowStockVariant{}
	for rows.Next() {
		var v models.LowStockVariant
		err := rows.Scan(
			&v.ID, &v.ProductID, &v.ProductName, &v.SKU, &v.VariantName,
			&v.Size, &v.Color, &v.StockQuantity, &v.LowStockThreshold, &v.AvailableStock,
		)
		if err != nil {
			continue
		}
		variants = append(variants, v)
	}

	return variants, nil
}

func (r *VariantRepository) GetStockSummary(productID int) ([]models.VariantStockSummary, error) {
	query := `SELECT variant_id, product_id, product_name, sku, variant_name,
		stock_quantity, reserved_quantity, available_quantity
		FROM variant_stock_summary
		WHERE product_id = $1`

	rows, err := r.db.Query(query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summaries := []models.VariantStockSummary{}
	for rows.Next() {
		var s models.VariantStockSummary
		err := rows.Scan(
			&s.VariantID, &s.ProductID, &s.ProductName, &s.SKU, &s.VariantName,
			&s.StockQuantity, &s.ReservedQuantity, &s.AvailableQuantity,
		)
		if err != nil {
			continue
		}
		summaries = append(summaries, s)
	}

	return summaries, nil
}

// Variant Attributes
func (r *VariantRepository) GetAttributes() ([]models.VariantAttribute, error) {
	query := `SELECT id, name, display_name, type, options, sort_order, is_active, created_at
		FROM variant_attributes
		WHERE is_active = true
		ORDER BY sort_order ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	attributes := []models.VariantAttribute{}
	for rows.Next() {
		var attr models.VariantAttribute
		err := rows.Scan(
			&attr.ID, &attr.Name, &attr.DisplayName, &attr.Type,
			&attr.Options, &attr.SortOrder, &attr.IsActive, &attr.CreatedAt,
		)
		if err != nil {
			continue
		}
		attributes = append(attributes, attr)
	}

	return attributes, nil
}

// Search and Filter
func (r *VariantRepository) Search(filters map[string]interface{}) ([]models.ProductVariant, error) {
	query := `
		SELECT id, product_id, sku, variant_name, size, color, color_hex,
			material, pattern, fit, sleeve, custom_attributes,
			price, compare_at_price, cost_per_item,
			stock_quantity, reserved_stock, low_stock_threshold,
			is_active, is_default, weight_grams, barcode, position,
			created_at, updated_at,
			get_available_stock(id) as available_stock
		FROM product_variants
		WHERE 1=1`

	args := []interface{}{}
	argCount := 1

	if productID, ok := filters["product_id"].(int); ok {
		query += fmt.Sprintf(" AND product_id = $%d", argCount)
		args = append(args, productID)
		argCount++
	}

	if size, ok := filters["size"].(string); ok {
		query += fmt.Sprintf(" AND size = $%d", argCount)
		args = append(args, size)
		argCount++
	}

	if color, ok := filters["color"].(string); ok {
		query += fmt.Sprintf(" AND color = $%d", argCount)
		args = append(args, color)
		argCount++
	}

	if isActive, ok := filters["is_active"].(bool); ok {
		query += fmt.Sprintf(" AND is_active = $%d", argCount)
		args = append(args, isActive)
		argCount++
	}

	if lowStock, ok := filters["low_stock"].(bool); ok && lowStock {
		query += " AND stock_quantity <= low_stock_threshold"
	}

	query += " ORDER BY position ASC, id ASC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	variants := []models.ProductVariant{}
	for rows.Next() {
		var v models.ProductVariant
		err := rows.Scan(
			&v.ID, &v.ProductID, &v.SKU, &v.VariantName,
			&v.Size, &v.Color, &v.ColorHex,
			&v.Material, &v.Pattern, &v.Fit, &v.Sleeve,
			&v.CustomAttributes, &v.Price, &v.CompareAtPrice,
			&v.CostPerItem, &v.StockQuantity, &v.ReservedStock,
			&v.LowStockThreshold, &v.IsActive, &v.IsDefault,
			&v.WeightGrams, &v.Barcode, &v.Position,
			&v.CreatedAt, &v.UpdatedAt, &v.AvailableStock,
		)
		if err != nil {
			continue
		}
		v.Images, _ = r.GetVariantImages(v.ID)
		variants = append(variants, v)
	}

	return variants, nil
}

func (r *VariantRepository) GetProductWithVariants(productID int) (*models.ProductWithVariants, error) {
	// Get product
	var product models.Product
	query := `SELECT id, name, slug, description, price, stock, weight, length, width, height, 
		is_active, category, subcategory, created_at, updated_at 
		FROM products WHERE id = $1`
	err := r.db.QueryRow(query, productID).Scan(
		&product.ID, &product.Name, &product.Slug, &product.Description, &product.Price,
		&product.Stock, &product.Weight, &product.Length, &product.Width, &product.Height,
		&product.IsActive, &product.Category, &product.Subcategory,
		&product.CreatedAt, &product.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Get variants
	variants, err := r.GetByProductID(productID)
	if err != nil {
		return nil, err
	}

	result := &models.ProductWithVariants{
		Product:  product,
		Variants: variants,
	}

	// Calculate price range
	if len(variants) > 0 {
		minPrice := product.Price
		maxPrice := product.Price

		for _, v := range variants {
			if !v.IsActive {
				continue
			}
			price := product.Price
			if v.Price != nil {
				price = *v.Price
			}
			if price < minPrice {
				minPrice = price
			}
			if price > maxPrice {
				maxPrice = price
			}
		}

		if minPrice != maxPrice {
			result.PriceRange = &models.PriceRange{
				MinPrice: minPrice,
				MaxPrice: maxPrice,
			}
		}
	}

	return result, nil
}

func (r *VariantRepository) GetUniqueValues(productID int, field string) ([]string, error) {
	validFields := map[string]bool{
		"size": true, "color": true, "material": true,
		"pattern": true, "fit": true, "sleeve": true,
	}

	if !validFields[field] {
		return nil, fmt.Errorf("invalid field: %s", field)
	}

	query := fmt.Sprintf(`
		SELECT DISTINCT %s 
		FROM product_variants 
		WHERE product_id = $1 AND %s IS NOT NULL AND is_active = true
		ORDER BY %s`, field, field, field)

	rows, err := r.db.Query(query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	values := []string{}
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			continue
		}
		values = append(values, value)
	}

	return values, nil
}
