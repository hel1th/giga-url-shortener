package domain

// LinkSummary struct contains stats about concrete link 
type LinkSummary struct {
	TotalClicks  uint64
	UniqueClicks uint64
	Browser      map[string]uint64
	OS           map[string]uint64
	Device       map[string]uint64
	Referer      map[string]uint64
}
