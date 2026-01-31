package routes

import (
	"database/sql"
	"os"
	"zavera/handler"
	"zavera/repository"
	"zavera/service"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, db *sql.DB) {
	// Initialize repositories
	productRepo := repository.NewProductRepository(db)
	variantRepo := repository.NewVariantRepository(db)
	cartRepo := repository.NewCartRepository(db)
	wishlistRepo := repository.NewWishlistRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	userRepo := repository.NewUserRepository(db)
	shippingRepo := repository.NewShippingRepository(db)
	emailRepo := repository.NewEmailRepository(db)

	// Initialize Core Payment repository
	orderPaymentRepo := repository.NewOrderPaymentRepository(db)

	// Initialize services
	productService := service.NewProductService(productRepo, variantRepo)
	variantService := service.NewVariantService(variantRepo, productRepo)
	cartService := service.NewCartService(cartRepo, productRepo)
	wishlistService := service.NewWishlistService(wishlistRepo, productRepo, cartRepo)
	orderService := service.NewOrderService(orderRepo, cartRepo, productRepo)
	paymentService := service.NewPaymentService(paymentRepo, orderRepo, shippingRepo, emailRepo)
	authService := service.NewAuthService(userRepo, shippingRepo)
	shippingService := service.NewShippingService(shippingRepo, cartRepo, productRepo, orderRepo)
	checkoutService := service.NewCheckoutService(orderRepo, cartRepo, productRepo, shippingRepo, emailRepo)

	// Initialize Core Payment service (Tokopedia-style VA payments)
	serverKey := os.Getenv("MIDTRANS_SERVER_KEY")
	emailService := service.NewEmailService(emailRepo)
	corePaymentService := service.NewCorePaymentService(orderPaymentRepo, orderRepo, serverKey, emailService)

	// Admin services
	adminProductService := service.NewAdminProductService(db)
	adminOrderService := service.NewAdminOrderService(db, orderRepo, paymentRepo, shippingRepo, emailRepo, shippingService)

	// Initialize handlers
	productHandler := handler.NewProductHandler(productService)
	variantHandler := handler.NewVariantHandler(variantService)
	cartHandler := handler.NewCartHandler(cartService)
	wishlistHandler := handler.NewWishlistHandler(wishlistService)
	orderHandler := handler.NewOrderHandler(orderService, paymentService, shippingService)
	paymentHandler := handler.NewPaymentHandler(paymentService)
	authHandler := handler.NewAuthHandler(authService, cartService)
	shippingHandler := handler.NewShippingHandler(shippingService)
	checkoutHandler := handler.NewCheckoutHandler(checkoutService, shippingService)
	trackingHandler := handler.NewTrackingHandler(shippingService, orderService)

	// Admin handlers
	adminProductHandler := handler.NewAdminProductHandler(adminProductService)
	adminOrderHandler := handler.NewAdminOrderHandler(adminOrderService)

	// Core Payment handler (Tokopedia-style VA payments)
	corePaymentHandler := handler.NewCorePaymentHandler(corePaymentService)

	// API routes
	api := router.Group("/api")
	{
		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/google", authHandler.GoogleLogin)
			auth.GET("/verify-email", authHandler.VerifyEmail)
			auth.POST("/resend-verification", authHandler.ResendVerification)
			auth.GET("/me", authHandler.AuthMiddleware(), authHandler.GetMe)
		}

		// User routes (protected)
		user := api.Group("/user")
		user.Use(authHandler.AuthMiddleware())
		{
			user.GET("/orders", authHandler.GetUserOrders)
			// User addresses
			user.GET("/addresses", shippingHandler.GetUserAddresses)
			user.POST("/addresses", shippingHandler.CreateAddress)
			user.GET("/addresses/:id", shippingHandler.GetAddress)
			user.PUT("/addresses/:id", shippingHandler.UpdateAddress)
			user.DELETE("/addresses/:id", shippingHandler.DeleteAddress)
			user.POST("/addresses/:id/default", shippingHandler.SetDefaultAddress)
		}

		// Customer refund routes (protected)
		customer := api.Group("/customer")
		customer.Use(authHandler.AuthMiddleware())
		{
			// Initialize refund repositories and services for customer endpoints
			refundRepo := repository.NewRefundRepository(db)
			auditRepo := repository.NewAdminAuditRepository(db)
			refundSvc := service.NewRefundService(refundRepo, orderRepo, paymentRepo, auditRepo)
			
			// Initialize customer refund handler
			customerRefundHandler := handler.NewCustomerRefundHandler(refundSvc, orderService)
			
			customer.GET("/orders/:code/refunds", customerRefundHandler.GetOrderRefunds)
			customer.GET("/refunds/:code", customerRefundHandler.GetRefundByCode)
		}

		// Product routes
		products := api.Group("/products")
		{
			products.GET("", productHandler.GetProducts)
			products.GET("/:id", productHandler.GetProductByID)
			products.GET("/:id/variants", variantHandler.GetProductVariants)
			products.GET("/:id/with-variants", variantHandler.GetProductWithVariants)
			products.GET("/:id/options", variantHandler.GetAvailableOptions)
			products.POST("/variants/find", variantHandler.FindVariant)
		}

		// Variant routes (public)
		variants := api.Group("/variants")
		{
			variants.GET("/:id", variantHandler.GetVariant)
			variants.GET("/sku/:sku", variantHandler.GetVariantBySKU)
			variants.GET("/:id/images", variantHandler.GetVariantImages)
			variants.POST("/check-availability", variantHandler.CheckAvailability)
			variants.GET("/attributes", variantHandler.GetVariantAttributes)
		}

		// Cart routes (with optional auth to persist cart for logged-in users)
		cart := api.Group("")
		cart.Use(authHandler.OptionalAuthMiddleware())
		{
			cart.GET("/cart", cartHandler.GetCart)
			cart.POST("/cart/items", cartHandler.AddToCart)
			cart.PUT("/cart/items/:id", cartHandler.UpdateCartItem)
			cart.DELETE("/cart/items/:id", cartHandler.RemoveFromCart)
			cart.DELETE("/cart", cartHandler.ClearCart)
			cart.GET("/cart/validate", cartHandler.ValidateCart)
		}

		// Wishlist routes (requires authentication)
		wishlist := api.Group("/wishlist")
		wishlist.Use(authHandler.AuthMiddleware())
		{
			wishlist.GET("", wishlistHandler.GetWishlist)
			wishlist.POST("", wishlistHandler.AddToWishlist)
			wishlist.DELETE("/:productId", wishlistHandler.RemoveFromWishlist)
			wishlist.POST("/:productId/move-to-cart", wishlistHandler.MoveToCart)
		}

		// Shipping routes (public, but with optional auth for user cart lookup)
		shipping := api.Group("/shipping")
		{
			shipping.GET("/providers", shippingHandler.GetProviders)
			shipping.GET("/areas", shippingHandler.SearchAreas)                // Biteship area search
			shipping.GET("/provinces", shippingHandler.GetProvinces)
			shipping.GET("/cities", shippingHandler.GetCities)
			shipping.GET("/districts", shippingHandler.GetDistricts)         // Kecamatan from Kommerce API (legacy)
			shipping.GET("/kelurahan", shippingHandler.GetSubdistrictsAPI)   // Kelurahan from Kommerce API (legacy)
			shipping.GET("/subdistricts", shippingHandler.GetSubdistricts)   // From local database (legacy)
			shipping.POST("/rates", authHandler.OptionalAuthMiddleware(), shippingHandler.GetShippingRates)
			shipping.GET("/preview", shippingHandler.GetCartShippingPreview)
		}

		// Checkout routes (with shipping)
		checkout := api.Group("/checkout")
		{
			checkout.GET("/shipping-options", checkoutHandler.GetShippingOptions)
			checkout.POST("/shipping", authHandler.OptionalAuthMiddleware(), checkoutHandler.CheckoutWithShipping)
		}

		// Legacy checkout (backward compatibility)
		api.POST("/checkout", authHandler.OptionalAuthMiddleware(), orderHandler.Checkout)
		
		// Order routes
		api.GET("/orders/:code", authHandler.OptionalAuthMiddleware(), orderHandler.GetOrder)
		api.GET("/orders/id/:id", authHandler.AuthMiddleware(), orderHandler.GetOrderByID)

		// Shipment routes (public - for tracking)
		shipments := api.Group("/shipments")
		{
			shipments.GET("/:id", shippingHandler.GetShipment)
			shipments.POST("/:id/refresh", shippingHandler.RefreshTracking)
		}
		
		// Public tracking route (by resi number)
		api.GET("/tracking/:resi", trackingHandler.GetTrackingByResi)

		// Payment routes (new Midtrans Snap integration)
		payments := api.Group("/payments")
		{
			payments.POST("/initiate", paymentHandler.InitiatePayment)
			payments.POST("/webhook", paymentHandler.Webhook)
		}

		// Core Payment routes (Tokopedia-style VA payments via Midtrans Core API)
		corePayments := api.Group("/payments/core")
		{
			// Public webhook endpoint (no auth required)
			corePayments.POST("/webhook", corePaymentHandler.CoreWebhook)
		}
		
		// Authenticated Core Payment routes
		corePaymentsAuth := api.Group("/payments/core")
		corePaymentsAuth.Use(authHandler.AuthMiddleware())
		{
			corePaymentsAuth.POST("/create", corePaymentHandler.CreateVAPayment)
			corePaymentsAuth.GET("/:order_id", corePaymentHandler.GetPaymentDetails)
			corePaymentsAuth.POST("/check", corePaymentHandler.CheckPaymentStatus)
		}

		// Pembelian routes (authenticated)
		pembelian := api.Group("/pembelian")
		pembelian.Use(authHandler.AuthMiddleware())
		{
			pembelian.GET("/pending", corePaymentHandler.GetPendingOrders)
			pembelian.GET("/history", corePaymentHandler.GetTransactionHistory)
		}

		// Midtrans webhook endpoint
		api.POST("/midtrans/webhook", paymentHandler.Webhook)

		// Midtrans Core API webhook endpoint (Tokopedia-style)
		api.POST("/webhook/midtrans/core", corePaymentHandler.CoreWebhook)

		// Legacy payment callback - DEPRECATED, use /api/payments/webhook instead
		// Kept for backward compatibility but with signature verification
		api.POST("/payment/callback", paymentHandler.Webhook)

		// SSE for admin notifications (token via Authorization header)
		api.GET("/admin/events", handler.HandleAdminSSE)

		// Admin routes (protected with auth + admin middleware)
		// Only ADMIN_GOOGLE_EMAIL can access these routes
		admin := api.Group("/admin")
		admin.Use(authHandler.AuthMiddleware())
		admin.Use(authHandler.AdminMiddleware())
		{
			// === ADMIN PRODUCT MANAGEMENT ===
			admin.GET("/products", adminProductHandler.GetAllProducts)
			admin.POST("/products", adminProductHandler.CreateProduct)
			admin.POST("/products/upload-image", adminProductHandler.UploadProductImage)
			admin.PUT("/products/:id", adminProductHandler.UpdateProduct)
			admin.PATCH("/products/:id/stock", adminProductHandler.UpdateStock)
			admin.DELETE("/products/:id", adminProductHandler.DeleteProduct)
			admin.POST("/products/:id/images", adminProductHandler.AddProductImage)
			admin.DELETE("/products/:id/images/:imageId", adminProductHandler.DeleteProductImage)

			// === ADMIN VARIANT MANAGEMENT ===
			admin.POST("/variants", variantHandler.CreateVariant)
			admin.PUT("/variants/:id", variantHandler.UpdateVariant)
			admin.DELETE("/variants/:id", variantHandler.DeleteVariant)
			admin.POST("/variants/bulk-generate", variantHandler.BulkGenerateVariants)
			admin.POST("/variants/images", variantHandler.AddVariantImage)
			admin.DELETE("/variants/images/:imageId", variantHandler.DeleteVariantImage)
			admin.POST("/variants/images/:variantId/primary", variantHandler.SetPrimaryImage)
			admin.POST("/variants/images/:variantId/reorder", variantHandler.ReorderImages)
			admin.PUT("/variants/stock/:id", variantHandler.UpdateStock)
			admin.POST("/variants/stock/:id/adjust", variantHandler.AdjustStock)
			admin.GET("/variants/low-stock", variantHandler.GetLowStockVariants)
			admin.GET("/variants/stock-summary/:id", variantHandler.GetStockSummary)
			admin.POST("/variants/reserve-stock", variantHandler.ReserveStock)

			// === ADMIN ORDER MANAGEMENT ===
			admin.GET("/orders", adminOrderHandler.GetAllOrders)
			admin.GET("/orders/stats", adminOrderHandler.GetOrderStats)
			admin.GET("/orders/:code", adminOrderHandler.GetOrderDetail)
			admin.PATCH("/orders/:code/status", adminOrderHandler.UpdateOrderStatus)
			admin.POST("/orders/:code/pack", adminOrderHandler.PackOrder)
			admin.POST("/orders/:code/generate-resi", adminOrderHandler.GenerateResi)
			admin.POST("/orders/:code/ship", adminOrderHandler.ShipOrder)
			admin.POST("/orders/:code/deliver", adminOrderHandler.DeliverOrder)
			admin.GET("/orders/:code/actions", adminOrderHandler.GetOrderActions)

			// Shipment management
			admin.PUT("/shipments/:id/tracking", shippingHandler.AdminUpdateTracking)
			admin.POST("/shipments/:id/ship", shippingHandler.AdminMarkShipped)
			admin.POST("/shipping/tracking-job", shippingHandler.AdminRunTrackingJob)

			// === COMMERCIAL HARDENING - PHASE 1 ===
			// Initialize hardening repositories
			refundRepo := repository.NewRefundRepository(db)
			auditRepo := repository.NewAdminAuditRepository(db)
			syncRepo := repository.NewPaymentSyncRepository(db)
			reconciliationRepo := repository.NewReconciliationRepository(db)

			// Initialize hardening services
			refundSvc := service.NewRefundService(refundRepo, orderRepo, paymentRepo, auditRepo)
			adminSvc := service.NewAdminService(orderRepo, paymentRepo, refundRepo, auditRepo, shippingRepo, refundSvc, db)
			recoverySvc := service.NewPaymentRecoveryService(paymentRepo, orderRepo, syncRepo, db)
			reconciliationSvc := service.NewReconciliationService(reconciliationRepo, syncRepo, db)

			// Initialize hardening handler
			hardeningHandler := handler.NewAdminHardeningHandler(adminSvc, refundSvc, recoverySvc, reconciliationSvc, db)

			// Force Actions
			admin.POST("/orders/:code/force-cancel", hardeningHandler.ForceCancel)
			admin.POST("/orders/:code/refund", hardeningHandler.ForceRefund)
			admin.POST("/orders/:code/reship", hardeningHandler.ForceReship)
			admin.POST("/payments/:id/reconcile", hardeningHandler.ReconcilePayment)

			// Refund Management
			refundHandler := handler.NewAdminRefundHandler(refundSvc)
			admin.POST("/refunds", hardeningHandler.CreateRefund)
			admin.GET("/refunds", refundHandler.ListRefunds)
			admin.GET("/refunds/:id", refundHandler.GetRefund)
			admin.POST("/refunds/:id/process", hardeningHandler.ProcessRefund)
			admin.POST("/refunds/:id/retry", refundHandler.RetryRefund)
			admin.POST("/refunds/:id/mark-completed", refundHandler.MarkRefundCompleted)
			admin.GET("/orders/:code/refunds", refundHandler.GetOrderRefunds)

			// Payment Recovery
			admin.POST("/payments/:id/sync", hardeningHandler.SyncPayment)
			admin.GET("/payments/stuck", hardeningHandler.GetStuckPayments)
			admin.POST("/payments/sync-all", hardeningHandler.RunPaymentSync)

			// Reconciliation
			admin.POST("/reconciliation/run", hardeningHandler.RunReconciliation)
			admin.GET("/reconciliation", hardeningHandler.GetReconciliation)
			admin.GET("/reconciliation/mismatches", hardeningHandler.GetMismatches)

			// Audit Logs
			admin.GET("/audit-logs", hardeningHandler.GetAuditLogs)

			// === EXECUTIVE DASHBOARD - P0 FEATURES ===
			adminDashboardService := service.NewAdminDashboardService(db, orderRepo, productRepo)
			adminDashboardHandler := handler.NewAdminDashboardHandler(adminDashboardService)

			admin.GET("/dashboard/executive", adminDashboardHandler.GetExecutiveDashboard)
			admin.GET("/dashboard/payments", adminDashboardHandler.GetPaymentMonitor)
			admin.GET("/dashboard/inventory", adminDashboardHandler.GetInventoryAlerts)
			admin.GET("/dashboard/customers", adminDashboardHandler.GetCustomerInsights)
			admin.GET("/dashboard/funnel", adminDashboardHandler.GetConversionFunnel)
			admin.GET("/dashboard/revenue-chart", adminDashboardHandler.GetRevenueChart)

			// Start background jobs if enabled
			if os.Getenv("ENABLE_RECOVERY_JOB") == "true" {
				recoverySvc.StartRecoveryJob(15) // Every 15 minutes
			}
			if os.Getenv("ENABLE_RECONCILIATION_JOB") == "true" {
				reconciliationSvc.StartReconciliationJob(2) // Run at 2 AM
			}

			// === SHIPPING & FULFILLMENT HARDENING - PHASE 2 ===
			disputeRepo := repository.NewDisputeRepository(db)

			// Initialize fulfillment services
			fulfillmentSvc := service.NewFulfillmentService(shippingRepo, disputeRepo, orderRepo, auditRepo, db)
			disputeSvc := service.NewDisputeService(disputeRepo, orderRepo, shippingRepo, refundSvc, db)
			monitorSvc := service.NewShipmentMonitorService(shippingRepo, disputeRepo, orderRepo, db)

			// Link services to avoid circular dependency
			disputeSvc.SetFulfillmentService(fulfillmentSvc)

			// Initialize fulfillment handler
			fulfillmentHandler := handler.NewFulfillmentHandler(fulfillmentSvc, disputeSvc, monitorSvc)

			// Shipment Control Endpoints
			// List endpoint must come BEFORE parameterized routes
			admin.GET("/shipments", fulfillmentHandler.GetShipmentsList)
			admin.GET("/shipments/stuck", fulfillmentHandler.GetStuckShipments)
			admin.GET("/shipments/pickup-failures", fulfillmentHandler.GetPickupFailures)
			admin.GET("/shipments/:id/details", fulfillmentHandler.GetShipmentDetails)
			admin.POST("/shipments/:id/investigate", fulfillmentHandler.InvestigateShipment)
			admin.POST("/shipments/:id/mark-lost", fulfillmentHandler.MarkLost)
			admin.POST("/shipments/:id/reship", fulfillmentHandler.Reship)
			admin.POST("/shipments/:id/override-status", fulfillmentHandler.OverrideStatus)
			admin.POST("/shipments/:id/schedule-pickup", fulfillmentHandler.SchedulePickup)
			admin.POST("/shipments/:id/mark-shipped", fulfillmentHandler.MarkShipped)

			// Dispute Endpoints
			admin.POST("/disputes", fulfillmentHandler.CreateDispute)
			admin.GET("/disputes/open", fulfillmentHandler.GetOpenDisputes)
			admin.GET("/disputes/:id", fulfillmentHandler.GetDispute)
			admin.GET("/disputes/code/:code", fulfillmentHandler.GetDisputeByCode)
			admin.POST("/disputes/:id/investigate", fulfillmentHandler.StartDisputeInvestigation)
			admin.POST("/disputes/:id/request-evidence", fulfillmentHandler.RequestEvidence)
			admin.POST("/disputes/:id/resolve", fulfillmentHandler.ResolveDispute)
			admin.POST("/disputes/:id/close", fulfillmentHandler.CloseDispute)
			admin.POST("/disputes/:id/messages", fulfillmentHandler.AddDisputeMessage)
			admin.GET("/disputes/:id/messages", fulfillmentHandler.GetDisputeMessages)

			// Fulfillment Dashboard & Monitoring
			admin.GET("/fulfillment/dashboard", fulfillmentHandler.GetFulfillmentDashboard)
			admin.POST("/fulfillment/run-monitors", fulfillmentHandler.RunMonitoringJob)

			// System Health & Analytics
			systemHealthService := service.NewSystemHealthService(db)
			systemHealthHandler := handler.NewSystemHealthHandler(systemHealthService)
			admin.GET("/system/health", systemHealthHandler.GetSystemHealth)
			admin.GET("/analytics/courier-performance", systemHealthHandler.GetCourierPerformance)

			// Start shipment monitoring jobs if enabled
			if os.Getenv("ENABLE_SHIPMENT_MONITOR") == "true" {
				monitorSvc.StartMonitoringJobs(30) // Every 30 minutes
			}
		}
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Debug endpoints (development only)
	debug := router.Group("/debug")
	{
		debugHandler := handler.NewDebugHandler()
		debug.GET("/midtrans", debugHandler.TestMidtrans)
	}
}

