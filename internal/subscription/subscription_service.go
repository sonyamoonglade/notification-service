package subscription

import (
	"context"

	"github.com/sonyamoonglade/notification-service/internal/entity"
	"github.com/sonyamoonglade/notification-service/internal/storage"
	"github.com/sonyamoonglade/notification-service/internal/subscription/response_object"
	"github.com/sonyamoonglade/notification-service/pkg/http_errors"
	"github.com/sonyamoonglade/notification-service/pkg/telegram_errors"
	"go.uber.org/zap"
)

type Service interface {
	GetSubscribersDataJoined(ctx context.Context) ([]*response_object.SubscriberRO, error)
	GetEventSubscribers(ctx context.Context, eventID uint64) ([]*entity.Subscriber, error)
	GetSubscriberByPhone(ctx context.Context, phoneNumber string) (*entity.Subscriber, error)
	GetTelegramSubscribers(ctx context.Context, phoneNumbers []string) ([]*entity.TelegramSubscriber, error)
	GetSubscription(ctx context.Context, subscriberID uint64, eventID uint64) (*entity.Subscription, error)
	GetTelegramSubscriber(ctx context.Context, phoneNumber string) (*entity.TelegramSubscriber, error)
	GetSubscribersWithoutSubs(ctx context.Context) ([]*response_object.SubscriberRO, error)
	RegisterSubscriber(ctx context.Context, phoneNumber string) (uint64, error)
	RegisterTelegramSubscriber(ctx context.Context, telegramID int64, subscriberID uint64) error
	SubscribeToEvent(ctx context.Context, subscriberID uint64, eventID uint64) error
	SelectPhones(subs []*entity.Subscriber) []string
	CancelSubscription(ctx context.Context, subscriptionID uint64) error
}

type subscriptionService struct {
	storage storage.DBStorage
	logger  *zap.SugaredLogger
}

func NewSubscriptionService(logger *zap.SugaredLogger, storage storage.DBStorage) Service {
	return &subscriptionService{logger: logger, storage: storage}
}

func (s *subscriptionService) GetSubscribersWithoutSubs(ctx context.Context) ([]*response_object.SubscriberRO, error) {
	return s.storage.GetSubscribersWithoutSubs(ctx)
}

func (s *subscriptionService) GetSubscribersDataJoined(ctx context.Context) ([]*response_object.SubscriberRO, error) {
	return s.storage.GetSubscribersDataJoined(ctx)
}

func (s *subscriptionService) SubscribeToEvent(ctx context.Context, subscriberID uint64, eventID uint64) error {
	subscriptionID, err := s.storage.SubscribeToEvent(ctx, subscriberID, eventID)
	if err != nil {
		return err
	}
	s.logger.Debugf("%d", subscriptionID)
	if subscriptionID == 0 {
		return http_errors.ErrSubscriptionAlreadyExists
	}
	return nil
}

func (s *subscriptionService) GetSubscription(ctx context.Context, subscriberID uint64, eventID uint64) (*entity.Subscription, error) {
	subscription, err := s.storage.GetSubscription(ctx, subscriberID, eventID)
	if err != nil {
		return nil, err
	}
	//No such subscription
	if subscription == nil {
		return nil, http_errors.ErrSubscriptionDoesNotExist
	}

	return subscription, nil
}

func (s *subscriptionService) GetEventSubscribers(ctx context.Context, eventID uint64) ([]*entity.Subscriber, error) {
	subs, err := s.storage.GetEventSubscribers(ctx, eventID)
	if err != nil {
		return nil, err
	}
	return subs, nil
}

func (s *subscriptionService) GetSubscriberByPhone(ctx context.Context, phoneNumber string) (*entity.Subscriber, error) {
	sub, err := s.storage.GetSubscriberByPhone(ctx, phoneNumber)
	if err != nil {
		return nil, err
	}
	//No such subscriber
	if sub == nil {
		return nil, http_errors.ErrSubscriberDoesNotExist
	}
	return sub, nil
}

func (s *subscriptionService) RegisterSubscriber(ctx context.Context, phoneNumber string) (uint64, error) {
	return s.storage.RegisterSubscriber(ctx, phoneNumber)
}

func (s *subscriptionService) SelectPhones(subs []*entity.Subscriber) []string {
	var ph []string
	for _, sub := range subs {
		ph = append(ph, sub.PhoneNumber)
	}
	return ph
}

func (s *subscriptionService) RegisterTelegramSubscriber(ctx context.Context, telegramID int64, subscriberID uint64) error {
	ok, err := s.storage.RegisterTelegramSubscriber(ctx, telegramID, subscriberID)
	if err != nil {
		return err
	}
	if ok != true {
		return telegram_errors.ErrTgSubscriberAlreadyExists
	}

	return nil
}

func (s *subscriptionService) GetTelegramSubscriber(ctx context.Context, phoneNumber string) (*entity.TelegramSubscriber, error) {
	tgsub, err := s.storage.GetTelegramSubscriber(ctx, phoneNumber)
	if err != nil {
		return nil, err
	}
	if tgsub == nil {
		return nil, telegram_errors.ErrNoSuchTelegramSubscriber
	}

	return tgsub, nil
}

func (s *subscriptionService) GetTelegramSubscribers(ctx context.Context, phoneNumbers []string) ([]*entity.TelegramSubscriber, error) {
	return s.storage.GetTelegramSubscribers(ctx, phoneNumbers)
}

func (s *subscriptionService) CancelSubscription(ctx context.Context, subscriptionID uint64) error {
	ok, err := s.storage.CancelSubscription(ctx, subscriptionID)
	if err != nil {
		return err
	}
	if ok != true {
		return http_errors.ErrSubscriptionDoesNotExist
	}
	return nil
}
