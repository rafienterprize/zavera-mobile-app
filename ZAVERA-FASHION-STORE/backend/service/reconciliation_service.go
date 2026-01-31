package service

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"
	"zavera/dto"
	"zavera/models"
	"zavera/repository"
)

// ReconciliationService handles daily reconciliation
type ReconciliationService interface {
	// Run reconciliation
	RunDailyReconciliation(date time.Time, runBy string) (*dto.ReconciliationSummary, error)
	RunReconciliationForDateRange(startDate, endDate time.Time, runBy string) ([]*dto.ReconciliationSummary, error)
	
	// Get results
	GetReconciliationSummary(date time.Time) (*dto.ReconciliationSummary, error)
	GetRecentReconciliations(limit int) ([]*models.ReconciliationLog, error)
	
	// Mismatch handling
	GetUnresolvedMismatches() ([]dto.MismatchDetail, error)
	ResolveMismatch(syncLogID int, resolvedBy int, action string) error
	
	// Cron job
	StartReconciliationJob(hour int) // Run daily at specified hour
	StopReconciliationJob()
}

type reconciliationService struct {
	reconciliationRepo repository.ReconciliationRepository
	syncRepo           repository.PaymentSyncRepository
	db                 *sql.DB
	
	// Job control
	jobRunning bool
	jobStop    chan struct{}
	jobMutex   sync.Mutex
}

func NewReconciliationService(
	reconciliationRepo repository.ReconciliationRepository,
	syncRepo repository.PaymentSyncRepository,
	db *sql.DB,
) ReconciliationService {
	return &reconciliationService{
		reconciliationRepo: reconciliationRepo,
		syncRepo:           syncRepo,
		db:                 db,
		jobStop:            make(chan struct{}),
	}
}

// RunDailyReconciliation runs reconciliation for a specific date
func (s *reconciliationService) RunDailyReconciliation(date time.Time, runBy string) (*dto.ReconciliationSummary, error) {
	log.Printf("📊 Starting reconciliation for %s by %s", date.Format("2006-01-02"), runBy)

	// Define period
	periodStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	periodEnd := periodStart.Add(24 * time.Hour)

	// Create or update reconciliation log
	reconLog := &models.ReconciliationLog{
		ReconciliationDate: periodStart,
		PeriodStart:        periodStart,
		PeriodEnd:          periodEnd,
		Status:             "RUNNING",
		RunBy:              runBy,
	}

	// Gather statistics
	var err error

	// 1. Order statistics
	reconLog.TotalOrders, reconLog.OrdersPending, reconLog.OrdersPaid, 
		reconLog.OrdersCancelled, reconLog.OrdersRefunded, err = s.getOrderStats(periodStart, periodEnd)
	if err != nil {
		log.Printf("⚠️ Error getting order stats: %v", err)
	}

	// 2. Payment statistics
	reconLog.TotalPayments, reconLog.PaymentsPending, reconLog.PaymentsSuccess,
		reconLog.PaymentsFailed, reconLog.TotalAmount, err = s.getPaymentStats(periodStart, periodEnd)
	if err != nil {
		log.Printf("⚠️ Error getting payment stats: %v", err)
	}

	// 3. Find mismatches
	mismatches, err := s.findMismatches(periodStart, periodEnd)
	if err != nil {
		log.Printf("⚠️ Error finding mismatches: %v", err)
	}
	reconLog.MismatchesFound = len(mismatches)
	reconLog.MismatchDetails = map[string]any{"mismatches": mismatches}

	// 4. Find orphans
	orphanOrders, orphanPayments, orphanDetails := s.findOrphans(periodStart, periodEnd)
	reconLog.OrphanOrders = orphanOrders
	reconLog.OrphanPayments = orphanPayments
	reconLog.OrphanDetails = orphanDetails

	// 5. Find stuck payments
	stuckPayments, stuckIDs := s.findStuckPayments(periodStart, periodEnd)
	reconLog.StuckPayments = stuckPayments
	reconLog.StuckPaymentIDs = stuckIDs

	// 6. Calculate revenue
	reconLog.ExpectedRevenue, reconLog.ActualRevenue, reconLog.RevenueVariance = s.calculateRevenue(periodStart, periodEnd)

	// 7. Get refund totals
	reconLog.TotalRefunds = s.getRefundTotal(periodStart, periodEnd)

	// Save reconciliation log
	if err := s.reconciliationRepo.Create(reconLog); err != nil {
		log.Printf("❌ Error saving reconciliation log: %v", err)
		return nil, err
	}

	// Mark as completed
	s.reconciliationRepo.MarkCompleted(reconLog.ID)

	log.Printf("✅ Reconciliation completed for %s", date.Format("2006-01-02"))

	// Build summary response
	summary := &dto.ReconciliationSummary{
		Date:               date.Format("2006-01-02"),
		TotalOrders:        reconLog.TotalOrders,
		TotalPayments:      reconLog.TotalPayments,
		TotalAmount:        reconLog.TotalAmount,
		OrdersByStatus: map[string]int{
			"pending":   reconLog.OrdersPending,
			"paid":      reconLog.OrdersPaid,
			"cancelled": reconLog.OrdersCancelled,
			"refunded":  reconLog.OrdersRefunded,
		},
		PaymentsByStatus: map[string]int{
			"pending": reconLog.PaymentsPending,
			"success": reconLog.PaymentsSuccess,
			"failed":  reconLog.PaymentsFailed,
		},
		MismatchesFound:    reconLog.MismatchesFound,
		MismatchesResolved: reconLog.MismatchesResolved,
		OrphanOrders:       reconLog.OrphanOrders,
		OrphanPayments:     reconLog.OrphanPayments,
		StuckPayments:      reconLog.StuckPayments,
		ExpectedRevenue:    reconLog.ExpectedRevenue,
		ActualRevenue:      reconLog.ActualRevenue,
		RevenueVariance:    reconLog.RevenueVariance,
		TotalRefunds:       reconLog.TotalRefunds,
		Status:             "COMPLETED",
		Mismatches:         mismatches,
	}

	return summary, nil
}

// RunReconciliationForDateRange runs reconciliation for a date range
func (s *reconciliationService) RunReconciliationForDateRange(startDate, endDate time.Time, runBy string) ([]*dto.ReconciliationSummary, error) {
	var summaries []*dto.ReconciliationSummary

	current := startDate
	for !current.After(endDate) {
		summary, err := s.RunDailyReconciliation(current, runBy)
		if err != nil {
			log.Printf("⚠️ Reconciliation failed for %s: %v", current.Format("2006-01-02"), err)
		} else {
			summaries = append(summaries, summary)
		}
		current = current.Add(24 * time.Hour)
	}

	return summaries, nil
}

// GetReconciliationSummary gets reconciliation summary for a date
func (s *reconciliationService) GetReconciliationSummary(date time.Time) (*dto.ReconciliationSummary, error) {
	reconLog, err := s.reconciliationRepo.FindByDate(date)
	if err != nil {
		return nil, err
	}

	// Extract mismatches from details
	var mismatches []dto.MismatchDetail
	if reconLog.MismatchDetails != nil {
		if m, ok := reconLog.MismatchDetails["mismatches"].([]dto.MismatchDetail); ok {
			mismatches = m
		}
	}

	summary := &dto.ReconciliationSummary{
		Date:               reconLog.ReconciliationDate.Format("2006-01-02"),
		TotalOrders:        reconLog.TotalOrders,
		TotalPayments:      reconLog.TotalPayments,
		TotalAmount:        reconLog.TotalAmount,
		OrdersByStatus: map[string]int{
			"pending":   reconLog.OrdersPending,
			"paid":      reconLog.OrdersPaid,
			"cancelled": reconLog.OrdersCancelled,
			"refunded":  reconLog.OrdersRefunded,
		},
		PaymentsByStatus: map[string]int{
			"pending": reconLog.PaymentsPending,
			"success": reconLog.PaymentsSuccess,
			"failed":  reconLog.PaymentsFailed,
		},
		MismatchesFound:    reconLog.MismatchesFound,
		MismatchesResolved: reconLog.MismatchesResolved,
		OrphanOrders:       reconLog.OrphanOrders,
		OrphanPayments:     reconLog.OrphanPayments,
		StuckPayments:      reconLog.StuckPayments,
		ExpectedRevenue:    reconLog.ExpectedRevenue,
		ActualRevenue:      reconLog.ActualRevenue,
		RevenueVariance:    reconLog.RevenueVariance,
		TotalRefunds:       reconLog.TotalRefunds,
		Status:             reconLog.Status,
		Mismatches:         mismatches,
	}

	return summary, nil
}

// GetRecentReconciliations gets recent reconciliation logs
func (s *reconciliationService) GetRecentReconciliations(limit int) ([]*models.ReconciliationLog, error) {
	return s.reconciliationRepo.FindRecent(limit)
}

// GetUnresolvedMismatches gets all unresolved mismatches
func (s *reconciliationService) GetUnresolvedMismatches() ([]dto.MismatchDetail, error) {
	logs, err := s.syncRepo.FindUnresolved()
	if err != nil {
		return nil, err
	}

	var mismatches []dto.MismatchDetail
	for _, log := range logs {
		mismatches = append(mismatches, dto.MismatchDetail{
			OrderCode:     log.OrderCode,
			PaymentID:     log.PaymentID,
			LocalStatus:   log.LocalPaymentStatus,
			GatewayStatus: log.GatewayStatus,
			MismatchType:  log.MismatchType,
		})
	}

	return mismatches, nil
}

// ResolveMismatch manually resolves a mismatch
func (s *reconciliationService) ResolveMismatch(syncLogID int, resolvedBy int, action string) error {
	return s.syncRepo.MarkResolved(syncLogID, resolvedBy, action)
}

// StartReconciliationJob starts the daily reconciliation job
func (s *reconciliationService) StartReconciliationJob(hour int) {
	s.jobMutex.Lock()
	if s.jobRunning {
		s.jobMutex.Unlock()
		return
	}
	s.jobRunning = true
	s.jobStop = make(chan struct{})
	s.jobMutex.Unlock()

	go func() {
		log.Printf("📅 Reconciliation job started (runs daily at %02d:00)", hour)

		for {
			// Calculate time until next run
			now := time.Now()
			nextRun := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, now.Location())
			if now.After(nextRun) {
				nextRun = nextRun.Add(24 * time.Hour)
			}
			duration := nextRun.Sub(now)

			select {
			case <-time.After(duration):
				// Run reconciliation for yesterday
				yesterday := time.Now().Add(-24 * time.Hour)
				s.RunDailyReconciliation(yesterday, "cron")
			case <-s.jobStop:
				log.Println("🛑 Reconciliation job stopped")
				return
			}
		}
	}()
}

// StopReconciliationJob stops the reconciliation job
func (s *reconciliationService) StopReconciliationJob() {
	s.jobMutex.Lock()
	defer s.jobMutex.Unlock()

	if s.jobRunning {
		close(s.jobStop)
		s.jobRunning = false
	}
}

// Helper methods
func (s *reconciliationService) getOrderStats(start, end time.Time) (total, pending, paid, cancelled, refunded int, err error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'PENDING') as pending,
			COUNT(*) FILTER (WHERE status = 'PAID') as paid,
			COUNT(*) FILTER (WHERE status = 'CANCELLED') as cancelled,
			COUNT(*) FILTER (WHERE refund_status IS NOT NULL) as refunded
		FROM orders
		WHERE created_at >= $1 AND created_at < $2
	`
	err = s.db.QueryRow(query, start, end).Scan(&total, &pending, &paid, &cancelled, &refunded)
	return
}

func (s *reconciliationService) getPaymentStats(start, end time.Time) (total, pending, success, failed int, amount float64, err error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'PENDING') as pending,
			COUNT(*) FILTER (WHERE status = 'SUCCESS') as success,
			COUNT(*) FILTER (WHERE status = 'FAILED') as failed,
			COALESCE(SUM(amount) FILTER (WHERE status = 'SUCCESS'), 0) as amount
		FROM payments
		WHERE created_at >= $1 AND created_at < $2
	`
	err = s.db.QueryRow(query, start, end).Scan(&total, &pending, &success, &failed, &amount)
	return
}

func (s *reconciliationService) findMismatches(start, end time.Time) ([]dto.MismatchDetail, error) {
	// Find orders where payment status doesn't match order status
	query := `
		SELECT o.order_code, p.id as payment_id, p.status as payment_status, 
		       o.status as order_status, p.amount
		FROM orders o
		JOIN payments p ON o.id = p.order_id
		WHERE o.created_at >= $1 AND o.created_at < $2
		AND (
			(p.status = 'SUCCESS' AND o.status NOT IN ('PAID', 'PROCESSING', 'SHIPPED', 'DELIVERED', 'COMPLETED'))
			OR (p.status = 'PENDING' AND o.status NOT IN ('PENDING'))
			OR (p.status IN ('FAILED', 'EXPIRED', 'CANCELLED') AND o.status NOT IN ('FAILED', 'EXPIRED', 'CANCELLED'))
		)
	`

	rows, err := s.db.Query(query, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mismatches []dto.MismatchDetail
	for rows.Next() {
		var m dto.MismatchDetail
		var paymentStatus, orderStatus string
		err := rows.Scan(&m.OrderCode, &m.PaymentID, &paymentStatus, &orderStatus, &m.Amount)
		if err != nil {
			continue
		}
		m.LocalStatus = fmt.Sprintf("order:%s", orderStatus)
		m.GatewayStatus = fmt.Sprintf("payment:%s", paymentStatus)
		m.MismatchType = "status_mismatch"
		mismatches = append(mismatches, m)
	}

	return mismatches, nil
}

func (s *reconciliationService) findOrphans(start, end time.Time) (int, int, map[string]any) {
	// Orphan orders (no payment)
	var orphanOrders int
	s.db.QueryRow(`
		SELECT COUNT(*) FROM orders o
		LEFT JOIN payments p ON o.id = p.order_id
		WHERE p.id IS NULL AND o.created_at >= $1 AND o.created_at < $2
	`, start, end).Scan(&orphanOrders)

	// Orphan payments (no order - shouldn't happen but check)
	var orphanPayments int
	s.db.QueryRow(`
		SELECT COUNT(*) FROM payments p
		LEFT JOIN orders o ON p.order_id = o.id
		WHERE o.id IS NULL AND p.created_at >= $1 AND p.created_at < $2
	`, start, end).Scan(&orphanPayments)

	details := map[string]any{
		"orphan_orders":   orphanOrders,
		"orphan_payments": orphanPayments,
	}

	return orphanOrders, orphanPayments, details
}

func (s *reconciliationService) findStuckPayments(start, end time.Time) (int, []int) {
	query := `
		SELECT id FROM payments
		WHERE status = 'PENDING'
		AND created_at >= $1 AND created_at < $2
		AND created_at < NOW() - INTERVAL '2 hours'
	`

	rows, _ := s.db.Query(query, start, end)
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		rows.Scan(&id)
		ids = append(ids, id)
	}

	return len(ids), ids
}

func (s *reconciliationService) calculateRevenue(start, end time.Time) (expected, actual, variance float64) {
	// Expected: sum of all paid orders
	s.db.QueryRow(`
		SELECT COALESCE(SUM(total_amount), 0) FROM orders
		WHERE status IN ('PAID', 'PROCESSING', 'SHIPPED', 'DELIVERED', 'COMPLETED')
		AND created_at >= $1 AND created_at < $2
	`, start, end).Scan(&expected)

	// Actual: sum of successful payments
	s.db.QueryRow(`
		SELECT COALESCE(SUM(amount), 0) FROM payments
		WHERE status = 'SUCCESS'
		AND created_at >= $1 AND created_at < $2
	`, start, end).Scan(&actual)

	variance = actual - expected
	return
}

func (s *reconciliationService) getRefundTotal(start, end time.Time) float64 {
	var total float64
	s.db.QueryRow(`
		SELECT COALESCE(SUM(refund_amount), 0) FROM refunds
		WHERE status = 'COMPLETED'
		AND created_at >= $1 AND created_at < $2
	`, start, end).Scan(&total)
	return total
}
