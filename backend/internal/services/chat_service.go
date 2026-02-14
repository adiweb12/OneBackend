package services

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"onechat/internal/models"
)

type ChatService struct {
	db *gorm.DB
}

func NewChatService(db *gorm.DB) *ChatService {
	return &ChatService{db: db}
}

func (s *ChatService) GetUserChats(userID uint) ([]models.Chat, error) {
	var chats []models.Chat
	err := s.db.Preload("LastMessage").
		Preload("LastMessage.Sender").
		Where("(user1_id = ? OR user2_id = ?) AND type = ?", userID, userID, "private").
		Or("id IN (?)", 
			s.db.Table("group_members").
				Select("group_id").
				Where("user_id = ?", userID)).
		Order("updated_at DESC").
		Find(&chats).Error
	
	return chats, err
}

func (s *ChatService) GetOrCreatePrivateChat(user1ID, user2ID uint) (*models.Chat, error) {
	var chat models.Chat
	err := s.db.Where(
		"((user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)) AND type = ?",
		user1ID, user2ID, user2ID, user1ID, "private",
	).First(&chat).Error

	if err == gorm.ErrRecordNotFound {
		chat = models.Chat{
			Type:    "private",
			User1ID: &user1ID,
			User2ID: &user2ID,
		}
		if err := s.db.Create(&chat).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return &chat, nil
}

func (s *ChatService) GetMessages(chatID uint, limit, offset int) ([]models.Message, error) {
	var messages []models.Message
	err := s.db.Preload("Sender").
		Where("chat_id = ?", chatID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error
	
	// Reverse to show oldest first
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	
	return messages, err
}

func (s *ChatService) CreateMessage(chatID, senderID uint, msgType, content, mediaURL string, replyToID *uint) (*models.Message, error) {
	message := &models.Message{
		ChatID:    chatID,
		SenderID:  senderID,
		Type:      msgType,
		Content:   content,
		MediaURL:  mediaURL,
		Status:    "sent",
		ReplyToID: replyToID,
	}

	if err := s.db.Create(message).Error; err != nil {
		return nil, err
	}

	// Update chat's last message
	s.db.Model(&models.Chat{}).Where("id = ?", chatID).Updates(map[string]interface{}{
		"last_message_id": message.ID,
		"updated_at":      time.Now(),
	})

	// Preload sender info
	s.db.Preload("Sender").First(message, message.ID)

	return message, nil
}

func (s *ChatService) UpdateMessageStatus(messageID, userID uint, status string) error {
	// Update message status
	if err := s.db.Model(&models.Message{}).
		Where("id = ? AND sender_id != ?", messageID, userID).
		Update("status", status).Error; err != nil {
		return err
	}

	// Create or update message status record
	messageStatus := &models.MessageStatus{
		MessageID: messageID,
		UserID:    userID,
		Status:    status,
		Timestamp: time.Now(),
	}

	return s.db.Create(messageStatus).Error
}

func (s *ChatService) DeleteMessage(messageID, userID uint) error {
	var message models.Message
	if err := s.db.First(&message, messageID).Error; err != nil {
		return err
	}

	if message.SenderID != userID {
		return errors.New("unauthorized to delete this message")
	}

	return s.db.Delete(&message).Error
}

func (s *ChatService) GetChatByID(chatID uint) (*models.Chat, error) {
	var chat models.Chat
	if err := s.db.Preload("LastMessage").First(&chat, chatID).Error; err != nil {
		return nil, err
	}
	return &chat, nil
}

func (s *ChatService) GetMessageByID(messageID uint) (*models.Message, error) {
	var message models.Message
	if err := s.db.Preload("Sender").First(&message, messageID).Error; err != nil {
		return nil, err
	}
	return &message, nil
}
