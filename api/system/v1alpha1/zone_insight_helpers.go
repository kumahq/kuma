package v1alpha1

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/kumahq/kuma/api/helpers"
)

func NewSubscriptionStatus() *KDSSubscriptionStatus {
	return &KDSSubscriptionStatus{
		Total: &KDSServiceStats{},
		Stat:  map[string]*KDSServiceStats{},
	}
}

func (x *ZoneInsight) GetSubscription(id string) (int, *KDSSubscription) {
	for i, s := range x.GetSubscriptions() {
		if s.Id == id {
			return i, s
		}
	}
	return -1, nil
}

// todo(lobkovilya): delete GetLatestSubscription, use GetLastSubscription instead
func (x *ZoneInsight) GetLatestSubscription() (*KDSSubscription, *time.Time) {
	if len(x.GetSubscriptions()) == 0 {
		return nil, nil
	}
	var idx = 0
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

func (x *ZoneInsight) GetLastSubscription() helpers.Subscription {
	if len(x.GetSubscriptions()) == 0 {
		return nil
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

func (x *KDSSubscription) SetCandidateForDisconnect(b bool) {
	x.CandidateForDisconnect = b
}

func (x *KDSSubscription) SetDisconnectTime(time time.Time) {
	x.DisconnectTime = timestamppb.New(time)
}

func (x *ZoneInsight) Sum(v func(*KDSSubscription) uint64) uint64 {
	var result uint64 = 0
	for _, s := range x.GetSubscriptions() {
		result += v(s)
	}
	return result
}

func (x *ZoneInsight) UpdateSubscription(s helpers.Subscription) {
	if x == nil {
		return
	}
	kdsSubscription, ok := s.(*KDSSubscription)
	if !ok {
		return
	}
	i, old := x.GetSubscription(kdsSubscription.Id)
	if old != nil {
		x.Subscriptions[i] = kdsSubscription
	} else {
		x.finalizeSubscriptions()
		x.Subscriptions = append(x.Subscriptions, kdsSubscription)
	}
}

// If Global CP was killed ungracefully then we can get a subscription without a DisconnectTime.
// Because of the way we process subscriptions the lack of DisconnectTime on old subscription
// will cause wrong status.
func (x *ZoneInsight) finalizeSubscriptions() {
	now := timestamppb.Now()
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
