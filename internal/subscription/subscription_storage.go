package subscription

import (
	"context"
	"fmt"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/sonyamoonglade/notification-service/internal/entity"
	"github.com/sonyamoonglade/notification-service/internal/subscription/response_object"
	"github.com/sonyamoonglade/notification-service/pkg/tables"
	"go.uber.org/zap"
)

type Storage interface {
	GetSubscribersDataJoined(ctx context.Context) ([]*response_object.SubscriberRO, error)
	GetEventSubscribers(ctx context.Context, eventID uint64) ([]*entity.Subscriber, error)
	GetSubscriberByPhone(ctx context.Context, phoneNumber string) (*entity.Subscriber, error)
	GetSubscription(ctx context.Context, subscriberID uint64, eventID uint64) (*entity.Subscription, error)
	GetTelegramSubscribers(ctx context.Context, phoneNumbers []string) ([]*entity.TelegramSubscriber, error)
	GetTelegramSubscriber(ctx context.Context, phoneNumber string) (*entity.TelegramSubscriber, error)
	RegisterSubscriber(ctx context.Context, phoneNumber string) (uint64, error)
	RegisterTelegramSubscriber(ctx context.Context, telegramID int64, subscriberID uint64) (bool, error)
	SubscribeToEvent(ctx context.Context, subscriberID uint64, eventID uint64) (uint64, error)
	CancelSubscription(ctx context.Context, subscriptionID uint64) (bool, error)
}

type subscriptionStorage struct {
	pool   *pgxpool.Pool
	logger *zap.SugaredLogger
}

func NewSubscriptionStorage(logger *zap.SugaredLogger, pool *pgxpool.Pool) Storage {
	return &subscriptionStorage{pool: pool, logger: logger}
}

func (s *subscriptionStorage) GetSubscribersDataJoined(ctx context.Context) ([]*response_object.SubscriberRO, error) {

	q := fmt.Sprintf(
		`SELECT sub.phone_number, COALESCE(tgsub.subscriber_id,0)::boolean as has_telegram_subscription,
				subs.subscription_id, e.name, e.translate FROM %s sub
				JOIN %s subs ON sub.subscriber_id = subs.subscriber_id
				JOIN %s e ON subs.event_id = e.event_id
				LEFT JOIN %s tgsub ON sub.subscriber_id = tgsub.subscriber_id`,
		tables.Subscribers, tables.Subscriptions, tables.Events, tables.TelegramSubscribers)
	rows, err := s.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}

	var subscribers []*response_object.SubscriberRO

	for rows.Next() {
		var subscriberRO response_object.SubscriberRO
		var subscriptionRO response_object.SubscriptionRO

		err = rows.Scan(
			&subscriberRO.PhoneNumber,
			&subscriberRO.HasTelegramSubscription,

			&subscriptionRO.SubscriptionID,
			&subscriptionRO.Event.Name,
			&subscriptionRO.Event.Translate,
		)

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return []*response_object.SubscriberRO{}, nil
			}
			return nil, err
		}

		lastIdx := len(subscribers)
		if lastIdx != 0 {
			lastIdx = lastIdx - 1
		}

		//Check if there's already this subscriber
		if len(subscribers) != 0 && subscribers[lastIdx].PhoneNumber == subscriberRO.PhoneNumber {
			//Push subscription to it's subscriptions
			subscribers[lastIdx].Subscriptions = append(subscribers[lastIdx].Subscriptions, subscriptionRO)
			continue
		}

		//Otherwise, push to the end and add subscription
		subscriberRO.Subscriptions = append(subscriberRO.Subscriptions, subscriptionRO)
		subscribers = append(subscribers, &subscriberRO)
	}

	return subscribers, nil
}

func (s *subscriptionStorage) SubscribeToEvent(ctx context.Context, subscriberID uint64, eventID uint64) (uint64, error) {
	var subscriptionID uint64
	q := fmt.Sprintf(
		"INSERT INTO %s (subscriber_id, event_id) VALUES ($1,$2) ON CONFLICT DO NOTHING RETURNING subscription_id",
		tables.Subscriptions)

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
	q := fmt.Sprintf("SELECT * FROM %s WHERE subscriber_id = $1 AND event_id = $2", tables.Subscriptions)
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
	q := fmt.Sprintf("INSERT INTO %s (phone_number) VALUES ($1) RETURNING subscriber_id", tables.Subscribers)
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
		tables.Subscribers, tables.Subscriptions)

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
	q := fmt.Sprintf("SELECT * FROM %s WHERE phone_number = $1", tables.Subscribers)
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
		tables.TelegramSubscribers)

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
		tables.TelegramSubscribers, tables.Subscribers)

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
		"SELECT tgsub.telegram_id, tgsub.subscriber_id FROM %s tgsub JOIN %s sub ON tgsub.subscriber_id = sub.subscriber_id WHERE %s",
		tables.TelegramSubscribers, tables.Subscribers, whereq)

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

func (s *subscriptionStorage) CancelSubscription(ctx context.Context, subscriptionID uint64) (bool, error) {
	q := fmt.Sprintf("DELETE FROM %s WHERE subscription_id = $1 RETURNING subscription_id", tables.Subscriptions)

	rows, err := s.pool.Query(ctx, q, subscriptionID)
	if err != nil {
		return false, err
	}

	return rows.Next(), nil

}
