package events

import (
	"sync"

	"github.com/kumahq/kuma/pkg/core"
)

func NewEventBus() EventBus {
	return &eventBus{
		subscribers: map[string]chan Event{},
	}
}

type eventBus struct {
	mtx         sync.RWMutex
	subscribers map[string]chan Event
}

func (b *eventBus) Subscribe() Listener {
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

func (b *eventBus) Send(event Event) {
	b.mtx.RLock()
	defer b.mtx.RUnlock()
	switch e := event.(type) {
	case ResourceChangedEvent:
		for _, channel := range b.subscribers {
			channel <- e
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
