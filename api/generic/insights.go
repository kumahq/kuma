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

func GetSubscription[S Subscription, T interface{ GetSubscriptions() []S }](t T, id string) Subscription {
	for _, s := range t.GetSubscriptions() {
		if s.GetId() == id {
			return s
		}
	}
	return nil
}

type Insight interface {
	proto.Message
	IsOnline() bool
	// GetLastSubscription returns that last subscription or nil if there are no subscriptions
	GetLastSubscription() Subscription
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
