package entity

type Events struct {
	Events []Event `json:"events"`
}

type Event struct {
	EventID   uint64 `json:"event_id,omitempty" db:"event_id"`
	Name      string `json:"name" db:"name"`
	Translate string `json:"translate" db:"translate"`
}
