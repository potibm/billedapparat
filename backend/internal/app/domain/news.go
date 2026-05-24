package domain

type News struct {
	ID          int64  `json:"id"`
	Source      string `json:"source"`
	ExternalID  string `json:"external_id"`
	Title       string `json:"title"`
	Body        string `json:"body"`
	IsUrgent    bool   `json:"is_urgent"`
	ExternalURL string `json:"external_url,omitempty"`
	IsHidden    bool   `json:"is_hidden"`
}
