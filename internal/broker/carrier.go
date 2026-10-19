package broker

import (
	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel/propagation"
)

// NewCarrier converts slice of headers to map which implmenets MapCarrier interface
func NewCarrier(headers []kgo.RecordHeader) propagation.MapCarrier {
	m := make(propagation.MapCarrier, len(headers))
	for _, h := range headers {
		m[h.Key] = string(h.Value)
	}
	return m
}
