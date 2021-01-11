package v3

import (
	"sort"

	"github.com/golang/protobuf/ptypes"

	v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	ctl_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
)

func ToDeltaDiscoveryResponse(s ctl_cache.Snapshot) (*v3.DeltaDiscoveryResponse, error) {
	resp := &v3.DeltaDiscoveryResponse{}
	for _, rs := range s.Resources {
		for _, name := range sortedResourceNames(rs) {
			r := rs.Items[name]
			pbany, err := ptypes.MarshalAny(r)
			if err != nil {
				return nil, err
			}
			resp.Resources = append(resp.Resources, &v3.Resource{
				Version:  rs.Version,
				Name:     name,
				Resource: pbany,
			})
		}
	}
	return resp, nil
}

func sortedResourceNames(rs ctl_cache.Resources) []string {
	names := make([]string, 0, len(rs.Items))
	for name := range rs.Items {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
