package subscription

import (
	"net/http"
	"reflect"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"github.com/sonyamoonglade/delivery-service/pkg/binder"
	"github.com/sonyamoonglade/delivery-service/pkg/validation"
	"github.com/sonyamoonglade/notification-service/internal/entity"
	"github.com/sonyamoonglade/notification-service/internal/events"
	"github.com/sonyamoonglade/notification-service/internal/events/middlewares"
	"github.com/sonyamoonglade/notification-service/internal/events/payload"
	"github.com/sonyamoonglade/notification-service/internal/subscription/dto"
	"github.com/sonyamoonglade/notification-service/pkg/bot"
	"github.com/sonyamoonglade/notification-service/pkg/formatter"
	"github.com/sonyamoonglade/notification-service/pkg/http_errors"
	"github.com/sonyamoonglade/notification-service/pkg/response"
	"github.com/sonyamoonglade/notification-service/pkg/template"
	"go.uber.org/zap"
)

type Transport interface {
	Fire(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	Subscribe(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	Cancel(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	RegisterSubscriber(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	GetSubscribersJoined(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	GetAvailableEvents(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	GetSubscribersWithoutSubs(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	InitRoutes(router *httprouter.Router)
}

type subscriptionTransport struct {
	subscriptionService Service
	eventsService       events.Service
	templateProvider    template.Provider
	formatter           formatter.Formatter
	de                  *event_middlewares.DoesExist
	logger              *zap.SugaredLogger
	bot                 bot.Bot
}

func (s *subscriptionTransport) InitRoutes(router *httprouter.Router) {
	router.POST("/api/events/fire/:eventName", s.de.Check(s.Fire))
	router.GET("/api/events", s.GetAvailableEvents)
	router.POST("/api/subscriptions", s.Subscribe)
	router.DELETE("/api/subscriptions/:subscriptionId", s.Cancel)
	router.GET("/api/subscriptions/subscribers/joined", s.GetSubscribersJoined)
	router.GET("/api/subscriptions/subscribers", s.GetSubscribersWithoutSubs)
	router.POST("/api/subscriptions/subscribers", s.RegisterSubscriber)
}

func NewSubscriptionTransport(logger *zap.SugaredLogger,
	service Service,
	de *event_middlewares.DoesExist,
	eventsService events.Service,
	templateProvider template.Provider,
	formatter formatter.Formatter,
	bot bot.Bot) Transport {

	return &subscriptionTransport{
		logger:              logger,
		subscriptionService: service,
		de:                  de,
		eventsService:       eventsService,
		templateProvider:    templateProvider,
		bot:                 bot,
		formatter:           formatter,
	}
}

func (s *subscriptionTransport) RegisterSubscriber(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	s.logger.Debug("register subscriber")

	var inp dto.RegisterSubscriberDto

	err := binder.Bind(r.Body, &inp)
	if err != nil {
		http_errors.MakeErrorResponse(w, err)
		s.logger.Error(err.Error())
		return
	}

	ok := validation.ValidatePhoneNumber(inp.PhoneNumber)
	if ok != true {
		http_errors.MakeErrorResponse(w, http_errors.ErrInvalidPayload)
		s.logger.Debug("invalid phone number")
		return
	}

	_, err = s.subscriptionService.RegisterSubscriber(r.Context(), inp.PhoneNumber)
	if err != nil {
		http_errors.MakeErrorResponse(w, err)
		s.logger.Error(err.Error())
		return
	}

	response.Created(w)
}

func (s *subscriptionTransport) GetSubscribersWithoutSubs(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	s.logger.Debug("get subscribers without subs")

	subscribers, err := s.subscriptionService.GetSubscribersWithoutSubs(r.Context())
	if err != nil {
		http_errors.MakeErrorResponse(w, err)
		s.logger.Error(err.Error())
		return
	}

	response.Json(s.logger, w, http.StatusOK, response.JSON{
		"subscribers": subscribers,
	})
	return
}

func (s *subscriptionTransport) GetAvailableEvents(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	s.logger.Debug("get available events")

	evnts, err := s.eventsService.GetAvailableEvents(r.Context())
	if err != nil {
		http_errors.MakeErrorResponse(w, err)
		s.logger.Error(err.Error())
		return
	}

	response.Json(s.logger, w, http.StatusOK, response.JSON{
		"events": evnts,
	})

}

func (s *subscriptionTransport) Fire(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := r.Context()
	eventID := ctx.Value("eventId").(uint64)

	subscribers, err := s.subscriptionService.GetEventSubscribers(ctx, eventID)
	if err != nil {
		http_errors.MakeErrorResponse(w, err)
		s.logger.Error(err.Error())
		return
	}
	//No actual subscribers whatsoever, so the rest of the code is a waste
	if len(subscribers) == 0 {
		response.NoContent(w)
		return
	}

	subsPhones := s.subscriptionService.SelectPhones(subscribers)
	telegramSubs, err := s.subscriptionService.GetTelegramSubscribers(ctx, subsPhones)
	if err != nil {
		http_errors.MakeErrorResponse(w, err)
		s.logger.Error(err.Error())
		return
	}

	//No actual subscribers in telegram, so the rest of the code is a waste
	if len(telegramSubs) == 0 {
		response.NoContent(w)
		return
	}

	tmpl, err := s.templateProvider.Find(eventID)
	if err != nil {
		http_errors.MakeErrorResponse(w, err)
		s.logger.Error(err.Error())
		return
	}
	//tmpl will be formatted into fmtTmpl passing ...args in the switch-case below
	var fmtTmpl string

	//Find payload type assigned to eventID
	payloadType := payload.GetProvider().MustGetType(eventID)

	//Iterate over payload types, assigned to events in payload package
	//Format the template, assigned to event in templates.json
	//Keep in mind, formatter.Format func is Variadic
	switch payloadType {
	case reflect.TypeOf(payload.WorkerLoginPayload{}):
		var p payload.WorkerLoginPayload
		err := binder.Bind(r.Body, &p)
		if err != nil {
			s.logger.Error(err.Error())
			http_errors.MakeErrorResponse(w, err)
			return
		}
		fmtTmpl = s.formatter.Format(
			tmpl,
			p.Username,
			s.formatter.FormatTime(p.LoginAt, p.TimeOffset))

		break
	case reflect.TypeOf(payload.UserOrderCreatePayload{}):
		var p payload.UserOrderCreatePayload
		err := binder.Bind(r.Body, &p)
		if err != nil {
			s.logger.Error(err.Error())
			http_errors.MakeErrorResponse(w, err)
			return
		}
		fmtTmpl = s.formatter.Format(tmpl,
			p.OrderID,
			p.Amount)

		break
	case reflect.TypeOf(payload.MasterOrderCreatePayload{}):
		var p payload.MasterOrderCreatePayload
		err := binder.Bind(r.Body, &p)
		if err != nil {
			s.logger.Error(err.Error())
			http_errors.MakeErrorResponse(w, err)
			return
		}

		ok := validation.ValidatePhoneNumber(p.PhoneNumber)
		if ok != true {
			http_errors.MakeErrorResponse(w, http_errors.ErrInvalidPayload)
			return
		}
		fmtTmpl = s.formatter.Format(tmpl,
			p.OrderID,
			p.Username,
			p.PhoneNumber,
			p.Amount)

		break
	}

	//Range over subscriber's associated telegram id's
	for _, sub := range telegramSubs {
		//Notify subscribers in telegram here with fmtTmpl text
		err := s.bot.Notify(sub.TelegramID, fmtTmpl)
		if err != nil {
			//todo: if err occurs there bot must send warning msg to admin
			http_errors.MakeErrorResponse(w, err)
			s.logger.Error(err.Error())
			return
		}
	}

	response.Ok(w)
	return
}

func (s *subscriptionTransport) Subscribe(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	var inp dto.SubscribeToEventInp
	ctx := r.Context()

	err := binder.Bind(r.Body, &inp)
	if err != nil {
		http_errors.MakeErrorResponse(w, err)
		s.logger.Error(err.Error())
		return
	}

	ok := validation.ValidatePhoneNumber(inp.PhoneNumber)
	if ok != true {
		err = http_errors.ErrInvalidPayload
		http_errors.MakeErrorResponse(w, err)
		s.logger.Debug(err.Error())
		return
	}

	eventID, err := s.eventsService.DoesExist(ctx, inp.EventName)
	if err != nil {
		http_errors.MakeErrorResponse(w, err)
		s.logger.Error(err.Error())
		return
	}

	subscriber, err := s.subscriptionService.GetSubscriberByPhone(ctx, inp.PhoneNumber)
	if err != nil {
		//If any internal error not SubscriberDoesNotExist
		if !errors.Is(err, http_errors.ErrSubscriberDoesNotExist) {
			http_errors.MakeErrorResponse(w, err)
			s.logger.Error(err.Error())
			return
		}
		//Register subscriber
		regSubID, err := s.subscriptionService.RegisterSubscriber(ctx, inp.PhoneNumber)
		if err != nil {
			http_errors.MakeErrorResponse(w, err)
			s.logger.Error(err.Error())
			return
		}
		s.logger.Debug("registered subscriber")
		//Assign newly registered regSubID and phoneNumber to subscriber
		regSub := entity.Subscriber{
			SubscriberID: regSubID,
			PhoneNumber:  inp.PhoneNumber,
		}
		subscriber = &regSub
	}

	//Create subscription
	err = s.subscriptionService.SubscribeToEvent(ctx, subscriber.SubscriberID, eventID)
	if err != nil {
		http_errors.MakeErrorResponse(w, err)
		s.logger.Error(err.Error())
		return
	}
	s.logger.Debugf("subscriber with phone %s has subscribed to event %d", inp.PhoneNumber, eventID)

	response.Created(w)
	return
}

func (s *subscriptionTransport) Cancel(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

	subscriptionIDstr := params.ByName("subscriptionId")

	if subscriptionIDstr == "" {
		http_errors.MakeErrorResponse(w, http_errors.ErrNoSubscriptionID)
		s.logger.Debug("missing subscription id")
		return
	}

	subscriptionID, err := strconv.ParseUint(subscriptionIDstr, 10, 64)
	if err != nil {
		http_errors.MakeErrorResponse(w, err)
		s.logger.Error(err.Error())
		return
	}

	err = s.subscriptionService.CancelSubscription(r.Context(), subscriptionID)
	if err != nil {
		http_errors.MakeErrorResponse(w, err)
		s.logger.Error(err.Error())
		return
	}

	response.Ok(w)
	return
}

func (s *subscriptionTransport) GetSubscribersJoined(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	s.logger.Debug("get all subscribers joined")

	subscribersData, err := s.subscriptionService.GetSubscribersDataJoined(r.Context())
	if err != nil {
		http_errors.MakeErrorResponse(w, err)
		s.logger.Error(err.Error())
		return
	}

	response.Json(s.logger, w, http.StatusOK, response.JSON{
		"subscribers": subscribersData,
	})
	return
}
