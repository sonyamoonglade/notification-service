package event_middlewares

import (
	"context"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sonyamoonglade/notification-service/internal/events"
	"github.com/sonyamoonglade/notification-service/pkg/http_errors"
	"go.uber.org/zap"
)

type DoesExist struct {
	logger       *zap.SugaredLogger
	eventService events.Service
}

func NewDoesExist(logger *zap.SugaredLogger, eventService events.Service) *DoesExist {
	return &DoesExist{
		logger:       logger,
		eventService: eventService,
	}
}

func (m *DoesExist) Check(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		ctx := r.Context()

		eventName := params.ByName("eventName")

		if eventName == "" {
			m.logger.Debug("empty eventName string")
			http_errors.MakeErrorResponse(w, http_errors.ErrNoEventName)
			return
		}

		eventID, err := m.eventService.DoesExist(ctx, eventName)
		if err != nil {
			m.logger.Error(err.Error())
			http_errors.MakeErrorResponse(w, err)
			return
		}

		withv := context.WithValue(ctx, "eventId", eventID)

		h(w, r.WithContext(withv), params)
		return

	}
}
