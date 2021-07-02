package v1alpha1

import (
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func NewSubscriptionStatus() *DiscoverySubscriptionStatus {
	return &DiscoverySubscriptionStatus{
		Total: &DiscoveryServiceStats{},
		Cds:   &DiscoveryServiceStats{},
		Eds:   &DiscoveryServiceStats{},
		Lds:   &DiscoveryServiceStats{},
		Rds:   &DiscoveryServiceStats{},
	}
}

func NewVersion() *Version {
	return &Version{
		KumaDp: &KumaDpVersion{
			Version:   "",
			GitTag:    "",
			GitCommit: "",
			BuildDate: "",
		},
		Envoy: &EnvoyVersion{
			Version: "",
			Build:   "",
		},
	}
}

func (ds *DataplaneInsight) IsOnline() bool {
	for _, s := range ds.GetSubscriptions() {
		if s.ConnectTime != nil && s.DisconnectTime == nil {
			return true
		}
	}
	return false
}

func (ds *DataplaneInsight) GetSubscription(id string) (int, *DiscoverySubscription) {
	for i, s := range ds.GetSubscriptions() {
		if s.Id == id {
			return i, s
		}
	}
	return -1, nil
}

func (ds *DataplaneInsight) UpdateCert(generation time.Time, expiration time.Time) error {
	if ds.MTLS == nil {
		ds.MTLS = &DataplaneInsight_MTLS{}
	}
	ts := timestamppb.New(expiration)
	if err := ts.CheckValid(); err != nil {
		return err
	}
	ds.MTLS.CertificateExpirationTime = ts
	ds.MTLS.CertificateRegenerations++
	ts = timestamppb.New(generation)
	if err := ts.CheckValid(); err != nil {
		return err
	}
	ds.MTLS.LastCertificateRegeneration = ts
	return nil
}

func (ds *DataplaneInsight) UpdateSubscription(s *DiscoverySubscription) {
	if ds == nil {
		return
	}
	i, old := ds.GetSubscription(s.Id)
	if old != nil {
		ds.Subscriptions[i] = s
	} else {
		ds.finalizeSubscriptions()
		ds.Subscriptions = append(ds.Subscriptions, s)
	}
}

// If Kuma CP was killed ungracefully then we can get a subscription without a DisconnectTime.
// Because of the way we process subscriptions the lack of DisconnectTime on old subscription
// will cause wrong status.
func (ds *DataplaneInsight) finalizeSubscriptions() {
	now := timestamppb.Now()
	for _, subscription := range ds.GetSubscriptions() {
		if subscription.DisconnectTime == nil {
			subscription.DisconnectTime = now
		}
	}
}

func (ds *DataplaneInsight) GetLatestSubscription() (*DiscoverySubscription, *time.Time) {
	if len(ds.GetSubscriptions()) == 0 {
		return nil, nil
	}
	var idx int = 0
	var latest *time.Time
	for i, s := range ds.GetSubscriptions() {
		if err := s.ConnectTime.CheckValid(); err != nil {
			continue
		}
		t := s.ConnectTime.AsTime()
		if latest == nil || latest.Before(t) {
			idx = i
			latest = &t
		}
	}
	return ds.Subscriptions[idx], latest
}

func (ds *DataplaneInsight) Sum(v func(*DiscoverySubscription) uint64) uint64 {
	var result uint64 = 0
	for _, s := range ds.GetSubscriptions() {
		result += v(s)
	}
	return result
}

func (s *DiscoverySubscriptionStatus) StatsOf(typeUrl string) *DiscoveryServiceStats {
	if s == nil {
		return &DiscoveryServiceStats{}
	}
	// we rely on type URL suffix to get rid of the dependency on concrete V2 / V3 implementation
	switch {
	case strings.HasSuffix(typeUrl, "Cluster"):
		if s.Cds == nil {
			s.Cds = &DiscoveryServiceStats{}
		}
		return s.Cds
	case strings.HasSuffix(typeUrl, "ClusterLoadAssignment"):
		if s.Eds == nil {
			s.Eds = &DiscoveryServiceStats{}
		}
		return s.Eds
	case strings.HasSuffix(typeUrl, "Listener"):
		if s.Lds == nil {
			s.Lds = &DiscoveryServiceStats{}
		}
		return s.Lds
	case strings.HasSuffix(typeUrl, "RouteConfiguration"):
		if s.Rds == nil {
			s.Rds = &DiscoveryServiceStats{}
		}
		return s.Rds
	default:
		return &DiscoveryServiceStats{}
	}
}
