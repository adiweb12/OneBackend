package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"onechat/internal/services"
)

type AIHandler struct {
	aiService *services.AIService
}

func NewAIHandler(aiService *services.AIService) *AIHandler {
	return &AIHandler{aiService: aiService}
}

type ResearchRequest struct {
	Query string `json:"query" binding:"required"`
}

type ExtractEventRequest struct {
	MessageText string `json:"message_text" binding:"required"`
}

func (h *AIHandler) Research(c *gin.Context) {
	var req ResearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.aiService.Research(req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"response": response,
	})
}

func (h *AIHandler) ExtractEvent(c *gin.Context) {
	var req ExtractEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event, err := h.aiService.ExtractEvent(req.MessageText)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"event": event,
	})
}
