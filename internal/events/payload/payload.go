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
	fmt.Println(provider.store)
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

/*
	OrderCreatedPayload
	Payload for event with id 1
	See events.json
*/
type OrderCreatedPayload struct {
	OrderID        int64  `json:"order_id"`
	TotalCartPrice int64  `json:"total_cart_price"`
	Username       string `json:"username"`
	PhoneNumber    string `json:"phone_number"`
}

/*
	WorkerLoginPayload
	Payload for event with id 2
	See events.json
*/
type WorkerLoginPayload struct {
	Username string    `json:"username"`
	LoginAt  time.Time `json:"login_at"`
}
