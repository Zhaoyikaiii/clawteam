// Package im provides event bus for real-time updates
package im

import (
	"context"
	"sync"
	"time"
)

// EventType represents the type of event
type EventType string

const (
	EventMessageCreated  EventType = "message.created"
	EventMessageUpdated  EventType = "message.updated"
	EventMessageDeleted  EventType = "message.deleted"
	EventUserJoined     EventType = "user.joined"
	EventUserLeft       EventType = "user.left"
	EventUserTyping     EventType = "user.typing"
	EventPresenceUpdate EventType = "presence.updated"
)

// Event represents a system event
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp int64                  `json:"timestamp"`
}

// NewEvent creates a new event
func NewEvent(eventType EventType, data map[string]interface{}) *Event {
	return &Event{
		ID:        time.Now().Format("20060102150405.000000"),
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
}

// EventHandler is a function that handles events
type EventHandler func(ctx context.Context, event *Event)

// EventBus manages event subscriptions and broadcasting
type EventBus interface {
	Publish(ctx context.Context, eventType EventType, data interface{})
	Subscribe(eventType EventType) (*EventSubscription, func())
}

// EventSubscription represents a subscription to events
type EventSubscription struct {
	ch   <-chan *Event
	unsub func()
}

// InMemoryEventBus is an in-process event bus implementation
type InMemoryEventBus struct {
	subscribers map[EventType][]chan *Event
	mu          sync.RWMutex
}

// NewInMemoryEventBus creates a new in-memory event bus
func NewInMemoryEventBus() *InMemoryEventBus {
	return &InMemoryEventBus{
		subscribers: make(map[EventType][]chan *Event),
	}
}

// Publish sends an event to all subscribers
func (b *InMemoryEventBus) Publish(ctx context.Context, eventType EventType, data interface{}) {
	event := NewEvent(eventType, map[string]interface{}{
		"data": data,
	})

	b.mu.RLock()
	subs := b.subscribers[eventType]
	b.mu.RUnlock()

	for _, ch := range subs {
		select {
		case ch <- event:
		case <-ctx.Done():
			return
		default:
			// Non-blocking send - if channel is full, skip
		}
	}
}

// Subscribe registers for events of a specific type
func (b *InMemoryEventBus) Subscribe(eventType EventType) *EventSubscription {
	ch := make(chan *Event, 100) // Buffered channel

	b.mu.Lock()
	b.subscribers[eventType] = append(b.subscribers[eventType], ch)
	b.mu.Unlock()

	return &EventSubscription{
		ch:   ch,
		unsub: func() {
			b.Unsubscribe(eventType, ch)
		},
	}
}

// Unsubscribe removes a channel from subscriptions
func (b *InMemoryEventBus) Unsubscribe(eventType EventType, ch chan *Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	subs := b.subscribers[eventType]
	for i, sub := range subs {
		if sub == ch {
			b.subscribers[eventType] = append(subs[:i], subs[i+1:]...)
			close(ch)
			return
		}
	}
}
