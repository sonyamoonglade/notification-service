package events

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/sonyamoonglade/notification-service/internal/entity"
	"go.uber.org/zap"
)

type Storage interface {
	DoesExist(ctx context.Context, eventID uint64) (bool, error)
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

func (e *eventStorage) DoesExist(ctx context.Context, eventID uint64) (bool, error) {
	var ok bool
	q := fmt.Sprintf("SELECT true FROM %s WHERE event_id = $1", EventTable)
	err := e.pool.QueryRow(ctx, q, eventID).Scan(&ok)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (e *eventStorage) RegisterEvent(ctx context.Context, ev entity.Event) error {
	q := fmt.Sprintf("INSERT INTO %s (event_id, name,translate) VALUES($1,$2,$3) ON CONFLICT DO NOTHING", EventTable)
	_, err := e.pool.Exec(ctx, q, ev.EventID, ev.Name, ev.Translate)
	if err != nil {
		return err
	}
	return nil
}
