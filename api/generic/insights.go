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
	SetDisconnectTime(time time.Time)
}
