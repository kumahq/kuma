package xds

import (
	"github.com/golang/protobuf/ptypes"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache"

	util_error "github.com/Kong/kuma/pkg/util/error"
)

// ResourcePayload is a convenience type alias.
type ResourcePayload = envoy_cache.Resource

// Resource represents a generic xDS resource with name and version.
type Resource struct {
	Name     string
	Version  string
	Resource ResourcePayload
}

// ResourceList represents a list of generic xDS resources.
type ResourceList []*Resource

func (rs ResourceList) ToDeltaDiscoveryResponse() *envoy.DeltaDiscoveryResponse {
	resp := &envoy.DeltaDiscoveryResponse{}
	for _, r := range rs {
		pbany, err := ptypes.MarshalAny(r.Resource)
		util_error.MustNot(err)
		resp.Resources = append(resp.Resources, &envoy.Resource{
			Name:     r.Name,
			Version:  r.Version,
			Resource: pbany,
		})
	}
	return resp
}
