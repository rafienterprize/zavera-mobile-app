package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"zavera/models"
)

type CartRepository interface {
	FindOrCreateBySessionID(sessionID string) (*models.Cart, error)
	FindByUserID(userID int) (*models.Cart, error)
	FindByID(id int) (*models.Cart, error)
	FindItemsByCartID(cartID int) ([]models.CartItem, error)
	AddItem(item *models.CartItem) error
	UpdateItem(item *models.CartItem) error
	DeleteItem(itemID int) error
	ClearCart(cartID int) error
	LinkCartToUser(cartID int, userID int) error
	MergeGuestCartToUser(guestCartID int, userCartID int) error
}

type cartRepository struct {
	db *sql.DB
}

func NewCartRepository(db *sql.DB) CartRepository {
	return &cartRepository{db: db}
}

func (r *cartRepository) FindOrCreateBySessionID(sessionID string) (*models.Cart, error) {
	// Try to find existing cart
	var cart models.Cart
	query := `
		SELECT id, user_id, session_id, created_at, updated_at
		FROM carts
		WHERE session_id = $1
	`
	
	err := r.db.QueryRow(query, sessionID).Scan(
		&cart.ID, &cart.UserID, &cart.SessionID, &cart.CreatedAt, &cart.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// Create new cart
		insertQuery := `
			INSERT INTO carts (session_id)
			VALUES ($1)
			RETURNING id, user_id, session_id, created_at, updated_at
		`
		err = r.db.QueryRow(insertQuery, sessionID).Scan(
			&cart.ID, &cart.UserID, &cart.SessionID, &cart.CreatedAt, &cart.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Load items
	items, _ := r.FindItemsByCartID(cart.ID)
	cart.Items = items

	return &cart, nil
}

func (r *cartRepository) FindByID(id int) (*models.Cart, error) {
	var cart models.Cart
	query := `
		SELECT id, user_id, session_id, created_at, updated_at
		FROM carts
		WHERE id = $1
	`
	
	err := r.db.QueryRow(query, id).Scan(
		&cart.ID, &cart.UserID, &cart.SessionID, &cart.CreatedAt, &cart.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Load items
	items, _ := r.FindItemsByCartID(cart.ID)
	cart.Items = items

	return &cart, nil
}

func (r *cartRepository) FindItemsByCartID(cartID int) ([]models.CartItem, error) {
	query := `
		SELECT ci.id, ci.cart_id, ci.product_id, ci.variant_id, ci.quantity, 
		       ci.price_snapshot, ci.metadata, ci.created_at, ci.updated_at
		FROM cart_items ci
		WHERE ci.cart_id = $1
		ORDER BY ci.created_at DESC
	`

	rows, err := r.db.Query(query, cartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.CartItem
	for rows.Next() {
		var item models.CartItem
		var metadataJSON []byte

		err := rows.Scan(
			&item.ID, &item.CartID, &item.ProductID, &item.VariantID, &item.Quantity,
			&item.PriceSnapshot, &metadataJSON, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			continue
		}

		// Parse metadata JSON
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &item.Metadata)
		}

		items = append(items, item)
	}

	return items, nil
}

func (r *cartRepository) AddItem(item *models.CartItem) error {
	// Check if item already exists with same product_id AND same metadata (size/color)
	var existingID int
	var existingQuantity int
	var existingMetadata []byte
	
	// Convert item metadata to JSON for comparison
	itemMetadataJSON, _ := json.Marshal(item.Metadata)
	
	log.Printf("🛒 AddItem: cart_id=%d, product_id=%d, variant_id=%v, quantity=%d, metadata=%s", 
		item.CartID, item.ProductID, item.VariantID, item.Quantity, string(itemMetadataJSON))
	
	checkQuery := `
		SELECT id, quantity, metadata FROM cart_items
		WHERE cart_id = $1 AND product_id = $2
	`
	rows, err := r.db.Query(checkQuery, item.CartID, item.ProductID)
	if err != nil {
		// No existing items, proceed to insert
		log.Printf("🛒 AddItem: No existing items found, inserting new")
		goto insertNew
	}
	defer rows.Close()

	// Check each existing item to see if metadata matches
	for rows.Next() {
		err = rows.Scan(&existingID, &existingQuantity, &existingMetadata)
		if err != nil {
			continue
		}
		
		log.Printf("🛒 AddItem: Comparing with existing item_id=%d, metadata=%s", existingID, string(existingMetadata))
		
		// Compare metadata (size, color, etc.)
		if string(existingMetadata) == string(itemMetadataJSON) {
			// Found exact match (same product + same variant)
			// Update quantity by SETTING to new value (frontend sends total)
			log.Printf("🛒 AddItem: Metadata match! Updating item_id=%d, quantity=%d", existingID, item.Quantity)
			updateQuery := `
				UPDATE cart_items
				SET quantity = $1, updated_at = NOW()
				WHERE id = $2
				RETURNING id
			`
			return r.db.QueryRow(updateQuery, item.Quantity, existingID).Scan(&item.ID)
		}
	}

insertNew:
	// No matching item found, insert as new item
	log.Printf("🛒 AddItem: No metadata match, inserting new item")
	metadataJSON, _ := json.Marshal(item.Metadata)
	insertQuery := `
		INSERT INTO cart_items (cart_id, product_id, variant_id, quantity, price_snapshot, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	
	return r.db.QueryRow(
		insertQuery,
		item.CartID, item.ProductID, item.VariantID, item.Quantity, item.PriceSnapshot, metadataJSON,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
}

func (r *cartRepository) UpdateItem(item *models.CartItem) error {
	query := `
		UPDATE cart_items
		SET quantity = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Exec(query, item.Quantity, item.ID)
	return err
}

func (r *cartRepository) DeleteItem(itemID int) error {
	query := `DELETE FROM cart_items WHERE id = $1`
	result, err := r.db.Exec(query, itemID)
	if err != nil {
		return err
	}
	
	// Check how many rows were affected
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("cart item not found or already deleted")
	}
	
	return nil
}

func (r *cartRepository) ClearCart(cartID int) error {
	query := `DELETE FROM cart_items WHERE cart_id = $1`
	_, err := r.db.Exec(query, cartID)
	return err
}

// FindByUserID finds cart by user ID (for logged-in users)
func (r *cartRepository) FindByUserID(userID int) (*models.Cart, error) {
	var cart models.Cart
	query := `
		SELECT id, user_id, session_id, created_at, updated_at
		FROM carts
		WHERE user_id = $1
		ORDER BY updated_at DESC
		LIMIT 1
	`

	err := r.db.QueryRow(query, userID).Scan(
		&cart.ID, &cart.UserID, &cart.SessionID, &cart.CreatedAt, &cart.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Load items
	items, _ := r.FindItemsByCartID(cart.ID)
	cart.Items = items

	return &cart, nil
}

// LinkCartToUser links a guest cart to a user account
func (r *cartRepository) LinkCartToUser(cartID int, userID int) error {
	query := `
		UPDATE carts
		SET user_id = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Exec(query, userID, cartID)
	return err
}

// MergeGuestCartToUser merges guest cart items into user's existing cart
func (r *cartRepository) MergeGuestCartToUser(guestCartID int, userCartID int) error {
	// Get guest cart items
	guestItems, err := r.FindItemsByCartID(guestCartID)
	if err != nil {
		return err
	}

	// Get user cart items for comparison
	userItems, err := r.FindItemsByCartID(userCartID)
	if err != nil {
		return err
	}

	// Create map of existing user cart items by product_id
	userItemMap := make(map[int]*models.CartItem)
	for i := range userItems {
		userItemMap[userItems[i].ProductID] = &userItems[i]
	}

	// Merge guest items into user cart
	for _, guestItem := range guestItems {
		if existingItem, exists := userItemMap[guestItem.ProductID]; exists {
			// Update quantity if product already exists in user cart
			newQuantity := existingItem.Quantity + guestItem.Quantity
			updateQuery := `
				UPDATE cart_items
				SET quantity = $1, updated_at = NOW()
				WHERE id = $2
			`
			r.db.Exec(updateQuery, newQuantity, existingItem.ID)
		} else {
			// Move item to user cart
			moveQuery := `
				UPDATE cart_items
				SET cart_id = $1, updated_at = NOW()
				WHERE id = $2
			`
			r.db.Exec(moveQuery, userCartID, guestItem.ID)
		}
	}

	// Delete guest cart (items already moved/merged)
	deleteItemsQuery := `DELETE FROM cart_items WHERE cart_id = $1`
	r.db.Exec(deleteItemsQuery, guestCartID)

	deleteCartQuery := `DELETE FROM carts WHERE id = $1`
	r.db.Exec(deleteCartQuery, guestCartID)

	return nil
}
