package mesh

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"

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
			Destination: mesh_proto.MatchAnyService(),
			LoadBalancer: &mesh_proto.TrafficRoute_LoadBalancer{
				LbType: &mesh_proto.TrafficRoute_LoadBalancer_RoundRobin_{},
			},
		},
	}
)

// TrafficRoute needs to contain mesh name inside it. Otherwise if the name is the same (ex. "allow-all") creating new mesh would fail because there is already resource of name "allow-all" which is unique key on K8S
func defaultTrafficRouteKey(meshName string) model.ResourceKey {
	return model.ResourceKey{
		Mesh: meshName,
		Name: fmt.Sprintf("route-all-%s", meshName),
	}
}

func ensureDefaultTrafficRoute(resManager manager.ResourceManager, meshName string) (err error, created bool) {
	tr := &core_mesh.TrafficRouteResource{
		Spec: &defaultTrafficRoute,
	}
	key := defaultTrafficRouteKey(meshName)
	err = resManager.Get(context.Background(), tr, store.GetBy(key))
	if err == nil {
		return nil, false
	}
	if !store.IsResourceNotFound(err) {
		return errors.Wrap(err, "could not retrieve a resource"), false
	}
	if err := resManager.Create(context.Background(), tr, store.CreateBy(key)); err != nil {
		return errors.Wrap(err, "could not create a resource"), false
	}
	return nil, true
}
