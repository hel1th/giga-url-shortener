package broker

// Message broker record header keys.
const (
	// HeaderEventID is the header key for the unique event identifier.
	HeaderEventID = "event_id"
	// HeaderEventType is the header key for the event type descriptor.
	HeaderEventType = "event_type"
	// HeaderTraceParent is the header key for W3C distributed traceparent propagation.
	HeaderTraceParent = "traceparent"
)
