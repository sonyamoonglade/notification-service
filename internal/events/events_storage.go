package events

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sonyamoonglade/notification-service/internal/events/dto"
	"go.uber.org/zap"
)

type Storage interface {
	IsExist(ctx context.Context, eventName string) error
	RegisterEvent(ctx context.Context, dto dto.RegisterEventDto) error
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

func (e *eventStorage) RegisterEvent(ctx context.Context, dto dto.RegisterEventDto) error {

	q := fmt.Sprintf("INSERT INTO %s (name,translate) VALUES($1,$2) ON CONFLICT DO NOTHING", EventTable)
	_, err := e.pool.Exec(ctx, q, dto.Name, dto.Translate)
	if err != nil {
		return err
	}
	return nil
}
