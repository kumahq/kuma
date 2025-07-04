package v1alpha1

import (
	"time"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/kumahq/kuma/api/generic"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ generic.Insight = &ZoneInsight{}

func NewSubscriptionStatus(now time.Time) *KDSSubscriptionStatus {
	return &KDSSubscriptionStatus{
		LastUpdateTime: util_proto.MustTimestampProto(now),
		Total:          &KDSServiceStats{},
		Stat:           map[string]*KDSServiceStats{},
	}
}

func (x *ZoneInsight) GetSubscription(id string) generic.Subscription {
	return generic.GetSubscription[*KDSSubscription](x, id)
}

func (x *ZoneInsight) GetLastSubscription() generic.Subscription {
	if len(x.GetSubscriptions()) == 0 {
		return (*KDSSubscription)(nil)
	}
	return x.GetSubscriptions()[len(x.GetSubscriptions())-1]
}

func (x *ZoneInsight) IsOnline() bool {
	for _, s := range x.GetSubscriptions() {
		if s.ConnectTime != nil && s.DisconnectTime == nil {
			return true
		}
	}
	return false
}

func (x *ZoneInsight) GetKDSStream(streamType string) *KDSStream {
	switch streamType {
	case "globalToZone":
		return x.GetKdsStreams().GetGlobalToZone()
	case "zoneToGlobal":
		return x.GetKdsStreams().GetZoneToGlobal()
	case "clusters":
		return x.GetKdsStreams().GetClusters()
	case "stats":
		return x.GetKdsStreams().GetStats()
	case "configDump":
		return x.GetKdsStreams().GetConfigDump()
	}
	return nil
}

func (x *ZoneInsight) AllSubscriptions() []generic.Subscription {
	return generic.AllSubscriptions[*KDSSubscription](x)
}

func (x *KDSSubscription) SetDisconnectTime(time time.Time) {
	x.DisconnectTime = timestamppb.New(time)
}

func (x *KDSSubscription) IsOnline() bool {
	return x.GetConnectTime() != nil && x.GetDisconnectTime() == nil
}

func (x *ZoneInsight) Sum(v func(*KDSSubscription) uint64) uint64 {
	var result uint64
	for _, s := range x.GetSubscriptions() {
		result += v(s)
	}
	return result
}

func (x *ZoneInsight) UpdateSubscription(s generic.Subscription) error {
	if x == nil {
		return nil
	}
	kdsSubscription, ok := s.(*KDSSubscription)
	if !ok {
		return errors.Errorf("invalid type %T for ZoneInsight", s)
	}
	for i, sub := range x.GetSubscriptions() {
		if sub.GetId() == kdsSubscription.Id {
			x.Subscriptions[i] = kdsSubscription
			return nil
		}
	}
	x.finalizeSubscriptions()
	x.Subscriptions = append(x.Subscriptions, kdsSubscription)
	return nil
}

// CompactFinished removes detailed information about finished subscriptions to trim the object size
// The last subscription always has details.
func (x *ZoneInsight) CompactFinished() {
	for i := 0; i < len(x.GetSubscriptions())-1; i++ {
		x.Subscriptions[i].Config = ""
		if status := x.Subscriptions[i].Status; status != nil {
			status.Stat = map[string]*KDSServiceStats{}
		}
	}
}

// If Global CP was killed ungracefully then we can get a subscription without a DisconnectTime.
// Because of the way we process subscriptions the lack of DisconnectTime on old subscription
// will cause wrong status.
func (x *ZoneInsight) finalizeSubscriptions() {
	now := util_proto.Now()
	for _, subscription := range x.GetSubscriptions() {
		if subscription.DisconnectTime == nil {
			subscription.DisconnectTime = now
		}
	}
}

func NewVersion() *Version {
	return &Version{
		KumaCp: &KumaCpVersion{
			Version:   "",
			GitTag:    "",
			GitCommit: "",
			BuildDate: "",
		},
	}
}
