package middleware

import (
	"context"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sonyamoonglade/notification-service/internal/events"
	"github.com/sonyamoonglade/notification-service/pkg/httpErrors"
	"go.uber.org/zap"
)

type EventsMiddlewares struct {
	eventService events.Service
	logger       *zap.SugaredLogger
}

func NewEventsMiddlewares(logger *zap.SugaredLogger, service events.Service) *EventsMiddlewares {
	return &EventsMiddlewares{eventService: service, logger: logger}
}

func (m *EventsMiddlewares) DoesEventExist(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		ctx := r.Context()

		eventName := params.ByName("eventName")

		if eventName == "" {
			m.logger.Debug("empty eventName string")
			httpErrors.MakeErrorResponse(w, httpErrors.ErrNoEventName)
			return
		}

		eventID, err := m.eventService.DoesExist(ctx, eventName)
		if err != nil {
			m.logger.Error(err.Error())
			httpErrors.MakeErrorResponse(w, err)
			return
		}

		withv := context.WithValue(ctx, "eventId", eventID)

		h(w, r.WithContext(withv), params)
		return

	}
}
