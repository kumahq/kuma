package v1alpha1

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/api/generic"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ generic.Insight = &ZoneEgressInsight{}

func (x *ZoneEgressInsight) GetSubscription(id string) (int, *DiscoverySubscription) {
	for i, s := range x.GetSubscriptions() {
		if s.Id == id {
			return i, s
		}
	}
	return -1, nil
}

func (x *ZoneEgressInsight) UpdateSubscription(s generic.Subscription) error {
	if x == nil {
		return nil
	}
	discoverySubscription, ok := s.(*DiscoverySubscription)
	if !ok {
		return errors.Errorf("invalid type %T for ZoneEgressInsight", s)
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
func (x *ZoneEgressInsight) finalizeSubscriptions() {
	now := util_proto.Now()
	for _, subscription := range x.GetSubscriptions() {
		if subscription.DisconnectTime == nil {
			subscription.DisconnectTime = now
		}
	}
}

func (x *ZoneEgressInsight) IsOnline() bool {
	for _, s := range x.GetSubscriptions() {
		if s.ConnectTime != nil && s.DisconnectTime == nil {
			return true
		}
	}
	return false
}

func (x *ZoneEgressInsight) GetLastSubscription() generic.Subscription {
	if len(x.GetSubscriptions()) == 0 {
		return (*DiscoverySubscription)(nil)
	}
	return x.GetSubscriptions()[len(x.GetSubscriptions())-1]
}

func (x *ZoneEgressInsight) Sum(v func(*DiscoverySubscription) uint64) uint64 {
	var result uint64 = 0
	for _, s := range x.GetSubscriptions() {
		result += v(s)
	}
	return result
}
