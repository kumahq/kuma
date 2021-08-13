package generic

import (
	"time"

	"github.com/golang/protobuf/proto"
)

type Insight interface {
	proto.Message
	IsOnline() bool
	GetLastSubscription() Subscription
	UpdateSubscription(Subscription) error
}

type Subscription interface {
	proto.Message
	GetId() string
	GetGeneration() uint32
	SetDisconnectTime(time time.Time)
}
