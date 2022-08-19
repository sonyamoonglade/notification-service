package telegram

import (
	"context"
	"fmt"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/sonyamoonglade/notification-service/internal/entity"
	"github.com/sonyamoonglade/notification-service/pkg/tables"
	"go.uber.org/zap"
)

type Storage interface {
	GetTelegramSubscribers(ctx context.Context, phoneNumbers []string) ([]*entity.TelegramSubscriber, error)
}

type telegramStorage struct {
	pool   *pgxpool.Pool
	logger *zap.SugaredLogger
}

func NewTelegramStorage(logger *zap.SugaredLogger, pool *pgxpool.Pool) Storage {
	return &telegramStorage{logger: logger, pool: pool}
}

func (s *telegramStorage) GetTelegramSubscribers(ctx context.Context, phoneNumbers []string) ([]*entity.TelegramSubscriber, error) {
	whereq := ""
	for i, ph := range phoneNumbers {
		if i != len(phoneNumbers)-1 {
			whereq += fmt.Sprintf("sub.phone_number = '%s' or", ph)
		} else {
			whereq += fmt.Sprintf("sub.phone_number = '%s'", ph)
		}
	}

	mainq := fmt.Sprintf(
		"SELECT * FROM %s tgsub JOIN %s sub ON tgsub.subscriber_id = sub.subscriber_id WHERE %s",
		tables.TelegramSubscribersTable, tables.SubscribersTable, whereq)

	rows, err := s.pool.Query(ctx, mainq)
	if err != nil {
		return nil, err
	}

	var telegramSubs []*entity.TelegramSubscriber

	err = pgxscan.ScanAll(&telegramSubs, rows)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	defer rows.Close()
	return telegramSubs, nil

}
