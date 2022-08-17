package events

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/sonyamoonglade/notification-service/pkg/httpErrors"
	"go.uber.org/zap"
	"io"
	"os"
)

type Service interface {
	ReadEvents() error
	IsExists(ctx context.Context, eventName string) error
}

type eventService struct {
	eventStorage Storage
	logger       *zap.SugaredLogger
}

func NewEventsService(logger *zap.SugaredLogger, storage Storage) Service {
	return &eventService{logger: logger, eventStorage: storage}
}

func (s *eventService) ReadEvents() error {
	_, err := os.Stat(path)
	if err != nil {
		return err
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}

	bytes, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	content := string(bytes)
	fmt.Println(content)
	return nil
}

func (s *eventService) IsExists(ctx context.Context, eventName string) error {
	err := s.eventStorage.IsExist(ctx, eventName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return httpErrors.NewErrEventDoesNotExist(eventName)
		}
		return err
	}
	return nil
}

var path = "./events.json"
