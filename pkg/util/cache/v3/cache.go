package v3

import (
	"sort"

	v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	ctl_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"google.golang.org/protobuf/runtime/protoimpl"
	"google.golang.org/protobuf/types/known/anypb"
)

func ToDeltaDiscoveryResponse(s ctl_cache.Snapshot) (*v3.DeltaDiscoveryResponse, error) {
	resp := &v3.DeltaDiscoveryResponse{}
	for _, rs := range s.Resources {
		for _, name := range sortedResourceNames(rs) {
			r := rs.Items[name]
			pbany, err := anypb.New(protoimpl.X.ProtoMessageV2Of(r.Resource))
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
