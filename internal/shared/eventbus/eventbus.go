// Package eventbus defines the domain event bus port.
//
// Why an event bus?
//   - Modules must not call each other directly (no import cycles).
//   - Side effects (email, cache invalidation, analytics) react to domain events.
//   - In the future the in-memory bus can be swapped for NATS/Kafka without
//     touching domain or application code.
package eventbus

import "context"

// Event is a marker interface. Domain events implement this to document intent.
type Event interface {
	EventName() string
}

// Handler processes a single event type.
type Handler func(ctx context.Context, event Event) error

// EventBus is the port used by domain/application layers.
type EventBus interface {
	Publish(ctx context.Context, event Event) error
	Subscribe(eventName string, handler Handler) error
}

// Noop is a no-op event bus for tests.
type Noop struct{}

func (Noop) Publish(context.Context, Event) error { return nil }
func (Noop) Subscribe(string, Handler) error      { return nil }
