// Package im provides WebSocket connection handling
package im

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to peer
	writeWait = 10 * time.Second

	// Time allowed to read next pong message from peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period
	pingPeriod = 30 * time.Second

	// Maximum message size allowed from peer
	maxMessageSize = 8192
)

// MessageType from client
type WSMessageType string

const (
	WSMessageTypeSubscribe   WSMessageType = "subscribe"
	WSMessageTypeUnsubscribe WSMessageType = "unsubscribe"
	WSMessageTypeTyping     WSMessageType = "typing"
	WSMessageTypeSend        WSMessageType = "send"
)

// WSMessage represents a WebSocket message from client
type WSMessage struct {
	Type    WSMessageType `json:"type"`
	Channel  string          `json:"channel,omitempty"`
	Data     json.RawMessage `json:"data,omitempty"`
}

// WSClient represents a connected WebSocket client
type WSClient struct {
	ID        string
	UserID    string
	SessionID string
	Conn      *websocket.Conn
	Send      chan *Event
	Presence  *PresenceManager
	logger    *slog.Logger
	mu        sync.RWMutex
	Subscribed map[string]bool // channel IDs
}

// NewWSClient creates a new WebSocket client
func NewWSClient(conn *websocket.Conn, userID string, presence *PresenceManager, logger *slog.Logger) *WSClient {
	sessionID := fmt.Sprintf("%s-%d", userID, time.Now().UnixNano())

	return &WSClient{
		ID:        sessionID,
		UserID:    userID,
		SessionID: sessionID,
		Conn:      conn,
		Send:      make(chan *Event, 256),
		Presence:  presence,
		logger:    logger,
		Subscribed: make(map[string]bool),
	}
}

// ReadPump reads messages from the WebSocket connection
func (c *WSClient) ReadPump(ctx context.Context) {
	defer c.Conn.Close()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, message, err := c.Conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err) || websocket.IsCloseError(err) {
					c.logger.Error("WebSocket close error", "error", err)
				}
				return
			}

			var wsMsg WSMessage
			if err := json.Unmarshal(message, &wsMsg); err != nil {
				c.logger.Error("Failed to unmarshal WS message", "error", err)
				continue
			}

			c.handleMessage(&wsMsg)
		}
	}
}

// WritePump writes messages to the WebSocket connection
func (c *WSClient) WritePump(ctx context.Context) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	defer c.Conn.Close()

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteJSON(event); err != nil {
				c.logger.Error("WebSocket write error", "error", err)
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

// handleMessage processes incoming WebSocket messages
func (c *WSClient) handleMessage(msg *WSMessage) {
	switch msg.Type {
	case WSMessageTypeSubscribe:
		if msg.Channel == "" {
			c.logger.Warn("Subscribe missing channel")
			return
		}
		c.mu.Lock()
		c.Subscribed[msg.Channel] = true
		c.mu.Unlock()
		c.logger.Info("Client subscribed to channel", "channel", msg.Channel, "user", c.UserID)

	case WSMessageTypeUnsubscribe:
		if msg.Channel == "" {
			return
		}
		c.mu.Lock()
		delete(c.Subscribed, msg.Channel)
		c.mu.Unlock()

	case WSMessageTypeTyping:
		// Broadcast typing indicator to channel
		// TODO: Implement via event bus

	case WSMessageTypeSend:
		// Direct message send (should go through API)
		c.logger.Warn("Received message via WS, use API instead")
	}
}

// IsSubscribed checks if client is subscribed to a channel
func (c *WSClient) IsSubscribed(channelID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Subscribed[channelID]
}

// Close closes the connection
func (c *WSClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-c.Send:
		// Channel already closed
	default:
		close(c.Send)
	}
	return c.Conn.Close()
}

// WSHub manages active WebSocket connections
type WSHub struct {
	clients map[string]*WSClient // sessionID -> client
	mu      sync.RWMutex
	logger  *slog.Logger
}

// NewWSHub creates a new WebSocket hub
func NewWSHub(logger *slog.Logger) *WSHub {
	return &WSHub{
		clients: make(map[string]*WSClient),
		logger:  logger,
	}
}

// Register adds a new client
func (h *WSHub) Register(client *WSClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[client.ID] = client
	h.logger.Info("WebSocket client registered", "user", client.UserID, "session", client.ID)
}

// Unregister removes a client
func (h *WSHub) Unregister(client *WSClient) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.clients[client.ID]; exists {
		delete(h.clients, client.ID)
		h.logger.Info("WebSocket client unregistered", "user", client.UserID, "session", client.ID)
	}
}

// GetClient returns a client by session ID
func (h *WSHub) GetClient(sessionID string) *WSClient {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.clients[sessionID]
}

// GetClientsByUser returns all clients for a user
func (h *WSHub) GetClientsByUser(userID string) []*WSClient {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients := make([]*WSClient, 0)
	for _, client := range h.clients {
		if client.UserID == userID {
			clients = append(clients, client)
		}
	}
	return clients
}

// BroadcastToChannel sends an event to all clients subscribed to a channel
func (h *WSHub) BroadcastToChannel(channelID string, event *Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		if client.IsSubscribed(channelID) {
			select {
			case client.Send <- event:
			default:
				// Channel full, skip this client
			}
		}
	}
}

// BroadcastToUser sends an event to all clients for a user
func (h *WSHub) BroadcastToUser(userID string, event *Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		if client.UserID == userID {
			select {
			case client.Send <- event:
			default:
				// Channel full, skip this client
			}
		}
	}
}

// Broadcast sends an event to all connected clients
func (h *WSHub) Broadcast(event *Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		select {
		case client.Send <- event:
		default:
			// Channel full, skip this client
		}
	}
}

// GetClientCount returns number of connected clients
func (h *WSHub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
