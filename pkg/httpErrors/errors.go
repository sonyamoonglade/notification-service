package httpErrors

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var ErrNoEventId = errors.New("missing eventId in url string")
var ErrInvalidEventId = errors.New("invalid eventId format")
var ErrInternalError = errors.New("internal error")
var ErrMissingTemplateServiceUnavailable = errors.New("service is unavailable due to missing template")
var ErrInvalidPayload = errors.New("invalid request payload")
var ErrSubscriberDoesNotExist = errors.New("subscriber does not exist")
var ErrSubscriptionDoesNotExist = errors.New("subscription does not exist")
var ErrSubscriptionAlreadyExists = errors.New("subscription already exists")
var ErrNoSubscriptions = errors.New("no subscriptions")
var ErrNoTelegramSubscribers = errors.New("no telegram subscribers")

func NewErrEventDoesNotExist(eventID uint64) error {
	return errors.New(fmt.Sprintf("event with id %d does not exist", eventID))
}

func MakeErrorResponse(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "text/plain")
	switch true {
	case strings.Contains(err.Error(), "with name"):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, ErrNoEventId):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case strings.Contains(err.Error(), "template for event"):
		http.Error(w, ErrMissingTemplateServiceUnavailable.Error(), http.StatusServiceUnavailable)
		return
	case strings.Contains(err.Error(), "Bad request"):
		http.Error(w, ErrInvalidPayload.Error(), http.StatusBadRequest)
		return
	case strings.Contains(err.Error(), "Validation"):
		http.Error(w, ErrInvalidPayload.Error(), http.StatusBadRequest)
		return
	case strings.Contains(err.Error(), "already exists"):
		http.Error(w, err.Error(), http.StatusConflict)
		return
	case strings.Contains(err.Error(), "subscriber does not exist"):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case strings.Contains(err.Error(), "subscription does not exist"):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case strings.Contains(err.Error(), "no subscriptions"):
		http.Error(w, "", http.StatusNoContent)
		return
	case strings.Contains(err.Error(), "no telegram subscribers"):
		http.Error(w, "", http.StatusNoContent)
		return
	case strings.Contains(err.Error(), "invalid request payload"):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	default:
		http.Error(w, ErrInternalError.Error(), http.StatusInternalServerError)
		return
	}
}
