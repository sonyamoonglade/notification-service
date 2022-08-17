package telegram

import "github.com/sonyamoonglade/notification-service/internal/entity"

type Service interface {
	GetTelegramSubscribers(phoneNumbers []string) ([]*entity.TelegramSubscriber, error)
}
