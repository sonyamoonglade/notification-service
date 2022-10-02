package events

import (
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/sonyamoonglade/notification-service/internal/entity"
	"github.com/sonyamoonglade/notification-service/internal/events/payload"
	"github.com/sonyamoonglade/notification-service/internal/storage"
	"github.com/sonyamoonglade/notification-service/pkg/http_errors"
	"github.com/sonyamoonglade/notification-service/pkg/template"

	"go.uber.org/zap"
)

var path = "./events.json"

type Service interface {
	ReadEvents(ctx context.Context) error
	DoesExist(ctx context.Context, eventName string) (uint64, error)
	RegisterEvent(ctx context.Context, e entity.Event) error
	GetAvailableEvents(ctx context.Context) ([]*entity.Event, error)
}

type eventService struct {
	storage          storage.DBStorage
	logger           *zap.SugaredLogger
	templateProvider template.Provider
}

func NewEventsService(logger *zap.SugaredLogger, storage storage.DBStorage, templateProvider template.Provider) Service {
	return &eventService{logger: logger, storage: storage, templateProvider: templateProvider}
}

func (s *eventService) GetAvailableEvents(ctx context.Context) ([]*entity.Event, error) {
	return s.storage.GetAvailableEvents(ctx)
}
func (s *eventService) RegisterEvent(ctx context.Context, e entity.Event) error {
	return s.storage.RegisterEvent(ctx, e)
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

		//Check if the developer prepared a template in templates.json for event in events.json
		_, err = s.templateProvider.Find(e.EventID)
		if err != nil {
			return err
		}
		s.logger.Infof("template for event %d is ok", e.EventID)

		//Register/justify event to be fired
		err = s.RegisterEvent(ctx, event)
		if err != nil {
			s.logger.Errorf("could not register base event. %s", err.Error())
			return err
		}
		s.logger.Infof("event %s is ready to be fired", e.Name)
	}

	return nil
}

func (s *eventService) DoesExist(ctx context.Context, eventName string) (uint64, error) {
	eventID, err := s.storage.DoesExist(ctx, eventName)
	if err != nil {
		return 0, err
	}
	if eventID == 0 {
		return 0, http_errors.NewErrEventDoesNotExist(eventName)
	}
	return eventID, nil
}
