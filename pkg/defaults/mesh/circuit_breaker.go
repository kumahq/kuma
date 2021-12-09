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
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var defaultCircuitBreaker = mesh_proto.CircuitBreaker{
	Sources: []*mesh_proto.Selector{{
		Match: mesh_proto.MatchAnyService(),
	}},
	Destinations: []*mesh_proto.Selector{{
		Match: mesh_proto.MatchAnyService(),
	}},
	Conf: &mesh_proto.CircuitBreaker_Conf{
		Thresholds: &mesh_proto.CircuitBreaker_Conf_Thresholds{
			MaxConnections:     util_proto.UInt32(1024),
			MaxPendingRequests: util_proto.UInt32(1024),
			MaxRequests:        util_proto.UInt32(1024),
			MaxRetries:         util_proto.UInt32(3),
		},
	},
}

// CircuitBreaker needs to contain mesh name inside it. Otherwise if the name is the same (ex. "allow-all") creating new mesh would fail because there is already resource of name "allow-all" which is unique key on K8S
func defaultCircuitBreakerKey(meshName string) model.ResourceKey {
	return model.ResourceKey{
		Mesh: meshName,
		Name: fmt.Sprintf("circuit-breaker-all-%s", meshName),
	}
}

func ensureDefaultCircuitBreaker(resManager manager.ResourceManager, meshName string) (err error, created bool) {
	circuitBreaker := &core_mesh.CircuitBreakerResource{
		Spec: &defaultCircuitBreaker,
	}
	key := defaultCircuitBreakerKey(meshName)
	err = resManager.Get(context.Background(), circuitBreaker, store.GetBy(key))
	if err == nil {
		return nil, false
	}
	if !store.IsResourceNotFound(err) {
		return errors.Wrap(err, "could not retrieve a resource"), false
	}
	if err := resManager.Create(context.Background(), circuitBreaker, store.CreateBy(key)); err != nil {
		return errors.Wrap(err, "could not create a resource"), false
	}
	return nil, true
}
