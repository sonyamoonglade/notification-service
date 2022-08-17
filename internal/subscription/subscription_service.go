package subscription

import (
	"context"
	"github.com/sonyamoonglade/notification-service/internal/entity"
	"go.uber.org/zap"
)

type Service interface {
	GetEventSubscribers(ctx context.Context, eventID uint64) ([]*entity.EventSubscriber, error)
}

type subscriptionService struct {
	storage Storage
	logger  *zap.SugaredLogger
}

func NewSubscriptionService(logger *zap.SugaredLogger, storage Storage) Service {
	return &subscriptionService{logger: logger, storage: storage}
}

func (s *subscriptionService) GetEventSubscribers(ctx context.Context, eventID uint64) ([]*entity.EventSubscriber, error) {

	return nil, nil
}
