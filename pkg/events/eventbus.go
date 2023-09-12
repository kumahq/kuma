package events

import (
	"sync"

	"github.com/kumahq/kuma/pkg/core"
)

var log = core.Log.WithName("eventbus")

type subscriber struct {
	ch         chan Event
	predicates []Predicate
}

func NewEventBus(bufferSize uint) EventBus {
	return &eventBus{
		subscribers: map[string]subscriber{},
		bufferSize:  bufferSize,
	}
}

type eventBus struct {
	mtx         sync.RWMutex
	subscribers map[string]subscriber
	bufferSize  uint
}

// Subscribe subscribes to a stream of events given Predicates
// Predicate should not block on I/O, otherwise the whole event bus can block.
func (b *eventBus) Subscribe(predicates ...Predicate) Listener {
	id := core.NewUUID()
	b.mtx.Lock()
	defer b.mtx.Unlock()

	events := make(chan Event, b.bufferSize)
	b.subscribers[id] = subscriber{
		ch:         events,
		predicates: predicates,
	}
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
	for _, sub := range b.subscribers {
		matched := true
		for _, predicate := range sub.predicates {
			if !predicate(event) {
				matched = false
			}
		}
		if matched {
			select {
			case sub.ch <- event:
			default:
				log.Info("event is not sent because the channel is full. Ignoring event. Consider increasing buffer size using KUMA_EVENT_BUS_BUFFER_SIZE",
					"bufferSize", b.bufferSize,
					"event", event,
				)
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
