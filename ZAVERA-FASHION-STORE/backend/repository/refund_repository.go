package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"
	"zavera/models"
)

type RefundRepository interface {
	Create(refund *models.Refund) error
	CreateWithTx(tx *sql.Tx, refund *models.Refund) error
	FindByID(id int) (*models.Refund, error)
	FindByCode(code string) (*models.Refund, error)
	FindByOrderID(orderID int) ([]*models.Refund, error)
	FindByIdempotencyKey(key string) (*models.Refund, error)
	FindAll(page, pageSize int, status, orderCode string) ([]*models.Refund, int, error)
	UpdateStatus(id int, status models.RefundStatus, gatewayResponse map[string]any) error
	UpdateStatusWithTx(tx *sql.Tx, id int, status models.RefundStatus, gatewayResponse map[string]any) error
	MarkCompleted(id int, gatewayRefundID string, gatewayResponse map[string]any) error
	MarkFailed(id int, errorMsg string, gatewayResponse map[string]any) error
	CreateRefundItem(item *models.RefundItem) error
	CreateRefundItemWithTx(tx *sql.Tx, item *models.RefundItem) error
	FindItemsByRefundID(refundID int) ([]models.RefundItem, error)
	MarkItemStockRestored(itemID int) error
	RecordStatusChange(refundID int, from, to models.RefundStatus, changedBy, reason string) error
	RecordStatusChangeWithTx(tx *sql.Tx, refundID int, from, to models.RefundStatus, changedBy, reason string) error
	GetStatusHistory(refundID int) ([]models.RefundStatusHistory, error)
	GetDB() *sql.DB
}

type refundRepository struct {
	db *sql.DB
}

func NewRefundRepository(db *sql.DB) RefundRepository {
	return &refundRepository{db: db}
}

func (r *refundRepository) Create(refund *models.Refund) error {
	return r.createInternal(r.db, refund)
}

func (r *refundRepository) CreateWithTx(tx *sql.Tx, refund *models.Refund) error {
	return r.createInternal(tx, refund)
}

type dbExecutor interface {
	QueryRow(query string, args ...any) *sql.Row
	Exec(query string, args ...any) (sql.Result, error)
}

func (r *refundRepository) createInternal(db dbExecutor, refund *models.Refund) error {
	query := `
		INSERT INTO refunds (
			refund_code, order_id, payment_id, refund_type, reason, reason_detail,
			original_amount, refund_amount, shipping_refund, items_refund,
			status, idempotency_key, requested_by, requested_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at, updated_at
	`

	err := db.QueryRow(
		query,
		refund.RefundCode, refund.OrderID, refund.PaymentID, refund.RefundType,
		refund.Reason, refund.ReasonDetail, refund.OriginalAmount, refund.RefundAmount,
		refund.ShippingRefund, refund.ItemsRefund, refund.Status, refund.IdempotencyKey,
		refund.RequestedBy, time.Now(),
	).Scan(&refund.ID, &refund.CreatedAt, &refund.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create refund %s: %w", refund.RefundCode, err)
	}
	return nil
}

func (r *refundRepository) FindByID(id int) (*models.Refund, error) {
	query := `
		SELECT id, refund_code, order_id, payment_id, refund_type, reason, reason_detail,
		       original_amount, refund_amount, shipping_refund, items_refund, status,
		       gateway_refund_id, gateway_status, gateway_response, idempotency_key,
		       processed_by, processed_at, requested_by, requested_at,
		       created_at, updated_at, completed_at
		FROM refunds WHERE id = $1
	`
	refund, err := r.scanRefund(r.db.QueryRow(query, id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("refund with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to find refund by ID %d: %w", id, err)
	}
	return refund, nil
}

func (r *refundRepository) FindByCode(code string) (*models.Refund, error) {
	query := `
		SELECT id, refund_code, order_id, payment_id, refund_type, reason, reason_detail,
		       original_amount, refund_amount, shipping_refund, items_refund, status,
		       gateway_refund_id, gateway_status, gateway_response, idempotency_key,
		       processed_by, processed_at, requested_by, requested_at,
		       created_at, updated_at, completed_at
		FROM refunds WHERE refund_code = $1
	`
	refund, err := r.scanRefund(r.db.QueryRow(query, code))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("refund with code %s not found", code)
		}
		return nil, fmt.Errorf("failed to find refund by code %s: %w", code, err)
	}
	return refund, nil
}

func (r *refundRepository) FindByOrderID(orderID int) ([]*models.Refund, error) {
	query := `
		SELECT id, refund_code, order_id, payment_id, refund_type, reason, reason_detail,
		       original_amount, refund_amount, shipping_refund, items_refund, status,
		       gateway_refund_id, gateway_status, gateway_response, idempotency_key,
		       processed_by, processed_at, requested_by, requested_at,
		       created_at, updated_at, completed_at
		FROM refunds WHERE order_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to find refunds for order %d: %w", orderID, err)
	}
	defer rows.Close()

	var refunds []*models.Refund
	for rows.Next() {
		refund, err := r.scanRefundFromRows(rows)
		if err != nil {
			continue
		}
		refunds = append(refunds, refund)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating refund rows for order %d: %w", orderID, err)
	}
	
	return refunds, nil
}

func (r *refundRepository) FindAll(page, pageSize int, status, orderCode string) ([]*models.Refund, int, error) {
	// Build query with filters
	baseQuery := `
		SELECT r.id, r.refund_code, r.order_id, r.payment_id, r.refund_type, r.reason, r.reason_detail,
		       r.original_amount, r.refund_amount, r.shipping_refund, r.items_refund, r.status,
		       r.gateway_refund_id, r.gateway_status, r.gateway_response, r.idempotency_key,
		       r.processed_by, r.processed_at, r.requested_by, r.requested_at,
		       r.created_at, r.updated_at, r.completed_at
		FROM refunds r
	`
	
	countQuery := `SELECT COUNT(*) FROM refunds r`
	
	// Add filters
	var whereClauses []string
	var args []interface{}
	argIndex := 1
	
	if status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("r.status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}
	
	if orderCode != "" {
		baseQuery += " JOIN orders o ON r.order_id = o.id"
		countQuery += " JOIN orders o ON r.order_id = o.id"
		whereClauses = append(whereClauses, fmt.Sprintf("o.order_code = $%d", argIndex))
		args = append(args, orderCode)
		argIndex++
	}
	
	if len(whereClauses) > 0 {
		whereClause := " WHERE " + whereClauses[0]
		for i := 1; i < len(whereClauses); i++ {
			whereClause += " AND " + whereClauses[i]
		}
		baseQuery += whereClause
		countQuery += whereClause
	}
	
	// Get total count (use same args for count query)
	var totalCount int
	err := r.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count refunds: %w", err)
	}
	
	// Add pagination to data query
	baseQuery += " ORDER BY r.created_at DESC"
	offset := (page - 1) * pageSize
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	
	// Append pagination args to the same args array
	queryArgs := append(args, pageSize, offset)
	
	// Debug logging
	log.Printf("🔍 FindAll SQL Query: %s", baseQuery)
	log.Printf("🔍 FindAll Query Args: %v", queryArgs)
	
	// Execute query
	rows, err := r.db.Query(baseQuery, queryArgs...)
	if err != nil {
		log.Printf("❌ FindAll Query Error: %v", err)
		return nil, 0, fmt.Errorf("failed to find refunds: %w", err)
	}
	defer rows.Close()
	
	var refunds []*models.Refund
	rowCount := 0
	for rows.Next() {
		refund, err := r.scanRefundFromRows(rows)
		if err != nil {
			log.Printf("⚠️ Failed to scan refund row: %v", err)
			continue
		}
		refunds = append(refunds, refund)
		rowCount++
	}
	
	log.Printf("✅ FindAll scanned %d refunds from %d total", rowCount, totalCount)
	
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating refund rows: %w", err)
	}
	
	return refunds, totalCount, nil
}

func (r *refundRepository) FindByIdempotencyKey(key string) (*models.Refund, error) {
	query := `
		SELECT id, refund_code, order_id, payment_id, refund_type, reason, reason_detail,
		       original_amount, refund_amount, shipping_refund, items_refund, status,
		       gateway_refund_id, gateway_status, gateway_response, idempotency_key,
		       processed_by, processed_at, requested_by, requested_at,
		       created_at, updated_at, completed_at
		FROM refunds WHERE idempotency_key = $1
	`
	refund, err := r.scanRefund(r.db.QueryRow(query, key))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Return nil without error if not found (idempotency check)
		}
		return nil, fmt.Errorf("failed to find refund by idempotency key: %w", err)
	}
	return refund, nil
}

func (r *refundRepository) scanRefund(row *sql.Row) (*models.Refund, error) {
	var refund models.Refund
	var gatewayResponseJSON []byte

	err := row.Scan(
		&refund.ID, &refund.RefundCode, &refund.OrderID, &refund.PaymentID,
		&refund.RefundType, &refund.Reason, &refund.ReasonDetail,
		&refund.OriginalAmount, &refund.RefundAmount, &refund.ShippingRefund,
		&refund.ItemsRefund, &refund.Status, &refund.GatewayRefundID,
		&refund.GatewayStatus, &gatewayResponseJSON, &refund.IdempotencyKey,
		&refund.ProcessedBy, &refund.ProcessedAt, &refund.RequestedBy,
		&refund.RequestedAt, &refund.CreatedAt, &refund.UpdatedAt, &refund.CompletedAt,
	)
	if err != nil {
		return nil, err
	}

	if len(gatewayResponseJSON) > 0 {
		json.Unmarshal(gatewayResponseJSON, &refund.GatewayResponse)
	}

	return &refund, nil
}

func (r *refundRepository) scanRefundFromRows(rows *sql.Rows) (*models.Refund, error) {
	var refund models.Refund
	var gatewayResponseJSON []byte

	err := rows.Scan(
		&refund.ID, &refund.RefundCode, &refund.OrderID, &refund.PaymentID,
		&refund.RefundType, &refund.Reason, &refund.ReasonDetail,
		&refund.OriginalAmount, &refund.RefundAmount, &refund.ShippingRefund,
		&refund.ItemsRefund, &refund.Status, &refund.GatewayRefundID,
		&refund.GatewayStatus, &gatewayResponseJSON, &refund.IdempotencyKey,
		&refund.ProcessedBy, &refund.ProcessedAt, &refund.RequestedBy,
		&refund.RequestedAt, &refund.CreatedAt, &refund.UpdatedAt, &refund.CompletedAt,
	)
	if err != nil {
		return nil, err
	}

	if len(gatewayResponseJSON) > 0 {
		json.Unmarshal(gatewayResponseJSON, &refund.GatewayResponse)
	}

	return &refund, nil
}

func (r *refundRepository) UpdateStatus(id int, status models.RefundStatus, gatewayResponse map[string]any) error {
	gatewayResponseJSON, _ := json.Marshal(gatewayResponse)
	query := `
		UPDATE refunds SET status = $1, gateway_response = $2, updated_at = NOW()
		WHERE id = $3
	`
	_, err := r.db.Exec(query, status, gatewayResponseJSON, id)
	if err != nil {
		return fmt.Errorf("failed to update status for refund %d: %w", id, err)
	}
	return nil
}

func (r *refundRepository) UpdateStatusWithTx(tx *sql.Tx, id int, status models.RefundStatus, gatewayResponse map[string]any) error {
	gatewayResponseJSON, _ := json.Marshal(gatewayResponse)
	query := `
		UPDATE refunds SET status = $1, gateway_response = $2, updated_at = NOW()
		WHERE id = $3
	`
	_, err := tx.Exec(query, status, gatewayResponseJSON, id)
	if err != nil {
		return fmt.Errorf("failed to update status for refund %d in transaction: %w", id, err)
	}
	return nil
}

func (r *refundRepository) MarkCompleted(id int, gatewayRefundID string, gatewayResponse map[string]any) error {
	gatewayResponseJSON, _ := json.Marshal(gatewayResponse)
	query := `
		UPDATE refunds 
		SET status = $1, gateway_refund_id = $2, gateway_status = 'success',
		    gateway_response = $3, completed_at = NOW(), updated_at = NOW()
		WHERE id = $4
	`
	_, err := r.db.Exec(query, models.RefundStatusCompleted, gatewayRefundID, gatewayResponseJSON, id)
	if err != nil {
		return fmt.Errorf("failed to mark refund %d as completed: %w", id, err)
	}
	return nil
}

func (r *refundRepository) MarkFailed(id int, errorMsg string, gatewayResponse map[string]any) error {
	gatewayResponseJSON, _ := json.Marshal(gatewayResponse)
	query := `
		UPDATE refunds 
		SET status = $1, gateway_status = 'failed', gateway_response = $2, 
		    reason_detail = COALESCE(reason_detail, '') || ' | Error: ' || $3, updated_at = NOW()
		WHERE id = $4
	`
	_, err := r.db.Exec(query, models.RefundStatusFailed, gatewayResponseJSON, errorMsg, id)
	if err != nil {
		return fmt.Errorf("failed to mark refund %d as failed: %w", id, err)
	}
	return nil
}

func (r *refundRepository) CreateRefundItem(item *models.RefundItem) error {
	return r.createRefundItemInternal(r.db, item)
}

func (r *refundRepository) CreateRefundItemWithTx(tx *sql.Tx, item *models.RefundItem) error {
	return r.createRefundItemInternal(tx, item)
}

func (r *refundRepository) createRefundItemInternal(db dbExecutor, item *models.RefundItem) error {
	query := `
		INSERT INTO refund_items (
			refund_id, order_item_id, product_id, product_name,
			quantity, price_per_unit, refund_amount, item_reason
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`
	err := db.QueryRow(
		query,
		item.RefundID, item.OrderItemID, item.ProductID, item.ProductName,
		item.Quantity, item.PricePerUnit, item.RefundAmount, item.ItemReason,
	).Scan(&item.ID, &item.CreatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create refund item for refund %d: %w", item.RefundID, err)
	}
	return nil
}

func (r *refundRepository) FindItemsByRefundID(refundID int) ([]models.RefundItem, error) {
	query := `
		SELECT id, refund_id, order_item_id, product_id, product_name,
		       quantity, price_per_unit, refund_amount, item_reason,
		       stock_restored, stock_restored_at, created_at
		FROM refund_items WHERE refund_id = $1
	`
	rows, err := r.db.Query(query, refundID)
	if err != nil {
		return nil, fmt.Errorf("failed to find items for refund %d: %w", refundID, err)
	}
	defer rows.Close()

	var items []models.RefundItem
	for rows.Next() {
		var item models.RefundItem
		err := rows.Scan(
			&item.ID, &item.RefundID, &item.OrderItemID, &item.ProductID,
			&item.ProductName, &item.Quantity, &item.PricePerUnit,
			&item.RefundAmount, &item.ItemReason, &item.StockRestored,
			&item.StockRestoredAt, &item.CreatedAt,
		)
		if err != nil {
			continue
		}
		items = append(items, item)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating refund items for refund %d: %w", refundID, err)
	}
	
	return items, nil
}

func (r *refundRepository) MarkItemStockRestored(itemID int) error {
	query := `
		UPDATE refund_items SET stock_restored = true, stock_restored_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.Exec(query, itemID)
	if err != nil {
		return fmt.Errorf("failed to mark stock restored for refund item %d: %w", itemID, err)
	}
	return nil
}

func (r *refundRepository) RecordStatusChange(refundID int, from, to models.RefundStatus, changedBy, reason string) error {
	return r.RecordStatusChangeWithTx(nil, refundID, from, to, changedBy, reason)
}

func (r *refundRepository) RecordStatusChangeWithTx(tx *sql.Tx, refundID int, from, to models.RefundStatus, changedBy, reason string) error {
	query := `
		INSERT INTO refund_status_history (refund_id, from_status, to_status, changed_by, reason)
		VALUES ($1, $2, $3, $4, $5)
	`
	var fromStatus *string
	if from != "" {
		fromStr := string(from)
		fromStatus = &fromStr
	}
	
	var err error
	if tx != nil {
		_, err = tx.Exec(query, refundID, fromStatus, string(to), changedBy, reason)
	} else {
		_, err = r.db.Exec(query, refundID, fromStatus, string(to), changedBy, reason)
	}
	
	if err != nil {
		return fmt.Errorf("failed to record status change for refund %d: %w", refundID, err)
	}
	return nil
}

func (r *refundRepository) GetStatusHistory(refundID int) ([]models.RefundStatusHistory, error) {
	query := `
		SELECT id, refund_id, from_status, to_status, changed_by, reason, created_at
		FROM refund_status_history
		WHERE refund_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(query, refundID)
	if err != nil {
		return nil, fmt.Errorf("failed to get status history for refund %d: %w", refundID, err)
	}
	defer rows.Close()

	var history []models.RefundStatusHistory
	for rows.Next() {
		var h models.RefundStatusHistory
		var fromStatus sql.NullString
		var newStatus string
		
		err := rows.Scan(
			&h.ID, &h.RefundID, &fromStatus, &newStatus,
			&h.Actor, &h.Reason, &h.CreatedAt,
		)
		if err != nil {
			continue
		}
		
		if fromStatus.Valid {
			h.OldStatus = &fromStatus.String
		}
		h.NewStatus = models.RefundStatus(newStatus)
		
		history = append(history, h)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating status history rows: %w", err)
	}
	
	return history, nil
}

func (r *refundRepository) GetDB() *sql.DB {
	return r.db
}

// GenerateRefundCode generates a unique refund code
func GenerateRefundCode() string {
	return fmt.Sprintf("RFD-%s-%s", time.Now().Format("20060102"), randomHex(4))
}
