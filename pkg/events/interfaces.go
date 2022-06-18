package events

import (
	"errors"

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
}

var ListenerStoppedErr = errors.New("listener closed")

type Listener interface {
	Recv(stop <-chan struct{}) (Event, error)
}

type Emitter interface {
	Send(Event)
}

type ListenerFactory interface {
	New() Listener
}
