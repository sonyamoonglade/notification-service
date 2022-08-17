package entity

type Event struct {
	EventID   uint64 `json:"event_id"`
	Name      string `json:"name"`
	Translate string `json:"translate"`
}
