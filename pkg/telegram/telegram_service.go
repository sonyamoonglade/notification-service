package telegram

import (
	"github.com/sonyamoonglade/notification-service/internal/entity"
	"go.uber.org/zap"
)

type Service interface {
	GetTelegramSubscribers(phoneNumbers []string) ([]*entity.TelegramSubscriber, error)
}

type telegramService struct {
	telegramStorage Storage
	logger          *zap.SugaredLogger
}

func NewTelegramService(logger *zap.SugaredLogger, storage Storage) Service {
	return &telegramService{logger: logger, telegramStorage: storage}
}

func (t telegramService) GetTelegramSubscribers(phoneNumbers []string) ([]*entity.TelegramSubscriber, error) {
	//TODO implement me
	panic("implement me")
}
