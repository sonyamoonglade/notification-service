package subscription

import (
	"context"

	"github.com/sonyamoonglade/notification-service/internal/entity"
	"github.com/sonyamoonglade/notification-service/pkg/httpErrors"
	"go.uber.org/zap"
)

type Service interface {
	GetEventSubscribers(ctx context.Context, eventID uint64) ([]*entity.Subscriber, error)
	GetSubscriberByPhone(ctx context.Context, phoneNumber string) (*entity.Subscriber, error)
	RegisterSubscriber(ctx context.Context, phoneNumber string) (uint64, error)
	GetSubscription(ctx context.Context, subscriberID uint64, eventID uint64) (*entity.Subscription, error)
	SubscribeToEvent(ctx context.Context, subscriberID uint64, eventID uint64) error
	SelectPhones(subs []*entity.Subscriber) []string
}

type subscriptionService struct {
	storage Storage
	logger  *zap.SugaredLogger
}

func NewSubscriptionService(logger *zap.SugaredLogger, storage Storage) Service {
	return &subscriptionService{logger: logger, storage: storage}
}

func (s *subscriptionService) SubscribeToEvent(ctx context.Context, subscriberID uint64, eventID uint64) error {
	subscriptionID, err := s.storage.SubscribeToEvent(ctx, subscriberID, eventID)
	if err != nil {
		return err
	}
	s.logger.Debugf("%d", subscriptionID)
	if subscriptionID == 0 {
		return httpErrors.ErrSubscriptionAlreadyExists
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
		return nil, httpErrors.ErrSubscriptionDoesNotExist
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
		return nil, httpErrors.ErrSubscriberDoesNotExist
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
