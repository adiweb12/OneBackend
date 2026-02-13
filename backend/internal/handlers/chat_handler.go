package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"onechat/internal/services"
	"onechat/internal/websocket"
)

type ChatHandler struct {
	chatService *services.ChatService
	hub         *websocket.Hub
}

func NewChatHandler(chatService *services.ChatService, hub *websocket.Hub) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
		hub:         hub,
	}
}

type CreateChatRequest struct {
	RecipientID uint `json:"recipient_id" binding:"required"`
}

type SendMessageRequest struct {
	Type      string `json:"type" binding:"required"`
	Content   string `json:"content"`
	MediaURL  string `json:"media_url"`
	ReplyToID *uint  `json:"reply_to_id"`
}

type UpdateMessageStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

func (h *ChatHandler) GetChats(c *gin.Context) {
	userID := c.GetUint("user_id")

	chats, err := h.chatService.GetUserChats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"chats": chats})
}

func (h *ChatHandler) CreateChat(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req CreateChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	chat, err := h.chatService.GetOrCreatePrivateChat(userID, req.RecipientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"chat": chat})
}

func (h *ChatHandler) GetMessages(c *gin.Context) {
	chatID, err := strconv.ParseUint(c.Param("chatId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat ID"})
		return
	}

	limit := 50
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsedLimit, err := strconv.Atoi(l); err == nil {
			limit = parsedLimit
		}
	}

	if o := c.Query("offset"); o != "" {
		if parsedOffset, err := strconv.Atoi(o); err == nil {
			offset = parsedOffset
		}
	}

	messages, err := h.chatService.GetMessages(uint(chatID), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

func (h *ChatHandler) SendMessage(c *gin.Context) {
	userID := c.GetUint("user_id")
	chatID, err := strconv.ParseUint(c.Param("chatId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat ID"})
		return
	}

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	message, err := h.chatService.CreateMessage(
		uint(chatID),
		userID,
		req.Type,
		req.Content,
		req.MediaURL,
		req.ReplyToID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast to WebSocket
	messageJSON, _ := json.Marshal(map[string]interface{}{
		"type":    "new_message",
		"message": message,
	})
	h.hub.BroadcastToChat(uint(chatID), messageJSON, userID)

	c.JSON(http.StatusCreated, gin.H{"message": message})
}

func (h *ChatHandler) UpdateMessageStatus(c *gin.Context) {
	userID := c.GetUint("user_id")
	messageID, err := strconv.ParseUint(c.Param("messageId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	var req UpdateMessageStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.chatService.UpdateMessageStatus(uint(messageID), userID, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get message to broadcast update
	message, _ := h.chatService.GetMessageByID(uint(messageID))
	if message != nil {
		statusUpdate, _ := json.Marshal(map[string]interface{}{
			"type":       "message_status",
			"message_id": messageID,
			"status":     req.Status,
			"user_id":    userID,
		})
		h.hub.BroadcastToChat(message.ChatID, statusUpdate, 0)
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *ChatHandler) DeleteMessage(c *gin.Context) {
	userID := c.GetUint("user_id")
	messageID, err := strconv.ParseUint(c.Param("messageId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	// Get message before deleting to get chat ID
	message, _ := h.chatService.GetMessageByID(uint(messageID))
	
	if err := h.chatService.DeleteMessage(uint(messageID), userID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// Broadcast deletion
	if message != nil {
		deleteNotif, _ := json.Marshal(map[string]interface{}{
			"type":       "message_deleted",
			"message_id": messageID,
		})
		h.hub.BroadcastToChat(message.ChatID, deleteNotif, 0)
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
