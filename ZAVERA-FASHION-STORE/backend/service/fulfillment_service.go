package service

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
	"zavera/dto"
	"zavera/models"
	"zavera/repository"
)

var (
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrAdminRequired           = errors.New("admin privileges required for this transition")
	ErrShipmentInFinalState    = errors.New("shipment is in final state")
	ErrDisputeNotFound         = errors.New("dispute not found")
	ErrDisputeAlreadyResolved  = errors.New("dispute already resolved")
	ErrCannotReship            = errors.New("cannot create reship for this shipment")
)

type FulfillmentService interface {
	// Status management
	UpdateShipmentStatus(shipmentID int, newStatus models.ShipmentStatus, reason string, isAdmin bool, changedBy string) error
	SchedulePickup(shipmentID int, req *dto.SchedulePickupRequest, changedBy string) error
	MarkShipped(shipmentID int, req *dto.MarkShippedRequest, changedBy string) error
	
	// Problem handling
	OpenInvestigation(shipmentID int, req *dto.InvestigateShipmentRequest, adminEmail string) error
	MarkLost(shipmentID int, req *dto.MarkLostRequest, adminEmail string) error
	OverrideStatus(shipmentID int, req *dto.OverrideStatusRequest, adminEmail string) error
	
	// Reship
	CreateReship(shipmentID int, req *dto.ReshipRequest, adminEmail string) (*models.Shipment, error)
	
	// Queries
	GetEnhancedShipment(shipmentID int) (*dto.EnhancedShipmentResponse, error)
	GetStuckShipments(daysThreshold int) ([]*dto.StuckShipmentResponse, error)
	GetPickupFailures() ([]*dto.PickupFailureResponse, error)
	GetFulfillmentDashboard() (*dto.FulfillmentDashboardResponse, error)
	GetShipmentsList(status string, page, pageSize int) ([]dto.ShipmentListItem, int, error)
}

type fulfillmentService struct {
	shippingRepo repository.ShippingRepository
	disputeRepo  repository.DisputeRepository
	orderRepo    repository.OrderRepository
	auditRepo    repository.AdminAuditRepository
	resiService  ResiService
	emailService EmailService
	db           *sql.DB
}

func NewFulfillmentService(
	shippingRepo repository.ShippingRepository,
	disputeRepo repository.DisputeRepository,
	orderRepo repository.OrderRepository,
	auditRepo repository.AdminAuditRepository,
	db *sql.DB,
) FulfillmentService {
	return &fulfillmentService{
		shippingRepo: shippingRepo,
		disputeRepo:  disputeRepo,
		orderRepo:    orderRepo,
		auditRepo:    auditRepo,
		resiService:  NewResiService(orderRepo),
		emailService: nil, // Will be set via SetEmailService if needed
		db:           db,
	}
}

// UpdateShipmentStatus updates shipment status with validation
func (s *fulfillmentService) UpdateShipmentStatus(shipmentID int, newStatus models.ShipmentStatus, reason string, isAdmin bool, changedBy string) error {
	shipment, err := s.shippingRepo.GetShipmentByID(shipmentID)
	if err != nil {
		return ErrShipmentNotFound
	}

	currentStatus := shipment.Status

	// Check if in final state
	if currentStatus.IsFinalStatus() && !isAdmin {
		return ErrShipmentInFinalState
	}

	// Validate transition
	if !currentStatus.IsValidTransition(newStatus) {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidStatusTransition, currentStatus, newStatus)
	}

	// Check if admin required
	if currentStatus.RequiresAdmin(newStatus) && !isAdmin {
		return ErrAdminRequired
	}

	// Update status
	if err := s.shippingRepo.UpdateShipmentStatus(shipmentID, newStatus); err != nil {
		return err
	}

	// Record history
	s.disputeRepo.RecordStatusChange(shipmentID, string(currentStatus), string(newStatus), changedBy, reason, nil)

	// Handle special status updates
	s.handleStatusSideEffects(shipment, newStatus, changedBy)

	log.Printf("📦 Shipment %d: %s -> %s by %s", shipmentID, currentStatus, newStatus, changedBy)
	return nil
}

// SchedulePickup schedules courier pickup
func (s *fulfillmentService) SchedulePickup(shipmentID int, req *dto.SchedulePickupRequest, changedBy string) error {
	shipment, err := s.shippingRepo.GetShipmentByID(shipmentID)
	if err != nil {
		return ErrShipmentNotFound
	}

	if shipment.Status != models.ShipmentStatusProcessing {
		return fmt.Errorf("can only schedule pickup for PROCESSING shipments, current: %s", shipment.Status)
	}

	pickupDate, err := time.Parse("2006-01-02", req.PickupDate)
	if err != nil {
		return fmt.Errorf("invalid pickup date format")
	}

	// Set pickup deadline (48 hours from scheduled)
	deadline := pickupDate.Add(48 * time.Hour)

	query := `
		UPDATE shipments SET 
			status = 'PICKUP_SCHEDULED',
			pickup_scheduled_at = $1,
			pickup_deadline = $2,
			pickup_notes = $3,
			updated_at = NOW()
		WHERE id = $4
	`
	_, err = s.db.Exec(query, pickupDate, deadline, req.Notes, shipmentID)
	if err != nil {
		return err
	}

	s.disputeRepo.RecordStatusChange(shipmentID, string(shipment.Status), "PICKUP_SCHEDULED", changedBy, "Pickup scheduled", map[string]any{
		"pickup_date": req.PickupDate,
		"deadline":    deadline,
	})

	return nil
}

// MarkShipped marks shipment as shipped with tracking number
func (s *fulfillmentService) MarkShipped(shipmentID int, req *dto.MarkShippedRequest, changedBy string) error {
	shipment, err := s.shippingRepo.GetShipmentByID(shipmentID)
	if err != nil {
		return ErrShipmentNotFound
	}

	validFromStatuses := []models.ShipmentStatus{
		models.ShipmentStatusProcessing,
		models.ShipmentStatusPickupScheduled,
		models.ShipmentStatusPickupFailed,
	}

	valid := false
	for _, s := range validFromStatuses {
		if shipment.Status == s {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("cannot mark as shipped from status: %s", shipment.Status)
	}

	// Generate resi if not provided
	trackingNumber := req.TrackingNumber
	if trackingNumber == "" {
		// Generate resi using resi service
		generatedResi, err := s.resiService.GenerateResi(shipment.OrderID, shipment.ProviderCode)
		if err != nil {
			return fmt.Errorf("failed to generate resi: %w", err)
		}
		trackingNumber = generatedResi
	}

	if err := s.shippingRepo.MarkShipmentShipped(shipmentID, trackingNumber); err != nil {
		return err
	}

	// Update order status and resi
	s.orderRepo.MarkAsShippedWithResi(shipment.OrderID, trackingNumber)

	s.disputeRepo.RecordStatusChange(shipmentID, string(shipment.Status), "SHIPPED", changedBy, "Marked as shipped", map[string]any{
		"tracking_number": trackingNumber,
		"auto_generated":  req.TrackingNumber == "",
	})

	// Send shipping email (non-blocking)
	if s.emailService != nil {
		go func() {
			order, err := s.orderRepo.FindByID(shipment.OrderID)
			if err == nil {
				// Get shipping address from order metadata
				shippingAddr := ""
				if addr, ok := order.Metadata["shipping_address_snapshot"].(string); ok {
					shippingAddr = addr
				}
				s.emailService.SendOrderShipped(order, shipment, shippingAddr)
			}
		}()
	}

	return nil
}

// OpenInvestigation opens an investigation for a shipment
func (s *fulfillmentService) OpenInvestigation(shipmentID int, req *dto.InvestigateShipmentRequest, adminEmail string) error {
	shipment, err := s.shippingRepo.GetShipmentByID(shipmentID)
	if err != nil {
		return ErrShipmentNotFound
	}

	// Can investigate from these statuses
	validStatuses := []models.ShipmentStatus{
		models.ShipmentStatusShipped,
		models.ShipmentStatusInTransit,
		models.ShipmentStatusOutForDelivery,
		models.ShipmentStatusDelivered, // For disputed deliveries
	}

	valid := false
	for _, st := range validStatuses {
		if shipment.Status == st {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("cannot investigate shipment in status: %s", shipment.Status)
	}

	query := `
		UPDATE shipments SET 
			status = 'INVESTIGATION',
			investigation_opened_at = NOW(),
			investigation_reason = $1,
			requires_admin_action = true,
			admin_action_reason = 'Under investigation',
			updated_at = NOW()
		WHERE id = $2
	`
	_, err = s.db.Exec(query, req.Reason, shipmentID)
	if err != nil {
		return err
	}

	// Create alert
	alert := &models.ShipmentAlert{
		ShipmentID:  shipmentID,
		AlertType:   "investigation",
		AlertLevel:  "critical",
		Title:       "Shipment Under Investigation",
		Description: req.Reason,
	}
	s.disputeRepo.CreateAlert(alert)

	s.disputeRepo.RecordStatusChange(shipmentID, string(shipment.Status), "INVESTIGATION", "admin:"+adminEmail, req.Reason, nil)

	log.Printf("🔍 Investigation opened for shipment %d by %s", shipmentID, adminEmail)
	return nil
}

// MarkLost marks a shipment as lost
func (s *fulfillmentService) MarkLost(shipmentID int, req *dto.MarkLostRequest, adminEmail string) error {
	shipment, err := s.shippingRepo.GetShipmentByID(shipmentID)
	if err != nil {
		return ErrShipmentNotFound
	}

	// Can mark lost from these statuses
	validStatuses := []models.ShipmentStatus{
		models.ShipmentStatusShipped,
		models.ShipmentStatusInTransit,
		models.ShipmentStatusInvestigation,
	}

	valid := false
	for _, st := range validStatuses {
		if shipment.Status == st {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("cannot mark as lost from status: %s", shipment.Status)
	}

	query := `
		UPDATE shipments SET 
			status = 'LOST',
			marked_lost_at = NOW(),
			lost_reason = $1,
			requires_admin_action = true,
			admin_action_reason = 'Package lost - requires resolution',
			updated_at = NOW()
		WHERE id = $2
	`
	_, err = s.db.Exec(query, req.Reason, shipmentID)
	if err != nil {
		return err
	}

	// Log courier failure
	failure := &models.CourierFailureLog{
		ShipmentID:      shipmentID,
		FailureType:     "lost",
		FailureReason:   req.Reason,
		CourierCode:     shipment.ProviderCode,
		CourierName:     shipment.ProviderName,
		CourierTracking: shipment.TrackingNumber,
	}
	s.disputeRepo.LogCourierFailure(failure)

	// Create alert
	alert := &models.ShipmentAlert{
		ShipmentID:  shipmentID,
		AlertType:   "lost",
		AlertLevel:  "urgent",
		Title:       "Package Lost",
		Description: req.Reason,
	}
	s.disputeRepo.CreateAlert(alert)

	// Create dispute if requested
	if req.CreateDispute {
		order, _ := s.orderRepo.FindByID(shipment.OrderID)
		if order != nil {
			dispute := &models.Dispute{
				DisputeCode:   repository.GenerateDisputeCode(),
				OrderID:       order.ID,
				ShipmentID:    &shipmentID,
				DisputeType:   models.DisputeTypeLostPackage,
				Status:        models.DisputeStatusOpen,
				Title:         "Lost Package",
				Description:   req.Reason,
				CustomerEmail: order.CustomerEmail,
				CustomerPhone: order.CustomerPhone,
			}
			if order.UserID != nil {
				dispute.CustomerUserID = order.UserID
			}
			s.disputeRepo.Create(dispute)
		}
	}

	s.disputeRepo.RecordStatusChange(shipmentID, string(shipment.Status), "LOST", "admin:"+adminEmail, req.Reason, nil)

	log.Printf("📦❌ Shipment %d marked as LOST by %s", shipmentID, adminEmail)
	return nil
}

// OverrideStatus allows admin to override status with bypass
func (s *fulfillmentService) OverrideStatus(shipmentID int, req *dto.OverrideStatusRequest, adminEmail string) error {
	shipment, err := s.shippingRepo.GetShipmentByID(shipmentID)
	if err != nil {
		return ErrShipmentNotFound
	}

	newStatus := models.ShipmentStatus(req.NewStatus)

	// If not bypassing, validate transition
	if !req.BypassValidation {
		if !shipment.Status.IsValidTransition(newStatus) {
			return fmt.Errorf("%w: %s -> %s", ErrInvalidStatusTransition, shipment.Status, newStatus)
		}
	}

	if err := s.shippingRepo.UpdateShipmentStatus(shipmentID, newStatus); err != nil {
		return err
	}

	// Clear admin action flag if resolved
	if !newStatus.RequiresAction() {
		s.db.Exec(`UPDATE shipments SET requires_admin_action = false, admin_action_reason = NULL WHERE id = $1`, shipmentID)
	}

	s.disputeRepo.RecordStatusChange(shipmentID, string(shipment.Status), string(newStatus), "admin:"+adminEmail, req.Reason, map[string]any{
		"bypass_validation": req.BypassValidation,
		"admin_override":    true,
	})

	log.Printf("⚡ Admin override: shipment %d %s -> %s by %s", shipmentID, shipment.Status, newStatus, adminEmail)
	return nil
}

// CreateReship creates a replacement shipment
func (s *fulfillmentService) CreateReship(shipmentID int, req *dto.ReshipRequest, adminEmail string) (*models.Shipment, error) {
	shipment, err := s.shippingRepo.GetShipmentByID(shipmentID)
	if err != nil {
		return nil, ErrShipmentNotFound
	}

	// HARDENING: Prevent reship loops - max 3 reships per original order
	const maxReshipCount = 3
	if shipment.ReshipCount >= maxReshipCount {
		return nil, fmt.Errorf("maximum reship limit (%d) reached for this shipment", maxReshipCount)
	}

	// Can reship from these statuses
	validStatuses := []models.ShipmentStatus{
		models.ShipmentStatusLost,
		models.ShipmentStatusReturnedToSender,
		models.ShipmentStatusInvestigation,
	}

	valid := false
	for _, st := range validStatuses {
		if shipment.Status == st {
			valid = true
			break
		}
	}
	if !valid {
		return nil, fmt.Errorf("%w: current status is %s", ErrCannotReship, shipment.Status)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Mark original as REPLACED
	_, err = tx.Exec(`
		UPDATE shipments SET 
			status = 'REPLACED',
			replaced_by_shipment_id = NULL,
			updated_at = NOW()
		WHERE id = $1
	`, shipmentID)
	if err != nil {
		return nil, err
	}

	// Create new shipment
	query := `
		INSERT INTO shipments (
			order_id, provider_code, provider_name, service_code, service_name,
			cost, etd, weight, status, origin_city_id, origin_city_name,
			destination_city_id, destination_city_name, tracking_number,
			is_replacement, original_shipment_id, reship_reason, reship_cost, reship_cost_bearer
		)
		SELECT 
			order_id, provider_code, provider_name, service_code, service_name,
			cost, etd, weight, 'PROCESSING', origin_city_id, origin_city_name,
			destination_city_id, destination_city_name, $2,
			true, $1, $3, cost, $4
		FROM shipments WHERE id = $1
		RETURNING id
	`

	var newShipmentID int
	err = tx.QueryRow(query, shipmentID, req.NewTrackingNo, req.Reason, req.CostBearer).Scan(&newShipmentID)
	if err != nil {
		return nil, err
	}

	// Update original with replacement ID
	_, err = tx.Exec(`UPDATE shipments SET replaced_by_shipment_id = $1 WHERE id = $2`, newShipmentID, shipmentID)
	if err != nil {
		return nil, err
	}

	// Increment reship count on original
	_, err = tx.Exec(`UPDATE shipments SET reship_count = reship_count + 1 WHERE id = $1`, shipmentID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Record history
	s.disputeRepo.RecordStatusChange(shipmentID, string(shipment.Status), "REPLACED", "admin:"+adminEmail, req.Reason, map[string]any{
		"new_shipment_id": newShipmentID,
		"cost_bearer":     req.CostBearer,
	})

	s.disputeRepo.RecordStatusChange(newShipmentID, "", "PROCESSING", "admin:"+adminEmail, "Replacement shipment created", map[string]any{
		"original_shipment_id": shipmentID,
	})

	log.Printf("📦🔄 Reship created: %d -> %d by %s", shipmentID, newShipmentID, adminEmail)

	return s.shippingRepo.GetShipmentByID(newShipmentID)
}

// GetEnhancedShipment returns detailed shipment info
func (s *fulfillmentService) GetEnhancedShipment(shipmentID int) (*dto.EnhancedShipmentResponse, error) {
	shipment, err := s.shippingRepo.GetShipmentByID(shipmentID)
	if err != nil {
		return nil, ErrShipmentNotFound
	}

	order, _ := s.orderRepo.FindByID(shipment.OrderID)
	orderCode := ""
	if order != nil {
		orderCode = order.OrderCode
	}

	resp := &dto.EnhancedShipmentResponse{
		ID:                  shipment.ID,
		OrderID:             shipment.OrderID,
		OrderCode:           orderCode,
		ProviderCode:        shipment.ProviderCode,
		ProviderName:        shipment.ProviderName,
		ServiceName:         shipment.ServiceName,
		TrackingNumber:      shipment.TrackingNumber,
		Status:              string(shipment.Status),
		Cost:                shipment.Cost,
		Weight:              shipment.Weight,
		Origin:              shipment.OriginCityName,
		Destination:         shipment.DestinationCityName,
		CreatedAt:           shipment.CreatedAt.Format(time.RFC3339),
		PickupAttempts:      shipment.PickupAttempts,
		DaysWithoutUpdate:   shipment.DaysWithoutUpdate,
		TrackingStale:       shipment.TrackingStale,
		IsProblematic:       shipment.Status.IsProblematic(),
		RequiresAdminAction: shipment.RequiresAdminAction,
		AdminActionReason:   shipment.AdminActionReason,
		IsReplacement:       shipment.IsReplacement,
		OriginalShipmentID:  shipment.OriginalShipmentID,
		ReplacedByID:        shipment.ReplacedByShipmentID,
	}

	if shipment.ShippedAt != nil {
		t := shipment.ShippedAt.Format(time.RFC3339)
		resp.ShippedAt = &t
	}
	if shipment.DeliveredAt != nil {
		t := shipment.DeliveredAt.Format(time.RFC3339)
		resp.DeliveredAt = &t
	}
	if shipment.PickupScheduledAt != nil {
		t := shipment.PickupScheduledAt.Format(time.RFC3339)
		resp.PickupScheduledAt = &t
	}
	if shipment.LastTrackingUpdate != nil {
		t := shipment.LastTrackingUpdate.Format(time.RFC3339)
		resp.LastTrackingUpdate = &t
	}

	// Load status history
	history, _ := s.disputeRepo.GetStatusHistory(shipmentID)
	for _, h := range history {
		resp.StatusHistory = append(resp.StatusHistory, dto.ShipmentStatusHistoryResponse{
			FromStatus: h.FromStatus,
			ToStatus:   h.ToStatus,
			ChangedBy:  h.ChangedBy,
			Reason:     h.Reason,
			CreatedAt:  h.CreatedAt.Format(time.RFC3339),
		})
	}

	// Load alerts
	alerts, _ := s.disputeRepo.GetAlertsByShipment(shipmentID)
	for _, a := range alerts {
		resp.Alerts = append(resp.Alerts, dto.ShipmentAlertResponse{
			ID:          a.ID,
			ShipmentID:  a.ShipmentID,
			AlertType:   a.AlertType,
			AlertLevel:  a.AlertLevel,
			Title:       a.Title,
			Description: a.Description,
			Acknowledged: a.Acknowledged,
			Resolved:    a.Resolved,
			CreatedAt:   a.CreatedAt.Format(time.RFC3339),
		})
	}

	return resp, nil
}

// GetStuckShipments returns shipments without tracking updates
func (s *fulfillmentService) GetStuckShipments(daysThreshold int) ([]*dto.StuckShipmentResponse, error) {
	query := `
		SELECT s.id, s.order_id, o.order_code, s.tracking_number, s.status,
		       s.provider_code, s.days_without_update, s.last_tracking_update, s.shipped_at
		FROM shipments s
		JOIN orders o ON s.order_id = o.id
		WHERE s.status IN ('SHIPPED', 'IN_TRANSIT', 'OUT_FOR_DELIVERY')
		AND s.days_without_update >= $1
		ORDER BY s.days_without_update DESC
	`

	rows, err := s.db.Query(query, daysThreshold)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stuck []*dto.StuckShipmentResponse
	for rows.Next() {
		var ss dto.StuckShipmentResponse
		err := rows.Scan(
			&ss.ShipmentID, &ss.OrderID, &ss.OrderCode, &ss.TrackingNumber,
			&ss.Status, &ss.ProviderCode, &ss.DaysWithoutUpdate,
			&ss.LastTrackingUpdate, &ss.ShippedAt,
		)
		if err != nil {
			continue
		}

		// Determine alert level
		if ss.DaysWithoutUpdate >= 14 {
			ss.AlertLevel = "urgent"
		} else if ss.DaysWithoutUpdate >= 7 {
			ss.AlertLevel = "critical"
		} else {
			ss.AlertLevel = "warning"
		}

		stuck = append(stuck, &ss)
	}

	return stuck, nil
}

// GetPickupFailures returns shipments with pickup failures
func (s *fulfillmentService) GetPickupFailures() ([]*dto.PickupFailureResponse, error) {
	query := `
		SELECT s.id, s.order_id, o.order_code, s.pickup_attempts,
		       s.last_pickup_attempt_at, s.pickup_scheduled_at, s.pickup_deadline
		FROM shipments s
		JOIN orders o ON s.order_id = o.id
		WHERE s.status = 'PICKUP_FAILED' OR (s.status = 'PICKUP_SCHEDULED' AND s.pickup_attempts > 0)
		ORDER BY s.pickup_attempts DESC, s.created_at ASC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var failures []*dto.PickupFailureResponse
	for rows.Next() {
		var pf dto.PickupFailureResponse
		err := rows.Scan(
			&pf.ShipmentID, &pf.OrderID, &pf.OrderCode, &pf.PickupAttempts,
			&pf.LastAttemptAt, &pf.ScheduledAt, &pf.Deadline,
		)
		if err != nil {
			continue
		}
		pf.RequiresAdmin = pf.PickupAttempts >= 3
		failures = append(failures, &pf)
	}

	return failures, nil
}

// GetFulfillmentDashboard returns fulfillment overview
func (s *fulfillmentService) GetFulfillmentDashboard() (*dto.FulfillmentDashboardResponse, error) {
	dashboard := &dto.FulfillmentDashboardResponse{
		StatusCounts: make(map[string]int),
	}

	// Get status counts
	statusQuery := `
		SELECT status, COUNT(*) FROM shipments 
		GROUP BY status
	`
	rows, err := s.db.Query(statusQuery)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var status string
			var count int
			rows.Scan(&status, &count)
			dashboard.StatusCounts[status] = count
		}
	}

	// Problem counts
	s.db.QueryRow(`SELECT COUNT(*) FROM shipments WHERE days_without_update >= 7 AND status IN ('SHIPPED', 'IN_TRANSIT')`).Scan(&dashboard.StuckShipments)
	s.db.QueryRow(`SELECT COUNT(*) FROM shipments WHERE status = 'PICKUP_FAILED' OR (status = 'PICKUP_SCHEDULED' AND pickup_attempts >= 3)`).Scan(&dashboard.PickupFailures)
	s.db.QueryRow(`SELECT COUNT(*) FROM shipments WHERE status = 'DELIVERY_FAILED'`).Scan(&dashboard.DeliveryFailures)
	s.db.QueryRow(`SELECT COUNT(*) FROM shipments WHERE status = 'LOST'`).Scan(&dashboard.LostPackages)
	s.db.QueryRow(`SELECT COUNT(*) FROM shipments WHERE status = 'INVESTIGATION'`).Scan(&dashboard.UnderInvestigation)

	// Alert counts
	s.db.QueryRow(`SELECT COUNT(*) FROM shipment_alerts WHERE resolved = false`).Scan(&dashboard.UnresolvedAlerts)
	s.db.QueryRow(`SELECT COUNT(*) FROM shipment_alerts WHERE resolved = false AND alert_level IN ('critical', 'urgent')`).Scan(&dashboard.CriticalAlerts)

	// Dispute counts
	s.db.QueryRow(`SELECT COUNT(*) FROM disputes WHERE status = 'OPEN'`).Scan(&dashboard.OpenDisputes)
	s.db.QueryRow(`SELECT COUNT(*) FROM disputes WHERE status = 'PENDING_RESOLUTION'`).Scan(&dashboard.PendingResolution)

	// Performance metrics
	s.db.QueryRow(`
		SELECT COALESCE(AVG(EXTRACT(DAY FROM (delivered_at - shipped_at))), 0)
		FROM shipments WHERE status = 'DELIVERED' AND delivered_at IS NOT NULL AND shipped_at IS NOT NULL
		AND delivered_at > NOW() - INTERVAL '30 days'
	`).Scan(&dashboard.AvgDeliveryDays)

	// Recent alerts
	alerts, _ := s.disputeRepo.GetUnresolvedAlerts()
	for i, a := range alerts {
		if i >= 10 {
			break
		}
		dashboard.RecentAlerts = append(dashboard.RecentAlerts, dto.ShipmentAlertResponse{
			ID:          a.ID,
			ShipmentID:  a.ShipmentID,
			AlertType:   a.AlertType,
			AlertLevel:  a.AlertLevel,
			Title:       a.Title,
			Description: a.Description,
			Acknowledged: a.Acknowledged,
			Resolved:    a.Resolved,
			CreatedAt:   a.CreatedAt.Format(time.RFC3339),
		})
	}

	return dashboard, nil
}

// handleStatusSideEffects handles side effects of status changes
func (s *fulfillmentService) handleStatusSideEffects(shipment *models.Shipment, newStatus models.ShipmentStatus, changedBy string) {
	switch newStatus {
	case models.ShipmentStatusDelivered:
		s.shippingRepo.MarkShipmentDelivered(shipment.ID)
		s.orderRepo.UpdateStatus(shipment.OrderID, models.OrderStatusDelivered)
		// Resolve any open alerts
		s.db.Exec(`UPDATE shipment_alerts SET resolved = true, resolution_notes = 'Delivered' WHERE shipment_id = $1 AND resolved = false`, shipment.ID)
		
		// Send delivery email (non-blocking)
		if s.emailService != nil {
			go func() {
				order, err := s.orderRepo.FindByID(shipment.OrderID)
				if err == nil {
					s.emailService.SendOrderDelivered(order, shipment)
				}
			}()
		}

	case models.ShipmentStatusPickupFailed:
		// Increment pickup attempts
		s.db.Exec(`UPDATE shipments SET pickup_attempts = pickup_attempts + 1, last_pickup_attempt_at = NOW() WHERE id = $1`, shipment.ID)
		// Log failure
		failure := &models.CourierFailureLog{
			ShipmentID:   shipment.ID,
			FailureType:  "pickup_failed",
			CourierCode:  shipment.ProviderCode,
			CourierName:  shipment.ProviderName,
		}
		s.disputeRepo.LogCourierFailure(failure)
		// Check if needs admin
		var attempts int
		s.db.QueryRow(`SELECT pickup_attempts FROM shipments WHERE id = $1`, shipment.ID).Scan(&attempts)
		if attempts >= 3 {
			s.db.Exec(`UPDATE shipments SET requires_admin_action = true, admin_action_reason = 'Pickup failed 3+ times' WHERE id = $1`, shipment.ID)
			alert := &models.ShipmentAlert{
				ShipmentID:  shipment.ID,
				AlertType:   "pickup_failed",
				AlertLevel:  "critical",
				Title:       "Pickup Failed Multiple Times",
				Description: fmt.Sprintf("Pickup has failed %d times", attempts),
			}
			s.disputeRepo.CreateAlert(alert)
		}

	case models.ShipmentStatusDeliveryFailed:
		// Increment delivery attempts
		s.db.Exec(`UPDATE shipments SET delivery_attempts = delivery_attempts + 1, last_delivery_attempt_at = NOW() WHERE id = $1`, shipment.ID)
		failure := &models.CourierFailureLog{
			ShipmentID:   shipment.ID,
			FailureType:  "delivery_failed",
			CourierCode:  shipment.ProviderCode,
			CourierName:  shipment.ProviderName,
		}
		s.disputeRepo.LogCourierFailure(failure)
	}
}


// GetShipmentsList returns paginated list of shipments
func (s *fulfillmentService) GetShipmentsList(status string, page, pageSize int) ([]dto.ShipmentListItem, int, error) {
	var shipments []dto.ShipmentListItem
	offset := (page - 1) * pageSize

	// Build query
	query := `
		SELECT 
			s.id,
			o.order_code,
			s.tracking_number,
			s.status,
			s.provider_code,
			s.provider_name,
			s.days_without_update,
			s.created_at
		FROM shipments s
		JOIN orders o ON s.order_id = o.id
		WHERE 1=1
	`
	
	args := []interface{}{}
	argCount := 1
	
	if status != "" {
		query += fmt.Sprintf(" AND s.status = $%d", argCount)
		args = append(args, status)
		argCount++
	}
	
	query += " ORDER BY s.created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, pageSize, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		log.Printf("❌ Failed to query shipments: %v", err)
		log.Printf("   Query: %s", query)
		log.Printf("   Args: %v", args)
		return nil, 0, fmt.Errorf("failed to query shipments: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item dto.ShipmentListItem
		var createdAt time.Time
		var trackingNumber sql.NullString
		if err := rows.Scan(&item.ID, &item.OrderCode, &trackingNumber, &item.Status,
			&item.ProviderCode, &item.ProviderName, &item.DaysWithoutUpdate, &createdAt); err != nil {
			log.Printf("⚠️ Failed to scan shipment row: %v", err)
			continue
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		if trackingNumber.Valid {
			item.TrackingNumber = trackingNumber.String
		} else {
			item.TrackingNumber = ""
		}
		shipments = append(shipments, item)
	}
	
	log.Printf("📦 GetShipmentsList: status=%s, page=%d, found %d shipments", status, page, len(shipments))

	// Get total count
	countQuery := "SELECT COUNT(*) FROM shipments s WHERE 1=1"
	countArgs := []interface{}{}
	if status != "" {
		countQuery += " AND s.status = $1"
		countArgs = append(countArgs, status)
	}
	
	var total int
	s.db.QueryRow(countQuery, countArgs...).Scan(&total)

	return shipments, total, nil
}
