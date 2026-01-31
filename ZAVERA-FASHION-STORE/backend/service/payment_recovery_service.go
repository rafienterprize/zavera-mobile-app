package service

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
	"zavera/dto"
	"zavera/models"
	"zavera/repository"
)

// PaymentRecoveryService handles stuck payment detection and recovery
type PaymentRecoveryService interface {
	// Sync operations
	SyncPaymentStatus(paymentID int, syncType string) (*dto.PaymentSyncResponse, error)
	SyncAllPendingPayments() (int, int, error) // returns (synced, errors)
	
	// Detection
	FindStuckPayments(hoursThreshold int) ([]*dto.StuckPaymentResponse, error)
	FindOrphanOrders() ([]int, error)
	FindOrphanPayments() ([]int, error)
	
	// Recovery
	RecoverStuckPayment(paymentID int) error
	ResolveOrphanOrder(orderID int) error
	
	// Cron job
	StartRecoveryJob(intervalMinutes int)
	StopRecoveryJob()
}

type paymentRecoveryService struct {
	paymentRepo  repository.PaymentRepository
	orderRepo    repository.OrderRepository
	syncRepo     repository.PaymentSyncRepository
	db           *sql.DB
	serverKey    string
	baseURL      string
	
	// Job control
	jobRunning   bool
	jobStop      chan struct{}
	jobMutex     sync.Mutex
}

func NewPaymentRecoveryService(
	paymentRepo repository.PaymentRepository,
	orderRepo repository.OrderRepository,
	syncRepo repository.PaymentSyncRepository,
	db *sql.DB,
) PaymentRecoveryService {
	baseURL := "https://api.sandbox.midtrans.com"
	if os.Getenv("MIDTRANS_ENVIRONMENT") == "production" {
		baseURL = "https://api.midtrans.com"
	}
	
	return &paymentRecoveryService{
		paymentRepo: paymentRepo,
		orderRepo:   orderRepo,
		syncRepo:    syncRepo,
		db:          db,
		serverKey:   os.Getenv("MIDTRANS_SERVER_KEY"),
		baseURL:     baseURL,
		jobStop:     make(chan struct{}),
	}
}

// SyncPaymentStatus syncs a single payment with Midtrans
func (s *paymentRecoveryService) SyncPaymentStatus(paymentID int, syncType string) (*dto.PaymentSyncResponse, error) {
	payment, err := s.paymentRepo.FindByOrderID(paymentID)
	if err != nil {
		// Try direct payment ID lookup
		query := `SELECT id, order_id, status, external_id FROM payments WHERE id = $1`
		var p struct {
			ID         int
			OrderID    int
			Status     string
			ExternalID string
		}
		err = s.db.QueryRow(query, paymentID).Scan(&p.ID, &p.OrderID, &p.Status, &p.ExternalID)
		if err != nil {
			return nil, fmt.Errorf("payment not found: %d", paymentID)
		}
		payment = &models.Payment{
			ID:         p.ID,
			OrderID:    p.OrderID,
			Status:     models.PaymentStatus(p.Status),
			ExternalID: p.ExternalID,
		}
	}

	order, err := s.orderRepo.FindByID(payment.OrderID)
	if err != nil {
		return nil, fmt.Errorf("order not found for payment: %d", paymentID)
	}

	// Create sync log
	syncLog := &models.PaymentSyncLog{
		PaymentID:          payment.ID,
		OrderID:            order.ID,
		OrderCode:          order.OrderCode,
		SyncType:           syncType,
		SyncStatus:         models.PaymentSyncStatusInProgress,
		LocalPaymentStatus: string(payment.Status),
		LocalOrderStatus:   string(order.Status),
	}
	s.syncRepo.Create(syncLog)

	// Call Midtrans status API
	gatewayStatus, gatewayResponse, err := s.checkMidtransStatus(order.OrderCode)
	if err != nil {
		s.syncRepo.UpdateSyncStatus(syncLog.ID, models.PaymentSyncStatusFailed, map[string]any{"error": err.Error()})
		return nil, err
	}

	syncLog.GatewayStatus = gatewayStatus
	syncLog.GatewayResponse = gatewayResponse

	// Check for mismatch
	localStatus := s.mapPaymentStatusToGateway(payment.Status)
	hasMismatch := localStatus != gatewayStatus

	response := &dto.PaymentSyncResponse{
		PaymentID:       payment.ID,
		OrderCode:       order.OrderCode,
		LocalStatus:     string(payment.Status),
		GatewayStatus:   gatewayStatus,
		HasMismatch:     hasMismatch,
		GatewayResponse: gatewayResponse,
	}

	if hasMismatch {
		mismatchType := fmt.Sprintf("local:%s vs gateway:%s", localStatus, gatewayStatus)
		s.syncRepo.MarkMismatch(syncLog.ID, mismatchType, fmt.Sprintf("Local status %s does not match gateway status %s", localStatus, gatewayStatus))
		response.MismatchType = mismatchType

		// Auto-resolve if gateway is authoritative
		if s.shouldAutoResolve(payment.Status, gatewayStatus) {
			err := s.resolveStatusMismatch(payment, order, gatewayStatus, gatewayResponse)
			if err == nil {
				s.syncRepo.MarkResolved(syncLog.ID, 0, "auto-resolved")
				response.Resolved = true
				response.ResolutionAction = "auto-resolved to gateway status"
			}
		}
	} else {
		s.syncRepo.UpdateSyncStatus(syncLog.ID, models.PaymentSyncStatusSynced, gatewayResponse)
		response.Resolved = true
	}

	// Update payment last synced
	s.db.Exec(`UPDATE payments SET last_synced_at = NOW(), sync_status = $1 WHERE id = $2`, 
		syncLog.SyncStatus, payment.ID)

	return response, nil
}

// SyncAllPendingPayments syncs all pending payments
func (s *paymentRecoveryService) SyncAllPendingPayments() (int, int, error) {
	query := `
		SELECT p.id FROM payments p
		JOIN orders o ON p.order_id = o.id
		WHERE p.status = 'PENDING' 
		AND p.created_at < NOW() - INTERVAL '30 minutes'
		AND (p.last_synced_at IS NULL OR p.last_synced_at < NOW() - INTERVAL '5 minutes')
		ORDER BY p.created_at ASC
		LIMIT 50
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()

	var paymentIDs []int
	for rows.Next() {
		var id int
		rows.Scan(&id)
		paymentIDs = append(paymentIDs, id)
	}

	synced := 0
	errors := 0

	for _, id := range paymentIDs {
		_, err := s.SyncPaymentStatus(id, "scheduled")
		if err != nil {
			errors++
			log.Printf("⚠️ Sync failed for payment %d: %v", id, err)
		} else {
			synced++
		}
		
		// Rate limiting
		time.Sleep(500 * time.Millisecond)
	}

	log.Printf("📊 Payment sync completed: %d synced, %d errors", synced, errors)
	return synced, errors, nil
}

// FindStuckPayments finds payments stuck in pending state
func (s *paymentRecoveryService) FindStuckPayments(hoursThreshold int) ([]*dto.StuckPaymentResponse, error) {
	query := `
		SELECT p.id, p.order_id, o.order_code, p.amount, p.status, p.created_at, p.last_synced_at,
		       EXTRACT(EPOCH FROM (NOW() - p.created_at)) / 3600 as hours_stuck,
		       COALESCE((SELECT retry_count FROM payment_sync_log WHERE payment_id = p.id ORDER BY created_at DESC LIMIT 1), 0) as retry_count
		FROM payments p
		JOIN orders o ON p.order_id = o.id
		WHERE p.status = 'PENDING'
		AND p.created_at < NOW() - INTERVAL '%d hours'
		ORDER BY p.created_at ASC
	`

	rows, err := s.db.Query(fmt.Sprintf(query, hoursThreshold))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stuck []*dto.StuckPaymentResponse
	for rows.Next() {
		var sp dto.StuckPaymentResponse
		err := rows.Scan(
			&sp.PaymentID, &sp.OrderID, &sp.OrderCode, &sp.Amount,
			&sp.Status, &sp.CreatedAt, &sp.LastSyncedAt, &sp.HoursStuck, &sp.RetryCount,
		)
		if err != nil {
			continue
		}
		stuck = append(stuck, &sp)
	}

	return stuck, nil
}

// FindOrphanOrders finds orders without payments
func (s *paymentRecoveryService) FindOrphanOrders() ([]int, error) {
	query := `
		SELECT o.id FROM orders o
		LEFT JOIN payments p ON o.id = p.order_id
		WHERE p.id IS NULL
		AND o.status = 'PENDING'
		AND o.created_at < NOW() - INTERVAL '1 hour'
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		rows.Scan(&id)
		ids = append(ids, id)
	}

	return ids, nil
}

// FindOrphanPayments finds payments without valid orders
func (s *paymentRecoveryService) FindOrphanPayments() ([]int, error) {
	query := `
		SELECT p.id FROM payments p
		LEFT JOIN orders o ON p.order_id = o.id
		WHERE o.id IS NULL
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		rows.Scan(&id)
		ids = append(ids, id)
	}

	return ids, nil
}

// RecoverStuckPayment attempts to recover a stuck payment
func (s *paymentRecoveryService) RecoverStuckPayment(paymentID int) error {
	resp, err := s.SyncPaymentStatus(paymentID, "recovery")
	if err != nil {
		return err
	}

	if resp.HasMismatch && !resp.Resolved {
		// Schedule for retry
		nextRetry := time.Now().Add(15 * time.Minute)
		s.db.Exec(`
			UPDATE payment_sync_log 
			SET next_retry_at = $1 
			WHERE payment_id = $2 AND resolved = false
			ORDER BY created_at DESC LIMIT 1
		`, nextRetry, paymentID)
	}

	return nil
}

// ResolveOrphanOrder handles an orphan order
func (s *paymentRecoveryService) ResolveOrphanOrder(orderID int) error {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return err
	}

	// If order is old and has no payment, expire it
	if time.Since(order.CreatedAt) > 24*time.Hour {
		s.orderRepo.UpdateStatus(orderID, models.OrderStatusExpired)
		if order.StockReserved {
			s.orderRepo.RestoreStock(orderID)
		}
		log.Printf("🗑️ Orphan order %s expired and stock restored", order.OrderCode)
	}

	return nil
}

// StartRecoveryJob starts the background recovery job
func (s *paymentRecoveryService) StartRecoveryJob(intervalMinutes int) {
	s.jobMutex.Lock()
	if s.jobRunning {
		s.jobMutex.Unlock()
		return
	}
	s.jobRunning = true
	s.jobStop = make(chan struct{})
	s.jobMutex.Unlock()

	go func() {
		ticker := time.NewTicker(time.Duration(intervalMinutes) * time.Minute)
		defer ticker.Stop()

		log.Printf("🔄 Payment recovery job started (interval: %d minutes)", intervalMinutes)

		for {
			select {
			case <-ticker.C:
				s.runRecoveryJob()
			case <-s.jobStop:
				log.Println("🛑 Payment recovery job stopped")
				return
			}
		}
	}()
}

// StopRecoveryJob stops the background recovery job
func (s *paymentRecoveryService) StopRecoveryJob() {
	s.jobMutex.Lock()
	defer s.jobMutex.Unlock()

	if s.jobRunning {
		close(s.jobStop)
		s.jobRunning = false
	}
}

func (s *paymentRecoveryService) runRecoveryJob() {
	log.Println("🔄 Running payment recovery job...")

	// 1. Sync pending payments
	synced, errors, _ := s.SyncAllPendingPayments()
	log.Printf("   Synced: %d, Errors: %d", synced, errors)

	// 2. Find and recover stuck payments
	stuck, _ := s.FindStuckPayments(2) // 2 hours threshold
	for _, sp := range stuck {
		if sp.RetryCount < 5 {
			s.RecoverStuckPayment(sp.PaymentID)
		}
	}
	log.Printf("   Stuck payments processed: %d", len(stuck))

	// 3. Handle orphan orders
	orphanOrders, _ := s.FindOrphanOrders()
	for _, id := range orphanOrders {
		s.ResolveOrphanOrder(id)
	}
	log.Printf("   Orphan orders resolved: %d", len(orphanOrders))

	// 4. Alert on critical issues
	if len(stuck) > 10 {
		s.sendAlert("HIGH_STUCK_PAYMENTS", fmt.Sprintf("%d payments stuck for >2 hours", len(stuck)))
	}

	log.Println("✅ Payment recovery job completed")
}

// Helper methods
func (s *paymentRecoveryService) checkMidtransStatus(orderCode string) (string, map[string]any, error) {
	url := fmt.Sprintf("%s/v2/%s/status", s.baseURL, orderCode)

	req, _ := http.NewRequest("GET", url, nil)
	auth := base64.StdEncoding.EncodeToString([]byte(s.serverKey + ":"))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var statusResp dto.MidtransStatusResponse
	if err := json.Unmarshal(body, &statusResp); err != nil {
		return "", nil, err
	}

	// Convert to map for storage
	var responseMap map[string]any
	json.Unmarshal(body, &responseMap)

	return statusResp.TransactionStatus, responseMap, nil
}

func (s *paymentRecoveryService) mapPaymentStatusToGateway(status models.PaymentStatus) string {
	switch status {
	case models.PaymentStatusPending:
		return "pending"
	case models.PaymentStatusSuccess:
		return "settlement"
	case models.PaymentStatusFailed:
		return "deny"
	case models.PaymentStatusExpired:
		return "expire"
	case models.PaymentStatusCancelled:
		return "cancel"
	default:
		return "pending"
	}
}

func (s *paymentRecoveryService) shouldAutoResolve(localStatus models.PaymentStatus, gatewayStatus string) bool {
	// Auto-resolve if local is pending but gateway has final status
	if localStatus == models.PaymentStatusPending {
		switch gatewayStatus {
		case "settlement", "capture", "deny", "cancel", "expire":
			return true
		}
	}
	return false
}

func (s *paymentRecoveryService) resolveStatusMismatch(payment *models.Payment, order *models.Order, gatewayStatus string, gatewayResponse map[string]any) error {
	var newPaymentStatus models.PaymentStatus
	var newOrderStatus models.OrderStatus

	switch gatewayStatus {
	case "settlement", "capture":
		newPaymentStatus = models.PaymentStatusSuccess
		newOrderStatus = models.OrderStatusPaid
	case "deny", "failure":
		newPaymentStatus = models.PaymentStatusFailed
		newOrderStatus = models.OrderStatusFailed
	case "cancel":
		newPaymentStatus = models.PaymentStatusCancelled
		newOrderStatus = models.OrderStatusCancelled
	case "expire":
		newPaymentStatus = models.PaymentStatusExpired
		newOrderStatus = models.OrderStatusExpired
	default:
		return nil // No action for pending
	}

	// Update payment
	s.paymentRepo.UpdateStatusWithResponse(payment.ID, newPaymentStatus, gatewayResponse)

	// Update order
	s.orderRepo.UpdateStatus(order.ID, newOrderStatus)

	// Restore stock if needed
	if newOrderStatus.RequiresStockRestore() && order.StockReserved {
		s.orderRepo.RestoreStock(order.ID)
	}

	log.Printf("🔧 Auto-resolved payment %d: %s -> %s", payment.ID, payment.Status, newPaymentStatus)
	return nil
}

func (s *paymentRecoveryService) sendAlert(alertType, message string) {
	// In production, this would send to Slack, PagerDuty, etc.
	log.Printf("🚨 ALERT [%s]: %s", alertType, message)
	
	// Could also write to a alerts table
	s.db.Exec(`
		INSERT INTO system_alerts (alert_type, message, created_at)
		VALUES ($1, $2, NOW())
	`, alertType, message)
}
