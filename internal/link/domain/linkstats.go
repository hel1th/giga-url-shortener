package domain

// LinkStats представляет собой доменную модель агрегированной статистики по ссылке.
type LinkStats struct {
	ShortCode    string
	TotalClicks  uint64
	UniqueClicks uint64
	Browsers     map[string]uint64
	OS           map[string]uint64
	Devices      map[string]uint64
}
