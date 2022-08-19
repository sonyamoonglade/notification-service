package entity

type Subscription struct {
	SubscriptionID uint64 `json:"subscription_id" db:"subscription_id"`
	EventID        uint64 `json:"event_id" db:"event_id"`
	SubscriberID   uint64 `json:"subscriber_id" db:"subscriber_id"`
}
