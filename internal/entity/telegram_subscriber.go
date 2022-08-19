package entity

type TelegramSubscriber struct {
	SubscriberID uint64 `json:"subscriber_id" db:"subscriber_id"`
	TelegramID   int64  `json:"telegram_id" db:"telegram_id"`
}
