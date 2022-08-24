package events

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
	DoesExist(ctx context.Context, eventName string) (uint64, error)
	GetAvailableEvents(ctx context.Context) ([]*entity.Event, error)
	RegisterEvent(ctx context.Context, e entity.Event) error
}

type eventStorage struct {
	pool   *pgxpool.Pool
	logger *zap.SugaredLogger
}

const (
	EventTable = "events"
)

func NewEventStorage(logger *zap.SugaredLogger, pool *pgxpool.Pool) Storage {
	return &eventStorage{logger: logger, pool: pool}
}

func (e *eventStorage) DoesExist(ctx context.Context, eventName string) (uint64, error) {
	var eventID uint64
	q := fmt.Sprintf("SELECT event_Id FROM %s WHERE name = $1", EventTable)
	err := e.pool.QueryRow(ctx, q, eventName).Scan(&eventID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return eventID, nil
}

func (e *eventStorage) RegisterEvent(ctx context.Context, ev entity.Event) error {
	q := fmt.Sprintf("INSERT INTO %s (event_id, name,translate) VALUES($1,$2,$3) ON CONFLICT DO NOTHING", EventTable)
	_, err := e.pool.Exec(ctx, q, ev.EventID, ev.Name, ev.Translate)
	if err != nil {
		return err
	}
	return nil
}

func (e *eventStorage) GetAvailableEvents(ctx context.Context) ([]*entity.Event, error) {
	q := fmt.Sprintf(`SELECT * FROM %s`, tables.Events)

	c, err := e.pool.Acquire(ctx)
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
