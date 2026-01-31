package service

import (
	"database/sql"
	"log"
	"sync"
	"time"
	"zavera/models"
	"zavera/repository"
)

type ShipmentMonitorService interface {
	// Manual triggers
	RunStuckShipmentDetector() error
	RunLostShipmentDetector() error
	RunPickupFailureDetector() error
	RunAllDetectors() error

	// Background jobs
	StartMonitoringJobs(intervalMinutes int)
	StopMonitoringJobs()
}

type shipmentMonitorService struct {
	shippingRepo repository.ShippingRepository
	disputeRepo  repository.DisputeRepository
	orderRepo    repository.OrderRepository
	db           *sql.DB
	stopChan     chan struct{}
	running      bool
	mu           sync.Mutex
}

func NewShipmentMonitorService(
	shippingRepo repository.ShippingRepository,
	disputeRepo repository.DisputeRepository,
	orderRepo repository.OrderRepository,
	db *sql.DB,
) ShipmentMonitorService {
	return &shipmentMonitorService{
		shippingRepo: shippingRepo,
		disputeRepo:  disputeRepo,
		orderRepo:    orderRepo,
		db:           db,
		stopChan:     make(chan struct{}),
	}
}


// RunStuckShipmentDetector detects shipments without tracking updates
// 7 days no update → INVESTIGATION, 14 days → LOST
func (s *shipmentMonitorService) RunStuckShipmentDetector() error {
	log.Println("🔍 Running stuck shipment detector...")

	// Update days_without_update for all active shipments
	updateQuery := `
		UPDATE shipments SET 
			days_without_update = COALESCE(
				EXTRACT(DAY FROM (NOW() - COALESCE(last_tracking_update, shipped_at, created_at))),
				0
			)::int,
			tracking_stale = CASE 
				WHEN EXTRACT(DAY FROM (NOW() - COALESCE(last_tracking_update, shipped_at, created_at))) >= 3 THEN true
				ELSE false
			END
		WHERE status IN ('SHIPPED', 'IN_TRANSIT', 'OUT_FOR_DELIVERY')
	`
	_, err := s.db.Exec(updateQuery)
	if err != nil {
		log.Printf("❌ Failed to update days_without_update: %v", err)
		return err
	}

	// Find shipments stuck for 7+ days (need investigation)
	investigateQuery := `
		SELECT id, order_id, tracking_number, days_without_update
		FROM shipments
		WHERE status IN ('SHIPPED', 'IN_TRANSIT')
		AND days_without_update >= 7
		AND days_without_update < 14
		AND requires_admin_action = false
	`
	rows, err := s.db.Query(investigateQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	var investigateCount int
	for rows.Next() {
		var id, orderID, days int
		var tracking string
		rows.Scan(&id, &orderID, &tracking, &days)

		// Mark for investigation
		s.db.Exec(`
			UPDATE shipments SET 
				status = 'INVESTIGATION',
				investigation_opened_at = NOW(),
				investigation_reason = 'Auto-detected: No tracking update for 7+ days',
				requires_admin_action = true,
				admin_action_reason = 'Stuck shipment - requires investigation'
			WHERE id = $1
		`, id)

		// Create alert
		alert := &models.ShipmentAlert{
			ShipmentID:      id,
			AlertType:       "stuck_shipment",
			AlertLevel:      "critical",
			Title:           "Shipment Stuck - No Updates",
			Description:     "No tracking update for 7+ days. Auto-moved to investigation.",
			AutoActionTaken: true,
			AutoActionType:  "status_change_investigation",
		}
		s.disputeRepo.CreateAlert(alert)

		// Record status change
		s.disputeRepo.RecordStatusChange(id, "IN_TRANSIT", "INVESTIGATION", "system:monitor", "Auto-detected stuck shipment", nil)

		investigateCount++
		log.Printf("⚠️ Shipment %d moved to INVESTIGATION (stuck %d days)", id, days)
	}

	// Find shipments stuck for 14+ days (mark as lost)
	lostQuery := `
		SELECT id, order_id, tracking_number, days_without_update
		FROM shipments
		WHERE status = 'INVESTIGATION'
		AND days_without_update >= 14
	`
	rows2, err := s.db.Query(lostQuery)
	if err != nil {
		return err
	}
	defer rows2.Close()

	var lostCount int
	for rows2.Next() {
		var id, orderID, days int
		var tracking string
		rows2.Scan(&id, &orderID, &tracking, &days)

		// Mark as lost
		s.db.Exec(`
			UPDATE shipments SET 
				status = 'LOST',
				marked_lost_at = NOW(),
				lost_reason = 'Auto-detected: No tracking update for 14+ days',
				requires_admin_action = true,
				admin_action_reason = 'Package presumed lost - requires resolution'
			WHERE id = $1
		`, id)

		// Create alert
		alert := &models.ShipmentAlert{
			ShipmentID:      id,
			AlertType:       "lost_package",
			AlertLevel:      "urgent",
			Title:           "Package Presumed Lost",
			Description:     "No tracking update for 14+ days. Auto-marked as LOST.",
			AutoActionTaken: true,
			AutoActionType:  "status_change_lost",
		}
		s.disputeRepo.CreateAlert(alert)

		// Log courier failure
		failure := &models.CourierFailureLog{
			ShipmentID:    id,
			FailureType:   "lost",
			FailureReason: "Auto-detected: No tracking update for 14+ days",
		}
		s.disputeRepo.LogCourierFailure(failure)

		// Auto-create dispute
		order, _ := s.orderRepo.FindByID(orderID)
		if order != nil {
			dispute := &models.Dispute{
				DisputeCode:   repository.GenerateDisputeCode(),
				OrderID:       orderID,
				ShipmentID:    &id,
				DisputeType:   models.DisputeTypeLostPackage,
				Status:        models.DisputeStatusOpen,
				Title:         "Auto-detected Lost Package",
				Description:   "Package has had no tracking updates for 14+ days and is presumed lost.",
				CustomerEmail: order.CustomerEmail,
				CustomerPhone: order.CustomerPhone,
			}
			s.disputeRepo.Create(dispute)
		}

		s.disputeRepo.RecordStatusChange(id, "INVESTIGATION", "LOST", "system:monitor", "Auto-detected lost package", nil)

		lostCount++
		log.Printf("❌ Shipment %d marked as LOST (stuck %d days)", id, days)
	}

	log.Printf("✅ Stuck detector complete: %d → investigation, %d → lost", investigateCount, lostCount)
	return nil
}


// RunLostShipmentDetector specifically handles lost shipment detection
func (s *shipmentMonitorService) RunLostShipmentDetector() error {
	log.Println("🔍 Running lost shipment detector...")

	// Find shipments in INVESTIGATION that have been there too long
	query := `
		SELECT id, order_id, investigation_opened_at
		FROM shipments
		WHERE status = 'INVESTIGATION'
		AND investigation_opened_at IS NOT NULL
		AND investigation_opened_at < NOW() - INTERVAL '7 days'
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var id, orderID int
		var openedAt time.Time
		rows.Scan(&id, &orderID, &openedAt)

		// Mark as lost
		s.db.Exec(`
			UPDATE shipments SET 
				status = 'LOST',
				marked_lost_at = NOW(),
				lost_reason = 'Investigation timeout - no resolution after 7 days'
			WHERE id = $1
		`, id)

		alert := &models.ShipmentAlert{
			ShipmentID:      id,
			AlertType:       "investigation_timeout",
			AlertLevel:      "urgent",
			Title:           "Investigation Timeout",
			Description:     "Investigation has been open for 7+ days without resolution. Marked as LOST.",
			AutoActionTaken: true,
			AutoActionType:  "status_change_lost",
		}
		s.disputeRepo.CreateAlert(alert)

		s.disputeRepo.RecordStatusChange(id, "INVESTIGATION", "LOST", "system:monitor", "Investigation timeout", nil)

		count++
		log.Printf("⏰ Shipment %d: investigation timeout, marked LOST", id)
	}

	log.Printf("✅ Lost detector complete: %d shipments marked lost", count)
	return nil
}

// RunPickupFailureDetector detects pickup failures
// 48h no pickup → alert, 3x fail → admin required
func (s *shipmentMonitorService) RunPickupFailureDetector() error {
	log.Println("🔍 Running pickup failure detector...")

	// Find shipments past pickup deadline
	deadlineQuery := `
		SELECT id, order_id, pickup_scheduled_at, pickup_deadline, pickup_attempts
		FROM shipments
		WHERE status = 'PICKUP_SCHEDULED'
		AND pickup_deadline IS NOT NULL
		AND pickup_deadline < NOW()
		AND requires_admin_action = false
	`
	rows, err := s.db.Query(deadlineQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	var alertCount int
	for rows.Next() {
		var id, orderID, attempts int
		var scheduledAt, deadline *time.Time
		rows.Scan(&id, &orderID, &scheduledAt, &deadline, &attempts)

		// Update status to PICKUP_FAILED
		s.db.Exec(`
			UPDATE shipments SET 
				status = 'PICKUP_FAILED',
				pickup_attempts = pickup_attempts + 1,
				last_pickup_attempt_at = NOW()
			WHERE id = $1
		`, id)

		// Check if needs admin (3+ failures)
		newAttempts := attempts + 1
		if newAttempts >= 3 {
			s.db.Exec(`
				UPDATE shipments SET 
					requires_admin_action = true,
					admin_action_reason = 'Pickup failed 3+ times - manual intervention required'
				WHERE id = $1
			`, id)

			alert := &models.ShipmentAlert{
				ShipmentID:      id,
				AlertType:       "pickup_failed_critical",
				AlertLevel:      "urgent",
				Title:           "Pickup Failed Multiple Times",
				Description:     "Pickup has failed 3+ times. Manual intervention required.",
				AutoActionTaken: true,
				AutoActionType:  "flag_admin_required",
			}
			s.disputeRepo.CreateAlert(alert)
		} else {
			alert := &models.ShipmentAlert{
				ShipmentID:      id,
				AlertType:       "pickup_deadline_missed",
				AlertLevel:      "warning",
				Title:           "Pickup Deadline Missed",
				Description:     "Courier did not pick up package within 48 hours.",
				AutoActionTaken: true,
				AutoActionType:  "status_change_pickup_failed",
			}
			s.disputeRepo.CreateAlert(alert)
		}

		// Log courier failure
		failure := &models.CourierFailureLog{
			ShipmentID:    id,
			FailureType:   "pickup_failed",
			FailureReason: "Pickup deadline missed",
		}
		s.disputeRepo.LogCourierFailure(failure)

		s.disputeRepo.RecordStatusChange(id, "PICKUP_SCHEDULED", "PICKUP_FAILED", "system:monitor", "Pickup deadline missed", nil)

		alertCount++
		log.Printf("📦 Shipment %d: pickup failed (attempt %d)", id, newAttempts)
	}

	log.Printf("✅ Pickup detector complete: %d failures detected", alertCount)
	return nil
}

// RunAllDetectors runs all monitoring detectors
func (s *shipmentMonitorService) RunAllDetectors() error {
	log.Println("🚀 Running all shipment detectors...")

	if err := s.RunStuckShipmentDetector(); err != nil {
		log.Printf("❌ Stuck detector error: %v", err)
	}

	if err := s.RunLostShipmentDetector(); err != nil {
		log.Printf("❌ Lost detector error: %v", err)
	}

	if err := s.RunPickupFailureDetector(); err != nil {
		log.Printf("❌ Pickup detector error: %v", err)
	}

	log.Println("✅ All detectors complete")
	return nil
}


// StartMonitoringJobs starts background monitoring jobs
func (s *shipmentMonitorService) StartMonitoringJobs(intervalMinutes int) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.stopChan = make(chan struct{})
	s.mu.Unlock()

	go func() {
		ticker := time.NewTicker(time.Duration(intervalMinutes) * time.Minute)
		defer ticker.Stop()

		// Run immediately on start
		s.RunAllDetectors()

		for {
			select {
			case <-ticker.C:
				s.RunAllDetectors()
			case <-s.stopChan:
				log.Println("🛑 Shipment monitoring jobs stopped")
				return
			}
		}
	}()

	log.Printf("🚀 Shipment monitoring jobs started (interval: %d minutes)", intervalMinutes)
}

// StopMonitoringJobs stops background monitoring jobs
func (s *shipmentMonitorService) StopMonitoringJobs() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	close(s.stopChan)
	s.running = false
}
