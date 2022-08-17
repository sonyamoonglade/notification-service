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
		if err := m.eventService.IsExists(ctx, eventName); err != nil {
			m.logger.Errorf("eventService.IsExists err. %s", err.Error())
			httpErrors.MakeErrorResponse(w, err)
		}

		withv := context.WithValue(ctx, "eventName", eventName)
		r.WithContext(withv)

		h(w, r, params)
		return
	}
}
