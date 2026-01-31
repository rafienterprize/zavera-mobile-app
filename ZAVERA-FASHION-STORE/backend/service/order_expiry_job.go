package service

import (
	"log"
	"time"
	"zavera/models"
	"zavera/repository"
)

// OrderExpiryJob handles automatic expiration of unpaid orders
// Tokopedia-style: Orders not paid within 24 hours are automatically cancelled
type OrderExpiryJob struct {
	orderRepo repository.OrderRepository
	ticker    *time.Ticker
	done      chan bool
}

func NewOrderExpiryJob(orderRepo repository.OrderRepository) *OrderExpiryJob {
	return &OrderExpiryJob{
		orderRepo: orderRepo,
		done:      make(chan bool),
	}
}

// Start begins the order expiry job scheduler
// Runs every 5 minutes to check for expired orders
func (j *OrderExpiryJob) Start() {
	j.ticker = time.NewTicker(5 * time.Minute)
	
	// Run immediately on start
	go j.expireOldOrders()
	
	go func() {
		for {
			select {
			case <-j.done:
				return
			case <-j.ticker.C:
				j.expireOldOrders()
			}
		}
	}()
	
	log.Println("⏰ Order expiry job started (checks every 5 minutes)")
}

// Stop stops the order expiry job scheduler
func (j *OrderExpiryJob) Stop() {
	if j.ticker != nil {
		j.ticker.Stop()
	}
	j.done <- true
	log.Println("⏰ Order expiry job stopped")
}

// expireOldOrders finds and expires orders older than 24 hours that are still PENDING
// This handles orders that never selected a payment method (stayed in PENDING status)
func (j *OrderExpiryJob) expireOldOrders() {
	log.Println("🔍 Checking for expired pending orders...")
	
	// Find orders that are PENDING and older than 24 hours
	expiredOrders, err := j.orderRepo.FindExpiredPendingOrders(24 * time.Hour)
	if err != nil {
		log.Printf("⚠️ Failed to find expired orders: %v", err)
		return
	}
	
	if len(expiredOrders) == 0 {
		log.Println("✅ No expired orders found")
		return
	}
	
	log.Printf("📋 Found %d expired pending orders", len(expiredOrders))
	
	var expiredCount int
	for _, order := range expiredOrders {
		// Idempotency check: skip if already expired
		if order.Status == models.OrderStatusExpired || order.Status == models.OrderStatusKadaluarsa {
			log.Printf("⏭️ Order %s already expired, skipping", order.OrderCode)
			continue
		}
		
		// Mark as expired
		err := j.orderRepo.MarkAsExpired(order.ID)
		if err != nil {
			log.Printf("⚠️ Failed to expire order %s: %v", order.OrderCode, err)
			continue
		}
		
		// Restore stock if reserved
		if order.StockReserved {
			err = j.orderRepo.RestoreStock(order.ID)
			if err != nil {
				log.Printf("⚠️ Failed to restore stock for order %s: %v", order.OrderCode, err)
			}
		}
		
		// Record status change
		j.orderRepo.RecordStatusChange(order.ID, order.Status, models.OrderStatusExpired, "order_expiry_job", "Order expired - payment method not selected within 24 hours")
		
		expiredCount++
		log.Printf("🗑️ Order %s expired (created: %s)", order.OrderCode, order.CreatedAt.Format("2006-01-02 15:04"))
	}
	
	if expiredCount > 0 {
		log.Printf("✅ Expired %d orders", expiredCount)
	}
}
