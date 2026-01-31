package repository

import (
	"database/sql"
	"zavera/models"
)

// EmailRepository handles email template and log operations
type EmailRepository interface {
	// Templates
	GetTemplateByKey(key string) (*models.EmailTemplate, error)
	GetAllTemplates() ([]models.EmailTemplate, error)
	UpdateTemplate(template *models.EmailTemplate) error
	
	// Logs
	CreateEmailLog(log *models.EmailLog) error
	GetEmailLogsByOrder(orderID int) ([]models.EmailLog, error)
	GetEmailLogsByUser(userID int) ([]models.EmailLog, error)
	GetPendingEmailLogs() ([]models.EmailLog, error)
	UpdateEmailLogStatus(logID int, status models.EmailLogStatus, errorMessage string) error
	
	// Duplicate check
	HasSentEmail(orderID int, templateKey string) (bool, error)
}

type emailRepository struct {
	db *sql.DB
}

// NewEmailRepository creates a new email repository
func NewEmailRepository(db *sql.DB) EmailRepository {
	return &emailRepository{db: db}
}

// GetTemplateByKey retrieves an email template by its key
func (r *emailRepository) GetTemplateByKey(key string) (*models.EmailTemplate, error) {
	query := `
		SELECT id, template_key, name, subject_template, html_template, is_active, created_at, updated_at
		FROM email_templates
		WHERE template_key = $1 AND is_active = true
	`

	var template models.EmailTemplate
	err := r.db.QueryRow(query, key).Scan(
		&template.ID, &template.TemplateKey, &template.Name,
		&template.SubjectTemplate, &template.HTMLTemplate,
		&template.IsActive, &template.CreatedAt, &template.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &template, nil
}

// GetAllTemplates retrieves all email templates
func (r *emailRepository) GetAllTemplates() ([]models.EmailTemplate, error) {
	query := `
		SELECT id, template_key, name, subject_template, html_template, is_active, created_at, updated_at
		FROM email_templates
		ORDER BY template_key
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []models.EmailTemplate
	for rows.Next() {
		var t models.EmailTemplate
		err := rows.Scan(
			&t.ID, &t.TemplateKey, &t.Name,
			&t.SubjectTemplate, &t.HTMLTemplate,
			&t.IsActive, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			continue
		}
		templates = append(templates, t)
	}

	return templates, nil
}

// UpdateTemplate updates an email template
func (r *emailRepository) UpdateTemplate(template *models.EmailTemplate) error {
	query := `
		UPDATE email_templates
		SET name = $1, subject_template = $2, html_template = $3, is_active = $4, updated_at = NOW()
		WHERE id = $5
	`

	_, err := r.db.Exec(query, template.Name, template.SubjectTemplate, template.HTMLTemplate, template.IsActive, template.ID)
	return err
}

// CreateEmailLog creates a new email log entry
func (r *emailRepository) CreateEmailLog(log *models.EmailLog) error {
	query := `
		INSERT INTO email_logs (order_id, user_id, template_key, recipient_email, subject, status, error_message, sent_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(
		query,
		log.OrderID, log.UserID, log.TemplateKey, log.RecipientEmail,
		log.Subject, log.Status, log.ErrorMessage, log.SentAt,
	).Scan(&log.ID, &log.CreatedAt)

	return err
}

// GetEmailLogsByOrder retrieves email logs for an order
func (r *emailRepository) GetEmailLogsByOrder(orderID int) ([]models.EmailLog, error) {
	query := `
		SELECT id, order_id, user_id, template_key, recipient_email, subject, status, error_message, sent_at, created_at
		FROM email_logs
		WHERE order_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.EmailLog
	for rows.Next() {
		var l models.EmailLog
		err := rows.Scan(
			&l.ID, &l.OrderID, &l.UserID, &l.TemplateKey, &l.RecipientEmail,
			&l.Subject, &l.Status, &l.ErrorMessage, &l.SentAt, &l.CreatedAt,
		)
		if err != nil {
			continue
		}
		logs = append(logs, l)
	}

	return logs, nil
}

// GetEmailLogsByUser retrieves email logs for a user
func (r *emailRepository) GetEmailLogsByUser(userID int) ([]models.EmailLog, error) {
	query := `
		SELECT id, order_id, user_id, template_key, recipient_email, subject, status, error_message, sent_at, created_at
		FROM email_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.EmailLog
	for rows.Next() {
		var l models.EmailLog
		err := rows.Scan(
			&l.ID, &l.OrderID, &l.UserID, &l.TemplateKey, &l.RecipientEmail,
			&l.Subject, &l.Status, &l.ErrorMessage, &l.SentAt, &l.CreatedAt,
		)
		if err != nil {
			continue
		}
		logs = append(logs, l)
	}

	return logs, nil
}

// GetPendingEmailLogs retrieves pending email logs for retry
func (r *emailRepository) GetPendingEmailLogs() ([]models.EmailLog, error) {
	query := `
		SELECT id, order_id, user_id, template_key, recipient_email, subject, status, error_message, sent_at, created_at
		FROM email_logs
		WHERE status IN ('PENDING', 'RETRY')
		ORDER BY created_at ASC
		LIMIT 100
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.EmailLog
	for rows.Next() {
		var l models.EmailLog
		err := rows.Scan(
			&l.ID, &l.OrderID, &l.UserID, &l.TemplateKey, &l.RecipientEmail,
			&l.Subject, &l.Status, &l.ErrorMessage, &l.SentAt, &l.CreatedAt,
		)
		if err != nil {
			continue
		}
		logs = append(logs, l)
	}

	return logs, nil
}

// UpdateEmailLogStatus updates the status of an email log
func (r *emailRepository) UpdateEmailLogStatus(logID int, status models.EmailLogStatus, errorMessage string) error {
	query := `
		UPDATE email_logs
		SET status = $1, error_message = $2, sent_at = CASE WHEN $1 = 'SENT' THEN NOW() ELSE sent_at END
		WHERE id = $3
	`

	_, err := r.db.Exec(query, status, errorMessage, logID)
	return err
}

// HasSentEmail checks if an email type has already been sent for an order
func (r *emailRepository) HasSentEmail(orderID int, templateKey string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM email_logs 
			WHERE order_id = $1 AND template_key = $2 AND status = 'SENT'
		)
	`

	var exists bool
	err := r.db.QueryRow(query, orderID, templateKey).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
