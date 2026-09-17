package domain

import "time"

const (
	EventTypeLinkClicked string = "link.clicked"
	EventTypeLinkCreated string = "link.created"
)

type LinkClicked struct {
	ShortCode  string
	Meta       ClickMeta
	OccurredAt time.Time
}

type LinkCreated struct {
	ShortCode   string
	OriginalURL string
	OccurredAt  time.Time
}
