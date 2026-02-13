package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Phone       string         `gorm:"unique;not null" json:"phone"`
	Username    string         `gorm:"unique;not null" json:"username"`
	Password    string         `gorm:"not null" json:"-"`
	ProfilePic  string         `json:"profile_pic"`
	Status      string         `json:"status"`
	LastSeen    *time.Time     `json:"last_seen"`
	IsOnline    bool           `json:"is_online"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type Chat struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Type      string         `gorm:"not null" json:"type"` // private or group
	User1ID   *uint          `json:"user1_id"`
	User2ID   *uint          `json:"user2_id"`
	GroupID   *uint          `json:"group_id"`
	LastMessage *Message     `gorm:"foreignKey:LastMessageID" json:"last_message,omitempty"`
	LastMessageID *uint      `json:"-"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Message struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	ChatID    uint           `gorm:"not null;index" json:"chat_id"`
	SenderID  uint           `gorm:"not null" json:"sender_id"`
	Sender    *User          `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
	Type      string         `gorm:"not null" json:"type"` // text, image, video, audio, document
	Content   string         `json:"content"`
	MediaURL  string         `json:"media_url"`
	Status    string         `gorm:"default:'sent'" json:"status"` // sent, delivered, read
	ReplyToID *uint          `json:"reply_to_id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Group struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"not null" json:"name"`
	Icon        string         `json:"icon"`
	Description string         `json:"description"`
	CreatedByID uint           `gorm:"not null" json:"created_by_id"`
	CreatedBy   *User          `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`
	Members     []GroupMember  `gorm:"foreignKey:GroupID" json:"members,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type GroupMember struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	GroupID   uint           `gorm:"not null;index" json:"group_id"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	User      *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Role      string         `gorm:"default:'member'" json:"role"` // admin, member
	JoinedAt  time.Time      `json:"joined_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Event struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	UserID          uint           `gorm:"not null;index" json:"user_id"`
	Title           string         `gorm:"not null" json:"title"`
	Description     string         `json:"description"`
	EventDate       time.Time      `json:"event_date"`
	Location        string         `json:"location"`
	SourceMessageID *uint          `json:"source_message_id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

type Media struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      uint           `gorm:"not null;index" json:"user_id"`
	Type        string         `gorm:"not null" json:"type"` // image, video, audio, document
	URL         string         `gorm:"not null" json:"url"`
	PublicID    string         `json:"public_id"`
	Size        int64          `json:"size"`
	ExpiresAt   time.Time      `json:"expires_at"`
	CreatedAt   time.Time      `json:"created_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type MessageStatus struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	MessageID uint      `gorm:"not null;index" json:"message_id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Status    string    `gorm:"not null" json:"status"` // delivered, read
	Timestamp time.Time `json:"timestamp"`
}
