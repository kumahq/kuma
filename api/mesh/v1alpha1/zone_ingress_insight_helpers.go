package v1alpha1

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/api/generic"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ generic.Insight = &ZoneIngressInsight{}

func (x *ZoneIngressInsight) GetSubscription(id string) generic.Subscription {
	return generic.GetSubscription[*DiscoverySubscription](x, id)
}

func (x *ZoneIngressInsight) UpdateSubscription(s generic.Subscription) error {
	if x == nil {
		return nil
	}
	discoverySubscription, ok := s.(*DiscoverySubscription)
	if !ok {
		return errors.Errorf("invalid type %T for ZoneIngressInsight", s)
	}
	for i, sub := range x.GetSubscriptions() {
		if sub.GetId() == discoverySubscription.Id {
			x.Subscriptions[i] = discoverySubscription
			return nil
		}
	}
	x.finalizeSubscriptions()
	x.Subscriptions = append(x.Subscriptions, discoverySubscription)
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

func (x *ZoneIngressInsight) AllSubscriptions() []generic.Subscription {
	return generic.AllSubscriptions[*DiscoverySubscription](x)
}

func (x *ZoneIngressInsight) GetLastSubscription() generic.Subscription {
	if len(x.GetSubscriptions()) == 0 {
		return (*DiscoverySubscription)(nil)
	}
	return x.GetSubscriptions()[len(x.GetSubscriptions())-1]
}

func (x *ZoneIngressInsight) Sum(v func(*DiscoverySubscription) uint64) uint64 {
	var result uint64
	for _, s := range x.GetSubscriptions() {
		result += v(s)
	}
	return result
}
