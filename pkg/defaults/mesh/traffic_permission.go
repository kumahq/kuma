package mesh

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

var defaultTrafficPermission = mesh_proto.TrafficPermission{
	Sources: []*mesh_proto.Selector{
		{
			Match: map[string]string{
				mesh_proto.ServiceTag: "*",
			},
		},
	},
	Destinations: []*mesh_proto.Selector{
		{
			Match: map[string]string{
				mesh_proto.ServiceTag: "*",
			},
		},
	},
}

// TrafficPermission needs to contain mesh name inside it. Otherwise if the name is the same (ex. "allow-all") creating new mesh would fail because there is already resource of name "allow-all" which is unique key on K8S
func defaultTrafficPermissionKey(meshName string) model.ResourceKey {
	return model.ResourceKey{
		Mesh: meshName,
		Name: fmt.Sprintf("allow-all-%s", meshName),
	}
}

func ensureDefaultTrafficPermission(resManager manager.ResourceManager, meshName string) (err error, created bool) {
	key := defaultTrafficPermissionKey(meshName)
	tp := &core_mesh.TrafficPermissionResource{
		Spec: &defaultTrafficPermission,
	}
	err = resManager.Get(context.Background(), tp, store.GetBy(key))
	if err == nil {
		return nil, false
	}
	if !store.IsResourceNotFound(err) {
		return errors.Wrap(err, "could not retrieve a resource"), false
	}
	if err := resManager.Create(context.Background(), tp, store.CreateBy(key)); err != nil {
		return errors.Wrap(err, "could not create a resource"), false
	}
	return nil, true
}
