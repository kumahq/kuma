package events

import (
	"sync"

	"github.com/pkg/errors"
)

func NewEventBus() *EventBus {
	return &EventBus{}
}

type EventBus struct {
	mtx         sync.RWMutex
	subscribers []chan Event
}

func (b *EventBus) New() Listener {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	events := make(chan Event, 10)
	b.subscribers = append(b.subscribers, events)
	return &reader{
		events: events,
	}
}

func (b *EventBus) Send(event Event) {
	b.mtx.RLock()
	defer b.mtx.RUnlock()

	switch e := event.(type) {
	case ResourceChangedEvent:
		for _, s := range b.subscribers {
			s <- ResourceChangedEvent{
				Operation: e.Operation,
				Type:      e.Type,
				Key:       e.Key,
			}
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
			return nil, errors.New("end of events channel")
		}
		return event, nil
	case <-stop:
		return nil, ListenerStoppedErr
	}
}
