package events

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/kumahq/kuma/pkg/core"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
)

var log = core.Log.WithName("eventbus")

type subscriber struct {
	ch         chan Event
	predicates []Predicate
}

func NewEventBus(bufferSize uint, metrics core_metrics.Metrics) (EventBus, error) {
	metric := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "events_dropped",
		Help: "Number of dropped events in event bus due to full channels",
	})
	if err := metrics.Register(metric); err != nil {
		return nil, err
	}
	return &eventBus{
		subscribers: map[string]subscriber{},
		bufferSize:  bufferSize,
		metric:      metric,
	}, nil
}

type eventBus struct {
	mtx         sync.RWMutex
	subscribers map[string]subscriber
	bufferSize  uint
	metric      prometheus.Counter
}

// Subscribe subscribes to a stream of events given Predicates
// Predicate should not block on I/O, otherwise the whole event bus can block.
// All predicates must pass for the event to enqueued.
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
				b.metric.Inc()
				log.Info("[WARNING] event is not sent because the channel is full. Ignoring event. Consider increasing buffer size using KUMA_EVENT_BUS_BUFFER_SIZE",
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
