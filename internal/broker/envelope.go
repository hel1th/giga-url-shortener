// Package broker defines messaging abstractions and common envelope structures.
package broker

import (
	"context"
	"time"
)

// Envelope represents an event payload wrapped with metadata for message brokers.
type Envelope struct {
	ID          string
	Type        string
	Key         string
	OccurredAt  time.Time
	TraceParent string
	Payload     []byte
}

// Producer defines an interface for publishing envelopes to a message broker.
type Producer interface {
	Publish(ctx context.Context, event Envelope)
}
