package repository

import (
	"database/sql"
	"errors"
	"zavera/models"
)

type WishlistRepository interface {
	FindByUserID(userID int) ([]models.Wishlist, error)
	FindByUserAndProduct(userID int, productID int) (*models.Wishlist, error)
	Add(userID int, productID int) (*models.Wishlist, error)
	Remove(userID int, productID int) error
	RemoveByID(id int, userID int) error
	Count(userID int) (int, error)
	IsInWishlist(userID int, productID int) (bool, error)
}

type wishlistRepository struct {
	db *sql.DB
}

func NewWishlistRepository(db *sql.DB) WishlistRepository {
	return &wishlistRepository{db: db}
}

// FindByUserID gets all wishlist items for a user
func (r *wishlistRepository) FindByUserID(userID int) ([]models.Wishlist, error) {
	query := `
		SELECT id, user_id, product_id, created_at, updated_at
		FROM wishlists
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Wishlist
	for rows.Next() {
		var item models.Wishlist
		err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.ProductID,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			continue
		}
		items = append(items, item)
	}

	return items, nil
}

// FindByUserAndProduct finds a specific wishlist item
func (r *wishlistRepository) FindByUserAndProduct(userID int, productID int) (*models.Wishlist, error) {
	query := `
		SELECT id, user_id, product_id, created_at, updated_at
		FROM wishlists
		WHERE user_id = $1 AND product_id = $2
	`

	var item models.Wishlist
	err := r.db.QueryRow(query, userID, productID).Scan(
		&item.ID,
		&item.UserID,
		&item.ProductID,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &item, nil
}

// Add adds a product to wishlist
func (r *wishlistRepository) Add(userID int, productID int) (*models.Wishlist, error) {
	// Check if already exists
	existing, err := r.FindByUserAndProduct(userID, productID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil // Already in wishlist
	}

	// Insert new wishlist item
	query := `
		INSERT INTO wishlists (user_id, product_id)
		VALUES ($1, $2)
		RETURNING id, user_id, product_id, created_at, updated_at
	`

	var item models.Wishlist
	err = r.db.QueryRow(query, userID, productID).Scan(
		&item.ID,
		&item.UserID,
		&item.ProductID,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &item, nil
}

// Remove removes a product from wishlist by product ID
func (r *wishlistRepository) Remove(userID int, productID int) error {
	query := `DELETE FROM wishlists WHERE user_id = $1 AND product_id = $2`
	result, err := r.db.Exec(query, userID, productID)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("wishlist item not found")
	}

	return nil
}

// RemoveByID removes a wishlist item by its ID
func (r *wishlistRepository) RemoveByID(id int, userID int) error {
	query := `DELETE FROM wishlists WHERE id = $1 AND user_id = $2`
	result, err := r.db.Exec(query, id, userID)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("wishlist item not found or unauthorized")
	}

	return nil
}

// Count returns the number of items in user's wishlist
func (r *wishlistRepository) Count(userID int) (int, error) {
	query := `SELECT COUNT(*) FROM wishlists WHERE user_id = $1`
	var count int
	err := r.db.QueryRow(query, userID).Scan(&count)
	return count, err
}

// IsInWishlist checks if a product is in user's wishlist
func (r *wishlistRepository) IsInWishlist(userID int, productID int) (bool, error) {
	item, err := r.FindByUserAndProduct(userID, productID)
	if err != nil {
		return false, err
	}
	return item != nil, nil
}
