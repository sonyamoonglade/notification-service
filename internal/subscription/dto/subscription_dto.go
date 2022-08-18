package dto

type SubscribeToEventInp struct {
	EventName   string `json:"event_name" validate:"required"`
	PhoneNumber string `json:"phone_number" validate:"required"`
}
