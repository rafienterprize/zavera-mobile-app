package handler

import (
	"net/http"
	"strconv"
	"zavera/dto"

	"github.com/gin-gonic/gin"
)

type AdminCustomerHandler struct {
	customerService CustomerService
}

type CustomerService interface {
	GetCustomers(page, limit int, search, segment string) (*dto.CustomersResponse, error)
	GetCustomerStats() (*dto.CustomerStatsResponse, error)
	GetCustomerDetail(userID int) (*dto.CustomerDetailResponse, error)
	ExportCustomers() ([]byte, error)
}

func NewAdminCustomerHandler(customerService CustomerService) *AdminCustomerHandler {
	return &AdminCustomerHandler{
		customerService: customerService,
	}
}

// GetCustomers returns paginated list of customers
// GET /api/admin/customers
func (h *AdminCustomerHandler) GetCustomers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")
	segment := c.Query("segment")

	customers, err := h.customerService.GetCustomers(page, limit, search, segment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "server_error",
			Message: "Failed to fetch customers",
		})
		return
	}

	stats, _ := h.customerService.GetCustomerStats()

	c.JSON(http.StatusOK, gin.H{
		"customers":   customers.Customers,
		"total":       customers.Total,
		"page":        page,
		"limit":       limit,
		"total_pages": (customers.Total + limit - 1) / limit,
		"stats":       stats,
	})
}

// GetCustomerDetail returns detailed customer information
// GET /api/admin/customers/:id
func (h *AdminCustomerHandler) GetCustomerDetail(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid customer ID",
		})
		return
	}

	detail, err := h.customerService.GetCustomerDetail(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Message: "Customer not found",
		})
		return
	}

	c.JSON(http.StatusOK, detail)
}

// ExportCustomers exports customer list as CSV
// GET /api/admin/customers/export
func (h *AdminCustomerHandler) ExportCustomers(c *gin.Context) {
	csvData, err := h.customerService.ExportCustomers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "export_failed",
			Message: "Failed to export customers",
		})
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=customers.csv")
	c.Data(http.StatusOK, "text/csv", csvData)
}
