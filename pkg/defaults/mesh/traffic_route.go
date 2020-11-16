package mesh

import (
	"context"
	"fmt"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

var (
	defaultTrafficRoute = mesh_proto.TrafficRoute{
		Sources: []*mesh_proto.Selector{{
			Match: mesh_proto.MatchAnyService(),
		}},
		Destinations: []*mesh_proto.Selector{{
			Match: mesh_proto.MatchAnyService(),
		}},
		Conf: &mesh_proto.TrafficRoute_Conf{
			Split: []*mesh_proto.TrafficRoute_Split{{
				Weight:      100,
				Destination: mesh_proto.MatchAnyService(),
			}},
		},
	}
)

// Default traffic route needs to be stored with default suffix so on K8S it will be stored in the default namespace
// This will be dropped when TrafficRoute will be converted to Global Scope on K8S instead of Namespace Scope
// TrafficRoute needs to contain mesh name inside it. Otherwise if the name is the same (ex. "allow-all") creating new mesh would fail because there is already resource of name "allow-all" which is unique key on K8S
func defaultTrafficRouteName(meshName string) string {
	return fmt.Sprintf("route-all-%s", meshName)
}

func createDefaultTrafficRoute(resManager manager.ResourceManager, meshName string) error {
	tp := &core_mesh.TrafficRouteResource{
		Spec: defaultTrafficRoute,
	}
	return resManager.Create(context.Background(), tp, store.CreateByKey(defaultTrafficRouteName(meshName), meshName))
}
