package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryService interface {
	UploadImage(file multipart.File, filename string) (string, error)
	DeleteImage(publicID string) error
}

type cloudinaryService struct {
	cld *cloudinary.Cloudinary
}

func NewCloudinaryService() (CloudinaryService, error) {
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")

	if cloudName == "" || apiKey == "" || apiSecret == "" {
		return nil, fmt.Errorf("cloudinary credentials not configured")
	}

	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cloudinary: %w", err)
	}

	return &cloudinaryService{cld: cld}, nil
}

// UploadImage uploads an image to Cloudinary and returns the URL
func (s *cloudinaryService) UploadImage(file multipart.File, filename string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Generate unique public ID
	ext := filepath.Ext(filename)
	publicID := fmt.Sprintf("zavera/products/%d_%s", time.Now().Unix(), strings.TrimSuffix(filename, ext))

	// Upload to Cloudinary
	uploadResult, err := s.cld.Upload.Upload(ctx, file, uploader.UploadParams{
		PublicID:       publicID,
		Folder:         "zavera/products",
		ResourceType:   "image",
		Transformation: "q_auto:good,f_auto",
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload to cloudinary: %w", err)
	}

	return uploadResult.SecureURL, nil
}

// DeleteImage deletes an image from Cloudinary
func (s *cloudinaryService) DeleteImage(publicID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := s.cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})

	if err != nil {
		return fmt.Errorf("failed to delete from cloudinary: %w", err)
	}

	return nil
}

// ExtractPublicIDFromURL extracts Cloudinary public ID from URL
func ExtractPublicIDFromURL(url string) string {
	// Example URL: https://res.cloudinary.com/dmofyz5tv/image/upload/v1234567890/zavera/products/image.jpg
	// Extract: zavera/products/image
	
	parts := strings.Split(url, "/upload/")
	if len(parts) < 2 {
		return ""
	}

	// Remove version and get path
	pathParts := strings.Split(parts[1], "/")
	if len(pathParts) < 2 {
		return ""
	}

	// Skip version (v1234567890) and join the rest
	publicIDParts := pathParts[1:]
	publicID := strings.Join(publicIDParts, "/")
	
	// Remove file extension
	publicID = strings.TrimSuffix(publicID, filepath.Ext(publicID))
	
	return publicID
}

// ValidateImageFile validates uploaded image file
func ValidateImageFile(fileHeader *multipart.FileHeader) error {
	// Check file size (max 5MB)
	maxSize := int64(5 * 1024 * 1024) // 5MB
	if fileHeader.Size > maxSize {
		return fmt.Errorf("file size exceeds 5MB limit")
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
	}

	if !allowedExts[ext] {
		return fmt.Errorf("invalid file type. Allowed: JPG, PNG, WEBP")
	}

	// Check MIME type
	file, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read first 512 bytes to detect content type
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Reset file pointer
	file.Seek(0, 0)

	// Validate MIME type
	contentType := fileHeader.Header.Get("Content-Type")
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/webp": true,
	}

	if !allowedTypes[contentType] {
		return fmt.Errorf("invalid content type: %s", contentType)
	}

	return nil
}
