package service

import (
	"zavera/dto"
	"zavera/models"
	"zavera/repository"
)

type ProductService interface {
	GetAllProducts() ([]dto.ProductResponse, error)
	GetProductsByCategory(category string) ([]dto.ProductResponse, error)
	GetProductByID(id int) (*dto.ProductResponse, error)
}

type productService struct {
	productRepo repository.ProductRepository
	variantRepo *repository.VariantRepository
}

func NewProductService(productRepo repository.ProductRepository, variantRepo *repository.VariantRepository) ProductService {
	return &productService{
		productRepo: productRepo,
		variantRepo: variantRepo,
	}
}

func (s *productService) GetAllProducts() ([]dto.ProductResponse, error) {
	products, err := s.productRepo.FindAll()
	if err != nil {
		return nil, err
	}

	var response []dto.ProductResponse
	for _, p := range products {
		productResp := s.toProductResponse(&p)
		response = append(response, productResp)
	}

	return response, nil
}

func (s *productService) GetProductsByCategory(category string) ([]dto.ProductResponse, error) {
	products, err := s.productRepo.FindByCategory(category)
	if err != nil {
		return nil, err
	}

	var response []dto.ProductResponse
	for _, p := range products {
		productResp := s.toProductResponse(&p)
		response = append(response, productResp)
	}

	return response, nil
}

func (s *productService) GetProductByID(id int) (*dto.ProductResponse, error) {
	product, err := s.productRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	response := s.toProductResponse(product)
	return &response, nil
}

func (s *productService) toProductResponse(p *models.Product) dto.ProductResponse {
	response := dto.ProductResponse{
		ID:          p.ID,
		Name:        p.Name,
		Slug:        p.Slug,
		Description: p.Description,
		Price:       p.Price,
		Stock:       p.Stock,
		Weight:      p.Weight,
		Category:    p.Category,
		Subcategory: p.Subcategory,
		Brand:       p.Brand,
		Material:    p.Material,
	}

	// Set primary image URL and all images
	var images []string
	for _, img := range p.Images {
		images = append(images, img.ImageURL)
		if img.IsPrimary {
			response.ImageURL = img.ImageURL
		}
	}

	// If no primary image, use first one
	if response.ImageURL == "" && len(images) > 0 {
		response.ImageURL = images[0]
	}

	response.Images = images

	// Get available sizes from active variants
	variants, err := s.variantRepo.GetByProductID(p.ID)
	if err == nil && len(variants) > 0 {
		sizeMap := make(map[string]bool)
		for _, v := range variants {
			if v.IsActive && v.Size != nil && *v.Size != "" && v.StockQuantity > 0 {
				sizeMap[*v.Size] = true
			}
		}
		
		// Convert map to sorted slice
		var sizes []string
		sizeOrder := []string{"XS", "S", "M", "L", "XL", "XXL"}
		for _, size := range sizeOrder {
			if sizeMap[size] {
				sizes = append(sizes, size)
			}
		}
		response.AvailableSizes = sizes
	}

	return response
}
