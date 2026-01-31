package service

import (
	"database/sql"
	"log"
	"time"
	"zavera/repository"
)

// PaymentExpiryJob handles automatic expiration of pending payments
// Runs every minute to check for expired VA/QRIS payments
type PaymentExpiryJob struct {
	db        *sql.DB
	orderRepo repository.OrderRepository
	ticker    *time.Ticker
	done      chan bool
}

func NewPaymentExpiryJob(db *sql.DB, orderRepo repository.OrderRepository) *PaymentExpiryJob {
	return &PaymentExpiryJob{
		db:        db,
		orderRepo: orderRepo,
		done:      make(chan bool),
	}
}

// Start begins the payment expiry job scheduler
// Runs every 1 minute to check for expired payments
func (j *PaymentExpiryJob) Start() {
	j.ticker = time.NewTicker(1 * time.Minute)

	// Run immediately on start
	go j.expirePayments()

	go func() {
		for {
			select {
			case <-j.done:
				return
			case <-j.ticker.C:
				j.expirePayments()
			}
		}
	}()

	log.Println("⏰ Payment expiry job started (checks every 1 minute)")
}

// Stop stops the payment expiry job scheduler
func (j *PaymentExpiryJob) Stop() {
	if j.ticker != nil {
		j.ticker.Stop()
	}
	j.done <- true
	log.Println("⏰ Payment expiry job stopped")
}

// expirePayments finds and expires payments that have passed their expiry_time
func (j *PaymentExpiryJob) expirePayments() {
	log.Println("🔍 Checking for expired payments...")

	// Find PENDING payments that have expired
	rows, err := j.db.Query(`
		SELECT p.id, p.order_id, o.order_code, o.stock_reserved
		FROM order_payments p
		JOIN orders o ON p.order_id = o.id
		WHERE p.payment_status = 'PENDING' 
		  AND p.expiry_time < NOW()
		  AND o.status = 'MENUNGGU_PEMBAYARAN'
		ORDER BY p.expiry_time ASC
		LIMIT 100
	`)
	if err != nil {
		log.Printf("⚠️ Failed to find expired payments: %v", err)
		return
	}
	defer rows.Close()

	var expiredCount int
	for rows.Next() {
		var paymentID, orderID int
		var orderCode string
		var stockReserved bool

		if err := rows.Scan(&paymentID, &orderID, &orderCode, &stockReserved); err != nil {
			log.Printf("⚠️ Failed to scan expired payment: %v", err)
			continue
		}

		// Process expiry in a transaction
		if err := j.processExpiry(paymentID, orderID, orderCode, stockReserved); err != nil {
			log.Printf("⚠️ Failed to expire payment %d: %v", paymentID, err)
			continue
		}

		expiredCount++
		log.Printf("🗑️ Payment %d expired for order %s", paymentID, orderCode)
	}

	if expiredCount > 0 {
		log.Printf("✅ Expired %d payments", expiredCount)
	} else {
		log.Println("✅ No expired payments found")
	}
}

// processExpiry handles a single payment expiry with transaction
func (j *PaymentExpiryJob) processExpiry(paymentID, orderID int, orderCode string, stockReserved bool) error {
	tx, err := j.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Lock the payment row and verify it's still PENDING
	var currentStatus string
	err = tx.QueryRow(`
		SELECT payment_status FROM order_payments 
		WHERE id = $1 FOR UPDATE
	`, paymentID).Scan(&currentStatus)
	if err != nil {
		return err
	}

	// Idempotency: skip if already processed
	if currentStatus != "PENDING" {
		log.Printf("⏭️ Payment %d already %s, skipping", paymentID, currentStatus)
		return nil
	}

	// Update payment status to EXPIRED
	_, err = tx.Exec(`
		UPDATE order_payments 
		SET payment_status = 'EXPIRED', updated_at = NOW()
		WHERE id = $1
	`, paymentID)
	if err != nil {
		return err
	}

	// Update order status to KADALUARSA
	_, err = tx.Exec(`
		UPDATE orders 
		SET status = 'KADALUARSA', updated_at = NOW()
		WHERE id = $1 AND status = 'MENUNGGU_PEMBAYARAN'
	`, orderID)
	if err != nil {
		return err
	}

	// Restore stock if reserved
	if stockReserved {
		j.orderRepo.RestoreStockTx(tx, orderID)
	}

	// Record status change for audit
	_, _ = tx.Exec(`
		INSERT INTO order_status_history (order_id, from_status, to_status, changed_by, reason)
		VALUES ($1, 'MENUNGGU_PEMBAYARAN', 'KADALUARSA', 'payment_expiry_job', 'Payment expired - VA/QRIS not paid before expiry')
	`, orderID)

	// Log to sync log
	_, _ = tx.Exec(`
		INSERT INTO core_payment_sync_log (
			payment_id, order_id, order_code, sync_type, sync_status,
			local_payment_status, local_order_status, has_mismatch
		) VALUES ($1, $2, $3, 'expiry_job', 'SYNCED', 'EXPIRED', 'KADALUARSA', false)
	`, paymentID, orderID, orderCode)

	return tx.Commit()
}
