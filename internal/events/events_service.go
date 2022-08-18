package events

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/jackc/pgx/v4"
	"github.com/sonyamoonglade/notification-service/internal/entity"
	"github.com/sonyamoonglade/notification-service/internal/events/payload"
	"github.com/sonyamoonglade/notification-service/pkg/httpErrors"

	"go.uber.org/zap"
)

var path = "./events.json"

type Service interface {
	ReadEvents(ctx context.Context) error
	IsExists(ctx context.Context, eventName string) (uint64, error)
	RegisterEvent(ctx context.Context, e entity.Event) error
}

type eventService struct {
	eventStorage Storage
	logger       *zap.SugaredLogger
}

func NewEventsService(logger *zap.SugaredLogger, storage Storage) Service {
	return &eventService{logger: logger, eventStorage: storage}
}

func (s *eventService) RegisterEvent(ctx context.Context, e entity.Event) error {
	return s.eventStorage.RegisterEvent(ctx, e)
}

func (s *eventService) ReadEvents(ctx context.Context) error {
	_, err := os.Stat(path)
	if err != nil {
		return err
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	var content entity.Events

	bytes, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(bytes, &content); err != nil {
		return err
	}

	for _, e := range content.Events {
		event := entity.Event{
			EventID:   e.EventID,
			Name:      e.Name,
			Translate: e.Translate,
		}
		//Check if the developer assigned payload for event registered in events.json
		_, err := payload.GetProvider().GetType(e.EventID)
		if err != nil {
			return err
		}
		s.logger.Infof("payload for event %d is ok", e.EventID)
		//Only then register it/prepare to be fired
		err = s.RegisterEvent(ctx, event)
		if err != nil {
			s.logger.Errorf("could not register base event. %s", err.Error())
			return err
		}
		s.logger.Infof("event %s is ready to be fired", e.Name)
	}

	return nil
}

func (s *eventService) IsExists(ctx context.Context, eventName string) (uint64, error) {
	eventID, err := s.eventStorage.IsExist(ctx, eventName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, httpErrors.NewErrEventDoesNotExist(eventName)
		}
		return 0, err
	}
	return eventID, nil
}
