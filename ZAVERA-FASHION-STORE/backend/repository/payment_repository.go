package repository

import (
	"database/sql"
	"encoding/json"
	"time"
	"zavera/models"
)

type PaymentRepository interface {
	Create(payment *models.Payment) error
	FindByOrderID(orderID int) (*models.Payment, error)
	FindByExternalID(externalID string) (*models.Payment, error)
	UpdateStatus(paymentID int, status models.PaymentStatus, providerResponse map[string]any) error
	UpdateStatusWithResponse(paymentID int, status models.PaymentStatus, providerResponse map[string]any) error
	UpdateToken(paymentID int, token, redirectURL string) error
	MarkAsPaid(paymentID int, transactionID string) error
	MarkAsPaidWithDetails(paymentID int, transactionID, paymentType string, providerResponse map[string]any) error
	GetDB() *sql.DB
}

type paymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(payment *models.Payment) error {
	providerResponseJSON, _ := json.Marshal(payment.ProviderResponse)
	
	query := `
		INSERT INTO payments (
			order_id, payment_method, payment_provider, amount, status,
			external_id, transaction_id, provider_response
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	return r.db.QueryRow(
		query,
		payment.OrderID, payment.PaymentMethod, payment.PaymentProvider,
		payment.Amount, payment.Status, payment.ExternalID,
		payment.TransactionID, providerResponseJSON,
	).Scan(&payment.ID, &payment.CreatedAt, &payment.UpdatedAt)
}

func (r *paymentRepository) FindByOrderID(orderID int) (*models.Payment, error) {
	query := `
		SELECT id, order_id, payment_method, payment_provider, amount, status,
		       external_id, transaction_id, provider_response, created_at, updated_at,
		       paid_at, expired_at
		FROM payments
		WHERE order_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var payment models.Payment
	var providerResponseJSON []byte
	var externalID, transactionID sql.NullString
	var paidAt, expiredAt sql.NullTime

	err := r.db.QueryRow(query, orderID).Scan(
		&payment.ID, &payment.OrderID, &payment.PaymentMethod, &payment.PaymentProvider,
		&payment.Amount, &payment.Status, &externalID, &transactionID,
		&providerResponseJSON, &payment.CreatedAt, &payment.UpdatedAt,
		&paidAt, &expiredAt,
	)

	if err != nil {
		return nil, err
	}

	// Handle NULL values
	if externalID.Valid {
		payment.ExternalID = externalID.String
	}
	if transactionID.Valid {
		payment.TransactionID = transactionID.String
	}
	if paidAt.Valid {
		payment.PaidAt = &paidAt.Time
	}
	if expiredAt.Valid {
		payment.ExpiredAt = &expiredAt.Time
	}

	// Parse provider response
	if len(providerResponseJSON) > 0 {
		json.Unmarshal(providerResponseJSON, &payment.ProviderResponse)
	}

	return &payment, nil
}

func (r *paymentRepository) FindByExternalID(externalID string) (*models.Payment, error) {
	query := `
		SELECT id, order_id, payment_method, payment_provider, amount, status,
		       external_id, transaction_id, provider_response, created_at, updated_at,
		       paid_at, expired_at
		FROM payments
		WHERE external_id = $1
	`

	var payment models.Payment
	var providerResponseJSON []byte

	err := r.db.QueryRow(query, externalID).Scan(
		&payment.ID, &payment.OrderID, &payment.PaymentMethod, &payment.PaymentProvider,
		&payment.Amount, &payment.Status, &payment.ExternalID, &payment.TransactionID,
		&providerResponseJSON, &payment.CreatedAt, &payment.UpdatedAt,
		&payment.PaidAt, &payment.ExpiredAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse provider response
	if len(providerResponseJSON) > 0 {
		json.Unmarshal(providerResponseJSON, &payment.ProviderResponse)
	}

	return &payment, nil
}

func (r *paymentRepository) UpdateStatus(paymentID int, status models.PaymentStatus, providerResponse map[string]any) error {
	providerResponseJSON, _ := json.Marshal(providerResponse)
	
	query := `
		UPDATE payments
		SET status = $1, provider_response = $2, updated_at = NOW()
		WHERE id = $3
	`
	_, err := r.db.Exec(query, status, providerResponseJSON, paymentID)
	return err
}

func (r *paymentRepository) MarkAsPaid(paymentID int, transactionID string) error {
	query := `
		UPDATE payments
		SET status = $1, transaction_id = $2, paid_at = NOW(), updated_at = NOW()
		WHERE id = $3
	`
	_, err := r.db.Exec(query, models.PaymentStatusSuccess, transactionID, paymentID)
	return err
}

// UpdateToken updates the Snap token for an existing payment
func (r *paymentRepository) UpdateToken(paymentID int, token, redirectURL string) error {
	providerResponse := map[string]any{
		"token":        token,
		"redirect_url": redirectURL,
	}
	providerResponseJSON, _ := json.Marshal(providerResponse)

	query := `
		UPDATE payments
		SET provider_response = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Exec(query, providerResponseJSON, paymentID)
	return err
}

// UpdateStatusWithResponse updates payment status with full provider response
func (r *paymentRepository) UpdateStatusWithResponse(paymentID int, status models.PaymentStatus, providerResponse map[string]any) error {
	providerResponseJSON, _ := json.Marshal(providerResponse)

	var query string
	if status == models.PaymentStatusExpired {
		query = `
			UPDATE payments
			SET status = $1, provider_response = $2, expired_at = NOW(), updated_at = NOW()
			WHERE id = $3
		`
	} else {
		query = `
			UPDATE payments
			SET status = $1, provider_response = $2, updated_at = NOW()
			WHERE id = $3
		`
	}

	_, err := r.db.Exec(query, status, providerResponseJSON, paymentID)
	return err
}

// MarkAsPaidWithDetails marks payment as paid with full transaction details
func (r *paymentRepository) MarkAsPaidWithDetails(paymentID int, transactionID, paymentMethod string, providerResponse map[string]any) error {
	providerResponseJSON, _ := json.Marshal(providerResponse)

	query := `
		UPDATE payments
		SET status = $1, transaction_id = $2, payment_method = $3, 
		    provider_response = $4, paid_at = $5, updated_at = NOW()
		WHERE id = $6
	`
	_, err := r.db.Exec(query, models.PaymentStatusSuccess, transactionID, paymentMethod, providerResponseJSON, time.Now(), paymentID)
	return err
}


// GetDB returns the database connection for direct queries
func (r *paymentRepository) GetDB() *sql.DB {
	return r.db
}
