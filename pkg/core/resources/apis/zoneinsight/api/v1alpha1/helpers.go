package v1alpha1

import (
	"time"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/v3/api/generic"
)

var _ generic.Insight = &ZoneInsight{}

func NewSubscriptionStatus(now time.Time) *KDSSubscriptionStatus {
	return &KDSSubscriptionStatus{
		LastUpdateTime: NewTime(now),
		Total:          &KDSServiceStats{},
		Stat:           map[string]*KDSServiceStats{},
	}
}

func NewVersion() *Version {
	return &Version{KumaCP: &KumaCpVersion{}}
}

func (x *ZoneInsight) GetSubscriptions() []*KDSSubscription {
	if x == nil {
		return nil
	}
	return x.Subscriptions
}

func (x *ZoneInsight) GetKdsStreams() *KDSStreams {
	if x == nil {
		return nil
	}
	return x.KDSStreams
}

func (x *ZoneInsight) GetHealthCheck() *HealthCheck {
	if x == nil {
		return nil
	}
	return x.HealthCheck
}

func (x *ZoneInsight) GetEnvoyAdminStreams() *EnvoyAdminStreams {
	if x == nil {
		return nil
	}
	return x.EnvoyAdminStreams
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

func (x *ZoneInsight) AllSubscriptions() []generic.Subscription {
	return generic.AllSubscriptions[*KDSSubscription](x)
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
		if sub.GetId() == kdsSubscription.ID {
			x.Subscriptions[i] = kdsSubscription
			return nil
		}
	}
	x.finalizeSubscriptions()
	x.Subscriptions = append(x.Subscriptions, kdsSubscription)
	return nil
}

// CompactFinished removes detailed information about finished subscriptions to trim the
// object size. The last subscription always keeps its details.
func (x *ZoneInsight) CompactFinished() {
	for i := 0; i < len(x.GetSubscriptions())-1; i++ {
		x.Subscriptions[i].Config = ""
		if status := x.Subscriptions[i].Status; status != nil {
			status.Stat = map[string]*KDSServiceStats{}
		}
	}
}

// finalizeSubscriptions closes subscriptions left open by an ungracefully killed Global
// CP. Without a DisconnectTime an old subscription makes the zone look online forever.
func (x *ZoneInsight) finalizeSubscriptions() {
	now := NewTime(time.Now())
	for _, subscription := range x.GetSubscriptions() {
		if subscription.DisconnectTime == nil {
			subscription.DisconnectTime = now
		}
	}
}

func (x *KDSStreams) GetGlobalToZone() *KDSStream {
	if x == nil {
		return nil
	}
	return x.GlobalToZone
}

func (x *KDSStreams) GetZoneToGlobal() *KDSStream {
	if x == nil {
		return nil
	}
	return x.ZoneToGlobal
}

func (x *KDSStreams) GetClusters() *KDSStream {
	if x == nil {
		return nil
	}
	return x.Clusters
}

func (x *KDSStreams) GetStats() *KDSStream {
	if x == nil {
		return nil
	}
	return x.Stats
}

func (x *KDSStreams) GetConfigDump() *KDSStream {
	if x == nil {
		return nil
	}
	return x.ConfigDump
}

func (x *KDSStream) GetGlobalInstanceID() string {
	if x == nil {
		return ""
	}
	return x.GlobalInstanceID
}

func (x *KDSStream) GetConnectTime() *Time {
	if x == nil {
		return nil
	}
	return x.ConnectTime
}

func (x *HealthCheck) GetTime() *Time {
	if x == nil {
		return nil
	}
	return x.Time
}

func (x *KDSSubscription) GetId() string {
	if x == nil {
		return ""
	}
	return x.ID
}

func (x *KDSSubscription) GetGeneration() uint32 {
	if x == nil {
		return 0
	}
	return x.Generation
}

func (x *KDSSubscription) GetGlobalInstanceID() string {
	if x == nil {
		return ""
	}
	return x.GlobalInstanceID
}

func (x *KDSSubscription) GetZoneInstanceID() string {
	if x == nil {
		return ""
	}
	return x.ZoneInstanceID
}

func (x *KDSSubscription) GetConfig() string {
	if x == nil {
		return ""
	}
	return x.Config
}

func (x *KDSSubscription) GetConnectTime() *Time {
	if x == nil {
		return nil
	}
	return x.ConnectTime
}

func (x *KDSSubscription) GetDisconnectTime() *Time {
	if x == nil {
		return nil
	}
	return x.DisconnectTime
}

func (x *KDSSubscription) GetStatus() *KDSSubscriptionStatus {
	if x == nil {
		return nil
	}
	return x.Status
}

func (x *KDSSubscription) GetVersion() *Version {
	if x == nil {
		return nil
	}
	return x.Version
}

func (x *KDSSubscription) SetDisconnectTime(t time.Time) {
	x.DisconnectTime = NewTime(t)
}

func (x *KDSSubscription) IsOnline() bool {
	return x.GetConnectTime() != nil && x.GetDisconnectTime() == nil
}

func (x *KDSSubscriptionStatus) GetLastUpdateTime() *Time {
	if x == nil {
		return nil
	}
	return x.LastUpdateTime
}

func (x *KDSSubscriptionStatus) GetTotal() *KDSServiceStats {
	if x == nil {
		return nil
	}
	return x.Total
}

func (x *KDSSubscriptionStatus) GetStat() map[string]*KDSServiceStats {
	if x == nil {
		return nil
	}
	return x.Stat
}

func (x *KDSServiceStats) GetResponsesSent() uint64 {
	if x == nil {
		return 0
	}
	return x.ResponsesSent
}

func (x *KDSServiceStats) GetResponsesAcknowledged() uint64 {
	if x == nil {
		return 0
	}
	return x.ResponsesAcknowledged
}

func (x *KDSServiceStats) GetResponsesRejected() uint64 {
	if x == nil {
		return 0
	}
	return x.ResponsesRejected
}

func (x *Version) GetKumaCp() *KumaCpVersion {
	if x == nil {
		return nil
	}
	return x.KumaCP
}

func (x *KumaCpVersion) GetVersion() string {
	if x == nil {
		return ""
	}
	return x.Version
}

func (x *KumaCpVersion) GetKumaCpGlobalCompatible() bool {
	if x == nil {
		return false
	}
	return x.KumaCpGlobalCompatible
}

func (x *EnvoyAdminStreams) GetConfigDumpGlobalInstanceID() string {
	if x == nil {
		return ""
	}
	return x.ConfigDumpGlobalInstanceID
}

func (x *EnvoyAdminStreams) GetStatsGlobalInstanceID() string {
	if x == nil {
		return ""
	}
	return x.StatsGlobalInstanceID
}

func (x *EnvoyAdminStreams) GetClustersGlobalInstanceID() string {
	if x == nil {
		return ""
	}
	return x.ClustersGlobalInstanceID
}
