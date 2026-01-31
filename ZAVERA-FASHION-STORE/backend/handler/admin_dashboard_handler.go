package handler

import (
	"net/http"
	"zavera/service"

	"github.com/gin-gonic/gin"
)

type AdminDashboardHandler struct {
	dashboardService service.AdminDashboardService
}

func NewAdminDashboardHandler(dashboardService service.AdminDashboardService) *AdminDashboardHandler {
	return &AdminDashboardHandler{
		dashboardService: dashboardService,
	}
}

// GetExecutiveDashboard returns executive-level financial and operational metrics
// GET /api/admin/dashboard/executive
func (h *AdminDashboardHandler) GetExecutiveDashboard(c *gin.Context) {
	period := c.DefaultQuery("period", "today") // today, week, month, year

	metrics, err := h.dashboardService.GetExecutiveMetrics(period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetPaymentMonitor returns real-time payment monitoring data
// GET /api/admin/dashboard/payments
func (h *AdminDashboardHandler) GetPaymentMonitor(c *gin.Context) {
	monitor, err := h.dashboardService.GetPaymentMonitor()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, monitor)
}

// GetInventoryAlerts returns low stock and out of stock products
// GET /api/admin/dashboard/inventory
func (h *AdminDashboardHandler) GetInventoryAlerts(c *gin.Context) {
	alerts, err := h.dashboardService.GetInventoryAlerts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, alerts)
}

// GetCustomerInsights returns customer analytics and segmentation
// GET /api/admin/dashboard/customers
func (h *AdminDashboardHandler) GetCustomerInsights(c *gin.Context) {
	insights, err := h.dashboardService.GetCustomerInsights()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, insights)
}

// GetConversionFunnel returns conversion funnel metrics
// GET /api/admin/dashboard/funnel
func (h *AdminDashboardHandler) GetConversionFunnel(c *gin.Context) {
	period := c.DefaultQuery("period", "today")

	funnel, err := h.dashboardService.GetConversionFunnel(period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, funnel)
}

// GetRevenueChart returns revenue data for charting
// GET /api/admin/dashboard/revenue-chart
func (h *AdminDashboardHandler) GetRevenueChart(c *gin.Context) {
	period := c.DefaultQuery("period", "7days") // 7days, 30days, 90days, year

	chart, err := h.dashboardService.GetRevenueChart(period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, chart)
}
