package events

import (
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
	Recv(stop <-chan struct{}) (Event, error)
}

type Writer interface {
	Send(Op, model.ResourceType, model.ResourceKey)
}

type ReaderFactory interface {
	New() Reader
}
