package domain

import (
	"time"

	analyticsv1 "github.com/hel1th/giga-url-shortener/proto/analytics/v1"
	"github.com/mileusna/useragent"
)

// Click struct for keeping analytics data.
type Click struct {
	ShortCode string
	ClickedAt time.Time
	IP        string
	UserAgent string
	Browser   string
	OS        string
	Device    string
	Referer   string
}

// NewClickFromEvent creates a new Click from an analyticsv1.LinkClicked event.
func NewClickFromEvent(event *analyticsv1.LinkClicked) Click {
	click := Click{
		ShortCode: event.ShortCode,
		ClickedAt: event.OccurredAt.AsTime(),
	}

	if event.Meta != nil {
		click.IP = event.Meta.Ip
		click.UserAgent = event.Meta.UserAgent
		click.Referer = event.Meta.Referer
	}

	ua := useragent.Parse(click.UserAgent)

	click.Device = extractDevice(ua)

	click.Browser = withDefault(ua.Name, "Unknown")
	click.OS = withDefault(ua.OS, "Unknown")

	return click
}

func extractDevice(ua useragent.UserAgent) string {
	switch {
	case ua.Mobile:
		return "Mobile"
	case ua.Tablet:
		return "Tablet"
	case ua.Desktop:
		return "Desktop"
	case ua.Bot:
		return "Bot"
	case ua.Device != "":
		return ua.Device
	default:
		return "Unknown"
	}
}

func withDefault(val, fallback string) string {
	if val == "" {
		return fallback
	}
	return val
}
