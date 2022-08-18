package telegram

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sonyamoonglade/notification-service/internal/entity"
	"go.uber.org/zap"
)

const (
	TelegramSubscribersTable = "telegram_subscribers"
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
			whereq += fmt.Sprintf("phone_number = '%s' or", ph)
		} else {
			whereq += fmt.Sprintf("phone_number = '%s'", ph)
		}
	}

	mainq := fmt.Sprintf("SELECT * FROM %s WHERE %s", TelegramSubscribersTable, whereq)

	rows, err := s.pool.Query(ctx, mainq)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	var telegramSubs []*entity.TelegramSubscriber

	for rows.Next() {
		var telegramSub entity.TelegramSubscriber
		err := rows.Scan(&telegramSub)
		if err != nil {
			return nil, err
		}
		telegramSubs = append(telegramSubs, &telegramSub)
	}

	return telegramSubs, nil

}
