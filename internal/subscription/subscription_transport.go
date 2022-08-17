package subscription

import (
	"github.com/julienschmidt/httprouter"
	"github.com/sonyamoonglade/notification-service/internal/events/middleware"
	"go.uber.org/zap"
	"net/http"
)

type Transport interface {
	Fire(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	InitRoutes(router *httprouter.Router)
}

type subscriptionTransport struct {
	service           Service
	logger            *zap.SugaredLogger
	eventsMiddlewares *middleware.EventsMiddlewares
}

func (s *subscriptionTransport) InitRoutes(router *httprouter.Router) {
	api := "/api"
	router.POST(api+"/events/fire/:eventName", s.eventsMiddlewares.DoesEventExist(s.Fire))
}

func NewSubscriptionTransport(logger *zap.SugaredLogger, service Service, eventMiddlewares *middleware.EventsMiddlewares) Transport {
	return &subscriptionTransport{logger: logger, service: service, eventsMiddlewares: eventMiddlewares}
}

func (s *subscriptionTransport) Fire(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Write([]byte("fire!"))
}
