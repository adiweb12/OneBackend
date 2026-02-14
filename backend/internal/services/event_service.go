package services

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"onechat/internal/models"
)

type EventService struct {
	db        *gorm.DB
	aiService *AIService
}

func NewEventService(db *gorm.DB, aiService *AIService) *EventService {
	return &EventService{
		db:        db,
		aiService: aiService,
	}
}

func (s *EventService) CreateEventFromMessage(userID, messageID uint, messageText string) (*models.Event, error) {
	// Extract event info using AI
	extraction, err := s.aiService.ExtractEvent(messageText)
	if err != nil {
		return nil, fmt.Errorf("failed to extract event: %w", err)
	}

	// Parse date and time
	eventDateTime, err := time.Parse("2006-01-02 15:04", extraction.Date+" "+extraction.Time)
	if err != nil {
		// Try with just date
		eventDateTime, err = time.Parse("2006-01-02", extraction.Date)
		if err != nil {
			return nil, fmt.Errorf("invalid date format: %w", err)
		}
	}

	// Create event
	event := &models.Event{
		UserID:          userID,
		Title:           extraction.Title,
		Description:     extraction.Description,
		EventDate:       eventDateTime,
		Location:        extraction.Location,
		SourceMessageID: &messageID,
	}

	if err := s.db.Create(event).Error; err != nil {
		return nil, err
	}

	return event, nil
}

func (s *EventService) CreateEvent(userID uint, title, description, location string, eventDate time.Time, sourceMessageID *uint) (*models.Event, error) {
	event := &models.Event{
		UserID:          userID,
		Title:           title,
		Description:     description,
		EventDate:       eventDate,
		Location:        location,
		SourceMessageID: sourceMessageID,
	}

	if err := s.db.Create(event).Error; err != nil {
		return nil, err
	}

	return event, nil
}

func (s *EventService) GetUserEvents(userID uint) ([]models.Event, error) {
	var events []models.Event
	err := s.db.Where("user_id = ?", userID).
		Order("event_date ASC").
		Find(&events).Error
	
	return events, err
}

func (s *EventService) GetUpcomingEvents(userID uint, limit int) ([]models.Event, error) {
	var events []models.Event
	err := s.db.Where("user_id = ? AND event_date > ?", userID, time.Now()).
		Order("event_date ASC").
		Limit(limit).
		Find(&events).Error
	
	return events, err
}

func (s *EventService) UpdateEvent(eventID, userID uint, updates map[string]interface{}) (*models.Event, error) {
	var event models.Event
	if err := s.db.Where("id = ? AND user_id = ?", eventID, userID).First(&event).Error; err != nil {
		return nil, err
	}

	if err := s.db.Model(&event).Updates(updates).Error; err != nil {
		return nil, err
	}

	return &event, nil
}

func (s *EventService) DeleteEvent(eventID, userID uint) error {
	return s.db.Where("id = ? AND user_id = ?", eventID, userID).Delete(&models.Event{}).Error
}

func (s *EventService) GetEventByID(eventID uint) (*models.Event, error) {
	var event models.Event
	if err := s.db.First(&event, eventID).Error; err != nil {
		return nil, err
	}
	return &event, nil
}
