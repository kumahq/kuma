package v1alpha1

import (
	"time"

	"github.com/golang/protobuf/ptypes"
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
		t, err := ptypes.Timestamp(s.ConnectTime)
		if err != nil {
			continue
		}
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
		m.Subscriptions = append(m.Subscriptions, s)
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
