package v1alpha1

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/kumahq/kuma/pkg/util/pointer"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/jsonpatch"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshproxypatch/api/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_metadata "github.com/kumahq/kuma/pkg/xds/envoy/metadata/v3"
)

type listenerModificator api.ListenerMod

func (l *listenerModificator) apply(resources *core_xds.ResourceSet) error {
	listener := &envoy_listener.Listener{}
	if l.Value != nil {
		if err := util_proto.FromYAML([]byte(*l.Value), listener); err != nil {
			return err
		}
	}
	switch l.Operation {
	case api.ModOpAdd:
		l.add(resources, listener)
	case api.ModOpRemove:
		l.remove(resources)
	case api.ModOpPatch:
		return l.patch(resources, listener)
	default:
		return errors.Errorf("invalid operation: %s", l.Operation)
	}
	return nil
}

func (l *listenerModificator) patch(resources *core_xds.ResourceSet, listenerPatch *envoy_listener.Listener) error {
	for _, listener := range resources.Resources(envoy_resource.ListenerType) {
		if l.listenerMatches(listener) {
			if len(pointer.Deref(l.JsonPatches)) > 0 {
				if err := jsonpatch.MergeJsonPatch(listener.Resource, pointer.Deref(l.JsonPatches)); err != nil {
					return err
				}

				continue
			}

			util_proto.Merge(listener.Resource, listenerPatch)
		}
	}

	return nil
}

func (l *listenerModificator) remove(resources *core_xds.ResourceSet) {
	for name, resource := range resources.Resources(envoy_resource.ListenerType) {
		if l.listenerMatches(resource) {
			resources.Remove(envoy_resource.ListenerType, name)
		}
	}
}

func (l *listenerModificator) add(resources *core_xds.ResourceSet, listener *envoy_listener.Listener) *core_xds.ResourceSet {
	return resources.Add(&core_xds.Resource{
		Name:     listener.Name,
		Origin:   Origin,
		Resource: listener,
	})
}

func (l *listenerModificator) listenerMatches(listener *core_xds.Resource) bool {
	if l.Match == nil {
		return true
	}
	if l.Match.Name != nil && *l.Match.Name != listener.Name {
		return false
	}
	if l.Match.Origin != nil && *l.Match.Origin != listener.Origin {
		return false
	}
	if len(pointer.Deref(l.Match.Tags)) > 0 {
		if listenerProto, ok := listener.Resource.(*envoy_listener.Listener); ok {
			listenerTags := envoy_metadata.ExtractTags(listenerProto.Metadata)
			if !mesh_proto.TagSelector(pointer.Deref(l.Match.Tags)).Matches(listenerTags) {
				return false
			}
		}
	}
	return true
}
