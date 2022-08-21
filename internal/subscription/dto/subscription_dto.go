package dto

type SubscribeToEventInp struct {
	PhoneNumber string `json:"phone_number" validate:"required"`
	EventName   string `json:"event_name" validate:"required"`
}
