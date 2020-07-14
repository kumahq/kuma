package modifications

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v2"
	envoy_wellknown "github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	model "github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func applyHTTPFilterModification(resources *model.ResourceSet, modification *mesh_proto.ProxyTemplate_Modifications_HttpFilter) error {
	for _, resource := range resources.Resources(envoy_resource.ListenerType) {
		if httpFilterListenerMatches(resource, modification.Match) {
			listener := resource.Resource.(*envoy_api.Listener)
			for _, chain := range listener.FilterChains {
				for _, networkFilter := range chain.Filters {
					if networkFilter.Name == envoy_wellknown.HTTPConnectionManager {
						hcm := &envoy_hcm.HttpConnectionManager{}
						err := ptypes.UnmarshalAny(networkFilter.ConfigType.(*envoy_listener.Filter_TypedConfig).TypedConfig, hcm)
						if err != nil {
							return err
						}
						if err := applyHCMModification(hcm, modification); err != nil {
							return err
						}
						any, err := util_proto.MarshalAnyDeterministic(hcm)
						if err != nil {
							return err
						}
						networkFilter.ConfigType.(*envoy_listener.Filter_TypedConfig).TypedConfig = any
					}
				}

			}
		}
	}
	return nil
}

func applyHCMModification(hcm *envoy_hcm.HttpConnectionManager, modification *mesh_proto.ProxyTemplate_Modifications_HttpFilter) error {
	filterMod := &envoy_hcm.HttpFilter{}
	if err := util_proto.FromYAML([]byte(modification.Value), filterMod); err != nil {
		return err
	}
	switch modification.Operation {
	case mesh_proto.OpAddFirst:
		hcm.HttpFilters = append([]*envoy_hcm.HttpFilter{filterMod}, hcm.HttpFilters...)
	case mesh_proto.OpAddLast:
		hcm.HttpFilters = append(hcm.HttpFilters, filterMod)
	case mesh_proto.OpAddAfter:
		idx := indexOfHttpMatchedFilter(hcm, modification.Match)
		if idx != -1 {
			hcm.HttpFilters = append(hcm.HttpFilters, nil)
			copy(hcm.HttpFilters[idx+2:], hcm.HttpFilters[idx+1:])
			hcm.HttpFilters[idx+1] = filterMod
		}
	case mesh_proto.OpAddBefore:
		idx := indexOfHttpMatchedFilter(hcm, modification.Match)
		if idx != -1 {
			hcm.HttpFilters = append(hcm.HttpFilters, nil)
			copy(hcm.HttpFilters[idx+1:], hcm.HttpFilters[idx:])
			hcm.HttpFilters[idx] = filterMod
		}
	case mesh_proto.OpRemove:
		var filters []*envoy_hcm.HttpFilter
		for _, filter := range hcm.HttpFilters {
			if !httpFilterMatches(filter, modification.Match) {
				filters = append(filters, filter)
			}
		}
		hcm.HttpFilters = filters
	case mesh_proto.OpPatch:
		for _, filter := range hcm.HttpFilters {
			if httpFilterMatches(filter, modification.Match) {
				proto.Merge(filter, filterMod)
			}
		}
	}
	return nil
}

func httpFilterMatches(filter *envoy_hcm.HttpFilter, match *mesh_proto.ProxyTemplate_Modifications_HttpFilter_Match) bool {
	if match == nil {
		return true
	}
	if match.Name == "" {
		return true
	}
	return filter.Name == match.Name
}

func httpFilterListenerMatches(resource *model.Resource, match *mesh_proto.ProxyTemplate_Modifications_HttpFilter_Match) bool {
	if match == nil {
		return true
	}
	if match.ListenerName == "" && match.Direction == "" {
		return true
	}
	if match.ListenerName == resource.Name {
		return true
	}
	if match.Direction == resource.GeneratedBy {
		return true
	}
	return false
}

func indexOfHttpMatchedFilter(hcm *envoy_hcm.HttpConnectionManager, match *mesh_proto.ProxyTemplate_Modifications_HttpFilter_Match) int {
	for i, filter := range hcm.HttpFilters {
		if filter.Name == match.Name {
			return i
		}
	}
	return -1
}
