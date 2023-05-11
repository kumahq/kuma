package events

import (
	"sync"
	"time"

	store_config "github.com/kumahq/kuma/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/pkg/core"
)

var log = core.Log.WithName("event-bus")

func NewEventBus(cfg store_config.EventBusConfig) *EventBus {
	return &EventBus{
		subscribers: map[string]chan Event{},
		sendTimeout: cfg.SendTimeout.Duration,
	}
}

type EventBus struct {
	mtx         sync.RWMutex
	sendTimeout time.Duration
	subscribers map[string]chan Event
}

func (b *EventBus) Subscribe() Listener {
	id := core.NewUUID()
	b.mtx.Lock()
	defer b.mtx.Unlock()

	events := make(chan Event, 10)
	b.subscribers[id] = events
	return &reader{
		events: events,
		close: func() {
			b.mtx.Lock()
			defer b.mtx.Unlock()
			delete(b.subscribers, id)
		},
	}
}

func (b *EventBus) Send(event Event) {
	b.mtx.RLock()
	defer b.mtx.RUnlock()
	switch e := event.(type) {
	case ResourceChangedEvent:
		for id, channel := range b.subscribers {
			select {
			case channel <- ResourceChangedEvent{
				Operation: e.Operation,
				Type:      e.Type,
				Key:       e.Key,
			}:
			case <-time.After(b.sendTimeout):
				log.V(1).Info("timeout occurred while sending event", "subscriber", id)
			}
		}
	}
}

type reader struct {
	events chan Event
	close  func()
}

func (k *reader) Recv() <-chan Event {
	return k.events
}

func (k *reader) Close() {
	k.close()
}
