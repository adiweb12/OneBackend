package services

import (
	"context"
	"errors"
	"fmt"
	// Removed unused "io" import
	"log"
	"mime/multipart"
	"time"

	"[github.com/cloudinary/cloudinary-go/v2](https://github.com/cloudinary/cloudinary-go/v2)"
	"[github.com/cloudinary/cloudinary-go/v2/api/uploader](https://github.com/cloudinary/cloudinary-go/v2/api/uploader)"
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

	contentType := fileHeader.Header.Get("Content-Type")
	var resourceType string
	var folder string

	switch {
	case len(contentType) >= 5 && contentType[:5] == "image":
		resourceType = "image"
		folder = "onechat/images"
	case len(contentType) >= 5 && contentType[:5] == "video":
		resourceType = "video"
		folder = "onechat/videos"
	case len(contentType) >= 5 && contentType[:5] == "audio":
		resourceType = "video" 
		folder = "onechat/audio"
	default:
		resourceType = "raw"
		folder = "onechat/documents"
	}

	ctx := context.Background()
	uploadParams := uploader.UploadParams{
		Folder:       folder,
		ResourceType: resourceType,
	}

	result, err := s.cloudinary.Upload.Upload(ctx, file, uploadParams)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to Cloudinary: %w", err)
	}

	media := &models.Media{
		UserID:    userID,
		Type:      resourceType,
		URL:       result.SecureURL,
		PublicID:  result.PublicID,
		Size:      fileHeader.Size,
		ExpiresAt: time.Now().Add(10 * 24 * time.Hour),
	}

	if s.db != nil {
		s.db.Create(media)
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
	_, err := s.cloudinary.Upload.Destroy(ctx, uploader.DestroyParams{PublicID: publicID})
	if err != nil {
		return err
	}

	if s.db != nil {
		s.db.Where("public_id = ?", publicID).Delete(&models.Media{})
	}
	return nil
}

func (s *MediaService) StartCleanupScheduler(interval time.Duration) {
	if s.cloudinary == nil || s.db == nil {
		return
	}
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			var expired []models.Media
			s.db.Where("expires_at < ?", time.Now()).Find(&expired)
			for _, m := range expired {
				s.Delete(m.PublicID)
			}
		}
	}()
}

func (s *MediaService) UploadFromBytes(data []byte, filename string, userID uint) (*UploadResult, error) {
	if s.cloudinary == nil {
		return nil, errors.New("Cloudinary not configured")
	}

	ctx := context.Background()
	result, err := s.cloudinary.Upload.Upload(ctx, bytes.NewReader(data), uploader.UploadParams{
		Folder: "onechat/files",
	})
	if err != nil {
		return nil, err
	}

	return &UploadResult{
		URL: result.SecureURL,
		PublicID: result.PublicID,
		Type: "file",
	}, nil
}
