package v1alpha1

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func NewSubscriptionStatus() *KDSSubscriptionStatus {
	return &KDSSubscriptionStatus{
		Total: &KDSServiceStats{},
		Stat:  map[string]*KDSServiceStats{},
	}
}

func (m *ZoneInsight) GetSubscription(id string) (int, *KDSSubscription) {
	for i, s := range m.GetSubscriptions() {
		if s.Id == id {
			return i, s
		}
	}
	return -1, nil
}

func (m *ZoneInsight) GetLatestSubscription() (*KDSSubscription, *time.Time) {
	if len(m.GetSubscriptions()) == 0 {
		return nil, nil
	}
	var idx = 0
	var latest *time.Time
	for i, s := range m.GetSubscriptions() {
		if err := s.ConnectTime.CheckValid(); err != nil {
			continue
		}
		t := s.ConnectTime.AsTime()
		if latest == nil || latest.Before(t) {
			idx = i
			latest = &t
		}
	}
	return m.Subscriptions[idx], latest
}

func (m *ZoneInsight) IsOnline() bool {
	for _, s := range m.GetSubscriptions() {
		if s.ConnectTime != nil && s.DisconnectTime == nil {
			return true
		}
	}
	return false
}

func (m *ZoneInsight) Sum(v func(*KDSSubscription) uint64) uint64 {
	var result uint64 = 0
	for _, s := range m.GetSubscriptions() {
		result += v(s)
	}
	return result
}

func (m *ZoneInsight) UpdateSubscription(s *KDSSubscription) {
	if m == nil {
		return
	}
	i, old := m.GetSubscription(s.Id)
	if old != nil {
		m.Subscriptions[i] = s
	} else {
		m.finalizeSubscriptions()
		m.Subscriptions = append(m.Subscriptions, s)
	}
}

// If Global CP was killed ungracefully then we can get a subscription without a DisconnectTime.
// Because of the way we process subscriptions the lack of DisconnectTime on old subscription
// will cause wrong status.
func (m *ZoneInsight) finalizeSubscriptions() {
	now := timestamppb.Now()
	for _, subscription := range m.GetSubscriptions() {
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
