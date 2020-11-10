package events

import (
	"sync"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/resources/model"
)

func NewEventBus() *EventBus {
	return &EventBus{}
}

type EventBus struct {
	mtx         sync.RWMutex
	subscribers []chan Event
}

func (b *EventBus) New() Reader {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	events := make(chan Event)
	b.subscribers = append(b.subscribers, events)
	return &reader{
		events: events,
	}
}

func (b *EventBus) Send(op Op, resourceType model.ResourceType, key model.ResourceKey) {
	b.mtx.RLock()
	defer b.mtx.RUnlock()

	for _, s := range b.subscribers {
		s <- Event{
			Operation: op,
			Type:      resourceType,
			Key:       key,
		}
	}
}

type reader struct {
	events chan Event
}

func (k *reader) Recv(stop <-chan struct{}) (Event, error) {
	select {
	case event, ok := <-k.events:
		if !ok {
			return Event{}, errors.New("end of events channel")
		}
		return event, nil
	case <-stop:
		return Event{}, errors.New("stop channel was closed")
	}
}
