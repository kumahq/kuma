package v1alpha1

import (
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache"
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
