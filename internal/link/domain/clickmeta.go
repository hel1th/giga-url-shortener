package domain

import "time"

type ClickMeta struct {
	ShortCode string
	ClickedAt time.Time
	IP        string
	UserAgent string
	Referer   string
	Device    string
	Browser   string
	OS        string
}
