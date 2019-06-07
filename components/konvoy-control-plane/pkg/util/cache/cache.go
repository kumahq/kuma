package cache

import (
	"sort"

	"github.com/gogo/protobuf/types"

	util_error "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/error"
	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	ctl_cache "github.com/envoyproxy/go-control-plane/pkg/cache"
)

func ToDeltaDiscoveryResponse(s ctl_cache.Snapshot) *v2.DeltaDiscoveryResponse {
	resp := &v2.DeltaDiscoveryResponse{}
	for _, rs := range []ctl_cache.Resources{s.Endpoints, s.Clusters, s.Routes, s.Listeners, s.Secrets} {
		for _, name := range sortedResourceNames(rs) {
			r := rs.Items[name]
			pbany, err := types.MarshalAny(r)
			util_error.MustNot(err)
			resp.Resources = append(resp.Resources, v2.Resource{
				Version:  rs.Version,
				Name:     name,
				Resource: pbany,
			})
		}
	}
	return resp
}

func sortedResourceNames(rs ctl_cache.Resources) []string {
	names := make([]string, 0, len(rs.Items))
	for name := range rs.Items {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
