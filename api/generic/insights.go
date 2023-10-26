package generic

import (
	"time"

	"google.golang.org/protobuf/proto"
)

type Insight interface {
	proto.Message
	IsOnline() bool
	GetSubscription(id string) Subscription
	AllSubscriptions() []Subscription
	UpdateSubscription(Subscription) error
}

type Subscription interface {
	proto.Message
	GetId() string
	GetGeneration() uint32
	IsOnline() bool
	SetDisconnectTime(time time.Time)
}
