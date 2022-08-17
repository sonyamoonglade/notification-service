package subscription

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sonyamoonglade/notification-service/internal/events"
	"github.com/sonyamoonglade/notification-service/internal/events/middleware"
	"github.com/sonyamoonglade/notification-service/pkg/httpErrors"
	"github.com/sonyamoonglade/notification-service/pkg/telegram"
	"github.com/sonyamoonglade/notification-service/pkg/template"
	"go.uber.org/zap"
)

type Transport interface {
	Fire(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	InitRoutes(router *httprouter.Router)
}

type subscriptionTransport struct {
	subscriptionService Service
	telegramService     telegram.Service
	eventsService       events.Service
	templateProvider    template.TemplateProvider
	logger              *zap.SugaredLogger
	eventsMiddlewares   *middleware.EventsMiddlewares
}

func (s *subscriptionTransport) InitRoutes(router *httprouter.Router) {

	router.POST("/api/events/fire/:eventName", s.eventsMiddlewares.DoesEventExist(s.Fire))
}

func NewSubscriptionTransport(logger *zap.SugaredLogger,
	service Service,
	eventMiddlewares *middleware.EventsMiddlewares,
	eventsService events.Service,
	telegramService telegram.Service,
	templateProvider template.TemplateProvider) Transport {

	return &subscriptionTransport{
		logger:              logger,
		subscriptionService: service,
		eventsMiddlewares:   eventMiddlewares,
		eventsService:       eventsService,
		telegramService:     telegramService,
		templateProvider:    templateProvider,
	}
}

func (s *subscriptionTransport) Fire(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := r.Context()
	eventID := ctx.Value("eventId").(uint64)
	subscribers, err := s.subscriptionService.GetEventSubscribers(ctx, eventID)
	if err != nil {
		httpErrors.MakeErrorResponse(w, err)
		s.logger.Error(err.Error())
		return
	}

	var subscriberPhones []string
	for _, s := range subscribers {
		subscriberPhones = append(subscriberPhones, s.PhoneNumber)
	}

	telegramIDs, err := s.telegramService.GetTelegramSubscribers(subscriberPhones)
	if err != nil {
		httpErrors.MakeErrorResponse(w, err)
		s.logger.Error(err.Error())
		return
	}

	template, err := s.templateProvider.Find(eventID)
	if err != nil {
		httpErrors.MakeErrorResponse(w, err)
		s.logger.Error(err.Error())
		return
	}

	for _, telegramSubID := range telegramIDs {

	}

}
