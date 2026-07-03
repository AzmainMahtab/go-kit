package nats

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
)

// Job is a unit of background work.
type Job struct {
	Queue   string          `json:"queue"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// WorkerEnqueue publishes a job to a JetStream work queue.
func (b *EventBus) WorkerEnqueue(ctx context.Context, queue, jobType string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal job payload: %w", err)
	}

	job := Job{Queue: queue, Type: jobType, Payload: data}
	body, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	subject := fmt.Sprintf("%s.queue.%s", b.stream, queue)
	if _, err := b.js.Publish(subject, body); err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}
	return nil
}

// WorkerRegister creates a pull consumer for a queue and invokes handler for each job.
func (b *EventBus) WorkerRegister(queue string, handler func(ctx context.Context, job Job) error) error {
	streamName := b.stream + "-queue-" + queue
	subject := fmt.Sprintf("%s.queue.%s", b.stream, queue)

	_, err := b.js.AddStream(&nats.StreamConfig{
		Name:      streamName,
		Subjects:  []string{subject},
		Retention: nats.WorkQueuePolicy,
	})
	if err != nil && !isStreamExists(err) {
		return fmt.Errorf("failed to create queue stream: %w", err)
	}

	// Subscription loop; in production run this in a dedicated worker process.
	go func() {
		sub, err := b.js.PullSubscribe(subject, queue+"-worker")
		if err != nil {
			return
		}

		for {
			msgs, err := sub.Fetch(1)
			if err != nil {
				continue
			}
			for _, msg := range msgs {
				var job Job
				if err := json.Unmarshal(msg.Data, &job); err != nil {
					_ = msg.Nak()
					continue
				}
				if err := handler(context.Background(), job); err != nil {
					_ = msg.Nak()
					continue
				}
				_ = msg.Ack()
			}
		}
	}()

	return nil
}
