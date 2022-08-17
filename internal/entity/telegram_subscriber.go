package entity

type TelegramSubscriber struct {
	SubscriberID uint64 `json:"subscriber_id"`
	TelegramID   int64  `json:"telegram_id"`
}
