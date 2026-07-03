package eventbus

import (
	"context"
	"fmt"
	"sync"
)

// InMemory is a synchronous, in-process event bus.
//
// It is useful for tests and for single-process monoliths. Handlers run in the
// same goroutine as the publisher; if a handler fails, Publish returns the error
// and the caller can decide whether to roll back the transaction.
type InMemory struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

// NewInMemory creates an in-memory event bus.
func NewInMemory() *InMemory {
	return &InMemory{handlers: make(map[string][]Handler)}
}

// Subscribe registers a handler for an event type.
func (b *InMemory) Subscribe(eventName string, handler Handler) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventName] = append(b.handlers[eventName], handler)
	return nil
}

// Publish sends an event to all subscribed handlers.
func (b *InMemory) Publish(ctx context.Context, event Event) error {
	b.mu.RLock()
	handlers := b.handlers[event.EventName()]
	b.mu.RUnlock()

	for _, h := range handlers {
		if err := h(ctx, event); err != nil {
			return fmt.Errorf("event handler for %s failed: %w", event.EventName(), err)
		}
	}
	return nil
}
