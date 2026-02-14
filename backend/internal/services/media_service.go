package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"gorm.io/gorm"
	"onechat/internal/models"
)

type MediaService struct {
	db            *gorm.DB
	cloudinary    *cloudinary.Cloudinary
	cloudinaryURL string
}

type UploadResult struct {
	URL      string `json:"url"`
	PublicID string `json:"public_id"`
	Type     string `json:"type"`
}

func NewMediaService(cloudinaryURL string) *MediaService {
	var cld *cloudinary.Cloudinary
	var err error

	if cloudinaryURL != "" {
		cld, err = cloudinary.NewFromURL(cloudinaryURL)
		if err != nil {
			log.Printf("Failed to initialize Cloudinary: %v", err)
		}
	}

	return &MediaService{
		cloudinary:    cld,
		cloudinaryURL: cloudinaryURL,
	}
}

func (s *MediaService) SetDB(db *gorm.DB) {
	s.db = db
}

func (s *MediaService) Upload(file multipart.File, fileHeader *multipart.FileHeader, userID uint) (*UploadResult, error) {
	if s.cloudinary == nil {
		return nil, errors.New("Cloudinary not configured")
	}

	// Determine file type
	contentType := fileHeader.Header.Get("Content-Type")
	var resourceType string
	var folder string

	switch {
	case contentType[:5] == "image":
		resourceType = "image"
		folder = "onechat/images"
	case contentType[:5] == "video":
		resourceType = "video"
		folder = "onechat/videos"
	case contentType[:5] == "audio":
		resourceType = "video" // Cloudinary uses video for audio
		folder = "onechat/audio"
	default:
		resourceType = "raw"
		folder = "onechat/documents"
	}

	// Upload to Cloudinary
	ctx := context.Background()
	uploadParams := uploader.UploadParams{
		Folder:       folder,
		ResourceType: resourceType,
		// Auto-delete after 10 days (864000 seconds)
		// Note: This requires a Cloudinary paid plan for scheduled deletion
		// For free tier, use the cleanup scheduler
	}

	result, err := s.cloudinary.Upload.Upload(ctx, file, uploadParams)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to Cloudinary: %w", err)
	}

	// Save to database
	media := &models.Media{
		UserID:    userID,
		Type:      resourceType,
		URL:       result.SecureURL,
		PublicID:  result.PublicID,
		Size:      fileHeader.Size,
		ExpiresAt: time.Now().Add(10 * 24 * time.Hour), // 10 days
	}

	if s.db != nil {
		if err := s.db.Create(media).Error; err != nil {
			log.Printf("Failed to save media to database: %v", err)
		}
	}

	return &UploadResult{
		URL:      result.SecureURL,
		PublicID: result.PublicID,
		Type:     resourceType,
	}, nil
}

func (s *MediaService) Delete(publicID string) error {
	if s.cloudinary == nil {
		return errors.New("Cloudinary not configured")
	}

	ctx := context.Background()
	_, err := s.cloudinary.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})

	if err != nil {
		return fmt.Errorf("failed to delete from Cloudinary: %w", err)
	}

	// Delete from database
	if s.db != nil {
		s.db.Where("public_id = ?", publicID).Delete(&models.Media{})
	}

	return nil
}

func (s *MediaService) StartCleanupScheduler(interval time.Duration) {
	if s.cloudinary == nil || s.db == nil {
		log.Println("Cloudinary or DB not configured, skipping cleanup scheduler")
		return
	}

	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			s.cleanupExpiredMedia()
		}
	}()

	log.Println("Media cleanup scheduler started")
}

func (s *MediaService) cleanupExpiredMedia() {
	var expiredMedia []models.Media
	if err := s.db.Where("expires_at < ?", time.Now()).Find(&expiredMedia).Error; err != nil {
		log.Printf("Error finding expired media: %v", err)
		return
	}

	log.Printf("Found %d expired media files to delete", len(expiredMedia))

	for _, media := range expiredMedia {
		if err := s.Delete(media.PublicID); err != nil {
			log.Printf("Error deleting media %s: %v", media.PublicID, err)
		} else {
			log.Printf("Deleted expired media: %s", media.PublicID)
		}
	}
}

func (s *MediaService) UploadFromBytes(data []byte, filename string, userID uint) (*UploadResult, error) {
	if s.cloudinary == nil {
		return nil, errors.New("Cloudinary not configured")
	}

	ctx := context.Background()
	uploadParams := uploader.UploadParams{
		Folder:   "onechat/files",
		PublicID: filename,
	}

	result, err := s.cloudinary.Upload.Upload(ctx, data, uploadParams)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to Cloudinary: %w", err)
	}

	return &UploadResult{
		URL:      result.SecureURL,
		PublicID: result.PublicID,
		Type:     "file",
	}, nil
}
