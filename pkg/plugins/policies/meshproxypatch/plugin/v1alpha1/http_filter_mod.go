package v1alpha1

import (
	"slices"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
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

type httpFilterModificator api.HTTPFilterMod

func (h *httpFilterModificator) apply(resources *core_xds.ResourceSet) error {
	for _, resource := range resources.Resources(envoy_resource.ListenerType) {
		if h.listenerMatches(resource) {
			listener := resource.Resource.(*envoy_listener.Listener)
			for _, chain := range listener.FilterChains { // apply on all filter chains. We could introduce filter chain matcher as an improvement.
				for _, networkFilter := range chain.Filters {
					if networkFilter.Name == "envoy.filters.network.http_connection_manager" {
						hcm := &envoy_hcm.HttpConnectionManager{}
						err := util_proto.UnmarshalAnyTo(networkFilter.ConfigType.(*envoy_listener.Filter_TypedConfig).TypedConfig, hcm)
						if err != nil {
							return err
						}
						if err := h.applyHCMModification(hcm); err != nil {
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

func (h *httpFilterModificator) applyHCMModification(hcm *envoy_hcm.HttpConnectionManager) error {
	filter := &envoy_hcm.HttpFilter{}
	if h.Value != nil {
		if err := util_proto.FromYAML([]byte(*h.Value), filter); err != nil {
			return err
		}
	}
	switch h.Operation {
	case api.ModOpAddFirst:
		h.addFirst(hcm, filter)
	case api.ModOpAddLast:
		h.addLast(hcm, filter)
	case api.ModOpAddAfter:
		h.addAfter(hcm, filter)
	case api.ModOpAddBefore:
		h.addBefore(hcm, filter)
	case api.ModOpRemove:
		h.remove(hcm)
	case api.ModOpPatch:
		if err := h.patch(hcm, filter); err != nil {
			return errors.Wrap(err, "could not patch the resource")
		}
	default:
		return errors.Errorf("invalid operation: %s", h.Operation)
	}
	return nil
}

func (h *httpFilterModificator) patch(hcm *envoy_hcm.HttpConnectionManager, filterPatch *envoy_hcm.HttpFilter) error {
	for _, filter := range hcm.HttpFilters {
		if h.filterMatches(filter) {
			var merged *anypb.Any
			var err error

			if len(pointer.Deref(h.JsonPatches)) > 0 {
				merged, err = jsonpatch.MergeJsonPatchAny(filter.GetTypedConfig(), pointer.Deref(h.JsonPatches))
			} else {
				merged, err = util_proto.MergeAnys(filter.GetTypedConfig(), filterPatch.GetTypedConfig())
			}

			if err != nil {
				return err
			}

			filter.ConfigType = &envoy_hcm.HttpFilter_TypedConfig{
				TypedConfig: merged,
			}
		}
	}
	return nil
}

func (h *httpFilterModificator) remove(hcm *envoy_hcm.HttpConnectionManager) {
	hcm.HttpFilters = slices.DeleteFunc(hcm.HttpFilters, func(filter *envoy_hcm.HttpFilter) bool {
		return h.filterMatches(filter)
	})
}

func (h *httpFilterModificator) addBefore(hcm *envoy_hcm.HttpConnectionManager, filterMod *envoy_hcm.HttpFilter) {
	idx := h.indexOfMatchedFilter(hcm)
	if idx != -1 {
		hcm.HttpFilters = append(hcm.HttpFilters, nil)
		copy(hcm.HttpFilters[idx+1:], hcm.HttpFilters[idx:])
		hcm.HttpFilters[idx] = filterMod
	}
}

func (h *httpFilterModificator) addAfter(hcm *envoy_hcm.HttpConnectionManager, filterMod *envoy_hcm.HttpFilter) {
	idx := h.indexOfMatchedFilter(hcm)
	if idx != -1 {
		hcm.HttpFilters = append(hcm.HttpFilters, nil)
		copy(hcm.HttpFilters[idx+2:], hcm.HttpFilters[idx+1:])
		hcm.HttpFilters[idx+1] = filterMod
	}
}

func (h *httpFilterModificator) addLast(hcm *envoy_hcm.HttpConnectionManager, filterMod *envoy_hcm.HttpFilter) {
	hcm.HttpFilters = append(hcm.HttpFilters, filterMod)
}

func (h *httpFilterModificator) addFirst(hcm *envoy_hcm.HttpConnectionManager, filterMod *envoy_hcm.HttpFilter) {
	hcm.HttpFilters = append([]*envoy_hcm.HttpFilter{filterMod}, hcm.HttpFilters...)
}

func (h *httpFilterModificator) filterMatches(filter *envoy_hcm.HttpFilter) bool {
	if h.Match == nil {
		return true
	}
	if h.Match.Name != nil && *h.Match.Name != filter.Name {
		return false
	}
	return true
}

func (h *httpFilterModificator) listenerMatches(resource *core_xds.Resource) bool {
	if h.Match == nil {
		return true
	}
	if h.Match.ListenerName != nil && *h.Match.ListenerName != resource.Name {
		return false
	}
	if h.Match.Origin != nil && *h.Match.Origin != resource.Origin {
		return false
	}
	if len(pointer.Deref(h.Match.ListenerTags)) > 0 {
		if listenerProto, ok := resource.Resource.(*envoy_listener.Listener); ok {
			listenerTags := envoy_metadata.ExtractTags(listenerProto.Metadata)
			if !mesh_proto.TagSelector(pointer.Deref(h.Match.ListenerTags)).Matches(listenerTags) {
				return false
			}
		}
	}
	return true
}

func (h *httpFilterModificator) indexOfMatchedFilter(hcm *envoy_hcm.HttpConnectionManager) int {
	for i, filter := range hcm.HttpFilters {
		if h.Match != nil && filter.Name == *h.Match.Name {
			return i
		}
	}
	return -1
}
