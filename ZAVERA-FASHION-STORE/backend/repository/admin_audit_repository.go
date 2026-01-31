package repository

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"time"
	"zavera/models"
)

type AdminAuditRepository interface {
	Create(log *models.AdminAuditLog) error
	CreateWithTx(tx *sql.Tx, log *models.AdminAuditLog) error
	FindByID(id int) (*models.AdminAuditLog, error)
	FindByIdempotencyKey(key string) (*models.AdminAuditLog, error)
	FindByTarget(targetType string, targetID int) ([]*models.AdminAuditLog, error)
	FindByAdmin(adminUserID int, limit int) ([]*models.AdminAuditLog, error)
	FindRecent(limit int) ([]*models.AdminAuditLog, error)
}

type adminAuditRepository struct {
	db *sql.DB
}

func NewAdminAuditRepository(db *sql.DB) AdminAuditRepository {
	return &adminAuditRepository{db: db}
}

func (r *adminAuditRepository) Create(log *models.AdminAuditLog) error {
	return r.createInternal(r.db, log)
}

func (r *adminAuditRepository) CreateWithTx(tx *sql.Tx, log *models.AdminAuditLog) error {
	return r.createInternal(tx, log)
}

func (r *adminAuditRepository) createInternal(db dbExecutor, log *models.AdminAuditLog) error {
	stateBeforeJSON, _ := json.Marshal(log.StateBefore)
	stateAfterJSON, _ := json.Marshal(log.StateAfter)
	metadataJSON, _ := json.Marshal(log.Metadata)

	query := `
		INSERT INTO admin_audit_log (
			admin_user_id, admin_email, admin_ip, admin_user_agent,
			action_type, action_detail, target_type, target_id, target_code,
			state_before, state_after, success, error_message,
			idempotency_key, metadata
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, created_at
	`

	return db.QueryRow(
		query,
		log.AdminUserID, log.AdminEmail, log.AdminIP, log.AdminUserAgent,
		log.ActionType, log.ActionDetail, log.TargetType, log.TargetID, log.TargetCode,
		stateBeforeJSON, stateAfterJSON, log.Success, log.ErrorMessage,
		log.IdempotencyKey, metadataJSON,
	).Scan(&log.ID, &log.CreatedAt)
}

func (r *adminAuditRepository) FindByID(id int) (*models.AdminAuditLog, error) {
	query := `
		SELECT id, admin_user_id, admin_email, admin_ip, admin_user_agent,
		       action_type, action_detail, target_type, target_id, target_code,
		       state_before, state_after, success, error_message,
		       idempotency_key, metadata, created_at
		FROM admin_audit_log WHERE id = $1
	`
	return r.scanAuditLog(r.db.QueryRow(query, id))
}

func (r *adminAuditRepository) FindByIdempotencyKey(key string) (*models.AdminAuditLog, error) {
	query := `
		SELECT id, admin_user_id, admin_email, admin_ip, admin_user_agent,
		       action_type, action_detail, target_type, target_id, target_code,
		       state_before, state_after, success, error_message,
		       idempotency_key, metadata, created_at
		FROM admin_audit_log WHERE idempotency_key = $1
	`
	return r.scanAuditLog(r.db.QueryRow(query, key))
}

func (r *adminAuditRepository) FindByTarget(targetType string, targetID int) ([]*models.AdminAuditLog, error) {
	query := `
		SELECT id, admin_user_id, admin_email, admin_ip, admin_user_agent,
		       action_type, action_detail, target_type, target_id, target_code,
		       state_before, state_after, success, error_message,
		       idempotency_key, metadata, created_at
		FROM admin_audit_log 
		WHERE target_type = $1 AND target_id = $2
		ORDER BY created_at DESC
	`
	return r.queryMultiple(query, targetType, targetID)
}

func (r *adminAuditRepository) FindByAdmin(adminUserID int, limit int) ([]*models.AdminAuditLog, error) {
	query := `
		SELECT id, admin_user_id, admin_email, admin_ip, admin_user_agent,
		       action_type, action_detail, target_type, target_id, target_code,
		       state_before, state_after, success, error_message,
		       idempotency_key, metadata, created_at
		FROM admin_audit_log 
		WHERE admin_user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	return r.queryMultiple(query, adminUserID, limit)
}

func (r *adminAuditRepository) FindRecent(limit int) ([]*models.AdminAuditLog, error) {
	query := `
		SELECT id, admin_user_id, admin_email, admin_ip, admin_user_agent,
		       action_type, action_detail, target_type, target_id, target_code,
		       state_before, state_after, success, error_message,
		       idempotency_key, metadata, created_at
		FROM admin_audit_log 
		ORDER BY created_at DESC
		LIMIT $1
	`
	return r.queryMultiple(query, limit)
}

func (r *adminAuditRepository) queryMultiple(query string, args ...any) ([]*models.AdminAuditLog, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.AdminAuditLog
	for rows.Next() {
		log, err := r.scanAuditLogFromRows(rows)
		if err != nil {
			continue
		}
		logs = append(logs, log)
	}
	return logs, nil
}

func (r *adminAuditRepository) scanAuditLog(row *sql.Row) (*models.AdminAuditLog, error) {
	var log models.AdminAuditLog
	var stateBeforeJSON, stateAfterJSON, metadataJSON []byte

	err := row.Scan(
		&log.ID, &log.AdminUserID, &log.AdminEmail, &log.AdminIP, &log.AdminUserAgent,
		&log.ActionType, &log.ActionDetail, &log.TargetType, &log.TargetID, &log.TargetCode,
		&stateBeforeJSON, &stateAfterJSON, &log.Success, &log.ErrorMessage,
		&log.IdempotencyKey, &metadataJSON, &log.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if len(stateBeforeJSON) > 0 {
		json.Unmarshal(stateBeforeJSON, &log.StateBefore)
	}
	if len(stateAfterJSON) > 0 {
		json.Unmarshal(stateAfterJSON, &log.StateAfter)
	}
	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &log.Metadata)
	}

	return &log, nil
}

func (r *adminAuditRepository) scanAuditLogFromRows(rows *sql.Rows) (*models.AdminAuditLog, error) {
	var log models.AdminAuditLog
	var stateBeforeJSON, stateAfterJSON, metadataJSON []byte

	err := rows.Scan(
		&log.ID, &log.AdminUserID, &log.AdminEmail, &log.AdminIP, &log.AdminUserAgent,
		&log.ActionType, &log.ActionDetail, &log.TargetType, &log.TargetID, &log.TargetCode,
		&stateBeforeJSON, &stateAfterJSON, &log.Success, &log.ErrorMessage,
		&log.IdempotencyKey, &metadataJSON, &log.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if len(stateBeforeJSON) > 0 {
		json.Unmarshal(stateBeforeJSON, &log.StateBefore)
	}
	if len(stateAfterJSON) > 0 {
		json.Unmarshal(stateAfterJSON, &log.StateAfter)
	}
	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &log.Metadata)
	}

	return &log, nil
}

// Helper function for generating random hex strings
func randomHex(n int) string {
	bytes := make([]byte, n)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// PaymentSyncRepository handles payment sync log operations
type PaymentSyncRepository interface {
	Create(log *models.PaymentSyncLog) error
	FindByPaymentID(paymentID int) ([]*models.PaymentSyncLog, error)
	FindUnresolved() ([]*models.PaymentSyncLog, error)
	FindPendingRetry() ([]*models.PaymentSyncLog, error)
	UpdateSyncStatus(id int, status models.PaymentSyncStatus, gatewayResponse map[string]any) error
	MarkMismatch(id int, mismatchType, detail string) error
	MarkResolved(id int, resolvedBy int, action string) error
	IncrementRetry(id int, nextRetryAt time.Time) error
}

type paymentSyncRepository struct {
	db *sql.DB
}

func NewPaymentSyncRepository(db *sql.DB) PaymentSyncRepository {
	return &paymentSyncRepository{db: db}
}

func (r *paymentSyncRepository) Create(log *models.PaymentSyncLog) error {
	gatewayResponseJSON, _ := json.Marshal(log.GatewayResponse)

	query := `
		INSERT INTO payment_sync_log (
			payment_id, order_id, order_code, sync_type, sync_status,
			local_payment_status, local_order_status, gateway_status,
			gateway_transaction_id, gateway_response
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`

	return r.db.QueryRow(
		query,
		log.PaymentID, log.OrderID, log.OrderCode, log.SyncType, log.SyncStatus,
		log.LocalPaymentStatus, log.LocalOrderStatus, log.GatewayStatus,
		log.GatewayTransactionID, gatewayResponseJSON,
	).Scan(&log.ID, &log.CreatedAt, &log.UpdatedAt)
}

func (r *paymentSyncRepository) FindByPaymentID(paymentID int) ([]*models.PaymentSyncLog, error) {
	query := `
		SELECT id, payment_id, order_id, order_code, sync_type, sync_status,
		       local_payment_status, local_order_status, gateway_status,
		       gateway_transaction_id, gateway_response, has_mismatch,
		       mismatch_type, mismatch_detail, resolved, resolved_by,
		       resolved_at, resolution_action, retry_count, last_retry_at,
		       next_retry_at, created_at, updated_at
		FROM payment_sync_log WHERE payment_id = $1 ORDER BY created_at DESC
	`
	return r.queryMultiple(query, paymentID)
}

func (r *paymentSyncRepository) FindUnresolved() ([]*models.PaymentSyncLog, error) {
	query := `
		SELECT id, payment_id, order_id, order_code, sync_type, sync_status,
		       local_payment_status, local_order_status, gateway_status,
		       gateway_transaction_id, gateway_response, has_mismatch,
		       mismatch_type, mismatch_detail, resolved, resolved_by,
		       resolved_at, resolution_action, retry_count, last_retry_at,
		       next_retry_at, created_at, updated_at
		FROM payment_sync_log WHERE has_mismatch = true AND resolved = false
		ORDER BY created_at DESC
	`
	return r.queryMultiple(query)
}

func (r *paymentSyncRepository) FindPendingRetry() ([]*models.PaymentSyncLog, error) {
	query := `
		SELECT id, payment_id, order_id, order_code, sync_type, sync_status,
		       local_payment_status, local_order_status, gateway_status,
		       gateway_transaction_id, gateway_response, has_mismatch,
		       mismatch_type, mismatch_detail, resolved, resolved_by,
		       resolved_at, resolution_action, retry_count, last_retry_at,
		       next_retry_at, created_at, updated_at
		FROM payment_sync_log 
		WHERE resolved = false AND next_retry_at <= NOW() AND retry_count < 5
		ORDER BY next_retry_at ASC
	`
	return r.queryMultiple(query)
}

func (r *paymentSyncRepository) queryMultiple(query string, args ...any) ([]*models.PaymentSyncLog, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.PaymentSyncLog
	for rows.Next() {
		var log models.PaymentSyncLog
		var gatewayResponseJSON []byte

		err := rows.Scan(
			&log.ID, &log.PaymentID, &log.OrderID, &log.OrderCode, &log.SyncType,
			&log.SyncStatus, &log.LocalPaymentStatus, &log.LocalOrderStatus,
			&log.GatewayStatus, &log.GatewayTransactionID, &gatewayResponseJSON,
			&log.HasMismatch, &log.MismatchType, &log.MismatchDetail, &log.Resolved,
			&log.ResolvedBy, &log.ResolvedAt, &log.ResolutionAction, &log.RetryCount,
			&log.LastRetryAt, &log.NextRetryAt, &log.CreatedAt, &log.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if len(gatewayResponseJSON) > 0 {
			json.Unmarshal(gatewayResponseJSON, &log.GatewayResponse)
		}
		logs = append(logs, &log)
	}
	return logs, nil
}

func (r *paymentSyncRepository) UpdateSyncStatus(id int, status models.PaymentSyncStatus, gatewayResponse map[string]any) error {
	gatewayResponseJSON, _ := json.Marshal(gatewayResponse)
	query := `
		UPDATE payment_sync_log 
		SET sync_status = $1, gateway_response = $2, updated_at = NOW()
		WHERE id = $3
	`
	_, err := r.db.Exec(query, status, gatewayResponseJSON, id)
	return err
}

func (r *paymentSyncRepository) MarkMismatch(id int, mismatchType, detail string) error {
	query := `
		UPDATE payment_sync_log 
		SET has_mismatch = true, mismatch_type = $1, mismatch_detail = $2,
		    sync_status = 'MISMATCH', updated_at = NOW()
		WHERE id = $3
	`
	_, err := r.db.Exec(query, mismatchType, detail, id)
	return err
}

func (r *paymentSyncRepository) MarkResolved(id int, resolvedBy int, action string) error {
	query := `
		UPDATE payment_sync_log 
		SET resolved = true, resolved_by = $1, resolved_at = NOW(),
		    resolution_action = $2, sync_status = 'RESOLVED', updated_at = NOW()
		WHERE id = $3
	`
	_, err := r.db.Exec(query, resolvedBy, action, id)
	return err
}

func (r *paymentSyncRepository) IncrementRetry(id int, nextRetryAt time.Time) error {
	query := `
		UPDATE payment_sync_log 
		SET retry_count = retry_count + 1, last_retry_at = NOW(),
		    next_retry_at = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Exec(query, nextRetryAt, id)
	return err
}

// ReconciliationRepository handles reconciliation log operations
type ReconciliationRepository interface {
	Create(log *models.ReconciliationLog) error
	FindByDate(date time.Time) (*models.ReconciliationLog, error)
	FindRecent(limit int) ([]*models.ReconciliationLog, error)
	Update(log *models.ReconciliationLog) error
	MarkCompleted(id int) error
	MarkFailed(id int, errors map[string]any) error
}

type reconciliationRepository struct {
	db *sql.DB
}

func NewReconciliationRepository(db *sql.DB) ReconciliationRepository {
	return &reconciliationRepository{db: db}
}

func (r *reconciliationRepository) Create(log *models.ReconciliationLog) error {
	mismatchDetailsJSON, _ := json.Marshal(log.MismatchDetails)
	orphanDetailsJSON, _ := json.Marshal(log.OrphanDetails)
	errorsJSON, _ := json.Marshal(log.Errors)

	query := `
		INSERT INTO reconciliation_log (
			reconciliation_date, period_start, period_end, status, run_by, started_at
		)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (reconciliation_date) DO UPDATE SET
			status = EXCLUDED.status, started_at = NOW(), updated_at = NOW()
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		query,
		log.ReconciliationDate, log.PeriodStart, log.PeriodEnd, "RUNNING", log.RunBy,
	).Scan(&log.ID, &log.CreatedAt, &log.UpdatedAt)

	if err != nil {
		return err
	}

	// Update with full data
	updateQuery := `
		UPDATE reconciliation_log SET
			total_orders = $1, total_payments = $2, total_amount = $3,
			orders_pending = $4, orders_paid = $5, orders_cancelled = $6, orders_refunded = $7,
			payments_pending = $8, payments_success = $9, payments_failed = $10,
			mismatches_found = $11, mismatches_resolved = $12, mismatch_details = $13,
			orphan_orders = $14, orphan_payments = $15, orphan_details = $16,
			stuck_payments = $17, expected_revenue = $18, actual_revenue = $19,
			revenue_variance = $20, total_refunds = $21, error_count = $22, errors = $23
		WHERE id = $24
	`

	_, err = r.db.Exec(updateQuery,
		log.TotalOrders, log.TotalPayments, log.TotalAmount,
		log.OrdersPending, log.OrdersPaid, log.OrdersCancelled, log.OrdersRefunded,
		log.PaymentsPending, log.PaymentsSuccess, log.PaymentsFailed,
		log.MismatchesFound, log.MismatchesResolved, mismatchDetailsJSON,
		log.OrphanOrders, log.OrphanPayments, orphanDetailsJSON,
		log.StuckPayments, log.ExpectedRevenue, log.ActualRevenue,
		log.RevenueVariance, log.TotalRefunds, log.ErrorCount, errorsJSON,
		log.ID,
	)

	return err
}

func (r *reconciliationRepository) FindByDate(date time.Time) (*models.ReconciliationLog, error) {
	query := `
		SELECT id, reconciliation_date, period_start, period_end,
		       total_orders, total_payments, total_amount,
		       orders_pending, orders_paid, orders_cancelled, orders_refunded,
		       payments_pending, payments_success, payments_failed,
		       mismatches_found, mismatches_resolved, mismatch_details,
		       orphan_orders, orphan_payments, orphan_details,
		       stuck_payments, expected_revenue, actual_revenue,
		       revenue_variance, total_refunds, status, started_at, completed_at,
		       run_by, error_count, errors, created_at, updated_at
		FROM reconciliation_log WHERE reconciliation_date = $1
	`
	return r.scanReconciliationLog(r.db.QueryRow(query, date.Format("2006-01-02")))
}

func (r *reconciliationRepository) FindRecent(limit int) ([]*models.ReconciliationLog, error) {
	query := `
		SELECT id, reconciliation_date, period_start, period_end,
		       total_orders, total_payments, total_amount,
		       orders_pending, orders_paid, orders_cancelled, orders_refunded,
		       payments_pending, payments_success, payments_failed,
		       mismatches_found, mismatches_resolved, mismatch_details,
		       orphan_orders, orphan_payments, orphan_details,
		       stuck_payments, expected_revenue, actual_revenue,
		       revenue_variance, total_refunds, status, started_at, completed_at,
		       run_by, error_count, errors, created_at, updated_at
		FROM reconciliation_log ORDER BY reconciliation_date DESC LIMIT $1
	`
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.ReconciliationLog
	for rows.Next() {
		log, err := r.scanReconciliationLogFromRows(rows)
		if err != nil {
			continue
		}
		logs = append(logs, log)
	}
	return logs, nil
}

func (r *reconciliationRepository) Update(log *models.ReconciliationLog) error {
	mismatchDetailsJSON, _ := json.Marshal(log.MismatchDetails)
	orphanDetailsJSON, _ := json.Marshal(log.OrphanDetails)
	errorsJSON, _ := json.Marshal(log.Errors)

	query := `
		UPDATE reconciliation_log SET
			total_orders = $1, total_payments = $2, total_amount = $3,
			orders_pending = $4, orders_paid = $5, orders_cancelled = $6, orders_refunded = $7,
			payments_pending = $8, payments_success = $9, payments_failed = $10,
			mismatches_found = $11, mismatches_resolved = $12, mismatch_details = $13,
			orphan_orders = $14, orphan_payments = $15, orphan_details = $16,
			stuck_payments = $17, expected_revenue = $18, actual_revenue = $19,
			revenue_variance = $20, total_refunds = $21, status = $22,
			error_count = $23, errors = $24, updated_at = NOW()
		WHERE id = $25
	`

	_, err := r.db.Exec(query,
		log.TotalOrders, log.TotalPayments, log.TotalAmount,
		log.OrdersPending, log.OrdersPaid, log.OrdersCancelled, log.OrdersRefunded,
		log.PaymentsPending, log.PaymentsSuccess, log.PaymentsFailed,
		log.MismatchesFound, log.MismatchesResolved, mismatchDetailsJSON,
		log.OrphanOrders, log.OrphanPayments, orphanDetailsJSON,
		log.StuckPayments, log.ExpectedRevenue, log.ActualRevenue,
		log.RevenueVariance, log.TotalRefunds, log.Status,
		log.ErrorCount, errorsJSON, log.ID,
	)
	return err
}

func (r *reconciliationRepository) MarkCompleted(id int) error {
	query := `UPDATE reconciliation_log SET status = 'COMPLETED', completed_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *reconciliationRepository) MarkFailed(id int, errors map[string]any) error {
	errorsJSON, _ := json.Marshal(errors)
	query := `UPDATE reconciliation_log SET status = 'FAILED', errors = $1, completed_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(query, errorsJSON, id)
	return err
}

func (r *reconciliationRepository) scanReconciliationLog(row *sql.Row) (*models.ReconciliationLog, error) {
	var log models.ReconciliationLog
	var mismatchDetailsJSON, orphanDetailsJSON, errorsJSON []byte

	err := row.Scan(
		&log.ID, &log.ReconciliationDate, &log.PeriodStart, &log.PeriodEnd,
		&log.TotalOrders, &log.TotalPayments, &log.TotalAmount,
		&log.OrdersPending, &log.OrdersPaid, &log.OrdersCancelled, &log.OrdersRefunded,
		&log.PaymentsPending, &log.PaymentsSuccess, &log.PaymentsFailed,
		&log.MismatchesFound, &log.MismatchesResolved, &mismatchDetailsJSON,
		&log.OrphanOrders, &log.OrphanPayments, &orphanDetailsJSON,
		&log.StuckPayments, &log.ExpectedRevenue, &log.ActualRevenue,
		&log.RevenueVariance, &log.TotalRefunds, &log.Status, &log.StartedAt, &log.CompletedAt,
		&log.RunBy, &log.ErrorCount, &errorsJSON, &log.CreatedAt, &log.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if len(mismatchDetailsJSON) > 0 {
		json.Unmarshal(mismatchDetailsJSON, &log.MismatchDetails)
	}
	if len(orphanDetailsJSON) > 0 {
		json.Unmarshal(orphanDetailsJSON, &log.OrphanDetails)
	}
	if len(errorsJSON) > 0 {
		json.Unmarshal(errorsJSON, &log.Errors)
	}

	return &log, nil
}

func (r *reconciliationRepository) scanReconciliationLogFromRows(rows *sql.Rows) (*models.ReconciliationLog, error) {
	var log models.ReconciliationLog
	var mismatchDetailsJSON, orphanDetailsJSON, errorsJSON []byte

	err := rows.Scan(
		&log.ID, &log.ReconciliationDate, &log.PeriodStart, &log.PeriodEnd,
		&log.TotalOrders, &log.TotalPayments, &log.TotalAmount,
		&log.OrdersPending, &log.OrdersPaid, &log.OrdersCancelled, &log.OrdersRefunded,
		&log.PaymentsPending, &log.PaymentsSuccess, &log.PaymentsFailed,
		&log.MismatchesFound, &log.MismatchesResolved, &mismatchDetailsJSON,
		&log.OrphanOrders, &log.OrphanPayments, &orphanDetailsJSON,
		&log.StuckPayments, &log.ExpectedRevenue, &log.ActualRevenue,
		&log.RevenueVariance, &log.TotalRefunds, &log.Status, &log.StartedAt, &log.CompletedAt,
		&log.RunBy, &log.ErrorCount, &errorsJSON, &log.CreatedAt, &log.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if len(mismatchDetailsJSON) > 0 {
		json.Unmarshal(mismatchDetailsJSON, &log.MismatchDetails)
	}
	if len(orphanDetailsJSON) > 0 {
		json.Unmarshal(orphanDetailsJSON, &log.OrphanDetails)
	}
	if len(errorsJSON) > 0 {
		json.Unmarshal(errorsJSON, &log.Errors)
	}

	return &log, nil
}
