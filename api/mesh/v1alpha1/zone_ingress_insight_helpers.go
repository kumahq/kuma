package v1alpha1

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (x *ZoneIngressInsight) GetSubscription(id string) (int, *DiscoverySubscription) {
	for i, s := range x.GetSubscriptions() {
		if s.Id == id {
			return i, s
		}
	}
	return -1, nil
}

func (x *ZoneIngressInsight) UpdateSubscription(s *DiscoverySubscription) {
	if x == nil {
		return
	}
	i, old := x.GetSubscription(s.Id)
	if old != nil {
		x.Subscriptions[i] = s
	} else {
		x.finalizeSubscriptions()
		x.Subscriptions = append(x.Subscriptions, s)
	}
}

// If Kuma CP was killed ungracefully then we can get a subscription without a DisconnectTime.
// Because of the way we process subscriptions the lack of DisconnectTime on old subscription
// will cause wrong status.
func (x *ZoneIngressInsight) finalizeSubscriptions() {
	now := timestamppb.Now()
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
