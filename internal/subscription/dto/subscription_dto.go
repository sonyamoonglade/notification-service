package dto

type SubscribeToEventInp struct {
	EventID     uint64 `json:"event_id" validate:"required"`
	PhoneNumber string `json:"phone_number" validate:"required"`
}
