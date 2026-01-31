package repository

import (
	"database/sql"
	"zavera/models"
)

type UserRepository interface {
	Create(user *models.User) error
	FindByID(id int) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	FindByGoogleID(googleID string) (*models.User, error)
	Update(user *models.User) error
	MarkAsVerified(userID int) error
	
	// Email verification tokens
	CreateVerificationToken(token *models.EmailVerificationToken) error
	FindVerificationToken(token string) (*models.EmailVerificationToken, error)
	MarkTokenAsUsed(tokenID int) error
	DeleteExpiredTokens(userID int) error
	
	// User orders
	FindOrdersByUserID(userID int, page, pageSize int) ([]models.Order, int, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (email, first_name, name, phone, password_hash, birthdate, is_verified, google_id, auth_provider)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(
		query,
		user.Email, user.FirstName, user.Name, user.Phone, user.PasswordHash,
		user.Birthdate, user.IsVerified, user.GoogleID, user.AuthProvider,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *userRepository) FindByID(id int) (*models.User, error) {
	query := `
		SELECT id, email, COALESCE(first_name, '') as first_name, COALESCE(name, '') as name, 
		       COALESCE(phone, '') as phone, COALESCE(password_hash, '') as password_hash,
		       birthdate, COALESCE(is_verified, false) as is_verified, google_id, 
		       COALESCE(auth_provider, 'local') as auth_provider, created_at, updated_at
		FROM users WHERE id = $1
	`
	var user models.User
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Email, &user.FirstName, &user.Name, &user.Phone,
		&user.PasswordHash, &user.Birthdate, &user.IsVerified, &user.GoogleID,
		&user.AuthProvider, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, email, COALESCE(first_name, '') as first_name, COALESCE(name, '') as name, 
		       COALESCE(phone, '') as phone, COALESCE(password_hash, '') as password_hash,
		       birthdate, COALESCE(is_verified, false) as is_verified, google_id, 
		       COALESCE(auth_provider, 'local') as auth_provider, created_at, updated_at
		FROM users WHERE email = $1
	`
	var user models.User
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Email, &user.FirstName, &user.Name, &user.Phone,
		&user.PasswordHash, &user.Birthdate, &user.IsVerified, &user.GoogleID,
		&user.AuthProvider, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByGoogleID(googleID string) (*models.User, error) {
	query := `
		SELECT id, email, COALESCE(first_name, '') as first_name, COALESCE(name, '') as name, 
		       COALESCE(phone, '') as phone, COALESCE(password_hash, '') as password_hash,
		       birthdate, COALESCE(is_verified, false) as is_verified, google_id, 
		       COALESCE(auth_provider, 'local') as auth_provider, created_at, updated_at
		FROM users WHERE google_id = $1
	`
	var user models.User
	err := r.db.QueryRow(query, googleID).Scan(
		&user.ID, &user.Email, &user.FirstName, &user.Name, &user.Phone,
		&user.PasswordHash, &user.Birthdate, &user.IsVerified, &user.GoogleID,
		&user.AuthProvider, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(user *models.User) error {
	query := `
		UPDATE users SET 
			first_name = $1, name = $2, phone = $3, birthdate = $4, 
			is_verified = $5, updated_at = NOW()
		WHERE id = $6
	`
	_, err := r.db.Exec(query, user.FirstName, user.Name, user.Phone, user.Birthdate, user.IsVerified, user.ID)
	return err
}

func (r *userRepository) MarkAsVerified(userID int) error {
	query := `UPDATE users SET is_verified = true, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(query, userID)
	return err
}


// Email verification token methods
func (r *userRepository) CreateVerificationToken(token *models.EmailVerificationToken) error {
	query := `
		INSERT INTO email_verification_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	return r.db.QueryRow(query, token.UserID, token.Token, token.ExpiresAt).Scan(&token.ID, &token.CreatedAt)
}

func (r *userRepository) FindVerificationToken(token string) (*models.EmailVerificationToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, used_at, created_at
		FROM email_verification_tokens
		WHERE token = $1
	`
	var t models.EmailVerificationToken
	err := r.db.QueryRow(query, token).Scan(
		&t.ID, &t.UserID, &t.Token, &t.ExpiresAt, &t.UsedAt, &t.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *userRepository) MarkTokenAsUsed(tokenID int) error {
	query := `UPDATE email_verification_tokens SET used_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(query, tokenID)
	return err
}

func (r *userRepository) DeleteExpiredTokens(userID int) error {
	query := `DELETE FROM email_verification_tokens WHERE user_id = $1 AND (expires_at < NOW() OR used_at IS NOT NULL)`
	_, err := r.db.Exec(query, userID)
	return err
}

// User orders
func (r *userRepository) FindOrdersByUserID(userID int, page, pageSize int) ([]models.Order, int, error) {
	// Count total
	var totalCount int
	countQuery := `SELECT COUNT(*) FROM orders WHERE user_id = $1`
	err := r.db.QueryRow(countQuery, userID).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Get orders with pagination
	offset := (page - 1) * pageSize
	query := `
		SELECT id, order_code, user_id, customer_name, customer_email, customer_phone,
		       subtotal, shipping_cost, tax, discount, total_amount, status,
		       COALESCE(stock_reserved, true) as stock_reserved, notes,
		       created_at, updated_at, paid_at, shipped_at, completed_at, cancelled_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(
			&order.ID, &order.OrderCode, &order.UserID, &order.CustomerName,
			&order.CustomerEmail, &order.CustomerPhone, &order.Subtotal, &order.ShippingCost,
			&order.Tax, &order.Discount, &order.TotalAmount, &order.Status, &order.StockReserved,
			&order.Notes, &order.CreatedAt, &order.UpdatedAt,
			&order.PaidAt, &order.ShippedAt, &order.CompletedAt, &order.CancelledAt,
		)
		if err != nil {
			continue
		}

		// Load items for each order
		items, _ := r.findItemsByOrderID(order.ID)
		order.Items = items

		orders = append(orders, order)
	}

	return orders, totalCount, nil
}

func (r *userRepository) findItemsByOrderID(orderID int) ([]models.OrderItem, error) {
	query := `
		SELECT oi.id, oi.order_id, oi.product_id, oi.product_name, oi.quantity,
		       oi.price_per_unit, oi.subtotal, oi.created_at,
		       COALESCE(
		           (SELECT image_url FROM product_images WHERE product_id = oi.product_id ORDER BY is_primary DESC, display_order ASC LIMIT 1),
		           ''
		       ) as product_image
		FROM order_items oi
		WHERE oi.order_id = $1
	`

	rows, err := r.db.Query(query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		err := rows.Scan(
			&item.ID, &item.OrderID, &item.ProductID, &item.ProductName,
			&item.Quantity, &item.PricePerUnit, &item.Subtotal, &item.CreatedAt,
			&item.ProductImage,
		)
		if err != nil {
			continue
		}
		items = append(items, item)
	}

	return items, nil
}
