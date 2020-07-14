package modifications

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v2"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	model "github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func applyListenerModification(resources *model.ResourceSet, modification *mesh_proto.ProxyTemplate_Modifications_Listener) error {
	listenerMod := &envoy_api.Listener{}
	if err := util_proto.FromYAML([]byte(modification.Value), listenerMod); err != nil {
		return err
	}
	switch modification.Operation {
	case mesh_proto.OpAdd:
		resources.Add(&model.Resource{
			Name:        listenerMod.Name,
			GeneratedBy: GeneratedByProxyTemplateModifications,
			Resource:    listenerMod,
		})
	case mesh_proto.OpRemove:
		for name, resource := range resources.Resources(envoy_resource.ListenerType) {
			if listenerMatches(resource, modification.Match) {
				resources.Remove(envoy_resource.ListenerType, name)
			}
		}
	case mesh_proto.OpPatch:
		for _, listener := range resources.Resources(envoy_resource.ListenerType) {
			if listenerMatches(listener, modification.Match) {
				proto.Merge(listener.Resource, listenerMod)
			}
		}
	default:
		return errors.New("invalid operation")
	}
	return nil
}

func listenerMatches(listener *model.Resource, match *mesh_proto.ProxyTemplate_Modifications_Listener_Match) bool {
	if match == nil {
		return true
	}
	if match.Name == listener.Name {
		return true
	}
	if match.Direction == listener.GeneratedBy {
		return true
	}
	return false
}
