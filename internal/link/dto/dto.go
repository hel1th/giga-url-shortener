// Package dto contains Data Transfer Objects for the HTTP transport layer.
package dto

// CreateLinkRequest defines the request body payload for creating a shortened link.
type CreateLinkRequest struct {
	OriginalURL string `json:"original_url"`
}

// CreateLinkResponse defines the response payload returned after creating a short link.
type CreateLinkResponse struct {
	ShortCode   string `json:"short_code"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// LinkStatsResponse defines the analytics response payload for a short link.
type LinkStatsResponse struct {
	ShortCode    string            `json:"short_code"`
	TotalClicks  uint64            `json:"total_clicks"`
	UniqueClicks uint64            `json:"unique_clicks"`
	Browsers     map[string]uint64 `json:"browsers"`
	OS           map[string]uint64 `json:"os"`
	Devices      map[string]uint64 `json:"devices"`
}

// ErrorResponse defines the standard JSON error payload returned by HTTP handlers.
type ErrorResponse struct {
	Error string `json:"error"`
}
