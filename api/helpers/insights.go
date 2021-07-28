package helpers

import (
	"time"

	"github.com/golang/protobuf/proto"
)

type Insight interface {
	proto.Message
	IsOnline() bool
	GetLastSubscription() Subscription
	UpdateSubscription(Subscription)
}

type Subscription interface {
	proto.Message
	GetCandidateForDisconnect() bool
	SetCandidateForDisconnect(bool)
	SetDisconnectTime(time time.Time)
}
