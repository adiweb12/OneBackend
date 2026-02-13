package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	ws "onechat/internal/websocket"
	"onechat/internal/services"
)

type WebSocketHandler struct {
	hub         *ws.Hub
	authService *services.AuthService
	upgrader    websocket.Upgrader
}

func NewWebSocketHandler(hub *ws.Hub, authService *services.AuthService) *WebSocketHandler {
	return &WebSocketHandler{
		hub:         hub,
		authService: authService,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in development
			},
		},
	}
}

func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	userID := c.GetUint("user_id")

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &ws.Client{
		ID:        userID,
		Hub:       h.hub,
		Conn:      conn,
		Send:      make(chan []byte, 256),
		ChatRooms: make(map[uint]bool),
	}

	client.Hub.register <- client

	// Start reading and writing in goroutines
	go client.WritePump()
	go client.ReadPump()
}
