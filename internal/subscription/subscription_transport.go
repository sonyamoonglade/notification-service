package subscription

import (
	"net/http"
	"reflect"

	"github.com/julienschmidt/httprouter"
	"github.com/sonyamoonglade/delivery-service/pkg/binder"
	"github.com/sonyamoonglade/notification-service/internal/events"
	"github.com/sonyamoonglade/notification-service/internal/events/middleware"
	"github.com/sonyamoonglade/notification-service/internal/events/payload"
	"github.com/sonyamoonglade/notification-service/pkg/bot"
	"github.com/sonyamoonglade/notification-service/pkg/formatter"
	"github.com/sonyamoonglade/notification-service/pkg/httpErrors"
	"github.com/sonyamoonglade/notification-service/pkg/httpRes"
	"github.com/sonyamoonglade/notification-service/pkg/telegram"
	"github.com/sonyamoonglade/notification-service/pkg/template"
	"go.uber.org/zap"
)

type Transport interface {
	Fire(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	Subscribe(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	InitRoutes(router *httprouter.Router)
}

type subscriptionTransport struct {
	subscriptionService Service
	telegramService     telegram.Service
	eventsService       events.Service
	templateProvider    template.Provider
	formatter           formatter.Formatter
	logger              *zap.SugaredLogger
	eventsMiddlewares   *middleware.EventsMiddlewares
	bot                 bot.Bot
}

func (s *subscriptionTransport) InitRoutes(router *httprouter.Router) {
	router.POST("/api/events/fire/:eventName", s.eventsMiddlewares.DoesEventExist(s.Fire))
}

func NewSubscriptionTransport(logger *zap.SugaredLogger,
	service Service,
	eventMiddlewares *middleware.EventsMiddlewares,
	eventsService events.Service,
	telegramService telegram.Service,
	templateProvider template.Provider,
	formatter formatter.Formatter,
	bot bot.Bot) Transport {

	return &subscriptionTransport{
		logger:              logger,
		subscriptionService: service,
		eventsMiddlewares:   eventMiddlewares,
		eventsService:       eventsService,
		telegramService:     telegramService,
		templateProvider:    templateProvider,
		bot:                 bot,
		formatter:           formatter,
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

	if len(subscribers) == 0 {
		httpRes.NoSubscribers(w)
		s.logger.Debugf("no subscribers for event %d", eventID)
		return
	}

	//todo: move to service
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

	templ, err := s.templateProvider.Find(eventID)
	if err != nil {
		httpErrors.MakeErrorResponse(w, err)
		s.logger.Error(err.Error())
		return
	}
	//templ will be formatted into fmtTempl passing ...args in below switch-case
	var fmtTempl string

	//Find payload type assigned to eventID
	//todo: move to service
	payloadType := payload.GetProvider().MustGetType(eventID)

	//todo: move to service
	switch payloadType {
	case reflect.TypeOf(payload.WorkerLoginPayload{}):
		var shouldBind payload.WorkerLoginPayload
		err := binder.Bind(r.Body, &shouldBind)
		if err != nil {
			httpErrors.MakeErrorResponse(w, err)
		}
		fmtTempl = s.formatter.Format(templ, shouldBind.Username, shouldBind.LoginAt)
	}

	for _, tgSub := range telegramIDs {
		//Notify subscribers in telegram here with fmtTempl text
		err := s.bot.Notify(tgSub.TelegramID, fmtTempl)
		if err != nil {
			//todo: if err occurs there bot must send warning msg to admin
			httpErrors.MakeErrorResponse(w, err)
			return
		}
	}

}

func (s *subscriptionTransport) Subscribe(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

}
