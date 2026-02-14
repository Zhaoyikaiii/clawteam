// Package im provides in-memory message repository (for MVP)
package im

import (
	"context"
	"sync"
)

// InMemoryMessageRepository stores messages in memory
type InMemoryMessageRepository struct {
	messages map[string]*Message
	mu       sync.RWMutex
}

// NewInMemoryMessageRepository creates a new in-memory repository
func NewInMemoryMessageRepository() *InMemoryMessageRepository {
	return &InMemoryMessageRepository{
		messages: make(map[string]*Message),
	}
}

// Save persists a message
func (r *InMemoryMessageRepository) Save(ctx context.Context, msg *Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.messages[msg.ID] = msg
	return nil
}

// FindByID retrieves a message by ID
func (r *InMemoryMessageRepository) FindByID(ctx context.Context, id string) (*Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	msg, exists := r.messages[id]
	if !exists {
		return nil, ErrNotFound("message", id)
	}
	return msg, nil
}

// FindByChannel retrieves messages for a channel with pagination
func (r *InMemoryMessageRepository) FindByChannel(ctx context.Context, channelID string, limit, offset int) ([]*Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Get all messages for channel
	var channelMessages []*Message
	for _, msg := range r.messages {
		if msg.ChannelID == channelID {
			channelMessages = append(channelMessages, msg)
		}
	}

	// Sort by created time descending
	// For MVP, simple slice sort
	// In production, would use database ORDER BY

	// Apply offset and limit
	total := len(channelMessages)
	start := offset
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}

	if start >= end {
		return []*Message{}, nil
	}

	return channelMessages[start:end], nil
}

// FindUnread retrieves unread messages for a user in a channel
func (r *InMemoryMessageRepository) FindUnread(ctx context.Context, userID, channelID string) ([]*Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// For MVP, return empty (would track read receipts in production)
	return []*Message{}, nil
}

// MarkAsRead marks a message as read by a user
func (r *InMemoryMessageRepository) MarkAsRead(ctx context.Context, messageID, userID string) error {
	// For MVP, no-op (would store read receipts in production)
	return nil
}

// Delete removes a message
func (r *InMemoryMessageRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.messages[id]; !exists {
		return ErrNotFound("message", id)
	}

	delete(r.messages, id)
	return nil
}
