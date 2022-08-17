package entity

type Subscriber struct {
	SubscriberID uint64 `json:"subscriber_id"`
	PhoneNumber  string `json:"phone_number"`
}
