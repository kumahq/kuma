package v1alpha1

import (
	"time"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v2"
	"github.com/golang/protobuf/ptypes"
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
	ts, err := ptypes.TimestampProto(expiration)
	if err != nil {
		return err
	}
	ds.MTLS.CertificateExpirationTime = ts
	ds.MTLS.CertificateRegenerations++
	ts, err = ptypes.TimestampProto(generation)
	if err != nil {
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
		ds.Subscriptions = append(ds.Subscriptions, s)
	}
}

func (ds *DataplaneInsight) GetLatestSubscription() (*DiscoverySubscription, *time.Time) {
	if len(ds.GetSubscriptions()) == 0 {
		return nil, nil
	}
	var idx int = 0
	var latest *time.Time
	for i, s := range ds.GetSubscriptions() {
		t, err := ptypes.Timestamp(s.ConnectTime)
		if err != nil {
			continue
		}
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
	switch typeUrl {
	case envoy_resource.ClusterType:
		if s.Cds == nil {
			s.Cds = &DiscoveryServiceStats{}
		}
		return s.Cds
	case envoy_resource.EndpointType:
		if s.Eds == nil {
			s.Eds = &DiscoveryServiceStats{}
		}
		return s.Eds
	case envoy_resource.ListenerType:
		if s.Lds == nil {
			s.Lds = &DiscoveryServiceStats{}
		}
		return s.Lds
	case envoy_resource.RouteType:
		if s.Rds == nil {
			s.Rds = &DiscoveryServiceStats{}
		}
		return s.Rds
	default:
		return &DiscoveryServiceStats{}
	}
}
