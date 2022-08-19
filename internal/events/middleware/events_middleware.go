package middleware

import (
	"context"
	"net/http"
	"strconv"

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

		eventIDstr := params.ByName("event_id")

		if eventIDstr == "" {
			m.logger.Debug("empty eventID string")
			httpErrors.MakeErrorResponse(w, httpErrors.ErrNoEventId)
			return
		}
		eventID, err := strconv.ParseUint(eventIDstr, 10, 64)
		if err != nil {
			m.logger.Error(err.Error())
			httpErrors.MakeErrorResponse(w, err)
			return
		}

		err = m.eventService.DoesExist(ctx, eventID)
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
