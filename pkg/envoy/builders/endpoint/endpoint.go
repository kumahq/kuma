package endpoint

import (
	envoy_endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	"google.golang.org/protobuf/types/known/wrapperspb"

	. "github.com/kumahq/kuma/pkg/envoy/builders/common"
)

// NonLocalPriority sets priority for all non-local locality endpoints
func NonLocalPriority(isLocalityAware bool, localZone string) Configurer[envoy_endpoint.ClusterLoadAssignment] {
	return func(cla *envoy_endpoint.ClusterLoadAssignment) error {
		var priority uint32
		if isLocalityAware {
			priority = 1
		}
		for _, localityLbEndpoints := range cla.Endpoints {
			if localityLbEndpoints.Locality != nil && localityLbEndpoints.Locality.Zone != localZone {
				localityLbEndpoints.Priority = priority
			}
		}
		return nil
	}
}

func Endpoints(endpoints []*envoy_endpoint.LocalityLbEndpoints) Configurer[envoy_endpoint.ClusterLoadAssignment] {
	return func(cla *envoy_endpoint.ClusterLoadAssignment) error {
		cla.Endpoints = endpoints
		return nil
	}
}

func OverprovisioningFactor(factor uint32) Configurer[envoy_endpoint.ClusterLoadAssignment] {
	return func(cla *envoy_endpoint.ClusterLoadAssignment) error {
		if cla.Policy == nil {
			cla.Policy = &envoy_endpoint.ClusterLoadAssignment_Policy{}
		}
		cla.Policy.OverprovisioningFactor = wrapperspb.UInt32(factor)
		return nil
	}
}
