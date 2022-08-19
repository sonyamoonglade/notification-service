package httpErrors

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var ErrNoEventId = errors.New("missing eventId in url string")
var ErrInvalidEventId = errors.New("invalid eventId format")
var InternalError = errors.New("internal error")
var MissingTemplateServiceUnavailable = errors.New("service is unavailable due to missing template")
var InvalidPayload = errors.New("invalid payload")
var SubscriberDoesNotExist = errors.New("subscriber does not exist")
var SubscriptionDoesNotExist = errors.New("subscription does not exist")
var SubscriptionAlreadyExists = errors.New("subscription already exists")

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
		http.Error(w, MissingTemplateServiceUnavailable.Error(), http.StatusServiceUnavailable)
		return
	case strings.Contains(err.Error(), "Validation"):
		http.Error(w, InvalidPayload.Error(), http.StatusBadRequest)
		return
	case strings.Contains(err.Error(), "already exists"):
		http.Error(w, err.Error(), http.StatusConflict)
		return
	case strings.Contains(err.Error(), "does not exist"):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	default:
		http.Error(w, InternalError.Error(), http.StatusInternalServerError)
		return
	}
}
