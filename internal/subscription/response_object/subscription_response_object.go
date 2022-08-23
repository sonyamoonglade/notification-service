package response_object

type EventRO struct {
	Name      string `json:"name"`
	Translate string `json:"translate"`
}

type SubscriptionRO struct {
	SubscriptionID uint64  `json:"subscription_id"`
	Event          EventRO `json:"event"`
}

type SubscriberRO struct {
	HasTelegramSubscription bool             `json:"has_telegram_subscription" db:"has_telegram_subscription"`
	PhoneNumber             string           `json:"phone_number" db:"phone_number"`
	Subscriptions           []SubscriptionRO `json:"subscriptions" db:"subscriptions"`
}
