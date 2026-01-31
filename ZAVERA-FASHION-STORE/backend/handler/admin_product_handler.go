package handler

import (
	"net/http"
	"strconv"
	"zavera/dto"
	"zavera/service"

	"github.com/gin-gonic/gin"
)

type AdminProductHandler struct {
	productService service.AdminProductService
}

func NewAdminProductHandler(productService service.AdminProductService) *AdminProductHandler {
	return &AdminProductHandler{
		productService: productService,
	}
}

// GetAllProducts returns all products for admin (including inactive)
// GET /api/admin/products
func (h *AdminProductHandler) GetAllProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	category := c.Query("category")
	includeInactive := c.Query("include_inactive") == "true"

	products, total, err := h.productService.GetAllProductsAdmin(page, pageSize, category, includeInactive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "server_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"products":    products,
		"total_count": total,
		"page":        page,
		"page_size":   pageSize,
	})
}

// CreateProduct creates a new product
// POST /api/admin/products
func (h *AdminProductHandler) CreateProduct(c *gin.Context) {
	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Log the detailed error for debugging
		println("❌ JSON Binding Error:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Data produk tidak valid: " + err.Error(),
		})
		return
	}

	// Log the received request
	println("✅ Received product creation request:")
	println("   Name:", req.Name)
	println("   Category:", req.Category)
	println("   Brand:", req.Brand)
	println("   Material:", req.Material)
	println("   Price:", req.Price)
	println("   Images count:", len(req.Images))

	product, err := h.productService.CreateProduct(req)
	if err != nil {
		println("❌ Product creation failed:", err.Error())
		
		// Provide user-friendly error messages
		var errorMsg string
		if err == service.ErrDuplicateSlug {
			errorMsg = "Produk dengan nama yang sama sudah ada. Silakan gunakan nama yang berbeda."
		} else if err == service.ErrInvalidSlug {
			errorMsg = "Nama produk tidak valid. Gunakan huruf, angka, dan spasi saja."
		} else {
			errorMsg = err.Error()
		}
		
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "create_failed",
			"message": errorMsg,
		})
		return
	}

	println("✅ Product created successfully! ID:", product.ID)
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Product created successfully",
		"data":    product,
	})
}

// UpdateProduct updates an existing product
// PUT /api/admin/products/:id
func (h *AdminProductHandler) UpdateProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid product ID",
		})
		return
	}

	var req dto.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	product, err := h.productService.UpdateProduct(id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, product)
}

// UpdateStock updates product stock (restock or adjust)
// PATCH /api/admin/products/:id/stock
func (h *AdminProductHandler) UpdateStock(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid product ID",
		})
		return
	}

	var req dto.UpdateStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	product, err := h.productService.UpdateStock(id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "stock_update_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, product)
}

// DeleteProduct soft-deletes a product (sets is_active = false)
// DELETE /api/admin/products/:id
func (h *AdminProductHandler) DeleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid product ID",
		})
		return
	}

	err = h.productService.DeleteProduct(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "delete_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted"})
}

// AddProductImage adds an image to a product
// POST /api/admin/products/:id/images
func (h *AdminProductHandler) AddProductImage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid product ID",
		})
		return
	}

	var req dto.AddProductImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	image, err := h.productService.AddProductImage(id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "image_add_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, image)
}

// UploadProductImage handles multipart file upload to Cloudinary
// POST /api/admin/products/upload-image
func (h *AdminProductHandler) UploadProductImage(c *gin.Context) {
	// Get file from form
	file, fileHeader, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_file",
			Message: "No file uploaded",
		})
		return
	}
	defer file.Close()

	// Validate file
	if err := service.ValidateImageFile(fileHeader); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_file",
			Message: err.Error(),
		})
		return
	}

	// Upload to Cloudinary
	cloudinaryService, err := service.NewCloudinaryService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "upload_failed",
			Message: "Failed to initialize upload service",
		})
		return
	}

	imageURL, err := cloudinaryService.UploadImage(file, fileHeader.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "upload_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"image_url": imageURL,
		"message":   "Image uploaded successfully",
	})
}

// DeleteProductImage removes an image from a product
// DELETE /api/admin/products/:id/images/:imageId
func (h *AdminProductHandler) DeleteProductImage(c *gin.Context) {
	imageIDStr := c.Param("imageId")
	imageID, err := strconv.Atoi(imageIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid image ID",
		})
		return
	}

	err = h.productService.DeleteProductImage(imageID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "image_delete_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Image deleted"})
}
