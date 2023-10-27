package generic

import (
	"time"

	"google.golang.org/protobuf/proto"
)

func AllSubscriptions[S Subscription, T interface{ GetSubscriptions() []S }](t T) []Subscription {
	var subs []Subscription
	for _, s := range t.GetSubscriptions() {
		subs = append(subs, s)
	}
	return subs
}

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
