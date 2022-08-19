package telegram

import (
	"context"

	"github.com/sonyamoonglade/notification-service/internal/entity"
	"go.uber.org/zap"
)

type Service interface {
	GetTelegramSubscribers(ctx context.Context, phoneNumbers []string) ([]*entity.TelegramSubscriber, error)
}

type telegramService struct {
	storage Storage
	logger  *zap.SugaredLogger
}

func NewTelegramService(logger *zap.SugaredLogger, storage Storage) Service {
	return &telegramService{logger: logger, storage: storage}
}

func (s *telegramService) GetTelegramSubscribers(ctx context.Context, phoneNumbers []string) ([]*entity.TelegramSubscriber, error) {
	return s.storage.GetTelegramSubscribers(ctx, phoneNumbers)
}
