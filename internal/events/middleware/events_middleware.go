package middleware

import (
	"context"
	"github.com/julienschmidt/httprouter"
	"github.com/sonyamoonglade/notification-service/internal/events"
	"github.com/sonyamoonglade/notification-service/pkg/httpErrors"
	"go.uber.org/zap"
	"net/http"
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
		eventName := params.ByName("eventName")
		if eventName == "" {
			httpErrors.MakeErrorResponse(w, httpErrors.ErrNoEventId)
			return
		}
		ctx := r.Context()
		eventID, err := m.eventService.IsExists(ctx, eventName)
		if err != nil {
			m.logger.Errorf("eventService.IsExists err. %s", err.Error())
			httpErrors.MakeErrorResponse(w, err)
			return
		}

		withv := context.WithValue(ctx, "eventId", eventID)

		h(w, r.WithContext(withv), params)
		return

	}
}
