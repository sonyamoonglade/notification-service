package response_object

import "github.com/sonyamoonglade/notification-service/internal/entity"

type SubscriptionRO struct {
	SubscriptionID uint64       `json:"subscription_id"`
	Event          entity.Event `json:"event"`
}

type SubscriberRO struct {
	PhoneNumber             string           `json:"phone_number" db:"phone_number"`
	HasTelegramSubscription bool             `json:"has_telegram_subscription" db:"has_telegram_subscription"`
	Subscriptions           []SubscriptionRO `json:"subscriptions,omitempty" db:"subscriptions"`
}
