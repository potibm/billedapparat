package mastodon

import "encoding/json"

const (
	EventTypeUpdate   = "update"
	EventStatusUpdate = "status.update"
	EventTypeDelete   = "delete"
)

type Event struct {
	Type    string
	Payload string
}

func (e *Event) Status() (MastoStatus, error) {
	var status MastoStatus

	if err := json.Unmarshal([]byte(e.Payload), &status); err != nil {
		return MastoStatus{}, err
	}

	return status, nil
}
