package broker

import (
	"context"
	"time"
)

type Envelope struct {
	ID          string
	Type        string
	Key         string
	OccurredAt  time.Time
	TraceParent string
	Payload     []byte
}

type Producer interface {
	Publish(ctx context.Context, event Envelope) error
}
