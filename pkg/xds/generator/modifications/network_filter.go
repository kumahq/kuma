package modifications

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_api_v2_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v2"
	"github.com/gogo/protobuf/proto"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	model "github.com/Kong/kuma/pkg/core/xds"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
)

func applyNetworkFilterModification(resources *model.ResourceSet, modification *mesh_proto.ProxyTemplate_Modifications_NetworkFilter) error {
	filterMod := &envoy_api_v2_listener.Filter{}
	if err := util_proto.FromYAML([]byte(modification.Value), filterMod); err != nil {
		return err
	}
	for _, resource := range resources.Resources(envoy_resource.ListenerType) {
		if listenerMatches(resource, modification.Match) {
			listener := resource.Resource.(*envoy_api.Listener)
			for _, chain := range listener.FilterChains {
				switch modification.Operation {
				case mesh_proto.OpAddFirst:
					chain.Filters = append([]*envoy_api_v2_listener.Filter{filterMod}, chain.Filters...)
				case mesh_proto.OpAddLast:
					chain.Filters = append(chain.Filters, filterMod)
				case mesh_proto.OpAddAfter:
					idx := indexOfMatchedFilter(chain, modification.Match)
					if idx != -1 {
						chain.Filters = append(chain.Filters, nil)
						copy(chain.Filters[idx+2:], chain.Filters[idx+1:])
						chain.Filters[idx+1] = filterMod
					}
				case mesh_proto.OpAddBefore:
					idx := indexOfMatchedFilter(chain, modification.Match)
					if idx != -1 {
						chain.Filters = append(chain.Filters, nil)
						copy(chain.Filters[idx+1:], chain.Filters[idx:])
						chain.Filters[idx] = filterMod
					}
				case mesh_proto.OpRemove:
					var filters []*envoy_api_v2_listener.Filter
					for _, filter := range chain.Filters {
						if !filterMatches(filter, modification.Match) {
							filters = append(filters, filter)
						}
					}
					chain.Filters = filters
				case mesh_proto.OpPatch:
					for _, filter := range chain.Filters {
						if filterMatches(filter, modification.Match) {
							proto.Merge(filter, filterMod)
						}
					}
				}
			}
		}
	}
	return nil
}

func filterMatches(filter *envoy_api_v2_listener.Filter, match *mesh_proto.ProxyTemplate_Modifications_NetworkFilter_Match) bool {
	if match == nil {
		return true
	}
	if match.Name == "" {
		return true
	}
	return filter.Name == match.Name
}

func listenerMatches(resource *model.Resource, match *mesh_proto.ProxyTemplate_Modifications_NetworkFilter_Match) bool {
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

func indexOfMatchedFilter(chain *envoy_api_v2_listener.FilterChain, match *mesh_proto.ProxyTemplate_Modifications_NetworkFilter_Match) int {
	for i, filter := range chain.Filters {
		if filter.Name == match.Name {
			return i
		}
	}
	return -1
}
