package events

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sonyamoonglade/notification-service/internal/entity"
	"go.uber.org/zap"
)

type Storage interface {
	IsExist(ctx context.Context, eventName string) (uint64, error)
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

func (e *eventStorage) IsExist(ctx context.Context, eventName string) (uint64, error) {
	var eventID uint64
	q := fmt.Sprintf("SELECT event_id FROM %s WHERE name = $1", EventTable)
	err := e.pool.QueryRow(ctx, q, eventName).Scan(&eventID)
	if err != nil {
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
