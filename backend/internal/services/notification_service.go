package services

import "log"

type NotificationService struct {
	// FCM client will go here in future
}

type Notification struct {
	UserID uint
	Title  string
	Body   string
	Data   map[string]string
}

func NewNotificationService() *NotificationService {
	return &NotificationService{}
}

func (s *NotificationService) SendNotification(notification *Notification) error {
	// Placeholder for FCM implementation
	log.Printf("Notification to user %d: %s - %s", notification.UserID, notification.Title, notification.Body)
	
	// TODO: Implement Firebase Cloud Messaging
	// This will be implemented when FCM tokens are stored in the database
	
	return nil
}

func (s *NotificationService) SendBulkNotifications(notifications []*Notification) error {
	for _, notif := range notifications {
		if err := s.SendNotification(notif); err != nil {
			log.Printf("Failed to send notification: %v", err)
		}
	}
	return nil
}
