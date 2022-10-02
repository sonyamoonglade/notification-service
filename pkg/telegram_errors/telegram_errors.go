package telegram_errors

import "github.com/pkg/errors"

var ErrNoSuchTelegramSubscriber = errors.New("no such telegram subscriber")
var ErrTgSubscriberAlreadyExists = errors.New("telegram subscriber already exists")
