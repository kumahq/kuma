package events

import (
	"context"
	"sync"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/resources/model"
)

type Op int

const (
	Create Op = iota
	Update
	Delete
)

type Event struct {
	Operation Op
	Type      model.ResourceType
	Key       model.ResourceKey
}

type Reader interface {
	Recv(context.Context) (Event, error)
}

type Writer interface {
	Send(Op, model.ResourceType, model.ResourceKey)
}

type ReaderFactory interface {
	New() Reader
}

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

func (k *reader) Recv(ctx context.Context) (Event, error) {
	select {
	case event, ok := <-k.events:
		if !ok {
			return Event{}, errors.New("end of events channel")
		}
		return event, nil
	case <-ctx.Done():
		return Event{}, ctx.Err()
	}
}
