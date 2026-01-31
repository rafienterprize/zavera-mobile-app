package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
	"zavera/models"

	"github.com/lib/pq"
)

type DisputeRepository interface {
	// Disputes
	Create(dispute *models.Dispute) error
	FindByID(id int) (*models.Dispute, error)
	FindByCode(code string) (*models.Dispute, error)
	FindByOrderID(orderID int) ([]*models.Dispute, error)
	FindByShipmentID(shipmentID int) ([]*models.Dispute, error)
	FindOpen() ([]*models.Dispute, error)
	FindByStatus(status models.DisputeStatus) ([]*models.Dispute, error)
	Update(dispute *models.Dispute) error
	UpdateStatus(id int, status models.DisputeStatus) error
	Resolve(id int, resolution models.DisputeStatus, notes string, amount *float64, resolvedBy int) error
	LinkRefund(disputeID, refundID int) error
	LinkReship(disputeID, shipmentID int) error

	// Messages
	AddMessage(msg *models.DisputeMessage) error
	GetMessages(disputeID int, includeInternal bool) ([]models.DisputeMessage, error)

	// Courier Failures
	LogCourierFailure(log *models.CourierFailureLog) error
	GetCourierFailures(shipmentID int) ([]models.CourierFailureLog, error)
	GetUnresolvedFailures() ([]models.CourierFailureLog, error)
	ResolveCourierFailure(id int, resolvedBy, action string) error

	// Shipment Status History
	RecordStatusChange(shipmentID int, from, to string, changedBy, reason string, metadata map[string]any) error
	GetStatusHistory(shipmentID int) ([]models.ShipmentStatusHistory, error)

	// Alerts
	CreateAlert(alert *models.ShipmentAlert) error
	GetAlertsByShipment(shipmentID int) ([]models.ShipmentAlert, error)
	GetUnresolvedAlerts() ([]models.ShipmentAlert, error)
	GetAlertsByLevel(level string) ([]models.ShipmentAlert, error)
	AcknowledgeAlert(id, userID int) error
	ResolveAlert(id, userID int, notes string) error

	GetDB() *sql.DB
}

type disputeRepository struct {
	db *sql.DB
}

func NewDisputeRepository(db *sql.DB) DisputeRepository {
	return &disputeRepository{db: db}
}

// ============================================
// DISPUTES
// ============================================

func (r *disputeRepository) Create(dispute *models.Dispute) error {
	metadataJSON, _ := json.Marshal(dispute.Metadata)

	query := `
		INSERT INTO disputes (
			dispute_code, order_id, shipment_id, dispute_type, status,
			title, description, customer_claim, customer_user_id,
			customer_email, customer_phone, evidence_urls,
			response_deadline, resolution_deadline, metadata
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, created_at, updated_at
	`

	// Set deadlines
	responseDeadline := time.Now().Add(48 * time.Hour)
	resolutionDeadline := time.Now().Add(7 * 24 * time.Hour)

	return r.db.QueryRow(
		query,
		dispute.DisputeCode, dispute.OrderID, dispute.ShipmentID,
		dispute.DisputeType, dispute.Status, dispute.Title, dispute.Description,
		dispute.CustomerClaim, dispute.CustomerUserID, dispute.CustomerEmail,
		dispute.CustomerPhone, pq.Array(dispute.EvidenceURLs),
		responseDeadline, resolutionDeadline, metadataJSON,
	).Scan(&dispute.ID, &dispute.CreatedAt, &dispute.UpdatedAt)
}

func (r *disputeRepository) FindByID(id int) (*models.Dispute, error) {
	query := `
		SELECT id, dispute_code, order_id, shipment_id, refund_id, dispute_type, status,
		       title, description, customer_claim, customer_user_id, customer_email,
		       customer_phone, evidence_urls, customer_evidence_urls, courier_evidence_urls,
		       investigation_notes, investigation_started_at, investigation_completed_at,
		       investigator_id, resolution, resolution_notes, resolution_amount,
		       resolved_by, resolved_at, reship_shipment_id, response_deadline,
		       resolution_deadline, metadata, created_at, updated_at
		FROM disputes WHERE id = $1
	`
	return r.scanDispute(r.db.QueryRow(query, id))
}

func (r *disputeRepository) FindByCode(code string) (*models.Dispute, error) {
	query := `
		SELECT id, dispute_code, order_id, shipment_id, refund_id, dispute_type, status,
		       title, description, customer_claim, customer_user_id, customer_email,
		       customer_phone, evidence_urls, customer_evidence_urls, courier_evidence_urls,
		       investigation_notes, investigation_started_at, investigation_completed_at,
		       investigator_id, resolution, resolution_notes, resolution_amount,
		       resolved_by, resolved_at, reship_shipment_id, response_deadline,
		       resolution_deadline, metadata, created_at, updated_at
		FROM disputes WHERE dispute_code = $1
	`
	return r.scanDispute(r.db.QueryRow(query, code))
}

func (r *disputeRepository) FindByOrderID(orderID int) ([]*models.Dispute, error) {
	query := `
		SELECT id, dispute_code, order_id, shipment_id, refund_id, dispute_type, status,
		       title, description, customer_claim, customer_user_id, customer_email,
		       customer_phone, evidence_urls, customer_evidence_urls, courier_evidence_urls,
		       investigation_notes, investigation_started_at, investigation_completed_at,
		       investigator_id, resolution, resolution_notes, resolution_amount,
		       resolved_by, resolved_at, reship_shipment_id, response_deadline,
		       resolution_deadline, metadata, created_at, updated_at
		FROM disputes WHERE order_id = $1 ORDER BY created_at DESC
	`
	return r.queryDisputes(query, orderID)
}

func (r *disputeRepository) FindByShipmentID(shipmentID int) ([]*models.Dispute, error) {
	query := `
		SELECT id, dispute_code, order_id, shipment_id, refund_id, dispute_type, status,
		       title, description, customer_claim, customer_user_id, customer_email,
		       customer_phone, evidence_urls, customer_evidence_urls, courier_evidence_urls,
		       investigation_notes, investigation_started_at, investigation_completed_at,
		       investigator_id, resolution, resolution_notes, resolution_amount,
		       resolved_by, resolved_at, reship_shipment_id, response_deadline,
		       resolution_deadline, metadata, created_at, updated_at
		FROM disputes WHERE shipment_id = $1 ORDER BY created_at DESC
	`
	return r.queryDisputes(query, shipmentID)
}

func (r *disputeRepository) FindOpen() ([]*models.Dispute, error) {
	query := `
		SELECT id, dispute_code, order_id, shipment_id, refund_id, dispute_type, status,
		       title, description, customer_claim, customer_user_id, customer_email,
		       customer_phone, evidence_urls, customer_evidence_urls, courier_evidence_urls,
		       investigation_notes, investigation_started_at, investigation_completed_at,
		       investigator_id, resolution, resolution_notes, resolution_amount,
		       resolved_by, resolved_at, reship_shipment_id, response_deadline,
		       resolution_deadline, metadata, created_at, updated_at
		FROM disputes 
		WHERE status IN ('OPEN', 'INVESTIGATING', 'EVIDENCE_REQUIRED', 'PENDING_RESOLUTION')
		ORDER BY created_at ASC
	`
	return r.queryDisputes(query)
}

func (r *disputeRepository) FindByStatus(status models.DisputeStatus) ([]*models.Dispute, error) {
	query := `
		SELECT id, dispute_code, order_id, shipment_id, refund_id, dispute_type, status,
		       title, description, customer_claim, customer_user_id, customer_email,
		       customer_phone, evidence_urls, customer_evidence_urls, courier_evidence_urls,
		       investigation_notes, investigation_started_at, investigation_completed_at,
		       investigator_id, resolution, resolution_notes, resolution_amount,
		       resolved_by, resolved_at, reship_shipment_id, response_deadline,
		       resolution_deadline, metadata, created_at, updated_at
		FROM disputes WHERE status = $1 ORDER BY created_at DESC
	`
	return r.queryDisputes(query, status)
}

func (r *disputeRepository) scanDispute(row *sql.Row) (*models.Dispute, error) {
	var d models.Dispute
	var metadataJSON []byte
	var evidenceURLs, customerEvidenceURLs, courierEvidenceURLs pq.StringArray

	err := row.Scan(
		&d.ID, &d.DisputeCode, &d.OrderID, &d.ShipmentID, &d.RefundID,
		&d.DisputeType, &d.Status, &d.Title, &d.Description, &d.CustomerClaim,
		&d.CustomerUserID, &d.CustomerEmail, &d.CustomerPhone,
		&evidenceURLs, &customerEvidenceURLs, &courierEvidenceURLs,
		&d.InvestigationNotes, &d.InvestigationStartedAt, &d.InvestigationCompletedAt,
		&d.InvestigatorID, &d.Resolution, &d.ResolutionNotes, &d.ResolutionAmount,
		&d.ResolvedBy, &d.ResolvedAt, &d.ReshipShipmentID, &d.ResponseDeadline,
		&d.ResolutionDeadline, &metadataJSON, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	d.EvidenceURLs = evidenceURLs
	d.CustomerEvidenceURLs = customerEvidenceURLs
	d.CourierEvidenceURLs = courierEvidenceURLs

	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &d.Metadata)
	}

	return &d, nil
}

func (r *disputeRepository) queryDisputes(query string, args ...any) ([]*models.Dispute, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var disputes []*models.Dispute
	for rows.Next() {
		var d models.Dispute
		var metadataJSON []byte
		var evidenceURLs, customerEvidenceURLs, courierEvidenceURLs pq.StringArray

		err := rows.Scan(
			&d.ID, &d.DisputeCode, &d.OrderID, &d.ShipmentID, &d.RefundID,
			&d.DisputeType, &d.Status, &d.Title, &d.Description, &d.CustomerClaim,
			&d.CustomerUserID, &d.CustomerEmail, &d.CustomerPhone,
			&evidenceURLs, &customerEvidenceURLs, &courierEvidenceURLs,
			&d.InvestigationNotes, &d.InvestigationStartedAt, &d.InvestigationCompletedAt,
			&d.InvestigatorID, &d.Resolution, &d.ResolutionNotes, &d.ResolutionAmount,
			&d.ResolvedBy, &d.ResolvedAt, &d.ReshipShipmentID, &d.ResponseDeadline,
			&d.ResolutionDeadline, &metadataJSON, &d.CreatedAt, &d.UpdatedAt,
		)
		if err != nil {
			continue
		}

		d.EvidenceURLs = evidenceURLs
		d.CustomerEvidenceURLs = customerEvidenceURLs
		d.CourierEvidenceURLs = courierEvidenceURLs

		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &d.Metadata)
		}

		disputes = append(disputes, &d)
	}

	return disputes, nil
}

func (r *disputeRepository) Update(dispute *models.Dispute) error {
	metadataJSON, _ := json.Marshal(dispute.Metadata)

	query := `
		UPDATE disputes SET
			status = $1, investigation_notes = $2, evidence_urls = $3,
			customer_evidence_urls = $4, courier_evidence_urls = $5,
			metadata = $6, updated_at = NOW()
		WHERE id = $7
	`

	_, err := r.db.Exec(query,
		dispute.Status, dispute.InvestigationNotes,
		pq.Array(dispute.EvidenceURLs), pq.Array(dispute.CustomerEvidenceURLs),
		pq.Array(dispute.CourierEvidenceURLs), metadataJSON, dispute.ID,
	)
	return err
}

func (r *disputeRepository) UpdateStatus(id int, status models.DisputeStatus) error {
	query := `UPDATE disputes SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(query, status, id)
	return err
}

func (r *disputeRepository) Resolve(id int, resolution models.DisputeStatus, notes string, amount *float64, resolvedBy int) error {
	query := `
		UPDATE disputes SET
			status = $1, resolution = $1, resolution_notes = $2,
			resolution_amount = $3, resolved_by = $4, resolved_at = NOW(),
			updated_at = NOW()
		WHERE id = $5
	`
	_, err := r.db.Exec(query, resolution, notes, amount, resolvedBy, id)
	return err
}

func (r *disputeRepository) LinkRefund(disputeID, refundID int) error {
	query := `UPDATE disputes SET refund_id = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(query, refundID, disputeID)
	return err
}

func (r *disputeRepository) LinkReship(disputeID, shipmentID int) error {
	query := `UPDATE disputes SET reship_shipment_id = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(query, shipmentID, disputeID)
	return err
}

// ============================================
// MESSAGES
// ============================================

func (r *disputeRepository) AddMessage(msg *models.DisputeMessage) error {
	query := `
		INSERT INTO dispute_messages (dispute_id, sender_type, sender_id, sender_name, message, attachment_urls, is_internal)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`
	return r.db.QueryRow(
		query,
		msg.DisputeID, msg.SenderType, msg.SenderID, msg.SenderName,
		msg.Message, pq.Array(msg.AttachmentURLs), msg.IsInternal,
	).Scan(&msg.ID, &msg.CreatedAt)
}

func (r *disputeRepository) GetMessages(disputeID int, includeInternal bool) ([]models.DisputeMessage, error) {
	query := `
		SELECT id, dispute_id, sender_type, sender_id, sender_name, message, attachment_urls, is_internal, created_at
		FROM dispute_messages
		WHERE dispute_id = $1
	`
	if !includeInternal {
		query += " AND is_internal = false"
	}
	query += " ORDER BY created_at ASC"

	rows, err := r.db.Query(query, disputeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.DisputeMessage
	for rows.Next() {
		var m models.DisputeMessage
		var attachmentURLs pq.StringArray

		err := rows.Scan(
			&m.ID, &m.DisputeID, &m.SenderType, &m.SenderID, &m.SenderName,
			&m.Message, &attachmentURLs, &m.IsInternal, &m.CreatedAt,
		)
		if err != nil {
			continue
		}
		m.AttachmentURLs = attachmentURLs
		messages = append(messages, m)
	}

	return messages, nil
}

// ============================================
// COURIER FAILURES
// ============================================

func (r *disputeRepository) LogCourierFailure(log *models.CourierFailureLog) error {
	query := `
		INSERT INTO courier_failure_log (
			shipment_id, failure_type, failure_reason, courier_code,
			courier_name, courier_tracking, failure_location, evidence_urls, notes
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at
	`
	return r.db.QueryRow(
		query,
		log.ShipmentID, log.FailureType, log.FailureReason, log.CourierCode,
		log.CourierName, log.CourierTracking, log.FailureLocation,
		pq.Array(log.EvidenceURLs), log.Notes,
	).Scan(&log.ID, &log.CreatedAt)
}

func (r *disputeRepository) GetCourierFailures(shipmentID int) ([]models.CourierFailureLog, error) {
	query := `
		SELECT id, shipment_id, failure_type, failure_reason, failure_time,
		       courier_code, courier_name, courier_tracking, failure_location,
		       resolved, resolved_at, resolved_by, resolution_action,
		       evidence_urls, notes, created_at
		FROM courier_failure_log WHERE shipment_id = $1 ORDER BY created_at DESC
	`
	return r.queryCourierFailures(query, shipmentID)
}

func (r *disputeRepository) GetUnresolvedFailures() ([]models.CourierFailureLog, error) {
	query := `
		SELECT id, shipment_id, failure_type, failure_reason, failure_time,
		       courier_code, courier_name, courier_tracking, failure_location,
		       resolved, resolved_at, resolved_by, resolution_action,
		       evidence_urls, notes, created_at
		FROM courier_failure_log WHERE resolved = false ORDER BY created_at ASC
	`
	return r.queryCourierFailures(query)
}

func (r *disputeRepository) queryCourierFailures(query string, args ...any) ([]models.CourierFailureLog, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var failures []models.CourierFailureLog
	for rows.Next() {
		var f models.CourierFailureLog
		var evidenceURLs pq.StringArray

		err := rows.Scan(
			&f.ID, &f.ShipmentID, &f.FailureType, &f.FailureReason, &f.FailureTime,
			&f.CourierCode, &f.CourierName, &f.CourierTracking, &f.FailureLocation,
			&f.Resolved, &f.ResolvedAt, &f.ResolvedBy, &f.ResolutionAction,
			&evidenceURLs, &f.Notes, &f.CreatedAt,
		)
		if err != nil {
			continue
		}
		f.EvidenceURLs = evidenceURLs
		failures = append(failures, f)
	}

	return failures, nil
}

func (r *disputeRepository) ResolveCourierFailure(id int, resolvedBy, action string) error {
	query := `
		UPDATE courier_failure_log 
		SET resolved = true, resolved_at = NOW(), resolved_by = $1, resolution_action = $2
		WHERE id = $3
	`
	_, err := r.db.Exec(query, resolvedBy, action, id)
	return err
}

// ============================================
// STATUS HISTORY
// ============================================

func (r *disputeRepository) RecordStatusChange(shipmentID int, from, to string, changedBy, reason string, metadata map[string]any) error {
	metadataJSON, _ := json.Marshal(metadata)

	query := `
		INSERT INTO shipment_status_history (shipment_id, from_status, to_status, changed_by, reason, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(query, shipmentID, from, to, changedBy, reason, metadataJSON)
	return err
}

func (r *disputeRepository) GetStatusHistory(shipmentID int) ([]models.ShipmentStatusHistory, error) {
	query := `
		SELECT id, shipment_id, from_status, to_status, changed_by, reason, metadata, created_at
		FROM shipment_status_history WHERE shipment_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query, shipmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []models.ShipmentStatusHistory
	for rows.Next() {
		var h models.ShipmentStatusHistory
		var metadataJSON []byte

		err := rows.Scan(
			&h.ID, &h.ShipmentID, &h.FromStatus, &h.ToStatus,
			&h.ChangedBy, &h.Reason, &metadataJSON, &h.CreatedAt,
		)
		if err != nil {
			continue
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &h.Metadata)
		}
		history = append(history, h)
	}

	return history, nil
}

// ============================================
// ALERTS
// ============================================

func (r *disputeRepository) CreateAlert(alert *models.ShipmentAlert) error {
	query := `
		INSERT INTO shipment_alerts (shipment_id, alert_type, alert_level, title, description)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	return r.db.QueryRow(
		query,
		alert.ShipmentID, alert.AlertType, alert.AlertLevel, alert.Title, alert.Description,
	).Scan(&alert.ID, &alert.CreatedAt)
}

func (r *disputeRepository) GetAlertsByShipment(shipmentID int) ([]models.ShipmentAlert, error) {
	query := `
		SELECT id, shipment_id, alert_type, alert_level, title, description,
		       acknowledged, acknowledged_by, acknowledged_at, resolved, resolved_by,
		       resolved_at, resolution_notes, auto_action_taken, auto_action_type,
		       auto_action_at, created_at
		FROM shipment_alerts WHERE shipment_id = $1 ORDER BY created_at DESC
	`
	return r.queryAlerts(query, shipmentID)
}

func (r *disputeRepository) GetUnresolvedAlerts() ([]models.ShipmentAlert, error) {
	query := `
		SELECT id, shipment_id, alert_type, alert_level, title, description,
		       acknowledged, acknowledged_by, acknowledged_at, resolved, resolved_by,
		       resolved_at, resolution_notes, auto_action_taken, auto_action_type,
		       auto_action_at, created_at
		FROM shipment_alerts WHERE resolved = false ORDER BY 
			CASE alert_level WHEN 'urgent' THEN 1 WHEN 'critical' THEN 2 ELSE 3 END,
			created_at ASC
	`
	return r.queryAlerts(query)
}

func (r *disputeRepository) GetAlertsByLevel(level string) ([]models.ShipmentAlert, error) {
	query := `
		SELECT id, shipment_id, alert_type, alert_level, title, description,
		       acknowledged, acknowledged_by, acknowledged_at, resolved, resolved_by,
		       resolved_at, resolution_notes, auto_action_taken, auto_action_type,
		       auto_action_at, created_at
		FROM shipment_alerts WHERE alert_level = $1 AND resolved = false ORDER BY created_at ASC
	`
	return r.queryAlerts(query, level)
}

func (r *disputeRepository) queryAlerts(query string, args ...any) ([]models.ShipmentAlert, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []models.ShipmentAlert
	for rows.Next() {
		var a models.ShipmentAlert
		err := rows.Scan(
			&a.ID, &a.ShipmentID, &a.AlertType, &a.AlertLevel, &a.Title, &a.Description,
			&a.Acknowledged, &a.AcknowledgedBy, &a.AcknowledgedAt, &a.Resolved, &a.ResolvedBy,
			&a.ResolvedAt, &a.ResolutionNotes, &a.AutoActionTaken, &a.AutoActionType,
			&a.AutoActionAt, &a.CreatedAt,
		)
		if err != nil {
			continue
		}
		alerts = append(alerts, a)
	}

	return alerts, nil
}

func (r *disputeRepository) AcknowledgeAlert(id, userID int) error {
	query := `UPDATE shipment_alerts SET acknowledged = true, acknowledged_by = $1, acknowledged_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(query, userID, id)
	return err
}

func (r *disputeRepository) ResolveAlert(id, userID int, notes string) error {
	query := `UPDATE shipment_alerts SET resolved = true, resolved_by = $1, resolved_at = NOW(), resolution_notes = $2 WHERE id = $3`
	_, err := r.db.Exec(query, userID, notes, id)
	return err
}

func (r *disputeRepository) GetDB() *sql.DB {
	return r.db
}

// GenerateDisputeCode generates a unique dispute code
func GenerateDisputeCode() string {
	return fmt.Sprintf("DSP-%s-%s", time.Now().Format("20060102"), randomHex(4))
}
