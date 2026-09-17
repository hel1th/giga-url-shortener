package dto

type CreateLinkRequest struct {
	OriginalURL string `json:"original_url"`
}

type CreateLinkResponse struct {
	ShortCode   string `json:"short_code"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type LinkStatsResponse struct {
	ShortCode    string            `json:"short_code"`
	TotalClicks  uint64            `json:"total_clicks"`
	UniqueClicks uint64            `json:"unique_clicks"`
	Browsers     map[string]uint64 `json:"browsers"`
	OS           map[string]uint64 `json:"os"`
	Devices      map[string]uint64 `json:"devices"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
