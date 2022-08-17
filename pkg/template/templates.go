package template

import (
	"fmt"

	"github.com/pkg/errors"
)

func newErrtemplateNotFound(eventID uint64) error {
	return errors.New(fmt.Sprintf("template for event %d not found", eventID))
}

type TemplateProvider interface {
	Find(eventID uint64) (string, error)
}

type templateProvider struct {
	store map[uint64]string
}

func (t *templateProvider) Find(eventID uint64) (string, error) {
	templ, ok := t.store[eventID]
	if ok != true {
		return "", newErrtemplateNotFound(eventID)
	}
	return templ, nil
}

func init() {

}
