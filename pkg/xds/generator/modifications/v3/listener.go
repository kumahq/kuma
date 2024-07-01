package v3

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/runtime/protoimpl"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_metadata "github.com/kumahq/kuma/pkg/xds/envoy/metadata/v3"
)

type listenerModificator mesh_proto.ProxyTemplate_Modifications_Listener

func (l *listenerModificator) apply(resources *core_xds.ResourceSet) error {
	listener := &envoy_listener.Listener{}
	if err := util_proto.FromYAML([]byte(l.Value), listener); err != nil {
		return err
	}
	switch l.Operation {
	case mesh_proto.OpAdd:
		l.add(resources, listener)
	case mesh_proto.OpRemove:
		l.remove(resources)
	case mesh_proto.OpPatch:
		l.patch(resources, listener)
	default:
		return errors.Errorf("invalid operation: %s", l.Operation)
	}
	return nil
}

func (l *listenerModificator) patch(resources *core_xds.ResourceSet, listenerPatch *envoy_listener.Listener) {
	for _, listener := range resources.Resources(envoy_resource.ListenerType) {
		if l.listenerMatches(listener) {
			util_proto.Merge(protoimpl.X.ProtoMessageV2Of(listener.Resource), listenerPatch)
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

func (l *listenerModificator) add(resources *core_xds.ResourceSet, listener *envoy_listener.Listener) *core_xds.ResourceSet {
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
	if len(l.Match.GetTags()) > 0 {
		if listenerProto, ok := listener.Resource.(*envoy_listener.Listener); ok {
			listenerTags := envoy_metadata.ExtractTags(listenerProto.Metadata)
			if !mesh_proto.TagSelector(l.Match.GetTags()).Matches(listenerTags) {
				return false
			}
		}
	}
	return true
}
