package events

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

type Storage interface {
	IsExist(ctx context.Context, eventName string) error
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

func (e *eventStorage) IsExist(ctx context.Context, eventName string) error {
	var ok bool
	q := fmt.Sprintf("SELECT true FROM %s WHERE name = $1", EventTable)
	row := e.pool.QueryRow(ctx, q, eventName)
	if err := row.Scan(&ok); err != nil {
		return err
	}
	return nil
}
