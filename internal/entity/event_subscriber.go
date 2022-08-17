package entity

type EventSubscriber struct {
	EventID     uint64 `json:"event_id"`
	PhoneNumber string `json:"phone_number"`
}
