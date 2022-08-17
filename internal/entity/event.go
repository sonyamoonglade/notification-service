package entity

type Event struct {
	EventID   uint64 `json:"event_id,omitempty"`
	Name      string `json:"name"`
	Translate string `json:"translate"`
}

type Events struct {
	Events []Event `json:"events"`
}