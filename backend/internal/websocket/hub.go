package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"onechat/internal/services"
)

type Client struct {
	ID       uint
	Hub      *Hub
	Conn     *websocket.Conn
	Send     chan []byte
	ChatRooms map[uint]bool
}

type Hub struct {
	clients       map[uint]*Client
	chatRooms     map[uint]map[*Client]bool
	register      chan *Client
	unregister    chan *Client
	broadcast     chan *BroadcastMessage
	mu            sync.RWMutex
	chatService   *services.ChatService
}

type BroadcastMessage struct {
	ChatID  uint
	Message []byte
	Exclude uint // User ID to exclude from broadcast
}

type WSMessage struct {
	Type    string          `json:"type"`
	ChatID  uint            `json:"chat_id,omitempty"`
	Payload json.RawMessage `json:"payload"`
}

func NewHub(chatService *services.ChatService) *Hub {
	return &Hub{
		clients:     make(map[uint]*Client),
		chatRooms:   make(map[uint]map[*Client]bool),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan *BroadcastMessage, 256),
		chatService: chatService,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()
			log.Printf("Client %d connected", client.ID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.Send)
				
				// Remove from all chat rooms
				for chatID := range client.ChatRooms {
					if room, exists := h.chatRooms[chatID]; exists {
						delete(room, client)
						if len(room) == 0 {
							delete(h.chatRooms, chatID)
						}
					}
				}
			}
			h.mu.Unlock()
			log.Printf("Client %d disconnected", client.ID)

		case message := <-h.broadcast:
			h.mu.RLock()
			if room, ok := h.chatRooms[message.ChatID]; ok {
				for client := range room {
					if client.ID != message.Exclude {
						select {
						case client.Send <- message.Message:
						default:
							close(client.Send)
							delete(h.clients, client.ID)
						}
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) JoinChatRoom(client *Client, chatID uint) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.chatRooms[chatID] == nil {
		h.chatRooms[chatID] = make(map[*Client]bool)
	}
	h.chatRooms[chatID][client] = true
	client.ChatRooms[chatID] = true
	
	log.Printf("Client %d joined chat room %d", client.ID, chatID)
}

func (h *Hub) LeaveChatRoom(client *Client, chatID uint) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if room, ok := h.chatRooms[chatID]; ok {
		delete(room, client)
		if len(room) == 0 {
			delete(h.chatRooms, chatID)
		}
	}
	delete(client.ChatRooms, chatID)
	
	log.Printf("Client %d left chat room %d", client.ID, chatID)
}

func (h *Hub) BroadcastToChat(chatID uint, message []byte, excludeUserID uint) {
	h.broadcast <- &BroadcastMessage{
		ChatID:  chatID,
		Message: message,
		Exclude: excludeUserID,
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		switch wsMsg.Type {
		case "join_chat":
			c.Hub.JoinChatRoom(c, wsMsg.ChatID)
		case "leave_chat":
			c.Hub.LeaveChatRoom(c, wsMsg.ChatID)
		case "typing":
			c.Hub.BroadcastToChat(wsMsg.ChatID, message, c.ID)
		case "message_delivered":
			c.Hub.BroadcastToChat(wsMsg.ChatID, message, c.ID)
		case "message_read":
			c.Hub.BroadcastToChat(wsMsg.ChatID, message, c.ID)
		}
	}
}

func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
	}()

	for message := range c.Send {
		if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("WebSocket write error: %v", err)
			return
		}
	}
}
