package domain

import "time"

const (
	// EventTypeLinkClicked is the event type name emitted when a short link is clicked.
	EventTypeLinkClicked string = "link.clicked"
	// EventTypeLinkCreated is the event type name emitted when a new short link is created.
	EventTypeLinkCreated string = "link.created"
)

// LinkClicked represents the domain event payload for a link click.
type LinkClicked struct {
	ShortCode  string
	Meta       ClickMeta
	OccurredAt time.Time
}

// LinkCreated represents the domain event payload when a link is shortened.
type LinkCreated struct {
	ShortCode   string
	OriginalURL string
	OccurredAt  time.Time
}
