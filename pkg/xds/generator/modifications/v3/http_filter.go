package v3

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_metadata "github.com/kumahq/kuma/pkg/xds/envoy/metadata/v3"
)

type httpFilterModificator mesh_proto.ProxyTemplate_Modifications_HttpFilter

func (h *httpFilterModificator) apply(resources *core_xds.ResourceSet) error {
	for _, resource := range resources.Resources(envoy_resource.ListenerType) {
		if h.listenerMatches(resource) {
			listener := resource.Resource.(*envoy_listener.Listener)
			for _, chain := range listener.FilterChains { // apply on all filter chains. We could introduce filter chain matcher as an improvement.
				for _, networkFilter := range chain.Filters {
					if networkFilter.Name != "envoy.filters.network.http_connection_manager" {
						continue
					}
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
	return nil
}

func (h *httpFilterModificator) applyHCMModification(hcm *envoy_hcm.HttpConnectionManager) error {
	filter := &envoy_hcm.HttpFilter{}
	if err := util_proto.FromYAML([]byte(h.Value), filter); err != nil {
		return err
	}
	switch h.Operation {
	case mesh_proto.OpAddFirst:
		h.addFirst(hcm, filter)
	case mesh_proto.OpAddLast:
		h.addLast(hcm, filter)
	case mesh_proto.OpAddAfter:
		h.addAfter(hcm, filter)
	case mesh_proto.OpAddBefore:
		h.addBefore(hcm, filter)
	case mesh_proto.OpRemove:
		h.remove(hcm)
	case mesh_proto.OpPatch:
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
			any, err := util_proto.MergeAnys(filter.GetTypedConfig(), filterPatch.GetTypedConfig())
			if err != nil {
				return err
			}

			filter.ConfigType = &envoy_hcm.HttpFilter_TypedConfig{
				TypedConfig: any,
			}
		}
	}
	return nil
}

func (h *httpFilterModificator) remove(hcm *envoy_hcm.HttpConnectionManager) {
	var filters []*envoy_hcm.HttpFilter
	for _, filter := range hcm.HttpFilters {
		if !h.filterMatches(filter) {
			filters = append(filters, filter)
		}
	}
	hcm.HttpFilters = filters
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
	if h.Match.GetName() != "" && h.Match.GetName() != filter.Name {
		return false
	}
	return true
}

func (h *httpFilterModificator) listenerMatches(resource *core_xds.Resource) bool {
	if h.Match.GetListenerName() != "" && h.Match.GetListenerName() != resource.Name {
		return false
	}
	if h.Match.GetOrigin() != "" && h.Match.GetOrigin() != resource.Origin {
		return false
	}
	if len(h.Match.GetListenerTags()) > 0 {
		if listenerProto, ok := resource.Resource.(*envoy_listener.Listener); ok {
			listenerTags := envoy_metadata.ExtractTags(listenerProto.Metadata)
			if !mesh_proto.TagSelector(h.Match.GetListenerTags()).Matches(listenerTags) {
				return false
			}
		}
	}
	return true
}

func (h *httpFilterModificator) indexOfMatchedFilter(hcm *envoy_hcm.HttpConnectionManager) int {
	for i, filter := range hcm.HttpFilters {
		if filter.Name == h.Match.Name {
			return i
		}
	}
	return -1
}
