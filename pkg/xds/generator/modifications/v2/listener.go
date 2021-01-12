package v2

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v2"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

type listenerModificator mesh_proto.ProxyTemplate_Modifications_Listener

func (l *listenerModificator) apply(resources *core_xds.ResourceSet) error {
	listener := &envoy_api.Listener{}
	if err := util_proto.FromYAML([]byte(l.Value), listener); err != nil {
		return err
	}
	switch l.Operation {
	case mesh_proto.OpAdd:
		l.app(resources, listener)
	case mesh_proto.OpRemove:
		l.remove(resources)
	case mesh_proto.OpPatch:
		l.patch(resources, listener)
	default:
		return errors.Errorf("invalid operation: %s", l.Operation)
	}
	return nil
}

func (l *listenerModificator) patch(resources *core_xds.ResourceSet, listenerPatch *envoy_api.Listener) {
	for _, listener := range resources.Resources(envoy_resource.ListenerType) {
		if l.listenerMatches(listener) {
			proto.Merge(listener.Resource, listenerPatch)
		}
	}
}

func (l *listenerModificator) remove(resources *core_xds.ResourceSet) {
	for name, resource := range resources.Resources(envoy_resource.ListenerType) {
		if l.listenerMatches(resource) {
			resources.Remove(envoy_resource.ListenerType, name)
		}
	}
}

func (l *listenerModificator) app(resources *core_xds.ResourceSet, listener *envoy_api.Listener) *core_xds.ResourceSet {
	return resources.Add(&core_xds.Resource{
		Name:     listener.Name,
		Origin:   OriginProxyTemplateModifications,
		Resource: listener,
	})
}

func (l *listenerModificator) listenerMatches(listener *core_xds.Resource) bool {
	if l.Match.GetName() != "" && l.Match.GetName() != listener.Name {
		return false
	}
	if l.Match.GetOrigin() != "" && l.Match.GetOrigin() != listener.Origin {
		return false
	}
	return true
}
