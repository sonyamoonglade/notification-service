package events

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v4"
	"github.com/sonyamoonglade/notification-service/internal/entity"
	"github.com/sonyamoonglade/notification-service/internal/events/dto"
	"github.com/sonyamoonglade/notification-service/pkg/httpErrors"
	"go.uber.org/zap"
	"io"
	"os"
)

type Service interface {
	ReadEvents(ctx context.Context) error
	IsExists(ctx context.Context, eventName string) error
	RegisterEvent(ctx context.Context, dto dto.RegisterEventDto) error
}

type eventService struct {
	eventStorage Storage
	logger       *zap.SugaredLogger
}

func NewEventsService(logger *zap.SugaredLogger, storage Storage) Service {
	return &eventService{logger: logger, eventStorage: storage}
}

func (s *eventService) RegisterEvent(ctx context.Context, dto dto.RegisterEventDto) error {
	return s.eventStorage.RegisterEvent(ctx, dto)
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
		evDto := dto.RegisterEventDto{
			Name:      e.Name,
			Translate: e.Translate,
		}
		err := s.RegisterEvent(ctx, evDto)
		if err != nil {
			s.logger.Errorf("could not register base event. %s", err.Error())
			return err
		}
		s.logger.Infof("event %s is ready to be fired", e.Name)
	}

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
