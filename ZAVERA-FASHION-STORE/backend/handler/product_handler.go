package handler

import (
	"net/http"
	"strconv"
	"zavera/dto"
	"zavera/service"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	productService service.ProductService
}

func NewProductHandler(productService service.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

// GetProducts godoc
// @Summary Get all products
// @Description Get list of all active products, optionally filtered by category
// @Tags products
// @Accept json
// @Produce json
// @Param category query string false "Filter by category (wanita, pria, anak, sports, luxury, beauty)"
// @Success 200 {array} dto.ProductResponse
// @Router /api/products [get]
func (h *ProductHandler) GetProducts(c *gin.Context) {
	category := c.Query("category")
	
	var products []dto.ProductResponse
	var err error
	
	if category != "" {
		products, err = h.productService.GetProductsByCategory(category)
	} else {
		products, err = h.productService.GetAllProducts()
	}
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, products)
}

// GetProductByID godoc
// @Summary Get product by ID
// @Description Get product details by ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} dto.ProductResponse
// @Router /api/products/{id} [get]
func (h *ProductHandler) GetProductByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid product ID",
		})
		return
	}

	product, err := h.productService.GetProductByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Message: "Product not found",
		})
		return
	}

	c.JSON(http.StatusOK, product)
}
