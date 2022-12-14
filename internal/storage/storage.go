package storage

import (
	"context"
	"fmt"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/sonyamoonglade/notification-service/internal/entity"
	"github.com/sonyamoonglade/notification-service/internal/subscription/response_object"
	"go.uber.org/zap"
)

type DBStorage interface {
	GetSubscribersDataJoined(ctx context.Context) ([]*response_object.SubscriberRO, error)
	GetEventSubscribers(ctx context.Context, eventID uint64) ([]*entity.Subscriber, error)
	GetSubscribersWithoutSubs(ctx context.Context) ([]*response_object.SubscriberRO, error)
	GetSubscriberByPhone(ctx context.Context, phoneNumber string) (*entity.Subscriber, error)
	GetSubscription(ctx context.Context, subscriberID uint64, eventID uint64) (*entity.Subscription, error)
	GetTelegramSubscribers(ctx context.Context, phoneNumbers []string) ([]*entity.TelegramSubscriber, error)
	GetTelegramSubscriber(ctx context.Context, phoneNumber string) (*entity.TelegramSubscriber, error)
	RegisterSubscriber(ctx context.Context, phoneNumber string) (uint64, error)
	RegisterTelegramSubscriber(ctx context.Context, telegramID int64, subscriberID uint64) (bool, error)
	SubscribeToEvent(ctx context.Context, subscriberID uint64, eventID uint64) (uint64, error)
	CancelSubscription(ctx context.Context, subscriptionID uint64) (bool, error)
	DoesExist(ctx context.Context, eventName string) (uint64, error)
	GetAvailableEvents(ctx context.Context) ([]*entity.Event, error)
	RegisterEvent(ctx context.Context, e entity.Event) error
}

const (
	subscribersTable         = "subscribers"
	subscriptionsTable       = "subscriptions"
	telegramSubscribersTable = "telegram_subscribers"
	eventsTable              = "events"
)

type PostgresStorage struct {
	pool   *pgxpool.Pool
	logger *zap.SugaredLogger
}

func NewPostgresStorage(logger *zap.SugaredLogger, pool *pgxpool.Pool) *PostgresStorage {
	return &PostgresStorage{pool: pool, logger: logger}
}

func (p *PostgresStorage) GetSubscribersWithoutSubs(ctx context.Context) ([]*response_object.SubscriberRO, error) {

	q := fmt.Sprintf(`
		SELECT sub.phone_number, COALESCE(tgsub.subscriber_id, 0)::boolean as has_telegram_subscription
		FROM %s sub LEFT JOIN %s tgsub ON sub.subscriber_id = tgsub.subscriber_id ORDER BY sub.phone_number ASC`,
		subscribersTable, telegramSubscribersTable)

	c, err := p.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Release()

	rows, err := c.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscribers []*response_object.SubscriberRO

	for rows.Next() {
		var subscriber response_object.SubscriberRO
		err = rows.Scan(&subscriber.PhoneNumber, &subscriber.HasTelegramSubscription)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return []*response_object.SubscriberRO{}, nil
			}
			return nil, err
		}

		subscribers = append(subscribers, &subscriber)
	}

	return subscribers, nil
}

func (p *PostgresStorage) GetSubscribersDataJoined(ctx context.Context) ([]*response_object.SubscriberRO, error) {

	q := fmt.Sprintf(
		`SELECT sub.phone_number, COALESCE(tgsub.subscriber_id,0)::boolean as has_telegram_subscription,
				subs.subscription_id, e.name, e.translate, e.event_id FROM %s sub
				JOIN %s subs ON sub.subscriber_id = subs.subscriber_id
				JOIN %s e ON subs.event_id = e.event_id
				LEFT JOIN %s tgsub ON sub.subscriber_id = tgsub.subscriber_id
				ORDER BY sub.phone_number ASC`,
		subscribersTable, subscriptionsTable, eventsTable, telegramSubscribersTable)

	c, err := p.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Release()

	rows, err := c.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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
			&subscriptionRO.Event.EventID,
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

func (p *PostgresStorage) SubscribeToEvent(ctx context.Context, subscriberID uint64, eventID uint64) (uint64, error) {
	var subscriptionID uint64
	q := fmt.Sprintf(
		"INSERT INTO %s (subscriber_id, event_id) VALUES ($1,$2) ON CONFLICT DO NOTHING RETURNING subscription_id",
		subscriptionsTable)

	c, err := p.pool.Acquire(ctx)
	if err != nil {
		return 0, nil
	}
	defer c.Release()

	err = c.QueryRow(ctx, q, subscriberID, eventID).Scan(&subscriptionID)
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

func (p *PostgresStorage) GetSubscription(ctx context.Context, subscriberID uint64, eventID uint64) (*entity.Subscription, error) {
	var sub entity.Subscription

	q := fmt.Sprintf("SELECT * FROM %s WHERE subscriber_id = $1 AND event_id = $2", subscriptionsTable)
	c, err := p.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Release()

	rows, err := c.Query(ctx, q, subscriberID, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	err = pgxscan.ScanOne(&sub, rows)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &sub, nil
}

func (p *PostgresStorage) RegisterSubscriber(ctx context.Context, phoneNumber string) (uint64, error) {
	var subscriberID uint64
	q := fmt.Sprintf("INSERT INTO %s (phone_number) VALUES ($1) RETURNING subscriber_id", subscribersTable)

	c, err := p.pool.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer c.Release()

	err = c.QueryRow(ctx, q, phoneNumber).Scan(&subscriberID)
	if err != nil {
		return 0, err
	}
	return subscriberID, nil
}

func (p *PostgresStorage) GetEventSubscribers(ctx context.Context, eventID uint64) ([]*entity.Subscriber, error) {
	var subs []*entity.Subscriber
	q := fmt.Sprintf(
		`SELECT sub.subscriber_id, sub.phone_number FROM %s sub JOIN %s subs ON sub.subscriber_id = subs.subscriber_id WHERE subs.event_id = $1`,
		subscribersTable, subscriptionsTable)

	c, err := p.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Release()

	rows, err := c.Query(ctx, q, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	err = pgxscan.ScanAll(&subs, rows)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return subs, nil
}

func (p *PostgresStorage) GetSubscriberByPhone(ctx context.Context, phoneNumber string) (*entity.Subscriber, error) {
	var sub entity.Subscriber
	q := fmt.Sprintf("SELECT * FROM %s WHERE phone_number = $1", subscribersTable)

	c, err := p.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Release()

	rows, err := c.Query(ctx, q, phoneNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	err = pgxscan.ScanOne(&sub, rows)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &sub, nil
}

func (p *PostgresStorage) RegisterTelegramSubscriber(ctx context.Context, telegramID int64, subscriberID uint64) (bool, error) {

	q := fmt.Sprintf(
		"INSERT INTO %s (telegram_id, subscriber_id) VALUES ($1,$2) ON CONFLICT DO NOTHING RETURNING telegram_id",
		telegramSubscribersTable)

	c, err := p.pool.Acquire(ctx)
	if err != nil {
		return false, err
	}
	defer c.Release()

	rows, err := c.Query(ctx, q, telegramID, subscriberID)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	var placeholder int64

	err = pgxscan.ScanOne(&placeholder, rows)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (p *PostgresStorage) GetTelegramSubscriber(ctx context.Context, phoneNumber string) (*entity.TelegramSubscriber, error) {
	q := fmt.Sprintf(
		`SELECT tgsub.telegram_id, tgsub.subscriber_id FROM %s tgsub JOIN %s sub ON
				tgsub.subscriber_id = sub.subscriber.id WHERE sub.phone_number = $1`,
		telegramSubscribersTable, subscribersTable)

	c, err := p.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Release()

	rows, err := c.Query(ctx, q, phoneNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tgsub entity.TelegramSubscriber

	err = pgxscan.ScanOne(&tgsub, rows)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &tgsub, nil
}

func (p *PostgresStorage) GetTelegramSubscribers(ctx context.Context, phoneNumbers []string) ([]*entity.TelegramSubscriber, error) {
	whereq := ""
	for i, ph := range phoneNumbers {
		if i != len(phoneNumbers)-1 {
			whereq += fmt.Sprintf("sub.phone_number = '%s' or ", ph)
		} else {
			whereq += fmt.Sprintf("sub.phone_number = '%s'", ph)
		}
	}

	mainq := fmt.Sprintf(
		"SELECT tgsub.telegram_id, tgsub.subscriber_id FROM %s tgsub JOIN %s sub ON tgsub.subscriber_id = sub.subscriber_id WHERE %s",
		telegramSubscribersTable, subscribersTable, whereq)

	c, err := p.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Release()

	rows, err := c.Query(ctx, mainq)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var telegramSubs []*entity.TelegramSubscriber

	err = pgxscan.ScanAll(&telegramSubs, rows)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return telegramSubs, nil

}

func (p *PostgresStorage) CancelSubscription(ctx context.Context, subscriptionID uint64) (bool, error) {
	q := fmt.Sprintf("DELETE FROM %s WHERE subscription_id = $1 RETURNING subscription_id", subscriptionsTable)

	c, err := p.pool.Acquire(ctx)
	if err != nil {
		return false, err
	}
	defer c.Release()

	rows, err := c.Query(ctx, q, subscriptionID)
	if err != nil {
		return false, err
	}

	defer rows.Close()

	return rows.Next(), nil

}

func (p *PostgresStorage) DoesExist(ctx context.Context, eventName string) (uint64, error) {
	var eventID uint64
	q := fmt.Sprintf("SELECT event_Id FROM %s WHERE name = $1", eventsTable)
	err := p.pool.QueryRow(ctx, q, eventName).Scan(&eventID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return eventID, nil
}

func (p *PostgresStorage) RegisterEvent(ctx context.Context, ev entity.Event) error {
	q := fmt.Sprintf("INSERT INTO %s (event_id, name,translate) VALUES($1,$2,$3) ON CONFLICT DO NOTHING", eventsTable)
	_, err := p.pool.Exec(ctx, q, ev.EventID, ev.Name, ev.Translate)
	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresStorage) GetAvailableEvents(ctx context.Context) ([]*entity.Event, error) {
	q := fmt.Sprintf(`SELECT * FROM %s`, eventsTable)

	c, err := p.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Release()

	rows, err := c.Query(ctx, q)
	if err != nil {
		return nil, err
	}

	var events []*entity.Event

	err = pgxscan.ScanAll(&events, rows)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []*entity.Event{}, nil
		}
		return nil, err
	}

	return events, nil
}
