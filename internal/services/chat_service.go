package services

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/satria/obrolan-api/internal/models"
	"github.com/satria/obrolan-api/internal/utils"
)

// Hub manages WebSocket rooms per thread
type Hub struct {
	rooms        map[string]map[*Client]bool
	RegisterCh   chan *Client
	UnregisterCh chan *Client
	StopCh       chan struct{}
	mu           sync.RWMutex
}

// Client represents a single WebSocket connection
type Client struct {
	Hub      *Hub
	Conn     *websocket.Conn
	Send     chan []byte
	UserID   uuid.UUID
	Username string
	ThreadID string

	// SaveMessage persists a chat message to DB (injected, avoids direct DB call)
	SaveMessage func(msg *models.Message) error
}

func NewClient(hub *Hub, conn *websocket.Conn, userID uuid.UUID, username, threadID string, saveFn func(msg *models.Message) error) *Client {
	return &Client{
		Hub:         hub,
		Conn:        conn,
		Send:        make(chan []byte, 256),
		UserID:      userID,
		Username:    username,
		ThreadID:    threadID,
		SaveMessage: saveFn,
	}
}

// Message payload sent over WebSocket
type WSMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

type WSMessageData struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	CreatedAt string    `json:"created_at"`
}

type WSUserData struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

func NewHub() *Hub {
	return &Hub{
		rooms:        make(map[string]map[*Client]bool),
		RegisterCh:   make(chan *Client, 256),
		UnregisterCh: make(chan *Client, 256),
		StopCh:       make(chan struct{}, 1),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.RegisterCh:
			h.mu.Lock()
			if h.rooms[client.ThreadID] == nil {
				h.rooms[client.ThreadID] = make(map[*Client]bool)
			}
			h.rooms[client.ThreadID][client] = true
			h.mu.Unlock()

			// Broadcast join event
			joinData, _ := json.Marshal(WSUserData{
				UserID:   client.UserID,
				Username: client.Username,
			})
			msg, _ := json.Marshal(WSMessage{Type: "join", Data: joinData})
			h.broadcastToRoom(client.ThreadID, msg, client)

		case client := <-h.UnregisterCh:
			h.mu.Lock()
			if clients, ok := h.rooms[client.ThreadID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.rooms, client.ThreadID)
					}
				}
			}
			h.mu.Unlock()

			// Broadcast leave event
			leaveData, _ := json.Marshal(WSUserData{
				UserID:   client.UserID,
				Username: client.Username,
			})
			msg, _ := json.Marshal(WSMessage{Type: "leave", Data: leaveData})
			h.broadcastToRoom(client.ThreadID, msg, nil)

		case <-h.StopCh:
			return
		}
	}
}

func (h *Hub) broadcastToRoom(roomID string, message []byte, sender *Client) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.rooms[roomID] {
		if client == sender {
			continue
		}
		select {
		case client.Send <- message:
		default:
			h.mu.RUnlock()
			h.UnregisterCh <- client
			h.mu.RLock()
		}
	}
}

// ReadPump reads messages from WebSocket connection
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.UnregisterCh <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			continue
		}

		if wsMsg.Type != "message" {
			continue
		}

		var msgData struct {
			Content string `json:"content"`
		}
		if err := json.Unmarshal(wsMsg.Data, &msgData); err != nil {
			continue
		}

		if msgData.Content == "" {
			continue
		}

		// Save to database
		threadUUID, err := uuid.Parse(c.ThreadID)
		if err != nil {
			continue
		}

		chatMsg := &models.Message{
			ID:       uuid.New(),
			ThreadID: threadUUID,
			UserID:   c.UserID,
			Content:  msgData.Content,
		}

		if c.SaveMessage != nil {
			if err := c.SaveMessage(chatMsg); err != nil {
				log.Printf("Failed to save message: %v", err)
				continue
			}
		} else {
			log.Printf("Warning: SaveMessage not set, message not persisted")
			continue
		}

		// Broadcast to room
		broadcastData, _ := json.Marshal(WSMessageData{
			ID:        chatMsg.ID,
			UserID:    c.UserID,
			Username:  c.Username,
			Content:   msgData.Content,
			CreatedAt: chatMsg.CreatedAt.Format(utils.TimeFormat),
		})
		broadcastMsg, _ := json.Marshal(WSMessage{Type: "message", Data: broadcastData})

		c.Hub.broadcastToRoom(c.ThreadID, broadcastMsg, nil)
	}
}

// WritePump writes messages to WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
