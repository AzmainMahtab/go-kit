// Package nats provides a production-grade NATS JetStream adapter for the
// shared event bus.
//
// Why NATS JetStream?
// - Single deployment covers both pub/sub event bus and durable work queues.
// - At-least-once delivery with acknowledgements.
// - Horizontal scaling: multiple API instances can publish and consume.
// - When we fan out to microservices, NATS becomes the service mesh backbone.
//
// Alternatives to consider later:
// - Redis Streams: simpler if you already run Redis, but less durable.
// - RabbitMQ: mature, complex routing, harder to operate at scale.
// - Kafka: best for very high throughput, heavier operational footprint.
package nats

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elite4print/elite4print-go/internal/shared/eventbus"
	"github.com/nats-io/nats.go"
)

// EventBus is a NATS JetStream-backed event bus.
type EventBus struct {
	conn   *nats.Conn
	js     nats.JetStreamContext
	stream string
}

// NewEventBus connects to NATS and ensures the stream exists.
func NewEventBus(url, stream string) (*EventBus, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to nats: %w", err)
	}

	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("failed to create jetstream context: %w", err)
	}

	_, err = js.AddStream(&nats.StreamConfig{
		Name:     stream,
		Subjects: []string{stream + ".*>"},
	})
	if err != nil && !isStreamExists(err) {
		return nil, fmt.Errorf("failed to create stream: %w", err)
	}

	return &EventBus{conn: nc, js: js, stream: stream}, nil
}

// Publish serializes and publishes an event.
func (b *EventBus) Publish(ctx context.Context, evt eventbus.Event) error {
	payload, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	subject := fmt.Sprintf("%s.%s", b.stream, evt.EventName())
	if _, err := b.js.Publish(subject, payload); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}
	return nil
}

// Subscribe registers a durable consumer for an event type.
func (b *EventBus) Subscribe(eventName string, handler eventbus.Handler) error {
	subject := fmt.Sprintf("%s.%s", b.stream, eventName)
	consumerName := eventName + "-consumer"

	_, err := b.js.Subscribe(subject, func(msg *nats.Msg) {
		// Decode a generic envelope so we know which concrete type to unmarshal.
		var envelope eventEnvelope
		if err := json.Unmarshal(msg.Data, &envelope); err != nil {
			_ = msg.Nak()
			return
		}

		// Handlers receive the raw Event; concrete decoding happens in the handler.
		if err := handler(context.Background(), envelope); err != nil {
			_ = msg.Nak()
			return
		}
		_ = msg.Ack()
	}, nats.Durable(consumerName), nats.ManualAck())

	if err != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", subject, err)
	}
	return nil
}

// Close closes the NATS connection.
func (b *EventBus) Close() {
	b.conn.Close()
}

type eventEnvelope struct {
	Name string          `json:"name"`
	Data json.RawMessage `json:"data"`
}

func (e eventEnvelope) EventName() string { return e.Name }

func isStreamExists(err error) bool {
	if apiErr, ok := err.(*nats.APIError); ok {
		return apiErr.ErrorCode == nats.JSErrCodeStreamNameInUse
	}
	return false
}
