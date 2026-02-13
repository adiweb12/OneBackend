package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"onechat/internal/services"
)

type EventHandler struct {
	eventService *services.EventService
}

func NewEventHandler(eventService *services.EventService) *EventHandler {
	return &EventHandler{eventService: eventService}
}

type CreateEventRequest struct {
	Title           string `json:"title" binding:"required"`
	Description     string `json:"description"`
	Location        string `json:"location"`
	EventDate       string `json:"event_date" binding:"required"`
	SourceMessageID *uint  `json:"source_message_id"`
}

func (h *EventHandler) GetEvents(c *gin.Context) {
	userID := c.GetUint("user_id")

	events, err := h.eventService.GetUserEvents(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"events": events})
}

func (h *EventHandler) CreateEvent(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse event date
	eventDate, err := time.Parse(time.RFC3339, req.EventDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event date format"})
		return
	}

	event, err := h.eventService.CreateEvent(
		userID,
		req.Title,
		req.Description,
		req.Location,
		eventDate,
		req.SourceMessageID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"event": event})
}

func (h *EventHandler) UpdateEvent(c *gin.Context) {
	userID := c.GetUint("user_id")
	eventID, err := strconv.ParseUint(c.Param("eventId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Remove protected fields
	delete(updates, "id")
	delete(updates, "user_id")
	delete(updates, "created_at")

	event, err := h.eventService.UpdateEvent(uint(eventID), userID, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"event": event})
}

func (h *EventHandler) DeleteEvent(c *gin.Context) {
	userID := c.GetUint("user_id")
	eventID, err := strconv.ParseUint(c.Param("eventId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	if err := h.eventService.DeleteEvent(uint(eventID), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
