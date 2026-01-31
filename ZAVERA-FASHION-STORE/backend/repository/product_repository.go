package repository

import (
	"database/sql"
	"zavera/models"
)

type ProductRepository interface {
	FindAll() ([]models.Product, error)
	FindByCategory(category string) ([]models.Product, error)
	FindByID(id int) (*models.Product, error)
	FindBySlug(slug string) (*models.Product, error)
	UpdateStock(productID int, quantity int) error
}

type productRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) FindAll() ([]models.Product, error) {
	query := `
		SELECT p.id, p.name, p.slug, p.description, p.price, p.stock, 
		       COALESCE(p.weight, 500) as weight,
		       COALESCE(p.length, 30) as length,
		       COALESCE(p.width, 20) as width,
		       COALESCE(p.height, 5) as height,
		       p.is_active, COALESCE(p.category, 'wanita') as category, 
		       COALESCE(p.subcategory, '') as subcategory, p.created_at, p.updated_at
		FROM products p
		WHERE p.is_active = true
		ORDER BY p.created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		err := rows.Scan(
			&p.ID, &p.Name, &p.Slug, &p.Description, &p.Price,
			&p.Stock, &p.Weight, &p.Length, &p.Width, &p.Height, &p.IsActive, &p.Category, &p.Subcategory, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Load images for product
		images, _ := r.findImagesByProductID(p.ID)
		p.Images = images

		products = append(products, p)
	}

	return products, nil
}

func (r *productRepository) FindByCategory(category string) ([]models.Product, error) {
	query := `
		SELECT p.id, p.name, p.slug, p.description, p.price, p.stock, 
		       COALESCE(p.weight, 500) as weight,
		       COALESCE(p.length, 30) as length,
		       COALESCE(p.width, 20) as width,
		       COALESCE(p.height, 5) as height,
		       p.is_active, COALESCE(p.category, 'wanita') as category, 
		       COALESCE(p.subcategory, '') as subcategory, p.created_at, p.updated_at
		FROM products p
		WHERE p.is_active = true AND LOWER(p.category) = LOWER($1)
		ORDER BY p.created_at DESC
	`

	rows, err := r.db.Query(query, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		err := rows.Scan(
			&p.ID, &p.Name, &p.Slug, &p.Description, &p.Price,
			&p.Stock, &p.Weight, &p.Length, &p.Width, &p.Height, &p.IsActive, &p.Category, &p.Subcategory, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Load images for product
		images, _ := r.findImagesByProductID(p.ID)
		p.Images = images

		products = append(products, p)
	}

	return products, nil
}

func (r *productRepository) FindByID(id int) (*models.Product, error) {
	query := `
		SELECT id, name, slug, description, price, stock, 
		       COALESCE(weight, 500) as weight,
		       COALESCE(length, 30) as length,
		       COALESCE(width, 20) as width,
		       COALESCE(height, 5) as height,
		       is_active, COALESCE(category, 'wanita') as category, 
		       COALESCE(subcategory, '') as subcategory, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	var p models.Product
	err := r.db.QueryRow(query, id).Scan(
		&p.ID, &p.Name, &p.Slug, &p.Description, &p.Price,
		&p.Stock, &p.Weight, &p.Length, &p.Width, &p.Height, &p.IsActive, &p.Category, &p.Subcategory, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Load images
	images, _ := r.findImagesByProductID(p.ID)
	p.Images = images

	return &p, nil
}

func (r *productRepository) FindBySlug(slug string) (*models.Product, error) {
	query := `
		SELECT id, name, slug, description, price, stock, 
		       COALESCE(weight, 500) as weight,
		       COALESCE(length, 30) as length,
		       COALESCE(width, 20) as width,
		       COALESCE(height, 5) as height,
		       is_active, COALESCE(category, 'wanita') as category, 
		       COALESCE(subcategory, '') as subcategory, created_at, updated_at
		FROM products
		WHERE slug = $1
	`

	var p models.Product
	err := r.db.QueryRow(query, slug).Scan(
		&p.ID, &p.Name, &p.Slug, &p.Description, &p.Price,
		&p.Stock, &p.Weight, &p.Length, &p.Width, &p.Height, &p.IsActive, &p.Category, &p.Subcategory, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Load images
	images, _ := r.findImagesByProductID(p.ID)
	p.Images = images

	return &p, nil
}

func (r *productRepository) UpdateStock(productID int, quantity int) error {
	query := `
		UPDATE products 
		SET stock = stock + $1
		WHERE id = $2
	`
	_, err := r.db.Exec(query, quantity, productID)
	return err
}

func (r *productRepository) findImagesByProductID(productID int) ([]models.ProductImage, error) {
	query := `
		SELECT id, product_id, image_url, is_primary, display_order, created_at
		FROM product_images
		WHERE product_id = $1
		ORDER BY display_order ASC
	`

	rows, err := r.db.Query(query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []models.ProductImage
	for rows.Next() {
		var img models.ProductImage
		err := rows.Scan(
			&img.ID, &img.ProductID, &img.ImageURL,
			&img.IsPrimary, &img.DisplayOrder, &img.CreatedAt,
		)
		if err != nil {
			continue
		}
		images = append(images, img)
	}

	return images, nil
}
