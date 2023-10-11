package events

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/resources/model"
)

type Event interface{}

type Op int

const (
	Create Op = iota
	Update
	Delete
)

type ResourceChangedEvent struct {
	Operation Op
	Type      model.ResourceType
	Key       model.ResourceKey
	TenantID  string
}

type TriggerInsightsComputationEvent struct {
	TenantID string
}

var ListenerStoppedErr = errors.New("listener closed")

type Listener interface {
	Recv() <-chan Event
	Close()
}

func NewNeverListener() Listener {
	return &neverRecvListener{}
}

type neverRecvListener struct{}

func (*neverRecvListener) Recv() <-chan Event {
	return nil
}

func (*neverRecvListener) Close() {
}

type Predicate = func(event Event) bool

type Emitter interface {
	Send(Event)
}

type ListenerFactory interface {
	Subscribe(...Predicate) Listener
}

type EventBus interface {
	Emitter
	ListenerFactory
}
