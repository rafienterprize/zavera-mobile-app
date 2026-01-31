package handler

import (
	"net/http"
	"zavera/service"

	"github.com/gin-gonic/gin"
)

type SystemHealthHandler struct {
	healthService service.SystemHealthService
}

func NewSystemHealthHandler(healthService service.SystemHealthService) *SystemHealthHandler {
	return &SystemHealthHandler{
		healthService: healthService,
	}
}

// GetSystemHealth returns system health metrics
// GET /api/admin/system/health
func (h *SystemHealthHandler) GetSystemHealth(c *gin.Context) {
	health, err := h.healthService.GetSystemHealth()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, health)
}

// GetCourierPerformance returns courier performance analytics
// GET /api/admin/analytics/courier-performance
func (h *SystemHealthHandler) GetCourierPerformance(c *gin.Context) {
	performance, err := h.healthService.GetCourierPerformance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "server_error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, performance)
}
