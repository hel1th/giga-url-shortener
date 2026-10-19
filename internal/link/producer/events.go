// Package producer provides event publishing adapters for the link service.
package producer

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hel1th/giga-url-shortener/internal/broker"
	"github.com/hel1th/giga-url-shortener/internal/link/domain"
	"github.com/hel1th/giga-url-shortener/internal/link/usecase"
	analyticsv1 "github.com/hel1th/giga-url-shortener/proto/analytics/v1"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ usecase.EventProducer = (*LinkEventProducer)(nil)

// LinkEventProducer adapts domain event publishing to the message broker.
type LinkEventProducer struct {
	baseProducer broker.Producer
}

// NewLinkEventProducer creates a new LinkEventProducer instance.
func NewLinkEventProducer(baseProducer broker.Producer) *LinkEventProducer {
	return &LinkEventProducer{
		baseProducer: baseProducer,
	}
}

// PublishLinkClicked serializes and publishes a link clicked domain event to the broker.
func (p *LinkEventProducer) PublishLinkClicked(ctx context.Context, event domain.LinkClicked) {
	payload, err := proto.Marshal(toProtoLinkClicked(event))
	if err != nil {
		_ = fmt.Errorf("marshal link clicked event: %w", err)
	}

	envelope := broker.Envelope{
		ID:          uuid.NewString(),
		Type:        domain.EventTypeLinkClicked,
		Key:         event.ShortCode,
		OccurredAt:  event.OccurredAt,
		Payload:     payload,
		TraceParent: extractTraceParent(ctx),
	}

	p.baseProducer.Publish(ctx, envelope)
}

// extractTraceParent extracts the W3C traceparent header value from the current span context.
func extractTraceParent(ctx context.Context) string {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	return carrier.Get(broker.HeaderTraceParent)
}

// toProtoLinkClicked converts a domain LinkClicked event into a Protobuf message.
func toProtoLinkClicked(event domain.LinkClicked) *analyticsv1.LinkClicked {
	return &analyticsv1.LinkClicked{
		ShortCode:  event.ShortCode,
		OccurredAt: timestamppb.New(event.OccurredAt),
		Meta: &analyticsv1.ClickMeta{
			Ip:        event.Meta.IP,
			UserAgent: event.Meta.UserAgent,
			Referer:   event.Meta.Referer,
		},
	}
}
