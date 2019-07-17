package v1alpha1

import (
	"time"

	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/gogo/protobuf/types"
)

func (ds *DataplaneStatus) GetSubscription(id string) (int, *DiscoverySubscription) {
	for i, s := range ds.Subscriptions {
		if s.Id == id {
			return i, s
		}
	}
	return -1, nil
}

func (ds *DataplaneStatus) UpdateSubscription(s *DiscoverySubscription) {
	i, old := ds.GetSubscription(s.Id)
	if old != nil {
		ds.Subscriptions[i] = s
	} else {
		ds.Subscriptions = append(ds.Subscriptions, s)
	}
}

func (ds *DataplaneStatus) GetLatestSubscription() (*DiscoverySubscription, *time.Time) {
	if len(ds.Subscriptions) == 0 {
		return nil, nil
	}
	var idx int = 0
	var latest *time.Time
	for i, s := range ds.Subscriptions {
		t, err := types.TimestampFromProto(s.ConnectTime)
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

func (ds *DataplaneStatus) Sum(v func(*DiscoverySubscription) uint64) uint64 {
	var result uint64 = 0
	for _, s := range ds.Subscriptions {
		result += v(s)
	}
	return result
}

func (s *DiscoverySubscriptionStatus) StatsOf(typeUrl string) *DiscoveryServiceStats {
	switch typeUrl {
	case envoy_cache.ClusterType:
		return &s.Cds
	case envoy_cache.EndpointType:
		return &s.Eds
	case envoy_cache.ListenerType:
		return &s.Lds
	case envoy_cache.RouteType:
		return &s.Rds
	default:
		return &DiscoveryServiceStats{}
	}
}
