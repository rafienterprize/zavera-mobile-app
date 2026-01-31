package service

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"zavera/dto"
)

var (
	ErrProductNotFound = errors.New("product not found")
	ErrInvalidSlug     = errors.New("invalid slug format")
	ErrDuplicateSlug   = errors.New("slug already exists")
)

type AdminProductService interface {
	GetAllProductsAdmin(page, pageSize int, category string, includeInactive bool) ([]dto.AdminProductResponse, int, error)
	CreateProduct(req dto.CreateProductRequest) (*dto.AdminProductResponse, error)
	UpdateProduct(id int, req dto.UpdateProductRequest) (*dto.AdminProductResponse, error)
	UpdateStock(id int, req dto.UpdateStockRequest) (*dto.AdminProductResponse, error)
	DeleteProduct(id int) error
	AddProductImage(productID int, req dto.AddProductImageRequest) (*dto.ProductImageResponse, error)
	DeleteProductImage(imageID int) error
}

type adminProductService struct {
	db *sql.DB
}

func NewAdminProductService(db *sql.DB) AdminProductService {
	return &adminProductService{db: db}
}

func (s *adminProductService) GetAllProductsAdmin(page, pageSize int, category string, includeInactive bool) ([]dto.AdminProductResponse, int, error) {
	offset := (page - 1) * pageSize

	// Build query
	whereClause := ""
	args := []interface{}{}
	argIndex := 1

	if !includeInactive {
		whereClause = "WHERE is_active = true"
	}

	if category != "" {
		if whereClause == "" {
			whereClause = "WHERE"
		} else {
			whereClause += " AND"
		}
		whereClause += fmt.Sprintf(" LOWER(category) = LOWER($%d)", argIndex)
		args = append(args, category)
		argIndex++
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM products %s", whereClause)
	var total int
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get products
	query := fmt.Sprintf(`
		SELECT id, name, slug, description, price, stock, 
		       COALESCE(weight, 500) as weight,
		       COALESCE(length, 30) as length,
		       COALESCE(width, 20) as width,
		       COALESCE(height, 5) as height,
		       COALESCE(category, 'wanita') as category, 
		       COALESCE(subcategory, '') as subcategory,
		       COALESCE(brand, '') as brand,
		       COALESCE(material, '') as material,
		       is_active, created_at, updated_at
		FROM products
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, pageSize, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []dto.AdminProductResponse
	for rows.Next() {
		var p dto.AdminProductResponse
		var createdAt, updatedAt sql.NullTime

		err := rows.Scan(
			&p.ID, &p.Name, &p.Slug, &p.Description, &p.Price, &p.Stock,
			&p.Weight, &p.Length, &p.Width, &p.Height, &p.Category, &p.Subcategory,
			&p.Brand, &p.Material, &p.IsActive,
			&createdAt, &updatedAt,
		)
		if err != nil {
			continue
		}

		if createdAt.Valid {
			p.CreatedAt = dto.FormatTime(createdAt.Time)
		}
		if updatedAt.Valid {
			p.UpdatedAt = dto.FormatTime(updatedAt.Time)
		}

		// Load images
		p.Images = s.getProductImages(p.ID)

		products = append(products, p)
	}

	return products, total, nil
}

func (s *adminProductService) CreateProduct(req dto.CreateProductRequest) (*dto.AdminProductResponse, error) {
	// Generate slug if not provided
	slug := req.Slug
	if slug == "" {
		slug = s.generateSlug(req.Name)
	}

	// Validate slug
	if !s.isValidSlug(slug) {
		return nil, ErrInvalidSlug
	}

	// Check slug uniqueness
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM products WHERE slug = $1)", slug).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrDuplicateSlug
	}

	// Default values
	weight := 500
	if req.Weight > 0 {
		weight = req.Weight
	}

	length := 30
	if req.Length > 0 {
		length = req.Length
	}

	width := 20
	if req.Width > 0 {
		width = req.Width
	}

	height := 5
	if req.Height > 0 {
		height = req.Height
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	// Insert product
	query := `
		INSERT INTO products (name, slug, description, price, stock, weight, length, width, height, category, subcategory, brand, material, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at, updated_at
	`

	var id int
	var createdAt, updatedAt sql.NullTime
	err = s.db.QueryRow(
		query,
		req.Name, slug, req.Description, req.Price, req.Stock, weight, length, width, height,
		req.Category, req.Subcategory, req.Brand, req.Material, isActive,
	).Scan(&id, &createdAt, &updatedAt)

	if err != nil {
		return nil, err
	}

	// Add images if provided
	for i, imageURL := range req.Images {
		isPrimary := i == 0
		err := s.addImage(id, imageURL, isPrimary, i)
		if err != nil {
			// Log error but don't fail the whole operation
			// Product is already created, just images failed
			fmt.Printf("Failed to add image %d for product %d: %v\n", i, id, err)
		}
	}

	// Return created product
	return s.getProductByID(id)
}

func (s *adminProductService) UpdateProduct(id int, req dto.UpdateProductRequest) (*dto.AdminProductResponse, error) {
	// Check product exists
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)", id).Scan(&exists)
	if err != nil || !exists {
		return nil, ErrProductNotFound
	}

	// Build update query dynamically
	updates := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Name != nil {
		updates = append(updates, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, *req.Name)
		argIndex++
	}
	if req.Slug != nil {
		if !s.isValidSlug(*req.Slug) {
			return nil, ErrInvalidSlug
		}
		updates = append(updates, fmt.Sprintf("slug = $%d", argIndex))
		args = append(args, *req.Slug)
		argIndex++
	}
	if req.Description != nil {
		updates = append(updates, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *req.Description)
		argIndex++
	}
	if req.Price != nil {
		updates = append(updates, fmt.Sprintf("price = $%d", argIndex))
		args = append(args, *req.Price)
		argIndex++
	}
	if req.Stock != nil {
		updates = append(updates, fmt.Sprintf("stock = $%d", argIndex))
		args = append(args, *req.Stock)
		argIndex++
	}
	if req.Weight != nil {
		updates = append(updates, fmt.Sprintf("weight = $%d", argIndex))
		args = append(args, *req.Weight)
		argIndex++
	}
	if req.Length != nil {
		updates = append(updates, fmt.Sprintf("length = $%d", argIndex))
		args = append(args, *req.Length)
		argIndex++
	}
	if req.Width != nil {
		updates = append(updates, fmt.Sprintf("width = $%d", argIndex))
		args = append(args, *req.Width)
		argIndex++
	}
	if req.Height != nil {
		updates = append(updates, fmt.Sprintf("height = $%d", argIndex))
		args = append(args, *req.Height)
		argIndex++
	}
	if req.Category != nil {
		updates = append(updates, fmt.Sprintf("category = $%d", argIndex))
		args = append(args, *req.Category)
		argIndex++
	}
	if req.Subcategory != nil {
		updates = append(updates, fmt.Sprintf("subcategory = $%d", argIndex))
		args = append(args, *req.Subcategory)
		argIndex++
	}
	if req.Brand != nil {
		updates = append(updates, fmt.Sprintf("brand = $%d", argIndex))
		args = append(args, *req.Brand)
		argIndex++
	}
	if req.Material != nil {
		updates = append(updates, fmt.Sprintf("material = $%d", argIndex))
		args = append(args, *req.Material)
		argIndex++
	}
	if req.IsActive != nil {
		updates = append(updates, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *req.IsActive)
		argIndex++
	}

	if len(updates) == 0 {
		return s.getProductByID(id)
	}

	updates = append(updates, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE products SET %s WHERE id = $%d", strings.Join(updates, ", "), argIndex)
	_, err = s.db.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	return s.getProductByID(id)
}

func (s *adminProductService) UpdateStock(id int, req dto.UpdateStockRequest) (*dto.AdminProductResponse, error) {
	// Check product exists and get current stock
	var currentStock int
	err := s.db.QueryRow("SELECT stock FROM products WHERE id = $1", id).Scan(&currentStock)
	if err != nil {
		return nil, ErrProductNotFound
	}

	newStock := currentStock + req.Quantity
	if newStock < 0 {
		return nil, errors.New("stock cannot be negative")
	}

	// Update stock
	_, err = s.db.Exec("UPDATE products SET stock = $1, updated_at = NOW() WHERE id = $2", newStock, id)
	if err != nil {
		return nil, err
	}

	// Log stock change (optional - could add stock_history table)
	// For now, just return updated product

	return s.getProductByID(id)
}

func (s *adminProductService) DeleteProduct(id int) error {
	// Hard delete - permanently remove from database
	// Note: This will cascade delete related records (images, variants, etc.)
	// based on foreign key constraints
	
	// Start transaction for safe deletion
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete product images first
	_, err = tx.Exec("DELETE FROM product_images WHERE product_id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete product images: %w", err)
	}

	// Delete product variants
	_, err = tx.Exec("DELETE FROM product_variants WHERE product_id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete product variants: %w", err)
	}

	// Delete the product itself
	result, err := tx.Exec("DELETE FROM products WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrProductNotFound
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *adminProductService) AddProductImage(productID int, req dto.AddProductImageRequest) (*dto.ProductImageResponse, error) {
	// Check product exists
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)", productID).Scan(&exists)
	if err != nil || !exists {
		return nil, ErrProductNotFound
	}

	// If setting as primary, unset other primaries
	if req.IsPrimary {
		s.db.Exec("UPDATE product_images SET is_primary = false WHERE product_id = $1", productID)
	}

	// Insert image
	query := `
		INSERT INTO product_images (product_id, image_url, is_primary, display_order)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var imageID int
	err = s.db.QueryRow(query, productID, req.ImageURL, req.IsPrimary, req.DisplayOrder).Scan(&imageID)
	if err != nil {
		return nil, err
	}

	return &dto.ProductImageResponse{
		ID:           imageID,
		ImageURL:     req.ImageURL,
		IsPrimary:    req.IsPrimary,
		DisplayOrder: req.DisplayOrder,
	}, nil
}

func (s *adminProductService) DeleteProductImage(imageID int) error {
	result, err := s.db.Exec("DELETE FROM product_images WHERE id = $1", imageID)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("image not found")
	}

	return nil
}

// Helper methods

func (s *adminProductService) getProductByID(id int) (*dto.AdminProductResponse, error) {
	query := `
		SELECT id, name, slug, description, price, stock, 
		       COALESCE(weight, 500) as weight,
		       COALESCE(length, 30) as length,
		       COALESCE(width, 20) as width,
		       COALESCE(height, 5) as height,
		       COALESCE(category, 'wanita') as category, 
		       COALESCE(subcategory, '') as subcategory,
		       COALESCE(brand, '') as brand,
		       COALESCE(material, '') as material,
		       is_active, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	var p dto.AdminProductResponse
	var createdAt, updatedAt sql.NullTime

	err := s.db.QueryRow(query, id).Scan(
		&p.ID, &p.Name, &p.Slug, &p.Description, &p.Price, &p.Stock,
		&p.Weight, &p.Length, &p.Width, &p.Height, &p.Category, &p.Subcategory,
		&p.Brand, &p.Material, &p.IsActive,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, ErrProductNotFound
	}

	if createdAt.Valid {
		p.CreatedAt = dto.FormatTime(createdAt.Time)
	}
	if updatedAt.Valid {
		p.UpdatedAt = dto.FormatTime(updatedAt.Time)
	}

	p.Images = s.getProductImages(id)

	return &p, nil
}

func (s *adminProductService) getProductImages(productID int) []dto.ProductImageResponse {
	query := `
		SELECT id, image_url, is_primary, display_order
		FROM product_images
		WHERE product_id = $1
		ORDER BY display_order ASC
	`

	rows, err := s.db.Query(query, productID)
	if err != nil {
		return []dto.ProductImageResponse{}
	}
	defer rows.Close()

	var images []dto.ProductImageResponse
	for rows.Next() {
		var img dto.ProductImageResponse
		err := rows.Scan(&img.ID, &img.ImageURL, &img.IsPrimary, &img.DisplayOrder)
		if err != nil {
			continue
		}
		images = append(images, img)
	}

	return images
}

func (s *adminProductService) addImage(productID int, imageURL string, isPrimary bool, displayOrder int) error {
	query := `
		INSERT INTO product_images (product_id, image_url, is_primary, display_order)
		VALUES ($1, $2, $3, $4)
	`
	_, err := s.db.Exec(query, productID, imageURL, isPrimary, displayOrder)
	return err
}

func (s *adminProductService) generateSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)
	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove special characters
	reg := regexp.MustCompile("[^a-z0-9-]")
	slug = reg.ReplaceAllString(slug, "")
	// Remove multiple hyphens
	reg = regexp.MustCompile("-+")
	slug = reg.ReplaceAllString(slug, "-")
	// Trim hyphens from ends
	slug = strings.Trim(slug, "-")
	return slug
}

func (s *adminProductService) isValidSlug(slug string) bool {
	if slug == "" {
		return false
	}
	matched, _ := regexp.MatchString("^[a-z0-9-]+$", slug)
	return matched
}
