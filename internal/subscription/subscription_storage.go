package subscription

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
	GetEventSubscribers(ctx context.Context, eventID uint64) ([]*entity.Subscriber, error)
	GetSubscriberByPhone(ctx context.Context, phoneNumber string) (*entity.Subscriber, error)
	GetSubscription(ctx context.Context, subscriberID uint64, eventID uint64) (*entity.Subscription, error)
	GetTelegramSubscribers(ctx context.Context, phoneNumbers []string) ([]*entity.TelegramSubscriber, error)
	GetTelegramSubscriber(ctx context.Context, phoneNumber string) (*entity.TelegramSubscriber, error)
	RegisterSubscriber(ctx context.Context, phoneNumber string) (uint64, error)
	RegisterTelegramSubscriber(ctx context.Context, telegramID int64, subscriberID uint64) (bool, error)
	SubscribeToEvent(ctx context.Context, subscriberID uint64, eventID uint64) (uint64, error)
}

type subscriptionStorage struct {
	pool   *pgxpool.Pool
	logger *zap.SugaredLogger
}

func NewSubscriptionStorage(logger *zap.SugaredLogger, pool *pgxpool.Pool) Storage {
	return &subscriptionStorage{pool: pool, logger: logger}
}

func (s *subscriptionStorage) SubscribeToEvent(ctx context.Context, subscriberID uint64, eventID uint64) (uint64, error) {
	var subscriptionID uint64
	q := fmt.Sprintf(
		"INSERT INTO %s (subscriber_id, event_id) VALUES ($1,$2) ON CONFLICT DO NOTHING RETURNING subscription_id",
		tables.SubscriptionsTable)

	err := s.pool.QueryRow(ctx, q, subscriberID, eventID).Scan(&subscriptionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	if subscriptionID == 0 {
		return 0, nil
	}
	return subscriptionID, nil
}

func (s *subscriptionStorage) GetSubscription(ctx context.Context, subscriberID uint64, eventID uint64) (*entity.Subscription, error) {
	var sub entity.Subscription
	q := fmt.Sprintf("SELECT * FROM %s WHERE subscriber_id = $1 AND event_id = $2", tables.SubscriptionsTable)
	rows, err := s.pool.Query(ctx, q, subscriberID, eventID)
	if err != nil {
		return nil, err
	}
	err = pgxscan.ScanOne(&sub, rows)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &sub, nil
}

func (s *subscriptionStorage) RegisterSubscriber(ctx context.Context, phoneNumber string) (uint64, error) {
	var subscriberID uint64
	q := fmt.Sprintf("INSERT INTO %s (phone_number) VALUES ($1) RETURNING subscriber_id", tables.SubscribersTable)
	err := s.pool.QueryRow(ctx, q, phoneNumber).Scan(&subscriberID)
	if err != nil {
		return 0, err
	}
	return subscriberID, nil
}

func (s *subscriptionStorage) GetEventSubscribers(ctx context.Context, eventID uint64) ([]*entity.Subscriber, error) {
	var subs []*entity.Subscriber
	q := fmt.Sprintf(
		`SELECT sub.subscriber_id, sub.phone_number FROM %s sub JOIN %s subs ON sub.subscriber_id = subs.subscriber_id WHERE subs.event_id = $1`,
		tables.SubscribersTable, tables.SubscriptionsTable)

	rows, err := s.pool.Query(ctx, q, eventID)
	if err != nil {
		return nil, err
	}

	err = pgxscan.ScanAll(&subs, rows)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	defer rows.Close()
	return subs, nil
}

func (s *subscriptionStorage) GetSubscriberByPhone(ctx context.Context, phoneNumber string) (*entity.Subscriber, error) {
	var sub entity.Subscriber
	q := fmt.Sprintf("SELECT * FROM %s WHERE phone_number = $1", tables.SubscribersTable)
	rows, err := s.pool.Query(ctx, q, phoneNumber)
	if err != nil {
		return nil, err
	}
	err = pgxscan.ScanOne(&sub, rows)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &sub, nil
}

func (s *subscriptionStorage) RegisterTelegramSubscriber(ctx context.Context, telegramID int64, subscriberID uint64) (bool, error) {

	q := fmt.Sprintf(
		"INSERT INTO %s (telegram_id, subscriber_id) VALUES ($1,$2) ON CONFLICT DO NOTHING RETURNING telegram_id",
		tables.TelegramSubscribersTable)

	rows, err := s.pool.Query(ctx, q, telegramID, subscriberID)
	if err != nil {
		return false, err
	}

	var placeholder int64

	err = pgxscan.ScanOne(&placeholder, rows)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	defer rows.Close()
	return true, nil
}

func (s *subscriptionStorage) GetTelegramSubscriber(ctx context.Context, phoneNumber string) (*entity.TelegramSubscriber, error) {
	q := fmt.Sprintf(
		`SELECT tgsub.telegram_id, tgsub.subscriber_id FROM %s tgsub JOIN %s sub ON
				tgsub.subscriber_id = sub.subscriber.id WHERE sub.phone_number = $1`,
		tables.TelegramSubscribersTable, tables.SubscribersTable)

	rows, err := s.pool.Query(ctx, q, phoneNumber)
	if err != nil {
		return nil, err
	}

	var tgsub entity.TelegramSubscriber

	err = pgxscan.ScanOne(&tgsub, rows)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	defer rows.Close()
	return &tgsub, nil
}

func (s *subscriptionStorage) GetTelegramSubscribers(ctx context.Context, phoneNumbers []string) ([]*entity.TelegramSubscriber, error) {
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
