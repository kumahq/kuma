package v1alpha1

import (
	"slices"


	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/anypb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/jsonpatch"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshproxypatch/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_metadata "github.com/kumahq/kuma/pkg/xds/envoy/metadata/v3"
)

type networkFilterModificator api.NetworkFilterMod

func (n *networkFilterModificator) apply(resources *core_xds.ResourceSet) error {
	filter := &envoy_listener.Filter{}
	if n.Value != nil {
		if err := util_proto.FromYAML([]byte(*n.Value), filter); err != nil {
			return err
		}
	}
	for _, resource := range resources.Resources(envoy_resource.ListenerType) {
		if n.listenerMatches(resource) {
			listener := resource.Resource.(*envoy_listener.Listener)
			for _, chain := range listener.FilterChains { // apply on all filter chains. We could introduce filter chain matcher as an improvement.
				switch n.Operation {
				case api.ModOpAddFirst:
					n.addFirst(chain, filter)
				case api.ModOpAddLast:
					n.addLast(chain, filter)
				case api.ModOpAddAfter:
					n.addAfter(chain, filter)
				case api.ModOpAddBefore:
					n.addBefore(chain, filter)
				case api.ModOpRemove:
					n.remove(chain)
				case api.ModOpPatch:
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
	chain.Filters = slices.DeleteFunc(chain.Filters, func(filter *envoy_listener.Filter) bool {
		return n.filterMatches(filter)
	})
}

func (n *networkFilterModificator) patch(chain *envoy_listener.FilterChain, filterPatch *envoy_listener.Filter) error {
	for _, filter := range chain.Filters {
		if n.filterMatches(filter) {
			var merged *anypb.Any
			var err error

			if len(pointer.Deref(n.JsonPatches)) > 0 {
				merged, err = jsonpatch.MergeJsonPatchAny(filter.GetTypedConfig(), pointer.Deref(n.JsonPatches))
			} else {
				merged, err = util_proto.MergeAnys(filter.GetTypedConfig(), filterPatch.GetTypedConfig())
			}

			if err != nil {
				return err
			}

			filter.ConfigType = &envoy_listener.Filter_TypedConfig{
				TypedConfig: merged,
			}
		}
	}
	return nil
}

func (n *networkFilterModificator) filterMatches(filter *envoy_listener.Filter) bool {
	if n.Match == nil {
		return true
	}
	if n.Match.Name != nil && *n.Match.Name != filter.Name {
		return false
	}
	return true
}

func (n *networkFilterModificator) listenerMatches(resource *core_xds.Resource) bool {
	if n.Match == nil {
		return true
	}
	if n.Match.ListenerName != nil && *n.Match.ListenerName != resource.Name {
		return false
	}
	if n.Match.Origin != nil && *n.Match.Origin != resource.Origin {
		return false
	}
	if len(pointer.Deref(n.Match.ListenerTags)) > 0 {
		if listenerProto, ok := resource.Resource.(*envoy_listener.Listener); ok {
			listenerTags := envoy_metadata.ExtractTags(listenerProto.Metadata)
			if !mesh_proto.TagSelector(pointer.Deref(n.Match.ListenerTags)).Matches(listenerTags) {
				return false
			}
		}
	}
	return true
}

func (n *networkFilterModificator) indexOfMatchedFilter(chain *envoy_listener.FilterChain) int {
	for i, filter := range chain.Filters {
		if n.Match != nil && filter.Name == *n.Match.Name {
			return i
		}
	}
	return -1
}
