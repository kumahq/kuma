package v1alpha1

import (
	"time"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/api/generic"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ generic.Insight = &ZoneIngressInsight{}

func (x *ZoneIngressInsight) GetSubscription(id string) (int, *DiscoverySubscription) {
	for i, s := range x.GetSubscriptions() {
		if s.Id == id {
			return i, s
		}
	}
	return -1, nil
}

func (x *ZoneIngressInsight) UpdateSubscription(s generic.Subscription) error {
	if x == nil {
		return nil
	}
	discoverySubscription, ok := s.(*DiscoverySubscription)
	if !ok {
		return errors.Errorf("invalid type %T for ZoneIngressInsight", s)
	}
	i, old := x.GetSubscription(discoverySubscription.Id)
	if old != nil {
		x.Subscriptions[i] = discoverySubscription
	} else {
		x.finalizeSubscriptions()
		x.Subscriptions = append(x.Subscriptions, discoverySubscription)
	}
	return nil
}

// If Kuma CP was killed ungracefully then we can get a subscription without a DisconnectTime.
// Because of the way we process subscriptions the lack of DisconnectTime on old subscription
// will cause wrong status.
func (x *ZoneIngressInsight) finalizeSubscriptions() {
	now := util_proto.Now()
	for _, subscription := range x.GetSubscriptions() {
		if subscription.DisconnectTime == nil {
			subscription.DisconnectTime = now
		}
	}
}

func (x *ZoneIngressInsight) IsOnline() bool {
	for _, s := range x.GetSubscriptions() {
		if s.ConnectTime != nil && s.DisconnectTime == nil {
			return true
		}
	}
	return false
}

func (x *ZoneIngressInsight) GetLastSubscription() generic.Subscription {
	if len(x.GetSubscriptions()) == 0 {
		return nil
	}
	return x.GetSubscriptions()[len(x.GetSubscriptions())-1]
}

// todo(lobkovilya): delete GetLatestSubscription, use GetLastSubscription instead
func (x *ZoneIngressInsight) GetLatestSubscription() (*DiscoverySubscription, *time.Time) {
	if len(x.GetSubscriptions()) == 0 {
		return nil, nil
	}
	var idx int = 0
	var latest *time.Time
	for i, s := range x.GetSubscriptions() {
		if err := s.ConnectTime.CheckValid(); err != nil {
			continue
		}
		t := s.ConnectTime.AsTime()
		if latest == nil || latest.Before(t) {
			idx = i
			latest = &t
		}
	}
	return x.Subscriptions[idx], latest
}

func (x *ZoneIngressInsight) Sum(v func(*DiscoverySubscription) uint64) uint64 {
	var result uint64 = 0
	for _, s := range x.GetSubscriptions() {
		result += v(s)
	}
	return result
}
