package template

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/sonyamoonglade/notification-service/internal/entity"
)

var path = "./templates.json"

type Provider interface {
	Find(eventID uint64) (string, error)
	ReadTemplates() error
}

type templateProvider struct {
	store map[uint64]string
}

func NewTemplateProvider() Provider {
	return &templateProvider{store: make(map[uint64]string)}
}

func (t *templateProvider) Find(eventID uint64) (string, error) {
	templ, ok := t.store[eventID]
	if ok != true {
		return "", fmt.Errorf("template for event %d not found", eventID)
	}
	return templ, nil
}

func (t *templateProvider) ReadTemplates() error {
	_, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("file %s does not exist", path)
		}
		return err
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)

	var result entity.Templates

	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return err
	}

	for _, tmpl := range result.Templates {
		t.store[tmpl.EventID] = tmpl.Text
	}

	return nil
}
