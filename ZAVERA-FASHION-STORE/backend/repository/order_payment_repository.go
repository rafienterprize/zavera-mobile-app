package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"
	"zavera/models"
)

var (
	ErrPaymentAlreadyExists = errors.New("active payment already exists for this order")
	ErrPaymentNotFound      = errors.New("payment not found")
	ErrPaymentExpired       = errors.New("payment has expired")
	ErrInvalidTransition    = errors.New("invalid payment status transition")
)

// OrderPaymentRepository defines the interface for order payment data access
type OrderPaymentRepository interface {
	// Create creates a new order payment record with row locking on order
	// Returns ErrPaymentAlreadyExists if a PENDING payment already exists
	Create(payment *models.OrderPayment) error

	// FindByOrderID finds the payment for an order (any status)
	FindByOrderID(orderID int) (*models.OrderPayment, error)

	// FindPendingByOrderID finds the PENDING payment for an order
	// Returns nil if no pending payment exists
	FindPendingByOrderID(orderID int) (*models.OrderPayment, error)

	// FindByMidtransOrderID finds payment by Midtrans order ID
	FindByMidtransOrderID(midtransOrderID string) (*models.OrderPayment, error)

	// UpdateStatus updates the payment status with optimistic locking
	UpdateStatus(paymentID int, currentStatus, newStatus models.CorePaymentStatus) error

	// UpdateToPaid marks payment as paid with timestamp
	UpdateToPaid(paymentID int, transactionID string) error

	// UpdateToExpired marks payment as expired
	UpdateToExpired(paymentID int) error

	// GetBankInstructions returns payment instructions for a bank
	GetBankInstructions(bank string) ([]models.PaymentInstructionGroup, error)

	// LogSync creates a sync log entry for audit
	LogSync(log *models.CorePaymentSyncLog) error

	// GetDB returns the underlying database connection
	GetDB() *sql.DB

	// GetPendingPaymentsForUser returns all pending orders with payment info for a user
	GetPendingPaymentsForUser(userID int, page, pageSize int) ([]map[string]interface{}, int, error)

	// GetTransactionHistoryForUser returns completed/cancelled orders for a user
	GetTransactionHistoryForUser(userID int, page, pageSize int) ([]map[string]interface{}, int, error)

	// GetTransactionHistoryForUserWithFilter returns filtered transaction history (Tokopedia-style)
	GetTransactionHistoryForUserWithFilter(userID int, filter string, page, pageSize int) ([]map[string]interface{}, int, error)
}

type orderPaymentRepository struct {
	db *sql.DB
}

// NewOrderPaymentRepository creates a new OrderPaymentRepository
func NewOrderPaymentRepository(db *sql.DB) OrderPaymentRepository {
	return &orderPaymentRepository{db: db}
}

func (r *orderPaymentRepository) GetDB() *sql.DB {
	return r.db
}

// Create creates a new order payment with row locking to prevent duplicates
func (r *orderPaymentRepository) Create(payment *models.OrderPayment) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Lock the order row to prevent race conditions
	var orderID int
	err = tx.QueryRow(`SELECT id FROM orders WHERE id = $1 FOR UPDATE`, payment.OrderID).Scan(&orderID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("order not found: %d", payment.OrderID)
		}
		return fmt.Errorf("failed to lock order: %w", err)
	}

	// Check if a PENDING payment already exists (enforced by partial unique index too)
	var existingID int
	err = tx.QueryRow(`
		SELECT id FROM order_payments 
		WHERE order_id = $1 AND payment_status = 'PENDING'
	`, payment.OrderID).Scan(&existingID)
	if err == nil {
		return ErrPaymentAlreadyExists
	}
	if err != sql.ErrNoRows {
		return fmt.Errorf("failed to check existing payment: %w", err)
	}

	// Serialize raw response to JSON
	rawResponseJSON, err := json.Marshal(payment.RawResponse)
	if err != nil {
		return fmt.Errorf("failed to marshal raw response: %w", err)
	}

	// Insert the payment record
	err = tx.QueryRow(`
		INSERT INTO order_payments (
			order_id, payment_method, bank, va_number, transaction_id,
			midtrans_order_id, expiry_time, payment_status, raw_response,
			qr_code_url, deeplink_url
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`,
		payment.OrderID,
		payment.PaymentMethod,
		payment.Bank,
		payment.VANumber,
		payment.TransactionID,
		payment.MidtransOrderID,
		payment.ExpiryTime,
		payment.PaymentStatus,
		rawResponseJSON,
		payment.QRCodeURL,
		payment.DeeplinkURL,
	).Scan(&payment.ID, &payment.CreatedAt, &payment.UpdatedAt)

	if err != nil {
		// Check for unique constraint violation (partial unique index)
		if isUniqueViolation(err) {
			return ErrPaymentAlreadyExists
		}
		return fmt.Errorf("failed to insert payment: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("✅ Created order payment: id=%d, order_id=%d, method=%s, va=%s",
		payment.ID, payment.OrderID, payment.PaymentMethod, payment.VANumber)

	return nil
}

// FindByOrderID finds the most recent payment for an order
func (r *orderPaymentRepository) FindByOrderID(orderID int) (*models.OrderPayment, error) {
	payment := &models.OrderPayment{}
	var rawResponseJSON []byte
	var qrCodeURL, deeplinkURL sql.NullString

	err := r.db.QueryRow(`
		SELECT id, order_id, payment_method, bank, va_number, transaction_id,
			   midtrans_order_id, expiry_time, payment_status, raw_response,
			   created_at, updated_at, paid_at, qr_code_url, deeplink_url
		FROM order_payments
		WHERE order_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, orderID).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.PaymentMethod,
		&payment.Bank,
		&payment.VANumber,
		&payment.TransactionID,
		&payment.MidtransOrderID,
		&payment.ExpiryTime,
		&payment.PaymentStatus,
		&rawResponseJSON,
		&payment.CreatedAt,
		&payment.UpdatedAt,
		&payment.PaidAt,
		&qrCodeURL,
		&deeplinkURL,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPaymentNotFound
		}
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}

	// Parse raw response JSON
	if len(rawResponseJSON) > 0 {
		if err := json.Unmarshal(rawResponseJSON, &payment.RawResponse); err != nil {
			log.Printf("⚠️ Failed to unmarshal raw response: %v", err)
		}
	}

	// Set GoPay fields
	if qrCodeURL.Valid {
		payment.QRCodeURL = qrCodeURL.String
	}
	if deeplinkURL.Valid {
		payment.DeeplinkURL = deeplinkURL.String
	}

	return payment, nil
}

// FindPendingByOrderID finds the PENDING payment for an order
func (r *orderPaymentRepository) FindPendingByOrderID(orderID int) (*models.OrderPayment, error) {
	payment := &models.OrderPayment{}
	var rawResponseJSON []byte
	var qrCodeURL, deeplinkURL sql.NullString

	err := r.db.QueryRow(`
		SELECT id, order_id, payment_method, bank, va_number, transaction_id,
			   midtrans_order_id, expiry_time, payment_status, raw_response,
			   created_at, updated_at, paid_at, qr_code_url, deeplink_url
		FROM order_payments
		WHERE order_id = $1 AND payment_status = 'PENDING'
		LIMIT 1
	`, orderID).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.PaymentMethod,
		&payment.Bank,
		&payment.VANumber,
		&payment.TransactionID,
		&payment.MidtransOrderID,
		&payment.ExpiryTime,
		&payment.PaymentStatus,
		&rawResponseJSON,
		&payment.CreatedAt,
		&payment.UpdatedAt,
		&payment.PaidAt,
		&qrCodeURL,
		&deeplinkURL,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No pending payment is not an error
		}
		return nil, fmt.Errorf("failed to find pending payment: %w", err)
	}

	// Parse raw response JSON
	if len(rawResponseJSON) > 0 {
		if err := json.Unmarshal(rawResponseJSON, &payment.RawResponse); err != nil {
			log.Printf("⚠️ Failed to unmarshal raw response: %v", err)
		}
	}

	// Set GoPay fields
	if qrCodeURL.Valid {
		payment.QRCodeURL = qrCodeURL.String
	}
	if deeplinkURL.Valid {
		payment.DeeplinkURL = deeplinkURL.String
	}

	return payment, nil
}

// FindByMidtransOrderID finds payment by Midtrans order ID
func (r *orderPaymentRepository) FindByMidtransOrderID(midtransOrderID string) (*models.OrderPayment, error) {
	payment := &models.OrderPayment{}
	var rawResponseJSON []byte

	err := r.db.QueryRow(`
		SELECT id, order_id, payment_method, bank, va_number, transaction_id,
			   midtrans_order_id, expiry_time, payment_status, raw_response,
			   created_at, updated_at, paid_at
		FROM order_payments
		WHERE midtrans_order_id = $1
		LIMIT 1
	`, midtransOrderID).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.PaymentMethod,
		&payment.Bank,
		&payment.VANumber,
		&payment.TransactionID,
		&payment.MidtransOrderID,
		&payment.ExpiryTime,
		&payment.PaymentStatus,
		&rawResponseJSON,
		&payment.CreatedAt,
		&payment.UpdatedAt,
		&payment.PaidAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPaymentNotFound
		}
		return nil, fmt.Errorf("failed to find payment by midtrans order id: %w", err)
	}

	// Parse raw response JSON
	if len(rawResponseJSON) > 0 {
		if err := json.Unmarshal(rawResponseJSON, &payment.RawResponse); err != nil {
			log.Printf("⚠️ Failed to unmarshal raw response: %v", err)
		}
	}

	return payment, nil
}

// UpdateStatus updates payment status with optimistic locking
func (r *orderPaymentRepository) UpdateStatus(paymentID int, currentStatus, newStatus models.CorePaymentStatus) error {
	result, err := r.db.Exec(`
		UPDATE order_payments
		SET payment_status = $1, updated_at = NOW()
		WHERE id = $2 AND payment_status = $3
	`, newStatus, paymentID, currentStatus)

	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrInvalidTransition
	}

	log.Printf("✅ Updated payment %d status: %s → %s", paymentID, currentStatus, newStatus)
	return nil
}

// UpdateToPaid marks payment as paid with timestamp
func (r *orderPaymentRepository) UpdateToPaid(paymentID int, transactionID string) error {
	result, err := r.db.Exec(`
		UPDATE order_payments
		SET payment_status = 'PAID', 
			transaction_id = COALESCE(NULLIF($1, ''), transaction_id),
			paid_at = NOW(), 
			updated_at = NOW()
		WHERE id = $2 AND payment_status = 'PENDING'
	`, transactionID, paymentID)

	if err != nil {
		return fmt.Errorf("failed to update payment to paid: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrInvalidTransition
	}

	log.Printf("✅ Payment %d marked as PAID, transaction_id=%s", paymentID, transactionID)
	return nil
}

// UpdateToExpired marks payment as expired
func (r *orderPaymentRepository) UpdateToExpired(paymentID int) error {
	result, err := r.db.Exec(`
		UPDATE order_payments
		SET payment_status = 'EXPIRED', updated_at = NOW()
		WHERE id = $1 AND payment_status = 'PENDING'
	`, paymentID)

	if err != nil {
		return fmt.Errorf("failed to update payment to expired: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrInvalidTransition
	}

	log.Printf("✅ Payment %d marked as EXPIRED", paymentID)
	return nil
}

// GetBankInstructions returns payment instructions grouped by channel
func (r *orderPaymentRepository) GetBankInstructions(bank string) ([]models.PaymentInstructionGroup, error) {
	rows, err := r.db.Query(`
		SELECT channel, instruction
		FROM bank_payment_instructions
		WHERE bank = $1 AND is_active = true
		ORDER BY channel, step_order
	`, bank)
	if err != nil {
		return nil, fmt.Errorf("failed to get bank instructions: %w", err)
	}
	defer rows.Close()

	// Group instructions by channel
	channelMap := make(map[string][]string)
	channelOrder := []string{}

	for rows.Next() {
		var channel, instruction string
		if err := rows.Scan(&channel, &instruction); err != nil {
			return nil, fmt.Errorf("failed to scan instruction: %w", err)
		}

		if _, exists := channelMap[channel]; !exists {
			channelOrder = append(channelOrder, channel)
		}
		channelMap[channel] = append(channelMap[channel], instruction)
	}

	// Convert to slice maintaining order
	result := make([]models.PaymentInstructionGroup, 0, len(channelOrder))
	for _, channel := range channelOrder {
		result = append(result, models.PaymentInstructionGroup{
			Channel: channel,
			Steps:   channelMap[channel],
		})
	}

	return result, nil
}

// LogSync creates a sync log entry for audit
func (r *orderPaymentRepository) LogSync(syncLog *models.CorePaymentSyncLog) error {
	_, err := r.db.Exec(`
		INSERT INTO core_payment_sync_log (
			payment_id, order_id, order_code, sync_type, sync_status,
			local_payment_status, local_order_status, gateway_status,
			gateway_transaction_id, has_mismatch, error_message
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`,
		syncLog.PaymentID,
		syncLog.OrderID,
		syncLog.OrderCode,
		syncLog.SyncType,
		syncLog.SyncStatus,
		syncLog.LocalPaymentStatus,
		syncLog.LocalOrderStatus,
		syncLog.GatewayStatus,
		syncLog.GatewayTransactionID,
		syncLog.HasMismatch,
		syncLog.ErrorMessage,
	)

	if err != nil {
		log.Printf("⚠️ Failed to log sync: %v", err)
		return err
	}

	return nil
}

// isUniqueViolation checks if the error is a unique constraint violation
func isUniqueViolation(err error) bool {
	// PostgreSQL unique violation error code is 23505
	return err != nil && (
		// Check for common PostgreSQL unique violation patterns
		contains(err.Error(), "duplicate key") ||
		contains(err.Error(), "unique constraint") ||
		contains(err.Error(), "23505"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ============================================
// TRANSACTION-AWARE METHODS FOR WEBHOOK
// ============================================

// FindByOrderIDForUpdateTx finds payment with row lock within a transaction
func (r *orderPaymentRepository) FindByOrderIDForUpdateTx(tx *sql.Tx, orderID int) (*models.OrderPayment, error) {
	payment := &models.OrderPayment{}
	var rawResponseJSON []byte

	err := tx.QueryRow(`
		SELECT id, order_id, payment_method, bank, va_number, transaction_id,
			   midtrans_order_id, expiry_time, payment_status, raw_response,
			   created_at, updated_at, paid_at
		FROM order_payments
		WHERE order_id = $1
		ORDER BY created_at DESC
		LIMIT 1
		FOR UPDATE
	`, orderID).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.PaymentMethod,
		&payment.Bank,
		&payment.VANumber,
		&payment.TransactionID,
		&payment.MidtransOrderID,
		&payment.ExpiryTime,
		&payment.PaymentStatus,
		&rawResponseJSON,
		&payment.CreatedAt,
		&payment.UpdatedAt,
		&payment.PaidAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPaymentNotFound
		}
		return nil, fmt.Errorf("failed to find payment for update: %w", err)
	}

	if len(rawResponseJSON) > 0 {
		json.Unmarshal(rawResponseJSON, &payment.RawResponse)
	}

	return payment, nil
}

// UpdateToPaidTx marks payment as paid within a transaction
func (r *orderPaymentRepository) UpdateToPaidTx(tx *sql.Tx, paymentID int, transactionID string) error {
	_, err := tx.Exec(`
		UPDATE order_payments
		SET payment_status = 'PAID', 
			transaction_id = COALESCE(NULLIF($1, ''), transaction_id),
			paid_at = NOW(), 
			updated_at = NOW()
		WHERE id = $2
	`, transactionID, paymentID)

	return err
}

// UpdateToExpiredTx marks payment as expired within a transaction
func (r *orderPaymentRepository) UpdateToExpiredTx(tx *sql.Tx, paymentID int) error {
	_, err := tx.Exec(`
		UPDATE order_payments
		SET payment_status = 'EXPIRED', updated_at = NOW()
		WHERE id = $1
	`, paymentID)

	return err
}

// UpdateStatusTx updates payment status within a transaction
func (r *orderPaymentRepository) UpdateStatusTx(tx *sql.Tx, paymentID int, newStatus models.CorePaymentStatus) error {
	_, err := tx.Exec(`
		UPDATE order_payments
		SET payment_status = $1, updated_at = NOW()
		WHERE id = $2
	`, newStatus, paymentID)

	return err
}

// BeginTx starts a new transaction
func (r *orderPaymentRepository) BeginTx() (*sql.Tx, error) {
	return r.db.Begin()
}

// GetPendingPaymentsForUser returns orders that are waiting for payment (already have payment generated)
func (r *orderPaymentRepository) GetPendingPaymentsForUser(userID int, page, pageSize int) ([]map[string]interface{}, int, error) {
	offset := (page - 1) * pageSize

	// Count total - PENDING orders (with or without payment)
	var totalCount int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM orders o
		WHERE o.user_id = $1 AND o.status = 'PENDING'
	`, userID).Scan(&totalCount)
	if err != nil {
		log.Printf("❌ GetPendingPaymentsForUser count error: %v", err)
		return nil, 0, err
	}

	log.Printf("📊 GetPendingPaymentsForUser: user_id=%d, total_count=%d", userID, totalCount)

	// Get orders with payment info - LEFT JOIN to include orders without payment
	rows, err := r.db.Query(`
		SELECT 
			o.id, o.order_code, o.total_amount, o.created_at,
			COALESCE((SELECT COUNT(*) FROM order_items WHERE order_id = o.id), 0) as item_count,
			COALESCE((SELECT product_name FROM order_items WHERE order_id = o.id LIMIT 1), 'Pesanan') as first_item,
			p.id as payment_id, p.payment_method, p.bank, p.va_number, p.expiry_time, p.payment_status
		FROM orders o
		LEFT JOIN order_payments p ON o.id = p.order_id AND p.payment_status = 'PENDING'
		WHERE o.user_id = $1 AND o.status = 'PENDING'
		ORDER BY o.created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, pageSize, offset)
	if err != nil {
		log.Printf("❌ GetPendingPaymentsForUser query error: %v", err)
		return nil, 0, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var (
			orderID, itemCount int
			orderCode, firstItem string
			totalAmount float64
			createdAt time.Time
			paymentID sql.NullInt64
			paymentMethod, bank, vaNumber sql.NullString
			expiryTime sql.NullTime
			paymentStatus sql.NullString
		)

		err := rows.Scan(
			&orderID, &orderCode, &totalAmount, &createdAt,
			&itemCount, &firstItem,
			&paymentID, &paymentMethod, &bank, &vaNumber, &expiryTime, &paymentStatus,
		)
		if err != nil {
			log.Printf("❌ GetPendingPaymentsForUser scan error: %v", err)
			return nil, 0, err
		}

		result := map[string]interface{}{
			"order_id":     orderID,
			"order_code":   orderCode,
			"total_amount": totalAmount,
			"item_count":   itemCount,
			"item_summary": firstItem,
			"created_at":   createdAt,
			"has_payment":  paymentID.Valid, // True if payment exists
		}

		// Only add payment details if payment exists
		if paymentID.Valid {
			result["payment_method"] = paymentMethod.String
			result["bank"] = bank.String
			result["va_number_masked"] = maskVANumber(vaNumber.String)
			result["expiry_time"] = expiryTime.Time
			result["remaining_seconds"] = int(time.Until(expiryTime.Time).Seconds())
		}

		results = append(results, result)
	}

	log.Printf("✅ GetPendingPaymentsForUser: found %d orders", len(results))
	return results, totalCount, nil
}

// GetTransactionHistoryForUser returns ALL orders for a user (Tokopedia-style)
// Supports filtering by status category: all, ongoing, completed, failed
func (r *orderPaymentRepository) GetTransactionHistoryForUser(userID int, page, pageSize int) ([]map[string]interface{}, int, error) {
	return r.GetTransactionHistoryForUserWithFilter(userID, "all", page, pageSize)
}

// GetTransactionHistoryForUserWithFilter returns orders filtered by status category
// filter: "all", "ongoing" (PAID, PACKING, SHIPPED), "completed" (DELIVERED, COMPLETED), "failed" (CANCELLED, EXPIRED, FAILED, REFUNDED)
func (r *orderPaymentRepository) GetTransactionHistoryForUserWithFilter(userID int, filter string, page, pageSize int) ([]map[string]interface{}, int, error) {
	offset := (page - 1) * pageSize

	// Build WHERE clause based on filter
	// Use only valid enum values from order_status enum
	var statusFilter string
	switch filter {
	case "ongoing":
		// Berlangsung: Pesanan yang sedang diproses
		statusFilter = "AND o.status IN ('PAID', 'PACKING', 'PROCESSING', 'SHIPPED')"
	case "completed":
		// Selesai: Pesanan yang sudah selesai
		statusFilter = "AND o.status IN ('DELIVERED', 'COMPLETED')"
	case "failed":
		// Tidak Berhasil: Pesanan yang gagal/dibatalkan
		statusFilter = "AND o.status IN ('CANCELLED', 'EXPIRED', 'KADALUARSA', 'FAILED', 'REFUNDED')"
	default:
		// Semua: Exclude hanya yang masih menunggu pembayaran (PENDING with payment)
		// Show all orders that have been paid or completed
		statusFilter = "AND o.status != 'PENDING'"
	}

	// Count total
	var totalCount int
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM orders o
		WHERE o.user_id = $1 %s
	`, statusFilter)
	err := r.db.QueryRow(countQuery, userID).Scan(&totalCount)
	if err != nil {
		log.Printf("❌ GetTransactionHistoryForUserWithFilter count error: %v", err)
		return nil, 0, err
	}

	// Get orders with full details (Tokopedia-style)
	// Use COALESCE for columns that might not exist in older schemas
	query := fmt.Sprintf(`
		SELECT 
			o.id, o.order_code, o.total_amount, o.status::text, o.created_at, o.paid_at,
			COALESCE(o.resi, '') as resi,
			COALESCE((SELECT COUNT(*) FROM order_items WHERE order_id = o.id), 0) as item_count,
			COALESCE((SELECT product_name FROM order_items WHERE order_id = o.id LIMIT 1), 'Pesanan') as first_item,
			COALESCE((SELECT pi.image_url FROM order_items oi 
				LEFT JOIN product_images pi ON oi.product_id = pi.product_id AND pi.is_primary = true
				WHERE oi.order_id = o.id LIMIT 1), '') as product_image,
			COALESCE(p.payment_method::text, '') as payment_method,
			COALESCE(p.bank, '') as bank,
			COALESCE(s.provider_name, '') as courier_name,
			COALESCE(s.service_name, '') as courier_service,
			COALESCE(s.status::text, '') as shipment_status,
			COALESCE(s.tracking_number, '') as tracking_number,
			s.shipped_at as shipment_shipped_at,
			s.delivered_at as shipment_delivered_at,
			o.cancelled_at
		FROM orders o
		LEFT JOIN order_payments p ON o.id = p.order_id
		LEFT JOIN shipments s ON o.id = s.order_id
		WHERE o.user_id = $1 %s
		ORDER BY o.created_at DESC
		LIMIT $2 OFFSET $3
	`, statusFilter)

	rows, err := r.db.Query(query, userID, pageSize, offset)
	if err != nil {
		log.Printf("❌ GetTransactionHistoryForUserWithFilter query error: %v", err)
		return nil, 0, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var (
			orderID, itemCount                                          int
			orderCode, status, firstItem, productImage                  string
			resi, paymentMethod, bank                                   string
			courierName, courierService, shipmentStatus, trackingNumber string
			totalAmount                                                 float64
			createdAt                                                   time.Time
			paidAt, shippedAt, deliveredAt, cancelledAt                 sql.NullTime
		)

		err := rows.Scan(
			&orderID, &orderCode, &totalAmount, &status, &createdAt, &paidAt,
			&resi, &itemCount, &firstItem, &productImage,
			&paymentMethod, &bank,
			&courierName, &courierService, &shipmentStatus, &trackingNumber,
			&shippedAt, &deliveredAt, &cancelledAt,
		)
		if err != nil {
			log.Printf("❌ GetTransactionHistoryForUserWithFilter scan error: %v", err)
			continue
		}

		result := map[string]interface{}{
			"order_id":      orderID,
			"order_code":    orderCode,
			"total_amount":  totalAmount,
			"status":        status,
			"item_count":    itemCount,
			"item_summary":  firstItem,
			"product_image": productImage,
			"created_at":    createdAt,
			"resi":          resi,
		}

		// Payment info
		if paymentMethod != "" {
			result["payment_method"] = paymentMethod
		}
		if bank != "" {
			result["bank"] = bank
		}

		// Timestamps
		if paidAt.Valid {
			result["paid_at"] = paidAt.Time
		}
		if shippedAt.Valid {
			result["shipped_at"] = shippedAt.Time
		}
		if deliveredAt.Valid {
			result["delivered_at"] = deliveredAt.Time
		}
		if cancelledAt.Valid {
			result["cancelled_at"] = cancelledAt.Time
		}
		// Shipment info
		if courierName != "" {
			result["courier_name"] = courierName
			result["courier_service"] = courierService
			result["shipment_status"] = shipmentStatus
			result["tracking_number"] = trackingNumber
		}

		results = append(results, result)
	}

	return results, totalCount, nil
}

func maskVANumber(va string) string {
	if len(va) <= 4 {
		return va
	}
	return "****" + va[len(va)-4:]
}
