package subscription

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sonyamoonglade/notification-service/internal/entity"
	"go.uber.org/zap"
)

type Storage interface {
	GetEventSubscribers(ctx context.Context, eventID uint64) ([]*entity.EventSubscriber, error)
}

type subscriptionStorage struct {
	pool   *pgxpool.Pool
	logger *zap.SugaredLogger
}

func NewSubscriptionStorage(logger *zap.SugaredLogger, pool *pgxpool.Pool) Storage {
	return &subscriptionStorage{pool: pool, logger: logger}
}

func (s *subscriptionStorage) GetEventSubscribers(ctx context.Context, eventID uint64) ([]*entity.EventSubscriber, error) {
	return nil, nil
}
