// Package domain contains domain entities, domain events, and core business rules.
package domain

import "time"

// ClickMeta captures metadata associated with a link click or redirect.
type ClickMeta struct {
	ShortCode string
	ClickedAt time.Time
	IP        string
	UserAgent string
	Referer   string
}
