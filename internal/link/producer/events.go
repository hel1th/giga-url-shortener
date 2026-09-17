package producer

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hel1th/giga-url-shortener/internal/broker"
	"github.com/hel1th/giga-url-shortener/internal/link/domain"
	"github.com/hel1th/giga-url-shortener/internal/link/usecase"
	analyticsv1 "github.com/hel1th/giga-url-shortener/proto/analytics/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ usecase.EventProducer = (*LinkEventProducer)(nil)

type LinkEventProducer struct {
	baseProducer broker.Producer
}

func NewLinkEventProducer(baseProducer broker.Producer) *LinkEventProducer {
	return &LinkEventProducer{
		baseProducer: baseProducer,
	}
}

func (p *LinkEventProducer) PublishLinkClicked(ctx context.Context, event domain.LinkClicked) error {
	payload, err := proto.Marshal(toProtoLinkClicked(event))
	if err != nil {
		return fmt.Errorf("marshal link clicked event: %w", err)
	}

	envelope := broker.Envelope{
		ID:         uuid.NewString(),
		Type:       domain.EventTypeLinkClicked,
		Key:        event.ShortCode,
		OccurredAt: event.OccurredAt,
		Payload:    payload,
		// TraceParent: extractTraceParent(ctx),
	}

	if err := p.baseProducer.Publish(ctx, envelope); err != nil {
		return fmt.Errorf("publish link clicked envelope: %w", err)
	}

	return nil
}

// toProtoLinkClicked транслирует чистые доменные типы в сгенерированные Protobuf-структуры
func toProtoLinkClicked(event domain.LinkClicked) *analyticsv1.LinkClicked {
	return &analyticsv1.LinkClicked{
		ShortCode:  event.ShortCode,
		OccurredAt: timestamppb.New(event.OccurredAt),
		Meta: &analyticsv1.ClickMeta{
			Ip:        event.Meta.IP,
			UserAgent: event.Meta.UserAgent,
			Referer:   event.Meta.Referer,
			Device:    event.Meta.Device,
			Browser:   event.Meta.Browser,
			Os:        event.Meta.OS,
		},
	}
}
