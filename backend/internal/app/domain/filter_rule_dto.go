package domain

type FilterType string

const (
	FilterTypeLanguage    FilterType = "language"
	FilterTypeDisplayName FilterType = "display_name"
)

type FilterRule struct {
	ID     int64      `json:"id"`
	Source string     `json:"source"`
	Type   FilterType `json:"type"`
	Value  string     `json:"value"`
}
