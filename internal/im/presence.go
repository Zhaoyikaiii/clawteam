// Package im provides user presence management
package im

import (
	"sync"
	"time"
)

// PresenceStatus represents user online status
type PresenceStatus string

const (
	PresenceStatusOnline  PresenceStatus = "online"
	PresenceStatusAway    PresenceStatus = "away"
	PresenceStatusOffline PresenceStatus = "offline"
)

// Presence represents a user's online state
type Presence struct {
	UserID    string         `json:"user_id"`
	Status     PresenceStatus `json:"status"`
	LastSeenAt time.Time     `json:"last_seen_at"`
	Clients    map[string]bool `json:"-"` // session IDs
	mu          sync.RWMutex  `json:"-"`
}

// NewPresence creates a new presence entry
func NewPresence(userID string) *Presence {
	return &Presence{
		UserID:    userID,
		Status:     PresenceStatusOnline,
		LastSeenAt: time.Now(),
		Clients:    make(map[string]bool),
	}
}

// IsOnline returns true if user has any active clients
func (p *Presence) IsOnline() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.Clients) > 0
}

// AddClient adds a session/client for this user
func (p *Presence) AddClient(sessionID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Clients[sessionID] = true
	p.Status = PresenceStatusOnline
	p.LastSeenAt = time.Now()
}

// RemoveClient removes a session/client for this user
func (p *Presence) RemoveClient(sessionID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.Clients, sessionID)
	if len(p.Clients) == 0 {
		p.Status = PresenceStatusOffline
	}
	p.LastSeenAt = time.Now()
}

// GetClientCount returns the number of active clients
func (p *Presence) GetClientCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.Clients)
}

// PresenceManager manages user presence across the system
type PresenceManager struct {
	users map[string]*Presence
	mu    sync.RWMutex
}

// NewPresenceManager creates a new presence manager
func NewPresenceManager() *PresenceManager {
	return &PresenceManager{
		users: make(map[string]*Presence),
	}
}

// GetOrCreate gets or creates a presence entry for a user
func (m *PresenceManager) GetOrCreate(userID string) *Presence {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.users[userID]; !exists {
		m.users[userID] = NewPresence(userID)
	}
	return m.users[userID]
}

// UserConnected registers a user connection
func (m *PresenceManager) UserConnected(userID, sessionID string) {
	p := m.GetOrCreate(userID)
	p.AddClient(sessionID)
}

// UserDisconnected registers a user disconnection
func (m *PresenceManager) UserDisconnected(userID, sessionID string) {
	m.mu.RLock()
	p, exists := m.users[userID]
	if !exists {
		m.mu.RUnlock()
		return
	}
	m.mu.RUnlock()

	p.RemoveClient(sessionID)

	// If no more clients, mark offline
	if !p.IsOnline() {
		m.mu.Lock()
		// Optionally remove from map after a timeout
		// For now, keep the record
		m.mu.Unlock()
	}
}

// GetStatus returns the current status of a user
func (m *PresenceManager) GetStatus(userID string) PresenceStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if p, exists := m.users[userID]; exists {
		if p.IsOnline() {
			return PresenceStatusOnline
		}
		return p.Status
	}
	return PresenceStatusOffline
}

// GetOnlineUsers returns IDs of all currently online users
func (m *PresenceManager) GetOnlineUsers() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	online := make([]string, 0)
	for userID, p := range m.users {
		if p.IsOnline() {
			online = append(online, userID)
		}
	}
	return online
}

// GetOnlineInChannel returns online user IDs for a channel
// This requires channel membership tracking (simplified for MVP)
func (m *PresenceManager) GetOnlineInChannel(channelID string) []string {
	// MVP: Return all online users
	// Full implementation would track channel membership
	return m.GetOnlineUsers()
}

// UpdateLastSeen updates the last seen time for a user
func (m *PresenceManager) UpdateLastSeen(userID string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if p, exists := m.users[userID]; exists {
		p.LastSeenAt = time.Now()
	}
}
