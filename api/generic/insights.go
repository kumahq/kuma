package generic

import (
	"time"

	"google.golang.org/protobuf/proto"
)

type Insight interface {
	proto.Message
	IsOnline() bool
	// TODO Deprecated: bad
	GetLastSubscription() Subscription
	GetSubscription(id string) Subscription
	GetOnlineSubscriptions() []Subscription
	UpdateSubscription(Subscription) error
}

type Subscription interface {
	proto.Message
	GetId() string
	GetGeneration() uint32
	SetDisconnectTime(time time.Time)
}
