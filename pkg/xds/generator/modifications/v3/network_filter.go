package v3

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_metadata "github.com/kumahq/kuma/pkg/xds/envoy/metadata/v3"
)

type networkFilterModificator mesh_proto.ProxyTemplate_Modifications_NetworkFilter

func (n *networkFilterModificator) apply(resources *core_xds.ResourceSet) error {
	filter := &envoy_listener.Filter{}
	if err := util_proto.FromYAML([]byte(n.Value), filter); err != nil {
		return err
	}
	for _, resource := range resources.Resources(envoy_resource.ListenerType) {
		if n.listenerMatches(resource) {
			listener := resource.Resource.(*envoy_listener.Listener)
			for _, chain := range listener.FilterChains { // apply on all filter chains. We could introduce filter chain matcher as an improvement.
				switch n.Operation {
				case mesh_proto.OpAddFirst:
					n.addFirst(chain, filter)
				case mesh_proto.OpAddLast:
					n.addLast(chain, filter)
				case mesh_proto.OpAddAfter:
					n.addAfter(chain, filter)
				case mesh_proto.OpAddBefore:
					n.addBefore(chain, filter)
				case mesh_proto.OpRemove:
					n.remove(chain)
				case mesh_proto.OpPatch:
					if err := n.patch(chain, filter); err != nil {
						return errors.Wrap(err, "could not patch the resource")
					}
				default:
					return errors.Errorf("invalid operation: %s", n.Operation)
				}
			}
		}
	}
	return nil
}

func (n *networkFilterModificator) addFirst(chain *envoy_listener.FilterChain, filter *envoy_listener.Filter) {
	chain.Filters = append([]*envoy_listener.Filter{filter}, chain.Filters...)
}

func (n *networkFilterModificator) addLast(chain *envoy_listener.FilterChain, filter *envoy_listener.Filter) {
	chain.Filters = append(chain.Filters, filter)
}

func (n *networkFilterModificator) addAfter(chain *envoy_listener.FilterChain, filter *envoy_listener.Filter) {
	idx := n.indexOfMatchedFilter(chain)
	if idx != -1 {
		chain.Filters = append(chain.Filters, nil)
		copy(chain.Filters[idx+2:], chain.Filters[idx+1:])
		chain.Filters[idx+1] = filter
	}
}

func (n *networkFilterModificator) addBefore(chain *envoy_listener.FilterChain, filter *envoy_listener.Filter) {
	idx := n.indexOfMatchedFilter(chain)
	if idx != -1 {
		chain.Filters = append(chain.Filters, nil)
		copy(chain.Filters[idx+1:], chain.Filters[idx:])
		chain.Filters[idx] = filter
	}
}

func (n *networkFilterModificator) remove(chain *envoy_listener.FilterChain) {
	var filters []*envoy_listener.Filter
	for _, filter := range chain.Filters {
		if !n.filterMatches(filter) {
			filters = append(filters, filter)
		}
	}
	chain.Filters = filters
}

func (n *networkFilterModificator) patch(chain *envoy_listener.FilterChain, filterPatch *envoy_listener.Filter) error {
	for _, filter := range chain.Filters {
		if n.filterMatches(filter) {
			any, err := util_proto.MergeAnys(filter.GetTypedConfig(), filterPatch.GetTypedConfig())
			if err != nil {
				return err
			}

			filter.ConfigType = &envoy_listener.Filter_TypedConfig{
				TypedConfig: any,
			}
		}
	}
	return nil
}

func (n *networkFilterModificator) filterMatches(filter *envoy_listener.Filter) bool {
	if n.Match.GetName() != "" && n.Match.GetName() != filter.Name {
		return false
	}
	return true
}

func (n *networkFilterModificator) listenerMatches(resource *core_xds.Resource) bool {
	if n.Match.GetListenerName() != "" && n.Match.GetListenerName() != resource.Name {
		return false
	}
	if n.Match.GetOrigin() != "" && n.Match.GetOrigin() != string(resource.Origin) {
		return false
	}
	if len(n.Match.GetListenerTags()) > 0 {
		if listenerProto, ok := resource.Resource.(*envoy_listener.Listener); ok {
			listenerTags := envoy_metadata.ExtractTags(listenerProto.Metadata)
			if !mesh_proto.TagSelector(n.Match.GetListenerTags()).Matches(listenerTags) {
				return false
			}
		}
	}
	return true
}

func (n *networkFilterModificator) indexOfMatchedFilter(chain *envoy_listener.FilterChain) int {
	for i, filter := range chain.Filters {
		if filter.Name == n.Match.Name {
			return i
		}
	}
	return -1
}
