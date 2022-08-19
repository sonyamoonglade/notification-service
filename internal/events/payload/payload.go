package payload

import (
	"fmt"
	"reflect"
	"time"
)

var provider *Provider

type Provider struct {
	store map[uint64]reflect.Type
}

/*
	init function should assign each event, listed in events.json
	unique payload type struct from this file.
	Each eventId should have its own payload type
*/
func init() {
	provider = &Provider{store: make(map[uint64]reflect.Type)}
	provider.Register(1, OrderCreatedPayload{})
	provider.Register(2, WorkerLoginPayload{})
}

func GetProvider() *Provider {
	return provider
}

func (p *Provider) Register(eventID uint64, t interface{}) {
	p.store[eventID] = reflect.TypeOf(t)
}

func (p *Provider) GetType(eventID uint64) (interface{}, error) {
	v, ok := p.store[eventID]
	if ok != true {
		return nil, fmt.Errorf("no such type registered on event %d", eventID)
	}
	return v, nil
}

func (p *Provider) MustGetType(eventID uint64) reflect.Type {
	v, _ := p.store[eventID]
	return v
}

/*
	OrderCreatedPayload
	Payload for event with id 1
	See events.json
*/
type OrderCreatedPayload struct {
	OrderID        int64  `json:"order_id" validate:"required"`
	TotalCartPrice int64  `json:"total_cart_price" validate:"required"`
	Username       string `json:"username" validate:"required"`
	PhoneNumber    string `json:"phone_number" validate:"required"`
}

/*
	WorkerLoginPayload
	Payload for event with id 2
	See events.json
*/
type WorkerLoginPayload struct {
	Username   string    `json:"username" validate:"required"`
	LoginAt    time.Time `json:"login_at" validate:"required"`
	TimeOffset int       `json:"time_offset" validate:"required"`
}
