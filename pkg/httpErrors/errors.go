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

func NewErrEventDoesNotExist(eventName string) error {
	return errors.New(fmt.Sprintf("event with name %s does not exist", eventName))
}

func MakeErrorResponse(w http.ResponseWriter, err error) {
	switch true {
	case strings.Contains(err.Error(), "with name"):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case errors.Is(err, ErrNoEventId):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	default:
		http.Error(w, InternalError.Error(), http.StatusInternalServerError)
		return
	}
}
