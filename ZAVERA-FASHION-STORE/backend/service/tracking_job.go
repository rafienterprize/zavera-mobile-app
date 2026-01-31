package service

import (
	"log"
	"time"
	"zavera/repository"
)

// TrackingJobRunner handles scheduled tracking updates
type TrackingJobRunner struct {
	shippingService ShippingService
	interval        time.Duration
	stopChan        chan struct{}
	running         bool
}

// NewTrackingJobRunner creates a new tracking job runner
func NewTrackingJobRunner(
	shippingRepo repository.ShippingRepository,
	cartRepo repository.CartRepository,
	productRepo repository.ProductRepository,
	orderRepo repository.OrderRepository,
) *TrackingJobRunner {
	return &TrackingJobRunner{
		shippingService: NewShippingService(shippingRepo, cartRepo, productRepo, orderRepo),
		interval:        30 * time.Minute, // Run every 30 minutes
		stopChan:        make(chan struct{}),
		running:         false,
	}
}

// Start begins the tracking job scheduler
func (r *TrackingJobRunner) Start() {
	if r.running {
		return
	}

	r.running = true
	log.Println("📦 Starting tracking job scheduler (interval: 30 minutes)")

	go func() {
		// Run immediately on start
		r.runJob()

		ticker := time.NewTicker(r.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				r.runJob()
			case <-r.stopChan:
				log.Println("📦 Tracking job scheduler stopped")
				return
			}
		}
	}()
}

// Stop stops the tracking job scheduler
func (r *TrackingJobRunner) Stop() {
	if !r.running {
		return
	}

	r.running = false
	close(r.stopChan)
}

func (r *TrackingJobRunner) runJob() {
	log.Println("📦 Running scheduled tracking job...")
	
	startTime := time.Now()
	err := r.shippingService.RunTrackingJob()
	duration := time.Since(startTime)

	if err != nil {
		log.Printf("⚠️ Tracking job completed with errors: %v (took %v)", err, duration)
	} else {
		log.Printf("✅ Tracking job completed successfully (took %v)", duration)
	}
}

// SetInterval changes the job interval
func (r *TrackingJobRunner) SetInterval(interval time.Duration) {
	r.interval = interval
}

// IsRunning returns whether the job is running
func (r *TrackingJobRunner) IsRunning() bool {
	return r.running
}
