// Package im provides core messaging functionality
package im

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// MessageType represents the type of message content
type MessageType string

const (
	MessageTypeText        MessageType = "text"
	MessageTypeAgentCard   MessageType = "agent_card"
	MessageTypeApprovalCard MessageType = "approval_card"
)

// Message represents a chat message
type Message struct {
	ID          string                 `json:"id"`
	ChannelID   string                 `json:"channel_id"`
	SenderID    string                 `json:"sender_id"`
	ParentID    *string                `json:"parent_id,omitempty"`
	Content     string                 `json:"content"`
	ContentType  MessageType            `json:"content_type"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// NewMessage creates a new message with required fields
func NewMessage(channelID, senderID, content string) *Message {
	now := time.Now()
	return &Message{
		ID:          uuid.New().String(),
		ChannelID:   channelID,
		SenderID:    senderID,
		Content:     content,
		ContentType:  MessageTypeText,
		Metadata:    make(map[string]interface{}),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Validate checks if a message is valid
func (m *Message) Validate() error {
	if m.ID == "" {
		return ErrInvalidInput("message ID cannot be empty")
	}
	if m.ChannelID == "" {
		return ErrInvalidInput("channel ID cannot be empty")
	}
	if m.SenderID == "" {
		return ErrInvalidInput("sender ID cannot be empty")
	}
	if m.Content == "" {
		return ErrInvalidInput("content cannot be empty")
	}
	return nil
}

// ToDTO converts Message to a map for API responses
func (m *Message) ToDTO() map[string]interface{} {
	return map[string]interface{}{
		"id":           m.ID,
		"channel_id":   m.ChannelID,
		"sender_id":    m.SenderID,
		"parent_id":    m.ParentID,
		"content":      m.Content,
		"content_type": m.ContentType,
		"metadata":     m.Metadata,
		"created_at":   m.CreatedAt,
		"updated_at":   m.UpdatedAt,
	}
}

// MessageRepository defines the interface for message persistence
type MessageRepository interface {
	Save(ctx context.Context, msg *Message) error
	FindByID(ctx context.Context, id string) (*Message, error)
	FindByChannel(ctx context.Context, channelID string, limit, offset int) ([]*Message, error)
	FindUnread(ctx context.Context, userID string, channelID string) ([]*Message, error)
	MarkAsRead(ctx context.Context, messageID, userID string) error
	Delete(ctx context.Context, id string) error
}

// MessageService handles message business logic
type MessageService struct {
	repo  MessageRepository
	bus   *InMemoryEventBus
}

// NewMessageService creates a new message service
func NewMessageService(repo MessageRepository, bus *InMemoryEventBus) *MessageService {
	return &MessageService{
		repo: repo,
		bus:  bus,
	}
}

// SendMessage creates and persists a new message
func (s *MessageService) SendMessage(ctx context.Context, channelID, senderID, content string, parentID *string) (*Message, error) {
	msg := NewMessage(channelID, senderID, content)
	if parentID != nil {
		msg.ParentID = parentID
	}

	if err := msg.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, msg); err != nil {
		return nil, ErrStorageFailed("failed to save message")
	}

	// Publish message event
	s.bus.Publish(ctx, EventMessageCreated, msg)

	return msg, nil
}

// GetMessages retrieves messages for a channel
func (s *MessageService) GetMessages(ctx context.Context, channelID string, limit, offset int) ([]*Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.FindByChannel(ctx, channelID, limit, offset)
}

// GetMessageByID retrieves a single message
func (s *MessageService) GetMessageByID(ctx context.Context, id string) (*Message, error) {
	return s.repo.FindByID(ctx, id)
}
