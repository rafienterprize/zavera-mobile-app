package service

import (
	"database/sql"
	"fmt"
	"time"
	"zavera/dto"
)

type SystemHealthService interface {
	GetSystemHealth() (*dto.SystemHealth, error)
	GetCourierPerformance() ([]dto.CourierPerformance, error)
}

type systemHealthService struct {
	db *sql.DB
}

func NewSystemHealthService(db *sql.DB) SystemHealthService {
	return &systemHealthService{
		db: db,
	}
}

// GetSystemHealth returns system health metrics
func (s *systemHealthService) GetSystemHealth() (*dto.SystemHealth, error) {
	health := &dto.SystemHealth{
		BackgroundJobsHealthy: true,
	}

	// Webhook success rate (mock for now - would need webhook logs table)
	// In production, query webhook_logs table for last 24h
	health.WebhookSuccessRate = 98.5

	// Payment gateway latency (mock - would measure actual Midtrans response time)
	health.PaymentGatewayLatency = 245

	// Check background jobs health
	// Check if payment expiry job ran recently (should run every 5 minutes)
	var lastPaymentCheck time.Time
	err := s.db.QueryRow(`
		SELECT MAX(created_at) 
		FROM orders 
		WHERE status = 'EXPIRED' 
		AND created_at > NOW() - INTERVAL '1 hour'
	`).Scan(&lastPaymentCheck)
	
	if err != nil && err != sql.ErrNoRows {
		health.BackgroundJobsHealthy = false
	}

	// Get last tracking update
	var lastTracking time.Time
	err = s.db.QueryRow(`
		SELECT MAX(updated_at) 
		FROM shipments 
		WHERE updated_at > NOW() - INTERVAL '1 hour'
	`).Scan(&lastTracking)
	
	if err == nil {
		health.LastTrackingUpdate = lastTracking.Format(time.RFC3339)
	} else {
		health.LastTrackingUpdate = time.Now().Format(time.RFC3339)
	}

	// Check if tracking job is running (should update within last hour)
	if time.Since(lastTracking) > 2*time.Hour {
		health.BackgroundJobsHealthy = false
	}

	return health, nil
}

// GetCourierPerformance returns courier performance analytics
func (s *systemHealthService) GetCourierPerformance() ([]dto.CourierPerformance, error) {
	var performance []dto.CourierPerformance

	rows, err := s.db.Query(`
		SELECT 
			provider_code as courier_name,
			COUNT(CASE WHEN status = 'DELIVERED' THEN 1 END) as delivered,
			COUNT(CASE WHEN status IN ('DELIVERY_FAILED', 'LOST', 'RETURNED_TO_SENDER') THEN 1 END) as failed,
			COALESCE(AVG(EXTRACT(DAY FROM (delivered_at - shipped_at))), 0) as avg_delivery_days,
			CASE 
				WHEN COUNT(*) > 0 THEN 
					(COUNT(CASE WHEN status = 'DELIVERED' THEN 1 END)::float / COUNT(*)::float * 100)
				ELSE 0 
			END as success_rate
		FROM shipments
		WHERE created_at > NOW() - INTERVAL '30 days'
		AND status NOT IN ('PENDING', 'PROCESSING', 'CANCELLED')
		GROUP BY provider_code
		HAVING COUNT(*) >= 5
		ORDER BY success_rate DESC
	`)

	if err != nil {
		return nil, fmt.Errorf("failed to get courier performance: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var perf dto.CourierPerformance
		if err := rows.Scan(&perf.CourierName, &perf.Delivered, &perf.Failed, 
			&perf.AvgDeliveryDays, &perf.SuccessRate); err == nil {
			performance = append(performance, perf)
		}
	}

	return performance, nil
}
