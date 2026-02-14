// Package im provides HTTP API handlers
package im

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// API handles HTTP requests for the IM service
type API struct {
	router    *mux.Router
	messages  *MessageService
	presence  *PresenceManager
	events     *InMemoryEventBus
	wshub      *WSHub
	logger     *slog.Logger
}

// NewAPI creates a new API handler
func NewAPI(msgService *MessageService, presence *PresenceManager, events *InMemoryEventBus, logger *slog.Logger) *API {
	api := &API{
		messages: msgService,
		presence: presence,
		events:   events,
		wshub:    NewWSHub(logger),
		logger:    logger,
	}

	api.setupRoutes()
	return api
}

// setupRoutes configures all HTTP routes
func (a *API) setupRoutes() {
	r := mux.NewRouter()

	// Health check
	r.HandleFunc("/health", a.handleHealth).Methods("GET")

	// WebSocket endpoint
	r.HandleFunc("/ws", a.handleWebSocket)

	// Message API
	r.HandleFunc("/api/v1/channels/{channel}/messages", a.handleGetMessages).Methods("GET")
	r.HandleFunc("/api/v1/channels/{channel}/messages", a.handlePostMessage).Methods("POST")
	r.HandleFunc("/api/v1/messages/{id}", a.handleGetMessage).Methods("GET")

	// Presence API
	r.HandleFunc("/api/v1/presence", a.handleGetPresence).Methods("GET")
	r.HandleFunc("/api/v1/presence/subscribe", a.handleSubscribePresence).Methods("POST")

	a.router = r
}

// handleHealth returns service health status
func (a *API) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "healthy",
		"service": "im-core",
	})
}

// handleWebSocket upgrades HTTP to WebSocket
func (a *API) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Get user from context (would be set by auth middleware)
	userID := r.URL.Query().Get("user")
	if userID == "" {
		http.Error(w, "user parameter required", http.StatusBadRequest)
		return
	}

	// Upgrade to WebSocket
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for MVP
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		a.logger.Error("Failed to upgrade to WebSocket", "error", err)
		http.Error(w, "Failed to upgrade to WebSocket", http.StatusBadRequest)
		return
	}

	// Create client and start pumps
	client := NewWSClient(conn, userID, a.presence, a.logger)
	a.wshub.Register(client)

	// Handle connection close
	defer a.wshub.Unregister(client)
	defer a.presence.UserDisconnected(userID, client.ID)

	// Start read/write pumps
	ctx := r.Context()
	go client.ReadPump(ctx)
	go client.WritePump(ctx)
}

// handleGetMessages retrieves messages for a channel
func (a *API) handleGetMessages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	channelID := vars["channel"]

	// Parse pagination params
	limit := 50
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if i, err := strconv.Atoi(l); err == nil {
			limit = i
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if i, err := strconv.Atoi(o); err == nil {
			offset = i
		}
	}

	messages, err := a.messages.GetMessages(r.Context(), channelID, limit, offset)
	if err != nil {
		a.sendError(w, err)
		return
	}

	// Mark as read for user (would get from auth context)
	userID := r.URL.Query().Get("user")

	// Get unread count
	unread := 0
	if userID != "" {
		a.presence.UpdateLastSeen(userID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"messages": messages,
		"limit":    limit,
		"offset":   offset,
		"unread":   unread,
	})
}

// handlePostMessage creates a new message
func (a *API) handlePostMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	channelID := vars["channel"]

	// Get user from context (would be set by auth middleware)
	userID := r.URL.Query().Get("user")
	if userID == "" {
		a.sendError(w, ErrUnauthorized)
		return
	}

	var req struct {
		Content  string `json:"content"`
		ParentID  *string `json:"parent_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	message, err := a.messages.SendMessage(r.Context(), channelID, userID, req.Content, req.ParentID)
	if err != nil {
		a.sendError(w, err)
		return
	}

	// Broadcast to WebSocket clients
	event := NewEvent(EventMessageCreated, map[string]interface{}{
		"message": message.ToDTO(),
	})
	a.wshub.BroadcastToChannel(channelID, event)

	// Also publish to event bus for other services
	a.events.Publish(r.Context(), EventMessageCreated, message)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(message.ToDTO())
}

// handleGetMessage retrieves a single message
func (a *API) handleGetMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	messageID := vars["id"]

	message, err := a.messages.GetMessageByID(r.Context(), messageID)
	if err != nil {
		a.sendError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(message.ToDTO())
}

// handleGetPresence returns online presence for users
func (a *API) handleGetPresence(w http.ResponseWriter, r *http.Request) {
	userIDs := strings.Split(r.URL.Query().Get("users"), ",")

	presence := make(map[string]PresenceStatus)
	for _, userID := range userIDs {
		if userID != "" {
			presence[userID] = a.presence.GetStatus(userID)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(presence)
}

// handleSubscribePresence subscribes to presence updates (SSE for MVP)
func (a *API) handleSubscribePresence(w http.ResponseWriter, r *http.Request) {
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Get user to track
	userID := r.URL.Query().Get("user")
	if userID == "" {
		http.Error(w, "user parameter required", http.StatusBadRequest)
		return
	}

	// Send initial presence
	presence := a.presence.GetStatus(userID)
	fmt.Fprintf(w, "event: presence\ndata: %s\n\n", presence)
	flusher.Flush()

	// In production, would listen to event bus and send updates
	<-r.Context().Done()
}

// sendError writes an error response
func (a *API) sendError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	errMsg := "internal error"

	if e, ok := err.(*Error); ok {
		switch e.Type {
		case ErrorTypeValidation:
			status = http.StatusBadRequest
		case ErrorTypeNotFound:
			status = http.StatusNotFound
		case ErrorTypeAuth:
			status = http.StatusUnauthorized
		}
		errMsg = e.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": errMsg,
		"code":   fmt.Sprintf("IM_%d", status),
	})
}

// Serve starts the HTTP server
func (a *API) Serve(addr string) error {
	a.logger.Info("Starting IM Core API", "addr", addr)
	return http.ListenAndServe(addr, a.router)
}
