package main

import (
	"log"
	"os"
	"zavera/config"
	"zavera/repository"
	"zavera/routes"
	"zavera/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ Warning: .env file not found")
	}

	// Connect to database
	db, err := config.ConnectDatabase()
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize Gin router
	router := gin.Default()

	// Configure CORS
	corsConfig := cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Session-ID"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}
	router.Use(cors.New(corsConfig))

	// Start SSE broker for admin notifications BEFORE setting up routes
	sseBroker := service.GetSSEBroker()
	sseBroker.Start()
	defer sseBroker.Shutdown()
	log.Println("📡 SSE Broker initialized and started")

	// Setup routes
	routes.SetupRoutes(router, db)

	// Start tracking job scheduler (if enabled)
	if os.Getenv("ENABLE_TRACKING_JOB") == "true" {
		shippingRepo := repository.NewShippingRepository(db)
		cartRepo := repository.NewCartRepository(db)
		productRepo := repository.NewProductRepository(db)
		orderRepo := repository.NewOrderRepository(db)
		
		trackingJob := service.NewTrackingJobRunner(shippingRepo, cartRepo, productRepo, orderRepo)
		trackingJob.Start()
		defer trackingJob.Stop()
		log.Println("📦 Tracking job scheduler enabled")
	}

	// Start order expiry job (Tokopedia-style: auto-cancel unpaid orders after 24h)
	{
		orderRepo := repository.NewOrderRepository(db)
		expiryJob := service.NewOrderExpiryJob(orderRepo)
		expiryJob.Start()
		defer expiryJob.Stop()
	}

	// Start payment expiry job (auto-expire VA/QRIS payments past expiry_time)
	{
		orderRepo := repository.NewOrderRepository(db)
		paymentExpiryJob := service.NewPaymentExpiryJob(db, orderRepo)
		paymentExpiryJob.Start()
		defer paymentExpiryJob.Stop()
	}

	// Start server
	log.Println("🚀 Server starting on :8080...")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}
